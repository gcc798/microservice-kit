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

type RedisConf struct {
	Addr     string
	Password string
	Db       int
}

type ExpiredAttachmentCleanupConf struct {
	Enabled        bool
	Cron           string
	LockTTLMinutes int
}

type WorkersConf struct {
	ExpiredAttachmentCleanup ExpiredAttachmentCleanupConf
}

type Config struct {
	zrpc.RpcServerConf
	Postgres PostgresConf
	Storage  StorageConf
	Redis    RedisConf
	Workers  WorkersConf
}
