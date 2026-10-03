package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type OperLogBatchCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperLogBatchCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperLogBatchCreateLogic {
	return &OperLogBatchCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperLogBatchCreateLogic) OperLogBatchCreate(in *pb.OperLogBatchReq) (*pb.Ack, error) {
	if len(in.Logs) == 0 {
		return &pb.Ack{Msg: "ok"}, nil
	}
	rows := make([]model.SOperLog, 0, len(in.Logs))
	for _, item := range in.Logs {
		rows = append(rows, newOperLog(item))
	}
	if err := gorm.G[model.SOperLog](l.svcCtx.DB).CreateInBatches(l.ctx, &rows, 100); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
