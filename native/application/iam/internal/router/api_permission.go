package router

import (
	"github.com/gcc798/microservice-kit/internal/constants"
	"github.com/gcc798/microservice-kit/internal/httpmiddleware"
	"github.com/gcc798/microservice-kit/internal/httpx"
)

// registerApiPermissionRoutes 注册 API 权限管理路由。
func registerApiPermissionRoutes(r *httpx.Router, ctx *RouterContext) {
	apiPermissions := r.Group("/api/v1/api-permission")
	apiPermissions.Use(ctx.AuthMiddleware)
	{
		apiPermissions.GET("/tree", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionRead, "read"), ctx.APIPermission.Tree)
		apiPermissions.GET("", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionRead, "read"), ctx.APIPermission.List)
		apiPermissions.POST("", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionCreate, "write"), ctx.APIPermission.Create)
		apiPermissions.PUT("/:id", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionUpdate, "write"), ctx.APIPermission.Update)
		apiPermissions.DELETE("/:id", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionDelete, "write"), ctx.APIPermission.Delete)
	}

	roles := r.Group("/api/v1/role")
	roles.Use(ctx.AuthMiddleware)
	{
		roles.GET("/:roleId/api-permissions", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionAssign, "write"), ctx.APIPermission.GetRolePermissions)
		roles.POST("/:roleId/api-permissions", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionAssign, "write"), ctx.APIPermission.AssignRolePermissions)
	}

	users := r.Group("/api/v1/user")
	users.Use(ctx.AuthMiddleware)
	{
		users.GET("/:id/api-permissions", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionAssign, "write"), ctx.APIPermission.GetUserPermissions)
		users.POST("/:id/api-permissions", middleware.Permission(ctx.PermissionService, constants.ResourceApiPermissionAssign, "write"), ctx.APIPermission.AssignUserPermissions)
	}
}
