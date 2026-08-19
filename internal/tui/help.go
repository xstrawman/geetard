package tui

import (
	"charm.land/lipgloss/v2"
	"geetard/internal/theme"
)

const helpBody = `terminal GEETARD talks to freetar.de.
It never requests ultimate-guitar.com.

/ search     enter open     n/p page
l library    j/k scroll     space page
a autoscroll [ ] speed      s settings
t theme      ? help         q back/quit
Wide chord sheets use 2–3 columns. Every opened tab is saved.`

func helpView(m Model) string {
	st := theme.SetBorder(theme.Apply(theme.Must(m.Sel.Theme)), m.Sel.Border)
	box := st.Box
	if m.Width > 0 {
		w := m.Width - box.GetHorizontalMargins()
		if w > 0 {
			box = box.Width(w)
		}
	}

	title := box.Render("terminal GEETARD — help")
	body := box.Render(st.Body.Render(helpBody))
	foot := "esc back   q quit"
	if m.Status != "" {
		foot = m.Status + "   " + foot
	}
	footer := st.Footer.Render(foot)
	return lipgloss.JoinVertical(lipgloss.Left, title, body, footer)
}
