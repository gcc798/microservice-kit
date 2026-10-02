package iamservicelogic

import (
	"context"
	"errors"

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
		return nil, err
	}
	if _, err := l.svcCtx.SysRpcClient.LoginLogCreate(l.ctx, &sysservice.LoginLogReq{
		UserName: user.UserName,
		Status:   0,
		Msg:      "登录成功",
		ClientId: client.ClientId,
	}); err != nil {
		l.Errorf("记录登录日志失败: %v", err)
	}
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
		},
	}, nil
}

func errUnsupportedGrantType() error {
	return errors.New("不支持的授权类型，支持的类型: password, email, sms, xcx, wechat")
}
