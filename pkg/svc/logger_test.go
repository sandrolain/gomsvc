package svc

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTest() {
	loggerLevel = new(slog.LevelVar)
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: loggerLevel})
	logger = slog.New(handler)
}

func TestLogLevel(t *testing.T) {
	setupTest()
	
	tests := []struct {
		level    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // default case
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			LogLevel(tt.level)
			assert.Equal(t, tt.expected, loggerLevel.Level())
		})
	}
}

func TestLoggerNamespace(t *testing.T) {
	setupTest()
	
	// Setup JSON logger to capture output
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: loggerLevel})
	logger = slog.New(handler)

	// Create namespaced logger and log a message
	nsLogger := LoggerNamespace("test-ns", "key", "value")
	nsLogger.Info("test message")

	// Parse the JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// Verify namespace and additional fields
	assert.Equal(t, "test-ns", logEntry["ns"])
	assert.Equal(t, "value", logEntry["key"])
	assert.Equal(t, "test message", logEntry["msg"])
}

func TestError(t *testing.T) {
	setupTest()
	
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: loggerLevel})
	logger = slog.New(handler)

	testErr := errors.New("test error")
	err := Error("error message", testErr, "key", "value")

	// Verify returned error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error message")
	assert.Contains(t, err.Error(), "test error")

	// Parse the JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// Verify log entry
	assert.Equal(t, "error message", logEntry["msg"])
	assert.Equal(t, "value", logEntry["key"])
	assert.Contains(t, logEntry["err"], "test error")
}

func TestInitLogger(t *testing.T) {
	setupTest()
	
	tests := []struct {
		name       string
		env        DefaultEnv
		checkJSON  bool
		checkColor bool
	}{
		{
			name: "JSON format",
			env: DefaultEnv{
				LogLevel:  "INFO",
				LogFormat: "JSON",
			},
			checkJSON: true,
		},
		{
			name: "Text format with color",
			env: DefaultEnv{
				LogLevel:  "INFO",
				LogFormat: "TEXT",
				LogColor:  "true",
			},
			checkColor: true,
		},
		{
			name: "Plain text format",
			env: DefaultEnv{
				LogLevel:  "INFO",
				LogFormat: "TEXT",
				LogColor:  "false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			initLogger(tt.env)
			logger.Info("test message")

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, err := buf.ReadFrom(r)
			assert.NoError(t, err)
			output := buf.String()

			if tt.checkJSON {
				var logEntry map[string]interface{}
				err = json.Unmarshal([]byte(output), &logEntry)
				assert.NoError(t, err)
				assert.Equal(t, "test message", logEntry["msg"])
			} else {
				assert.Contains(t, output, "test message")
				if tt.checkColor {
					assert.True(t, strings.Contains(output, "\x1b["))
				}
			}
		})
	}
}
