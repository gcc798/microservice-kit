package iamservicelogic

import (
	"context"
	"errors"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleRemoveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleRemoveLogic {
	return &RoleRemoveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleRemoveLogic) RoleRemove(in *pb.RemoveRoleReq) (*pb.Ack, error) {
	if _, err := getRoleByID(l.ctx, l.svcCtx, in.RoleId); err != nil {
		return nil, err
	}
	result := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ? AND role_id = ?", in.UserId, in.RoleId).Delete(&model.MUserRole{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("用户角色关系不存在")
	}
	return &pb.Ack{Msg: "ok"}, nil
}
