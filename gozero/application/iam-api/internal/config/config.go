// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	registry "github.com/gcc798/microservice-kit/internal/registry"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type RedisConf struct {
	Addr     string
	Password string
	Db       int
}

type JwtConf struct {
	Secret string
	Expire int64
}

type AuthConf struct {
	TokenHeader     string
	AllowConcurrent bool
}

type Config struct {
	rest.RestConf
	HTTPRegistry registry.HTTPConfig
	IamRpc       zrpc.RpcClientConf
	Redis        RedisConf
	Jwt          JwtConf
	Auth         AuthConf
}
