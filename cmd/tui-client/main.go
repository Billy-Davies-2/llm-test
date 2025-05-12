package main

import (
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Billy-Davies-2/llm-test/pkg/tui"
	"github.com/Billy-Davies-2/llm-test/pkg/tui/clipboard"
)

func main() {
	logger := func() *slog.Logger {
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: false})
		return slog.New(h)
	}()

	// seed our in-memory clipboard from the OS
	clipboard.Init()

	// run the TUI
	if _, err := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen(), tea.WithMouseAllMotion()).Run(); err != nil {
		logger.Error("TUI exited with error:", err)
		os.Exit(1)
	}
}
