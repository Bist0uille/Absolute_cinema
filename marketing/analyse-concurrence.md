# Analyse concurrentielle — Absolute Cinema vs Plex / Jellyfin / Emby / Kodi

*Mise à jour : juillet 2026. Contexte : Plex a triplé son pass lifetime à 749 $
(07/2026), Jellyfin (gratuit) domine chez les technophiles mais exige un serveur.*

## Notre position

Tous les concurrents sont des **serveurs** : une machine allumée, une installation,
des clients à connecter. Nous sommes **le disque** : zéro machine dédiée, zéro
installation, zéro compte. Chaque fonctionnalité s'évalue à l'aune d'une règle :
**est-ce que ça reste « je branche, ça marche » ?**

## Ce qu'ils ont — et ce qu'on en fait

| Fonctionnalité | Plex | Jellyfin | Nous (v1.3) | Décision |
|---|---|---|---|---|
| Reconnaissance auto + affiches | ✓ | ✓ | ✓ (99,8 % mesuré) | ✅ fait |
| Reprise de lecture / « Continuer » | ✓ | ✓ | ✓ (suit le disque, pas l'ordinateur !) | ✅ fait (v1.3) |
| Lecture dans le navigateur | ✓ | ✓ | ✓ h264/webm, repli VLC sinon | ✅ fait (v1.3) |
| Accueil type Netflix (héros, rangées) | ✓ | ✓ | ✓ héros + Reprendre + Récents | ✅ fait (v1.3) |
| Épisode suivant / suivi des vus | ✓ | ✓ | ✓ | ✅ fait (v1.3) |
| Contrôle parental | Plex Pass | ✓ | ✓ PIN + visas officiels, côté serveur | ✅ fait (v1.2) |
| Scan automatique | ✓ | ✓ | ✓ à chaque ouverture | ✅ fait (v1.3) |
| Multilingue | ✓ | ✓ | ✓ FR/EN | ✅ fait (v1.2) |
| Accès TV/tablette du foyer | ✓ | ✓ | ✓ option réseau local (navigateur TV) | ✅ fait (v1.3, basique) |
| **Transcodage à la volée** | Pass | ✓ | ✗ | ❌ refusé : c'est LE truc qui exige un serveur puissant, de la config, du matériel. Notre réponse : repli VLC (lit tout, nativement). |
| **Streaming à distance (hors du foyer)** | ✓ | ~ | ✗ | ❌ refusé : comptes, relais, sécurité, bande passante — l'anti-thèse du produit. Le disque EST le transport. |
| Applis mobiles/TV natives | ✓ | ✓ | ✗ (web responsive) | ⏸ plus tard si traction ; le navigateur suffit au pilote |
| Skip intro / crédits | Pass | plugin | ✗ | ❌ refusé : détection lourde, gadget |
| DVR / TV en direct | Pass | ✓ | ✗ | ❌ hors sujet |
| Profils multi-utilisateurs | ✓ | ✓ | ✗ (mode enfant = le vrai besoin famille) | ⏸ à envisager en « profils légers » si demandé |
| Plugins / extensions | ✓ | ✓ | ✗ | ❌ refusé : surface de maintenance et de sécurité |
| Collections / listes | ✓ | ✓ | ✗ | ⏸ candidat simple (« sagas » auto via TMDB collections) |
| Bandes-annonces | Pass | plugin | ✗ | ⏸ lien YouTube possible, à 1 clic — candidat facile |
| Téléchargement de sous-titres | ✓ | plugin | ✗ | ⏸ OpenSubtitles a une API payante ; à étudier avec la monétisation |
| DLNA (TV anciennes) | ✓ | ✓ | ✗ | ⏸ le mode réseau local couvre les TV modernes (navigateur) ; DLNA = gros chantier UPnP pour un parc en voie de disparition |

## Nos avantages structurels (à marteler dans la comm)

1. **La bibliothèque suit le disque** — reprise de lecture comprise. Chez Plex/Jellyfin,
   changer de machine = tout reconfigurer ; chez nous c'est le cas d'usage nominal.
2. **Hors ligne total** après le premier scan (cabane, van, avion, coupure).
3. **Aucune maintenance** : pas de serveur qui tourne, pas de mises à jour de NAS.
4. **Vie privée par construction** : aucun compte, aucune télémétrie, rien ne sort.

## Faiblesses assumées (réponses prêtes pour les forums)

- « Pas de transcodage ? » → VLC lit tout nativement, y compris ce que les serveurs
  transcodent péniblement. Un clic de plus, zéro serveur de moins.
- « Pas de streaming distant ? » → Le produit s'appelle vidéothèque *de poche* :
  elle voyage physiquement. Pour le reste, Plex/Jellyfin existent.
- « Une appli TV ? » → Option réseau local + navigateur de la TV. Natif à l'étude
  si la demande le justifie.
