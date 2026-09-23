package bootstrap

import (
	"time"

	sys "github.com/gcc798/microservice-kit/application/sys/internal/domain"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	"github.com/gcc798/microservice-kit/internal/logger"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
)

type domainDeps struct {
	security   iamv1.API
	systemAPI  *sys.API
	logCleanup sys.LogCleanupService
	grpcServer *sys.GRPCServer
}

func newDomain(resources *sysResources, pool *sharedtransport.ClientPool, log logger.Logger) domainDeps {
	security := iamv1.NewCached(iamv1.NewRemote(pool), 5*time.Second)
	systemAPI := sys.NewAPI(resources.DB, log)
	loginLogs := sys.NewLoginLogService(resources.DB, log)
	operationLogs := sys.NewOperLogService(resources.DB, log)
	logCleanup := sys.NewLogCleanupService(loginLogs, operationLogs)
	return domainDeps{security: security, systemAPI: systemAPI, logCleanup: logCleanup, grpcServer: sys.NewGRPCServer(systemAPI)}
}
