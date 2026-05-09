package tui

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func (m *model) updateAdminUsers(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.adminIsAdding || m.adminIsEditing {
		return m.updateAdminForm(msg)
	}
	if m.adminIsDeleting {
		return m.updateAdminDelete(msg)
	}

	key := msg.String()
	switch key {
	case "up", "k":
		if m.adminUserCursor > 0 {
			m.adminUserCursor--
		}
	case "down", "j":
		if len(m.adminUsers) > 0 && m.adminUserCursor < len(m.adminUsers)-1 {
			m.adminUserCursor++
		}
	case "n", "N":
		m.adminIsAdding = true
		m.adminIsEditing = false
		m.adminForm = [2]string{}
		m.adminFormCursor = 0
		m.adminError = ""
	case "enter", "e":
		if len(m.adminUsers) > 0 {
			m.adminIsAdding = false
			m.adminIsEditing = true
			u := m.adminUsers[m.adminUserCursor]
			m.adminForm = [2]string{u.Username, ""}
			m.adminFormCursor = 1 // Start at password since username is fixed
			m.adminError = ""
		}
	case "delete", "backspace":
		if len(m.adminUsers) > 0 {
			u := m.adminUsers[m.adminUserCursor]
			if u.Username == "admin" {
				m.adminError = "Cannot delete the default admin user."
			} else {
				m.adminIsDeleting = true
				m.adminError = ""
			}
		}
	case "esc", "left", "h":
		m.focusContent = false
	}
	return m, nil
}

func (m *model) updateAdminDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "y", "Y", "s", "S":
		u := m.adminUsers[m.adminUserCursor]
		storage.DeleteUser(m.baseDataDir, u.Username)
		m.adminUsers, _ = storage.LoadUsers(m.baseDataDir)
		if m.adminUserCursor >= len(m.adminUsers) {
			m.adminUserCursor = len(m.adminUsers) - 1
		}
		if m.adminUserCursor < 0 {
			m.adminUserCursor = 0
		}
		m.adminError = "User deleted successfully."
		m.adminIsDeleting = false
	case "n", "N", "esc":
		m.adminIsDeleting = false
		m.adminError = ""
	}
	return m, nil
}

func (m *model) updateAdminResetVault(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "y", "Y", "s", "S":
		if m.adminUserCursor < len(m.adminUsers) {
			u := m.adminUsers[m.adminUserCursor]
			dataDir := storage.GetUserDir(m.baseDataDir, u.Username)
			storage.DeleteVault(dataDir)
		} else {
			idx := m.adminUserCursor - len(m.adminUsers)
			f := m.adminFamilies[idx]
			dataDir := storage.GetFamilyDir(m.baseDataDir, f.ID)
			storage.DeleteVault(dataDir)
		}
		m.adminError = "Vault deleted successfully."
		m.adminIsResettingVault = false
	case "n", "N", "esc":
		m.adminIsResettingVault = false
		m.adminError = ""
	}
	return m, nil
}

func (m *model) updateAdminForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		if m.adminIsEditing {
			m.adminFormCursor = 1
		} else {
			m.adminFormCursor = (m.adminFormCursor + 1) % 2
		}
	case "shift+tab", "up":
		if m.adminIsEditing {
			m.adminFormCursor = 1
		} else {
			m.adminFormCursor = (m.adminFormCursor - 1 + 2) % 2
		}
	case "esc":
		m.adminIsAdding = false
		m.adminIsEditing = false
		m.adminError = ""
	case "enter":
		user := strings.TrimSpace(m.adminForm[0])
		pass := m.adminForm[1]

		if user == "" {
			m.adminError = "Username cannot be empty."
			return m, nil
		}

		if m.adminIsAdding {
			if pass == "" {
				m.adminError = "Password cannot be empty."
				return m, nil
			}
			err := storage.CreateUser(m.baseDataDir, user, pass)
			if err != nil {
				m.adminError = "Error: " + err.Error()
			} else {
				m.adminIsAdding = false
				m.adminUsers, _ = storage.LoadUsers(m.baseDataDir)
				m.adminError = "User created successfully."
			}
		} else if m.adminIsEditing {
			if pass == "" {
				m.adminError = "Password cannot be empty."
				return m, nil
			}
			err := storage.UpdateUserPassword(m.baseDataDir, user, pass, false)
			if err != nil {
				m.adminError = "Error: " + err.Error()
			} else {
				m.adminIsEditing = false
				m.adminUsers, _ = storage.LoadUsers(m.baseDataDir)
				m.adminError = "User updated successfully."
			}
		}
	case "backspace":
		field := &m.adminForm[m.adminFormCursor]
		if len(*field) > 0 {
			runes := []rune(*field)
			*field = string(runes[:len(runes)-1])
		}
	default:
		if key == "space" {
			key = " "
		}
		runes := []rune(key)
		if len(runes) == 1 && unicode.IsPrint(runes[0]) {
			m.adminForm[m.adminFormCursor] += key
		}
	}
	return m, nil
}

