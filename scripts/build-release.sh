#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$ROOT"
mkdir -p dist
ldflags='-s -w'

echo "building linux/amd64"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="$ldflags" -o dist/geetard-linux-amd64 ./cmd/geetard
echo "building linux/arm64"
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="$ldflags" -o dist/geetard-linux-arm64 ./cmd/geetard

build_windows() {
	arch=$1
	echo "building windows/$arch"
	CGO_ENABLED=0 GOOS=windows GOARCH="$arch" go build -trimpath -ldflags="$ldflags" -o "dist/geetard-windows-${arch}.exe" ./cmd/geetard
	wstage=$(mktemp -d)
	cp "dist/geetard-windows-${arch}.exe" "$wstage/geetard.exe"
	rm -f "dist/geetard-windows-${arch}.zip"
	( cd "$wstage" && zip -q -X "geetard-windows-${arch}.zip" geetard.exe )
	mv "$wstage/geetard-windows-${arch}.zip" dist/
	rm -rf "$wstage"
	rm -f "dist/geetard-windows-${arch}.exe"
}
build_windows amd64
build_windows arm64

stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
mkdir -p "$stage/geetard"
cp dist/geetard-linux-amd64 dist/geetard-linux-arm64 "$stage/geetard/"
cp scripts/install.sh "$stage/geetard/install.sh"
cp packaging/geetard.desktop "$stage/geetard/geetard.desktop"
cp README.md "$stage/geetard/README.md"
chmod +x "$stage/geetard/install.sh" "$stage/geetard/geetard-linux-amd64" "$stage/geetard/geetard-linux-arm64"

tar -C "$stage" -czf dist/geetard-chromeos.tar.gz geetard
(
	cd dist
	sha256sum \
		geetard-linux-amd64 \
		geetard-linux-arm64 \
		geetard-chromeos.tar.gz \
		geetard-windows-amd64.zip \
		geetard-windows-arm64.zip \
		> SHA256SUMS
)
echo "wrote dist/geetard-linux-amd64 dist/geetard-linux-arm64 dist/geetard-chromeos.tar.gz dist/geetard-windows-amd64.zip dist/geetard-windows-arm64.zip dist/SHA256SUMS"
