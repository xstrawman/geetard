package theme

import "charm.land/lipgloss/v2"

type Theme struct {
	Name, Primary, Secondary, Tertiary string
}

var Order = []string{"amber", "green", "mono", "nord"}

var catalog = map[string]Theme{
	"amber": {Name: "amber", Primary: "#E09A3E", Secondary: "#E6C07B", Tertiary: "#F0C14B"},
	"green": {Name: "green", Primary: "#3D8B4A", Secondary: "#8FBC8F", Tertiary: "#7CFC00"},
	"mono":  {Name: "mono", Primary: "#888888", Secondary: "#CCCCCC", Tertiary: "#FFFFFF"},
	"nord":  {Name: "nord", Primary: "#88C0D0", Secondary: "#D8DEE9", Tertiary: "#EBCB8B"},
}

func Named(name string) (Theme, bool) {
	t, ok := catalog[name]
	return t, ok
}

func Must(name string) Theme {
	t, ok := Named(name)
	if !ok {
		t, _ = Named("amber")
	}
	return t
}

func Cycle(name string) string {
	for i, n := range Order {
		if n == name {
			return Order[(i+1)%len(Order)]
		}
	}
	return Order[0]
}

type Styles struct {
	Box, Body, Chord, Title, Footer, Selected lipgloss.Style
}

func Apply(t Theme) Styles {
	prim := lipgloss.Color(t.Primary)
	sec := lipgloss.Color(t.Secondary)
	tert := lipgloss.Color(t.Tertiary)
	box := lipgloss.NewStyle().Padding(1).Margin(0, 1).Border(lipgloss.ASCIIBorder()).BorderForeground(prim).Foreground(sec)
	return Styles{
		Box:      box,
		Body:     lipgloss.NewStyle().Foreground(sec),
		Chord:    lipgloss.NewStyle().Foreground(tert).Bold(true),
		Title:    lipgloss.NewStyle().Foreground(prim),
		Footer:   lipgloss.NewStyle().Faint(true).Foreground(tert).Align(lipgloss.Center),
		Selected: lipgloss.NewStyle().Foreground(tert),
	}
}

func SetBorder(s Styles, name string) Styles {
	var b lipgloss.Border
	switch name {
	case "normal":
		b = lipgloss.NormalBorder()
	case "rounded":
		b = lipgloss.RoundedBorder()
	case "none":
		b = lipgloss.HiddenBorder()
	default:
		b = lipgloss.ASCIIBorder()
	}
	s.Box = s.Box.BorderStyle(b)
	return s
}
