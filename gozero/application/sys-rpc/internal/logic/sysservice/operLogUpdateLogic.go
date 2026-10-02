package sysservicelogic

import (
	"context"
	"database/sql"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperLogUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperLogUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperLogUpdateLogic {
	return &OperLogUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperLogUpdateLogic) OperLogUpdate(in *pb.OperLogUpdateReq) (*pb.Ack, error) {
	updates := map[string]any{
		"title": sql.NullString{String: in.Title, Valid: in.Title != ""}, "business_type": sql.NullString{String: in.BusinessType, Valid: in.BusinessType != ""},
		"method": sql.NullString{String: in.Method, Valid: in.Method != ""}, "request_method": sql.NullString{String: in.RequestMethod, Valid: in.RequestMethod != ""},
		"device_type": sql.NullString{String: in.DeviceType, Valid: in.DeviceType != ""}, "oper_name": sql.NullString{String: in.OperName, Valid: in.OperName != ""},
		"oper_url": sql.NullString{String: in.OperUrl, Valid: in.OperUrl != ""}, "oper_ip": sql.NullString{String: in.OperIp, Valid: in.OperIp != ""},
		"oper_location": sql.NullString{String: in.OperLocation, Valid: in.OperLocation != ""}, "oper_param": sql.NullString{String: in.OperParam, Valid: in.OperParam != ""},
		"json_result": sql.NullString{String: in.JsonResult, Valid: in.JsonResult != ""}, "status": sql.NullString{String: in.Status, Valid: in.Status != ""},
		"error_msg": sql.NullString{String: in.ErrorMsg, Valid: in.ErrorMsg != ""}, "cost_time": sql.NullInt64{Int64: in.CostTime, Valid: in.CostTime > 0},
		"user_agent": sql.NullString{String: in.UserAgent, Valid: in.UserAgent != ""},
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SOperLog{}).Where("id = ?", in.Id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
