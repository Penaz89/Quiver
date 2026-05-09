package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	fSectionAnalytics                       // Analytics (Statistiche)
	fSectionDaily                           // Daily Expenses (Spese Quotidiane)
	fSectionHousing                         // Housing (Casa)
	fSectionSubscriptions                   // Subscriptions (Abbonamenti)
	fSectionSalaries                        // Salaries (Stipendi)
	fSectionGoals                           // Goals (Obiettivi)
)

const (
	fViewList finSubView = iota
	fViewAdd
	fViewEdit
	fViewDelete
)

const (
	dailyFDate = iota
	dailyFCategory
	dailyFDescription
	dailyFAmount
	dailyFCount
)

const (
	houseFExpense = iota
	houseFType
	houseFCost
	houseFCount
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
			if m.finMenuCursor < 7-1 { // 7 items
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
	case fSectionAnalytics:
		return m.updateAnalytics(msg)
	case fSectionDaily:
		return m.updateDaily(msg)
	case fSectionHousing:
		return m.updateHousing(msg)
	case fSectionSubscriptions:
		return m.updateSubscriptions(msg)
	case fSectionSalaries:
		return m.updateSalaries(msg)
	case fSectionGoals:
		return m.updateGoals(msg)
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

	labels := []string{
		strings.ToUpper(t(m.lang, "finances.fixedExp")),
		t(m.lang, "finances.analytics"),
		t(m.lang, "finances.daily"),
		t(m.lang, "finances.housing"),
		t(m.lang, "finances.subscriptions"),
		t(m.lang, "finances.salaries"),
		t(m.lang, "finances.goals"),
	}
	var lines []string
	for i, l := range labels {
		if m.finSection == fSectionMenu && m.finMenuCursor == i {
			if m.focusContent {
				lines = append(lines, s.menuSelected.Width(submenuWidth).Render(l))
			} else {
				lines = append(lines, s.menuActiveDim.Width(submenuWidth).Render(l))
			}
		} else if m.finSection == finSection(i+1) {
			lines = append(lines, s.menuActiveDim.Width(submenuWidth).Render(l))
		} else {
			lines = append(lines, s.menuNormal.Width(submenuWidth).Render(l))
		}
		
		if i == 1 {
			lines = append(lines, s.dim.Render(strings.Repeat("─", submenuWidth)))
		}
	}
	menu := strings.Join(lines, "\n")
	col2 := title + "\n" + desc + "\n\n" + menu

	targetSection := m.finSection
	if targetSection == fSectionMenu {
		targetSection = finSection(m.finMenuCursor + 1)
	}

	var col3 string
	switch targetSection {
	case fSectionFixedExp:
		col3 = m.renderFixedExpenses(s)
	case fSectionAnalytics:
		col3 = m.renderAnalytics(s)
	case fSectionDaily:
		col3 = m.renderDaily(s)
	case fSectionHousing:
		col3 = m.renderHousing(s)
	case fSectionSubscriptions:
		col3 = m.renderSubscriptions(s)
	case fSectionSalaries:
		col3 = m.renderSalaries(s)
	case fSectionGoals:
		col3 = m.renderGoals(s)
	}

	if m.finSection == fSectionMenu {
		// Optional: add a visual indicator that it's a preview or change help text
		// But for now, just rendering the target section is enough.
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
		BorderForeground(lipgloss.Color(m.theme.Border)).
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
		// Tagliando
		annualTotal += parseEuro(v.ServiceCost)

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

	// --- CATEGORY: CASA ---
	var houseBlock string
	if len(m.housing) > 0 {
		catHouseTitle := s.info.Render("  " + strings.ToUpper(t(m.lang, "finances.housing")))
		
		hdrHouse := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "col.expense"), t(m.lang, "col.annual"), t(m.lang, "col.monthly"))
		headerHouse := s.subtitle.Render(hdrHouse)
		dividerHouse := s.dim.Render("  " + strings.Repeat("─", 63))
		
		var houseRows []string
		var houseCatAnnual float64
		var houseCatMonthly float64

		for _, h := range m.housing {
			var annualTotal float64
			cost := parseEuro(h.Cost)
			
			if h.Type == "type.monthly" {
				annualTotal = cost * 12.0
			} else {
				annualTotal = cost
			}

			monthlyTotal := annualTotal / 12.0
			
			houseCatAnnual += annualTotal
			houseCatMonthly += monthlyTotal

			annStr := fmt.Sprintf("€ %.2f", annualTotal)
			monStr := fmt.Sprintf("€ %.2f", monthlyTotal)
			
			row := fmt.Sprintf("  %-31s %-14s %-14s",
				truncate(h.Expense, 30),
				annStr,
				monStr,
			)
			houseRows = append(houseRows, row)
		}
		
		grandTotalAnnual += houseCatAnnual
		grandTotalMonthly += houseCatMonthly

		houseTable := strings.Join(houseRows, "\n")
		
		houseSubtotalStr := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "finances.subtotal")+" "+t(m.lang, "finances.housing"),
			fmt.Sprintf("€ %.2f", houseCatAnnual),
			fmt.Sprintf("€ %.2f", houseCatMonthly),
		)
		
		houseBlock = "\n\n" + catHouseTitle + "\n" + dividerHouse + "\n" + headerHouse + "\n" + dividerHouse + "\n" + houseTable + "\n" + dividerHouse + "\n" + s.highlight.Render(houseSubtotalStr)
	}



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

	// --- CATEGORY: SPESE GIORNALIERE ---
	var dailyBlock string
	if len(m.daily) > 0 {
		now := time.Now()
		var monthlyTotal float64
		for _, exp := range m.daily {
			if !exp.Date.IsZero() && exp.Date.Year() == now.Year() && exp.Date.Month() == now.Month() {
				monthlyTotal += parseEuro(exp.Amount)
			}
		}
		
		if monthlyTotal > 0 {
			catDailyTitle := s.info.Render("  " + strings.ToUpper(t(m.lang, "finances.daily")))
			
			hdrDaily := fmt.Sprintf("  %-31s %-14s %-14s",
				t(m.lang, "finances.daily"), t(m.lang, "col.annual"), t(m.lang, "col.monthly"))
			headerDaily := s.subtitle.Render(hdrDaily)
			dividerDaily := s.dim.Render("  " + strings.Repeat("─", 63))
			
			annualTotal := monthlyTotal * 12.0
			grandTotalAnnual += annualTotal
			grandTotalMonthly += monthlyTotal
			
			monthKey := fmt.Sprintf("month.%02d", now.Month())
			row := fmt.Sprintf("  %-31s %-14s %-14s",
				truncate(t(m.lang, monthKey) + " " + fmt.Sprintf("%d", now.Year()), 30),
				fmt.Sprintf("€ %.2f", annualTotal),
				fmt.Sprintf("€ %.2f", monthlyTotal),
			)
			
			dailySubtotalStr := fmt.Sprintf("  %-31s %-14s %-14s",
				t(m.lang, "finances.subtotal")+" "+t(m.lang, "finances.daily"),
				fmt.Sprintf("€ %.2f", annualTotal),
				fmt.Sprintf("€ %.2f", monthlyTotal),
			)
			
			dailyBlock = "\n\n" + catDailyTitle + "\n" + dividerDaily + "\n" + headerDaily + "\n" + dividerDaily + "\n" + row + "\n" + dividerDaily + "\n" + s.highlight.Render(dailySubtotalStr)
		}
	}

	// --- CATEGORY: GOALS ---
	var goalBlock string
	if len(m.goals) > 0 {
		catGoalTitle := s.info.Render("  " + t(m.lang, "cat.goals"))
		
		hdrGoal := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "finances.goals"), t(m.lang, "col.goal"), t(m.lang, "col.monthly"))
		headerGoal := s.subtitle.Render(hdrGoal)
		dividerGoal := s.dim.Render("  " + strings.Repeat("─", 63))
		
		var goalRows []string
		var goalCatTotal float64
		var goalCatMonthly float64

		for _, g := range m.goals {
			target := parseEuro(g.Target)
			current := parseEuro(g.Current)
			remainingMoney := target - current
			var monthlyNeeded float64
			
			if remainingMoney > 0 && !g.Deadline.IsZero() {
				now := time.Now()
				remainingMonths := g.Deadline.Sub(now).Hours() / (24 * 30.44)
				if remainingMonths < 1 {
					remainingMonths = 1
				}
				monthlyNeeded = remainingMoney / remainingMonths
			}

			if monthlyNeeded > 0 {
				goalCatTotal += remainingMoney
				goalCatMonthly += monthlyNeeded

				annStr := fmt.Sprintf("€ %.2f", remainingMoney)
				monStr := fmt.Sprintf("€ %.2f", monthlyNeeded)
				
				row := fmt.Sprintf("  %-31s %-14s %-14s",
					truncate(g.Name, 30),
					annStr,
					monStr,
				)
				goalRows = append(goalRows, row)
			}
		}
		
		if len(goalRows) > 0 {
			// Do NOT add goalCatTotal to grandTotalAnnual because goals are not annual expenses
			grandTotalMonthly += goalCatMonthly

			goalTable := strings.Join(goalRows, "\n")
			
			goalSubtotalStr := fmt.Sprintf("  %-31s %-14s %-14s",
				t(m.lang, "finances.subtotal")+" "+t(m.lang, "finances.goals"),
				fmt.Sprintf("€ %.2f", goalCatTotal),
				fmt.Sprintf("€ %.2f", goalCatMonthly),
			)
			
			goalBlock = "\n\n" + catGoalTitle + "\n" + dividerGoal + "\n" + headerGoal + "\n" + dividerGoal + "\n" + goalTable + "\n" + dividerGoal + "\n" + s.highlight.Render(goalSubtotalStr)
		}
	}

	// --- GRAND TOTAL ---
	grandDivider := s.dim.Render("  " + strings.Repeat("═", 63))
	grandStr := fmt.Sprintf("  %-31s %-14s %-14s",
		t(m.lang, "finances.grandTotal"),
		fmt.Sprintf("€ %.2f", grandTotalAnnual),
		fmt.Sprintf("€ %.2f", grandTotalMonthly),
	)
	
	grandBlock := grandDivider + "\n" + s.title.Render(grandStr) + "\n" + grandDivider

	content := vehBlock + houseBlock + subBlock + dailyBlock + goalBlock + "\n\n\n" + grandBlock
	help := s.dim.Render(fmt.Sprintf("\n\n←: %s", t(m.lang, "help.goBack")))
	
	return title + "\n\n" + content + help
}

