package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewSystem renders up to 4 servers as spokes around a central Hub.
// Any extra servers are listed below the graph.
func (m model) viewSystem() string {
	// Helper: style a server’s box
	renderNode := func(name, body string, err bool) string {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)
		if err {
			style = style.Foreground(lipgloss.Color("#FF0000"))
		}
		return style.Render(fmt.Sprintf("%s\n%s", name, body))
	}

	// Gather the server boxes (up to 4)
	var extra []string
	boxes := make([]string, len(m.servers))
	for i, srv := range m.servers {
		body := ""
		if srv.Err != nil {
			body = "ERROR"
		} else {
			d := srv.Data
			body = fmt.Sprintf("CPU: %.1f%%\nRAM: %.1f/%.1f MB",
				d.CpuUsagePercent,
				d.MemoryUsedMb, d.MemoryTotalMb,
			)
			if gpu := d.Gpu; gpu.Name != "" {
				body += fmt.Sprintf("\nGPU: %s (%.0f°C)", gpu.Name, gpu.TemperatureCelsius)
			} else {
				body += "\nGPU: n/a"
			}
		}
		box := renderNode(srv.URL, body, srv.Err != nil)
		if i < 4 {
			boxes[i] = box
		} else {
			// extra after the 4 spokes
			extra = append(extra, fmt.Sprintf("%s  %s", srv.URL, func() string {
				if srv.Err != nil {
					return "[ERROR]"
				}
				return fmt.Sprintf("CPU %.1f%%, RAM %.1fMB", srv.Data.CpuUsagePercent, srv.Data.MemoryUsedMb)
			}()))
		}
	}

	// Fill missing with empty strings so our slice has len=4
	for len(boxes) < 4 {
		boxes = append(boxes, "")
	}

	// Build the 5 rows of our graph:
	//    [0]     Top
	//      │
	// [3]─ Hub ─[1]
	//      │
	//    [2]     Bottom

	// Row 0:    center the top box
	row0 := lipgloss.PlaceHorizontal(
		m.width, lipgloss.Center,
		boxes[0],
	)

	// Row 1: a vertical connector under the top
	connV := lipgloss.NewStyle().Render("│")
	row1 := lipgloss.Place(
		m.width, 1,
		lipgloss.Center, lipgloss.Center,
		connV,
	)

	// Row 2: left box, hub, right box with horizontal lines
	horiz := lipgloss.NewStyle().Render("─")
	row2 := lipgloss.PlaceHorizontal(
		m.width, lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			boxes[3],
			horiz, horiz, // two dashes
			lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).Render(" HUB "),
			horiz, horiz,
			boxes[1],
		),
	)

	// Row 3: vertical connector above bottom
	row3 := row1

	// Row 4: bottom box centered
	row4 := row0
	if boxes[2] != "" {
		row4 = lipgloss.PlaceHorizontal(
			m.width, lipgloss.Center,
			boxes[2],
		)
	}

	// Combine the graph rows
	graph := strings.Join([]string{row0, row1, row2, row3, row4}, "\n")

	// If any extra servers, append them as a simple list
	extras := ""
	if len(extra) > 0 {
		extras = "\n\nAdditional servers:\n  " + strings.Join(extra, "\n  ")
	}

	// Footer hint
	footer := lipgloss.NewStyle().
		Faint(true).
		Align(lipgloss.Center).
		Render("C:Chat | M:System | q:Quit")

	return graph + extras + "\n\n" + footer
}
