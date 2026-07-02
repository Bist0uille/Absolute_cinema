package server

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"absolute_cinema/internal/config"
	"absolute_cinema/internal/library"
)

// --- PIN ---

func hashPIN(salt, pin string) string {
	h := sha256.Sum256([]byte(salt + ":" + pin))
	return hex.EncodeToString(h[:])
}

func (s *Server) checkPIN(pin string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.KidPINHash == "" {
		return false
	}
	got := hashPIN(s.cfg.KidPINSalt, pin)
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.cfg.KidPINHash)) == 1
}

// --- Classifications manuelles (ratings.json : "movie:238" -> 0) ---

func (s *Server) ratingsFile() string {
	return filepath.Join(s.Paths.DataDir, "ratings.json")
}

func (s *Server) loadRatings() map[string]int {
	out := map[string]int{}
	if data, err := os.ReadFile(s.ratingsFile()); err == nil {
		_ = json.Unmarshal(data, &out)
	}
	return out
}

func (s *Server) saveRatings(m map[string]int) error {
	data, _ := json.MarshalIndent(m, "", " ")
	tmp := s.ratingsFile() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.ratingsFile())
}

// applyManualRatings écrase les classifications TMDB par celles définies à la
// main. À appeler (sous mu) après chargement, scan ou modification.
func (s *Server) applyManualRatings() {
	ratings := s.loadRatings()
	if len(ratings) == 0 {
		return
	}
	for _, m := range s.lib.Movies {
		if age, ok := ratings["movie:"+itoa(m.TmdbID)]; ok {
			m.AgeRating = age
		}
	}
	for _, se := range s.lib.Series {
		if age, ok := ratings["tv:"+itoa(se.TmdbID)]; ok {
			se.AgeRating = age
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// --- Filtrage ---

func kidAllowedMovie(m *library.Movie, maxAge int) bool {
	return m.AgeRating >= 0 && m.AgeRating <= maxAge
}

func kidAllowedSeries(se *library.Series, maxAge int) bool {
	return se.AgeRating >= 0 && se.AgeRating <= maxAge
}

// kidLibrary retourne une copie filtrée de la bibliothèque (mu déjà pris).
func (s *Server) kidLibrary() *library.Library {
	maxAge := s.cfg.KidMaxAge
	out := &library.Library{SchemaVersion: s.lib.SchemaVersion, ScannedAt: s.lib.ScannedAt}
	for _, m := range s.lib.Movies {
		if kidAllowedMovie(m, maxAge) {
			out.Movies = append(out.Movies, m)
		}
	}
	for _, se := range s.lib.Series {
		if kidAllowedSeries(se, maxAge) {
			out.Series = append(out.Series, se)
		}
	}
	return out
}

// kidAllowedPath vérifie qu'un chemin appartient à un titre autorisé (mu pris).
func (s *Server) kidAllowedPath(rel string) bool {
	maxAge := s.cfg.KidMaxAge
	for _, m := range s.lib.Movies {
		if !kidAllowedMovie(m, maxAge) {
			continue
		}
		for _, v := range m.Versions {
			if v.Path == rel {
				return true
			}
			for _, p := range v.Parts {
				if p == rel {
					return true
				}
			}
			for _, sub := range v.ExternalSubs {
				if sub.Path == rel {
					return true
				}
			}
		}
	}
	for _, se := range s.lib.Series {
		if !kidAllowedSeries(se, maxAge) {
			continue
		}
		for _, season := range se.Seasons {
			for _, ep := range season.Episodes {
				for _, f := range ep.Files {
					if f.Path == rel {
						return true
					}
					for _, sub := range f.ExternalSubs {
						if sub.Path == rel {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// requireParent renvoie 403 si le mode enfant est actif.
func (s *Server) requireParent(w http.ResponseWriter) bool {
	s.mu.Lock()
	kid := s.cfg.KidMode
	s.mu.Unlock()
	if kid {
		writeErr(w, 403, "mode enfant actif : action verrouillée (code PIN requis)")
		return false
	}
	return true
}

// --- Endpoints ---

type kidModeRequest struct {
	Enable bool   `json:"enable"`
	PIN    string `json:"pin"`      // pour désactiver, ou définir au 1er usage
	MaxAge *int   `json:"max_age"`  // optionnel à l'activation
}

func (s *Server) handleKidMode(w http.ResponseWriter, r *http.Request) {
	var req kidModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	if req.Enable {
		s.mu.Lock()
		if s.cfg.KidPINHash == "" {
			pin := strings.TrimSpace(req.PIN)
			if len(pin) < 4 {
				s.mu.Unlock()
				writeErr(w, 400, "définissez d'abord un code PIN d'au moins 4 chiffres")
				return
			}
			salt := make([]byte, 8)
			_, _ = rand.Read(salt)
			s.cfg.KidPINSalt = hex.EncodeToString(salt)
			s.cfg.KidPINHash = hashPIN(s.cfg.KidPINSalt, pin)
		}
		if req.MaxAge != nil {
			s.cfg.KidMaxAge = *req.MaxAge
		}
		if s.cfg.KidMaxAge == 0 {
			s.cfg.KidMaxAge = 10
		}
		s.cfg.KidMode = true
		err := config.Save(s.cfg, s.Paths.ConfigFile())
		s.mu.Unlock()
		if err != nil {
			writeErr(w, 500, "sauvegarde impossible : "+err.Error())
			return
		}
		writeJSON(w, 200, map[string]bool{"kid_mode": true})
		return
	}
	// désactivation : PIN obligatoire
	if !s.checkPIN(req.PIN) {
		writeErr(w, 403, "code PIN incorrect")
		return
	}
	s.mu.Lock()
	s.cfg.KidMode = false
	err := config.Save(s.cfg, s.Paths.ConfigFile())
	s.mu.Unlock()
	if err != nil {
		writeErr(w, 500, "sauvegarde impossible : "+err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"kid_mode": false})
}

type ratingRequest struct {
	MediaType string `json:"media_type"` // movie | tv
	TmdbID    int    `json:"tmdb_id"`
	Age       int    `json:"age"` // 0/10/12/16/18, -1 pour effacer
}

func (s *Server) handleRating(w http.ResponseWriter, r *http.Request) {
	if !s.requireParent(w) {
		return
	}
	var req ratingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		req.TmdbID == 0 || (req.MediaType != "movie" && req.MediaType != "tv") {
		writeErr(w, 400, "requête invalide (media_type, tmdb_id, age)")
		return
	}
	key := req.MediaType + ":" + itoa(req.TmdbID)
	ratings := s.loadRatings()
	if req.Age < 0 {
		delete(ratings, key)
	} else {
		ratings[key] = req.Age
	}
	if err := s.saveRatings(ratings); err != nil {
		writeErr(w, 500, "sauvegarde impossible : "+err.Error())
		return
	}
	s.mu.Lock()
	// re-appliquer : d'abord restaurer depuis library.json n'est pas nécessaire,
	// on écrase simplement la valeur ciblée (l'effacement reprendra au prochain scan)
	for _, m := range s.lib.Movies {
		if req.MediaType == "movie" && m.TmdbID == req.TmdbID {
			if req.Age >= 0 {
				m.AgeRating = req.Age
			}
		}
	}
	for _, se := range s.lib.Series {
		if req.MediaType == "tv" && se.TmdbID == req.TmdbID {
			if req.Age >= 0 {
				se.AgeRating = req.Age
			}
		}
	}
	s.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}
