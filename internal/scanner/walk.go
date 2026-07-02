// Package scanner parcourt le disque et produit des candidats films/épisodes
// à partir des fichiers vidéo trouvés (aucun accès réseau).
package scanner

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"absolute_cinema/internal/library"
	"absolute_cinema/internal/parse"
)

// FileEntry est un fichier vidéo candidat, avec son analyse de nom.
type FileEntry struct {
	RelPath      string // relatif au mediaRoot, séparateur /
	Size         int64
	RawName      string
	Container    string
	Parsed       parse.Parsed // meilleure analyse (fichier vs dossier)
	SeriesTitle  string       // si épisode : titre de série deviné
	Season       int
	Episode      int
	IsEpisode    bool
	FromFilms    bool // trouvé sous [Films]
	ExternalSubs []library.ExternalSub
	PartPaths    []string // fichiers CD2, CD3… regroupés sous cette entrée
}

// Result est la sortie brute du scan filesystem.
type Result struct {
	Movies     []*FileEntry
	Episodes   []*FileEntry
	Skipped    []string // junk ignoré (info)
	OrphanSubs []string // sous-titres sans vidéo dans leur dossier (chemins relatifs)
}

var junkDirs = map[string]bool{
	"plex versions": true, "sample": true, "samples": true, "extras": true,
	"$recycle.bin": true, "system volume information": true,
}

var reJunkVideo = regexp.MustCompile(`(?i)^sample[\s._-]?|[\s._-]sample\.|^(rarbg|etrg)\b|rarbg\.com|^encoded by|^advert`)

const junkSizeLimit = 300 << 20 // 300 Mo

// épisode de secours : "Ep 12", "E12", " - 12 - ", numéro final
var reEpFallback = regexp.MustCompile(`(?i)(?:\be?p?[\s._-]?(\d{1,3})\b)[^\d]*$`)

