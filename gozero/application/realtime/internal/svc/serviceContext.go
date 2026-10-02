package svc

import (
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/gcc798/microservice-kit/application/realtime/internal/config"
	"github.com/gcc798/microservice-kit/application/realtime/internal/hub"
	"github.com/gcc798/microservice-kit/application/realtime/internal/relay"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
	Hub    *hub.Hub
	Relay  *relay.Relay
	IamRpc iamservice.IamService
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := redis.NewClient(&redis.Options{Addr: c.RelayRedis.Addr, Password: c.RelayRedis.Password, DB: c.RelayRedis.DB})
	connections := hub.New()
	return &ServiceContext{
		Config: c,
		Redis:  client,
		Hub:    connections,
		Relay:  relay.New(client, connections),
		IamRpc: iamservice.NewIamService(zrpc.MustNewClient(c.IamRpc)),
	}
}
