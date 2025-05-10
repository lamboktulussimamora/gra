package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	logger := Get()
	if logger == nil {
		t.Fatal("Get() returned nil")
	}

	// Get should always return the same logger instance
	logger2 := Get()
	if logger != logger2 {
		t.Error("Get() should return the same logger instance")
	}
}

func TestNew(t *testing.T) {
	testPrefix := "TEST_LOGGER"
	logger := New(testPrefix)

	if logger == nil {
		t.Fatal("New() returned nil")
	}

	if logger.prefix != testPrefix {
		t.Errorf("Expected prefix %s, got %s", testPrefix, logger.prefix)
	}

	if logger.level != INFO {
		t.Errorf("Expected default level INFO, got %v", logger.level)
	}
}

func TestSetLevel(t *testing.T) {
	logger := New("TEST")

	// Default level should be INFO
	if logger.level != INFO {
		t.Errorf("Expected default level INFO, got %v", logger.level)
	}

	// Set level to DEBUG
	logger.SetLevel(DEBUG)
	if logger.level != DEBUG {
		t.Errorf("Expected level DEBUG, got %v", logger.level)
	}

	// Set level to ERROR
	logger.SetLevel(ERROR)
	if logger.level != ERROR {
		t.Errorf("Expected level ERROR, got %v", logger.level)
	}
}

func TestSetPrefix(t *testing.T) {
	logger := New("ORIGINAL")

	// Test prefix set by New
	if logger.prefix != "ORIGINAL" {
		t.Errorf("Expected prefix ORIGINAL, got %s", logger.prefix)
	}

	// Change prefix
	newPrefix := "CHANGED"
	logger.SetPrefix(newPrefix)
	if logger.prefix != newPrefix {
		t.Errorf("Expected prefix %s, got %s", newPrefix, logger.prefix)
	}
}

func TestInfoLogs(t *testing.T) {
	// Use a buffer to capture log output
	var buf bytes.Buffer
	logger := &Logger{
		level:  INFO,
		prefix: "TEST",
		logger: log.New(&buf, "", 0), // No timestamp/flags for easier testing
	}

	// Test Info
	buf.Reset()
	logger.Info("This is a test")
	output := buf.String()
	if !strings.Contains(output, "[TEST] INFO: This is a test") {
		t.Errorf("Unexpected log output: %s", output)
	}

	// Test Infof
	buf.Reset()
	logger.Infof("This is %s #%d", "test", 2)
	output = buf.String()
	if !strings.Contains(output, "[TEST] INFO: This is test #2") {
		t.Errorf("Unexpected log output: %s", output)
	}
}

func TestDebugLogs(t *testing.T) {
	// Use a buffer to capture log output
	var buf bytes.Buffer
	logger := &Logger{
		level:  DEBUG,
		prefix: "TEST",
		logger: log.New(&buf, "", 0), // No timestamp/flags for easier testing
	}

	// Test Debug
	buf.Reset()
	logger.Debug("This is a debug message")
	output := buf.String()
	if !strings.Contains(output, "[TEST] DEBUG: This is a debug message") {
		t.Errorf("Unexpected log output: %s", output)
	}

	// Test Debugf
	buf.Reset()
	logger.Debugf("This is debug %s #%d", "test", 2)
	output = buf.String()
	if !strings.Contains(output, "[TEST] DEBUG: This is debug test #2") {
		t.Errorf("Unexpected log output: %s", output)
	}

	// Test that debug messages aren't shown when level is INFO
	logger.SetLevel(INFO)
	buf.Reset()
	logger.Debug("This should not appear")
	output = buf.String()
	if output != "" {
		t.Errorf("Expected no output for Debug with INFO level, got: %s", output)
	}
}

func TestWarnLogs(t *testing.T) {
	// Use a buffer to capture log output
	var buf bytes.Buffer
	logger := &Logger{
		level:  WARN,
		prefix: "TEST",
		logger: log.New(&buf, "", 0), // No timestamp/flags for easier testing
	}

	// Test Warn
	buf.Reset()
	logger.Warn("This is a warning")
	output := buf.String()
	if !strings.Contains(output, "[TEST] WARN: This is a warning") {
		t.Errorf("Unexpected log output: %s", output)
	}

	// Test Warnf
	buf.Reset()
	logger.Warnf("This is warning %s #%d", "test", 2)
	output = buf.String()
	if !strings.Contains(output, "[TEST] WARN: This is warning test #2") {
		t.Errorf("Unexpected log output: %s", output)
	}
}

func TestErrorLogs(t *testing.T) {
	// Use a buffer to capture log output
	var buf bytes.Buffer
	logger := &Logger{
		level:  ERROR,
		prefix: "TEST",
		logger: log.New(&buf, "", 0), // No timestamp/flags for easier testing
	}

	// Test Error
	buf.Reset()
	logger.Error("This is an error")
	output := buf.String()
	if !strings.Contains(output, "[TEST] ERROR: This is an error") {
		t.Errorf("Unexpected log output: %s", output)
	}

	// Test Errorf
	buf.Reset()
	logger.Errorf("This is error %s #%d", "test", 2)
	output = buf.String()
	if !strings.Contains(output, "[TEST] ERROR: This is error test #2") {
		t.Errorf("Unexpected log output: %s", output)
	}
}

// We'll skip testing fatal logs since they call os.Exit
// which terminates the test process

// Override os.Exit for testing Fatal logs
var osExit = os.Exit
