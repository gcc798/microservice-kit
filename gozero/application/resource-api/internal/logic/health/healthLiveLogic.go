// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package health

import (
	"context"

	"github.com/gcc798/microservice-kit/application/resource-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLiveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLiveLogic {
	return &HealthLiveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLiveLogic) HealthLive() (resp *types.CommonResp, err error) {
	return &types.CommonResp{
		Code: 200,
		Msg:  "success",
		Data: map[string]any{"status": "alive"},
	}, nil
}
