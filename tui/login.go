package tui

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func (m *model) updateLogin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.isChangingPassword {
		return m.updateChangePassword(msg)
	}

	key := msg.String()
	switch key {
	case "tab", "down":
		m.loginCursor = (m.loginCursor + 1) % 4
	case "shift+tab", "up":
		m.loginCursor = (m.loginCursor - 1 + 4) % 4
	case "enter":
		if m.loginCursor == 3 {
			m.isRegistering = !m.isRegistering
			m.loginError = ""
			m.loginForm[1] = ""
			m.loginCursor = 0
			return m, nil
		}

		user := strings.TrimSpace(m.loginForm[0])
		pass := m.loginForm[1]

		if user == "" || pass == "" {
			m.loginError = "Username and password cannot be empty."
			return m, nil
		}

		if m.isRegistering {
			err := storage.CreateUser(m.baseDataDir, user, pass)
			if err != nil {
				m.loginError = "Error: " + err.Error()
				return m, nil
			}
			m.isRegistering = false
			m.loginError = "Registration successful! You can now log in."
			m.loginForm[1] = "" // clear password
			m.loginCursor = 1
			return m, nil
		} else {
			userRecord, ok := storage.CheckUserAuth(m.baseDataDir, user, pass)
			if ok {
				m.user = user // set the current app user to the authenticated user
				m.dataDir = storage.GetUserDir(m.baseDataDir, user)
				m.personalDataDir = m.dataDir
				
				// Read personal settings to determine default workspace
				personalSettings := storage.LoadSettings(m.personalDataDir)
				m.currentWorkspace = personalSettings.DefaultWorkspace
				
				// Ensure workspace exists, fallback to personal if not
				if m.currentWorkspace != "Personal" {
					families, _ := storage.GetUserFamilies(m.baseDataDir, user)
					found := false
					for _, f := range families {
						if f.ID == m.currentWorkspace {
							found = true
							m.dataDir = storage.GetFamilyDir(m.baseDataDir, m.currentWorkspace)
							break
						}
					}
					if !found {
						m.currentWorkspace = "Personal"
						m.dataDir = m.personalDataDir
					}
				}
				
				registerSessionLogin(m.ctx, m.user)
				if userRecord.Role == "admin" {
					m.isAdmin = true
				}
				if userRecord.MustChange {
					m.isChangingPassword = true
					m.loginError = ""
					m.changePasswordForm = [2]string{}
					m.changePasswordCur = 0
					m.changePasswordErr = ""
					return m, nil
				}

				m.isLoggedIn = true
				m.loadUserData()
				m.loginError = ""
				m.loginForm = [2]string{} // clear credentials

				var cmds []tea.Cmd
				if m.settings.WeatherLoc != "" {
					cmds = append(cmds, fetchWeatherCmd(m.settings.WeatherLoc))
				}
				return m, tea.Batch(cmds...)
			}
			m.loginError = "Invalid username or password."
			m.loginForm[1] = "" // clear password
			return m, nil
		}
	case "esc":
		if m.isRegistering {
			m.isRegistering = false
			m.loginError = ""
		} else {
			return m, tea.Quit
		}
	case "backspace":
		if m.loginCursor < 2 {
			field := &m.loginForm[m.loginCursor]
			if len(*field) > 0 {
				runes := []rune(*field)
				*field = string(runes[:len(runes)-1])
			}
		}
	default:
		if m.loginCursor < 2 {
			if key == "space" {
				key = " "
			}
			runes := []rune(key)
			if len(runes) == 1 && unicode.IsPrint(runes[0]) {
				m.loginForm[m.loginCursor] += key
			}
		}
	}
	return m, nil
}

func (m *model) renderLoginView(s *styles) string {
	boxWidth := 56

	var titleText string
	if m.isRegistering {
		titleText = "CREATE NEW ACCOUNT"
	} else {
		titleText = "LOGIN"
	}

	logo := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(s.logo.Render(sidebarLogo))
	title := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(s.title.Render(titleText))

	labelStyle := lipgloss.NewStyle().Width(14).Align(lipgloss.Right).MarginRight(1).Foreground(lipgloss.Color("241"))
	inputStyle := lipgloss.NewStyle().Width(20).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).PaddingLeft(1)
	emptyStyle := lipgloss.NewStyle().Width(20).PaddingLeft(1).Foreground(lipgloss.Color("241"))
	filledStyle := lipgloss.NewStyle().Width(20).PaddingLeft(1).Foreground(lipgloss.Color("252"))

	// Username field
	usrLabel := labelStyle.Render("Username:")
	var usrVal string
	if m.loginCursor == 0 {
		usrVal = inputStyle.Render(m.loginForm[0] + s.highlight.Render("_"))
	} else {
		if m.loginForm[0] == "" {
			usrVal = emptyStyle.Render("(empty)")
		} else {
			usrVal = filledStyle.Render(m.loginForm[0])
		}
	}
	usrLine := lipgloss.JoinHorizontal(lipgloss.Top, usrLabel, usrVal)

	// Password field
	pwdLabel := labelStyle.Render("Password:")
	pwdValRaw := m.loginForm[1]
	pwdValMasked := strings.Repeat("*", len(pwdValRaw))
	var pwdVal string
	if m.loginCursor == 1 {
		pwdVal = inputStyle.Render(pwdValMasked + s.highlight.Render("_"))
	} else {
		if pwdValRaw == "" {
			pwdVal = emptyStyle.Render("(empty)")
		} else {
			pwdVal = filledStyle.Render(pwdValMasked)
		}
	}
	pwdLine := lipgloss.JoinHorizontal(lipgloss.Top, pwdLabel, pwdVal)

	// Action buttons
	btnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).Padding(0, 2)
	activeBtnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("42")).Padding(0, 2).Bold(true)

	var submitText string
	var toggleText string
	if m.isRegistering {
		submitText = "REGISTER"
		toggleText = "Back to Login"
	} else {
		submitText = "LOGIN"
		toggleText = "Create Account"
	}

	submitBtn := btnStyle.Render(submitText)
	if m.loginCursor == 2 {
		submitBtn = activeBtnStyle.Render(submitText)
	}

	toggleBtn := btnStyle.Render(toggleText)
	if m.loginCursor == 3 {
		toggleBtn = activeBtnStyle.Render(toggleText)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, submitBtn, "   ", toggleBtn)
	buttonsBlock := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(buttons)

	formBlock := usrLine + "\n\n" + pwdLine + "\n\n\n" + buttonsBlock
	form := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(formBlock)

	// Error line
	errLine := ""
	if m.loginError != "" {
		if strings.HasPrefix(m.loginError, "Registration successful") {
			errLine = lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("42")).Render(m.loginError)
		} else {
			errLine = lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("196")).Render(m.loginError)
		}
	}

	// Help instructions
	helpText := "Tab: Navigate • Enter: Select • Esc: Quit"
	help := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("241")).Render(helpText)

	content := logo + "\n\n" + title + "\n\n" + form
	if errLine != "" {
		content += "\n\n" + errLine
	}
	content += "\n\n" + help

	// Center the box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(2, 2).
		Width(boxWidth).
		Render(content)

	// Use lipgloss to place the box in the middle of the terminal
	centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	return centered
}

