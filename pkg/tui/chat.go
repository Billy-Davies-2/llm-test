package tui

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Billy-Davies-2/tui-chat/pkg/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Chat-Page Commands ────────────────────────────────────────────────

func thinkCmd() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
		return thinkMsg{}
	})
}

func pasteTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return pasteTickMsg{}
	})
}

// codeStyle highlights the input area
var codeStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#002b36")).
	Foreground(lipgloss.Color("#93a1a1")).
	Padding(0, 1)

// chunkByWidth splits a string into rune-chunks of at most width.
func chunkByWidth(s string, width int) []string {
	var out []string
	r := []rune(s)
	for len(r) > 0 {
		n := width
		if n > len(r) {
			n = len(r)
		}
		out = append(out, string(r[:n]))
		r = r[n:]
	}
	return out
}

// wrapText wraps a text at maxWidth, indenting subsequent lines.
func wrapText(text string, maxWidth int, indent string) string {
	if maxWidth <= 0 {
		return text
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if len(line) <= maxWidth {
			lines = append(lines, line)
			continue
		}
		for len(line) > 0 {
			cut := maxWidth
			for cut > 0 && (cut > len(line) || line[cut-1] != ' ') {
				cut--
			}
			if cut <= 0 {
				cut = maxWidth
			}
			lines = append(lines, line[:cut])
			line = indent + strings.TrimLeft(line[cut:], " ")
		}
	}
	return strings.Join(lines, "\n")
}

// ── UpdateChat (keys) ──────────────────────────────────────────────────

func (m model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	cur := &m.tabs[m.currentTab]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		log.Printf("chat Key %q insert=%v lastKey=%q", s, m.insertMode, m.lastKey)

		// NORMAL vs INSERT
		if !m.insertMode {
			// yy yank
			if s == "y" {
				if m.lastKey == "y" && len(cur.messages) > 0 {
					clipboard.WriteAll(cur.messages[len(cur.messages)-1])
				}
				m.lastKey = ""
				return m, nil
			}
			// g prefix for gt/gT
			if s == "g" {
				m.lastKey = "g"
				return m, nil
			}
			if s == "t" && m.lastKey == "g" {
				m.currentTab = (m.currentTab + 1) % len(m.tabs)
				m.lastKey = ""
				return m, nil
			}
			if s == "T" && m.lastKey == "g" {
				m.currentTab = (m.currentTab - 1 + len(m.tabs)) % len(m.tabs)
				m.lastKey = ""
				return m, nil
			}

			switch s {
			case "q":
				return m, tea.Quit
			case "i":
				m.insertMode = true
				return m, nil
			case "p", "P", tea.KeyCtrlV.String():
				if clip, _ := clipboard.ReadAll(); clip != "" {
					sw := 0
					if m.showSidebar {
						sw = 18
					}
					iw := m.width - sw - 4
					if iw < 1 {
						iw = 1
					}
					m.pasteQueue = chunkByWidth(clip, iw)
					return m, pasteTick()
				}
				return m, nil
			case "z":
				m.showSidebar = !m.showSidebar
				return m, nil
			case "j":
				if m.showSidebar {
					m.currentTab = (m.currentTab + 1) % len(m.tabs)
				}
				return m, nil
			case "k":
				if m.showSidebar {
					m.currentTab = (m.currentTab - 1 + len(m.tabs)) % len(m.tabs)
				}
				return m, nil
			case "T":
				n := len(m.tabs) + 1
				m.tabs = append(m.tabs, tab{title: fmt.Sprintf("Tab %d", n), messages: []string{"New tab"}})
				m.currentTab = len(m.tabs) - 1
				return m, nil
			case "d":
				if m.lastKey == "d" && len(m.tabs) > 1 {
					m.tabs = slices.Delete(m.tabs, m.currentTab, m.currentTab+1)
					if len(m.tabs) == 0 {
						return m, tea.Quit
					}
					if m.currentTab >= len(m.tabs) {
						m.currentTab = len(m.tabs) - 1
					}
				}
				m.lastKey = "d"
				return m, nil
			}
			m.lastKey = ""
			return m, nil
		}

		// INSERT MODE
		if s == "esc" {
			m.insertMode = false
			return m, nil
		}
		switch s {
		case "enter":
			cur.messages = append(cur.messages, "You: "+cur.input)
			cur.input = ""
			cur.thinking = true
			cur.dots = 0
			return m, thinkCmd()
		case "backspace":
			if len(cur.input) > 0 {
				_, sz := utf8.DecodeLastRuneInString(cur.input)
				cur.input = cur.input[:len(cur.input)-sz]
			}
			return m, nil
		default:
			if len(msg.Runes) > 0 {
				cur.input += string(msg.Runes)
			}
			return m, nil
		}

	case pasteTickMsg:
		if len(m.pasteQueue) > 0 {
			line := m.pasteQueue[0]
			m.pasteQueue = m.pasteQueue[1:]
			if len(cur.input) > 0 {
				cur.input += "\n"
			}
			cur.input += line
			return m, pasteTick()
		}
		return m, nil

	case thinkMsg:
		if cur.thinking {
			if cur.dots < 3 {
				cur.dots++
				return m, thinkCmd()
			}
			cur.messages = append(cur.messages, "AI: epic response")
			cur.thinking = false
			cur.dots = 0
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.MouseMsg:
		return m.updateChatMouse(msg)
	}

	return m, nil
}

