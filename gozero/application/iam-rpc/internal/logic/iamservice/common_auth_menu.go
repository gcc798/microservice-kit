package iamservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authClientRow struct {
	ClientId      string         `db:"client_id"`
	ClientKey     string         `db:"client_key"`
	ClientSecret  string         `db:"client_secret"`
	GrantType     sql.NullString `db:"grant_type"`
	DeviceType    sql.NullString `db:"device_type"`
	Status        int64          `db:"status"`
	Timeout       int64          `db:"timeout"`
	ActiveTimeout int64          `db:"active_timeout"`
}

type userAuthRow struct {
	Id          int64          `db:"id"`
	OrgId       sql.NullInt64  `db:"org_id"`
	UserName    string         `db:"user_name"`
	NickName    sql.NullString `db:"nick_name"`
	UserType    int64          `db:"user_type"`
	Email       sql.NullString `db:"email"`
	Phonenumber sql.NullString `db:"phonenumber"`
	Avatar      sql.NullString `db:"avatar"`
	Password    sql.NullString `db:"password"`
	Status      int64          `db:"status"`
}

type menuRow struct {
	Id          int64          `db:"id"`
	MenuName    string         `db:"menu_name"`
	ParentId    int64          `db:"parent_id"`
	Sort        int64          `db:"sort"`
	Path        sql.NullString `db:"path"`
	Component   sql.NullString `db:"component"`
	Query       sql.NullString `db:"query"`
	IsFrame     int64          `db:"is_frame"`
	IsCache     int64          `db:"is_cache"`
	MenuType    int64          `db:"menu_type"`
	Visible     int64          `db:"visible"`
	Status      int64          `db:"status"`
	Perms       sql.NullString `db:"perms"`
	Icon        sql.NullString `db:"icon"`
	Remark      sql.NullString `db:"remark"`
	CreateBy    sql.NullInt64  `db:"create_by"`
	UpdateBy    sql.NullInt64  `db:"update_by"`
	CreatedTime sql.NullTime   `db:"created_time"`
	UpdatedTime sql.NullTime   `db:"updated_time"`
}

type roleKeyRow struct {
	RoleKey sql.NullString `db:"role_key"`
}

func authenticateClient(ctx context.Context, svcCtx *svc.ServiceContext, clientId, grantType string) (*authClientRow, error) {
	if clientId == "" {
		return nil, fmt.Errorf("clientId不能为空")
	}
	if grantType == "" {
		return nil, fmt.Errorf("grantType不能为空")
	}
	var row authClientRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SAuthClient{}).Where("client_id = ?", clientId).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("客户端不存在")
		}
		return nil, fmt.Errorf("查询客户端失败")
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("客户端已停用")
	}
	if !grantTypeSupported(row.GrantType.String, grantType) {
		return nil, fmt.Errorf("客户端不支持该授权类型")
	}
	return &row, nil
}

func authenticatePassword(ctx context.Context, svcCtx *svc.ServiceContext, username, password string) (*userAuthRow, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("用户名和密码不能为空")
	}
	var row userAuthRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("user_name = ?", username).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		return nil, fmt.Errorf("登录失败")
	}
	if !row.Password.Valid || row.Password.String == "" {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.Password.String), []byte(password)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	return &row, nil
}

func authenticateEmail(ctx context.Context, svcCtx *svc.ServiceContext, email, uuidValue, code string) (*userAuthRow, error) {
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}
	if err := verifyEmailCaptcha(ctx, svcCtx, uuidValue, email, code); err != nil {
		return nil, err
	}
	var row userAuthRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("email = ?", email).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("邮箱或验证码错误")
		}
		return nil, fmt.Errorf("登录失败")
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	return &row, nil
}

func authenticateSms(ctx context.Context, svcCtx *svc.ServiceContext, phonenumber, uuidValue, code string) (*userAuthRow, error) {
	if phonenumber == "" {
		return nil, fmt.Errorf("手机号不能为空")
	}
	if err := verifySmsCaptcha(ctx, svcCtx, uuidValue, phonenumber, code); err != nil {
		return nil, err
	}
	var row userAuthRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("phonenumber = ?", phonenumber).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("手机号或验证码错误")
		}
		return nil, fmt.Errorf("登录失败")
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	return &row, nil
}

