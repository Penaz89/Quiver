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
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

// ─── Vehicle section sub-menu ────────────────────────────────────────

type vehicleSection int

const (
	vSectionMenu    vehicleSection = iota // sub-menu view
	vSectionMgmt                          // Vehicle Management
	vSectionInsurance                     // Insurance
	vSectionRoadTax                       // Road Tax
	vSectionNTC                           // NTC
)

// vehicleSectionLabels returns the localized section names.
func vehicleSectionLabels(lang string) []string {
	return []string{
		t(lang, "vehicles.management"),
		t(lang, "vehicles.insurance"),
		t(lang, "vehicles.roadTax"),
		t(lang, "vehicles.ntc"),
	}
}

// ─── Vehicle sub-view states (within each section) ───────────────────

type vehicleSubView int

const (
	vViewList   vehicleSubView = iota
	vViewAdd
	vViewEdit
	vViewDelete
)

// ─── Form field indices ──────────────────────────────────────────────

// Vehicle Management fields
const (
	fBrand = iota
	fModel
	fPlate
	fOwner
	fMgmtCount
)

// Road Tax fields
const (
	fRoadTax      = 0
	fRoadTaxCount = 1
)

// NTC fields
const (
	fNTC      = 0
	fNTCCount = 1
)

// mgmtFieldKeys maps to i18n keys for vehicle form labels.
var mgmtFieldKeys = [fMgmtCount]string{
	"field.brand",
	"field.model",
	"field.licensePlate",
	"field.owner",
}

// fCount is the max across all sections (used for formFields array size)
const fCount = fMgmtCount

// ─── Update handlers ─────────────────────────────────────────────────

// updateVehicleSection routes input based on current vehicle section.
func (m *model) updateVehicleSection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.vehicleSection {
	case vSectionMenu:
		return m.updateVehicleSectionMenu(msg)
	case vSectionMgmt:
		switch m.vehicleView {
		case vViewAdd, vViewEdit:
			return m.updateVehicleForm(msg)
		case vViewDelete:
			return m.updateVehicleDelete(msg)
		default:
			return m.updateVehicleList(msg)
		}
	case vSectionInsurance:
		return m.updateInsuranceSection(msg)
	case vSectionRoadTax:
		switch m.vehicleView {
		case vViewAdd, vViewEdit:
			return m.updateSingleFieldForm(msg, fRoadTaxCount)
		default:
			return m.updateSingleFieldList(msg)
		}
	case vSectionNTC:
		switch m.vehicleView {
		case vViewAdd, vViewEdit:
			return m.updateSingleFieldForm(msg, fNTCCount)
		default:
			return m.updateSingleFieldList(msg)
		}
	}
	return m, nil
}

// updateVehicleSectionMenu handles input in the vehicle sub-menu.
func (m *model) updateVehicleSectionMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.vehicleSectionCursor > 0 {
			m.vehicleSectionCursor--
		}
	case "down", "j":
		if m.vehicleSectionCursor < len(vehicleSectionLabels(m.lang))-1 {
			m.vehicleSectionCursor++
		}
	case "enter", "right":
		switch m.vehicleSectionCursor {
		case 0:
			m.vehicleSection = vSectionMgmt
		case 1:
			m.vehicleSection = vSectionInsurance
		case 2:
			m.vehicleSection = vSectionRoadTax
		case 3:
			m.vehicleSection = vSectionNTC
		}
		m.vehicleView = vViewList
		m.vehicleCursor = 0
	case "esc", "left":
		m.focusContent = false
	}
	return m, nil
}

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
		m.vehicleView = vViewAdd
		m.formFields = [fCount]string{}
		m.formCursor = 0
	case "e", "enter":
		if len(m.vehicles) > 0 {
			v := m.vehicles[m.vehicleCursor]
			m.vehicleView = vViewEdit
			m.editIndex = m.vehicleCursor
			m.formFields = [fCount]string{v.Brand, v.Model, v.LicensePlate, v.Owner}
			m.formCursor = 0
		}
	case "d", "x":
		if len(m.vehicles) > 0 {
			m.vehicleView = vViewDelete
		}
	case "esc", "left":
		m.vehicleSection = vSectionMenu
	}
	return m, nil
}

