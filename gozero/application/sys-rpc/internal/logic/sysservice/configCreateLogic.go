package sysservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfigCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigCreateLogic {
	return &ConfigCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfigCreateLogic) ConfigCreate(in *pb.ConfigCreateReq) (*pb.Ack, error) {
	if in.Name == "" || in.Code == "" {
		return nil, errors.New("名称和编码不能为空")
	}
	exists, err := configNameExists(l.ctx, l.svcCtx, in.Name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("配置名称已存在")
	}
	now := time.Now()
	row := model.SConfig{
		Name: in.Name, Code: in.Code, Data: in.DataJson,
		Remark:      sql.NullString{String: in.Remark, Valid: in.Remark != ""},
		CreateBy:    sql.NullInt64{Int64: in.CreateBy, Valid: in.CreateBy != 0},
		UpdateBy:    sql.NullInt64{Int64: in.UpdateBy, Valid: in.UpdateBy != 0},
		CreatedTime: sql.NullTime{Time: now, Valid: true}, UpdatedTime: sql.NullTime{Time: now, Valid: true},
	}
	if err := gorm.G[model.SConfig](l.svcCtx.DB).Create(l.ctx, &row); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
