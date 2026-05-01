// Quiver - An SSH TUI Application
// Copyright (C) 2026  penaz
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ssh"
)

// Re-export types so main doesn't need to import bubbletea directly.
type Model = tea.Model
type ProgramOption = tea.ProgramOption

// view is the application's view state.
type view int

const (
	viewHome view = iota
)

// model holds the TUI application state.
type model struct {
	// Session info
	user    string
	term    string
	width   int
	height  int
	bg      string
	profile string

	// App state
	currentView view
	dataDir     string
	version     string

	// Styles
	styles *styles
}

// styles holds all lipgloss styles used by the TUI.
type styles struct {
	title     lipgloss.Style
	subtitle  lipgloss.Style
	info      lipgloss.Style
	highlight lipgloss.Style
	dim       lipgloss.Style
	border    lipgloss.Style
	status    lipgloss.Style
}

func newStyles() *styles {
	return &styles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")).
			Italic(true),
		info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true),
		dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			MarginTop(1),
		status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),
	}
}

// NewModel creates a new TUI model for the given SSH session.
func NewModel(s ssh.Session, dataDir, version string) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	user := s.User()

	m := &model{
		user:        user,
		term:        pty.Term,
		width:       pty.Window.Width,
		height:      pty.Window.Height,
		bg:          "light",
		currentView: viewHome,
		dataDir:     dataDir,
		version:     version,
		styles:      newStyles(),
	}
	return m, []tea.ProgramOption{}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.ColorProfileMsg:
		m.profile = msg.String()
	case tea.BackgroundColorMsg:
		if msg.IsDark() {
			m.bg = "dark"
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() tea.View {
	s := m.styles

	// Header
	header := s.title.Render("🏹 Q U I V E R")
	subtitle := s.subtitle.Render(fmt.Sprintf("v%s", m.version))

	// Welcome message
	welcome := s.info.Render(fmt.Sprintf("Welcome, %s!", s.highlight.Render(m.user)))

	// Session info box
	sessionInfo := fmt.Sprintf(
		"%s %s\n%s %dx%d\n%s %s\n%s %s\n%s %s",
		s.dim.Render("Terminal:"),
		s.info.Render(m.term),
		s.dim.Render("Window:"),
		m.width, m.height,
		s.dim.Render("Background:"),
		s.info.Render(m.bg),
		s.dim.Render("Color Profile:"),
		s.info.Render(m.profile),
		s.dim.Render("Data Directory:"),
		s.info.Render(m.dataDir),
	)
	infoBox := s.border.Render(sessionInfo)

	// Status line
	status := s.status.Render("● Connected")

	// Help
	help := s.dim.Render("Press 'q' to quit")

	// Compose the full view
	content := fmt.Sprintf(
		"%s  %s\n\n%s\n%s\n\n%s\n\n%s",
		header, subtitle,
		welcome,
		infoBox,
		status,
		help,
	)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
