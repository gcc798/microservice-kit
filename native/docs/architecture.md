# Native 架构设计

## 1. 目标和范围

`native` 是当前仓库的业务语义与 HTTP 契约基线，只保留真实微服务部署形态。系统按数据所有权、事务边界和独立部署需求拆分，不同时维护单体装配入口，也不为未来需求预建服务。

当前部署单元只有五类业务进程：

| 进程 | 主要职责 | 数据所有权 | 可否水平扩展 |
| --- | --- | --- | --- |
| Gateway | HTTP/WS 唯一入口、鉴权前置、动态路由、反向代理 | 无业务数据 | 可以 |
| IAM | 用户、角色、API 权限、Token、Session | IAM 表和 Redis 会话 | 可以 |
| SYS | 组织、菜单、字典、配置、登录日志、操作日志 | SYS 表 | 可以 |
| Resource | 附件元数据和对象存储 | Resource 表和对象存储 | 可以 |
| Realtime | WebSocket 连接、心跳、实时消息广播 | 连接状态，不拥有领域表 | 可以 |

Consul、etcd、Nacos 是可选注册中心实现；PostgreSQL、Redis、RustFS/S3 是基础设施，不属于业务服务进程。

## 2. DDD 的使用程度

本项目采用“足够表达边界、不过度建模”的轻量 DDD，而不是完整战术 DDD 框架。

### 必须使用的部分

- **限界上下文**：IAM、SYS、Resource、Realtime 是独立上下文和部署边界。
- **聚合和不变量**：同一事务内需要保持一致的业务规则由所属 domain service 和 model 负责，例如角色授权来源、用户会话轮换、附件状态变更。
- **领域服务**：跨多个模型或依赖外部能力的业务流程放在 `application/<service>/internal/domain`。
- **端口/适配**：跨进程调用通过 Proto、`contracts.go` 的最小 API 和 Remote 适配；对象存储、Redis、注册中心通过基础设施接口或启动装配注入。
- **领域事件式通知**：SYS、Resource 通过 Realtime gRPC 发布尽力而为的实时消息；不能丢失的事实必须先由所属领域持久化。

### 不强制使用的部分

- 不为每个 CRUD 方法创建独立聚合、仓储接口或命令对象。
- 不为单一实现增加通用工厂、容器或共享业务基类。
- 不引入 Casbin、事件总线或分布式事务，除非出现明确的数据所有权和一致性需求。
- 请求 DTO、响应 DTO 和数据库模型可以保持简单；它们只有在需要隔离边界或承载规则时才单独建模。

判断标准是：规则是否需要事务一致性、是否跨进程、是否被多个入口复用。自描述的查询和简单映射留在当前服务的 controller/domain 中即可。

## 3. 组合根

每个进程的 `main.go` 只负责加载配置、初始化日志、创建组合根并运行生命周期；具体依赖装配位于该服务自己的 `internal/bootstrap`。组合根是唯一允许知道完整依赖图的位置。

```text
application/<service>/main.go
  -> internal/config.Load
  -> internal/bootstrap.New
       -> registry / database / redis / storage / jwt
       -> domain services
       -> HTTP server / gRPC server / workers
  -> App.Run(ctx)
  -> App.Stop(ctx)
```

组合根约束：

1. 服务私有配置只能在 `application/<service>/internal/config` 定义；共享 `internal/config` 只提供原子配置类型、环境变量绑定和加载器。
2. 领域代码不读取环境变量、不创建数据库/Redis/HTTP 客户端、不查询注册中心。
3. `internal/transport.StartRegisteredGRPC` 只接收端口、实例、路由和注册回调，不接收共享全量配置。
4. Realtime HTTP server 只接收 HTTP、CORS、Auth、WebSocket 参数和业务依赖。
5. 组合根创建的资源必须由同一个组合根关闭；关闭顺序与依赖方向相反。

## 4. 目录边界

