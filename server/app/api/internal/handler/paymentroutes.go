package handler

import (
	"net/http"

	"server/app/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func registerPaymentRoutes(server *rest.Server, serverCtx *svc.ServiceContext, middlewares ...rest.Middleware) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/stripe/webhook",
				Handler: StripeWebhookHandler(serverCtx),
			},
		},
		rest.WithPrefix("/payment"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(middlewares,
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/detail",
				Handler: PaymentDetailHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/check",
				Handler: PaymentTradeCheckHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/status",
				Handler: PaymentTradeCheckHandler(serverCtx),
			},
		),
		rest.WithPrefix("/payment"),
	)
}
