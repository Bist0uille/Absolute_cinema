// Package catalog orchestre le scan disque + le matching TMDB pour produire
// la bibliothèque finale.
package catalog

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"absolute_cinema/internal/library"
	"absolute_cinema/internal/parse"
	"absolute_cinema/internal/scanner"
	"absolute_cinema/internal/tmdb"
)

// Override est une correction manuelle : chemin (fichier ou dossier) → fiche TMDB.
type Override struct {
	MediaType string `json:"media_type"` // "movie" | "tv"
	TmdbID    int    `json:"tmdb_id"`
}

// Progress est appelé pendant la construction (phase, fait, total, élément courant).
type Progress func(phase string, done, total int, current string)

const confidenceThreshold = 0.5

// Build scanne mediaRoot et construit la bibliothèque complète.
func Build(mediaRoot, dataDir string, client *tmdb.Client, overrides map[string]Override, progress Progress) (*library.Library, error) {
	if progress == nil {
		progress = func(string, int, int, string) {}
	}

	filmsDirs, seriesDirs, err := scanner.DetectRoots(mediaRoot)
	if err != nil {
		return nil, err
	}
	if len(filmsDirs) == 0 && len(seriesDirs) == 0 {
		return nil, fmt.Errorf("aucun dossier de films ou de séries trouvé dans %s", mediaRoot)
	}

	progress("walk", 0, 0, "parcours du disque…")
	res, err := scanner.Scan(mediaRoot, filmsDirs, seriesDirs)
	if err != nil {
		return nil, err
	}

	lib := &library.Library{SchemaVersion: 1, ScannedAt: time.Now().UTC()}

	buildMovies(lib, res.Movies, client, dataDir, overrides, progress)
	buildSeries(lib, res.Episodes, client, dataDir, overrides, progress)
	attachOrphanSubs(lib, res.OrphanSubs)

	sort.Slice(lib.Movies, func(i, j int) bool { return lib.Movies[i].Title < lib.Movies[j].Title })
	sort.Slice(lib.Series, func(i, j int) bool { return lib.Series[i].Title < lib.Series[j].Title })
	return lib, nil
}

// --- Films ---

func buildMovies(lib *library.Library, files []*scanner.FileEntry, client *tmdb.Client, dataDir string, overrides map[string]Override, progress Progress) {
	if !client.HasKey() {
		buildMoviesLocal(lib, files, progress)
		return
	}
	byTmdb := map[int]*library.Movie{}
	total := len(files)

	for i, f := range files {
		progress("tmdb", i, total, f.Parsed.Title)

		match, conf := matchMovie(f, client, overrides)
		if match == nil {
			lib.Unmatched = append(lib.Unmatched, library.Unmatched{
				Path: f.RelPath, ParsedTitle: f.Parsed.Title, Year: f.Parsed.Year,
				Reason: "aucun résultat TMDB", SizeBytes: f.Size,
			})
			continue
		}

		v := makeVersion(f)
		if m, ok := byTmdb[match.ID]; ok {
			m.Versions = append(m.Versions, v)
			if conf > m.MatchConfidence {
				m.MatchConfidence = conf
				m.NeedsReview = conf < confidenceThreshold
			}
			continue
		}

		movie := &library.Movie{
			ID:              library.PathID("mv", f.RelPath),
			TmdbID:          match.ID,
			MatchConfidence: conf,
			NeedsReview:     conf < confidenceThreshold,
			Title:           match.DisplayTitle(),
			OriginalTitle:   match.DisplayOriginal(),
			Year:            match.Year(),
			Overview:        match.Overview,
			VoteAverage:     match.VoteAverage,
			AgeRating:       -1,
			Versions:        []*library.Version{v},
		}
		// détails (genres, durée, synopsis fr complet, classification)
		if d, err := client.MovieDetails(match.ID, client.Lang); err == nil {
			movie.Title = or(d.Title, movie.Title)
			movie.OriginalTitle = or(d.OriginalTitle, movie.OriginalTitle)
			movie.Overview = or(d.Overview, movie.Overview)
			movie.RuntimeMin = d.Runtime
			movie.VoteAverage = d.VoteAverage
			movie.AgeRating = d.AgeRating()
			for _, g := range d.Genres {
				movie.Genres = append(movie.Genres, g.Name)
			}
			if movie.Overview == "" {
				if den, err := client.MovieDetails(match.ID, "en-US"); err == nil {
					movie.Overview = den.Overview
				}
			}
			if p, err := tmdb.DownloadImage(d.PosterPath, "w342", dataDir, "posters", "mv_"+fmt.Sprint(match.ID)); err == nil {
				movie.Poster = p
			}
			if b, err := tmdb.DownloadImage(d.BackdropPath, "w780", dataDir, "backdrops", "mv_"+fmt.Sprint(match.ID)); err == nil {
				movie.Backdrop = b
			}
		} else {
			if p, err := tmdb.DownloadImage(match.PosterPath, "w342", dataDir, "posters", "mv_"+fmt.Sprint(match.ID)); err == nil {
				movie.Poster = p
			}
		}
		byTmdb[match.ID] = movie
		lib.Movies = append(lib.Movies, movie)
	}
	progress("tmdb", total, total, "")
}

