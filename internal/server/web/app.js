/* absolute_cinema — interface (vanilla JS, routing par hash) */
"use strict";

const APP_VERSION = "1.2.0";
const GITHUB_REPO = "Bist0uille/Absolute_cinema";

const $ = (sel, el = document) => el.querySelector(sel);
const app = $("#app");

const state = {
  lib: { movies: [], series: [], unmatched: [] },
  tab: "films",
  filters: { genre: "", lang: "", decade: "", sort: "titre" },
  search: "",
  scanning: false,
  lang: "fr",
};

/* ---------- i18n ----------
   Le français est la langue des clés ; EN est le dictionnaire. */

const EN = {
  "Bibliothèque": "Library",
  "Doublons": "Duplicates",
  "Réglages": "Settings",
  "Rechercher un titre…": "Search a title…",
  "Films": "Movies",
  "Séries": "TV Shows",
  "Tous les genres": "All genres",
  "Toutes langues": "All languages",
  "Toutes décennies": "All decades",
  "Tri : titre": "Sort: title",
  "Tri : année": "Sort: year",
  "Tri : note": "Sort: rating",
  "films": "movies",
  "séries": "shows",
  "à vérifier": "to review",
  "saison(s)": "season(s)",
  "ép.": "ep.",
  "égarée": "misplaced",
  "La bibliothèque est vide.": "The library is empty.",
  "Scanner le disque": "Scan the drive",
  "Re-scanner le disque": "Re-scan the drive",
  "Film introuvable.": "Movie not found.",
  "Série introuvable.": "Show not found.",
  "Pas de synopsis.": "No synopsis.",
  "match incertain — corrigez dans Réglages": "uncertain match — fix it in Settings",
  "Fichier": "File",
  "versions — voir Doublons": "versions — see Duplicates",
  "Sous-titres : aucun": "Subtitles: none",
  "Sous-titres": "Subtitles",
  "ST : aucun": "Subs: none",
  "▶ Lire": "▶ Play",
  "partie(s)": "part(s)",
  "Saison": "Season",
  "épisode(s)": "episode(s)",
  "Épisode": "Episode",
  "des fichiers sont dans [Films]": "some files sit in the movies folder",
  "Mauvaise fiche ? Corriger le film/la série associé(e)…": "Wrong match? Fix the linked movie/show…",
  "Chercher le bon titre sur TMDB…": "Search the right title on TMDB…",
  "Chercher": "Search",
  "Chargement…": "Loading…",
  "Disque en lecture seule : la suppression est désactivée.": "Drive is read-only: deletion is disabled.",
  "Aucun doublon détecté.": "No duplicates found.",
  "Résolution": "Resolution",
  "Langue": "Language",
  "Taille": "Size",
  "✓ garder": "✓ keep",
  "✗ supprimer ?": "✗ delete?",
  "🗑 Corbeille": "🗑 Trash",
  "Épisodes en double": "Duplicate episodes",
  "Séries égarées dans [Films]": "TV shows misplaced in the movies folder",
  "Ces séries ont des fichiers rangés dans le dossier [Films]. Elles apparaissent quand même dans l'onglet Séries ; à déplacer à la main si vous voulez ranger le disque.":
    "These shows have files stored in the movies folder. They still appear under TV Shows; move them by hand if you want a tidy drive.",
  "Corbeille": "Trash",
  "↩ Restaurer": "↩ Restore",
  "Vider la corbeille…": "Empty trash…",
  "Les fichiers sont déplacés dans .absolute_cinema/corbeille/ sur le disque — rien n'est effacé définitivement.":
    "Files are moved to .absolute_cinema/corbeille/ on the drive — nothing is permanently deleted.",
  "Mettre à la corbeille ?": "Move to trash?",
  "Le fichier sera déplacé vers .absolute_cinema/corbeille/ (restaurable) :": "The file will be moved to .absolute_cinema/corbeille/ (restorable):",
  "Annuler": "Cancel",
  "Mettre à la corbeille": "Move to trash",
  "Déplacé en corbeille ✓": "Moved to trash ✓",
  "Échec :": "Failed:",
  "Restauré ✓ — relancez un scan pour le réintégrer": "Restored ✓ — run a scan to bring it back",
  "Vider définitivement la corbeille ?": "Permanently empty the trash?",
  "Tous les fichiers de la corbeille seront définitivement supprimés. Cette action est irréversible.":
    "All trashed files will be permanently deleted. This cannot be undone.",
  "Vider": "Empty",
  "Corbeille vidée ✓": "Trash emptied ✓",
  "Clé API TMDB": "TMDB API key",
  "Enregistrer": "Save",
  "Collez d'abord votre clé.": "Paste your key first.",
  "Clé validée et enregistrée ✓": "Key validated and saved ✓",
  "Langue / Language": "Language / Langue",
  "Langue de l'interface": "Interface language",
  "Langue des affiches et synopsis": "Posters & synopsis language",
  "Après un changement de langue des métadonnées, relancez un scan.": "After changing metadata language, run a scan again.",
  "Français": "Français",
  "Scan de la bibliothèque": "Library scan",
  "Médiathèque :": "Media root:",
  "Données :": "Data:",
  "lecture seule": "read-only",
  "Dernier scan :": "Last scan:",
  "jamais": "never",
  "🧸 Mode enfant": "🧸 Kids mode",
  "Activer le mode enfant…": "Enable kids mode…",
  "Activer le mode enfant": "Enable kids mode",
  "Quitter le mode enfant (code PIN)": "Exit kids mode (PIN code)",
  "Quitter le mode enfant": "Exit kids mode",
  "Entrez le code PIN parent :": "Enter the parent PIN code:",
  "Mode enfant désactivé": "Kids mode disabled",
  "🧸 Activer le mode enfant": "🧸 Enable kids mode",
  "Seuls les films et séries classés jusqu'à l'âge choisi seront visibles et lisibles. Les titres sans classification sont cachés par prudence (classez-les dans Réglages). Pour ressortir du mode enfant, il faudra le code PIN.":
    "Only movies and shows rated up to the chosen age will be visible and playable. Unrated titles are hidden as a precaution (rate them in Settings). Exiting kids mode requires the PIN code.",
  "Âge maximum :": "Maximum age:",
  "ans": "y/o",
  "Choisissez un code PIN parent (au moins 4 chiffres) :": "Choose a parent PIN code (at least 4 digits):",
  "Activer": "Enable",
  "Valider": "OK",
  "Mode enfant activé 🧸": "Kids mode enabled 🧸",
  "Mode enfant — seuls les titres jusqu'à": "Kids mode — only titles up to",
  "ans sont visibles": "y/o are visible",
  "Classification :": "Rating:",
  "Classification effacée": "Rating cleared",
  "Classé": "Rated",
  "tous publics": "all audiences",
  "Revenir à la classification TMDB": "Back to the TMDB rating",
  "Le mode enfant ne montre (et ne lit) que les titres classés jusqu'à l'âge choisi — classification officielle TMDB. Les titres sans classification sont cachés par prudence. La sortie du mode enfant demande un code PIN.":
    "Kids mode only shows (and plays) titles rated up to the chosen age — official TMDB ratings. Unrated titles are hidden as a precaution. Exiting kids mode requires a PIN code.",
  "Un code PIN est défini.": "A PIN code is set.",
  "Le code PIN sera défini à la première activation.": "The PIN code will be set on first activation.",
  "titre(s) sans classification — classer manuellement": "unrated title(s) — rate manually",
  "Cliquez sur un âge pour classer le titre (mémorisé, survit aux re-scans). Sans classement, le titre reste invisible en mode enfant.":
    "Click an age to rate the title (remembered across scans). Unrated titles stay hidden in kids mode.",
  "Tous les titres de la bibliothèque ont une classification ✓": "Every title in the library has a rating ✓",
  "série": "show",
  "film": "movie",
  "À vérifier": "To review",
  "Le match TMDB est incertain pour ces fiches. Cherchez le bon titre et cliquez dessus pour corriger.":
    "The TMDB match is uncertain for these entries. Search the right title and click it to fix.",
  "confiance": "confidence",
  "Chercher le bon titre…": "Search the right title…",
  "Non identifiés": "Unidentified",
  "Aucune fiche TMDB trouvée pour ces fichiers. Cherchez manuellement, ou ignorez (concerts, bonus, documents…).":
    "No TMDB entry found for these files. Search manually, or ignore (concerts, extras, documents…).",
  "Chercher sur TMDB…": "Search on TMDB…",
  "Film…": "Movie…",
  "Série…": "Show…",
  "(sans titre)": "(untitled)",
  "Recherche…": "Searching…",
  "Aucun résultat.": "No results.",
  "Correction enregistrée — re-scan en cours…": "Fix saved — re-scanning…",
  "À propos": "About",
  "Scan :": "Scan:",
  "Scan terminé ✓": "Scan finished ✓",
  "Parcours du disque": "Walking the drive",
  "Films — récupération TMDB": "Movies — fetching TMDB",
  "Séries — récupération TMDB": "TV shows — fetching TMDB",
  "Lecture lancée dans VLC ▶": "Now playing in VLC ▶",
  "Lecture lancée dans le lecteur par défaut ▶": "Now playing in the default player ▶",
  "Lecture impossible :": "Cannot play:",
  "VLC introuvable : lecture avec le lecteur par défaut, sous-titre non transmis.": "VLC not found: playing with the default player, subtitle not passed.",
  "VLC introuvable : lecture avec le lecteur par défaut.": "VLC not found: playing with the default player.",
  "Erreur de chargement :": "Loading error:",
  "Nouvelle version disponible :": "New version available:",
  "Télécharger": "Download",
  "Classification inconnue": "Unknown rating",
  "Âge minimum": "Minimum age",
  /* onboarding */
  "Bienvenue !": "Welcome!",
  "Transformons ce disque en vidéothèque. Deux petites choses et c'est parti :":
    "Let's turn this drive into a movie library. Two quick things and you're set:",
  "1. Choisissez vos langues": "1. Pick your languages",
  "Interface :": "Interface:",
  "Affiches & synopsis :": "Posters & synopsis:",
  "2. Collez votre clé TMDB gratuite": "2. Paste your free TMDB key",
  "Les affiches et synopsis viennent de The Movie Database. Créez un compte gratuit (2 min), puis copiez la « clé d'API » depuis":
    "Posters and synopses come from The Movie Database. Create a free account (2 min), then copy the “API key” from",
  "Paramètres → API": "Settings → API",
  "Votre clé API TMDB": "Your TMDB API key",
  "Lancer le scan 🎬": "Start the scan 🎬",
  "La clé reste sur votre disque, rien n'est envoyé ailleurs.": "The key stays on your drive; nothing is sent anywhere else.",
  "Pensez aussi à installer VLC pour la lecture :": "Also install VLC for playback:",
};

