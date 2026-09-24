# Native 配置说明

本文以当前代码为准，说明五类进程的配置文件、环境变量和启动约束。配置加载器位于 `internal/config`，每个服务的完整配置类型位于自己的 `application/<service>/internal/config`。

## 加载规则

1. 必须设置 `MS_K_APP_ENV=dev` 或 `MS_K_APP_ENV=prod`。
2. 服务只读取自身目录的 `conf.dev.yaml` 或 `conf.prod.yaml`，不会读取 `conf.example.yaml`。
3. 执行 `make init-config` 可从模板创建缺失的 dev/prod 文件，已有文件不会覆盖。
4. 非空的 `MS_K_*` 环境变量覆盖 YAML；空值不会覆盖 YAML。布尔值使用 `true`/`false`，整数使用十进制。
5. 配置没有隐式业务默认值。服务实际使用的键必须在 YAML 或环境变量中显式出现；允许空值的键也必须在 YAML 中声明。
6. `zaplogger.dev.yaml` 和 `zaplogger.prod.yaml` 只从文件加载，不接受字段级环境变量覆盖。

## 共享配置段

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `registry.driver` | `MS_K_REGISTRY_DRIVER` | `consul`、`etcd`、`nacos` 或测试用 `inprocess`。 |
| `registry.address` | `MS_K_REGISTRY_ADDRESS` | 注册中心地址；Nacos 可使用逗号分隔的地址列表。 |
| `registry.prefix` | `MS_K_REGISTRY_PREFIX` | etcd 注册键前缀。 |
| `registry.namespace` | `MS_K_REGISTRY_NAMESPACE` | Nacos 命名空间。 |
| `registry.group` | `MS_K_REGISTRY_GROUP` | Nacos 服务分组。 |
| `registry.username` / `registry.password` | `MS_K_REGISTRY_USERNAME` / `MS_K_REGISTRY_PASSWORD` | Nacos 认证信息。 |
| `server.port` | `MS_K_SERVER_PORT` | HTTP 监听端口，范围 1-65535。 |
| `grpc.port` | `MS_K_GRPC_PORT` | IAM、SYS、Resource、Realtime 的 gRPC 监听端口。 |
| `service.id` | `MS_K_SERVICE_ID` | 注册中心实例 ID；留空时按服务名和主机名生成。 |
| `service.advertiseHost` | `MS_K_SERVICE_ADVERTISE_HOST` | 可选；写入注册中心、供其他实例访问的地址。未配置时自动选择启用的非回环 IPv4；它不是监听地址。 |
| `database.dsn` | `MS_K_DATABASE_DSN` | PostgreSQL 连接串。 |
| `database.maxOpenConns` | `MS_K_DATABASE_MAX_OPEN_CONNS` | 最大打开连接数，必须大于 0。 |
| `database.maxIdleConns` | `MS_K_DATABASE_MAX_IDLE_CONNS` | 最大空闲连接数，不能超过最大打开连接数。 |
| `database.connMaxLifetimeMinutes` | `MS_K_DATABASE_CONN_MAX_LIFETIME_MINUTES` | 连接最大存活时间，单位为分钟。 |
| `database.slowThreshold` | `MS_K_DATABASE_SLOW_THRESHOLD` | 慢 SQL 阈值，单位为毫秒。 |
| `redis.addr` | `MS_K_REDIS_ADDR` | Redis 地址。 |
| `redis.password` | `MS_K_REDIS_PASSWORD` | Redis 密码；无密码时显式设置为空字符串。 |
| `redis.db` | `MS_K_REDIS_DB` | Redis 逻辑库编号，不能为负数。 |
| `auth.tokenHeader` | `MS_K_AUTH_TOKEN_HEADER` | 访问令牌请求头名称，通常为 `Authorization`。 |
| `cors.enabled` | `MS_K_CORS_ENABLED` | 是否启用跨域；生产环境建议关闭。 |

## Gateway

Gateway 是唯一对外 HTTP/WS 入口，不连接数据库，不拥有业务数据。

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `server.tlsCertFile` | `MS_K_SERVER_TLS_CERT_FILE` | TLS 证书路径；与私钥必须同时为空或同时非空。 |
| `server.tlsKeyFile` | `MS_K_SERVER_TLS_KEY_FILE` | TLS 私钥路径。 |
| `gateway.rateLimitPerMinute` | `MS_K_GATEWAY_RATE_LIMIT_PER_MINUTE` | 单来源 IP 每分钟请求上限，`0` 表示不限制。 |

Gateway 使用 `registry` 发现 IAM、SYS、Resource、Realtime，并根据服务实例发布的 HTTP 路由动态构建代理路由表。

## IAM

IAM 拥有用户、角色、API 权限、登录会话和 JWT 能力，使用 PostgreSQL 与 Redis。

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `jwt.secret` | `MS_K_JWT_SECRET` | JWT 密钥，至少 32 个字符。 |
| `jwt.expire` | `MS_K_JWT_EXPIRE` | Access Token 有效期，单位为秒，必须大于 0。 |
| `auth.allowConcurrent` | `MS_K_AUTH_ALLOW_CONCURRENT` | 是否允许同一用户保留多个并发会话。 |

其余 `server`、`grpc`、`registry`、`service`、`database`、`redis`、`auth.tokenHeader`、`cors` 使用共享配置段。

## SYS

