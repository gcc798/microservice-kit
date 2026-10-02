// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package attachment

import (
	"context"

	"github.com/gcc798/microservice-kit/application/resource-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-api/internal/types"
	"github.com/gcc798/microservice-kit/application/resource-rpc/client/resourceservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentDownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAttachmentDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentDownloadLogic {
	return &AttachmentDownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AttachmentDownloadLogic) AttachmentDownload(req *types.AttachmentIdPathReq) (*resourceservice.AttachmentDownloadResp, error) {
	return l.svcCtx.ResourceRpcClient.AttachmentDownload(l.ctx, &resourceservice.AttachmentDownloadReq{AttachmentId: req.AttachmentId})
}
