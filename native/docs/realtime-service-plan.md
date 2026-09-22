# Realtime 服务拆分规划

本文规划把当前位于 IAM 进程内的 WebSocket 能力拆分为可多实例部署的 `realtime` 服务。`native` 是首个落地目标和 HTTP 契约基线；Kratos 与 go-zero 在 native 验证完成后提供相同的对外行为。

本文只定义当前需要的用户实时消息能力，不预建房间、群组、离线消息、IoT 设备接入或通用事件平台。

## 目标

- IAM 只负责身份、会话、Token 和权限，不再维护长连接。
- Realtime 独立维护 WebSocket 连接、心跳和用户消息投递。
- Realtime 支持多实例部署，任一实例收到发布请求后都能把消息送到持有目标用户连接的实例。
- 业务服务通过稳定的内部契约发布消息，不直接依赖 Redis channel。
- 前端只使用一套 WebSocket 契约，不感知 Realtime 实例数量。

## 当前问题

当前 WebSocket Hub 由 `application/iam` 启动，连接只存在于单个 IAM 进程内：

- `application/iam/main.go` 注册 `WebSocketModule`。
- `application/iam/internal/router/common.go` 提供 `/resource/websocket`。
- `internal/platform/websocket.Hub` 使用进程内 map 保存用户连接。
- 实例之间没有连接目录或消息传播机制。

因此，用户连接到 IAM-A 后，IAM-B 无法通过自己的 Hub 向该连接发送消息。继续把连接能力放在 IAM 中，还会让 IAM 的扩容、发布和故障边界受到长连接约束。

## 已确定的设计

### 服务边界

| 组件 | 职责 |
| --- | --- |
| Gateway | 统一 HTTP/WS 入口、Token 前置校验、服务发现和反向代理。 |
| IAM | 登录、会话、Token 签发与校验、权限判断。 |
| Realtime | WebSocket 升级、连接与心跳管理、Redis 订阅、本地消息投递。 |
| SYS、Resource、Scheduler 等 | 产生业务结果，通过 Realtime 内部 API 请求实时推送。 |
| Redis Pub/Sub | 只负责 Realtime 实例之间的瞬时消息广播。 |

Realtime 不拥有 IAM 数据库，也不执行数据库迁移。它只依赖注册中心、Redis 和 IAM gRPC。

### 进程与代码位置

新增真实部署单元：

```text
application/realtime/
├── main.go
├── conf.example.yaml
├── zaplogger.example.yaml
└── internal/
    ├── domain/
    └── relay/
```

Realtime 复用 `internal/platform/websocket` 中的连接基础设施，业务发布和 Redis Relay 保持在 Realtime 私有目录。IAM 删除 WebSocket 路由、配置和模块注册；不为单一实现增加跨进程共享业务抽象。

服务注册名使用 `realtime`。建议默认 HTTP 端口为 `9013`，内部 gRPC 端口为 `9113`。

### 对外 WebSocket 契约

新的规范路径为：

```text
GET /realtime/websocket
```

这是新工程，迁移时直接删除旧的 `/resource/websocket`，不保留兼容别名。Gateway 根据 Realtime 注册的路由把升级请求代理到任一健康实例；WebSocket 建立后，该 TCP 连接在整个生命周期内固定由选中的实例处理，不需要额外粘性会话配置。

浏览器和小程序无法统一设置自定义握手 Header，因此第一阶段沿用当前 query Token 方式。Gateway 和 Realtime 都不得记录完整 URL 或 Token。后续若安全要求提高，再引入一次性 WebSocket ticket，不在本次拆分中同时实现。

### 握手鉴权

握手流程保持 IAM 为身份事实来源：

1. 客户端携带 Access Token 请求 Gateway。
2. Gateway 调用 IAM `ValidateAccessToken`，校验通过后选择 Realtime 实例并代理升级请求。
3. Realtime 再通过 IAM `ValidateAccessToken` 校验原始 Token，并从返回的 claims 获取 `user_id`、`client_id` 和设备类型。
4. Realtime 只使用校验结果建立连接，不接受客户端直接提交的 `userId` 作为身份。

