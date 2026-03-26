package main

import (
	"flag"

	"server/app/rpc/order/internal/config"
	"server/app/rpc/order/internal/logic"
	"server/app/rpc/order/internal/server"
	"server/app/rpc/order/internal/svc"
	"server/app/rpc/order/orderpb"
	"server/pkg/logging"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/order.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logRuntime := logging.MustSetup(c.Name, c.Mode, c.Logging)
	defer func() {
		_ = logRuntime.Close()
	}()

	ctx := svc.NewServiceContext(c)
	logic.StartBackgroundWorkers(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		orderpb.RegisterOrderServiceServer(grpcServer, server.NewOrderServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logging.RebindLogx()
	logRuntime.Logger().Info("order rpc starting",
		zap.String("listen_on", c.ListenOn),
	)
	s.Start()
}