// buildMoviesLocal construit les fiches films sans TMDB, à partir du seul
// parsing des noms de fichiers (aucun appel réseau). Les fichiers d'un même
// titre + année sont regroupés en une fiche, poster laissé vide (l'UI affiche
// alors un placeholder).
func buildMoviesLocal(lib *library.Library, files []*scanner.FileEntry, progress Progress) {
	byKey := map[string]*library.Movie{}
	total := len(files)
	for i, f := range files {
		progress("local", i, total, f.Parsed.Title)

		title := f.Parsed.Title
		if title == "" {
			title = strings.TrimSpace(f.RawName)
		}
		key := parse.Fold(title)
		if f.Parsed.Year > 0 {
			key = fmt.Sprintf("%s|%d", key, f.Parsed.Year)
		}

		v := makeVersion(f)
		if m, ok := byKey[key]; ok {
			m.Versions = append(m.Versions, v)
			continue
		}
		m := &library.Movie{
			ID:              library.PathID("mv", f.RelPath),
			MatchConfidence: 1, // pas de revue en mode local
			Title:           title,
			Year:            f.Parsed.Year,
			AgeRating:       -1,
			Versions:        []*library.Version{v},
		}
		byKey[key] = m
		lib.Movies = append(lib.Movies, m)
	}
	progress("local", total, total, "")
}

func makeVersion(f *scanner.FileEntry) *library.Version {
	return &library.Version{
		Path:         f.RelPath,
		Parts:        f.PartPaths,
		SizeBytes:    f.Size,
		Container:    f.Container,
		Resolution:   f.Parsed.Resolution,
		LangBadge:    f.Parsed.Badge(),
		LangTags:     f.Parsed.LangTags,
		ExternalSubs: dedupSubs(f.ExternalSubs),
		IsISO:        f.Parsed.IsISO,
		Edition:      f.Parsed.Edition,
		RawName:      f.RawName,
		AddedAt:      f.ModTime,
	}
}

func dedupSubs(subs []library.ExternalSub) []library.ExternalSub {
	seen := map[string]bool{}
	var out []library.ExternalSub
	for _, s := range subs {
		if !seen[s.Path] {
			seen[s.Path] = true
			out = append(out, s)
		}
	}
	return out
}

// matchMovie applique l'échelle de recherche TMDB et retourne le meilleur
// résultat avec son score de confiance.
func matchMovie(f *scanner.FileEntry, client *tmdb.Client, overrides map[string]Override) (*tmdb.SearchResult, float64) {
	if ov := findOverride(overrides, f.RelPath); ov != nil && ov.MediaType == "movie" {
		if d, err := client.MovieDetails(ov.TmdbID, client.Lang); err == nil {
			return &tmdb.SearchResult{
				ID: d.ID, Title: d.Title, OriginalTitle: d.OriginalTitle,
				ReleaseDate: d.ReleaseDate, Overview: d.Overview,
				PosterPath: d.PosterPath, BackdropPath: d.BackdropPath,
				VoteAverage: d.VoteAverage,
			}, 1.0
		}
		return nil, 0
	}

	var best *tmdb.SearchResult
	var bestScore float64
	consider := func(results []tmdb.SearchResult, queryTitle string) {
		for i := range results {
			r := &results[i]
			s := score(queryTitle, f.Parsed.Year, r)
			if s > bestScore {
				best, bestScore = r, s
			}
		}
	}

	for _, q := range titleVariants(f.Parsed.Title) {
		if q == "" {
			continue
		}
		if f.Parsed.Year > 0 {
			if rs, err := client.SearchMovie(q, f.Parsed.Year); err == nil {
				consider(rs, q)
			}
			if bestScore >= 0.85 {
				return best, bestScore
			}
		}
		if rs, err := client.SearchMovie(q, 0); err == nil {
			consider(rs, q)
		}
		if bestScore >= 0.85 {
			return best, bestScore
		}
	}
	// dernier recours : multi
	if bestScore < confidenceThreshold {
		if rs, err := client.SearchMulti(f.Parsed.Title); err == nil {
			for i := range rs {
				if rs[i].MediaType != "movie" {
					continue
				}
				s := score(f.Parsed.Title, f.Parsed.Year, &rs[i])
				if s > bestScore {
					best, bestScore = &rs[i], s
				}
			}
		}
	}
	if best == nil || bestScore < 0.2 {
		return nil, 0
	}
	return best, bestScore
}

