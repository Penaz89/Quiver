package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func nextYear(current string) string {
	y, err := strconv.Atoi(current)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().Year())
	}
	return fmt.Sprintf("%d", y+1)
}

func prevYear(current string) string {
	y, err := strconv.Atoi(current)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().Year())
	}
	return fmt.Sprintf("%d", y-1)
}

func nextMonth(current string) string {
	months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}
	for i, m := range months {
		if m == current {
			return months[(i+1)%12]
		}
	}
	return "01"
}

func prevMonth(current string) string {
	months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}
	for i, m := range months {
		if m == current {
			return months[(i-1+12)%12]
		}
	}
	return "12"
}

func (m *model) updateSalaries(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch m.finView {
	case fViewList:
		switch key {
		case "esc", "left":
			if m.salaryYearFilter != "" {
				m.salaryYearFilter = "" // go back to years list
				m.salaryCursor = 0
			} else {
				m.finSection = fSectionMenu
			}
		case "up", "k":
			if m.salaryCursor > 0 {
				m.salaryCursor--
			}
		case "down", "j":
			max := 0
			if m.salaryYearFilter == "" {
				max = len(m.getSalaryYears()) - 1
			} else {
				max = len(m.getSalariesForYear(m.salaryYearFilter)) - 1
			}
			if m.salaryCursor < max {
				m.salaryCursor++
			}
		case "enter", "right":
			if m.salaryYearFilter == "" {
				years := m.getSalaryYears()
				if len(years) > 0 && m.salaryCursor < len(years) {
					m.salaryYearFilter = years[m.salaryCursor]
					m.salaryCursor = 0
				}
			}
		case "n", "a":
			m.finView = fViewAdd
			m.salaryFormCur = 0
			m.salaryForm = [4]string{fmt.Sprintf("%d", time.Now().Year()), "01", "", ""}
		case "e":
			if m.salaryYearFilter != "" {
				salaries := m.getSalariesForYear(m.salaryYearFilter)
				if len(salaries) > 0 && m.salaryCursor < len(salaries) {
					m.finView = fViewEdit
					m.salaryEditIdx = m.getGlobalSalaryIndex(salaries[m.salaryCursor])
					sal := m.salaries[m.salaryEditIdx]
					m.salaryFormCur = 0
					m.salaryForm = [4]string{sal.Year, sal.Month, sal.Gross, sal.Net}
				}
			}
		case "d", "del":
			if m.salaryYearFilter != "" {
				salaries := m.getSalariesForYear(m.salaryYearFilter)
				if len(salaries) > 0 && m.salaryCursor < len(salaries) {
					m.finView = fViewDelete
					m.salaryEditIdx = m.getGlobalSalaryIndex(salaries[m.salaryCursor])
				}
			}
		}
	case fViewAdd, fViewEdit:
		switch key {
		case "esc":
			m.finView = fViewList
		case "tab", "down":
			m.salaryFormCur = (m.salaryFormCur + 1) % 4
		case "shift+tab", "up":
			m.salaryFormCur = (m.salaryFormCur - 1 + 4) % 4
		case "enter":
			if m.salaryForm[0] != "" && m.salaryForm[1] != "" && m.salaryForm[2] != "" && m.salaryForm[3] != "" {
				newSal := storage.Salary{
					Year:   m.salaryForm[0],
					Month:  m.salaryForm[1],
					Gross:  m.salaryForm[2],
					Net:    m.salaryForm[3],
					Author: m.user,
				}
				if m.finView == fViewAdd {
					m.salaries = append(m.salaries, newSal)
				} else {
					m.salaries[m.salaryEditIdx] = newSal
				}
				storage.SaveSalaries(m.dataDir, m.salaries)
				m.finView = fViewList
			}
		case "left", "h":
			if m.salaryFormCur == 0 {
				m.salaryForm[0] = prevYear(m.salaryForm[0])
			} else if m.salaryFormCur == 1 {
				m.salaryForm[1] = prevMonth(m.salaryForm[1])
			}
		case "right", "l":
			if m.salaryFormCur == 0 {
				m.salaryForm[0] = nextYear(m.salaryForm[0])
			} else if m.salaryFormCur == 1 {
				m.salaryForm[1] = nextMonth(m.salaryForm[1])
			}
		case "backspace":
			if m.salaryFormCur != 0 && m.salaryFormCur != 1 && len(m.salaryForm[m.salaryFormCur]) > 0 {
				s := m.salaryForm[m.salaryFormCur]
				m.salaryForm[m.salaryFormCur] = s[:len(s)-1]
			}
		default:
			if m.salaryFormCur != 0 && m.salaryFormCur != 1 {
				if key == "space" {
					key = " "
				}
				if len(key) == 1 {
					m.salaryForm[m.salaryFormCur] += key
				}
			}
		}
	case fViewDelete:
		switch key {
		case "y", "Y", "enter":
			if m.salaryEditIdx >= 0 && m.salaryEditIdx < len(m.salaries) {
				m.salaries = append(m.salaries[:m.salaryEditIdx], m.salaries[m.salaryEditIdx+1:]...)
				storage.SaveSalaries(m.dataDir, m.salaries)
			}
			m.finView = fViewList
			m.salaryCursor = 0
		case "n", "N", "esc":
			m.finView = fViewList
		}
	}
	return m, nil
}

