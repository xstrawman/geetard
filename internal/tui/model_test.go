package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"geetard/internal/client"
	"geetard/internal/config"
)

func enterKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyEnter}
}

func viewString(v tea.View) string { return v.Content }

func press(s string) tea.KeyPressMsg {
	switch s {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	default:
		r := []rune(s)[0]
		return tea.KeyPressMsg{Text: s, Code: r}
	}
}

func TestNew_startsOnSearch(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	if m.State != StateSearch {
		t.Fatalf("%s", m.State)
	}
}

func TestKeys_settingsAndHelpAndBack(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	next, _ := m.Update(press("s"))
	if next.(Model).State != StateSettings {
		t.Fatal(next.(Model).State)
	}
	next, _ = next.(Model).Update(press("esc"))
	if next.(Model).State != StateSearch {
		t.Fatal(next.(Model).State)
	}
	next, _ = next.(Model).Update(press("?"))
	if next.(Model).State != StateHelp {
		t.Fatal(next.(Model).State)
	}
}

func TestKeys_themeCycle(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	next, _ := m.Update(press("t"))
	if next.(Model).Sel.Theme != "green" {
		t.Fatal(next.(Model).Sel.Theme)
	}
}

func TestSearch_enterOnEmptyDoesNotSearch(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Querying = true
	m.Query = ""
	_, cmd := m.Update(enterKey())
	if cmd != nil {
		t.Fatal("empty query must not fire a command")
	}
}

func TestSearchView_listsResults(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Width, m.Height = 80, 24
	m.Results = client.SearchPage{Results: []client.SearchResult{{
		Artist: "Willie Nelson", Song: "Always On My Mind", Type: "Chords", Rating: 4.8,
	}}}
	if !strings.Contains(viewString(m.View()), "Willie Nelson") {
		t.Fatal(viewString(m.View()))
	}
}

func TestReader_autoscrollTickAdvances(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.State = StateReader
	m.AutoOn = true
	m.Lines = make([]client.DisplayLine, 40)
	m.Height = 10
	next, _ := m.Update(autoTick{})
	if next.(Model).Scroll != 1 {
		t.Fatal(next.(Model).Scroll)
	}
}

func TestReaderView_showsChordsAndTitle(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Width, m.Height = 100, 24
	m.State = StateReader
	m.Tab = client.TabDetail{Artist: "Willie Nelson", Song: "Always On My Mind", Version: 3, Capo: "1",
		Shapes: []client.ChordShape{{Name: "G", Lines: []string{"  e |-3-"}}}}
	m.Lines = client.Render("[ch]G[/ch]\nMaybe I didn't love you")
	m.Sel.Diagrams = "sidebar"
	body := viewString(m.View())
	if !strings.Contains(body, "Always On My Mind") || !strings.Contains(body, "Maybe I didn't") {
		t.Fatal(body)
	}
	if !strings.Contains(body, "e |-3-") {
		t.Fatal("diagram missing", body)
	}
}
