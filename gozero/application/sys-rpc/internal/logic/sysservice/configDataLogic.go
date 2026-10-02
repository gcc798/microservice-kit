package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfigDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigDataLogic {
	return &ConfigDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfigDataLogic) ConfigData(in *pb.ConfigCodeQueryReq) (*pb.ConfigDataResp, error) {
	row, err := gorm.G[model.SConfig](l.svcCtx.DB).Select("data").Where("code = ?", in.Code).Order("id DESC").First(l.ctx)
	if err != nil {
		return nil, err
	}
	return &pb.ConfigDataResp{Code: in.Code, DataJson: row.Data}, nil
}
