#!/bin/bash
# Lance Absolute Cinema depuis la racine du disque (Mac).
cd "$(dirname "$0")"
ROOT="$(pwd)"
if [ "$(uname -m)" = "arm64" ]; then
  BIN=".absolute_cinema/bin/AbsoluteCinema-Mac-AppleSilicon"
else
  BIN=".absolute_cinema/bin/AbsoluteCinema-Mac-Intel"
fi
chmod +x "$BIN" 2>/dev/null
# Disque en lecture seule (NTFS sur Mac) : copie locale pour pouvoir exécuter
if [ ! -x "$BIN" ]; then
  TMPBIN="${TMPDIR:-/tmp}/absolute_cinema_bin"
  cp "$BIN" "$TMPBIN" && chmod +x "$TMPBIN"
  exec "$TMPBIN" -root "$ROOT"
fi
exec "$BIN" -root "$ROOT"
