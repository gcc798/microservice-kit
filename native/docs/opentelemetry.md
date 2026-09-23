# Native OpenTelemetry 链路追踪

Native 在组合根初始化 OpenTelemetry。未配置 OTLP 接收端时仍生成 `trace_id`、`span_id` 并写入日志，但不导出 Span，也不影响服务启动。

## 自动插桩边界

- Gateway HTTP 入口和反向代理：HTTP Server/Client Span。
- IAM、SYS、Resource、Realtime HTTP 入口：Echo Server Span。
- gRPC：共享传输层为客户端和服务端传播 Trace Context。
- PostgreSQL：IAM、SYS、Resource 的组合根通过共享数据库连接注册 SQL Trace。
- Redis：各服务组合根注册 go-redis Trace Hook；业务调用必须继续传递入口 `ctx`。
- SYS、Resource Worker：使用所属服务生命周期 Context。
- `/health/*` 和 `/metrics` 不创建 HTTP Span，避免探针噪声。

业务 controller、domain、model 不应重复为 HTTP、gRPC、PostgreSQL、Redis 创建基础设施 Span，也不应使用 `context.Background()` 替换入口传入的上下文。

## 业务 Span

只有框架 Span 无法表达且确实有排障价值的业务阶段才手动创建 Span。当前登录流程在 `application/iam/internal/domain/auth.go` 中使用 `auth.login`：

```go
ctx, span := otel.Tracer("github.com/gcc798/microservice-kit/application/iam").Start(ctx, "auth.login")
defer span.End()
```

业务错误同时写入 Span：

```go
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
```

Span 属性只能记录排障需要的非敏感信息，禁止记录密码、Token、验证码、Cookie、完整请求体和 SQL 参数。

## 日志关联

```go
logging.WithContext(ctx, log).Info("login succeeded", zap.Int64("user_id", userID))
```

`logger.WithContext` 会把当前有效 Span 的 Trace ID 和 Span ID 附加到日志字段：

```json
{"msg":"login succeeded","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","span_id":"00f067aa0ba902b7"}
```

## OTLP 导出

不部署 Collector 时无需设置 `OTEL_*` 变量。需要导出时至少设置：

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```

Native 只有检测到 `OTEL_EXPORTER_OTLP_ENDPOINT` 或 `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` 时才创建 OTLP exporter；进程退出时由各入口执行 Provider shutdown。`OTEL_TRACES_EXPORTER=none` 可强制关闭导出。

服务名、实例 ID 和环境由服务配置写入 OTel Resource，不重复使用 `OTEL_SERVICE_NAME`。
