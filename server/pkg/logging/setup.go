package logging

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultLevel           = "info"
	defaultEncoding        = "json"
	defaultDirectory       = "logs"
	defaultLogstashNetwork = "tcp"
	defaultConnectTimeout  = 3 * time.Second
	defaultWriteTimeout    = 3 * time.Second
	logstashErrorCooldown  = 30 * time.Second
	logstashReconnectDelay = 5 * time.Second
)

type Runtime struct {
	logger  *zap.Logger
	writer  *logxWriter
	level   zapcore.Level
	closers []closer
}

type closer interface {
	Close() error
}

var (
	runtimeMu      sync.RWMutex
	currentRuntime *Runtime
)

func MustSetup(serviceName, mode string, cfg Config) *Runtime {
	rt, err := Setup(serviceName, mode, cfg)
	if err != nil {
		panic(err)
	}

	return rt
}

func Setup(serviceName, mode string, cfg Config) (*Runtime, error) {
	serviceName = normalizeServiceName(serviceName)
	mode = strings.TrimSpace(mode)
	cfg = normalizeConfig(serviceName, cfg)

	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	cores := make([]zapcore.Core, 0, 3)
	closers := make([]closer, 0, 2)

	if cfg.Console {
		cores = append(cores, zapcore.NewCore(
			buildEncoder(cfg.Encoding),
			zapcore.Lock(os.Stdout),
			level,
		))
	}

	if cfg.Directory != "" {
		file, err := openLogFile(cfg.Directory, cfg.Filename)
		if err != nil {
			return nil, err
		}

		closers = append(closers, file)
		cores = append(cores, zapcore.NewCore(
			buildEncoder(defaultEncoding),
			zapcore.AddSync(file),
			level,
		))
	}

	if cfg.Logstash.Enabled {
		syncer, err := newLogstashSyncer(cfg.Logstash)
		if err != nil {
			return nil, err
		}

		closers = append(closers, syncer)
		cores = append(cores, zapcore.NewCore(
			buildEncoder(defaultEncoding),
			syncer,
			level,
		))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(
			buildEncoder(defaultEncoding),
			zapcore.Lock(os.Stdout),
			level,
		))
	}

	fields := []zap.Field{
		zap.String("service", serviceName),
	}
	if mode != "" {
		fields = append(fields, zap.String("mode", mode))
	}
	if hostname, err := os.Hostname(); err == nil && strings.TrimSpace(hostname) != "" {
		fields = append(fields, zap.String("hostname", hostname))
	}
	fields = append(fields, zap.Int("pid", os.Getpid()))

	logger := zap.New(
		zapcore.NewTee(cores...),
		zap.ErrorOutput(zapcore.Lock(os.Stderr)),
	).With(fields...)

	rt := &Runtime{
		logger:  logger,
		writer:  newLogxWriter(logger),
		level:   level,
		closers: closers,
	}

	runtimeMu.Lock()
	currentRuntime = rt
	runtimeMu.Unlock()

	zap.ReplaceGlobals(logger)
	RebindLogx()

	return rt, nil
}

func Logger() *zap.Logger {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()

	if currentRuntime == nil || currentRuntime.logger == nil {
		return zap.NewNop()
	}

	return currentRuntime.logger
}

func RebindLogx() {
	runtimeMu.RLock()
	rt := currentRuntime
	runtimeMu.RUnlock()

	if rt == nil || rt.writer == nil {
		return
	}

	logx.SetWriter(rt.writer)
	logx.SetLevel(toLogxLevel(rt.level))
}

func (rt *Runtime) Logger() *zap.Logger {
	if rt == nil || rt.logger == nil {
		return zap.NewNop()
	}

	return rt.logger
}

func (rt *Runtime) Close() error {
	if rt == nil {
		return nil
	}

	var errs []error
	if rt.logger != nil {
		if err := rt.logger.Sync(); err != nil && !isIgnorableSyncError(err) {
			errs = append(errs, err)
		}
	}

	for _, resource := range rt.closers {
		if err := resource.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func normalizeConfig(serviceName string, cfg Config) Config {
	cfg.Level = normalizeOrDefault(cfg.Level, defaultLevel)
	cfg.Encoding = normalizeOrDefault(cfg.Encoding, defaultEncoding)
	if cfg.Directory == "" {
		cfg.Directory = defaultDirectory
	}
	if cfg.Filename == "" {
		cfg.Filename = sanitizeFilename(serviceName) + ".log"
	}
	cfg.Logstash.Network = normalizeOrDefault(cfg.Logstash.Network, defaultLogstashNetwork)
	if cfg.Logstash.ConnectTimeoutSeconds <= 0 {
		cfg.Logstash.ConnectTimeoutSeconds = int(defaultConnectTimeout / time.Second)
	}
	if cfg.Logstash.WriteTimeoutSeconds <= 0 {
		cfg.Logstash.WriteTimeoutSeconds = int(defaultWriteTimeout / time.Second)
	}

	return cfg
}

func normalizeServiceName(serviceName string) string {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return "unknown-service"
	}

	return serviceName
}

func normalizeOrDefault(value, fallback string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return fallback
	}

	return value
}

