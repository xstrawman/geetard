# GEETARD so far — work summary

Terminal guitar tab reader in Go (`geetard`). Full-screen TUI, no browser, no UG site. Catalog is public freetar instances; content still originates as UG chord sheets. Repo: [github.com/xstrawman/geetard](https://github.com/xstrawman/geetard) (`feat/v1-tui`). Tests green.

Written 2026-09-13. Snapshot of corpus counts from that day; they will move as packing continues.

---

## What exists (product)

A playable TUI: search → open sheet → autoscroll, themes (amber/green/mono/nord), settings, help, newspaper columns on wide screens, library of opened tabs. Static linux/amd64 + arm64 builds and a ChromeOS Crostini tarball. Never requests `ultimate-guitar.com` from the app.

**Catalog survival:** `freetar.de` search is dead; packer/client tries `habedieeh.re` → `adast.dk` → `freetar.de`. HTML parse + `js-store` JSON. One song, one sheet (no 70 Wonderwalls).

---

## Library / corpus (the real work)

**Policy:** original **record** performance. Shapes as played; tuning noted (Collective Soul *Run* = C shapes + ½-step down, not concert-pitch B). Cover-list third column is a distractor. A+C else B+C. No LLM this round.

**Seeds**

| List | Rows | What it is |
|---|---|---|
| Cover songs (`COVER SONGS.md`) | 850 | Famous cover act in column 3; need original writer |
| Mood boosters (`top_800_mood_boosters.xlsx`) | 800 | Artist **is** the recording act — no Wikipedia |

**On disk (`corpus/tabs/`, 2026-09-13)**

| | |
|---|---|
| Sheets | **535** JSON files (~1.7 MB of `raw` text) |
| With lyrics + `[ch]` chords | **503** |
| Original artist filled | **170** |
| Cover-sheet fallback (B) | **365** |
| Cover-list ranks seen | 1–608 (71 fetch fails, packing still walking) |
| Examples that are real charts | NIN *Hurt*, Dylan *Watchtower*, Prine *Angel From Montgomery*, Cohen *Hallelujah* |

Windows `geetard-orig` does **not** download music. It only fills `original_artist`. Linux `geetard-pack` is what writes lyrics and chords.

**Windows tool:** [release orig-windows-1](https://github.com/xstrawman/geetard/releases/tag/orig-windows-1) — unzip, off VPN, `run.bat`. Mood-boosters JSON is in that zip too.

---

## What broke (and what didn’t)

| Problem | Reality |
|---|---|
| freetar.de / public proxy | Dead. Mirrors + HTML parse fixed search. |
| Wikipedia originals | API works. **429** from the Linux box on **Proton VPN**. 153× in the pack log. Home Windows IP is the fix, not a bigger CPU. |
| Title-only “best other artist” | Picked random third covers (DMB Watchtower). Killed. |
| Angel From Montgomery | **John Prine**. Bonnie Raitt was B after 429. |

---

## Is rented compute worthwhile?

**Not for this round**, if “compute” means GPU or a fat VM.

- Packing 850 + 800 sheets is **polite HTTP**, ~1–2 req/s, a few hours. CPU is idle. A GPU does nothing.
- Wikipedia is an **IP / rate-limit** problem. A datacenter VPS often 429s the same way. A **home Windows PC off VPN** (the exe you have) is the right worker.
- The 20k-song dream is still I/O and bans, not FLOPs. Overnight on one clean IP beats a $200 GPU box.

**When renting *would* pay:**

1. **LLM judge** (parked) — original vs concert-pitch vs ½-down, “is this the record.” That’s API tokens or a small rented GPU, and only after the 850+800 sheets exist.
2. A **non-VPN VPS with a clean IP** solely to fetch freetar 24/7 with backoff — cheap ($5–12/mo), not “AI compute.”
3. Later: embeddings / search over 20k sheets — modest CPU/RAM, still not a GPU.

**Cheapest path to a finished gig bag:** finish Windows originals → `geetard-pack` the 850 → pack mood-boosters (no wiki) → optional $6 VPS for a second fetch worker. Rent GPU only if you turn the judge back on.

---

## Still open

- Finish 850 pack; re-key B rows after Windows `originals.json`
- Pack the 800 mood boosters (straight `artist + title`)
- Strip junk originals (`soul singer-songwriter Otis Redding`)
- ChromeOS tarball is built, not a current mandate
- Transpose, sqlite+zstd ship format, 20k fill — later

**Bottom line:** the TUI is real; the library is starting to be real (535 playable sheets). The bottleneck is Wikipedia’s attitude toward VPNs and freetar politeness, not horsepower. Rent a quiet IP if you want speed; don’t rent a GPU until you want the model to pick charts.