var (
	reParenGroup = regexp.MustCompile(`\([^)]*\)`)
	reAKA        = regexp.MustCompile(`(?i)\bA\.?K\.?A\.?\b`)
)

// titleVariants génère des variantes de recherche pour un titre ambigu :
// tel quel, sans parenthèses, contenu des parenthèses, segments autour de
// " - " et de "AKA", et en dernier recours le titre amputé du dernier mot
// (noms de réalisateurs collés au titre).
func titleVariants(title string) []string {
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(strings.Trim(s, "-–— .,"))
		if s == "" {
			return
		}
		for _, x := range out {
			if x == s {
				return
			}
		}
		out = append(out, s)
	}
	// parenthèse ouverte jamais refermée ("Hiroshima mon amour (Alain Resnais") :
	// on coupe à la dernière parenthèse orpheline
	if strings.Count(title, "(") > strings.Count(title, ")") {
		if i := strings.LastIndex(title, "("); i > 0 {
			title = strings.TrimSpace(title[:i])
		}
	}
	add(title)
	if noParen := reParenGroup.ReplaceAllString(title, " "); noParen != title {
		add(noParen)
	}
	for _, m := range reParenGroup.FindAllString(title, -1) {
		add(strings.Trim(m, "()"))
	}
	if m := reAKA.FindStringIndex(title); m != nil {
		add(title[:m[0]])
		add(title[m[1]:])
	}
	if i := strings.Index(title, " - "); i > 0 {
		add(title[:i])
		add(title[i+3:])
	}
	// dernier recours : sans le dernier mot (souvent un nom de réalisateur)
	base := reParenGroup.ReplaceAllString(title, " ")
	words := strings.Fields(base)
	if len(words) >= 3 {
		add(strings.Join(words[:len(words)-1], " "))
	}
	return out
}

// score = 0.7 × similarité titre (max fr/original) + 0.3 × concordance année.
func score(query string, year int, r *tmdb.SearchResult) float64 {
	simFr := parse.Similarity(query, r.DisplayTitle())
	simOrig := parse.Similarity(query, r.DisplayOriginal())
	sim := simFr
	if simOrig > sim {
		sim = simOrig
	}
	yearScore := 0.5 // année inconnue d'un côté : neutre
	ry := r.Year()
	if year > 0 && ry > 0 {
		diff := year - ry
		if diff < 0 {
			diff = -diff
		}
		switch {
		case diff == 0:
			yearScore = 1
		case diff == 1:
			yearScore = 0.8
		default:
			yearScore = 0
		}
	}
	s := 0.7*sim + 0.3*yearScore
	// petit bonus de popularité pour départager les homonymes
	if r.VoteCount > 100 {
		s += 0.02
	}
	return s
}

