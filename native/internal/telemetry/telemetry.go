package telemetry

import (
	"context"
	"errors"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

// Init 安装进程级链路追踪。
// 未配置 OTLP 端点时仍会生成可用于日志关联的追踪 ID，但不会导出链路数据。
func Init(ctx context.Context, serviceName, instanceID, environment string) (func(context.Context) error, error) {
	if strings.TrimSpace(serviceName) == "" {
		return nil, errors.New("telemetry service name is required")
	}
	if strings.TrimSpace(instanceID) == "" {
		return nil, errors.New("telemetry service instance ID is required")
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceInstanceID(instanceID),
		semconv.DeploymentEnvironmentName(environment),
	)
	options := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if exporterConfigured() {
		exporter, err := otlptracegrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		options = append(options, sdktrace.WithBatcher(exporter))
	}

	provider := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return provider.Shutdown, nil
}

// exporterConfigured 判断是否配置了 OTLP 链路导出端点。
func exporterConfigured() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_TRACES_EXPORTER")), "none") {
		return false
	}
	return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" || os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""
}

// TraceHTTPPath 判断 HTTP 路径是否应创建追踪跨度。
// 高频基础设施探针不创建跨度，以减少无意义的追踪数据。
func TraceHTTPPath(path string) bool {
	switch path {
	case "/health", "/health/live", "/health/ready", "/health/startup", "/metrics":
		return false
	default:
		return true
	}
}
