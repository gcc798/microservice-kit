package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictPageLogic {
	return &DictPageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictPageLogic) DictPage(in *pb.DictPageReq) (*pb.DictPageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SDictData{}).Where("parent_id = ?", 0)
	if in.DictType != "" {
		query = query.Where("dict_type LIKE ?", "%"+in.DictType+"%")
	}
	if in.DictLabel != "" {
		query = query.Where("dict_label LIKE ?", "%"+in.DictLabel+"%")
	}
	if in.Status == 0 || in.Status == 1 {
		query = query.Where("status = ?", in.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.SDictData
	if err := query.Order("sort ASC, id ASC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.DictPageResp{Records: toDictList(rows), Page: toPageInfo(total, pageNum, pageSize)}, nil
}
