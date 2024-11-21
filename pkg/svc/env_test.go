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
		os.Setenv("REQUIRED_VALUE", "test")
		defer os.Unsetenv("REQUIRED_VALUE")

		cfg, err := GetEnv[TestEnvConfig]()
		assert.NoError(t, err)
		assert.Equal(t, "test", cfg.Required)
		assert.Empty(t, cfg.Optional)
	})

	t.Run("failure with missing required field", func(t *testing.T) {
		os.Unsetenv("REQUIRED_VALUE")

		_, err := GetEnv[TestEnvConfig]()
		assert.Error(t, err)
	})
}
