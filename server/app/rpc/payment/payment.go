package main

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"server/app/rpc/payment/internal/config"
	"server/app/rpc/payment/internal/server"
	"server/app/rpc/payment/internal/svc"
	"server/app/rpc/payment/paymentpb"
	"server/pkg/logging"
)

var configFile = flag.String("f", "etc/payment.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logRuntime := logging.MustSetup(c.Name, c.Mode, c.Logging)
	defer func() {
		_ = logRuntime.Close()
	}()

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		paymentpb.RegisterPaymentServiceServer(grpcServer, server.NewPaymentServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logging.RebindLogx()
	logRuntime.Logger().Info("payment rpc starting",
		zap.String("listen_on", c.ListenOn),
	)
	s.Start()
}
