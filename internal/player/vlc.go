// Package player lance la lecture d'un fichier vidéo dans VLC (ou à défaut
// le lecteur par défaut du système). VLC est cherché d'abord dans sa version
// portable livrée sur le disque, puis parmi les installations locales.
package player

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"absolute_cinema/internal/config"
	"absolute_cinema/internal/fsutil"
)

// Result indique comment la lecture a été lancée.
type Result struct {
	Player  string `json:"player"` // "vlc" | "system"
	Warning string `json:"warning,omitempty"`
}

// Play lance la lecture de absPath (+ parties supplémentaires éventuelles),
// avec un sous-titre optionnel. mediaRoot sert à localiser le VLC embarqué.
func Play(mediaRoot, absPath string, extraParts []string, subPath string) (*Result, error) {
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("fichier introuvable : %s", absPath)
	}
	// mode démo (enregistrement de captures) : ne lance pas de lecteur
	if os.Getenv("AC_DEMO") == "1" {
		return &Result{Player: "vlc"}, nil
	}
	if vlc := findVLC(mediaRoot); vlc != "" {
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

// VLCKind indique quel VLC serait utilisé pour la lecture :
// "embedded" (livré sur le disque), "installed" (présent sur la machine),
// ou "none". Sert à l'UI pour masquer l'invitation à installer VLC.
func VLCKind(mediaRoot string) string {
	if embeddedVLC(mediaRoot) != "" {
		return "embedded"
	}
	if installedVLC() != "" {
		return "installed"
	}
	return "none"
}

func findVLC(mediaRoot string) string {
	if p := embeddedVLC(mediaRoot); p != "" {
		return p
	}
	return installedVLC()
}

// embeddedVLC retourne le chemin de l'exécutable VLC portable livré sur le
// disque, ou "" s'il est absent/inutilisable. Sur Mac, si l'exécutable ne peut
// être lancé en place (bit exécutable absent sur certains montages en lecture
// seule), VLC.app est copié une seule fois dans ~/Library/Application Support.
func embeddedVLC(mediaRoot string) string {
	if mediaRoot == "" {
		return ""
	}
	tools := filepath.Join(mediaRoot, config.DataDirName, "tools")
	switch runtime.GOOS {
	case "windows":
		p := filepath.Join(tools, "vlc-win", "vlc.exe")
		if fileExists(p) {
			return p
		}
	case "darwin":
		app := filepath.Join(tools, "vlc-mac", "VLC.app")
		bin := filepath.Join(app, "Contents", "MacOS", "VLC")
		if isExecutable(bin) {
			return bin
		}
		if dirExists(app) {
			if local := ensureMacVLCCopy(app); local != "" {
				return local
			}
		}
	}
	return ""
}

func installedVLC() string {
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
		if fileExists(c) {
			return c
		}
	}
	return ""
}

// ensureMacVLCCopy copie VLC.app dans le dossier de support de l'application
// (une seule fois) et retourne le chemin de l'exécutable copié, ou "".
func ensureMacVLCCopy(srcApp string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dstApp := filepath.Join(home, "Library", "Application Support", "AbsoluteCinema", "VLC.app")
	bin := filepath.Join(dstApp, "Contents", "MacOS", "VLC")
	if isExecutable(bin) {
		return bin // déjà copié
	}
	if err := fsutil.CopyTree(srcApp, dstApp); err != nil {
		return ""
	}
	if isExecutable(bin) {
		return bin
	}
	return ""
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func isExecutable(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir() && fi.Mode()&0o111 != 0
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