// ─── Housing logic ───────────────────────────────────────────────────

func (m *model) updateHousing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.finView {
	case fViewAdd, fViewEdit:
		return m.updateHouseForm(msg)
	case fViewDelete:
		return m.updateHouseDelete(msg)
	default:
		return m.updateHouseList(msg)
	}
}

func (m *model) updateHouseList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.houseCursor > 0 {
			m.houseCursor--
		}
	case "down", "j":
		if m.houseCursor < len(m.housing)-1 {
			m.houseCursor++
		}
	case "a":
		m.finView = fViewAdd
		m.houseForm = [houseFCount]string{"", "type.monthly", ""}
		m.houseFormCur = 0
	case "e", "enter":
		if len(m.housing) > 0 {
			m.finView = fViewEdit
			m.houseEditIdx = m.houseCursor
			hType := m.housing[m.houseCursor].Type
			if hType == "" {
				hType = "type.monthly"
			}
			m.houseForm = [houseFCount]string{
				m.housing[m.houseCursor].Expense,
				hType,
				m.housing[m.houseCursor].Cost,
			}
			m.houseFormCur = 0
		}
	case "d", "x":
		if len(m.housing) > 0 {
			m.finView = fViewDelete
		}
	case "esc", "left":
		m.finSection = fSectionMenu
	}
	return m, nil
}

