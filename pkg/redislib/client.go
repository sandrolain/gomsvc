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
	Address  string `env:"REDIS_ADDR" validate:"required"`
	Password string `env:"REDIS_PWD" validate:"required"`
}

type Config struct {
	Address  string        `validation:"required"`
	Password string        `validation:"required"`
	Timeout  time.Duration `validation:"required"`
	TLS      *tls.Config
}

func Connect(config Config) (err error) {
	v := validator.New()
	err = v.Struct(config)
	if err != nil {
		err = fmt.Errorf("redis config not valid: %w", err)
		return
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:      config.Address,
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
