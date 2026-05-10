package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

// ─── Installments logic ──────────────────────────────────────────

const (
	instFName = iota
	instFAmount
	instFTotalCount
	instFPaidCount
	instFFrequency
	instFStartDate
	instFAccount
	instFCount
)

func getInstFrequencies() []string {
	return []string{"type.monthly", "type.bimonthly", "type.quarterly", "type.semiannual", "type.annual"}
}

func (m *model) updateInstallments(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.finView {
	case fViewAdd, fViewEdit:
		return m.updateInstForm(msg)
	case fViewDelete:
		return m.updateInstDelete(msg)
	default:
		return m.updateInstList(msg)
	}
}

func (m *model) updateInstList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "left":
		m.finSection = fSectionMenu
	case "up", "k":
		if m.instCursor > 0 {
			m.instCursor--
		}
	case "down", "j":
		if m.instCursor < len(m.installments)-1 {
			m.instCursor++
		}
	case "a":
		m.finView = fViewAdd
		defaultAcc := ""
		if len(m.accounts) > 0 {
			defaultAcc = m.accounts[0].Name
		}
		m.instForm = [instFCount]string{"", "", "0", "0", getInstFrequencies()[0], time.Now().Format("02/01/2006"), defaultAcc}
		m.instFormCur = 0
	case "e", "enter":
		if len(m.installments) > 0 {
			m.finView = fViewEdit
			m.instEditIdx = m.instCursor
			inst := m.installments[m.instCursor]
			dateStr := inst.StartDate.Format("02/01/2006")
			if inst.StartDate.IsZero() {
				dateStr = ""
			}
			freq := inst.Frequency
			if freq == "" {
				freq = getInstFrequencies()[0]
			}
			accName := inst.Account
			if accName == "" && len(m.accounts) > 0 {
				accName = m.accounts[0].Name
			}
			m.instForm = [instFCount]string{
				inst.Name,
				inst.Amount,
				fmt.Sprintf("%d", inst.TotalCount),
				fmt.Sprintf("%d", inst.PaidCount),
				freq,
				dateStr,
				accName,
			}
			m.instFormCur = 0
		}
	case "d", "x", "del":
		if len(m.installments) > 0 {
			m.finView = fViewDelete
			m.instEditIdx = m.instCursor
		}
	case "space": // Pay installment action
		if len(m.installments) > 0 {
			inst := &m.installments[m.instCursor]
			if inst.TotalCount == 0 || inst.PaidCount < inst.TotalCount {
				inst.PaidCount++
				_ = storage.SaveInstallments(m.dataDir, m.installments)
				if inst.Account != "" {
					m.adjustAccountBalance(inst.Account, -parseEuro(inst.Amount))
				}
			}
		}
	case "backspace": // Undo pay
		if len(m.installments) > 0 {
			inst := &m.installments[m.instCursor]
			if inst.PaidCount > 0 {
				inst.PaidCount--
				_ = storage.SaveInstallments(m.dataDir, m.installments)
				if inst.Account != "" {
					m.adjustAccountBalance(inst.Account, parseEuro(inst.Amount))
				}
			}
		}
	}
	return m, nil
}

