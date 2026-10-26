package server

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"absolute_cinema/internal/library"
)

// NativePlayback contains only files belonging to a catalogue entry. Native
// bindings use the same root and kid-mode checks as the HTTP player.
type NativePlayback struct {
	Paths    []string
	Subtitle string
}

// RecordNativeProgress is also called before the native video window closes.
func (s *Server) RecordNativeProgress(path string, position, duration float64, part int, watched bool) error {
	if position < 0 || duration <= 0 {
		return nil
	}
	s.progressMu.Lock()
	defer s.progressMu.Unlock()
	progress := s.loadProgress()
	entry := progress[path]
	entry.Position, entry.Duration, entry.Part = position, duration, part
	entry.Watched = entry.Watched || watched
	entry.UpdatedAt = time.Now().UTC()
	progress[path] = entry
	return s.saveProgress(progress)
}

func (s *Server) ResolveNativePlayback(path, subtitle string) (NativePlayback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.KidMode && !s.kidAllowedPath(path) {
		return NativePlayback{}, fmt.Errorf("titre non autorisé en mode enfant")
	}
	var version *library.Version
	for _, movie := range s.lib.Movies {
		for _, candidate := range movie.Versions {
			if candidate.Path == path {
				version = candidate
			}
		}
	}
	for _, series := range s.lib.Series {
		for _, season := range series.Seasons {
			for _, episode := range season.Episodes {
				for _, candidate := range episode.Files {
					if candidate.Path == path {
						version = candidate
					}
				}
			}
		}
	}
	if version == nil {
		return NativePlayback{}, fmt.Errorf("fichier absent du catalogue")
	}
	var result NativePlayback
	for _, relative := range append([]string{version.Path}, version.Parts...) {
		absolute, err := s.absMediaPath(relative)
		if err != nil {
			return NativePlayback{}, err
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return NativePlayback{}, err
		}
		if info.IsDir() {
			return NativePlayback{}, fmt.Errorf("le média est un dossier")
		}
		// Matroska/WebM start with an EBML header. Incomplete downloads can
		// have the expected size while containing only preallocated zeros.
		ext := strings.ToLower(filepath.Ext(absolute))
		if ext == ".mkv" || ext == ".webm" {
			file, err := os.Open(absolute)
			if err != nil {
				return NativePlayback{}, err
			}
			var header [4]byte
			_, err = io.ReadFull(file, header[:])
			file.Close()
			if err != nil || !bytes.Equal(header[:], []byte{0x1a, 0x45, 0xdf, 0xa3}) {
				return NativePlayback{}, fmt.Errorf("fichier vidéo endommagé ou incomplet : %s", filepath.Base(absolute))
			}
		}
		result.Paths = append(result.Paths, absolute)
	}
	if subtitle != "" {
		allowed := false
		for _, sub := range version.ExternalSubs {
			if sub.Path == subtitle {
				allowed = true
			}
		}
		if !allowed {
			return NativePlayback{}, fmt.Errorf("sous-titre absent de cette version")
		}
		var err error
		result.Subtitle, err = s.absMediaPath(subtitle)
		if err != nil {
			return NativePlayback{}, err
		}
	}
	return result, nil
}
