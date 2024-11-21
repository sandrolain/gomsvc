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

var (
	serviceUuid string
	serviceUuidOnce sync.Once
	serviceUuidMu sync.RWMutex
)

var (
	exitCallbacksMu sync.RWMutex
	exitCallbacks   = make([]OnExitFunc, 0)
)

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

var (
	optionsMu sync.RWMutex
	options   *ServiceOptions
)

var (
	globalConfigMu sync.RWMutex
	globalConfig   interface{}
)

var osExit = os.Exit // allow to be mocked in tests

func Service[C any](opts ServiceOptions, fn ServiceFunc[C]) {
	v := validator.New()
	PanicIfError(v.Struct(opts))
	
	serviceUuidOnce.Do(func() {
		uuid := PanicWithError(typeid.From(cleanTypeIdName(opts.Name), "")).String()
		serviceUuidMu.Lock()
		serviceUuid = uuid
		serviceUuidMu.Unlock()
	})

	optionsMu.Lock()
	options = &opts
	optionsMu.Unlock()

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

	serviceUuidMu.RLock()
	svcUuid := serviceUuid
	serviceUuidMu.RUnlock()

	optionsMu.RLock()
	svcOpts := options
	optionsMu.RUnlock()

	slog.Info(`Starting service`, "name", svcOpts.Name, "version", opts.Version, "ID", svcUuid)

	globalConfigMu.Lock()
	globalConfig = config
	globalConfigMu.Unlock()

	go fn(config)
	<-exitCh
	Exit(0)
}

func Exit(code int) {
	var wg sync.WaitGroup
	
	exitCallbacksMu.RLock()
	callbacks := make([]OnExitFunc, len(exitCallbacks))
	copy(callbacks, exitCallbacks)
	exitCallbacksMu.RUnlock()
	
	for _, fn := range callbacks {
		wg.Add(1)
		go func(callback OnExitFunc) {
			defer wg.Done()
			callback()
		}(fn)
	}
	wg.Wait()
	logger.Info("Exit service", "code", code)
	osExit(code)
}

type OnExitFunc func()

func OnExit(fn OnExitFunc) {
	exitCallbacksMu.Lock()
	exitCallbacks = append(exitCallbacks, fn)
	exitCallbacksMu.Unlock()
}

func Config[T any]() T {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.(T)
}

func ServiceID() string {
	serviceUuidMu.RLock()
	defer serviceUuidMu.RUnlock()
	return serviceUuid
}

func ServiceName() string {
	optionsMu.RLock()
	defer optionsMu.RUnlock()
	return options.Name
}

func ServiceVersion() string {
	optionsMu.RLock()
	defer optionsMu.RUnlock()
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
