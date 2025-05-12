package main

import (
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Billy-Davies-2/llm-test/pkg/tui"
	"github.com/Billy-Davies-2/llm-test/pkg/tui/clipboard"
)

func main() {
	f, err := os.OpenFile("tui.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		// If we can’t open the file, fall back to stderr
		f = os.Stderr
	}
	defer f.Close()
	logger := func() *slog.Logger {
		h := slog.NewTextHandler(f, &slog.HandlerOptions{
			Level:     slog.LevelError,
			AddSource: false,
		})
		return slog.New(h)
	}()

	// seed our in-memory clipboard from the OS
	clipboard.Init()

	// run the TUI
	if _, err := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen(), tea.WithMouseAllMotion()).Run(); err != nil {
		logger.Error("TUI exited with error", "error", err)
		os.Exit(1)
	}
}
