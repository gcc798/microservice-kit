package svc

import (
	"errors"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/config"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/migrations"
	"github.com/gcc798/microservice-kit/internal/database"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Client
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
	rdb := redis.NewClient(&redis.Options{Addr: c.Redis.Addr, Password: c.Redis.Password, DB: c.Redis.Db})
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		_ = rdb.Close()
		_ = database.Close(db)
		return nil, err
	}
	return &ServiceContext{Config: c, DB: db, Redis: rdb}, nil
}

func (s *ServiceContext) Close() error {
	return errors.Join(s.Redis.Close(), database.Close(s.DB))
}
