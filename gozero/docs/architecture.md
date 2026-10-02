# go-zero 最终架构

`gozero` 以 `native` 的业务语义和 HTTP 契约为基线，采用 go-zero 常见的 API/RPC 分层。API 进程只处理 HTTP 边界与 RPC 适配；数据库、GORM model、Goose migration 和领域写入只存在于所属 RPC 进程。

## 部署单元

| 进程 | 端口 | 职责 | 数据所有权 |
| --- | --- | --- | --- |
| Gateway | HTTP 9009 | 唯一外部入口、IAM Token 与 API 权限前置校验、HTTP/WS 反向代理 | 无 |
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

Gateway 使用配置中的后端地址构建固定路由表。go-zero 版本当前不引入额外注册中心抽象；需要多实例发现时再将 `Backends` 替换为 go-zero 服务发现，不改变领域 HTTP 契约。

受保护路由在 Gateway 中显式声明 `resource` 和 `action`，并调用 IAM RPC 基于当前 GORM 授权表检查；不信任 JWT 中可能过期的权限，也不维护 Casbin/Redis 权限副本。未登记的 `/api/v1` 路由默认拒绝转发。

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
