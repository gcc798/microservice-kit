package config

import "github.com/zeromicro/go-zero/zrpc"

type PostgresConf struct {
	Dsn                    string
	MaxIdleConns           int
	MaxOpenConns           int
	ConnMaxLifetimeMinutes int
}

type StorageConf struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

type Config struct {
	zrpc.RpcServerConf
	Postgres PostgresConf
	Storage  StorageConf
}
