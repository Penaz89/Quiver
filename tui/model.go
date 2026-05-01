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
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ssh"
)

// Re-export types so main doesn't need to import bubbletea directly.
type Model = tea.Model
type ProgramOption = tea.ProgramOption

// ─── Menu items ──────────────────────────────────────────────────────

var menuItems = []string{
	"HOME",
	"VEHICLES",
	"WORK",
	"SETTINGS",
}

// ─── Model ───────────────────────────────────────────────────────────

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
	menuCursor int
	dataDir    string
	version    string

	// Styles (computed based on window size)
	styles *styles
}

// ─── Styles ──────────────────────────────────────────────────────────

const menuWidth = 26

type styles struct {
	// Layout
	sidebar      lipgloss.Style
	content      lipgloss.Style
	// Text
	logo         lipgloss.Style
	version      lipgloss.Style
	menuNormal   lipgloss.Style
	menuSelected lipgloss.Style
	title        lipgloss.Style
	subtitle     lipgloss.Style
	info         lipgloss.Style
	highlight    lipgloss.Style
	dim          lipgloss.Style
	infoBox      lipgloss.Style
	status       lipgloss.Style
	helpBar      lipgloss.Style
}

func newStyles(width, height int) *styles {
	contentWidth := width - menuWidth - 4 // account for borders/padding
	if contentWidth < 20 {
		contentWidth = 20
	}
	contentHeight := height - 4
	if contentHeight < 10 {
		contentHeight = 10
	}

	return &styles{
		sidebar: lipgloss.NewStyle().
			Width(menuWidth).
			Height(contentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 1),
		content: lipgloss.NewStyle().
			Width(contentWidth).
			Height(contentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2),

		logo: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")),
		version: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
		menuNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(1).
			Width(menuWidth - 4),
		menuSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Background(lipgloss.Color("236")).
			PaddingLeft(1).
			Width(menuWidth - 4),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")).
			Italic(true).
			MarginBottom(1),
		info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true),
		dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		infoBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			MarginTop(1),
		status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),
		helpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1),
	}
}

// ─── Constructor ─────────────────────────────────────────────────────

// NewModel creates a new TUI model for the given SSH session.
func NewModel(s ssh.Session, dataDir, version string) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	user := s.User()

	m := &model{
		user:       user,
		term:       pty.Term,
		width:      pty.Window.Width,
		height:     pty.Window.Height,
		bg:         "light",
		menuCursor: 0,
		dataDir:    dataDir,
		version:    version,
		styles:     newStyles(pty.Window.Width, pty.Window.Height),
	}
	return m, []tea.ProgramOption{}
}

// ─── Bubble Tea interface ────────────────────────────────────────────

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
		m.styles = newStyles(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < len(menuItems)-1 {
				m.menuCursor++
			}
		}
	}
	return m, nil
}

func (m *model) View() tea.View {
	s := m.styles

	// ── Sidebar ──────────────────────────────────────────────
	logoLine := s.logo.Render("QUIVER")
	versionLine := s.version.Render(fmt.Sprintf("v%s", m.version))
	sep := s.dim.Render(strings.Repeat("─", menuWidth-2))

	var menuLines []string
	for i, item := range menuItems {
		if i == m.menuCursor {
			menuLines = append(menuLines, s.menuSelected.Render("▸ "+item))
		} else {
			menuLines = append(menuLines, s.menuNormal.Render("  "+item))
		}
	}
	menu := strings.Join(menuLines, "\n")

	userLine := s.status.Render("● ") + s.info.Render(m.user)

	sidebarContent := logoLine + "\n" + versionLine + "\n" + sep + "\n\n" + menu + "\n\n" + sep + "\n" + userLine
	sidebar := s.sidebar.Render(sidebarContent)

	// ── Content area ─────────────────────────────────────────
	var contentStr string
	switch m.menuCursor {
	case 0:
		contentStr = m.viewHome()
	case 1:
		contentStr = m.viewVehicles()
	case 2:
		contentStr = m.viewWork()
	case 3:
		contentStr = m.viewSettings()
	}
	content := s.content.Render(contentStr)

	// ── Compose layout ───────────────────────────────────────
	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	// ── Help bar ─────────────────────────────────────────────
	help := s.helpBar.Render("  ↑/↓ navigate • q quit")

	full := layout + "\n" + help

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}

// ─── Views ───────────────────────────────────────────────────────────

func (m *model) viewHome() string {
	s := m.styles

	banner := s.highlight.Render(
		"  ██████╗ ██╗   ██╗██╗██╗   ██╗███████╗██████╗ \n" +
			" ██╔═══██╗██║   ██║██║██║   ██║██╔════╝██╔══██╗\n" +
			" ██║   ██║██║   ██║██║██║   ██║█████╗  ██████╔╝\n" +
			" ██║▄▄ ██║██║   ██║██║╚██╗ ██╔╝██╔══╝  ██╔══██╗\n" +
			" ╚██████╔╝╚██████╔╝██║ ╚████╔╝ ███████╗██║  ██║\n" +
			"  ╚══▀▀═╝  ╚═════╝ ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝")

	versionLine := s.version.Render(fmt.Sprintf("  v%s", m.version))

	welcome := s.info.Render(fmt.Sprintf(
		"Welcome back, %s!",
		s.highlight.Render(m.user),
	))

	sessionInfo := fmt.Sprintf(
		"%s  %s\n%s  %dx%d\n%s  %s\n%s  %s\n%s  %s",
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
	infoBox := s.infoBox.Render(sessionInfo)

	return banner + "\n" + versionLine + "\n\n" + welcome + "\n" + infoBox
}

func (m *model) viewVehicles() string {
	s := m.styles

	title := s.title.Render("🚗  Vehicles")
	desc := s.subtitle.Render("Vehicle management")
	placeholder := s.dim.Render("No vehicles registered yet.")

	return title + "\n" + desc + "\n\n" + placeholder
}

func (m *model) viewWork() string {
	s := m.styles

	title := s.title.Render("🔧  Work")
	desc := s.subtitle.Render("Work log & tasks")
	placeholder := s.dim.Render("No work entries yet.")

	return title + "\n" + desc + "\n\n" + placeholder
}

func (m *model) viewSettings() string {
	s := m.styles

	title := s.title.Render("⚙️  Settings")
	desc := s.subtitle.Render("Application configuration")
	placeholder := s.dim.Render("No settings available yet.")

	return title + "\n" + desc + "\n\n" + placeholder
}
