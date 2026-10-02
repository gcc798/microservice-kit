package resourceservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentPageLogic {
	return &AttachmentPageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentPageLogic) AttachmentPage(in *pb.AttachmentPageReq) (*pb.AttachmentPageResp, error) {
	pageNum, pageSize := pageValues(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.BizAttachment{}).Where("status = ?", attachmentNormal)
	if in.FileName != "" {
		query = query.Where("file_name LIKE ?", "%"+in.FileName+"%")
	}
	if in.FileType != "" {
		query = query.Where("file_type = ?", in.FileType)
	}
	if in.BusinessType != "" {
		query = query.Where("business_type = ?", in.BusinessType)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.BizAttachment
	if err := query.Order("create_time DESC, id DESC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.AttachmentPageResp{Records: toAttachmentList(rows), Page: pageInfo(total, pageNum, pageSize)}, nil
}
