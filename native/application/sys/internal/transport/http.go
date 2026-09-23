package transport

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/sys/internal/config"
	"github.com/gcc798/microservice-kit/application/sys/internal/router"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	"github.com/gcc798/microservice-kit/internal/httpserver"
	"github.com/gcc798/microservice-kit/internal/httpx"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/gcc798/microservice-kit/internal/runtimeconfig"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HTTPOptions struct {
	Config        *serviceconfig.Config
	DB            *gorm.DB
	Redis         *redis.Client
	RuntimeConfig *runtimeconfig.Store
	Logger        logger.Logger
	Security      iamv1.API
	SystemAPI     sysv1.API
}

func NewHTTP(opts HTTPOptions) (*httpserver.Server, []registry.HTTPRoute, error) {
	return httpserver.New(httpserver.Options{
		Port: opts.Config.Server.Port, CORS: opts.Config.CORS.Enabled,
		Logger: opts.Logger, SystemAPI: opts.SystemAPI,
	}, func(r *httpx.Router) error {
		return router.Setup(r, router.RouterContext{
			DB: opts.DB, Redis: opts.Redis, Logger: opts.Logger,
			RuntimeConfig: opts.RuntimeConfig, TokenHeader: opts.Config.Auth.TokenHeader,
		}, opts.Security, opts.SystemAPI)
	})
}
