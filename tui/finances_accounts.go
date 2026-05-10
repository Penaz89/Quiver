package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

// ─── Accounts logic ───────────────────────────────────────────────

const (
	accFName = iota
	accFBalance
	accFType
	accFCount
)

func getAccTypes() []string {
	return []string{"type.bank", "type.cash", "type.crypto", "type.investment"}
}

func (m *model) adjustAccountBalance(name string, delta float64) {
	if name == "" || delta == 0 {
		return
	}
	for i, a := range m.accounts {
		if a.Name == name {
			current := parseEuro(a.Balance)
			newBalance := current + delta
			m.accounts[i].Balance = fmt.Sprintf("%.2f", newBalance)
			_ = storage.SaveAccounts(m.dataDir, m.accounts)
			break
		}
	}
}

func (m *model) updateAccounts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.finView {
	case fViewAdd, fViewEdit:
		return m.updateAccForm(msg)
	case fViewDelete:
		return m.updateAccDelete(msg)
	default:
		return m.updateAccList(msg)
	}
}

func (m *model) updateAccList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "left":
		m.finSection = fSectionMenu
	case "up", "k":
		if m.accCursor > 0 {
			m.accCursor--
		}
	case "down", "j":
		if m.accCursor < len(m.accounts)-1 {
			m.accCursor++
		}
	case "a":
		m.finView = fViewAdd
		m.accForm = [accFCount]string{"", "0.00", getAccTypes()[0]}
		m.accFormCur = 0
	case "e", "enter":
		if len(m.accounts) > 0 {
			m.finView = fViewEdit
			m.accEditIdx = m.accCursor
			acc := m.accounts[m.accCursor]
			m.accForm = [accFCount]string{
				acc.Name,
				acc.Balance,
				acc.Type,
			}
			m.accFormCur = 0
		}
	case "d", "x", "del":
		if len(m.accounts) > 0 {
			m.finView = fViewDelete
			m.accEditIdx = m.accCursor
		}
	}
	return m, nil
}

func (m *model) updateAccForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	types := getAccTypes()

	switch key {
	case "tab", "down":
		m.accFormCur = (m.accFormCur + 1) % accFCount
	case "shift+tab", "up":
		m.accFormCur = (m.accFormCur - 1 + accFCount) % accFCount
	case "left", "right":
		if m.accFormCur == accFType {
			current := m.accForm[accFType]
			idx := 0
			for i, t := range types {
				if t == current {
					idx = i
					break
				}
			}
			if key == "right" {
				idx = (idx + 1) % len(types)
			} else {
				idx = (idx - 1 + len(types)) % len(types)
			}
			m.accForm[accFType] = types[idx]
		}
	case "enter":
		if m.accForm[accFName] == "" || m.accForm[accFBalance] == "" {
			m.finView = fViewList
			return m, nil
		}

		newAcc := storage.Account{
			Name:    strings.TrimSpace(m.accForm[accFName]),
			Balance: strings.TrimSpace(m.accForm[accFBalance]),
			Type:    m.accForm[accFType],
			Author:  m.user,
		}

		if m.finView == fViewAdd {
			m.accounts = append(m.accounts, newAcc)
		} else {
			m.accounts[m.accEditIdx] = newAcc
		}

		_ = storage.SaveAccounts(m.dataDir, m.accounts)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		if m.accFormCur != accFType {
			field := &m.accForm[m.accFormCur]
			if len(*field) > 0 {
				runes := []rune(*field)
				*field = string(runes[:len(runes)-1])
			}
		}
	case "space":
		if m.accFormCur == accFType {
			current := m.accForm[accFType]
			idx := 0
			for i, t := range types {
				if t == current {
					idx = i
					break
				}
			}
			idx = (idx + 1) % len(types)
			m.accForm[accFType] = types[idx]
		} else {
			m.accForm[m.accFormCur] += " "
		}
	default:
		if m.accFormCur != accFType {
			runes := []rune(key)
			if len(runes) == 1 {
				field := &m.accForm[m.accFormCur]
				if m.accFormCur == accFBalance {
					if strings.ContainsRune("0123456789.,-", runes[0]) {
						*field += key
					}
				} else {
					*field += key
				}
			}
		}
	}
	return m, nil
}

