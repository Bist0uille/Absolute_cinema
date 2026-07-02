package library

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Load lit library.json ; retourne une bibliothèque vide si le fichier n'existe pas.
func Load(path string) (*Library, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Library{SchemaVersion: 1}, nil
	}
	if err != nil {
		return nil, err
	}
	var lib Library
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, err
	}
	return &lib, nil
}

// Save écrit library.json de façon atomique (tmp + rename).
func Save(lib *Library, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(lib, "", " ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
