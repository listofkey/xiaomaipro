package svc

import (
	"context"
	"time"

	"server/app/rpc/dao"
	"server/app/rpc/order/internal/config"
	"server/app/rpc/payment/paymentservice"
	"server/pkg/logging"
	"server/pkg/monitoring"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config          config.Config
	DB              *gorm.DB
	Query           *dao.Query
	Redis           *redis.Client
	PaymentRpc      paymentservice.PaymentService
	KafkaWriter     *kafka.Writer
	LocalAsyncQueue chan []byte
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(c.DB.DSN), &gorm.Config{
		Logger: logging.NewGormLogger("gorm"),
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	monitoring.StartDBMonitor("order", "postgres", db)

	dao.SetDefault(db)
	query := dao.Use(db)

	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisConfig.Host,
		Password: c.RedisConfig.Password,
		DB:       c.RedisConfig.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		logx.Errorf("redis connect failed: %v", err)
	} else {
		logx.Infof("redis connected: %s", c.RedisConfig.Host)
	}
	monitoring.InstrumentRedis("order", c.RedisConfig.Host, rdb)

	applyDefaults(&c)
	paymentRpc := paymentservice.NewPaymentService(zrpc.MustNewClient(c.PaymentRpc))

	var writer *kafka.Writer
	if c.Kafka.Enabled && len(c.Kafka.Brokers) > 0 && c.Kafka.Topic != "" {
		writer = &kafka.Writer{
			Addr:         kafka.TCP(c.Kafka.Brokers...),
			Topic:        c.Kafka.Topic,
			RequiredAcks: kafka.RequireOne,
			Async:        false,
			Balancer:     &kafka.LeastBytes{},
		}
	}
	for _, target := range c.Etcd.Hosts {
		monitoring.StartTCPMonitor("order", "etcd", target, 0)
	}
	for _, target := range c.PaymentRpc.Etcd.Hosts {
		monitoring.StartTCPMonitor("order", "etcd", target, 0)
	}
	if c.Kafka.Enabled {
		monitoring.StartKafkaBrokerMonitor("order", c.Kafka.Brokers, 0)
	}

	return &ServiceContext{
		Config:          c,
		DB:              db,
		Query:           query,
		Redis:           rdb,
		PaymentRpc:      paymentRpc,
		KafkaWriter:     writer,
		LocalAsyncQueue: make(chan []byte, c.Order.LocalAsyncBuffer),
	}
}

func applyDefaults(c *config.Config) {
	if c.KeyPrefix == "" {
		c.KeyPrefix = "xiaomaipro:order"
	}
	if c.Pay.NotifySuccessResult == "" {
		c.Pay.NotifySuccessResult = "success"
	}
	if c.Pay.MockPayFormTemplate == "" {
		c.Pay.MockPayFormTemplate = "mock://pay/{order_no}"
	}
	if c.Order.PayTimeoutMinutes <= 0 {
		c.Order.PayTimeoutMinutes = 15
	}
	if c.Order.QueueStatusTTLMinutes <= 0 {
		c.Order.QueueStatusTTLMinutes = 30
	}
	if c.Order.InventoryTTLHours <= 0 {
		c.Order.InventoryTTLHours = 48
	}
	if c.Order.PurchaseCounterTTLHours <= 0 {
		c.Order.PurchaseCounterTTLHours = 48
	}
	if c.Order.RefundDeadlineHours < 0 {
		c.Order.RefundDeadlineHours = 24
	}
	if c.Order.AutoCancelIntervalSeconds <= 0 {
		c.Order.AutoCancelIntervalSeconds = 30
	}
	if c.Order.LocalAsyncBuffer <= 0 {
		c.Order.LocalAsyncBuffer = 1024
	}
}

func (s *ServiceContext) PublishOrderMessage(ctx context.Context, payload []byte) error {
	start := time.Now()
	topic := s.Config.Kafka.Topic
	if topic == "" {
		topic = "local_async"
	}

	if s.KafkaWriter != nil {
		err := s.KafkaWriter.WriteMessages(ctx, kafka.Message{Value: payload})
		monitoring.RecordKafkaMessage("order", topic, "produce", monitoring.ResultFromError(err), time.Since(start))
		return err
	}

	select {
	case s.LocalAsyncQueue <- payload:
		monitoring.RecordKafkaMessage("order", topic, "produce", "success", time.Since(start))
		return nil
	case <-ctx.Done():
		err := ctx.Err()
		monitoring.RecordKafkaMessage("order", topic, "produce", monitoring.ResultFromError(err), time.Since(start))
		return err
	}
}
