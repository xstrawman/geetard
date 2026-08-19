package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"geetard/internal/client"
	"geetard/internal/theme"
)

type searchMsg struct {
	Page client.SearchPage
	Err  error
}

type tabMsg struct {
	Tab client.TabDetail
	Err error
}

type errMsg struct {
	Err error
}

func SearchCmd(m Model, query string, page int) tea.Cmd {
	return func() tea.Msg {
		p, err := m.Client.Search(query, page)
		return searchMsg{Page: p, Err: err}
	}
}

func TabCmd(m Model, path string) tea.Cmd {
	return func() tea.Msg {
		tab, err := m.Client.Tab(path)
		return tabMsg{Tab: tab, Err: err}
	}
}

func (m Model) updateSearchQuery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		if m.Query == "" {
			return m, nil
		}
		m.Querying = false
		return m, SearchCmd(m, m.Query, 1)
	case "backspace":
		if q := []rune(m.Query); len(q) > 0 {
			m.Query = string(q[:len(q)-1])
		}
		return m, nil
	case "esc", "down":
		m.Querying = false
		return m, nil
	default:
		if t := msg.Key().Text; t != "" {
			m.Query += t
			return m, nil
		}
		if msg.String() == "space" {
			m.Query += " "
		}
		return m, nil
	}
}

func (m Model) updateSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if n := len(m.Results.Results); n > 0 && m.Cursor < n-1 {
			m.Cursor++
		}
	case "k", "up":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "n":
		return m, m.searchPage(1)
	case "p":
		return m, m.searchPage(-1)
	case "enter", "space":
		if i := m.Cursor; i >= 0 && i < len(m.Results.Results) {
			path := m.Results.Results[i].Path
			if path != "" {
				return m, TabCmd(m, path)
			}
		}
	}
	return m, nil
}

func searchVisible(m Model) int {
	n := m.Height - 12
	if n < 3 {
		return 3
	}
	return n
}

func (m Model) searchPage(delta int) tea.Cmd {
	q := m.Results.Query
	if q == "" {
		q = m.Query
	}
	if q == "" {
		return nil
	}
	total := m.Results.TotalPages
	if total < 1 {
		total = 1
	}
	page := m.Results.Page
	if page < 1 {
		page = 1
	}
	page += delta
	if page < 1 {
		page = 1
	}
	if page > total {
		page = total
	}
	return SearchCmd(m, q, page)
}

func searchView(m Model) string {
	st := theme.SetBorder(theme.Apply(theme.Must(m.Sel.Theme)), m.Sel.Border)
	box := st.Box
	if m.Width > 0 {
		w := m.Width - box.GetHorizontalMargins()
		if w > 0 {
			box = box.Width(w)
		}
	}

	title := box.Render("terminal GEETARD — search")

	var b strings.Builder
	b.WriteString("/ ")
	b.WriteString(m.Query)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "%-22s %-28s %-8s %s\n", "ARTIST", "SONG", "TYPE", "★")
	rows := searchVisible(m)
	start := 0
	if m.Cursor >= rows {
		start = m.Cursor - rows + 1
	}
	end := start + rows
	if end > len(m.Results.Results) {
		end = len(m.Results.Results)
	}
	for i := start; i < end; i++ {
		r := m.Results.Results[i]
		line := fmt.Sprintf("%-22s %-28s %-8s %.1f", r.Artist, r.Song, r.Type, r.Rating)
		if i == m.Cursor {
			line = st.Selected.Render(line)
		} else {
			line = st.Body.Render(line)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}

	body := box.Render(strings.TrimRight(b.String(), "\n"))
	foot := "enter open   n/p page   / search   l library   s settings   ? help   q quit"
	if m.Status != "" {
		foot = m.Status + "   " + foot
	}
	footer := st.Footer.Render(foot)
	return lipgloss.JoinVertical(lipgloss.Left, title, body, footer)
}
