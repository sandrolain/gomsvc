package dblib

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"log/slog"

	"github.com/sandrolain/gomsvc/pkg/svc"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormSlog struct {
	slogger       *slog.Logger
	SlowThreshold time.Duration
}

func (l *GormSlog) log(level slog.Level, ctx context.Context, msg string, args []interface{}) {
	if !l.slogger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = l.slogger.Handler().Handle(ctx, r)
}

func (l *GormSlog) LogMode(lvl logger.LogLevel) logger.Interface {
	return l
}
func (l *GormSlog) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log(slog.LevelInfo, ctx, msg, args)
}
func (l *GormSlog) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log(slog.LevelWarn, ctx, msg, args)
}
func (l *GormSlog) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log(slog.LevelError, ctx, msg, args)
}
func (l *GormSlog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {

	elapsed := time.Since(begin)
	switch {
	case err != nil && err != gorm.ErrRecordNotFound:
		sql, rows := fc()
		l.log(slog.LevelError, ctx, err.Error(), []interface{}{"rows", rows, "sql", sql, "elapsed", float64(elapsed.Nanoseconds()) / 1e6})
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.log(slog.LevelWarn, ctx, slowLog, []interface{}{"rows", rows, "sql", sql, "elapsed", float64(elapsed.Nanoseconds()) / 1e6})
	default:
		sql, rows := fc()
		l.log(slog.LevelDebug, ctx, "sql trace", []interface{}{"rows", rows, "sql", sql, "elapsed", float64(elapsed.Nanoseconds()) / 1e6})
	}

}

func NewGormSlog(slowThreshold time.Duration) logger.Interface {
	if slowThreshold == 0 {
		slowThreshold = 200 * time.Millisecond
	}
	return &GormSlog{
		slogger:       svc.Logger(),
		SlowThreshold: slowThreshold,
	}
}
