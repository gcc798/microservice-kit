package transport

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/gcc798/microservice-kit/internal/registry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

// GRPCServer 封装 gRPC 服务端及其监听器。
type GRPCServer struct {
	server   *grpc.Server // gRPC 服务端。
	listener net.Listener // 网络监听器。
}

// NewGRPCServer 创建 gRPC 服务端并监听指定地址。
func NewGRPCServer(address string) (*GRPCServer, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen for grpc: %w", err)
	}
	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	return &GRPCServer{server: server, listener: listener}, nil
}

// Server 返回底层 gRPC 服务端，用于注册服务实现。
func (s *GRPCServer) Server() *grpc.Server { return s.server }

// Address 返回监听器实际绑定的地址。
func (s *GRPCServer) Address() string { return s.listener.Addr().String() }

// Serve 阻塞运行 gRPC 服务端。
func (s *GRPCServer) Serve() error { return s.server.Serve(s.listener) }

// GracefulStop 优雅停止 gRPC 服务端。
func (s *GRPCServer) GracefulStop() { s.server.GracefulStop() }

// ClientPool 根据服务发现结果复用 gRPC 客户端连接。
type ClientPool struct {
	registry registry.Registry           // 服务注册发现客户端。
	selector *registry.Selector          // 实例选择器。
	mu       sync.Mutex                  // 连接缓存互斥锁。
	conns    map[string]*grpc.ClientConn // 按地址缓存的连接。
}

// NewClientPool 创建 gRPC 客户端连接池。
func NewClientPool(reg registry.Registry) *ClientPool {
	return &ClientPool{registry: reg, selector: registry.NewSelector(), conns: make(map[string]*grpc.ClientConn)}
}

// Conn 解析服务实例并返回可复用的 gRPC 连接。
func (p *ClientPool) Conn(ctx context.Context, service string) (*grpc.ClientConn, error) {
	instances, err := p.registry.Resolve(ctx, service)
	if err != nil {
		return nil, err
	}
	instance, err := p.selector.Pick(service, instances)
	if err != nil {
		return nil, err
	}
	address := instance.Endpoints[registry.EndpointGRPC]
	if address == "" {
		return nil, fmt.Errorf("service %q instance %q has no grpc endpoint", service, instance.ID)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if conn := p.conns[address]; conn != nil {
		return conn, nil
	}
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to %s at %s: %w", service, address, err)
	}
	p.conns[address] = conn
	return conn, nil
}

// Close 关闭连接池中的全部 gRPC 连接。
func (p *ClientPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var first error
	for address, conn := range p.conns {
		if err := conn.Close(); err != nil && first == nil {
			first = err
		}
		delete(p.conns, address)
	}
	return first
}
