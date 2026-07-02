package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// --- Diffusion des vidéos vers le lecteur intégré (avec support Range) ---

var contentTypes = map[string]string{
	".mp4": "video/mp4", ".m4v": "video/mp4", ".webm": "video/webm",
	".mkv": "video/x-matroska", ".avi": "video/x-msvideo", ".mov": "video/quicktime",
	".mpg": "video/mpeg", ".mpeg": "video/mpeg", ".ts": "video/mp2t",
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	s.mu.Lock()
	if s.cfg.KidMode && !s.kidAllowedPath(rel) {
		s.mu.Unlock()
		writeErr(w, 403, "titre non autorisé en mode enfant")
		return
	}
	s.mu.Unlock()
	abs, err := s.absMediaPath(rel)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	f, err := os.Open(abs)
	if err != nil {
		writeErr(w, 404, "fichier introuvable")
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if ct := contentTypes[strings.ToLower(filepath.Ext(abs))]; ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	// ServeContent gère les requêtes Range (seek dans le lecteur)
	http.ServeContent(w, r, filepath.Base(abs), st.ModTime(), f)
}

// --- Sous-titres : conversion SRT -> WebVTT à la volée ---

var reSrtTime = regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})`)

func (s *Server) handleSubtitle(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	s.mu.Lock()
	if s.cfg.KidMode && !s.kidAllowedPath(rel) {
		s.mu.Unlock()
		writeErr(w, 403, "non autorisé")
		return
	}
	s.mu.Unlock()
	abs, err := s.absMediaPath(rel)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		writeErr(w, 404, "sous-titre introuvable")
		return
	}
	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")

	ext := strings.ToLower(filepath.Ext(abs))
	if ext == ".vtt" {
		w.Write(data)
		return
	}
	// SRT -> VTT : BOM retiré, virgules des timestamps -> points, en-tête WEBVTT
	text := strings.TrimPrefix(string(fixEncoding(data)), "\uFEFF")
	text = reSrtTime.ReplaceAllString(text, "$1.$2")
	w.Write([]byte("WEBVTT\n\n" + text))
}

// fixEncoding convertit les .srt Latin-1/Windows-1252 (fréquents) en UTF-8.
func fixEncoding(b []byte) []byte {
	// déjà de l'UTF-8 valide ? on ne touche à rien
	if isValidUTF8(b) {
		return b
	}
	out := make([]byte, 0, len(b)*2)
	for _, c := range b {
		if c < 0x80 {
			out = append(out, c)
		} else {
			// Latin-1 -> UTF-8
			out = append(out, 0xC0|c>>6, 0x80|c&0x3F)
		}
	}
	return out
}

func isValidUTF8(b []byte) bool {
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c < 0x80:
			i++
		case c>>5 == 0x6:
			if i+1 >= len(b) || b[i+1]>>6 != 0x2 {
				return false
			}
			i += 2
		case c>>4 == 0xE:
			if i+2 >= len(b) || b[i+1]>>6 != 0x2 || b[i+2]>>6 != 0x2 {
				return false
			}
			i += 3
		case c>>3 == 0x1E:
			if i+3 >= len(b) || b[i+1]>>6 != 0x2 || b[i+2]>>6 != 0x2 || b[i+3]>>6 != 0x2 {
				return false
			}
			i += 4
		default:
			return false
		}
	}
	return true
}

// --- Progression de lecture (reprise « à la Netflix ») ---

// ProgressEntry est l'état de visionnage d'un fichier.
type ProgressEntry struct {
	Position  float64   `json:"position"` // secondes
	Duration  float64   `json:"duration"`
	Watched   bool      `json:"watched"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Server) progressFile() string {
	return filepath.Join(s.Paths.DataDir, "progress.json")
}

func (s *Server) loadProgress() map[string]ProgressEntry {
	out := map[string]ProgressEntry{}
	if data, err := os.ReadFile(s.progressFile()); err == nil {
		_ = json.Unmarshal(data, &out)
	}
	return out
}

func (s *Server) saveProgress(m map[string]ProgressEntry) error {
	data, _ := json.MarshalIndent(m, "", " ")
	tmp := s.progressFile() + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(f)
	bw.Write(data)
	bw.Flush()
	f.Close()
	return os.Rename(tmp, s.progressFile())
}

func (s *Server) handleProgressGet(w http.ResponseWriter, r *http.Request) {
	prog := s.loadProgress()
	s.mu.Lock()
	kid := s.cfg.KidMode
	s.mu.Unlock()
	if kid {
		// ne divulguer que les chemins autorisés
		s.mu.Lock()
		filtered := map[string]ProgressEntry{}
		for p, e := range prog {
			if s.kidAllowedPath(p) {
				filtered[p] = e
			}
		}
		s.mu.Unlock()
		prog = filtered
	}
	writeJSON(w, 200, prog)
}

type progressRequest struct {
	Path     string  `json:"path"`
	Position float64 `json:"position"`
	Duration float64 `json:"duration"`
	Watched  *bool   `json:"watched"`
}

func (s *Server) handleProgressSet(w http.ResponseWriter, r *http.Request) {
	var req progressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		writeErr(w, 400, "requête invalide")
		return
	}
	if _, err := s.absMediaPath(req.Path); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	prog := s.loadProgress()
	e := prog[req.Path]
	if req.Position >= 0 {
		e.Position = req.Position
	}
	if req.Duration > 0 {
		e.Duration = req.Duration
	}
	if req.Watched != nil {
		e.Watched = *req.Watched
	} else if e.Duration > 0 && e.Position/e.Duration > 0.92 {
		e.Watched = true
	}
	e.UpdatedAt = time.Now().UTC()
	prog[req.Path] = e
	if err := s.saveProgress(prog); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
