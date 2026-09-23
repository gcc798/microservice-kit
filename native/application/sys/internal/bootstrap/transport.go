package bootstrap

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/sys/internal/config"
	systransport "github.com/gcc798/microservice-kit/application/sys/internal/transport"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	"github.com/gcc798/microservice-kit/internal/httpserver"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
)

type transportDeps struct {
	httpServer *httpserver.Server
	routes     []registry.HTTPRoute
}

func newTransport(cfg *serviceconfig.Config, resources *sysResources, security iamv1.API, systemAPI sysv1.API, log logger.Logger) (transportDeps, error) {
	server, routes, err := systransport.NewHTTP(systransport.HTTPOptions{
		Config: cfg, DB: resources.DB, Redis: resources.Redis, RuntimeConfig: resources.RuntimeConfig, Logger: log, Security: security, SystemAPI: systemAPI,
	})
	if err != nil {
		return transportDeps{}, err
	}
	return transportDeps{httpServer: server, routes: routes}, nil
}
