#!/usr/bin/env bash
# Cross-compile absolute_cinema pour les 4 cibles et prépare le dossier dist/.
# Usage : build/build.sh [dossier destination]   (défaut : dist/)
set -euo pipefail
cd "$(dirname "$0")/.."

GO="${GO:-$HOME/go-sdk/bin/go}"
DEST="${1:-dist}"
mkdir -p "$DEST"

echo "→ windows/amd64 (sans console)"
GOOS=windows GOARCH=amd64 "$GO" build -ldflags "-s -w -H=windowsgui" -o "$DEST/AbsoluteCinema-Windows.exe" ./cmd/absolute_cinema
echo "→ darwin/arm64 (Mac Apple Silicon)"
GOOS=darwin GOARCH=arm64 "$GO" build -ldflags "-s -w" -o "$DEST/AbsoluteCinema-Mac-AppleSilicon" ./cmd/absolute_cinema
echo "→ darwin/amd64 (Mac Intel)"
GOOS=darwin GOARCH=amd64 "$GO" build -ldflags "-s -w" -o "$DEST/AbsoluteCinema-Mac-Intel" ./cmd/absolute_cinema
echo "→ linux/amd64"
GOOS=linux GOARCH=amd64 "$GO" build -ldflags "-s -w" -o "$DEST/absolute-cinema-linux" ./cmd/absolute_cinema

cp launchers/* "$DEST/"
echo
echo "Binaires et lanceurs prêts dans $DEST/ :"
ls -lh "$DEST"
