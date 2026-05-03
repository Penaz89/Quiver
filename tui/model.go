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
	"io"
	"net/http"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ssh"
	"github.com/penaz/quiver/storage"
)

// Re-export types so main doesn't need to import bubbletea directly.
type Model = tea.Model
type ProgramOption = tea.ProgramOption

// ─── Menu items ──────────────────────────────────────────────────────

var defaultMenuItems = []string{
	"HOME",
	"HABITS",
	"VEHICLES",
	"FINANCES",
	"WEATHER",
	"SETTINGS",
	"LOGOUT",
}

// ─── ASCII banners ───────────────────────────────────────────────────

// Sidebar logo (~22 cols) for normal sidebar
const sidebarLogo = "" +
	"╔═╗ ╦ ╦ ╦ ╦  ╦ ╔═╗ ╦═╗\n" +
	"║ ║ ║ ║ ║ ╚╗╔╝ ╠═  ╠╦╝\n" +
	"╚═╣ ╚═╝ ╩  ╚╝  ╚═╝ ╩╚═"

// Sidebar logo narrow (~14 cols) for narrow sidebar
const sidebarLogoNarrow = "" +
	"  ╔═╗ ╦ ╦ ╦\n" +
	"  ║ ║ ║ ║ ║\n" +
	"  ╚═╣ ╚═╝ ╩\n" +
	"           \n" +
	"  ╦  ╦ ╔═╗ ╦═╗\n" +
	"  ╚╗╔╝ ╠═  ╠╦╝\n" +
	"   ╚╝  ╚═╝ ╩╚═"

// ─── Layout breakpoints ─────────────────────────────────────────────

const (
	// Terminal width below which we hide the sidebar entirely
	breakpointCompact = 50
	// Terminal width below which we use narrow sidebar
	breakpointNarrow = 80
	// Content width below which we use the small ASCII banner
	breakpointBannerSmall = 50
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

	// Auth state
	isLoggedIn         bool
	isRegistering      bool
	loginForm          [2]string
	loginCursor        int
	loginError         string
	baseDataDir        string
	isAdmin            bool
	isChangingPassword bool
	changePasswordForm [2]string
	changePasswordCur  int
	changePasswordErr  string

	// App state
	menuItems    []string
	menuCursor   int
	focusContent bool
	dataDir      string
	version      string
	lang         string // "en" or "it"
	settings     storage.Settings
	vp           viewport.Model

	// Admin state
	adminUsers       []storage.User
	adminUserCursor  int
	adminForm        [2]string
	adminFormCursor  int
	adminIsAdding    bool
	adminIsEditing   bool
	adminIsDeleting  bool
	adminError       string

	// Vehicle state
	vehicles             []storage.Vehicle
	vehicleSection       vehicleSection
	vehicleSectionCursor int
	vehicleView          vehicleSubView
	vehicleCursor        int
	formFields           [fCount]string
	formCursor           int
	editIndex            int

	// Insurance state
	insurances      []storage.Insurance
	insuranceCursor int
	insFormFields   [insFCount]string
	insFormCursor   int
	insPickerMode   bool
	insPickerCursor int

	// Finances state
	finSection    finSection
	finMenuCursor int
	finView       finSubView
	
	// Housing state
	housing       []storage.Housing
	houseCursor   int
	houseForm     [houseFCount]string
	houseFormCur  int
	houseEditIdx  int
	
	// Holidays state
	holidays      []storage.Holiday
	holiCursor    int
	holiForm      [holiFCount]string
	holiFormCur   int
	holiEditIdx   int
	
	// Subscriptions state
	subs        []storage.Subscription
	subCursor   int
	subForm     [subFCount]string
	subFormCur  int
	subEditIdx  int

	// Settings state
	settingsSection    setSection
	settingsMenuCursor int
	settingsCursor     int

	// Habits state
	habits          []storage.Habit
	habitCursor     int
	habitIsAdding   bool
	habitIsDeleting bool
	habitForm       string

	// Weather
	weatherData string
}

// ─── Styles ──────────────────────────────────────────────────────────

type styles struct {
	sidebar        lipgloss.Style
	content        lipgloss.Style
	contentFocused lipgloss.Style
	logo           lipgloss.Style
	version        lipgloss.Style
	menuNormal     lipgloss.Style
	menuSelected   lipgloss.Style
	menuActiveDim  lipgloss.Style
	title          lipgloss.Style
	subtitle       lipgloss.Style
	info           lipgloss.Style
	highlight      lipgloss.Style
	dim            lipgloss.Style
	infoBox        lipgloss.Style
	status         lipgloss.Style
	helpBar        lipgloss.Style
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
		contentFocused: lipgloss.NewStyle().
			Width(contentWidth).
			Height(contentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 2),
		logo: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")),
		version: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
		menuNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(1),
		menuSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("63")).
			Bold(true).
			PaddingLeft(1),
		menuActiveDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("237")).
			Bold(true).
			PaddingLeft(1),
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
		user:        user,
		term:        pty.Term,
		width:       pty.Window.Width,
		height:      pty.Window.Height,
		bg:          "light",
		menuItems:   append([]string(nil), defaultMenuItems...),
		menuCursor:  0,
		dataDir:     dataDir,
		baseDataDir: dataDir,
		version:     version,
		lang:        "en",
		vp:          viewport.New(),
	}
	
	if user != "" && user != "anonymous" {
		m.loginForm[0] = user
		m.loginCursor = 1
	}
	
	// Disable up/down in viewport so it doesn't conflict with forms
	m.vp.KeyMap.Up.SetEnabled(false)
	m.vp.KeyMap.Down.SetEnabled(false)
	
	m.updateMenuLabels()
	return m, []tea.ProgramOption{}
}

