package iamservicelogic

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserAuthContextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserAuthContextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserAuthContextLogic {
	return &UserAuthContextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserAuthContextLogic) UserAuthContext(in *pb.UserAuthContextReq) (*pb.UserAuthContextResp, error) {
	if in.UserId == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}
	orgID, err := getUserOrgID(l.ctx, l.svcCtx, in.UserId)
	if err != nil {
		return nil, err
	}
	roles, err := getUserRoleKeys(l.ctx, l.svcCtx, in.UserId)
	if err != nil {
		return nil, err
	}
	permissions, err := getUserPermissionCodes(l.ctx, l.svcCtx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.UserAuthContextResp{
		OrgId:       orgID,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func getUserOrgID(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) (int64, error) {
	var orgID int64
	if err := svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("id = ?", userID).Select("org_id").Scan(&orgID).Error; err != nil {
		return 0, fmt.Errorf("查询用户组织失败: %w", err)
	}
	return orgID, nil
}

func getUserRoleKeys(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) ([]string, error) {
	var rows []string
	if err := svcCtx.DB.WithContext(ctx).Model(&model.SRole{}).
		Joins("JOIN m_user_role ON m_user_role.role_id = s_role.id").
		Where("m_user_role.user_id = ? AND s_role.status = ? AND s_role.role_key <> ''", userID, 0).
		Distinct().Order("s_role.role_key ASC").Pluck("s_role.role_key", &rows).Error; err != nil {
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}
	return uniqueStringValues(rows), nil
}

func getUserPermissionCodes(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) ([]string, error) {
	permissions := make(map[string]struct{})
	var rolePermissions []string
	if err := svcCtx.DB.WithContext(ctx).Model(&model.SApiPermission{}).
		Joins("JOIN m_role_api_permission ON m_role_api_permission.permission_id = s_api_permission.id").
		Joins("JOIN m_user_role ON m_user_role.role_id = m_role_api_permission.role_id").
		Joins("JOIN s_role ON s_role.id = m_user_role.role_id").
		Where("m_user_role.user_id = ? AND s_role.status = ? AND s_api_permission.status = ? AND s_api_permission.code <> ''", userID, 0, 0).
		Distinct().Pluck("s_api_permission.code", &rolePermissions).Error; err != nil {
		return nil, fmt.Errorf("查询用户权限失败: %w", err)
	}
	var userPermissions []string
	if err := svcCtx.DB.WithContext(ctx).Model(&model.SApiPermission{}).
		Joins("JOIN m_user_api_permission ON m_user_api_permission.permission_id = s_api_permission.id").
		Where("m_user_api_permission.user_id = ? AND s_api_permission.status = ? AND s_api_permission.code <> ''", userID, 0).
		Distinct().Pluck("s_api_permission.code", &userPermissions).Error; err != nil {
		return nil, fmt.Errorf("查询用户权限失败: %w", err)
	}
	for _, row := range append(rolePermissions, userPermissions...) {
		addPermission(permissions, row)
	}
	return sortedStringKeys(permissions), nil
}

func addPermission(permissions map[string]struct{}, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	permissions[value] = struct{}{}
}

func uniqueStringValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sortedStringKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
