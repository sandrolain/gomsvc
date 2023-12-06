package redislib

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/sandrolain/gomsvc/pkg/svc"
)

const DefaultTimeout = 2 * time.Second

var redisClient *redis.Client
var timeout time.Duration

func timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout*time.Second)
}

type EnvClientConfig struct {
	Hostname string        `env:"REDIS_HOST" validate:"required,hostname"`
	Port     int           `env:"REDIS_PORT" validate:"required,numeric"`
	Password string        `env:"REDIS_PASSWORD" validate:"required"`
	Timeout  time.Duration `env:"REDIS_TIMEOUT" validate:"numeric"`
}

type ClientOptions struct {
	Hostname string        `validation:"required,hostname"`
	Port     int           `validation:"required,numeric"`
	Password string        `validation:"required"`
	Timeout  time.Duration `validation:"required"`
	TLS      *tls.Config
}

func ClientOptionsFromEnvConfig(cfg EnvClientConfig) ClientOptions {
	return ClientOptions{
		Hostname: cfg.Hostname,
		Port:     cfg.Port,
		Password: cfg.Password,
		Timeout:  cfg.Timeout,
	}
}

func (o ClientOptions) FormatHost() string {
	return fmt.Sprintf("%s:%d", o.Hostname, o.Port)
}

func (o ClientOptions) FormatUri() string {
	return fmt.Sprintf("redis://:%s@%s", o.Password, o.FormatHost())
}

func Connect(config ClientOptions) (err error) {
	if e := validator.New().Struct(config); e != nil {
		err = svc.Error("redis config not valid", e)
		return
	}

	timeout = config.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	addr := config.FormatHost()

	redisClient = redis.NewClient(&redis.Options{
		Addr:      addr,
		Password:  config.Password,
		DB:        0, // use default DB
		TLSConfig: config.TLS,
	})

	ctx, cancel := timeoutCtx()
	defer cancel()

	if e := redisClient.Ping(ctx).Err(); e != nil {
		err = svc.Error("cannot ping redis", e)
		return
	}
	return
}