func (m *model) loadUserData() {
	m.vehicles, _ = storage.LoadVehicles(m.dataDir)
	m.insurances, _ = storage.LoadInsurance(m.dataDir)
	m.settings = storage.LoadSettings(m.dataDir)
	m.subs, _ = storage.LoadSubscriptions(m.dataDir)
	m.housing, _ = storage.LoadHousing(m.dataDir)
	m.holidays, _ = storage.LoadHolidays(m.dataDir)
	m.habits, _ = storage.LoadHabits(m.dataDir)
	
	if m.settings.Language != "" {
		m.lang = m.settings.Language
	}
	m.updateMenuLabels()
}

// ─── Bubble Tea interface ────────────────────────────────────────────

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		m.vp.Init(),
		fetchWeatherCmd(m.settings.WeatherLoc),
	)
}

type weatherMsg string

func fetchWeatherCmd(loc string) tea.Cmd {
	return func() tea.Msg {
		if loc == "" {
			return weatherMsg("No weather location set")
		}
		
		req, _ := http.NewRequest("GET", "https://wttr.in/"+loc+"?2Q", nil)
		req.Header.Set("User-Agent", "curl/7.68.0")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return weatherMsg("Weather unavailable")
		}
		defer resp.Body.Close()
		
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return weatherMsg("Error reading weather")
		}
		return weatherMsg(string(b))
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case weatherMsg:
		m.weatherData = string(msg)
		return m, nil
	case tea.ColorProfileMsg:
		m.profile = msg.String()
	case tea.BackgroundColorMsg:
		if msg.IsDark() {
			m.bg = "dark"
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.MouseMsg:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		// Always allow ctrl+c
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if !m.isLoggedIn {
			return m.updateLogin(msg)
		}

		// Handle viewport scrolling explicitly
		if msg.String() == "pgup" {
			m.vp.HalfPageUp()
			return m, nil
		}
		if msg.String() == "pgdown" {
			m.vp.HalfPageDown()
			return m, nil
		}

		if m.isAdmin {
			if m.menuCursor == 0 && m.focusContent {
				return m.updateAdminUsers(msg)
			}
		} else {
			if m.focusContent {
				item := m.menuItems[m.menuCursor]
				if item == t(m.lang, "menu.vehicles") {
					return m.updateVehicleSection(msg)
				} else if item == t(m.lang, "menu.finances") {
					return m.updateFinances(msg)
				} else if item == t(m.lang, "menu.habits") {
					return m.updateHabits(msg)
				} else if item == t(m.lang, "menu.settings") {
					return m.updateSettings(msg)
				}
			}
		}

		// Toggle focus between sidebar and content
		switch msg.String() {
		case "tab", "enter", "right":
			if !m.focusContent {
				item := m.menuItems[m.menuCursor]
				if item == t(m.lang, "menu.logout") {
					m.isLoggedIn = false
					m.isAdmin = false
					m.menuCursor = 0
					m.user = ""
					m.loginForm = [2]string{}
					m.loginCursor = 0
					m.loginError = ""
					m.vehicles = nil
					m.insurances = nil
					m.subs = nil
					m.housing = nil
					m.holidays = nil
					m.updateMenuLabels()
					return m, nil
				}
				m.focusContent = true
				return m, nil
			}
		case "q":
			if !m.focusContent {
				return m, tea.Quit
			}
		case "esc", "left":
			m.focusContent = false
			return m, nil
		}

		// Sidebar navigation (only when sidebar has focus)
		if !m.focusContent {
			switch msg.String() {
			case "up", "k":
				if m.menuCursor > 0 {
					m.menuCursor--
				}
			case "down", "j":
				if m.menuCursor < len(m.menuItems)-1 {
					m.menuCursor++
				}
			}
		}
	}
	return m, nil
}

