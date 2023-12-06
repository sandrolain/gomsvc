package main

import (
	"github.com/sandrolain/gomsvc/example/models"
	"github.com/sandrolain/gomsvc/pkg/devlib"
	"github.com/sandrolain/gomsvc/pkg/redislib"
	"github.com/sandrolain/gomsvc/pkg/svc"
)

type Config struct {
	Redis redislib.EnvClientConfig
}

func main() {
	svc.Service(svc.ServiceOptions{
		Name:    "svcb",
		Version: "1.2.3",
	}, func(cfg Config) {
		svc.PanicIfError(
			redislib.Connect(redislib.ClientOptionsFromEnvConfig(cfg.Redis)),
		)

		cs := svc.PanicWithError(
			redislib.NewStreamConsumer[models.MessageData](redislib.StreamConsumerConfig{
				Stream:   "mystream",
				Group:    "group1",
				Consumer: svc.ServiceName(),
			}),
		)

		cs.Emitter.Subscribe(
			func(m *redislib.Message[models.MessageData]) error {
				devlib.P(m)
				return nil
			},
			func(err error) {
				devlib.P(err)
			},
		)

		svc.PanicIfError(cs.Consume())
	})
}
