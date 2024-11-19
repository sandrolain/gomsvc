package svc

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

var logger *slog.Logger
var loggerLevel *slog.LevelVar

func initLogger(env DefaultEnv) {
	loggerLevel = new(slog.LevelVar)
	LogLevel(env.LogLevel)
	var handler slog.Handler
	if strings.ToUpper(env.LogFormat) == "JSON" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel, AddSource: true})
	} else if env.LogColor == "true" {
		handler = tint.NewHandler(os.Stdout, &tint.Options{Level: loggerLevel, AddSource: true})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel, AddSource: true})
	}
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

func Logger() *slog.Logger {
	if logger == nil {
		initLogger(PanicWithError(GetEnv[DefaultEnv]()))
	}
	return logger
}

func LogLevel(level string) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		loggerLevel.Set(slog.LevelDebug)
	case "INFO":
		loggerLevel.Set(slog.LevelInfo)
	case "WARN":
		loggerLevel.Set(slog.LevelWarn)
	case "ERROR":
		loggerLevel.Set(slog.LevelError)
	default:
		loggerLevel.Set(slog.LevelInfo)
	}
}

func LoggerNamespace(ns string, args ...any) *slog.Logger {
	args = append([]any{"ns", ns}, args...)
	return logger.With(args...)
}

func Error(msg string, err error, args ...any) error {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
	r.Add("err", err)
	r.Add(args...)
	_ = logger.Handler().Handle(context.Background(), r)
	return fmt.Errorf("%s: %w", msg, err)
}

func Fatal(msg string, args ...interface{}) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
	r.Add(args...)
	_ = logger.Handler().Handle(context.Background(), r)
	Exit(1)
}
