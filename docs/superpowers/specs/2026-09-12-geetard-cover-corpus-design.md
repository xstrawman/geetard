# GEETARD cover corpus — design

**Date:** 2026-09-12  
**Status:** Policy locked (A+C, else B+C). Mission: original **record**. No LLM this round.  
**Seed:** `testdata/cover-songs-seed.json` from `/home/mountaindewurbest/COVER SONGS.md` (850 ranks, no gaps)

## Goal

One library row per song. No seventy Wonderwalls. The 850-line cover list is the first pack: a resolver test, then a one-sheet test.

## Mission

Get as close to the **original record performance** as reasonably possible. A skilled player already knows what that means. The sheet is a transcription of how it was *played on the record*, not a concert-pitch rewrite of what the listener hears.

Canonical example: Collective Soul — *Run*.

- Record: guitars **½ step down**.
- Good chart: **C shapes** + a one-line tuning note (`½ step down` / Eb standard).
- Bad chart: written in **B** because that is the sounding pitch to a listener in concert pitch.

The bad chart is “accurate” to the ear and wrong for the player. We keep the shapes the guitarist used, and we **note** the tuning. Same rule for capo, drop D, dropped-C 90s tunings, Nashville, etc. Do not bake the transposition into the chord names.

This overrides the old “prefer standard tuning” tie-break. Standard is only a default when the record *is* standard.

## Policy: A+C, else B+C

**A** = original writer / original recording is the song’s identity.  
**C** = skilled transcription: record tuning and record shapes, not listener pitch, not “easy capo” fanfic.  
**B** = if the original cannot be resolved, or there is no decent original-record Chords sheet, fall back — still **one** row.

The 850 list’s third column is a **distractor** for identity. We do not key the library on Johnny Cash, Jimi Hendrix, or Jeff Buckley when the record we want is NIN / Dylan / Cohen. A famous cover chart is only in play under **B**, when the original-record sheet does not exist or is junk.

Together:

1. Resolve `original_artist` (MusicBrainz / SecondHandSongs). Not UG. Not the cover column.
2. **Library key** = `(original_artist_norm, title_norm)` when original is known. Else `(cover_artist_norm, title_norm)`.
3. **Sheet pool** = original artist + title first (A). Cover artist + title only if that pool has no decent Chords sheet (B).
4. Keep **one** Chords sheet. Drop Pro, Official, GP, uke, bass, tab.
5. **Judge (C):** among remaining candidates, prefer the chart that documents the record (tuning line, capo as played, original shapes). Bayesian stars×votes break ties. Then fewer wrong-key concert-pitch respells. Then lower version number. Never store two bodies.

`sheet_artist` is whose chart we saved. If A+C succeeded it should usually be the original act. Cover `sheet_artist` means we were in B.

## Fields (seed + library)

Seed JSON (`testdata/cover-songs-seed.json`):

- `rank`, `title`, `cover_artist` (from the md list)
- `original_artist` (null until resolved)

Library row (product DB):

- `original_artist` — writer / first recording; may be null only if resolve failed (then B)
- `cover_artist` — the list’s famous cover; metadata
- `sheet_artist` — whose chart `raw` actually is (cover or original)
- `title`, `path` of the winner, rating, votes, capo, tuning
- `raw` + shapes
- unique on the library key above

The TUI shows the original act. If `sheet_artist` differs (B fallback), footnote that. Tuning/capo from the record stay in the header; chord shapes stay as played.

## Golden judge tests (C)

- Collective Soul — *Run*: winner notes ½ step down (or Eb) and uses **C** shapes, not a concert-pitch **B** chart.
- A ½-down 90s chart that says nothing and writes sounding chords **loses** to one that keeps record shapes + a tuning line, even if the sounding chart has a slightly higher star rating and few votes.
- Capo-on-record stays capo + shapes; do not rewrite to open chords in a new key.

## Not this round

- LLM judge
- Seventy candidate bodies in the product DB
- Shipping `.html` / `.rtf` as the library (HTML is fetch input; store GEETARD `raw` text; `.txt` export later)
- Bulk zip of UG
- Copying `ultimateultimateguitar` (AGPL)

## Pipeline

```
COVER SONGS.md
  → cover-songs-seed.json
  → original_artist resolve
  → search original+title, then cover+title if needed
  → one Chords winner (record shapes + tuning note)
  → geetard.library (sqlite, then zstd)
```

Resume on disk. Git does not take the dump. The tarball may.

## Success on this 850

- 850 unique keys. No duplicate Wonderwalls.
- Rank 1–3: key is NIN / Dylan / Cohen. Sheet is the original-record Chords chart if one exists; Cash / Hendrix / Buckley only as B.
- Cover column never becomes the key when original resolved.
- *Run* / Collective Soul style: ½-down + original shapes beats concert-pitch respell.
- Low-confidence originals stay `needs_review`; they do not fork extra rows.
