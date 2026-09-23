package bootstrap

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/iam/internal/config"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
)

type infraDeps struct {
	iam      *iamResources
	registry registry.Registry
}

func newInfra(cfg *serviceconfig.Config, log logger.Logger) (infraDeps, *sharedtransport.ClientPool, error) {
	iamInfra, err := newIAMResources(cfg, log)
	if err != nil {
		return infraDeps{}, nil, err
	}
	reg, err := newIAMRegistry(cfg)
	if err != nil {
		_ = iamInfra.Close()
		return infraDeps{}, nil, err
	}
	return infraDeps{iam: iamInfra, registry: reg}, newIAMClientPool(reg), nil
}
