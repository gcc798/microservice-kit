package role

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/zeromicro/go-zero/core/logx"
)

type RoleAssignLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleAssignLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleAssignLogic {
	return &RoleAssignLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *RoleAssignLogic) RoleAssign(req *types.AssignRoleReq) (resp *types.CommonResp, err error) {
	if _, err := l.svcCtx.IamRpcClient.RoleAssign(l.ctx, &iamservice.AssignRoleReq{UserId: req.UserId, RoleId: req.RoleId}); err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: "ok"}, nil
}