双重校验与现有领域服务的安全边界一致，也避免绕过 Gateway 直连 Realtime 时伪造用户身份。Realtime 不读取 IAM 数据库，不持有 JWT 签名密钥。

### 业务服务发布接口

新增 `api/realtime/v1/realtime.proto`，只提供当前需要的按用户集合推送。同一批用户共享一份消息，避免业务服务为相同内容重复发起 gRPC 请求：

```proto
service RealtimeService {
  rpc PublishToUsers(PublishToUsersRequest) returns (PublishToUsersResponse);
}

message PublishToUsersRequest {
  repeated int64 user_ids = 1;
  string type = 2;
  bytes data_json = 3;
}

message PublishToUsersResponse {}
```

业务服务通过现有注册中心和 `transport.ClientPool` 调用 Realtime，不直接执行 Redis `PUBLISH`。这样 Redis channel、序列化和多实例广播都由 Realtime 自己拥有。Realtime 在发布前去重用户 ID，拒绝空列表和非正数 ID，并限制单次请求的用户数量，避免一次请求占满发布和连接写入资源。

第一阶段不增加全体广播、房间、按角色推送或设备定向接口；出现真实调用方后再扩展契约。

`PublishToUsers` 成功只表示消息已发布到 Realtime 集群，不表示任一用户在线、客户端已经收到或已经处理。调用方不能把它当作业务事务成功条件。

### Redis Pub/Sub 模型

所有 Realtime 实例订阅同一个版本化 channel：

```text
microservice-kit:realtime:deliver:v1
```

内部消息格式：

```json
{
  "userId": 42,
  "type": "notification.created",
  "data": {}
}
```

完整链路如下：

```text
业务服务
  -> Realtime gRPC PublishToUsers（注册中心选择任一实例）
  -> 该实例 PUBLISH 到 Redis
  -> 每个 Realtime 实例收到同一消息
  -> 持有目标 userId 本地连接的实例执行发送
  -> 其他实例忽略
```

订阅不能使用竞争消费或让多个实例共享同一消费者，因为目标连接可能位于任意实例。一个用户同时连接多个实例时，每个本地连接都收到消息，这是预期行为。

第一阶段使用单 channel 广播，复杂度最低。当实例数乘以消息量已经成为实际瓶颈时，再引入分片 channel 或用户到实例的路由表。

### 交付语义

Redis Pub/Sub 提供即时、尽力而为的投递：

- Realtime 实例断线期间不会补收消息。
- 客户端离线时消息不会保存。
- 网络中断时可能出现消息未送达。
- 本次不实现客户端 ACK、重放和去重。

不能丢失的通知必须先由所属领域持久化，WebSocket 只负责实时提醒；客户端重连后通过领域 HTTP API 查询最终状态。未来需要统一离线通知时，再单独设计 Notification 领域和 Outbox，不把可靠性伪装在 Pub/Sub 之上。

### 连接生命周期

- 一个用户允许存在多个连接，消息发送到该用户的全部本地连接。
- 心跳、读写超时和最大连续超时次数沿用当前行为。
- Realtime 收到终止信号后先停止接收新连接，再关闭现有连接；客户端负责退避重连并由 Gateway 选择其他实例。
- 实例退出不迁移连接状态，也不把连接明细写入 Redis。
- IAM 注销或强制下线需要关闭连接时，通过 `PublishToUsers` 的后续专用控制接口扩展；本阶段不复用业务消息伪装控制命令。

## 配置归属

Realtime 配置只包含自身实际依赖：

```yaml
server:
  port: 9013
grpc:
  port: 9113
service:
  id: ""
  advertiseHost: "127.0.0.1"
registry:
  driver: consul
  address: http://127.0.0.1:8500
  prefix: microservice-kit
  namespace: ""
  group: ""
  username: ""
  password: ""
redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
auth:
  tokenHeader: Authorization
cors:
  enabled: false
websocket:
  timeoutEnabled: true
  readTimeoutSeconds: 60
  writeTimeoutSeconds: 10
  heartbeatEnabled: true
  maxReadTimeouts: 3
```

