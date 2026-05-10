package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

// ─── Daily Expenses logic ──────────────────────────────────────────────────

func (m *model) getDailyYears() []string {
	yearSet := make(map[string]bool)
	for _, d := range m.daily {
		yearSet[d.Date.Format("2006")] = true
	}
	var years []string
	for y := range yearSet {
		years = append(years, y)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))
	return years
}

func (m *model) getDailyMonthsForYear(year string) []string {
	monthSet := make(map[string]bool)
	for _, d := range m.daily {
		if d.Date.Format("2006") == year {
			monthSet[d.Date.Format("01")] = true
		}
	}
	var months []string
	for m := range monthSet {
		months = append(months, m)
	}
	sort.Strings(months)
	return months
}

func (m *model) getDailyForMonth(year, month string) []storage.DailyExpense {
	var filtered []storage.DailyExpense
	for _, d := range m.daily {
		if d.Date.Format("2006") == year && d.Date.Format("01") == month {
			filtered = append(filtered, d)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date.After(filtered[j].Date)
	})
	return filtered
}

func (m *model) getGlobalDailyIndex(target storage.DailyExpense) int {
	for i, d := range m.daily {
		if d.Date.Equal(target.Date) && d.Category == target.Category && d.Description == target.Description && d.Amount == target.Amount {
			return i
		}
	}
	return -1
}

func (m *model) updateDaily(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.finView {
	case fViewAdd, fViewEdit:
		return m.updateDailyForm(msg)
	case fViewDelete:
		return m.updateDailyDelete(msg)
	default:
		return m.updateDailyList(msg)
	}
}

func (m *model) updateDailyList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.dailyCursor > 0 {
			m.dailyCursor--
		}
	case "down", "j":
		max := 0
		if m.dailyYearFilter == "" {
			max = len(m.getDailyYears()) - 1
		} else if m.dailyMonthFilter == "" {
			max = len(m.getDailyMonthsForYear(m.dailyYearFilter)) - 1
		} else {
			max = len(m.getDailyForMonth(m.dailyYearFilter, m.dailyMonthFilter)) - 1
		}
		if m.dailyCursor < max {
			m.dailyCursor++
		}
	case "enter", "right":
		if m.dailyYearFilter == "" {
			years := m.getDailyYears()
			if len(years) > 0 && m.dailyCursor < len(years) {
				m.dailyYearFilter = years[m.dailyCursor]
				m.dailyCursor = 0
			}
		} else if m.dailyMonthFilter == "" {
			months := m.getDailyMonthsForYear(m.dailyYearFilter)
			if len(months) > 0 && m.dailyCursor < len(months) {
				m.dailyMonthFilter = months[m.dailyCursor]
				m.dailyCursor = 0
			}
		}
	case "a":
		m.finView = fViewAdd
		defaultCat := ""
		if len(m.categories) > 0 {
			defaultCat = m.categories[0]
		}
		m.dailyForm = [dailyFCount]string{"", defaultCat, "", ""}
		m.dailyFormCur = 0
	case "e":
		if m.dailyYearFilter != "" && m.dailyMonthFilter != "" {
			expenses := m.getDailyForMonth(m.dailyYearFilter, m.dailyMonthFilter)
			if len(expenses) > 0 && m.dailyCursor < len(expenses) {
				m.finView = fViewEdit
				m.dailyEditIdx = m.getGlobalDailyIndex(expenses[m.dailyCursor])
				d := m.daily[m.dailyEditIdx]
				dateStr := d.Date.Format("02/01/2006")
				if d.Date.IsZero() {
					dateStr = ""
				}
				m.dailyForm = [dailyFCount]string{
					dateStr,
					d.Category,
					d.Description,
					d.Amount,
				}
				m.dailyFormCur = 0
			}
		}
	case "d", "del", "x":
		if m.dailyYearFilter != "" && m.dailyMonthFilter != "" {
			expenses := m.getDailyForMonth(m.dailyYearFilter, m.dailyMonthFilter)
			if len(expenses) > 0 && m.dailyCursor < len(expenses) {
				m.finView = fViewDelete
				m.dailyEditIdx = m.getGlobalDailyIndex(expenses[m.dailyCursor])
			}
		}
	case "esc", "left":
		if m.dailyMonthFilter != "" {
			m.dailyMonthFilter = ""
			m.dailyCursor = 0
		} else if m.dailyYearFilter != "" {
			m.dailyYearFilter = ""
			m.dailyCursor = 0
		} else {
			m.finSection = fSectionMenu
		}
	}
	return m, nil
}

