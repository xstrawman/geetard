# terminal GEETARD — design

**Date:** 2026-08-17  
**Status:** Written; waiting on user review  
**Name:** terminal GEETARD  
**Command:** `geetard`  
**Repo:** `~/Projects/apps/terminal-geetard`

## Goal

A full-screen terminal tab reader you can play guitar in front of. Search a song, open a chord sheet or tab, sit in an old-school color theme. No browser. No banner. No popup. Nothing that looks like a website.

v1 is useful the first night: search, open, read, autoscroll, change theme.

## What this is not

- Not a Firefox redirector (that was the first idea; dropped).
- Not a hosted freetar clone. We do not run Flask or scrape `freetar.de` HTML.
- Not independent of Ultimate Guitar as a *catalog*. Content still originates on UG. We are independent of UG as a *place you open*: the program never requests `ultimate-guitar.com`.
- Not TABS on YOU (Android). That was the learning run. This is the TUI product.
- Not a dual Go+Python app. One contract, one implementation.

## Approach

**Client contract + one TUI.** Fetch/parse is a small, testable core. The TUI is two screens. Implementation language is chosen in the plan (Go + Bubble Tea or Python + Textual). v1 ships in exactly one of those.

0.2 (later): save opened tabs to disk so a song you already opened does not need the network.

## Architecture

Two layers. No local daemon. No webview.

```
TUI (search | reader)
        │
        ▼
Client (HTTP + js-store parse + tab markup)
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

### TUI

One job: keys and screens. Completely terminal. No HTML, no images, no “toast” that looks like an ad.

**Search**

```
 terminal GEETARD                      theme: amber
 ─────────────────────────────────────────────────
 / nelson always
 ─────────────────────────────────────────────────
 ARTIST            SONG                TYPE    ★
 Willie Nelson     Always On My Mind   Chords  4.8
 Willie Nelson     On The Road Again   Tab     4.6
 ─────────────────────────────────────────────────
 enter open   n/p page   / search   t theme   q quit
```

**Reader**

```
 Willie Nelson — Always On My Mind     v3  capo 1
 ─────────────────────────────────────────────────
 G            D            Em
 Maybe I didn't love you
 ─────────────────────────────────────────────────
 j/k scroll   space page   a autoscroll   [ ] speed   t theme   q back
```

**Keys**

| Key | Search | Reader |
|---|---|---|
| `/` | Focus query | — |
| Enter | Open selected | — |
| `j` / `k` / arrows | Move selection | Scroll one line |
| `n` / `p` | Next / previous search page | — |
| Space | — | Page down |
| `a` | — | Toggle autoscroll |
| `[` / `]` | — | Autoscroll slower / faster |
| `t` | Next theme | Next theme |
| `q` | Quit | Back to search |

While the query field is focused, letters type into the query. `Enter` in the query runs search. `Esc` or `Down` leaves the query and focuses the list.

Font size is the terminal emulator’s job (Konsole `Ctrl`+`+`). The app does not fake zoom.

**Autoscroll**

- Off by default.
- When on, the reader advances one line every N milliseconds.
- Default interval 400ms. `[` / `]` step by 50ms. Clamp 100–2000ms.
- No bouncing “now playing” animation. A `*` in the footer is enough (`autoscroll * 400ms`).

### Themes

Cycle with `t`. Four built-in 16-color palettes. No theme editor in v1.

| Name | Feel |
|---|---|
| `amber` | CRT terminal (default) |
| `green` | phosphor |
| `mono` | white on black |
| `nord` | dim cool |

Each palette sets: background, body text, chord accent, header, footer, selection, search slash. Chords use the accent. Body lines use the body color. Nothing blinks except we do not blink at all.

Last theme (and last autoscroll interval) persist in:

```
~/.config/geetard/config.toml
```

```toml
theme = "amber"
autoscroll_ms = 400
```

XDG: if `XDG_CONFIG_HOME` is set, use `$XDG_CONFIG_HOME/geetard/config.toml`.

### Config / env

| Source | Purpose |
|---|---|
| `FREETAR_SEARCH_HOST` | Override search base (no trailing slash) |
| `FREETAR_TAB_HOST` | Override tab base |
| `~/.config/geetard/config.toml` | Theme + autoscroll speed |

No account. No telemetry. No update ping.

## Layout (once language is picked)

Language-agnostic shape. Names map onto Go packages or Python modules the same way.

```
terminal-geetard/
  docs/superpowers/specs/     # this design
  client/                     # HTTP, js-store parse, render()
  tui/                        # search + reader
  testdata/                   # captured js-store JSON/HTML fixtures
```

The binary is `geetard`. Running it with no args opens the TUI on the search screen.

## Testing

- Checked-in fixtures under `testdata/` (search page HTML, tab page HTML, plus the extracted JSON).
- Unit tests for: extract `js-store`, map search results (including dropping Pro/Official), map tab metadata, `render()` chord spans, URL path extraction.
- Client tests that the constructed URLs never contain `ultimate-guitar.com`.
- No live network in automated tests.

A manual check after v1: search a known song, open it, confirm chords color with the accent, cycle all four themes, toggle autoscroll.

## Error handling (UI)

- Failed search: footer shows the error string; list stays or clears to empty. Query remains.
- Failed tab open: stay on search, footer shows the error. Do not open an empty reader.
- Mid-read network is not a thing in v1 (tab is fetched once). 0.2 can reopen from disk.

## 0.2 (explicitly later)

After a successful `tab()`, write `{artist, song, raw, fetched_at}` under `~/.local/share/geetard/tabs/`. Reopen without the proxy. Search still needs the network. Not in v1.

## Out of scope (v1)

- Firefox extension / UG URL redirect
- Hosting freetar or a UG fetch proxy
- Offline library (0.2)
- Transpose
- ASCII chord diagrams / fretboard
- Theme editor, 24-bit custom colors
- Favorites UI
- Printing
- Hitting `ultimate-guitar.com` directly
- Shipping both Go and Python

## Success

You type `geetard`, search, open a sheet, switch to amber or green, hit `a`, and play. You never see UG. You never see an ad.
