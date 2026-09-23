package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	serviceconfig "github.com/gcc798/microservice-kit/application/realtime/internal/config"
	"github.com/gcc798/microservice-kit/application/realtime/internal/domain"
	"github.com/gcc798/microservice-kit/application/realtime/internal/relay"
	"github.com/gcc798/microservice-kit/application/realtime/internal/server"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	realtimev1 "github.com/gcc798/microservice-kit/internal/api/realtime/v1"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	redisplatform "github.com/gcc798/microservice-kit/internal/platform/redis"
	websocketx "github.com/gcc798/microservice-kit/internal/platform/websocket"
	"github.com/gcc798/microservice-kit/internal/registry"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type App struct {
	log        logging.Logger
	redis      *redis.Client
	registry   registry.Registry
	clientPool *sharedtransport.ClientPool
	relay      *relay.Relay
	hub        *websocketx.Hub
	httpServer *http.Server
	grpc       sharedtransport.RegisteredGRPCOptions
	runCancel  context.CancelFunc
}

func New(cfg *serviceconfig.Config, log logging.Logger) (*App, error) {
	redisClient := redisplatform.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		_ = redisClient.Close()
		return nil, err
	}
	reg, err := registry.New(registry.Options{
		Driver: cfg.Registry.Driver, Address: cfg.Registry.Address, Prefix: cfg.Registry.Prefix,
		Namespace: cfg.Registry.Namespace, Group: cfg.Registry.Group, Username: cfg.Registry.Username, Password: cfg.Registry.Password,
	})
	if err != nil {
		_ = redisClient.Close()
		return nil, err
	}
	pool := sharedtransport.NewClientPool(reg)
	hub := websocketx.NewHub(log)
	relayService := relay.New(redisClient, hub, log)
	security := iamv1.NewRemote(pool)
	httpServer := server.NewHTTP(server.HTTPOptions{
		Port: cfg.Server.Port, CORS: cfg.CORS.Enabled, TokenHeader: cfg.Auth.TokenHeader,
		TimeoutEnabled: cfg.WebSocket.TimeoutEnabled, ReadTimeoutSeconds: cfg.WebSocket.ReadTimeoutSeconds,
		WriteTimeoutSeconds: cfg.WebSocket.WriteTimeoutSeconds, HeartbeatEnabled: cfg.WebSocket.HeartbeatEnabled,
		MaxReadTimeouts: cfg.WebSocket.MaxReadTimeouts,
	}, security, redisClient, relayService, hub, log)
	return &App{
		redis: redisClient, registry: reg, clientPool: pool, relay: relayService, hub: hub, httpServer: httpServer, log: log,
		grpc: sharedtransport.RegisteredGRPCOptions{
			GRPCPort: cfg.GRPC.Port, HTTPPort: cfg.Server.Port, ServiceID: cfg.Service.ID,
			ServiceName: realtimev1.ServiceName, AdvertiseHost: cfg.Service.AdvertiseHost,
			Routes: []registry.HTTPRoute{{Method: http.MethodGet, Path: server.WebSocketPath}},
		},
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	a.runCancel = cancel
	defer cancel()
	if err := a.redis.Ping(runCtx).Err(); err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	a.hub.Start()
	a.relay.Start(runCtx)
	defer func() {
		a.hub.Close()
		shutdown, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		_ = a.httpServer.Shutdown(shutdown)
	}()
	readyCtx, cancelReady := context.WithTimeout(runCtx, 10*time.Second)
	defer cancelReady()
	if err := relay.WaitReady(readyCtx, a.relay); err != nil {
		return fmt.Errorf("start redis subscription: %w", err)
	}

	httpErr := make(chan error, 1)
	go serveHTTP(a.httpServer, a.log, httpErr)
	grpcServer, err := sharedtransport.StartRegisteredGRPC(runCtx, a.registry, a.grpc, func(server *grpc.Server) {
		realtimev1.RegisterRealtimeServiceServer(server, domain.NewServer(a.relay))
	})
	if err != nil {
		return err
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = a.httpServer.Shutdown(shutdown)
		_ = grpcServer.Stop(shutdown)
	}()
	select {
	case err := <-httpErr:
		return err
	case <-runCtx.Done():
		return nil
	}
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	if a.runCancel != nil {
		a.runCancel()
	}
	var err error
	if a.httpServer != nil {
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = errors.Join(err, a.httpServer.Shutdown(shutdown))
		cancel()
	}
	if a.hub != nil {
		a.hub.Close()
	}
	if a.clientPool != nil {
		err = errors.Join(err, a.clientPool.Close())
	}
	if a.registry != nil {
		err = errors.Join(err, a.registry.Close())
	}
	if a.redis != nil {
		err = errors.Join(err, a.redis.Close())
	}
	return err
}

func serveHTTP(server *http.Server, log logging.Logger, errCh chan<- error) {
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("realtime http server failed")
		errCh <- err
	}
}
