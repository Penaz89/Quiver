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
	"github.com/penaz/quiver/storage"
)

// ─── Language options ────────────────────────────────────────────────

var langOptions = []struct {
	code  string
	flag  string
}{
	{"en", "EN"},
	{"it", "IT"},
}

type setSection int

const (
	sSectionMenu setSection = iota
	sSectionLang
	sSectionWeather
	sSectionTheme
	sSectionWorkspace
)

// ─── Settings update ─────────────────────────────────────────────────

func (m *model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settingsSection == sSectionMenu {
		switch msg.String() {
		case "up", "k":
			if m.settingsMenuCursor > 0 {
				m.settingsMenuCursor--
			}
		case "down", "j":
			if m.settingsMenuCursor < 3 { // 4 items
				m.settingsMenuCursor++
			}
		case "enter", "right":
			m.settingsSection = setSection(m.settingsMenuCursor + 1)
			m.settingsCursor = 0
		case "esc", "left":
			m.focusContent = false
		}
		return m, nil
	}

	key := msg.String()
	switch m.settingsSection {
	case sSectionLang:
		switch key {
		case "up", "shift+tab":
			if m.settingsCursor > 0 {
				m.settingsCursor--
			}
		case "down", "tab":
			if m.settingsCursor < len(langOptions)-1 {
				m.settingsCursor++
			}
		case "enter":
			selected := langOptions[m.settingsCursor].code
			if selected != m.lang {
				m.lang = selected
				m.settings.Language = selected
				_ = storage.SaveSettings(m.personalDataDir, m.settings)
				m.updateMenuLabels()
			}
		case "esc", "left":
			m.settingsSection = sSectionMenu
		}
	case sSectionWeather:
		switch key {
		case "esc", "left":
			_ = storage.SaveSettings(m.personalDataDir, m.settings)
			m.settingsSection = sSectionMenu
			return m, fetchWeatherCmd(m.settings.WeatherLoc)
		case "enter":
			_ = storage.SaveSettings(m.personalDataDir, m.settings)
			m.settingsSection = sSectionMenu
			m.weatherData = "Loading weather..."
			return m, fetchWeatherCmd(m.settings.WeatherLoc)
		case "backspace":
			if len(m.settings.WeatherLoc) > 0 {
				runes := []rune(m.settings.WeatherLoc)
				m.settings.WeatherLoc = string(runes[:len(runes)-1])
			}
		default:
			if key == "space" {
				key = " "
			}
			runes := []rune(key)
			if len(runes) == 1 {
				m.settings.WeatherLoc += key
			}
		}
	case sSectionTheme:
		themes := storage.GetAvailableThemes(m.dataDir)
		switch key {
		case "up", "shift+tab":
			if m.settingsCursor > 0 {
				m.settingsCursor--
			}
		case "down", "tab":
			if m.settingsCursor < len(themes)-1 {
				m.settingsCursor++
			}
		case "enter":
			selected := themes[m.settingsCursor]
			if selected != m.settings.Theme {
				m.settings.Theme = selected
				m.theme = storage.LoadTheme(m.personalDataDir, selected)
				_ = storage.SaveSettings(m.personalDataDir, m.settings)
			}
		case "esc", "left":
			m.settingsSection = sSectionMenu
		}
	case sSectionWorkspace:
		if m.familyIsAdding {
			switch key {
			case "esc":
				m.familyIsAdding = false
				m.familyForm = ""
				m.familyError = ""
			case "enter":
				if strings.TrimSpace(m.familyForm) != "" {
					_, err := storage.CreateFamily(m.baseDataDir, m.familyForm, m.user)
					if err != nil {
						m.familyError = "Error creating family: " + err.Error()
					} else {
						m.familyIsAdding = false
						m.familyForm = ""
						m.familyError = ""
						m.userFamilies, _ = storage.GetUserFamilies(m.baseDataDir, m.user)
					}
				}
			case "backspace":
				if len(m.familyForm) > 0 {
					runes := []rune(m.familyForm)
					m.familyForm = string(runes[:len(runes)-1])
				}
			default:
				if key == "space" {
					key = " "
				}
				runes := []rune(key)
				if len(runes) == 1 {
					m.familyForm += key
				}
			}
		} else if m.familyIsInviting {
			switch key {
			case "esc":
				m.familyIsInviting = false
				m.familyInviteForm = ""
				m.familyError = ""
			case "enter":
				if strings.TrimSpace(m.familyInviteForm) != "" && m.settingsCursor > 0 && m.settingsCursor <= len(m.userFamilies) {
					familyID := m.userFamilies[m.settingsCursor-1].ID
					err := storage.AddMemberToFamily(m.baseDataDir, familyID, m.familyInviteForm)
					if err != nil {
						m.familyError = "Error adding user: " + err.Error()
					} else {
						m.familyIsInviting = false
						m.familyInviteForm = ""
						m.familyError = ""
						m.userFamilies, _ = storage.GetUserFamilies(m.baseDataDir, m.user)
					}
				}
			case "backspace":
				if len(m.familyInviteForm) > 0 {
					runes := []rune(m.familyInviteForm)
					m.familyInviteForm = string(runes[:len(runes)-1])
				}
			default:
				if key == "space" {
					key = " "
				}
				runes := []rune(key)
				if len(runes) == 1 {
					m.familyInviteForm += key
				}
			}
		} else {
			switch key {
			case "esc", "left":
				m.settingsSection = sSectionMenu
			case "up", "k":
				if m.settingsCursor > 0 {
					m.settingsCursor--
				}
			case "down", "j":
				if m.settingsCursor < len(m.userFamilies) {
					m.settingsCursor++
				}
			case "enter":
				if m.settingsCursor == 0 {
					if m.currentWorkspace != "Personal" {
						m.currentWorkspace = "Personal"
						m.dataDir = m.personalDataDir
						m.loadUserData()
					}
				} else if m.settingsCursor > 0 && m.settingsCursor <= len(m.userFamilies) {
					familyID := m.userFamilies[m.settingsCursor-1].ID
					if m.currentWorkspace != familyID {
						m.currentWorkspace = familyID
						m.dataDir = storage.GetFamilyDir(m.baseDataDir, familyID)
						m.loadUserData()
					}
				}
			case "n", "a":
				m.familyIsAdding = true
				m.familyForm = ""
				m.familyError = ""
			case "i":
				if m.settingsCursor > 0 && m.settingsCursor <= len(m.userFamilies) {
					m.familyIsInviting = true
					m.familyInviteForm = ""
					m.familyError = ""
				}
			case "s": // Set default
				if m.settingsCursor == 0 {
					m.settings.DefaultWorkspace = "Personal"
				} else if m.settingsCursor > 0 && m.settingsCursor <= len(m.userFamilies) {
					m.settings.DefaultWorkspace = m.userFamilies[m.settingsCursor-1].ID
				}
				_ = storage.SaveSettings(m.personalDataDir, m.settings)
			case "delete":
				if m.settingsCursor > 0 && m.settingsCursor <= len(m.userFamilies) {
					familyID := m.userFamilies[m.settingsCursor-1].ID
					_ = storage.RemoveMemberFromFamily(m.baseDataDir, familyID, m.user)
					m.userFamilies, _ = storage.GetUserFamilies(m.baseDataDir, m.user)
					// If we just left the current workspace, switch back to personal
					if m.currentWorkspace == familyID {
						m.currentWorkspace = "Personal"
						m.dataDir = m.personalDataDir
						m.loadUserData()
					}
					if m.settingsCursor > len(m.userFamilies) {
						m.settingsCursor = len(m.userFamilies)
					}
				}
			}
		}
	}
	return m, nil
}

