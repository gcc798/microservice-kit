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

type ApiPermissionDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApiPermissionDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiPermissionDeleteLogic {
	return &ApiPermissionDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApiPermissionDeleteLogic) ApiPermissionDelete(in *pb.IdReq) (*pb.Ack, error) {
	childCount, err := gorm.G[model.SApiPermission](l.svcCtx.DB).Where("parent_id = ?", in.Id).Count(l.ctx, "id")
	if err != nil {
		return nil, fmt.Errorf("检查子权限失败: %w", err)
	}
	if childCount > 0 {
		return nil, fmt.Errorf("存在子权限，无法删除")
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("permission_id = ?", in.Id).Delete(&model.MRoleApiPermission{}).Error; err != nil {
			return fmt.Errorf("删除角色 API 权限关联失败: %w", err)
		}
		if err := tx.Where("permission_id = ?", in.Id).Delete(&model.MUserApiPermission{}).Error; err != nil {
			return fmt.Errorf("删除用户 API 权限关联失败: %w", err)
		}
		result := tx.Delete(&model.SApiPermission{}, in.Id)
		if result.Error != nil {
			return fmt.Errorf("删除 API 权限失败: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("API 权限不存在")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
