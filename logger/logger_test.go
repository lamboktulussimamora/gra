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
	// Test error messages
	unexpectedLogOutput   = "Unexpected log output: %s"
	msgExpectedNoOutput   = "Expected no output for %s with %s level, got: %s"
	msgSameLoggerInstance = "Get() should return the same logger instance"
	msgExpectedPrefix     = "Expected prefix %s, got %s"
	msgExpectedLevel      = "Expected level %s, got %v"
	msgOsExitNotCalled    = "os.Exit was not called for %s"
	msgExpectedExitCode   = "Expected exit code %d, got %d"

	// Test values
	testPrefix       = "TEST"
	testInfoMessage  = "This is info"
	testDebugMessage = "This is debug"
	testWarnMessage  = "This is a warning"
	testErrorMessage = "This is an error"
	testFatalMessage = "This is fatal"
	testArgName      = "test"
	testHiddenMsg    = "This should not appear"
)

func TestGet(t *testing.T) {
	logger := Get()
	if logger == nil {
		t.Fatal("Get() returned nil")
	}

	// Get should always return the same logger instance
	logger2 := Get()
	if logger != logger2 {
		t.Error(msgSameLoggerInstance)
	}
}

// TestCustomLogger tests creating a custom logger with a specific prefix
func TestCustomLogger(t *testing.T) {
	logger := &Logger{
		level:  INFO,
		prefix: testPrefix,
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}

	// Verify the correct prefix was set
	if logger.prefix != testPrefix {
		t.Errorf(msgExpectedPrefix, testPrefix, logger.prefix)
	}

	// Verify the correct level was set
	if logger.level != INFO {
		t.Errorf(msgExpectedLevel, "INFO", logger.level)
	}

	// Verify the logger was created
	if logger.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestSetLevel(t *testing.T) {
	logger := Get()
	origLevel := logger.level // Save original level for restoration
	defer func(level LogLevel) {
		logger.SetLevel(level) // restore original level
	}(origLevel)

	// Define test cases for setting different log levels
	testCases := []struct {
		name  string
		level LogLevel
	}{
		{"DEBUG", DEBUG},
		{"INFO", INFO},
		{"WARN", WARN},
		{"ERROR", ERROR},
		{"FATAL", FATAL},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger.SetLevel(tc.level)
			if logger.level != tc.level {
				t.Errorf(msgExpectedLevel, tc.name, logger.level)
			}
		})
	}
}

// createTestLogger creates a logger with the specified level for testing
func createTestLogger(level LogLevel) (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  level,
		prefix: testPrefix,
		logger: log.New(&buf, "", 0), // No timestamp/flags for easier testing
	}
	return logger, &buf
}

// testLogOutput tests if a given log function produces the expected output
func testLogOutput(t *testing.T, logFn func(), buf *bytes.Buffer, expectedContent string) {
	buf.Reset()
	logFn()
	output := buf.String()
	if !strings.Contains(output, expectedContent) {
		t.Errorf(unexpectedLogOutput, output)
	}
}

func TestInfoLogs(t *testing.T) {
	logger, buf := createTestLogger(INFO)

	// Test cases for Info logs
	t.Run("Info Basic", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Info(testInfoMessage)
		}, buf, "[TEST] INFO: This is info")
	})

	t.Run("Infof Format", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Infof("This is info %s #%d", testArgName, 1)
		}, buf, "[TEST] INFO: This is info test #1")
	})
}

func TestDebugLogs(t *testing.T) {
	logger, buf := createTestLogger(DEBUG)

	// Test cases for Debug logs
	t.Run("Debug Basic", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Debug(testDebugMessage)
		}, buf, "[TEST] DEBUG: This is debug")
	})

	t.Run("Debugf Format", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Debugf("This is debug %s #%d", testArgName, 2)
		}, buf, "[TEST] DEBUG: This is debug test #2")
	})

	t.Run("Debug Filtering", func(t *testing.T) {
		// Test that debug messages aren't shown when level is INFO
		logger.SetLevel(INFO)
		buf.Reset()
		logger.Debug(testHiddenMsg)
		output := buf.String()
		if output != "" {
			t.Errorf(msgExpectedNoOutput, "Debug", "INFO", output)
		}
	})
}

func TestWarnLogs(t *testing.T) {
	logger, buf := createTestLogger(WARN)

	// Test cases for Warn logs
	t.Run("Warn Basic", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Warn(testWarnMessage)
		}, buf, "[TEST] WARN: This is a warning")
	})

	t.Run("Warnf Format", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Warnf("This is warning %s #%d", testArgName, 2)
		}, buf, "[TEST] WARN: This is warning test #2")
	})
}

func TestErrorLogs(t *testing.T) {
	logger, buf := createTestLogger(ERROR)

	// Test cases for Error logs
	t.Run("Error Basic", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Error(testErrorMessage)
		}, buf, "[TEST] ERROR: This is an error")
	})

	t.Run("Errorf Format", func(t *testing.T) {
		testLogOutput(t, func() {
			logger.Errorf("This is error %s #%d", testArgName, 2)
		}, buf, "[TEST] ERROR: This is error test #2")
	})
}

// No need to redefine osExit - using the one from logger.go

