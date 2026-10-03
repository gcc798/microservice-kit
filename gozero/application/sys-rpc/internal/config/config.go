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

type LogCleanupConf struct {
	Enabled        bool
	Cron           string
	RetentionDays  int
	LockTTLMinutes int
}

type WorkersConf struct {
	LogCleanup LogCleanupConf
}

type Config struct {
	zrpc.RpcServerConf
	Postgres PostgresConf
	Redis    RedisConf
	Workers  WorkersConf
}
