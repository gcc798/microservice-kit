package svc

import (
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/config"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/migrations"
	"github.com/gcc798/microservice-kit/internal/database"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
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
	return &ServiceContext{Config: c, DB: db}, nil
}

func (s *ServiceContext) Close() error {
	return database.Close(s.DB)
}
