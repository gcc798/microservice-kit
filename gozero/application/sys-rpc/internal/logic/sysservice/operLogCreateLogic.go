package sysservicelogic

import (
	"context"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperLogCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperLogCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperLogCreateLogic {
	return &OperLogCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperLogCreateLogic) OperLogCreate(in *pb.OperLogReq) (*pb.Ack, error) {
	row := newOperLog(in)
	if err := gorm.G[model.SOperLog](l.svcCtx.DB).Create(l.ctx, &row); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
