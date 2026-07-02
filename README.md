# 🎬 Absolute Cinema — la vidéothèque de poche

**Vos films, sur n'importe quel écran, avec un simple disque dur.**

Copiez vos films sur un disque dur externe, posez Absolute Cinema à côté, double-cliquez :
votre collection devient une belle bibliothèque avec affiches, synopsis et sous-titres —
sur n'importe quel PC, Mac ou Linux. Sans compte, sans abonnement, sans serveur à
configurer, et sans internet une fois le premier scan terminé.

**➡ Site & démo : https://bist0uille.github.io/Absolute_cinema/**

## Fonctionnalités

- 🔍 **Reconnaissance automatique** des films et séries, même mal nommés (plus de 99 % de
  reconnaissance mesurée sur une vraie collection de ~1000 vidéos) — affiches, synopsis
  en français, dates, genres, notes via TMDB
- 📺 **Séries** : saisons, épisodes, résumés d'épisodes en français
- 🈶 **Langues** : détection VF / VO / VOSTFR / MULTI, association automatique des
  sous-titres externes (.srt/.ass/.sub)
- 🧸 **Mode enfant** : code PIN + âge maximum (classification officielle des films),
  filtrage appliqué côté serveur — pas un simple masquage
- 🧹 **Doublons** : détection des films en double avec recommandation de la meilleure
  version, corbeille restaurable
- ▶ **Lecture en un clic** dans VLC (avec choix des sous-titres)
- 🔌 **Vraiment portable** : la bibliothèque, le cache et les réglages vivent sur le
  disque, dans `.absolute_cinema/`. Changez d'ordinateur, tout suit.

## Installation

1. Téléchargez le binaire de votre système dans les
   [Releases](https://github.com/Bist0uille/Absolute_cinema/releases), et posez-le
   **à la racine de votre disque dur de films**, à côté de vos dossiers de films/séries
   (nommés par ex. `[Films]` et `[Séries]`, ou tout nom contenant « films »/« séries »).
2. Double-cliquez (Windows : `AbsoluteCinema-Windows.exe`). Le navigateur s'ouvre.
3. Au premier lancement, collez votre **clé API TMDB** dans Réglages (compte gratuit sur
   [themoviedb.org](https://www.themoviedb.org/signup), clé v3 dans
   [Paramètres → API](https://www.themoviedb.org/settings/api)) et lancez le scan.
4. Installez [VLC](https://www.videolan.org) pour la lecture, si ce n'est pas déjà fait.

Sur Mac et Linux, utilisez les scripts `Lancer - Mac.command` / `lancer-linux.sh` fournis
dans la release (ils gèrent le bit exécutable perdu sur les disques NTFS).

## Compiler depuis les sources

```bash
go build ./cmd/absolute_cinema        # binaire local
bash build/build.sh                   # les 4 cibles (Windows, Mac Intel/ARM, Linux)
go test ./...                         # tests (le parseur de noms est le cœur testé)
```

Go ≥ 1.22, aucune dépendance externe (stdlib uniquement). L'interface web (vanilla JS)
est embarquée dans le binaire via `go:embed`.

En développement : `go run ./cmd/absolute_cinema -root "/chemin/vers/le/disque"`.

## Architecture (en bref)

```
cmd/absolute_cinema/   point d'entrée — le binaire se repère via son propre emplacement
internal/parse/        extraction titre/année/langue/épisode depuis les noms de fichiers
internal/scanner/      parcours du disque, filtrage, association des sous-titres
internal/tmdb/         client TMDB : throttle, retries, cache disque intégral (offline)
internal/catalog/      orchestration scan → matching → bibliothèque
internal/library/      modèle de données, persistance JSON, détection de doublons
internal/player/       découverte de VLC par OS, lancement de la lecture
internal/server/       API HTTP + interface web embarquée + mode enfant
```

Tous les chemins stockés sont **relatifs à la racine du disque** : la bibliothèque
fonctionne quelle que soit la lettre de lecteur ou le point de montage.

## Notes légales

Ce logiciel organise des fichiers vidéo **que vous possédez déjà**. Il n'inclut, ne
fournit, n'indexe publiquement et ne télécharge aucun film.

Métadonnées et affiches fournies par [The Movie Database (TMDB)](https://www.themoviedb.org).
Ce produit utilise l'API TMDB sans être approuvé ni certifié par TMDB. Chaque utilisateur
emploie sa propre clé API personnelle.

La lecture s'appuie sur [VLC](https://www.videolan.org), logiciel indépendant non inclus.

## Licence

[GPL-3.0](LICENSE) — libre d'utiliser, d'étudier et de modifier ; toute redistribution
modifiée doit rester sous la même licence.
