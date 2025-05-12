package tui

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	// Create a temporary log file for the test
	logFile, err := os.CreateTemp("", "tui_test.log")
	if err != nil {
		t.Fatal("Failed to create temp log file:", err)
	}
	defer os.Remove(logFile.Name()) // Clean up after the test

	// Configure slog to write to the temp file
	handler := slog.NewTextHandler(logFile, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Generate a log message
	slog.Info("Test message")

	// Read the log file
	logContent, err := os.ReadFile(logFile.Name())
	if err != nil {
		t.Fatal("Failed to read log file:", err)
	}

	// Verify the log content
	expected := "Test message"
	if !strings.Contains(string(logContent), expected) {
		t.Errorf("Expected log to contain %q, got %q", expected, string(logContent))
	}
}
