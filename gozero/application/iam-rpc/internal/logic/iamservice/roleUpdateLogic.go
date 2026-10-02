package iamservicelogic

import (
	"context"
	"database/sql"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleUpdateLogic {
	return &RoleUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleUpdateLogic) RoleUpdate(in *pb.RoleUpdateReq) (*pb.Ack, error) {
	row, err := getRoleByID(l.ctx, l.svcCtx, in.RoleId)
	if err != nil {
		return nil, err
	}
	roleName := row.RoleName
	if in.RoleName != "" {
		roleName = in.RoleName
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SRole{}).Where("id = ?", in.RoleId).Updates(map[string]any{
		"role_name": roleName, "sort": in.Sort, "status": in.Status, "data_scope": in.DataScope,
		"remark": sql.NullString{String: in.Remark, Valid: in.Remark != ""}, "updated_time": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
