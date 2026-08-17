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
	case searchMsg:
		m.Results = message.Page
		m.Cursor = 0
		m.Status = ""
		if message.Err != nil {
			m.Status = message.Err.Error()
		}
		return m, nil
	case tabMsg:
		if message.Err != nil {
			m.Status = message.Err.Error()
			return m, nil
		}
		m.Tab = message.Tab
		m.Lines = client.Render(message.Tab.Raw)
		m.Scroll = 0
		m.AutoOn = m.Sel.Autoscroll == "on"
		m.State = StateReader
		m.Status = ""
		return m, nil
	case errMsg:
		if message.Err != nil {
			m.Status = message.Err.Error()
		}
		return m, nil
	case tea.KeyMsg:
		if m.State == StateSearch && m.Querying {
			return m.updateSearchQuery(message)
		}
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
		if m.State == StateSearch {
			return m.updateSearchKeys(message)
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	var content string
	switch m.State {
	case StateSearch:
		content = searchView(m)
	default:
		st := theme.Apply(theme.Must(m.Sel.Theme))
		title := st.Title.Render(fmt.Sprintf("terminal GEETARD — %s", m.State))
		body := st.Body.Render(placeholder(m.State))
		footer := st.Footer.Render(footerKeys(m.State))
		content = title + "\n\n" + body + "\n\n" + footer
	}
	v := tea.NewView("")
	v.AltScreen = true
	v.SetContent(content)
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
