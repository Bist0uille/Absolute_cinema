#!/usr/bin/env bash
# Assemble un disque « prêt à l'emploi » : programme + lanceurs + VLC embarqué.
# À lancer après build/build.sh et build/fetch-tools.sh.
# Usage : build/prepare-disk.sh <racine-du-disque>
#   ex.  build/prepare-disk.sh /Volumes/AbsoluteCinema
#        build/prepare-disk.sh /mnt/d
#
# Recommandation : formater le disque en exFAT (lecture/écriture Windows + Mac,
# VLC embarqué exécutable des deux côtés, cache et corbeille actifs).
set -euo pipefail
cd "$(dirname "$0")/.."

DEST="${1:?Usage: build/prepare-disk.sh <racine-du-disque>}"
DIST="${DIST:-dist}"
TEMPLATE="build/disk-template"

[ -f "$DIST/Absolute Cinema.exe" ] || { echo "Lancez d'abord build/build.sh" >&2; exit 1; }
[ -d "$DEST" ] || { echo "Introuvable : $DEST" >&2; exit 1; }

echo "→ Programme et lanceurs vers $DEST"
mkdir -p "$DEST/.absolute_cinema/bin"
cp "$DIST/Absolute Cinema.exe" "$DEST/"
cp "$DIST/Lancer sur Mac.command" "$DEST/" 2>/dev/null || true
cp "$DIST/lancer-linux.sh" "$DEST/" 2>/dev/null || true
cp "$DIST/LISEZMOI.txt" "$DEST/" 2>/dev/null || true
cp "$DIST/.absolute_cinema/bin/"* "$DEST/.absolute_cinema/bin/"

if [ -d "$TEMPLATE/.absolute_cinema/tools" ]; then
  echo "→ VLC embarqué"
  mkdir -p "$DEST/.absolute_cinema/tools"
  cp -R "$TEMPLATE/.absolute_cinema/tools/." "$DEST/.absolute_cinema/tools/"
else
  echo "⚠ VLC non préparé (lancez build/fetch-tools.sh) — disque sans VLC embarqué." >&2
fi

echo "✓ Disque prêt : $DEST"
echo
echo "Pour un catalogue pré-configuré (affiches & résumés sans clé côté client) :"
echo "  1. Lancez l'application une fois sur ce disque AVEC votre clé TMDB."
echo "  2. Avant livraison, supprimez  $DEST/.absolute_cinema/config.json"
echo "     (il contient VOTRE clé — à ne jamais distribuer). Le library.json"
echo "     et les affiches en cache, eux, restent et s'affichent sans clé."
