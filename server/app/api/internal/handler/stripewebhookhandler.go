package handler

import (
	"io"
	"net/http"
	"strings"

	"server/app/api/internal/logic"
	"server/app/api/internal/svc"
)

func StripeWebhookHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		headers := make(map[string]string, len(r.Header))
		for key, values := range r.Header {
			if len(values) == 0 {
				continue
			}
			headers[key] = values[0]
		}

		l := logic.NewStripeWebhookLogic(r.Context(), svcCtx)
		ackText, err := l.StripeWebhook(rawBody, r.Header.Get("Stripe-Signature"), headers)
		if err != nil {
			statusCode := http.StatusInternalServerError
			errText := strings.ToLower(err.Error())
			if strings.Contains(errText, "signature") || strings.Contains(errText, "webhook body") {
				statusCode = http.StatusBadRequest
			}
			http.Error(w, "failure", statusCode)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(ackText))
	}
}
