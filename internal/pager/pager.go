// Package pager provides a bubbletea-based terminal pager
// using the viewport bubble for scrolling rendered content.
package pager

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var barStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("7")).
	Background(lipgloss.Color("8"))

type model struct {
	viewport viewport.Model
	content  string
	ready    bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := 0
		footerHeight := 1
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	return m.viewport.View() + "\n" + m.statusView()
}

func (m model) statusView() string {
	pct := m.viewport.ScrollPercent() * 100
	info := fmt.Sprintf(" %3.0f%% ", pct)
	help := " q quit • ↑↓/jk scroll "

	w := m.viewport.Width - lipgloss.Width(info) - lipgloss.Width(help)
	if w < 0 {
		w = 0
	}
	gap := barStyle.Render(strings.Repeat(" ", w))
	return barStyle.Render(help) + gap + barStyle.Render(info)
}

// Run launches the pager with pre-rendered ANSI content.
func Run(content string) error {
	m := model{content: content}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// RunMarkdown renders markdown with glamour and launches the pager.
func RunMarkdown(markdown string) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}

	rendered, err := renderer.Render(markdown)
	if err != nil {
		return fmt.Errorf("render markdown: %w", err)
	}

	return Run(rendered)
}