func TestFatalLogs(t *testing.T) {
	// Save original os.Exit and restore it at the end
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	logger, buf := createTestLogger(FATAL)

	// Test cases for Fatal logs
	testCases := []struct {
		name     string
		logFn    func()
		expected string
		funcName string
	}{
		{
			name: "Fatal Basic",
			logFn: func() {
				logger.Fatal(testFatalMessage)
			},
			expected: "[TEST] FATAL: This is fatal",
			funcName: "Fatal",
		},
		{
			name: "Fatalf Format",
			logFn: func() {
				logger.Fatalf("This is fatal %s #%d", testArgName, 2)
			},
			expected: "[TEST] FATAL: This is fatal test #2",
			funcName: "Fatalf",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock for os.Exit
			exitCalled := false
			exitCode := 0
			osExit = func(code int) {
				exitCalled = true
				exitCode = code
			}

			buf.Reset()
			tc.logFn()
			output := buf.String()

			if !strings.Contains(output, tc.expected) {
				t.Errorf(unexpectedLogOutput, output)
			}

			if !exitCalled {
				t.Errorf(msgOsExitNotCalled, tc.funcName)
			}

			if exitCode != 1 {
				t.Errorf(msgExpectedExitCode, 1, exitCode)
			}
		})
	}
}

// TestLogFiltering tests that log messages are properly filtered based on level
func TestLogFiltering(t *testing.T) {
	// Test cases for filtering behavior
	testCases := []struct {
		name          string
		setLevel      LogLevel
		logLevel      LogLevel
		logFunc       func(*Logger, *bytes.Buffer) string
		shouldDisplay bool
	}{
		{
			name:     "Debug filtered by INFO",
			setLevel: INFO,
			logLevel: DEBUG,
			logFunc: func(logger *Logger, buf *bytes.Buffer) string {
				buf.Reset()
				logger.Debug(testDebugMessage)
				return buf.String()
			},
			shouldDisplay: false,
		},
		{
			name:     "Debug displayed by DEBUG",
			setLevel: DEBUG,
			logLevel: DEBUG,
			logFunc: func(logger *Logger, buf *bytes.Buffer) string {
				buf.Reset()
				logger.Debug(testDebugMessage)
				return buf.String()
			},
			shouldDisplay: true,
		},
		{
			name:     "Info filtered by WARN",
			setLevel: WARN,
			logLevel: INFO,
			logFunc: func(logger *Logger, buf *bytes.Buffer) string {
				buf.Reset()
				logger.Info(testInfoMessage)
				return buf.String()
			},
			shouldDisplay: false,
		},
		{
			name:     "Warn displayed by WARN",
			setLevel: WARN,
			logLevel: WARN,
			logFunc: func(logger *Logger, buf *bytes.Buffer) string {
				buf.Reset()
				logger.Warn(testWarnMessage)
				return buf.String()
			},
			shouldDisplay: true,
		},
		{
			name:     "Error displayed by WARN",
			setLevel: WARN,
			logLevel: ERROR,
			logFunc: func(logger *Logger, buf *bytes.Buffer) string {
				buf.Reset()
				logger.Error(testErrorMessage)
				return buf.String()
			},
			shouldDisplay: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, buf := createTestLogger(tc.setLevel)
			output := tc.logFunc(logger, buf)

			if tc.shouldDisplay && output == "" {
				t.Errorf("Expected log output for %s with level %v, but got empty output",
					tc.name, tc.setLevel)
			}

			if !tc.shouldDisplay && output != "" {
				levelNames := map[LogLevel]string{
					DEBUG: "DEBUG",
					INFO:  "INFO",
					WARN:  "WARN",
					ERROR: "ERROR",
					FATAL: "FATAL",
				}
				t.Errorf(msgExpectedNoOutput, levelNames[tc.logLevel], levelNames[tc.setLevel], output)
			}
		})
	}
}

// TestSetPrefix tests that setting a prefix affects log output
func TestSetPrefix(t *testing.T) {
	const (
		newPrefix   = "NEW_PREFIX"
		emptyPrefix = ""
		testMessage = "test message"
	)

	testCases := []struct {
		name           string
		prefix         string
		expectedOutput string
	}{
		{
			name:           "With prefix",
			prefix:         newPrefix,
			expectedOutput: "[NEW_PREFIX] INFO: test message",
		},
		{
			name:           "Empty prefix",
			prefix:         emptyPrefix,
			expectedOutput: "INFO: test message",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, buf := createTestLogger(INFO)

			// Set the test prefix
			logger.SetPrefix(tc.prefix)

			// Log a message and check the output
			buf.Reset()
			logger.Info(testMessage)
			output := buf.String()

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output containing '%s', got: '%s'",
					tc.expectedOutput, output)
			}
		})
	}
}

// TestNew tests the creation of a new logger with a custom prefix
func TestNew(t *testing.T) {
	const customPrefix = "CUSTOM"

	// Create a new logger with a custom prefix
	logger := New(customPrefix)

	// Verify the logger was created correctly
	if logger == nil {
		t.Fatal("New() returned nil")
		return
	}

	if logger.prefix != customPrefix {
		t.Errorf(msgExpectedPrefix, customPrefix, logger.prefix)
	}

	// Default level should be INFO
	if logger.level != INFO {
		t.Errorf("Expected default level to be INFO, got %v", logger.level)
	}

	// Test that the logger works correctly
	var buf bytes.Buffer
	logger.logger.SetOutput(&buf)
	logger.Info("This is info")
	output := buf.String()

	expectedOutput := "[CUSTOM] INFO: This is info"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output containing '%s', got: '%s'", expectedOutput, output)
	}
}