func (m *model) renderAdminUsersView(s *styles) string {
	if m.adminUsers == nil {
		m.adminUsers, _ = storage.LoadUsers(m.baseDataDir)
	}

	if m.adminIsAdding || m.adminIsEditing {
		return m.renderAdminForm(s)
	}
	if m.adminIsDeleting {
		return m.renderAdminDelete(s)
	}

	title := s.title.Render("USER MANAGEMENT")
	
	var list string
	if len(m.adminUsers) == 0 {
		list = s.dim.Render("No users found.")
	} else {
		for i, u := range m.adminUsers {
			cursor := "  "
			rowStyle := s.info
			if i == m.adminUserCursor {
				cursor = s.highlight.Render("> ")
				rowStyle = s.highlight
			}
			
			roleTag := ""
			if u.Role == "admin" {
				roleTag = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(" [ADMIN]")
			}
			
			statusTag := ""
			if u.MustChange {
				statusTag = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(" (Must Change Pwd)")
			}

			list += fmt.Sprintf("%s%s%s%s\n", cursor, rowStyle.Render(u.Username), roleTag, statusTag)
		}
	}

	content := title + "\n\n" + list
	
	if m.adminError != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		if strings.HasSuffix(m.adminError, "successfully.") {
			errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		}
		content += "\n\n" + errStyle.Render(m.adminError)
	}

	help := s.dim.Render("\n\nN: New User • Enter: Edit Password • Del: Delete User • Esc: Back")
	return content + help
}

func (m *model) renderAdminForm(s *styles) string {
	titleText := "ADD NEW USER"
	if m.adminIsEditing {
		titleText = "EDIT USER PASSWORD"
	}
	title := s.title.Render(titleText)

	labelStyle := lipgloss.NewStyle().Width(12).Align(lipgloss.Right).MarginRight(1).Foreground(lipgloss.Color("241"))
	inputStyle := lipgloss.NewStyle().Width(20).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).PaddingLeft(1)
	emptyStyle := lipgloss.NewStyle().Width(20).PaddingLeft(1).Foreground(lipgloss.Color("241"))
	filledStyle := lipgloss.NewStyle().Width(20).PaddingLeft(1).Foreground(lipgloss.Color("252"))
	dimFilledStyle := lipgloss.NewStyle().Width(20).PaddingLeft(1).Foreground(lipgloss.Color("241"))

	// Username field
	usrLabel := labelStyle.Render("Username:")
	var usrVal string
	if m.adminIsEditing {
		usrVal = dimFilledStyle.Render(m.adminForm[0] + " (Read-Only)")
	} else if m.adminFormCursor == 0 {
		usrVal = inputStyle.Render(m.adminForm[0] + s.highlight.Render("_"))
	} else {
		if m.adminForm[0] == "" {
			usrVal = emptyStyle.Render("(empty)")
		} else {
			usrVal = filledStyle.Render(m.adminForm[0])
		}
	}
	usrLine := lipgloss.JoinHorizontal(lipgloss.Top, usrLabel, usrVal)

	// Password field
	pwdLabel := labelStyle.Render("Password:")
	pwdValRaw := m.adminForm[1]
	pwdValMasked := strings.Repeat("*", len(pwdValRaw))
	var pwdVal string
	if m.adminFormCursor == 1 {
		pwdVal = inputStyle.Render(pwdValMasked + s.highlight.Render("_"))
	} else {
		if pwdValRaw == "" {
			pwdVal = emptyStyle.Render("(empty)")
		} else {
			pwdVal = filledStyle.Render(pwdValMasked)
		}
	}
	pwdLine := lipgloss.JoinHorizontal(lipgloss.Top, pwdLabel, pwdVal)

	form := usrLine + "\n\n" + pwdLine

	content := title + "\n\n" + form

	if m.adminError != "" {
		content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.adminError)
	}

	help := s.dim.Render("\n\nEnter: Save • Esc: Cancel")
	return content + help
}

