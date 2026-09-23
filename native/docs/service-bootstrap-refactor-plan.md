# 非中间件服务启动装配终态重构规划

## 1. 规划范围

本文是在《服务启动装配重构思想》基础上的终态方案，直接描述目标结构，不设计兼容层、双轨运行或临时适配器。

本次规划覆盖：

- IAM；
- SYS；
- Resource；
- Realtime；
- Gateway。

现有 Scheduler 不再作为终态服务保留。它当前唯一的 `data-cleanup` Job 会按数据所有权拆回 SYS 和 Resource，完成验证后删除 Scheduler 进程及其专属配置、部署和 RPC 调用链。

SYS 试点已经完成并通过运行闭环验证。其余服务继续逐个实施，不要求一次性改造全部服务。

本规划不引入依赖注入框架，不改变对外 HTTP、WebSocket、数据库和消息语义。仅供已删除 Scheduler 调用的内部 gRPC 方法会一并移除。

本规划也不追求全面 DDD 化。DDD 只用于帮助划分限界上下文、明确数据所有权和集中复杂业务规则；简单 CRUD 不额外引入完整 DDD 战术层次。

## 2. 终态设计原则

整体结构遵循以下关系：

```text
组合根决定谁负责连接依赖
依赖方向决定谁不能反向依赖
资源所有权决定谁负责生命周期
显式依赖决定代码是否真正可理解
配置边界决定服务是否只看到自己的能力
Worker 所有权决定后台协程是否可控
DDD 边界决定业务规则和数据所有权归属
```

### 2.1 DDD 的适度使用

DDD 在本项目中主要解决业务边界问题：

- IAM 拥有身份、认证、授权和用户权限规则；
- SYS 拥有组织、菜单、字典、配置和系统日志规则；
- Resource 拥有附件、对象存储和资源元数据规则；
- Realtime 拥有连接、投递和实时消息规则；
- 定时任务属于其操作数据的领域服务，不单独建立 Scheduler 限界上下文。

不要求所有模块都具备完整的 Aggregate、Repository、Application Service 和 Domain Event。只有业务规则复杂、事务边界明确、需要保护不变量时，才引入对应 DDD 战术模式。

简单查询和 CRUD 可以保持：

```text
Controller → Domain Service → Repository
```

复杂业务再演进为：

```text
Controller → Application Service → Aggregate / Domain Service → Repository
```

DDD 不负责：

- 初始化 Redis、数据库和注册中心；
- 启动 HTTP、gRPC 和后台 Worker；
- 读取进程配置；
- 管理服务优雅关闭。

这些职责仍归 Bootstrap、Infra 和 App Lifecycle。

### 2.2 组合根

每个服务拥有独立的组合根：

```text
native/application/<service>/internal/bootstrap/
```

组合根是该服务唯一允许同时依赖以下内容的地方：

- 服务配置；
- 基础设施实现；
- 领域实现；
- HTTP/gRPC/WebSocket 传输实现；
- 服务私有 Worker。

### 2.3 依赖方向

```text
main
  → bootstrap
      → transport
      → domain
      → infra
```

禁止以下反向依赖：

- domain → bootstrap；
- domain → main；
- controller → Container；
- router → 全量服务配置；
- worker → main；
- 业务服务 → 其他服务数据库。

### 2.4 生命周期边界

所有长期运行组件必须明确实现以下一种生命周期：

```go
Run(context.Context) error
```

或：

```go
Start(context.Context) error
Stop(context.Context) error
```

推荐长期运行 Worker 使用 `Run`，由服务级 App 统一管理错误和取消。

## 3. 配置终态：按服务能力拆分

### 3.1 已完成的边界收口

共享 `internal/config` 不再定义全量 `Config`，只提供加载器、环境变量绑定和可复用的原子配置类型及校验。每个服务的顶层配置和组合校验都位于自己的 `internal/config`，因此 Realtime 不会看到数据库字段，服务也不会通过共享配置访问其他领域能力。

### 3.2 终态配置模型

配置结构必须按服务定义，而不是按仓库所有能力定义。

建议每个服务在自己的 `internal/config` 中拥有顶层配置结构：

```text
application/iam/internal/config.Config
application/sys/internal/config.Config
application/resource/internal/config.Config
application/realtime/internal/config.Config
application/gateway/internal/config.Config
```

共享包只提供可复用的原子配置类型和加载能力，不提供全量服务配置：

```text
internal/config/
├── loader.go
├── environment.go
├── server.go
├── registry.go
├── database.go
├── redis.go
├── storage.go
└── telemetry.go
```

