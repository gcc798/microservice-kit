package svc

import (
	"errors"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/config"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/migrations"
	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
	"github.com/gcc798/microservice-kit/internal/database"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type SMSProvider interface {
	SendSMS(phone, code string) error
}

type EmailProvider interface {
	SendEmail(email, code string) error
}

type consoleSMSProvider struct{}

func (p *consoleSMSProvider) SendSMS(phone, code string) error {
	logx.Infof("[验证码] 短信验证码发送至 %s: %s", phone, code)
	return nil
}

type consoleEmailProvider struct{}

func (p *consoleEmailProvider) SendEmail(email, code string) error {
	logx.Infof("[验证码] 邮箱验证码发送至 %s: %s", email, code)
	return nil
}

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	Redis         *redis.Client
	SMSProvider   SMSProvider
	EmailProvider EmailProvider
	SysRpcClient  sysservice.SysService
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	db, err := database.Open(database.Config{
		DSN:                    c.Postgres.Dsn,
		MaxIdleConns:           c.Postgres.MaxIdleConns,
		MaxOpenConns:           c.Postgres.MaxOpenConns,
		ConnMaxLifetimeMinutes: c.Postgres.ConnMaxLifetimeMinutes,
	})
	if err != nil {
		return nil, err
	}
	pool, err := db.DB()
	if err != nil {
		_ = database.Close(db)
		return nil, err
	}
	if err := migrations.Up(pool); err != nil {
		_ = database.Close(db)
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.CacheRedis.Addr,
		Password: c.CacheRedis.Password,
		DB:       c.CacheRedis.Db,
	})

	return &ServiceContext{
		Config:        c,
		DB:            db,
		Redis:         rdb,
		SMSProvider:   &consoleSMSProvider{},
		EmailProvider: &consoleEmailProvider{},
		SysRpcClient:  sysservice.NewSysService(zrpc.MustNewClient(c.SysRpc)),
	}, nil
}

func (s *ServiceContext) Close() error {
	return errors.Join(s.Redis.Close(), database.Close(s.DB))
}
