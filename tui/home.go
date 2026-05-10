package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

func (m *model) calculateTotalFinances() (float64, float64) {
	var annual, monthly float64

	// Vehicles
	for _, v := range m.vehicles {
		var a float64
		a += parseEuro(v.RoadTaxCost)
		a += parseEuro(v.NTCCost) / 2.0
		a += parseEuro(v.ServiceCost)
		for _, ins := range m.insurances {
			if ins.LicensePlate == v.LicensePlate {
				cost := parseEuro(ins.TotalCost)
				if ins.Type == "type.semiannual" {
					a += cost * 2.0
				} else {
					a += cost
				}
			}
		}
		annual += a
		monthly += a / 12.0
	}

	// Housing
	for _, h := range m.housing {
		cost := parseEuro(h.Cost)
		if h.Type == "type.monthly" {
			monthly += cost
			annual += cost * 12.0
		} else {
			annual += cost
			monthly += cost / 12.0
		}
	}

	// Subs
	for _, sub := range m.subs {
		cost := parseEuro(sub.Cost)
		if sub.Type == "type.monthly" {
			monthly += cost
			annual += cost * 12.0
		} else {
			annual += cost
			monthly += cost / 12.0
		}
	}

	// Installments
	for _, inst := range m.installments {
		if inst.TotalCount > 0 && inst.PaidCount >= inst.TotalCount {
			continue // Paid off
		}
		cost := parseEuro(inst.Amount)
		switch inst.Frequency {
		case "type.monthly":
			monthly += cost
			annual += cost * 12.0
		case "type.bimonthly":
			monthly += cost / 2.0
			annual += cost * 6.0
		case "type.quarterly":
			monthly += cost / 3.0
			annual += cost * 4.0
		case "type.semiannual":
			monthly += cost / 6.0
			annual += cost * 2.0
		case "type.annual":
			monthly += cost / 12.0
			annual += cost
		default:
			monthly += cost
			annual += cost * 12.0
		}
	}


	// Daily Expenses (Current Month)
	now := time.Now()
	var dailyMonthly float64
	for _, exp := range m.daily {
		if !exp.Date.IsZero() && exp.Date.Year() == now.Year() && exp.Date.Month() == now.Month() {
			dailyMonthly += parseEuro(exp.Amount)
		}
	}
	if dailyMonthly > 0 {
		monthly += dailyMonthly
		annual += dailyMonthly * 12.0
	}

	// Goals
	for _, g := range m.goals {
		target := parseEuro(g.Target)
		current := parseEuro(g.Current)
		remainingMoney := target - current
		if remainingMoney > 0 && !g.Deadline.IsZero() {
			now := time.Now()
			remainingMonths := g.Deadline.Sub(now).Hours() / (24 * 30.44)
			if remainingMonths < 1 {
				remainingMonths = 1
			}
			monthlyNeeded := remainingMoney / remainingMonths
			monthly += monthlyNeeded
		}
	}

	return annual, monthly
}

