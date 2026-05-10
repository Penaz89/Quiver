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

const (
	trFDate = iota
	trFFrom
	trFTo
	trFAmount
	trFDescription
	trFFrequency
	trFCount
)

func (m *model) updateTransfers(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch m.finView {
	case fViewList:
		switch key {
		case "esc", "left":
			m.finSection = fSectionMenu
		case "up", "k":
			if m.trCursor > 0 {
				m.trCursor--
			}
		case "down", "j":
			if m.trCursor < len(m.transfers)-1 {
				m.trCursor++
			}
		case "a":
			m.finView = fViewAdd
			m.trForm = [trFCount]string{time.Now().Format("02/01/2006"), "", "", "", "", "type.none"}
			if len(m.accounts) > 0 {
				m.trForm[trFFrom] = m.accounts[0].Name
				if len(m.accounts) > 1 {
					m.trForm[trFTo] = m.accounts[1].Name
				} else {
					m.trForm[trFTo] = m.accounts[0].Name
				}
			}
			m.trFormCur = 0
		case "e", "enter":
			if len(m.transfers) > 0 {
				m.finView = fViewEdit
				m.trEditIdx = m.trCursor
				t := m.transfers[m.trCursor]
				dateStr := t.Date.Format("02/01/2006")
				if t.Date.IsZero() {
					dateStr = ""
				}
				freq := t.Frequency
				if freq == "" {
					freq = "type.none"
				}
				m.trForm = [trFCount]string{dateStr, t.FromAccount, t.ToAccount, t.Amount, t.Description, freq}
				m.trFormCur = 0
			}
		case "d", "del", "x":
			if len(m.transfers) > 0 {
				m.finView = fViewDelete
				m.trEditIdx = m.trCursor
			}
		}

	case fViewAdd, fViewEdit:
		switch key {
		case "esc":
			m.finView = fViewList
		case "tab", "down":
			m.trFormCur = (m.trFormCur + 1) % trFCount
		case "shift+tab", "up":
			m.trFormCur = (m.trFormCur - 1 + trFCount) % trFCount
		case "left", "h", "right", "l":
			if (m.trFormCur == trFFrom || m.trFormCur == trFTo) && len(m.accounts) > 0 {
				current := m.trForm[m.trFormCur]
				idx := -1
				for i, a := range m.accounts {
					if a.Name == current {
						idx = i
						break
					}
				}
				if key == "right" || key == "l" {
					idx = (idx + 1) % len(m.accounts)
				} else {
					if idx == -1 {
						idx = 0
					}
					idx = (idx - 1 + len(m.accounts)) % len(m.accounts)
				}
				m.trForm[m.trFormCur] = m.accounts[idx].Name
			} else if m.trFormCur == trFFrequency {
				freqs := []string{"type.none", "type.monthly"}
				current := m.trForm[trFFrequency]
				idx := 0
				for i, f := range freqs {
					if f == current {
						idx = i
						break
					}
				}
				if key == "right" || key == "l" {
					idx = (idx + 1) % len(freqs)
				} else {
					idx = (idx - 1 + len(freqs)) % len(freqs)
				}
				m.trForm[trFFrequency] = freqs[idx]
			}
		case "enter":
			dateStr := strings.TrimSpace(m.trForm[trFDate])
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

			amt := strings.TrimSpace(m.trForm[trFAmount])
			from := m.trForm[trFFrom]
			to := m.trForm[trFTo]
			
			if amt == "" || from == "" || to == "" || from == to {
				m.finView = fViewList
				return m, nil
			}

			freq := m.trForm[trFFrequency]
			if freq == "" {
				freq = "type.none"
			}
			var nextDate time.Time
			if freq == "type.monthly" {
				nextDate = parsedDate.AddDate(0, 1, 0)
			}

			tRec := storage.Transfer{
				Date:        parsedDate,
				FromAccount: from,
				ToAccount:   to,
				Amount:      amt,
				Description: strings.TrimSpace(m.trForm[trFDescription]),
				Author:      m.user,
				Frequency:   freq,
				NextDate:    nextDate,
			}

			amtEuro := parseEuro(amt)

			if m.finView == fViewAdd {
				m.transfers = append(m.transfers, tRec)
				m.adjustAccountBalance(from, -amtEuro)
				m.adjustAccountBalance(to, amtEuro)
			} else {
				oldFrom := m.transfers[m.trEditIdx].FromAccount
				oldTo := m.transfers[m.trEditIdx].ToAccount
				oldAmt := parseEuro(m.transfers[m.trEditIdx].Amount)

				// Revert old
				if oldFrom != "" {
					m.adjustAccountBalance(oldFrom, oldAmt)
				}
				if oldTo != "" {
					m.adjustAccountBalance(oldTo, -oldAmt)
				}

				m.transfers[m.trEditIdx] = tRec

				// Apply new
				m.adjustAccountBalance(from, -amtEuro)
				m.adjustAccountBalance(to, amtEuro)
			}

			// Sort by date descending
			sort.Slice(m.transfers, func(i, j int) bool {
				return m.transfers[i].Date.After(m.transfers[j].Date)
			})
			_ = storage.SaveTransfers(m.dataDir, m.transfers)
			m.finView = fViewList

		case "backspace":
			if m.trFormCur != trFFrom && m.trFormCur != trFTo && m.trFormCur != trFFrequency {
				field := &m.trForm[m.trFormCur]
				if len(*field) > 0 {
					runes := []rune(*field)
					*field = string(runes[:len(runes)-1])
				}
			}
			if m.trFormCur != trFFrom && m.trFormCur != trFTo && m.trFormCur != trFFrequency {
				runes := []rune(key)
				if len(runes) == 1 {
					field := &m.trForm[m.trFormCur]
					if m.trFormCur == trFDate {
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
					} else if m.trFormCur == trFAmount {
						if !strings.ContainsRune("0123456789.,", runes[0]) {
							return m, nil
						}
						*field += key
					} else {
						if key == "space" {
							key = " "
						}
						*field += key
					}
				}
			} else if (m.trFormCur == trFFrom || m.trFormCur == trFTo) && len(m.accounts) > 0 && key == "space" {
				current := m.trForm[m.trFormCur]
				idx := -1
				for i, a := range m.accounts {
					if a.Name == current {
						idx = i
						break
					}
				}
				idx = (idx + 1) % len(m.accounts)
				m.trForm[m.trFormCur] = m.accounts[idx].Name
			} else if m.trFormCur == trFFrequency && key == "space" {
				freqs := []string{"type.none", "type.monthly"}
				current := m.trForm[trFFrequency]
				idx := 0
				for i, f := range freqs {
					if f == current {
						idx = i
						break
					}
				}
				idx = (idx + 1) % len(freqs)
				m.trForm[trFFrequency] = freqs[idx]
			}
		}

	case fViewDelete:
		switch key {
		case "y", "Y", "s", "S", "enter":
			if m.trEditIdx >= 0 && m.trEditIdx < len(m.transfers) {
				tRec := m.transfers[m.trEditIdx]
				amt := parseEuro(tRec.Amount)
				if tRec.FromAccount != "" {
					m.adjustAccountBalance(tRec.FromAccount, amt)
				}
				if tRec.ToAccount != "" {
					m.adjustAccountBalance(tRec.ToAccount, -amt)
				}

				m.transfers = append(m.transfers[:m.trEditIdx], m.transfers[m.trEditIdx+1:]...)
				_ = storage.SaveTransfers(m.dataDir, m.transfers)
			}
			m.finView = fViewList
			if m.trCursor >= len(m.transfers) && m.trCursor > 0 {
				m.trCursor--
			}
		case "n", "N", "esc":
			m.finView = fViewList
		}
	}
	return m, nil
}

