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

	PaymentRpc zrpc.RpcClientConf

	Pay struct {
		NotifySuccessResult string
		MockPayFormTemplate string
	}

	Kafka struct {
		Enabled bool
		Brokers []string
		Topic   string
		GroupID string
	}

	Order struct {
		PayTimeoutMinutes         int
		QueueStatusTTLMinutes     int
		InventoryTTLHours         int
		PurchaseCounterTTLHours   int
		RefundDeadlineHours       int
		AutoCancelIntervalSeconds int
		LocalAsyncBuffer          int
	}

	KeyPrefix string
}
