package session

import (
	"damnTerminal/config"
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Side bool

func RenderFooter(cfg config.UIConfig, sessions []*Session, activeIdx int, width int) string {
	// Styles
	styleBar := lipgloss.NewStyle().
		Background(lipgloss.Color(cfg.BarBackgroundColor)).
		Foreground(lipgloss.Color(cfg.BarForegroundColor)).
		Padding(0, 1)

	styleTab := lipgloss.NewStyle().
		Background(lipgloss.Color(cfg.TabBackgroundColor)).
		Foreground(lipgloss.Color(cfg.TabForegroundColor)).
		Padding(0, 1)

	styleActiveTab := lipgloss.NewStyle().
		Background(lipgloss.Color(cfg.ActiveTabBackgroundColor)).
		Foreground(lipgloss.Color(cfg.ActiveTabForegroundColor)).
		Padding(0, 1)

	displayIdx := activeIdx + 1
	totalSessions := len(sessions)

	// Build Tab List
	var tabs []string
	for i := 0; i < totalSessions; i++ {
		t := fmt.Sprintf("Tab %d", i+1)
		if i == activeIdx {
			tabs = append(tabs, styleActiveTab.Render(t))
		} else {
			tabs = append(tabs, styleTab.Render(t))
		}
	}

	statusText := fmt.Sprintf("TaskFlow | %d/%d | Ctrl+N: New | Ctrl+B: Switch | Ctrl+Q: Quit", displayIdx, totalSessions)
	status := styleBar.Width(width - lipgloss.Width(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))).Render(statusText)

	footer := lipgloss.JoinHorizontal(lipgloss.Top, append(tabs, status)...)

	if lipgloss.Width(footer) < width {
		footer = lipgloss.NewStyle().Width(width).Background(lipgloss.Color("62")).Render(footer) // fill remaining
	}

	return footer
}

// RenderTopBar renders the top status bar
func RenderTopBar(cfg config.UIConfig, width int) string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color(cfg.BarBackgroundColor)).
		Foreground(lipgloss.Color(cfg.BarForegroundColor)).
		Padding(0, 1).
		Width(width)

	return style.Render("DamnTerminal - Go Terminal Multiplexer")
}

// RenderSideBar renders the side bar with session list
func RenderSideBar(cfg config.UIConfig, sessions []*Session, activeIdx int, height int) string {
	styleBar := lipgloss.NewStyle().
		Background(lipgloss.Color(cfg.BarBackgroundColor)).
		Foreground(lipgloss.Color(cfg.BarForegroundColor)).
		Width(20). // Fixed width for sidebar
		Height(height)

	styleTab := lipgloss.NewStyle().
		Background(lipgloss.Color(cfg.TabBackgroundColor)).
		Foreground(lipgloss.Color(cfg.TabForegroundColor)).
		Padding(0, 1).
		Width(20)

	styleActiveTab := lipgloss.NewStyle().
		Background(lipgloss.Color(cfg.ActiveTabBackgroundColor)).
		Foreground(lipgloss.Color(cfg.ActiveTabForegroundColor)).
		Padding(0, 1).
		Width(20)

	var tabs []string
	for i, _ := range sessions {
		t := fmt.Sprintf("Tab %d", i+1)
		if i == activeIdx {
			tabs = append(tabs, styleActiveTab.Render(t))
		} else {
			tabs = append(tabs, styleTab.Render(t))
		}
	}

	// Fill remaining height
	contentHeight := len(tabs) * 1 // Assuming 1 line per tab
	if contentHeight < height {
		remaining := height - contentHeight
		filler := styleBar.
			Height(remaining).
			Render("")
		tabs = append(tabs, filler)
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabs...)
}
