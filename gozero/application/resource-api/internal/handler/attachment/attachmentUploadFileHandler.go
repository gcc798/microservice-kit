// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package attachment

import (
	"io"
	"net/http"

	"github.com/gcc798/microservice-kit/application/resource-api/internal/logic/attachment"
	"github.com/gcc798/microservice-kit/application/resource-api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AttachmentUploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := attachment.NewAttachmentUploadFileLogic(r.Context(), svcCtx)
		resp, err := l.AttachmentUploadFile(header.Filename, header.Header.Get("Content-Type"), content)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
