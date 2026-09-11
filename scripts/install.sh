#!/bin/sh
# Install geetard into ~/.local/bin and add a Linux app launcher entry.
# Safe on ChromeOS Crostini (Debian) and ordinary Linux.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname "$0")" && pwd)

arch=$(uname -m)
case "$arch" in
	x86_64|amd64) name=geetard-linux-amd64 ;;
	aarch64|arm64) name=geetard-linux-arm64 ;;
	*)
		echo "unsupported architecture: $arch" >&2
		echo "geetard ships linux/amd64 and linux/arm64 binaries." >&2
		exit 1
		;;
esac

bin=""
for d in "$ROOT" "$ROOT/dist" "$(pwd)" "$(pwd)/dist"; do
	if [ -f "$d/$name" ]; then
		bin="$d/$name"
		break
	fi
done
if [ -z "$bin" ] && [ -f "$ROOT/cmd/geetard/main.go" ]; then
	echo "no prebuilt $name; building for this machine…"
	( cd "$ROOT" && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "/tmp/$name" ./cmd/geetard )
	bin="/tmp/$name"
fi
if [ -z "$bin" ] || [ ! -f "$bin" ]; then
	echo "missing $name. Run scripts/build-release.sh or unpack geetard-chromeos.tar.gz." >&2
	exit 1
fi

dest="${HOME}/.local/bin"
mkdir -p "$dest"
install -m 755 "$bin" "$dest/geetard"

appdir="${HOME}/.local/share/applications"
mkdir -p "$appdir"
desktop_src="$ROOT/geetard.desktop"
if [ ! -f "$desktop_src" ]; then
	desktop_src="$ROOT/packaging/geetard.desktop"
fi
if [ -f "$desktop_src" ]; then
	sed "s|^Exec=.*|Exec=env TERM=xterm-256color COLORTERM=truecolor $dest/geetard|" \
		"$desktop_src" > "$appdir/geetard.desktop"
	chmod 644 "$appdir/geetard.desktop"
fi

if command -v update-desktop-database >/dev/null 2>&1; then
	update-desktop-database "$appdir" 2>/dev/null || true
fi

echo "installed $dest/geetard"
if command -v geetard >/dev/null 2>&1; then
	echo "run: geetard"
else
	echo "open a new terminal, or run: $dest/geetard"
	echo "(ChromeOS: ~/.local/bin is on PATH after a new Linux Terminal tab)"
fi
