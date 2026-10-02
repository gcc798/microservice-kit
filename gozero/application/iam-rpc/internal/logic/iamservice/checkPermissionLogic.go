package iamservicelogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermissionLogic {
	return &CheckPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckPermissionLogic) CheckPermission(in *pb.CheckPermissionReq) (*pb.CheckPermissionResp, error) {
	if in == nil || in.UserId <= 0 || strings.TrimSpace(in.Resource) == "" || strings.TrimSpace(in.Action) == "" {
		return &pb.CheckPermissionResp{}, nil
	}
	roles, err := getUserRoleKeys(l.ctx, l.svcCtx, in.UserId)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role == "super_admin" {
			return &pb.CheckPermissionResp{Allowed: true}, nil
		}
	}

	type grant struct {
		Code   string
		Action string
	}
	var roleGrants []grant
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SApiPermission{}).
		Select("s_api_permission.code, s_api_permission.action").
		Joins("JOIN m_role_api_permission ON m_role_api_permission.permission_id = s_api_permission.id").
		Joins("JOIN m_user_role ON m_user_role.role_id = m_role_api_permission.role_id").
		Joins("JOIN s_role ON s_role.id = m_user_role.role_id").
		Where("m_user_role.user_id = ? AND s_role.status = ? AND s_api_permission.status = ?", in.UserId, 0, 0).
		Find(&roleGrants).Error; err != nil {
		return nil, fmt.Errorf("查询角色 API 权限失败: %w", err)
	}
	var userGrants []grant
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SApiPermission{}).
		Select("s_api_permission.code, s_api_permission.action").
		Joins("JOIN m_user_api_permission ON m_user_api_permission.permission_id = s_api_permission.id").
		Where("m_user_api_permission.user_id = ? AND s_api_permission.status = ?", in.UserId, 0).
		Find(&userGrants).Error; err != nil {
		return nil, fmt.Errorf("查询用户 API 权限失败: %w", err)
	}
	for _, grant := range append(roleGrants, userGrants...) {
		if permissionMatches(grant.Code, in.Resource) && (grant.Action == in.Action || grant.Action == "*") {
			return &pb.CheckPermissionResp{Allowed: true}, nil
		}
	}

	return &pb.CheckPermissionResp{}, nil
}

func permissionMatches(granted, required string) bool {
	return granted == required || granted == "*" ||
		(strings.HasSuffix(granted, "*") && strings.HasPrefix(required, strings.TrimSuffix(granted, "*"))) ||
		(strings.HasPrefix(granted, "*") && strings.HasSuffix(required, strings.TrimPrefix(granted, "*")))
}
