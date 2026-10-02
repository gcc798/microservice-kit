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
	"github.com/gcc798/microservice-kit/common/middleware"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	iam := iamservice.NewIamService(zrpc.MustNewClient(c.IamRpc))
	proxy, err := gatewayserver.New(gatewayserver.Targets{
		IAM: c.Backends.IAM, SYS: c.Backends.SYS, Resource: c.Backends.Resource, Realtime: c.Backends.Realtime,
	}, iam, c.Auth.TokenHeader)
	if err != nil {
		panic(err)
	}

	handler := middleware.PanicRecoveryMiddleware(proxy.Handler().ServeHTTP)
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
