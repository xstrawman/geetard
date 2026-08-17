package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"geetard/internal/config"
	"geetard/internal/theme"
)

type settingSpec struct {
	Name        string
	Description string
	Values      []string
	Get         func(config.Selections) string
	Set         func(*config.Selections, string)
}

func settingsCatalog() []settingSpec {
	return []settingSpec{
		{
			Name:        "Theme",
			Description: "Color theme. Primary = borders, secondary = body, tertiary = chords.",
			Values:      theme.Order,
			Get:         func(s config.Selections) string { return s.Theme },
			Set:         func(s *config.Selections, v string) { s.Theme = v },
		},
		{
			Name:        "Border",
			Description: "Lipgloss border on the boxes.",
			Values:      []string{"ascii", "normal", "rounded", "none"},
			Get:         func(s config.Selections) string { return s.Border },
			Set:         func(s *config.Selections, v string) { s.Border = v },
		},
		{
			Name:        "Density",
			Description: "Line gap 0 / 1 / 2. The TUI stand-in for font size.",
			Values:      []string{"compact", "normal", "airy"},
			Get:         func(s config.Selections) string { return s.Density },
			Set:         func(s *config.Selections, v string) { s.Density = v },
		},
		{
			Name:        "Autoscroll",
			Description: "Default when a reader opens. a still toggles for this song.",
			Values:      []string{"off", "on"},
			Get:         func(s config.Selections) string { return s.Autoscroll },
			Set:         func(s *config.Selections, v string) { s.Autoscroll = v },
		},
		{
			Name:        "Scroll speed",
			Description: "Line interval. [ ] still nudge this session.",
			Values:      []string{"slow", "medium", "fast"},
			Get:         func(s config.Selections) string { return s.ScrollSpeed },
			Set:         func(s *config.Selections, v string) { s.ScrollSpeed = v },
		},
		{
			Name:        "Diagrams",
			Description: "ASCII chord shapes from UG applicature. Preview = a G shape.",
			Values:      []string{"off", "sidebar", "below"},
			Get:         func(s config.Selections) string { return s.Diagrams },
			Set:         func(s *config.Selections, v string) { s.Diagrams = v },
		},
	}
}

func (m Model) updateSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cat := settingsCatalog()
	m.SettingIdx = clampIdx(m.SettingIdx, len(cat))
	vals := cat[m.SettingIdx].Values
	m.ValueIdx = clampIdx(m.ValueIdx, len(vals))

	switch msg.String() {
	case "h", "left":
		m.InValues = false
	case "l", "right":
		m.InValues = true
	case "j", "down":
		if m.InValues {
			if m.ValueIdx < len(vals)-1 {
				m.ValueIdx++
			}
		} else if m.SettingIdx < len(cat)-1 {
			m.SettingIdx++
			m.ValueIdx = valueIndex(cat[m.SettingIdx], m.Sel)
		}
	case "k", "up":
		if m.InValues {
			if m.ValueIdx > 0 {
				m.ValueIdx--
			}
		} else if m.SettingIdx > 0 {
			m.SettingIdx--
			m.ValueIdx = valueIndex(cat[m.SettingIdx], m.Sel)
		}
	case "enter", "space":
		if m.InValues && m.ValueIdx >= 0 && m.ValueIdx < len(vals) {
			m = m.applySetting(cat[m.SettingIdx], vals[m.ValueIdx])
		}
	}
	return m, nil
}

func clampIdx(i, n int) int {
	if n <= 0 {
		return 0
	}
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}

func valueIndex(s settingSpec, sel config.Selections) int {
	cur := s.Get(sel)
	for i, v := range s.Values {
		if v == cur {
			return i
		}
	}
	return 0
}

func (m Model) applySetting(s settingSpec, val string) Model {
	s.Set(&m.Sel, val)
	switch s.Name {
	case "Autoscroll":
		m.AutoOn = val == "on"
	case "Scroll speed":
		m.AutoMs = config.IntervalMs(val)
	}
	_ = config.Save(m.ConfigDir, m.Sel)
	return m
}

