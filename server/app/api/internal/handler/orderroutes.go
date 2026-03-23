package handler

import (
	"net/http"

	"server/app/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func registerOrderRoutes(server *rest.Server, serverCtx *svc.ServiceContext, middlewares ...rest.Middleware) {
	server.AddRoutes(
		rest.WithMiddlewares(middlewares,
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/create",
				Handler: CreateOrderHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/queue-status",
				Handler: GetOrderQueueStatusHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/pay",
				Handler: PayOrderHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/cancel",
				Handler: CancelOrderHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/refund",
				Handler: ApplyRefundHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/list",
				Handler: ListOrderHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/detail",
				Handler: GetOrderDetailHandler(serverCtx),
			},
		),
		rest.WithPrefix("/order"),
	)
}
