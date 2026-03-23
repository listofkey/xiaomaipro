package handler

import (
	"net/http"

	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
)

func AlipayNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "failure", http.StatusBadRequest)
			return
		}

		params := make(map[string]string, len(r.Form))
		for key, values := range r.Form {
			if len(values) == 0 {
				continue
			}
			params[key] = values[0]
		}

		l := logic.NewAlipayNotifyLogic(r.Context(), svcCtx)
		ackText, err := l.AlipayNotify(params)
		if err != nil {
			http.Error(w, "failure", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(ackText))
	}
}