// updateVehicleForm handles input in the add/edit form (Vehicle Management).
func (m *model) updateVehicleForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		m.formCursor = (m.formCursor + 1) % fMgmtCount
	case "shift+tab", "up":
		m.formCursor = (m.formCursor - 1 + fMgmtCount) % fMgmtCount
	case "enter":
		v := storage.Vehicle{
			Brand:        strings.TrimSpace(m.formFields[fBrand]),
			Model:        strings.TrimSpace(m.formFields[fModel]),
			LicensePlate: strings.TrimSpace(m.formFields[fPlate]),
			Owner:        strings.TrimSpace(m.formFields[fOwner]),
		}
		if v.Brand == "" && v.Model == "" && v.LicensePlate == "" && v.Owner == "" {
			m.vehicleView = vViewList
			return m, nil
		}
		if m.vehicleView == vViewAdd {
			m.vehicles = append(m.vehicles, v)
		} else if m.vehicleView == vViewEdit {
			// Preserve existing RoadTax/NTC
			v.RoadTax = m.vehicles[m.editIndex].RoadTax
			v.NTC = m.vehicles[m.editIndex].NTC
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

// updateSingleFieldList handles a list view for Insurance/RoadTax/NTC sections.
func (m *model) updateSingleFieldList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "e", "enter":
		if len(m.vehicles) > 0 {
			m.vehicleView = vViewEdit
			m.editIndex = m.vehicleCursor
			v := m.vehicles[m.vehicleCursor]
			var valCost, valDate string
			switch m.vehicleSection {
			case vSectionRoadTax:
				valCost = v.RoadTaxCost
				if !v.RoadTax.IsZero() {
					valDate = v.RoadTax.Format("02/01/2006")
				}
			case vSectionNTC:
				valCost = v.NTCCost
				if !v.NTC.IsZero() {
					valDate = v.NTC.Format("02/01/2006")
				}
			}
			m.formFields = [fCount]string{valCost, valDate}
			m.formCursor = 0
		}
	case "esc", "left":
		m.vehicleSection = vSectionMenu
	}
	return m, nil
}

// updateSingleFieldForm handles editing a single field (Insurance/RoadTax/NTC).
func (m *model) updateSingleFieldForm(msg tea.KeyMsg, _ int) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "shift+tab":
		if m.formCursor > 0 {
			m.formCursor--
		}
	case "down", "tab":
		if m.formCursor < 1 { // Cost -> Date
			m.formCursor++
		}
	case "enter":
		if len(m.vehicles) > 0 && m.editIndex < len(m.vehicles) {
			cost := strings.TrimSpace(m.formFields[0])
			dateStr := strings.TrimSpace(m.formFields[1])
			var date time.Time
			var err error
			if dateStr != "" {
				date, err = time.Parse("02/01/2006", dateStr)
				if err != nil {
					return m, nil // Stay on form if invalid
				}
			}

			switch m.vehicleSection {
			case vSectionRoadTax:
				m.vehicles[m.editIndex].RoadTax = date
				m.vehicles[m.editIndex].RoadTaxCost = cost
			case vSectionNTC:
				m.vehicles[m.editIndex].NTC = date
				m.vehicles[m.editIndex].NTCCost = cost
			}
			_ = storage.SaveVehicles(m.dataDir, m.vehicles)
		}
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
		runes := []rune(key)
		if len(runes) == 1 && unicode.IsPrint(runes[0]) {
			field := &m.formFields[m.formCursor]
			if m.formCursor == 1 { // Date
				if !unicode.IsDigit(runes[0]) {
					return m, nil
				}
				if len(*field) >= 10 {
					return m, nil
				}
				*field += key
				if len(*field) == 2 || len(*field) == 5 {
					*field += "/"
				}
			} else { // Cost
				if !unicode.IsDigit(runes[0]) && runes[0] != '.' && runes[0] != ',' {
					return m, nil
				}
				if len(*field) >= 15 {
					return m, nil
				}
				*field += key
			}
		}
	}
	return m, nil
}

