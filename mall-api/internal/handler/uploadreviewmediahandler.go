// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"mall-api/internal/logic"
	"mall-api/internal/svc"
)

func UploadReviewMediaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		fhs := r.MultipartForm.File["file"]
		parts := make([]logic.UploadPart, 0, len(fhs))
		for _, fh := range fhs {
			f, err := fh.Open()
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			defer f.Close()
			parts = append(parts, logic.UploadPart{
				Filename: fh.Filename,
				Reader:   f,
				Size:     fh.Size,
				MIME:     fh.Header.Get("Content-Type"),
			})
		}
		l := logic.NewUploadReviewMediaLogic(r.Context(), svcCtx)
		resp, err := l.Upload(parts)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
