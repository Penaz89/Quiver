package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

// ─── Finances sub-menu ───────────────────────────────────────────────

type finSection int
type finSubView int

const (
	fSectionMenu          finSection = iota // sub-menu view
	fSectionFixedExp                        // Fixed Expenses (Spese Fisse)
	fSectionSubscriptions                   // Subscriptions (Abbonamenti)
)

const (
	fViewList finSubView = iota
	fViewAdd
	fViewEdit
	fViewDelete
)

const (
	subFService = iota
	subFType
	subFCost
	subFCount
)

func (m *model) updateFinances(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.finSection == fSectionMenu {
		switch msg.String() {
		case "up", "k":
			if m.finMenuCursor > 0 {
				m.finMenuCursor--
			}
		case "down", "j":
			if m.finMenuCursor < 2-1 { // 2 items
				m.finMenuCursor++
			}
		case "enter", "right":
			m.finSection = finSection(m.finMenuCursor + 1)
			m.finView = fViewList
		case "esc", "left":
			m.focusContent = false
		}
		return m, nil
	}

	// Route to specific section
	switch m.finSection {
	case fSectionFixedExp:
		return m.updateFixedExpenses(msg)
	case fSectionSubscriptions:
		return m.updateSubscriptions(msg)
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

	labels := []string{t(m.lang, "finances.fixedExp"), t(m.lang, "finances.subscriptions")}
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
	case fSectionSubscriptions:
		col3 = m.renderSubscriptions(s)
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

	// --- CATEGORY: ABBONAMENTI ---
	var subBlock string
	if len(m.subs) > 0 {
		catSubTitle := s.info.Render("  " + strings.ToUpper(t(m.lang, "finances.subscriptions")))
		
		hdrSub := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "col.service"), t(m.lang, "col.annual"), t(m.lang, "col.monthly"))
		headerSub := s.subtitle.Render(hdrSub)
		dividerSub := s.dim.Render("  " + strings.Repeat("─", 63))
		
		var subRows []string
		var subCatAnnual float64
		var subCatMonthly float64

		for _, sub := range m.subs {
			var annualTotal float64
			cost := parseEuro(sub.Cost)
			
			if sub.Type == "type.monthly" {
				annualTotal = cost * 12.0
			} else {
				annualTotal = cost
			}

			monthlyTotal := annualTotal / 12.0
			
			subCatAnnual += annualTotal
			subCatMonthly += monthlyTotal

			annStr := fmt.Sprintf("€ %.2f", annualTotal)
			monStr := fmt.Sprintf("€ %.2f", monthlyTotal)
			
			row := fmt.Sprintf("  %-31s %-14s %-14s",
				truncate(sub.Service, 30),
				annStr,
				monStr,
			)
			subRows = append(subRows, row)
		}
		
		grandTotalAnnual += subCatAnnual
		grandTotalMonthly += subCatMonthly

		subTable := strings.Join(subRows, "\n")
		
		subSubtotalStr := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "finances.subtotal")+" "+t(m.lang, "finances.subscriptions"),
			fmt.Sprintf("€ %.2f", subCatAnnual),
			fmt.Sprintf("€ %.2f", subCatMonthly),
		)
		
		subBlock = "\n\n" + catSubTitle + "\n" + dividerSub + "\n" + headerSub + "\n" + dividerSub + "\n" + subTable + "\n" + dividerSub + "\n" + s.highlight.Render(subSubtotalStr)
	}

	// --- GRAND TOTAL ---
	grandDivider := s.dim.Render("  " + strings.Repeat("═", 63))
	grandStr := fmt.Sprintf("  %-31s %-14s %-14s",
		t(m.lang, "finances.grandTotal"),
		fmt.Sprintf("€ %.2f", grandTotalAnnual),
		fmt.Sprintf("€ %.2f", grandTotalMonthly),
	)
	
	grandBlock := grandDivider + "\n" + s.title.Render(grandStr) + "\n" + grandDivider

	content := vehBlock + subBlock + "\n\n\n" + grandBlock
	help := s.dim.Render(fmt.Sprintf("\n\n←: %s", t(m.lang, "help.goBack")))
	
	return title + "\n\n" + content + help
}

