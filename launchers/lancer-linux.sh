#!/bin/bash
# Lance Absolute Cinema depuis la racine du disque (Linux).
cd "$(dirname "$0")"
ROOT="$(pwd)"
BIN=".absolute_cinema/bin/absolute-cinema-linux"
chmod +x "$BIN" 2>/dev/null
if [ ! -x "$BIN" ]; then
  TMPBIN="${TMPDIR:-/tmp}/absolute_cinema_bin"
  cp "$BIN" "$TMPBIN" && chmod +x "$TMPBIN"
  exec "$TMPBIN" -root "$ROOT"
fi
exec "$BIN" -root "$ROOT"
