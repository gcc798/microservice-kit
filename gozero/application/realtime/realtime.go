package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gcc798/microservice-kit/application/realtime/internal/config"
	httpserver "github.com/gcc798/microservice-kit/application/realtime/internal/server"
	realtimeserviceServer "github.com/gcc798/microservice-kit/application/realtime/internal/server/realtimeservice"
	"github.com/gcc798/microservice-kit/application/realtime/internal/svc"
	"github.com/gcc798/microservice-kit/application/realtime/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/realtime.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterRealtimeServiceServer(grpcServer, realtimeserviceServer.NewRealtimeServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	runtime := newRuntime(ctx)
	group := service.NewServiceGroup()
	group.Add(s)
	group.Add(runtime)
	defer group.Stop()

	fmt.Printf("Starting realtime rpc at %s and websocket at %s...\n", c.ListenOn, runtime.http.Addr)
	group.Start()
}

type runtimeService struct {
	ctx    context.Context
	cancel context.CancelFunc
	svc    *svc.ServiceContext
	http   *http.Server
}

func newRuntime(ctx *svc.ServiceContext) *runtimeService {
	runCtx, cancel := context.WithCancel(context.Background())
	return &runtimeService{ctx: runCtx, cancel: cancel, svc: ctx, http: httpserver.NewHTTP(ctx)}
}

func (r *runtimeService) Start() {
	go r.svc.Relay.Run(r.ctx)
	if err := r.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func (r *runtimeService) Stop() {
	r.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = r.http.Shutdown(ctx)
	r.svc.Hub.Close()
	_ = r.svc.Redis.Close()
}
