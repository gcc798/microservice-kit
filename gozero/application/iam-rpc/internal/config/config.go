package config

import "github.com/zeromicro/go-zero/zrpc"

type PostgresConf struct {
	Dsn                    string
	MaxIdleConns           int
	MaxOpenConns           int
	ConnMaxLifetimeMinutes int
}

type RedisConf struct {
	Addr     string
	Password string
	Db       int
}

type JwtConf struct {
	Secret string
}

type Config struct {
	zrpc.RpcServerConf
	SysRpc     zrpc.RpcClientConf
	Postgres   PostgresConf
	CacheRedis RedisConf
	Jwt        JwtConf
}
