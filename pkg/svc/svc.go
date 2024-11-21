package svc

import (
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"log/slog"

	"github.com/go-playground/validator/v10"
	typeid "go.jetpack.io/typeid"
)

var serviceUuid string

var exitCallbacks = make([]OnExitFunc, 0)

type DefaultEnv struct {
	LogLevel  string `env:"LOG_LEVEL"`
	LogFormat string `env:"LOG_FORMAT"`
	LogColor  string `env:"LOG_COLOR"`
}

type ServiceOptions struct {
	Name    string `validate:"required"`
	Version string `validate:"required,semver"`
}
type ServiceFunc[T any] func(T)

var options *ServiceOptions

var globalConfig interface{}

var osExit = os.Exit // allow to be mocked in tests

func Service[C any](opts ServiceOptions, fn ServiceFunc[C]) {
	v := validator.New()
	PanicIfError(v.Struct(opts))
	serviceUuid = PanicWithError(typeid.From(cleanTypeIdName(opts.Name), "")).String()
	options = &opts

	env := PanicWithError(GetEnv[DefaultEnv]())

	initLogger(env)

	config := PanicWithError(GetEnv[C]())

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh,
		syscall.SIGTERM, // terminate: stopped by `kill -9 PID`
		syscall.SIGINT,  // interrupt: stopped by Ctrl + C
		syscall.SIGHUP,
		syscall.SIGQUIT,
		os.Interrupt,
	)

	slog.Info(`Starting service`, "name", options.Name, "version", opts.Version, "ID", serviceUuid)

	globalConfig = config

	go fn(config)
	<-exitCh
	Exit(0)
}

func Exit(code int) {
	var wg sync.WaitGroup
	for _, fn := range exitCallbacks {
		wg.Add(1)
		go func() {
			fn()
			wg.Done()
		}()
	}
	wg.Wait()
	logger.Info("Exit service", "code", code)
	osExit(code)
}

type OnExitFunc func()

func OnExit(fn OnExitFunc) {
	exitCallbacks = append(exitCallbacks, fn)
}

func Config[T any]() T {
	return globalConfig.(T)
}

func ServiceID() string {
	return serviceUuid
}

func ServiceName() string {
	return options.Name
}

func ServiceVersion() string {
	return options.Version
}

type EmptyConfig struct{}

func cleanTypeIdName(name string) string {
	name = strings.ToLower(name)
	re := regexp.MustCompile("[^a-z]")
	name = re.ReplaceAllString(name, "")
	if len(name) > 63 {
		name = name[0:63]
	}
	return name
}
