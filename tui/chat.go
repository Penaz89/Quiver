package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type chatMsg struct {
	msg ChatMessage
}

func waitForChatActivity(ch chan ChatMessage) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return chatMsg{msg}
	}
}

func (m *model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.chatInput, tiCmd = m.chatInput.Update(msg)
	m.chatViewport, vpCmd = m.chatViewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.focusContent = false
			return m, nil
		case "enter":
			text := strings.TrimSpace(m.chatInput.Value())
			if text != "" {
				broadcastChat(m.user, text)
				m.chatInput.Reset()
			}
			return m, nil
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *model) renderChatView(s *styles) string {
	title := s.title.Render("CHAT INTERNA")
	
	// Create viewport content
	history := GetChatHistory()
	var lines []string
	for _, msg := range history {
		timeStr := s.dim.Render(msg.Timestamp.Format("15:04"))
		senderStr := s.highlight.Render(msg.Sender)
		if msg.Sender == m.user {
			senderStr = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(msg.Sender)
		}
		lines = append(lines, fmt.Sprintf("%s %s: %s", timeStr, senderStr, msg.Text))
	}
	
	// Empty state
	if len(lines) == 0 {
		lines = append(lines, s.dim.Render("Nessun messaggio. Scrivi qualcosa per iniziare!"))
	}
	
	content := strings.Join(lines, "\n")
	m.chatViewport.SetContent(content)
	m.chatViewport.GotoBottom()
	
	divider := s.dim.Render("  " + strings.Repeat("─", m.chatViewport.Width()-2))
	
	vpStr := lipgloss.NewStyle().MarginLeft(2).Render(m.chatViewport.View())
	inputStr := lipgloss.NewStyle().MarginLeft(2).Render(m.chatInput.View())
	
	help := s.dim.Render("\n\nEnter: Invia • Esc: Indietro")
	
	return title + "\n\n" + vpStr + "\n" + divider + "\n" + inputStr + help
}
