package transport

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/iam/internal/config"
	"github.com/gcc798/microservice-kit/application/iam/internal/controller"
	iam "github.com/gcc798/microservice-kit/application/iam/internal/domain"
	"github.com/gcc798/microservice-kit/application/iam/internal/router"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	"github.com/gcc798/microservice-kit/internal/health"
	"github.com/gcc798/microservice-kit/internal/httpmiddleware"
	"github.com/gcc798/microservice-kit/internal/httpserver"
	"github.com/gcc798/microservice-kit/internal/httpx"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/modules"
	"github.com/gcc798/microservice-kit/internal/platform/jwt"
	"github.com/gcc798/microservice-kit/internal/registry"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HTTPOptions struct {
	Config        *serviceconfig.Config
	DB            *gorm.DB
	Redis         *redis.Client
	JWT           *jwt.Jwt
	Logger        logger.Logger
	Security      iamv1.API
	Auth          iam.AuthService
	Captcha       iam.CaptchaService
	User          iam.UserService
	Role          iam.RoleService
	APIPermission iam.ApiPermissionService
	Org           iam.OrgService
	Menu          *iam.MenuService
	SMS           modules.SMS
}

func NewHTTP(opts HTTPOptions) (*httpserver.Server, []registry.HTTPRoute, error) {
	return httpserver.New(httpserver.Options{
		Port: opts.Config.Server.Port, CORS: opts.Config.CORS.Enabled,
		Logger: opts.Logger, SystemAPI: nil,
	}, func(r *httpx.Router) error {
		security := opts.Security
		router.Setup(r, router.RouterContext{
			Health: health.NewHandler(opts.DB, opts.Redis), PermissionService: security,
			AuthMiddleware: middleware.Auth(security, middleware.AuthOptions{TokenHeader: opts.Config.Auth.TokenHeader}),
			Auth:           controller.NewAuthController(opts.Auth, opts.Config.Auth.TokenHeader, opts.Logger),
			Captcha:        controller.NewCaptchaController(opts.Captcha, opts.SMS),
			User:           controller.NewUserController(opts.DB, opts.JWT, opts.Config.Auth.TokenHeader, opts.Logger, opts.User),
			Role:           controller.NewRoleController(opts.Role, opts.Logger),
			APIPermission:  controller.NewApiPermissionController(opts.APIPermission),
			Org:            controller.NewOrgController(opts.JWT, opts.Config.Auth.TokenHeader, opts.Logger, opts.Org),
			Menu:           controller.NewMenuController(opts.Menu),
		})
		return nil
	})
}
