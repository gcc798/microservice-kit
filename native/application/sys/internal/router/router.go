package router

import (
	sys "github.com/gcc798/microservice-kit/application/sys/internal/domain"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	middleware "github.com/gcc798/microservice-kit/internal/httpmiddleware"
	"github.com/gcc798/microservice-kit/internal/httpx"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/runtimeconfig"
	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RouterContext struct {
	DB                *gorm.DB
	Redis             *goredis.Client
	Logger            logging.Logger
	RuntimeConfig     *runtimeconfig.Store
	Modules           sys.ModuleRefresher
	TokenHeader       string
	PermissionService middleware.PermissionChecker
	AuthMiddleware    echo.MiddlewareFunc
	SystemAPI         sysv1.API
}

func Setup(r *httpx.Router, deps RouterContext, security iamv1.API, systemAPI sysv1.API) error {
	deps.PermissionService = security
	deps.AuthMiddleware = middleware.Auth(security, middleware.AuthOptions{TokenHeader: deps.TokenHeader})
	deps.SystemAPI = systemAPI
	ctx := &deps
	r.Use(middleware.PrometheusMiddleware())
	r.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	registerCommonRoutes(r, ctx, false)
	registerDictRoutes(r, ctx)
	registerConfigRoutes(r, ctx)
	registerLoginLogRoutes(r, ctx)
	registerOperLogRoutes(r, ctx)
	return nil
}
