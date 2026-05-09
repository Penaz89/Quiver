package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

var taskStatuses = []string{"TODO", "DOING", "DONE"}

// getColumnTasks returns the indices of tasks in the specified column
func (m *model) getColumnTasks(colIdx int) []int {
	if colIdx < 0 || colIdx > 2 {
		return []int{}
	}
	status := taskStatuses[colIdx]
	var res []int
	for i, t := range m.tasks {
		if t.Status == status {
			res = append(res, i)
		}
	}
	return res
}

func (m *model) updateTasks(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.taskIsAdding || m.taskIsEditing {
		switch msg.String() {
		case "esc":
			m.taskIsAdding = false
			m.taskIsEditing = false
			m.taskFormError = ""
		case "tab", "down":
			m.taskFormCursor = (m.taskFormCursor + 1) % 4
		case "shift+tab", "up":
			m.taskFormCursor--
			if m.taskFormCursor < 0 {
				m.taskFormCursor = 3
			}
		case "enter":
			if m.taskFormCursor == 3 {
				m.taskFormError = ""
				deadline := strings.TrimSpace(m.taskFormFields[3])
				if deadline != "" {
					_, err := time.Parse("02/01/2006", deadline)
					if err != nil {
						m.taskFormError = t(m.lang, "tasks.invalidDate")
						return m, nil
					}
				}
				// Save
				t := storage.Task{
					Title:    strings.TrimSpace(m.taskFormFields[0]),
					Project:  strings.TrimSpace(m.taskFormFields[1]),
					Priority: strings.TrimSpace(m.taskFormFields[2]),
					Deadline: strings.TrimSpace(m.taskFormFields[3]),
					Author:   m.user,
				}
				if t.Title != "" {
					if m.taskIsAdding {
						t.Status = taskStatuses[m.taskColumn]
						m.tasks = append(m.tasks, t)
					} else {
						// Editing
						colTasks := m.getColumnTasks(m.taskColumn)
						if m.taskCursor >= 0 && m.taskCursor < len(colTasks) {
							idx := colTasks[m.taskCursor]
							t.Status = m.tasks[idx].Status
							m.tasks[idx] = t
						}
					}
					_ = storage.SaveTasks(m.dataDir, m.tasks)
				}
				m.taskIsAdding = false
				m.taskIsEditing = false
			} else {
				m.taskFormCursor++
			}
		case "backspace":
			if len(m.taskFormFields[m.taskFormCursor]) > 0 {
				m.taskFormFields[m.taskFormCursor] = m.taskFormFields[m.taskFormCursor][:len(m.taskFormFields[m.taskFormCursor])-1]
			}
		case "space":
			m.taskFormFields[m.taskFormCursor] += " "
		default:
			if len(msg.String()) == 1 {
				m.taskFormFields[m.taskFormCursor] += msg.String()
			}
		}
		return m, nil
	}

	colTasks := m.getColumnTasks(m.taskColumn)

	switch msg.String() {
	case "q", "esc":
		m.focusContent = false
	case "h", "left":
		if m.taskColumn > 0 {
			m.taskColumn--
			m.taskCursor = 0
		} else {
			m.focusContent = false
		}
	case "l", "right":
		if m.taskColumn < 2 {
			m.taskColumn++
			m.taskCursor = 0
		}
	case "j", "down":
		if m.taskCursor < len(colTasks)-1 {
			m.taskCursor++
		}
	case "k", "up":
		if m.taskCursor > 0 {
			m.taskCursor--
		}
	case "H": // Move task left
		if len(colTasks) > 0 && m.taskColumn > 0 {
			idx := colTasks[m.taskCursor]
			m.tasks[idx].Status = taskStatuses[m.taskColumn-1]
			_ = storage.SaveTasks(m.dataDir, m.tasks)
			m.taskColumn--
			m.taskCursor = len(m.getColumnTasks(m.taskColumn)) - 1
		}
	case "L": // Move task right
		if len(colTasks) > 0 && m.taskColumn < 2 {
			idx := colTasks[m.taskCursor]
			m.tasks[idx].Status = taskStatuses[m.taskColumn+1]
			_ = storage.SaveTasks(m.dataDir, m.tasks)
			m.taskColumn++
			m.taskCursor = len(m.getColumnTasks(m.taskColumn)) - 1
		}
	case "J": // Move down in order
		if m.taskCursor < len(colTasks)-1 {
			idx1 := colTasks[m.taskCursor]
			idx2 := colTasks[m.taskCursor+1]
			m.tasks[idx1], m.tasks[idx2] = m.tasks[idx2], m.tasks[idx1]
			_ = storage.SaveTasks(m.dataDir, m.tasks)
			m.taskCursor++
		}
	case "K": // Move up in order
		if m.taskCursor > 0 {
			idx1 := colTasks[m.taskCursor]
			idx2 := colTasks[m.taskCursor-1]
			m.tasks[idx1], m.tasks[idx2] = m.tasks[idx2], m.tasks[idx1]
			_ = storage.SaveTasks(m.dataDir, m.tasks)
			m.taskCursor--
		}
	case "n":
		m.taskIsAdding = true
		m.taskFormFields = [4]string{}
		m.taskFormCursor = 0
		m.taskFormError = ""
	case "enter":
		if len(colTasks) > 0 {
			m.taskIsEditing = true
			idx := colTasks[m.taskCursor]
			tsk := m.tasks[idx]
			m.taskFormFields = [4]string{tsk.Title, tsk.Project, tsk.Priority, tsk.Deadline}
			m.taskFormCursor = 0
			m.taskFormError = ""
		}
	case "d", "delete":
		if len(colTasks) > 0 {
			idx := colTasks[m.taskCursor]
			m.tasks = append(m.tasks[:idx], m.tasks[idx+1:]...)
			if m.taskCursor >= len(m.getColumnTasks(m.taskColumn)) {
				m.taskCursor--
				if m.taskCursor < 0 {
					m.taskCursor = 0
				}
			}
			_ = storage.SaveTasks(m.dataDir, m.tasks)
		}
	}
	return m, nil
}

