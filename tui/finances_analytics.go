package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) updateAnalytics(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "left":
		m.finSection = fSectionMenu
	case "up", "k":
		m.vp.ScrollUp(1)
	case "down", "j":
		m.vp.ScrollDown(1)
	}
	return m, nil
}

// renderBar draws a bar of barLen characters with `ratio` filled.
func renderBar(ratio float64, barLen int, color string) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1.0 {
		ratio = 1.0
	}
	filled := int(ratio * float64(barLen))
	empty := barLen - filled
	if empty < 0 {
		empty = 0
	}
	filledStr := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(strings.Repeat("━", filled))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(strings.Repeat("─", empty))
	return filledStr + emptyStr
}

func (m *model) renderAnalytics(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.analytics"))

	// ── Dynamically compute barLen & labelW based on available width ──
	sw := sidebarWidth(m.width)
	contentW := m.width - sw - 6 // content panel (borders + padding)
	if sw == 0 {
		contentW = m.width - 4
	}
	vpWidth := contentW - 4

	// col2 width is the max of title, subtitle, and submenu strings
	titleStr := s.title.Render(t(m.lang, "finances.title"))
	descStr := s.subtitle.Render(t(m.lang, "finances.subtitle"))
	col2W := lipgloss.Width(titleStr)
	if lipgloss.Width(descStr) > col2W {
		col2W = lipgloss.Width(descStr)
	}
	submenuW := sw - 4
	if submenuW < 10 {
		submenuW = 10
	}
	if submenuW > col2W {
		col2W = submenuW
	}

	// Subtract col2 width + paddingRight(2) + marginRight(2) + right border(1) + safety margin(3)
	col3W := vpWidth - col2W - 8
	if col3W < 30 {
		col3W = 30
	}

	maxLabel := 0
	labelsList := []string{
		t(m.lang, "analytics.avgIncome"),
		t(m.lang, "analytics.avgExpense"),
		t(m.lang, "analytics.cashFlow"),
		t(m.lang, "analytics.savingRate"),
		t(m.lang, "finances.projectedAnnual"),
		t(m.lang, "finances.fixedAnnual"),
		t(m.lang, "finances.impactPct"),
	}
	for _, l := range labelsList {
		w := lipgloss.Width(l)
		if w > maxLabel {
			maxLabel = w
		}
	}
	maxLabel += 2 // Padding

	isCompact := false
	if col3W < maxLabel+30 {
		isCompact = true
	}

	var labelW, barLen, divLen int
	var fmtRow string

	if isCompact {
		labelW = col3W - 4
		barLen = col3W - 25
		if barLen < 8 {
			barLen = 8
		}
		if barLen > 25 {
			barLen = 25
		}
		divLen = col3W - 4
		if divLen < 1 {
			divLen = 1
		}
		fmtRow = "  %s\n  %s  %s\n"
	} else {
		labelW = maxLabel
		barLen = col3W - labelW - 35
		if barLen < 8 {
			barLen = 8
		}
		if barLen > 25 {
			barLen = 25
		}
		divLen = labelW + barLen + 16
		if divLen < 1 {
			divLen = 1
		}
		fmtRow = fmt.Sprintf("  %%-%ds  %%s  %%s\n", labelW)
	}

	// ── Compute data ────────────────────────────────────────
	annualExp, monthlyExp := m.calculateTotalFinances()

	var totalNet float64
	var count int
	for _, sal := range m.salaries {
		if sal.Net != "" {
			totalNet += parseEuro(sal.Net)
			count++
		}
	}
	avgMonthlyNet := 0.0
	if count > 0 {
		avgMonthlyNet = totalNet / float64(count)
	}
	cashFlow := avgMonthlyNet - monthlyExp
	savingRate := 0.0
	if avgMonthlyNet > 0 {
		savingRate = (cashFlow / avgMonthlyNet) * 100.0
	}

	// ── Section 1: Monthly Balance ──────────────────────────
	sec1Title := s.subtitle.Render("  " + t(m.lang, "analytics.ratioBar"))
	sec1Div := s.dim.Render("  " + strings.Repeat("─", divLen))

	incomeBar := renderBar(1.0, barLen, "39")  // cyan = income reference
	expRatio := 0.0
	if avgMonthlyNet > 0 {
		expRatio = monthlyExp / avgMonthlyNet
	}
	expBar := renderBar(expRatio, barLen, "196") // red = expenses

	saveRatio := savingRate / 100.0
	saveColor := "42" // green
	if savingRate < 10 {
		saveColor = "196"
	} else if savingRate < 25 {
		saveColor = "214"
	}
	saveBar := renderBar(saveRatio, barLen, saveColor)

	sec1 := sec1Title + "\n" + sec1Div + "\n\n"
	sec1 += fmt.Sprintf(fmtRow, t(m.lang, "analytics.avgIncome"), incomeBar, s.info.Render(fmt.Sprintf("€ %.2f", avgMonthlyNet)))
	sec1 += fmt.Sprintf(fmtRow, t(m.lang, "analytics.avgExpense"), expBar, s.info.Render(fmt.Sprintf("€ %.2f", monthlyExp)))
	sec1 += "  " + s.dim.Render(strings.Repeat("·", divLen)) + "\n"

	// Cash flow value colored
	cfColor := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	if cashFlow < 0 {
		cfColor = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	}
	cfValStr := cfColor.Render(fmt.Sprintf("€ %.2f", cashFlow))
	sec1 += fmt.Sprintf(fmtRow, t(m.lang, "analytics.cashFlow"), saveBar, cfValStr)
	sec1 += fmt.Sprintf(fmtRow, t(m.lang, "analytics.savingRate"), s.dim.Render(strings.Repeat(" ", barLen)), s.info.Render(fmt.Sprintf("%.1f%%", savingRate)))

	// ── Section 2: Salary Impact (annual) ───────────────────
	var sec2 string

	currentYearStr := fmt.Sprintf("%d", time.Now().Year())
	var yearNet float64
	var monthsCount int
	for _, sal := range m.salaries {
		if sal.Year == currentYearStr {
			yearNet += parseEuro(sal.Net)
			monthsCount++
		}
	}

	if monthsCount > 0 {
		avgNet := yearNet / float64(monthsCount)
		projectedAnnual := avgNet * 12.0
		impactPct := 0.0
		if projectedAnnual > 0 {
			impactPct = (annualExp / projectedAnnual) * 100.0
		}

		sec2Title := s.subtitle.Render("  " + t(m.lang, "finances.salaryImpact"))
		sec2Div := s.dim.Render("  " + strings.Repeat("─", divLen))

		netBar := renderBar(1.0, barLen, "75") // light blue = net reference
		fixRatio := 0.0
		if projectedAnnual > 0 {
			fixRatio = annualExp / projectedAnnual
		}
		fixColor := "196" // red = fixed expenses
		fixBar := renderBar(fixRatio, barLen, fixColor)
		impBar := renderBar(impactPct/100.0, barLen, "208") // orange = impact

		// Impact percentage colored
		pctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)

		sec2 = "\n\n" + sec2Title + "\n" + sec2Div + "\n\n"
		sec2 += fmt.Sprintf(fmtRow, t(m.lang, "finances.projectedAnnual"), netBar, s.info.Render(fmt.Sprintf("€ %.2f", projectedAnnual)))
		sec2 += fmt.Sprintf(fmtRow, t(m.lang, "finances.fixedAnnual"), fixBar, s.info.Render(fmt.Sprintf("€ %.2f", annualExp)))
		sec2 += "  " + s.dim.Render(strings.Repeat("·", divLen)) + "\n"
		sec2 += fmt.Sprintf(fmtRow, t(m.lang, "finances.impactPct"), impBar, pctStyle.Render(fmt.Sprintf("%.1f%%", impactPct)))
	}

	// ── Section 3: Monthly Daily Expenses (Current Year) ────────
	var sec3 string
	monthlyDailyExp := make([]float64, 12)
	maxMonthlyExp := 0.0
	hasDailyExp := false
	
	for _, exp := range m.daily {
		if !exp.Date.IsZero() {
			if fmt.Sprintf("%d", exp.Date.Year()) == currentYearStr {
				monthIdx := int(exp.Date.Month()) - 1
				monthlyDailyExp[monthIdx] += parseEuro(exp.Amount)
				hasDailyExp = true
			}
		}
	}
	
	for _, amt := range monthlyDailyExp {
		if amt > maxMonthlyExp {
			maxMonthlyExp = amt
		}
	}
	
	if hasDailyExp {
		sec3Title := s.subtitle.Render("  " + t(m.lang, "finances.daily") + " (" + currentYearStr + ")")
		sec3Div := s.dim.Render("  " + strings.Repeat("─", divLen))
		sec3 = "\n\n" + sec3Title + "\n" + sec3Div + "\n\n"
		
		for i := 0; i < 12; i++ {
			amt := monthlyDailyExp[i]
			ratio := 0.0
			if maxMonthlyExp > 0 {
				ratio = amt / maxMonthlyExp
			}
			bar := renderBar(ratio, barLen, "135") // Purple for daily expenses
			monthKey := fmt.Sprintf("month.%02d", i+1)
			monthName := t(m.lang, monthKey)
			
			sec3 += fmt.Sprintf(fmtRow, monthName, bar, s.info.Render(fmt.Sprintf("€ %.2f", amt)))
		}
		
		totalDailyExp := 0.0
		for _, amt := range monthlyDailyExp {
			totalDailyExp += amt
		}
		
		sec3 += "  " + s.dim.Render(strings.Repeat("·", divLen)) + "\n"
		totalStr := s.highlight.Render(fmt.Sprintf("€ %.2f", totalDailyExp))
		sec3 += fmt.Sprintf(fmtRow, t(m.lang, "daily.annualTotal"), s.dim.Render(strings.Repeat(" ", barLen)), totalStr)
	}

	// ── Section 4: Budget per Categoria (Current Month) ────────
	var sec4 string
	currentMonth := time.Now().Month()
	
	budgetedCategories := make(map[string]float64)
	
	for cat, bStr := range m.budgets {
		if bStr != "" {
			b := parseEuro(bStr)
			if b > 0 {
				budgetedCategories[cat] = 0 // initialize
			}
		}
	}
	
	if len(budgetedCategories) > 0 {
		for _, exp := range m.daily {
			if !exp.Date.IsZero() {
				if fmt.Sprintf("%d", exp.Date.Year()) == currentYearStr && exp.Date.Month() == currentMonth {
					if _, ok := budgetedCategories[exp.Category]; ok {
						budgetedCategories[exp.Category] += parseEuro(exp.Amount)
					}
				}
			}
		}
		
		sec4Title := s.subtitle.Render("  " + t(m.lang, "analytics.budgets") + " (" + t(m.lang, fmt.Sprintf("month.%02d", currentMonth)) + ")")
		sec4Div := s.dim.Render("  " + strings.Repeat("─", divLen))
		sec4 = "\n\n" + sec4Title + "\n" + sec4Div + "\n\n"
		
		for cat, spent := range budgetedCategories {
			budget := parseEuro(m.budgets[cat])
			ratio := 0.0
			if budget > 0 {
				ratio = spent / budget
			}
			
			barColor := "42" // Green
			if ratio >= 1.0 {
				barColor = "196" // Red (over budget)
			} else if ratio >= 0.8 {
				barColor = "214" // Orange (close to budget)
			}
			
			bar := renderBar(ratio, barLen, barColor)
			
			catLabel := truncate(cat, labelW-1)
			valStr := fmt.Sprintf("€ %.2f / € %.2f", spent, budget)
			if ratio >= 1.0 {
				valStr = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(valStr)
			} else {
				valStr = s.info.Render(valStr)
			}
			
			sec4 += fmt.Sprintf(fmtRow, catLabel, bar, valStr)
		}
	}

	help := s.dim.Render("\n\n←: " + t(m.lang, "help.goBack"))

	return title + "\n\n" + sec1 + sec2 + sec3 + sec4 + help
}
