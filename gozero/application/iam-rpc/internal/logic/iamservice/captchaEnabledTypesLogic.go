package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaEnabledTypesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCaptchaEnabledTypesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaEnabledTypesLogic {
	return &CaptchaEnabledTypesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CaptchaEnabledTypesLogic) CaptchaEnabledTypes(in *pb.CaptchaReq) (*pb.CaptchaTypesResp, error) {
	return &pb.CaptchaTypesResp{Types: captchaEnabledTypes(l.svcCtx)}, nil
}
