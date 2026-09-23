package bootstrap

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/sys/internal/config"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
)

type infraDeps struct {
	infra    *sysResources
	registry registry.Registry
	close    func() error
}

func newInfra(cfg *serviceconfig.Config, log logger.Logger) (infraDeps, *sharedtransport.ClientPool, error) {
	resources, err := newSYSResources(cfg, log)
	if err != nil {
		return infraDeps{}, nil, err
	}
	reg, err := newSYSRegistry(cfg)
	if err != nil {
		_ = resources.Close()
		return infraDeps{}, nil, err
	}
	return infraDeps{infra: resources, registry: reg, close: resources.Close}, sharedtransport.NewClientPool(reg), nil
}
