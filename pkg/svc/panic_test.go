package svc

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupPanicTest() {
	loggerLevel = new(slog.LevelVar)
	logger = slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: loggerLevel}))
}

func TestPanicWithError(t *testing.T) {
	setupPanicTest()
	
	t.Run("no error returns value", func(t *testing.T) {
		result := PanicWithError("test", nil)
		assert.Equal(t, "test", result)
	})

	t.Run("error causes panic", func(t *testing.T) {
		assert.Panics(t, func() {
			PanicWithError("test", errors.New("test error"))
		})
	})
}

func TestPanicIfError(t *testing.T) {
	setupPanicTest()
	
	t.Run("no error does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			PanicIfError(nil)
		})
	})

	t.Run("error causes panic", func(t *testing.T) {
		assert.Panics(t, func() {
			PanicIfError(errors.New("test error"))
		})
	})
}