func grantTypeSupported(supported, actual string) bool {
	if supported == "" || supported == actual {
		return true
	}
	for _, item := range splitAndTrim(supported) {
		if item == actual {
			return true
		}
	}
	return false
}

func splitAndTrim(s string) []string {
	out := make([]string, 0)
	cur := ""
	for _, r := range s {
		if r == ',' || r == '|' || r == ' ' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func getUserByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*pb.User, error) {
	var row struct {
		Id          int64          `db:"id"`
		UserName    string         `db:"user_name"`
		NickName    sql.NullString `db:"nick_name"`
		UserType    int64          `db:"user_type"`
		Email       sql.NullString `db:"email"`
		Phonenumber sql.NullString `db:"phonenumber"`
		Sex         int64          `db:"sex"`
		Avatar      sql.NullString `db:"avatar"`
		Status      int64          `db:"status"`
		Sort        int64          `db:"sort"`
		LoginIp     sql.NullString `db:"login_ip"`
		LoginDate   sql.NullInt64  `db:"login_date"`
		OpenId      sql.NullString `db:"open_id"`
		UnionId     sql.NullString `db:"union_id"`
		Remark      sql.NullString `db:"remark"`
		CreateBy    sql.NullInt64  `db:"create_by"`
		UpdateBy    sql.NullInt64  `db:"update_by"`
		CreatedTime sql.NullTime   `db:"created_time"`
		UpdatedTime sql.NullTime   `db:"updated_time"`
		OrgId       sql.NullInt64  `db:"org_id"`
	}
	err := svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("id = ?", id).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	return &pb.User{
		UserId:      row.Id,
		UserName:    row.UserName,
		NickName:    nullString(row.NickName),
		UserType:    int32(row.UserType),
		Email:       nullString(row.Email),
		Phonenumber: nullString(row.Phonenumber),
		Sex:         int32(row.Sex),
		Avatar:      nullString(row.Avatar),
		Status:      int32(row.Status),
		Sort:        row.Sort,
		LoginIp:     nullString(row.LoginIp),
		LoginDate:   nullInt64(row.LoginDate),
		OpenId:      nullString(row.OpenId),
		UnionId:     nullString(row.UnionId),
		Remark:      nullString(row.Remark),
		CreateBy:    nullInt64(row.CreateBy),
		UpdateBy:    nullInt64(row.UpdateBy),
		CreatedAt:   nullTime(row.CreatedTime),
		UpdatedAt:   nullTime(row.UpdatedTime),
		OrgId:       nullInt64(row.OrgId),
	}, nil
}

func getAllMenus(ctx context.Context, svcCtx *svc.ServiceContext) ([]menuRow, error) {
	var rows []menuRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SMenu{}).Where("status = ?", 0).Order("sort ASC, id ASC").Find(&rows).Error
	return rows, err
}

func getUserMenus(ctx context.Context, svcCtx *svc.ServiceContext, userId int64) ([]menuRow, error) {
	var roles []roleKeyRow
	if err := svcCtx.DB.WithContext(ctx).Model(&model.SRole{}).Select("s_role.role_key").
		Joins("JOIN m_user_role ON m_user_role.role_id = s_role.id").
		Where("m_user_role.user_id = ? AND s_role.status = ?", userId, 0).Scan(&roles).Error; err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.RoleKey.Valid && role.RoleKey.String == "super_admin" {
			return getAllMenus(ctx, svcCtx)
		}
	}
	var rows []menuRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SMenu{}).Distinct("s_menu.*").
		Joins("JOIN m_role_menu ON m_role_menu.menu_id = s_menu.id").
		Joins("JOIN m_user_role ON m_user_role.role_id = m_role_menu.role_id").
		Joins("JOIN s_role ON s_role.id = m_user_role.role_id").
		Where("m_user_role.user_id = ? AND s_menu.status = ? AND s_role.status = ?", userId, 0, 0).
		Order("s_menu.sort ASC, s_menu.id ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return withAncestorMenus(ctx, svcCtx, rows)
}

