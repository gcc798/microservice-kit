package sysservicelogic

import (
	"context"
	"errors"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogBatchDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogBatchDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogBatchDeleteLogic {
	return &LoginLogBatchDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogBatchDeleteLogic) LoginLogBatchDelete(in *pb.BatchIdsReq) (*pb.Ack, error) {
	if len(in.Ids) == 0 {
		return nil, errors.New("ids 不能为空")
	}
	if _, err := gorm.G[model.SLoginLog](l.svcCtx.DB).Where("id IN ?", in.Ids).Delete(l.ctx); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
