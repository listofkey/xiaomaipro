package handler

import (
	"net/http"

	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SearchEventsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchEventsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewSearchEventsLogic(r.Context(), svcCtx)
		resp, err := l.SearchEvents(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