function t(s) { return state.lang === "en" ? (EN[s] ?? s) : s; }
function locale() { return state.lang === "en" ? "en-GB" : "fr-FR"; }

function resolveLang() {
  const cfg = state.lib.ui_lang;
  if (cfg === "fr" || cfg === "en") return cfg;
  const saved = localStorage.getItem("ac_lang");
  if (saved === "fr" || saved === "en") return saved;
  return (navigator.language || "fr").toLowerCase().startsWith("fr") ? "fr" : "en";
}

function applyChromeLang() {
  document.documentElement.lang = state.lang;
  document.querySelector('[data-nav="home"]').textContent = t("Bibliothèque");
  document.querySelector('[data-nav="doublons"]').textContent = t("Doublons");
  document.querySelector('[data-nav="reglages"]').textContent = t("Réglages");
  $("#search").placeholder = t("Rechercher un titre…");
}

/* ---------- utilitaires ---------- */

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

function fmtSize(b) {
  if (!b) return "";
  const gb = b / (1 << 30);
  const units = state.lang === "en" ? ["GB", "MB"] : ["Go", "Mo"];
  return gb >= 1 ? gb.toFixed(1) + " " + units[0] : Math.round(b / (1 << 20)) + " " + units[1];
}

function toast(msg, ms = 3500) {
  const el = $("#toast");
  el.textContent = msg;
  el.classList.remove("hidden");
  clearTimeout(el._h);
  el._h = setTimeout(() => el.classList.add("hidden"), ms);
}

async function api(path, opts) {
  const r = await fetch(path, opts);
  const data = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(data.error || r.statusText);
  return data;
}

function posterImg(p, cls = "poster") {
  if (p) return `<img class="${cls}" loading="lazy" src="/media-cache/${esc(p)}" alt="">`;
  return `<div class="${cls} placeholder">🎞️</div>`;
}

function bestBadge(versions) {
  const order = ["MULTI", "VOSTFR", "VF", "VO"];
  let best = "";
  for (const v of versions || []) {
    const b = v.lang_badge || "";
    if (b && (best === "" || order.indexOf(b) < order.indexOf(best))) best = b;
  }
  return best;
}

function badgeHTML(b) { return b ? `<span class="badge ${esc(b)}">${esc(b)}</span>` : ""; }

function ageBadge(a) {
  if (a === undefined || a === null || a < 0) return `<span class="badge ageunk" title="${t("Classification inconnue")}">?</span>`;
  const cls = a >= 18 ? "age18" : a >= 16 ? "age16" : a >= 12 ? "age12" : a >= 10 ? "age10" : "age0";
  const label = a === 0 ? (state.lang === "en" ? "ALL" : "TP") : a;
  return `<span class="badge ${cls}" title="${t("Âge minimum")}">${label}</span>`;
}

/* ---------- mode enfant ---------- */

function updateChrome() {
  applyChromeLang();
  const kid = state.lib.kid_mode;
  document.querySelector('[data-nav="doublons"]').style.display = kid ? "none" : "";
  document.querySelector('[data-nav="reglages"]').style.display = kid ? "none" : "";
  const btn = $("#kidbtn");
  btn.classList.toggle("on", !!kid);
  btn.textContent = kid ? "🔓" : "🧸";
  btn.title = kid ? t("Quitter le mode enfant (code PIN)") : t("Activer le mode enfant");
  let banner = $("#kidbanner");
  if (kid && !banner) {
    banner = document.createElement("div");
    banner.id = "kidbanner";
    banner.textContent = `🧸 ${t("Mode enfant — seuls les titres jusqu'à")} ${state.lib.kid_max_age} ${t("ans sont visibles")}`;
    $("#topbar").after(banner);
  } else if (!kid && banner) banner.remove();
}