// ─── Settings render ─────────────────────────────────────────────────

func (m *model) renderSettingsView(s *styles) string {
	sw := sidebarWidth(m.width)
	submenuWidth := sw - 4
	if submenuWidth < 10 {
		submenuWidth = 10
	}

	title := s.title.Render(t(m.lang, "settings.title"))
	desc := s.subtitle.Render(t(m.lang, "settings.subtitle"))

	labels := []string{strings.ToUpper(t(m.lang, "settings.language")), strings.ToUpper(t(m.lang, "settings.weatherLoc")), strings.ToUpper(t(m.lang, "settings.theme")), strings.ToUpper(t(m.lang, "settings.workspaces"))}
	var lines []string
	for i, l := range labels {
		if m.settingsSection == sSectionMenu && m.settingsMenuCursor == i {
			if m.focusContent {
				lines = append(lines, s.menuSelected.Width(submenuWidth).Render(l))
			} else {
				lines = append(lines, s.menuActiveDim.Width(submenuWidth).Render(l))
			}
		} else if m.settingsSection == setSection(i+1) {
			lines = append(lines, s.menuActiveDim.Width(submenuWidth).Render(l))
		} else {
			lines = append(lines, s.menuNormal.Width(submenuWidth).Render(l))
		}
	}
	menu := strings.Join(lines, "\n")
	col2 := title + "\n" + desc + "\n\n" + menu

	targetSection := m.settingsSection
	if targetSection == sSectionMenu {
		targetSection = setSection(m.settingsMenuCursor + 1)
	}

	var col3 string
	switch targetSection {
	case sSectionLang:
		col3 = m.renderSettingsLang(s)
	case sSectionWeather:
		col3 = m.renderSettingsWeather(s)
	case sSectionTheme:
		col3 = m.renderSettingsTheme(s)
	case sSectionWorkspace:
		col3 = m.renderSettingsWorkspace(s)
	}

	col2Height := lipgloss.Height(col2)
	col3Height := lipgloss.Height(col3)
	minHeight := m.height - 8
	
	maxHeight := col2Height
	if col3Height > maxHeight {
		maxHeight = col3Height
	}
	if minHeight > maxHeight {
		maxHeight = minHeight
	}

	col2Style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color(m.theme.Border)).
		PaddingRight(2).
		MarginRight(2).
		Height(maxHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, col2Style.Render(col2), col3)
}

