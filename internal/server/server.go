// Package server expose l'API HTTP et l'interface web embarquée.
package server

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"absolute_cinema/internal/catalog"
	"absolute_cinema/internal/config"
	"absolute_cinema/internal/fsutil"
	"absolute_cinema/internal/library"
	"absolute_cinema/internal/tmdb"
)

//go:embed web
var webFS embed.FS

// ScanState est l'état de progression du scan en cours.
type ScanState struct {
	Scanning bool   `json:"scanning"`
	Phase    string `json:"phase"` // walk | tmdb | series | done | error
	Done     int    `json:"done"`
	Total    int    `json:"total"`
	Current  string `json:"current"`
	Error    string `json:"error,omitempty"`
}

// Server porte l'état partagé de l'application.
type Server struct {
	Paths config.Paths

	mu        sync.Mutex
	lib       *library.Library
	cfg       *config.Config
	overrides map[string]catalog.Override
	scan      ScanState
}

// New charge l'état depuis le disque.
func New(paths config.Paths) (*Server, error) {
	seedFromDisk(paths)
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil, fmt.Errorf("config.json : %w", err)
	}
	lib, err := library.Load(paths.LibraryFile())
	if err != nil {
		return nil, fmt.Errorf("library.json : %w", err)
	}
	ov, err := config.LoadOverrides(paths.OverridesFile())
	if err != nil {
		return nil, fmt.Errorf("overrides.json : %w", err)
	}
	s := &Server{Paths: paths, lib: lib, cfg: cfg, overrides: ov}
	s.applyManualRatings()
	return s, nil
}

// seedFromDisk amorce le cache local depuis un catalogue pré-scanné livré sur
// le disque, quand celui-ci est en lecture seule (les données sont alors
// stockées dans le cache local, où le library.json du disque ne serait pas lu).
// Copie une seule fois : ne fait rien si le cache local a déjà un catalogue.
func seedFromDisk(paths config.Paths) {
	if !paths.ReadOnly {
		return // le disque est inscriptible : library.json y est lu directement
	}
	diskData := filepath.Join(paths.MediaRoot, config.DataDirName)
	diskLib := filepath.Join(diskData, "library.json")
	if _, err := os.Stat(diskLib); err != nil {
		return // aucun catalogue pré-scanné à amorcer
	}
	if _, err := os.Stat(paths.LibraryFile()); err == nil {
		return // cache local déjà peuplé
	}
	if err := fsutil.CopyFile(diskLib, paths.LibraryFile()); err != nil {
		log.Printf("amorçage library.json : %v", err)
		return
	}
	// médias en cache (affiches, fonds, réponses TMDB) pour l'affichage hors ligne.
	// On ne copie PAS config.json : il contient la clé TMDB du vendeur, qui ne
	// doit pas être transmise (CGU TMDB + fuite de clé).
	for _, sub := range []string{"posters", "backdrops", "tmdb_cache"} {
		if _, err := os.Stat(filepath.Join(diskData, sub)); err == nil {
			_ = fsutil.CopyTree(filepath.Join(diskData, sub), filepath.Join(paths.DataDir, sub))
		}
	}
}

// HasLibrary indique si un scan a déjà été fait.
func (s *Server) HasLibrary() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.lib.Movies) > 0 || len(s.lib.Series) > 0
}

// HasKey indique si une clé TMDB est configurée.
func (s *Server) HasKey() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg.TmdbAPIKey != ""
}

