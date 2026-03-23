package main

import (
	"flag"
	"fmt"

	"server/app/rpc/program/internal/config"
	"server/app/rpc/program/internal/server"
	"server/app/rpc/program/internal/svc"
	"server/app/rpc/program/programpb/programpb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/program.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		programpb.RegisterProgramServiceServer(grpcServer, server.NewProgramServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting program rpc server at %s...\n", c.ListenOn)
	s.Start()
}
