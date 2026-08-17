# terminal GEETARD Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `geetard`, a Go Bubble Tea TUI that searches via `proxy.freetar.de`, opens tabs from `tabs.proxy.freetar.de`, and lets you play in ttune-shaped themes, settings, autoscroll, and ASCII chord diagrams.

**Architecture:** `internal/client` is a testable HTTP + parse + markup library. `internal/config` and `internal/theme` are data. `internal/tui` is Bubble Tea states (search, reader, settings, help). `cmd/geetard` only starts the program. No live network in tests.

**Tech Stack:** Go 1.26, Bubble Tea v2 + Lipgloss v2 (`charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`), stdlib `net/http` + `encoding/json` + `net/http/httptest`.

**Spec:** `docs/superpowers/specs/2026-08-17-terminal-geetard-design.md`

---

## File map

```
terminal-geetard/
  go.mod
  README.md
  cmd/geetard/main.go
  internal/client/
    types.go           # data types + ClientError
    parse.go           # js-store extract, flexible numbers
    map.go             # HTML/JSON → SearchPage / TabDetail
    url.go             # hosts, search URL, tab path, reject UG
    render.go          # [ch] / [tab] → DisplayLine
    shapes.go          # applicature → ChordShape ASCII
    http.go            # Client.Search / Client.Tab
    *_test.go
  internal/config/
    config.go          # selections.json load/save
    config_test.go
  internal/theme/
    theme.go           # four palettes, Styles
    theme_test.go
  internal/tui/
    model.go           # Model, State, NewModel, Update, View
    search.go
    reader.go
    settings.go
    help.go
    model_test.go
  testdata/
    search_ok.html
    tab_ok.html
    tab_no_applicature.html
    no_store.html
    selections_ok.json
    selections_bad.json
```

Work from the repo root: `~/Projects/apps/terminal-geetard`.

---

### Task 1: Go module and README

**Files:**
- Create: `go.mod`
- Create: `README.md`

- [ ] **Step 1: Init the module**

Run:

```bash
cd /home/mountaindewurbest/Projects/apps/terminal-geetard
go mod init geetard
```

Expected: `go.mod` contains `module geetard` and `go 1.26`.

- [ ] **Step 2: Write README**

```markdown
# terminal GEETARD

TUI tab reader. Search via freetar's public proxy, read the sheet, change theme.

```bash
go test ./...
go run ./cmd/geetard
```

Never talks to `ultimate-guitar.com`. See `docs/superpowers/specs/2026-08-17-terminal-geetard-design.md`.
```

- [ ] **Step 3: Commit**

```bash
git add go.mod README.md
git commit -m "chore: init geetard Go module"
```

---

### Task 2: Fixtures and js-store extract

**Files:**
- Create: `testdata/search_ok.html`
- Create: `testdata/tab_ok.html`
- Create: `testdata/tab_no_applicature.html`
- Create: `testdata/no_store.html`
- Create: `internal/client/types.go`
- Create: `internal/client/parse.go`
- Test: `internal/client/parse_test.go`

- [ ] **Step 1: Write fixtures**

`testdata/search_ok.html`:

```html
<html><body>
<div class="js-store" data-content="{&quot;store&quot;:{&quot;page&quot;:{&quot;data&quot;:{&quot;results&quot;:[{&quot;artist_name&quot;:&quot;Willie Nelson&quot;,&quot;song_name&quot;:&quot;Always On My Mind&quot;,&quot;type&quot;:&quot;Chords&quot;,&quot;version&quot;:&quot;3&quot;,&quot;votes&quot;:12,&quot;rating&quot;:4.8,&quot;tab_url&quot;:&quot;https://tabs.ultimate-guitar.com/tab/willie-nelson/always-on-my-mind-123&quot;},{&quot;artist_name&quot;:&quot;Paid Act&quot;,&quot;song_name&quot;:&quot;Official Hit&quot;,&quot;type&quot;:&quot;Official&quot;,&quot;version&quot;:1,&quot;votes&quot;:9,&quot;rating&quot;:5,&quot;tab_url&quot;:&quot;https://tabs.ultimate-guitar.com/tab/paid/official-1&quot;},{&quot;artist_name&quot;:&quot;Pro Shop&quot;,&quot;song_name&quot;:&quot;Pro Tab&quot;,&quot;type&quot;:&quot;Pro&quot;,&quot;version&quot;:1,&quot;votes&quot;:1,&quot;rating&quot;:5,&quot;tab_url&quot;:&quot;https://tabs.ultimate-guitar.com/tab/pro/pro-1&quot;}],&quot;pagination&quot;:{&quot;current&quot;:1,&quot;total&quot;:3}}}}}"></div>
</body></html>
```

`testdata/tab_ok.html` (body uses `[ch]` markup; applicature frets are low-E → high-e):

```html
<html><body>
<div class="js-store" data-content="{&quot;store&quot;:{&quot;page&quot;:{&quot;data&quot;:{&quot;tab&quot;:{&quot;artist_name&quot;:&quot;Willie Nelson&quot;,&quot;song_name&quot;:&quot;Always On My Mind&quot;,&quot;version&quot;:3,&quot;type&quot;:&quot;Chords&quot;,&quot;rating&quot;:&quot;5&quot;},&quot;tab_view&quot;:{&quot;wiki_tab&quot;:{&quot;content&quot;:&quot;[tab][ch]G[/ch]            [ch]D[/ch]\nMaybe I didn't love you[/tab]&quot;},&quot;ug_difficulty&quot;:&quot;novice&quot;,&quot;meta&quot;:{&quot;capo&quot;:1,&quot;tuning&quot;:{&quot;value&quot;:&quot;E A D G B E&quot;,&quot;name&quot;:&quot;Standard&quot;}},&quot;applicature&quot;:{&quot;G&quot;:[{&quot;frets&quot;:[3,2,0,0,0,3],&quot;fingers&quot;:[2,1,0,0,0,3]}]}}}}}}"></div>
</body></html>
```

`testdata/tab_no_applicature.html`: same as `tab_ok.html` but omit the `"applicature"` key entirely (keep wiki_tab + meta).

`testdata/no_store.html`:

