package sysservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigUpdateLogic {
	return &ConfigUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfigUpdateLogic) ConfigUpdate(in *pb.ConfigUpdateReq) (*pb.Ack, error) {
	oldRow, err := getConfigByID(l.ctx, l.svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	name := in.Name
	if name == "" {
		name = oldRow.Name
	}
	exists, err := configNameExists(l.ctx, l.svcCtx, name, in.Id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("配置名称已存在")
	}
	code := in.Code
	if code == "" {
		code = oldRow.Code
	}
	updates := map[string]any{
		"name": name, "code": code, "data": in.DataJson,
		"remark":       sql.NullString{String: in.Remark, Valid: in.Remark != ""},
		"update_by":    sql.NullInt64{Int64: in.UpdateBy, Valid: in.UpdateBy != 0},
		"updated_time": time.Now(),
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SConfig{}).Where("id = ?", in.Id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