func (m *model) getSalaryYears() []string {
	yearSet := make(map[string]bool)
	for _, s := range m.salaries {
		yearSet[s.Year] = true
	}
	var years []string
	for y := range yearSet {
		years = append(years, y)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))
	return years
}

func (m *model) getSalariesForYear(year string) []storage.Salary {
	var filtered []storage.Salary
	for _, s := range m.salaries {
		if s.Year == year {
			filtered = append(filtered, s)
		}
	}
	// Sort by month (assuming month is 01, 02... or similar, string sort works for numeric months)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Month < filtered[j].Month
	})
	return filtered
}

func (m *model) getGlobalSalaryIndex(target storage.Salary) int {
	for i, s := range m.salaries {
		if s.Year == target.Year && s.Month == target.Month && s.Gross == target.Gross && s.Net == target.Net {
			return i
		}
	}
	return -1
}

func (m *model) renderSalaries(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.salaries"))
	
	switch m.finView {
	case fViewList:
		if m.salaryYearFilter == "" {
			return title + "\n\n" + m.renderSalariesYearList(s)
		}
		return title + "\n\n" + m.renderSalariesMonthList(s)
	case fViewAdd, fViewEdit:
		return title + "\n\n" + m.renderSalariesForm(s)
	case fViewDelete:
		return title + "\n\n" + m.renderSalariesDelete(s)
	}
	return title
}

func (m *model) renderSalariesYearList(s *styles) string {
	years := m.getSalaryYears()
	if len(years) == 0 {
		return s.dim.Render(t(m.lang, "salaries.noRecords")) + "\n\n" + s.dim.Render("n/a: "+t(m.lang, "action.add"))
	}

	var lines []string
	for i, y := range years {
		row := fmt.Sprintf("  %s", y)
		if i == m.salaryCursor {
			isActive := m.finSection != fSectionMenu && m.focusContent
			if isActive {
				lines = append(lines, s.menuSelected.Render(row))
			} else {
				lines = append(lines, s.menuActiveDim.Render(row))
			}
		} else {
			lines = append(lines, s.menuNormal.Render(row))
		}
	}

	help := s.dim.Render(fmt.Sprintf("n/a: %s • enter: %s • ←: %s", t(m.lang, "action.add"), t(m.lang, "help.select"), t(m.lang, "help.goBack")))
	
	stats := m.renderSalariesStats(s)
	if stats != "" {
		divider := s.dim.Render("  " + strings.Repeat("─", 40))
		return strings.Join(lines, "\n") + "\n\n" + divider + "\n\n" + stats + "\n\n" + help
	}
	
	return strings.Join(lines, "\n") + "\n\n" + help
}

func (m *model) renderSalariesStats(s *styles) string {
	currentYearStr := fmt.Sprintf("%d", time.Now().Year())
	prevYearStr := fmt.Sprintf("%d", time.Now().Year()-1)

	var currentGross, currentNet float64
	var prevGross, prevNet float64

	for _, sal := range m.salaries {
		if sal.Year == currentYearStr {
			currentGross += parseEuro(sal.Gross)
			currentNet += parseEuro(sal.Net)
		} else if sal.Year == prevYearStr {
			prevGross += parseEuro(sal.Gross)
			prevNet += parseEuro(sal.Net)
		}
	}

	if currentGross == 0 && prevGross == 0 {
		return ""
	}

	maxGross := currentGross
	if prevGross > maxGross {
		maxGross = prevGross
	}

	barWidth := 20

	renderYearStats := func(year string, gross, net float64) string {
		if gross == 0 {
			return ""
		}
		grossLen := 0
		netLen := 0
		if maxGross > 0 {
			grossLen = int((gross / maxGross) * float64(barWidth))
			netLen = int((net / maxGross) * float64(barWidth))
		}
		
		taxes := gross - net
		taxPct := 0.0
		if gross > 0 {
			taxPct = (taxes / gross) * 100
		}
		taxLen := int((taxPct / 100.0) * float64(barWidth))

		grossBar := strings.Repeat("━", grossLen)
		grossEmpty := strings.Repeat("─", barWidth-grossLen)
		netBar := strings.Repeat("━", netLen)
		netEmpty := strings.Repeat("─", barWidth-netLen)
		taxBar := strings.Repeat("━", taxLen)
		taxEmpty := strings.Repeat("─", barWidth-taxLen)

		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

		grossColor := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render(grossBar) + emptyStyle.Render(grossEmpty)
		netColor := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render(netBar) + emptyStyle.Render(netEmpty)
		taxColor := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(taxBar) + emptyStyle.Render(taxEmpty)
		
		lblGross := t(m.lang, "col.gross")
		lblNet := t(m.lang, "col.net")
		lblTaxes := t(m.lang, "col.taxes")

		res := fmt.Sprintf("  %s\n", s.info.Render(year))
		res += fmt.Sprintf("    %-8s %s  € %.2f\n", lblGross, grossColor, gross)
		res += fmt.Sprintf("    %-8s %s  € %.2f\n", lblNet, netColor, net)
		res += fmt.Sprintf("    %-8s %s  %.1f%%\n", lblTaxes, taxColor, taxPct)
		return res
	}

	title := s.subtitle.Render("  " + t(m.lang, "salaries.comparison"))

	res := title + "\n\n"
	
	curStats := renderYearStats(currentYearStr, currentGross, currentNet)
	if curStats != "" {
		res += curStats + "\n"
	}
	
	prevStats := renderYearStats(prevYearStr, prevGross, prevNet)
	if prevStats != "" {
		res += prevStats
	}

	return res
}

