package resourceservicelogic

import (
	"context"
	"database/sql"
	"time"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentBindLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentBindLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentBindLogic {
	return &AttachmentBindLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentBindLogic) AttachmentBind(in *pb.AttachmentBindReq) (*pb.Ack, error) {
	row, err := getAttachment(l.ctx, l.svcCtx, in.AttachmentId)
	if err != nil {
		return nil, err
	}
	accessURL := ""
	if in.IsPublic {
		accessURL, _ = l.svcCtx.Storage.URL(l.ctx, row.FileKey, 0)
	}
	expires, err := parseTime(in.ExpireTime)
	if err != nil {
		return nil, err
	}
	updates := map[string]any{
		"business_type": in.BusinessType, "business_id": in.BusinessId, "business_field": in.BusinessField,
		"is_public": in.IsPublic, "access_url": sql.NullString{String: accessURL, Valid: accessURL != ""},
		"metadata": sql.NullString{String: in.MetadataJson, Valid: in.MetadataJson != ""}, "update_time": time.Now(),
	}
	if !expires.IsZero() {
		updates["expire_time"] = expires
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.BizAttachment{}).Where("id = ?", in.AttachmentId).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
