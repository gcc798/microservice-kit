// Package modules 定义由应用入口组合的运行时模块契约。
package modules

import (
	"context"

	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/runtimeconfig"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Dependencies 描述模块可以使用的基础设施依赖。
type Dependencies struct {
	DB            *gorm.DB             // 数据库连接。
	Redis         *redis.Client        // Redis 客户端。
	Logger        logger.Logger        // 日志记录器。
	RuntimeConfig *runtimeconfig.Store // 运行时配置存储。
}

// Module 定义可由应用组合的运行时模块。
type Module interface {
	// Name 返回模块的唯一名称。
	Name() string
	// Init 初始化模块并注入共享依赖。
	Init(ctx context.Context, deps Dependencies) error
	// Start 启动模块运行逻辑。
	Start(ctx context.Context) error
	// Stop 停止模块并释放运行资源。
	Stop(ctx context.Context) error
	// Refresh 按请求刷新模块的本地配置或缓存。
	Refresh(ctx context.Context, req ModuleRefreshRequest) error
}

// ModuleRefreshRequest 描述一次显式的本地刷新请求。
type ModuleRefreshRequest struct {
	Codes  []string // 需要刷新的配置编码；为空表示由模块自行决定范围。
	Reason string   // 触发刷新的原因。
}
