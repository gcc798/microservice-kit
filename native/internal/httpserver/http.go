package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	middleware "github.com/gcc798/microservice-kit/internal/httpmiddleware"
	"github.com/gcc798/microservice-kit/internal/httpx"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/gcc798/microservice-kit/internal/telemetry"
	"github.com/gcc798/microservice-kit/internal/validator"
	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

// Server 封装 Echo HTTP 服务及其生命周期资源。
type Server struct {
	http       *http.Server   // 底层 HTTP 服务。
	log        logging.Logger // 日志记录器。
	stopWriter func()         // 停止操作日志写入器。
}

// Options 描述 HTTP 服务启动参数。
type Options struct {
	Port      int            // 监听端口。
	CORS      bool           // 是否启用跨域中间件。
	Logger    logging.Logger // 日志记录器。
	SystemAPI sysv1.API      // 操作日志写入所需的 SYS 接口，可为空。
}

// New 创建 HTTP 服务，并返回用于服务注册的业务路由列表。
func New(opts Options, setup func(*httpx.Router) error) (*Server, []registry.HTTPRoute, error) {
	e := echo.New()
	validator.Init()
	e.Validator = validator.EchoValidator{}
	e.Binder = &validator.EchoBinder{}
	e.Use(
		echootel.NewMiddlewareWithConfig(echootel.Config{
			ServerName: fmt.Sprintf("0.0.0.0:%d", opts.Port),
			Skipper: func(c *echo.Context) bool {
				return !telemetry.TraceHTTPPath(c.Request().URL.Path)
			},
		}),
		middleware.Recovery(opts.Logger),
		echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
			LogLatency: true, LogRemoteIP: true, LogMethod: true, LogURIPath: true,
			LogRoutePath: true, LogRequestID: true, LogStatus: true,
			LogValuesFunc: func(c *echo.Context, values echoMiddleware.RequestLoggerValues) error {
				fields := []zap.Field{
					zap.String("method", values.Method),
					zap.String("path", values.URIPath),
					zap.String("route", values.RoutePath),
					zap.Int("status", values.Status),
					zap.Duration("latency", values.Latency),
					zap.String("remote_ip", values.RemoteIP),
					zap.String("request_id", values.RequestID),
				}
				if values.Error != nil {
					fields = append(fields, zap.Error(values.Error))
				}
				logging.WithContext(c.Request().Context(), opts.Logger).Info("http request", fields...)
				return nil
			},
		}),
	)
	if opts.CORS {
		e.Use(middleware.CORS())
	}
	e.Use(middleware.StringIDConverter())
	stopWriter := func() {}
	if opts.SystemAPI != nil {
		writer := middleware.NewOperLogWriter(opts.SystemAPI, opts.Logger)
		stopWriter = writer.Stop
		e.Use(middleware.OperationLog(writer))
	}
	if err := setup(httpx.NewRouter(e)); err != nil {
		stopWriter()
		return nil, nil, err
	}
	addr := fmt.Sprintf(":%d", opts.Port)
	srv := &http.Server{Addr: addr, Handler: e, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, MaxHeaderBytes: 1 << 20}
	routes := make([]registry.HTTPRoute, 0, len(e.Router().Routes()))
	for _, route := range e.Router().Routes() {
		if route.Path == "/metrics" || route.Path == "/health" || strings.HasPrefix(route.Path, "/health/") {
			continue
		}
		routes = append(routes, registry.HTTPRoute{Method: route.Method, Path: route.Path})
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
	return &Server{http: srv, log: opts.Logger, stopWriter: stopWriter}, routes, nil
}

func (s *Server) Run(ctx context.Context) error {
	defer s.stopWriter()
	errs := make(chan error, 1)
	go func() {
		s.log.Info("starting http server", zap.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.http.Shutdown(shutdown)
	case err := <-errs:
		return err
	}
}
