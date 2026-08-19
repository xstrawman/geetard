package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"geetard/internal/client"
	"geetard/internal/config"
	"geetard/internal/library"
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

func TestSettings_selectThemeWritesFile(t *testing.T) {
	dir := t.TempDir()
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), dir)
	m.State = StateSettings
	m.SettingIdx = 0
	m.InValues = true
	m.ValueIdx = 3 // nord
	next, _ := m.Update(enterKey())
	got := next.(Model)
	if got.Sel.Theme != "nord" {
		t.Fatal(got.Sel.Theme)
	}
	s, _, err := config.Load(dir)
	if err != nil || s.Theme != "nord" {
		t.Fatal(s, err)
	}
}

func TestSettingsView_hasPreview(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Width, m.Height = 100, 30
	m.State = StateSettings
	body := viewString(m.View())
	for _, w := range []string{"Theme", "Border", "Density", "Autoscroll", "Scroll speed", "Diagrams", "Preview"} {
		if !strings.Contains(body, w) {
			t.Fatal("missing", w, body)
		}
	}
}

func TestTabOpen_savesToLibrary(t *testing.T) {
	dir := t.TempDir()
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.LibraryDir = dir
	next, _ := m.Update(tabMsg{Tab: client.TabDetail{
		Path:   "willie-nelson/always-on-my-mind-chords-77919",
		Artist: "Willie Nelson",
		Song:   "Always On My Mind",
		Raw:    "[ch]G[/ch]\nhello",
	}})
	got := next.(Model)
	if got.State != StateReader {
		t.Fatal(got.State)
	}
	list, err := library.List(dir)
	if err != nil || len(list) != 1 || list[0].Song != "Always On My Mind" {
		t.Fatalf("%+v %v", list, err)
	}
}

func TestLibrary_openFromDisk(t *testing.T) {
	dir := t.TempDir()
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.LibraryDir = dir
	_ = library.Save(dir, library.FromTab(client.TabDetail{
		Path: "a/b", Artist: "A", Song: "B", Raw: "body",
	}))
	next, _ := m.Update(press("l"))
	lib := next.(Model)
	if lib.State != StateLibrary || len(lib.Saved) != 1 {
		t.Fatalf("%s %d", lib.State, len(lib.Saved))
	}
	next, _ = lib.Update(enterKey())
	got := next.(Model)
	if got.State != StateReader || got.Tab.Song != "B" {
		t.Fatal(got.State, got.Tab.Song)
	}
}

func TestHelpView_mentionsProxyAndKeys(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Width, m.Height = 80, 24
	m.State = StateHelp
	body := viewString(m.View())
	for _, w := range []string{"freetar.de", "ultimate-guitar.com", "s settings", "a autoscroll", "l library"} {
		if !strings.Contains(body, w) {
			t.Fatal("missing", w, body)
		}
	}
}

func TestKeys_settingsDoesNotClobberPrev(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.State = StateReader
	next, _ := m.Update(press("s"))
	next, _ = next.(Model).Update(press("s"))
	next, _ = next.(Model).Update(press("esc"))
	if next.(Model).State != StateReader {
		t.Fatal(next.(Model).State)
	}
}

func TestSettings_openSyncsValueIdx(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Sel.Theme = "nord"
	next, _ := m.Update(press("s"))
	got := next.(Model)
	if got.ValueIdx != 3 {
		t.Fatal(got.ValueIdx)
	}
}
