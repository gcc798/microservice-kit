package bootstrap

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/iam/internal/config"
	iamtransport "github.com/gcc798/microservice-kit/application/iam/internal/transport"
	"github.com/gcc798/microservice-kit/internal/httpserver"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
)

type transportDeps struct {
	httpServer *httpserver.Server
	routes     []registry.HTTPRoute
}

func newTransport(cfg *serviceconfig.Config, iamInfra *iamResources, domain domainDeps, log logger.Logger) (transportDeps, error) {
	server, routes, err := iamtransport.NewHTTP(iamtransport.HTTPOptions{
		Config: cfg, DB: iamInfra.DB, Redis: iamInfra.Redis, JWT: iamInfra.JWT, Logger: log, Security: domain.security,
		Auth: domain.auth, Captcha: domain.captcha, User: domain.user, Role: domain.role,
		APIPermission: domain.apiPermission, Org: domain.org, Menu: domain.menu, SMS: domain.sms,
	})
	if err != nil {
		return transportDeps{}, err
	}
	return transportDeps{httpServer: server, routes: routes}, nil
}
