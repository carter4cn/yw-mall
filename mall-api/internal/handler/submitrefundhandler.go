package handler

import (
	"net/http"

	"mall-api/internal/logic"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SubmitRefundHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitRefundReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewSubmitRefundLogic(r.Context(), svcCtx)
		resp, err := l.SubmitRefund(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
