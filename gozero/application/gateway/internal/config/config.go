package config

import (
	registry "github.com/gcc798/microservice-kit/internal/registry"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

type Auth struct {
	TokenHeader string
}

type Config struct {
	service.ServiceConf
	Host            string
	Port            int
	Auth            Auth
	HTTPRegistry    registry.HTTPConfig
	ServiceRegistry registry.HTTPConfig
	IamRpc          zrpc.RpcClientConf
	SysRpc          zrpc.RpcClientConf
}
