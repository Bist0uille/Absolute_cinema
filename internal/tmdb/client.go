// Package tmdb est un client minimal de l'API The Movie Database, avec
// throttle, retry sur 429 et cache disque intégral (re-scan hors-ligne).
package tmdb

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.themoviedb.org/3"

// ErrNoKey est retourné quand aucune clé API n'est configurée et que la
// réponse n'est pas en cache.
var ErrNoKey = errors.New("clé API TMDB manquante")

// Client interroge TMDB. CacheDir est obligatoire ; APIKey peut être vide
// (mode hors-ligne : seules les réponses en cache sont servies).
type Client struct {
	APIKey   string
	CacheDir string
	Lang     string // langue des métadonnées ("fr-FR" par défaut)
	HTTP     *http.Client
	throttle <-chan time.Time
}

func New(apiKey, cacheDir string) *Client {
	return &Client{
		APIKey:   apiKey,
		CacheDir: cacheDir,
		Lang:     "fr-FR",
		HTTP:     &http.Client{Timeout: 30 * time.Second},
		throttle: time.Tick(250 * time.Millisecond), // ~4 req/s
	}
}

// HasKey indique qu'une clé API est configurée. Sinon, le catalogue est
// construit en « mode local » (titres depuis les noms de fichiers, sans
// affiches ni synopsis).
func (c *Client) HasKey() bool { return c.APIKey != "" }

// get effectue GET path?params, en passant par le cache disque.
// La clé de cache ne contient jamais la clé API.
func (c *Client) get(path string, params url.Values, out any) error {
	if params == nil {
		params = url.Values{}
	}
	cacheKey := path + "?" + params.Encode()
	h := sha1.Sum([]byte(cacheKey))
	cacheFile := filepath.Join(c.CacheDir, hex.EncodeToString(h[:])+".json")

	if data, err := os.ReadFile(cacheFile); err == nil {
		return json.Unmarshal(data, out)
	}
	if c.APIKey == "" {
		return ErrNoKey
	}

	params.Set("api_key", c.APIKey)
	reqURL := apiBase + path + "?" + params.Encode()

	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		<-c.throttle
		resp, err := c.HTTP.Get(reqURL)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(attempt+1))
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		switch {
		case resp.StatusCode == 200:
			_ = os.MkdirAll(c.CacheDir, 0o755)
			_ = os.WriteFile(cacheFile, body, 0o644)
			return json.Unmarshal(body, out)
		case resp.StatusCode == 429:
			wait := 2 * time.Second
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if s, err := strconv.Atoi(ra); err == nil {
					wait = time.Duration(s+1) * time.Second
				}
			}
			time.Sleep(wait)
			lastErr = fmt.Errorf("TMDB 429")
		case resp.StatusCode == 401:
			return fmt.Errorf("clé API TMDB invalide (401)")
		case resp.StatusCode == 404:
			return fmt.Errorf("TMDB 404: %s", path)
		default:
			lastErr = fmt.Errorf("TMDB %d: %s", resp.StatusCode, strings.TrimSpace(string(body[:min(len(body), 200)])))
			time.Sleep(time.Second * time.Duration(attempt+1))
		}
	}
	return lastErr
}

// ValidateKey vérifie la clé API auprès de TMDB (sans cache).
func ValidateKey(key string) error {
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get(apiBase + "/configuration?api_key=" + url.QueryEscape(key))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return fmt.Errorf("clé invalide")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("TMDB a répondu %d", resp.StatusCode)
	}
	return nil
}

// --- Types de réponses ---

type SearchResult struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`      // films
	Name          string  `json:"name"`       // séries
	OriginalTitle string  `json:"original_title"`
	OriginalName  string  `json:"original_name"`
	ReleaseDate   string  `json:"release_date"`   // films
	FirstAirDate  string  `json:"first_air_date"` // séries
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	MediaType     string  `json:"media_type"` // search/multi
}

func (r SearchResult) DisplayTitle() string {
	if r.Title != "" {
		return r.Title
	}
	return r.Name
}

func (r SearchResult) DisplayOriginal() string {
	if r.OriginalTitle != "" {
		return r.OriginalTitle
	}
	return r.OriginalName
}

func (r SearchResult) Year() int {
	d := r.ReleaseDate
	if d == "" {
		d = r.FirstAirDate
	}
	if len(d) >= 4 {
		y, _ := strconv.Atoi(d[:4])
		return y
	}
	return 0
}

type searchResponse struct {
	Results []SearchResult `json:"results"`
}

type Genre struct {
	Name string `json:"name"`
}

// MovieDetails est la réponse de /movie/{id} (+release_dates).
type MovieDetails struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	ReleaseDate   string  `json:"release_date"`
	Overview      string  `json:"overview"`
	Genres        []Genre `json:"genres"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	VoteAverage   float64 `json:"vote_average"`
	Runtime       int     `json:"runtime"`
	Adult         bool    `json:"adult"`
	ReleaseDates  struct {
		Results []struct {
			Iso      string `json:"iso_3166_1"`
			Releases []struct {
				Certification string `json:"certification"`
			} `json:"release_dates"`
		} `json:"results"`
	} `json:"release_dates"`
}

