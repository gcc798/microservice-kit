package bootstrap

import (
	serviceconfig "github.com/gcc798/microservice-kit/application/iam/internal/config"
	iam "github.com/gcc798/microservice-kit/application/iam/internal/domain"
	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	"github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/modules"
	sharedtransport "github.com/gcc798/microservice-kit/internal/transport"
)

type domainDeps struct {
	security      iamv1.API
	grpcServer    *iam.GRPCServer
	auth          iam.AuthService
	captcha       iam.CaptchaService
	user          iam.UserService
	role          iam.RoleService
	apiPermission iam.ApiPermissionService
	org           iam.OrgService
	menu          *iam.MenuService
	sms           modules.SMS
	modules       []modules.Module
}

func newDomain(cfg *serviceconfig.Config, iamInfra *iamResources, pool *sharedtransport.ClientPool, log logger.Logger) domainDeps {
	sms := modules.NewSMSModule()
	email := modules.NewEmailModule()
	wechat := modules.NewWeChatModule()
	captchaModule := modules.NewCaptchaModule(sms, email)
	captchaService := iam.NewCaptchaService(captchaModule)
	tokens := iam.NewTokenManager(iamInfra.JWT, iamInfra.Redis, log)
	security := iam.NewAuthzAPI(tokens, iam.NewPermissionService(iamInfra.DB, log))
	auth := iam.NewAuthService(
		iamInfra.DB, iamInfra.Redis, cfg.Auth.AllowConcurrent, log,
		iam.NewClientService(iamInfra.DB, iamInfra.Redis, log), tokens,
		captchaService, wechat, sysv1.NewRemote(pool),
	)
	return domainDeps{
		security: security, grpcServer: iam.NewGRPCServer(security), auth: auth,
		captcha: captchaService, user: iam.NewUserService(iamInfra.DB, log),
		role: iam.NewRoleService(iamInfra.DB, log), apiPermission: iam.NewApiPermissionService(iamInfra.DB),
		org: iam.NewOrgService(iamInfra.DB, log), menu: iam.NewMenuService(iamInfra.DB), sms: sms,
		modules: []modules.Module{sms, email, wechat, captchaModule},
	}
}
