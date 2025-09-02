package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected logrus.Level
	}{
		{"Debug level", "debug", logrus.DebugLevel},
		{"Info level", "info", logrus.InfoLevel},
		{"Warn level", "warn", logrus.WarnLevel},
		{"Error level", "error", logrus.ErrorLevel},
		{"Invalid level defaults to info", "invalid", logrus.InfoLevel},
		{"Empty level defaults to info", "", logrus.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger
			log = nil

			InitLogger(tt.level)
			logger := GetLogger()

			if logger.GetLevel() != tt.expected {
				t.Errorf("Expected log level %v, got %v", tt.expected, logger.GetLevel())
			}

			// Test that formatter is set correctly
			formatter, ok := logger.Formatter.(*logrus.TextFormatter)
			if !ok {
				t.Error("Expected TextFormatter")
			} else {
				if !formatter.FullTimestamp {
					t.Error("Expected FullTimestamp to be true")
				}
				if formatter.TimestampFormat != "2006-01-02 15:04:05" {
					t.Errorf("Expected timestamp format '2006-01-02 15:04:05', got %s", formatter.TimestampFormat)
				}
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	// Reset global logger
	log = nil

	logger1 := GetLogger()
	logger2 := GetLogger()

	// Should return the same instance
	if logger1 != logger2 {
		t.Error("GetLogger should return the same instance")
	}

	// Should initialize with info level by default
	if logger1.GetLevel() != logrus.InfoLevel {
		t.Errorf("Expected default level to be InfoLevel, got %v", logger1.GetLevel())
	}
}

func TestLoggerOutput(t *testing.T) {
	// Reset global logger
	log = nil

	// Create a buffer to capture output
	var buf bytes.Buffer

	InitLogger("info")
	logger := GetLogger()
	logger.SetOutput(&buf)

	// Test logging
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "level=info") {
		t.Errorf("Expected output to contain 'level=info', got: %s", output)
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		logFunc   func(*logrus.Logger)
		shouldLog bool
	}{
		{
			name:      "Debug message with debug level",
			level:     "debug",
			logFunc:   func(l *logrus.Logger) { l.Debug("debug message") },
			shouldLog: true,
		},
		{
			name:      "Debug message with info level",
			level:     "info",
			logFunc:   func(l *logrus.Logger) { l.Debug("debug message") },
			shouldLog: false,
		},
		{
			name:      "Info message with info level",
			level:     "info",
			logFunc:   func(l *logrus.Logger) { l.Info("info message") },
			shouldLog: true,
		},
		{
			name:      "Info message with warn level",
			level:     "warn",
			logFunc:   func(l *logrus.Logger) { l.Info("info message") },
			shouldLog: false,
		},
		{
			name:      "Error message with error level",
			level:     "error",
			logFunc:   func(l *logrus.Logger) { l.Error("error message") },
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger
			log = nil

			var buf bytes.Buffer
			InitLogger(tt.level)
			logger := GetLogger()
			logger.SetOutput(&buf)

			tt.logFunc(logger)

			output := buf.String()
			hasOutput := len(strings.TrimSpace(output)) > 0

			if tt.shouldLog && !hasOutput {
				t.Errorf("Expected log output but got none")
			}
			if !tt.shouldLog && hasOutput {
				t.Errorf("Expected no log output but got: %s", output)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetLogger(b *testing.B) {
	log = nil
	InitLogger("info")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetLogger()
	}
}

func BenchmarkLoggerInfo(b *testing.B) {
	log = nil
	InitLogger("info")
	logger := GetLogger()
	logger.SetOutput(&bytes.Buffer{}) // Discard output for benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}
