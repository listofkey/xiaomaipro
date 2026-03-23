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

	Pay struct {
		NotifySuccessResult string
		NotifyFailureResult string
		DefaultPayScene     string
		DefaultChannel      string
	}

	Business struct {
		RefundDeadlineHours int
	}

	Stripe struct {
		SecretKey             string
		WebhookSecret         string
		SuccessURL            string
		CancelURL             string
		Currency              string
		PaymentMethodTypes    []string
		CheckoutExpireMinutes int
		RequestTimeoutSeconds int
	}

	Lock struct {
		TTLSeconds          int
		RetryTimes          int
		RetryIntervalMillis int
	}

	KeyPrefix string
}
