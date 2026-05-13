package handler

import (
	"net/http"

	"mall-api/internal/logic"
	"mall-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AuthLogoutHandler revokes the bearer token used for the request. Auth-only
// route — middleware has already pinned the access token into context.
func AuthLogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAuthLogoutLogic(r.Context(), svcCtx)
		resp, err := l.AuthLogout()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
