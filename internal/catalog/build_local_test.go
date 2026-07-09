package catalog

import (
	"os"
	"path/filepath"
	"testing"

	"absolute_cinema/internal/tmdb"
)

// TestBuildLocalNoKey vérifie qu'un scan sans clé TMDB construit un catalogue
// à partir des seuls noms de fichiers, sans appel réseau ni titres « unmatched ».
func TestBuildLocalNoKey(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(rel string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		// >0 octet : les fichiers vides sont ignorés par le scanner
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("[Films]/Inception (2010).mkv")
	mustWrite("[Films]/The Matrix (1999).mp4")
	mustWrite("[Séries]/Breaking Bad/Breaking Bad S01E01.mkv")
	mustWrite("[Séries]/Breaking Bad/Breaking Bad S01E02.mkv")

	client := tmdb.New("", filepath.Join(root, "cache")) // pas de clé → mode local
	if client.HasKey() {
		t.Fatal("le client ne devrait pas avoir de clé")
	}

	lib, err := Build(root, filepath.Join(root, "data"), client, nil, nil)
	if err != nil {
		t.Fatalf("Build : %v", err)
	}

	if len(lib.Movies) != 2 {
		var titles []string
		for _, m := range lib.Movies {
			titles = append(titles, m.Title)
		}
		t.Errorf("films : %d, attendu 2 (%v)", len(lib.Movies), titles)
	}
	if len(lib.Series) != 1 {
		t.Errorf("séries : %d, attendu 1", len(lib.Series))
	}
	if len(lib.Unmatched) != 0 {
		t.Errorf("unmatched : %d, attendu 0 (le mode local n'en produit pas)", len(lib.Unmatched))
	}
	for _, m := range lib.Movies {
		if m.Poster != "" {
			t.Errorf("%q : poster non vide en mode local", m.Title)
		}
		if m.NeedsReview {
			t.Errorf("%q : ne devrait pas nécessiter de revue en mode local", m.Title)
		}
	}
	if len(lib.Series) == 1 {
		s := lib.Series[0]
		eps := 0
		for _, se := range s.Seasons {
			eps += len(se.Episodes)
		}
		if eps != 2 {
			t.Errorf("épisodes de %q : %d, attendu 2", s.Title, eps)
		}
	}
}