```html
<html><body><p>no store here</p></body></html>
```

- [ ] **Step 2: Write the failing test**

`internal/client/parse_test.go`:

```go
package client

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdata(t *testing.T, name string) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	p := filepath.Join(filepath.Dir(file), "..", "..", "testdata", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestExtractStore_unescapesAndReturnsObject(t *testing.T) {
	store, err := ExtractStore(testdata(t, "search_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := store["store"].(map[string]any)["page"].(map[string]any)
	data := page["data"].(map[string]any)
	results := data["results"].([]any)
	if len(results) != 3 {
		t.Fatalf("got %d results", len(results))
	}
}

func TestExtractStore_missingDiv(t *testing.T) {
	_, err := ExtractStore(testdata(t, "no_store.html"))
	if err == nil || err.Error() != "page layout changed" {
		t.Fatalf("got %v", err)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/client -run TestExtractStore -v
```

Expected: FAIL, `ExtractStore` undefined.

- [ ] **Step 4: Minimal implementation**

`internal/client/types.go`:

```go
package client

import "fmt"

type ClientError struct{ Msg string }

func (e ClientError) Error() string { return e.Msg }

func errf(msg string, args ...any) error {
	if len(args) == 0 {
		return ClientError{Msg: msg}
	}
	return ClientError{Msg: fmt.Sprintf(msg, args...)}
}

type SearchPage struct {
	Query      string
	Page       int
	TotalPages int
	Results    []SearchResult
}

type SearchResult struct {
	Artist  string
	Song    string
	Type    string
	Version int
	Rating  float64
	Votes   int
	Path    string
}

type TabDetail struct {
	Artist     string
	Song       string
	Version    int
	Type       string
	Rating     float64
	Capo       string
	Tuning     string
	Difficulty string
	Raw        string
	Shapes     []ChordShape
}

type ChordShape struct {
	Name  string
	Lines []string
}

type Kind int

const (
	KindBody Kind = iota
	KindChord
)

type Part struct {
	Text string
	Kind Kind
}

type DisplayLine struct {
	Parts []Part
}
```

`internal/client/parse.go`:

```go
package client

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
)

var storeDiv = regexp.MustCompile(`(?is)<div[^>]*class="[^"]*\bjs-store\b[^"]*"[^>]*data-content="([^"]*)"`)

func ExtractStore(htmlPage string) (map[string]any, error) {
	m := storeDiv.FindStringSubmatch(htmlPage)
	if m == nil {
		return nil, errf("page layout changed")
	}
	raw := html.UnescapeString(m[1])
	if strings.TrimSpace(raw) == "" {
		return nil, errf("page layout changed")
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, errf("could not parse page")
	}
	return out, nil
}

func storeData(store map[string]any) (map[string]any, error) {
	s, _ := store["store"].(map[string]any)
	if s == nil {
		return nil, errf("could not parse page")
	}
	page, _ := s["page"].(map[string]any)
	if page == nil {
		return nil, errf("could not parse page")
	}
	data, _ := page["data"].(map[string]any)
	if data == nil {
		return nil, errf("could not parse page")
	}
	return data, nil
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./internal/client -run TestExtractStore -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add testdata internal/client
git commit -m "feat: extract js-store JSON from proxy HTML"
```

---

### Task 3: Flexible numbers

**Files:**
- Modify: `internal/client/parse.go`
- Test: `internal/client/parse_test.go`

- [ ] **Step 1: Write the failing test**

Append to `parse_test.go`:

```go
func TestAsInt_stringOrNumber(t *testing.T) {
	if AsInt("3", 0) != 3 || AsInt(3.0, 0) != 3 || AsInt(nil, 7) != 7 {
		t.Fatal(AsInt("3", 0), AsInt(3.0, 0), AsInt(nil, 7))
	}
}

func TestAsFloat_stringOrNumber(t *testing.T) {
	if AsFloat("4.8", 0) != 4.8 || AsFloat(5, 0) != 5 || AsFloat(nil, 1) != 1 {
		t.Fatal(AsFloat("4.8", 0), AsFloat(5, 0), AsFloat(nil, 1))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client -run 'TestAsInt|TestAsFloat' -v
```

Expected: FAIL, `AsInt` / `AsFloat` undefined.

- [ ] **Step 3: Minimal implementation**

Append to `parse.go`:

```go
func AsInt(v any, def int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return def
		}
		return int(i)
	case string:
		var i int
		if _, err := fmt.Sscanf(n, "%d", &i); err != nil {
			return def
		}
		return i
	default:
		return def
	}
}

func AsFloat(v any, def float64) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return def
		}
		return f
	case string:
		var f float64
		if _, err := fmt.Sscanf(n, "%f", &f); err != nil {
			return def
		}
		return f
	default:
		return def
	}
}
```

Add `"fmt"` to the `parse.go` import list.

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/client -run 'TestAsInt|TestAsFloat' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/client/parse.go internal/client/parse_test.go
git commit -m "feat: parse UG numbers that arrive as strings"
```

---

### Task 4: Map search results (drop Pro/Official)

**Files:**
- Create: `internal/client/map.go`
- Test: `internal/client/map_test.go`

- [ ] **Step 1: Write the failing test**

```go
package client

import "testing"

func TestMapSearch_dropsProOfficialAndKeepsChords(t *testing.T) {
	store, err := ExtractStore(testdata(t, "search_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	page, err := MapSearch(store, "nelson", 1)
	if err != nil {
		t.Fatal(err)
	}
	if page.Query != "nelson" || page.Page != 1 || page.TotalPages != 3 {
		t.Fatalf("%+v", page)
	}
	if len(page.Results) != 1 {
		t.Fatalf("want 1 result, got %+v", page.Results)
	}
	r := page.Results[0]
	if r.Artist != "Willie Nelson" || r.Song != "Always On My Mind" || r.Type != "Chords" {
		t.Fatalf("%+v", r)
	}
	if r.Version != 3 || r.Votes != 12 || r.Rating != 4.8 {
		t.Fatalf("numbers %+v", r)
	}
	if r.Path != "willie-nelson/always-on-my-mind-123" {
		t.Fatalf("path %q", r.Path)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client -run TestMapSearch -v
```

