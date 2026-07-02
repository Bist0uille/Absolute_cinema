package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"absolute_cinema/internal/catalog"
	"absolute_cinema/internal/config"
	"absolute_cinema/internal/library"
	"absolute_cinema/internal/player"
	"absolute_cinema/internal/tmdb"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// absMediaPath convertit un chemin relatif de la bibliothèque en chemin absolu,
// en refusant toute évasion hors du mediaRoot.
func (s *Server) absMediaPath(rel string) (string, error) {
	if rel == "" || strings.Contains(rel, "..") {
		return "", errors.New("chemin invalide")
	}
	abs := filepath.Join(s.Paths.MediaRoot, filepath.FromSlash(rel))
	root := filepath.Clean(s.Paths.MediaRoot)
	if !strings.HasPrefix(filepath.Clean(abs), root+string(filepath.Separator)) {
		return "", errors.New("chemin hors de la bibliothèque")
	}
	return abs, nil
}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lib := s.lib
	var unmatched any = s.lib.Unmatched
	if s.cfg.KidMode {
		lib = s.kidLibrary()
		unmatched = nil
	}
	writeJSON(w, 200, map[string]any{
		"scanned_at":  lib.ScannedAt,
		"movies":      lib.Movies,
		"series":      lib.Series,
		"unmatched":   unmatched,
		"readonly":    s.Paths.ReadOnly,
		"has_key":     s.cfg.TmdbAPIKey != "",
		"scanning":    s.scan.Scanning,
		"kid_mode":    s.cfg.KidMode,
		"kid_max_age": s.cfg.KidMaxAge,
		"has_pin":     s.cfg.KidPINHash != "",
	})
}

func (s *Server) handleScanStart(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	if !s.StartScan() {
		writeErr(w, 409, "un scan est déjà en cours")
		return
	}
	writeJSON(w, 200, map[string]bool{"started": true})
}

func (s *Server) handleScanStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, 200, s.scan)
}

type playRequest struct {
	Path    string   `json:"path"`
	Parts   []string `json:"parts"`
	SubPath string   `json:"sub_path"`
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	var req playRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	s.mu.Lock()
	if s.cfg.KidMode && !s.kidAllowedPath(req.Path) {
		s.mu.Unlock()
		writeErr(w, 403, "titre non autorisé en mode enfant")
		return
	}
	s.mu.Unlock()
	abs, err := s.absMediaPath(req.Path)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	var parts []string
	for _, p := range req.Parts {
		if a, err := s.absMediaPath(p); err == nil {
			parts = append(parts, a)
		}
	}
	sub := ""
	if req.SubPath != "" {
		if a, err := s.absMediaPath(req.SubPath); err == nil {
			sub = a
		}
	}
	res, err := player.Play(abs, parts, sub)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, res)
}

func (s *Server) handleDuplicates(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, 200, library.FindDuplicates(s.lib))
}

// --- Corbeille ---

type trashRequest struct {
	Paths []string `json:"paths"`
}

type trashEntry struct {
	OriginalPath string    `json:"original_path"`
	TrashPath    string    `json:"trash_path"` // relatif au TrashDir
	DeletedAt    time.Time `json:"deleted_at"`
	SizeBytes    int64     `json:"size_bytes"`
}

func (s *Server) trashJournalFile() string {
	return filepath.Join(s.Paths.TrashDir(), "journal.json")
}

func (s *Server) loadTrashJournal() []trashEntry {
	var j []trashEntry
	if data, err := os.ReadFile(s.trashJournalFile()); err == nil {
		_ = json.Unmarshal(data, &j)
	}
	return j
}

func (s *Server) saveTrashJournal(j []trashEntry) error {
	if err := os.MkdirAll(s.Paths.TrashDir(), 0o755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(j, "", " ")
	return os.WriteFile(s.trashJournalFile(), data, 0o644)
}

func (s *Server) handleTrash(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	if s.Paths.ReadOnly {
		writeErr(w, 403, "disque en lecture seule : suppression impossible")
		return
	}
	var req trashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Paths) == 0 {
		writeErr(w, 400, "requête invalide")
		return
	}
	journal := s.loadTrashJournal()
	stamp := time.Now().Format("2006-01-02_150405")
	var moved, failed []string
	for _, rel := range req.Paths {
		abs, err := s.absMediaPath(rel)
		if err != nil {
			failed = append(failed, rel+" : "+err.Error())
			continue
		}
		st, err := os.Stat(abs)
		if err != nil {
			failed = append(failed, rel+" : introuvable")
			continue
		}
		trashRel := filepath.Join(stamp, filepath.FromSlash(rel))
		dest := filepath.Join(s.Paths.TrashDir(), trashRel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			failed = append(failed, rel+" : "+err.Error())
			continue
		}
		if err := os.Rename(abs, dest); err != nil {
			failed = append(failed, rel+" : "+err.Error())
			continue
		}
		journal = append(journal, trashEntry{
			OriginalPath: rel, TrashPath: filepath.ToSlash(trashRel),
			DeletedAt: time.Now().UTC(), SizeBytes: st.Size(),
		})
		moved = append(moved, rel)
	}
	_ = s.saveTrashJournal(journal)
	s.removeFromLibrary(moved)
	writeJSON(w, 200, map[string]any{"moved": moved, "failed": failed})
}

