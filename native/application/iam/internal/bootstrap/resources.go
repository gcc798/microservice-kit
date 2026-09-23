package bootstrap

import (
	"context"
	"errors"
	"fmt"

	serviceconfig "github.com/gcc798/microservice-kit/application/iam/internal/config"
	sharedconfig "github.com/gcc798/microservice-kit/internal/config"
	"github.com/gcc798/microservice-kit/internal/database"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/modules"
	"github.com/gcc798/microservice-kit/internal/platform/jwt"
	redisplatform "github.com/gcc798/microservice-kit/internal/platform/redis"
	"github.com/gcc798/microservice-kit/internal/platform/redislock"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/gcc798/microservice-kit/internal/runtimeconfig"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type iamResources struct {
	DB            *gorm.DB
	Redis         *redis.Client
	JWT           *jwt.Jwt
	RuntimeConfig *runtimeconfig.Store
	Modules       *modules.Manager
}

func newIAMResources(cfg *serviceconfig.Config, log logger.Logger) (*iamResources, error) {
	db, err := database.Open(cfg.Database, log)
	if err != nil {
		return nil, err
	}
	redisClient, err := newIAMRedis(cfg.Redis)
	if err != nil {
		_ = database.Close(db)
		return nil, err
	}
	store := runtimeconfig.NewStore(redisClient, runtimeconfig.NewGormSource(db), redislock.New(redisClient))
	return &iamResources{
		DB: db, Redis: redisClient, JWT: jwt.New(cfg.JWT.Secret, cfg.JWT.Expire), RuntimeConfig: store,
		Modules: modules.NewManager(modules.Dependencies{DB: db, Redis: redisClient, Logger: log, RuntimeConfig: store}),
	}, nil
}

func newIAMRedis(cfg sharedconfig.Redis) (*redis.Client, error) {
	client := redisplatform.NewRedis(cfg.Addr, cfg.Password, cfg.DB)
	if err := redisotel.InstrumentTracing(client); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("instrument redis tracing: %w", err)
	}
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return client, nil
}

func (i *iamResources) RegisterModules(ctx context.Context, candidates ...modules.Module) error {
	return i.Modules.Register(ctx, candidates...)
}

func (i *iamResources) StartModules(ctx context.Context) error { return i.Modules.Start(ctx) }
func (i *iamResources) StopModules(ctx context.Context) error  { return i.Modules.Stop(ctx) }

func (i *iamResources) Close() error {
	if i == nil {
		return nil
	}
	return errors.Join(i.Redis.Close(), database.Close(i.DB))
}

func newIAMRegistry(cfg *serviceconfig.Config) (registry.Registry, error) {
	return registry.New(registry.Options{
		Driver: cfg.Registry.Driver, Address: cfg.Registry.Address, Prefix: cfg.Registry.Prefix,
		Namespace: cfg.Registry.Namespace, Group: cfg.Registry.Group, Username: cfg.Registry.Username, Password: cfg.Registry.Password,
	})
}

func newIAMClientPool(reg registry.Registry) *sharedtransport.ClientPool {
	return sharedtransport.NewClientPool(reg)
}
