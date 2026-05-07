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
	fSectionHousing                         // Housing (Casa)
	fSectionHolidays                        // Holidays (Vacanze)
	fSectionSubscriptions                   // Subscriptions (Abbonamenti)
	fSectionSalaries                        // Salaries (Stipendi)
)

const (
	fViewList finSubView = iota
	fViewAdd
	fViewEdit
	fViewDelete
)

const (
	houseFExpense = iota
	houseFType
	houseFCost
	houseFCount
)

const (
	holiFDestination = iota
	holiFFlightDesc
	holiFFlightCost
	holiFAccomDesc
	holiFAccomCost
	holiFCarDesc
	holiFCarCost
	holiFInsDesc
	holiFInsCost
	holiFCount
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
			if m.finMenuCursor < 5-1 { // 5 items
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
	case fSectionHousing:
		return m.updateHousing(msg)
	case fSectionHolidays:
		return m.updateHolidays(msg)
	case fSectionSubscriptions:
		return m.updateSubscriptions(msg)
	case fSectionSalaries:
		return m.updateSalaries(msg)
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

	labels := []string{strings.ToUpper(t(m.lang, "finances.fixedExp")), t(m.lang, "finances.housing"), t(m.lang, "finances.holidays"), t(m.lang, "finances.subscriptions"), t(m.lang, "finances.salaries")}
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
		
		if i == 0 {
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
	case fSectionHousing:
		col3 = m.renderHousing(s)
	case fSectionHolidays:
		col3 = m.renderHolidays(s)
	case fSectionSubscriptions:
		col3 = m.renderSubscriptions(s)
	case fSectionSalaries:
		col3 = m.renderSalaries(s)
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

	// --- CATEGORY: VACANZE ---
	var holiBlock string
	if len(m.holidays) > 0 {
		catHoliTitle := s.info.Render("  " + strings.ToUpper(t(m.lang, "finances.holidays")))
		
		hdrHoli := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "col.destination"), t(m.lang, "col.annual"), t(m.lang, "col.monthly"))
		headerHoli := s.subtitle.Render(hdrHoli)
		dividerHoli := s.dim.Render("  " + strings.Repeat("─", 63))
		
		var holiRows []string
		var holiCatAnnual float64
		var holiCatMonthly float64

		for _, h := range m.holidays {
			fCost := parseEuro(h.FlightCost)
			aCost := parseEuro(h.AccomCost)
			cCost := parseEuro(h.CarCost)
			iCost := parseEuro(h.InsCost)

			annualTotal := fCost + aCost + cCost + iCost
			monthlyTotal := annualTotal / 12.0
			
			holiCatAnnual += annualTotal
			holiCatMonthly += monthlyTotal

			annStr := fmt.Sprintf("€ %.2f", annualTotal)
			monStr := fmt.Sprintf("€ %.2f", monthlyTotal)
			
			row := fmt.Sprintf("  %-31s %-14s %-14s",
				truncate(h.Destination, 30),
				annStr,
				monStr,
			)
			holiRows = append(holiRows, row)
		}
		
		grandTotalAnnual += holiCatAnnual
		grandTotalMonthly += holiCatMonthly

		holiTable := strings.Join(holiRows, "\n")
		
		holiSubtotalStr := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "finances.subtotal")+" "+t(m.lang, "finances.holidays"),
			fmt.Sprintf("€ %.2f", holiCatAnnual),
			fmt.Sprintf("€ %.2f", holiCatMonthly),
		)
		
		holiBlock = "\n\n" + catHoliTitle + "\n" + dividerHoli + "\n" + headerHoli + "\n" + dividerHoli + "\n" + holiTable + "\n" + dividerHoli + "\n" + s.highlight.Render(holiSubtotalStr)
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

	// --- GRAND TOTAL ---
	grandDivider := s.dim.Render("  " + strings.Repeat("═", 63))
	grandStr := fmt.Sprintf("  %-31s %-14s %-14s",
		t(m.lang, "finances.grandTotal"),
		fmt.Sprintf("€ %.2f", grandTotalAnnual),
		fmt.Sprintf("€ %.2f", grandTotalMonthly),
	)
	
	grandBlock := grandDivider + "\n" + s.title.Render(grandStr) + "\n" + grandDivider

	// --- IMPACT ON SALARY ---
	var impactBlock string
	currentYearStr := fmt.Sprintf("%d", time.Now().Year())
	var totalNet float64
	var monthsCount int
	for _, sal := range m.salaries {
		if sal.Year == currentYearStr {
			totalNet += parseEuro(sal.Net)
			monthsCount++
		}
	}
	
	if monthsCount > 0 {
		avgNet := totalNet / float64(monthsCount)
		projectedAnnual := avgNet * 12.0
		impactPct := (grandTotalAnnual / projectedAnnual) * 100.0
		
		impactTitle := s.info.Render("  " + t(m.lang, "finances.salaryImpact"))
		impactDivider := s.dim.Render("  " + strings.Repeat("─", 63))
		
		impactStr := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "finances.projectedAnnual"),
			"",
			fmt.Sprintf("€ %.2f", projectedAnnual),
		)
		
		impactStr2 := fmt.Sprintf("  %-31s %-14s %-14s",
			t(m.lang, "finances.fixedAnnual"),
			"",
			fmt.Sprintf("€ %.2f", grandTotalAnnual),
		)
		
		pctColor := s.highlight
		if impactPct > 35 { 
			pctColor = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
		} else if impactPct < 20 {
			pctColor = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
		}
		
		impactStr3 := fmt.Sprintf("  %-31s %-14s %s",
			t(m.lang, "finances.impactPct"),
			"",
			pctColor.Render(fmt.Sprintf("%.1f%%", impactPct)),
		)
		
		impactBlock = "\n\n" + impactTitle + "\n" + impactDivider + "\n" + impactStr + "\n" + impactStr2 + "\n" + impactStr3 + "\n" + impactDivider
	}

	content := vehBlock + houseBlock + holiBlock + subBlock + "\n\n\n" + grandBlock + impactBlock
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
	isActive := m.finSection != fSectionMenu
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



