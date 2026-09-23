package bootstrap

import (
	"context"
	"errors"
	"time"

	serviceconfig "github.com/gcc798/microservice-kit/application/iam/internal/config"
	"github.com/gcc798/microservice-kit/application/iam/internal/migrations"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/transport"
	"google.golang.org/grpc"
)

// App 负责 IAM 进程生命周期。
type App struct {
	infra      infraDeps
	domain     domainDeps
	transport  transportDeps
	clientPool *transport.ClientPool
	grpc       transport.RegisteredGRPCOptions
}

func New(cfg *serviceconfig.Config, log logger.Logger) (*App, error) {
	infraDeps, pool, err := newInfra(cfg, log)
	if err != nil {
		return nil, err
	}
	domainDeps := newDomain(cfg, infraDeps.iam, pool, log)
	if err := infraDeps.iam.RegisterModules(context.Background(), domainDeps.modules...); err != nil {
		_ = pool.Close()
		_ = infraDeps.registry.Close()
		_ = infraDeps.iam.Close()
		return nil, err
	}
	transportDeps, err := newTransport(cfg, infraDeps.iam, domainDeps, log)
	if err != nil {
		_ = pool.Close()
		_ = infraDeps.registry.Close()
		_ = infraDeps.iam.Close()
		return nil, err
	}
	return &App{
		infra: infraDeps, domain: domainDeps, transport: transportDeps, clientPool: pool,
		grpc: transport.RegisteredGRPCOptions{
			GRPCPort: cfg.GRPC.Port, HTTPPort: cfg.Server.Port, ServiceID: cfg.Service.ID,
			ServiceName: iamv1.ServiceName, AdvertiseHost: cfg.Service.AdvertiseHost, Routes: transportDeps.routes,
		},
	}, nil
}

func (a *App) Run(ctx context.Context) (err error) {
	db, err := a.infra.iam.DB.DB()
	if err != nil {
		return err
	}
	if err := migrations.Up(db); err != nil {
		return err
	}
	if err := a.infra.iam.StartModules(ctx); err != nil {
		return err
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = errors.Join(err, a.infra.iam.StopModules(shutdown))
	}()
	grpcServer, err := transport.StartRegisteredGRPC(ctx, a.infra.registry, a.grpc, func(server *grpc.Server) {
		iamv1.RegisterIAMServiceServer(server, a.domain.grpcServer)
	})
	if err != nil {
		return err
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = errors.Join(err, grpcServer.Stop(shutdown))
	}()
	return a.transport.httpServer.Run(ctx)
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	var closeErr error
	if a.clientPool != nil {
		closeErr = errors.Join(closeErr, a.clientPool.Close())
	}
	if a.infra.registry != nil {
		closeErr = errors.Join(closeErr, a.infra.registry.Close())
	}
	return errors.Join(closeErr, a.infra.iam.Close())
}