window._kidToggle = () => {
  if (state.lib.kid_mode) {
    pinModal(t("Quitter le mode enfant"), t("Entrez le code PIN parent :"), async (pin) => {
      await api("/api/kidmode", { method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enable: false, pin }) });
      toast(t("Mode enfant désactivé"));
      location.hash = "#/";
      await loadLibrary();
    });
  } else {
    kidEnableModal();
  }
};

function pinModal(title, label, onOk) {
  const root = $("#modal-root");
  root.innerHTML = `<div class="modal-back" onclick="if(event.target===this)this.remove()">
    <div class="modal">
      <h3>${esc(title)}</h3>
      <p class="help">${esc(label)}</p>
      <input class="pin" type="password" inputmode="numeric" maxlength="8" autofocus>
      <div class="actions">
        <button class="ghost" onclick="this.closest('.modal-back').remove()">${t("Annuler")}</button>
        <button class="play" id="pin-ok">${t("Valider")}</button>
      </div>
    </div>
  </div>`;
  const input = $(".pin", root);
  const go = async () => {
    try { await onOk(input.value); root.innerHTML = ""; }
    catch (e) { toast(e.message, 5000); input.value = ""; input.focus(); }
  };
  $("#pin-ok", root).onclick = go;
  input.onkeydown = e => { if (e.key === "Enter") go(); };
  input.focus();
}

function kidEnableModal() {
  const hasPin = state.lib.has_pin;
  const curAge = state.lib.kid_max_age || 10;
  const root = $("#modal-root");
  root.innerHTML = `<div class="modal-back" onclick="if(event.target===this)this.remove()">
    <div class="modal">
      <h3>${t("🧸 Activer le mode enfant")}</h3>
      <p class="help">${t("Seuls les films et séries classés jusqu'à l'âge choisi seront visibles et lisibles. Les titres sans classification sont cachés par prudence (classez-les dans Réglages). Pour ressortir du mode enfant, il faudra le code PIN.")}</p>
      <div class="field" style="align-items:center">
        <span>${t("Âge maximum :")}</span>
        <div class="agepick" id="agepick">
          ${[6, 10, 12, 16].map(a => `<button data-age="${a}" class="${a === curAge ? "sel" : ""}">${a} ${t("ans")}</button>`).join("")}
        </div>
      </div>
      ${hasPin ? "" : `<p class="help">${t("Choisissez un code PIN parent (au moins 4 chiffres) :")}</p>
      <input class="pin" type="password" inputmode="numeric" maxlength="8" placeholder="••••">`}
      <div class="actions">
        <button class="ghost" onclick="this.closest('.modal-back').remove()">${t("Annuler")}</button>
        <button class="play" id="kid-ok">${t("Activer")}</button>
      </div>
    </div>
  </div>`;
  let age = curAge;
  root.querySelectorAll("#agepick button").forEach(b => b.onclick = () => {
    root.querySelectorAll("#agepick button").forEach(x => x.classList.remove("sel"));
    b.classList.add("sel"); age = +b.dataset.age;
  });
  $("#kid-ok", root).onclick = async () => {
    const pinInput = $(".pin", root);
    try {
      await api("/api/kidmode", { method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enable: true, pin: pinInput ? pinInput.value : "", max_age: age }) });
      root.innerHTML = "";
      toast(t("Mode enfant activé 🧸"));
      location.hash = "#/";
      await loadLibrary();
    } catch (e) { toast(e.message, 5000); }
  };
}

window._setRating = async (mediaType, tmdbID, age) => {
  try {
    await api("/api/rating", { method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ media_type: mediaType, tmdb_id: tmdbID, age }) });
    toast(age < 0 ? t("Classification effacée") : `${t("Classé")} ${age === 0 ? t("tous publics") : age + " " + t("ans")} ✓`);
    await loadLibrary();
  } catch (e) { toast(e.message, 5000); }
};

function ratingPicker(mediaType, tmdbID, current) {
  const zero = state.lang === "en" ? "ALL" : "TP";
  return `<div class="agepick">
    <span class="help">${t("Classification :")}</span>
    ${[[0, zero], [10, "10"], [12, "12"], [16, "16"], [18, "18"]].map(([a, l]) =>
      `<button class="${current === a ? "sel" : ""}" onclick="_setRating('${mediaType}', ${tmdbID}, ${a})">${l}</button>`).join("")}
    ${current >= 0 ? `<button onclick="_setRating('${mediaType}', ${tmdbID}, -1)" title="${t("Revenir à la classification TMDB")}">↺</button>` : ""}
  </div>`;
}

/* ---------- chargement ---------- */

async function loadLibrary() {
  const data = await api("/api/library");
  state.lib = data;
  state.scanning = data.scanning;
  state.lang = resolveLang();
  if (data.scanning) pollScan();
  updateChrome();
  render();
}

let pollTimer = null;
function pollScan() {
  if (pollTimer) return;
  const bar = $("#scanbar");
  bar.classList.remove("hidden");
  pollTimer = setInterval(async () => {
    try {
      const st = await api("/api/scan/status");
      const label = $(".scan-label", bar);
      const fill = $(".scan-fill", bar);
      const phases = { walk: t("Parcours du disque"), tmdb: t("Films — récupération TMDB"), series: t("Séries — récupération TMDB") };
      if (st.scanning) {
        const pct = st.total ? Math.round(100 * st.done / st.total) : 0;
        label.textContent = `${phases[st.phase] || st.phase}… ${st.total ? `${st.done}/${st.total} (${pct} %)` : ""} ${st.current || ""}`;
        fill.style.width = (st.total ? pct : 5) + "%";
      } else {
        clearInterval(pollTimer); pollTimer = null;
        bar.classList.add("hidden");
        if (st.error) toast(t("Scan :") + " " + st.error, 8000);
        else if (st.phase === "done") { toast(t("Scan terminé ✓")); await loadLibrary(); }
      }
    } catch { /* serveur occupé, on réessaie */ }
  }, 800);
}

async function startScan() {
  try {
    await api("/api/scan", { method: "POST" });
    state.scanning = true;
    pollScan();
  } catch (e) { toast(e.message); }
}

/* ---------- vérification de mise à jour (API GitHub, silencieuse) ---------- */

async function checkUpdate() {
  try {
    const cached = JSON.parse(localStorage.getItem("ac_update") || "null");
    if (cached && Date.now() - cached.at < 86400e3) { state.update = cached; return; }
    const r = await fetch(`https://api.github.com/repos/${GITHUB_REPO}/releases/latest`);
    if (!r.ok) return;
    const rel = await r.json();
    const latest = (rel.tag_name || "").replace(/^v/, "").replace(/-beta$/, "");
    const info = { at: Date.now(), latest, url: rel.html_url, newer: isNewer(latest, APP_VERSION) };
    localStorage.setItem("ac_update", JSON.stringify(info));
    state.update = info;
  } catch { /* hors ligne : tant pis */ }
}

function isNewer(a, b) {
  const pa = a.split(".").map(Number), pb = b.split(".").map(Number);
  for (let i = 0; i < 3; i++) {
    if ((pa[i] || 0) > (pb[i] || 0)) return true;
    if ((pa[i] || 0) < (pb[i] || 0)) return false;
  }
  return false;
}

function updateBanner() {
  if (!state.update?.newer) return "";
  return `<div class="block" style="border-left:3px solid var(--accent)">
    ${t("Nouvelle version disponible :")} <strong>${esc(state.update.latest)}</strong>
    — <a style="color:var(--accent)" href="${esc(state.update.url)}" target="_blank">${t("Télécharger")}</a>
  </div>`;
}