// ─── Render ──────────────────────────────────────────────────────────

func (m *model) renderVehiclesView(s *styles) string {
	// ── Column 2: Submenu ────────────────────────────────────────
	sw := sidebarWidth(m.width)
	submenuWidth := sw - 4
	if submenuWidth < 10 {
		submenuWidth = 10
	}

	title := s.title.Render(t(m.lang, "vehicles.title"))
	desc := s.subtitle.Render(t(m.lang, "vehicles.selectSection"))

	sections := vehicleSectionLabels(m.lang)
	var lines []string
	for i, section := range sections {
		if i == m.vehicleSectionCursor {
			if m.vehicleSection == vSectionMenu {
				lines = append(lines, s.menuSelected.Width(submenuWidth).Render(section))
			} else {
				lines = append(lines, s.menuActiveDim.Width(submenuWidth).Render(section))
			}
		} else {
			lines = append(lines, s.menuNormal.Width(submenuWidth).Render(section))
		}
	}
	menu := strings.Join(lines, "\n")
	col2 := title + "\n" + desc + "\n\n" + menu

	// ── Column 3: Content / Preview ──────────────────────────────
	var col3 string
	switch m.vehicleSection {
	case vSectionMgmt:
		col3 = m.renderVehicleMgmt(s)
	case vSectionInsurance:
		col3 = m.renderInsuranceView(s)
	case vSectionRoadTax:
		col3 = m.renderSingleFieldSection(s, "Road Tax")
	case vSectionNTC:
		col3 = m.renderSingleFieldSection(s, "NTC")
	default:
		// Preview mode (Stats & Expiries)
		statsLeft, expiriesRight := m.renderVehicleStats(s)
		if expiriesRight != "" {
			divider := s.dim.Render(strings.Repeat("─", 30))
			col3 = statsLeft + "\n\n" + divider + "\n\n" + expiriesRight
		} else {
			col3 = statsLeft
		}
		
		help := s.dim.Render(fmt.Sprintf("↑/↓: %s  →: %s  ←: %s",
			t(m.lang, "help.navigate"), t(m.lang, "help.enter"), t(m.lang, "help.goBack")))
		col3 += "\n\n" + help
	}

	// Calculate heights to stretch the divider
	col2Height := lipgloss.Height(col2)
	col3Height := lipgloss.Height(col3)
	minHeight := m.height - 8 // Viewport inner height (m.height - 4 for content box, - 4 for borders/padding)
	
	maxHeight := col2Height
	if col3Height > maxHeight {
		maxHeight = col3Height
	}
	if minHeight > maxHeight {
		maxHeight = minHeight
	}

	col2Style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("63")).
		PaddingRight(2).
		MarginRight(2).
		Height(maxHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, col2Style.Render(col2), col3)
}

