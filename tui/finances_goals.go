package tui

import (
	"fmt"
	"strings"
	"time"

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
				deadlineStr := ""
				if !g.Deadline.IsZero() {
					deadlineStr = g.Deadline.Format("02/01/2006")
				}
				m.goalForm = [4]string{g.Name, g.Target, g.Current, deadlineStr}
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
				deadlineStr := strings.TrimSpace(m.goalForm[3])
				var deadline time.Time
				if deadlineStr != "" {
					d, err := time.Parse("02/01/2006", deadlineStr)
					if err != nil {
						return m, nil
					}
					deadline = d
				}
				g := storage.Goal{
					Name:     strings.TrimSpace(m.goalForm[0]),
					Target:   strings.TrimSpace(m.goalForm[1]),
					Current:  strings.TrimSpace(m.goalForm[2]),
					Deadline: deadline,
				}
				if g.Name != "" {
					if m.finView == fViewAdd {
						g.CreatedAt = time.Now()
						m.goals = append(m.goals, g)
						m.goalCursor = len(m.goals) - 1
					} else {
						g.CreatedAt = m.goals[m.goalEditIdx].CreatedAt
						if g.CreatedAt.IsZero() {
							g.CreatedAt = time.Now()
						}
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
			runes := []rune(key)
			if len(runes) == 1 {
				field := &m.goalForm[m.goalFormCur]
				if m.goalFormCur == 3 {
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
				} else if m.goalFormCur == 1 || m.goalFormCur == 2 {
					if !strings.ContainsRune("0123456789.,", runes[0]) {
						return m, nil
					}
					*field += key
				} else {
					*field += key
				}
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
		if !g.Deadline.IsZero() {
			row += " | " + g.Deadline.Format("02/01/2006")
		}
		
		var extraInfo string
		if !g.Deadline.IsZero() {
			deadline := g.Deadline
			now := time.Now()
			
			// Countdown
			remainingDays := int(deadline.Sub(now).Hours() / 24)
			
			if remainingDays < 0 {
				extraInfo = "\n    └ " + s.dim.Render(t(m.lang, "goals.expired"))
			} else {
				remainingMoney := target - current
				if remainingMoney > 0 {
					// Months remaining from today
					remainingMonths := deadline.Sub(now).Hours() / (24 * 30.44)
					if remainingMonths < 1 {
						remainingMonths = 1
					}
					monthlyNeeded := remainingMoney / remainingMonths
					
					expiresStr := fmt.Sprintf(t(m.lang, "goals.expiresIn"), remainingDays)
					saveStr := fmt.Sprintf(t(m.lang, "goals.savePerMonth"), monthlyNeeded)
					extraInfo = fmt.Sprintf("\n    └ %s. %s", expiresStr, saveStr)
				} else {
					expiresStr := fmt.Sprintf(t(m.lang, "goals.expiresIn"), remainingDays)
					extraInfo = fmt.Sprintf("\n    └ %s. %s", expiresStr, s.highlight.Render(t(m.lang, "goals.reached")))
				}
			}
		}

		if i == m.goalCursor {
			items = append(items, s.highlight.Render("▸ "+row) + extraInfo)
		} else {
			items = append(items, "  "+row + extraInfo)
		}
	}

	content += strings.Join(items, "\n\n")
	help := "\n\nn: new • enter: edit • d: delete • esc: back"
	return title + "\n\n" + content + s.dim.Render(help)
}
