package svc

import (
	"context"
	"time"

	"server/app/rpc/payment/internal/config"
	"server/app/rpc/payment/internal/pay"
	"server/pkg/logging"
	"server/pkg/monitoring"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	Redis         *redis.Client
	PayStrategies *pay.StrategyContext
}

func NewServiceContext(c config.Config) *ServiceContext {
	normalizeConfig(&c)

	db, err := gorm.Open(postgres.Open(c.DB.DSN), &gorm.Config{
		Logger: logging.NewGormLogger("gorm"),
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	monitoring.StartDBMonitor("payment", "postgres", db)

	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisConfig.Host,
		Password: c.RedisConfig.Password,
		DB:       c.RedisConfig.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		logx.Errorf("redis connect failed: %v", err)
	}
	monitoring.InstrumentRedis("payment", c.RedisConfig.Host, rdb)
	for _, target := range c.Etcd.Hosts {
		monitoring.StartTCPMonitor("payment", "etcd", target, 0)
	}

	return &ServiceContext{
		Config:        c,
		DB:            db,
		Redis:         rdb,
		PayStrategies: initStrategyContext(c),
	}
}

func initStrategyContext(c config.Config) *pay.StrategyContext {
	strategies := make([]pay.Strategy, 0, 1)
	if c.Stripe.SecretKey != "" {
		strategy, err := pay.NewStripeStrategy(pay.StripeConfig{
			SecretKey:             c.Stripe.SecretKey,
			WebhookSecret:         c.Stripe.WebhookSecret,
			SuccessURL:            c.Stripe.SuccessURL,
			CancelURL:             c.Stripe.CancelURL,
			Currency:              c.Stripe.Currency,
			PaymentMethodTypes:    c.Stripe.PaymentMethodTypes,
			CheckoutExpireMinutes: c.Stripe.CheckoutExpireMinutes,
			RequestTimeoutSeconds: c.Stripe.RequestTimeoutSeconds,
		})
		if err != nil {
			logx.Errorf("init stripe strategy failed: %v", err)
		} else {
			strategies = append(strategies, strategy)
		}
	} else {
		logx.Infof("stripe strategy not configured, secret key is empty")
	}
	return pay.NewStrategyContext(strategies...)
}

func normalizeConfig(c *config.Config) {
	if c.Pay.NotifySuccessResult == "" {
		c.Pay.NotifySuccessResult = "success"
	}
	if c.Pay.NotifyFailureResult == "" {
		c.Pay.NotifyFailureResult = "failure"
	}
	if c.Pay.DefaultPayScene == "" {
		c.Pay.DefaultPayScene = "order"
	}
	if c.Pay.DefaultChannel == "" {
		c.Pay.DefaultChannel = pay.ChannelStripe
	}
	if c.Business.RefundDeadlineHours <= 0 {
		c.Business.RefundDeadlineHours = 24
	}
	if c.Stripe.Currency == "" {
		c.Stripe.Currency = "cny"
	}
	if len(c.Stripe.PaymentMethodTypes) == 0 {
		c.Stripe.PaymentMethodTypes = []string{"card"}
	}
	if c.Stripe.CheckoutExpireMinutes <= 0 {
		c.Stripe.CheckoutExpireMinutes = 30
	}
	if c.Stripe.RequestTimeoutSeconds <= 0 {
		c.Stripe.RequestTimeoutSeconds = 45
	}
	if c.Lock.TTLSeconds <= 0 {
		c.Lock.TTLSeconds = 5
	}
	if c.Lock.RetryTimes <= 0 {
		c.Lock.RetryTimes = 20
	}
	if c.Lock.RetryIntervalMillis <= 0 {
		c.Lock.RetryIntervalMillis = 50
	}
	if c.KeyPrefix == "" {
		c.KeyPrefix = "xiaomaipro:payment"
	}
}