func findOverride(overrides map[string]Override, relPath string) *Override {
	if overrides == nil {
		return nil
	}
	p := relPath
	for {
		if ov, ok := overrides[p]; ok {
			return &ov
		}
		i := strings.LastIndex(p, "/")
		if i < 0 {
			return nil
		}
		p = p[:i]
	}
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// --- Séries ---

func buildSeries(lib *library.Library, files []*scanner.FileEntry, client *tmdb.Client, dataDir string, overrides map[string]Override, progress Progress) {
	if !client.HasKey() {
		buildSeriesLocal(lib, files, progress)
		return
	}
	// regrouper par titre de série plié
	groups := map[string][]*scanner.FileEntry{}
	var order []string
	for _, f := range files {
		key := parse.Fold(f.SeriesTitle)
		if key == "" {
			key = "?"
		}
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], f)
	}
	sort.Strings(order)

	byTmdb := map[int]*library.Series{}
	total := len(order)

	for gi, key := range order {
		group := groups[key]
		title := group[0].SeriesTitle
		progress("series", gi, total, title)

		match, conf := matchTV(title, group[0].RelPath, client, overrides)
		if match == nil {
			for _, f := range group {
				lib.Unmatched = append(lib.Unmatched, library.Unmatched{
					Path: f.RelPath, ParsedTitle: title, Reason: "série non trouvée sur TMDB", SizeBytes: f.Size,
				})
			}
			continue
		}

		s, ok := byTmdb[match.ID]
		if !ok {
			s = &library.Series{
				ID:              library.PathID("tv", group[0].RelPath),
				TmdbID:          match.ID,
				MatchConfidence: conf,
				NeedsReview:     conf < confidenceThreshold,
				Title:           match.DisplayTitle(),
				OriginalTitle:   match.DisplayOriginal(),
				Year:            match.Year(),
				Overview:        match.Overview,
				VoteAverage:     match.VoteAverage,
				AgeRating:       -1,
			}
			if d, err := client.TVDetails(match.ID, client.Lang); err == nil {
				s.Title = or(d.Name, s.Title)
				s.Overview = or(d.Overview, s.Overview)
				s.VoteAverage = d.VoteAverage
				s.AgeRating = d.AgeRating()
				for _, g := range d.Genres {
					s.Genres = append(s.Genres, g.Name)
				}
				if p, err := tmdb.DownloadImage(d.PosterPath, "w342", dataDir, "posters", "tv_"+fmt.Sprint(match.ID)); err == nil {
					s.Poster = p
				}
				if b, err := tmdb.DownloadImage(d.BackdropPath, "w780", dataDir, "backdrops", "tv_"+fmt.Sprint(match.ID)); err == nil {
					s.Backdrop = b
				}
			}
			byTmdb[match.ID] = s
			lib.Series = append(lib.Series, s)
		}

		// dossiers sources + drapeau « égaré dans [Films] »
		folderSeen := map[string]bool{}
		for _, x := range s.Folders {
			folderSeen[x] = true
		}
		for _, f := range group {
			top := f.RelPath
			if i := strings.Index(top, "/"); i > 0 {
				if j := strings.Index(top[i+1:], "/"); j > 0 {
					top = top[:i+1+j]
				}
			}
			if !folderSeen[top] {
				folderSeen[top] = true
				s.Folders = append(s.Folders, top)
			}
			if f.FromFilms {
				s.Misplaced = true
			}
		}

		attachEpisodes(s, group, client, dataDir)
	}
	progress("series", total, total, "")
}

// buildSeriesLocal construit les fiches séries sans TMDB, à partir du parsing
// des noms de fichiers (titre de série, saison, épisode).
func buildSeriesLocal(lib *library.Library, files []*scanner.FileEntry, progress Progress) {
	groups := map[string][]*scanner.FileEntry{}
	var order []string
	for _, f := range files {
		key := parse.Fold(f.SeriesTitle)
		if key == "" {
			key = "?"
		}
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], f)
	}
	sort.Strings(order)
	total := len(order)

	for gi, key := range order {
		group := groups[key]
		title := group[0].SeriesTitle
		progress("local", gi, total, title)

		s := &library.Series{
			ID:              library.PathID("tv", group[0].RelPath),
			MatchConfidence: 1,
			Title:           title,
			AgeRating:       -1,
		}
		folderSeen := map[string]bool{}
		for _, f := range group {
			top := f.RelPath
			if i := strings.Index(top, "/"); i > 0 {
				if j := strings.Index(top[i+1:], "/"); j > 0 {
					top = top[:i+1+j]
				}
			}
			if !folderSeen[top] {
				folderSeen[top] = true
				s.Folders = append(s.Folders, top)
			}
			if f.FromFilms {
				s.Misplaced = true
			}
		}
		attachEpisodesLocal(s, group)
		lib.Series = append(lib.Series, s)
	}
	progress("local", total, total, "")
}

// attachEpisodesLocal rattache les fichiers aux saisons/épisodes sans TMDB.
func attachEpisodesLocal(s *library.Series, group []*scanner.FileEntry) {
	seasons := map[int]*library.Season{}
	for _, f := range group {
		sn := f.Season
		if sn == 0 {
			sn = 1
		}
		season, ok := seasons[sn]
		if !ok {
			season = &library.Season{Number: sn}
			seasons[sn] = season
			s.Seasons = append(s.Seasons, season)
		}
		var ep *library.Episode
		for _, e := range season.Episodes {
			if e.Episode == f.Episode {
				ep = e
				break
			}
		}
		if ep == nil {
			ep = &library.Episode{Season: sn, Episode: f.Episode}
			season.Episodes = append(season.Episodes, ep)
		}
		ep.Files = append(ep.Files, makeVersion(f))
	}
	sort.Slice(s.Seasons, func(i, j int) bool { return s.Seasons[i].Number < s.Seasons[j].Number })
	for _, se := range s.Seasons {
		sort.Slice(se.Episodes, func(i, j int) bool { return se.Episodes[i].Episode < se.Episodes[j].Episode })
	}
}

