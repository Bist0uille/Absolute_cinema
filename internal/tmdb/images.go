package tmdb

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const imageBase = "https://image.tmdb.org/t/p/"

// DownloadImage télécharge une image TMDB (poster_path/backdrop_path) vers
// destDir/name.jpg si elle n'y est pas déjà. Retourne le chemin relatif
// "destSubdir/name.jpg" ou "" si tmdbPath est vide.
func DownloadImage(tmdbPath, size, destDir, destSubdir, name string) (string, error) {
	if tmdbPath == "" {
		return "", nil
	}
	rel := destSubdir + "/" + name + ".jpg"
	dest := filepath.Join(destDir, destSubdir, name+".jpg")
	if st, err := os.Stat(dest); err == nil && st.Size() > 0 {
		return rel, nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(imageBase + size + tmdbPath)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("image TMDB %d", resp.StatusCode)
	}
	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return "", err
	}
	f.Close()
	return rel, os.Rename(tmp, dest)
}
