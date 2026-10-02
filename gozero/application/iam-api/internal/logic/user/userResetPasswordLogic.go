// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserResetPasswordLogic {
	return &UserResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserResetPasswordLogic) UserResetPassword(req *types.UserPasswordPathReq) (resp *types.CommonResp, err error) {
	if _, err := l.svcCtx.IamRpcClient.UserResetPassword(l.ctx, &iamservice.UserPasswordReq{Id: req.Id, NewPassword: req.NewPassword}); err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: "ok"}, nil
}
