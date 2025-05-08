package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"github.com/Billy-Davies-2/llm-test/pkg/clipboard"
	"github.com/Billy-Davies-2/llm-test/pkg/tui"
)

func main() {
	// 1) load .env (for GRPC_URL or other vars)
	_ = godotenv.Load()

	// 2) set up logging to tui.log
	f, err := os.OpenFile("tui.log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Println("log open error:", err)
		os.Exit(1)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// 3) seed our in-memory clipboard from the OS
	clipboard.Init()

	// 4) run the TUI
	if _, err := tea.NewProgram(
		tui.InitialModel(),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	).Run(); err != nil {
		fmt.Println("TUI exited with error:", err)
		os.Exit(1)
	}
}
