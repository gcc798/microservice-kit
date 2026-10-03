// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/config"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/handler"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/common/middleware"
	"github.com/gcc798/microservice-kit/common/requestmeta"
	"github.com/gcc798/microservice-kit/internal/observability"
	registry "github.com/gcc798/microservice-kit/internal/registry"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/iam-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	observability.Configure(&c.ServiceConf)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(middleware.PanicRecoveryMiddleware)
	server.Use(middleware.HTTPStatus)
	server.Use(requestmeta.Middleware)
	server.Use(middleware.StringIDConverter)
	server.Use(middleware.NewJWTAuthMiddleware(middleware.JWTAuthConfig{
		Secret:      c.Jwt.Secret,
		TokenHeader: c.Auth.TokenHeader,
		WhiteList: []string{
			"/login",
			"/auth/refresh",
			"/captcha/*",
			"/resource/sms/code",
			"/health",
			"/health/ready",
			"/health/live",
			"/health/startup",
		},
	}).Handle)

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}
	defer ctx.Redis.Close()
	handler.RegisterHandlers(server, ctx)
	publisher, err := registry.PublishHTTP(c.HTTPRegistry, c.Name, c.Port, server.Routes())
	if err != nil {
		panic(err)
	}
	defer publisher.Stop()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
