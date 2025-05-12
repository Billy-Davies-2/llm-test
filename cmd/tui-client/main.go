package main

import (
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Billy-Davies-2/llm-test/pkg/tui"
	"github.com/Billy-Davies-2/llm-test/pkg/tui/clipboard"
)

func main() {
	logFile, err := os.OpenFile("tui.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("Failed to open log file: " + err.Error()) // Replace with proper error handling if needed
	}
	defer logFile.Close()

	// Create a new slog handler that writes to the log file
	handler := slog.NewTextHandler(logFile, nil)
	logger := slog.New(handler)

	slog.SetDefault(logger)

	slog.Info("TUI started")

	// seed our in-memory clipboard from the OS
	clipboard.Init()
	slog.Info("Copied clipboard into in-memory clipboard")

	// run the TUI
	if _, err := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen(), tea.WithMouseAllMotion()).Run(); err != nil {
		logger.Error("TUI exited with error", "error", err)
		os.Exit(1)
	}
}
