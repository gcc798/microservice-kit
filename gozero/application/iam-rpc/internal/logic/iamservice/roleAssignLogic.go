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

type RoleAssignLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleAssignLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleAssignLogic {
	return &RoleAssignLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleAssignLogic) RoleAssign(in *pb.AssignRoleReq) (*pb.Ack, error) {
	if _, err := getUserByID(l.ctx, l.svcCtx, in.UserId); err != nil {
		return nil, err
	}
	if _, err := getRoleByID(l.ctx, l.svcCtx, in.RoleId); err != nil {
		return nil, err
	}
	totalCount, err := gorm.G[model.MUserRole](l.svcCtx.DB).Where("user_id = ? AND role_id = ?", in.UserId, in.RoleId).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if totalCount > 0 {
		return nil, errors.New("用户已拥有该角色")
	}
	now := sql.NullTime{Time: time.Now(), Valid: true}
	if err := gorm.G[model.MUserRole](l.svcCtx.DB).Create(l.ctx, &model.MUserRole{
		UserId: in.UserId, RoleId: in.RoleId, CreatedTime: now, UpdatedTime: now,
	}); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
