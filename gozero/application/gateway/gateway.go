package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gcc798/microservice-kit/application/gateway/internal/config"
	gatewayserver "github.com/gcc798/microservice-kit/application/gateway/internal/server"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
	"github.com/gcc798/microservice-kit/common/middleware"
	"github.com/gcc798/microservice-kit/internal/observability"
	registry "github.com/gcc798/microservice-kit/internal/registry"
	"github.com/zeromicro/go-zero/core/conf"
	resthandler "github.com/zeromicro/go-zero/rest/handler"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	observability.Configure(&c.ServiceConf)
	c.ServiceConf.MustSetUp()

	iam := iamservice.NewIamService(zrpc.MustNewClient(c.IamRpc))
	system := sysservice.NewSysService(zrpc.MustNewClient(c.SysRpc))
	proxy, err := gatewayserver.New(c.HTTPRegistry, iam, system, c.Auth.TokenHeader)
	if err != nil {
		panic(err)
	}
	defer proxy.Close()
	publisher, err := registry.PublishService(c.ServiceRegistry, c.Name, c.Port)
	if err != nil {
		panic(err)
	}
	defer publisher.Stop()

	baseHandler := resthandler.TraceHandler(c.Name, "gateway", resthandler.WithTraceIgnorePaths([]string{"/health", "/health/live", "/health/ready", "/health/startup", "/metrics"}))(proxy.Handler())
	baseHandler = resthandler.PrometheusHandler("gateway", "*")(baseHandler)
	handler := middleware.PanicRecoveryMiddleware(baseHandler.ServeHTTP)
	handler = middleware.CORS(middleware.CORSConfig{})(handler)

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", c.Host, c.Port),
		Handler:           http.HandlerFunc(handler),
		ReadHeaderTimeout: 10 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
