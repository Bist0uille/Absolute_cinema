# Lecture native Windows

Sur Windows amd64, le lancement normal ouvre le catalogue dans une fenêtre
WebView2. La vidéo est décodée directement par LibVLC 3 dans un HWND enfant,
placé au-dessus de la zone vidéo. Les commandes HTML restent autour de cette
surface ; aucun élément `<video>` et aucun transcodage ne sont utilisés pour
la lecture native.

La barre de titre Windows utilise le mode sombre DWM et la palette du catalogue.
L'icône du ticket doré est intégrée à l'exécutable (tailles 16 à 256 pixels), à la
fenêtre, à la barre des tâches et au favicon. Le source vectoriel est dans
`docs/img/app-icon.svg` ; l'export Windows est `docs/img/app-icon.ico`.
Après modification de l'icône, copier cet ICO dans `internal/server/web/` et
régénérer la ressource avec :

```powershell
go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -ico docs/img/app-icon.ico -o cmd/absolute_cinema/rsrc_windows_amd64.syso
```

VLC est cherché dans `.absolute_cinema/tools/vlc-win`, puis parmi les
installations locales. La distribution complète est nécessaire : `libvlc.dll`,
`libvlccore.dll` et les plugins doivent rester ensemble. Le moteur est chargé
par chemin absolu avec une recherche des DLL limitée aux emplacements Windows
de confiance et au dossier du moteur.

Le chargement de LibVLC se fait en arrière-plan pour que la fenêtre reste
réactive, même sur un disque USB occupé. Le lecteur attend que le moteur soit
prêt avant de lancer le film ; le catalogue reste accessible pendant ce délai.

La fenêtre native fonctionne actuellement sur Windows 64 bits. Mac/Linux,
les appareils du réseau et l'option `-browser` gardent le lecteur web et VLC
externe. `-no-browser` conserve le fonctionnement serveur seul. En l'absence
de VLC 3 ou de WebView2, le lancement revient à l'interface web.

Les commandes natives sont accessibles uniquement par les bindings WebView2,
avec une clé aléatoire propre à la session. Aucun endpoint HTTP de pilotage
natif n'est exposé. Le serveur valide l'appartenance des fichiers au catalogue,
les chemins relatifs, les sous-titres associés et le mode enfant avant lecture.
Les liens externes s'ouvrent dans le navigateur du système.

La lecture conserve le volume lors du passage entre médias. La progression
est enregistrée régulièrement, à la fermeture du lecteur et avant la destruction
de la fenêtre. Les films en plusieurs parties s'enchaînent et conservent l'indice
de la partie en cours. LibVLC est arrêté avant de détruire sa surface vidéo.

Commandes : espace pour pause/reprise, flèches gauche/droite pour ±10 secondes,
F pour le plein écran, Échap pour sortir du plein écran puis fermer le lecteur.
F11 et le bouton ⛶ activent aussi le mode cinéma. Les commandes s'effacent
après un court délai de lecture, et le mouvement de la souris les fait revenir,
y compris au-dessus de la surface vidéo native. La vidéo occupe alors tout
l'écran, conserve son format d'origine et affiche des bandes noires sans recadrage.
Les sélecteurs proposent les pistes audio, les sous-titres internes et les
sous-titres externes associés au film.

## Vérifications

```powershell
go test ./...
$env:AC_LIBVLC_TEST_MEDIA = 'D:\chemin\film.mkv'
go test ./internal/desktop -run TestLibVLCPlayback -v -count=1 -timeout 45s
go build -ldflags '-s -w -H=windowsgui' -o 'Absolute Cinema.exe' ./cmd/absolute_cinema
```

Le test d'intégration est facultatif dans la suite courante. Il utilise un vrai
fichier local, vérifie l'avancement, une sortie vidéo non vide, le volume,
la pause et le déplacement, sans modifier le média.

Références : [LibVLC 3](https://videolan.videolan.me/vlc-3.0/group__libvlc__media__player.html),
[go-webview2](https://github.com/jchv/go-webview2). Les licences des dépendances
sont rassemblées dans `THIRD_PARTY_NOTICES.txt` ; celles de VLC accompagnent
sa distribution portable.
