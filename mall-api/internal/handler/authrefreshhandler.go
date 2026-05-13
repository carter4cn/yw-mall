package handler

import (
	"net/http"

	"mall-api/internal/logic"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AuthRefreshHandler rotates a refresh_token into a new access/refresh pair.
// Public route — the refresh token itself is the credential.
func AuthRefreshHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuthRefreshReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewAuthRefreshLogic(r.Context(), svcCtx)
		resp, err := l.AuthRefresh(&req)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
