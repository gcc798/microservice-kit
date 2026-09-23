# 新增或删除微服务

## 新增前决策

- 只有出现独立数据所有权和独立部署需求才新增 `native/application/<service>`。
- 明确服务负责的数据、HTTP/gRPC 能力、外部依赖、实例数量和故障边界。
- 检查依赖图，禁止新增循环依赖；能归入现有 IAM、SYS、Resource 或 Realtime 时不新增服务。

## 最小结构

服务应拥有自己的 `main.go`、私有 `internal/config`、`internal/bootstrap`、业务 controller/router/domain/model，以及需要时的迁移和 worker。组合根在服务自己的 `internal/bootstrap`，根 `native/internal` 只放跨进程技术设施。

配置模板只声明该服务真正使用的依赖。数据库迁移归属数据所有者，并使用该服务自己的版本表。

新增或删除服务时检查：

- Consul/注册发现、HTTP/gRPC 注册和健康检查；
- Docker Compose、部署脚本和横向扩展配置；
- README、AGENTS.md、架构文档、能力地图；
- 服务配置、日志配置、迁移、生成代码和测试；
- 删除所有旧入口、旧容器名、旧链接和旧规划文档。

不要创建共享业务包、聚合 API 服务、Scheduler 聚合进程或只为未来需求预建的抽象。
