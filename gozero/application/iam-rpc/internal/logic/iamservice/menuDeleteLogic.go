package iamservicelogic

import (
	"context"
	"fmt"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type MenuDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMenuDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuDeleteLogic {
	return &MenuDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MenuDeleteLogic) MenuDelete(in *pb.IdReq) (*pb.Ack, error) {
	if in.Id <= 0 {
		return nil, fmt.Errorf("菜单ID不能为空")
	}
	if _, err := getMenuByID(l.ctx, l.svcCtx, in.Id); err != nil {
		return nil, err
	}
	count, err := gorm.G[model.SMenu](l.svcCtx.DB).Where("parent_id = ?", in.Id).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("存在子菜单，无法删除")
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&model.SMenu{}, in.Id).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
