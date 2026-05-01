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

// ─── Settings update ─────────────────────────────────────────────────

func (m *model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case "down", "j":
		if m.settingsCursor < len(langOptions)-1 {
			m.settingsCursor++
		}
	case "enter":
		selected := langOptions[m.settingsCursor].code
		if selected != m.lang {
			m.lang = selected
			m.settings.Language = selected
			_ = storage.SaveSettings(m.dataDir, m.settings)
			// Update dynamic menu items
			m.updateMenuLabels()
		}
	case "esc", "left":
		m.focusContent = false
	}
	return m, nil
}

// ─── Settings render ─────────────────────────────────────────────────

func (m *model) renderSettingsView(s *styles) string {
	title := s.title.Render(t(m.lang, "settings.title"))
	desc := s.subtitle.Render(t(m.lang, "settings.subtitle"))

	// Language section
	langTitle := s.info.Render(fmt.Sprintf("  %s", t(m.lang, "settings.language")))
	current := s.dim.Render(fmt.Sprintf("  %s ", t(m.lang, "settings.currentLang"))) +
		s.highlight.Render(langDisplayName(m.lang))

	var options []string
	for i, opt := range langOptions {
		label := fmt.Sprintf("  %s  %s", opt.flag, langDisplayName(opt.code))
		if i == m.settingsCursor {
			row := s.menuSelected.Width(0).Render("  ▸ " + label)
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

	optList := ""
	for _, o := range options {
		optList += o + "\n"
	}

	help := s.dim.Render(fmt.Sprintf("↑/↓: %s  Enter: %s  ←: %s",
		t(m.lang, "help.navigate"),
		t(m.lang, "help.select"),
		t(m.lang, "help.goBack"),
	))

	return title + "\n" + desc + "\n\n" +
		langTitle + "\n" + current + "\n\n" +
		optList + "\n" + help
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
