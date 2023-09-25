package redislib

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var timeout time.Duration

func timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout*time.Second)
}

type EnvConfig struct {
	Host     string `env:"REDIS_HOST" validate:"required,hostname"`
	Port     int    `env:"REDIS_PORT" validate:"required,numeric"`
	Password string `env:"REDIS_PASSWORD" validate:"required"`
}

type Config struct {
	Host     string        `validation:"required,hostname"`
	Port     int           `validation:"required,numeric"`
	Password string        `validation:"required"`
	Timeout  time.Duration `validation:"required"`
	TLS      *tls.Config
}

func FormatUri(cfg Config) string {
	return fmt.Sprintf("redis://:%s@%s:%d", cfg.Password, cfg.Host, cfg.Port)
}

func Connect(config Config) (err error) {
	v := validator.New()
	err = v.Struct(config)
	if err != nil {
		err = fmt.Errorf("redis config not valid: %w", err)
		return
	}

	addr := fmt.Sprintf("%v:%v", config.Host, config.Port)

	redisClient = redis.NewClient(&redis.Options{
		Addr:      addr,
		Password:  config.Password,
		DB:        0, // use default DB
		TLSConfig: config.TLS,
	})

	timeout = config.Timeout

	ctx, cancel := timeoutCtx()
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return
}
