package resourceservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/resource-rpc/pb"
	"gorm.io/gorm"
)

const (
	attachmentNormal  = 0
	attachmentDeleted = 1
)

func getAttachment(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*model.BizAttachment, error) {
	row, err := gorm.G[model.BizAttachment](svcCtx.DB).
		Where("id = ? AND status = ?", id, attachmentNormal).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("附件不存在")
	}
	return &row, err
}

func toAttachmentPB(row model.BizAttachment) *pb.Attachment {
	return &pb.Attachment{
		AttachmentId: row.Id, FileName: row.FileName, FileKey: row.FileKey, FileSize: row.FileSize,
		FileType: row.FileType.String, FileExt: row.FileExt.String, BusinessType: row.BusinessType.String,
		BusinessId: row.BusinessId.String, BusinessField: row.BusinessField.String, IsPublic: row.IsPublic,
		AccessUrl: row.AccessUrl.String, MetadataJson: row.Metadata.String, Status: int32(row.Status),
		ExpireTime: formatTime(row.ExpireTime), CreateBy: valueInt64(row.CreateBy),
		CreateTime: formatTime(row.CreateTime), UpdateTime: formatTime(row.UpdateTime),
	}
}

func toAttachmentList(rows []model.BizAttachment) []*pb.Attachment {
	result := make([]*pb.Attachment, 0, len(rows))
	for _, row := range rows {
		result = append(result, toAttachmentPB(row))
	}
	return result
}

func formatTime(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format(time.RFC3339)
	}
	return ""
}

func valueInt64(value sql.NullInt64) int64 {
	if value.Valid {
		return value.Int64
	}
	return 0
}

func pageValues(pageNum, pageSize int64) (int64, int64) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageNum, pageSize
}

func pageInfo(total, pageNum, pageSize int64) *pb.PageInfo {
	return &pb.PageInfo{Total: total, Size: pageSize, Current: pageNum, Pages: (total + pageSize - 1) / pageSize}
}

func storageKey(filename string) string {
	return fmt.Sprintf("temp/%s/%d_%s", time.Now().Format("20060102"), time.Now().UnixNano(), filepath.Base(filename))
}

func downloadContent(ctx context.Context, svcCtx *svc.ServiceContext, row *model.BizAttachment) ([]byte, string, error) {
	reader, err := svcCtx.Storage.Download(ctx, row.FileKey)
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}
	contentType := row.FileType.String
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	return content, contentType, nil
}

func parseTime(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}
