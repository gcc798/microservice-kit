package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gcc798/microservice-kit/application/realtime/internal/domain"
	"github.com/gcc798/microservice-kit/application/realtime/internal/relay"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	realtimev1 "github.com/gcc798/microservice-kit/internal/api/realtime/v1"
	"github.com/gcc798/microservice-kit/internal/config"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/platform/redis"
	websocketx "github.com/gcc798/microservice-kit/internal/platform/websocket"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/gcc798/microservice-kit/internal/transport"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, cfg *config.Config, log logging.Logger) error {
	redisClient := redis.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		return err
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	defer redisClient.Close()

	reg, err := registry.New(registry.Options{
		Driver: cfg.Registry.Driver, Address: cfg.Registry.Address, Prefix: cfg.Registry.Prefix,
		Namespace: cfg.Registry.Namespace, Group: cfg.Registry.Group, Username: cfg.Registry.Username, Password: cfg.Registry.Password,
	})
	if err != nil {
		return err
	}
	defer reg.Close()
	pool := transport.NewClientPool(reg)
	defer pool.Close()

	security := iamv1.NewRemote(pool)
	hub := websocketx.NewHub(log)
	hub.Start()
	defer hub.Close()
	relayService := relay.New(redisClient, hub, log)
	relayService.Start(ctx)
	readyCtx, cancelReady := context.WithTimeout(ctx, 10*time.Second)
	if err := relay.WaitReady(readyCtx, relayService); err != nil {
		cancelReady()
		return fmt.Errorf("start redis subscription: %w", err)
	}
	cancelReady()

	httpServer := NewHTTP(cfg, security, redisClient, relayService, hub, log)
	httpErr := make(chan error, 1)
	go serveHTTP(httpServer, log, httpErr)
	grpcServer, err := transport.StartRegisteredGRPC(ctx, reg, cfg, realtimev1.ServiceName, []registry.HTTPRoute{{Method: http.MethodGet, Path: WebSocketPath}}, func(server *grpc.Server) {
		realtimev1.RegisterRealtimeServiceServer(server, domain.NewServer(relayService))
	})
	if err != nil {
		return err
	}

	select {
	case err := <-httpErr:
		_ = grpcServer.Stop(context.Background())
		return err
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdown)
	_ = grpcServer.Stop(shutdown)
	return nil
}

func serveHTTP(server *http.Server, log logging.Logger, errCh chan<- error) {
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("realtime http server failed", zap.Error(err))
		errCh <- err
	}
}
