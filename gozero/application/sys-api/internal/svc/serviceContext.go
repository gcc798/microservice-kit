// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"github.com/gcc798/microservice-kit/application/sys-api/internal/config"
	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	SysRpcClient sysservice.SysService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:       c,
		SysRpcClient: sysservice.NewSysService(zrpc.MustNewClient(c.SysRpc)),
	}
}
