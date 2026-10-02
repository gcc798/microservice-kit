package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfigCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigCodeLogic {
	return &ConfigCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfigCodeLogic) ConfigCode(in *pb.ConfigCodeQueryReq) (*pb.ConfigListResp, error) {
	rows, err := gorm.G[model.SConfig](l.svcCtx.DB).Where("code = ?", in.Code).Order("id ASC").Find(l.ctx)
	if err != nil {
		return nil, err
	}
	return &pb.ConfigListResp{Records: toConfigList(rows)}, nil
}
