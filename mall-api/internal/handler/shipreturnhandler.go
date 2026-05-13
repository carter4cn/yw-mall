package handler

import (
	"net/http"

	"mall-api/internal/logic"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ShipReturnHandler is the buyer-side endpoint POST /api/refund/:id/ship-return
// invoked after the merchant has approved a return_refund / exchange request.
func ShipReturnHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ShipReturnReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewShipReturnLogic(r.Context(), svcCtx)
		resp, err := l.ShipReturn(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