func (m *model) renderVehicleStats(s *styles) (string, string) {
	statsTitle := s.subtitle.Render("  " + t(m.lang, "vehicles.statistics"))

	total := len(m.vehicles)
	if total == 0 {
		return statsTitle + "\n" + s.dim.Render("  " + t(m.lang, "vehicles.noVehicles")), ""
	}

	// Count populated fields
	var taxed, inspected int
	var nextRoadTaxDate, nextNTCDate time.Time
	var nextRoadTax, nextNTC string
	var nextRTVehicle, nextNTCVehicle string

	// Insurance stats from separate insurance records
	insuredPlates := make(map[string]bool)
	var nextInsuranceDate time.Time
	var nextInsurance, nextInsVehicle string
	for _, ins := range m.insurances {
		insuredPlates[ins.LicensePlate] = true
		if !ins.ExpireDate.IsZero() {
			if nextInsuranceDate.IsZero() || ins.ExpireDate.Before(nextInsuranceDate) {
				nextInsuranceDate = ins.ExpireDate
				nextInsurance = ins.ExpireDate.Format("02/01/2006")
				// Find vehicle name for this plate
				for _, v := range m.vehicles {
					if v.LicensePlate == ins.LicensePlate {
						nextInsVehicle = v.Brand + " " + v.Model
						break
					}
				}
			}
		}
	}
	insured := len(insuredPlates)

	for _, v := range m.vehicles {
		vName := v.Brand + " " + v.Model
		if !v.RoadTax.IsZero() {
			taxed++
			if nextRoadTaxDate.IsZero() || v.RoadTax.Before(nextRoadTaxDate) {
				nextRoadTaxDate = v.RoadTax
				nextRoadTax = v.RoadTax.Format("02/01/2006")
				nextRTVehicle = vName
			}
		}
		if !v.NTC.IsZero() {
			inspected++
			if nextNTCDate.IsZero() || v.NTC.Before(nextNTCDate) {
				nextNTCDate = v.NTC
				nextNTC = v.NTC.Format("02/01/2006")
				nextNTCVehicle = vName
			}
		}
	}

	// Summary line
	totalLine := fmt.Sprintf("  %s  %s",
		s.dim.Render(t(m.lang, "vehicles.totalVehicles")),
		s.highlight.Render(fmt.Sprintf("%d", total)),
	)

	// Coverage bars
	insLine := fmt.Sprintf("  %s  %s",
		s.dim.Render(t(m.lang, "vehicles.insurance")+":"),
		renderCoverage(s, insured, total),
	)
	taxLine := fmt.Sprintf("  %s  %s",
		s.dim.Render(t(m.lang, "vehicles.roadTax")+":"),
		renderCoverage(s, taxed, total),
	)
	ntcLine := fmt.Sprintf("  %s  %s",
		s.dim.Render(t(m.lang, "vehicles.ntc")+":"),
		renderCoverage(s, inspected, total),
	)

	leftStats := statsTitle + "\n\n" + totalLine + "\n\n" + insLine + "\n" + taxLine + "\n" + ntcLine

	// Next expiry section (Right side)
	var expiries []string
	if nextInsurance != "" {
		expiries = append(expiries, fmt.Sprintf("%s\n%s %s %s",
			s.dim.Render(t(m.lang, "vehicles.insurance")+":"),
			s.info.Render(nextInsurance),
			s.dim.Render("("+nextInsVehicle+")"),
			formatDaysRemaining(m.lang, s, nextInsuranceDate),
		))
	}
	if nextRoadTax != "" {
		expiries = append(expiries, fmt.Sprintf("%s\n%s %s %s",
			s.dim.Render(t(m.lang, "vehicles.roadTax")+":"),
			s.info.Render(nextRoadTax),
			s.dim.Render("("+nextRTVehicle+")"),
			formatDaysRemaining(m.lang, s, nextRoadTaxDate),
		))
	}
	if nextNTC != "" {
		expiries = append(expiries, fmt.Sprintf("%s\n%s %s %s",
			s.dim.Render(t(m.lang, "vehicles.ntc")+":"),
			s.info.Render(nextNTC),
			s.dim.Render("("+nextNTCVehicle+")"),
			formatDaysRemaining(m.lang, s, nextNTCDate),
		))
	}

	var rightExpiries string
	if len(expiries) > 0 {
		expiryTitle := s.subtitle.Render(t(m.lang, "vehicles.nextExpiry"))
		rightExpiries = expiryTitle + "\n\n" + strings.Join(expiries, "\n\n")
	}

	return leftStats, rightExpiries
}

// renderCoverage shows a fraction like "3/5" colored by coverage level.
func renderCoverage(s *styles, count, total int) string {
	label := fmt.Sprintf("%d/%d", count, total)
	if count == total {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(label) // green
	} else if count == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(label) // red
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(label) // yellow/orange
}

