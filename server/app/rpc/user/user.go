package main

import (
	"flag"

	"server/app/rpc/user/internal/config"
	"server/app/rpc/user/internal/server"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"
	"server/pkg/logging"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

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
		userpb.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logging.RebindLogx()
	logRuntime.Logger().Info("user rpc starting",
		zap.String("listen_on", c.ListenOn),
	)
	s.Start()
}
