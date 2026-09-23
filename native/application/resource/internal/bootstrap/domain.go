package bootstrap

import (
	"time"

	resource "github.com/gcc798/microservice-kit/application/resource/internal/domain"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
)

type domainDeps struct {
	security    iamv1.API
	systemAPI   sysv1.API
	attachments resource.AttachmentService
	grpcServer  *resource.GRPCServer
}

func newDomain(resources *resourceResources, pool *sharedtransport.ClientPool, log logging.Logger) domainDeps {
	attachments := resource.NewAttachmentService(resources.DB, resources.Storage, log)
	return domainDeps{
		security: iamv1.NewCached(iamv1.NewRemote(pool), 5*time.Second), systemAPI: sysv1.NewRemote(pool),
		attachments: attachments, grpcServer: resource.NewGRPCServer(),
	}
}