func parseLevel(level string) (zapcore.Level, error) {
	switch normalizeOrDefault(level, defaultLevel) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unsupported log level: %s", level)
	}
}

func toLogxLevel(level zapcore.Level) uint32 {
	switch {
	case level <= zapcore.DebugLevel:
		return logx.DebugLevel
	case level <= zapcore.InfoLevel, level == zapcore.WarnLevel:
		return logx.InfoLevel
	case level == zapcore.ErrorLevel:
		return logx.ErrorLevel
	default:
		return logx.SevereLevel
	}
}

func buildEncoder(kind string) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
	}

	if normalizeOrDefault(kind, defaultEncoding) == "console" {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func openLogFile(directory, filename string) (*os.File, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory %s: %w", directory, err)
	}

	fullpath := filepath.Join(directory, filename)
	file, err := os.OpenFile(fullpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file %s: %w", fullpath, err)
	}

	return file, nil
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(".", "-", "/", "-", "\\", "-", ":", "-", " ", "-")
	name = replacer.Replace(strings.TrimSpace(name))
	name = strings.Trim(name, "-")
	if name == "" {
		return "service"
	}

	return name
}

func isIgnorableSyncError(err error) bool {
	if err == nil {
		return false
	}

	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "invalid argument")
}

type logstashSyncer struct {
	mu               sync.Mutex
	network          string
	address          string
	connectTimeout   time.Duration
	writeTimeout     time.Duration
	conn             net.Conn
	lastReportedTime time.Time
	nextRetryTime    time.Time
}

func newLogstashSyncer(cfg LogstashConfig) (*logstashSyncer, error) {
	address := strings.TrimSpace(cfg.Address)
	if address == "" {
		return nil, errors.New("logstash address is required when logstash logging is enabled")
	}

	return &logstashSyncer{
		network:        normalizeOrDefault(cfg.Network, defaultLogstashNetwork),
		address:        address,
		connectTimeout: time.Duration(cfg.ConnectTimeoutSeconds) * time.Second,
		writeTimeout:   time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
	}, nil
}

func (s *logstashSyncer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureConnLocked(); err != nil {
		s.reportLocked(err)
		return len(p), nil
	}

	written, err := s.writeLocked(p)
	if err == nil {
		return written, nil
	}

	s.closeLocked()

	if retryErr := s.ensureConnLocked(); retryErr != nil {
		s.reportLocked(retryErr)
		return len(p), nil
	}

	written, err = s.writeLocked(p)
	if err != nil {
		s.reportLocked(err)
		s.closeLocked()
		return len(p), nil
	}

	return written, nil
}

func (s *logstashSyncer) Sync() error {
	return nil
}

func (s *logstashSyncer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.closeLocked()
}

func (s *logstashSyncer) ensureConnLocked() error {
	if s.conn != nil {
		return nil
	}

	now := time.Now()
	if now.Before(s.nextRetryTime) {
		return fmt.Errorf("logstash reconnect delayed until %s", s.nextRetryTime.Format(time.RFC3339))
	}

	dialer := &net.Dialer{Timeout: s.connectTimeout}
	conn, err := dialer.Dial(s.network, s.address)
	if err != nil {
		s.nextRetryTime = now.Add(logstashReconnectDelay)
		return err
	}

	s.conn = conn
	s.nextRetryTime = time.Time{}
	return nil
}

func (s *logstashSyncer) writeLocked(p []byte) (int, error) {
	if s.writeTimeout > 0 {
		if err := s.conn.SetWriteDeadline(time.Now().Add(s.writeTimeout)); err != nil {
			return 0, err
		}
	}

	total := 0
	for total < len(p) {
		written, err := s.conn.Write(p[total:])
		total += written
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (s *logstashSyncer) closeLocked() error {
	if s.conn == nil {
		return nil
	}

	err := s.conn.Close()
	s.conn = nil
	s.nextRetryTime = time.Now().Add(logstashReconnectDelay)
	return err
}

func (s *logstashSyncer) reportLocked(err error) {
	now := time.Now()
	if now.Sub(s.lastReportedTime) < logstashErrorCooldown {
		return
	}

	s.lastReportedTime = now
	_, _ = fmt.Fprintf(os.Stderr, "logstash writer unavailable (%s): %v\n", s.address, err)
}
