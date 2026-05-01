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

// ─── Vehicle sub-view states ─────────────────────────────────────────

type vehicleSubView int

const (
	vViewList   vehicleSubView = iota
	vViewAdd
	vViewEdit
	vViewDelete
)

// ─── Form field indices ──────────────────────────────────────────────

const (
	fBrand = iota
	fModel
	fPlate
	fOwner
	fCount
)

var fieldLabels = [fCount]string{
	"Brand",
	"Model",
	"License Plate",
	"Owner",
}

// ─── Update handlers ─────────────────────────────────────────────────

// updateVehicleList handles input when the vehicle list has focus.
func (m *model) updateVehicleList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.vehicleCursor > 0 {
			m.vehicleCursor--
		}
	case "down", "j":
		if m.vehicleCursor < len(m.vehicles)-1 {
			m.vehicleCursor++
		}
	case "a":
		// Add new vehicle
		m.vehicleView = vViewAdd
		m.formFields = [fCount]string{}
		m.formCursor = 0
	case "e", "enter":
		// Edit selected vehicle
		if len(m.vehicles) > 0 {
			v := m.vehicles[m.vehicleCursor]
			m.vehicleView = vViewEdit
			m.editIndex = m.vehicleCursor
			m.formFields = [fCount]string{v.Brand, v.Model, v.LicensePlate, v.Owner}
			m.formCursor = 0
		}
	case "d", "x":
		// Delete selected vehicle
		if len(m.vehicles) > 0 {
			m.vehicleView = vViewDelete
		}
	case "esc":
		m.focusContent = false
	}
	return m, nil
}

// updateVehicleForm handles input in the add/edit form.
func (m *model) updateVehicleForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		m.formCursor = (m.formCursor + 1) % fCount
	case "shift+tab", "up":
		m.formCursor = (m.formCursor - 1 + fCount) % fCount
	case "enter":
		// Save the vehicle
		v := storage.Vehicle{
			Brand:        strings.TrimSpace(m.formFields[fBrand]),
			Model:        strings.TrimSpace(m.formFields[fModel]),
			LicensePlate: strings.TrimSpace(m.formFields[fPlate]),
			Owner:        strings.TrimSpace(m.formFields[fOwner]),
		}
		if v.Brand == "" && v.Model == "" && v.LicensePlate == "" && v.Owner == "" {
			// Don't save empty vehicles
			m.vehicleView = vViewList
			return m, nil
		}
		if m.vehicleView == vViewAdd {
			m.vehicles = append(m.vehicles, v)
		} else if m.vehicleView == vViewEdit {
			m.vehicles[m.editIndex] = v
		}
		_ = storage.SaveVehicles(m.dataDir, m.vehicles)
		m.vehicleView = vViewList
	case "esc":
		m.vehicleView = vViewList
	case "backspace":
		field := &m.formFields[m.formCursor]
		if len(*field) > 0 {
			runes := []rune(*field)
			*field = string(runes[:len(runes)-1])
		}
	default:
		// Append printable characters
		runes := []rune(key)
		if len(runes) == 1 && unicode.IsPrint(runes[0]) {
			m.formFields[m.formCursor] += key
		}
	}
	return m, nil
}

// updateVehicleDelete handles the delete confirmation.
func (m *model) updateVehicleDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "y":
		// Confirm delete
		m.vehicles = append(m.vehicles[:m.vehicleCursor], m.vehicles[m.vehicleCursor+1:]...)
		_ = storage.SaveVehicles(m.dataDir, m.vehicles)
		if m.vehicleCursor >= len(m.vehicles) && m.vehicleCursor > 0 {
			m.vehicleCursor--
		}
		m.vehicleView = vViewList
	case "n", "esc":
		m.vehicleView = vViewList
	}
	return m, nil
}

// ─── Render ──────────────────────────────────────────────────────────

func (m *model) renderVehiclesView(s *styles) string {
	switch m.vehicleView {
	case vViewAdd:
		return m.renderVehicleForm(s, "Add Vehicle")
	case vViewEdit:
		return m.renderVehicleForm(s, "Edit Vehicle")
	case vViewDelete:
		return m.renderVehicleDeleteConfirm(s)
	default:
		return m.renderVehicleList(s)
	}
}

func (m *model) renderVehicleList(s *styles) string {
	title := s.title.Render("Vehicles")

	if len(m.vehicles) == 0 {
		empty := s.dim.Render("No vehicles registered yet.")
		var help string
		if m.focusContent {
			help = s.dim.Render("\n\na: add vehicle  Esc: back to menu")
		} else {
			help = s.dim.Render("\n\nEnter/Tab: focus list")
		}
		return title + "\n\n" + empty + help
	}

	// Table header
	hdr := fmt.Sprintf("  %-3s %-14s %-14s %-14s %-14s", "#", "BRAND", "MODEL", "PLATE", "OWNER")
	header := s.subtitle.Render(hdr)
	divider := s.dim.Render("  " + strings.Repeat("─", 60))

	// Table rows
	var rows []string
	for i, v := range m.vehicles {
		row := fmt.Sprintf("  %-3d %-14s %-14s %-14s %-14s",
			i+1,
			truncate(v.Brand, 13),
			truncate(v.Model, 13),
			truncate(v.LicensePlate, 13),
			truncate(v.Owner, 13),
		)
		if m.focusContent && i == m.vehicleCursor {
			row = s.menuSelected.Width(0).Render(row)
		} else {
			row = s.info.Render(row)
		}
		rows = append(rows, row)
	}
	table := strings.Join(rows, "\n")

	// Help line
	var help string
	if m.focusContent {
		help = s.dim.Render("a: add  e: edit  d: delete  Esc: back")
	} else {
		help = s.dim.Render("Enter/Tab: focus list")
	}

	return title + "\n" + header + "\n" + divider + "\n" + table + "\n\n" + help
}

func (m *model) renderVehicleForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	var fields []string
	for i := 0; i < fCount; i++ {
		label := s.dim.Render(fmt.Sprintf("  %-15s", fieldLabels[i]+":"))
		value := m.formFields[i]

		var rendered string
		if i == m.formCursor {
			// Focused field: show cursor
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

	help := s.dim.Render("Tab/↑↓: switch field  Enter: save  Esc: cancel")

	return title + "\n\n" + form + "\n\n" + help
}

func (m *model) renderVehicleDeleteConfirm(s *styles) string {
	if m.vehicleCursor >= len(m.vehicles) {
		m.vehicleView = vViewList
		return m.renderVehicleList(s)
	}
	v := m.vehicles[m.vehicleCursor]

	title := s.title.Render("Delete Vehicle")
	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Render("Are you sure you want to delete this vehicle?")

	info := fmt.Sprintf(
		"\n  %s %s\n  %s %s\n  %s %s\n  %s %s",
		s.dim.Render("Brand:"), s.info.Render(v.Brand),
		s.dim.Render("Model:"), s.info.Render(v.Model),
		s.dim.Render("Plate:"), s.info.Render(v.LicensePlate),
		s.dim.Render("Owner:"), s.info.Render(v.Owner),
	)

	help := s.dim.Render("y: confirm  n/Esc: cancel")

	return title + "\n\n" + warning + info + "\n\n" + help
}

// ─── Helpers ─────────────────────────────────────────────────────────

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
