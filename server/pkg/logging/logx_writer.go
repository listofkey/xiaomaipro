package logging

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type logxWriter struct {
	logger *zap.Logger
}

func newLogxWriter(logger *zap.Logger) *logxWriter {
	return &logxWriter{
		logger: logger,
	}
}

func (w *logxWriter) Alert(v any) {
	w.withField(zap.String("category", "alert")).Error(messageFrom(v))
}

func (w *logxWriter) Close() error {
	return nil
}

func (w *logxWriter) Debug(v any, fields ...logx.LogField) {
	w.withFields(fields...).Debug(messageFrom(v))
}

func (w *logxWriter) Error(v any, fields ...logx.LogField) {
	w.withFields(fields...).Error(messageFrom(v))
}

func (w *logxWriter) Info(v any, fields ...logx.LogField) {
	w.withFields(fields...).Info(messageFrom(v))
}

func (w *logxWriter) Severe(v any) {
	w.withField(zap.String("category", "severe")).Error(messageFrom(v))
}

func (w *logxWriter) Slow(v any, fields ...logx.LogField) {
	logger := w.withFields(fields...)
	logger.With(zap.String("category", "slow")).Warn(messageFrom(v))
}

func (w *logxWriter) Stack(v any) {
	w.withField(zap.String("category", "stack")).Error(messageFrom(v))
}

func (w *logxWriter) Stat(v any, fields ...logx.LogField) {
	logger := w.withFields(fields...)
	logger.With(zap.String("category", "stat")).Info(messageFrom(v))
}

func (w *logxWriter) withFields(fields ...logx.LogField) *zap.Logger {
	if w == nil || w.logger == nil {
		return zap.NewNop()
	}

	if len(fields) == 0 {
		return w.logger
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}

	return w.logger.With(zapFields...)
}

func (w *logxWriter) withField(field zap.Field) *zap.Logger {
	if w == nil || w.logger == nil {
		return zap.NewNop()
	}

	return w.logger.With(field)
}

func messageFrom(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case error:
		return value.Error()
	case fmt.Stringer:
		return value.String()
	default:
		return fmt.Sprint(v)
	}
}