func (m *model) updateAdminVault(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.adminIsResettingVault {
		return m.updateAdminResetVault(msg)
	}
	
	if m.adminUsers == nil {
		m.adminUsers, _ = storage.LoadUsers(m.baseDataDir)
	}
	if m.adminFamilies == nil {
		m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
	}
	
	maxCursor := len(m.adminUsers) + len(m.adminFamilies) - 1
	
	key := msg.String()
	switch key {
	case "up", "k":
		if m.adminUserCursor > 0 {
			m.adminUserCursor--
		}
	case "down", "j":
		if maxCursor >= 0 && m.adminUserCursor < maxCursor {
			m.adminUserCursor++
		}
	case "enter", "v", "V":
		if maxCursor >= 0 {
			m.adminIsResettingVault = true
			m.adminError = ""
		}
	case "esc", "left", "h":
		m.focusContent = false
	}
	return m, nil
}

func (m *model) renderAdminVaultView(s *styles) string {
	if m.adminUsers == nil {
		m.adminUsers, _ = storage.LoadUsers(m.baseDataDir)
	}

	if m.adminFamilies == nil {
		m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
	}

	if m.adminIsResettingVault {
		return m.renderAdminResetVault(s)
	}

	title := s.title.Render("VAULT MANAGEMENT")
	subtitle := s.subtitle.Render("Reset encrypted vaults")
	
	var list string
	maxCursor := len(m.adminUsers) + len(m.adminFamilies) - 1
	if maxCursor < 0 {
		list = s.dim.Render("No users or families found.")
	} else {
		if len(m.adminUsers) > 0 {
			list += s.dim.Render("--- USERS ---") + "\n"
			for i, u := range m.adminUsers {
				cursor := "  "
				rowStyle := s.info
				if i == m.adminUserCursor {
					cursor = s.highlight.Render("> ")
					rowStyle = s.highlight
				}
				
				roleTag := ""
				if u.Role == "admin" {
					roleTag = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(" [ADMIN]")
				}
				
				dataDir := storage.GetUserDir(m.baseDataDir, u.Username)
				vaultStatus := ""
				if storage.VaultExists(dataDir) {
					vaultStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(" (Vault Active)")
				} else if u.Role == "admin" {
					vaultStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(" (No Vault)")
				}

				list += fmt.Sprintf("%s%s%s%s\n", cursor, rowStyle.Render(u.Username), roleTag, vaultStatus)
			}
		}
		
		if len(m.adminFamilies) > 0 {
			if len(m.adminUsers) > 0 {
				list += "\n"
			}
			list += s.dim.Render("--- FAMILIES ---") + "\n"
			for j, f := range m.adminFamilies {
				i := j + len(m.adminUsers)
				cursor := "  "
				rowStyle := s.info
				if i == m.adminUserCursor {
					cursor = s.highlight.Render("> ")
					rowStyle = s.highlight
				}
				
				dataDir := storage.GetFamilyDir(m.baseDataDir, f.ID)
				vaultStatus := ""
				if storage.VaultExists(dataDir) {
					vaultStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(" (Vault Active)")
				}
				
				list += fmt.Sprintf("%s%s%s\n", cursor, rowStyle.Render(f.Name), vaultStatus)
			}
		}
	}

	content := title + "\n" + subtitle + "\n\n" + list
	
	if m.adminError != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		if strings.HasSuffix(m.adminError, "successfully.") {
			errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		}
		content += "\n\n" + errStyle.Render(m.adminError)
	}

	help := s.dim.Render("\n\nEnter: Reset Vault • Esc: Back")
	return content + help
}

