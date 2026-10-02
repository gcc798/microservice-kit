package iamservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"gorm.io/gorm"
)

type apiPermissionRow = model.SApiPermission

func listApiPermissions(ctx context.Context, svcCtx *svc.ServiceContext) ([]apiPermissionRow, error) {
	var rows []apiPermissionRow
	err := svcCtx.DB.WithContext(ctx).Order("sort ASC, created_time ASC").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询 API 权限失败: %w", err)
	}
	return rows, nil
}

func getApiPermissionByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*apiPermissionRow, error) {
	var row apiPermissionRow
	err := svcCtx.DB.WithContext(ctx).Where("id = ?", id).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("API 权限不存在")
		}
		return nil, err
	}
	return &row, nil
}

func toApiPermissionPB(row apiPermissionRow) *pb.ApiPermission {
	return &pb.ApiPermission{
		Id:          row.Id,
		ParentId:    row.ParentId,
		Module:      row.Module,
		Code:        row.Code,
		Name:        row.Name,
		NodeType:    int32(row.NodeType),
		Action:      row.Action,
		Method:      nullString(row.Method),
		Path:        nullString(row.Path),
		Sort:        row.Sort,
		Status:      int32(row.Status),
		Remark:      nullString(row.Remark),
		CreateBy:    nullInt64(row.CreateBy),
		UpdateBy:    nullInt64(row.UpdateBy),
		CreatedTime: nullTime(row.CreatedTime),
		UpdatedTime: nullTime(row.UpdatedTime),
	}
}

func toApiPermissionList(rows []apiPermissionRow) []*pb.ApiPermission {
	list := make([]*pb.ApiPermission, 0, len(rows))
	for _, row := range rows {
		list = append(list, toApiPermissionPB(row))
	}
	return list
}

func buildApiPermissionTree(rows []apiPermissionRow, parentID int64) []*pb.ApiPermission {
	tree := make([]*pb.ApiPermission, 0)
	for _, row := range rows {
		if row.ParentId != parentID {
			continue
		}
		node := toApiPermissionPB(row)
		node.Children = buildApiPermissionTree(rows, row.Id)
		tree = append(tree, node)
	}
	return tree
}

func validateApiPermission(ctx context.Context, svcCtx *svc.ServiceContext, req *pb.ApiPermissionSaveReq) error {
	if strings.TrimSpace(req.Module) == "" || strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("模块、权限标识和权限名称不能为空")
	}
	if strings.TrimSpace(req.Action) == "" {
		return fmt.Errorf("操作类型不能为空")
	}
	if req.Id != 0 && req.ParentId == req.Id {
		return fmt.Errorf("不能将自己设置为父级权限")
	}
	if err := validateApiPermissionParent(ctx, svcCtx, req.ParentId, req.Id); err != nil {
		return err
	}
	count, err := gorm.G[model.SApiPermission](svcCtx.DB).Where("code = ? AND id <> ?", req.Code, req.Id).Count(ctx, "id")
	if err != nil {
		return fmt.Errorf("检查权限标识失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("权限标识已存在")
	}
	return nil
}

func validateApiPermissionParent(ctx context.Context, svcCtx *svc.ServiceContext, parentID, selfID int64) error {
	visited := make(map[int64]struct{})
	for parentID != 0 {
		if _, ok := visited[parentID]; ok {
			return fmt.Errorf("父级权限存在循环引用")
		}
		visited[parentID] = struct{}{}
		if selfID != 0 && parentID == selfID {
			return fmt.Errorf("不能将自己的下级设置为父级权限")
		}
		parent, err := getApiPermissionByID(ctx, svcCtx, parentID)
		if err != nil {
			return fmt.Errorf("父级权限不存在")
		}
		parentID = parent.ParentId
	}
	return nil
}