Expected: FAIL, `MapSearch` undefined.

- [ ] **Step 3: Minimal implementation**

`internal/client/url.go` (path helper used by map):

```go
package client

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultSearchHost = "https://proxy.freetar.de"
	DefaultTabHost    = "https://tabs.proxy.freetar.de"
)

func TabPath(tabURL string) string {
	u, err := url.Parse(tabURL)
	path := tabURL
	if err == nil && u.Path != "" {
		path = u.Path
	}
	path = strings.TrimPrefix(path, "/tab/")
	path = strings.TrimPrefix(path, "tab/")
	return strings.TrimPrefix(path, "/")
}

func SearchURL(host, query string, page int) (string, error) {
	if err := rejectUG(host); err != nil {
		return "", err
	}
	u, err := url.Parse(strings.TrimRight(host, "/") + "/search.php")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("search_type", "title")
	q.Set("value", query)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func TabURL(host, path string) (string, error) {
	if err := rejectUG(host); err != nil {
		return "", err
	}
	return strings.TrimRight(host, "/") + "/tab/" + TabPath(path), nil
}

func rejectUG(host string) error {
	if strings.Contains(strings.ToLower(host), "ultimate-guitar.com") {
		return errf("refusing ultimate-guitar.com")
	}
	return nil
}
```

`internal/client/map.go`:

```go
package client

import "strings"

func MapSearch(store map[string]any, query string, page int) (SearchPage, error) {
	data, err := storeData(store)
	if err != nil {
		return SearchPage{}, err
	}
	raw, _ := data["results"].([]any)
	var results []SearchResult
	for _, item := range raw {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		typ, _ := m["type"].(string)
		if strings.EqualFold(typ, "Pro") || strings.EqualFold(typ, "Official") {
			continue
		}
		tabURL, _ := m["tab_url"].(string)
		path := TabPath(tabURL)
		if path == "" {
			continue
		}
		artist, _ := m["artist_name"].(string)
		song, _ := m["song_name"].(string)
		if artist == "" {
			artist = "Unknown artist"
		}
		if song == "" {
			song = "Unknown song"
		}
		results = append(results, SearchResult{
			Artist:  artist,
			Song:    song,
			Type:    typ,
			Version: AsInt(m["version"], 1),
			Rating:  AsFloat(m["rating"], 0),
			Votes:   AsInt(m["votes"], 0),
			Path:    path,
		})
	}
	pag, _ := data["pagination"].(map[string]any)
	total := 1
	cur := page
	if pag != nil {
		total = AsInt(pag["total"], 1)
		cur = AsInt(pag["current"], page)
	}
	return SearchPage{Query: query, Page: cur, TotalPages: total, Results: results}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/client -run TestMapSearch -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/client/url.go internal/client/map.go internal/client/map_test.go
git commit -m "feat: map search results and drop Pro/Official"
```

---

### Task 5: Reject UG hosts and build URLs

**Files:**
- Test: `internal/client/url_test.go`
- Modify: `internal/client/url.go` (already created)

- [ ] **Step 1: Write the failing test**

```go
package client

import (
	"strings"
	"testing"
)

func TestSearchURL_usesProxyAndNeverUG(t *testing.T) {
	u, err := SearchURL(DefaultSearchHost, "nelson always", 2)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(u, "ultimate-guitar.com") {
		t.Fatal(u)
	}
	if !strings.HasPrefix(u, "https://proxy.freetar.de/search.php?") {
		t.Fatal(u)
	}
	if !strings.Contains(u, "search_type=title") || !strings.Contains(u, "page=2") {
		t.Fatal(u)
	}
}

func TestSearchURL_rejectsUGHost(t *testing.T) {
	_, err := SearchURL("https://www.ultimate-guitar.com", "x", 1)
	if err == nil {
		t.Fatal("expected reject")
	}
}

func TestTabURL_stripsPrefix(t *testing.T) {
	u, err := TabURL(DefaultTabHost, "/tab/willie-nelson/always-on-my-mind-123")
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://tabs.proxy.freetar.de/tab/willie-nelson/always-on-my-mind-123" {
		t.Fatal(u)
	}
}
```

- [ ] **Step 2: Run tests**

```bash
go test ./internal/client -run 'TestSearchURL|TestTabURL' -v
```

Expected: PASS (implementation already in Task 4). If `SearchURL` query escaping uses `+` vs `%20`, assert with `url.Parse` instead of a raw substring on the query — do not weaken the UG rejection checks.

- [ ] **Step 3: Commit**

```bash
git add internal/client/url_test.go
git commit -m "test: proxy URLs never touch ultimate-guitar.com"
```

---

### Task 6: Map tab detail

**Files:**
- Modify: `internal/client/map.go`
- Test: `internal/client/map_test.go`

- [ ] **Step 1: Write the failing test**

Append to `map_test.go`:

```go
func TestMapTab_readsMetaAndRaw(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	tab, err := MapTab(store)
	if err != nil {
		t.Fatal(err)
	}
	if tab.Artist != "Willie Nelson" || tab.Song != "Always On My Mind" {
		t.Fatalf("%+v", tab)
	}
	if tab.Version != 3 || tab.Type != "Chords" || tab.Rating != 5 {
		t.Fatalf("meta %+v", tab)
	}
	if tab.Capo != "1" || tab.Tuning != "E A D G B E (Standard)" || tab.Difficulty != "novice" {
		t.Fatalf("extra %+v", tab)
	}
	if !strings.Contains(tab.Raw, "[ch]G[/ch]") {
		t.Fatalf("raw %q", tab.Raw)
	}
}

func TestMapTab_missingApplicatureStillWorks(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_no_applicature.html"))
	if err != nil {
		t.Fatal(err)
	}
	tab, err := MapTab(store)
	if err != nil {
		t.Fatal(err)
	}
	if tab.Raw == "" {
		t.Fatal("empty raw")
	}
}
```

