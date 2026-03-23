package handler

import (
	"net/http"

	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetEventDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetEventDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetEventDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetEventDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
