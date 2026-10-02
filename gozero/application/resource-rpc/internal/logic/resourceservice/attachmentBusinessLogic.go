package resourceservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentBusinessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentBusinessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentBusinessLogic {
	return &AttachmentBusinessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentBusinessLogic) AttachmentBusiness(in *pb.AttachmentBusinessQueryReq) (*pb.AttachmentListResp, error) {
	rows, err := gorm.G[model.BizAttachment](l.svcCtx.DB).
		Where("business_type = ? AND business_id = ? AND status = ?", in.BusinessType, in.BusinessId, attachmentNormal).
		Order("create_time DESC").Find(l.ctx)
	if err != nil {
		return nil, err
	}
	return &pb.AttachmentListResp{Records: toAttachmentList(rows)}, nil
}
