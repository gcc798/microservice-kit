package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type RolePageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRolePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePageLogic {
	return &RolePageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RolePageLogic) RolePage(in *pb.RolePageReq) (*pb.RolePageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SRole{})
	if in.RoleName != "" {
		query = query.Where("role_name LIKE ?", "%"+in.RoleName+"%")
	}
	if in.Status == 0 || in.Status == 1 {
		query = query.Where("status = ?", in.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []roleRow
	if err := query.Order("id DESC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.RolePageResp{Records: toRoleList(rows), Page: toPageInfo(total, pageNum, pageSize)}, nil
}
