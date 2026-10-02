// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"github.com/gcc798/microservice-kit/application/resource-api/internal/config"
	"github.com/gcc798/microservice-kit/application/resource-rpc/client/resourceservice"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config            config.Config
	ResourceRpcClient resourceservice.ResourceService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:            c,
		ResourceRpcClient: resourceservice.NewResourceService(zrpc.MustNewClient(c.ResourceRpc)),
	}
}