// ─── Subscriptions logic ─────────────────────────────────────────────

func (m *model) updateSubscriptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.finView {
	case fViewAdd, fViewEdit:
		return m.updateSubForm(msg)
	case fViewDelete:
		return m.updateSubDelete(msg)
	default:
		return m.updateSubList(msg)
	}
}

func (m *model) updateSubList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.subCursor > 0 {
			m.subCursor--
		}
	case "down", "j":
		if m.subCursor < len(m.subs)-1 {
			m.subCursor++
		}
	case "a":
		m.finView = fViewAdd
		m.subForm = [subFCount]string{"", "type.monthly", ""}
		m.subFormCur = 0
	case "e", "enter":
		if len(m.subs) > 0 {
			m.finView = fViewEdit
			m.subEditIdx = m.subCursor
			subType := m.subs[m.subCursor].Type
			if subType == "" {
				subType = "type.monthly"
			}
			m.subForm = [subFCount]string{
				m.subs[m.subCursor].Service,
				subType,
				m.subs[m.subCursor].Cost,
			}
			m.subFormCur = 0
		}
	case "d", "x":
		if len(m.subs) > 0 {
			m.finView = fViewDelete
		}
	case "esc", "left":
		m.finSection = fSectionMenu
	}
	return m, nil
}

func (m *model) updateSubForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		if m.subFormCur < subFCount-1 {
			m.subFormCur++
		} else {
			m.subFormCur = 0
		}
	case "shift+tab", "up":
		if m.subFormCur > 0 {
			m.subFormCur--
		} else {
			m.subFormCur = subFCount - 1
		}
	case "left", "right", " ":
		if m.subFormCur == subFType {
			if m.subForm[subFType] == "type.monthly" {
				m.subForm[subFType] = "type.annual"
			} else {
				m.subForm[subFType] = "type.monthly"
			}
		}
	case "enter":
		sub := storage.Subscription{
			Service: strings.TrimSpace(m.subForm[subFService]),
			Type:    strings.TrimSpace(m.subForm[subFType]),
			Cost:    strings.TrimSpace(m.subForm[subFCost]),
		}
		if sub.Service == "" {
			m.finView = fViewList
			return m, nil
		}
		if m.finView == fViewAdd {
			m.subs = append(m.subs, sub)
		} else {
			m.subs[m.subEditIdx] = sub
		}
		_ = storage.SaveSubscriptions(m.dataDir, m.subs)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		if m.subFormCur != subFType {
			field := &m.subForm[m.subFormCur]
			if len(*field) > 0 {
				runes := []rune(*field)
				*field = string(runes[:len(runes)-1])
			}
		}
	default:
		if m.subFormCur != subFType {
			runes := []rune(key)
			if len(runes) == 1 {
				field := &m.subForm[m.subFormCur]
				if m.subFormCur == subFCost {
					if strings.ContainsRune("0123456789.,", runes[0]) {
						if len(*field) < 15 {
							*field += key
						}
					}
				} else {
					*field += key
				}
			}
		}
	}
	return m, nil
}

func (m *model) updateSubDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.subs = append(m.subs[:m.subCursor], m.subs[m.subCursor+1:]...)
		_ = storage.SaveSubscriptions(m.dataDir, m.subs)
		if m.subCursor >= len(m.subs) && m.subCursor > 0 {
			m.subCursor--
		}
		m.finView = fViewList
	case "n", "esc":
		m.finView = fViewList
	}
	return m, nil
}