/* ---------- lecture ---------- */

async function play(path, subPath, parts) {
  try {
    const res = await api("/api/play", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path, sub_path: subPath || "", parts: parts || [] }),
    });
    if (res.warning) toast(t(res.warning));
    else toast(res.player === "vlc" ? t("Lecture lancée dans VLC ▶") : t("Lecture lancée dans le lecteur par défaut ▶"));
  } catch (e) { toast(t("Lecture impossible :") + " " + e.message, 6000); }
}

window._playVersion = (btn) => {
  const row = btn.closest("[data-path]");
  const sel = $("select", row);
  play(row.dataset.path, sel ? sel.value : "", JSON.parse(row.dataset.parts || "[]"));
};

/* ---------- routing ---------- */

window.addEventListener("hashchange", render);

function route() {
  const h = location.hash || "#/";
  const parts = h.slice(2).split("/");
  return { page: parts[0] || "home", arg: decodeURIComponent(parts[1] || "") };
}

function setNav(name) {
  document.querySelectorAll("#topbar nav a").forEach(a => {
    a.classList.toggle("active", a.dataset.nav === name);
  });
}

function render() {
  let { page, arg } = route();
  if (state.lib.kid_mode && (page === "doublons" || page === "reglages")) {
    location.hash = "#/";
    page = "home";
  }
  if (page === "home" || page === "") { setNav("home"); renderHome(); }
  else if (page === "film") { setNav("home"); renderMovie(arg); }
  else if (page === "serie") { setNav("home"); renderSeries(arg); }
  else if (page === "doublons") { setNav("doublons"); renderDupes(); }
  else if (page === "reglages") { setNav("reglages"); renderSettings(); }
  else { setNav("home"); renderHome(); }
}

/* ---------- onboarding (premier lancement) ---------- */

function renderOnboarding() {
  app.innerHTML = `
  <div style="max-width:640px;margin:30px auto">
    <div class="block" style="text-align:center">
      <div style="font-size:46px;margin-bottom:8px">🎬</div>
      <h1 style="margin-bottom:10px">${t("Bienvenue !")}</h1>
      <p class="help">${t("Transformons ce disque en vidéothèque. Deux petites choses et c'est parti :")}</p>
    </div>
    <div class="block">
      <h2>${t("1. Choisissez vos langues")}</h2>
      <div class="field" style="align-items:center">
        <span>${t("Interface :")}</span>
        <div class="agepick">
          <button class="${state.lang === "fr" ? "sel" : ""}" onclick="_obLang('fr')">Français</button>
          <button class="${state.lang === "en" ? "sel" : ""}" onclick="_obLang('en')">English</button>
        </div>
      </div>
      <div class="field" style="align-items:center">
        <span>${t("Affiches & synopsis :")}</span>
        <div class="agepick" id="ob-meta">
          <button class="${(state.obMeta || (state.lang === "en" ? "en-US" : "fr-FR")) === "fr-FR" ? "sel" : ""}" onclick="_obMeta('fr-FR')">Français</button>
          <button class="${(state.obMeta || (state.lang === "en" ? "en-US" : "fr-FR")) === "en-US" ? "sel" : ""}" onclick="_obMeta('en-US')">English</button>
        </div>
      </div>
    </div>
    <div class="block">
      <h2>${t("2. Collez votre clé TMDB gratuite")}</h2>
      <p class="help">${t("Les affiches et synopsis viennent de The Movie Database. Créez un compte gratuit (2 min), puis copiez la « clé d'API » depuis")}
        <a href="https://www.themoviedb.org/settings/api" target="_blank">${t("Paramètres → API")}</a>.</p>
      <div class="field">
        <input type="password" id="ob-key" placeholder="${t("Votre clé API TMDB")}">
      </div>
      <div class="field">
        <button class="play" style="width:100%" onclick="_obGo()">${t("Lancer le scan 🎬")}</button>
      </div>
      <p class="help">${t("La clé reste sur votre disque, rien n'est envoyé ailleurs.")}
        ${t("Pensez aussi à installer VLC pour la lecture :")} <a href="https://www.videolan.org" target="_blank">videolan.org</a></p>
    </div>
  </div>`;
}

window._obLang = (l) => { state.lang = l; localStorage.setItem("ac_lang", l); updateChrome(); render(); };
window._obMeta = (m) => { state.obMeta = m; render(); };
window._obGo = async () => {
  const key = $("#ob-key").value.trim();
  if (!key) { toast(t("Collez d'abord votre clé.")); return; }
  try {
    await api("/api/config", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        tmdb_api_key: key,
        ui_lang: state.lang,
        metadata_lang: state.obMeta || (state.lang === "en" ? "en-US" : "fr-FR"),
      }),
    });
    toast(t("Clé validée et enregistrée ✓"));
    state.lib.has_key = true;
    startScan();
    app.innerHTML = `<div class="empty"><div class="big">🎬</div></div>`;
  } catch (e) { toast(e.message, 6000); }
};

/* ---------- page bibliothèque ---------- */

function collectGenres(items) {
  const s = new Set();
  items.forEach(m => (m.genres || []).forEach(g => s.add(g)));
  return [...s].sort();
}

function itemDecade(y) { return y ? Math.floor(y / 10) * 10 : 0; }

function renderHome() {
  const isFilms = state.tab === "films";
  const items = isFilms ? (state.lib.movies || []) : (state.lib.series || []);

  if (!(state.lib.movies || []).length && !(state.lib.series || []).length) {
    if (!state.lib.has_key && !state.lib.kid_mode) { renderOnboarding(); return; }
    app.innerHTML = `<div class="empty">
      <div class="big">🎬</div>
      <p>${t("La bibliothèque est vide.")}</p>
      <p style="margin:14px 0"><button class="play" onclick="_scan()">${t("Scanner le disque")}</button></p>
    </div>`;
    return;
  }

  const genres = collectGenres(items);
  const decades = [...new Set(items.map(m => itemDecade(m.year)).filter(Boolean))].sort((a, b) => b - a);

  let filtered = items.filter(m => {
    if (state.filters.genre && !(m.genres || []).includes(state.filters.genre)) return false;
    if (state.filters.decade && itemDecade(m.year) !== +state.filters.decade) return false;
    if (state.filters.lang) {
      const b = isFilms ? bestBadge(m.versions) : bestBadge((m.seasons || []).flatMap(s => (s.episodes || []).flatMap(e => e.files || [])));
      if (b !== state.filters.lang) return false;
    }
    if (state.search) {
      const q = state.search.toLowerCase();
      if (!(m.title_fr || "").toLowerCase().includes(q) && !(m.original_title || "").toLowerCase().includes(q)) return false;
    }
    return true;
  });

  const sort = state.filters.sort;
  filtered.sort((a, b) =>
    sort === "annee" ? (b.year || 0) - (a.year || 0) :
    sort === "note" ? (b.vote_average || 0) - (a.vote_average || 0) :
    (a.title_fr || "").localeCompare(b.title_fr || "", locale()));

  const nbReview = items.filter(m => m.needs_review).length;

  app.innerHTML = `
  <div class="filters">
    <div class="tabs">
      <button class="${isFilms ? "active" : ""}" onclick="_setTab('films')">${t("Films")} (${(state.lib.movies || []).length})</button>
      <button class="${!isFilms ? "active" : ""}" onclick="_setTab('series')">${t("Séries")} (${(state.lib.series || []).length})</button>
    </div>
    <select onchange="_setFilter('genre', this.value)">
      <option value="">${t("Tous les genres")}</option>
      ${genres.map(g => `<option ${state.filters.genre === g ? "selected" : ""}>${esc(g)}</option>`).join("")}
    </select>
    <select onchange="_setFilter('lang', this.value)">
      <option value="">${t("Toutes langues")}</option>
      ${["MULTI", "VF", "VOSTFR", "VO"].map(l => `<option ${state.filters.lang === l ? "selected" : ""}>${l}</option>`).join("")}
    </select>
    <select onchange="_setFilter('decade', this.value)">
      <option value="">${t("Toutes décennies")}</option>
      ${decades.map(d => `<option value="${d}" ${state.filters.decade == d ? "selected" : ""}>${d}s</option>`).join("")}
    </select>
    <select onchange="_setFilter('sort', this.value)">
      <option value="titre" ${sort === "titre" ? "selected" : ""}>${t("Tri : titre")}</option>
      <option value="annee" ${sort === "annee" ? "selected" : ""}>${t("Tri : année")}</option>
      <option value="note" ${sort === "note" ? "selected" : ""}>${t("Tri : note")}</option>
    </select>
    <span class="count">${filtered.length} ${isFilms ? t("films") : t("séries")}${nbReview && !state.lib.kid_mode ? ` · <a href="#/reglages" style="color:var(--accent2)">${nbReview} ${t("à vérifier")}</a>` : ""}</span>
  </div>
  <div class="grid">
    ${filtered.map(m => isFilms ? movieCard(m) : seriesCard(m)).join("")}
  </div>`;
}

