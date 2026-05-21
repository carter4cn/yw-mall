// mall-api/internal/handler/registerv2handler.go
package handler

import (
	"net/http"

	"mall-api/internal/logic"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterV2Handler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterV2Req
		if err := httpx.Parse(r, &req); err != nil {
			writeBizError(r.Context(), w, err)
			return
		}
		resp, err := logic.RegisterV2(r.Context(), svcCtx, &req)
		if err != nil {
			writeBizError(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