// StartScan lance un scan en arrière-plan (un seul à la fois).
func (s *Server) StartScan() bool {
	s.mu.Lock()
	if s.scan.Scanning {
		s.mu.Unlock()
		return false
	}
	s.scan = ScanState{Scanning: true, Phase: "walk"}
	apiKey := s.cfg.TmdbAPIKey
	metaLang := s.cfg.MetadataLang
	overrides := make(map[string]catalog.Override, len(s.overrides))
	for k, v := range s.overrides {
		overrides[k] = v
	}
	s.mu.Unlock()

	go func() {
		client := tmdb.New(apiKey, s.Paths.TmdbCacheDir())
		if metaLang != "" {
			client.Lang = metaLang
		}
		lib, err := catalog.Build(s.Paths.MediaRoot, s.Paths.DataDir, client, overrides, func(phase string, done, total int, current string) {
			s.mu.Lock()
			s.scan.Phase, s.scan.Done, s.scan.Total, s.scan.Current = phase, done, total, current
			s.mu.Unlock()
		})
		s.mu.Lock()
		defer s.mu.Unlock()
		if err != nil {
			s.scan = ScanState{Phase: "error", Error: err.Error()}
			log.Printf("scan : %v", err)
			return
		}
		// garde-fou hors ligne : si beaucoup de titres auparavant reconnus
		// basculent en « non identifiés » (cache TMDB incomplet + pas de
		// réseau), on conserve l'ancienne bibliothèque plutôt que de la dégrader
		oldMatched := len(s.lib.Movies) + len(s.lib.Series)
		newMatched := len(lib.Movies) + len(lib.Series)
		if oldMatched > 10 && newMatched < oldMatched*8/10 &&
			len(lib.Unmatched) > len(s.lib.Unmatched)+(oldMatched-newMatched)/2 {
			s.scan = ScanState{Phase: "done", Error: "scan partiel (réseau indisponible ?) — bibliothèque précédente conservée"}
			log.Printf("scan dégradé ignoré : %d reconnus contre %d avant", newMatched, oldMatched)
			return
		}
		s.lib = lib
		s.applyManualRatings()
		s.scan = ScanState{Phase: "done"}
		if err := library.Save(lib, s.Paths.LibraryFile()); err != nil {
			s.scan.Error = "sauvegarde library.json : " + err.Error()
			log.Printf("sauvegarde : %v", err)
		}
	}()
	return true
}

// Listen ouvre le premier port libre à partir de preferred (8484 par défaut).
// En mode LAN, écoute sur toutes les interfaces (TV, tablettes du foyer).
func Listen(preferred int, lan bool) (net.Listener, int, error) {
	if preferred == 0 {
		preferred = 8484
	}
	host := "127.0.0.1"
	if lan {
		host = "0.0.0.0"
	}
	for p := preferred; p < preferred+20; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, p))
		if err == nil {
			return ln, p, nil
		}
	}
	return nil, 0, fmt.Errorf("aucun port libre entre %d et %d", preferred, preferred+19)
}

// LanMode indique si l'écoute réseau local est activée dans la config.
func (s *Server) LanMode() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg.LanMode
}

// Handler construit le routeur HTTP complet.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	webRoot, _ := fs.Sub(webFS, "web")
	mux.Handle("GET /", http.FileServerFS(webRoot))
	mux.Handle("GET /media-cache/", http.StripPrefix("/media-cache/", http.FileServer(http.Dir(s.Paths.DataDir))))

	mux.HandleFunc("GET /api/library", s.handleLibrary)
	mux.HandleFunc("GET /api/capabilities", s.handleCapabilities)
	mux.HandleFunc("POST /api/scan", s.handleScanStart)
	mux.HandleFunc("GET /api/scan/status", s.handleScanStatus)
	mux.HandleFunc("POST /api/play", s.handlePlay)
	mux.HandleFunc("GET /api/duplicates", s.handleDuplicates)
	mux.HandleFunc("POST /api/trash", s.handleTrash)
	mux.HandleFunc("GET /api/trash", s.handleTrashList)
	mux.HandleFunc("POST /api/trash/restore", s.handleTrashRestore)
	mux.HandleFunc("POST /api/trash/empty", s.handleTrashEmpty)
	mux.HandleFunc("GET /api/config", s.handleConfigGet)
	mux.HandleFunc("POST /api/config", s.handleConfigSet)
	mux.HandleFunc("POST /api/override", s.handleOverride)
	mux.HandleFunc("GET /api/tmdb/search", s.handleTmdbSearch)
	mux.HandleFunc("POST /api/kidmode", s.handleKidMode)
	mux.HandleFunc("POST /api/rating", s.handleRating)
	mux.HandleFunc("GET /media/stream", s.handleStream)
	mux.HandleFunc("GET /media/subtitle", s.handleSubtitle)
	mux.HandleFunc("GET /api/progress", s.handleProgressGet)
	mux.HandleFunc("POST /api/progress", s.handleProgressSet)

	return mux
}
