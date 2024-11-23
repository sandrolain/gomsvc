package redislib

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sandrolain/gomsvc/pkg/eventlib"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"github.com/vmihailenco/msgpack/v5"
	"go.jetpack.io/typeid"
)

type StreamPublisherConfig struct {
	Stream  string
	Type    string
	Origin  string
	Timeout time.Duration
}

func NewStreamPublisher[T any](cfg StreamPublisherConfig) (res *StreamPublisher[T], err error) {
	// TODO: config validation
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	res = &StreamPublisher[T]{
		stream:        cfg.Stream,
		messageType:   cfg.Type,
		messageOrigin: cfg.Origin,
		timeout:       timeout,
	}
	return
}

type StreamPublisher[T any] struct {
	stream        string
	messageType   string
	messageOrigin string
	timeout       time.Duration
}

func (s *StreamPublisher[T]) Publish(payload T) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	t, e := typeid.From(s.stream, "")
	if e != nil {
		return svc.Error("cannot generate message id", e)
	}

	pl, e := msgpack.Marshal(payload)
	if e != nil {
		return svc.Error("cannot marshal payload", e)
	}

	values := map[string]interface{}{
		"tms": time.Now().Format(time.RFC3339Nano),
		"ids": t.String(),
		"typ": s.messageType,
		"ori": s.messageOrigin,
		"pld": pl,
	}

	e = redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: s.stream,
		MaxLen: 0,
		ID:     "",
		Values: values,
	}).Err()

	if e != nil {
		return svc.Error("cannot publish message", e)
	}
	return
}

type StreamConsumerConfig struct {
	Stream   string
	Group    string
	Consumer string
	Size     int
}

func NewStreamConsumer[T any](cfg StreamConsumerConfig) (res *StreamConsumer[T], err error) {
	ctx, cancel := context.WithCancel(context.Background())
	// TODO: config validation
	res = &StreamConsumer[T]{
		stream:   cfg.Stream,
		group:    cfg.Group,
		consumer: cfg.Consumer,
		ctx:      ctx,
		cancel:   cancel,
		Emitter:  eventlib.NewEmitter[*Message[T]](context.Background(), cfg.Size),
	}
	return
}

type StreamConsumer[T any] struct {
	stream   string
	group    string
	consumer string
	ctx      context.Context
	cancel   context.CancelFunc
	Emitter  *eventlib.Emitter[*Message[T]]
}

func (s *StreamConsumer[T]) Cancel() {
	s.cancel()
}

func (s *StreamConsumer[T]) Consume() error {
	id := "0"

	e := redisClient.XGroupCreateMkStream(s.ctx, s.stream, s.group, id).Err()
	if e != nil && e.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("cannot create group stream: %w", e)
	}

	e = redisClient.XGroupCreateConsumer(s.ctx, s.stream, s.group, s.consumer).Err()
	if e != nil {
		return fmt.Errorf("cannot create group consumer: %w", e)
	}

	go func() {
		for {
			if s.ctx.Err() != nil {
				return
			}
			stream, err := redisClient.XReadGroup(s.ctx, &redis.XReadGroupArgs{
				Group:    s.group,
				Consumer: s.consumer,
				Streams:  []string{s.stream, ">"},
				//count is number of entries we want to read from redis
				Count: 1,
				//we use the block command to make sure if no entry is found we wait
				//until an entry is found
				Block: 0,
			}).Result()

			if err != nil {
				_ = svc.Error("cannot read messages stream",
					err,
					"stream", s.stream,
					"group", s.group,
					"consumer", s.consumer,
				)
				if err != redis.Nil {
					// Only sleep on real errors, not on empty results
					time.Sleep(time.Second)
				}
				continue
			}

			///we have received the data we should loop it and queue the messages
			//so that our jobs can start processing
			for _, item := range stream {
				for _, msg := range item.Messages {
					message, err := parseStreamMessage[T](&msg)
					if err != nil {
						svc.Logger().Error("cannot parse message",
							"error", err,
							"message_id", msg.ID,
							"stream", s.stream,
						)
						// Acknowledge the message even if we can't parse it
						// to prevent endless retry of unparseable messages
						if ackErr := redisClient.XAck(s.ctx, s.stream, s.group, msg.ID).Err(); ackErr != nil {
							svc.Logger().Error("failed to acknowledge unparseable message",
								"error", ackErr,
								"original_error", err,
								"message_id", msg.ID,
								"stream", s.stream,
							)
						}
						continue
					}
					s.Emitter.Emit(message)
				}
			}
		}
	}()
	return nil
}

func parseStreamMessage[T any](msg *redis.XMessage) (res *Message[T], err error) {
	pldBytes, err := getValue(msg, "pld")
	if err != nil {
		return
	}

	var pld T
	e := msgpack.Unmarshal([]byte(pldBytes), &pld)
	if e != nil {
		err = fmt.Errorf("cannot unmarshal payload in message %v", msg.ID)
		return
	}

	id, _ := getValue(msg, "ids")
	ts, _ := getValue(msg, "tms")
	ty, _ := getValue(msg, "typ")
	or, _ := getValue(msg, "ori")

	timestamp, _ := time.Parse(time.RFC3339Nano, ts)

	res = &Message[T]{
		Id:        id,
		Timestamp: timestamp,
		Type:      ty,
		Origin:    or,
		Payload:   pld,
	}

	return
}

func getValue(msg *redis.XMessage, key string) (res string, err error) {
	raw, ok := msg.Values[key]
	if !ok {
		err = fmt.Errorf("%s not available in message %v", key, msg.ID)
		return
	}

	res, ok = raw.(string)
	if !ok {
		err = fmt.Errorf("bad %s interface in message %v", key, msg.ID)
		return
	}

	return
}
