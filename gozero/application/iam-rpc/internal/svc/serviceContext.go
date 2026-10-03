package svc

import (
	"errors"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/config"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/migrations"
	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
	"github.com/gcc798/microservice-kit/internal/database"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	Redis        *redis.Client
	SysRpcClient sysservice.SysService
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	db, err := database.Open(database.Config{
		DSN:                    c.Postgres.Dsn,
		MaxIdleConns:           c.Postgres.MaxIdleConns,
		MaxOpenConns:           c.Postgres.MaxOpenConns,
		ConnMaxLifetimeMinutes: c.Postgres.ConnMaxLifetimeMinutes,
	})
	if err != nil {
		return nil, err
	}
	pool, err := db.DB()
	if err != nil {
		_ = database.Close(db)
		return nil, err
	}
	if err := migrations.Up(pool); err != nil {
		_ = database.Close(db)
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.CacheRedis.Addr,
		Password: c.CacheRedis.Password,
		DB:       c.CacheRedis.Db,
	})
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		_ = rdb.Close()
		_ = database.Close(db)
		return nil, err
	}

	return &ServiceContext{
		Config:       c,
		DB:           db,
		Redis:        rdb,
		SysRpcClient: sysservice.NewSysService(zrpc.MustNewClient(c.SysRpc)),
	}, nil
}

func (s *ServiceContext) Close() error {
	return errors.Join(s.Redis.Close(), database.Close(s.DB))
}
