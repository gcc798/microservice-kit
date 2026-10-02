package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperLogDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperLogDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperLogDeleteLogic {
	return &OperLogDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperLogDeleteLogic) OperLogDelete(in *pb.IdReq) (*pb.Ack, error) {
	if _, err := gorm.G[model.SOperLog](l.svcCtx.DB).Where("id = ?", in.Id).Delete(l.ctx); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
