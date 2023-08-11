package main

import (
	"fmt"

	"github.com/sandrolain/gomscv/example/models"
	"github.com/sandrolain/gomscv/pkg/red"
	"github.com/sandrolain/gomscv/pkg/svc"
)

type Config struct {
	Port      int    `env:"PORT" validate:"required"`
	RedisAddr string `env:"REDIS_ADDR" validate:"required"`
	RedisPwd  string `env:"REDIS_PWD" validate:"required"`
}

func main() {
	svc.Service(svc.ServiceOptions{
		Name:    "svcb",
		Version: "1.2.3",
	}, func(cfg Config) {
		fmt.Printf("cfg: %v\n", cfg)

		red.Connect(cfg.RedisAddr, cfg.RedisPwd)

		// red.Subscribe("signup", func(payload red.Message[models.MessageData]) error {
		// 	svc.Logger().Debug("Message received", "payload", payload)
		// 	return nil
		// }, func(err error) {
		// 	fmt.Printf("err: %v\n", err)
		// })
		red.StreamConsumer("mystream", "group1", svc.ServiceName(), func(payload red.Message[models.MessageData]) error {
			svc.Logger().Debug("Message received", "payload", payload)
			return nil
		}, func(err error) {
			fmt.Printf("err: %v\n", err)
		})
	})
}
