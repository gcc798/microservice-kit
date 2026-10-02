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
	row := model.SOperLog{
		Title: sql.NullString{String: in.Title, Valid: in.Title != ""}, BusinessType: sql.NullString{String: in.BusinessType, Valid: in.BusinessType != ""},
		Method: sql.NullString{String: in.Method, Valid: in.Method != ""}, RequestMethod: sql.NullString{String: in.RequestMethod, Valid: in.RequestMethod != ""},
		DeviceType: sql.NullString{String: in.DeviceType, Valid: in.DeviceType != ""}, OperName: sql.NullString{String: in.OperName, Valid: in.OperName != ""},
		OperUrl: sql.NullString{String: in.OperUrl, Valid: in.OperUrl != ""}, OperIp: sql.NullString{String: in.OperIp, Valid: in.OperIp != ""},
		OperLocation: sql.NullString{String: in.OperLocation, Valid: in.OperLocation != ""}, OperParam: sql.NullString{String: in.OperParam, Valid: in.OperParam != ""},
		JsonResult: sql.NullString{String: in.JsonResult, Valid: in.JsonResult != ""}, Status: sql.NullString{String: in.Status, Valid: in.Status != ""},
		ErrorMsg: sql.NullString{String: in.ErrorMsg, Valid: in.ErrorMsg != ""}, OperTime: sql.NullTime{Time: time.Now(), Valid: true},
		CostTime: sql.NullInt64{Int64: in.CostTime, Valid: in.CostTime > 0}, UserAgent: sql.NullString{String: in.UserAgent, Valid: in.UserAgent != ""},
	}
	if err := gorm.G[model.SOperLog](l.svcCtx.DB).Create(l.ctx, &row); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
