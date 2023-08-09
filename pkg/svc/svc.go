package svc

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"log/slog"

	"github.com/caarlos0/env/v9"
	"github.com/go-playground/validator/v10"
	"github.com/lmittmann/tint"
	"github.com/sandrolain/gomscv/pkg/control"
	typeid "go.jetpack.io/typeid"
)

var serviceUuid string

var exitCallbacks = make([]OnExitFunc, 0)

type DefaultEnv struct {
	LogLevel  string `env:"LOG_LEVEL"`
	LogFormat string `env:"LOG_FORMAT"`
}

type ServiceOptions struct {
	Name    string `validate:"required"`
	Version string `validate:"required,semver"`
}
type ServiceFunc[T any] func(T)

var options *ServiceOptions
var logger *slog.Logger
var loggerLevel *slog.LevelVar

func Service[C any](opts ServiceOptions, fn ServiceFunc[C]) {
	v := validator.New()
	control.PanicIfError(v.Struct(opts))
	serviceUuid = control.PanicWithError(typeid.New(opts.Name)).String()
	options = &opts

	env := control.PanicWithError(GetEnv[DefaultEnv]())

	initLogger(env)

	config := control.PanicWithError(GetEnv[C]())

	exitCh := make(chan os.Signal)
	signal.Notify(exitCh,
		syscall.SIGTERM, // terminate: stopped by `kill -9 PID`
		syscall.SIGINT,  // interrupt: stopped by Ctrl + C
	)

	slog.Info(`Starting service`, "name", options.Name, "version", opts.Version, "ID", serviceUuid)

	go fn(config)
	<-exitCh
	exit()
}

func exit() {
	var wg sync.WaitGroup
	for _, fn := range exitCallbacks {
		wg.Add(1)
		go func() {
			fn()
			wg.Done()
		}()
	}
	wg.Wait()
	logger.Info("Exit service")
	os.Exit(0)
}

type OnExitFunc func()

func OnExit(fn OnExitFunc) {
	exitCallbacks = append(exitCallbacks, fn)
}

func GetEnv[T any]() (config T, err error) {
	err = env.Parse(&config)
	if e, ok := err.(*env.AggregateError); ok {
		for _, er := range e.Errors {
			err = fmt.Errorf("Env parse error: %v\n", er)
			return
		}
	}
	v := validator.New()
	err = v.Struct(config)
	if e, ok := err.(validator.ValidationErrors); ok {
		for _, er := range e {
			err = fmt.Errorf("Env validation error: %v\n", er)
			return
		}
	}
	return
}

func initLogger(env DefaultEnv) {
	loggerLevel = new(slog.LevelVar)
	LogLevel(env.LogLevel)
	var handler slog.Handler
	if strings.ToUpper(env.LogFormat) == "JSON" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{Level: loggerLevel})
		// handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel})
	}
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

func Logger() *slog.Logger {
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
