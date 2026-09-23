# 服务启动装配重构思想

## 1. 文档目的

本文只定义重构思想、目标边界和判断标准，不描述具体文件改造顺序，也不直接指导代码迁移。

目标是解决非中间件服务启动代码中的两个问题：

1. `main.go` 同时承担进程启动、依赖创建、业务装配、服务启动和资源关闭；
2. 仅把面条式代码从 `main.go` 移到 `internal/server` 或 `internal/app`，代码仍然缺乏边界。

本次重构借鉴开源项目的不是某个依赖注入框架，而是“组合根”和“生命周期边界”：

- WuKongIM 的 `New → Start → Stop` 应用生命周期；
- Kratos 的 Server 构造、服务注册和统一 App 生命周期；
- go-zero 的显式服务上下文和依赖传递。

同时适度借助 DDD，但不追求全面 DDD 化：DDD 用于明确业务边界、数据所有权和核心业务规则；组合根、配置边界、Worker 和生命周期仍使用独立的工程设计解决。

参考：

- [WuKongIM 启动入口](https://raw.githubusercontent.com/WuKongIM/WuKongIM/main/cmd/wukongim/main.go)
- [WuKongIM Server 生命周期](https://raw.githubusercontent.com/WuKongIM/WuKongIM/main/internal/server/server.go)
- [Kratos HTTP Server 装配](https://go-kratos.dev/docs/component/transport/http/)
- [go-zero 服务配置与上下文](https://go-zero.dev/reference/configuration/service-config/)

## 2. 核心结论

### 2.0 DDD 的使用范围

DDD 不是本次重构的总架构模板，而是解决业务边界问题的一组思想。

本项目采用：

- 战略层：使用限界上下文明确 IAM、SYS、Resource、Realtime 的职责和数据所有权；定时任务归属其操作数据的领域服务；
- 战术层：只在权限派生、认证规则、资源生命周期等复杂业务处使用领域服务或聚合；
- 简单 CRUD：保持直接的 Controller → Domain Service → Repository 结构，不为每张表创建 Aggregate、Factory 或完整 Use Case 层；
- 启动装配：由 Bootstrap 组合根负责，不交给 Domain；
- 运行管理：由 App 和 Worker 生命周期负责，不交给 Domain。

因此目标不是“全面 DDD”，而是：

```text
DDD                  定义业务边界和核心规则
依赖方向             约束代码依赖关系
组合根               连接具体实现
生命周期             管理启动、运行和关闭
服务私有配置         限制能力可见性
```

### 2.1 `main.go` 不是业务装配容器

`main.go` 只负责进程级职责：

- 创建进程上下文；
- 加载配置；
- 初始化进程日志和 Telemetry；
- 创建服务应用；
- 启动应用并等待退出；
- 处理最终错误。

`main.go` 不负责知道：

- 某个领域服务使用哪些数据库或缓存；
- 某个 HTTP 路由注册哪些 Controller；
- 某个 gRPC Server 注册哪些实现；
- 某个后台模块的启动顺序；
- 每个基础设施对象如何关闭。

### 2.2 组合根是唯一的依赖组装位置

每个服务应有一个服务私有的组合根，建议位于：

```text
application/<service>/internal/bootstrap/
```

组合根负责把具体实现连接起来：

```text
配置
  → 基础设施
      → 数据库 / Redis / 注册发现 / RPC Client / Storage
  → 领域对象
      → Repository / Domain Service / Module
  → 传输对象
      → HTTP / gRPC / WebSocket
  → 生命周期对象
```

组合根可以依赖所有服务私有实现，但业务层不能反向依赖组合根。

### 2.3 构造、启动、运行、关闭必须分离

服务装配必须区分四个阶段：

```text
New       只构造对象和依赖关系
Start     按依赖顺序启动资源
Run       等待上下文取消或运行错误
Stop      按逆序关闭资源
```

构造函数不应隐式启动 goroutine、监听端口或注册后台任务。

启动失败必须能够停止已经启动成功的组件，避免部分启动状态泄漏。

关闭顺序应与依赖关系相反：

```text
HTTP / WebSocket
  → gRPC
  → 领域模块
  → 注册发现 / RPC Client
  → Redis / Storage / Database
```

### 2.4 不使用依赖注入框架，但依赖必须显式

本项目不引入 Wire、Fx、Dig 或其他依赖注入框架。

采用普通 Go 构造函数表达依赖：

```go
service := domain.NewUserService(db, logger)
controller := controller.NewUserController(service, logger)
httpServer := transport.NewHTTPServer(controller, config)
```

允许使用小型、服务私有的依赖结构：

```go
type Dependencies struct {
    DB     *gorm.DB
    Logger logging.Logger
}
```

不允许使用一个拥有所有资源访问权的通用对象隐藏依赖。

## 3. Container 的终态决策

### 3.1 不保留业务层 Container 依赖

终态中，以下层不应接收 `container.Container`：

- Controller；
- Router；
- Domain Service；
- HTTP Server；
- gRPC Server；
- Health Handler。

这些组件只接收自己真正需要的依赖。

例如，不使用：

```go
controller.NewUserController(container)
```

而使用：

```go
controller.NewUserController(userService, logger)
```

### 3.2 不保留通用 Service Locator

终态不接受以下形式作为业务依赖：

```go
GetDB()
GetRedis()
GetStorage()
GetLogger()
GetModule("xxx")
```

这些方法可以存在于基础设施内部，但不能作为业务对象获取依赖的主要方式。

### 3.3 模块系统也采用显式依赖

模块接口不再依赖通用 Container，而是接收模块需要的明确依赖：

```go
type ModuleDependencies struct {
    DB            *gorm.DB
    Redis         *redis.Client
    Logger        logging.Logger
    RuntimeConfig *runtimeconfig.Store
}
```

如果某个模块需要访问其他模块，应通过明确的接口依赖表达，而不是通过字符串查找：

```go
type CaptchaDependencies struct {
    SMS   SMSProvider
    Email EmailProvider
}
```

## 4. 领域、传输和基础设施边界

### 4.1 基础设施层

负责创建和关闭外部资源：

- PostgreSQL；
- Redis；
- 对象存储；
- 注册中心；
- RPC Client Pool；
- Telemetry Exporter。

基础设施层不注册业务路由，不创建业务 Controller。

### 4.2 领域层

负责业务对象和业务规则：

- IAM Token、权限、用户、角色；
- SYS 菜单、字典、配置、操作日志；
- Resource 附件和资源元数据；
- Realtime Relay 和连接投递。

领域层不负责监听端口、不读取进程配置、不创建 Redis Client。

### 4.3 传输层

负责协议适配：

- HTTP 路由；
- gRPC 注册；
- WebSocket 握手和连接适配；
- 健康检查和服务注册元数据。

传输层只调用领域接口，不负责创建领域实现。

### 4.4 生命周期层

负责：

- 启动顺序；
- 启动失败回滚；
- 运行期间等待错误或取消；
- 逆序关闭。

生命周期层不包含业务规则。

## 5. 终态目录思想

每个非中间件服务可以采用以下结构：

```text
application/<service>/
├── main.go
└── internal/
    ├── bootstrap/
    │   ├── app.go
    │   ├── infra.go
    │   ├── domain.go
    │   ├── transport.go
    │   └── lifecycle.go
    ├── domain/
    ├── controller/
    ├── router/
    ├── transport/
    └── migrations/
```

不要求每个服务机械地创建全部文件。只有在职责真实存在时才创建对应文件。

例如 Realtime 没有数据库，就不创建迁移目录。

## 6. SYS 作为第一个试点是否合理

合理，SYS 适合作为第一个试点，原因如下：

1. 依赖关系比 IAM 简单，不包含登录、Token、第三方认证模块等复杂流程；
2. 同时具有数据库、Redis、IAM RPC Client、HTTP 和 gRPC，能够覆盖完整装配链路；
3. 有迁移、领域服务、路由和操作日志等典型业务能力；
4. 改造后可以验证 Container 是否真正从 Router、Controller、Domain 和 HTTP 层消失；
5. 不会直接影响 Realtime 的新架构，也不会先碰 Gateway 的复杂代理逻辑。

SYS 试点应验证以下终态目标：

```text
main.go
  → bootstrap.Load
  → bootstrap.New
  → app.Run

bootstrap.New
  → infra
  → migrations
  → IAM Client
  → SYS Domain
  → HTTP Server
  → gRPC Server

业务层
  → 只接收显式依赖
  → 不接收 container.Container
```

## 7. 判断重构是否成功

不能只看 `main.go` 是否变短，应检查：

- 依赖是否可以从构造函数参数直接看见；
- Controller 是否还需要 Container；
- Router 是否还需要 Container；
- Domain 是否还读取进程配置；
- 构造函数是否会隐式启动后台任务；
- HTTP、gRPC、模块是否拥有独立生命周期；
- 启动失败是否能够回滚已启动组件；
- 关闭顺序是否与依赖顺序相反；
- 是否可以在不启动完整进程的情况下测试领域对象；
- 是否减少了 `GetXXX()` 和字符串模块查找。
- 业务边界是否与数据所有权一致；
- 复杂业务规则是否集中在领域层，而不是散落在 Controller、Router 和 SQL 中；
- 简单 CRUD 是否避免了无必要的 DDD 层次和抽象。

如果只是从：

```text
main.go 的面条代码
```

变成：

```text
internal/server.Run 的面条代码
```

则不算完成重构。

## 8. 明确不做的事情

- 不引入依赖注入框架；
- 不创建跨服务的万能 `Application`；
- 不为了目录整齐给所有服务增加相同的空文件；
- 不保留业务层对通用 Container 的依赖；
- 不把所有基础设施强行塞到一个全局运行时对象；
- 不在本阶段改变 HTTP、gRPC、数据库或消息语义；
- 不在 SYS 试点完成后自动改造其他服务。
- 不把所有模块强行改造成完整 DDD 战术模型；
- 不为简单 CRUD 增加 Aggregate、Factory、Domain Event 等无实际收益的抽象。

## 9. 试点确认点

在继续输出完整终态重构文档前，需要确认以下原则：

1. SYS 作为第一个试点；
2. 终态移除业务层对 `container.Container` 的依赖；
3. 模块系统也采用显式依赖，不保留通用 Service Locator；
4. 使用服务私有 `internal/bootstrap` 作为组合根；
5. 不使用依赖注入框架；
6. 允许按服务逐个重构和验证；
7. 第一阶段只改 SYS；试点通过后再按规划逐个改造 Resource、IAM、Realtime、Gateway，并删除无独立职责的 Scheduler。