// ─── Holidays logic ──────────────────────────────────────────────────

func (m *model) updateHolidays(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.finView {
	case fViewAdd, fViewEdit:
		return m.updateHoliForm(msg)
	case fViewDelete:
		return m.updateHoliDelete(msg)
	default:
		return m.updateHoliList(msg)
	}
}

func (m *model) updateHoliList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.holiCursor > 0 {
			m.holiCursor--
		}
	case "down", "j":
		if m.holiCursor < len(m.holidays)-1 {
			m.holiCursor++
		}
	case "a":
		m.finView = fViewAdd
		m.holiForm = [holiFCount]string{}
		m.holiFormCur = 0
	case "e", "enter":
		if len(m.holidays) > 0 {
			m.finView = fViewEdit
			m.holiEditIdx = m.holiCursor
			h := m.holidays[m.holiCursor]
			m.holiForm = [holiFCount]string{
				h.Destination,
				h.FlightDesc,
				h.FlightCost,
				h.AccomDesc,
				h.AccomCost,
				h.CarDesc,
				h.CarCost,
				h.InsDesc,
				h.InsCost,
			}
			m.holiFormCur = 0
		}
	case "d", "x":
		if len(m.holidays) > 0 {
			m.finView = fViewDelete
		}
	case "esc", "left":
		m.finSection = fSectionMenu
	}
	return m, nil
}

func (m *model) updateHoliForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		if m.holiFormCur < holiFCount-1 {
			m.holiFormCur++
		} else {
			m.holiFormCur = 0
		}
	case "shift+tab", "up":
		if m.holiFormCur > 0 {
			m.holiFormCur--
		} else {
			m.holiFormCur = holiFCount - 1
		}
	case "left", "right":
		// No toggle fields anymore
	case " ":
		m.holiForm[m.holiFormCur] += key
	case "enter":
		h := storage.Holiday{
			Destination: strings.TrimSpace(m.holiForm[holiFDestination]),
			FlightDesc:  strings.TrimSpace(m.holiForm[holiFFlightDesc]),
			FlightCost:  strings.TrimSpace(m.holiForm[holiFFlightCost]),
			AccomDesc:   strings.TrimSpace(m.holiForm[holiFAccomDesc]),
			AccomCost:   strings.TrimSpace(m.holiForm[holiFAccomCost]),
			CarDesc:     strings.TrimSpace(m.holiForm[holiFCarDesc]),
			CarCost:     strings.TrimSpace(m.holiForm[holiFCarCost]),
			InsDesc:     strings.TrimSpace(m.holiForm[holiFInsDesc]),
			InsCost:     strings.TrimSpace(m.holiForm[holiFInsCost]),
		}
		if h.Destination == "" {
			m.finView = fViewList
			return m, nil
		}
		if m.finView == fViewAdd {
			m.holidays = append(m.holidays, h)
		} else {
			m.holidays[m.holiEditIdx] = h
		}
		_ = storage.SaveHolidays(m.dataDir, m.holidays)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		field := &m.holiForm[m.holiFormCur]
		if len(*field) > 0 {
			runes := []rune(*field)
			*field = string(runes[:len(runes)-1])
		}
	default:
		if key == "space" {
			key = " "
		}
		runes := []rune(key)
		if len(runes) == 1 {
			field := &m.holiForm[m.holiFormCur]
			isCostField := m.holiFormCur == holiFFlightCost || m.holiFormCur == holiFAccomCost || m.holiFormCur == holiFCarCost || m.holiFormCur == holiFInsCost
			if isCostField {
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
	return m, nil
}

func (m *model) updateHoliDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "s", "S":
		m.holidays = append(m.holidays[:m.holiCursor], m.holidays[m.holiCursor+1:]...)
		_ = storage.SaveHolidays(m.dataDir, m.holidays)
		if m.holiCursor >= len(m.holidays) && m.holiCursor > 0 {
			m.holiCursor--
		}
		m.finView = fViewList
	case "n", "esc":
		m.finView = fViewList
	}
	return m, nil
}