func (m *model) renderHome(s *styles) string {
	welcome := s.info.Render(fmt.Sprintf(
		t(m.lang, "home.welcome"),
		s.highlight.Render(m.user),
	))

	// Get Finance Totals
	annual, monthly := m.calculateTotalFinances()
	sw := sidebarWidth(m.width)
	contentW := m.width - sw - 8
	blockW := contentW - 4
	if blockW < 40 {
		blockW = 40
	}

	finBlock := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(blockW).
		Render(
			s.title.Render(t(m.lang, "home.finances")) + "\n\n" +
			fmt.Sprintf("%-10s € %.2f", t(m.lang, "home.annual"), annual) + "\n" +
			fmt.Sprintf("%-10s € %.2f", t(m.lang, "home.monthly"), monthly),
		)

	// Upcoming deadlines (Tasks & Vehicles/Ins)
	now := time.Now()
	twoMonths := now.AddDate(0, 2, 0)
	var upcoming []string

	for _, tsk := range m.tasks {
		if tsk.Deadline != "" && tsk.Status != "DONE" {
			d, err := time.Parse("02/01/2006", tsk.Deadline)
			if err == nil {
				if d.After(now.AddDate(0, 0, -1)) && d.Before(twoMonths) {
					upcoming = append(upcoming, fmt.Sprintf("%s - %s", tsk.Deadline, tsk.Title))
				}
			}
		}
	}
	
	for _, v := range m.vehicles {
		if !v.NTC.IsZero() && v.NTC.After(now.AddDate(0, 0, -1)) && v.NTC.Before(twoMonths) {
			upcoming = append(upcoming, fmt.Sprintf("%s - %s %s", v.NTC.Format("02/01/2006"), t(m.lang, "home.tagNTC"), v.LicensePlate))
		}
		if !v.RoadTax.IsZero() && v.RoadTax.After(now.AddDate(0, 0, -1)) && v.RoadTax.Before(twoMonths) {
			upcoming = append(upcoming, fmt.Sprintf("%s - %s %s", v.RoadTax.Format("02/01/2006"), t(m.lang, "home.tagTax"), v.LicensePlate))
		}
		if !v.Service.IsZero() && v.Service.After(now.AddDate(0, 0, -1)) && v.Service.Before(twoMonths) {
			upcoming = append(upcoming, fmt.Sprintf("%s - %s %s", v.Service.Format("02/01/2006"), t(m.lang, "home.tagService"), v.LicensePlate))
		}
	}
	
	for _, ins := range m.insurances {
		if !ins.ExpireDate.IsZero() && ins.ExpireDate.After(now.AddDate(0, 0, -1)) && ins.ExpireDate.Before(twoMonths) {
			upcoming = append(upcoming, fmt.Sprintf("%s - %s %s", ins.ExpireDate.Format("02/01/2006"), t(m.lang, "home.tagIns"), ins.LicensePlate))
		}
	}
	
	sort.Strings(upcoming)
	if len(upcoming) == 0 {
		upcoming = append(upcoming, s.dim.Render(t(m.lang, "home.noDeadlines")))
	}
	
	// Limit to top 5
	if len(upcoming) > 5 {
		upcoming = upcoming[:5]
	}

	deadlinesBlock := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Width(blockW).
		Render(
			s.title.Render(t(m.lang, "home.deadlines")) + "\n\n" +
			strings.Join(upcoming, "\n"),
		)
		
	// Recent Journal Entries
	var journalDates []string
	if m.journal.Entries != nil {
		for d := range m.journal.Entries {
			journalDates = append(journalDates, d)
		}
		sort.Strings(journalDates)
	}
	
	var recentNotes []string
	if len(journalDates) > 0 {
		for i := len(journalDates) - 1; i >= 0 && len(recentNotes) < 3; i-- {
			d := journalDates[i]
			entry := m.journal.Entries[d]
			// preview
			preview := entry
			if len(preview) > blockW - 12 {
				preview = preview[:blockW - 15] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			recentNotes = append(recentNotes, fmt.Sprintf("%s: %s", d, preview))
		}
	} else {
		recentNotes = append(recentNotes, s.dim.Render(t(m.lang, "home.noNotes")))
	}

	notesBlock := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2).
		Width(blockW).
		Render(
			s.title.Render(t(m.lang, "home.recentNotes")) + "\n\n" +
			strings.Join(recentNotes, "\n"),
		)

	// Recent Tasks
	var recentTasks []string
	if len(m.tasks) > 0 {
		for i := len(m.tasks) - 1; i >= 0 && len(recentTasks) < 3; i-- {
			tsk := m.tasks[i]
			statusCol := "240"
			statusText := tsk.Status
			if tsk.Status == "TODO" {
				statusCol = "214"
				statusText = t(m.lang, "tasks.todo")
			} else if tsk.Status == "DOING" {
				statusCol = "63"
				statusText = t(m.lang, "tasks.doing")
			} else if tsk.Status == "DONE" {
				statusCol = "42"
				statusText = t(m.lang, "tasks.done")
			}
			statusTag := lipgloss.NewStyle().Foreground(lipgloss.Color(statusCol)).Render("[" + statusText + "]")
			recentTasks = append(recentTasks, fmt.Sprintf("%s %s", statusTag, tsk.Title))
		}
	} else {
		recentTasks = append(recentTasks, s.dim.Render(t(m.lang, "home.noTasks")))
	}

	tasksBlock := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(blockW).
		Render(
			s.title.Render(t(m.lang, "home.recentTasks")) + "\n\n" +
			strings.Join(recentTasks, "\n"),
		)

	board := lipgloss.JoinVertical(lipgloss.Left, finBlock, " ", deadlinesBlock, " ", tasksBlock, " ", notesBlock)
	dashboard := welcome + "\n\n" + board

	return dashboard
}