func (m *model) updateChangePassword(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		m.changePasswordCur = (m.changePasswordCur + 1) % 2
	case "shift+tab", "up":
		m.changePasswordCur = (m.changePasswordCur - 1 + 2) % 2
	case "enter":
		pass1 := m.changePasswordForm[0]
		pass2 := m.changePasswordForm[1]

		if pass1 == "" || pass2 == "" {
			m.changePasswordErr = "Password cannot be empty."
			return m, nil
		}
		if pass1 != pass2 {
			m.changePasswordErr = "Passwords do not match."
			return m, nil
		}

		err := storage.UpdateUserPassword(m.baseDataDir, m.user, pass1, false)
		if err != nil {
			m.changePasswordErr = "Error: " + err.Error()
			return m, nil
		}

		m.isChangingPassword = false
		m.isLoggedIn = true
		m.loadUserData()
		m.loginError = ""
		m.loginForm = [2]string{} // clear credentials

		var cmds []tea.Cmd
		if m.settings.WeatherLoc != "" {
			cmds = append(cmds, fetchWeatherCmd(m.settings.WeatherLoc))
		}
		return m, tea.Batch(cmds...)
	case "esc":
		// Cannot skip password change, quit application
		return m, tea.Quit
	case "backspace":
		field := &m.changePasswordForm[m.changePasswordCur]
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
			m.changePasswordForm[m.changePasswordCur] += key
		}
	}
	return m, nil
}

func (m *model) renderChangePasswordView(s *styles) string {
	boxWidth := 56
	titleText := "CHANGE DEFAULT PASSWORD"

	logo := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(s.logo.Render(sidebarLogo))
	title := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(s.title.Render(titleText))

	labelStyle := lipgloss.NewStyle().Width(14).Align(lipgloss.Right).MarginRight(1).Foreground(lipgloss.Color("241"))
	inputStyle := lipgloss.NewStyle().Width(20).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).PaddingLeft(1)
	emptyStyle := lipgloss.NewStyle().Width(20).PaddingLeft(1).Foreground(lipgloss.Color("241"))
	filledStyle := lipgloss.NewStyle().Width(20).PaddingLeft(1).Foreground(lipgloss.Color("252"))

	// Password field 1
	pwdLabel := labelStyle.Render("New Password:")
	pwdValRaw := m.changePasswordForm[0]
	pwdValMasked := strings.Repeat("*", len(pwdValRaw))
	var pwdVal string
	if m.changePasswordCur == 0 {
		pwdVal = inputStyle.Render(pwdValMasked + s.highlight.Render("_"))
	} else {
		if pwdValRaw == "" {
			pwdVal = emptyStyle.Render("(empty)")
		} else {
			pwdVal = filledStyle.Render(pwdValMasked)
		}
	}
	pwdLine := lipgloss.JoinHorizontal(lipgloss.Top, pwdLabel, pwdVal)

	// Password field 2
	pwd2Label := labelStyle.Render("Confirm:")
	pwd2ValRaw := m.changePasswordForm[1]
	pwd2ValMasked := strings.Repeat("*", len(pwd2ValRaw))
	var pwd2Val string
	if m.changePasswordCur == 1 {
		pwd2Val = inputStyle.Render(pwd2ValMasked + s.highlight.Render("_"))
	} else {
		if pwd2ValRaw == "" {
			pwd2Val = emptyStyle.Render("(empty)")
		} else {
			pwd2Val = filledStyle.Render(pwd2ValMasked)
		}
	}
	pwd2Line := lipgloss.JoinHorizontal(lipgloss.Top, pwd2Label, pwd2Val)

	formBlock := pwdLine + "\n\n" + pwd2Line
	form := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(formBlock)

	// Error line
	errLine := ""
	if m.changePasswordErr != "" {
		errLine = lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("196")).Render(m.changePasswordErr)
	}

	help := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("241")).Render("Enter: Confirm • Esc: Quit")

	content := logo + "\n\n" + title + "\n\n" + form
	if errLine != "" {
		content += "\n\n" + errLine
	}
	content += "\n\n" + help

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("208")).
		Padding(2, 2).
		Width(boxWidth).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
