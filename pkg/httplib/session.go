package httplib

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
)

var sessionStore *session.Store

func UseSession(sessionProvider fiber.Storage) {
	sessionStore = session.New(session.Config{
		Storage: sessionProvider,
	})
}

type RedisSessionConfig struct {
	URL string
}

func UseRedisSession(cfg RedisSessionConfig) (err error) {
	UseSession(redis.New(redis.Config{
		URL: cfg.URL,
	}))
	return
}
