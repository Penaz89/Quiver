package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func (m *model) getFilteredVaultSecrets() []int {
	var indices []int
	search := strings.ToLower(m.vaultSearch)
	for i, sec := range m.vaultSecrets {
		if search == "" || strings.Contains(strings.ToLower(sec.Title), search) || strings.Contains(strings.ToLower(sec.Username), search) || strings.Contains(strings.ToLower(sec.Notes), search) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m *model) updateVault(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 1. Password input screen
	if !m.vaultUnlocked {
		switch msg.String() {
		case "esc":
			m.focusContent = false
			return m, nil
		case "enter":
			if m.vaultPwdForm == "" {
				return m, nil
			}
			secrets, err := storage.OpenVault(m.dataDir, m.vaultPwdForm)
			if err != nil {
				m.vaultPwdError = err.Error()
				if !storage.VaultExists(m.dataDir) {
					// It's a new vault, so the first password entered becomes master pwd
					m.vaultMasterPwd = m.vaultPwdForm
					m.vaultSecrets = []storage.Secret{}
					m.vaultUnlocked = true
					m.vaultExists = true
					m.vaultPwdError = ""
					storage.SaveVault(m.dataDir, m.vaultMasterPwd, m.vaultSecrets)
				}
				return m, nil
			}
			// Success
			m.vaultMasterPwd = m.vaultPwdForm
			m.vaultSecrets = secrets
			m.vaultUnlocked = true
			m.vaultExists = true
			m.vaultPwdError = ""
			m.vaultPwdForm = ""
			return m, nil
		case "backspace":
			if len(m.vaultPwdForm) > 0 {
				m.vaultPwdForm = m.vaultPwdForm[:len(m.vaultPwdForm)-1]
				m.vaultPwdError = ""
			}
		default:
			if len(msg.String()) == 1 {
				m.vaultPwdForm += msg.String()
				m.vaultPwdError = ""
			} else if msg.String() == "space" {
				m.vaultPwdForm += " "
				m.vaultPwdError = ""
			}
		}
		return m, nil
	}

	// 2. Add / Edit form
	if m.vaultIsAdding || m.vaultIsEditing {
		switch msg.String() {
		case "esc":
			m.vaultIsAdding = false
			m.vaultIsEditing = false
			return m, nil
		case "tab", "down":
			m.vaultFormCursor = (m.vaultFormCursor + 1) % 4
		case "shift+tab", "up":
			m.vaultFormCursor--
			if m.vaultFormCursor < 0 {
				m.vaultFormCursor = 3
			}
		case "enter":
			if m.vaultFormCursor == 3 {
				// Save
				sec := storage.Secret{
					ID:       fmt.Sprintf("%d", time.Now().UnixNano()),
					Title:    m.vaultForm[0],
					Username: m.vaultForm[1],
					Password: m.vaultForm[2],
					Notes:    m.vaultForm[3],
				}
				if m.vaultIsAdding {
					m.vaultSecrets = append(m.vaultSecrets, sec)
				} else if m.vaultIsEditing {
					sec.ID = m.vaultSecrets[m.vaultEditIndex].ID
					m.vaultSecrets[m.vaultEditIndex] = sec
				}
				err := storage.SaveVault(m.dataDir, m.vaultMasterPwd, m.vaultSecrets)
				if err != nil {
					// Handle error? Let's just ignore or log
				}
				m.vaultIsAdding = false
				m.vaultIsEditing = false
			} else {
				m.vaultFormCursor++
			}
		case "backspace":
			if len(m.vaultForm[m.vaultFormCursor]) > 0 {
				m.vaultForm[m.vaultFormCursor] = m.vaultForm[m.vaultFormCursor][:len(m.vaultForm[m.vaultFormCursor])-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.vaultForm[m.vaultFormCursor] += msg.String()
			} else if msg.String() == "space" {
				m.vaultForm[m.vaultFormCursor] += " "
			}
		}
		return m, nil
	}

	if m.vaultIsDeleting {
		switch msg.String() {
		case "y", "Y", "s", "S":
			m.vaultSecrets = append(m.vaultSecrets[:m.vaultEditIndex], m.vaultSecrets[m.vaultEditIndex+1:]...)
			storage.SaveVault(m.dataDir, m.vaultMasterPwd, m.vaultSecrets)
			m.vaultIsDeleting = false
			if m.vaultCursor >= len(m.vaultSecrets) && m.vaultCursor > 0 {
				m.vaultCursor--
			}
		case "n", "N", "esc":
			m.vaultIsDeleting = false
		}
		return m, nil
	}

	filtered := m.getFilteredVaultSecrets()
	if m.vaultCursor >= len(filtered) {
		m.vaultCursor = len(filtered) - 1
	}
	if m.vaultCursor < 0 {
		m.vaultCursor = 0
	}

	// 4. Search input
	if m.vaultIsSearching {
		switch msg.String() {
		case "esc", "enter":
			m.vaultIsSearching = false
			m.vaultCursor = 0
		case "backspace":
			if len(m.vaultSearch) > 0 {
				runes := []rune(m.vaultSearch)
				m.vaultSearch = string(runes[:len(runes)-1])
			}
			m.vaultCursor = 0
		default:
			key := msg.String()
			if key == "space" {
				key = " "
			}
			runes := []rune(key)
			if len(runes) == 1 { // naive check for printable
				m.vaultSearch += key
			}
			m.vaultCursor = 0
		}
		return m, nil
	}

	// 5. List navigation
	switch msg.String() {
	case "esc", "left":
		m.focusContent = false
	case "j", "down":
		if m.vaultCursor < len(filtered)-1 {
			m.vaultCursor++
		}
	case "k", "up":
		if m.vaultCursor > 0 {
			m.vaultCursor--
		}
	case "/":
		m.vaultIsSearching = true
	case "n":
		m.vaultIsAdding = true
		m.vaultForm = [4]string{}
		m.vaultFormCursor = 0
	case "enter":
		if len(filtered) > 0 {
			m.vaultIsEditing = true
			m.vaultEditIndex = filtered[m.vaultCursor]
			sec := m.vaultSecrets[m.vaultEditIndex]
			m.vaultForm = [4]string{sec.Title, sec.Username, sec.Password, sec.Notes}
			m.vaultFormCursor = 0
		}
	case "d", "delete":
		if len(filtered) > 0 {
			m.vaultEditIndex = filtered[m.vaultCursor]
			m.vaultIsDeleting = true
		}
	case "L": // lock
		m.vaultUnlocked = false
		m.vaultMasterPwd = ""
		m.vaultSecrets = nil
		m.vaultPwdForm = ""
	}

	return m, nil
}

func (m *model) renderVaultView(s *styles) string {
	if !m.vaultUnlocked {
		title := s.title.Render(t(m.lang, "vault.title"))
		prompt := t(m.lang, "vault.enterMasterPwd")
		if !storage.VaultExists(m.dataDir) {
			prompt = t(m.lang, "vault.createMasterPwd")
		}

		pwdMask := strings.Repeat("*", len(m.vaultPwdForm))
		pwdField := s.highlight.Render(pwdMask)
		if len(pwdMask) == 0 {
			pwdField = s.dim.Render("...")
		}
		
		errStr := ""
		if m.vaultPwdError != "" {
			errStr = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.vaultPwdError) + "\n"
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 4).
			Render(fmt.Sprintf("%s\n\n%s\n> %s█", prompt, errStr, pwdField))

		return title + "\n\n" + box
	}

	title := s.title.Render(t(m.lang, "vault.title"))
	subtitle := s.subtitle.Render(t(m.lang, "vault.subtitle"))
	header := lipgloss.JoinVertical(lipgloss.Left, title, subtitle)

	if m.vaultIsAdding || m.vaultIsEditing {
		action := t(m.lang, "vault.add")
		if m.vaultIsEditing {
			action = t(m.lang, "vault.edit")
		}
		
		labels := []string{
			t(m.lang, "vault.fieldTitle"),
			t(m.lang, "vault.fieldUsername"),
			t(m.lang, "vault.fieldPassword"),
			t(m.lang, "vault.fieldNotes"),
		}
		
		var lines []string
		lines = append(lines, s.highlight.Render(action))
		lines = append(lines, "")
		
		for i, label := range labels {
			val := m.vaultForm[i]
			if i == m.vaultFormCursor {
				lines = append(lines, s.menuSelected.Render(fmt.Sprintf("> %-15s: %s█", label, val)))
			} else {
				lines = append(lines, s.info.Render(fmt.Sprintf("  %-15s: %s", label, val)))
			}
		}
		lines = append(lines, "\n"+s.dim.Render("tab: next • enter: save • esc: cancel"))
		
		return header + "\n\n" + strings.Join(lines, "\n")
	}

	if m.vaultIsDeleting {
		msg := s.highlight.Render(t(m.lang, "vault.confirmDelete"))
		sec := m.vaultSecrets[m.vaultEditIndex]
		info := s.info.Render(sec.Title)
		return header + "\n\n" + msg + "\n" + info
	}

	filtered := m.getFilteredVaultSecrets()

	var searchBar string
	if m.vaultIsSearching || m.vaultSearch != "" {
		prompt := "Search: "
		if m.vaultIsSearching {
			searchBar = s.highlight.Render(prompt + m.vaultSearch + "█")
		} else {
			searchBar = s.dim.Render(prompt + m.vaultSearch)
		}
		searchBar += "\n\n"
	}

	if len(filtered) == 0 {
		return header + "\n\n" + searchBar + s.dim.Render(t(m.lang, "vault.noSecrets"))
	}

	// Calculate max widths for nice columns
	maxTitle := 15
	maxUser := 15
	for _, idx := range filtered {
		sec := m.vaultSecrets[idx]
		if len(sec.Title) > maxTitle {
			maxTitle = len(sec.Title)
		}
		if len(sec.Username) > maxUser {
			maxUser = len(sec.Username)
		}
	}

	var lines []string
	head := s.dim.Render(fmt.Sprintf("  %-*s │ %-*s", maxTitle, t(m.lang, "vault.colTitle"), maxUser, t(m.lang, "vault.colUsername")))
	lines = append(lines, head)
	lines = append(lines, s.dim.Render(strings.Repeat("─", maxTitle+maxUser+5)))

	for i, idx := range filtered {
		sec := m.vaultSecrets[idx]
		row := fmt.Sprintf("%-*s │ %-*s", maxTitle, sec.Title, maxUser, sec.Username)
		if i == m.vaultCursor {
			if m.focusContent {
				lines = append(lines, s.menuSelected.Render("> "+row))
			} else {
				lines = append(lines, s.menuActiveDim.Render("> "+row))
			}
		} else {
			lines = append(lines, s.info.Render("  "+row))
		}
	}
	
	help := s.dim.Render(t(m.lang, "vault.help"))
	if m.vaultIsSearching {
		help = s.dim.Render(t(m.lang, "vault.helpSearchActive"))
	}
	
	// Show details if selected
	details := ""
	if len(filtered) > 0 && m.vaultCursor >= 0 && m.vaultCursor < len(filtered) {
		sec := m.vaultSecrets[filtered[m.vaultCursor]]
		detailsBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			Render(fmt.Sprintf("%s: %s\n%s: %s\n%s: %s\n%s: %s", 
				t(m.lang, "vault.fieldTitle"), sec.Title,
				t(m.lang, "vault.fieldUsername"), sec.Username,
				t(m.lang, "vault.fieldPassword"), sec.Password,
				t(m.lang, "vault.fieldNotes"), sec.Notes))
		details = "\n\n" + detailsBox
	}

	return header + "\n\n" + searchBar + strings.Join(lines, "\n") + details + "\n\n" + help
}
