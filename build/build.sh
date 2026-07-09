#!/usr/bin/env bash
# Cross-compile absolute_cinema et prépare le dossier dist/ avec une racine
# « propre » : seuls l'exécutable Windows et les lanceurs sont visibles ; les
# binaires Mac/Linux sont rangés dans .absolute_cinema/bin/.
# Usage : build/build.sh [dossier destination]   (défaut : dist/)
set -euo pipefail
cd "$(dirname "$0")/.."

GO="${GO:-$HOME/go-sdk/bin/go}"
DEST="${1:-dist}"
BIN="$DEST/.absolute_cinema/bin"
mkdir -p "$BIN"

echo "→ windows/amd64 (sans console)"
GOOS=windows GOARCH=amd64 "$GO" build -ldflags "-s -w -H=windowsgui" -o "$DEST/Absolute Cinema.exe" ./cmd/absolute_cinema
echo "→ darwin/arm64 (Mac Apple Silicon)"
GOOS=darwin GOARCH=arm64 "$GO" build -ldflags "-s -w" -o "$BIN/AbsoluteCinema-Mac-AppleSilicon" ./cmd/absolute_cinema
echo "→ darwin/amd64 (Mac Intel)"
GOOS=darwin GOARCH=amd64 "$GO" build -ldflags "-s -w" -o "$BIN/AbsoluteCinema-Mac-Intel" ./cmd/absolute_cinema
echo "→ linux/amd64"
GOOS=linux GOARCH=amd64 "$GO" build -ldflags "-s -w" -o "$BIN/absolute-cinema-linux" ./cmd/absolute_cinema

# Lanceurs visibles + LISEZMOI à la racine (l'exe Windows se double-clique directement)
cp launchers/*.command launchers/*.sh launchers/LISEZMOI.txt "$DEST/" 2>/dev/null || true

echo
echo "Prêt dans $DEST/ (racine propre) :"
ls -A "$DEST"
