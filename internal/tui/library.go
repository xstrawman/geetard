package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"geetard/internal/client"
	"geetard/internal/library"
	"geetard/internal/theme"
)

func (m Model) openLibrary() (tea.Model, tea.Cmd) {
	list, err := library.List(m.LibraryDir)
	m.Saved = list
	m.LibCursor = 0
	m.Prev = m.State
	m.State = StateLibrary
	m.Status = ""
	if err != nil {
		m.Status = err.Error()
	}
	return m, nil
}

func (m Model) updateLibraryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if n := len(m.Saved); n > 0 && m.LibCursor < n-1 {
			m.LibCursor++
		}
	case "k", "up":
		if m.LibCursor > 0 {
			m.LibCursor--
		}
	case "enter", "space":
		if i := m.LibCursor; i >= 0 && i < len(m.Saved) {
			tab := m.Saved[i].Tab()
			m.Tab = tab
			m.Lines = client.Render(tab.Raw)
			m.Scroll = 0
			m.AutoOn = m.Sel.Autoscroll == "on"
			m.State = StateReader
			m.Status = ""
			if m.AutoOn {
				return m, autoCmd(m.AutoMs)
			}
			return m, nil
		}
	}
	return m, nil
}

func libraryView(m Model) string {
	st := theme.SetBorder(theme.Apply(theme.Must(m.Sel.Theme)), m.Sel.Border)
	box := st.Box
	if m.Width > 0 {
		w := m.Width - box.GetHorizontalMargins()
		if w > 0 {
			box = box.Width(w)
		}
	}
	title := box.Render("terminal GEETARD — library")
	var b strings.Builder
	if len(m.Saved) == 0 {
		b.WriteString(st.Body.Render("No saved tabs yet. Open a song from search; it is kept here."))
	} else {
		fmt.Fprintf(&b, "%-22s %-32s %-8s\n", "ARTIST", "SONG", "TYPE")
		rows := searchVisible(m)
		start := 0
		if m.LibCursor >= rows {
			start = m.LibCursor - rows + 1
		}
		end := start + rows
		if end > len(m.Saved) {
			end = len(m.Saved)
		}
		for i := start; i < end; i++ {
			r := m.Saved[i]
			line := fmt.Sprintf("%-22s %-32s %-8s", r.Artist, r.Song, r.Type)
			if i == m.LibCursor {
				line = st.Selected.Render(line)
			} else {
				line = st.Body.Render(line)
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	body := box.Render(strings.TrimRight(b.String(), "\n"))
	foot := "enter open   j/k move   q back"
	if m.Status != "" {
		foot = m.Status + "   " + foot
	}
	footer := st.Footer.Render(foot)
	return lipgloss.JoinVertical(lipgloss.Left, title, body, footer)
}
