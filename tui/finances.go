package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ─── Finances sub-menu ───────────────────────────────────────────────

type finSection int

const (
	fSectionMenu     finSection = iota // sub-menu view
	fSectionFixedExp                   // Fixed Expenses (Spese Fisse)
)

func (m *model) updateFinances(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.finSection == fSectionMenu {
		switch msg.String() {
		case "up", "k":
			if m.finMenuCursor > 0 {
				m.finMenuCursor--
			}
		case "down", "j":
			if m.finMenuCursor < 1-1 { // Currently 1 item
				m.finMenuCursor++
			}
		case "enter", "right":
			m.finSection = finSection(m.finMenuCursor + 1)
		case "esc", "left":
			m.focusContent = false
		}
		return m, nil
	}

	// Route to specific section
	switch m.finSection {
	case fSectionFixedExp:
		return m.updateFixedExpenses(msg)
	}
	return m, nil
}

func (m *model) updateFixedExpenses(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc", "left":
		m.finSection = fSectionMenu
	}
	return m, nil
}

// ─── Render ──────────────────────────────────────────────────────────

func (m *model) renderFinancesView(s *styles) string {
	sw := sidebarWidth(m.width)
	submenuWidth := sw - 4
	if submenuWidth < 10 {
		submenuWidth = 10
	}

	title := s.title.Render(t(m.lang, "finances.title"))
	desc := s.subtitle.Render(t(m.lang, "finances.subtitle"))

	labels := []string{t(m.lang, "finances.fixedExp")}
	var lines []string
	for i, l := range labels {
		if m.finSection == fSectionMenu && m.finMenuCursor == i {
			lines = append(lines, s.menuSelected.Width(submenuWidth).Render(l))
		} else if m.finSection == finSection(i+1) {
			lines = append(lines, s.menuActiveDim.Width(submenuWidth).Render(l))
		} else {
			lines = append(lines, s.menuNormal.Width(submenuWidth).Render(l))
		}
	}
	menu := strings.Join(lines, "\n")
	col2 := title + "\n" + desc + "\n\n" + menu

	var col3 string
	switch m.finSection {
	case fSectionFixedExp:
		col3 = m.renderFixedExpenses(s)
	default:
		// Preview mode
		placeholder := s.dim.Render(t(m.lang, "finances.noEntries"))
		help := s.dim.Render(fmt.Sprintf("\n\n↑/↓: %s  →: %s  ←: %s",
			t(m.lang, "help.navigate"), t(m.lang, "help.enter"), t(m.lang, "help.goBack")))
		col3 = "\n" + placeholder + help
	}

	// Calculate heights to stretch the divider
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
		BorderForeground(lipgloss.Color("63")).
		PaddingRight(2).
		MarginRight(2).
		Height(maxHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, col2Style.Render(col2), col3)
}

func parseEuro(s string) float64 {
	s = strings.ReplaceAll(s, ",", ".")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (m *model) renderFixedExpenses(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.fixedExp"))
	if len(m.vehicles) == 0 {
		empty := s.dim.Render(t(m.lang, "vehicles.noVehicles"))
		help := s.dim.Render(fmt.Sprintf("\n\n←: %s", t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	var grandTotalAnnual float64
	var grandTotalMonthly float64

	// --- CATEGORY: VEICOLI ---
	catVehTitle := s.info.Render("  " + t(m.lang, "cat.vehicles"))
	
	hdrVeh := fmt.Sprintf("  %-14s %-16s %-14s %-14s",
		t(m.lang, "col.plate"), t(m.lang, "col.model"), t(m.lang, "col.annual"), t(m.lang, "col.monthly"))
	headerVeh := s.subtitle.Render(hdrVeh)
	dividerVeh := s.dim.Render("  " + strings.Repeat("─", 63))
	
	var vehRows []string
	var vehCatAnnual float64
	var vehCatMonthly float64

	for _, v := range m.vehicles {
		var annualTotal float64

		// Bollo
		annualTotal += parseEuro(v.RoadTaxCost)
		// Revisione
		annualTotal += parseEuro(v.NTCCost) / 2.0

		// Insurance
		for _, ins := range m.insurances {
			if ins.LicensePlate == v.LicensePlate {
				cost := parseEuro(ins.TotalCost)
				if ins.Type == "type.semiannual" {
					annualTotal += cost * 2.0
				} else {
					annualTotal += cost
				}
			}
		}

		monthlyTotal := annualTotal / 12.0
		
		vehCatAnnual += annualTotal
		vehCatMonthly += monthlyTotal

		annStr := fmt.Sprintf("€ %.2f", annualTotal)
		monStr := fmt.Sprintf("€ %.2f", monthlyTotal)
		
		row := fmt.Sprintf("  %-14s %-16s %-14s %-14s",
			truncate(v.LicensePlate, 13),
			truncate(v.Brand+" "+v.Model, 15),
			annStr,
			monStr,
		)
		vehRows = append(vehRows, row)
	}
	
	grandTotalAnnual += vehCatAnnual
	grandTotalMonthly += vehCatMonthly

	vehTable := strings.Join(vehRows, "\n")
	
	subtotalStr := fmt.Sprintf("  %-31s %-14s %-14s",
		t(m.lang, "finances.subtotal")+" "+t(m.lang, "cat.vehicles"),
		fmt.Sprintf("€ %.2f", vehCatAnnual),
		fmt.Sprintf("€ %.2f", vehCatMonthly),
	)
	
	vehBlock := catVehTitle + "\n" + dividerVeh + "\n" + headerVeh + "\n" + dividerVeh + "\n" + vehTable + "\n" + dividerVeh + "\n" + s.highlight.Render(subtotalStr)

	// --- GRAND TOTAL ---
	grandDivider := s.dim.Render("  " + strings.Repeat("═", 63))
	grandStr := fmt.Sprintf("  %-31s %-14s %-14s",
		t(m.lang, "finances.grandTotal"),
		fmt.Sprintf("€ %.2f", grandTotalAnnual),
		fmt.Sprintf("€ %.2f", grandTotalMonthly),
	)
	
	grandBlock := grandDivider + "\n" + s.title.Render(grandStr) + "\n" + grandDivider

	content := vehBlock + "\n\n\n" + grandBlock
	help := s.dim.Render(fmt.Sprintf("\n\n←: %s", t(m.lang, "help.goBack")))
	
	return title + "\n\n" + content + help
}
