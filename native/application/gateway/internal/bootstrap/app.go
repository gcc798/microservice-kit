package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	serviceconfig "github.com/gcc798/microservice-kit/application/gateway/internal/config"
	"github.com/gcc798/microservice-kit/application/gateway/internal/proxy"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	sharedconfig "github.com/gcc798/microservice-kit/internal/config"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/gcc798/microservice-kit/internal/telemetry"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type App struct {
	cfg      *serviceconfig.Config
	log      logging.Logger
	registry registry.Registry
	pool     *sharedtransport.ClientPool
	proxy    *proxyGateway
	server   *http.Server
	instance registry.ServiceInstance
}

type proxyGateway struct {
	http   http.Handler
	routes interface {
		RefreshRoutes(context.Context) error
		ServeHTTP(http.ResponseWriter, *http.Request)
	}
}

func New(cfg *serviceconfig.Config, log logging.Logger) (*App, error) {
	reg, err := registry.New(registry.Options{
		Driver: cfg.Registry.Driver, Address: cfg.Registry.Address, Prefix: cfg.Registry.Prefix,
		Namespace: cfg.Registry.Namespace, Group: cfg.Registry.Group, Username: cfg.Registry.Username, Password: cfg.Registry.Password,
	})
	if err != nil {
		return nil, err
	}
	pool := sharedtransport.NewClientPool(reg)
	gateway := proxy.New(reg, iamv1.NewCached(iamv1.NewRemote(pool), 5*time.Second))
	handler := withMiddleware(gateway, cfg, log)
	handler = otelhttp.NewHandler(handler, "gateway HTTP", otelhttp.WithFilter(func(request *http.Request) bool {
		return telemetry.TraceHTTPPath(request.URL.Path)
	}))
	return &App{
		cfg: cfg, log: log, registry: reg, pool: pool,
		proxy:  &proxyGateway{http: handler, routes: gateway},
		server: &http.Server{Addr: fmt.Sprintf(":%d", cfg.Server.Port), Handler: handler, ReadHeaderTimeout: 10 * time.Second, MaxHeaderBytes: 1 << 20},
	}, nil
}

func (a *App) Run(ctx context.Context) (err error) {
	if err := a.proxy.routes.RefreshRoutes(ctx); err != nil {
		a.log.Warn("initial gateway route discovery failed", zap.Error(err))
	}
	go a.refreshLoop(ctx)
	errorsCh := make(chan error, 1)
	go func() {
		var serveErr error
		if a.cfg.Server.TLSCertFile != "" || a.cfg.Server.TLSKeyFile != "" {
			serveErr = a.server.ListenAndServeTLS(a.cfg.Server.TLSCertFile, a.cfg.Server.TLSKeyFile)
		} else {
			serveErr = a.server.ListenAndServe()
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errorsCh <- serveErr
		}
	}()
	scheme := "http"
	if a.cfg.Server.TLSCertFile != "" {
		scheme = "https"
	}
	instance, err := sharedtransport.RegisterService(ctx, a.registry, string(sharedconfig.ServiceGateway), a.cfg.Service.ID, map[string]string{
		registry.EndpointHTTP: fmt.Sprintf("%s://%s:%d", scheme, a.cfg.Service.AdvertiseHost, a.cfg.Server.Port),
	})
	if err != nil {
		return err
	}
	a.instance = instance
	select {
	case err := <-errorsCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (a *App) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := a.proxy.routes.RefreshRoutes(ctx); err != nil {
				a.log.Warn("gateway route refresh failed", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var err error
	if a.instance.ID != "" {
		err = errors.Join(err, a.registry.Deregister(shutdown, a.instance))
	}
	if a.server != nil {
		err = errors.Join(err, a.server.Shutdown(shutdown))
	}
	if a.pool != nil {
		err = errors.Join(err, a.pool.Close())
	}
	if a.registry != nil {
		err = errors.Join(err, a.registry.Close())
	}
	return err
}

type rateLimiter struct {
	mu     sync.Mutex
	window time.Time
	counts map[string]int
	limit  int
}

func withMiddleware(next http.Handler, cfg *serviceconfig.Config, log logging.Logger) http.Handler {
	limiter := &rateLimiter{window: time.Now(), counts: make(map[string]int), limit: cfg.Gateway.RateLimitPerMinute}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		requestID := request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
			request.Header.Set("X-Request-ID", requestID)
		}
		writer.Header().Set("X-Request-ID", requestID)
		defer func() {
			logging.WithContext(request.Context(), log).Info("http request", zap.String("method", request.Method), zap.String("path", request.URL.Path), zap.Duration("latency", time.Since(started)), zap.String("request_id", requestID))
		}()
		if cfg.CORS.Enabled {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, clientid, X-Request-ID")
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusNoContent)
				return
			}
		}
		host, _, _ := net.SplitHostPort(request.RemoteAddr)
		if !limiter.allow(host) {
			writer.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func (l *rateLimiter) allow(key string) bool {
	if l.limit <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if time.Since(l.window) >= time.Minute {
		l.window = time.Now()
		clear(l.counts)
	}
	l.counts[key]++
	return l.counts[key] <= l.limit
}
