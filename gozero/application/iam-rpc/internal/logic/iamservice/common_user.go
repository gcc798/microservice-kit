package iamservicelogic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
)

type userRow struct {
	Id          int64          `db:"id"`
	OrgId       sql.NullInt64  `db:"org_id"`
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
}

func listUsersByRole(ctx context.Context, svcCtx *svc.ServiceContext, roleID int64) ([]userRow, error) {
	var rows []userRow
	err := svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).
		Select("s_user.id, s_user.user_name, s_user.nick_name, s_user.user_type, s_user.email, s_user.phonenumber, s_user.sex, s_user.avatar, s_user.status, s_user.sort, s_user.login_ip, s_user.login_date, s_user.open_id, s_user.union_id, s_user.remark, s_user.create_by, s_user.update_by, s_user.created_time, s_user.updated_time, s_user.org_id").
		Joins("JOIN m_user_role ON m_user_role.user_id = s_user.id").
		Where("m_user_role.role_id = ?", roleID).Order("s_user.created_time DESC").Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询角色用户失败: %w", err)
	}
	return rows, nil
}

func toUserPB(row userRow) *pb.User {
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
	}
}

func toUserListPB(rows []userRow) []*pb.User {
	list := make([]*pb.User, 0, len(rows))
	for _, row := range rows {
		list = append(list, toUserPB(row))
	}
	return list
}
