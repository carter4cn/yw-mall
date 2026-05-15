package handler

import (
	"net/http"

	"mall-api/internal/logic"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AuthLoginHandler is the P0 login revamp endpoint. Issues an opaque access +
// refresh token pair from user-rpc and ships full session info to the client.
//
// S4.2: stashes the client IP into ctx so the lockout/whitelist logic can pick
// it up without taking *http.Request as a parameter.
func AuthLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuthLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		ctx := logic.WithIP(r.Context(), logic.ClientIP(r))
		l := logic.NewAuthLoginLogic(ctx, svcCtx)
		resp, err := l.AuthLogin(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