Add `"strings"` to the `map_test.go` import list.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client -run TestMapTab -v
```

Expected: FAIL, `MapTab` undefined.

- [ ] **Step 3: Minimal implementation**

Append to `map.go` (add `"strconv"` to imports):

```go
func MapTab(store map[string]any) (TabDetail, error) {
	data, err := storeData(store)
	if err != nil {
		return TabDetail{}, err
	}
	tab, _ := data["tab"].(map[string]any)
	view, _ := data["tab_view"].(map[string]any)
	if tab == nil || view == nil {
		return TabDetail{}, errf("could not parse page")
	}
	wiki, _ := view["wiki_tab"].(map[string]any)
	raw, _ := wiki["content"].(string)
	artist, _ := tab["artist_name"].(string)
	song, _ := tab["song_name"].(string)
	typ, _ := tab["type"].(string)
	diff, _ := view["ug_difficulty"].(string)
	tuning := ""
	if meta, ok := view["meta"].(map[string]any); ok {
		if tun, ok := meta["tuning"].(map[string]any); ok {
			val, _ := tun["value"].(string)
			name, _ := tun["name"].(string)
			switch {
			case val != "" && name != "":
				tuning = val + " (" + name + ")"
			case val != "":
				tuning = val
			default:
				tuning = name
			}
		}
	}
	if artist == "" {
		artist = "Unknown artist"
	}
	if song == "" {
		song = "Unknown song"
	}
	return TabDetail{
		Artist:     artist,
		Song:       song,
		Version:    AsInt(tab["version"], 1),
		Type:       typ,
		Rating:     AsFloat(tab["rating"], 0),
		Capo:       formatCapo(view),
		Tuning:     tuning,
		Difficulty: diff,
		Raw:        raw,
	}, nil
}

