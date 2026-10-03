package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/netx"
	"github.com/zeromicro/go-zero/rest"
)

type (
	HTTPConfig struct {
		Etcd          discov.EtcdConf
		AdvertiseHost string `json:",optional"`
	}

	HTTPRoute struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}

	HTTPInstance struct {
		Service  string      `json:"service"`
		Endpoint string      `json:"endpoint"`
		Routes   []HTTPRoute `json:"routes"`
	}

	ServiceInstance struct {
		Service  string `json:"service"`
		Endpoint string `json:"endpoint"`
	}
)

func PublishHTTP(config HTTPConfig, service string, port int, routes []rest.Route) (*discov.Publisher, error) {
	return PublishRoutes(config, service, port, Routes(routes))
}

func PublishRoutes(config HTTPConfig, service string, port int, routes []HTTPRoute) (*discov.Publisher, error) {
	host := advertiseHost(config.AdvertiseHost)
	if host == "" {
		return nil, errors.New("HTTP advertise host is empty")
	}
	return publish(config, HTTPInstance{
		Service:  service,
		Endpoint: "http://" + net.JoinHostPort(host, strconv.Itoa(port)),
		Routes:   routes,
	})
}

func PublishService(config HTTPConfig, service string, port int) (*discov.Publisher, error) {
	host := advertiseHost(config.AdvertiseHost)
	if host == "" || service == "" {
		return nil, errors.New("service name and advertise host are required")
	}
	return publishValue(config, ServiceInstance{Service: service, Endpoint: "http://" + net.JoinHostPort(host, strconv.Itoa(port))})
}

func Routes(routes []rest.Route) []HTTPRoute {
	registered := make([]HTTPRoute, 0, len(routes))
	for _, route := range routes {
		if route.Method == "" || route.Path == "" || isInternalRoute(route.Path) {
			continue
		}
		registered = append(registered, HTTPRoute{Method: route.Method, Path: route.Path})
	}
	return registered
}

func Subscribe(config HTTPConfig) (*discov.Subscriber, error) {
	if err := config.Etcd.Validate(); err != nil {
		return nil, err
	}
	return discov.NewSubscriber(config.Etcd.Hosts, config.Etcd.Key, subscriberOptions(config.Etcd)...)
}

func Decode(values []string) ([]HTTPInstance, error) {
	instances := make([]HTTPInstance, 0, len(values))
	for _, value := range values {
		var instance HTTPInstance
		if err := json.Unmarshal([]byte(value), &instance); err != nil {
			return nil, fmt.Errorf("decode HTTP service instance: %w", err)
		}
		if err := validate(instance); err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func publish(config HTTPConfig, instance HTTPInstance) (*discov.Publisher, error) {
	if err := validate(instance); err != nil {
		return nil, err
	}
	return publishValue(config, instance)
}

func publishValue(config HTTPConfig, instance any) (*discov.Publisher, error) {
	if err := config.Etcd.Validate(); err != nil {
		return nil, err
	}
	value, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}
	publisher := discov.NewPublisher(config.Etcd.Hosts, config.Etcd.Key, string(value), publisherOptions(config.Etcd)...)
	if err := publisher.KeepAlive(); err != nil {
		return nil, err
	}
	return publisher, nil
}

func publisherOptions(config discov.EtcdConf) []discov.PubOption {
	var options []discov.PubOption
	if config.HasID() {
		options = append(options, discov.WithId(config.ID))
	}
	if config.HasAccount() {
		options = append(options, discov.WithPubEtcdAccount(config.User, config.Pass))
	}
	if config.HasTLS() {
		options = append(options, discov.WithPubEtcdTLS(config.CertFile, config.CertKeyFile, config.CACertFile, config.InsecureSkipVerify))
	}
	return options
}

func subscriberOptions(config discov.EtcdConf) []discov.SubOption {
	var options []discov.SubOption
	if config.HasAccount() {
		options = append(options, discov.WithSubEtcdAccount(config.User, config.Pass))
	}
	if config.HasTLS() {
		options = append(options, discov.WithSubEtcdTLS(config.CertFile, config.CertKeyFile, config.CACertFile, config.InsecureSkipVerify))
	}
	return options
}

func advertiseHost(configured string) string {
	if configured != "" && configured != "0.0.0.0" && configured != "::" {
		return configured
	}
	if podIP := strings.TrimSpace(os.Getenv("POD_IP")); podIP != "" {
		return podIP
	}
	return netx.InternalIp()
}

func isInternalRoute(path string) bool {
	return path == "/metrics" || path == "/health" || strings.HasPrefix(path, "/health/")
}

func validate(instance HTTPInstance) error {
	if instance.Service == "" || len(instance.Routes) == 0 {
		return errors.New("HTTP service instance requires a service and routes")
	}
	endpoint, err := url.Parse(instance.Endpoint)
	if err != nil || endpoint.Hostname() == "" || endpoint.Path != "" || endpoint.RawQuery != "" || endpoint.Fragment != "" || endpoint.User != nil || endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return fmt.Errorf("invalid HTTP service endpoint %q", instance.Endpoint)
	}
	for _, route := range instance.Routes {
		if route.Method == "" || !strings.HasPrefix(route.Path, "/") {
			return errors.New("HTTP service route requires a method and path")
		}
	}
	return nil
}
