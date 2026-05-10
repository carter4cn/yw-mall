package middleware

import (
	"net/http"

	"mall-common/configcenter"
	"mall-common/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AdminTokenMiddleware validates the X-Admin-Token header.
// token is hot-reloadable: changing the etcd config takes effect immediately
// without restarting the server.
type AdminTokenMiddleware struct {
	token *configcenter.HotConfig[string]
}

func NewAdminTokenMiddleware(token *configcenter.HotConfig[string]) *AdminTokenMiddleware {
	return &AdminTokenMiddleware{token: token}
}

func (m *AdminTokenMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := m.token.Get()
		if t == "" || r.Header.Get("X-Admin-Token") != t {
			httpx.WriteJson(w, http.StatusUnauthorized, errorx.NewCodeError(errorx.AdminTokenInvalid))
			return
		}
		next(w, r)
	}
}