func (m *model) updateInstForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	freqs := getInstFrequencies()

	switch key {
	case "tab", "down":
		m.instFormCur = (m.instFormCur + 1) % instFCount
	case "shift+tab", "up":
		m.instFormCur = (m.instFormCur - 1 + instFCount) % instFCount
	case "left", "right":
		if m.instFormCur == instFFrequency {
			current := m.instForm[instFFrequency]
			idx := 0
			for i, f := range freqs {
				if f == current {
					idx = i
					break
				}
			}
			if key == "right" {
				idx = (idx + 1) % len(freqs)
			} else {
				idx = (idx - 1 + len(freqs)) % len(freqs)
			}
			m.instForm[instFFrequency] = freqs[idx]
		} else if m.instFormCur == instFAccount && len(m.accounts) > 0 {
			current := m.instForm[instFAccount]
			idx := -1
			for i, a := range m.accounts {
				if a.Name == current {
					idx = i
					break
				}
			}
			if key == "right" {
				idx = (idx + 1) % len(m.accounts)
			} else {
				if idx == -1 {
					idx = 0
				}
				idx = (idx - 1 + len(m.accounts)) % len(m.accounts)
			}
			m.instForm[instFAccount] = m.accounts[idx].Name
		}
	case "enter":
		if m.instForm[instFName] == "" || m.instForm[instFAmount] == "" {
			m.finView = fViewList
			return m, nil
		}

		var parsedDate time.Time
		dateStr := strings.TrimSpace(m.instForm[instFStartDate])
		if dateStr != "" {
			d, err := time.Parse("02/01/2006", dateStr)
			if err == nil {
				parsedDate = d
			} else {
				parsedDate = time.Now()
			}
		}

		totalCount := 0
		fmt.Sscanf(m.instForm[instFTotalCount], "%d", &totalCount)
		paidCount := 0
		fmt.Sscanf(m.instForm[instFPaidCount], "%d", &paidCount)

		newInst := storage.Installment{
			Name:       strings.TrimSpace(m.instForm[instFName]),
			Amount:     strings.TrimSpace(m.instForm[instFAmount]),
			TotalCount: totalCount,
			PaidCount:  paidCount,
			Frequency:  m.instForm[instFFrequency],
			StartDate:  parsedDate,
			Account:    m.instForm[instFAccount],
			Author:     m.user,
		}

		if m.finView == fViewAdd {
			m.installments = append(m.installments, newInst)
		} else {
			m.installments[m.instEditIdx] = newInst
		}

		_ = storage.SaveInstallments(m.dataDir, m.installments)
		m.finView = fViewList
	case "esc":
		m.finView = fViewList
	case "backspace":
		if m.instFormCur != instFFrequency && m.instFormCur != instFAccount {
			field := &m.instForm[m.instFormCur]
			if len(*field) > 0 {
				runes := []rune(*field)
				*field = string(runes[:len(runes)-1])
			}
		}
	case "space":
		if m.instFormCur == instFFrequency {
			current := m.instForm[instFFrequency]
			idx := 0
			for i, f := range freqs {
				if f == current {
					idx = i
					break
				}
			}
			idx = (idx + 1) % len(freqs)
			m.instForm[instFFrequency] = freqs[idx]
		} else if m.instFormCur == instFAccount && len(m.accounts) > 0 {
			current := m.instForm[instFAccount]
			idx := -1
			for i, a := range m.accounts {
				if a.Name == current {
					idx = i
					break
				}
			}
			idx = (idx + 1) % len(m.accounts)
			m.instForm[instFAccount] = m.accounts[idx].Name
		} else {
			m.instForm[m.instFormCur] += " "
		}
	default:
		if m.instFormCur != instFFrequency && m.instFormCur != instFAccount {
			runes := []rune(key)
			if len(runes) == 1 {
				field := &m.instForm[m.instFormCur]
				if m.instFormCur == instFAmount {
					if strings.ContainsRune("0123456789.,", runes[0]) {
						*field += key
					}
				} else if m.instFormCur == instFTotalCount || m.instFormCur == instFPaidCount {
					if strings.ContainsRune("0123456789", runes[0]) {
						*field += key
					}
				} else if m.instFormCur == instFStartDate {
					if strings.ContainsRune("0123456789", runes[0]) {
						if len(*field) < 10 {
							*field += key
							if len(*field) == 2 || len(*field) == 5 {
								*field += "/"
							}
						}
					}
				} else {
					*field += key
				}
			}
		}
	}
	return m, nil
}

func (m *model) updateInstDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.instEditIdx >= 0 && m.instEditIdx < len(m.installments) {
			m.installments = append(m.installments[:m.instEditIdx], m.installments[m.instEditIdx+1:]...)
			_ = storage.SaveInstallments(m.dataDir, m.installments)
		}
		m.finView = fViewList
		if m.instCursor >= len(m.installments) && m.instCursor > 0 {
			m.instCursor--
		}
	case "n", "N", "esc":
		m.finView = fViewList
	}
	return m, nil
}

func calculateNextPaymentDate(inst storage.Installment) time.Time {
	if inst.StartDate.IsZero() {
		return time.Time{}
	}
	nextDate := inst.StartDate
	
	// Fast-forward based on PaidCount
	for i := 0; i < inst.PaidCount; i++ {
		switch inst.Frequency {
		case "type.monthly":
			nextDate = nextDate.AddDate(0, 1, 0)
		case "type.bimonthly":
			nextDate = nextDate.AddDate(0, 2, 0)
		case "type.quarterly":
			nextDate = nextDate.AddDate(0, 3, 0)
		case "type.semiannual":
			nextDate = nextDate.AddDate(0, 6, 0)
		case "type.annual":
			nextDate = nextDate.AddDate(1, 0, 0)
		default:
			nextDate = nextDate.AddDate(0, 1, 0)
		}
	}
	
	return nextDate
}

func (m *model) renderInstallments(s *styles) string {
	switch m.finView {
	case fViewAdd:
		return m.renderInstForm(s, t(m.lang, "action.add")+" "+t(m.lang, "finances.installments"))
	case fViewEdit:
		return m.renderInstForm(s, t(m.lang, "action.edit")+" "+t(m.lang, "finances.installments"))
	case fViewDelete:
		return m.renderInstDelete(s)
	default:
		return m.renderInstList(s)
	}
}