function movieCard(m) {
  const badge = bestBadge(m.versions);
  const multi = (m.versions || []).length > 1;
  return `<div class="card" onclick="location.hash='#/film/${m.id}'">
    <div class="flags">${badgeHTML(badge)}${ageBadge(m.age_rating)}${m.needs_review ? `<span class="badge warn">?</span>` : ""}${multi ? `<span class="badge res">×${m.versions.length}</span>` : ""}</div>
    ${posterImg(m.poster)}
    <div class="title">${esc(m.title_fr)}</div>
    <div class="sub">${m.year || ""}${m.vote_average ? ` · ★ ${m.vote_average.toFixed(1)}` : ""}</div>
  </div>`;
}

function seriesCard(s) {
  const nbEp = (s.seasons || []).reduce((n, se) => n + (se.episodes || []).length, 0);
  return `<div class="card" onclick="location.hash='#/serie/${s.id}'">
    <div class="flags">${ageBadge(s.age_rating)}${s.misplaced ? `<span class="badge warn">${t("égarée")}</span>` : ""}</div>
    ${posterImg(s.poster)}
    <div class="title">${esc(s.title_fr)}</div>
    <div class="sub">${(s.seasons || []).length} ${t("saison(s)")} · ${nbEp} ${t("ép.")}</div>
  </div>`;
}

window._setTab = tab => { state.tab = tab; state.filters = { genre: "", lang: "", decade: "", sort: "titre" }; render(); };
window._setFilter = (k, v) => { state.filters[k] = v; render(); };
window._scan = startScan;

/* ---------- détail film ---------- */

function versionRow(v) {
  const subs = v.external_subs || [];
  return `<div class="vrow" data-path="${esc(v.path)}" data-parts='${esc(JSON.stringify(v.parts || []))}'>
    ${badgeHTML(v.lang_badge)}
    ${v.resolution_tag ? `<span class="badge res">${esc(v.resolution_tag)}</span>` : ""}
    ${v.is_iso ? `<span class="badge iso">ISO</span>` : ""}
    ${v.edition ? `<span class="badge res">${esc(v.edition)}</span>` : ""}
    <span class="fname">${esc(v.raw_name)}${(v.parts || []).length ? ` (+${v.parts.length} ${t("partie(s)")})` : ""}</span>
    <span class="size">${fmtSize(v.size_bytes)}</span>
    ${subs.length ? `<select title="${t("Sous-titres")}">
        <option value="">${t("Sous-titres : aucun")}</option>
        ${subs.map(s => `<option value="${esc(s.path)}">${esc(s.lang ? s.lang.toUpperCase() + " — " : "")}${esc(s.path.split("/").pop())}</option>`).join("")}
      </select>` : ""}
    <button class="play" onclick="_playVersion(this)">${t("▶ Lire")}</button>
  </div>`;
}

function renderMovie(id) {
  const m = (state.lib.movies || []).find(x => x.id === id);
  if (!m) { app.innerHTML = `<div class="empty">${t("Film introuvable.")}</div>`; return; }
  app.innerHTML = `
  <a class="back" href="#/">← ${t("Bibliothèque")}</a>
  <div class="detail-hero">
    ${m.backdrop ? `<div class="backdrop" style="background-image:url('/media-cache/${esc(m.backdrop)}')"></div>` : ""}
    <div class="detail-inner">
      ${posterImg(m.poster)}
      <div class="detail-meta">
        <h1>${esc(m.title_fr)}</h1>
        ${m.original_title && m.original_title !== m.title_fr ? `<div class="orig">${esc(m.original_title)}</div>` : ""}
        <div class="line">
          ${m.year ? `<span>${m.year}</span>` : ""}
          ${m.runtime_min ? `<span>${Math.floor(m.runtime_min / 60)} h ${String(m.runtime_min % 60).padStart(2, "0")}</span>` : ""}
          ${m.vote_average ? `<span class="note">★ ${m.vote_average.toFixed(1)}</span>` : ""}
          ${ageBadge(m.age_rating)}
          ${(m.genres || []).map(g => `<span class="badge res">${esc(g)}</span>`).join("")}
          ${m.needs_review ? `<span class="badge warn">${t("match incertain — corrigez dans Réglages")}</span>` : ""}
        </div>
        <p class="overview">${esc(m.overview_fr || t("Pas de synopsis."))}</p>
        ${state.lib.kid_mode ? "" : `<div style="margin-top:14px">${ratingPicker("movie", m.tmdb_id, m.age_rating)}</div>`}
      </div>
    </div>
  </div>
  <div class="block">
    <h2>${t("Fichier")}${(m.versions || []).length > 1 ? "s (" + m.versions.length + " " + t("versions — voir Doublons") + ")" : ""}</h2>
    ${(m.versions || []).map(versionRow).join("")}
  </div>
  ${state.lib.kid_mode ? "" : fixBlock(m.title_fr, "movie", m.versions?.[0]?.path)}`;
}

/* bloc repliable « mauvaise fiche ? » présent sur chaque page de détail */
function fixBlock(title, type, path) {
  if (!path) return "";
  return `<details class="block">
    <summary style="cursor:pointer;color:var(--muted)">${t("Mauvaise fiche ? Corriger le film/la série associé(e)…")}</summary>
    <div class="review-item" style="margin-top:12px">
      <div class="field">
        <input type="text" placeholder="${t("Chercher le bon titre sur TMDB…")}" value="${esc(title)}"
          onkeydown="if(event.key==='Enter')_fixSearch(this, '${type}')">
        <button class="ghost" onclick="_fixSearch(this.previousElementSibling, '${type}')">${t("Chercher")}</button>
      </div>
      <div class="fixresults" data-path="${esc(path)}"></div>
    </div>
  </details>`;
}

/* ---------- détail série ---------- */