// AgeRating retourne l'âge minimal (0, 10, 12, 16, 18) ou -1 si inconnu.
func (d *MovieDetails) AgeRating() int {
	if d.Adult {
		return 18
	}
	certs := map[string]string{}
	for _, r := range d.ReleaseDates.Results {
		for _, rel := range r.Releases {
			if rel.Certification != "" && certs[r.Iso] == "" {
				certs[r.Iso] = rel.Certification
			}
		}
	}
	for _, country := range []string{"FR", "US", "GB", "DE"} {
		if c, ok := certs[country]; ok {
			if age := CertToAge(country, c); age >= 0 {
				return age
			}
		}
	}
	for country, c := range certs {
		if age := CertToAge(country, c); age >= 0 {
			return age
		}
	}
	return -1
}

// CertToAge convertit une certification nationale en âge minimal (-1 = inconnu).
func CertToAge(country, cert string) int {
	c := strings.ToUpper(strings.TrimSpace(cert))
	if c == "" {
		return -1
	}
	switch country {
	case "FR":
		switch c {
		case "U", "TP", "TOUS PUBLICS", "T":
			return 0
		case "10":
			return 10
		case "12":
			return 12
		case "16":
			return 16
		case "18", "X":
			return 18
		}
	case "US":
		switch c {
		case "G", "TV-Y", "TV-G":
			return 0
		case "PG", "TV-Y7", "TV-PG":
			return 10
		case "PG-13":
			return 12
		case "TV-14":
			return 14
		case "R", "TV-MA":
			return 16
		case "NC-17":
			return 18
		}
	case "GB":
		switch c {
		case "U":
			return 0
		case "PG":
			return 10
		case "12", "12A":
			return 12
		case "15":
			return 16
		case "18", "R18":
			return 18
		}
	case "DE":
		switch c {
		case "0":
			return 0
		case "6":
			return 10
		case "12":
			return 12
		case "16":
			return 16
		case "18":
			return 18
		}
	}
	// certification purement numérique (beaucoup de pays)
	if n, err := strconv.Atoi(c); err == nil && n >= 0 && n <= 21 {
		return n
	}
	return -1
}

// TVDetails est la réponse de /tv/{id} (+content_ratings).
type TVDetails struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	OriginalName string  `json:"original_name"`
	FirstAirDate string  `json:"first_air_date"`
	Overview     string  `json:"overview"`
	Genres       []Genre `json:"genres"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	VoteAverage  float64 `json:"vote_average"`
	Seasons      []struct {
		SeasonNumber int    `json:"season_number"`
		Name         string `json:"name"`
		PosterPath   string `json:"poster_path"`
	} `json:"seasons"`
	ContentRatings struct {
		Results []struct {
			Iso    string `json:"iso_3166_1"`
			Rating string `json:"rating"`
		} `json:"results"`
	} `json:"content_ratings"`
}

// AgeRating retourne l'âge minimal (0…18) ou -1 si inconnu.
func (d *TVDetails) AgeRating() int {
	certs := map[string]string{}
	for _, r := range d.ContentRatings.Results {
		if r.Rating != "" && certs[r.Iso] == "" {
			certs[r.Iso] = r.Rating
		}
	}
	for _, country := range []string{"FR", "US", "GB", "DE"} {
		if c, ok := certs[country]; ok {
			if age := CertToAge(country, c); age >= 0 {
				return age
			}
		}
	}
	for country, c := range certs {
		if age := CertToAge(country, c); age >= 0 {
			return age
		}
	}
	return -1
}

// SeasonDetails est la réponse de /tv/{id}/season/{n}.
type SeasonDetails struct {
	SeasonNumber int    `json:"season_number"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	Episodes     []struct {
		EpisodeNumber int    `json:"episode_number"`
		Name          string `json:"name"`
		Overview      string `json:"overview"`
		AirDate       string `json:"air_date"`
		StillPath     string `json:"still_path"`
	} `json:"episodes"`
}

// --- Appels ---

func (c *Client) SearchMovie(query string, year int) ([]SearchResult, error) {
	p := url.Values{"query": {query}, "language": {c.Lang}, "include_adult": {"true"}}
	if year > 0 {
		p.Set("year", strconv.Itoa(year))
	}
	var r searchResponse
	err := c.get("/search/movie", p, &r)
	return r.Results, err
}

func (c *Client) SearchTV(query string) ([]SearchResult, error) {
	p := url.Values{"query": {query}, "language": {c.Lang}, "include_adult": {"true"}}
	var r searchResponse
	err := c.get("/search/tv", p, &r)
	return r.Results, err
}

func (c *Client) SearchMulti(query string) ([]SearchResult, error) {
	p := url.Values{"query": {query}, "language": {c.Lang}, "include_adult": {"true"}}
	var r searchResponse
	err := c.get("/search/multi", p, &r)
	return r.Results, err
}

func (c *Client) MovieDetails(id int, lang string) (*MovieDetails, error) {
	var d MovieDetails
	err := c.get("/movie/"+strconv.Itoa(id), url.Values{"language": {lang}, "append_to_response": {"release_dates"}}, &d)
	return &d, err
}

func (c *Client) TVDetails(id int, lang string) (*TVDetails, error) {
	var d TVDetails
	err := c.get("/tv/"+strconv.Itoa(id), url.Values{"language": {lang}, "append_to_response": {"content_ratings"}}, &d)
	return &d, err
}

func (c *Client) SeasonDetails(tvID, season int) (*SeasonDetails, error) {
	var d SeasonDetails
	err := c.get(fmt.Sprintf("/tv/%d/season/%d", tvID, season), url.Values{"language": {c.Lang}}, &d)
	return &d, err
}
