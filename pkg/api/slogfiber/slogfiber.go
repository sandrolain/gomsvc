package slogfiber

import (
	"context"
	"net/http"
	"sync"
	"time"

	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithRequestID bool
}

// New returns a fiber.Handler (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func New(logger *slog.Logger) fiber.Handler {
	return NewWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithRequestID: true,
	})
}

// NewWithConfig returns a fiber.Handler (middleware) that logs requests using slog.
func NewWithConfig(logger *slog.Logger, config Config) fiber.Handler {
	var once sync.Once
	var errHandler fiber.ErrorHandler
	return func(c *fiber.Ctx) error {
		once.Do(func() {
			errHandler = c.App().ErrorHandler
		})

		c.Path()
		start := time.Now()
		path := c.Path()

		requestID := uuid.New().String()
		if config.WithRequestID {
			c.Context().SetUserValue("request-id", requestID)
			c.Set("X-Request-ID", requestID)
		}

		chainErr := c.Next()

		end := time.Now()
		latency := end.Sub(start)

		ip := c.Context().RemoteIP().String()
		if len(c.IPs()) > 0 {
			ip = c.IPs()[0]
		}

		attributes := []slog.Attr{
			slog.Time("time", end),
			slog.Duration("latency", latency),
			slog.String("method", string(c.Context().Method())),
			slog.String("host", c.Hostname()),
			slog.String("path", path),
			slog.String("route", c.Route().Path),
			slog.Int("status", c.Response().StatusCode()),
			slog.String("ip", ip),
			slog.String("user-agent", string(c.Context().UserAgent())),
			slog.String("referer", c.Get(fiber.HeaderReferer)),
		}

		if len(c.IPs()) > 0 {
			attributes = append(attributes, slog.Any("x-forwarded-for", c.IPs()))
		}

		if config.WithRequestID {
			attributes = append(attributes, slog.String("request-id", requestID))
		}

		if chainErr != nil {
			attributes = append(attributes, slog.Any("err", chainErr))
			if err := errHandler(c, chainErr); err != nil {
				c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		status := c.Response().StatusCode()
		msg := http.StatusText(status)
		if chainErr != nil {
			msg = msg + ": " + chainErr.Error()
		}

		switch {
		case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
			logger.LogAttrs(context.Background(), config.ClientErrorLevel, msg, attributes...)
		case status >= http.StatusInternalServerError:
			logger.LogAttrs(context.Background(), config.ServerErrorLevel, msg, attributes...)
		default:
			logger.LogAttrs(context.Background(), config.DefaultLevel, msg, attributes...)
		}

		return nil
	}
}

// GetRequestID returns the request identifier
func GetRequestID(c *fiber.Ctx) string {
	requestID, ok := c.Context().UserValue("request-id").(string)
	if !ok {
		return ""
	}

	return requestID
}