func (m *model) renderSettingsLang(s *styles) string {
	isActive := m.settingsSection != sSectionMenu
	langTitle := s.info.Render("  " + strings.ToUpper(t(m.lang, "settings.language")))
	current := s.dim.Render(fmt.Sprintf("  %s ", t(m.lang, "settings.currentLang"))) +
		s.highlight.Render(langDisplayName(m.lang))

	var options []string
	for i, opt := range langOptions {
		label := fmt.Sprintf("  %s  %s", opt.flag, langDisplayName(opt.code))
		if i == m.settingsCursor {
			var row string
			if isActive {
				row = s.menuSelected.Width(0).Render("  ▸ " + label)
			} else {
				row = s.menuActiveDim.Width(0).Render("  ▸ " + label)
			}
			if opt.code == m.lang {
				check := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(" ✓")
				row += check
			}
			options = append(options, row)
		} else {
			row := s.menuNormal.Width(0).Render("    " + label)
			if opt.code == m.lang {
				check := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(" ✓")
				row += check
			}
			options = append(options, row)
		}
	}

	optList := strings.Join(options, "\n")
	help := s.dim.Render(fmt.Sprintf("\n\n↑/↓: %s  Enter: %s  ←: %s",
		t(m.lang, "help.navigate"), t(m.lang, "action.save"), t(m.lang, "help.goBack")))

	return langTitle + "\n" + current + "\n\n" + optList + help
}

func (m *model) renderSettingsWeather(s *styles) string {
	isActive := m.settingsSection != sSectionMenu
	weatherTitle := s.info.Render("  " + strings.ToUpper(t(m.lang, "settings.weatherLoc")))
	
	locVal := m.settings.WeatherLoc
	if locVal == "" {
		locVal = s.dim.Render("...")
	}

	cursor := ""
	if isActive {
		cursor = s.highlight.Render("_")
	}
	fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
	var weatherInput string
	if isActive {
		weatherInput = s.menuSelected.Width(0).Render(fmt.Sprintf("  ▸ %s: %s", t(m.lang, "settings.location"), fieldStyle.Render(locVal)+cursor))
	} else {
		weatherInput = s.menuActiveDim.Width(0).Render(fmt.Sprintf("  ▸ %s: %s", t(m.lang, "settings.location"), fieldStyle.Render(locVal)+cursor))
	}

	help := s.dim.Render(fmt.Sprintf("\n\nEnter: %s  Esc/←: %s",
		t(m.lang, "action.save"), t(m.lang, "help.goBack")))

	return weatherTitle + "\n\n" + weatherInput + help
}

func (m *model) renderSettingsTheme(s *styles) string {
	isActive := m.settingsSection != sSectionMenu
	themeTitle := s.info.Render("  " + strings.ToUpper(t(m.lang, "settings.theme")))
	currentName := storage.GetThemeName(m.dataDir, m.settings.Theme)
	current := s.dim.Render(fmt.Sprintf("  %s ", t(m.lang, "settings.currentTheme"))) +
		s.highlight.Render(currentName)

	themes := storage.GetAvailableThemes(m.dataDir)
	var options []string
	for i, opt := range themes {
		displayName := storage.GetThemeName(m.dataDir, opt)
		label := fmt.Sprintf("  %s", displayName)
		if i == m.settingsCursor {
			var row string
			if isActive {
				row = s.menuSelected.Width(0).Render("  ▸ " + label)
			} else {
				row = s.menuActiveDim.Width(0).Render("  ▸ " + label)
			}
			if opt == m.settings.Theme {
				check := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Status)).Render(" ✓")
				row += check
			}
			options = append(options, row)
		} else {
			row := s.menuNormal.Width(0).Render("    " + label)
			if opt == m.settings.Theme {
				check := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Status)).Render(" ✓")
				row += check
			}
			options = append(options, row)
		}
	}

	optList := strings.Join(options, "\n")
	help := s.dim.Render(fmt.Sprintf("\n\n↑/↓: %s  Enter: %s  ←: %s",
		t(m.lang, "help.navigate"), t(m.lang, "action.save"), t(m.lang, "help.goBack")))

	return themeTitle + "\n" + current + "\n\n" + optList + help
}