SYS 拥有组织、菜单、字典、系统配置、登录日志和操作日志。它的私有 Worker 负责定期清理过期系统日志。

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `workers.logCleanup.enabled` | `MS_K_WORKERS_LOG_CLEANUP_ENABLED` | 是否启用日志清理。 |
| `workers.logCleanup.cron` | `MS_K_WORKERS_LOG_CLEANUP_CRON` | 六段式 Cron 表达式。 |
| `workers.logCleanup.retentionDays` | `MS_K_WORKERS_LOG_CLEANUP_RETENTION_DAYS` | 日志保留天数，必须大于 0。 |
| `workers.logCleanup.lockTTLMinutes` | `MS_K_WORKERS_LOG_CLEANUP_LOCK_TTL_MINUTES` | Redis 抢占锁有效期，必须短于 Cron 周期。 |

## Resource

Resource 拥有附件元数据和对象存储能力。它的私有 Worker 负责清理过期附件。

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `storage.endpoint` | `MS_K_STORAGE_ENDPOINT` | S3 兼容服务地址，不含协议。 |
| `storage.accessKey` | `MS_K_STORAGE_ACCESS_KEY` | 对象存储访问密钥标识。 |
| `storage.secretKey` | `MS_K_STORAGE_SECRET_KEY` | 对象存储访问密钥。 |
| `storage.region` | `MS_K_STORAGE_REGION` | 对象存储区域。 |
| `storage.bucket` | `MS_K_STORAGE_BUCKET` | 附件存储桶。 |
| `storage.useSSL` | `MS_K_STORAGE_USE_SSL` | 是否使用 TLS。 |
| `workers.expiredAttachmentCleanup.enabled` | `MS_K_WORKERS_EXPIRED_ATTACHMENT_CLEANUP_ENABLED` | 是否启用附件清理。 |
| `workers.expiredAttachmentCleanup.cron` | `MS_K_WORKERS_EXPIRED_ATTACHMENT_CLEANUP_CRON` | 六段式 Cron 表达式。 |
| `workers.expiredAttachmentCleanup.lockTTLMinutes` | `MS_K_WORKERS_EXPIRED_ATTACHMENT_CLEANUP_LOCK_TTL_MINUTES` | Redis 抢占锁有效期，必须短于 Cron 周期。 |

## Realtime

Realtime 只拥有 WebSocket 连接、心跳和跨实例投递，不连接数据库。握手通过 IAM gRPC 校验令牌，实例之间通过 Redis Pub/Sub 广播。

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `websocket.enabled` | `MS_K_WEBSOCKET_ENABLED` | 是否启用 WebSocket。 |
| `websocket.timeoutEnabled` | `MS_K_WEBSOCKET_TIMEOUT_ENABLED` | 是否启用读写超时。 |
| `websocket.readTimeoutSeconds` | `MS_K_WEBSOCKET_READ_TIMEOUT_SECONDS` | 读取超时秒数。 |
| `websocket.writeTimeoutSeconds` | `MS_K_WEBSOCKET_WRITE_TIMEOUT_SECONDS` | 写入超时秒数。 |
| `websocket.heartbeatEnabled` | `MS_K_WEBSOCKET_HEARTBEAT_ENABLED` | 是否启用心跳。 |
| `websocket.maxReadTimeouts` | `MS_K_WEBSOCKET_MAX_READ_TIMEOUTS` | 连续读取超时上限。 |

Realtime 还使用共享的 `server`、`grpc`、`registry`、`service`、`redis`、`auth` 和 `cors` 配置段。

## User Manager

`cmd/usermgr` 是一次性管理员维护工具，不是部署服务，也不读取任何服务配置。数据库连接通过可选的 `--database-dsn` 参数指定；未指定时使用工具内置的本地 PostgreSQL 默认连接。

```bash
export MS_K_USERMGR_PASSWORD='至少 8 个字符的强密码'
go run ./cmd/usermgr --operation=create --username=admin --nickname=管理员 --role=super_admin
go run ./cmd/usermgr --operation=reset --username=admin
# 覆盖数据库连接：--database-dsn='host=... user=... password=... dbname=... port=5432 sslmode=disable'
```

## OpenTelemetry 环境变量

以下变量由 OpenTelemetry SDK 读取，不属于 `MS_K_*` 配置：

| 环境变量 | 说明 |
| --- | --- |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | 通用 OTLP gRPC 地址。 |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Trace 专用 OTLP 地址。 |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | 当前使用 gRPC。 |
| `OTEL_EXPORTER_OTLP_HEADERS` | Collector 认证请求头。 |
| `OTEL_TRACES_SAMPLER` / `OTEL_TRACES_SAMPLER_ARG` | Trace 采样器及参数。 |
| `OTEL_TRACES_EXPORTER=none` | 强制关闭 Trace 导出。 |

未配置 OTLP 地址时仍生成可用于日志关联的 `trace_id` 和 `span_id`，但不会导出 Span。

## Docker 示例

```bash
make init-config
export MS_K_JWT_SECRET='replace-with-at-least-32-random-characters'
./scripts/microservices.sh start
```

该脚本构建并启动 Gateway 1 个、IAM 3 个、SYS 5 个、Resource 1 个、Realtime 2 个实例。停止但保留数据使用 `./scripts/microservices.sh stop`；删除容器和数据卷使用 `./scripts/microservices.sh destroy`。
