package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"mall-api/internal/logic"
	"mall-api/internal/svc"
)

func Kuaidi100WebhookHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		param := r.FormValue("param")
		sign := r.FormValue("sign")
		l := logic.NewKuaidi100WebhookLogic(r.Context(), svcCtx)
		resp, err := l.Process(param, sign)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
