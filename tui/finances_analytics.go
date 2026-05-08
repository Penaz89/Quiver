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
	filledStr := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(strings.Repeat("█", filled))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(strings.Repeat("░", empty))
	return filledStr + emptyStr
}

func (m *model) renderAnalytics(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.analytics"))

	barLen := 30
	labelW := 26

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
	sec1Div := s.dim.Render("  " + strings.Repeat("─", labelW+barLen+18))

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

	fmtRow := fmt.Sprintf("  %%-%ds  %%s  %%s\n", labelW)

	sec1 := sec1Title + "\n" + sec1Div + "\n\n"
	sec1 += fmt.Sprintf(fmtRow, t(m.lang, "analytics.avgIncome"), incomeBar, s.info.Render(fmt.Sprintf("€ %.2f", avgMonthlyNet)))
	sec1 += fmt.Sprintf(fmtRow, t(m.lang, "analytics.avgExpense"), expBar, s.info.Render(fmt.Sprintf("€ %.2f", monthlyExp)))
	sec1 += "  " + s.dim.Render(strings.Repeat("·", labelW+barLen+16)) + "\n"

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
		sec2Div := s.dim.Render("  " + strings.Repeat("─", labelW+barLen+18))

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
		sec2 += "  " + s.dim.Render(strings.Repeat("·", labelW+barLen+16)) + "\n"
		sec2 += fmt.Sprintf(fmtRow, t(m.lang, "finances.impactPct"), impBar, pctStyle.Render(fmt.Sprintf("%.1f%%", impactPct)))
	}

	help := s.dim.Render("\n\n←: " + t(m.lang, "help.goBack"))

	return title + "\n\n" + sec1 + sec2 + help
}