func withAncestorMenus(ctx context.Context, svcCtx *svc.ServiceContext, menus []menuRow) ([]menuRow, error) {
	if len(menus) == 0 {
		return menus, nil
	}
	menuByID := make(map[int64]menuRow, len(menus))
	for _, m := range menus {
		menuByID[m.Id] = m
	}
	missingParentIDs := make([]int64, 0)
	missingSet := make(map[int64]struct{})
	for _, m := range menus {
		if m.ParentId != 0 {
			if _, ok := menuByID[m.ParentId]; !ok {
				if _, already := missingSet[m.ParentId]; !already {
					missingParentIDs = append(missingParentIDs, m.ParentId)
					missingSet[m.ParentId] = struct{}{}
				}
			}
		}
	}
	for len(missingParentIDs) > 0 {
		var parents []menuRow
		if err := svcCtx.DB.WithContext(ctx).Model(&model.SMenu{}).Where("id IN ? AND status = ?", missingParentIDs, 0).Find(&parents).Error; err != nil {
			return nil, fmt.Errorf("failed to get ancestor menus: %w", err)
		}
		missingParentIDs = nil
		missingSet = make(map[int64]struct{})
		for _, parent := range parents {
			if _, exists := menuByID[parent.Id]; exists {
				continue
			}
			menuByID[parent.Id] = parent
			if parent.ParentId != 0 {
				if _, ok := menuByID[parent.ParentId]; !ok {
					if _, already := missingSet[parent.ParentId]; !already {
						missingParentIDs = append(missingParentIDs, parent.ParentId)
						missingSet[parent.ParentId] = struct{}{}
					}
				}
			}
		}
	}
	result := make([]menuRow, 0, len(menuByID))
	for _, m := range menuByID {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Sort == result[j].Sort {
			return result[i].Id < result[j].Id
		}
		return result[i].Sort < result[j].Sort
	})
	return result, nil
}

func getMenuByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*pb.Menu, error) {
	var row menuRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SMenu{}).Where("id = ?", id).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("菜单不存在")
		}
		return nil, err
	}
	return toMenuPB(row), nil
}

func buildMenuTree(rows []menuRow, parentID int64) []*pb.Menu {
	tree := make([]*pb.Menu, 0)
	for _, row := range rows {
		if row.ParentId != parentID {
			continue
		}
		node := toMenuPB(row)
		node.Children = buildMenuTree(rows, row.Id)
		tree = append(tree, node)
	}
	return tree
}

func toMenuList(rows []menuRow) []*pb.Menu {
	list := make([]*pb.Menu, 0, len(rows))
	for _, row := range rows {
		list = append(list, toMenuPB(row))
	}
	return list
}

func toMenuPB(row menuRow) *pb.Menu {
	return &pb.Menu{
		Id:          row.Id,
		MenuName:    row.MenuName,
		ParentId:    row.ParentId,
		Sort:        row.Sort,
		Path:        nullString(row.Path),
		Component:   nullString(row.Component),
		Query:       nullString(row.Query),
		IsFrame:     row.IsFrame,
		IsCache:     row.IsCache,
		MenuType:    row.MenuType,
		Visible:     row.Visible,
		Status:      row.Status,
		Perms:       nullString(row.Perms),
		Icon:        nullString(row.Icon),
		Remark:      nullString(row.Remark),
		CreateBy:    nullInt64(row.CreateBy),
		UpdateBy:    nullInt64(row.UpdateBy),
		CreatedTime: nullTime(row.CreatedTime),
		UpdatedTime: nullTime(row.UpdatedTime),
	}
}

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}
func nullInt64(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}
func nullTime(v sql.NullTime) string {
	if v.Valid {
		return v.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}

func nullableString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullableInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value != 0}
}
