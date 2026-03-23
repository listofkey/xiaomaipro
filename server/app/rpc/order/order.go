package main

import (
	"flag"
	"fmt"

	"server/app/rpc/order/internal/config"
	"server/app/rpc/order/internal/logic"
	"server/app/rpc/order/internal/server"
	"server/app/rpc/order/internal/svc"
	"server/app/rpc/order/orderpb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/order.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	logic.StartBackgroundWorkers(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		orderpb.RegisterOrderServiceServer(grpcServer, server.NewOrderServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting order rpc server at %s...\n", c.ListenOn)
	s.Start()
}
