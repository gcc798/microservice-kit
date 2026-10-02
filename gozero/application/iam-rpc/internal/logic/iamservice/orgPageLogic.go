package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrgPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrgPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgPageLogic {
	return &OrgPageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrgPageLogic) OrgPage(in *pb.OrgPageReq) (*pb.OrgPageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SOrg{})
	if in.OrgName != "" {
		query = query.Where("org_name LIKE ?", "%"+in.OrgName+"%")
	}
	if in.OrgCode != "" {
		query = query.Where("org_code = ?", in.OrgCode)
	}
	if in.Status == 0 || in.Status == 1 {
		query = query.Where("status = ?", in.Status)
	}
	if in.ParentId > 0 {
		query = query.Where("parent_id = ?", in.ParentId)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []orgRow
	if err := query.Order("sort ASC, id ASC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.OrgPageResp{Records: toOrgList(rows), Page: toPageInfo(total, pageNum, pageSize)}, nil
}
