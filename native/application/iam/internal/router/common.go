package router

import "github.com/gcc798/microservice-kit/internal/httpx"

// 注册公共路由（健康检查等）。
func registerCommonRoutes(r *httpx.Router, ctx *RouterContext) {
	// 健康检查接口（公开接口，无需认证）
	r.GET("/health", ctx.Health.Health)          // 基础健康检查
	r.GET("/health/ready", ctx.Health.Ready)     // 就绪探针
	r.GET("/health/live", ctx.Health.Live)       // 存活探针
	r.GET("/health/startup", ctx.Health.Startup) // 启动探针

}
