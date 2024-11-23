package httplib

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
)

type RedisSessionConfig struct {
	URL string
}

func RedisSession(cfg RedisSessionConfig) fiber.Storage {
	return redis.New(redis.Config{
		URL: cfg.URL,
	})
}
