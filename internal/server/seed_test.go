package server

import (
	"os"
	"path/filepath"
	"testing"

	"absolute_cinema/internal/config"
)

// TestSeedFromDisk vérifie qu'un catalogue pré-scanné livré sur un disque en
// lecture seule est amorcé dans le cache local (là où library.Load le lira).
func TestSeedFromDisk(t *testing.T) {
	disk := t.TempDir()
	cache := t.TempDir()

	diskData := filepath.Join(disk, config.DataDirName)
	if err := os.MkdirAll(filepath.Join(diskData, "posters"), 0o755); err != nil {
		t.Fatal(err)
	}
	libJSON := `{"schema_version":1,"movies":[{"id":"mv_x","title_fr":"Amélie","year":2001}]}`
	if err := os.WriteFile(filepath.Join(diskData, "library.json"), []byte(libJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(diskData, "posters", "mv_x.jpg"), []byte("img"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths := config.Paths{MediaRoot: disk, DataDir: cache, ReadOnly: true}
	seedFromDisk(paths)

	if _, err := os.Stat(paths.LibraryFile()); err != nil {
		t.Fatalf("library.json non amorcé dans le cache : %v", err)
	}
	if _, err := os.Stat(filepath.Join(cache, "posters", "mv_x.jpg")); err != nil {
		t.Errorf("affiches non copiées : %v", err)
	}

	// le catalogue amorcé doit se charger correctement
	srv, err := New(paths)
	if err != nil {
		t.Fatalf("New : %v", err)
	}
	if !srv.HasLibrary() {
		t.Error("HasLibrary faux : le catalogue amorcé n'a pas été chargé")
	}

	// idempotence : un cache déjà peuplé n'est pas ré-amorcé (pas d'erreur)
	seedFromDisk(paths)
}

// TestSeedFromDiskWritable : disque inscriptible → pas d'amorçage (library.json
// du disque est lu directement, DataDir n'est pas redirigé).
func TestSeedFromDiskWritable(t *testing.T) {
	disk := t.TempDir()
	diskData := filepath.Join(disk, config.DataDirName)
	if err := os.MkdirAll(diskData, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(diskData, "library.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// DataDir == disque, ReadOnly false : seedFromDisk ne doit rien faire
	paths := config.Paths{MediaRoot: disk, DataDir: diskData, ReadOnly: false}
	seedFromDisk(paths) // ne doit pas paniquer ni dupliquer
}
