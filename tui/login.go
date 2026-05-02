package tui

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func (m *model) updateLogin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down":
		m.loginCursor = (m.loginCursor + 1) % 2
	case "shift+tab", "up":
		m.loginCursor = (m.loginCursor - 1 + 2) % 2
	case "enter":
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
			if storage.CheckUserAuth(m.baseDataDir, user, pass) {
				m.isLoggedIn = true
				m.user = user // set the current app user to the authenticated user
				m.dataDir = storage.GetUserDir(m.baseDataDir, user)
				m.loadUserData()
				m.loginError = ""
				m.loginForm = [2]string{} // clear credentials
				return m, nil
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
		field := &m.loginForm[m.loginCursor]
		if len(*field) > 0 {
			runes := []rune(*field)
			*field = string(runes[:len(runes)-1])
		}
	case "ctrl+n":
		m.isRegistering = !m.isRegistering
		m.loginError = ""
		m.loginForm[1] = ""
	default:
		if key == "space" {
			key = " "
		}
		runes := []rune(key)
		if len(runes) == 1 && unicode.IsPrint(runes[0]) {
			m.loginForm[m.loginCursor] += key
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
	
	formBlock := usrLine + "\n\n" + pwdLine
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
	var helpText string
	if m.isRegistering {
		helpText = "Enter: Register • Esc: Cancel"
	} else {
		helpText = "Enter: Login • Ctrl+N: Create Account • Esc: Quit"
	}
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
