package handler

import (
	"net/http"

	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewListOrderLogic(r.Context(), svcCtx)
		resp, err := l.ListOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
