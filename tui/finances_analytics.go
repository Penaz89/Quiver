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
		m.vp.LineUp(1)
	case "down", "j":
		m.vp.LineDown(1)
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
		barLen = col3W - 18
		if barLen < 8 {
			barLen = 8
		}
		if barLen > 40 {
			barLen = 40
		}
		divLen = col3W - 4
		if divLen < 1 {
			divLen = 1
		}
		fmtRow = "  %s\n  %s  %s\n"
	} else {
		labelW = maxLabel
		barLen = col3W - labelW - 18
		if barLen < 8 {
			barLen = 8
		}
		if barLen > 40 {
			barLen = 40
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

	help := s.dim.Render("\n\n←: " + t(m.lang, "help.goBack"))

	return title + "\n\n" + sec1 + sec2 + help
}
