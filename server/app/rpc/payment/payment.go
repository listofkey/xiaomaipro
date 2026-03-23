package main

import (
	"flag"
	"fmt"

	"server/app/rpc/payment/internal/config"
	"server/app/rpc/payment/internal/server"
	"server/app/rpc/payment/internal/svc"
	"server/app/rpc/payment/paymentpb"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/payment.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		paymentpb.RegisterPaymentServiceServer(grpcServer, server.NewPaymentServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting payment rpc server at %s...\n", c.ListenOn)
	s.Start()
}
