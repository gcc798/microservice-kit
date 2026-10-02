package iamservicelogic

import (
	"context"
	"fmt"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleRemoveUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleRemoveUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleRemoveUsersLogic {
	return &RoleRemoveUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleRemoveUsersLogic) RoleRemoveUsers(in *pb.RoleUsersReq) (*pb.Ack, error) {
	if in.RoleId <= 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}
	if _, err := getRoleByID(l.ctx, l.svcCtx, in.RoleId); err != nil {
		return nil, err
	}
	userIDs := uniqueInt64Values(in.UserIds)
	if len(userIDs) == 0 {
		return &pb.Ack{Msg: "ok"}, nil
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("role_id = ? AND user_id IN ?", in.RoleId, userIDs).Delete(&model.MUserRole{}).Error; err != nil {
		return nil, fmt.Errorf("批量移除角色用户失败: %w", err)
	}
	return &pb.Ack{Msg: "ok"}, nil
}
