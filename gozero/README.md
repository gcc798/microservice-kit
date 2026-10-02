# go-zero 实现

`gozero` 使用 go-zero 提供 HTTP/RPC 传输和代码生成，业务语义以 `../native` 为基线。

## 服务拓扑

- `application/gateway`：唯一外部 HTTP/WS 入口，完成 IAM Token 与 API 权限前置校验并代理到各领域 API 和 Realtime。
- `application/iam-api` / `iam-rpc`：认证、验证码、用户、角色、组织、菜单和 API 权限。IAM API 管理 Token/Refresh Token，IAM RPC 持有关系数据、GORM model 与 Goose migration。
- `application/sys-api` / `sys-rpc`：字典、配置、登录日志和操作日志。SYS RPC 独立持有数据库、GORM model 与 Goose migration。
- `application/resource-api` / `resource-rpc`：附件元数据和对象存储。Resource RPC 独立持有数据库、GORM model 与 Goose migration。
- `application/realtime`：同一进程提供 `PublishToUsers` RPC 和 `/realtime/websocket`，维护连接心跳，并通过 Redis Pub/Sub 向持有目标用户连接的实例投递消息。

完整边界、调用链和端口见 [`docs/architecture.md`](docs/architecture.md)。长期开发约束见 [`AGENTS.md`](AGENTS.md)。

## 代码生成

项目通过 `go.mod` 的 `tool` 指令固定 goctl 版本，统一使用 `go tool goctl`，不依赖开发机全局安装版本。

API 和 RPC 使用 goctl 官方模板：

```bash
make gen-iam-api
make gen-sys-api
make gen-resource-api
make gen-iam-rpc
make gen-sys-rpc
make gen-resource-rpc
make gen-realtime
```

数据库模型使用仓库内 `goctl-template/model` 的 GORM 模板：

```bash
make gen-models PG_URL='postgres://user:pass@127.0.0.1:5432/db?sslmode=disable'
```

模型按 IAM、SYS、Resource 的数据所有权分别生成；组织与菜单归 IAM。生成的模型仅依赖 GORM，不使用 go-zero SQLX。模板生成、编译和真实 PostgreSQL CRUD 可以通过以下命令验证：

```bash
make verify-gorm-template
```

默认连接 `postgres://postgres:post123@127.0.0.1:5433/postgres?sslmode=disable`。需要使用其他测试库时传入 `GOCTL_GORM_TEST_DSN`。验证命令只创建并删除自己的临时 schema。

常规检查使用：

```bash
make verify
```

该命令会拒绝重新引入 go-zero SQLX。

构建当前全部进程：

```bash
make build-all
```

示例配置使用 Gateway `9009`；RPC `9001` 至 `9004`；IAM、SYS、Resource、Realtime 的内部 HTTP 端口为 `9011` 至 `9014`。前端只连接 Gateway。

本地启动顺序：

```bash
make run-sys-rpc
make run-resource-rpc
make run-iam-rpc
make run-realtime
make run-iam-api
make run-sys-api
make run-resource-api
make run-gateway
```

## 数据访问约定

- 简单 CRUD 优先使用 GORM API；泛型 API 更清晰时使用 `gorm.G[T]`，否则使用传统 API。
- 复杂联表、统计或数据库特有查询才使用原生 SQL。
- 事务使用 `db.WithContext(ctx).Transaction`，事务内只使用回调提供的 `tx`。
- 表结构由所属服务的 Goose migration 管理，不使用 `AutoMigrate`。
- 不生成 Repository 基类、单实现接口或 ORM 包装层。
