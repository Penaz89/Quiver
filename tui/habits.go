package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func (m *model) updateHabits(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.habitIsAdding {
		switch msg.String() {
		case "enter":
			if m.habitForm != "" {
				m.habits = append(m.habits, storage.Habit{
					Name:      m.habitForm,
					Completed: make(map[string]bool),
					CreatedAt: time.Now().Format("2006-01-02"),
				})
				_ = storage.SaveHabits(m.dataDir, m.habits)
				m.habitForm = ""
				m.habitIsAdding = false
				m.habitCursor = len(m.habits) - 1
			}
		case "esc":
			m.habitIsAdding = false
			m.habitForm = ""
		case "backspace":
			if len(m.habitForm) > 0 {
				runes := []rune(m.habitForm)
				m.habitForm = string(runes[:len(runes)-1])
			}
		default:
			if len(msg.String()) == 1 {
				m.habitForm += msg.String()
			} else if msg.String() == "space" {
				m.habitForm += " "
			}
		}
		return m, nil
	}

	if m.habitIsDeleting {
		switch strings.ToLower(msg.String()) {
		case "y", "s":
			if len(m.habits) > 0 {
				m.habits = append(m.habits[:m.habitCursor], m.habits[m.habitCursor+1:]...)
				if m.habitCursor >= len(m.habits) && m.habitCursor > 0 {
					m.habitCursor--
				}
				_ = storage.SaveHabits(m.dataDir, m.habits)
			}
			m.habitIsDeleting = false
		default:
			m.habitIsDeleting = false
		}
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.habitCursor > 0 {
			m.habitCursor--
		}
	case "down", "j":
		if m.habitCursor < len(m.habits)-1 {
			m.habitCursor++
		}
	case "n":
		m.habitIsAdding = true
		m.habitForm = ""
	case "d", "del":
		if len(m.habits) > 0 {
			m.habitIsDeleting = true
		}
	case "space":
		if len(m.habits) > 0 {
			today := time.Now().Format("2006-01-02")
			if m.habits[m.habitCursor].Completed == nil {
				m.habits[m.habitCursor].Completed = make(map[string]bool)
			}
			m.habits[m.habitCursor].Completed[today] = !m.habits[m.habitCursor].Completed[today]
			_ = storage.SaveHabits(m.dataDir, m.habits)
		}
	case "esc", "left":
		m.focusContent = false
	}
	return m, nil
}

func (m *model) renderHabitsView(s *styles) string {
	title := s.title.Render(t(m.lang, "habits.title"))
	desc := s.subtitle.Render(t(m.lang, "habits.subtitle"))

	if m.habitIsAdding {
		formTitle := s.info.Render(t(m.lang, "habits.add"))
		input := s.menuSelected.Width(0).Render("  ▸ " + t(m.lang, "habits.name") + ": " + m.habitForm + s.highlight.Render("_"))
		help := s.dim.Render("\n\nEnter: " + t(m.lang, "action.save") + "  Esc: " + t(m.lang, "help.goBack"))
		return title + "\n" + desc + "\n\n" + formTitle + "\n\n" + input + help
	}

	if len(m.habits) == 0 {
		empty := s.dim.Render(t(m.lang, "habits.noItems"))
		return title + "\n" + desc + "\n\n" + empty
	}

	var listLines []string
	for i, h := range m.habits {
		today := time.Now().Format("2006-01-02")
		doneToday := h.Completed[today]
		status := " "
		if doneToday {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
		}

		if i == m.habitCursor {
			listLines = append(listLines, s.menuSelected.Width(0).Render(fmt.Sprintf("  ▸ [%s] %s", status, h.Name)))
		} else {
			listLines = append(listLines, s.menuNormal.Width(0).Render(fmt.Sprintf("    [%s] %s", status, h.Name)))
		}
	}

	// Render heatmap for the selected habit
	var heatmap string
	if len(m.habits) > 0 && m.habitCursor < len(m.habits) {
		selected := m.habits[m.habitCursor]
		heatmap = m.renderHeatmap(selected, s)
	}

	listStr := strings.Join(listLines, "\n")
	
	// Layout: List on top, heatmap below
	layout := listStr + "\n\n" + heatmap

	if m.habitIsDeleting {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		layout = listStr + "\n\n" + warningStyle.Render(t(m.lang, "habits.confirmDelete"))
	}

	help := s.dim.Render(fmt.Sprintf("\n\n↑/↓: %s  %s  %s  %s  ←: %s",
		t(m.lang, "help.navigate"),
		t(m.lang, "habits.toggleHelp"),
		t(m.lang, "habits.addHelp"),
		t(m.lang, "habits.deleteHelp"),
		t(m.lang, "help.goBack"),
	))

	return title + "\n" + desc + "\n\n" + layout + help
}

func (m *model) renderHeatmap(habit storage.Habit, s *styles) string {
	// 7 rows (Sun-Sat) x cols (weeks)
	cols := 28
	
	now := time.Now()
	
	grid := make([][]string, 7)
	for i := range grid {
		grid[i] = make([]string, cols)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	cellDone := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("■")
	cellMiss := lipgloss.NewStyle().Foreground(lipgloss.Color("237")).Render("■")
	cellFuture := s.dim.Render("·")

	totalDays := (cols-1)*7 + int(now.Weekday()) + 1
	startDate := now.AddDate(0, 0, -totalDays+1)

	currentDate := startDate
	for c := 0; c < cols; c++ {
		for r := 0; r < 7; r++ {
			if currentDate.After(now) {
				grid[r][c] = cellFuture
				continue
			}

			dateStr := currentDate.Format("2006-01-02")
			if habit.Completed[dateStr] {
				grid[r][c] = cellDone
			} else {
				grid[r][c] = cellMiss
			}
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	var rows []string
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if m.lang == "it" {
		days = []string{"Dom", "Lun", "Mar", "Mer", "Gio", "Ven", "Sab"}
	}

	for r := 0; r < 7; r++ {
		rowStr := s.dim.Render(fmt.Sprintf("    %3s ", days[r]))
		for c := 0; c < cols; c++ {
			rowStr += grid[r][c] + " "
		}
		rows = append(rows, rowStr)
	}

	return s.info.Render("  Heatmap:") + "\n" + strings.Join(rows, "\n")
}
