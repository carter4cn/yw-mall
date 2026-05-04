package middleware

import (
	"net/http"

	"mall-common/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type AdminTokenMiddleware struct {
	Token string
}

func NewAdminTokenMiddleware(token string) *AdminTokenMiddleware {
	return &AdminTokenMiddleware{Token: token}
}

func (m *AdminTokenMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.Token == "" || r.Header.Get("X-Admin-Token") != m.Token {
			httpx.WriteJson(w, http.StatusUnauthorized, errorx.NewCodeError(errorx.AdminTokenInvalid))
			return
		}
		next(w, r)
	}
}