func apiPermissionIDsByOwner(ctx context.Context, svcCtx *svc.ServiceContext, table, ownerColumn string, ownerID int64) ([]int64, error) {
	var ids []int64
	var query *gorm.DB
	switch table {
	case "m_role_api_permission":
		query = svcCtx.DB.WithContext(ctx).Model(&model.MRoleApiPermission{}).Where("role_id = ?", ownerID)
	case "m_user_api_permission":
		query = svcCtx.DB.WithContext(ctx).Model(&model.MUserApiPermission{}).Where("user_id = ?", ownerID)
	default:
		return nil, fmt.Errorf("不支持的权限关联: %s.%s", table, ownerColumn)
	}
	if err := query.Distinct().Order("permission_id ASC").Pluck("permission_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func resolveAssignableApiPermissions(ctx context.Context, svcCtx *svc.ServiceContext, ids []int64) ([]int64, error) {
	uniqueIDs := uniqueInt64Values(ids)
	if len(uniqueIDs) == 0 {
		return nil, nil
	}
	var rows []apiPermissionRow
	if err := svcCtx.DB.WithContext(ctx).Where("id IN ? AND status = ?", uniqueIDs, 0).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询 API 权限失败: %w", err)
	}
	if len(rows) != len(uniqueIDs) {
		return nil, fmt.Errorf("存在无效或停用的 API 权限")
	}
	normalized := normalizeCoveredApiPermissions(rows)
	normalizedIDs := make([]int64, 0, len(normalized))
	for _, row := range normalized {
		normalizedIDs = append(normalizedIDs, row.Id)
	}
	sort.Slice(normalizedIDs, func(i, j int) bool { return normalizedIDs[i] < normalizedIDs[j] })
	return normalizedIDs, nil
}

func replaceRoleApiPermissions(ctx context.Context, svcCtx *svc.ServiceContext, roleID int64, permissionIDs []int64, operatorID int64) error {
	normalizedIDs, err := resolveAssignableApiPermissions(ctx, svcCtx, permissionIDs)
	if err != nil {
		return err
	}
	return svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND source = ?", roleID, 0).Delete(&model.MRoleApiPermission{}).Error; err != nil {
			return fmt.Errorf("删除旧角色 API 权限失败: %w", err)
		}
		if len(normalizedIDs) == 0 {
			return nil
		}
		now := sql.NullTime{Time: time.Now(), Valid: true}
		rows := make([]model.MRoleApiPermission, 0, len(normalizedIDs))
		for _, permissionID := range normalizedIDs {
			rows = append(rows, model.MRoleApiPermission{
				RoleId: roleID, PermissionId: permissionID, Source: 0,
				CreateBy: nullableInt64(operatorID), UpdateBy: nullableInt64(operatorID), CreatedTime: now, UpdatedTime: now,
			})
		}
		return tx.Create(&rows).Error
	})
}

func replaceUserApiPermissions(ctx context.Context, svcCtx *svc.ServiceContext, userID int64, permissionIDs []int64, operatorID int64) error {
	normalizedIDs, err := resolveAssignableApiPermissions(ctx, svcCtx, permissionIDs)
	if err != nil {
		return err
	}
	return svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND source = ?", userID, 0).Delete(&model.MUserApiPermission{}).Error; err != nil {
			return fmt.Errorf("删除旧用户 API 权限失败: %w", err)
		}
		if len(normalizedIDs) == 0 {
			return nil
		}
		now := sql.NullTime{Time: time.Now(), Valid: true}
		rows := make([]model.MUserApiPermission, 0, len(normalizedIDs))
		for _, permissionID := range normalizedIDs {
			rows = append(rows, model.MUserApiPermission{
				UserId: userID, PermissionId: permissionID, Source: 0,
				CreateBy: nullableInt64(operatorID), UpdateBy: nullableInt64(operatorID), CreatedTime: now, UpdatedTime: now,
			})
		}
		return tx.Create(&rows).Error
	})
}

func normalizeCoveredApiPermissions(permissions []apiPermissionRow) []apiPermissionRow {
	sort.Slice(permissions, func(i, j int) bool {
		return len(permissions[i].Code) < len(permissions[j].Code)
	})
	selected := make([]apiPermissionRow, 0, len(permissions))
	wildcards := make([]string, 0)
	for _, permission := range permissions {
		covered := false
		for _, wildcard := range wildcards {
			prefix := strings.TrimSuffix(wildcard, "*")
			if permission.Code != wildcard && strings.HasPrefix(permission.Code, prefix) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		selected = append(selected, permission)
		if strings.HasSuffix(permission.Code, ".*") || permission.Code == "*" {
			wildcards = append(wildcards, permission.Code)
		}
	}
	return selected
}

func normalizeApiPermissionAction(code, action string) string {
	if code == "*" || strings.HasSuffix(code, ".*") {
		return "*"
	}
	if action == "" {
		if strings.HasSuffix(code, ".read") {
			return "read"
		}
		return "write"
	}
	return action
}

func uniqueInt64Values(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
