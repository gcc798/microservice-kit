# Native Proto/gRPC 开发约定

Proto 是跨进程契约；业务代码通过手写适配层和最小 API 接口使用 gRPC，不直接保存连接或硬编码服务地址。

## 目录和职责

```text
api/<service>/v1/*.proto                 手写跨进程契约
internal/api/<service>/v1/*.pb.go        protoc-gen-go 生成代码
internal/api/<service>/v1/*_grpc.pb.go   gRPC 客户端、服务端和注册描述
internal/api/<service>/v1/contracts.go   手写 Remote 和最小 API 适配层
application/<service>/internal/domain/rpc.go  领域服务到 gRPC 的适配
internal/transport                         服务端生命周期和动态 Client Pool
```

当前跨进程契约：

| Proto | 服务 | 服务端实现 |
| --- | --- | --- |
| `api/iam/v1/iam.proto` | IAM | `application/iam/internal/domain/rpc.go` |
| `api/sys/v1/sys.proto` | SYS | `application/sys/internal/domain/rpc.go` |
| `api/resource/v1/resource.proto` | Resource | `application/resource/internal/domain/rpc.go` |
| `api/realtime/v1/realtime.proto` | Realtime | `application/realtime/internal/domain/rpc.go` |

## 编写规则

1. 使用 `syntax = "proto3"`，包名为 `microservice_kit.<service>.v1`。
2. `go_package` 指向 `github.com/gcc798/microservice-kit/internal/api/<service>/v1`，并使用稳定别名。
3. `service`、`rpc`、`message` 和字段都要紧邻中文注释；单位、状态值和约束一并说明。
4. 已发布字段编号不能复用或改变语义；删除字段使用 `reserved`。
5. 破坏性契约变更新建 `v2` 包，不在旧包中改语义。
6. Proto 不放数据库模型、HTTP DTO 或实现细节。
7. 不手工修改 `*.pb.go` 和 `*_grpc.pb.go`。

## 生成代码

在 `native` 目录执行：

```bash
make proto
make proto-check
```

生成器使用本机的 `protoc`、`protoc-gen-go` 和 `protoc-gen-go-grpc`。新增 Proto 文件时必须同步更新 `native/Makefile` 的生成目标。

## 服务端注册

IAM、SYS、Resource、Realtime 都由各自组合根创建注册中心管理的 gRPC 服务：

```go
grpcServer, err := transport.StartRegisteredGRPC(ctx, registry, transport.RegisteredGRPCOptions{
	GRPCPort: cfg.GRPC.Port,
	HTTPPort: cfg.Server.Port,
	ServiceID: cfg.Service.ID,
	ServiceName: serviceName,
	AdvertiseHost: cfg.Service.AdvertiseHost,
	Routes: routes,
}, func(server *grpc.Server) {
	// 在此注册当前服务的生成 gRPC 实现。
})
```

`StartRegisteredGRPC` 只接受 gRPC 端口、HTTP 端口、实例信息和路由，不依赖共享全量配置。退出时 `RegisteredGRPC.Stop` 先注销实例，再优雅停止服务端。

## 客户端调用

```go
pool := transport.NewClientPool(registry)
defer pool.Close()
systemAPI := sysv1.NewRemote(pool)
if err := systemAPI.RecordLogin(ctx, request); err != nil {
	return err
}
```

`Remote` 通过 `ClientPool.Conn` 查询健康实例并复用连接。Gateway 使用 IAM Remote 校验令牌；SYS 和 Resource 使用 IAM Remote 做权限检查；SYS、Resource 通过 Realtime gRPC 发布实时消息。业务代码不直接依赖 Redis Pub/Sub。

## 修改闭环

1. 修改正确服务的 Proto，并补齐新增声明的中文注释。
2. 执行 `make proto`，不要手工编辑生成文件。
3. 更新 `contracts.go` 和目标服务 `internal/domain/rpc.go`。
4. 执行 `make proto-check`、`go test ./...`、`make verify` 和 `git diff --check`。
