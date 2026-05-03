package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/penaz/quiver/storage"
)

func (m *model) updateJournal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.journalIsEditing {
		switch msg.String() {
		case "esc":
			m.journalIsEditing = false
			dateStr := m.journalDate.Format("2006-01-02")
			if m.journal.Entries == nil {
				m.journal.Entries = make(map[string]string)
			}
			m.journal.Entries[dateStr] = m.journalTextArea.Value()
			_ = storage.SaveJournal(m.dataDir, m.journal)
			return m, nil
		default:
			var cmd tea.Cmd
			m.journalTextArea, cmd = m.journalTextArea.Update(msg)
			return m, cmd
		}
	}

	if m.journalIsDeleting {
		switch msg.String() {
		case "y", "Y", "s", "S":
			dateStr := m.journalDate.Format("2006-01-02")
			if m.journal.Entries != nil {
				delete(m.journal.Entries, dateStr)
				_ = storage.SaveJournal(m.dataDir, m.journal)
			}
			m.journalIsDeleting = false
		default:
			m.journalIsDeleting = false
		}
		return m, nil
	}

	m.journalMsg = "" // clear message

	switch msg.String() {
	case "left", "h":
		m.journalDate = m.journalDate.AddDate(0, 0, -1)
	case "right", "l":
		m.journalDate = m.journalDate.AddDate(0, 0, 1)
	case "enter":
		m.journalIsEditing = true
		dateStr := m.journalDate.Format("2006-01-02")
		if m.journal.Entries == nil {
			m.journal.Entries = make(map[string]string)
		}
		m.journalTextArea.SetValue(m.journal.Entries[dateStr])
		m.journalTextArea.Focus()
	case "d", "del":
		dateStr := m.journalDate.Format("2006-01-02")
		if m.journal.Entries != nil && m.journal.Entries[dateStr] != "" {
			m.journalIsDeleting = true
		}
	case "e":
		path, err := storage.ExportJournalMarkdown(m.dataDir, m.journal)
		if err == nil {
			m.journalMsg = fmt.Sprintf("%s %s", t(m.lang, "journal.exported"), path)
		}
	case "esc":
		m.focusContent = false
	}
	return m, nil
}

func (m *model) renderJournalView(s *styles) string {
	title := s.title.Render(t(m.lang, "journal.title"))
	desc := s.subtitle.Render(t(m.lang, "journal.subtitle"))

	dateStr := m.journalDate.Format("2006-01-02 (Mon)")
	if m.lang == "it" {
		// Just a simple translation for the day of week could be complex, keeping English layout for simplicity
		dateStr = m.journalDate.Format("2006-01-02")
	}
	
	dateHeader := s.info.Render("  ◄ " + dateStr + " ►")

	var content string
	var help string

	sw := sidebarWidth(m.width)
	contentW := m.width - sw - 14

	if m.journalIsEditing {
		m.journalTextArea.SetWidth(contentW)
		m.journalTextArea.SetHeight(m.height - 16)
		content = "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.journalTextArea.View())
		help = s.dim.Render("\n\n" + t(m.lang, "journal.saveHelp"))
	} else {
		entry := ""
		if m.journal.Entries != nil {
			entry = m.journal.Entries[m.journalDate.Format("2006-01-02")]
		}
		
		if entry == "" {
			entry = s.dim.Render("(No entry for this date)")
		}
		
		contentWidget := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			MarginLeft(2).
			Width(contentW).
			Height(m.height - 16).
			Render(entry)
		content = "\n" + contentWidget

		if m.journalIsDeleting {
			warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).MarginLeft(2)
			content = "\n\n" + warningStyle.Render(t(m.lang, "journal.confirmDelete"))
		}

		helpStr := fmt.Sprintf("\n\n%s  %s  %s  %s  %s",
			t(m.lang, "journal.navHelp"),
			t(m.lang, "journal.editHelp"),
			t(m.lang, "journal.deleteHelp"),
			t(m.lang, "journal.exportHelp"),
			t(m.lang, "help.goBack"),
		)
		
		if m.journalMsg != "" {
			helpStr += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(m.journalMsg)
		}
		
		help = s.dim.Render(helpStr)
	}

	return title + "\n" + desc + "\n\n" + dateHeader + content + help
}
