package main

import (
	"flag"
	"fmt"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/config"
	iamserviceServer "github.com/gcc798/microservice-kit/application/iam-rpc/internal/server/iamservice"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "application/iam-rpc/etc/iam-rpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := ctx.Close(); err != nil {
			fmt.Printf("close iam-rpc: %v\n", err)
		}
	}()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterIamServiceServer(grpcServer, iamserviceServer.NewIamServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
