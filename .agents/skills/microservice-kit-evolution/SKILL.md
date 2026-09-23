---
name: microservice-kit-evolution
description: Use when changing the native microservice topology, adding or removing a service, changing cross-service contracts, or completing a refactor that must keep AGENTS.md, README.md, final architecture docs, protocols, configuration, observability, and the capability map synchronized. Do not use for isolated business fixes that do not change architecture or public contracts.
---

# Microservice Kit 演进

把一次架构改动完成为“代码、契约、文档、生成产物、验证”一致的最终状态。

## 工作顺序

1. 先阅读仓库根目录 `AGENTS.md` 和 `native/docs/architecture.md`，确认当前架构，不根据旧对话或旧规划文档推断规则。
2. 判断改动类型，再按需阅读参考资料：
   - 新增、删除或拆分服务：`references/service-addition.md`
   - 跨服务 HTTP/gRPC 调用：`references/cross-service-call.md`
   - Proto、配置、链路追踪或公开契约变化：`references/protocol-change.md`
   - 文档和能力地图同步：`references/documentation-governance.md`
   - 完成前的验证：`references/validation-matrix.md`
3. 先确定数据所有权、部署边界和依赖方向，再修改代码。没有独立数据所有权或部署需求时，不新增服务。
4. 使用现有的服务私有 `internal/config`、`internal/bootstrap`、共享 `internal/transport`、注册发现和 `ClientPool` 模式；不要重新引入共享全量配置、聚合业务装配层或根 `internal` 业务包。
5. 源文件变更完成后，同步最终文档和生成产物；删除失效规划文档、旧术语、旧链接和不再使用的代码。
6. 按验证矩阵执行检查。只报告实际执行通过的检查；跳过的外部集成检查必须说明缺少的环境或依赖。

## 文档职责

- `AGENTS.md`：永久有效的仓库规则、边界和禁止事项。
- `native/docs/architecture.md`：当前最终架构，不记录重构过程。
- `native/docs/configuration.md`、`protobuf.md`、`opentelemetry.md`：各自领域的稳定契约。
- `README.md`：入口、快速开始和常用命令。
- 能力地图 JSON：唯一源文件；HTML、图片和检查 sidecar 是生成产物。

不要新增“规划说明”“重构过程”类长期文档。一个事实只保留一个详细来源，其他位置只链接或简述。

## 完成标准

改动只有同时满足以下条件才算完成：

- 代码遵守当前服务边界和组合根规则；
- Proto/OpenAPI/能力地图等生成产物已更新；
- 相关文档、README、AGENTS.md 已同步，且没有失效链接或旧架构残留；
- 测试、静态检查和适用的真实服务验证均已执行；
- `git diff --check` 通过，工作区中没有无关生成物或临时文件。

若用户只要求局部业务修复且不涉及上述边界，不要为了形式修改全部文档。
