package sysservicelogic

import (
	"context"
	"database/sql"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogCreateLogic {
	return &LoginLogCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogCreateLogic) LoginLogCreate(in *pb.LoginLogReq) (*pb.Ack, error) {
	row := model.SLoginLog{
		UserName: sql.NullString{String: in.UserName, Valid: in.UserName != ""}, Ipaddr: sql.NullString{String: in.Ipaddr, Valid: in.Ipaddr != ""},
		LoginLocation: sql.NullString{String: in.LoginLocation, Valid: in.LoginLocation != ""}, Browser: sql.NullString{String: in.Browser, Valid: in.Browser != ""},
		Os: sql.NullString{String: in.Os, Valid: in.Os != ""}, Status: int64(in.Status), Msg: sql.NullString{String: in.Msg, Valid: in.Msg != ""},
		LoginTime: sql.NullTime{Time: time.Now(), Valid: true}, ClientId: sql.NullString{String: in.ClientId, Valid: in.ClientId != ""},
	}
	if err := gorm.G[model.SLoginLog](l.svcCtx.DB).Create(l.ctx, &row); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
