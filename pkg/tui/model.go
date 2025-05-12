package tui

import (
	"time"

	metrics "github.com/Billy-Davies-2/llm-test/pkg/proto/metrics"
	tea "github.com/charmbracelet/bubbletea"
)

// ── Messages ─────────────────────────────────────────────────────────
type tickMsg struct{}
type thinkMsg struct{}
type pasteTickMsg struct{}
type sysTickMsg struct{}

// ── Pages ────────────────────────────────────────────────────────────
const (
	pageChat = iota
	pageSystem
)

// ── Multi‐server metrics types ───────────────────────────────────────────
type ServerMetrics struct {
	URL    string
	Client metrics.MetricsServiceClient
	Data   *metrics.MetricsResponse
	Err    error
}

// ── Tab & Model ──────────────────────────────────────────────────────
type tab struct {
	title    string
	messages []string
	input    string
	thinking bool
	dots     int
}

type model struct {
	// chat state
	tabs       []tab
	currentTab int

	// UI state
	blink         bool
	lastKey       string
	showSidebar   bool
	width, height int

	// which page is visible
	page int

	// paste animation
	pasteQueue   []string
	pasteRunning bool

	// vim‐style
	insertMode bool

	// drag & drop reorder
	dragging  bool
	dragIndex int

	// slice of servers to poll
	servers []ServerMetrics
}

// InitialModel constructs the starting model
func InitialModel() model {
	return model{
		tabs:        []tab{{title: "Tab 1", messages: []string{"Welcome!"}}},
		currentTab:  0,
		blink:       true,
		showSidebar: true,
		width:       80,
		height:      24,
		page:        pageChat,
		servers:     []ServerMetrics{},
	}
}

// ── Commands ─────────────────────────────────────────────────────────
func blinkCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond,
		func(t time.Time) tea.Msg { return tickMsg{} })
}
func sysTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second,
		func(t time.Time) tea.Msg { return sysTickMsg{} })
}

// ── Tea.Init ────────────────────────────────────────────────────────
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.EnableMouseAllMotion,
		blinkCmd(),
		sysTickCmd(),
	)
}

func (m model) NewModel(peers []string) model {
	// Create a new model with the given peers and logger
	m.servers = make([]ServerMetrics, len(peers))
	for i, peer := range peers {
		m.servers[i] = ServerMetrics{URL: peer}
	}
	return m
}

// ── Tea.Update ──────────────────────────────────────────────────────
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		// Global page‐switch (only in NORMAL)
		if !m.insertMode {
			switch s {
			case "M":
				m.page = (m.page + 1) % 2
				return m, nil
			case "C":
				m.page = pageChat
				return m, nil
			}
		}
		// Route key into chat or system
		if m.page == pageChat {
			return m.updateChat(msg)
		}
		// system page: only q/C/M
		return m, nil

	case thinkMsg, pasteTickMsg:
		// fire animations in chat
		if m.page == pageChat {
			return m.updateChat(msg)
		}
		return m, nil

	case tickMsg:
		m.blink = !m.blink
		return m, blinkCmd()

	case sysTickMsg:
		return m, sysTickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.MouseMsg:
		if m.page == pageChat {
			return m.updateChatMouse(msg)
		}
		return m, nil

	default:
		return m, nil
	}
}

// ── Tea.View ────────────────────────────────────────────────────────
func (m model) View() string {
	if m.page == pageSystem {
		return m.viewSystem()
	}
	return m.viewChat()
}
