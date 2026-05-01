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
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

// ─── Insurance form fields ───────────────────────────────────────────

const (
	insFPlate  = iota // License plate (selected via picker)
	insFCost          // Total cost
	insFExpiry        // Expire date
	insFCount
)

// insFieldKeys maps to i18n keys for insurance form labels.
var insFieldKeys = [insFCount]string{
	"field.licensePlate",
	"field.totalCost",
	"field.expireDate",
}

// ─── Insurance update handlers ───────────────────────────────────────

func (m *model) updateInsuranceSection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.vehicleView {
	case vViewAdd, vViewEdit:
		if m.insPickerMode {
			return m.updateInsurancePicker(msg)
		}
		return m.updateInsuranceForm(msg)
	case vViewDelete:
		return m.updateInsuranceDelete(msg)
	default:
		return m.updateInsuranceList(msg)
	}
}

func (m *model) updateInsuranceList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.insuranceCursor > 0 {
			m.insuranceCursor--
		}
	case "down", "j":
		if m.insuranceCursor < len(m.insurances)-1 {
			m.insuranceCursor++
		}
	case "a":
		if len(m.vehicles) == 0 {
			return m, nil // can't add without vehicles
		}
		m.vehicleView = vViewAdd
		m.insFormFields = [insFCount]string{}
		m.insFormCursor = 0
		m.insPickerMode = true
		m.insPickerCursor = 0
	case "e", "enter":
		if len(m.insurances) > 0 {
			ins := m.insurances[m.insuranceCursor]
			m.vehicleView = vViewEdit
			m.editIndex = m.insuranceCursor
			m.insFormFields = [insFCount]string{ins.LicensePlate, ins.TotalCost, ins.ExpireDate}
			m.insFormCursor = 1 // start on TotalCost, plate is pre-selected
			m.insPickerMode = false
		}
	case "d", "x":
		if len(m.insurances) > 0 {
			m.vehicleView = vViewDelete
		}
	case "esc":
		m.vehicleSection = vSectionMenu
	}
	return m, nil
}

func (m *model) updateInsurancePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.insPickerCursor > 0 {
			m.insPickerCursor--
		}
	case "down", "j":
		if m.insPickerCursor < len(m.vehicles)-1 {
			m.insPickerCursor++
		}
	case "enter":
		// Select this vehicle's plate
		m.insFormFields[insFPlate] = m.vehicles[m.insPickerCursor].LicensePlate
		m.insPickerMode = false
		m.insFormCursor = 1 // move to TotalCost
	case "esc":
		m.vehicleView = vViewList
		m.insPickerMode = false
	}
	return m, nil
}

func (m *model) updateInsuranceForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		next := m.insFormCursor + 1
		if next >= insFCount {
			next = 1 // skip plate field (index 0), wrap to cost
		}
		if next == 0 {
			next = 1
		}
		m.insFormCursor = next
	case "shift+tab", "up":
		prev := m.insFormCursor - 1
		if prev < 1 {
			prev = insFCount - 1 // skip plate field, wrap to last
		}
		m.insFormCursor = prev
	case "enter":
		ins := storage.Insurance{
			LicensePlate: strings.TrimSpace(m.insFormFields[insFPlate]),
			TotalCost:    strings.TrimSpace(m.insFormFields[insFCost]),
			ExpireDate:   strings.TrimSpace(m.insFormFields[insFExpiry]),
		}
		if ins.LicensePlate == "" {
			m.vehicleView = vViewList
			return m, nil
		}
		if m.vehicleView == vViewAdd {
			m.insurances = append(m.insurances, ins)
		} else if m.vehicleView == vViewEdit {
			m.insurances[m.editIndex] = ins
		}
		_ = storage.SaveInsurance(m.dataDir, m.insurances)
		m.vehicleView = vViewList
	case "esc":
		m.vehicleView = vViewList
	case "backspace":
		if m.insFormCursor > 0 { // don't edit plate
			field := &m.insFormFields[m.insFormCursor]
			if len(*field) > 0 {
				runes := []rune(*field)
				*field = string(runes[:len(runes)-1])
			}
		}
	default:
		if m.insFormCursor > 0 { // don't type in plate field
			runes := []rune(key)
			if len(runes) == 1 && unicode.IsPrint(runes[0]) {
				m.insFormFields[m.insFormCursor] += key
			}
		}
	}
	return m, nil
}

func (m *model) updateInsuranceDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "y":
		m.insurances = append(m.insurances[:m.insuranceCursor], m.insurances[m.insuranceCursor+1:]...)
		_ = storage.SaveInsurance(m.dataDir, m.insurances)
		if m.insuranceCursor >= len(m.insurances) && m.insuranceCursor > 0 {
			m.insuranceCursor--
		}
		m.vehicleView = vViewList
	case "n", "esc":
		m.vehicleView = vViewList
	}
	return m, nil
}

// ─── Insurance render ────────────────────────────────────────────────

