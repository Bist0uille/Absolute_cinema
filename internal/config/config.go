// Package config gère config.json et overrides.json dans .absolute_cinema/.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"absolute_cinema/internal/catalog"
)

const DataDirName = ".absolute_cinema"

// Config est le contenu de config.json.
type Config struct {
	TmdbAPIKey   string `json:"tmdb_api_key"`
	Port         int    `json:"port"`          // 0 = auto (8484+)
	MetadataLang string `json:"metadata_lang"` // "fr-FR" (défaut) ou "en-US"
	UILang       string `json:"ui_lang"`       // "fr" (défaut) ou "en"
	LanMode      bool   `json:"lan_mode"`      // écouter sur le réseau local (TV, tablettes)

	// Mode enfant : bibliothèque filtrée par âge, verrouillée par code PIN.
	KidMode    bool   `json:"kid_mode"`
	KidMaxAge  int    `json:"kid_max_age"`  // âge max autorisé (défaut 10)
	KidPINHash string `json:"kid_pin_hash"` // sha256(salt+pin)
	KidPINSalt string `json:"kid_pin_salt"`
}

// Paths regroupe les chemins de données de l'application.
type Paths struct {
	MediaRoot string
	DataDir   string // .absolute_cinema (sur le disque, ou cache local si lecture seule)
	ReadOnly  bool   // le disque n'est pas inscriptible (NTFS sur Mac…)
}

// ResolvePaths choisit le dossier de données : sur le disque si possible,
// sinon dans le cache utilisateur local.
func ResolvePaths(mediaRoot string) Paths {
	onDisk := filepath.Join(mediaRoot, DataDirName)
	if writable(onDisk) {
		return Paths{MediaRoot: mediaRoot, DataDir: onDisk}
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = os.TempDir()
	}
	local := filepath.Join(cache, "absolute_cinema")
	_ = os.MkdirAll(local, 0o755)
	return Paths{MediaRoot: mediaRoot, DataDir: local, ReadOnly: true}
}

func writable(dir string) bool {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	probe := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return false
	}
	os.Remove(probe)
	return true
}

func (p Paths) ConfigFile() string    { return filepath.Join(p.DataDir, "config.json") }
func (p Paths) LibraryFile() string   { return filepath.Join(p.DataDir, "library.json") }
func (p Paths) OverridesFile() string { return filepath.Join(p.DataDir, "overrides.json") }
func (p Paths) TmdbCacheDir() string  { return filepath.Join(p.DataDir, "tmdb_cache") }
func (p Paths) TrashDir() string      { return filepath.Join(p.DataDir, "corbeille") }

// Load lit config.json (vide si absent).
func Load(path string) (*Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &c, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, json.Unmarshal(data, &c)
}

// Save écrit config.json.
func Save(c *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(c, "", " ")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadOverrides lit overrides.json (vide si absent).
func LoadOverrides(path string) (map[string]catalog.Override, error) {
	out := map[string]catalog.Override{}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	return out, json.Unmarshal(data, &out)
}

// SaveOverrides écrit overrides.json.
func SaveOverrides(m map[string]catalog.Override, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(m, "", " ")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
