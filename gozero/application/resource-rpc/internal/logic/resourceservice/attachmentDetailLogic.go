package resourceservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentDetailLogic {
	return &AttachmentDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentDetailLogic) AttachmentDetail(in *pb.IdReq) (*pb.Attachment, error) {
	row, err := getAttachment(l.ctx, l.svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	return toAttachmentPB(*row), nil
}
