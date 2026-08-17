package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"geetard/internal/client"
	"geetard/internal/config"
	"geetard/internal/theme"
)

type State string

const (
	StateSearch   State = "search"
	StateReader   State = "read"
	StateSettings State = "settings"
	StateHelp     State = "help"
)

type Model struct {
	Width, Height int
	State         State
	Prev          State
	Client        *client.Client
	Sel           config.Selections
	ConfigDir     string
	Query         string
	Querying      bool
	Results       client.SearchPage
	Cursor        int
	Status        string
	Tab           client.TabDetail
	Lines         []client.DisplayLine
	Scroll        int
	AutoOn        bool
	AutoMs        int
	SettingIdx    int
	ValueIdx      int
	InValues      bool
}

func New(c *client.Client, sel config.Selections, dir string) Model {
	return Model{
		State:     StateSearch,
		Client:    c,
		Sel:       sel,
		ConfigDir: dir,
		AutoOn:    sel.Autoscroll == "on",
		AutoMs:    config.IntervalMs(sel.ScrollSpeed),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = message.Width
		m.Height = message.Height
		return m, nil
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.State == StateReader {
				m.State = StateSearch
				return m, nil
			}
			return m, tea.Quit
		case "s":
			m.Prev = m.State
			m.State = StateSettings
			return m, nil
		case "?":
			m.Prev = m.State
			m.State = StateHelp
			return m, nil
		case "esc", "backspace":
			if m.State == StateSettings || m.State == StateHelp {
				m.State = m.Prev
			}
			return m, nil
		case "t":
			if m.State == StateSearch || m.State == StateReader {
				m.Sel.Theme = theme.Cycle(m.Sel.Theme)
				_ = config.Save(m.ConfigDir, m.Sel)
			}
			return m, nil
		case "/":
			if m.State == StateSearch {
				m.Querying = true
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	st := theme.Apply(theme.Must(m.Sel.Theme))
	title := st.Title.Render(fmt.Sprintf("terminal GEETARD — %s", m.State))
	body := st.Body.Render(placeholder(m.State))
	footer := st.Footer.Render(footerKeys(m.State))
	v := tea.NewView("")
	v.AltScreen = true
	v.SetContent(title + "\n\n" + body + "\n\n" + footer)
	return v
}

func placeholder(s State) string {
	switch s {
	case StateSearch:
		return "type / to search"
	case StateReader:
		return "tab"
	case StateSettings:
		return "settings"
	case StateHelp:
		return "help"
	default:
		return ""
	}
}

func footerKeys(s State) string {
	switch s {
	case StateReader:
		return "s settings   ? help   t theme   q back"
	case StateSettings, StateHelp:
		return "esc back   q quit"
	default:
		return "/ search   s settings   ? help   t theme   q quit"
	}
}
