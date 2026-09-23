package bootstrap

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/resource/internal/config"
	resourcetransport "github.com/gcc798/microservice-kit/application/resource/internal/transport"
	"github.com/gcc798/microservice-kit/internal/httpserver"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
)

type transportDeps struct {
	httpServer *httpserver.Server
	routes     []registry.HTTPRoute
}

func newTransport(cfg *serviceconfig.Config, resources *resourceResources, domain domainDeps, log logging.Logger) (transportDeps, error) {
	server, routes, err := resourcetransport.NewHTTP(resourcetransport.HTTPOptions{
		Config: cfg, DB: resources.DB, Redis: resources.Redis, Logger: log, Security: domain.security,
		SystemAPI: domain.systemAPI, Attachments: domain.attachments,
	})
	if err != nil {
		return transportDeps{}, err
	}
	return transportDeps{httpServer: server, routes: routes}, nil
}
