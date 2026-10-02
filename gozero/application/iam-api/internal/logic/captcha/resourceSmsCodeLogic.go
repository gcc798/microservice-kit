package captcha

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/zeromicro/go-zero/core/logx"
)

type ResourceSmsCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceSmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceSmsCodeLogic {
	return &ResourceSmsCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResourceSmsCodeLogic) ResourceSmsCode(req *types.ResourceSmsCodeReq) (resp *types.CommonResp, err error) {
	phone := req.Phonenumber
	if phone == "" {
		phone = req.Phone
	}
	_, err = l.svcCtx.IamRpcClient.CaptchaSms(l.ctx, &iamservice.CaptchaPhoneReq{Phonenumber: phone})
	if err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功"}, nil
}