服务自己的 `Config` 只组合实际需要的字段。

### 3.3 各服务配置边界

#### IAM

```text
Service
Server
Database
Redis
Registry
JWT
CORS
Auth
Telemetry
Module configuration
```

#### SYS

```text
Service
Server
Database
Redis
Registry
CORS
Auth
Telemetry
Workers
```

SYS 不包含：

- Storage；
- WebSocket；
- JWT 签发配置；
- Realtime Relay 配置。

#### Resource

```text
Service
Server
Database
Redis
Storage
Registry
CORS
Auth
Telemetry
Workers
```

Resource 不包含：

- Redis Session 配置；
- WebSocket。

SYS 和 Resource 的 `Workers` 只声明真实任务所需的 `enabled`、`cron`、保留策略和锁 TTL；修改后重启服务生效，不保留 Scheduler 的数据库热更新机制。

#### Realtime

```text
Service
Server
Redis
Registry
Auth
CORS
WebSocket
Telemetry
```

Realtime 不包含 Database 配置。

#### Gateway

```text
Service
Server
Registry
CORS
Telemetry
Proxy
```

Gateway 不包含领域数据库、领域 Redis、JWT 签发和对象存储配置。

### 3.4 配置加载约束

服务配置加载必须返回服务私有类型，由服务私有配置包提供：

```go
cfg, err := serviceconfig.Load("application/sys")
```

加载器只负责：

- 环境文件选择；
- 环境变量覆盖；
- YAML 解码；
- 基础字段校验。

服务私有配置负责：

- 必填字段校验；
- 服务依赖校验；
- 端口和地址校验；
- 服务专属配置默认值。

终态不允许用一个全量 `Config` 作为所有服务的参数类型。

## 4. 后台 Worker 终态

后台协程不是异常情况，而是服务能力的一部分。区别在于它必须有明确归属。

### 4.1 Worker 模型

每个服务可以有自己的 Worker：

```text
application/sys/internal/workers/
application/resource/internal/workers/
application/iam/internal/workers/
application/realtime/internal/workers/
```

Worker 由所属服务的 Bootstrap 创建，由所属服务的 App 启动和关闭。

### 4.2 现有 Scheduler 的处理结论

当前 Scheduler 只有一个 `data-cleanup` Job，但它混合了两个领域动作：

- `SYS.CleanLogs(90)`：清理 90 天前的登录日志和操作日志；
- `Resource.CleanExpired()`：清理过期附件及其对象存储内容。

它不是通用调度能力，而是把两个领域任务通过 RPC 聚合到一个额外进程。终态按数据所有权拆分：

```text
SYS Worker      → 直接调用 SYS Domain 清理系统日志
Resource Worker → 直接调用 Resource Domain 清理过期附件
```

两个 Worker 不再经由 gRPC 调用本服务，也不共享一个 `data-cleanup` 事务或成败状态。Scheduler 进程删除。

### 4.3 业务服务私有 Worker

业务服务根据自身业务决定是否拥有 Worker：

```text
SYS:
  系统日志清理（保留 90 天）

Resource:
  过期附件清理

IAM:
  Token 清理、第三方认证同步

Realtime:
  Redis Relay、连接状态清理、设备连接维护
```

未来任务仍优先归入拥有数据和业务规则的服务。只有出现没有领域数据所有者、需要独立扩缩容的真实调度产品时，才重新评估独立调度服务；不为假设需求保留空 Scheduler。

### 4.4 多实例执行约束

SYS 和 Resource 均可水平扩容，同一任务在同一触发周期只能由一个实例执行：

- 复用 Redis，通过 `SET key instanceID NX EX lockTTL` 非阻塞声明本次执行窗口；
- SYS 和 Resource 使用不同且稳定的锁 Key；
- 未抢到执行权的实例立即跳过，不排队等待；
- 成功实例不提前删除 Key，由 TTL 覆盖任务最长耗时和实例时钟偏差，避免先完成的实例释放后其他实例重复执行；
- `lockTTL` 必须小于相邻两次计划执行的最短间隔，启动时校验；
- Worker 启动失败属于进程启动失败，单次任务业务失败只记录日志并等待下一周期；
- 最少保留一个多实例并发测试，证明一次触发只执行一次领域操作。

Resource 因分布式锁新增 Redis 这一真实运行依赖，并在自己的 Config、Compose 和 Kubernetes 配置中显式声明，不从全量 Container 隐式取得。

### 4.5 Worker 的启动关系