func (m *model) renderAdminDelete(s *styles) string {
	title := s.title.Render("DELETE USER")
	
	u := m.adminUsers[m.adminUserCursor]
	
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	
	msg := msgStyle.Render("Are you sure you want to delete the user ") + userStyle.Render(u.Username) + msgStyle.Render("?")
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("\nThis action will also delete the user's data directory and cannot be undone.")
	
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render("\n\nPress (Y) to confirm or (N) to cancel.")
	
	content := title + "\n\n" + msg + warning + prompt
	
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *model) renderAdminResetVault(s *styles) string {
	title := s.title.Render("RESET VAULT")
	
	targetName := ""
	if m.adminUserCursor < len(m.adminUsers) {
		targetName = "user " + m.adminUsers[m.adminUserCursor].Username
	} else {
		idx := m.adminUserCursor - len(m.adminUsers)
		targetName = "family " + m.adminFamilies[idx].Name
	}
	
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	targetStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	
	msg := msgStyle.Render("Are you sure you want to reset the vault for ") + targetStyle.Render(targetName) + msgStyle.Render("?")
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("\nThis action will delete all their encrypted secrets permanently.")
	
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render("\n\nPress (Y) to confirm or (N) to cancel.")
	
	content := title + "\n\n" + msg + warning + prompt
	
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *model) updateAdminWorkspaces(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.adminFamilies == nil {
		m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
	}

	if m.adminFamilyIsEditing || m.adminFamilyIsInviting {
		return m.updateAdminWorkspaceForm(msg)
	}
	if m.adminFamilyIsDeleting {
		return m.updateAdminWorkspaceDelete(msg)
	}

	key := msg.String()
	switch key {
	case "up", "k":
		if m.adminFamilyCursor > 0 {
			m.adminFamilyCursor--
		}
	case "down", "j":
		if len(m.adminFamilies) > 0 && m.adminFamilyCursor < len(m.adminFamilies)-1 {
			m.adminFamilyCursor++
		}
	case "e": // rename
		if len(m.adminFamilies) > 0 {
			m.adminFamilyIsEditing = true
			m.adminFamilyForm = m.adminFamilies[m.adminFamilyCursor].Name
			m.adminFamilyError = ""
		}
	case "i": // invite member
		if len(m.adminFamilies) > 0 {
			m.adminFamilyIsInviting = true
			m.adminFamilyForm = ""
			m.adminFamilyError = ""
		}
	case "d", "delete", "backspace": // remove member or delete
		if len(m.adminFamilies) > 0 {
			m.adminFamilyIsDeleting = true
			m.adminFamilyForm = ""
			m.adminFamilyError = ""
		}
	case "esc", "left", "h":
		m.focusContent = false
	}
	return m, nil
}

func (m *model) updateAdminWorkspaceForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		m.adminFamilyIsEditing = false
		m.adminFamilyIsInviting = false
		m.adminFamilyError = ""
	case "enter":
		val := strings.TrimSpace(m.adminFamilyForm)
		if val == "" {
			m.adminFamilyError = "Field cannot be empty."
			return m, nil
		}
		
		f := m.adminFamilies[m.adminFamilyCursor]

		if m.adminFamilyIsEditing {
			err := storage.RenameFamily(m.baseDataDir, f.ID, val)
			if err != nil {
				m.adminFamilyError = "Error: " + err.Error()
			} else {
				m.adminFamilyIsEditing = false
				m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
				m.adminFamilyError = "Workspace renamed successfully."
			}
		} else if m.adminFamilyIsInviting {
			err := storage.AddMemberToFamily(m.baseDataDir, f.ID, val)
			if err != nil {
				m.adminFamilyError = "Error: " + err.Error()
			} else {
				m.adminFamilyIsInviting = false
				m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
				m.adminFamilyError = "User added successfully."
			}
		}
	case "backspace":
		if len(m.adminFamilyForm) > 0 {
			runes := []rune(m.adminFamilyForm)
			m.adminFamilyForm = string(runes[:len(runes)-1])
		}
	default:
		if key == "space" {
			key = " "
		}
		runes := []rune(key)
		if len(runes) == 1 && unicode.IsPrint(runes[0]) {
			m.adminFamilyForm += key
		}
	}
	return m, nil
}

func (m *model) updateAdminWorkspaceDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		m.adminFamilyIsDeleting = false
		m.adminFamilyError = ""
	case "enter":
		val := strings.TrimSpace(m.adminFamilyForm)
		if val == "" {
			m.adminFamilyError = "Type a username to remove, or 'ALL' to delete family."
			return m, nil
		}
		f := m.adminFamilies[m.adminFamilyCursor]
		if val == "ALL" {
			// To delete a family, we remove all members one by one.
			for _, mem := range f.Members {
				_ = storage.RemoveMemberFromFamily(m.baseDataDir, f.ID, mem)
			}
			m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
			m.adminFamilyIsDeleting = false
			m.adminFamilyError = "Workspace deleted successfully."
			if m.adminFamilyCursor >= len(m.adminFamilies) {
				m.adminFamilyCursor = len(m.adminFamilies) - 1
			}
			if m.adminFamilyCursor < 0 { m.adminFamilyCursor = 0 }
		} else {
			err := storage.RemoveMemberFromFamily(m.baseDataDir, f.ID, val)
			if err != nil {
				m.adminFamilyError = "Error: " + err.Error()
			} else {
				m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
				m.adminFamilyIsDeleting = false
				m.adminFamilyError = "Member removed successfully."
				if m.adminFamilyCursor >= len(m.adminFamilies) {
					m.adminFamilyCursor = len(m.adminFamilies) - 1
				}
				if m.adminFamilyCursor < 0 { m.adminFamilyCursor = 0 }
			}
		}
	case "backspace":
		if len(m.adminFamilyForm) > 0 {
			runes := []rune(m.adminFamilyForm)
			m.adminFamilyForm = string(runes[:len(runes)-1])
		}
	default:
		if key == "space" {
			key = " "
		}
		runes := []rune(key)
		if len(runes) == 1 && unicode.IsPrint(runes[0]) {
			m.adminFamilyForm += key
		}
	}
	
	return m, nil
}

func (m *model) renderAdminWorkspacesView(s *styles) string {
	if m.adminFamilies == nil {
		m.adminFamilies, _ = storage.GetAllFamilies(m.baseDataDir)
	}

	title := s.title.Render("WORKSPACES MANAGEMENT")
	
	if m.adminFamilyIsEditing || m.adminFamilyIsInviting {
		return m.renderAdminWorkspaceForm(s, title)
	}
	if m.adminFamilyIsDeleting {
		return m.renderAdminWorkspaceDelete(s, title)
	}

	var list string
	if len(m.adminFamilies) == 0 {
		list = s.dim.Render("No family workspaces found.")
	} else {
		for i, f := range m.adminFamilies {
			cursor := "  "
			rowStyle := s.info
			if i == m.adminFamilyCursor {
				cursor = s.highlight.Render("> ")
				rowStyle = s.highlight
			}
			
			idTag := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fmt.Sprintf(" [%s]", f.ID))
			membersTag := s.dim.Render(fmt.Sprintf("\n    Members: %s", strings.Join(f.Members, ", ")))

			list += fmt.Sprintf("%s%s%s%s\n", cursor, rowStyle.Render(f.Name), idTag, membersTag)
		}
	}

	content := title + "\n\n" + list
	
	if m.adminFamilyError != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		if strings.HasSuffix(m.adminFamilyError, "successfully.") {
			errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		}
		content += "\n\n" + errStyle.Render(m.adminFamilyError)
	}

	help := s.dim.Render("\n\ne: Rename • i: Invite User • d: Remove User/Delete • Esc: Back")
	return content + help
}

func (m *model) renderAdminWorkspaceForm(s *styles, title string) string {
	subtitle := "RENAME WORKSPACE"
	label := "New Name:"
	if m.adminFamilyIsInviting {
		subtitle = "INVITE USER"
		label = "Username:"
	}
	
	subtitleStr := s.subtitle.Render(subtitle)
	
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).PaddingLeft(1)
	
	val := inputStyle.Render(m.adminFamilyForm + s.highlight.Render("_"))
	form := fmt.Sprintf("  %s %s", s.dim.Render(label), val)

	content := title + "\n" + subtitleStr + "\n\n" + form
	
	if m.adminFamilyError != "" {
		content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.adminFamilyError)
	}

	help := s.dim.Render("\n\nEnter: Save • Esc: Cancel")
	return content + help
}

func (m *model) renderAdminWorkspaceDelete(s *styles, title string) string {
	subtitle := s.subtitle.Render("REMOVE MEMBER OR DELETE")
	
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).PaddingLeft(1)
	
	val := inputStyle.Render(m.adminFamilyForm + s.highlight.Render("_"))
	form := fmt.Sprintf("  %s %s", s.dim.Render("Username to remove (or 'ALL' to delete workspace):"), val)

	content := title + "\n" + subtitle + "\n\n" + form
	
	if m.adminFamilyError != "" {
		content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.adminFamilyError)
	}

	help := s.dim.Render("\n\nEnter: Confirm • Esc: Cancel")
	return content + help
}
