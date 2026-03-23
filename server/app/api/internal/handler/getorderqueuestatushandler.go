package handler

import (
	"net/http"

	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetOrderQueueStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrderQueueStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetOrderQueueStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetOrderQueueStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
