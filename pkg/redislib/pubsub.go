package redislib

import (
	"context"
	"encoding/json"
	"fmt"
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
		t, err := typeid.From(config.Type, "")
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

type ReceiverFunc[T any] func(msg Message[T])
type ErrorFunc func(error)

func Subscribe[T any](channel string, receiver ReceiverFunc[T], onError ErrorFunc) func() {
	if onError == nil {
		onError = func(err error) {}
	}

	ctx, cancel := context.WithCancel(context.Background())
	subscription := redisClient.Subscribe(ctx, channel)

	closed := false

	go func() {
		defer func() {
			closed = true
			err := subscription.Close()
			if err != nil {
				onError(fmt.Errorf("failed to close subscription: %v", err))
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-subscription.Channel():
				if msg == nil {
					continue
				}

				var message Message[T]
				err := json.Unmarshal([]byte(msg.Payload), &message)
				if err != nil {
					onError(fmt.Errorf("failed to unmarshal message: %v", err))
					continue
				}

				receiver(message)
			}
		}
	}()

	return func() {
		cancel()
		for !closed {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

type Message[T any] struct {
	Timestamp time.Time `json:"tsp"`
	Id        string    `json:"idx"`
	Type      string    `json:"typ"`
	Origin    string    `json:"org"`
	Payload   T         `json:"pld"`
}
