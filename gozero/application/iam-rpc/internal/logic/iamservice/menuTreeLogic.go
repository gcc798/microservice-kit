package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type MenuTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuTreeLogic {
	return &MenuTreeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *MenuTreeLogic) MenuTree(in *pb.Empty) (*pb.MenuListResp, error) {
	rows, err := getAllMenus(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	return &pb.MenuListResp{Records: buildMenuTree(rows, 0)}, nil
}
