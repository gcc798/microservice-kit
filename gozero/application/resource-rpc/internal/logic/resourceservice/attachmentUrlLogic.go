package resourceservicelogic

import (
	"context"
	"time"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentUrlLogic {
	return &AttachmentUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentUrlLogic) AttachmentUrl(in *pb.AttachmentUrlQueryReq) (*pb.AttachmentUrlResp, error) {
	row, err := getAttachment(l.ctx, l.svcCtx, in.AttachmentId)
	if err != nil {
		return nil, err
	}
	expires := in.Expires
	if expires <= 0 {
		expires = 3600
	}
	url := row.AccessUrl.String
	if url == "" {
		url, err = l.svcCtx.Storage.URL(l.ctx, row.FileKey, time.Duration(expires)*time.Second)
		if err != nil {
			return nil, err
		}
	}
	return &pb.AttachmentUrlResp{AttachmentId: row.Id, Url: url, Expires: expires}, nil
}
