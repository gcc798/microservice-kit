package router

import (
	"github.com/gcc798/microservice-kit/application/iam/internal/controller"
	"github.com/gcc798/microservice-kit/internal/health"
	middleware "github.com/gcc798/microservice-kit/internal/httpmiddleware"
	"github.com/gcc798/microservice-kit/internal/httpx"
	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterContext struct {
	Health            health.Handler
	PermissionService middleware.PermissionChecker
	AuthMiddleware    echo.MiddlewareFunc
	Auth              controller.AuthController
	Captcha           *controller.CaptchaController
	User              controller.UserController
	Role              controller.RoleController
	APIPermission     controller.ApiPermissionController
	Org               controller.OrgController
	Menu              controller.MenuController
}

func Setup(r *httpx.Router, ctx RouterContext) {
	r.Use(middleware.PrometheusMiddleware())
	r.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	registerCommonRoutes(r, &ctx)
	registerAuthRoutes(r, &ctx)
	registerCaptchaRoutes(r, &ctx)
	registerUserRoutes(r, &ctx)
	registerRoleRoutes(r, &ctx)
	registerApiPermissionRoutes(r, &ctx)
	registerOrgRoutes(r, &ctx)
	registerMenuRoutes(r, &ctx)
}
