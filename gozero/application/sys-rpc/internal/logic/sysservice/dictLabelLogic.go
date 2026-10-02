package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictLabelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictLabelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictLabelLogic {
	return &DictLabelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictLabelLogic) DictLabel(in *pb.DictLabelQueryReq) (*pb.DictLabelResp, error) {
	row, err := gorm.G[model.SDictData](l.svcCtx.DB).Select("dict_label").
		Where("dict_type = ? AND dict_value = ? AND status = ?", in.DictType, in.DictValue, 0).
		Order("sort ASC, id ASC").First(l.ctx)
	if err != nil {
		return nil, err
	}
	return &pb.DictLabelResp{Label: row.DictLabel.String}, nil
}
