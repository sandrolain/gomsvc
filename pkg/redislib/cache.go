package redislib

import (
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type Key []string

func (k *Key) String() string {
	return strings.Join(*k, ":")
}

func SetNX(key Key, ttl time.Duration, value interface{}) (err error) {
	ctx, cancel := timeoutCtx()
	defer cancel()
	b, err := msgpack.Marshal(&value)
	if err != nil {
		return err
	}
	err = redisClient.SetNX(ctx, key.String(), b, ttl).Err()
	return
}

func Set(key Key, ttl time.Duration, value interface{}) (err error) {
	ctx, cancel := timeoutCtx()
	defer cancel()
	b, err := msgpack.Marshal(&value)
	if err != nil {
		return err
	}
	err = redisClient.Set(ctx, key.String(), b, ttl).Err()
	return
}

func Get[T any](key Key) (res T, err error) {
	ctx, cancel := timeoutCtx()
	defer cancel()
	b, err := redisClient.Get(ctx, key.String()).Bytes()
	if err != nil {
		return
	}
	err = msgpack.Unmarshal(b, &res)
	return
}

func GetOrSet[T any](key Key, ttl time.Duration, fn func() (T, error)) (res T, err error) {
	res, err = Get[T](key)
	if err != nil && !IsNil(err) {
		return
	}
	if IsNil(err) {
		res, err = fn()
		if err != nil {
			return
		}
		err = Set(key, ttl, &res)
	}
	return
}

func IsNil(err error) bool {
	return err == redis.Nil
}
