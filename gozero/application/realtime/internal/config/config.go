package config

import (
	registry "github.com/gcc798/microservice-kit/internal/registry"
	"github.com/zeromicro/go-zero/zrpc"
)

type HTTP struct {
	Host string
	Port int
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

type Auth struct {
	TokenHeader string
}

type Config struct {
	zrpc.RpcServerConf
	HTTP         HTTP
	HTTPRegistry registry.HTTPConfig
	RelayRedis   Redis
	Security     Auth
	IamRpc       zrpc.RpcClientConf
}