```text
业务服务 Bootstrap
  → 创建 Domain
  → 创建服务私有 Worker
  → 创建 HTTP/gRPC
  → App 统一 Run
```

Worker 出错时由 App 处理：

```text
Worker 返回不可恢复错误
  → 取消 App Context
  → 停止其他组件
  → 逆序释放资源
  → 进程返回错误
```

## 5. 服务终态结构

### 5.1 通用服务结构

```text
application/<service>/
├── main.go
└── internal/
    ├── bootstrap/
    │   ├── app.go
    │   ├── infra.go
    │   ├── domain.go
    │   ├── transport.go
    ├── config/
    ├── domain/
    ├── controller/
    ├── router/
    ├── transport/
    ├── workers/
    │   └── .keep
    └── migrations/
```

没有对应能力时不创建实现文件；可按约定只保留 `workers/.keep` 表达扩展位置。例如 Realtime 不创建数据库迁移目录。

### 5.2 main.go 终态

```go
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}

func run() error {
    ctx, stop := signal.NotifyContext(...)
    defer stop()

    cfg, log, telemetry, err := bootstrap.Load()
    if err != nil {
        return err
    }
    defer telemetry.Close()

    app, err := bootstrap.New(cfg, log)
    if err != nil {
        return err
    }

    return app.Run(ctx)
}
```

`main.go` 不包含任何具体 Controller、Repository、Redis Client、RPC Server 或 Worker 的创建代码。

## 6. Bootstrap 终态职责

### `bootstrap.Load`

负责：

- 读取服务私有配置；
- 创建日志；
- 创建 Telemetry；
- 创建进程上下文。

### `bootstrap.New`

负责：

- 构造基础设施；
- 构造领域对象；
- 构造服务私有 Worker；
- 构造 HTTP/gRPC/WebSocket；
- 组装 `App`。

不负责开始运行。

### `App.Run`

负责：

- 按顺序启动基础设施依赖；
- 执行迁移；
- 启动领域 Worker；
- 启动 gRPC/HTTP；
- 等待取消或组件错误；
- 触发逆序关闭。

### `App.Stop`

负责：

- 停止 HTTP/WebSocket；
- 停止 gRPC；
- 停止服务私有 Worker；
- 停止领域模块；
- 关闭 Registry、RPC Pool、Redis、Database、Storage。

## 7. SYS 试点设计

SYS 作为第一个试点是合理的。

### 7.1 SYS 终态依赖图

```text
SYS Config
  → PostgreSQL
  → Redis
  → Registry
  → IAM RPC Client
  → SYS Domain Services
  → SYS Workers
  → HTTP Router / Controllers
  → SYS gRPC Server
```

### 7.2 SYS 终态组合根

```text
application/sys/internal/
├── bootstrap/
│   ├── app.go
│   ├── infra.go
│   ├── domain.go
│   └── transport.go
└── workers/
    └── log_cleanup.go
```

### 7.3 SYS 的明确依赖

Controller 不再接收 `container.Container`：

```go
type UserControllerDeps struct {
    Service domain.UserService
    Logger  logging.Logger
}
```

Router 不再持有 Container：

```go
type RouterDeps struct {
    Permission middleware.PermissionChecker
    Auth       echo.MiddlewareFunc
    SystemAPI  sysv1.API
}
```

HTTP Server 不再接收全量 Container：

```go
type HTTPDeps struct {
    Logger logging.Logger
    Setup  func(*httpx.Router) error
}
```

### 7.4 SYS Worker

SYS 增加系统日志清理 Worker，直接调用 SYS Domain，不再通过自身 gRPC：

```text
SYS Bootstrap
  → NewLogCleanupWorker(retention=90天)
  → Redis 非阻塞分布式锁
  → App.Run 统一管理
```

## 8. 其他服务终态方向

### IAM

- 采用 IAM 私有 Config；
- 保留 IAM 专属 JWT、认证模块和第三方模块；
- Controller 和 Router 使用显式依赖；
- Token 清理等后台任务归 IAM；
- 不让 IAM Container 暴露给业务层。

### Resource

- 采用 Resource 私有 Config；
- Storage 只在 Resource 配置和 Bootstrap 中出现；
- Redis 只用于服务私有 Worker 的分布式锁；
- 附件 Controller 只依赖 Attachment Service；
- 过期附件清理归 Resource Worker，并直接调用 Attachment Service；
- 不把 Storage 放进通用 Container。

### Realtime

