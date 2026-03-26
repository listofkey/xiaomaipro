package monitoring

import (
	"context"
	"database/sql"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

const (
	namespace         = "xiaomaipro"
	defaultInterval   = 15 * time.Second
	dependencyTimeout = 3 * time.Second
)

var (
	operationTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "business",
		Name:      "operation_total",
		Help:      "Business operation executions grouped by service, operation and result.",
	}, []string{"service", "operation", "result"})

	operationDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "business",
		Name:      "operation_duration_seconds",
		Help:      "Business operation duration grouped by service and operation.",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"service", "operation"})

	dependencyUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "dependency",
		Name:      "up",
		Help:      "Dependency reachability grouped by service, dependency and target.",
	}, []string{"service", "dependency", "target"})

	dbConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "db",
		Name:      "connections",
		Help:      "Database connection pool status grouped by service and state.",
	}, []string{"service", "state"})

	dbWaitCountTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "db",
		Name:      "wait_count_total",
		Help:      "Database wait count deltas grouped by service.",
	}, []string{"service"})

	dbWaitDurationSecondsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "db",
		Name:      "wait_duration_seconds_total",
		Help:      "Database wait duration deltas grouped by service.",
	}, []string{"service"})

	dbClosedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "db",
		Name:      "closed_total",
		Help:      "Database pool close counters grouped by service and reason.",
	}, []string{"service", "reason"})

	redisCommandTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "redis",
		Name:      "command_total",
		Help:      "Redis command executions grouped by service, command and result.",
	}, []string{"service", "command", "result"})

	redisCommandDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "redis",
		Name:      "command_duration_seconds",
		Help:      "Redis command duration grouped by service and command.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"service", "command"})

	kafkaMessageTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "kafka",
		Name:      "message_total",
		Help:      "Kafka message operations grouped by service, topic, direction and result.",
	}, []string{"service", "topic", "direction", "result"})

	kafkaMessageDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "kafka",
		Name:      "message_duration_seconds",
		Help:      "Kafka message operation duration grouped by service, topic and direction.",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"service", "topic", "direction"})

	websocketConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "websocket",
		Name:      "connections",
		Help:      "Current websocket connections grouped by service.",
	}, []string{"service"})

	websocketMessagesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "websocket",
		Name:      "messages_total",
		Help:      "Websocket message operations grouped by service, direction and result.",
	}, []string{"service", "direction", "result"})

	cacheRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "cache",
		Name:      "request_total",
		Help:      "Cache requests grouped by service, cache and result.",
	}, []string{"service", "cache", "result"})

	backgroundJobTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "background",
		Name:      "job_total",
		Help:      "Background job executions grouped by service, job and result.",
	}, []string{"service", "job", "result"})

	backgroundJobDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "background",
		Name:      "job_duration_seconds",
		Help:      "Background job duration grouped by service and job.",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"service", "job"})

	dependencyLoops sync.Map
	redisHooks      sync.Map
	dbLoops         sync.Map
)

func init() {
	prometheus.MustRegister(
		operationTotal,
		operationDuration,
		dependencyUp,
		dbConnections,
		dbWaitCountTotal,
		dbWaitDurationSecondsTotal,
		dbClosedTotal,
		redisCommandTotal,
		redisCommandDuration,
		kafkaMessageTotal,
		kafkaMessageDuration,
		websocketConnections,
		websocketMessagesTotal,
		cacheRequestTotal,
		backgroundJobTotal,
		backgroundJobDuration,
	)
}

func RecordOperation(service, operation, result string, duration time.Duration) {
	service = sanitizeLabel(service)
	operation = sanitizeLabel(operation)
	result = sanitizeLabel(result)

	operationTotal.WithLabelValues(service, operation, result).Inc()
	operationDuration.WithLabelValues(service, operation).Observe(duration.Seconds())
}

func ResultFromError(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

func RecordCacheRequest(service, cache, result string) {
	cacheRequestTotal.WithLabelValues(
		sanitizeLabel(service),
		sanitizeLabel(cache),
		sanitizeLabel(result),
	).Inc()
}

func RecordBackgroundJob(service, job, result string, duration time.Duration) {
	service = sanitizeLabel(service)
	job = sanitizeLabel(job)
	result = sanitizeLabel(result)

	backgroundJobTotal.WithLabelValues(service, job, result).Inc()
	backgroundJobDuration.WithLabelValues(service, job).Observe(duration.Seconds())
}

func RecordKafkaMessage(service, topic, direction, result string, duration time.Duration) {
	service = sanitizeLabel(service)
	topic = sanitizeLabel(topic)
	direction = sanitizeLabel(direction)
	result = sanitizeLabel(result)

	kafkaMessageTotal.WithLabelValues(service, topic, direction, result).Inc()
	kafkaMessageDuration.WithLabelValues(service, topic, direction).Observe(duration.Seconds())
}

func RecordWebsocketMessage(service, direction, result string) {
	websocketMessagesTotal.WithLabelValues(
		sanitizeLabel(service),
		sanitizeLabel(direction),
		sanitizeLabel(result),
	).Inc()
}

func SetWebsocketConnections(service string, count int) {
	websocketConnections.WithLabelValues(sanitizeLabel(service)).Set(float64(count))
}

func StartDependencyMonitor(service, dependency, target string, interval time.Duration, check func(context.Context) error) {
	service = sanitizeLabel(service)
	dependency = sanitizeLabel(dependency)
	target = sanitizeLabel(target)
	if interval <= 0 {
		interval = defaultInterval
	}

	key := strings.Join([]string{service, dependency, target}, "|")
	if _, loaded := dependencyLoops.LoadOrStore(key, struct{}{}); loaded {
		return
	}

	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), dependencyTimeout)
		defer cancel()

		if err := check(ctx); err != nil {
			dependencyUp.WithLabelValues(service, dependency, target).Set(0)
			return
		}
		dependencyUp.WithLabelValues(service, dependency, target).Set(1)
	}

	run()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			run()
		}
	}()
}

