// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package captcha

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/logic/commonutil"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaImageLogic {
	return &CaptchaImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaptchaImageLogic) CaptchaImage() (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.IamRpcClient.CaptchaImage(l.ctx, &iamservice.CaptchaReq{})
	if err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: map[string]interface{}{
		"id":       data.Id,
		"type":     data.Type,
		"data":     commonutil.JSONStringToValue(data.DataJson),
		"expireAt": data.ExpireAt,
	}}, nil
}
