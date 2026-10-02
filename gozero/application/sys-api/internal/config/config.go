// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type JwtConf struct {
	Secret string
	Expire int64
}

type AuthConf struct {
	TokenHeader string
}

type Config struct {
	rest.RestConf
	SysRpc zrpc.RpcClientConf
	Jwt    JwtConf
	Auth   AuthConf
}
