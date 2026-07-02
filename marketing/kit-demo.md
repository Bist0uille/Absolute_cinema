# Kit démo — vidéothèque de poche

## ⚠️ Règle d'or pour TOUS les visuels publics

Aucun nom de fichier pirate visible (pas de « BluRay.x264-YIFY » à l'écran). Deux options :
- Cadrer sur la grille d'affiches et les fiches (les noms de fichiers ne sont visibles qu'en bas des fiches — éviter de scroller jusque-là en vidéo publique) ;
- Ou préparer un petit disque de démo avec des films libres de droits (Blender Open Movies : *Big Buck Bunny*, *Sintel*, *Tears of Steel* — gratuits et légaux) pour les gros plans.
Ne jamais écrire « téléchargez vos films » : toujours « vos films », « votre collection ».

## Script vidéo 60–90 s (screen recording + téléphone)

| Temps | Plan | Voix off / texte à l'écran |
|---|---|---|
| 0–8 s | Main qui branche le disque USB sur un PC portable | « Tout est sur ce disque. » |
| 8–15 s | Double-clic sur « Lancer », le navigateur s'ouvre | « Un double-clic… » |
| 15–35 s | Scroll fluide de la grille d'affiches, filtre par genre, recherche d'un titre | « …et votre collection devient une vraie vidéothèque. Affiches, synopsis, tout est là. » |
| 35–45 s | Clic sur un film → fiche → bouton ▶ → VLC démarre | « Un clic pour lancer la lecture, sous-titres compris. » |
| 45–60 s | Bouton 🧸 → saisie du PIN → la grille se filtre (que des films famille) | « Mode enfant : code PIN, et les petits ne voient que ce qui est de leur âge. » |
| 60–75 s | On débranche, on rebranche sur un Mac, même bibliothèque | « PC, Mac, Linux — la bibliothèque voyage avec le disque. Sans compte, sans abonnement, sans internet. » |
| 75–85 s | Écran final | « La vidéothèque de poche. Bientôt disponible — lien en description. » |

Outils gratuits : OBS Studio (capture écran), CapCut/DaVinci Resolve (montage), téléphone sur trépied improvisé pour les plans réels.

## Les 5 captures d'écran à faire (pour landing + posts)

1. La grille d'affiches complète (onglet Films, plein écran, thème sombre) — LA capture héro.
2. Une fiche film avec fond d'écran, synopsis FR, badges VF/1080p.
3. Le mode enfant actif (bandeau vert + grille filtrée famille).
4. La page Doublons avec une recommandation « garder / supprimer ».
5. Une série dépliée (Breaking Bad : saisons, titres d'épisodes en français).

## Annonce Leboncoin (test de demande — disque pré-équipé)

**Titre** : Vidéothèque de poche 2 To — vos films avec affiches sur n'importe quel ordi

**Texte** :
> Disque dur 2 To (neuf) livré avec un logiciel exclusif : copiez vos films dessus,
> il les reconnaît tout seul et crée une magnifique bibliothèque avec affiches,
> résumés et sous-titres. Branchez ensuite le disque sur n'importe quel PC, Mac ou
> Linux : votre vidéothèque s'ouvre en un double-clic. Sans abonnement, sans compte,
> sans internet. Mode enfant avec code PIN inclus.
> Le disque est vendu vide : vous y mettez vos propres films.
> Mise en route accompagnée (15 min au téléphone) incluse. 99 €.

**Objectif du test** : compter les contacts sérieux sur 2–3 semaines. Ne pas conclure de vente avant d'avoir créé la micro-entreprise (répondre « premier stock en cours, je vous recontacte très vite » si ça mord).

## Post forum / groupe Facebook (feedback, pas vente)

**Titre** : J'ai fabriqué une « vidéothèque de poche » — votre avis ?

> Salut ! J'en avais marre des solutions compliquées (serveur Plex, NAS…) juste pour
> montrer mes films avec des jolies affiches. Alors j'ai fait un petit logiciel qui
> vit SUR le disque dur : on branche le disque sur n'importe quel ordi, double-clic,
> et toute la collection s'affiche avec affiches, synopsis, sous-titres, mode enfant.
> Pas de compte, pas d'abonnement, pas d'internet nécessaire.
> [vidéo/captures]
> Question honnête : est-ce que ça vous serait utile, ou vous avez déjà mieux ?
> Qu'est-ce qui manquerait pour que vous l'utilisiez ?

Où poster (France) : groupes Facebook « Home Cinéma France », « Plex France », « NAS et serveurs multimédia » ; forum HomeCinema-fr ; r/france (fil tech du week-end) ; Discord ciné/tech où vous êtes déjà. Étaler sur 2 semaines, adapter le ton à chaque endroit, répondre à TOUS les commentaires (c'est là que sont les vrais enseignements).

## Tableau de bord du test (à tenir à jour, simple fichier ou papier)

| Indicateur | Compteur | Objectif GO |
|---|---|---|
| Emails collectés (landing) | | ≥ 30 |
| Demandes d'achat spontanées | | ≥ 5 |
| Contacts Leboncoin sérieux | | ≥ 3 |
| Entretiens réalisés | | 10 |
| Verdict entretiens (paieraient ?) | | majorité oui |

---

## Posts anglophones (marché « réfugiés Plex », fenêtre 2026)

### r/selfhosted ou r/DataHoarder (titre + corps)

**Title:** I built a "pocket media library" — a single binary that lives ON the external drive (no server, no account, works offline)

> Plex tripling its lifetime price got me thinking about the opposite approach:
> instead of a server, the library lives on the drive itself.
>
> You drop a single binary at the root of your external drive, double-click on any
> Windows/Mac/Linux machine, and your movie collection opens in the browser with
> posters, synopses, subtitles, a PIN-locked kids mode and duplicate detection.
> All metadata is cached on the drive, so after the first scan it works fully
> offline. Nothing gets installed on the host computer.
>
> Free beta, open source (Go, GPL-3.0, zero dependencies): [site] / [github]
> It measured 99%+ recognition on my own messy 1000-file collection. I'd love to
> know how it handles yours — feedback and issues very welcome.

Règles Reddit : lire les règles d'autopromo de chaque sub (r/selfhosted a un fil
« What are you working on » hebdo qui est l'endroit idéal) ; répondre à tous les
commentaires ; ne JAMAIS mentionner de sources de films.

### Hacker News (Show HN)

**Title:** Show HN: A movie library that lives on the hard drive itself (Go, offline, no server)

> Premier commentaire (à poster soi-même) : expliquer la genèse (une vraie collection
> familiale en désordre), l'architecture (binaire unique go:embed, chemins relatifs
> pour la portabilité, cache TMDB intégral pour l'offline), et le choix GPL.
> Les threads HN aiment les détails techniques honnêtes et les limites assumées
> (pas de transcodage, VLC requis, pas de streaming distant).

### Où poster, dans l'ordre

1. r/selfhosted (fil hebdo) — jour 1
2. r/DataHoarder — jour 3
3. Show HN — jour 7 (avec la version EN au point et 2-3 retours déjà intégrés)
4. Forums Jellyfin/Kodi (sections « alternatives/outils ») — au fil de l'eau