func (m *model) renderInstList(s *styles) string {
	isActive := m.finSection != fSectionMenu && m.focusContent
	title := s.title.Render(t(m.lang, "finances.installments"))
	
	if len(m.installments) == 0 {
		empty := s.dim.Render(t(m.lang, "installments.noRecords"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	headerStr := fmt.Sprintf("  %-20s %-12s %-12s %-15s %-15s %s", 
		t(m.lang, "col.description"), t(m.lang, "col.amount"), t(m.lang, "col.frequency"), t(m.lang, "col.progress"), t(m.lang, "col.nextPayment"), truncate(t(m.lang, "col.account"), 10))
	header := s.subtitle.Render(headerStr)
	divider := s.dim.Render("  " + strings.Repeat("─", 78))

	var rows []string
	
	for i, inst := range m.installments {
		amt := parseEuro(inst.Amount)
		
		freqStr := truncate(t(m.lang, inst.Frequency), 11)
		
		progressStr := ""
		if inst.TotalCount > 0 {
			progressStr = fmt.Sprintf("%d / %d", inst.PaidCount, inst.TotalCount)
		} else {
			progressStr = fmt.Sprintf("%d (%s)", inst.PaidCount, truncate(t(m.lang, "type.indefinite"), 8))
		}
		
		nextDate := calculateNextPaymentDate(inst)
		nextDateStr := ""
		if !nextDate.IsZero() {
			if inst.TotalCount > 0 && inst.PaidCount >= inst.TotalCount {
				nextDateStr = "COMPLETED"
			} else {
				nextDateStr = nextDate.Format("02/01/2006")
			}
		}

		accountName := inst.Account
		if accountName == "" {
			accountName = "-"
		}

		row := fmt.Sprintf("  %-20s € %-10.2f %-12s %-15s %-15s %s", truncate(inst.Name, 19), amt, freqStr, progressStr, nextDateStr, truncate(accountName, 10))
		
		if i == m.instCursor {
			if isActive {
				row = s.menuSelected.Width(0).Render(row)
			} else {
				row = s.menuActiveDim.Width(0).Render(row)
			}
		} else {
			if inst.TotalCount > 0 && inst.PaidCount >= inst.TotalCount {
				row = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(row)
			} else {
				row = s.info.Render(row)
			}
		}
		rows = append(rows, row)
	}

	table := strings.Join(rows, "\n")
	help := s.dim.Render(fmt.Sprintf("\n\na: %s • e: %s • d: %s • Space: Paga Rata • Back: Annulla Rata • ←: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))

	return title + "\n" + header + "\n" + divider + "\n" + table + help
}

func (m *model) renderInstForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	fields := []string{
		t(m.lang, "col.description"),
		t(m.lang, "col.amount") + " (€)",
		"Tot. Rate (0=infinito)",
		"Rate Pagate",
		t(m.lang, "col.frequency"),
		t(m.lang, "col.date") + " Inizio",
		t(m.lang, "col.account"),
	}

	var renderedLines []string
	for i, f := range fields {
		val := m.instForm[i]
		if i == instFFrequency {
			val = t(m.lang, val)
		}

		if i == m.instFormCur {
			if i == instFFrequency || i == instFAccount {
				renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-22s ", f+":"))+s.fieldBg.Render("< "+val+" >"))
			} else {
				renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-22s ", f+":"))+s.fieldBg.Render(val+s.highlight.Render("_")))
			}
		} else {
			renderedLines = append(renderedLines, s.dim.Render(fmt.Sprintf("  %-22s ", f+":"))+s.info.Render(val))
		}
	}

	rendered := strings.Join(renderedLines, "\n\n")

	help := s.dim.Render(fmt.Sprintf("\n\nEnter: %s  Esc: %s", t(m.lang, "action.save"), t(m.lang, "action.cancel")))

	return title + "\n\n" + rendered + help
}

func (m *model) renderInstDelete(s *styles) string {
	if m.instEditIdx < 0 || m.instEditIdx >= len(m.installments) {
		return m.renderInstList(s)
	}
	inst := m.installments[m.instEditIdx]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.installments"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("Delete this installment/recurring expense? (y/N)")

	info := fmt.Sprintf(
		"\n  %s %s\n  %s € %s",
		s.dim.Render(t(m.lang, "col.description")+":"), s.info.Render(inst.Name),
		s.dim.Render(t(m.lang, "col.amount")+":"), s.info.Render(inst.Amount),
	)

	help := s.dim.Render(fmt.Sprintf("\n\ny/s: %s  n/Esc: %s",
		t(m.lang, "action.delete"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + "\n" + info + help
}
