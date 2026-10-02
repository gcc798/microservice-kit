package iamservicelogic

import (
	"context"
	"errors"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RoleDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleDeleteLogic {
	return &RoleDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleDeleteLogic) RoleDelete(in *pb.IdReq) (*pb.Ack, error) {
	row, err := getRoleByID(l.ctx, l.svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	if row.IsSystem {
		return nil, errors.New("系统角色不可删除")
	}
	count, err := gorm.G[model.MUserRole](l.svcCtx.DB).Where("role_id = ?", in.Id).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("该角色已被用户使用，无法删除")
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", in.Id).Delete(&model.MRoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.SRole{}, in.Id).Error
	}); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
