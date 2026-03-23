package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
)

func UpdateTicketBuyerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateTicketBuyerReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUpdateTicketBuyerLogic(r.Context(), svcCtx)
		resp, err := l.UpdateTicketBuyer(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