func (m *model) renderHolidays(s *styles) string {
	switch m.finView {
	case fViewAdd:
		return m.renderHoliForm(s, t(m.lang, "action.add")+" "+t(m.lang, "finances.holidays"))
	case fViewEdit:
		return m.renderHoliForm(s, t(m.lang, "action.edit")+" "+t(m.lang, "finances.holidays"))
	case fViewDelete:
		return m.renderHoliDelete(s)
	default:
		return m.renderHoliList(s)
	}
}

func (m *model) renderHoliList(s *styles) string {
	isActive := m.finSection != fSectionMenu
	title := s.title.Render(t(m.lang, "finances.holidays"))
	if len(m.holidays) == 0 {
		empty := s.dim.Render(t(m.lang, "holidays.noRecords"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	hdr := fmt.Sprintf("  %-3s %-20s %-12s %-12s %-12s",
		t(m.lang, "col.num"), t(m.lang, "col.destination"), t(m.lang, "col.flight"), t(m.lang, "col.accom"), t(m.lang, "col.totalCost"))
	header := s.subtitle.Render(hdr)
	divider := s.dim.Render("  " + strings.Repeat("─", 63))

	var rows []string
	for i, h := range m.holidays {
		fCost := parseEuro(h.FlightCost)
		aCost := parseEuro(h.AccomCost)
		cCost := parseEuro(h.CarCost)
		iCost := parseEuro(h.InsCost)
		totCost := fCost + aCost + cCost + iCost

		fStr := h.FlightCost
		if fStr == "" { fStr = "-" } else { fStr = "€ " + fStr }

		aStr := h.AccomCost
		if aStr == "" { aStr = "-" } else { aStr = "€ " + aStr }

		row := fmt.Sprintf("  %-3d %-20s %-12s %-12s € %.2f",
			i+1,
			truncate(h.Destination, 19),
			truncate(fStr, 11),
			truncate(aStr, 11),
			totCost,
		)
		if i == m.holiCursor {
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

func (m *model) renderHoliForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	labels := []string{
		t(m.lang, "field.destination"),
		t(m.lang, "field.flightDesc"),
		t(m.lang, "field.flightCost"),
		t(m.lang, "field.accomDesc"),
		t(m.lang, "field.accomCost"),
		t(m.lang, "field.carDesc"),
		t(m.lang, "field.carCost"),
		t(m.lang, "field.insDesc"),
		t(m.lang, "field.insCost"),
	}

	var fields []string
	for i := 0; i < holiFCount; i++ {
		label := s.dim.Render(fmt.Sprintf("  %-25s", labels[i]+":"))
		val := m.holiForm[i]
		
		isCostField := i == holiFFlightCost || i == holiFAccomCost || i == holiFCarCost || i == holiFInsCost

		var rendered string
		if i == m.holiFormCur {
			cursor := s.highlight.Render("_")
			fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
			
			if isCostField {
				if val == "" { val = s.dim.Render("0.00") }
				rendered = label + " € " + fieldStyle.Render(val) + cursor
			} else {
				rendered = label + " " + fieldStyle.Render(val) + cursor
			}
		} else {
			if isCostField {
				if val == "" { val = s.dim.Render("0.00") }
				rendered = label + " € " + s.info.Render(val)
			} else {
				rendered = label + " " + s.info.Render(val)
			}
		}
		fields = append(fields, rendered)
	}

	form := strings.Join(fields, "\n")
	help := s.dim.Render(fmt.Sprintf("\n\nTab/↑↓: %s  Enter: %s  Esc: %s",
		t(m.lang, "help.switchField"), t(m.lang, "action.save"), t(m.lang, "action.cancel")))

	return title + "\n\n" + form + help
}

func (m *model) renderHoliDelete(s *styles) string {
	if m.holiCursor >= len(m.holidays) {
		return m.renderHoliList(s)
	}
	h := m.holidays[m.holiCursor]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.holidays"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(t(m.lang, "delete.confirmInsurance"))

	fCost := parseEuro(h.FlightCost)
	aCost := parseEuro(h.AccomCost)
	cCost := parseEuro(h.CarCost)
	iCost := parseEuro(h.InsCost)
	totCost := fCost + aCost + cCost + iCost

	info := fmt.Sprintf(
		"\n  %s %s\n  %s € %.2f",
		s.dim.Render(t(m.lang, "field.destination")+":"), s.info.Render(h.Destination),
		s.dim.Render(t(m.lang, "col.totalCost")+":"), totCost,
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
	isActive := m.finSection != fSectionMenu
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
