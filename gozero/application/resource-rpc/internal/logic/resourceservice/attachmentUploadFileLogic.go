package resourceservicelogic

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentUploadFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentUploadFileLogic {
	return &AttachmentUploadFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentUploadFileLogic) AttachmentUploadFile(in *pb.AttachmentUploadReq) (*pb.Attachment, error) {
	if in.FileName == "" || len(in.Content) == 0 {
		return nil, errors.New("文件不能为空")
	}
	key := storageKey(in.FileName)
	if err := l.svcCtx.Storage.Upload(l.ctx, key, in.ContentType, bytes.NewReader(in.Content), int64(len(in.Content))); err != nil {
		return nil, err
	}
	now := time.Now()
	row := model.BizAttachment{
		FileName: in.FileName, FileKey: key, FileSize: int64(len(in.Content)),
		FileType: sql.NullString{String: in.ContentType, Valid: in.ContentType != ""},
		FileExt:  sql.NullString{String: strings.ToLower(strings.TrimPrefix(filepath.Ext(in.FileName), ".")), Valid: filepath.Ext(in.FileName) != ""},
		Status:   attachmentNormal, CreateTime: sql.NullTime{Time: now, Valid: true}, UpdateTime: sql.NullTime{Time: now, Valid: true},
	}
	if err := gorm.G[model.BizAttachment](l.svcCtx.DB).Create(l.ctx, &row); err != nil {
		_ = l.svcCtx.Storage.Delete(l.ctx, key)
		return nil, err
	}
	return toAttachmentPB(row), nil
}
