#!/bin/bash
# Lance absolute_cinema depuis la racine du disque (Mac).
cd "$(dirname "$0")"
if [ "$(uname -m)" = "arm64" ]; then
  BIN="./AbsoluteCinema-Mac-AppleSilicon"
else
  BIN="./AbsoluteCinema-Mac-Intel"
fi
chmod +x "$BIN" 2>/dev/null
# Si le disque est en lecture seule (NTFS sur Mac), copie locale pour exécuter
if [ ! -x "$BIN" ]; then
  TMPBIN="$TMPDIR/absolute_cinema_bin"
  cp "$BIN" "$TMPBIN" && chmod +x "$TMPBIN"
  exec "$TMPBIN" -root "$(pwd)"
fi
exec "$BIN"