func (m *model) View() tea.View {
	s := newStyles(m.width, m.height)
	
	if !m.isLoggedIn {
		if m.isChangingPassword {
			v := tea.NewView(m.renderChangePasswordView(s))
			v.AltScreen = true
			return v
		}
		v := tea.NewView(m.renderLoginView(s))
		v.AltScreen = true
		return v
	}
	
	sw := sidebarWidth(m.width)

	// ── Content area ─────────────────────────────────────────
	var contentStr string
	if m.isAdmin {
		switch m.menuCursor {
		case 0:
			contentStr = m.renderAdminUsersView(s)
		}
	} else {
		item := m.menuItems[m.menuCursor]
		if item == t(m.lang, "menu.home") {
			contentStr = m.renderHome(s)
		} else if item == t(m.lang, "menu.vehicles") {
			contentStr = m.renderVehiclesView(s)
		} else if item == t(m.lang, "menu.finances") {
			contentStr = m.renderFinancesView(s)
		} else if item == t(m.lang, "menu.habits") {
			contentStr = m.renderHabitsView(s)
		} else if item == t(m.lang, "menu.weather") {
			contentStr = m.renderWeatherView(s)
		} else if item == t(m.lang, "menu.settings") {
			contentStr = m.renderSettingsView(s)
		}
	}
	// Configure viewport size based on the layout
	contentWidth := m.width - sw - 6 // borders + padding
	if sw == 0 {
		contentWidth = m.width - 4
	}
	if contentWidth < 10 {
		contentWidth = 10
	}

	contentHeight := m.height - 4
	if contentHeight < 6 {
		contentHeight = 6
	}

	// Inner dimensions for the viewport (excluding borders and padding)
	vpWidth := contentWidth - 4
	if vpWidth < 1 {
		vpWidth = 1
	}
	vpHeight := contentHeight - 4
	if vpHeight < 1 {
		vpHeight = 1
	}

	m.vp.SetWidth(vpWidth)
	m.vp.SetHeight(vpHeight)

	// Set content and get rendered viewport
	m.vp.SetContent(contentStr)
	
	contentStyle := s.content
	if m.focusContent {
		contentStyle = s.contentFocused
	}
	content := contentStyle.Render(m.vp.View())

	var layout string

	if sw > 0 {
		// ── Sidebar ──────────────────────────────────────────
		var logoLine string
		if sw >= 26 {
			logoLine = s.logo.Render(sidebarLogo)
		} else {
			logoLine = s.logo.Render(sidebarLogoNarrow)
		}
		versionLine := s.version.Render(fmt.Sprintf("v%s", m.version))
		sep := s.dim.Render(strings.Repeat("─", sw-4))

		var menuLines []string
		for i, item := range m.menuItems {
			if i == m.menuCursor {
				menuLines = append(menuLines, s.menuSelected.Width(sw-4).Render(item))
			} else {
				menuLines = append(menuLines, s.menuNormal.Width(sw-4).Render(item))
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
			s.highlight.Render(m.menuItems[m.menuCursor]) +
			s.dim.Render(" ▸")
		layout = indicator + "\n" + content
	}

	// ── Help bar ─────────────────────────────────────────────
	var helpText string
	if sw == 0 {
		helpText = fmt.Sprintf("  ↑/↓ %s • q %s", t(m.lang, "help.navigate"), t(m.lang, "help.quit"))
	} else if m.focusContent {
		helpText = fmt.Sprintf("  ←: %s • PgUp/PgDn: %s • %s", t(m.lang, "help.goBack"), "scroll", t(m.lang, "help.contentFocused"))
	} else {
		helpText = fmt.Sprintf("  ↑/↓ %s • →: %s • q %s", t(m.lang, "help.navigate"), t(m.lang, "help.enter"), t(m.lang, "help.quit"))
	}
	help := s.helpBar.Render(helpText)

	full := layout + "\n" + help

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}

// ─── Views ───────────────────────────────────────────────────────────

func (m *model) renderHome(s *styles) string {
	welcome := s.info.Render(fmt.Sprintf(
		t(m.lang, "home.welcome"),
		s.highlight.Render(m.user),
	))

	return welcome
}

func (m *model) renderWeatherView(s *styles) string {
	weatherBox := s.dim.Render("Loading weather...")
	if m.weatherData != "" {
		weatherBox = m.weatherData
	}
	
	// Remove trailing blank lines from weather data to keep it compact
	weatherBox = strings.TrimRight(weatherBox, "\n")
	
	weatherWidget := lipgloss.NewStyle().
		Padding(1, 2).
		Render(weatherBox)

	return weatherWidget
}

// renderVehicles is now in vehicles.go as renderVehiclesView



// updateMenuLabels refreshes menu item labels based on the current language.
func (m *model) updateMenuLabels() {
	if m.isAdmin {
		m.menuItems = []string{
			t(m.lang, "menu.users"),
			t(m.lang, "menu.logout"),
		}
	} else {
		m.menuItems = []string{
			t(m.lang, "menu.home"),
			t(m.lang, "menu.habits"),
			t(m.lang, "menu.vehicles"),
			t(m.lang, "menu.finances"),
			t(m.lang, "menu.weather"),
			t(m.lang, "menu.settings"),
			t(m.lang, "menu.logout"),
		}
	}
}
