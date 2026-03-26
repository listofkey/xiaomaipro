package handler

import (
	"net/http"

	"server/app/api/internal/svc"
)

func WebsocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svcCtx.WebsocketHub == nil {
			http.Error(w, "websocket hub is not configured", http.StatusServiceUnavailable)
			return
		}

		svcCtx.WebsocketHub.ServeHTTP(w, r)
	}
}
