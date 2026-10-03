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
	"github.com/gcc798/microservice-kit/internal/observability"
	registry "github.com/gcc798/microservice-kit/internal/registry"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/discov"
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
	observability.Configure(&c.ServiceConf)
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterRealtimeServiceServer(grpcServer, realtimeserviceServer.NewRealtimeServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	runtime, err := newRuntime(ctx)
	if err != nil {
		panic(err)
	}
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
	pub    *discov.Publisher
}

func newRuntime(ctx *svc.ServiceContext) (*runtimeService, error) {
	runCtx, cancel := context.WithCancel(context.Background())
	pub, err := registry.PublishRoutes(ctx.Config.HTTPRegistry, "realtime", ctx.Config.HTTP.Port, []registry.HTTPRoute{{Method: http.MethodGet, Path: httpserver.WebSocketPath}})
	if err != nil {
		cancel()
		return nil, err
	}
	return &runtimeService{ctx: runCtx, cancel: cancel, svc: ctx, http: httpserver.NewHTTP(ctx), pub: pub}, nil
}

func (r *runtimeService) Start() {
	go r.svc.Relay.Run(r.ctx)
	if err := r.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func (r *runtimeService) Stop() {
	r.pub.Stop()
	r.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = r.http.Shutdown(ctx)
	r.svc.Hub.Close()
	_ = r.svc.Redis.Close()
}
