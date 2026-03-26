package config

import (
	"server/pkg/logging"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Logging logging.Config

	DB struct {
		DSN string
	}

	RedisConfig struct {
		Host     string
		Password string
		DB       int
	}
}