// ── Mouse for Chat ────────────────────────────────────────────────────

func (m model) updateChatMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// right-click to paste
	if msg.Action == tea.MouseActionPress &&
		msg.Button == tea.MouseButtonRight &&
		!m.insertMode {
		if clip, _ := clipboard.ReadAll(); clip != "" {
			sw := 0
			if m.showSidebar {
				sw = 18
			}
			iw := m.width - sw - 4
			if iw < 1 {
				iw = 1
			}
			m.pasteQueue = chunkByWidth(clip, iw)
			return m, pasteTick()
		}
		return m, nil
	}

	// left-click sidebar select/close/drag
	if msg.Action == tea.MouseActionPress &&
		msg.Button == tea.MouseButtonLeft &&
		m.showSidebar {
		const padX, padY = 1, 1
		const sw = 16
		if msg.X >= padX && msg.X < padX+sw && msg.Y >= padY {
			idx := msg.Y - padY
			line := fmt.Sprintf("> %s [x]", m.tabs[idx].title)
			closeX := padX + len(line) - 3
			if msg.X >= closeX {
				m.tabs = slices.Delete(m.tabs, idx, idx+1)
				if len(m.tabs) == 0 {
					return m, tea.Quit
				}
				if m.currentTab >= len(m.tabs) {
					m.currentTab = len(m.tabs) - 1
				}
			} else {
				m.dragging = true
				m.dragIndex = idx
			}
		}
		return m, nil
	}

	// drag release → reorder
	if msg.Action == tea.MouseActionRelease &&
		msg.Button == tea.MouseButtonLeft &&
		m.dragging {
		m.dragging = false
		const padY = 1
		if msg.Y >= padY {
			target := msg.Y - padY
			if target >= 0 && target < len(m.tabs) && target != m.dragIndex {
				t := m.tabs[m.dragIndex]
				m.tabs = slices.Delete(m.tabs, m.dragIndex, m.dragIndex+1)
				if target > m.dragIndex {
					target--
				}
				m.tabs = slices.Insert(m.tabs, target, t)
				// adjust currentTab
				switch {
				case m.currentTab == m.dragIndex:
					m.currentTab = target
				case m.currentTab > m.dragIndex && m.currentTab <= target:
					m.currentTab--
				case m.currentTab < m.dragIndex && m.currentTab >= target:
					m.currentTab++
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// ── Render Chat ───────────────────────────────────────────────────────

func (m model) viewChat() string {
	// Sidebar
	sb := ""
	if m.showSidebar {
		var lines []string
		for i, t := range m.tabs {
			mrk := " "
			if i == m.currentTab {
				mrk = ">"
			}
			lines = append(lines, lipgloss.NewStyle().
				BorderBottom(true).
				BorderForeground(lipgloss.Color("#00FF00")).
				Render(fmt.Sprintf("%s %s [x]", mrk, t.title)))
		}
		sb = lipgloss.NewStyle().
			Width(16).
			Padding(1).
			Background(lipgloss.Color("#000")).
			Foreground(lipgloss.Color("#0f0")).
			Render(strings.Join(lines, "\n"))
	}

	// Chat area dimensions
	head := 2
	chatH := m.height - head - 2
	chatW := m.width
	if m.showSidebar {
		chatW -= 18
	}
	innerW := chatW - 4
	if innerW < 10 {
		innerW = 10
	}

	// Build chat lines
	cur := m.tabs[m.currentTab]
	var chatLines []string
	style := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Left)
	for _, msg := range cur.messages {
		chatLines = append(chatLines, style.Render(msg))
	}
	if cur.thinking {
		dots := strings.Repeat(".", cur.dots)
		chatLines = append(chatLines, style.Render("AI is thinking"+dots))
	}

	if len(chatLines) > chatH {
		chatLines = chatLines[len(chatLines)-chatH:]
	}

	// Render input
	inputSt := codeStyle.Width(innerW).Align(lipgloss.Left)
	parts := strings.Split(cur.input, "\n")
	for i, l := range parts {
		pref := "> "
		if i > 0 {
			pref = "  "
		}
		line := pref + l
		if i == len(parts)-1 && m.blink {
			line += "_"
		}
		chatLines = append(chatLines, inputSt.Render(line))
	}

	chatPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0f0")).
		Padding(1, 2).
		Background(lipgloss.Color("#000")).
		Width(chatW).
		Height(chatH).
		Render(strings.Join(chatLines, "\n"))

	panel := chatPane
	if m.showSidebar {
		panel = lipgloss.JoinHorizontal(lipgloss.Top, sb, chatPane)
	}

	footer := lipgloss.NewStyle().
		Faint(true).
		Align(lipgloss.Center).
		Width(m.width).
		Render("q:Quit | M:Metrics | i:Insert | dd:Close | p:Paste | yy:Copy | gt/gT:Tabs | z:Sidebar")

	return panel + "\n" + footer
}