func (m *model) updateAccDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.accEditIdx >= 0 && m.accEditIdx < len(m.accounts) {
			m.accounts = append(m.accounts[:m.accEditIdx], m.accounts[m.accEditIdx+1:]...)
			_ = storage.SaveAccounts(m.dataDir, m.accounts)
		}
		m.finView = fViewList
		if m.accCursor >= len(m.accounts) && m.accCursor > 0 {
			m.accCursor--
		}
	case "n", "N", "esc":
		m.finView = fViewList
	}
	return m, nil
}

func (m *model) renderAccounts(s *styles) string {
	switch m.finView {
	case fViewAdd:
		return m.renderAccForm(s, t(m.lang, "action.add")+" "+t(m.lang, "finances.accounts"))
	case fViewEdit:
		return m.renderAccForm(s, t(m.lang, "action.edit")+" "+t(m.lang, "finances.accounts"))
	case fViewDelete:
		return m.renderAccDelete(s)
	default:
		return m.renderAccList(s)
	}
}

func (m *model) renderAccList(s *styles) string {
	isActive := m.finSection != fSectionMenu && m.focusContent
	title := s.title.Render(t(m.lang, "finances.accounts"))

	if len(m.accounts) == 0 {
		empty := s.dim.Render(t(m.lang, "accounts.noRecords"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	headerStr := fmt.Sprintf("  %-20s %-15s %-15s",
		t(m.lang, "col.account"), t(m.lang, "col.balance"), t(m.lang, "col.type"))
	header := s.subtitle.Render(headerStr)
	divider := s.dim.Render("  " + strings.Repeat("─", 52))

	var rows []string
	var totalBalance float64

	for i, acc := range m.accounts {
		bal := parseEuro(acc.Balance)
		totalBalance += bal

		balStr := fmt.Sprintf("€ %.2f", bal)
		typeStr := t(m.lang, acc.Type)
		row := fmt.Sprintf("  %-20s %-15s %-15s", truncate(acc.Name, 19), balStr, truncate(typeStr, 14))

		if i == m.accCursor {
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
	
	totalDiv := s.dim.Render("  " + strings.Repeat("=", 52))
	totalStr := s.highlight.Render(fmt.Sprintf("  %-20s € %.2f", "TOTALE", totalBalance))

	help := s.dim.Render(fmt.Sprintf("\n\na: %s • e: %s • d: %s • ←: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))

	return title + "\n" + header + "\n" + divider + "\n" + table + "\n" + totalDiv + "\n" + totalStr + help
}

func (m *model) renderAccForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	fields := []string{
		t(m.lang, "col.account") + " Nome",
		t(m.lang, "col.balance") + " (€)",
		t(m.lang, "col.type"),
	}

	var renderedLines []string
	for i, f := range fields {
		val := m.accForm[i]
		
		var displayVal string
		if i == accFType {
			displayVal = t(m.lang, val)
		} else {
			displayVal = val
		}
		
		if i == m.accFormCur {
			if i == accFType {
				renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-20s ", f+":"))+s.fieldBg.Render("< "+displayVal+" >"))
			} else {
				renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-20s ", f+":"))+s.fieldBg.Render(displayVal+s.highlight.Render("_")))
			}
		} else {
			renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-20s ", f+":"))+s.info.Render(displayVal))
		}
	}

	rendered := strings.Join(renderedLines, "\n\n")
	help := s.dim.Render(fmt.Sprintf("\n\nEnter: %s  Esc: %s", t(m.lang, "action.save"), t(m.lang, "action.cancel")))
	return title + "\n\n" + rendered + help
}

func (m *model) renderAccDelete(s *styles) string {
	if m.accEditIdx < 0 || m.accEditIdx >= len(m.accounts) {
		return m.renderAccList(s)
	}
	acc := m.accounts[m.accEditIdx]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.accounts"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("Delete this account? (y/N)")

	info := fmt.Sprintf(
		"\n  %s %s\n  %s %s",
		s.dim.Render(t(m.lang, "col.account")+":"), s.info.Render(acc.Name),
		s.dim.Render(t(m.lang, "col.balance")+":"), s.info.Render("€ "+acc.Balance),
	)

	help := s.dim.Render(fmt.Sprintf("\n\ny/s: %s  n/Esc: %s",
		t(m.lang, "action.delete"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + "\n" + info + help
}
