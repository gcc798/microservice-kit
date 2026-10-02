package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserPageLogic {
	return &UserPageLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UserPageLogic) UserPage(in *pb.UserPageReq) (*pb.UserPageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SUser{})
	if in.Username != "" {
		query = query.Where("user_name LIKE ?", "%"+in.Username+"%")
	}
	if in.Phonenumber != "" {
		query = query.Where("phonenumber LIKE ?", "%"+in.Phonenumber+"%")
	}
	if in.Status == 0 || in.Status == 1 {
		query = query.Where("status = ?", in.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.SUser
	if err := query.Order("sort ASC, created_time DESC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]*pb.User, 0, len(rows))
	for _, row := range rows {
		records = append(records, &pb.User{
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
			OrgId:       row.OrgId,
		})
	}
	return &pb.UserPageResp{Records: records, Page: toPageInfo(total, pageNum, pageSize)}, nil
}
