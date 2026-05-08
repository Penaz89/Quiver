package tui

import (
	"fmt"
	"strings"

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

func (m *model) renderAnalytics(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.analytics"))

	_, monthlyExp := m.calculateTotalFinances()
	
	var totalNet float64
	var count int
	for _, sal := range m.salaries {
		if sal.Net != "" {
			netVal := parseEuro(sal.Net)
			totalNet += netVal
			count++
		}
	}
	
	avgMonthlyNet := 0.0
	if count > 0 {
		avgMonthlyNet = totalNet / float64(count)
	}

	savingRate := 0.0
	if avgMonthlyNet > 0 {
		savingRate = ((avgMonthlyNet - monthlyExp) / avgMonthlyNet) * 100.0
	}

	barLen := 40
	expRatio := 0.0
	if avgMonthlyNet > 0 {
		expRatio = monthlyExp / avgMonthlyNet
	}
	if expRatio > 1.0 {
		expRatio = 1.0
	}

	filled := int(expRatio * float64(barLen))
	empty := barLen - filled
	if empty < 0 {
		empty = 0
	}

	barStr := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	color := "42" // green
	if expRatio > 0.8 {
		color = "196" // red
	} else if expRatio > 0.5 {
		color = "214" // orange
	}
	coloredBar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(barStr)

	content := fmt.Sprintf("%-25s € %.2f\n", t(m.lang, "analytics.avgIncome"), avgMonthlyNet)
	content += fmt.Sprintf("%-25s € %.2f\n", t(m.lang, "analytics.avgExpense"), monthlyExp)
	content += fmt.Sprintf("%-25s € %.2f\n\n", t(m.lang, "analytics.cashFlow"), avgMonthlyNet-monthlyExp)
	
	content += fmt.Sprintf("%-25s %s\n", t(m.lang, "analytics.savingRate"), fmt.Sprintf("%.1f%%", savingRate))
	content += fmt.Sprintf("\n%s\n[%s]", t(m.lang, "analytics.ratioBar"), coloredBar)

	return title + "\n\n" + content + "\n\n" + s.dim.Render("esc: back")
}
