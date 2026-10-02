package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperLogPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperLogPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperLogPageLogic {
	return &OperLogPageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperLogPageLogic) OperLogPage(in *pb.OperLogPageReq) (*pb.OperLogPageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SOperLog{})
	if in.Title != "" {
		query = query.Where("title LIKE ?", "%"+in.Title+"%")
	}
	if in.OperName != "" {
		query = query.Where("oper_name LIKE ?", "%"+in.OperName+"%")
	}
	if in.BusinessType != "" {
		query = query.Where("business_type = ?", in.BusinessType)
	}
	if in.Status != "" {
		query = query.Where("status = ?", in.Status)
	}
	if in.StartTime != "" {
		query = query.Where("oper_time >= ?", in.StartTime)
	}
	if in.EndTime != "" {
		query = query.Where("oper_time <= ?", in.EndTime)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.SOperLog
	if err := query.Order("oper_time DESC NULLS LAST, id DESC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.OperLogPageResp{Records: toOperLogList(rows), Page: toPageInfo(total, pageNum, pageSize)}, nil
}
