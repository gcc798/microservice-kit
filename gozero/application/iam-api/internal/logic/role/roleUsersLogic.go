// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleUsersLogic {
	return &RoleUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleUsersLogic) RoleUsers(req *types.RoleUsersPathReq) (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.IamRpcClient.RoleUsers(l.ctx, &iamservice.RoleUsersReq{RoleId: req.RoleId})
	if err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: data.Records}, nil
}
