package logging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const defaultSlowThreshold = 200 * time.Millisecond

type gormLogger struct {
	logger               *zap.Logger
	level                gormlogger.LogLevel
	slowThreshold        time.Duration
	ignoreRecordNotFound bool
}

func NewGormLogger(component string) gormlogger.Interface {
	base := Logger()
	if component != "" {
		base = base.Named(component)
	}

	return &gormLogger{
		logger:               base,
		level:                gormlogger.Info,
		slowThreshold:        defaultSlowThreshold,
		ignoreRecordNotFound: true,
	}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *gormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Info {
		return
	}

	l.logger.Info(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Warn {
		return
	}

	l.logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Error {
		return
	}

	l.logger.Error(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.ignoreRecordNotFound):
		l.logger.Error("gorm query failed", append(fields, zap.Error(err))...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		l.logger.Warn("gorm slow query", append(fields, zap.Duration("slow_threshold", l.slowThreshold))...)
	case l.level >= gormlogger.Info:
		l.logger.Info("gorm query", fields...)
	}
}
