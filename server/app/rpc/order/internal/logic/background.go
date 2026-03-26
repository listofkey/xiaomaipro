package logic

import (
	"context"
	"time"

	"server/app/rpc/order/internal/svc"
	"server/pkg/monitoring"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

func StartBackgroundWorkers(svcCtx *svc.ServiceContext) {
	if svcCtx == nil {
		return
	}

	if svcCtx.Config.Kafka.Enabled && len(svcCtx.Config.Kafka.Brokers) > 0 && svcCtx.Config.Kafka.Topic != "" {
		go consumeKafkaLoop(svcCtx)
	} else {
		go consumeLocalLoop(svcCtx)
	}

	go autoCancelLoop(svcCtx)
}

func consumeKafkaLoop(svcCtx *svc.ServiceContext) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  svcCtx.Config.Kafka.Brokers,
		Topic:    svcCtx.Config.Kafka.Topic,
		GroupID:  svcCtx.Config.Kafka.GroupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	for {
		start := time.Now()
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			monitoring.RecordBackgroundJob("order", "kafka_read", "error", time.Since(start))
			logx.Errorf("order kafka consume failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		core := NewOrderCore(context.Background(), svcCtx)
		processStart := time.Now()
		err = core.processQueuedOrderMessage(msg.Value)
		monitoring.RecordKafkaMessage("order", svcCtx.Config.Kafka.Topic, "consume", monitoring.ResultFromError(err), time.Since(processStart))
		monitoring.RecordBackgroundJob("order", "kafka_process", monitoring.ResultFromError(err), time.Since(processStart))
		if err != nil {
			logx.Errorf("order kafka process failed: %v", err)
		}
	}
}

func consumeLocalLoop(svcCtx *svc.ServiceContext) {
	for payload := range svcCtx.LocalAsyncQueue {
		core := NewOrderCore(context.Background(), svcCtx)
		start := time.Now()
		err := core.processQueuedOrderMessage(payload)
		monitoring.RecordKafkaMessage("order", "local_async", "consume", monitoring.ResultFromError(err), time.Since(start))
		monitoring.RecordBackgroundJob("order", "local_async_process", monitoring.ResultFromError(err), time.Since(start))
		if err != nil {
			logx.Errorf("order local async process failed: %v", err)
		}
	}
}

func autoCancelLoop(svcCtx *svc.ServiceContext) {
	interval := time.Duration(svcCtx.Config.Order.AutoCancelIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		core := NewOrderCore(context.Background(), svcCtx)
		start := time.Now()
		count, err := core.expirePendingOrders(100)
		monitoring.RecordBackgroundJob("order", "expire_pending_orders", monitoring.ResultFromError(err), time.Since(start))
		if err != nil {
			logx.Errorf("expire pending orders failed: %v", err)
			continue
		}
		if count > 0 {
			logx.Infof("expired pending orders: %d", count)
		}
	}
}
