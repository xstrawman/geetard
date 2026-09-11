# terminal GEETARD

TUI tab reader. Search a song, open a chord sheet, sit in an old-school theme.

Never talks to `ultimate-guitar.com`. Catalog comes from public freetar instances
(`freetar.habedieeh.re`, `tabs.adast.dk`, then `freetar.de` as a last try).

```bash
go test ./...
go run ./cmd/geetard
```

Install on this machine:

```bash
./scripts/build-release.sh
./scripts/install.sh
geetard
```

Keys: `/` search, enter open, `l` library, `s` settings, `a` autoscroll, `t` theme, `q` back/quit.

See `docs/superpowers/specs/2026-08-17-terminal-geetard-design.md`.

## ChromeOS (Chromebook)

This is a terminal app. It runs in **Linux (Crostini)**, not in the Chrome browser
and not in crosh (`Ctrl+Alt+T`).

1. On the Chromebook: **Settings → Advanced → Developers → Linux development environment → Turn on**.
   Wait for the penguin container to finish installing, then open **Terminal**.
2. Copy `dist/geetard-chromeos.tar.gz` onto the Chromebook (USB, Drive, or `scp`)
   and move it into Linux files.
3. In the Linux Terminal:

```bash
mkdir -p ~/apps && tar -C ~/apps -xf geetard-chromeos.tar.gz
cd ~/apps/geetard
./install.sh
geetard
```

If `geetard` is not found, open a **new** Terminal tab (so `~/.local/bin` is on `PATH`)
or run `~/.local/bin/geetard`.

Intel/AMD Chromebooks use the amd64 binary; ARM Chromebooks use arm64.
`install.sh` picks the right one.

Override catalog hosts if you self-host freetar:

```bash
export FREETAR_SEARCH_HOST=https://freetar.example
export FREETAR_TAB_HOST=https://freetar.example
geetard
```
