package iamservicelogic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RoleAssignMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleAssignMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleAssignMenusLogic {
	return &RoleAssignMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleAssignMenusLogic) RoleAssignMenus(in *pb.RoleMenusAssignReq) (*pb.Ack, error) {
	if in.RoleId <= 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}
	if _, err := getRoleByID(l.ctx, l.svcCtx, in.RoleId); err != nil {
		return nil, err
	}
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", in.RoleId).Delete(&model.MRoleMenu{}).Error; err != nil {
			return err
		}
		if len(in.MenuIds) == 0 {
			return nil
		}
		now := sql.NullTime{Time: time.Now(), Valid: true}
		rows := make([]model.MRoleMenu, 0, len(in.MenuIds))
		for _, menuID := range in.MenuIds {
			rows = append(rows, model.MRoleMenu{RoleId: in.RoleId, MenuId: menuID, CreatedTime: now, UpdatedTime: now})
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