function renderSeries(id) {
  const s = (state.lib.series || []).find(x => x.id === id);
  if (!s) { app.innerHTML = `<div class="empty">${t("Série introuvable.")}</div>`; return; }
  app.innerHTML = `
  <a class="back" href="#/">← ${t("Bibliothèque")}</a>
  <div class="detail-hero">
    ${s.backdrop ? `<div class="backdrop" style="background-image:url('/media-cache/${esc(s.backdrop)}')"></div>` : ""}
    <div class="detail-inner">
      ${posterImg(s.poster)}
      <div class="detail-meta">
        <h1>${esc(s.title_fr)}</h1>
        ${s.original_title && s.original_title !== s.title_fr ? `<div class="orig">${esc(s.original_title)}</div>` : ""}
        <div class="line">
          ${s.year ? `<span>${s.year}</span>` : ""}
          ${s.vote_average ? `<span class="note">★ ${s.vote_average.toFixed(1)}</span>` : ""}
          ${ageBadge(s.age_rating)}
          ${(s.genres || []).map(g => `<span class="badge res">${esc(g)}</span>`).join("")}
          ${s.misplaced ? `<span class="badge warn">${t("des fichiers sont dans [Films]")}</span>` : ""}
        </div>
        <p class="overview">${esc(s.overview_fr || t("Pas de synopsis."))}</p>
        ${state.lib.kid_mode ? "" : `<div style="margin-top:14px">${ratingPicker("tv", s.tmdb_id, s.age_rating)}</div>`}
      </div>
    </div>
  </div>
  ${(s.seasons || []).map((se, i) => `
  <details class="season block" ${i === 0 ? "open" : ""}>
    <summary>${esc(se.name_fr || t("Saison") + " " + se.number)} <span class="size">${(se.episodes || []).length} ${t("épisode(s)")}</span></summary>
    <div class="eps">
      ${(se.episodes || []).map(ep => epRow(ep)).join("")}
    </div>
  </details>`).join("")}
  ${state.lib.kid_mode ? "" : fixBlock(s.title_fr, "tv", s.folders?.[0])}`;
}

function epRow(ep) {
  const f = (ep.files || [])[0];
  if (!f) return "";
  const dup = (ep.files || []).length > 1;
  return `<div class="eprow" data-path="${esc(f.path)}" data-parts="[]">
    <span class="epnum">S${String(ep.season).padStart(2, "0")}E${String(ep.episode).padStart(2, "0")}</span>
    <div style="flex:1;min-width:0">
      <div class="epname">${esc(ep.name_fr || t("Épisode") + " " + ep.episode)} ${badgeHTML(f.lang_badge)} ${dup ? `<span class="badge warn">×${ep.files.length}</span>` : ""}</div>
      ${ep.overview_fr ? `<div class="epover">${esc(ep.overview_fr)}</div>` : ""}
    </div>
    ${(f.external_subs || []).length ? `<select title="${t("Sous-titres")}">
        <option value="">${t("ST : aucun")}</option>
        ${f.external_subs.map(s => `<option value="${esc(s.path)}">${esc((s.lang || "?").toUpperCase())} ${esc(s.path.split("/").pop().slice(0, 40))}</option>`).join("")}
      </select>` : ""}
    <button class="play" onclick="_playVersion(this)">▶</button>
  </div>`;
}

/* ---------- doublons ---------- */

async function renderDupes() {
  app.innerHTML = `<div class="empty">${t("Chargement…")}</div>`;
  let d;
  try { d = await api("/api/duplicates"); } catch (e) { app.innerHTML = `<div class="empty">${esc(e.message)}</div>`; return; }
  const trash = await api("/api/trash").catch(() => []);

  const movieGroups = d.movies || [];
  const epDups = d.episodes || [];
  const misplaced = d.misplaced_series || [];

  app.innerHTML = `
  <h1 style="margin-bottom:18px">${t("Doublons")}</h1>
  ${state.lib.readonly ? `<div class="block" style="border-left:3px solid var(--accent2)">${t("Disque en lecture seule : la suppression est désactivée.")}</div>` : ""}
  ${!movieGroups.length && !epDups.length && !misplaced.length ? `<div class="empty"><div class="big">✨</div>${t("Aucun doublon détecté.")}</div>` : ""}

  ${movieGroups.map(g => `
  <div class="block">
    <h2>${esc(g.title)} ${g.year ? `(${g.year})` : ""} — ${g.versions.length} versions</h2>
    <table>
      <tr><th></th><th>${t("Fichier")}</th><th>${t("Résolution")}</th><th>${t("Langue")}</th><th>ST</th><th>${t("Taille")}</th><th></th></tr>
      ${g.versions.map(v => {
        const keep = v.path === g.keep_path;
        return `<tr class="${keep ? "keep" : "del"}">
          <td class="reco ${keep ? "keep" : "del"}">${keep ? t("✓ garder") : t("✗ supprimer ?")}</td>
          <td style="word-break:break-all;font-size:12.5px">${esc(v.path)}</td>
          <td>${esc(v.resolution_tag || "?")}${v.is_iso ? " ISO" : ""}</td>
          <td>${badgeHTML(v.lang_badge) || "?"}</td>
          <td>${(v.external_subs || []).length || ""}</td>
          <td>${fmtSize(v.size_bytes)}</td>
          <td>${keep ? "" : `<button class="ghost" ${state.lib.readonly ? "disabled" : ""} onclick='_confirmTrash(${JSON.stringify(JSON.stringify([v.path]))})'>${t("🗑 Corbeille")}</button>`}</td>
        </tr>`;
      }).join("")}
    </table>
  </div>`).join("")}

  ${epDups.length ? `<div class="block">
    <h2>${t("Épisodes en double")}</h2>
    ${epDups.map(e => `
      <div style="margin-bottom:14px">
        <strong>${esc(e.title)} S${String(e.season).padStart(2, "0")}E${String(e.episode).padStart(2, "0")}</strong>
        <table>
        ${e.files.map(f => `<tr>
          <td style="word-break:break-all;font-size:12.5px">${esc(f.path)}</td>
          <td>${esc(f.resolution_tag || "?")}</td><td>${badgeHTML(f.lang_badge) || "?"}</td>
          <td>${fmtSize(f.size_bytes)}</td>
          <td><button class="ghost" ${state.lib.readonly ? "disabled" : ""} onclick='_confirmTrash(${JSON.stringify(JSON.stringify([f.path]))})'>🗑</button></td>
        </tr>`).join("")}
        </table>
      </div>`).join("")}
  </div>` : ""}

  ${misplaced.length ? `<div class="block">
    <h2>${t("Séries égarées dans [Films]")}</h2>
    <p class="help">${t("Ces séries ont des fichiers rangés dans le dossier [Films]. Elles apparaissent quand même dans l'onglet Séries ; à déplacer à la main si vous voulez ranger le disque.")}</p>
    <ul style="margin:10px 0 0 20px">${misplaced.map(x => `<li>${esc(x)}</li>`).join("")}</ul>
  </div>` : ""}

  ${trash.length ? `<div class="block">
    <h2>${t("Corbeille")} (${trash.length})</h2>
    <table>
      ${trash.map(x => `<tr>
        <td style="word-break:break-all;font-size:12.5px">${esc(x.original_path)}</td>
        <td>${fmtSize(x.size_bytes)}</td>
        <td>${new Date(x.deleted_at).toLocaleDateString(locale())}</td>
        <td><button class="ghost" onclick='_restore(${JSON.stringify(JSON.stringify(x.trash_path))})'>${t("↩ Restaurer")}</button></td>
      </tr>`).join("")}
    </table>
    <p class="help" style="margin-top:8px">${t("Les fichiers sont déplacés dans .absolute_cinema/corbeille/ sur le disque — rien n'est effacé définitivement.")}</p>
    <div class="field"><button class="danger" onclick="_confirmEmptyTrash()">${t("Vider la corbeille…")}</button></div>
  </div>` : ""}`;
}

