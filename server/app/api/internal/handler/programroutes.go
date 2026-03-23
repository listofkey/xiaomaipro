package handler

import (
	"net/http"

	"server/app/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func registerProgramRoutes(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/events",
				Handler: ListEventsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/events/search",
				Handler: SearchEventsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/events/detail",
				Handler: GetEventDetailHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/categories",
				Handler: ListCategoriesHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/hot-recommend",
				Handler: GetHotRecommendHandler(serverCtx),
			},
		},
		rest.WithPrefix("/program"),
	)
}
