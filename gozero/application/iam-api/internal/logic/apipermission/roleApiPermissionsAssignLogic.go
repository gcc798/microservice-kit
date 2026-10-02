// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package apipermission

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/logic/commonutil"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleApiPermissionsAssignLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleApiPermissionsAssignLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleApiPermissionsAssignLogic {
	return &RoleApiPermissionsAssignLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleApiPermissionsAssignLogic) RoleApiPermissionsAssign(req *types.RoleApiPermissionsAssignReq) (resp *types.CommonResp, err error) {
	if _, err := l.svcCtx.IamRpcClient.RoleApiPermissionsAssign(l.ctx, &iamservice.RoleApiPermissionsReq{
		RoleId:        req.RoleId,
		PermissionIds: req.PermissionIds,
		OperatorId:    commonutil.UserIDFromContext(l.ctx),
	}); err != nil {
		return failure(err), nil
	}
	return success("ok"), nil
}
