package router

import "github.com/gcc798/microservice-kit/internal/httpx"

// 注册认证相关路由。
func registerAuthRoutes(r *httpx.Router, ctx *RouterContext) {
	// 公开路由（无需认证）
	r.POST("/login", ctx.Auth.Login)               // 统一登录接口
	r.POST("/logout", ctx.Auth.Logout)             // 登出
	r.POST("/auth/refresh", ctx.Auth.RefreshToken) // 刷新Token
}
