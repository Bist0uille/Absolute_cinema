# Brouillon d'email — licence commerciale TMDB

À envoyer via le formulaire « API for Business » (https://www.themoviedb.org/api-for-business)
ou à l'adresse indiquée dans leurs conditions. Garder une copie de la réponse : elle
conditionne le modèle économique (clé embarquée vs clé personnelle par client).

---

**Subject:** Commercial licensing inquiry — small offline personal media library app (France)

Hello,

I am an independent developer in France preparing to commercialize a small desktop
application and I would like to understand your commercial licensing terms before
any launch.

**What the product does:** it is a portable, offline personal media library. Users
point it at their own video files on an external hard drive; the app matches titles
against the TMDB API (search, movie/TV details, posters) and caches metadata and
images locally on the user's drive. After the initial scan, the app works fully
offline. TMDB attribution is displayed in the app and on the website.

**Expected usage:**
- Very low volume at launch: likely under 100 customers in the first year (hobby-scale,
  French market first).
- Each customer's library is scanned once (~300–1000 titles), then requests are only
  made for newly added files. All responses are cached locally, so per-user API usage
  is minimal and one-time.
- No redistribution of your data beyond the user's own local cache; no bulk export.

**My questions:**
1. What licensing terms and pricing would apply to this kind of low-volume commercial
   desktop application?
2. Is there a minimum fee, or a tier suitable for an early-stage/indie product?
3. Would you prefer the app to embed a single licensed API key, or is a model where
   each end user registers their own personal key acceptable?

Thank you for your time — I want to do this properly from day one.

Best regards,
[Nom]
[email] — France

---

## Replis selon la réponse

| Réponse TMDB | Action |
|---|---|
| Tarif abordable (ex. < 50 €/mois ou % raisonnable) | Signer, embarquer la clé → grosse simplification UX |
| Trop cher pour le stade actuel | Modèle « chaque client crée sa clé perso gratuite » + assistant intégré (déjà en place) — à re-négocier quand le volume le justifie |
| Pas de réponse sous 3 semaines | Relance, puis étudier OMDb (licence commerciale ~1 $/mois par palier) ou Wikidata (libre, moins riche en affiches) |
