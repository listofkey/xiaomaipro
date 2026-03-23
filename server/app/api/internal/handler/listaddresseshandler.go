package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
)

func ListAddressesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListAddressesLogic(r.Context(), svcCtx)
		resp, err := l.ListAddresses()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
