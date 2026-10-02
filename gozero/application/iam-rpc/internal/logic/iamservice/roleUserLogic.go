package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleUserLogic {
	return &RoleUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleUserLogic) RoleUser(in *pb.UserRoleQueryReq) (*pb.RoleListResp, error) {
	var rows []roleRow
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SRole{}).
		Select("s_role.id, s_role.role_key, s_role.role_name, s_role.sort, s_role.status, s_role.data_scope, s_role.is_system, s_role.remark, s_role.create_by, s_role.created_time").
		Joins("JOIN m_user_role ON m_user_role.role_id = s_role.id").
		Where("m_user_role.user_id = ?", in.UserId).
		Order("s_role.sort ASC, s_role.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.RoleListResp{Records: toRoleList(rows)}, nil
}
