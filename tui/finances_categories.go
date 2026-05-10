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
		m.catForm = ""
	case "e", "enter":
		if len(m.categories) > 0 {
			m.finView = fViewEdit
			m.catEditIdx = m.catCursor
			m.catForm = m.categories[m.catCursor]
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
	case "enter":
		val := strings.TrimSpace(m.catForm)
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
			}
		}
		_ = storage.SaveCategories(m.dataDir, m.categories)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		if len(m.catForm) > 0 {
			runes := []rune(m.catForm)
			m.catForm = string(runes[:len(runes)-1])
		}
	case "space":
		m.catForm += " "
	default:
		runes := []rune(key)
		if len(runes) == 1 {
			m.catForm += key
		}
	}
	return m, nil
}

func (m *model) updateCatDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "s", "S", "enter":
		if m.catEditIdx >= 0 && m.catEditIdx < len(m.categories) {
			m.categories = append(m.categories[:m.catEditIdx], m.categories[m.catEditIdx+1:]...)
			_ = storage.SaveCategories(m.dataDir, m.categories)
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

	header := s.subtitle.Render(fmt.Sprintf("  %-3s %-20s", t(m.lang, "col.num"), t(m.lang, "col.category")))
	divider := s.dim.Render("  " + strings.Repeat("─", 25))

	var rows []string
	for i, c := range m.categories {
		row := fmt.Sprintf("  %-3d %-20s", i+1, truncate(c, 19))
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

	label := s.dim.Render(fmt.Sprintf("  %-15s", t(m.lang, "categories.name")+":"))
	cursor := s.highlight.Render("_")
	fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
	
	rendered := label + " " + fieldStyle.Render(m.catForm) + cursor

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
