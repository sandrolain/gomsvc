package redislib

import (
	"context"
	"encoding/json"
	"time"

	"go.jetpack.io/typeid"
)

type PublisherConfig struct {
	Type    string
	Origin  string
	Timeout time.Duration
}

func Publisher[T any](channel string, config PublisherConfig) func(T) error {
	return func(payload T) error {
		to := config.Timeout
		if to == 0 {
			to = time.Second * 10
		}
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()
		t, err := typeid.New(config.Type)
		if err != nil {
			return err
		}
		message := Message[T]{
			Timestamp: time.Now(),
			Id:        t.String(),
			Type:      config.Type,
			Origin:    config.Origin,
			Payload:   payload,
		}
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		return redisClient.Publish(ctx, channel, data).Err()
	}
}

type ReceiverFunc[T any] func(payload Message[T]) error
type ErrorFunc func(error)

func Subscribe[T any](channel string, receiver ReceiverFunc[T], onError ErrorFunc) func() {
	if onError == nil {
		onError = func(err error) {}
	}

	ctx := context.Background()
	subscription := redisClient.Subscribe(ctx, channel)

	closed := false

	go func() {
		for {
			if closed {
				return
			}

			msg, err := subscription.ReceiveMessage(ctx)
			if err != nil {
				onError(err)
				continue
			}

			var payload Message[T]
			err = json.Unmarshal([]byte(msg.Payload), &payload)
			if err != nil {
				onError(err)
				continue
			}

			err = receiver(payload)
			if err != nil {
				onError(err)
				continue
			}
		}
	}()

	return func() {
		closed = true
		subscription.Close()
	}
}

type Message[T any] struct {
	Timestamp time.Time `json:"tsp"`
	Id        string    `json:"idx"`
	Type      string    `json:"typ"`
	Origin    string    `json:"org"`
	Payload   T         `json:"pld"`
}
