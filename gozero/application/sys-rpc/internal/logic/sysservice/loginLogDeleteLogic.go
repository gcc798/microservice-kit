package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogDeleteLogic {
	return &LoginLogDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogDeleteLogic) LoginLogDelete(in *pb.IdReq) (*pb.Ack, error) {
	if _, err := gorm.G[model.SLoginLog](l.svcCtx.DB).Where("id = ?", in.Id).Delete(l.ctx); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
