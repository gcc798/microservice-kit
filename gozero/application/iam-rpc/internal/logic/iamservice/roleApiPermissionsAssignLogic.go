package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleApiPermissionsAssignLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleApiPermissionsAssignLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleApiPermissionsAssignLogic {
	return &RoleApiPermissionsAssignLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleApiPermissionsAssignLogic) RoleApiPermissionsAssign(in *pb.RoleApiPermissionsReq) (*pb.Ack, error) {
	if _, err := getRoleByID(l.ctx, l.svcCtx, in.RoleId); err != nil {
		return nil, err
	}
	if err := replaceRoleApiPermissions(l.ctx, l.svcCtx, in.RoleId, in.PermissionIds, in.OperatorId); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