// formatDaysRemaining calculates the days left until the target date and returns a formatted string.
func formatDaysRemaining(lang string, s *styles, target time.Time) string {
	if target.IsZero() {
		return ""
	}
	now := time.Now()
	// calculate absolute days difference without time of day
	targetDate := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	
	days := int(targetDate.Sub(today).Hours() / 24)
	
	if days < 0 {
		return s.highlight.Render(fmt.Sprintf(t(lang, "vehicles.expiredDays"), -days))
	} else if days == 0 {
		return s.highlight.Render(t(lang, "vehicles.expiresToday"))
	}
	return s.dim.Render(fmt.Sprintf(t(lang, "vehicles.expiresIn"), days))
}

func (m *model) renderVehicleMgmt(s *styles) string {
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
	title := s.title.Render(t(m.lang, "vehicles.management"))

	if len(m.vehicles) == 0 {
		empty := s.dim.Render(t(m.lang, "vehicles.noVehicles"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	hdr := fmt.Sprintf("  %-3s %-14s %-14s %-14s %-14s",
		t(m.lang, "col.num"), t(m.lang, "col.brand"), t(m.lang, "col.model"), t(m.lang, "col.plate"), t(m.lang, "col.owner"))
	header := s.subtitle.Render(hdr)
	divider := s.dim.Render("  " + strings.Repeat("─", 60))

	var rows []string
	for i, v := range m.vehicles {
		row := fmt.Sprintf("  %-3d %-14s %-14s %-14s %-14s",
			i+1,
			truncate(v.Brand, 13),
			truncate(v.Model, 13),
			truncate(v.LicensePlate, 13),
			truncate(v.Owner, 13),
		)
		if i == m.vehicleCursor {
			row = s.menuSelected.Width(0).Render(row)
		} else {
			row = s.info.Render(row)
		}
		rows = append(rows, row)
	}
	table := strings.Join(rows, "\n")

	help := s.dim.Render(fmt.Sprintf("a: %s  e: %s  d: %s  ←: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))

	return title + "\n" + header + "\n" + divider + "\n" + table + "\n\n" + help
}

func (m *model) renderVehicleForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	var fields []string
	for i := 0; i < fMgmtCount; i++ {
		// Use localized field label
		label := s.dim.Render(fmt.Sprintf("  %-15s", t(m.lang, mgmtFieldKeys[i])+":"))
		value := m.formFields[i]

		var rendered string
		if i == m.formCursor {
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

func (m *model) renderVehicleDeleteConfirm(s *styles) string {
	if m.vehicleCursor >= len(m.vehicles) {
		m.vehicleView = vViewList
		return m.renderVehicleList(s)
	}
	v := m.vehicles[m.vehicleCursor]

	title := s.title.Render(t(m.lang, "delete.vehicle"))
	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Render(t(m.lang, "delete.confirmVehicle"))

	info := fmt.Sprintf(
		"\n  %s %s\n  %s %s\n  %s %s\n  %s %s",
		s.dim.Render(t(m.lang, "field.brand")+":"), s.info.Render(v.Brand),
		s.dim.Render(t(m.lang, "field.model")+":"), s.info.Render(v.Model),
		s.dim.Render(t(m.lang, "field.licensePlate")+":"), s.info.Render(v.LicensePlate),
		s.dim.Render(t(m.lang, "field.owner")+":"), s.info.Render(v.Owner),
	)

	help := s.dim.Render(fmt.Sprintf("y: %s  n/Esc: %s", t(m.lang, "action.confirm"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + info + "\n\n" + help
}

// renderSingleFieldSection renders Insurance/RoadTax/NTC section.
func (m *model) renderSingleFieldSection(s *styles, sectionName string) string {
	title := s.title.Render(sectionName)

	if m.vehicleView == vViewEdit {
		return m.renderSingleFieldEdit(s, sectionName)
	}

	if len(m.vehicles) == 0 {
		empty := s.dim.Render(t(m.lang, "vehicles.addFirst"))
		help := s.dim.Render("\n\n←: " + t(m.lang, "help.goBack"))
		return title + "\n\n" + empty + help
	}

	hdr := fmt.Sprintf("  %-3s %-14s %-14s %-10s %-12s",
		t(m.lang, "col.num"), t(m.lang, "col.brand"), t(m.lang, "col.model"), t(m.lang, "col.cost"), t(m.lang, "col.expires"))
	header := s.subtitle.Render(hdr)
	divider := s.dim.Render("  " + strings.Repeat("─", 57))

	var rows []string
	for i, v := range m.vehicles {
		var cost, dateStr string
		switch m.vehicleSection {
		case vSectionRoadTax:
			cost = v.RoadTaxCost
			if !v.RoadTax.IsZero() {
				dateStr = v.RoadTax.Format("02/01/2006")
			}
		case vSectionNTC:
			cost = v.NTCCost
			if !v.NTC.IsZero() {
				dateStr = v.NTC.Format("02/01/2006")
			}
		}
		
		if cost == "" {
			cost = "-"
		} else {
			cost = "€ " + cost
		}
		if dateStr == "" {
			dateStr = "-"
		}
		
		row := fmt.Sprintf("  %-3d %-14s %-14s %-10s %-12s",
			i+1,
			truncate(v.Brand, 13),
			truncate(v.Model, 13),
			truncate(cost, 9),
			dateStr,
		)
		if i == m.vehicleCursor {
			row = s.menuSelected.Width(0).Render(row)
		} else {
			row = s.info.Render(row)
		}
		rows = append(rows, row)
	}
	table := strings.Join(rows, "\n")

	help := s.dim.Render(fmt.Sprintf("e/Enter: %s  ←: %s", t(m.lang, "action.edit"), t(m.lang, "help.goBack")))

	return title + "\n" + header + "\n" + divider + "\n" + table + "\n\n" + help
}

func (m *model) renderSingleFieldEdit(s *styles, sectionName string) string {
	title := s.title.Render(fmt.Sprintf("Edit %s", sectionName))

	if m.editIndex >= len(m.vehicles) {
		m.vehicleView = vViewList
		return ""
	}
	v := m.vehicles[m.editIndex]
	vehicleInfo := s.dim.Render(fmt.Sprintf("  Vehicle: %s %s (%s)", v.Brand, v.Model, v.LicensePlate))

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236"))

	labels := []string{t(m.lang, "col.cost"), t(m.lang, "col.expires")}
	var fields []string
	for i := 0; i < 2; i++ {
		label := s.dim.Render(fmt.Sprintf("  %-15s", labels[i]+":"))
		value := m.formFields[i]
		cursor := ""
		if m.formCursor == i {
			cursor = s.highlight.Render("_")
		}

		valDisp := value
		if valDisp == "" && i == 1 {
			valDisp = s.dim.Render("GG/MM/AAAA")
			fields = append(fields, label+" "+valDisp+cursor)
		} else if valDisp == "" && i == 0 {
			valDisp = s.dim.Render("0.00")
			fields = append(fields, label+" € "+valDisp+cursor)
		} else {
			if i == 0 {
				fields = append(fields, label+" € "+fieldStyle.Render(valDisp)+cursor)
			} else {
				fields = append(fields, label+" "+fieldStyle.Render(valDisp)+cursor)
			}
		}
	}
	formContent := strings.Join(fields, "\n\n")

	help := s.dim.Render(fmt.Sprintf("Tab/↑/↓: %s  Enter: %s  Esc: %s", t(m.lang, "help.navigate"), t(m.lang, "action.save"), t(m.lang, "action.cancel")))

	return title + "\n\n" + vehicleInfo + "\n\n" + formContent + "\n\n" + help
}

// ─── Helpers ─────────────────────────────────────────────────────────

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
