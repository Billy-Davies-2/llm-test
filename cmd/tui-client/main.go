package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/Billy-Davies-2/llm-test/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	logger := func() *slog.Logger {
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: false})
		return slog.New(h)
	}()
	peers := flag.String("peers", "localhost:50051", "Comma-separated list of peer gRPC addresses")
	flag.Parse()

	addrList := strings.Split(*peers, ",")
	model := tui.NewModel(addrList)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		logger.Error("TUI error: %v", err)
	}
}