func (m *model) renderInsuranceView(s *styles) string {
	switch m.vehicleView {
	case vViewAdd:
		if m.insPickerMode {
			return m.renderInsurancePicker(s)
		}
		return m.renderInsuranceForm(s, t(m.lang, "insurance.add"))
	case vViewEdit:
		return m.renderInsuranceForm(s, t(m.lang, "insurance.edit"))
	case vViewDelete:
		return m.renderInsuranceDeleteConfirm(s)
	default:
		return m.renderInsuranceList(s)
	}
}

func (m *model) renderInsuranceList(s *styles) string {
	title := s.title.Render(t(m.lang, "insurance.title"))

	if len(m.insurances) == 0 {
		empty := s.dim.Render(t(m.lang, "insurance.noRecords"))
		var extra string
		if len(m.vehicles) == 0 {
			extra = "\n" + s.dim.Render(t(m.lang, "vehicles.addFirst"))
		}
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  Esc: %s", t(m.lang, "action.add"), t(m.lang, "action.back")))
		return title + "\n\n" + empty + extra + help
	}

	hdr := fmt.Sprintf("  %-3s %-14s %-14s %-14s",
		t(m.lang, "col.num"), t(m.lang, "col.plate"), t(m.lang, "col.cost"), t(m.lang, "col.expires"))
	header := s.subtitle.Render(hdr)
	divider := s.dim.Render("  " + strings.Repeat("─", 48))

	var rows []string
	for i, ins := range m.insurances {
		row := fmt.Sprintf("  %-3d %-14s %-14s %-14s",
			i+1,
			truncate(ins.LicensePlate, 13),
			truncate(ins.TotalCost, 13),
			truncate(ins.ExpireDate, 13),
		)
		if i == m.insuranceCursor {
			row = s.menuSelected.Width(0).Render(row)
		} else {
			row = s.info.Render(row)
		}
		rows = append(rows, row)
	}
	table := strings.Join(rows, "\n")

	help := s.dim.Render(fmt.Sprintf("a: %s  e: %s  d: %s  Esc: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "action.back")))

	return title + "\n" + header + "\n" + divider + "\n" + table + "\n\n" + help
}

func (m *model) renderInsurancePicker(s *styles) string {
	title := s.title.Render(t(m.lang, "action.selectVeh"))
	desc := s.subtitle.Render(t(m.lang, "action.chooseByPlate"))

	var lines []string
	for i, v := range m.vehicles {
		label := fmt.Sprintf("  %-14s  %s %s", v.LicensePlate, v.Brand, v.Model)
		if i == m.insPickerCursor {
			lines = append(lines, s.menuSelected.Width(0).Render("▸ "+label))
		} else {
			lines = append(lines, s.menuNormal.Width(0).Render("  "+label))
		}
	}
	list := strings.Join(lines, "\n")

	help := s.dim.Render(fmt.Sprintf("↑/↓: %s  Enter: %s  Esc: %s",
		t(m.lang, "help.navigate"), t(m.lang, "help.select"), t(m.lang, "action.cancel")))

	return title + "\n" + desc + "\n\n" + list + "\n\n" + help
}

func (m *model) renderInsuranceForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	var fields []string
	for i := 0; i < insFCount; i++ {
		label := s.dim.Render(fmt.Sprintf("  %-15s", t(m.lang, insFieldKeys[i])+":"))
		value := m.insFormFields[i]

		var rendered string
		if i == 0 {
			// Plate is read-only (selected via picker)
			plateStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true)
			rendered = label + " " + plateStyle.Render(value) + s.dim.Render(" (" + t(m.lang, "action.locked") + ")")
		} else if i == m.insFormCursor {
			cursor := s.highlight.Render("_")
			fieldStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236"))
			rendered = label + " " + fieldStyle.Render(value) + cursor
		} else {
			rendered = label + " " + s.info.Render(value)
		}
		fields = append(fields, rendered)
	}
	form := strings.Join(fields, "\n\n")

	help := s.dim.Render(fmt.Sprintf("Tab/↑↓: %s  Enter: %s  Esc: %s",
		t(m.lang, "help.switchField"), t(m.lang, "action.save"), t(m.lang, "action.cancel")))

	return title + "\n\n" + form + "\n\n" + help
}

func (m *model) renderInsuranceDeleteConfirm(s *styles) string {
	if m.insuranceCursor >= len(m.insurances) {
		m.vehicleView = vViewList
		return m.renderInsuranceList(s)
	}
	ins := m.insurances[m.insuranceCursor]

	title := s.title.Render(t(m.lang, "delete.insurance"))
	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Render(t(m.lang, "delete.confirmInsurance"))

	info := fmt.Sprintf(
		"\n  %s %s\n  %s %s\n  %s %s",
		s.dim.Render(t(m.lang, "field.licensePlate")+":"), s.info.Render(ins.LicensePlate),
		s.dim.Render(t(m.lang, "field.totalCost")+":"), s.info.Render(ins.TotalCost),
		s.dim.Render(t(m.lang, "field.expireDate")+":"), s.info.Render(ins.ExpireDate),
	)

	help := s.dim.Render(fmt.Sprintf("y: %s  n/Esc: %s", t(m.lang, "action.confirm"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + info + "\n\n" + help
}
