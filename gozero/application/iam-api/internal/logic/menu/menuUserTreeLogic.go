package menu

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/zeromicro/go-zero/core/logx"
)

type MenuUserTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuUserTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuUserTreeLogic {
	return &MenuUserTreeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *MenuUserTreeLogic) MenuUserTree(userId int64) (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.IamRpcClient.MenuUserTree(l.ctx, &iamservice.IdReq{Id: userId})
	if err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: toNativeMenuTreeList(data.Records)}, nil
}