func (m *model) renderTasksView(s *styles) string {
	if m.taskIsAdding || m.taskIsEditing {
		title := t(m.lang, "tasks.add")
		if m.taskIsEditing {
			title = t(m.lang, "tasks.edit")
		}
		
		form := s.title.Render(title) + "\n\n"
		labels := []string{
			t(m.lang, "tasks.titleLabel"),
			t(m.lang, "tasks.projectLabel"),
			t(m.lang, "tasks.priorityLabel"),
			t(m.lang, "tasks.deadlineLabel"),
		}

		for i, label := range labels {
			cursor := "  "
			if m.taskFormCursor == i {
				cursor = s.highlight.Render("> ")
			}
			val := m.taskFormFields[i]
			if m.taskFormCursor == i {
				val += "█"
			}
			form += fmt.Sprintf("%s%-25s %s\n", cursor, label+":", val)
		}
		
		if m.taskFormError != "" {
			form += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.taskFormError)
		}
		
		form += "\n\n" + s.dim.Render("tab: next • enter: save • esc: cancel")
		return form
	}

	title := s.title.Render(t(m.lang, "tasks.title"))
	desc := s.subtitle.Render(t(m.lang, "tasks.subtitle"))

	sw := sidebarWidth(m.width)
	contentW := m.width - sw - 8
	vpWidth := contentW - 4
	if vpWidth < 10 {
		vpWidth = 10
	}
	colW := vpWidth / 3
	if colW < 20 {
		colW = 20 // minimum column width
	}

	// Render the 3 columns
	cols := make([]string, 3)
	headers := []string{
		t(m.lang, "tasks.todo"),
		t(m.lang, "tasks.doing"),
		t(m.lang, "tasks.done"),
	}

	for i := 0; i < 3; i++ {
		colHeader := headers[i]
		if m.taskColumn == i {
			colHeader = s.highlight.Render(colHeader)
		} else {
			colHeader = s.info.Render(colHeader)
		}
		
		colTasks := m.getColumnTasks(i)
		
		var taskStrs []string
		for j, idx := range colTasks {
			tsk := m.tasks[idx]
			
			// Format task
			tStr := tsk.Title
			if tsk.Project != "" {
				tStr += s.dim.Render(" [" + tsk.Project + "]")
			}
			
			meta := ""
			if tsk.Author != "" {
				meta += s.dim.Render("[" + tsk.Author + "] ")
			}
			
			if tsk.Priority != "" {
				pCol := "240"
				if tsk.Priority == "H" || tsk.Priority == "A" { pCol = "196" }
				if tsk.Priority == "M" { pCol = "214" }
				meta += lipgloss.NewStyle().Foreground(lipgloss.Color(pCol)).Render(tsk.Priority) + " "
			}
			if tsk.Deadline != "" {
				meta += s.dim.Render(tsk.Deadline)
			}
			
			if meta != "" {
				tStr += "\n" + meta
			}
			
			// Check if selected
			boxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Width(colW - 6).
				Padding(0, 1)
				
			if m.taskColumn == i && m.taskCursor == j {
				boxStyle = boxStyle.BorderForeground(lipgloss.Color("42"))
			} else {
				boxStyle = boxStyle.BorderForeground(lipgloss.Color("240"))
			}
			
			taskStrs = append(taskStrs, boxStyle.Render(tStr))
		}
		
		colContent := colHeader + "\n\n" + strings.Join(taskStrs, "\n")
		
		cols[i] = lipgloss.NewStyle().
			Width(colW - 2).
			MarginRight(2).
			Render(colContent)
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	help := s.dim.Render(fmt.Sprintf("%s  %s  %s\n%s  %s  %s",
		t(m.lang, "tasks.navHelp"),
		t(m.lang, "tasks.addHelp"),
		t(m.lang, "tasks.editHelp"),
		t(m.lang, "tasks.moveHelp"),
		t(m.lang, "tasks.statusHelp"),
		t(m.lang, "tasks.deleteHelp"),
	))

	return title + "\n" + desc + "\n\n" + board + "\n\n" + help
}
