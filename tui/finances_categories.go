package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

// ─── Categories logic ───────────────────────────────────────────────

func (m *model) updateCategories(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.finView {
	case fViewAdd, fViewEdit:
		return m.updateCatForm(msg)
	case fViewDelete:
		return m.updateCatDelete(msg)
	default:
		return m.updateCatList(msg)
	}
}

func (m *model) updateCatList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.catCursor > 0 {
			m.catCursor--
		}
	case "down", "j":
		if m.catCursor < len(m.categories)-1 {
			m.catCursor++
		}
	case "a":
		m.finView = fViewAdd
		m.catForm = [2]string{"", ""}
		m.catFormCur = 0
	case "e", "enter":
		if len(m.categories) > 0 {
			m.finView = fViewEdit
			m.catEditIdx = m.catCursor
			catName := m.categories[m.catCursor]
			m.catForm = [2]string{catName, m.budgets[catName]}
			m.catFormCur = 0
		}
	case "d", "x", "del":
		if len(m.categories) > 0 {
			m.finView = fViewDelete
			m.catEditIdx = m.catCursor
		}
	case "esc", "left":
		m.finSection = fSectionMenu
	}
	return m, nil
}

func (m *model) updateCatForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		m.catFormCur = (m.catFormCur + 1) % 2
	case "shift+tab", "up":
		m.catFormCur = (m.catFormCur - 1 + 2) % 2
	case "enter":
		val := strings.TrimSpace(m.catForm[0])
		budgetVal := strings.TrimSpace(m.catForm[1])
		if val == "" {
			m.finView = fViewList
			return m, nil
		}
		if m.finView == fViewAdd {
			m.categories = append(m.categories, val)
		} else {
			oldCat := m.categories[m.catEditIdx]
			m.categories[m.catEditIdx] = val
			
			// Also update existing expenses with the new category name
			if oldCat != val {
				changed := false
				for i, d := range m.daily {
					if d.Category == oldCat {
						m.daily[i].Category = val
						changed = true
					}
				}
				if changed {
					_ = storage.SaveDailyExpenses(m.dataDir, m.daily)
				}
				delete(m.budgets, oldCat)
			}
		}
		if budgetVal != "" {
			if m.budgets == nil {
				m.budgets = make(map[string]string)
			}
			m.budgets[val] = budgetVal
		} else {
			if m.budgets != nil {
				delete(m.budgets, val)
			}
		}
		_ = storage.SaveCategories(m.dataDir, m.categories)
		_ = storage.SaveBudgets(m.dataDir, m.budgets)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		if len(m.catForm[m.catFormCur]) > 0 {
			runes := []rune(m.catForm[m.catFormCur])
			m.catForm[m.catFormCur] = string(runes[:len(runes)-1])
		}
	case "space":
		m.catForm[m.catFormCur] += " "
	default:
		runes := []rune(key)
		if len(runes) == 1 {
			if m.catFormCur == 1 {
				// Only allow numbers and decimal separators for budget
				if strings.ContainsRune("0123456789.,", runes[0]) {
					m.catForm[1] += key
				}
			} else {
				m.catForm[0] += key
			}
		}
	}
	return m, nil
}

func (m *model) updateCatDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "s", "S", "enter":
		if m.catEditIdx >= 0 && m.catEditIdx < len(m.categories) {
			catName := m.categories[m.catEditIdx]
			m.categories = append(m.categories[:m.catEditIdx], m.categories[m.catEditIdx+1:]...)
			if m.budgets != nil {
				delete(m.budgets, catName)
			}
			_ = storage.SaveCategories(m.dataDir, m.categories)
			_ = storage.SaveBudgets(m.dataDir, m.budgets)
		}
		m.finView = fViewList
		if m.catCursor >= len(m.categories) && m.catCursor > 0 {
			m.catCursor--
		}
	case "n", "N", "esc":
		m.finView = fViewList
	}
	return m, nil
}

func (m *model) renderCategories(s *styles) string {
	switch m.finView {
	case fViewAdd:
		return m.renderCatForm(s, t(m.lang, "categories.add"))
	case fViewEdit:
		return m.renderCatForm(s, t(m.lang, "categories.edit"))
	case fViewDelete:
		return m.renderCatDelete(s)
	default:
		return m.renderCatList(s)
	}
}

func (m *model) renderCatList(s *styles) string {
	isActive := m.finSection != fSectionMenu && m.focusContent
	title := s.title.Render(t(m.lang, "categories.title"))
	if len(m.categories) == 0 {
		empty := s.dim.Render(t(m.lang, "categories.noRecords"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	header := s.subtitle.Render(fmt.Sprintf("  %-3s %-20s %-10s", t(m.lang, "col.num"), t(m.lang, "col.category"), t(m.lang, "col.budget")))
	divider := s.dim.Render("  " + strings.Repeat("─", 36))

	var rows []string
	for i, c := range m.categories {
		budgStr := "-"
		if b, ok := m.budgets[c]; ok && b != "" {
			budgStr = "€ " + b
		}
		row := fmt.Sprintf("  %-3d %-20s %-10s", i+1, truncate(c, 19), budgStr)
		if i == m.catCursor {
			if isActive {
				row = s.menuSelected.Width(0).Render(row)
			} else {
				row = s.menuActiveDim.Width(0).Render(row)
			}
		} else {
			row = s.info.Render(row)
		}
		rows = append(rows, row)
	}

	table := strings.Join(rows, "\n")
	help := s.dim.Render(fmt.Sprintf("\n\na: %s  e: %s  d: %s  ←: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))

	return title + "\n" + header + "\n" + divider + "\n" + table + help
}

func (m *model) renderCatForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	fields := []string{
		t(m.lang, "categories.name"),
		t(m.lang, "col.budget") + " (€)",
	}

	var renderedLines []string
	for i, f := range fields {
		val := m.catForm[i]
		if i == m.catFormCur {
			renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-20s ", f+":"))+s.fieldBg.Render(val+s.highlight.Render("_")))
		} else {
			renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-20s ", f+":"))+val)
		}
	}

	rendered := strings.Join(renderedLines, "\n\n")

	help := s.dim.Render(fmt.Sprintf("\n\nEnter: %s  Esc: %s", t(m.lang, "action.save"), t(m.lang, "action.cancel")))

	return title + "\n\n" + rendered + help
}

func (m *model) renderCatDelete(s *styles) string {
	if m.catEditIdx < 0 || m.catEditIdx >= len(m.categories) {
		return m.renderCatList(s)
	}
	c := m.categories[m.catEditIdx]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.categories"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(t(m.lang, "categories.confirmDelete"))

	info := fmt.Sprintf(
		"\n  %s %s",
		s.dim.Render(t(m.lang, "col.category")+":"), s.info.Render(c),
	)

	help := s.dim.Render(fmt.Sprintf("\n\ny/s: %s  n/Esc: %s",
		t(m.lang, "action.delete"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + "\n" + info + help
}
