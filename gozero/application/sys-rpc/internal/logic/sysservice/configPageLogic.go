package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfigPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigPageLogic {
	return &ConfigPageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfigPageLogic) ConfigPage(in *pb.ConfigPageReq) (*pb.ConfigPageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SConfig{})
	if in.Name != "" {
		query = query.Where("name LIKE ?", "%"+in.Name+"%")
	}
	if in.Code != "" {
		query = query.Where("code LIKE ?", "%"+in.Code+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.SConfig
	if err := query.Order("code ASC, id ASC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.ConfigPageResp{Records: toConfigList(rows), Page: toPageInfo(total, pageNum, pageSize)}, nil
}
