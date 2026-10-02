package config

import "github.com/zeromicro/go-zero/zrpc"

type Auth struct {
	TokenHeader string
}

type Backends struct {
	IAM      string
	SYS      string
	Resource string
	Realtime string
}

type Config struct {
	Name     string
	Host     string
	Port     int
	Auth     Auth
	Backends Backends
	IamRpc   zrpc.RpcClientConf
}