IAM 删除 `websocket.*` 配置，但继续保留自己的 Redis 会话配置。Realtime 不配置数据库、JWT secret、短信、邮件或对象存储。

## 部署目标

Compose、Kubernetes 和构建脚本新增 `realtime` 目标：

- Gateway 发现并代理 Realtime。
- Realtime 默认启动两个副本以验证跨实例消息投递。
- IAM 可以独立扩容，不再受 WebSocket Hub 限制。
- Realtime readiness 检查 Redis 订阅和自身注册状态；IAM 在实际握手时通过 Token 校验确认可用。
- Realtime 不配置 HPA 初始策略；先记录连接数、发送失败数和 Redis 重连次数，再根据实际负载设置。

Kubernetes 终止宽限期必须覆盖 HTTP server 停止和现有连接关闭。发布 Realtime 会导致连接重连，但不影响 IAM 的登录和权限接口。

## 可观测性

沿用现有日志和 OpenTelemetry 基础设施，至少记录：

- 当前连接总数和按实例连接数；
- WebSocket 建立、正常关闭和异常关闭次数；
- Redis 订阅状态与重连次数；
- 发布成功、序列化失败、Redis 发布失败；
- 本地命中用户连接数和实际发送失败数。

日志包含 `userId`、消息 `type` 和服务实例 ID，但不记录 Access Token 或完整消息正文。

## 实施阶段

### 阶段一：建立 Realtime 服务

1. 新增 Realtime 配置类型、示例配置和服务入口。
2. 新增 Realtime Proto、生成代码、Remote 适配和 gRPC server。
3. 由 Realtime 接入现有 WebSocket Hub 与 Handler，保持连接基础设施与业务发布逻辑分离。
4. 增加 Redis publisher/subscriber，完成多实例本地投递。
5. 为 Hub、Pub/Sub 消息解析和 gRPC 发布链路保留最小测试。

### 阶段二：切换流量并清理 IAM

1. Gateway 服务发现列表加入 Realtime，WebSocket 特殊鉴权路径改为 `/realtime/websocket`。
2. 删除 IAM 的 WebSocket 路由、模块注册和配置。
3. 删除不再使用的 `internal/modules.WebSocketModule`；连接基础设施保留在 `internal/platform/websocket`。
4. 更新 README、配置文档和 OpenAPI 生成结果。

阶段二与阶段一在同一版本完成，避免同时存在两个连接入口和双写逻辑。

### 阶段三：部署验证

1. Dockerfile、Makefile、Compose、启动脚本和 Kubernetes 清单加入 Realtime。
2. 至少启动两个 Realtime 实例。
3. 分别建立落到不同实例的两个用户连接。
4. 调用 `PublishToUsers` 发送一批用户，确认无论 gRPC 请求落到哪个实例，目标用户的所有连接都收到消息。
5. 重启其中一个 Realtime 实例，确认其连接自动重连到健康实例，IAM 请求不受影响。

## 验收标准

- IAM 代码、配置和部署清单不再包含 WebSocket Hub 或连接参数。
- Realtime 在两个及以上实例运行时，可以向任意实例持有的用户连接推送消息。
- 业务服务只依赖 Realtime gRPC 契约，不依赖 Redis channel 名和内部消息结构。
- 伪造 `userId` query 或内部 Header 不能建立其他用户的连接。
- Redis 不可用时 Realtime readiness 失败，发布请求返回明确错误，不误报为已交付。
- Realtime 实例退出后，IAM 的登录、Token 校验和权限接口继续正常工作。
- `make verify` 通过，Compose 和 Kubernetes 多实例场景完成手工或集成验证。

## 本次不做

- 持久消息、离线补发、客户端 ACK 和重放；
- 房间、群组、角色广播和通配订阅；
- IoT 设备接入、MQTT 或设备在线状态；
- Redis 中保存全部连接目录；
- Kafka、NATS 或 Redis Streams；
- 为未来协议预建通用插件或接口层。

这些能力出现真实业务需求后再设计，不阻塞当前从 IAM 中拆出用户实时连接能力。
