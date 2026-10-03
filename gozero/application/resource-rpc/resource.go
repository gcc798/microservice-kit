package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/config"
	resourceserviceServer "github.com/gcc798/microservice-kit/application/resource-rpc/internal/server/resourceservice"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/workers"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"
	"github.com/gcc798/microservice-kit/internal/observability"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/resource.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	observability.Configure(&c.ServiceConf)
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := ctx.Close(); err != nil {
			fmt.Printf("close resource-rpc: %v\n", err)
		}
	}()
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	cleanup, err := workers.NewExpiredAttachmentCleanup(workerCtx, c.Workers.ExpiredAttachmentCleanup, ctx)
	if err != nil {
		panic(err)
	}
	cleanup.Start()
	defer func() {
		cancelWorkers()
		cleanup.Stop()
	}()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterResourceServiceServer(grpcServer, resourceserviceServer.NewResourceServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
