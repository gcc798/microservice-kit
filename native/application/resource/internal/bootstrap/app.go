package bootstrap

import (
	"context"
	"errors"
	"time"

	serviceconfig "github.com/gcc798/microservice-kit/application/resource/internal/config"
	"github.com/gcc798/microservice-kit/application/resource/internal/migrations"
	resourcev1 "github.com/gcc798/microservice-kit/internal/api/resource/v1"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
	"google.golang.org/grpc"
)

type App struct {
	infra      infraDeps
	domain     domainDeps
	transport  transportDeps
	workers    workerDeps
	clientPool *sharedtransport.ClientPool
	grpc       sharedtransport.RegisteredGRPCOptions
}

func New(cfg *serviceconfig.Config, log logging.Logger) (*App, error) {
	infraDeps, pool, err := newInfra(cfg, log)
	if err != nil {
		return nil, err
	}
	domainDeps := newDomain(infraDeps.infra, pool, log)
	workerDeps, err := newWorkers(cfg, infraDeps.infra, domainDeps, log)
	if err != nil {
		return nil, errors.Join(err, pool.Close(), infraDeps.close())
	}
	transportDeps, err := newTransport(cfg, infraDeps.infra, domainDeps, log)
	if err != nil {
		return nil, errors.Join(err, pool.Close(), infraDeps.close())
	}
	return &App{
		infra: infraDeps, domain: domainDeps, transport: transportDeps, workers: workerDeps, clientPool: pool,
		grpc: sharedtransport.RegisteredGRPCOptions{
			GRPCPort: cfg.GRPC.Port, HTTPPort: cfg.Server.Port, ServiceID: cfg.Service.ID,
			ServiceName: resourcev1.ServiceName, AdvertiseHost: cfg.Service.AdvertiseHost, Routes: transportDeps.routes,
		},
	}, nil
}

func (a *App) Run(ctx context.Context) (err error) {
	sqlDB, err := a.infra.infra.DB.DB()
	if err != nil {
		return err
	}
	if err := migrations.Up(sqlDB); err != nil {
		return err
	}
	a.workers.start(ctx)
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = errors.Join(err, a.workers.stop(shutdown))
	}()
	grpcServer, err := sharedtransport.StartRegisteredGRPC(ctx, a.infra.registry, a.grpc, func(server *grpc.Server) {
		resourcev1.RegisterResourceServiceServer(server, a.domain.grpcServer)
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
	if a.infra.close != nil {
		closeErr = errors.Join(closeErr, a.infra.close())
	}
	return closeErr
}
