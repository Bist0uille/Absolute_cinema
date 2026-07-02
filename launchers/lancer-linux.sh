#!/bin/bash
# Lance absolute_cinema depuis la racine du disque (Linux).
cd "$(dirname "$0")"
BIN="./absolute-cinema-linux"
chmod +x "$BIN" 2>/dev/null
if [ ! -x "$BIN" ]; then
  TMPBIN="${TMPDIR:-/tmp}/absolute_cinema_bin"
  cp "$BIN" "$TMPBIN" && chmod +x "$TMPBIN"
  exec "$TMPBIN" -root "$(pwd)"
fi
exec "$BIN"
