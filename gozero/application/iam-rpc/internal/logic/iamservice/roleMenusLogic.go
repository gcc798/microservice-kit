package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleMenusLogic {
	return &RoleMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleMenusLogic) RoleMenus(in *pb.RoleMenusReq) (*pb.MenuIdsResp, error) {
	var menuIds []int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.MRoleMenu{}).Where("role_id = ?", in.RoleId).Order("menu_id ASC").Pluck("menu_id", &menuIds).Error; err != nil {
		return nil, err
	}
	return &pb.MenuIdsResp{MenuIds: menuIds}, nil
}
