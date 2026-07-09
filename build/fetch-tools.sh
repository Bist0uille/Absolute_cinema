#!/usr/bin/env bash
# Télécharge VLC portable (Windows + Mac) dans build/disk-template/.absolute_cinema/tools/.
# À lancer UNE fois avant build/prepare-disk.sh. ~250 Mo, mis en cache.
# Ces binaires ne sont PAS embarqués dans le programme ni dans les zips de la
# release GitHub : ils ne servent qu'à préparer les disques « prêts à l'emploi ».
set -euo pipefail
cd "$(dirname "$0")/.."

VLC_VER="${VLC_VER:-3.0.21}"
TOOLS="build/disk-template/.absolute_cinema/tools"
CACHE="build/.tools-cache"
mkdir -p "$TOOLS" "$CACHE"

# --- Windows : archive portable (vlc.exe fonctionne tel quel) ---
WIN_ZIP="$CACHE/vlc-$VLC_VER-win64.zip"
if [ ! -f "$WIN_ZIP" ]; then
  echo "→ Téléchargement VLC Windows $VLC_VER"
  curl -fL -o "$WIN_ZIP" "https://download.videolan.org/vlc/$VLC_VER/win64/vlc-$VLC_VER-win64.zip"
fi
if [ ! -f "$TOOLS/vlc-win/vlc.exe" ]; then
  echo "→ Extraction VLC Windows"
  rm -rf "$TOOLS/vlc-win"; mkdir -p "$TOOLS/vlc-win"
  tmp="$(mktemp -d)"; unzip -q "$WIN_ZIP" -d "$tmp"
  mv "$tmp"/vlc-*/* "$TOOLS/vlc-win/"   # le zip contient un dossier vlc-<ver>/ ; on l'aplatit
  rm -rf "$tmp"
fi

# --- Mac : VLC.app extrait du dmg (signé & notarisé par VideoLAN) ---
MAC_DMG="$CACHE/vlc-$VLC_VER-universal.dmg"
if [ ! -f "$MAC_DMG" ]; then
  echo "→ Téléchargement VLC Mac $VLC_VER"
  curl -fL -o "$MAC_DMG" "https://download.videolan.org/vlc/$VLC_VER/macosx/vlc-$VLC_VER-universal.dmg"
fi
if [ ! -d "$TOOLS/vlc-mac/VLC.app" ]; then
  echo "→ Extraction VLC.app"
  mkdir -p "$TOOLS/vlc-mac"
  if command -v hdiutil >/dev/null 2>&1; then            # macOS
    mnt="$(mktemp -d)"
    hdiutil attach -nobrowse -mountpoint "$mnt" "$MAC_DMG"
    cp -R "$mnt/VLC.app" "$TOOLS/vlc-mac/"
    hdiutil detach "$mnt" >/dev/null
  elif command -v 7z >/dev/null 2>&1; then               # Linux/WSL (p7zip-full)
    tmp="$(mktemp -d)"; 7z x -o"$tmp" "$MAC_DMG" >/dev/null || true
    app="$(find "$tmp" -maxdepth 4 -name VLC.app -type d | head -1)"
    [ -n "$app" ] && cp -R "$app" "$TOOLS/vlc-mac/" || echo "⚠ VLC.app introuvable dans le dmg." >&2
    rm -rf "$tmp"
  else
    echo "⚠ Ni hdiutil ni 7z : impossible d'extraire VLC.app." >&2
    echo "  Installez p7zip-full (Linux) ou lancez ce script sur un Mac." >&2
  fi
fi

# --- Licence GPL + offre de source (obligatoire pour redistribuer VLC) ---
curl -fsL -o "$TOOLS/VLC-LICENSE.txt" "https://raw.githubusercontent.com/videolan/vlc/master/COPYING" || \
  echo "⚠ Copiez manuellement la licence GPL de VLC dans $TOOLS/VLC-LICENSE.txt" >&2
cat > "$TOOLS/SOURCES.txt" <<EOF
VLC media player $VLC_VER — VideoLAN, licence GNU GPL v2 ou ultérieure.
Binaires portables redistribués tels quels depuis https://www.videolan.org.
Code source : https://code.videolan.org/videolan/vlc (branche/tag $VLC_VER)
Archives officielles : https://download.videolan.org/vlc/$VLC_VER/
EOF

echo "✓ VLC prêt dans $TOOLS"
