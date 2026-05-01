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

// ─── ASCII banners ───────────────────────────────────────────────────

// Large banner (~48 cols)
const bannerLarge = "" +
	"  ██████╗ ██╗   ██╗██╗██╗   ██╗███████╗██████╗ \n" +
	" ██╔═══██╗██║   ██║██║██║   ██║██╔════╝██╔══██╗\n" +
	" ██║   ██║██║   ██║██║██║   ██║█████╗  ██████╔╝\n" +
	" ██║▄▄ ██║██║   ██║██║╚██╗ ██╔╝██╔══╝  ██╔══██╗\n" +
	" ╚██████╔╝╚██████╔╝██║ ╚████╔╝ ███████╗██║  ██║\n" +
	"  ╚══▀▀═╝  ╚═════╝ ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝"

// Small banner (~22 cols)
const bannerSmall = "" +
	" ╔═╗╦ ╦╦╦  ╦╔═╗╦═╗\n" +
	" ║═╬╗║ ║║╚╗╔╝║╣ ╠╦╝\n" +
	" ╚═╝╚╚═╝╩ ╚╝ ╚═╝╩╚═"

// ─── Layout breakpoints ─────────────────────────────────────────────

const (
	// Terminal width below which we hide the sidebar entirely
	breakpointCompact = 50
	// Terminal width below which we use narrow sidebar
	breakpointNarrow = 80
	// Content width below which we use the small ASCII banner
	breakpointBannerSmall = 55
)

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
}

// ─── Styles ──────────────────────────────────────────────────────────

type styles struct {
	sidebar      lipgloss.Style
	content      lipgloss.Style
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

// sidebarWidth returns the sidebar width for the given terminal width.
func sidebarWidth(termWidth int) int {
	if termWidth < breakpointCompact {
		return 0 // no sidebar
	}
	if termWidth < breakpointNarrow {
		return 18 // narrow
	}
	return 26 // normal
}

func newStyles(width, height int) *styles {
	sw := sidebarWidth(width)

	// Content panel fills the remaining space
	contentWidth := width - sw - 6 // borders + padding
	if sw == 0 {
		contentWidth = width - 4
	}
	if contentWidth < 10 {
		contentWidth = 10
	}

	contentHeight := height - 4
	if contentHeight < 6 {
		contentHeight = 6
	}

	sidebarStyle := lipgloss.NewStyle()
	if sw > 0 {
		sidebarStyle = sidebarStyle.
			Width(sw).
			Height(contentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 1)
	}

	return &styles{
		sidebar: sidebarStyle,
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
			Width(sw - 2),
		menuSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Background(lipgloss.Color("236")).
			PaddingLeft(1).
			Width(sw - 2),
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
			Foreground(lipgloss.Color("241")),
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
	s := newStyles(m.width, m.height)
	sw := sidebarWidth(m.width)

	// ── Content area ─────────────────────────────────────────
	var contentStr string
	switch m.menuCursor {
	case 0:
		contentStr = m.renderHome(s)
	case 1:
		contentStr = m.renderVehicles(s)
	case 2:
		contentStr = m.renderWork(s)
	case 3:
		contentStr = m.renderSettings(s)
	}
	content := s.content.Render(contentStr)

	var layout string

	if sw > 0 {
		// ── Sidebar ──────────────────────────────────────────
		logoLine := s.logo.Render("QUIVER")
		versionLine := s.version.Render(fmt.Sprintf("v%s", m.version))
		sep := s.dim.Render(strings.Repeat("─", sw-4))

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

		layout = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	} else {
		// ── Compact: no sidebar, show current view name at top ─
		indicator := s.dim.Render("◂ ") +
			s.highlight.Render(menuItems[m.menuCursor]) +
			s.dim.Render(" ▸")
		layout = indicator + "\n" + content
	}

	// ── Help bar ─────────────────────────────────────────────
	helpText := "  ↑/↓ navigate • q quit"
	if sw == 0 {
		helpText = "  ↑/↓ switch view • q quit"
	}
	help := s.helpBar.Render(helpText)

	full := layout + "\n" + help

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}

// ─── Views ───────────────────────────────────────────────────────────

func (m *model) renderHome(s *styles) string {
	// Choose banner size based on available content width
	contentW := m.width - sidebarWidth(m.width) - 6
	if sidebarWidth(m.width) == 0 {
		contentW = m.width - 4
	}

	var header string
	if contentW >= breakpointBannerSmall {
		header = s.highlight.Render(bannerLarge)
	} else if contentW >= 24 {
		header = s.highlight.Render(bannerSmall)
	} else {
		header = s.title.Render("QUIVER")
	}

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

	return header + "\n" + versionLine + "\n\n" + welcome + "\n" + infoBox
}

func (m *model) renderVehicles(s *styles) string {
	title := s.title.Render("Vehicles")
	desc := s.subtitle.Render("Vehicle management")
	placeholder := s.dim.Render("No vehicles registered yet.")

	return title + "\n" + desc + "\n\n" + placeholder
}

func (m *model) renderWork(s *styles) string {
	title := s.title.Render("Work")
	desc := s.subtitle.Render("Work log & tasks")
	placeholder := s.dim.Render("No work entries yet.")

	return title + "\n" + desc + "\n\n" + placeholder
}

func (m *model) renderSettings(s *styles) string {
	title := s.title.Render("Settings")
	desc := s.subtitle.Render("Application configuration")
	placeholder := s.dim.Render("No settings available yet.")

	return title + "\n" + desc + "\n\n" + placeholder
}
