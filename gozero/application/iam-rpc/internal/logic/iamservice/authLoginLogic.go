package iamservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLoginLogic {
	return &AuthLoginLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AuthLoginLogic) AuthLogin(in *pb.AuthLoginReq) (*pb.AuthLoginResp, error) {
	client, err := authenticateClient(l.ctx, l.svcCtx, in.ClientKey, in.GrantType)
	if err != nil {
		l.recordLogin(in, loginAccount(in), in.ClientKey, 1, err.Error())
		return nil, err
	}
	var user *userAuthRow
	switch in.GrantType {
	case "password":
		if err := verifyImageCaptcha(l.ctx, l.svcCtx, in.Uuid, in.Code); err != nil {
			return nil, err
		}
		user, err = authenticatePassword(l.ctx, l.svcCtx, in.Username, in.Password)
	case "email":
		user, err = authenticateEmail(l.ctx, l.svcCtx, in.Email, in.Uuid, in.Code)
	case "sms":
		user, err = authenticateSms(l.ctx, l.svcCtx, in.Phonenumber, in.Uuid, in.Code)
	case "xcx":
		user, err = authenticateXcx(l.ctx, l.svcCtx, in.Phonenumber, in.Code, in.WxCode)
	case "wechat":
		user, err = authenticateWechat(l.ctx, l.svcCtx, in.WxCode)
	default:
		return nil, errUnsupportedGrantType()
	}
	if err != nil {
		l.recordLogin(in, loginAccount(in), client.ClientId, 1, err.Error())
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SUser{}).Where("id = ?", user.Id).Updates(map[string]any{
		"login_ip": in.LoginIp, "login_date": time.Now().Unix(), "updated_time": time.Now(),
	}).Error; err != nil {
		l.Errorf("更新登录信息失败: %v", err)
	}
	l.recordLogin(in, user.UserName, client.ClientId, 0, "登录成功")
	return &pb.AuthLoginResp{
		ClientId:      client.ClientId,
		ClientKey:     client.ClientKey,
		DeviceType:    nullString(client.DeviceType),
		Timeout:       client.Timeout,
		ActiveTimeout: client.ActiveTimeout,
		UserInfo: &pb.UserInfo{
			UserId:      user.Id,
			Username:    user.UserName,
			Nickname:    nullString(user.NickName),
			Phonenumber: nullString(user.Phonenumber),
			Email:       nullString(user.Email),
			Avatar:      nullString(user.Avatar),
			UserType:    int32(user.UserType),
			OrgId:       nullInt64(user.OrgId),
			OpenId:      nullString(user.OpenId),
			UnionId:     nullString(user.UnionId),
		},
	}, nil
}

func (l *AuthLoginLogic) recordLogin(in *pb.AuthLoginReq, username, clientID string, status int32, message string) {
	browser, osName := parseUserAgent(in.UserAgent)
	if _, err := l.svcCtx.SysRpcClient.LoginLogCreate(l.ctx, &sysservice.LoginLogReq{
		UserName: username, Ipaddr: in.LoginIp, Browser: browser, Os: osName,
		Status: status, Msg: message, ClientId: clientID,
	}); err != nil {
		l.Errorf("记录登录日志失败: %v", err)
	}
}

func loginAccount(in *pb.AuthLoginReq) string {
	for _, account := range []string{in.Username, in.Email, in.Phonenumber} {
		if account != "" {
			return account
		}
	}
	return "wechat"
}

func parseUserAgent(userAgent string) (browser, operatingSystem string) {
	switch {
	case strings.Contains(userAgent, "Edg/"):
		browser = "Edge"
	case strings.Contains(userAgent, "Chrome/"):
		browser = "Chrome"
	case strings.Contains(userAgent, "Safari/"):
		browser = "Safari"
	case strings.Contains(userAgent, "Firefox/"):
		browser = "Firefox"
	default:
		browser = "Unknown"
	}
	switch {
	case strings.Contains(userAgent, "Windows"):
		operatingSystem = "Windows"
	case strings.Contains(userAgent, "Mac OS X"):
		operatingSystem = "macOS"
	case strings.Contains(userAgent, "Android"):
		operatingSystem = "Android"
	case strings.Contains(userAgent, "iPhone"), strings.Contains(userAgent, "iPad"):
		operatingSystem = "iOS"
	case strings.Contains(userAgent, "Linux"):
		operatingSystem = "Linux"
	default:
		operatingSystem = "Unknown"
	}
	return browser, operatingSystem
}

func errUnsupportedGrantType() error {
	return errors.New("不支持的授权类型，支持的类型: password, email, sms, xcx, wechat")
}