func (m *model) updateDailyForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		if m.dailyFormCur < dailyFCount-1 {
			m.dailyFormCur++
		} else {
			m.dailyFormCur = 0
		}
	case "shift+tab", "up":
		if m.dailyFormCur > 0 {
			m.dailyFormCur--
		} else {
			m.dailyFormCur = dailyFCount - 1
		}
	case "left", "right":
		if m.dailyFormCur == dailyFCategory && len(m.categories) > 0 {
			current := m.dailyForm[dailyFCategory]
			idx := -1
			for i, c := range m.categories {
				if c == current {
					idx = i
					break
				}
			}
			if key == "right" {
				idx++
				if idx >= len(m.categories) {
					idx = 0
				}
			} else {
				idx--
				if idx < 0 {
					idx = len(m.categories) - 1
				}
			}
			m.dailyForm[dailyFCategory] = m.categories[idx]
		}
	case "enter":
		dateStr := strings.TrimSpace(m.dailyForm[dailyFDate])
		var parsedDate time.Time
		if dateStr != "" {
			d, err := time.Parse("02/01/2006", dateStr)
			if err != nil {
				return m, nil
			}
			parsedDate = d
		} else {
			parsedDate = time.Now()
		}

		cat := strings.TrimSpace(m.dailyForm[dailyFCategory])
		desc := strings.TrimSpace(m.dailyForm[dailyFDescription])
		amt := strings.TrimSpace(m.dailyForm[dailyFAmount])

		if cat == "" || amt == "" {
			m.finView = fViewList
			return m, nil
		}

		d := storage.DailyExpense{
			Date:        parsedDate,
			Category:    cat,
			Description: desc,
			Amount:      amt,
			Author:      m.user,
		}

		if m.finView == fViewAdd {
			m.daily = append(m.daily, d)
		} else {
			m.daily[m.dailyEditIdx] = d
		}

		_ = storage.SaveDailyExpenses(m.dataDir, m.daily)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		if m.dailyFormCur != dailyFCategory {
			field := &m.dailyForm[m.dailyFormCur]
			if len(*field) > 0 {
				runes := []rune(*field)
				*field = string(runes[:len(runes)-1])
			}
		}
	case "space":
		if m.dailyFormCur == dailyFCategory && len(m.categories) > 0 {
			current := m.dailyForm[dailyFCategory]
			idx := -1
			for i, c := range m.categories {
				if c == current {
					idx = i
					break
				}
			}
			idx++
			if idx >= len(m.categories) {
				idx = 0
			}
			m.dailyForm[dailyFCategory] = m.categories[idx]
		} else if m.dailyFormCur != dailyFCategory {
			m.dailyForm[m.dailyFormCur] += " "
		}
	default:
		if m.dailyFormCur != dailyFCategory {
			runes := []rune(key)
			if len(runes) == 1 {
				field := &m.dailyForm[m.dailyFormCur]
				if m.dailyFormCur == dailyFDate {
					if !strings.ContainsRune("0123456789", runes[0]) {
						return m, nil
					}
					if len(*field) >= 10 {
						return m, nil
					}
					*field += key
					if len(*field) == 2 || len(*field) == 5 {
						*field += "/"
					}
				} else if m.dailyFormCur == dailyFAmount {
					if !strings.ContainsRune("0123456789.,", runes[0]) {
						return m, nil
					}
					*field += key
				} else {
					*field += key
				}
			}
		}
	}
	return m, nil
}

func (m *model) updateDailyDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "s", "S", "enter":
		if m.dailyEditIdx >= 0 && m.dailyEditIdx < len(m.daily) {
			m.daily = append(m.daily[:m.dailyEditIdx], m.daily[m.dailyEditIdx+1:]...)
			_ = storage.SaveDailyExpenses(m.dataDir, m.daily)
		}
		m.finView = fViewList
		m.dailyCursor = 0
	case "n", "N", "esc":
		m.finView = fViewList
	}
	return m, nil
}

func (m *model) renderDaily(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.daily"))
	
	switch m.finView {
	case fViewList:
		if m.dailyYearFilter == "" {
			return title + "\n\n" + m.renderDailyYearList(s)
		} else if m.dailyMonthFilter == "" {
			return title + "\n\n" + m.renderDailyMonthList(s)
		}
		return title + "\n\n" + m.renderDailyExpenseList(s)
	case fViewAdd, fViewEdit:
		return m.renderDailyForm(s, t(m.lang, "action.edit")+" "+t(m.lang, "finances.daily"))
	case fViewDelete:
		return m.renderDailyDelete(s)
	}
	return title
}

