// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package apipermission

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleApiPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleApiPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleApiPermissionsLogic {
	return &RoleApiPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleApiPermissionsLogic) RoleApiPermissions(req *types.RoleApiPermissionsPathReq) (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.IamRpcClient.RoleApiPermissions(l.ctx, &iamservice.IdReq{Id: req.RoleId})
	if err != nil {
		return failure(err), nil
	}
	return success(data.PermissionIds), nil
}