func settingsView(m Model) string {
	st := theme.SetBorder(theme.Apply(theme.Must(m.Sel.Theme)), m.Sel.Border)
	box := st.Box
	if m.Width > 0 {
		w := m.Width - box.GetHorizontalMargins()
		if w > 0 {
			box = box.Width(w)
		}
	}

	title := box.Render("terminal GEETARD — settings")

	cat := settingsCatalog()
	idx := clampIdx(m.SettingIdx, len(cat))
	spec := cat[idx]
	vIdx := clampIdx(m.ValueIdx, len(spec.Values))

	var left, mid strings.Builder
	left.WriteString("Settings\n\n")
	for i, s := range cat {
		mark := "[ ]"
		if i == idx {
			mark = "[o]"
		}
		line := mark + " " + s.Name
		if !m.InValues && i == idx {
			line = st.Selected.Render(line)
		} else {
			line = st.Body.Render(line)
		}
		left.WriteString(line)
		left.WriteByte('\n')
	}

	mid.WriteString("Options\n\n")
	applied := spec.Get(m.Sel)
	for i, v := range spec.Values {
		mark := "[ ]"
		if v == applied {
			mark = "[o]"
		}
		line := mark + " " + v
		if m.InValues && i == vIdx {
			line = st.Selected.Render(line)
		} else {
			line = st.Body.Render(line)
		}
		mid.WriteString(line)
		mid.WriteByte('\n')
	}

	hovered := spec.Values[vIdx]
	preview := "Preview\n\n" + settingsPreview(spec, hovered, st)
	desc := "Description\n\n" + spec.Description

	leftW, midW := 22, 20
	prevW := 36
	if m.Width > 0 {
		avail := m.Width - box.GetHorizontalFrameSize() - box.GetHorizontalMargins()
		if avail > 60 {
			leftW = avail / 4
			midW = avail / 4
			prevW = avail - leftW - midW
		}
	}
	if leftW < 1 {
		leftW = 1
	}
	if midW < 1 {
		midW = 1
	}
	if prevW < 1 {
		prevW = 1
	}

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(strings.TrimRight(left.String(), "\n")),
		lipgloss.NewStyle().Width(midW).Render(strings.TrimRight(mid.String(), "\n")),
		lipgloss.NewStyle().Width(prevW).Render(preview),
	)
	body := box.Render(cols + "\n\n" + desc)
	foot := "esc back   j/k move   h/l pane   enter select   q quit"
	if m.Status != "" {
		foot = m.Status + "   " + foot
	}
	footer := st.Footer.Render(foot)
	return lipgloss.JoinVertical(lipgloss.Left, title, body, footer)
}

func settingsPreview(spec settingSpec, hovered string, st theme.Styles) string {
	switch spec.Name {
	case "Theme":
		th := theme.Apply(theme.Must(hovered))
		return th.Chord.Render("G        D        Em") + "\n" + th.Body.Render("Maybe I didn't love you")
	case "Border":
		return theme.SetBorder(st, hovered).Box.Width(14).Height(5).Render(hovered)
	case "Density":
		gap := config.DensityGap(hovered)
		var b strings.Builder
		for i, line := range []string{"line one", "line two", "line three"} {
			if i > 0 {
				for range gap {
					b.WriteByte('\n')
				}
			}
			b.WriteString(st.Body.Render(line))
			b.WriteByte('\n')
		}
		return strings.TrimRight(b.String(), "\n")
	case "Autoscroll":
		return st.Body.Render("autoscroll " + hovered)
	case "Scroll speed":
		return st.Body.Render(fmt.Sprintf("%s (%dms)", hovered, config.IntervalMs(hovered)))
	case "Diagrams":
		if hovered == "off" {
			return st.Body.Render("diagrams off")
		}
		var b strings.Builder
		b.WriteString(st.Chord.Render("G"))
		b.WriteByte('\n')
		for _, line := range []string{
			"  e |-3-",
			"  B |-0-",
			"  G |-0-",
			"  D |-0-",
			"  A |-2-",
			"  E |-3-",
		} {
			b.WriteString(st.Body.Render(line))
			b.WriteByte('\n')
		}
		return strings.TrimRight(b.String(), "\n")
	default:
		return st.Body.Render(hovered)
	}
}