```text
application/<service>/
  main.go                         进程入口
  internal/config                 服务私有配置
  internal/bootstrap              组合根和依赖装配
  internal/controller             HTTP 控制器
  internal/router                 HTTP 路由
  internal/request / response     HTTP 边界模型
  internal/domain                 领域服务、模型和 RPC 适配
  internal/migrations              该服务独有的 Goose 迁移
  internal/workers                 该服务独有的后台任务

internal/
  api                             生成的 gRPC 契约和薄 Remote 适配
  database                        数据库连接和 GORM 插件
  httpserver / httpx               共享 HTTP 生命周期和路由适配
  registry                        注册发现实现
  transport                       gRPC 服务端和 Client Pool
  platform                        Redis、JWT、对象存储、验证码等技术能力
  modules                         IAM 运行期集成模块生命周期
  utils                            分页、密码、时间、IP、错误等稳定工具
```

根 `internal` 不放 IAM、SYS、Resource 的业务实现，不创建共享 API 应用，也不重新引入 `internal/infra`、`internal/container` 或 Scheduler 聚合装配层。

## 5. 请求和调用路径

### HTTP / WebSocket

```text
web-react
  -> Gateway
       -> IAM：校验 Access Token
       -> IAM / SYS / Resource：按注册元数据动态代理
       -> Realtime：代理 WebSocket 握手
            -> IAM gRPC：再次校验 Token
            -> Realtime 本地 Hub：持有连接和心跳
```

Gateway 启动时和运行期间从注册中心读取服务实例及 HTTP 路由，每 5 秒刷新。健康探针和 Metrics 不发布为业务代理路由。

### 实时消息

```text
SYS / Resource
  -> Realtime gRPC PublishToUsers
  -> Redis Pub/Sub
  -> 每个 Realtime 实例
  -> 仅持有目标用户连接的实例发送 WebSocket 消息
```

该通知是尽力而为的实时提醒。不能丢失的业务通知由所属领域持久化，客户端重新连接后通过领域 API 查询最终状态。

## 6. 数据和迁移边界

- IAM、SYS、Resource 各自拥有迁移目录和 Goose 版本表：`goose_iam_version`、`goose_sys_version`、`goose_resource_version`。
- PostgreSQL 可以是同一实例，但表结构、迁移执行和业务模型按服务隔离。
- Realtime 不创建数据库迁移目录；连接状态只存在于进程内 Hub，跨实例消息通过 Redis Pub/Sub。
- Redis 既承载 IAM Session，也承载运行时配置缓存、分布式锁和 Realtime 广播；使用不同键空间，不共享领域模型。
- Resource 通过 S3 兼容接口访问 RustFS；业务只依赖 `storage.Storage`，不依赖具体 SDK。

## 7. Worker 和生命周期

后台任务属于数据所有者，不单独创建 Scheduler 进程：

- SYS 的 `logCleanup` 清理 SYS 日志。
- Resource 的 `expiredAttachmentCleanup` 清理 Resource 附件。
- 每个 Worker 使用所属服务 Context，通过 Redis 锁抢占执行窗口，多个副本可以同时运行而不会重复执行同一时间窗口。

启动顺序是资源初始化、领域服务构建、HTTP/gRPC 注册、Worker 启动；关闭顺序相反：停止接收新请求、停止 Worker、注销服务实例、关闭 gRPC/HTTP、关闭 Redis/数据库/对象存储。

## 8. 横向扩展原则

- Gateway 可以多副本部署；前端只访问 Gateway。
- IAM、SYS、Resource 使用无状态 HTTP/gRPC 进程，状态放在各自数据库和共享 Redis 中。
- Realtime 不要求粘性会话；任意实例都可以接受 WebSocket 握手，消息通过 Redis 广播到全部实例。
- 服务发现实例必须注册真实可达的 HTTP/gRPC 地址，不能注册容器内部不可达的 `127.0.0.1`。
- Consul 中的实例健康状态决定 Gateway 和 Client Pool 是否选择该实例。

## 9. 变更规则

新增能力优先归入 IAM、SYS、Resource 或 Realtime。只有同时满足以下条件才新增进程：

1. 存在清晰、独立的数据所有权；
2. 存在独立部署、扩容或故障隔离需求；
3. 跨进程契约比进程内调用更能降低耦合；
4. 已有服务无法在不破坏事务边界的前提下承载该能力。

跨进程行为先修改 Proto 和契约适配，再修改服务端 domain 实现，最后更新 Gateway 路由和前端契约。生成代码始终通过生成命令更新，不手工编辑。
