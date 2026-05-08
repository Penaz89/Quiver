package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func (m *model) updateGoals(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch m.finView {
	case fViewList:
		switch key {
		case "esc", "left":
			m.finSection = fSectionMenu
		case "j", "down":
			if m.goalCursor < len(m.goals)-1 {
				m.goalCursor++
			}
		case "k", "up":
			if m.goalCursor > 0 {
				m.goalCursor--
			}
		case "n":
			m.finView = fViewAdd
			m.goalForm = [4]string{}
			m.goalFormCur = 0
		case "enter":
			if len(m.goals) > 0 {
				m.finView = fViewEdit
				g := m.goals[m.goalCursor]
				m.goalForm = [4]string{g.Name, g.Target, g.Current, g.Deadline}
				m.goalFormCur = 0
				m.goalEditIdx = m.goalCursor
			}
		case "d", "delete":
			if len(m.goals) > 0 {
				m.finView = fViewDelete
			}
		}

	case fViewAdd, fViewEdit:
		switch key {
		case "esc":
			m.finView = fViewList
		case "tab", "down":
			m.goalFormCur = (m.goalFormCur + 1) % 4
		case "shift+tab", "up":
			m.goalFormCur--
			if m.goalFormCur < 0 {
				m.goalFormCur = 3
			}
		case "enter":
			if m.goalFormCur == 3 {
				// save
				g := storage.Goal{
					Name:     strings.TrimSpace(m.goalForm[0]),
					Target:   strings.TrimSpace(m.goalForm[1]),
					Current:  strings.TrimSpace(m.goalForm[2]),
					Deadline: strings.TrimSpace(m.goalForm[3]),
				}
				if g.Name != "" {
					if m.finView == fViewAdd {
						m.goals = append(m.goals, g)
						m.goalCursor = len(m.goals) - 1
					} else {
						m.goals[m.goalEditIdx] = g
					}
					_ = storage.SaveGoals(m.dataDir, m.goals)
				}
				m.finView = fViewList
			} else {
				m.goalFormCur++
			}
		case "backspace":
			if len(m.goalForm[m.goalFormCur]) > 0 {
				m.goalForm[m.goalFormCur] = m.goalForm[m.goalFormCur][:len(m.goalForm[m.goalFormCur])-1]
			}
		case "space":
			m.goalForm[m.goalFormCur] += " "
		default:
			if len(key) == 1 {
				m.goalForm[m.goalFormCur] += key
			}
		}

	case fViewDelete:
		switch key {
		case "y", "Y":
			idx := m.goalCursor
			m.goals = append(m.goals[:idx], m.goals[idx+1:]...)
			if m.goalCursor >= len(m.goals) && m.goalCursor > 0 {
				m.goalCursor--
			}
			_ = storage.SaveGoals(m.dataDir, m.goals)
			m.finView = fViewList
		case "n", "N", "esc", "enter":
			m.finView = fViewList
		}
	}
	return m, nil
}

func (m *model) renderGoals(s *styles) string {
	title := s.title.Render(t(m.lang, "finances.goals"))

	if m.finView == fViewAdd || m.finView == fViewEdit {
		tStr := t(m.lang, "action.add")
		if m.finView == fViewEdit {
			tStr = t(m.lang, "action.edit")
		}
		form := s.title.Render(tStr) + "\n\n"

		labels := []string{
			t(m.lang, "field.goalName"),
			t(m.lang, "field.goalTarget"),
			t(m.lang, "field.goalCurrent"),
			t(m.lang, "field.goalDeadline"),
		}

		for i, label := range labels {
			cursor := "  "
			if m.goalFormCur == i {
				cursor = s.highlight.Render("> ")
			}
			val := m.goalForm[i]
			if m.goalFormCur == i {
				val += "█"
			}
			form += fmt.Sprintf("%s%-20s %s\n", cursor, label+":", val)
		}
		return form + "\n" + s.dim.Render("tab: next • enter: save • esc: cancel")
	}

	if len(m.goals) == 0 {
		return title + "\n\n" + s.dim.Render(t(m.lang, "goals.noRecords")) + "\n\n" + s.dim.Render("n: " + t(m.lang, "action.add"))
	}

	var content string
	if m.finView == fViewDelete {
		content = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(t(m.lang, "delete.confirmGoal")) + "\n\n"
	}

	sw := sidebarWidth(m.width)
	maxW := m.width - sw - 12
	if maxW < 40 {
		maxW = 40
	}

	var items []string
	for i, g := range m.goals {
		target := parseEuro(g.Target)
		current := parseEuro(g.Current)
		pct := 0.0
		if target > 0 {
			pct = (current / target) * 100
		}
		if pct > 100 {
			pct = 100
		}

		barLen := 30
		filled := int((pct / 100.0) * float64(barLen))
		empty := barLen - filled
		if empty < 0 {
			empty = 0
		}

		barStr := strings.Repeat("█", filled) + strings.Repeat("░", empty)
		var color string
		if pct < 33 {
			color = "196" // red
		} else if pct < 66 {
			color = "214" // orange
		} else if pct < 100 {
			color = "226" // yellow
		} else {
			color = "42" // green
		}

		coloredBar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(barStr)

		row := fmt.Sprintf("%-20s | %s | %5.1f%% | €%.2f / €%.2f", g.Name, coloredBar, pct, current, target)
		if g.Deadline != "" {
			row += " | " + g.Deadline
		}

		if i == m.goalCursor {
			items = append(items, s.highlight.Render("▸ "+row))
		} else {
			items = append(items, "  "+row)
		}
	}

	content += strings.Join(items, "\n")
	help := "\n\nn: new • enter: edit • d: delete • esc: back"
	return title + "\n\n" + content + s.dim.Render(help)
}
