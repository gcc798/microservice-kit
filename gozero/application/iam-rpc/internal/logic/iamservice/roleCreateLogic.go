package iamservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RoleCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleCreateLogic {
	return &RoleCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleCreateLogic) RoleCreate(in *pb.RoleCreateReq) (*pb.Ack, error) {
	if in.RoleKey == "" || in.RoleName == "" {
		return nil, errors.New("角色标识和角色名称不能为空")
	}
	count, err := gorm.G[model.SRole](l.svcCtx.DB).Where("role_key = ?", in.RoleKey).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("角色标识已存在")
	}
	now := time.Now()
	if err := gorm.G[model.SRole](l.svcCtx.DB).Create(l.ctx, &model.SRole{
		RoleKey: in.RoleKey, RoleName: in.RoleName, Sort: in.Sort, Status: int64(in.Status),
		DataScope: int64(in.DataScope), Remark: sql.NullString{String: in.Remark, Valid: in.Remark != ""},
		CreatedTime: sql.NullTime{Time: now, Valid: true}, UpdatedTime: sql.NullTime{Time: now, Valid: true},
	}); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