// langDisplayName returns the full display name for a language code.
func langDisplayName(code string) string {
	switch code {
	case "en":
		return "English"
	case "it":
		return "Italiano"
	default:
		return code
	}
}

func (m *model) renderSettingsWorkspace(s *styles) string {
	title := s.info.Render(t(m.lang, "settings.workspacesSubtitle"))
	
	var currentName string
	if m.currentWorkspace == "Personal" {
		currentName = "Personal"
	} else {
		for _, f := range m.userFamilies {
			if f.ID == m.currentWorkspace {
				currentName = "Family: " + f.Name
				break
			}
		}
	}
	currentInfo := s.dim.Render(t(m.lang, "settings.currentWorkspace")) + s.highlight.Render(currentName)
	
	var contentLines []string
	if m.familyIsAdding {
		contentLines = append(contentLines, s.dim.Render(t(m.lang, "settings.createNewFamily")))
		cursor := s.highlight.Render("_")
		fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
		contentLines = append(contentLines, s.menuSelected.Width(0).Render(fmt.Sprintf("  ▸ Name: %s", fieldStyle.Render(m.familyForm)+cursor)))
		if m.familyError != "" {
			contentLines = append(contentLines, s.status.Foreground(lipgloss.Color("196")).Render("  "+m.familyError))
		}
	} else if m.familyIsInviting {
		contentLines = append(contentLines, s.dim.Render(t(m.lang, "settings.inviteUser")))
		cursor := s.highlight.Render("_")
		fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
		contentLines = append(contentLines, s.menuSelected.Width(0).Render(fmt.Sprintf("  ▸ Username: %s", fieldStyle.Render(m.familyInviteForm)+cursor)))
		if m.familyError != "" {
			contentLines = append(contentLines, s.status.Foreground(lipgloss.Color("196")).Render("  "+m.familyError))
		}
	} else {
		contentLines = append(contentLines, s.dim.Render(t(m.lang, "settings.yourWorkspaces")))
		
		isActive := m.settingsSection == sSectionWorkspace
		
		getLabel := func(base string, isDefault bool) string {
			if isDefault {
				return base + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(" (Default)")
			}
			return base
		}

		if m.settingsCursor == 0 {
			label := getLabel(t(m.lang, "settings.personal"), m.settings.DefaultWorkspace == "Personal")
			if isActive {
				contentLines = append(contentLines, s.menuSelected.Width(0).Render("  ▸ "+label))
			} else {
				contentLines = append(contentLines, s.menuActiveDim.Width(0).Render("  ▸ "+label))
			}
		} else {
			label := getLabel(t(m.lang, "settings.personal"), m.settings.DefaultWorkspace == "Personal")
			contentLines = append(contentLines, s.menuNormal.Width(0).Render("    "+label))
		}
		
		for i, f := range m.userFamilies {
			label := getLabel(fmt.Sprintf(t(m.lang, "settings.familyLabel"), f.Name, len(f.Members)), m.settings.DefaultWorkspace == f.ID)
			if m.settingsCursor == i+1 {
				if isActive {
					contentLines = append(contentLines, s.menuSelected.Width(0).Render("  ▸ "+label))
				} else {
					contentLines = append(contentLines, s.menuActiveDim.Width(0).Render("  ▸ "+label))
				}
			} else {
				contentLines = append(contentLines, s.menuNormal.Width(0).Render("    "+label))
			}
		}
		if m.familyError != "" {
			contentLines = append(contentLines, "\n"+s.status.Foreground(lipgloss.Color("196")).Render("  "+m.familyError))
		}
	}

	content := strings.Join(contentLines, "\n")
	
	var help string
	if m.familyIsAdding {
		help = s.dim.Render(fmt.Sprintf("\n\nEnter: %s  Esc: %s", t(m.lang, "action.save"), t(m.lang, "help.cancel")))
	} else if m.familyIsInviting {
		help = s.dim.Render(fmt.Sprintf("\n\nEnter: %s  Esc: %s", t(m.lang, "action.save"), t(m.lang, "help.cancel")))
	} else {
		help = s.dim.Render(fmt.Sprintf("\n\n↑/↓: %s  %s\n%s  ←: %s", t(m.lang, "help.navigate"), t(m.lang, "settings.workspaceHelp1"), t(m.lang, "settings.workspaceHelp2"), t(m.lang, "help.goBack")))
	}

	return title + "\n" + currentInfo + "\n\n" + content + help
}
