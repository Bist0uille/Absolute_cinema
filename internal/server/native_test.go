package server

import (
	"os"
	"path/filepath"
	"testing"

	"absolute_cinema/internal/config"
	"absolute_cinema/internal/library"
)

func TestNativePlaybackOnlyResolvesCatalogueFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"movie.mkv", "second.mkv", "sub.srt", "private.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte{0x1a, 0x45, 0xdf, 0xa3}, 0600); err != nil {
			t.Fatal(err)
		}
	}
	s := &Server{
		Paths: config.Paths{MediaRoot: root, DataDir: root}, cfg: &config.Config{},
		lib: &library.Library{Movies: []*library.Movie{{AgeRating: 18, Versions: []*library.Version{{
			Path: "movie.mkv", Parts: []string{"second.mkv"}, ExternalSubs: []library.ExternalSub{{Path: "sub.srt"}},
		}}}}},
	}
	media, err := s.ResolveNativePlayback("movie.mkv", "sub.srt")
	if err != nil {
		t.Fatal(err)
	}
	if len(media.Paths) != 2 || media.Paths[1] != filepath.Join(root, "second.mkv") || media.Subtitle != filepath.Join(root, "sub.srt") {
		t.Fatalf("bad resolved files: %#v", media)
	}
	for _, input := range [][2]string{{"private.txt", ""}, {"../movie.mkv", ""}, {"movie.mkv", "private.txt"}} {
		if _, err := s.ResolveNativePlayback(input[0], input[1]); err == nil {
			t.Errorf("allowed %#v", input)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "movie.mkv"), make([]byte, 4096), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ResolveNativePlayback("movie.mkv", ""); err == nil {
		t.Fatal("accepted a preallocated zero-filled MKV")
	}
	s.cfg.KidMode, s.cfg.KidMaxAge = true, 10
	if _, err := s.ResolveNativePlayback("movie.mkv", ""); err == nil {
		t.Fatal("native player bypassed kid mode")
	}
}

func TestMediaPathsAtFilesystemRoot(t *testing.T) {
	root := filepath.VolumeName(t.TempDir()) + string(filepath.Separator)
	s := &Server{Paths: config.Paths{MediaRoot: root}}
	got, err := s.absMediaPath("Films/movie.mkv")
	if err != nil || got != filepath.Join(root, "Films", "movie.mkv") {
		t.Fatalf("root path rejected: %q %v", got, err)
	}
	if _, err := s.absMediaPath(filepath.Join(root, "movie.mkv")); err == nil {
		t.Fatal("accepted an absolute media path")
	}
}

func TestNativeResumeRetainsMultipartPosition(t *testing.T) {
	dir := t.TempDir()
	s := &Server{Paths: config.Paths{DataDir: dir}}
	if err := s.RecordNativeProgress("movie.mkv", 125, 1000, 1, false); err != nil {
		t.Fatal(err)
	}
	entry := s.loadProgress()["movie.mkv"]
	if entry.Part != 1 || entry.Position != 125 || entry.Watched {
		t.Fatalf("bad resume entry: %#v", entry)
	}
	if err := s.RecordNativeProgress("movie.mkv", 1000, 1000, 1, true); err != nil {
		t.Fatal(err)
	}
	if !s.loadProgress()["movie.mkv"].Watched {
		t.Fatal("completion not saved")
	}
}
