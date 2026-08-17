package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"geetard/internal/client"
	"geetard/internal/config"
	"geetard/internal/theme"
)

type autoTick struct{}

func autoCmd(ms int) tea.Cmd {
	if ms < 100 {
		ms = 100
	}
	return tea.Tick(time.Duration(ms)*time.Millisecond, func(time.Time) tea.Msg {
		return autoTick{}
	})
}

func (m Model) updateAutoTick() (tea.Model, tea.Cmd) {
	if !m.AutoOn || m.State != StateReader {
		return m, nil
	}
	m.Scroll++
	m = m.clampScroll()
	return m, autoCmd(m.AutoMs)
}

func (m Model) updateReaderKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m = m.scrollBy(1)
	case "k", "up":
		m = m.scrollBy(-1)
	case "space":
		page := m.Height - 8
		if page < 1 {
			page = 1
		}
		m = m.scrollBy(page)
	case "a":
		m.AutoOn = !m.AutoOn
		if m.AutoOn {
			return m, autoCmd(m.AutoMs)
		}
	case "[":
		m.AutoMs += 50
		m.AutoMs = clampAutoMs(m.AutoMs)
	case "]":
		m.AutoMs -= 50
		m.AutoMs = clampAutoMs(m.AutoMs)
	}
	return m, nil
}

func (m Model) scrollBy(delta int) Model {
	m.Scroll += delta
	return m.clampScroll()
}

func (m Model) clampScroll() Model {
	max := len(m.Lines) - 1
	if max < 0 {
		max = 0
	}
	if m.Scroll < 0 {
		m.Scroll = 0
	}
	if m.Scroll > max {
		m.Scroll = max
	}
	return m
}

func sheetVisible(m Model) int {
	n := m.Height - 10
	if n < 3 {
		return 3
	}
	return n
}

func clampAutoMs(ms int) int {
	if ms < 100 {
		return 100
	}
	if ms > 2000 {
		return 2000
	}
	return ms
}

func readerView(m Model) string {
	st := theme.SetBorder(theme.Apply(theme.Must(m.Sel.Theme)), m.Sel.Border)
	box := st.Box
	if m.Width > 0 {
		w := m.Width - box.GetHorizontalMargins()
		if w > 0 {
			box = box.Width(w)
		}
	}

	title := box.Render(readerTitle(m))
	sheet := renderSheet(m, st)
	diagrams := renderDiagrams(st, m.Tab.Shapes)
	body := layoutReaderBody(m, st, box, sheet, diagrams)
	footer := st.Footer.Render(readerFooter(m))
	return lipgloss.JoinVertical(lipgloss.Left, title, body, footer)
}

func readerTitle(m Model) string {
	return fmt.Sprintf("terminal GEETARD — read ── %s — %s  v%d  capo %s",
		m.Tab.Artist, m.Tab.Song, m.Tab.Version, m.Tab.Capo)
}

func readerFooter(m Model) string {
	foot := "j/k scroll   space page   a autoscroll   s settings   ? help   q back"
	if m.AutoOn {
		foot += fmt.Sprintf("   autoscroll * %dms", m.AutoMs)
	}
	if m.Status != "" {
		foot = m.Status + "   " + foot
	}
	return foot
}

func layoutReaderBody(m Model, st theme.Styles, box lipgloss.Style, sheet, diagrams string) string {
	mode := m.Sel.Diagrams
	sidebar := mode == "sidebar" && m.Width >= 80 && diagrams != ""
	below := diagrams != "" && (mode == "below" || (mode == "sidebar" && m.Width < 80))
	switch {
	case sidebar:
		avail := m.Width
		if hm := st.Box.GetHorizontalMargins(); avail > hm {
			avail -= hm
		}
		sheetW := avail * 2 / 3
		diagW := avail - sheetW
		if sheetW < 1 {
			sheetW = 1
		}
		if diagW < 1 {
			diagW = 1
		}
		left := st.Box.Width(sheetW).Render(sheet)
		right := st.Box.Width(diagW).Render(diagrams)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	case below:
		return lipgloss.JoinVertical(lipgloss.Left, box.Render(diagrams), box.Render(sheet))
	default:
		return box.Render(sheet)
	}
}

func renderSheet(m Model, st theme.Styles) string {
	start := m.Scroll
	if start < 0 {
		start = 0
	}
	if start > len(m.Lines) {
		start = len(m.Lines)
	}
	end := start + sheetVisible(m)
	if end > len(m.Lines) {
		end = len(m.Lines)
	}
	gap := config.DensityGap(m.Sel.Density)
	var b strings.Builder
	for i, line := range m.Lines[start:end] {
		if i > 0 {
			for range gap {
				b.WriteByte('\n')
			}
		}
		b.WriteString(renderDisplayLine(st, line))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderDisplayLine(st theme.Styles, line client.DisplayLine) string {
	var b strings.Builder
	for _, p := range line.Parts {
		switch p.Kind {
		case client.KindChord:
			b.WriteString(st.Chord.Render(p.Text))
		default:
			b.WriteString(st.Body.Render(p.Text))
		}
	}
	return b.String()
}

func renderDiagrams(st theme.Styles, shapes []client.ChordShape) string {
	if len(shapes) == 0 {
		return ""
	}
	var b strings.Builder
	for i, sh := range shapes {
		if i > 0 {
			b.WriteByte('\n')
		}
		if sh.Name != "" {
			b.WriteString(st.Chord.Render(sh.Name))
			b.WriteByte('\n')
		}
		for _, line := range sh.Lines {
			b.WriteString(st.Body.Render(line))
			b.WriteByte('\n')
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
