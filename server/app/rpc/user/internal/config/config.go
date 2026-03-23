package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	DB struct {
		DSN string
	}

	RedisConfig struct {
		Host     string
		Password string
		DB       int
	}

	JWT struct {
		AccessSecret  string
		AccessExpire  int64
		RefreshExpire int64
	}

	AES struct {
		Key string
	}
}
