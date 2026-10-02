// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package health

import (
	"context"

	"github.com/gcc798/microservice-kit/application/resource-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-api/internal/types"
	"github.com/gcc798/microservice-kit/application/resource-rpc/client/resourceservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLogic) Health() (resp *types.CommonResp, err error) {
	pingResp, pingErr := l.svcCtx.ResourceRpcClient.Ping(l.ctx, &resourceservice.PingReq{})
	if pingErr != nil {
		return nil, pingErr
	}

	return &types.CommonResp{
		Code: 200,
		Msg:  "success",
		Data: map[string]any{
			"status": "ok",
			"rpc":    pingResp.Message,
		},
	}, nil
}