func (m *model) renderSalariesMonthList(s *styles) string {
	salaries := m.getSalariesForYear(m.salaryYearFilter)
	
	var totalGross, totalNet float64
	
	headerStr := fmt.Sprintf("  %-13s %-13s %-13s %-13s %-8s %s", t(m.lang, "col.month"), t(m.lang, "col.gross"), t(m.lang, "col.net"), t(m.lang, "col.deductions"), t(m.lang, "col.taxes"), "AUTORE")
	header := s.subtitle.Render(headerStr)
	divider := s.dim.Render("  " + strings.Repeat("─", 80))
	
	var lines []string
	for i, sal := range salaries {
		gross := parseEuro(sal.Gross)
		net := parseEuro(sal.Net)
		totalGross += gross
		totalNet += net
		
		taxes := gross - net
		taxPct := 0.0
		if gross > 0 {
			taxPct = (taxes / gross) * 100
		}
		
		authorStr := ""
		if sal.Author != "" {
			authorStr = s.dim.Render("[" + sal.Author + "]")
		}
		
		monthLabel := fmt.Sprintf("%s - %s", sal.Month, truncate(t(m.lang, "month."+sal.Month), 6))
		row := fmt.Sprintf("  %-13s € %-11.2f € %-11.2f € %-11.2f %-8.1f%% %s", monthLabel, gross, net, taxes, taxPct, authorStr)
		if i == m.salaryCursor {
			isActive := m.finSection != fSectionMenu && m.focusContent
			if isActive {
				lines = append(lines, s.menuSelected.Render(row))
			} else {
				lines = append(lines, s.menuActiveDim.Render(row))
			}
		} else {
			lines = append(lines, s.menuNormal.Render(row))
		}
	}
	
	// Annual summary
	sumDivider := s.dim.Render("  " + strings.Repeat("=", 80))
	
	totalTaxes := totalGross - totalNet
	totalTaxPct := 0.0
	if totalGross > 0 {
		totalTaxPct = (totalTaxes / totalGross) * 100
	}
	
	sumTitle := s.info.Render("  " + t(m.lang, "salaries.annualSum") + " " + m.salaryYearFilter)
	sumRow := fmt.Sprintf("  %-13s € %-11.2f € %-11.2f € %-11.2f %.1f%%", "TOT", totalGross, totalNet, totalTaxes, totalTaxPct)
	
	summaryBlock := sumTitle + "\n" + sumRow
	
	help := s.dim.Render(fmt.Sprintf("e: %s • d/del: %s • ←: %s", t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))
	
	return header + "\n" + divider + "\n" + strings.Join(lines, "\n") + "\n" + sumDivider + "\n" + summaryBlock + "\n\n" + help
}

func (m *model) renderSalariesForm(s *styles) string {
	fields := []string{
		t(m.lang, "field.year"),
		t(m.lang, "field.month"),
		t(m.lang, "field.gross"),
		t(m.lang, "field.net"),
	}

	var lines []string
	for i, f := range fields {
		val := m.salaryForm[i]
		if i == m.salaryFormCur {
			if i == 0 || i == 1 {
				val = s.fieldBg.Render("< " + val + " >")
			} else {
				val = s.fieldBg.Render(val + s.highlight.Render("_"))
			}
		}
		lines = append(lines, fmt.Sprintf("%-15s %s", f+":", val))
	}

	helpStr := fmt.Sprintf("enter: %s • esc: %s", t(m.lang, "action.save"), t(m.lang, "action.cancel"))
	if m.salaryFormCur == 0 || m.salaryFormCur == 1 {
		helpStr += " • ←/→: select"
	}
	help := s.dim.Render(helpStr)
	return strings.Join(lines, "\n\n") + "\n\n" + help
}

func (m *model) renderSalariesDelete(s *styles) string {
	prompt := s.status.Render(t(m.lang, "delete.confirmInsurance")) // Reuse string
	help := s.dim.Render("y: " + t(m.lang, "action.confirm") + " • n/esc: " + t(m.lang, "action.cancel"))
	return prompt + "\n\n" + help
}