func (m *model) renderTransfers(s *styles) string {
	switch m.finView {
	case fViewAdd:
		return m.renderTransfersForm(s, t(m.lang, "action.add")+" "+t(m.lang, "finances.transfers"))
	case fViewEdit:
		return m.renderTransfersForm(s, t(m.lang, "action.edit")+" "+t(m.lang, "finances.transfers"))
	case fViewDelete:
		return m.renderTransfersDelete(s)
	default:
		return m.renderTransfersList(s)
	}
}

func (m *model) renderTransfersList(s *styles) string {
	isActive := m.finSection != fSectionMenu && m.focusContent
	title := s.title.Render(t(m.lang, "finances.transfers"))

	if len(m.transfers) == 0 {
		empty := s.dim.Render(t(m.lang, "transfers.noRecords"))
		help := s.dim.Render(fmt.Sprintf("\n\na: %s  ←: %s", t(m.lang, "action.add"), t(m.lang, "help.goBack")))
		return title + "\n\n" + empty + help
	}

	headerStr := fmt.Sprintf("  %-12s %-15s %-15s %-10s %-20s",
		t(m.lang, "col.date"), t(m.lang, "col.fromAccount"), t(m.lang, "col.toAccount"), t(m.lang, "col.amount"), t(m.lang, "field.description"))
	header := s.subtitle.Render(headerStr)
	divider := s.dim.Render("  " + strings.Repeat("─", 80))

	var rows []string

	for i, tRec := range m.transfers {
		dateStr := tRec.Date.Format("02/01/2006")
		amtStr := fmt.Sprintf("€ %.2f", parseEuro(tRec.Amount))

		desc := tRec.Description
		if tRec.Frequency == "type.monthly" {
			desc = "↺ " + desc
		}
		row := fmt.Sprintf("  %-12s %-15s %-15s %-10s %-20s",
			dateStr, truncate(tRec.FromAccount, 14), truncate(tRec.ToAccount, 14), amtStr, truncate(desc, 19))

		if i == m.trCursor {
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
	
	help := s.dim.Render(fmt.Sprintf("\n\na: %s • e: %s • d: %s • ←: %s",
		t(m.lang, "action.add"), t(m.lang, "action.edit"), t(m.lang, "action.delete"), t(m.lang, "help.goBack")))

	return title + "\n" + header + "\n" + divider + "\n" + table + help
}

func (m *model) renderTransfersForm(s *styles, formTitle string) string {
	title := s.title.Render(formTitle)

	fields := []string{
		t(m.lang, "field.date"),
		t(m.lang, "col.fromAccount"),
		t(m.lang, "col.toAccount"),
		t(m.lang, "field.amount") + " (€)",
		t(m.lang, "field.description"),
		t(m.lang, "field.frequency"),
	}

	var renderedLines []string
	for i, f := range fields {
		val := m.trForm[i]
		
		displayVal := val
		if i == trFFrequency {
			displayVal = t(m.lang, val)
		}
		
		if i == m.trFormCur {
			if i == trFFrom || i == trFTo || i == trFFrequency {
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

func (m *model) renderTransfersDelete(s *styles) string {
	if m.trEditIdx < 0 || m.trEditIdx >= len(m.transfers) {
		return m.renderTransfersList(s)
	}
	tRec := m.transfers[m.trEditIdx]

	title := s.title.Render(t(m.lang, "action.delete") + " " + t(m.lang, "finances.transfers"))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(t(m.lang, "delete.confirmInsurance"))

	info := fmt.Sprintf(
		"\n  %s %s\n  %s %s\n  %s %s\n  %s € %s",
		s.dim.Render(t(m.lang, "col.date")+":"), s.info.Render(tRec.Date.Format("02/01/2006")),
		s.dim.Render(t(m.lang, "col.fromAccount")+":"), s.info.Render(tRec.FromAccount),
		s.dim.Render(t(m.lang, "col.toAccount")+":"), s.info.Render(tRec.ToAccount),
		s.dim.Render(t(m.lang, "col.amount")+":"), s.info.Render(tRec.Amount),
	)

	help := s.dim.Render(fmt.Sprintf("\n\ny/s: %s  n/Esc: %s",
		t(m.lang, "action.delete"), t(m.lang, "action.cancel")))

	return title + "\n\n" + warning + "\n" + info + help
}

func (m *model) processRecurringTransfers() {
	now := time.Now()
	modified := false
	for i, tRec := range m.transfers {
		if tRec.Frequency == "type.monthly" && !tRec.NextDate.IsZero() {
			for !tRec.NextDate.After(now) {
				// execute transfer
				amt := parseEuro(tRec.Amount)
				m.adjustAccountBalance(tRec.FromAccount, -amt)
				m.adjustAccountBalance(tRec.ToAccount, amt)

				// record execution as a new history item (non-recurring)
				newTr := storage.Transfer{
					Date:        tRec.NextDate,
					FromAccount: tRec.FromAccount,
					ToAccount:   tRec.ToAccount,
					Amount:      tRec.Amount,
					Description: tRec.Description + " (Auto)",
					Author:      "System",
					Frequency:   "type.none",
				}
				m.transfers = append(m.transfers, newTr)

				// advance next date
				tRec.NextDate = tRec.NextDate.AddDate(0, 1, 0)
				modified = true
			}
			m.transfers[i] = tRec
		}
	}

	if modified {
		sort.Slice(m.transfers, func(i, j int) bool {
			return m.transfers[i].Date.After(m.transfers[j].Date)
		})
		_ = storage.SaveTransfers(m.dataDir, m.transfers)
	}
}
