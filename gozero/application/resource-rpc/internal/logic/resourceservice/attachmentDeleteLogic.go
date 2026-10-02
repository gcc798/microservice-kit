package resourceservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentDeleteLogic {
	return &AttachmentDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentDeleteLogic) AttachmentDelete(in *pb.IdReq) (*pb.Ack, error) {
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		row, err := getAttachment(l.ctx, l.svcCtx, in.Id)
		if err != nil {
			return err
		}
		_ = l.svcCtx.Storage.Delete(l.ctx, row.FileKey)
		return tx.Model(&model.BizAttachment{}).Where("id = ?", in.Id).Update("status", attachmentDeleted).Error
	})
	if err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