func matchTV(title, relPath string, client *tmdb.Client, overrides map[string]Override) (*tmdb.SearchResult, float64) {
	if ov := findOverride(overrides, relPath); ov != nil && ov.MediaType == "tv" {
		if d, err := client.TVDetails(ov.TmdbID, client.Lang); err == nil {
			return &tmdb.SearchResult{
				ID: d.ID, Name: d.Name, OriginalName: d.OriginalName,
				FirstAirDate: d.FirstAirDate, Overview: d.Overview,
				PosterPath: d.PosterPath, BackdropPath: d.BackdropPath,
				VoteAverage: d.VoteAverage,
			}, 1.0
		}
		return nil, 0
	}
	var best *tmdb.SearchResult
	var bestScore float64
	for _, q := range titleVariants(title) {
		rs, err := client.SearchTV(q)
		if err != nil {
			continue
		}
		for i := range rs {
			s := score(q, 0, &rs[i])
			if s > bestScore {
				best, bestScore = &rs[i], s
			}
		}
		if bestScore >= 0.85 {
			break
		}
	}
	if best == nil || bestScore < 0.2 {
		return nil, 0
	}
	return best, bestScore
}

func attachEpisodes(s *library.Series, group []*scanner.FileEntry, client *tmdb.Client, dataDir string) {
	seasons := map[int]*library.Season{}
	for _, se := range s.Seasons {
		seasons[se.Number] = se
	}

	for _, f := range group {
		sn := f.Season
		if sn == 0 {
			sn = 1
		}
		season, ok := seasons[sn]
		if !ok {
			season = &library.Season{Number: sn}
			if sd, err := client.SeasonDetails(s.TmdbID, sn); err == nil {
				season.Name = sd.Name
				season.Overview = sd.Overview
				if p, err := tmdb.DownloadImage(sd.PosterPath, "w342", dataDir, "posters", fmt.Sprintf("tv_%d_s%d", s.TmdbID, sn)); err == nil {
					season.Poster = p
				}
			}
			seasons[sn] = season
			s.Seasons = append(s.Seasons, season)
		}

		var ep *library.Episode
		for _, e := range season.Episodes {
			if e.Episode == f.Episode {
				ep = e
				break
			}
		}
		if ep == nil {
			ep = &library.Episode{Season: sn, Episode: f.Episode}
			if sd, err := client.SeasonDetails(s.TmdbID, sn); err == nil {
				for _, e := range sd.Episodes {
					if e.EpisodeNumber == f.Episode {
						ep.Name = e.Name
						ep.Overview = e.Overview
						ep.AirDate = e.AirDate
						break
					}
				}
			}
			season.Episodes = append(season.Episodes, ep)
		}
		ep.Files = append(ep.Files, makeVersion(f))
	}

	sort.Slice(s.Seasons, func(i, j int) bool { return s.Seasons[i].Number < s.Seasons[j].Number })
	for _, se := range s.Seasons {
		sort.Slice(se.Episodes, func(i, j int) bool { return se.Episodes[i].Episode < se.Episodes[j].Episode })
	}
}

// attachOrphanSubs rattache les sous-titres trouvés dans des dossiers sans
// vidéo (ex. les .srt de Vikings dans [Séries] alors que les vidéos sont
// dans [Films]) aux épisodes correspondants, par titre de série + SxxEyy.
func attachOrphanSubs(lib *library.Library, orphans []string) {
	if len(orphans) == 0 {
		return
	}
	// index série pliée -> série
	byTitle := map[string]*library.Series{}
	for _, s := range lib.Series {
		byTitle[parse.Fold(s.Title)] = s
		byTitle[parse.Fold(s.OriginalTitle)] = s
	}
	for _, sub := range orphans {
		base := sub
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		p := parse.ParseName(base)
		if !p.HasEpisode || p.Title == "" {
			continue
		}
		s := byTitle[parse.Fold(p.Title)]
		if s == nil {
			continue
		}
		for _, season := range s.Seasons {
			if season.Number != p.Season {
				continue
			}
			for _, ep := range season.Episodes {
				if ep.Episode != p.Episode {
					continue
				}
				es := library.ExternalSub{Path: sub, Lang: parse.SubLang(base)}
				for _, f := range ep.Files {
					exists := false
					for _, x := range f.ExternalSubs {
						if x.Path == sub {
							exists = true
						}
					}
					if !exists {
						f.ExternalSubs = append(f.ExternalSubs, es)
					}
				}
			}
		}
	}
}
