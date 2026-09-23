package transport

import (
	"context"
	"errors"
	"strconv"

	"github.com/gcc798/microservice-kit/internal/registry"
	"google.golang.org/grpc"
)

// RegisteredGRPC 表示已启动并完成注册的 gRPC 服务。
type RegisteredGRPC struct {
	Server   *GRPCServer              // gRPC 服务端。
	Registry registry.Registry        // 服务注册中心。
	Instance registry.ServiceInstance // 当前服务实例信息。
}

// RegisteredGRPCOptions 描述 gRPC 服务启动和注册所需的最小参数。
type RegisteredGRPCOptions struct {
	GRPCPort      int                  // gRPC 监听端口。
	HTTPPort      int                  // HTTP 入口端口。
	ServiceID     string               // 服务实例唯一标识。
	ServiceName   string               // 服务名称。
	AdvertiseHost string               // 对外通告的主机名或地址。
	Routes        []registry.HTTPRoute // 网关需要代理的 HTTP 路由。
}

// RegisterService 注册一个带端点信息的服务实例。
func RegisterService(ctx context.Context, reg registry.Registry, name, id string, endpoints map[string]string) (registry.ServiceInstance, error) {
	if id == "" {
		return registry.ServiceInstance{}, errors.New("service instance ID is required")
	}
	if endpoints == nil {
		endpoints = map[string]string{}
	}
	instance := registry.ServiceInstance{ID: id, Name: name, Endpoints: endpoints}
	return instance, reg.Register(ctx, instance)
}

// StartRegisteredGRPC 创建、启动并注册一个 gRPC 服务。
func StartRegisteredGRPC(ctx context.Context, reg registry.Registry, opts RegisteredGRPCOptions, register func(*grpc.Server)) (*RegisteredGRPC, error) {
	if opts.ServiceID == "" {
		return nil, errors.New("service instance ID is required")
	}
	server, err := NewGRPCServer(":" + strconv.Itoa(opts.GRPCPort))
	if err != nil {
		return nil, err
	}
	register(server.Server())
	go func() { _ = server.Serve() }()
	instance := registry.ServiceInstance{
		ID: opts.ServiceID, Name: opts.ServiceName, Routes: opts.Routes,
		Endpoints: map[string]string{
			registry.EndpointHTTP: "http://" + opts.AdvertiseHost + ":" + strconv.Itoa(opts.HTTPPort),
			registry.EndpointGRPC: opts.AdvertiseHost + ":" + strconv.Itoa(opts.GRPCPort),
		},
	}
	if err := reg.Register(ctx, instance); err != nil {
		server.GracefulStop()
		return nil, err
	}
	return &RegisteredGRPC{Server: server, Registry: reg, Instance: instance}, nil
}

// Stop 注销服务实例并优雅停止 gRPC 服务端。
func (s *RegisteredGRPC) Stop(ctx context.Context) error {
	err := s.Registry.Deregister(ctx, s.Instance)
	s.Server.GracefulStop()
	return err
}
