// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
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
	TokenHeader string
}

type Config struct {
	rest.RestConf
	IamRpc zrpc.RpcClientConf
	Redis  RedisConf
	Jwt    JwtConf
	Auth   AuthConf
}
