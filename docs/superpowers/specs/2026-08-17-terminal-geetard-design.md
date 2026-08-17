# terminal GEETARD — design

**Date:** 2026-08-17  
**Status:** Written; waiting on user review  
**Name:** terminal GEETARD  
**Command:** `geetard`  
**Repo:** `~/Projects/apps/terminal-geetard`  
**Look/feel reference:** [ttune](https://github.com/SteveMCWin/ttune) (Bubble Tea v2 + Lipgloss v2). Steal the skeleton, keys, and settings delivery. Do not steal the tuner, ASCII guitar, or Portaudio.

## Goal

A full-screen terminal tab reader you can play guitar in front of. Search a song, open a chord sheet or tab, sit in an old-school color theme. No browser. No banner. No popup. Nothing that looks like a website.

v1 is useful the first night: search, open, read, autoscroll, settings that actually change the room.

## What this is not

- Not a Firefox redirector (that was the first idea; dropped).
- Not a hosted freetar clone. We do not run Flask or scrape `freetar.de` HTML.
- Not independent of Ultimate Guitar as a *catalog*. Content still originates on UG. We are independent of UG as a *place you open*: the program never requests `ultimate-guitar.com`.
- Not TABS on YOU (Android). That was the learning run. This is the TUI product.
- Not a dual Go+Python app. One contract, one implementation: **Go**.
- Not a clone of ttune. Same *delivery*. Different job.

## Approach

**Client contract + one TUI, in Go.** Fetch/parse is a small, testable core. The TUI is Bubble Tea v2 + Lipgloss v2 (same stack as ttune). Four states: search, reader, settings, help.

0.2 (later): save opened tabs to disk so a song you already opened does not need the network.

## Architecture

Two layers. No local daemon. No webview.

```
TUI (search | reader | settings | help)
        │
        ▼
Client (HTTP + js-store parse + tab markup + chord shapes)
        │
        ▼
proxy.freetar.de              tabs.proxy.freetar.de
  /search.php                   /tab/<artist>/<song>
```

The proxy returns UG pages. We throw away the HTML chrome and keep the JSON in `div.js-store[data-content]`. Same mechanism freetar and TABS on YOU use.

### Client

One job: talk to the proxy, return plain data.

| Call | Meaning |
|---|---|
| `search(query, page)` | Title search. Returns results + pagination. |
| `tab(path)` | Load one tab by UG path (`artist/song-slug-id`). |
| `render(raw)` | Turn `[ch]…[/ch]` / `[tab]…[/tab]` into display lines. |
| `shapes(tab)` | Unique chord names → one ASCII diagram each, from UG `applicature` when present. |

Rules:

- Never request `ultimate-guitar.com` or `*.ultimate-guitar.com`.
- Default hosts: `https://proxy.freetar.de` and `https://tabs.proxy.freetar.de`.
- Override with `FREETAR_SEARCH_HOST` and `FREETAR_TAB_HOST` (future self-hosted fetcher, not required for v1).
- Drop results whose `type` is `Pro` or `Official` (unreadable without a paid UG session). Same as freetar.
- User-Agent: a normal desktop browser string. No custom “bot” UA.
- Timeouts: 10s connect, 15s read. One retry on 5xx/timeout, then surface the error.
- UG ships some numbers as JSON numbers and some as strings. Parse rating, votes, version, and pagination with a flexible number reader (string or number → value; missing → 0).
- Empty query does not hit the network.

**Search request**

```
GET {SEARCH_HOST}/search.php?page={n}&search_type=title&value={query}
```

JSON path: `store.page.data.results[]`  
Pagination: `store.page.data.pagination.current`, `.total`

**Tab request**

```
GET {TAB_HOST}/tab/{path}
```

`path` is the UG path with the leading slash stripped, e.g. `willie-nelson/always-on-my-mind-123`.  
Taken from a search result’s `tab_url` path.

JSON path for body: `store.page.data.tab_view.wiki_tab.content`  
Metadata: `store.page.data.tab` (`artist_name`, `song_name`, `version`, `type`, `rating`) and `tab_view.meta` (`capo`, `tuning`) plus `tab_view.ug_difficulty`.  
Chord shapes: `store.page.data.tab_view.applicature` (same blob freetar already walks). If missing, diagrams setting shows nothing extra — the sheet still renders.

**Errors the TUI can show (plain strings, not stack traces)**

| Condition | Message |
|---|---|
| Timeout / network | `proxy timed out` |
| 5xx after retry | `proxy down ({code})` |
| 404 | `not found` |
| Missing `div.js-store` or empty `data-content` | `page layout changed` |
| JSON/shape mismatch | `could not parse page` |
| Search with no usable results | empty list, not an error |

### Data shapes

```text
SearchPage {
  query: string
  page: int
  total_pages: int
  results: [SearchResult]
}

SearchResult {
  artist: string
  song: string
  type: string          # Chords | Tab | Ukulele | …
  version: int
  rating: float         # 0–5, one decimal
  votes: int
  path: string          # "artist/song-id" for tab()
}

TabDetail {
  artist: string
  song: string
  version: int
  type: string
  rating: float
  capo: string | none
  tuning: string | none
  difficulty: string | none
  raw: string           # wiki_tab.content, still has [ch] markers
  shapes: [ChordShape]  # may be empty
}

ChordShape {
  name: string          # "G", "Em", "D/F#"
  lines: [string]       # 6 ASCII strings, high e on top
}

DisplayLine {
  parts: [ { text: string, kind: chord | body } ]
}
```

`render(raw)`:

- Strip `[tab]` / `[/tab]`.
- Split on newlines (`\r\n` or `\n`).
- Replace `[ch]…[/ch]` with a `chord` span. Keep the inner text as-is (no HTML).
- Everything else is `body`.
- Do not convert spaces to `&nbsp;` or newlines to `<br/>`. That is freetar’s web renderer. We are not a browser.

`shapes(tab)`:

- One diagram per unique chord name in the sheet.
- First fingering variant only (YAGNI on cycling shapes).
- ASCII, six strings, muted = `x`, open = `0`:

```
  e |-0-
  B |-1-
  G |-0-
  D |-2-
  A |-3-
  E |-x-
```

## TUI

Same skeleton as ttune, every screen:

1. **Title box** — `terminal GEETARD — search|read|settings|help`
2. **Content boxes** — Lipgloss borders, primary color
3. **Faint key footer** — always visible

Alt-screen. No HTML, no images, no toast that looks like an ad. The program does **not** paint the terminal background. Themes sit on Konsole’s background, like ttune.

### Search

```
┌─ terminal GEETARD — search ─────────────────────────┐
│ / nelson always                                     │
├─────────────────────────────────────────────────────┤
│ ARTIST            SONG                TYPE    ★     │
│ Willie Nelson     Always On My Mind   Chords  4.8   │
│ Willie Nelson     On The Road Again   Tab     4.6   │
└─────────────────────────────────────────────────────┘
 enter open   n/p page   / search   s settings   ? help   q quit
```

### Reader

```
┌─ terminal GEETARD — read ── Willie Nelson — Always On My Mind  v3  capo 1 ─┐
│ G            D            Em                                               │
│ Maybe I didn't love you                                                    │
└────────────────────────────────────────────────────────────────────────────┘
 j/k scroll   space page   a autoscroll   s settings   ? help   q back
```

When **chord diagrams** are `sidebar` and the window is wide enough, a second box sits on the right with the unique shapes. When `below`, they sit under the header, above the sheet. When `off`, the reader is only the sheet.

### Settings (ttune delivery)

`s` from search or reader. Left list of settings, right list of values, description under the values, **live preview** of the hovered value (ttune’s three-pane settings). `[ ]` / `[o]` selection. `hjkl` / arrows. Enter or space selects. Esc / backspace back to wherever you came from.

```
┌─ terminal GEETARD — settings ──────────────────────────────────────────────┐
│ Settings          Options              Preview                             │
│ [o] Theme         [o] amber            G        D        Em                │
│ [ ] Border        [ ] green            Maybe I didn't love you             │
│ [ ] Density       [ ] mono                                                 │
│ [ ] Autoscroll    [ ] nord                                                 │
│ [ ] Scroll speed                                                           │
│ [ ] Diagrams                                                               │
│                                                                            │
│ Description                                                                │
│ Color theme. Primary = borders, secondary = body, tertiary = chords.       │
└────────────────────────────────────────────────────────────────────────────┘
 esc back   j/k move   h/l pane   enter select   q quit
```

Changing a value applies **immediately** and writes `~/.config/geetard/selections.json`. No Apply button.

### Help

`?` opens a short help screen: keys, what the proxy is, what we never request. One page is enough. Esc back.

### Keys

| Key | Search | Reader | Settings | Help |
|---|---|---|---|---|
| `/` | Focus query | — | — | — |
| Enter / space | Open selected | — | Select value | — |
| `j` `k` / arrows | Move | Scroll one line | Move in pane | — |
| `h` `l` | — | — | Switch pane | — |
| `n` / `p` | Next / prev page | — | — | — |
| Space | — | Page down | Select value | — |
| `a` | — | Toggle autoscroll | — | — |
| `[` / `]` | — | Scroll slower / faster | — | — |
| `s` | Settings | Settings | — | Settings |
| `?` | Help | Help | Help | — |
| `t` | Next theme | Next theme | — | — |
| Esc / backspace | — | — | Back | Back |
| `q` | Quit | Back to search | Quit | Quit |

While the query field is focused, letters type into the query. `Enter` in the query runs search. `Esc` or `Down` leaves the query and focuses the list.

### Settings catalog (v1)

Few knobs. Each is a list of named values, like ttune. No free-text except we do not need any in v1.

| Setting | Values | Default | What it actually does |
|---|---|---|---|
| **Theme** | `amber`, `green`, `mono`, `nord` | `amber` | Three hex roles (below). Preview = fake chord line. |
| **Border** | `ascii`, `normal`, `rounded`, `none` | `ascii` | Lipgloss border on the boxes. You already run ttune on Ascii. |
| **Density** | `compact`, `normal`, `airy` | `normal` | Line gap 0 / 1 / 2. **This is the TUI stand-in for font size.** A program cannot change Konsole’s typeface; `Ctrl`+`+` still does that. Density is what we own. |
| **Autoscroll** | `off`, `on` | `off` | Default when a reader opens. `a` still toggles for this song. |
| **Scroll speed** | `slow` (800ms), `medium` (400ms), `fast` (200ms) | `medium` | Line interval. `[` `]` still nudge ±50ms this session (clamp 100–2000ms) without rewriting the named default unless you set it here. |
| **Diagrams** | `off`, `sidebar`, `below` | `off` | ASCII chord shapes from UG `applicature`. Preview = a G shape. |

No theme editor. No custom hex in v1. No YIN-style pile of numeric internals.

### Autoscroll (reader)

- Starts from the **Autoscroll** setting when a tab opens.
- Footer shows `autoscroll * 400ms` when on. No bouncing animation.
- `a` and `[` `]` are session overrides. Opening settings and picking a speed writes the named default.

### Themes

ttune model: three hex colors. **Do not set the terminal background.**

| Role | Use |
|---|---|
| primary | box borders, title, rules |
| secondary | body / tab ASCII / lyrics |
| tertiary | chords, selection `[o]`, footer hints |

| Name | primary | secondary | tertiary | Feel |
|---|---|---|---|---|
| `amber` | `#E09A3E` | `#E6C07B` | `#F0C14B` | CRT |
| `green` | `#3D8B4A` | `#8FBC8F` | `#7CFC00` | phosphor |
| `mono` | `#888888` | `#CCCCCC` | `#FFFFFF` | white/black |
| `nord` | `#88C0D0` | `#D8DEE9` | `#EBCB8B` | same trio you use in ttune |

`t` cycles the four. Settings → Theme does the same with a preview.

### Config

XDG: `$XDG_CONFIG_HOME/geetard/` or `~/.config/geetard/`.

`selections.json` (ttune-shaped, only what we chose):

```json
{
  "theme": "amber",
  "border": "ascii",
  "density": "normal",
  "autoscroll": "off",
  "scroll_speed": "medium",
  "diagrams": "off"
}
```

Missing file → defaults. Corrupt file → defaults + footer `bad config, using defaults`.

Env (not in the settings screen):

| Source | Purpose |
|---|---|
| `FREETAR_SEARCH_HOST` | Override search base (no trailing slash) |
| `FREETAR_TAB_HOST` | Override tab base |

No account. No telemetry. No update ping.

## Layout

```
terminal-geetard/
  docs/superpowers/specs/     # this design
  client/                     # HTTP, js-store parse, render(), shapes()
  tui/                        # search, reader, settings, help
  testdata/                   # captured js-store JSON/HTML fixtures
```

Go module. Binary `geetard`. No args → search screen.

## Testing

- Checked-in fixtures under `testdata/` (search page HTML, tab page HTML, plus the extracted JSON, including one page that has `applicature` and one that does not).
- Unit tests for: extract `js-store`, map search results (including dropping Pro/Official), map tab metadata, `render()` chord spans, `shapes()` ASCII, URL path extraction.
- Client tests that the constructed URLs never contain `ultimate-guitar.com`.
- Settings: load defaults, load a valid `selections.json`, reject a corrupt file.
- No live network in automated tests.

A manual check after v1: search a known song, open it, cycle themes, open settings, turn diagrams to `sidebar`, turn autoscroll on, play.

## Error handling (UI)

- Failed search: footer shows the error string; list stays or clears to empty. Query remains.
- Failed tab open: stay on search, footer shows the error. Do not open an empty reader.
- Mid-read network is not a thing in v1 (tab is fetched once). 0.2 can reopen from disk.

## 0.2 (explicitly later)

After a successful `tab()`, write `{artist, song, raw, shapes, fetched_at}` under `~/.local/share/geetard/tabs/`. Reopen without the proxy. Search still needs the network. Not in v1.

## Out of scope (v1)

- Firefox extension / UG URL redirect
- Hosting freetar or a UG fetch proxy
- Offline library (0.2)
- Transpose
- Changing the terminal emulator’s typeface (not possible; Density is the in-app stand-in)
- Cycling alternate fingerings for one chord
- Theme editor / user hex files
- Favorites UI
- Printing
- Hitting `ultimate-guitar.com` directly
- A second language implementation
- ASCII guitar art, pitch detection, Portaudio

## Success

You type `geetard`, search, open a sheet, hit `s`, pick amber and sidebar diagrams, hit `a`, and play. It feels like ttune’s cousin, not like a website. You never see UG. You never see an ad.
