package main

import (
	"flag"

	"server/app/rpc/program/internal/config"
	"server/app/rpc/program/internal/server"
	"server/app/rpc/program/internal/svc"
	"server/app/rpc/program/programpb/programpb"
	"server/pkg/logging"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/program.yaml", "the config file")

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
		programpb.RegisterProgramServiceServer(grpcServer, server.NewProgramServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logging.RebindLogx()
	logRuntime.Logger().Info("program rpc starting",
		zap.String("listen_on", c.ListenOn),
	)
	s.Start()
}
