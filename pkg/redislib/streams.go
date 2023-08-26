package redislib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.jetpack.io/typeid"
)

type SenderConfig struct {
	Type    string
	Origin  string
	Timeout time.Duration
}

func StreamSender[T any](stream string, config SenderConfig) func(T) error {
	return func(payload T) error {
		to := config.Timeout
		if to == 0 {
			to = time.Second * 10
		}
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()
		t, err := typeid.New(stream)
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
		values := map[string]interface{}{
			"message": data,
		}
		return redisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			MaxLen: 0,
			ID:     "",
			Values: values,
		}).Err()
	}
}

func StreamConsumer[T any](stream string, group string, consumer string, receiver ReceiverFunc[T], onError ErrorFunc) func() {
	closed := false
	go func() {
		id := "0"
		ctx := context.Background()
		redisClient.XGroupCreate(ctx, stream, group, id)
		redisClient.XGroupCreateConsumer(ctx, stream, group, consumer)
		for {
			if closed {
				return
			}
			data, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				//count is number of entries we want to read from redis
				Count: 1,
				//we use the block command to make sure if no entry is found we wait
				//until an entry is found
				Block: 0,
			}).Result()
			if err != nil {
				onError(err)
				continue
			}
			///we have received the data we should loop it and queue the messages
			//so that our jobs can start processing
			for _, result := range data {
				for _, message := range result.Messages {
					fmt.Printf("message: %v\n", message)
					id = message.ID
					redisClient.XDel(ctx, stream, id)
				}
			}
		}
	}()

	return func() {
		closed = true
	}
}