func formatCapo(view map[string]any) string {
	meta, ok := view["meta"].(map[string]any)
	if !ok {
		return ""
	}
	v, exists := meta["capo"]
	if !exists || v == nil {
		return ""
	}
	if n, ok := v.(string); ok {
		return n
	}
	return strconv.Itoa(AsInt(v, 0))
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/client -run TestMapTab -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/client/map.go internal/client/map_test.go
git commit -m "feat: map tab metadata and raw wiki_tab content"
```

---

### Task 7: Render `[ch]` markup

**Files:**
- Create: `internal/client/render.go`
- Test: `internal/client/render_test.go`

- [ ] **Step 1: Write the failing test**

```go
package client

import "testing"

func TestRender_splitsChordsAndBody(t *testing.T) {
	lines := Render("[tab][ch]Em[/ch]  [ch]G[/ch]\nverse line[/tab]")
	if len(lines) != 2 {
		t.Fatalf("%+v", lines)
	}
	if len(lines[0].Parts) != 3 ||
		lines[0].Parts[0] != (Part{Text: "Em", Kind: KindChord}) ||
		lines[0].Parts[1] != (Part{Text: "  ", Kind: KindBody}) ||
		lines[0].Parts[2] != (Part{Text: "G", Kind: KindChord}) {
		t.Fatalf("line0 %+v", lines[0].Parts)
	}
	if len(lines[1].Parts) != 1 || lines[1].Parts[0].Text != "verse line" || lines[1].Parts[0].Kind != KindBody {
		t.Fatalf("line1 %+v", lines[1].Parts)
	}
}

func TestRender_empty(t *testing.T) {
	if len(Render("")) != 0 || len(Render("[tab][/tab]")) != 0 {
		t.Fatal(Render(""), Render("[tab][/tab]"))
	}
}

func TestRender_noHTMLEntities(t *testing.T) {
	lines := Render("hello  world")
	if len(lines) != 1 || lines[0].Parts[0].Text != "hello  world" {
		t.Fatalf("%+v", lines)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client -run TestRender -v
```

Expected: FAIL, `Render` undefined.

- [ ] **Step 3: Minimal implementation**

```go
package client

import (
	"regexp"
	"strings"
)

var chordRE = regexp.MustCompile(`\[ch\](.*?)\[/ch\]`)

func Render(raw string) []DisplayLine {
	s := strings.ReplaceAll(raw, "[tab]", "")
	s = strings.ReplaceAll(s, "[/tab]", "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if s == "" {
		return nil
	}
	var lines []DisplayLine
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, DisplayLine{Parts: splitParts(line)})
	}
	if len(lines) == 1 && len(lines[0].Parts) == 0 {
		return nil
	}
	return lines
}

func splitParts(s string) []Part {
	var parts []Part
	last := 0
	for _, m := range chordRE.FindAllStringSubmatchIndex(s, -1) {
		if m[0] > last {
			parts = append(parts, Part{Text: s[last:m[0]], Kind: KindBody})
		}
		inner := s[m[2]:m[3]]
		if inner != "" {
			parts = append(parts, Part{Text: inner, Kind: KindChord})
		}
		last = m[1]
	}
	if last < len(s) {
		parts = append(parts, Part{Text: s[last:], Kind: KindBody})
	}
	return parts
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/client -run TestRender -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/client/render.go internal/client/render_test.go
git commit -m "feat: render [ch] markup into display lines"
```

---

### Task 8: ASCII chord shapes

**Files:**
- Create: `internal/client/shapes.go`
- Modify: `internal/client/map.go` (`MapTab` fills `Shapes`)
- Test: `internal/client/shapes_test.go`

- [ ] **Step 1: Write the failing test**

```go
package client

import (
	"strings"
	"testing"
)

func TestShapes_gFromFixture(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	view := store["store"].(map[string]any)["page"].(map[string]any)["data"].(map[string]any)["tab_view"].(map[string]any)
	shapes := Shapes(view)
	if len(shapes) != 1 || shapes[0].Name != "G" {
		t.Fatalf("%+v", shapes)
	}
	joined := strings.Join(shapes[0].Lines, "\n")
	want := strings.Join([]string{
		"  e |-3-",
		"  B |-0-",
		"  G |-0-",
		"  D |-0-",
		"  A |-2-",
		"  E |-3-",
	}, "\n")
	if joined != want {
		t.Fatalf("got\n%s\nwant\n%s", joined, want)
	}
}

func TestShapes_missingIsEmpty(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_no_applicature.html"))
	if err != nil {
		t.Fatal(err)
	}
	view := store["store"].(map[string]any)["page"].(map[string]any)["data"].(map[string]any)["tab_view"].(map[string]any)
	if len(Shapes(view)) != 0 {
		t.Fatal(Shapes(view))
	}
}

func TestShapes_muteIsX(t *testing.T) {
	view := map[string]any{
		"applicature": map[string]any{
			"Em": []any{map[string]any{"frets": []any{0.0, 2.0, 2.0, 0.0, 0.0, -1.0}}},
		},
	}
	shapes := Shapes(view)
	if len(shapes) != 1 || !strings.Contains(strings.Join(shapes[0].Lines, "\n"), "  e |-x-") {
		t.Fatalf("%+v", shapes)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client -run TestShapes -v
```

Expected: FAIL, `Shapes` undefined.

- [ ] **Step 3: Minimal implementation**

```go
package client

import "fmt"

var stringNames = []string{"e", "B", "G", "D", "A", "E"}

func Shapes(tabView map[string]any) []ChordShape {
	app, _ := tabView["applicature"].(map[string]any)
	if app == nil {
		return nil
	}
	var out []ChordShape
	for name, raw := range app {
		variants, _ := raw.([]any)
		if len(variants) == 0 {
			continue
		}
		first, _ := variants[0].(map[string]any)
		if first == nil {
			continue
		}
		fretsAny, _ := first["frets"].([]any)
		if len(fretsAny) < 6 {
			continue
		}
		// frets[0] is low E. Display high e on top.
		lines := make([]string, 6)
		for i := 0; i < 6; i++ {
			fret := AsInt(fretsAny[5-i], 0)
			cell := "0"
			if fret < 0 {
				cell = "x"
			} else {
				cell = fmt.Sprintf("%d", fret)
			}
			lines[i] = fmt.Sprintf("  %s |-%s-", stringNames[i], cell)
		}
		out = append(out, ChordShape{Name: name, Lines: lines})
	}
	return out
}
```

In `MapTab`, after building `TabDetail`, set `Shapes: Shapes(view)` on the returned struct.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/client -run 'TestShapes|TestMapTab' -v
```

Expected: PASS. Map iteration order of `app` is random — `TestShapes_gFromFixture` only has one chord so it is stable. Do not assert shape slice order across multiple names.

- [ ] **Step 5: Commit**

```bash
git add internal/client/shapes.go internal/client/shapes_test.go internal/client/map.go
git commit -m "feat: build ASCII chord diagrams from applicature"
```

---

### Task 9: HTTP client (httptest only)

**Files:**
- Create: `internal/client/http.go`
- Test: `internal/client/http_test.go`

- [ ] **Step 1: Write the failing test**

```go
package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientSearch_emptyQueryDoesNotHitNetwork(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	page, err := c.Search("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 0 || len(page.Results) != 0 {
		t.Fatalf("hits=%d page=%+v", hits, page)
	}
}

func TestClientSearch_parsesFixture(t *testing.T) {
	html := testdata(t, "search_ok.html")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search.php" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	page, err := c.Search("nelson", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Results) != 1 || page.Results[0].Artist != "Willie Nelson" {
		t.Fatalf("%+v", page)
	}
}

func TestClientTab_parsesFixture(t *testing.T) {
	html := testdata(t, "tab_ok.html")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tab/willie-nelson/always-on-my-mind-123" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	tab, err := c.Tab("willie-nelson/always-on-my-mind-123")
	if err != nil {
		t.Fatal(err)
	}
	if tab.Song != "Always On My Mind" || len(tab.Shapes) != 1 {
		t.Fatalf("%+v", tab)
	}
}

func TestClientSearch_5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	c.HTTP.Timeout = time.Second
	_, err := c.Search("x", 1)
	if err == nil || err.Error() != "proxy down (503)" {
		t.Fatalf("got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client -run TestClient -v
```

Expected: FAIL, `New` undefined.

- [ ] **Step 3: Minimal implementation**

```go
package client

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	SearchHost string
	TabHost    string
	HTTP       *http.Client
	UA         string
}

func New(searchHost, tabHost string) *Client {
	if searchHost == "" {
		searchHost = DefaultSearchHost
	}
	if tabHost == "" {
		tabHost = DefaultTabHost
	}
	return &Client{
		SearchHost: searchHost,
		TabHost:    tabHost,
		HTTP:       &http.Client{Timeout: 15 * time.Second},
		UA:         "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0",
	}
}

func FromEnv() *Client {
	return New(os.Getenv("FREETAR_SEARCH_HOST"), os.Getenv("FREETAR_TAB_HOST"))
}

func (c *Client) Search(query string, page int) (SearchPage, error) {
	if query == "" {
		return SearchPage{Query: query, Page: page, TotalPages: 1}, nil
	}
	if page < 1 {
		page = 1
	}
	u, err := SearchURL(c.SearchHost, query, page)
	if err != nil {
		return SearchPage{}, err
	}
	body, err := c.get(u)
	if err != nil {
		return SearchPage{}, err
	}
	store, err := ExtractStore(body)
	if err != nil {
		return SearchPage{}, err
	}
	return MapSearch(store, query, page)
}

func (c *Client) Tab(path string) (TabDetail, error) {
	u, err := TabURL(c.TabHost, path)
	if err != nil {
		return TabDetail{}, err
	}
	body, err := c.get(u)
	if err != nil {
		return TabDetail{}, err
	}
	store, err := ExtractStore(body)
	if err != nil {
		return TabDetail{}, err
	}
	return MapTab(store)
}

func (c *Client) get(u string) (string, error) {
	var last error
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", c.UA)
		resp, err := c.HTTP.Do(req)
		if err != nil {
			last = errf("proxy timed out")
			continue
		}
		b, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			last = errf("proxy timed out")
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			return "", errf("not found")
		}
		if resp.StatusCode >= 500 {
			last = errf("proxy down (%d)", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return "", errf("proxy down (%d)", resp.StatusCode)
		}
		return string(b), nil
	}
	if last == nil {
		last = errf("proxy timed out")
	}
	return "", last
}
```

Remove unused `"fmt"` if the compiler complains; `errf` already lives in `types.go`.

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/client -run TestClient -v
```

Expected: PASS. 5xx retries twice and still returns `proxy down (503)`.

- [ ] **Step 5: Commit**

```bash
git add internal/client/http.go internal/client/http_test.go
git commit -m "feat: HTTP client for search and tab via httptest"
```

---

### Task 10: selections.json config

**Files:**
- Create: `testdata/selections_ok.json`
- Create: `testdata/selections_bad.json`
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write fixtures and failing tests**

`testdata/selections_ok.json`:

```json
{
  "theme": "nord",
  "border": "ascii",
  "density": "airy",
  "autoscroll": "on",
  "scroll_speed": "fast",
  "diagrams": "sidebar"
}
```

`testdata/selections_bad.json`:

```json
{ this is not json
```

`internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	s := Defaults()
	if s.Theme != "amber" || s.Border != "ascii" || s.Density != "normal" ||
		s.Autoscroll != "off" || s.ScrollSpeed != "medium" || s.Diagrams != "off" {
		t.Fatalf("%+v", s)
	}
}

func TestLoad_ok(t *testing.T) {
	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "selections_ok.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "selections.json"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	s, warn, err := Load(dir)
	if err != nil || warn != "" {
		t.Fatal(err, warn)
	}
	if s.Theme != "nord" || s.Diagrams != "sidebar" || s.Autoscroll != "on" {
		t.Fatalf("%+v", s)
	}
}

func TestLoad_missingUsesDefaults(t *testing.T) {
	s, warn, err := Load(t.TempDir())
	if err != nil || warn != "" || s.Theme != "amber" {
		t.Fatal(s, warn, err)
	}
}

func TestLoad_corrupt(t *testing.T) {
	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "selections_bad.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "selections.json"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	s, warn, err := Load(dir)
	if err != nil || warn != "bad config, using defaults" || s.Theme != "amber" {
		t.Fatal(s, warn, err)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := Defaults()
	in.Theme = "green"
	if err := Save(dir, in); err != nil {
		t.Fatal(err)
	}
	out, warn, err := Load(dir)
	if err != nil || warn != "" || out.Theme != "green" {
		t.Fatal(out, warn, err)
	}
}

func TestIntervalMs(t *testing.T) {
	if IntervalMs("slow") != 800 || IntervalMs("medium") != 400 || IntervalMs("fast") != 200 {
		t.Fatal(IntervalMs("slow"), IntervalMs("medium"), IntervalMs("fast"))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/config -v
```

Expected: FAIL, package undefined symbols.

- [ ] **Step 3: Minimal implementation**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Selections struct {
	Theme       string `json:"theme"`
	Border      string `json:"border"`
	Density     string `json:"density"`
	Autoscroll  string `json:"autoscroll"`
	ScrollSpeed string `json:"scroll_speed"`
	Diagrams    string `json:"diagrams"`
}

func Defaults() Selections {
	return Selections{
		Theme:       "amber",
		Border:      "ascii",
		Density:     "normal",
		Autoscroll:  "off",
		ScrollSpeed: "medium",
		Diagrams:    "off",
	}
}

func Dir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "geetard")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "geetard")
}

func Load(dir string) (Selections, string, error) {
	def := Defaults()
	b, err := os.ReadFile(filepath.Join(dir, "selections.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return def, "", nil
		}
		return def, "", err
	}
	var s Selections
	if json.Unmarshal(b, &s) != nil {
		return def, "bad config, using defaults", nil
	}
	return applyDefaults(s), "", nil
}

func applyDefaults(s Selections) Selections {
	d := Defaults()
	if s.Theme == "" {
		s.Theme = d.Theme
	}
	if s.Border == "" {
		s.Border = d.Border
	}
	if s.Density == "" {
		s.Density = d.Density
	}
	if s.Autoscroll == "" {
		s.Autoscroll = d.Autoscroll
	}
	if s.ScrollSpeed == "" {
		s.ScrollSpeed = d.ScrollSpeed
	}
	if s.Diagrams == "" {
		s.Diagrams = d.Diagrams
	}
	return s
}

func Save(dir string, s Selections) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "selections.json"), append(b, '\n'), 0o644)
}

func IntervalMs(speed string) int {
	switch speed {
	case "slow":
		return 800
	case "fast":
		return 200
	default:
		return 400
	}
}

func DensityGap(density string) int {
	switch density {
	case "compact":
		return 0
	case "airy":
		return 2
	default:
		return 1
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/config -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add testdata/selections_ok.json testdata/selections_bad.json internal/config
git commit -m "feat: load and save geetard selections.json"
```

---

### Task 11: Themes (three hex roles, no background)

**Files:**
- Create: `internal/theme/theme.go`
- Test: `internal/theme/theme_test.go`

- [ ] **Step 1: Write the failing test**

```go
package theme

import "testing"

func TestNamed_nordMatchesTtuneTrio(t *testing.T) {
	th, ok := Named("nord")
	if !ok || th.Primary != "#88C0D0" || th.Secondary != "#D8DEE9" || th.Tertiary != "#EBCB8B" {
		t.Fatalf("%+v", th)
	}
}

func TestNamed_unknown(t *testing.T) {
	if _, ok := Named("papaya"); ok {
		t.Fatal("expected miss")
	}
}

func TestCycle(t *testing.T) {
	if Cycle("amber") != "green" || Cycle("nord") != "amber" {
		t.Fatal(Cycle("amber"), Cycle("nord"))
	}
}

func TestApply_doesNotSetBackground(t *testing.T) {
	st := Apply(Must("amber"))
	if st.Body.GetBackground() != "" && st.Body.GetBackground() != nil {
		// lipgloss may return nil; a set background is the bug
		if s, ok := st.Body.GetBackground().(string); ok && s != "" {
			t.Fatalf("painted background %q", s)
		}
	}
}
```

If `GetBackground` is awkward in Lipgloss v2, replace the last test with: `Apply` returns styles whose `Foreground` is the secondary/tertiary hex, and never call `.Background(...)` in `theme.go`. Grep the file in the test:

```go
func TestApply_setsForegroundRoles(t *testing.T) {
	st := Apply(Must("amber"))
	if st.Chord.GetForeground() == nil {
		t.Fatal("chord has no foreground")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go get charm.land/lipgloss/v2@v2.0.0
go test ./internal/theme -v
```

Expected: FAIL, `Named` undefined.

- [ ] **Step 3: Minimal implementation**

```go
package theme

import "charm.land/lipgloss/v2"

type Theme struct {
	Name      string
	Primary   string
	Secondary string
	Tertiary  string
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
```

Adjust `TestApply_*` to compile against Lipgloss v2. If `GetForeground` does not exist, drop that assertion and keep `Named` / `Cycle` tests only.

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/theme -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/theme
git commit -m "feat: four CRT themes with ttune color roles"
```

---

### Task 12: TUI model and keys

**Files:**
- Create: `internal/tui/model.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

```go
package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"geetard/internal/client"
	"geetard/internal/config"
)

func TestNew_startsOnSearch(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	if m.State != StateSearch {
		t.Fatalf("%s", m.State)
	}
}

func TestKeys_settingsAndHelpAndBack(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m, _ = m.Update(tea.KeyPressMsg{Code: 's'})
	if m.(Model).State != StateSettings {
		t.Fatal(m.(Model).State)
	}
	m, _ = m.(Model).Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.(Model).State != StateSearch {
		t.Fatal(m.(Model).State)
	}
	m, _ = m.(Model).Update(tea.KeyPressMsg{Code: '?'})
	if m.(Model).State != StateHelp {
		t.Fatal(m.(Model).State)
	}
}

func TestKeys_themeCycle(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m, _ = m.Update(tea.KeyPressMsg{Code: 't'})
	if m.(Model).Sel.Theme != "green" {
		t.Fatal(m.(Model).Sel.Theme)
	}
}
```

Bubble Tea v2 key types may be `tea.KeyMsg` with `.String()` `"s"` / `"esc"` / `"?"` — match **ttune** (`/tmp/ttune-src/ttune-master/model.go` after line 149). If `KeyPressMsg` does not compile, use the same `tea.KeyMsg` + `message.String()` switch ttune uses. Rewrite the tests to send whatever type `Update` actually receives.

- [ ] **Step 2: Run test to verify it fails**

```bash
go get charm.land/bubbletea/v2@v2.0.0
go get charm.land/bubbles/v2@v2.0.0
go test ./internal/tui -v
```

Expected: FAIL, `New` undefined.

- [ ] **Step 3: Minimal implementation**

`internal/tui/model.go` — keep this file as the state machine only. Follow ttune: `Init`, `Update`, `View() tea.View` with `AltScreen = true`.

Required fields:

```go
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
```

`Update` handles `tea.WindowSizeMsg` and keys:

- `q` → `tea.Quit` from search/settings/help; from reader → `StateSearch`
- `s` → `Prev = State; State = StateSettings`
- `?` → `Prev = State; State = StateHelp`
- `esc` / `backspace` from settings/help → `State = Prev`
- `t` on search/reader → `Sel.Theme = theme.Cycle(Sel.Theme)` and `config.Save`
- `/` on search → `Querying = true`

`View()` returns `tea.NewView` with `AltScreen: true` and placeholder content per state (`searchView`, `readerView`, `settingsView`, `helpView` can live in this file until later tasks split them).

`New(c *client.Client, sel config.Selections, dir string) Model` sets `StateSearch`, `AutoOn` from `sel.Autoscroll == "on"`, `AutoMs` from `config.IntervalMs`.

Copy ttune’s key extraction — do not invent a new key API.

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/tui -v
```

Expected: PASS. Fix compile errors against Bubble Tea v2 by matching ttune, not by deleting tests.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/tui
git commit -m "feat: TUI state machine with ttune-style keys"
```

---

### Task 13: Search screen

**Files:**
- Create: `internal/tui/search.go`
- Modify: `internal/tui/model.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestSearch_enterOnEmptyDoesNotSearch(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Querying = true
	m.Query = ""
	nm, cmd := m.Update(enterKey())
	if cmd != nil {
		t.Fatal("empty query must not fire a command")
	}
	_ = nm
}

func TestSearchView_listsResults(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Width, m.Height = 80, 24
	m.Results = client.SearchPage{Results: []client.SearchResult{{
		Artist: "Willie Nelson", Song: "Always On My Mind", Type: "Chords", Rating: 4.8,
	}}}
	v := m.View()
	if !strings.Contains(viewString(v), "Willie Nelson") {
		t.Fatal(viewString(v))
	}
}
```

Helpers `enterKey()` and `viewString(tea.View) string` should wrap whatever Bubble Tea v2 exposes (`v` content field or `fmt.Sprint`). Look at `tea.View` in the module cache.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/tui -run TestSearch -v
```

Expected: FAIL until search view lists rows.

- [ ] **Step 3: Implement search**

`search.go`:

- Title box: `terminal GEETARD — search`
- Query line prefixed `/ `
- Table columns ARTIST SONG TYPE ★
- Footer: `enter open   n/p page   / search   s settings   ? help   q quit`
- `j`/`k` move `Cursor` in `Results.Results`
- `n`/`p` issue `SearchCmd` with page±1 (clamp 1..TotalPages)
- Enter on a result issues `TabCmd(path)`
- Enter in query with non-empty Query issues `SearchCmd(query, 1)`

`SearchCmd` / `TabCmd` are `tea.Cmd` functions that call `m.Client.Search` / `Tab` and return a `searchMsg` / `tabMsg` or `errMsg`. Handle those in `Update`: set `Results` or `Tab`+`Lines`+`StateReader`, or `Status` on error.

Use `theme.Apply` + `theme.SetBorder` from `m.Sel`.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/tui -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui
git commit -m "feat: search screen lists proxy results"
```

---

### Task 14: Reader, autoscroll, density, diagrams

**Files:**
- Create: `internal/tui/reader.go`
- Modify: `internal/tui/model.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestReader_autoscrollTickAdvances(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.State = StateReader
	m.AutoOn = true
	m.Lines = make([]client.DisplayLine, 40)
	m.Height = 10
	m, _ = m.Update(autoTick{})
	if m.(Model).Scroll != 1 {
		t.Fatal(m.(Model).Scroll)
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/tui -run TestReader -v
```

Expected: FAIL.

- [ ] **Step 3: Implement reader**

- Title: `terminal GEETARD — read` plus `Artist — Song  vN  capo X`
- Render `Lines` with `theme.Styles.Chord` / `.Body`
- Insert `DensityGap(m.Sel.Density)` blank lines between sheet lines
- `j`/`k` change `Scroll`; space pages by `Height-8`
- `a` toggles `AutoOn`; `[` / `]` adjust `AutoMs` by 50, clamp 100–2000
- When `AutoOn`, `Init`/toggle starts `tea.Tick` of `AutoMs` that sends `autoTick{}` and reschedules if still on
- Footer: `j/k scroll   space page   a autoscroll   s settings   ? help   q back` and `autoscroll * 400ms` when on
- Diagrams `off`: sheet only. `sidebar`: JoinHorizontal sheet + shapes if `Width >= 80`. `below`: shapes under the title, above the sheet
- `q` returns to search (does not quit)

- [ ] **Step 4: Run tests**

```bash
go test ./internal/tui -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui
git commit -m "feat: reader with autoscroll density and diagrams"
```

---

### Task 15: Settings screen

**Files:**
- Create: `internal/tui/settings.go`
- Modify: `internal/tui/model.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestSettings_selectThemeWritesFile(t *testing.T) {
	dir := t.TempDir()
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), dir)
	m.State = StateSettings
	m.SettingIdx = 0 // Theme
	m.InValues = true
	m.ValueIdx = 3 // nord in theme.Order
	m, _ = m.Update(enterKey())
	got := m.(Model)
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/tui -run TestSettings -v
```

Expected: FAIL.

- [ ] **Step 3: Implement settings**

Catalog in order (indices used by tests):

0. Theme — `theme.Order`
1. Border — `ascii`, `normal`, `rounded`, `none`
2. Density — `compact`, `normal`, `airy`
3. Autoscroll — `off`, `on`
4. Scroll speed — `slow`, `medium`, `fast`
5. Diagrams — `off`, `sidebar`, `below`

Layout like ttune: left names with `[ ]`/`[o]`, options list, description, preview.

- Theme preview: `G        D        Em` / `Maybe I didn't love you` using hovered theme styles
- Diagrams preview: a hard-coded G shape
- Density preview: three lines with the hovered gap

Enter/space on a value: write the field on `Sel`, `config.Save(ConfigDir, Sel)`, apply immediately (`AutoOn`/`AutoMs` if those fields). `h`/`l` toggle `InValues`. `j`/`k` move in the active pane.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/tui -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui
git commit -m "feat: ttune-style settings with live preview"
```

---

### Task 16: Help screen and main

**Files:**
- Create: `internal/tui/help.go`
- Create: `cmd/geetard/main.go`
- Modify: `internal/tui/model.go` (help view)
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestHelpView_mentionsProxyAndKeys(t *testing.T) {
	m := New(client.New("http://127.0.0.1:1", "http://127.0.0.1:1"), config.Defaults(), t.TempDir())
	m.Width, m.Height = 80, 24
	m.State = StateHelp
	body := viewString(m.View())
	for _, w := range []string{"proxy.freetar.de", "ultimate-guitar.com", "s settings", "a autoscroll"} {
		if !strings.Contains(body, w) {
			t.Fatal("missing", w, body)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/tui -run TestHelp -v
```

Expected: FAIL.

- [ ] **Step 3: Implement help + main**

`help.go` one page:

```
terminal GEETARD talks to proxy.freetar.de and tabs.proxy.freetar.de.
It never requests ultimate-guitar.com.

/ search     enter open     n/p page
j/k scroll   space page     a autoscroll
[ ] speed    s settings     t theme
? help       q back/quit
```

`cmd/geetard/main.go`:

```go
package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"geetard/internal/client"
	"geetard/internal/config"
	"geetard/internal/tui"
)

func main() {
	dir := config.Dir()
	sel, warn, err := config.Load(dir)
	if err != nil {
		log.Fatal(err)
	}
	m := tui.New(client.FromEnv(), sel, dir)
	m.Status = warn
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Println("Unable to run tui:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Run all tests and a compile check**

```bash
go test ./...
go build -o /tmp/geetard ./cmd/geetard
```

Expected: all PASS, binary builds.

- [ ] **Step 5: Commit**

```bash
git add internal/tui cmd/geetard README.md
git commit -m "feat: help screen and geetard entrypoint"
```

---

### Task 17: Manual check (human, not CI)

**Files:** none.

- [ ] **Step 1: Run the TUI against the public proxy**

```bash
go run ./cmd/geetard
```

- Search `nelson always`
- Open a Chords result
- Confirm chords are tertiary-colored, no ads
- `t` cycles amber → green → mono → nord
- `s` → Diagrams → sidebar, Autoscroll → on
- Confirm `~/.config/geetard/selections.json` updated
- `a` and `[` `]` change the footer speed
- `q` from reader returns to search; `q` from search quits
- Confirm no request to `ultimate-guitar.com` (footer/help text is enough; optional: `ss` / mitm not required)

- [ ] **Step 2: If something fails, fix with a test first, then commit the fix**

Do not skip this. The spec’s success line is this session, not the unit tests alone.

---

## Self-review (plan vs spec)

| Spec item | Task |
|---|---|
| `search` / `tab` / never UG | 4, 5, 9 |
| flexible numbers, drop Pro/Official | 3, 4 |
| `render` / no HTML entities | 7 |
| `shapes` from applicature | 8 |
| empty query no network | 9 |
| error strings | 9 (`proxy down`, `not found`, `page layout changed`, …) |
| env hosts | 9 `FromEnv` |
| Search / Reader / Settings / Help | 13–16 |
| keys `s` `?` `t` `a` `[` `]` `q` | 12–14 |
| 3-color themes, no background | 11 |
| Density / autoscroll / diagrams | 10, 14, 15 |
| `selections.json` + corrupt fallback | 10 |
| fixtures, no live CI network | 2, 9 |
| 0.2 offline save | **out of plan** (spec says later) |
| Firefox / self-host freetar / transpose | **out of plan** |

No TBD left in tasks. Types (`SearchPage`, `TabDetail`, `Selections`) are named the same in every task.
