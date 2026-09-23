package router

import (
	"github.com/gcc798/microservice-kit/internal/constants"
	"github.com/gcc798/microservice-kit/internal/httpmiddleware"
	"github.com/gcc798/microservice-kit/internal/httpx"
)

// registerUserRoutes 注册用户管理路由
func registerUserRoutes(r *httpx.Router, ctx *RouterContext) {
	// 用户管理路由组（需要认证和权限）
	users := r.Group("/api/v1/user")
	users.Use(ctx.AuthMiddleware)
	{
		// 用户创建 - 需要 user.create 权限
		users.POST("", middleware.Permission(ctx.PermissionService, constants.ResourceUserCreate, "write"), ctx.User.Create)
		users.POST("/import", middleware.Permission(ctx.PermissionService, constants.ResourceUserCreate, "write"), ctx.User.BatchImport)

		// 用户查询 - 需要 user.read 权限
		users.POST("/page", middleware.Permission(ctx.PermissionService, constants.ResourceUserRead, "read"), ctx.User.PageUser)

		// 批量删除 - 需要 user.delete 权限
		users.DELETE("/batch", middleware.Permission(ctx.PermissionService, constants.ResourceUserDelete, "write"), ctx.User.BatchDelete)

		// 用户修改密码 - 不需要特殊权限（用户修改自己的密码）
		users.POST("/password/change", ctx.User.ChangePassword)

		// 用户更新 - 需要 user.update 权限（带参数的路由放在后面）
		users.PUT("/:id", middleware.Permission(ctx.PermissionService, constants.ResourceUserUpdate, "write"), ctx.User.Update)
		users.PUT("/:id/password", middleware.Permission(ctx.PermissionService, constants.ResourceUserUpdate, "write"), ctx.User.ResetPassword)

		// 用户查询 - 需要 user.read 权限（带参数的路由放在最后）
		users.GET("/:id", middleware.Permission(ctx.PermissionService, constants.ResourceUserRead, "read"), ctx.User.GetById)

		// 用户删除 - 需要 user.delete 权限（带参数的路由放在最后）
		users.DELETE("/:id", middleware.Permission(ctx.PermissionService, constants.ResourceUserDelete, "write"), ctx.User.Delete)
	}

	systemUsers := r.Group("/system/user", ctx.AuthMiddleware)
	{
		systemUsers.POST("/xcxGetInfo", ctx.User.XcxGetInfo)
	}
}
