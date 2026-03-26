package handler

import (
	"net/http"

	"server/app/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterWebsocketRoutes(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/connect",
				Handler: WebsocketHandler(serverCtx),
			},
		},
		rest.WithPrefix("/ws"),
	)
}
