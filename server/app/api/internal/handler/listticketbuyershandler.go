package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
)

func ListTicketBuyersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListTicketBuyersLogic(r.Context(), svcCtx)
		resp, err := l.ListTicketBuyers()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
