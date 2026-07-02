// Package library définit le modèle de données de la bibliothèque et sa persistance.
package library

import (
	"crypto/sha1"
	"encoding/hex"
	"time"
)

// Library est le contenu complet de library.json.
type Library struct {
	SchemaVersion int        `json:"schema_version"`
	ScannedAt     time.Time  `json:"scanned_at"`
	Movies        []*Movie   `json:"movies"`
	Series        []*Series  `json:"series"`
	Unmatched     []Unmatched `json:"unmatched"`
}

// Movie est une fiche film (regroupe éventuellement plusieurs fichiers/versions).
type Movie struct {
	ID              string   `json:"id"`
	TmdbID          int      `json:"tmdb_id"`
	MatchConfidence float64  `json:"match_confidence"`
	NeedsReview     bool     `json:"needs_review"`
	Title           string   `json:"title_fr"`
	OriginalTitle   string   `json:"original_title"`
	Year            int      `json:"year"`
	Overview        string   `json:"overview_fr"`
	Genres          []string `json:"genres"`
	Poster          string   `json:"poster,omitempty"`   // chemin relatif à .absolute_cinema/
	Backdrop        string   `json:"backdrop,omitempty"`
	VoteAverage     float64  `json:"vote_average"`
	RuntimeMin      int      `json:"runtime_min"`
	AgeRating       int      `json:"age_rating"` // 0/10/12/16/18, -1 = inconnu
	Versions        []*Version `json:"versions"`
}

// Version est un fichier vidéo (ou un groupe de parties CD1/CD2) d'un film.
type Version struct {
	Path         string        `json:"path"` // relatif au mediaRoot, séparateur /
	Parts        []string      `json:"parts,omitempty"` // chemins additionnels (CD2…)
	SizeBytes    int64         `json:"size_bytes"`
	Container    string        `json:"container"`
	Resolution   string        `json:"resolution_tag,omitempty"`
	LangBadge    string        `json:"lang_badge,omitempty"`
	LangTags     []string      `json:"filename_lang_tags,omitempty"`
	AudioLangs   []string      `json:"audio_langs,omitempty"` // ffprobe (phase 2)
	SubLangs     []string      `json:"sub_langs,omitempty"`
	ExternalSubs []ExternalSub `json:"external_subs,omitempty"`
	IsISO        bool          `json:"is_iso,omitempty"`
	Edition      string        `json:"edition,omitempty"`
	RawName      string        `json:"raw_name"`
	Probed       bool          `json:"probed,omitempty"` // pistes analysées par ffprobe
}

// ExternalSub est un fichier de sous-titres externe associé.
type ExternalSub struct {
	Path string `json:"path"`
	Lang string `json:"lang,omitempty"`
}

// Series est une fiche série.
type Series struct {
	ID              string    `json:"id"`
	TmdbID          int       `json:"tmdb_id"`
	MatchConfidence float64   `json:"match_confidence"`
	NeedsReview     bool      `json:"needs_review"`
	Title           string    `json:"title_fr"`
	OriginalTitle   string    `json:"original_title"`
	Year            int       `json:"year"` // première diffusion
	Overview        string    `json:"overview_fr"`
	Genres          []string  `json:"genres"`
	Poster          string    `json:"poster,omitempty"`
	Backdrop        string    `json:"backdrop,omitempty"`
	VoteAverage     float64   `json:"vote_average"`
	AgeRating       int       `json:"age_rating"` // 0…18, -1 = inconnu
	Folders         []string  `json:"folders"` // dossiers sources (peut inclure [Films]/…)
	Misplaced       bool      `json:"misplaced,omitempty"` // trouvée dans [Films]
	Seasons         []*Season `json:"seasons"`
}

// Season est une saison d'une série.
type Season struct {
	Number   int        `json:"number"`
	Name     string     `json:"name_fr,omitempty"`
	Overview string     `json:"overview_fr,omitempty"`
	Poster   string     `json:"poster,omitempty"`
	Episodes []*Episode `json:"episodes"`
}

// Episode est un épisode avec son fichier vidéo.
type Episode struct {
	Season       int           `json:"season"`
	Episode      int           `json:"episode"`
	Name         string        `json:"name_fr,omitempty"`
	Overview     string        `json:"overview_fr,omitempty"`
	AirDate      string        `json:"air_date,omitempty"`
	StillPath    string        `json:"still,omitempty"`
	Files        []*Version    `json:"files"` // >1 = doublon d'épisode
}

// Unmatched est un fichier vidéo qu'on n'a pas su rattacher à TMDB.
type Unmatched struct {
	Path        string `json:"path"`
	ParsedTitle string `json:"parsed_title"`
	Year        int    `json:"year,omitempty"`
	Reason      string `json:"reason"`
	SizeBytes   int64  `json:"size_bytes"`
}

// PathID fabrique un identifiant stable à partir d'un chemin relatif.
func PathID(prefix, relPath string) string {
	h := sha1.Sum([]byte(relPath))
	return prefix + "_" + hex.EncodeToString(h[:6])
}
