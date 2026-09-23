package registry

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const (
	EndpointHTTP = "http"
	EndpointGRPC = "grpc"
)

// ServiceInstance 描述一个可被发现的服务实例。
type ServiceInstance struct {
	ID        string            `json:"id"`               // 实例唯一标识。
	Name      string            `json:"name"`             // 服务名称。
	Endpoints map[string]string `json:"endpoints"`        // 协议到访问地址的映射。
	Routes    []HTTPRoute       `json:"routes,omitempty"` // 对外暴露的 HTTP 路由。
}

// HTTPRoute 描述网关可代理的 HTTP 路由。
type HTTPRoute struct {
	Method string `json:"m"` // HTTP 方法。
	Path   string `json:"p"` // 路径模式。
}

// Options 描述服务注册中心的连接配置。
type Options struct {
	Driver    string // 注册中心类型。
	Address   string // 注册中心地址。
	Prefix    string // 键前缀。
	Namespace string // 命名空间。
	Group     string // 服务分组。
	Username  string // 认证用户名。
	Password  string // 认证密码。
}

// Watcher 持续获取指定服务的实例变化。
type Watcher interface {
	// Next 返回下一次服务实例快照。
	Next(context.Context) ([]ServiceInstance, error)
	// Stop 停止监听。
	Stop()
}

// Registry 定义服务注册、注销、发现和监听能力。
type Registry interface {
	// Register 注册服务实例。
	Register(context.Context, ServiceInstance) error
	// Deregister 注销服务实例。
	Deregister(context.Context, ServiceInstance) error
	// Resolve 获取当前可用的服务实例。
	Resolve(context.Context, string) ([]ServiceInstance, error)
	// Watch 监听服务实例变化。
	Watch(context.Context, string) (Watcher, error)
	// Close 关闭注册中心客户端。
	Close() error
}

// New 按配置创建服务注册中心客户端。
func New(options Options) (Registry, error) {
	switch options.Driver {
	case "consul":
		return NewConsul(options.Address), nil
	case "etcd":
		return NewEtcd(options.Address, options.Prefix), nil
	case "nacos":
		return NewNacos(options)
	case "inprocess", "":
		// 测试场景使用进程内注册中心，确保每个进程相互隔离。
		return NewInProcess(), nil
	default:
		return nil, fmt.Errorf("unsupported registry driver %q", options.Driver)
	}
}

// Selector 按轮询策略选择服务实例。
type Selector struct {
	mu   sync.Mutex
	next map[string]uint64
}

// NewSelector 创建服务实例选择器。
func NewSelector() *Selector { return &Selector{next: make(map[string]uint64)} }

// Pick 从健康实例列表中选择下一个实例。
func (s *Selector) Pick(name string, instances []ServiceInstance) (ServiceInstance, error) {
	if len(instances) == 0 {
		return ServiceInstance{}, fmt.Errorf("service %q has no healthy instances", name)
	}
	s.mu.Lock()
	index := s.next[name] % uint64(len(instances))
	s.next[name]++
	s.mu.Unlock()
	return instances[index], nil
}

var ErrWatcherStopped = errors.New("registry watcher stopped")
