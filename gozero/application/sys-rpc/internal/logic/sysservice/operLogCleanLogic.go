package sysservicelogic

import (
	"context"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperLogCleanLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperLogCleanLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperLogCleanLogic {
	return &OperLogCleanLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperLogCleanLogic) OperLogClean(in *pb.LogCleanReq) (*pb.Ack, error) {
	if in.Days <= 0 {
		return nil, fmt.Errorf("天数必须大于0")
	}
	cutoff := time.Now().AddDate(0, 0, -int(in.Days))
	if _, err := gorm.G[model.SOperLog](l.svcCtx.DB).Where("oper_time < ?", cutoff).Delete(l.ctx); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
