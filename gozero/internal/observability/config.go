package observability

import (
	"os"
	"strings"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel/attribute"
)

func Configure(config *service.ServiceConf) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
	endpoint = strings.TrimSuffix(endpoint, "/v1/traces")
	if endpoint != "" {
		config.Telemetry.Endpoint = endpoint
		config.Telemetry.Batcher = "otlpgrpc"
	}
	if strings.EqualFold(os.Getenv("OTEL_TRACES_EXPORTER"), "none") {
		config.Telemetry.Disabled = true
	}
	hostname, _ := os.Hostname()
	trace.AddResources(
		attribute.String("service.instance.id", config.Name+"-"+hostname),
		attribute.String("deployment.environment.name", config.Mode),
	)
}
