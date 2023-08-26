package main

import (
	"fmt"

	"github.com/sandrolain/gomsvc/example/models"
	"github.com/sandrolain/gomsvc/pkg/redislib"
	"github.com/sandrolain/gomsvc/pkg/svc"
)

type Config struct {
	Port      int    `env:"PORT" validate:"required"`
	RedisAddr string `env:"REDIS_ADDR" validate:"required"`
	RedisPwd  string `env:"REDIS_PWD" validate:"required"`
}

func main() {
	svc.Service(svc.ServiceOptions{
		Name:    "svcc",
		Version: "1.2.3",
	}, func(cfg Config) {
		fmt.Printf("cfg: %v\n", cfg)

		redislib.Connect(redislib.Config{
			Address:  cfg.RedisAddr,
			Password: cfg.RedisPwd,
		})

		// redislib.Subscribe("signup", func(payload redislib.Message[models.MessageData]) error {
		// 	svc.Logger().Debug("Message received", "payload", payload)
		// 	return nil
		// }, func(err error) {
		// 	fmt.Printf("err: %v\n", err)
		// })
		redislib.StreamConsumer("mystream", "group1", svc.ServiceName(), func(payload redislib.Message[models.MessageData]) error {
			svc.Logger().Debug("Message received", "payload", payload)
			return nil
		}, func(err error) {
			fmt.Printf("err: %v\n", err)
		})
	})
}
