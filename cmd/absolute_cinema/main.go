// absolute_cinema : bibliothèque de films & séries portable.
// Le binaire vit à la racine du disque dur ; il scanne [Films]/[Séries],
// sert l'interface web sur localhost et lance la lecture dans VLC.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"absolute_cinema/internal/config"
	"absolute_cinema/internal/player"
	"absolute_cinema/internal/scanner"
	"absolute_cinema/internal/server"
)

func main() {
	rootFlag := flag.String("root", "", "racine de la médiathèque (défaut : dossier de l'exécutable)")
	portFlag := flag.Int("port", 0, "port HTTP (défaut : 8484, ou le suivant libre)")
	noBrowser := flag.Bool("no-browser", false, "ne pas ouvrir le navigateur au démarrage")
	flag.Parse()

	mediaRoot := *rootFlag
	if mediaRoot == "" {
		exe, err := os.Executable()
		if err != nil {
			log.Fatalf("impossible de localiser l'exécutable : %v", err)
		}
		mediaRoot = filepath.Dir(exe)
	}
	mediaRoot, err := filepath.Abs(mediaRoot)
	if err != nil {
		log.Fatalf("chemin invalide : %v", err)
	}

	films, series, err := scanner.DetectRoots(mediaRoot)
	if err != nil {
		log.Fatalf("lecture de %s : %v", mediaRoot, err)
	}
	if len(films) == 0 && len(series) == 0 {
		log.Fatalf("aucun dossier de films/séries trouvé dans %s\n"+
			"Placez le programme à la racine du disque (à côté de [Films] et [Séries]),\n"+
			"ou lancez-le avec -root \"D:\\Films & Séries\".", mediaRoot)
	}

	paths := config.ResolvePaths(mediaRoot)
	if paths.ReadOnly {
		fmt.Println("⚠ Disque en lecture seule : cache et corbeille désactivés sur le disque,")
		fmt.Println("  données stockées localement dans", paths.DataDir)
	}

	srv, err := server.New(paths)
	if err != nil {
		log.Fatalf("démarrage : %v", err)
	}

	ln, port, err := server.Listen(*portFlag)
	if err != nil {
		log.Fatalf("%v", err)
	}
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	// premier lancement avec clé déjà configurée : scan automatique
	if !srv.HasLibrary() && srv.HasKey() {
		srv.StartScan()
	}

	fmt.Println()
	fmt.Println("  🎬 absolute_cinema")
	fmt.Println("  Médiathèque :", mediaRoot)
	fmt.Println("  Interface   :", url)
	fmt.Println("  (Ctrl+C pour quitter)")
	fmt.Println()

	if !*noBrowser {
		player.OpenBrowser(url)
	}
	log.Fatal(http.Serve(ln, srv.Handler()))
}
