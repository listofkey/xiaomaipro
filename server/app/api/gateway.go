package main

import (
	"flag"
	"fmt"

	"server/app/api/internal/config"
	"server/app/api/internal/handler"
	"server/app/api/internal/svc"
	"server/pkg/logging"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"go.uber.org/zap"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logRuntime := logging.MustSetup(c.Name, c.Mode, c.Logging)
	defer func() {
		_ = logRuntime.Close()
	}()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	logging.RebindLogx()
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	logRuntime.Logger().Info("gateway starting",
		zap.String("listen_on", fmt.Sprintf("%s:%d", c.Host, c.Port)),
	)
	server.Start()
}
