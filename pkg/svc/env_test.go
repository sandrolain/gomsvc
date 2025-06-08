package svc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestEnvConfig struct {
	Required string `env:"REQUIRED_VALUE" validate:"required"`
	Optional string `env:"OPTIONAL_VALUE"`
}

func TestGetEnv(t *testing.T) {
	t.Run("success with required field", func(t *testing.T) {
		err := os.Setenv("REQUIRED_VALUE", "test")
		assert.NoError(t, err)

		defer func() {
			_ = os.Unsetenv("REQUIRED_VALUE")
		}()

		cfg, err := GetEnv[TestEnvConfig]()
		assert.NoError(t, err)
		assert.Equal(t, "test", cfg.Required)
		assert.Empty(t, cfg.Optional)
	})

	t.Run("failure with missing required field", func(t *testing.T) {
		err := os.Unsetenv("REQUIRED_VALUE")
		assert.NoError(t, err)

		_, err = GetEnv[TestEnvConfig]()
		assert.Error(t, err)
	})
}
