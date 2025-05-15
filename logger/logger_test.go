// Package logger provides logging functionality.
package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

const (
	unexpectedLogOutput = "Unexpected log output: %s"
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

// TestCustomLogger tests creating a custom logger with a specific prefix
func TestCustomLogger(t *testing.T) {
	logger := &Logger{
		level:  INFO,
		prefix: "TEST",
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}

	// A struct created with the & operator can never be nil

	if logger.prefix != "TEST" {
		t.Errorf("Expected prefix TEST, got %s", logger.prefix)
	}
}

func TestSetLevel(t *testing.T) {
	logger := Get()
	origLevel := logger.level
	defer logger.SetLevel(origLevel) // restore original level

	// Test setting each level
	logger.SetLevel(DEBUG)
	if logger.level != DEBUG {
		t.Errorf("Expected level DEBUG, got %v", logger.level)
	}

	logger.SetLevel(INFO)
	if logger.level != INFO {
		t.Errorf("Expected level INFO, got %v", logger.level)
	}

	logger.SetLevel(WARN)
	if logger.level != WARN {
		t.Errorf("Expected level WARN, got %v", logger.level)
	}

	logger.SetLevel(ERROR)
	if logger.level != ERROR {
		t.Errorf("Expected level ERROR, got %v", logger.level)
	}

	logger.SetLevel(FATAL)
	if logger.level != FATAL {
		t.Errorf("Expected level FATAL, got %v", logger.level)
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
	logger.Info("This is info")
	output := buf.String()
	if !strings.Contains(output, "[TEST] INFO: This is info") {
		t.Errorf(unexpectedLogOutput, output)
	}

	// Test Infof
	buf.Reset()
	logger.Infof("This is info %s #%d", "test", 1)
	output = buf.String()
	if !strings.Contains(output, "[TEST] INFO: This is info test #1") {
		t.Errorf(unexpectedLogOutput, output)
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
	logger.Debug("This is debug")
	output := buf.String()
	if !strings.Contains(output, "[TEST] DEBUG: This is debug") {
		t.Errorf(unexpectedLogOutput, output)
	}

	// Test Debugf
	buf.Reset()
	logger.Debugf("This is debug %s #%d", "test", 2)
	output = buf.String()
	if !strings.Contains(output, "[TEST] DEBUG: This is debug test #2") {
		t.Errorf(unexpectedLogOutput, output)
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
		t.Errorf(unexpectedLogOutput, output)
	}

	// Test Warnf
	buf.Reset()
	logger.Warnf("This is warning %s #%d", "test", 2)
	output = buf.String()
	if !strings.Contains(output, "[TEST] WARN: This is warning test #2") {
		t.Errorf(unexpectedLogOutput, output)
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
		t.Errorf(unexpectedLogOutput, output)
	}

	// Test Errorf
	buf.Reset()
	logger.Errorf("This is error %s #%d", "test", 2)
	output = buf.String()
	if !strings.Contains(output, "[TEST] ERROR: This is error test #2") {
		t.Errorf(unexpectedLogOutput, output)
	}
}

// No need to redefine osExit - using the one from logger.go

func TestFatalLogs(t *testing.T) {
	// Save original os.Exit and restore it at the end
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	// Create a mock for os.Exit
	exitCalled := false
	exitCode := 0
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	// Use a buffer to capture log output
	var buf bytes.Buffer
	logger := &Logger{
		level:  FATAL,
		prefix: "TEST",
		logger: log.New(&buf, "", 0), // No timestamp/flags for easier testing
	}

	// Test Fatal
	buf.Reset()
	exitCalled = false
	logger.Fatal("This is fatal")
	output := buf.String()

	if !strings.Contains(output, "[TEST] FATAL: This is fatal") {
		t.Errorf(unexpectedLogOutput, output)
	}

	if !exitCalled {
		t.Error("os.Exit was not called for Fatal")
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Test Fatalf
	buf.Reset()
	exitCalled = false
	logger.Fatalf("This is fatal %s #%d", "test", 2)
	output = buf.String()

	if !strings.Contains(output, "[TEST] FATAL: This is fatal test #2") {
		t.Errorf(unexpectedLogOutput, output)
	}

	if !exitCalled {
		t.Error("os.Exit was not called for Fatalf")
	}
}
