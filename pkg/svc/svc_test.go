package svc

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	TestValue string `env:"TEST_VALUE" validate:"required"`
}

func TestService(t *testing.T) {
	// Set required environment variables
	os.Setenv("TEST_VALUE", "test")
	os.Setenv("LOG_LEVEL", "INFO")
	
	done := make(chan bool)
	
	// Test service initialization
	go Service(ServiceOptions{
		Name:    "test-service",
		Version: "1.0.0",
	}, func(cfg TestConfig) {
		assert.Equal(t, "test", cfg.TestValue)
		assert.NotEmpty(t, ServiceID())
		assert.Equal(t, "test-service", ServiceName())
		assert.Equal(t, "1.0.0", ServiceVersion())
		
		// Test config retrieval
		retrievedConfig := Config[TestConfig]()
		assert.Equal(t, cfg, retrievedConfig)
		
		done <- true
	})

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out")
	}
}

func TestOnExit(t *testing.T) {
	// Save original osExit and defer its restoration
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()
	
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		assert.Equal(t, 0, code)
	}
	
	callCount := 0
	OnExit(func() {
		callCount++
	})
	
	OnExit(func() {
		callCount++
	})
	
	Exit(0)
	assert.True(t, exitCalled)
	assert.Equal(t, 2, callCount)
}
