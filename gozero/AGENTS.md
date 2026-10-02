# AGENTS.md

本目录是 `native/` 业务基线的 go-zero 框架实现。`native/` 定义业务语义和对外 HTTP 契约；本目录只选择适合 go-zero 的工程组织方式，不得自行改变接口路径、请求/响应、错误语义、鉴权规则或业务约束。

## 长期服务边界

- 按 `native` 的限界上下文划分 IAM、SYS、Resource、Realtime 和 Gateway。
- IAM 拥有用户、角色、组织、菜单、API 权限、认证客户端和登录会话。
- SYS 拥有配置、字典、登录日志和操作日志。
- Resource 拥有附件及对象存储访问。
- Realtime 拥有实时连接；Gateway 只负责统一入口和转发，不拥有业务数据。
- go-zero 领域服务可以保留 API/RPC 分层，但数据模型、迁移和数据库连接只能由所属 RPC 服务持有；API 服务不得直接访问数据库。
- 不跨服务共享业务 model、logic 或数据库表；跨边界调用使用明确的 RPC 契约。

## GORM 与 goctl

- 数据访问统一使用 GORM，不使用 go-zero `sqlx`、`sqlc` 或其缓存 model。
- `goctl-template/model/` 是本工程的 GORM model 模板；通过根 Makefile 中的生成目标调用固定版本的 `go tool goctl`。
- 不手工修改 `*_gen.go`。先修改模板、数据库迁移或协议，再重新生成。
- 简单 CRUD、筛选、分页、关联和事务优先使用 GORM API；只有 GORM 表达明显更差的复杂查询才使用参数化原生 SQL。
- 所有数据库操作传递 `context.Context`；多步写入使用 GORM 事务。
- 数据库结构由各服务私有 Goose migration 管理，不使用 `AutoMigrate`。
- 新增或移动数据表时，必须同步更新 Makefile 中对应服务的 model 生成表清单。

## 生成与验证

```bash
cd gozero
make gen-models
make verify-gorm-template
make verify
```

- `make verify-gorm-template` 必须证明自定义模板生成的代码可编译，并能对临时 PostgreSQL 表执行 CRUD。
- 新增 API/RPC 生成目标时沿用 `GOCTL := go tool goctl`，禁止依赖开发机全局安装的未知 goctl 版本。
- 本文件只记录长期约束，不记录迁移进度、临时拓扑或具体端口。普通架构演进只更新 `README.md` 和架构文档；仅当本目录的长期约束发生变化时才修改本文件。
- 修改协议、配置或运行方式时，同步更新 `README.md` 和最终架构文档中的相关内容。
- 本项目不要求兼容旧服务布局、旧接口、旧配置或旧数据；除非开发者明确提出，否则完成迁移后删除过渡代码。