// removeFromLibrary retire de la bibliothèque en mémoire les versions déplacées
// en corbeille, puis persiste.
func (s *Server) removeFromLibrary(paths []string) {
	if len(paths) == 0 {
		return
	}
	gone := map[string]bool{}
	for _, p := range paths {
		gone[p] = true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var movies []*library.Movie
	for _, m := range s.lib.Movies {
		var vs []*library.Version
		for _, v := range m.Versions {
			if !gone[v.Path] {
				vs = append(vs, v)
			}
		}
		m.Versions = vs
		if len(vs) > 0 {
			movies = append(movies, m)
		}
	}
	s.lib.Movies = movies
	for _, se := range s.lib.Series {
		for _, season := range se.Seasons {
			for _, ep := range season.Episodes {
				var fs []*library.Version
				for _, f := range ep.Files {
					if !gone[f.Path] {
						fs = append(fs, f)
					}
				}
				ep.Files = fs
			}
		}
	}
	_ = library.Save(s.lib, s.Paths.LibraryFile())
}

func (s *Server) handleTrashList(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	writeJSON(w, 200, s.loadTrashJournal())
}

type restoreRequest struct {
	TrashPath string `json:"trash_path"`
}

func (s *Server) handleTrashRestore(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	var req restoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	journal := s.loadTrashJournal()
	var kept []trashEntry
	restored := false
	for _, e := range journal {
		if e.TrashPath == req.TrashPath && !restored {
			src := filepath.Join(s.Paths.TrashDir(), filepath.FromSlash(e.TrashPath))
			dst, err := s.absMediaPath(e.OriginalPath)
			if err == nil {
				_ = os.MkdirAll(filepath.Dir(dst), 0o755)
				if os.Rename(src, dst) == nil {
					restored = true
					continue
				}
			}
		}
		kept = append(kept, e)
	}
	_ = s.saveTrashJournal(kept)
	if !restored {
		writeErr(w, 404, "restauration impossible")
		return
	}
	writeJSON(w, 200, map[string]any{"restored": true, "note": "relancez un scan pour réintégrer le fichier à la bibliothèque"})
}

// --- Config ---

func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	masked := ""
	if k := s.cfg.TmdbAPIKey; len(k) > 6 {
		masked = k[:3] + "…" + k[len(k)-3:]
	} else if k != "" {
		masked = "•••"
	}
	writeJSON(w, 200, map[string]any{
		"tmdb_key_masked": masked,
		"has_key":         s.cfg.TmdbAPIKey != "",
		"readonly":        s.Paths.ReadOnly,
		"media_root":      s.Paths.MediaRoot,
		"data_dir":        s.Paths.DataDir,
	})
}

type configRequest struct {
	TmdbAPIKey string `json:"tmdb_api_key"`
}

func (s *Server) handleConfigSet(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.TmdbAPIKey) == "" {
		writeErr(w, 400, "clé manquante")
		return
	}
	key := strings.TrimSpace(req.TmdbAPIKey)
	if err := tmdb.ValidateKey(key); err != nil {
		writeErr(w, 400, "clé refusée par TMDB : "+err.Error())
		return
	}
	s.mu.Lock()
	s.cfg.TmdbAPIKey = key
	err := config.Save(s.cfg, s.Paths.ConfigFile())
	s.mu.Unlock()
	if err != nil {
		writeErr(w, 500, "sauvegarde impossible : "+err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// --- Override (correction manuelle d'un match TMDB) ---

type overrideRequest struct {
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
	TmdbID    int    `json:"tmdb_id"`
}

func (s *Server) handleOverride(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	var req overrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		req.Path == "" || req.TmdbID == 0 || (req.MediaType != "movie" && req.MediaType != "tv") {
		writeErr(w, 400, "requête invalide (path, media_type movie|tv, tmdb_id)")
		return
	}
	s.mu.Lock()
	s.overrides[req.Path] = catalog.Override{MediaType: req.MediaType, TmdbID: req.TmdbID}
	err := config.SaveOverrides(s.overrides, s.Paths.OverridesFile())
	s.mu.Unlock()
	if err != nil {
		writeErr(w, 500, "sauvegarde overrides : "+err.Error())
		return
	}
	// re-scan pour appliquer (rapide : tout le reste vient du cache TMDB)
	s.StartScan()
	writeJSON(w, 200, map[string]bool{"ok": true, "rescan": true})
}

// --- Recherche TMDB pour l'UI de correction ---

func (s *Server) handleTmdbSearch(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	kind := r.URL.Query().Get("type")
	if q == "" {
		writeErr(w, 400, "paramètre q manquant")
		return
	}
	s.mu.Lock()
	key := s.cfg.TmdbAPIKey
	s.mu.Unlock()
	client := tmdb.New(key, s.Paths.TmdbCacheDir())
	var results []tmdb.SearchResult
	var err error
	if kind == "tv" {
		results, err = client.SearchTV(q)
	} else {
		results, err = client.SearchMovie(q, 0)
	}
	if err != nil {
		writeErr(w, 502, fmt.Sprintf("TMDB : %v", err))
		return
	}
	if len(results) > 8 {
		results = results[:8]
	}
	writeJSON(w, 200, results)
}
