package sysservicelogic

import (
	"context"
	"database/sql"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogUpdateLogic {
	return &LoginLogUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogUpdateLogic) LoginLogUpdate(in *pb.LoginLogUpdateReq) (*pb.Ack, error) {
	updates := map[string]any{
		"user_name": sql.NullString{String: in.UserName, Valid: in.UserName != ""}, "ipaddr": sql.NullString{String: in.Ipaddr, Valid: in.Ipaddr != ""},
		"login_location": sql.NullString{String: in.LoginLocation, Valid: in.LoginLocation != ""}, "browser": sql.NullString{String: in.Browser, Valid: in.Browser != ""},
		"os": sql.NullString{String: in.Os, Valid: in.Os != ""}, "status": in.Status,
		"msg": sql.NullString{String: in.Msg, Valid: in.Msg != ""}, "client_id": sql.NullString{String: in.ClientId, Valid: in.ClientId != ""},
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SLoginLog{}).Where("id = ?", in.Id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
