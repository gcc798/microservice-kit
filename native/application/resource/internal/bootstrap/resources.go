package bootstrap

import (
	"context"
	"errors"
	"fmt"

	serviceconfig "github.com/gcc798/microservice-kit/application/resource/internal/config"
	"github.com/gcc798/microservice-kit/internal/database"
	"github.com/gcc798/microservice-kit/internal/logger"
	redisplatform "github.com/gcc798/microservice-kit/internal/platform/redis"
	"github.com/gcc798/microservice-kit/internal/platform/storage"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type resourceResources struct {
	DB      *gorm.DB
	Redis   *redis.Client
	Storage storage.Storage
}

func newResourceResources(cfg *serviceconfig.Config, log logger.Logger) (*resourceResources, error) {
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
	store, err := storage.New(cfg.Storage)
	if err != nil {
		_ = redisClient.Close()
		_ = database.Close(db)
		return nil, fmt.Errorf("initialize storage: %w", err)
	}
	log.Info("storage initialized")
	return &resourceResources{DB: db, Redis: redisClient, Storage: store}, nil
}

func (r *resourceResources) Close() error {
	if r == nil {
		return nil
	}
	return errors.Join(r.Redis.Close(), database.Close(r.DB))
}

func newResourceRegistry(cfg *serviceconfig.Config) (registry.Registry, error) {
	return registry.New(registry.Options{
		Driver: cfg.Registry.Driver, Address: cfg.Registry.Address, Prefix: cfg.Registry.Prefix,
		Namespace: cfg.Registry.Namespace, Group: cfg.Registry.Group, Username: cfg.Registry.Username, Password: cfg.Registry.Password,
	})
}
