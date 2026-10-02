package menu

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/zeromicro/go-zero/core/logx"
)

type MenuDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuDetailLogic {
	return &MenuDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *MenuDetailLogic) MenuDetail(req *types.IdPathReq) (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.IamRpcClient.MenuDetail(l.ctx, &iamservice.IdReq{Id: req.Id})
	if err != nil {
		return &types.CommonResp{Code: 404, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: data}, nil
}
