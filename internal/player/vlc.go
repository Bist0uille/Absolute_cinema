// Package player lance la lecture d'un fichier vidéo dans VLC (ou à défaut
// le lecteur par défaut du système).
package player

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Result indique comment la lecture a été lancée.
type Result struct {
	Player   string `json:"player"`   // "vlc" | "system"
	Warning  string `json:"warning,omitempty"`
}

// Play lance la lecture de absPath (+ parties supplémentaires éventuelles),
// avec un sous-titre optionnel.
func Play(absPath string, extraParts []string, subPath string) (*Result, error) {
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("fichier introuvable : %s", absPath)
	}
	// mode démo (enregistrement de captures) : ne lance pas de lecteur
	if os.Getenv("AC_DEMO") == "1" {
		return &Result{Player: "vlc"}, nil
	}
	if vlc := findVLC(); vlc != "" {
		args := []string{absPath}
		args = append(args, extraParts...)
		if subPath != "" {
			args = append(args, "--sub-file="+subPath)
		}
		cmd := exec.Command(vlc, args...)
		detach(cmd)
		if err := cmd.Start(); err == nil {
			go cmd.Wait()
			return &Result{Player: "vlc"}, nil
		}
	}
	// repli : lecteur par défaut du système
	if err := openWithSystem(absPath); err != nil {
		return nil, err
	}
	w := ""
	if subPath != "" {
		w = "VLC introuvable : lecture avec le lecteur par défaut, sous-titre non transmis."
	} else {
		w = "VLC introuvable : lecture avec le lecteur par défaut."
	}
	return &Result{Player: "system", Warning: w}, nil
}

func findVLC() string {
	if p, err := exec.LookPath("vlc"); err == nil {
		return p
	}
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "ProgramW6432"} {
			if base := os.Getenv(env); base != "" {
				candidates = append(candidates, filepath.Join(base, "VideoLAN", "VLC", "vlc.exe"))
			}
		}
		candidates = append(candidates,
			`C:\Program Files\VideoLAN\VLC\vlc.exe`,
			`C:\Program Files (x86)\VideoLAN\VLC\vlc.exe`,
		)
	case "darwin":
		home, _ := os.UserHomeDir()
		candidates = []string{
			"/Applications/VLC.app/Contents/MacOS/VLC",
			filepath.Join(home, "Applications", "VLC.app", "Contents", "MacOS", "VLC"),
		}
	default: // linux & co
		candidates = []string{
			"/usr/bin/vlc", "/usr/local/bin/vlc", "/snap/bin/vlc",
			"/var/lib/flatpak/exports/bin/org.videolan.VLC",
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func openWithSystem(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}

// OpenBrowser ouvre l'URL dans le navigateur par défaut.
func OpenBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	detach(cmd)
	if err := cmd.Start(); err == nil {
		go cmd.Wait()
	}
}