func StartTCPMonitor(service, dependency, target string, interval time.Duration) {
	StartDependencyMonitor(service, dependency, target, interval, func(ctx context.Context) error {
		dialer := &net.Dialer{}
		conn, err := dialer.DialContext(ctx, "tcp", target)
		if err != nil {
			return err
		}
		return conn.Close()
	})
}

func StartKafkaBrokerMonitor(service string, brokers []string, interval time.Duration) {
	if len(brokers) == 0 {
		return
	}

	target := strings.TrimSpace(brokers[0])
	if target == "" {
		return
	}

	StartDependencyMonitor(service, "kafka", target, interval, func(ctx context.Context) error {
		conn, err := kafka.DialContext(ctx, "tcp", target)
		if err != nil {
			return err
		}
		return conn.Close()
	})
}

func StartDBMonitor(service, target string, db *gorm.DB) {
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil || sqlDB == nil {
		return
	}

	service = sanitizeLabel(service)
	key := service + "|" + sanitizeLabel(target)
	if _, loaded := dbLoops.LoadOrStore(key, struct{}{}); loaded {
		return
	}

	updateDBMetrics(service, sqlDB)
	StartDependencyMonitor(service, "postgres", target, defaultInterval, func(ctx context.Context) error {
		return sqlDB.PingContext(ctx)
	})

	go func() {
		ticker := time.NewTicker(defaultInterval)
		defer ticker.Stop()

		for range ticker.C {
			updateDBMetrics(service, sqlDB)
		}
	}()
}

func InstrumentRedis(service, target string, client *redis.Client) {
	if client == nil {
		return
	}

	service = sanitizeLabel(service)
	key := service + "|" + sanitizeLabel(target)
	if _, loaded := redisHooks.LoadOrStore(key, struct{}{}); !loaded {
		client.AddHook(redisMetricsHook{service: service})
	}

	StartDependencyMonitor(service, "redis", target, defaultInterval, func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	})
}

func sanitizeLabel(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "unknown"
	}

	value = strings.ReplaceAll(value, " ", "_")
	return value
}

type dbSnapshot struct {
	waitCount     int64
	waitDuration  time.Duration
	maxIdleClosed int64
	maxIdleTime   int64
	maxLifetime   int64
}

var lastDBStats sync.Map

func updateDBMetrics(service string, db *sql.DB) {
	stats := db.Stats()

	dbConnections.WithLabelValues(service, "max_open").Set(float64(stats.MaxOpenConnections))
	dbConnections.WithLabelValues(service, "open").Set(float64(stats.OpenConnections))
	dbConnections.WithLabelValues(service, "in_use").Set(float64(stats.InUse))
	dbConnections.WithLabelValues(service, "idle").Set(float64(stats.Idle))

	current := dbSnapshot{
		waitCount:     stats.WaitCount,
		waitDuration:  stats.WaitDuration,
		maxIdleClosed: stats.MaxIdleClosed,
		maxIdleTime:   stats.MaxIdleTimeClosed,
		maxLifetime:   stats.MaxLifetimeClosed,
	}

	previousAny, _ := lastDBStats.LoadOrStore(service, current)
	previous, _ := previousAny.(dbSnapshot)
	lastDBStats.Store(service, current)

	if delta := stats.WaitCount - previous.waitCount; delta > 0 {
		dbWaitCountTotal.WithLabelValues(service).Add(float64(delta))
	}
	if delta := stats.WaitDuration - previous.waitDuration; delta > 0 {
		dbWaitDurationSecondsTotal.WithLabelValues(service).Add(delta.Seconds())
	}
	if delta := stats.MaxIdleClosed - previous.maxIdleClosed; delta > 0 {
		dbClosedTotal.WithLabelValues(service, "max_idle").Add(float64(delta))
	}
	if delta := stats.MaxIdleTimeClosed - previous.maxIdleTime; delta > 0 {
		dbClosedTotal.WithLabelValues(service, "max_idle_time").Add(float64(delta))
	}
	if delta := stats.MaxLifetimeClosed - previous.maxLifetime; delta > 0 {
		dbClosedTotal.WithLabelValues(service, "max_lifetime").Add(float64(delta))
	}
}

type redisMetricsHook struct {
	service string
}

func (h redisMetricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h redisMetricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		recordRedisCommand(h.service, cmd.Name(), ResultFromError(err), time.Since(start))
		return err
	}
}

func (h redisMetricsHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		recordRedisCommand(h.service, "pipeline", ResultFromError(err), time.Since(start))
		return err
	}
}

func recordRedisCommand(service, command, result string, duration time.Duration) {
	service = sanitizeLabel(service)
	command = sanitizeLabel(command)
	result = sanitizeLabel(result)

	redisCommandTotal.WithLabelValues(service, command, result).Inc()
	redisCommandDuration.WithLabelValues(service, command).Observe(duration.Seconds())
}
