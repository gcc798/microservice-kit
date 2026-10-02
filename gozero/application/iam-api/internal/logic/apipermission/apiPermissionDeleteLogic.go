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

type ApiPermissionDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApiPermissionDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiPermissionDeleteLogic {
	return &ApiPermissionDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApiPermissionDeleteLogic) ApiPermissionDelete(req *types.IdPathReq) (resp *types.CommonResp, err error) {
	if _, err := l.svcCtx.IamRpcClient.ApiPermissionDelete(l.ctx, &iamservice.IdReq{Id: req.Id}); err != nil {
		return failure(err), nil
	}
	return success("ok"), nil
}
