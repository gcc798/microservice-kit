// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"github.com/gcc798/microservice-kit/application/iam-api/internal/config"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	IamRpcClient iamservice.IamService
	Redis        *redis.Client
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       c.Redis.Db,
	})
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		_ = rdb.Close()
		return nil, err
	}

	return &ServiceContext{
		Config:       c,
		IamRpcClient: iamservice.NewIamService(zrpc.MustNewClient(c.IamRpc)),
		Redis:        rdb,
	}, nil
}