window._confirmTrash = (pathsJSON) => {
  const paths = JSON.parse(pathsJSON);
  const root = $("#modal-root");
  root.innerHTML = `<div class="modal-back" onclick="if(event.target===this)this.remove()">
    <div class="modal">
      <h3>${t("Mettre à la corbeille ?")}</h3>
      <p class="help">${t("Le fichier sera déplacé vers .absolute_cinema/corbeille/ (restaurable) :")}</p>
      <ul>${paths.map(p => `<li>${esc(p)}</li>`).join("")}</ul>
      <div class="actions">
        <button class="ghost" onclick="this.closest('.modal-back').remove()">${t("Annuler")}</button>
        <button class="danger" onclick='_doTrash(${JSON.stringify(pathsJSON)})'>${t("Mettre à la corbeille")}</button>
      </div>
    </div>
  </div>`;
};

window._doTrash = async (pathsJSON) => {
  $("#modal-root").innerHTML = "";
  try {
    const res = await api("/api/trash", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ paths: JSON.parse(pathsJSON) }),
    });
    if ((res.failed || []).length) toast(t("Échec :") + " " + res.failed.join(" ; "), 7000);
    else toast(t("Déplacé en corbeille ✓"));
    await loadLibrary();
    renderDupes();
  } catch (e) { toast(e.message, 6000); }
};

window._restore = async (tpJSON) => {
  try {
    await api("/api/trash/restore", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ trash_path: JSON.parse(tpJSON) }),
    });
    toast(t("Restauré ✓ — relancez un scan pour le réintégrer"));
    renderDupes();
  } catch (e) { toast(e.message, 6000); }
};

window._confirmEmptyTrash = () => {
  const root = $("#modal-root");
  root.innerHTML = `<div class="modal-back" onclick="if(event.target===this)this.remove()">
    <div class="modal">
      <h3>${t("Vider définitivement la corbeille ?")}</h3>
      <p class="help">${t("Tous les fichiers de la corbeille seront définitivement supprimés. Cette action est irréversible.")}</p>
      <div class="actions">
        <button class="ghost" onclick="this.closest('.modal-back').remove()">${t("Annuler")}</button>
        <button class="danger" onclick="_doEmptyTrash()">${t("Vider")}</button>
      </div>
    </div>
  </div>`;
};

window._doEmptyTrash = async () => {
  $("#modal-root").innerHTML = "";
  try {
    await api("/api/trash/empty", { method: "POST" });
    toast(t("Corbeille vidée ✓"));
    renderDupes();
  } catch (e) { toast(e.message, 6000); }
};

/* ---------- réglages ---------- */

async function renderSettings() {
  let cfg = {};
  try { cfg = await api("/api/config"); } catch { }
  const reviewMovies = (state.lib.movies || []).filter(m => m.needs_review);
  const reviewSeries = (state.lib.series || []).filter(s => s.needs_review);
  const unmatched = state.lib.unmatched || [];
  const metaLang = cfg.metadata_lang || "fr-FR";

  app.innerHTML = `
  <h1 style="margin-bottom:18px">${t("Réglages")}</h1>
  ${updateBanner()}

  <div class="block">
    <h2>${t("Langue / Language")}</h2>
    <div class="field" style="align-items:center">
      <span>${t("Langue de l'interface")} :</span>
      <div class="agepick">
        <button class="${state.lang === "fr" ? "sel" : ""}" onclick="_setUILang('fr')">Français</button>
        <button class="${state.lang === "en" ? "sel" : ""}" onclick="_setUILang('en')">English</button>
      </div>
    </div>
    <div class="field" style="align-items:center">
      <span>${t("Langue des affiches et synopsis")} :</span>
      <div class="agepick">
        <button class="${metaLang === "fr-FR" ? "sel" : ""}" onclick="_setMetaLang('fr-FR')">Français</button>
        <button class="${metaLang === "en-US" ? "sel" : ""}" onclick="_setMetaLang('en-US')">English</button>
      </div>
    </div>
    <p class="help">${t("Après un changement de langue des métadonnées, relancez un scan.")}</p>
  </div>

  <div class="block">
    <h2>${t("Clé API TMDB")}</h2>
    <p class="help">
      <a href="https://www.themoviedb.org" target="_blank">The Movie Database</a> —
      <a href="https://www.themoviedb.org/settings/api" target="_blank">${t("Paramètres → API")}</a>.</p>
    <div class="field">
      <input type="password" id="tmdbkey" placeholder="${cfg.has_key ? "(" + esc(cfg.tmdb_key_masked) + ")" : t("Votre clé API TMDB")}">
      <button class="play" onclick="_saveKey()">${t("Enregistrer")}</button>
    </div>
  </div>

  <div class="block">
    <h2>${t("Scan de la bibliothèque")}</h2>
    <p class="help">
      ${t("Médiathèque :")} <code>${esc(cfg.media_root || "")}</code><br>
      ${t("Données :")} <code>${esc(cfg.data_dir || "")}</code>${cfg.readonly ? ` <span class="badge warn">${t("lecture seule")}</span>` : ""}<br>
      ${t("Dernier scan :")} ${state.lib.scanned_at && state.lib.scanned_at !== "0001-01-01T00:00:00Z" ? new Date(state.lib.scanned_at).toLocaleString(locale()) : t("jamais")}
    </p>
    <div class="field">
      <button class="play" onclick="_scan()">${state.lib.movies?.length ? t("Re-scanner le disque") : t("Scanner le disque")}</button>
    </div>
  </div>

  ${kidSettingsBlock()}

  ${reviewMovies.length || reviewSeries.length ? `<div class="block">
    <h2>${t("À vérifier")} (${reviewMovies.length + reviewSeries.length})</h2>
    <p class="help">${t("Le match TMDB est incertain pour ces fiches. Cherchez le bon titre et cliquez dessus pour corriger.")}</p>
    ${reviewMovies.map(m => reviewItem(m, "movie", m.versions?.[0]?.path)).join("")}
    ${reviewSeries.map(s => reviewItem(s, "tv", s.folders?.[0])).join("")}
  </div>` : ""}

  ${unmatched.length ? `<div class="block">
    <h2>${t("Non identifiés")} (${unmatched.length})</h2>
    <p class="help">${t("Aucune fiche TMDB trouvée pour ces fichiers. Cherchez manuellement, ou ignorez (concerts, bonus, documents…).")}</p>
    ${unmatched.map(u => `
    <div class="review-item">
      <div><strong>${esc(u.parsed_title || t("(sans titre)"))}</strong> ${u.year ? `(${u.year})` : ""} <span class="size">${fmtSize(u.size_bytes)}</span></div>
      <div class="path">${esc(u.path)}</div>
      <div class="field">
        <input type="text" placeholder="${t("Chercher sur TMDB…")}" value="${esc(u.parsed_title || "")}"
          onkeydown="if(event.key==='Enter')_fixSearch(this, 'movie')">
        <button class="ghost" onclick="_fixSearch(this.previousElementSibling, 'movie')">${t("Film…")}</button>
        <button class="ghost" onclick="_fixSearch(this.previousElementSibling.previousElementSibling, 'tv')">${t("Série…")}</button>
      </div>
      <div class="fixresults" data-path="${esc(u.path)}"></div>
    </div>`).join("")}
  </div>` : ""}

  <div class="block">
    <h2>${t("À propos")}</h2>
    <p class="help">
      <strong>absolute cinema</strong> — version ${APP_VERSION} ·
      <a href="https://github.com/${GITHUB_REPO}" target="_blank">GitHub</a> ·
      <a href="https://bist0uille.github.io/Absolute_cinema/" target="_blank">Site</a><br><br>
      ${state.lang === "en"
        ? `This software organizes files you already own; it does not include, provide or download any movie.<br><br>
           Metadata and artwork by <a href="https://www.themoviedb.org" target="_blank">The Movie Database (TMDB)</a>.
           This product uses the TMDB API but is not endorsed or certified by TMDB.
           Playback via <a href="https://www.videolan.org" target="_blank">VLC</a> (independent software, installed separately).`
        : `Ce logiciel organise des fichiers que vous possédez déjà ; il n'inclut, ne fournit
           et ne télécharge aucun film.<br><br>
           Métadonnées et affiches fournies par
           <a href="https://www.themoviedb.org" target="_blank">The Movie Database (TMDB)</a>.
           Ce produit utilise l'API TMDB sans être approuvé ni certifié par TMDB.
           Lecture vidéo assurée par <a href="https://www.videolan.org" target="_blank">VLC</a>
           (logiciel indépendant, à installer séparément).`}
    </p>
  </div>`;
}

