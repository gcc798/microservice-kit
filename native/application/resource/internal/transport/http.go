package transport

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/resource/internal/config"
	resource "github.com/gcc798/microservice-kit/application/resource/internal/domain"
	"github.com/gcc798/microservice-kit/application/resource/internal/router"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	"github.com/gcc798/microservice-kit/internal/httpserver"
	"github.com/gcc798/microservice-kit/internal/httpx"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HTTPOptions struct {
	Config      *serviceconfig.Config
	DB          *gorm.DB
	Redis       *redis.Client
	Logger      logging.Logger
	Security    iamv1.API
	SystemAPI   sysv1.API
	Attachments resource.AttachmentService
}

func NewHTTP(opts HTTPOptions) (*httpserver.Server, []registry.HTTPRoute, error) {
	return httpserver.New(httpserver.Options{
		Port: opts.Config.Server.Port, CORS: opts.Config.CORS.Enabled,
		Logger: opts.Logger, SystemAPI: opts.SystemAPI,
	}, func(r *httpx.Router) error {
		return router.Setup(r, router.RouterContext{
			DB: opts.DB, Redis: opts.Redis, Attachments: opts.Attachments,
			Logger: opts.Logger, TokenHeader: opts.Config.Auth.TokenHeader,
		}, opts.Security, opts.SystemAPI)
	})
}