func (m *model) renderDailyYearList(s *styles) string {
	years := m.getDailyYears()
	if len(years) == 0 {
		return s.dim.Render(t(m.lang, "daily.noRecords")) + "\n\n" + s.dim.Render("a: "+t(m.lang, "action.add"))
	}

	var lines []string
	for i, y := range years {
		row := fmt.Sprintf("  %s", y)
		if i == m.dailyCursor {
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

	help := s.dim.Render(fmt.Sprintf("a: %s • enter: %s • ←: %s", t(m.lang, "action.add"), t(m.lang, "help.select"), t(m.lang, "help.goBack")))
	return strings.Join(lines, "\n") + "\n\n" + help
}

func (m *model) renderDailyMonthList(s *styles) string {
	months := m.getDailyMonthsForYear(m.dailyYearFilter)
	
	header := s.subtitle.Render("  " + t(m.lang, "col.year") + ": " + m.dailyYearFilter)
	divider := s.dim.Render("  " + strings.Repeat("─", 40))

	var lines []string
	for i, mth := range months {
		row := fmt.Sprintf("  %s - %s", mth, t(m.lang, "month."+mth))
		if i == m.dailyCursor {
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

	help := s.dim.Render(fmt.Sprintf("a: %s • enter: %s • ←: %s", t(m.lang, "action.add"), t(m.lang, "help.select"), t(m.lang, "help.goBack")))
	return header + "\n" + divider + "\n" + strings.Join(lines, "\n") + "\n\n" + help
}

func (m *model) renderDailyExpenseList(s *styles) string {
	expenses := m.getDailyForMonth(m.dailyYearFilter, m.dailyMonthFilter)
	
	headerStr := fmt.Sprintf("  %-12s %-15s %-20s %-10s %s", t(m.lang, "col.date"), t(m.lang, "col.category"), t(m.lang, "col.description"), t(m.lang, "col.amount"), "AUTORE")
	header := s.subtitle.Render(headerStr)
	divider := s.dim.Render("  " + strings.Repeat("─", 80))
	
	var lines []string
	var totalAmount float64

	for i, d := range expenses {
		dateStr := d.Date.Format("02/01/2006")
		amtStr := "€ " + d.Amount
		totalAmount += parseEuro(d.Amount)
		
		authorStr := ""
		if d.Author != "" {
			authorStr = s.dim.Render("[" + d.Author + "]")
		}
		
		row := fmt.Sprintf("  %-12s %-15s %-20s %-10s %s",
			dateStr,
			truncate(d.Category, 14),
			truncate(d.Description, 19),
			amtStr,
			authorStr,
		)

		if i == m.dailyCursor {
			isActive := m.finSection != fSectionMenu && m.focusContent
			if isActive {
				lines = append(lines, s.menuSelected.Width(0).Render(row))
			} else {
				lines = append(lines, s.menuActiveDim.Width(0).Render(row))
			}
		} else {
			lines = append(lines, s.info.Render(row))
		}
	}

	sumDivider := s.dim.Render("  " + strings.Repeat("=", 80))
	sumTitle := s.info.Render("  " + t(m.lang, "finances.monthlyTotal") + " " + m.dailyMonthFilter + "/" + m.dailyYearFilter)
	sumRow := fmt.Sprintf("  %-12s %-15s %-20s € %.2f", "TOT", "", "", totalAmount)
	summaryBlock := sumTitle + "\n" + s.highlight.Render(sumRow)

	help := s.dim.Render(fmt.Sprintf("a: %s  e: %s  d: %s  ←: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))

	return header + "\n" + divider + "\n" + strings.Join(lines, "\n") + "\n" + sumDivider + "\n" + summaryBlock + "\n\n" + help
}

func (m *model) renderDailyForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	labels := []string{
		t(m.lang, "field.date"),
		t(m.lang, "field.category"),
		t(m.lang, "field.description"),
		t(m.lang, "field.amount"),
	}

	var fields []string
	for i := 0; i < dailyFCount; i++ {
		label := s.dim.Render(fmt.Sprintf("  %-25s", labels[i]+":"))
		val := m.dailyForm[i]

		var rendered string
		if i == m.dailyFormCur {
			cursor := s.highlight.Render("_")
			fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
			if i == dailyFCategory {
				rendered = label + " < " + fieldStyle.Render(val) + " >"
			} else {
				rendered = label + " " + fieldStyle.Render(val) + cursor
			}
		} else {
			rendered = label + " " + s.info.Render(val)
		}
		fields = append(fields, rendered)
	}

	form := strings.Join(fields, "\n\n")
	help := s.dim.Render(fmt.Sprintf("\n\nTab/↑↓: %s  Enter: %s  Esc: %s",
		t(m.lang, "help.switchField"), t(m.lang, "action.save"), t(m.lang, "action.cancel")))

	return title + "\n\n" + form + help
}

func (m *model) renderDailyDelete(s *styles) string {
	if m.dailyEditIdx < 0 || m.dailyEditIdx >= len(m.daily) {
		return m.renderDailyExpenseList(s)
	}
	d := m.daily[m.dailyEditIdx]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.daily"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(t(m.lang, "delete.confirmInsurance"))

	info := fmt.Sprintf(
		"\n  %s %s\n  %s %s\n  %s € %s",
		s.dim.Render(t(m.lang, "col.date")+":"), s.info.Render(d.Date.Format("02/01/2006")),
		s.dim.Render(t(m.lang, "col.category")+":"), s.info.Render(d.Category),
		s.dim.Render(t(m.lang, "col.amount")+":"), s.info.Render(d.Amount),
	)

	help := s.dim.Render(fmt.Sprintf("\n\ny/s: %s  n/Esc: %s",
		t(m.lang, "action.delete"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + "\n" + info + help
}
