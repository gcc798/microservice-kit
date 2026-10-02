package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeLogic {
	return &DictTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictTypeLogic) DictType(in *pb.DictTypeQueryReq) (*pb.DictListResp, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Where("dict_type = ? AND status = ?", in.DictType, 0)
	if in.ParentId > 0 {
		query = query.Where("parent_id = ?", in.ParentId)
	}
	var rows []model.SDictData
	if err := query.Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.DictListResp{Records: toDictList(rows)}, nil
}
