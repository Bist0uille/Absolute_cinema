# Changelog

## v1.3.2-beta — 2026-07-03

- New brand identity: golden cinema-ticket logo (app top bar, favicon, website,
  demo video) and proper capitalization of the name — Absolute Cinema

## v1.3.1-beta — 2026-07-03

- Safety guard: the automatic startup scan no longer degrades a good library
  when the network is unavailable and the TMDB cache is incomplete
- Multi-part movies (CD1/CD2) now open in VLC (which chains the parts) instead
  of playing only part one in the browser
- Smoother search (debounced rendering); hover play button hidden on touch screens
- New demo video showcasing the built-in player, resume and the new home

## v1.3.0-beta — 2026-07-03

- **Built-in player**: movies and episodes now play inside the app (h264 mp4/mkv,
  webm), with subtitles (on-the-fly SRT→VTT conversion, encoding repair), and a
  one-click "Play in VLC" fallback for exotic formats (avi, x265, ISO)
- **Resume playback**: positions are saved on the drive itself — pause on one
  computer, resume on another; episodes are marked watched at 92%
- **Netflix-style home**: hero banner, "Resume" row with progress bars,
  "Recently added" row
- **One-click play**: ▶ on poster hover, big Play/Resume button on detail pages,
  TV shows start/resume the next unwatched episode automatically
- **Auto-scan on every launch**: new files appear by themselves
- **TV & tablets (local network)**: opt-in setting exposing the library to your
  home Wi-Fi devices, with the TV address displayed

## v1.2.0-beta — 2026-07-03

- **English support**: fully bilingual interface (auto-detected, switchable in
  Settings) and metadata language choice (posters/synopses in English or French)
- **Welcome wizard** on first launch: pick languages, paste the TMDB key, scan —
  guided in two steps
- **Windows**: the app now starts without a console window
- **Update check**: a banner in Settings signals new releases (silent when offline)
- **Trash**: new "Empty trash" button (double confirmation)
- English folder names (`[Movies]`, `[Shows]`…) detected at the drive root
- New: `README.en.md`, English landing page (`/en/`)

## v1.1.0-beta — 2026-07-02

- First public beta
- Automatic movie/TV recognition via TMDB (99.8% on a real 994-video collection)
- French metadata, posters, backdrops, episode summaries
- Language badges (VF/VO/VOSTFR/MULTI), external subtitle association
- Kids mode with PIN lock and official age ratings, enforced server-side
- Duplicate detection with recommendation + restorable trash
- One-click VLC playback with subtitle choice
- Fully portable: single binary per OS, data cached on the drive, works offline
