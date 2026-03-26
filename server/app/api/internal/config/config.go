package config

import (
	"server/pkg/logging"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Logging    logging.Config
	UserRpc    zrpc.RpcClientConf
	ProgramRpc zrpc.RpcClientConf
	OrderRpc   zrpc.RpcClientConf
	PaymentRpc zrpc.RpcClientConf
}