// Scan parcourt mediaRoot. filmsDirs et seriesDirs sont les noms des dossiers
// racine (ex. "[Films]", "[Séries]").
func Scan(mediaRoot string, filmsDirs, seriesDirs []string) (*Result, error) {
	res := &Result{}
	for _, d := range filmsDirs {
		if err := walkFilms(mediaRoot, d, res); err != nil {
			return nil, err
		}
	}
	for _, d := range seriesDirs {
		if err := walkSeries(mediaRoot, d, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// DetectRoots trouve les dossiers [Films] / [Séries] sous mediaRoot.
func DetectRoots(mediaRoot string) (films, series []string, err error) {
	entries, err := os.ReadDir(mediaRoot)
	if err != nil {
		return nil, nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		switch {
		case strings.Contains(name, "série") || strings.Contains(name, "serie") || strings.Contains(name, "series") || strings.Contains(name, "show") || name == "tv" || strings.Contains(name, "[tv]"):
			series = append(series, e.Name())
		case strings.Contains(name, "film") || strings.Contains(name, "movie"):
			films = append(films, e.Name())
		}
	}
	return films, series, nil
}

type dirContent struct {
	videos []videoFile
	subs   []string // chemins relatifs
}

type videoFile struct {
	relPath string
	size    int64
}

// collectDirs regroupe vidéos et sous-titres par dossier sous root (relatif au mediaRoot).
func collectDirs(mediaRoot, root string, skipped *[]string) (map[string]*dirContent, error) {
	dirs := map[string]*dirContent{}
	absRoot := filepath.Join(mediaRoot, filepath.FromSlash(root))
	err := filepath.WalkDir(absRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // dossier illisible : on continue
		}
		name := d.Name()
		if d.IsDir() {
			if strings.HasPrefix(name, ".") || junkDirs[strings.ToLower(name)] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(mediaRoot, p)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		dir := path.Dir(rel)
		ext := strings.ToLower(filepath.Ext(name))
		if parse.IsSubExt(ext) {
			dc := dirs[dir]
			if dc == nil {
				dc = &dirContent{}
				dirs[dir] = dc
			}
			dc.subs = append(dc.subs, rel)
			return nil
		}
		if !parse.IsVideoExt(ext) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if reJunkVideo.MatchString(name) && info.Size() < junkSizeLimit {
			*skipped = append(*skipped, rel)
			return nil
		}
		if info.Size() == 0 {
			*skipped = append(*skipped, rel)
			return nil
		}
		dc := dirs[dir]
		if dc == nil {
			dc = &dirContent{}
			dirs[dir] = dc
		}
		dc.videos = append(dc.videos, videoFile{relPath: rel, size: info.Size()})
		return nil
	})
	return dirs, err
}

func walkFilms(mediaRoot, filmsDir string, res *Result) error {
	dirs, err := collectDirs(mediaRoot, filmsDir, &res.Skipped)
	if err != nil {
		return err
	}
	for dir, dc := range dirs {
		if len(dc.videos) == 0 {
			res.OrphanSubs = append(res.OrphanSubs, dc.subs...)
			continue
		}
		entries := buildEntries(mediaRoot, filmsDir, dir, dc, true)
		for _, e := range entries {
			if e.IsEpisode {
				res.Episodes = append(res.Episodes, e)
			} else {
				res.Movies = append(res.Movies, e)
			}
		}
	}
	return nil
}

func walkSeries(mediaRoot, seriesDir string, res *Result) error {
	dirs, err := collectDirs(mediaRoot, seriesDir, &res.Skipped)
	if err != nil {
		return err
	}
	for dir, dc := range dirs {
		if len(dc.videos) == 0 {
			res.OrphanSubs = append(res.OrphanSubs, dc.subs...)
			continue
		}
		entries := buildEntries(mediaRoot, seriesDir, dir, dc, false)
		for _, e := range entries {
			// sous [Séries], tout est épisode ; le titre de série vient du
			// dossier de tête si le nom de fichier n'a pas suffi
			if !e.IsEpisode {
				e.IsEpisode = true
				if e.Season == 0 {
					e.Season = seasonFromAncestors(dir, seriesDir)
				}
				if e.Episode == 0 {
					e.Episode = fallbackEpisode(e.RawName)
				}
			}
			if e.SeriesTitle == "" || topFolder(e.RelPath, seriesDir) != "" {
				// le dossier de tête sous [Séries] est le nom canonique de la série
				if tf := topFolder(e.RelPath, seriesDir); tf != "" {
					e.SeriesTitle = parse.ParseName(tf).Title
				}
			}
			res.Episodes = append(res.Episodes, e)
		}
	}
	return nil
}

// topFolder retourne le premier composant du chemin sous rootDir ("" si le
// fichier est directement à la racine).
func topFolder(relPath, rootDir string) string {
	rest := strings.TrimPrefix(relPath, rootDir+"/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func seasonFromAncestors(dir, rootDir string) int {
	rest := strings.TrimPrefix(dir, rootDir+"/")
	parts := strings.Split(rest, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if s := parse.SeasonFromFolder(parts[i]); s > 0 {
			return s
		}
	}
	return 0
}

func fallbackEpisode(name string) int {
	if i := strings.LastIndex(name, "."); i > 0 {
		name = name[:i]
	}
	if m := reEpFallback.FindStringSubmatch(name); m != nil {
		n, _ := strconv.Atoi(m[1])
		if n >= 1 && n <= 500 {
			return n
		}
	}
	return 0
}

// buildEntries construit les FileEntry d'un dossier : choix fichier-vs-dossier,
// regroupement CD1/CD2, association des sous-titres.
func buildEntries(mediaRoot, rootDir, dir string, dc *dirContent, fromFilms bool) []*FileEntry {
	dirName := path.Base(dir)
	isRoot := dir == rootDir
	folderP := parse.Parsed{}
	if !isRoot {
		folderP = parse.ParseName(dirName)
	}

	var entries []*FileEntry
	for _, v := range dc.videos {
		base := path.Base(v.relPath)
		fileP := parse.ParseName(base)
		chosen := chooseParse(fileP, folderP, isRoot, len(dc.videos))
		e := &FileEntry{
			RelPath:   v.relPath,
			Size:      v.size,
			RawName:   base,
			Container: strings.TrimPrefix(strings.ToLower(path.Ext(base)), "."),
			Parsed:    chosen,
			FromFilms: fromFilms,
		}
		if chosen.HasEpisode {
			e.IsEpisode = true
			e.Season, e.Episode = chosen.Season, chosen.Episode
			e.SeriesTitle = chosen.Title
			if e.SeriesTitle == "" && !isRoot {
				e.SeriesTitle = folderP.Title
			}
		}
		entries = append(entries, e)
	}

	entries = groupParts(entries)
	attachSubs(entries, dc.subs)
	return entries
}

// chooseParse choisit la meilleure analyse entre nom de fichier et nom de dossier.
func chooseParse(fileP, folderP parse.Parsed, isRoot bool, nVideos int) parse.Parsed {
	if isRoot || folderP.Title == "" {
		return fileP
	}
	// un SxxEyy dans le nom de fichier prime toujours
	if fileP.HasEpisode {
		return mergeTags(fileP, folderP)
	}
	if folderP.HasEpisode && fileP.Title == "" {
		return folderP
	}
	if nVideos > 1 {
		// dossier multi-films (trilogies, collections) : le fichier prime
		if fileP.Title == "" {
			return mergeTags(folderP, fileP)
		}
		return mergeTags(fileP, folderP)
	}
	// un seul fichier vidéo dans le dossier
	switch {
	case fileP.Year > 0 && folderP.Year == 0:
		return mergeTags(fileP, folderP)
	case folderP.Year > 0 && fileP.Year == 0:
		return mergeTags(folderP, fileP)
	case fileP.Year > 0: // les deux ont une année : le fichier est plus précis
		return mergeTags(fileP, folderP)
	default: // aucune année : le dossier est en général plus propre
		if len(folderP.Title) >= 3 {
			return mergeTags(folderP, fileP)
		}
		return mergeTags(fileP, folderP)
	}
}

// mergeTags complète primary avec les tags de langue/résolution/édition de other.
func mergeTags(primary, other parse.Parsed) parse.Parsed {
	for _, t := range other.LangTags {
		found := false
		for _, x := range primary.LangTags {
			if x == t {
				found = true
			}
		}
		if !found {
			primary.LangTags = append(primary.LangTags, t)
		}
	}
	for _, l := range other.AudioLangs {
		found := false
		for _, x := range primary.AudioLangs {
			if x == l {
				found = true
			}
		}
		if !found {
			primary.AudioLangs = append(primary.AudioLangs, l)
		}
	}
	for _, l := range other.SubLangs {
		found := false
		for _, x := range primary.SubLangs {
			if x == l {
				found = true
			}
		}
		if !found {
			primary.SubLangs = append(primary.SubLangs, l)
		}
	}
	if primary.Resolution == "" {
		primary.Resolution = other.Resolution
	}
	if primary.Edition == "" {
		primary.Edition = other.Edition
	}
	if primary.Year == 0 {
		primary.Year = other.Year
	}
	return primary
}

// groupParts fusionne les fichiers CD1/CD2 d'un même film en une seule entrée.
func groupParts(entries []*FileEntry) []*FileEntry {
	byTitle := map[string][]*FileEntry{}
	var order []string
	for _, e := range entries {
		key := parse.Fold(e.Parsed.Title)
		if e.Parsed.Part == 0 || e.IsEpisode {
			key = e.RelPath // pas de regroupement
		}
		if _, ok := byTitle[key]; !ok {
			order = append(order, key)
		}
		byTitle[key] = append(byTitle[key], e)
	}
	var out []*FileEntry
	for _, key := range order {
		group := byTitle[key]
		if len(group) == 1 {
			out = append(out, group[0])
			continue
		}
		// trier par numéro de partie
		primary := group[0]
		for _, e := range group {
			if e.Parsed.Part < primary.Parsed.Part {
				primary = e
			}
		}
		for _, e := range group {
			if e == primary {
				continue
			}
			primary.Size += e.Size
			primary.ExternalSubs = append(primary.ExternalSubs, e.ExternalSubs...)
			primary.PartPaths = append(primary.PartPaths, e.RelPath)
		}
		out = append(out, primary)
	}
	return out
}

// attachSubs associe les fichiers de sous-titres aux vidéos du même dossier.
func attachSubs(entries []*FileEntry, subs []string) {
	if len(subs) == 0 || len(entries) == 0 {
		return
	}
	for _, sub := range subs {
		subBase := path.Base(sub)
		subStem := stem(subBase)
		lang := parse.SubLang(subBase)
		es := library.ExternalSub{Path: sub, Lang: lang}

		// 1. correspondance SxxEyy
		sp := parse.ParseName(subBase)
		if sp.HasEpisode {
			matched := false
			for _, e := range entries {
				if e.IsEpisode && e.Season == sp.Season && e.Episode == sp.Episode {
					e.ExternalSubs = append(e.ExternalSubs, es)
					matched = true
				}
			}
			if matched {
				continue
			}
		}
		// 2. préfixe de nom le plus long
		var best *FileEntry
		bestLen := 0
		for _, e := range entries {
			vStem := stem(path.Base(e.RelPath))
			n := commonPrefix(strings.ToLower(vStem), strings.ToLower(subStem))
			if n > bestLen && n >= len(vStem)*3/4 {
				bestLen, best = n, e
			}
		}
		if best != nil {
			best.ExternalSubs = append(best.ExternalSubs, es)
			continue
		}
		// 3. vidéo unique dans le dossier
		if len(entries) == 1 {
			entries[0].ExternalSubs = append(entries[0].ExternalSubs, es)
		}
	}
}

func stem(name string) string {
	if i := strings.LastIndex(name, "."); i > 0 {
		return name[:i]
	}
	return name
}

func commonPrefix(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}