window._setUILang = async (l) => {
  state.lang = l;
  localStorage.setItem("ac_lang", l);
  try {
    await api("/api/config", { method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ui_lang: l }) });
  } catch { }
  state.lib.ui_lang = l;
  updateChrome();
  render();
};

window._setMetaLang = async (m) => {
  try {
    await api("/api/config", { method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ metadata_lang: m }) });
    toast("✓ — " + t("Après un changement de langue des métadonnées, relancez un scan."), 5000);
    render();
  } catch (e) { toast(e.message, 5000); }
};

function kidSettingsBlock() {
  const unratedM = (state.lib.movies || []).filter(m => m.age_rating < 0);
  const unratedS = (state.lib.series || []).filter(s => s.age_rating < 0);
  const nbUnrated = unratedM.length + unratedS.length;
  return `<div class="block">
    <h2>${t("🧸 Mode enfant")}</h2>
    <p class="help">${t("Le mode enfant ne montre (et ne lit) que les titres classés jusqu'à l'âge choisi — classification officielle TMDB. Les titres sans classification sont cachés par prudence. La sortie du mode enfant demande un code PIN.")}
      ${state.lib.has_pin ? t("Un code PIN est défini.") : t("Le code PIN sera défini à la première activation.")}</p>
    <div class="field">
      <button class="play" onclick="_kidToggle()">${t("Activer le mode enfant…")}</button>
    </div>
    ${nbUnrated ? `
    <details style="margin-top:10px">
      <summary style="cursor:pointer;color:var(--accent)">${nbUnrated} ${t("titre(s) sans classification — classer manuellement")}</summary>
      <p class="help" style="margin:10px 0">${t("Cliquez sur un âge pour classer le titre (mémorisé, survit aux re-scans). Sans classement, le titre reste invisible en mode enfant.")}</p>
      ${[...unratedM.map(m => ({ ty: "movie", x: m })), ...unratedS.map(s => ({ ty: "tv", x: s }))].map(({ ty, x }) => `
      <div class="vrow">
        <span style="flex:1">${esc(x.title_fr)} ${x.year ? `(${x.year})` : ""} <span class="badge res">${ty === "tv" ? t("série") : t("film")}</span></span>
        ${ratingPicker(ty, x.tmdb_id, x.age_rating)}
      </div>`).join("")}
    </details>` : `<p class="help">${t("Tous les titres de la bibliothèque ont une classification ✓")}</p>`}
  </div>`;
}

function reviewItem(m, type, path) {
  return `
  <div class="review-item">
    <div><strong>${esc(m.title_fr)}</strong> ${m.year ? `(${m.year})` : ""}
      <span class="size">${t("confiance")} ${(m.match_confidence * 100).toFixed(0)} %</span>
      <span class="badge res">${type === "tv" ? t("série") : t("film")}</span></div>
    <div class="path">${esc(path || "")}</div>
    <div class="field">
      <input type="text" placeholder="${t("Chercher le bon titre…")}" value="${esc(m.title_fr)}"
        onkeydown="if(event.key==='Enter')_fixSearch(this, '${type}')">
      <button class="ghost" onclick="_fixSearch(this.previousElementSibling, '${type}')">${t("Chercher")}</button>
    </div>
    <div class="fixresults" data-path="${esc(path || "")}"></div>
  </div>`;
}

window._fixSearch = async (input, type) => {
  const item = input.closest(".review-item");
  const box = $(".fixresults", item);
  box.innerHTML = `<p class="help">${t("Recherche…")}</p>`;
  try {
    const rs = await api(`/api/tmdb/search?type=${type}&q=` + encodeURIComponent(input.value));
    if (!rs.length) { box.innerHTML = `<p class="help">${t("Aucun résultat.")}</p>`; return; }
    box.innerHTML = rs.map(r => `
      <div class="fr-item" onclick='_applyFix(this, ${r.id}, "${type}")'>
        ${r.poster_path ? `<img src="https://image.tmdb.org/t/p/w92${esc(r.poster_path)}">` : `<img>`}
        <div><strong>${esc(r.title || r.name)}</strong>
          <span class="size">${(r.release_date || r.first_air_date || "").slice(0, 4)}</span><br>
          <span class="help">${esc((r.overview || "").slice(0, 110))}…</span></div>
      </div>`).join("");
  } catch (e) { box.innerHTML = `<p class="help">${esc(e.message)}</p>`; }
};

window._applyFix = async (el, tmdbID, type) => {
  const path = el.closest(".fixresults").dataset.path;
  try {
    await api("/api/override", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path, media_type: type, tmdb_id: tmdbID }),
    });
    toast(t("Correction enregistrée — re-scan en cours…"));
    pollScan();
  } catch (e) { toast(e.message, 6000); }
};

window._saveKey = async () => {
  const key = $("#tmdbkey").value.trim();
  if (!key) { toast(t("Collez d'abord votre clé.")); return; }
  try {
    await api("/api/config", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tmdb_api_key: key }),
    });
    toast(t("Clé validée et enregistrée ✓"));
    state.lib.has_key = true;
    if (!(state.lib.movies || []).length) startScan();
    renderSettings();
  } catch (e) { toast(e.message, 6000); }
};

/* ---------- recherche globale ---------- */

$("#search").addEventListener("input", e => {
  state.search = e.target.value;
  if (route().page !== "home") location.hash = "#/";
  else render();
});

/* ---------- démarrage ---------- */

loadLibrary().catch(e => {
  app.innerHTML = `<div class="empty">${t("Erreur de chargement :")} ${esc(e.message)}</div>`;
});
checkUpdate();