- 采用 Realtime 私有 Config；
- 不包含 Database 配置；
- Redis Relay、连接 Hub、WebSocket Server、gRPC Server 由 Realtime Bootstrap 组装；
- 连接维护和设备长连接能力归 Realtime Worker/Domain；
- Realtime Worker 由 Realtime App 管理。

### Gateway

- 采用 Gateway 私有 Config；
- 只负责入口、服务发现、动态代理和流量治理；
- 不创建任何领域对象；
- 不读取 IAM、SYS、Resource 的数据库或 Redis。

## 9. 终态验收标准

### 配置

- 每个服务有独立顶层 Config 类型；
- Realtime 编译期不可访问 Database 配置；
- Gateway 编译期不可访问领域数据库配置；
- 配置模板只声明当前服务实际依赖。

### 依赖

- 业务层不依赖 `container.Container`；
- Controller 构造函数只接收实际依赖；
- Router 不持有全量服务运行时对象；
- Domain 不创建基础设施；
- 服务之间只通过公开 RPC/HTTP 契约通信。
- 业务边界与数据所有权一致；
- 复杂业务规则集中在领域层；
- 简单 CRUD 不引入无实际收益的完整 DDD 层次。

### 生命周期

- 每个服务有明确的 `App.Run` 和 `App.Stop`；
- Worker 有明确创建者、启动者、停止者；
- 多实例定时 Worker 每个触发周期至多执行一次；
- 启动失败能够逆序回滚；
- 所有后台协程响应 Context；
- 服务只在关键依赖就绪后发布 Ready。

### 结构

- `main.go` 不包含具体业务装配；
- 不出现新的万能 `Application` 或跨服务容器；
- 不使用反射式依赖注入；
- 不为了目录对称创建无实际职责的包。
- 不存在 `application/scheduler`、Scheduler 构建产物或 Scheduler 部署单元；
- SYS 和 Resource 的清理任务不经过 RPC，直接调用各自 Domain Service。

## 10. 实施顺序

SYS 组合根试点已经通过启动、健康检查、注册发现、跨服务调用和优雅关闭验证。后续按以下顺序实施，不保留双轨装配：

### 阶段一：Resource 重构与任务归位

- 按 SYS 模式建立 Resource 私有 Config、Bootstrap、Infra、Domain、Transport 和 App；
- Controller、Router、Domain 移除 `container.Container`；
- 增加 Resource 过期附件清理 Worker 和 Redis 非阻塞分布式锁；
- 在 SYS 增加系统日志清理 Worker和同样的多实例锁约束；
- 两个 Worker 完成集成验证后，在同一变更中停止旧 `data-cleanup` 调度，避免重复执行。

### 阶段二：删除 Scheduler

- 删除 `application/scheduler/`；
- 删除 `internal/modules/scheduler.go` 及其测试；
- 删除 `runtimeconfig.CodeScheduler`、`SchedulerConfig` 和 SYS 初始化数据中的 `scheduler` 配置；
- 删除仅供 Scheduler 使用的 `CleanLogs`、`CleanExpired` gRPC 方法并重新生成 Proto；
- 删除 `internal/platform/scheduler/`；每个定时 Worker 直接拥有自己的单个 Cron entry，不保留动态 Job 注册、更新和查询 API；
- 从 Makefile、Docker Compose、Kubernetes、README 和运维脚本删除 Scheduler 构建与部署项；
- 最终拓扑为 `gateway + iam + sys + resource + realtime` 五个进程类型。

### 阶段三：IAM

- 建立 IAM 私有 Config 与 Bootstrap；
- 显式装配认证、授权、JWT 和第三方模块；
- 移除 Controller、Router 对 Container 的依赖；
- 当前没有真实 IAM 定时任务，不新增空 Worker 实现。

### 阶段四：Realtime

- 将现有 `internal/server` 中的组合职责拆到 Bootstrap、Infra、Domain、Transport；
- App 统一管理 Redis Relay、Hub、WebSocket、gRPC 和优雅关闭；
- 不引入数据库配置。

### 阶段五：Gateway

- 建立 Gateway 私有 Config 与 Bootstrap；
- App 管理注册中心、RPC Client Pool、路由刷新循环和 HTTP Server；
- Gateway 继续只做入口与代理，不引入领域对象或 Worker 框架。

每个服务完成后，必须独立通过：

```text
go test ./...
go vet ./...
服务启动验证
健康检查验证
服务注册与调用验证
优雅关闭验证
```

每个阶段验收通过后再进入下一阶段；失败时修正当前服务，不通过兼容层或重新引入全量 Container 绕过问题。