func (m *model) renderSubscriptions(s *styles) string {
	switch m.finView {
	case fViewAdd:
		return m.renderSubForm(s, t(m.lang, "action.add")+" "+t(m.lang, "finances.subscriptions"))
	case fViewEdit:
		return m.renderSubForm(s, t(m.lang, "action.edit")+" "+t(m.lang, "finances.subscriptions"))
	case fViewDelete:
		return m.renderSubDelete(s)
	default:
		return m.renderSubList(s)
	}
}

func (m *model) renderSubList(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.subscriptions"))
	if len(m.subs) == 0 {
		empty := s.dim.Render(t(m.lang, "subscriptions.noRecords"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	hdr := fmt.Sprintf("  %-3s %-20s %-14s %-14s",
		t(m.lang, "col.num"), t(m.lang, "col.service"), t(m.lang, "col.type"), t(m.lang, "col.cost"))
	header := s.subtitle.Render(hdr)
	divider := s.dim.Render("  " + strings.Repeat("─", 54))

	var rows []string
	for i, sub := range m.subs {
		cost := sub.Cost
		if cost == "" {
			cost = "-"
		} else {
			cost = "€ " + cost
		}

		sType := sub.Type
		if sType != "" {
			sType = t(m.lang, sType)
		} else {
			sType = "-"
		}

		row := fmt.Sprintf("  %-3d %-20s %-14s %-14s",
			i+1,
			truncate(sub.Service, 19),
			truncate(sType, 13),
			truncate(cost, 13),
		)
		if i == m.subCursor {
			row = s.menuSelected.Width(0).Render(row)
		} else {
			row = s.info.Render(row)
		}
		rows = append(rows, row)
	}

	table := strings.Join(rows, "\n")
	help := s.dim.Render(fmt.Sprintf("\n\na: %s  e: %s  d: %s  ←: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))

	return title + "\n" + header + "\n" + divider + "\n" + table + help
}

func (m *model) renderSubForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	labels := []string{
		t(m.lang, "field.service"),
		t(m.lang, "field.insType"),
		t(m.lang, "field.totalCost"),
	}

	var fields []string
	for i := 0; i < subFCount; i++ {
		label := s.dim.Render(fmt.Sprintf("  %-15s", labels[i]+":"))
		val := m.subForm[i]
		
		var rendered string
		if i == m.subFormCur {
			cursor := s.highlight.Render("_")
			fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
			if i == subFType {
				rendered = label + " < " + fieldStyle.Render(t(m.lang, val)) + " >"
			} else if i == subFCost {
				if val == "" { val = s.dim.Render("0.00") }
				rendered = label + " € " + fieldStyle.Render(val) + cursor
			} else {
				rendered = label + " " + fieldStyle.Render(val) + cursor
			}
		} else {
			if i == subFType {
				rendered = label + " " + s.info.Render(t(m.lang, val))
			} else if i == subFCost {
				if val == "" { val = s.dim.Render("0.00") }
				rendered = label + " € " + s.info.Render(val)
			} else {
				rendered = label + " " + s.info.Render(val)
			}
		}
		fields = append(fields, rendered)
	}

	form := strings.Join(fields, "\n\n")
	help := s.dim.Render(fmt.Sprintf("\n\nTab/↑↓: %s  Enter: %s  Esc: %s",
		t(m.lang, "help.switchField"), t(m.lang, "action.save"), t(m.lang, "action.cancel")))

	return title + "\n\n" + form + help
}

func (m *model) renderSubDelete(s *styles) string {
	if m.subCursor >= len(m.subs) {
		return m.renderSubList(s)
	}
	sub := m.subs[m.subCursor]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.subscriptions"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(t(m.lang, "delete.confirmInsurance"))

	info := fmt.Sprintf(
		"\n  %s %s\n  %s € %s",
		s.dim.Render(t(m.lang, "field.service")+":"), s.info.Render(sub.Service),
		s.dim.Render(t(m.lang, "field.totalCost")+":"), s.info.Render(sub.Cost),
	)

	help := s.dim.Render(fmt.Sprintf("\n\ny: %s  n/Esc: %s",
		t(m.lang, "action.delete"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + "\n" + info + help
}
