# go-zero 最终架构

`gozero` 以 `native` 的业务语义和 HTTP 契约为基线，采用 go-zero 常见的 API/RPC 分层。API 进程只处理 HTTP 边界与 RPC 适配；数据库、GORM model、Goose migration 和领域写入只存在于所属 RPC 进程。

## 部署单元

| 进程 | 端口 | 职责 | 数据所有权 |
| --- | --- | --- | --- |
| Gateway | HTTP 9009 | 唯一外部入口、IAM Token 与 API 权限前置校验、HTTP/WS 反向代理、统一 Swagger、操作审计 | 无 |
| IAM API / RPC | HTTP 9011 / gRPC 9001 | 认证、验证码、用户、角色、组织、菜单、API 权限 | IAM 表、Redis 登录状态 |
| SYS API / RPC | HTTP 9012 / gRPC 9002 | 字典、配置、登录日志、操作日志 | SYS 表 |
| Resource API / RPC | HTTP 9013 / gRPC 9003 | 附件、对象存储 | Resource 表、S3/RustFS |
| Realtime | HTTP 9014 / gRPC 9004 | WebSocket 连接、心跳和面向用户的实时投递 | 连接状态、Redis Pub/Sub |

端口来自示例 YAML，可按部署环境修改。领域 API 端口只供 Gateway 访问，不作为前端兼容入口。

## 请求路径

```text
web-react
  -> Gateway
       -> IAM RPC ValidateAccessToken
       -> IAM API -> IAM RPC -> PostgreSQL / Redis
       -> SYS API -> SYS RPC -> PostgreSQL
       -> Resource API -> Resource RPC -> PostgreSQL / S3
       -> Realtime /realtime/websocket
            -> IAM RPC ValidateAccessToken
            -> 本地连接 Hub
```

所有 RPC Server 使用 go-zero 原生 etcd 注册，RPC Client 按服务 key 发现实例，不配置 `Target` 直连地址。IAM、SYS、Resource API 启动时直接读取 `rest.Server.Routes()`；Realtime 在注册 WebSocket handler 的同一处声明其公开路由。它们将服务名、HTTP endpoint 和公开路由注册到 `/microservice-kit/http/<service>`，健康检查与指标路由不发布。

Gateway 订阅 `/microservice-kit/http` 前缀并原子更新动态路由表，不维护 `Backends` 或手写的路径到服务映射。同一服务的多个 endpoint 采用轮询；如果两个不同服务声明同一组 method/path（路径参数名不同也视为同一路由），Gateway 拒绝冲突快照并保留上一份有效路由表。实例租约失效后会自动移出路由表。

Gateway 自身注册到 `/microservice-kit/services/gateway`，用于运维侧发现统一入口，不参与业务路由表。Gateway 的 `/swagger/index.html` 使用 `make swagger` 从 IAM、SYS、Resource 的 `.api` 文件生成统一文档。

受保护路由在 Gateway 中显式声明 `resource` 和 `action`，并调用 IAM RPC 基于当前 GORM 授权表检查；不信任 JWT 中可能过期的权限，也不维护 Casbin/Redis 权限副本。未登记的 `/api/v1` 路由默认拒绝转发。

AccessToken 除签名和用户状态外还必须命中 Redis 活跃会话。默认禁止同一用户、同一客户端并发登录；刷新令牌使用 `GETDEL` 单次消费，刷新、重新登录和登出都会撤销旧会话。密码登录连续失败 5 次后锁定 10 分钟，成功登录会更新用户最后登录信息，并将成功或失败记录写入 SYS。

验证码、短信、邮件和微信配置以 SYS `s_config` 为唯一运行时来源。IAM 每次使用前读取对应配置，因此修改后无需重启；SYS 在写入已知配置时校验 JSON 结构和启用能力所需凭据。短信使用阿里云，邮件使用 SMTP。

Gateway 将外部 HTTP 操作批量写入 SYS，跳过健康检查和操作日志自身接口，查询参数和 JSON 请求体中的密码、Token、验证码和密钥会脱敏。SYS 每日清理超过 90 天的登录/操作日志，Resource 每日清理过期附件；两个 Worker 都由数据所有者 RPC 进程运行，并以 Redis 执行窗口避免多实例重复执行。

go-zero 原生 REST/gRPC 中间件负责 HTTP、RPC 指标和追踪；Gateway、Realtime 的自定义 HTTP Server 使用同一套中间件，GORM 与 Redis 也接入 OpenTelemetry。指标由每个进程的独立 DevServer 暴露，避免同机部署时端口冲突。

## 实时消息

```text
业务调用方 -> Realtime gRPC PublishToUsers
           -> Redis channel microservice-kit:realtime:deliver:v1
           -> 每个 Realtime 实例
           -> 当前实例持有的目标用户 WebSocket
```

发布请求最多包含 1000 个正数用户 ID，服务端去重，并要求消息类型非空、数据为合法 JSON。该链路是尽力而为的实时提醒，不能代替领域持久化。

## 数据与生成边界

- IAM、SYS、Resource 分别使用 `goose_iam_version`、`goose_sys_version`、`goose_resource_version`。
- Model 由 `goctl-template/model` 生成，统一依赖 GORM；普通 CRUD、筛选、分页和事务使用 GORM API。
- 只有递归字典查询/删除保留参数化原生 SQL，因为递归 CTE 比 ORM 拼装更清晰。
- API/RPC 代码通过 `make gen-all` 生成，模型通过 `make gen-models` 生成；`make verify-gorm-template` 对自定义模板执行生成、编译和真实 PostgreSQL CRUD 验证。
