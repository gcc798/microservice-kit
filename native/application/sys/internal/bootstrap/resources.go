package bootstrap

import (
	"context"
	"errors"
	"fmt"

	serviceconfig "github.com/gcc798/microservice-kit/application/sys/internal/config"
	"github.com/gcc798/microservice-kit/internal/database"
	"github.com/gcc798/microservice-kit/internal/logger"
	redisplatform "github.com/gcc798/microservice-kit/internal/platform/redis"
	"github.com/gcc798/microservice-kit/internal/platform/redislock"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/gcc798/microservice-kit/internal/runtimeconfig"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type sysResources struct {
	DB            *gorm.DB
	Redis         *redis.Client
	RuntimeConfig *runtimeconfig.Store
}

func newSYSResources(cfg *serviceconfig.Config, log logger.Logger) (*sysResources, error) {
	db, err := database.Open(cfg.Database, log)
	if err != nil {
		return nil, err
	}
	redisClient := redisplatform.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		_ = redisClient.Close()
		_ = database.Close(db)
		return nil, fmt.Errorf("instrument redis tracing: %w", err)
	}
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		_ = redisClient.Close()
		_ = database.Close(db)
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return &sysResources{
		DB: db, Redis: redisClient,
		RuntimeConfig: runtimeconfig.NewStore(redisClient, runtimeconfig.NewGormSource(db), redislock.New(redisClient)),
	}, nil
}

func (s *sysResources) Close() error {
	if s == nil {
		return nil
	}
	return errors.Join(s.Redis.Close(), database.Close(s.DB))
}

func newSYSRegistry(cfg *serviceconfig.Config) (registry.Registry, error) {
	return registry.New(registry.Options{
		Driver: cfg.Registry.Driver, Address: cfg.Registry.Address, Prefix: cfg.Registry.Prefix,
		Namespace: cfg.Registry.Namespace, Group: cfg.Registry.Group, Username: cfg.Registry.Username, Password: cfg.Registry.Password,
	})
}