func (m *model) updateHouseForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		if m.houseFormCur < houseFCount-1 {
			m.houseFormCur++
		} else {
			m.houseFormCur = 0
		}
	case "shift+tab", "up":
		if m.houseFormCur > 0 {
			m.houseFormCur--
		} else {
			m.houseFormCur = houseFCount - 1
		}
	case "left", "right":
		if m.houseFormCur == houseFType {
			if m.houseForm[houseFType] == "type.monthly" {
				m.houseForm[houseFType] = "type.annual"
			} else {
				m.houseForm[houseFType] = "type.monthly"
			}
		}
	case " ":
		if m.houseFormCur == houseFType {
			if m.houseForm[houseFType] == "type.monthly" {
				m.houseForm[houseFType] = "type.annual"
			} else {
				m.houseForm[houseFType] = "type.monthly"
			}
		} else {
			m.houseForm[m.houseFormCur] += key
		}
	case "enter":
		h := storage.Housing{
			Expense: strings.TrimSpace(m.houseForm[houseFExpense]),
			Type:    strings.TrimSpace(m.houseForm[houseFType]),
			Cost:    strings.TrimSpace(m.houseForm[houseFCost]),
		}
		if h.Expense == "" {
			m.finView = fViewList
			return m, nil
		}
		if m.finView == fViewAdd {
			m.housing = append(m.housing, h)
		} else {
			m.housing[m.houseEditIdx] = h
		}
		_ = storage.SaveHousing(m.dataDir, m.housing)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		if m.houseFormCur != houseFType {
			field := &m.houseForm[m.houseFormCur]
			if len(*field) > 0 {
				runes := []rune(*field)
				*field = string(runes[:len(runes)-1])
			}
		}
	default:
		if key == "space" {
			key = " "
		}
		if m.houseFormCur != houseFType {
			runes := []rune(key)
			if len(runes) == 1 {
				field := &m.houseForm[m.houseFormCur]
				if m.houseFormCur == houseFCost {
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

func (m *model) updateHouseDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "s", "S":
		m.housing = append(m.housing[:m.houseCursor], m.housing[m.houseCursor+1:]...)
		_ = storage.SaveHousing(m.dataDir, m.housing)
		if m.houseCursor >= len(m.housing) && m.houseCursor > 0 {
			m.houseCursor--
		}
		m.finView = fViewList
	case "n", "esc":
		m.finView = fViewList
	}
	return m, nil
}

func (m *model) renderHousing(s *styles) string {
	switch m.finView {
	case fViewAdd:
		return m.renderHouseForm(s, t(m.lang, "action.add")+" "+t(m.lang, "finances.housing"))
	case fViewEdit:
		return m.renderHouseForm(s, t(m.lang, "action.edit")+" "+t(m.lang, "finances.housing"))
	case fViewDelete:
		return m.renderHouseDelete(s)
	default:
		return m.renderHouseList(s)
	}
}

func (m *model) renderHouseList(s *styles) string {
	isActive := m.finSection != fSectionMenu && m.focusContent
	title := s.title.Render(t(m.lang, "finances.housing"))
	if len(m.housing) == 0 {
		empty := s.dim.Render(t(m.lang, "housing.noRecords"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	hdr := fmt.Sprintf("  %-3s %-20s %-14s %-14s",
		t(m.lang, "col.num"), t(m.lang, "col.expense"), t(m.lang, "col.type"), t(m.lang, "col.cost"))
	header := s.subtitle.Render(hdr)
	divider := s.dim.Render("  " + strings.Repeat("─", 54))

	var rows []string
	for i, h := range m.housing {
		cost := h.Cost
		if cost == "" {
			cost = "-"
		} else {
			cost = "€ " + cost
		}

		hType := h.Type
		if hType != "" {
			hType = t(m.lang, hType)
		} else {
			hType = "-"
		}

		row := fmt.Sprintf("  %-3d %-20s %-14s %-14s",
			i+1,
			truncate(h.Expense, 19),
			truncate(hType, 13),
			truncate(cost, 13),
		)
		if i == m.houseCursor {
			if isActive {
				row = s.menuSelected.Width(0).Render(row)
			} else {
				row = s.menuActiveDim.Width(0).Render(row)
			}
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

func (m *model) renderHouseForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	labels := []string{
		t(m.lang, "field.expense"),
		t(m.lang, "field.insType"),
		t(m.lang, "field.totalCost"),
	}

	var fields []string
	for i := 0; i < houseFCount; i++ {
		label := s.dim.Render(fmt.Sprintf("  %-15s", labels[i]+":"))
		val := m.houseForm[i]
		
		var rendered string
		if i == m.houseFormCur {
			cursor := s.highlight.Render("_")
			fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
			if i == houseFType {
				rendered = label + " < " + fieldStyle.Render(t(m.lang, val)) + " >"
			} else if i == houseFCost {
				if val == "" { val = s.dim.Render("0.00") }
				rendered = label + " € " + fieldStyle.Render(val) + cursor
			} else {
				rendered = label + " " + fieldStyle.Render(val) + cursor
			}
		} else {
			if i == houseFType {
				rendered = label + " " + s.info.Render(t(m.lang, val))
			} else if i == houseFCost {
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

func (m *model) renderHouseDelete(s *styles) string {
	if m.houseCursor >= len(m.housing) {
		return m.renderHouseList(s)
	}
	h := m.housing[m.houseCursor]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.housing"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(t(m.lang, "delete.confirmInsurance"))

	info := fmt.Sprintf(
		"\n  %s %s\n  %s € %s",
		s.dim.Render(t(m.lang, "field.expense")+":"), s.info.Render(h.Expense),
		s.dim.Render(t(m.lang, "field.totalCost")+":"), s.info.Render(h.Cost),
	)

	help := s.dim.Render(fmt.Sprintf("\n\ny: %s  n/Esc: %s",
		t(m.lang, "action.delete"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + "\n" + info + help
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
	case "left", "right":
		if m.subFormCur == subFType {
			if m.subForm[subFType] == "type.monthly" {
				m.subForm[subFType] = "type.annual"
			} else {
				m.subForm[subFType] = "type.monthly"
			}
		}
	case " ":
		if m.subFormCur == subFType {
			if m.subForm[subFType] == "type.monthly" {
				m.subForm[subFType] = "type.annual"
			} else {
				m.subForm[subFType] = "type.monthly"
			}
		} else {
			m.subForm[m.subFormCur] += key
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
		if key == "space" {
			key = " "
		}
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
	case "y", "Y", "s", "S":
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
	isActive := m.finSection != fSectionMenu && m.focusContent
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
			if isActive {
				row = s.menuSelected.Width(0).Render(row)
			} else {
				row = s.menuActiveDim.Width(0).Render(row)
			}
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
