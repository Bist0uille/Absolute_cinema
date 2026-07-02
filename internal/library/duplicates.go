package library

import "sort"

// DuplicateGroup est un film présent en plusieurs versions.
type DuplicateGroup struct {
	MovieID     string     `json:"movie_id"`
	Title       string     `json:"title"`
	Year        int        `json:"year"`
	Poster      string     `json:"poster,omitempty"`
	Versions    []*Version `json:"versions"`
	KeepPath    string     `json:"keep_path"`    // version recommandée à garder
	DeletePaths []string   `json:"delete_paths"` // versions recommandées à supprimer
}

// EpisodeDup est un épisode présent en plusieurs fichiers.
type EpisodeDup struct {
	SeriesID string     `json:"series_id"`
	Title    string     `json:"title"`
	Season   int        `json:"season"`
	Episode  int        `json:"episode"`
	Files    []*Version `json:"files"`
}

// Duplicates liste les doublons de films et d'épisodes, avec recommandation.
type Duplicates struct {
	Movies   []DuplicateGroup `json:"movies"`
	Episodes []EpisodeDup     `json:"episodes"`
	// Séries dont des fichiers traînent dans [Films]
	MisplacedSeries []string `json:"misplaced_series"`
}

// FindDuplicates détecte les doublons dans la bibliothèque.
func FindDuplicates(lib *Library) *Duplicates {
	d := &Duplicates{}
	for _, m := range lib.Movies {
		if len(m.Versions) < 2 {
			continue
		}
		// les fichiers multi-parties (CD1+CD2) regroupés ne comptent pas :
		// ici chaque Version est déjà un film complet
		g := DuplicateGroup{MovieID: m.ID, Title: m.Title, Year: m.Year, Poster: m.Poster, Versions: m.Versions}
		ranked := make([]*Version, len(m.Versions))
		copy(ranked, m.Versions)
		sort.SliceStable(ranked, func(i, j int) bool {
			return versionScore(ranked[i]) > versionScore(ranked[j])
		})
		g.KeepPath = ranked[0].Path
		for _, v := range ranked[1:] {
			g.DeletePaths = append(g.DeletePaths, v.Path)
		}
		d.Movies = append(d.Movies, g)
	}
	for _, s := range lib.Series {
		if s.Misplaced {
			d.MisplacedSeries = append(d.MisplacedSeries, s.Title)
		}
		for _, se := range s.Seasons {
			for _, ep := range se.Episodes {
				if len(ep.Files) > 1 {
					d.Episodes = append(d.Episodes, EpisodeDup{
						SeriesID: s.ID, Title: s.Title, Season: se.Number, Episode: ep.Episode, Files: ep.Files,
					})
				}
			}
		}
	}
	return d
}

// versionScore classe les versions : résolution > langues > codec > taille.
func versionScore(v *Version) float64 {
	score := 0.0
	switch v.Resolution {
	case "2160p":
		score += 4000
	case "1080p":
		score += 3000
	case "720p":
		score += 2000
	case "576p", "480p":
		score += 1000
	}
	switch v.LangBadge {
	case "MULTI":
		score += 400
	case "VOSTFR":
		score += 300
	case "VF":
		score += 200
	case "VO":
		score += 150
	}
	score += float64(len(v.ExternalSubs)) * 10
	if v.IsISO {
		score -= 500 // moins pratique à lire
	}
	// à critères égaux, le plus gros fichier gagne (1 point par Go)
	score += float64(v.SizeBytes) / (1 << 30)
	return score
}
