# <img src="docs/img/logo.svg" height="30" alt=""> Absolute Cinema — the pocket movie library

**Your movies, on any screen, with a plain hard drive.**

Copy your movies to an external hard drive, drop Absolute Cinema next to them,
double-click: your collection becomes a beautiful library with posters, synopses and
subtitles — on any Windows, Mac or Linux computer. No account, no subscription, no
server to set up, and no internet once the first scan is done.

**➡ Website & demo: https://bist0uille.github.io/Absolute_cinema/en/**

*Version française : [README.md](README.md)*

## Features

- 🔍 **Automatic recognition** of movies and TV shows, even badly named files (over
  99% recognition measured on a real ~1000-video collection) — posters, synopses,
  dates, genres and ratings via TMDB, in English or French
- 📺 **TV shows**: seasons, episodes, episode summaries
- 🈶 **Languages**: audio version detection, automatic association of external
  subtitles (.srt/.ass/.sub)
- 🧸 **Kids mode**: PIN code + maximum age (official movie ratings), enforced
  server-side — not mere hiding
- 🧹 **Duplicates**: detection with a "best version" recommendation and a restorable trash
- ▶ **One-click playback** in VLC (with subtitle selection)
- 🔌 **Truly portable**: the library, cache and settings live on the drive in
  `.absolute_cinema/`. Switch computers, everything follows.

## Install

1. Download your platform's zip from the
   [Releases](https://github.com/Bist0uille/Absolute_cinema/releases) and unzip it
   **at the root of your movie hard drive**, next to your movie folders (any folder
   name containing "movies"/"films" or "series"/"shows" is detected).
2. Double-click (Windows: `AbsoluteCinema-Windows.exe`). Your browser opens.
3. On first launch, a welcome screen asks for your **free TMDB API key**
   ([sign up](https://www.themoviedb.org/signup), then copy the v3 key from
   [Settings → API](https://www.themoviedb.org/settings/api)) and starts the scan.
4. Install [VLC](https://www.videolan.org) for playback if you don't have it.

## Build from source

```bash
go build ./cmd/absolute_cinema        # local binary
bash build/build.sh                   # all 4 targets (Windows, Mac Intel/ARM, Linux)
go test ./...
```

Go ≥ 1.22, standard library only. The web UI (vanilla JS) is embedded in the binary
via `go:embed`. Development: `go run ./cmd/absolute_cinema -root "/path/to/drive"`.

## Legal notes

This software organizes video files **you already own**. It does not include,
provide, publicly index or download any movie.

Metadata and artwork by [The Movie Database (TMDB)](https://www.themoviedb.org).
This product uses the TMDB API but is not endorsed or certified by TMDB. Each user
brings their own personal API key.

Playback relies on [VLC](https://www.videolan.org), independent software not included.

## License

[GPL-3.0](LICENSE).
