/* absolute_cinema — interface (vanilla JS, routing par hash) */
"use strict";

const APP_VERSION = "1.1.0";

const $ = (sel, el = document) => el.querySelector(sel);
const app = $("#app");

const state = {
  lib: { movies: [], series: [], unmatched: [] },
  tab: "films",
  filters: { genre: "", lang: "", decade: "", sort: "titre" },
  search: "",
  scanning: false,
};

/* ---------- utilitaires ---------- */

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

function fmtSize(b) {
  if (!b) return "";
  const gb = b / (1 << 30);
  return gb >= 1 ? gb.toFixed(1) + " Go" : Math.round(b / (1 << 20)) + " Mo";
}

function toast(msg, ms = 3500) {
  const t = $("#toast");
  t.textContent = msg;
  t.classList.remove("hidden");
  clearTimeout(t._h);
  t._h = setTimeout(() => t.classList.add("hidden"), ms);
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
  if (a === undefined || a === null || a < 0) return `<span class="badge ageunk" title="Classification inconnue">?</span>`;
  const cls = a >= 18 ? "age18" : a >= 16 ? "age16" : a >= 12 ? "age12" : a >= 10 ? "age10" : "age0";
  return `<span class="badge ${cls}" title="Âge minimum">${a === 0 ? "TP" : a}</span>`;
}

/* ---------- mode enfant ---------- */

function updateChrome() {
  const kid = state.lib.kid_mode;
  document.querySelector('[data-nav="doublons"]').style.display = kid ? "none" : "";
  document.querySelector('[data-nav="reglages"]').style.display = kid ? "none" : "";
  const btn = $("#kidbtn");
  btn.classList.toggle("on", !!kid);
  btn.textContent = kid ? "🔓" : "🧸";
  btn.title = kid ? "Quitter le mode enfant (code PIN)" : "Activer le mode enfant";
  let banner = $("#kidbanner");
  if (kid && !banner) {
    banner = document.createElement("div");
    banner.id = "kidbanner";
    banner.textContent = `🧸 Mode enfant — seuls les titres jusqu'à ${state.lib.kid_max_age} ans sont visibles`;
    $("#topbar").after(banner);
  } else if (!kid && banner) banner.remove();
}

window._kidToggle = () => {
  if (state.lib.kid_mode) {
    pinModal("Quitter le mode enfant", "Entrez le code PIN parent :", async (pin) => {
      await api("/api/kidmode", { method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enable: false, pin }) });
      toast("Mode enfant désactivé");
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
        <button class="ghost" onclick="this.closest('.modal-back').remove()">Annuler</button>
        <button class="play" id="pin-ok">Valider</button>
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
      <h3>🧸 Activer le mode enfant</h3>
      <p class="help">Seuls les films et séries classés jusqu'à l'âge choisi seront visibles et lisibles.
      Les titres sans classification sont cachés par prudence (classez-les dans Réglages).
      Pour ressortir du mode enfant, il faudra le code PIN.</p>
      <div class="field" style="align-items:center">
        <span>Âge maximum :</span>
        <div class="agepick" id="agepick">
          ${[6, 10, 12, 16].map(a => `<button data-age="${a}" class="${a === curAge ? "sel" : ""}">${a} ans</button>`).join("")}
        </div>
      </div>
      ${hasPin ? "" : `<p class="help">Choisissez un code PIN parent (au moins 4 chiffres) :</p>
      <input class="pin" type="password" inputmode="numeric" maxlength="8" placeholder="••••">`}
      <div class="actions">
        <button class="ghost" onclick="this.closest('.modal-back').remove()">Annuler</button>
        <button class="play" id="kid-ok">Activer</button>
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
      toast("Mode enfant activé 🧸");
      location.hash = "#/";
      await loadLibrary();
    } catch (e) { toast(e.message, 5000); }
  };
}

window._setRating = async (mediaType, tmdbID, age, el) => {
  try {
    await api("/api/rating", { method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ media_type: mediaType, tmdb_id: tmdbID, age }) });
    toast(age < 0 ? "Classification effacée" : `Classé ${age === 0 ? "tous publics" : age + " ans"} ✓`);
    await loadLibrary();
  } catch (e) { toast(e.message, 5000); }
};

function ratingPicker(mediaType, tmdbID, current) {
  return `<div class="agepick">
    <span class="help">Classification :</span>
    ${[[0, "TP"], [10, "10"], [12, "12"], [16, "16"], [18, "18"]].map(([a, l]) =>
      `<button class="${current === a ? "sel" : ""}" onclick="_setRating('${mediaType}', ${tmdbID}, ${a}, this)">${l}</button>`).join("")}
    ${current >= 0 ? `<button onclick="_setRating('${mediaType}', ${tmdbID}, -1, this)" title="Revenir à la classification TMDB">↺</button>` : ""}
  </div>`;
}

/* ---------- chargement ---------- */

async function loadLibrary() {
  const data = await api("/api/library");
  state.lib = data;
  state.scanning = data.scanning;
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
      const phases = { walk: "Parcours du disque", tmdb: "Films — récupération TMDB", series: "Séries — récupération TMDB" };
      if (st.scanning) {
        const pct = st.total ? Math.round(100 * st.done / st.total) : 0;
        label.textContent = `${phases[st.phase] || st.phase}… ${st.total ? `${st.done}/${st.total} (${pct} %)` : ""} ${st.current || ""}`;
        fill.style.width = (st.total ? pct : 5) + "%";
      } else {
        clearInterval(pollTimer); pollTimer = null;
        bar.classList.add("hidden");
        if (st.error) toast("Scan : " + st.error, 8000);
        else if (st.phase === "done") { toast("Scan terminé ✓"); await loadLibrary(); }
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

/* ---------- lecture ---------- */

async function play(path, subPath, parts) {
  try {
    const res = await api("/api/play", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path, sub_path: subPath || "", parts: parts || [] }),
    });
    toast(res.warning || "Lecture lancée dans " + (res.player === "vlc" ? "VLC" : "le lecteur par défaut") + " ▶");
  } catch (e) { toast("Lecture impossible : " + e.message, 6000); }
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

  if (!items.length && !(state.lib.movies || []).length && !(state.lib.series || []).length) {
    app.innerHTML = `<div class="empty">
      <div class="big">🎬</div>
      <p>La bibliothèque est vide.</p>
      <p style="margin:14px 0">${state.lib.has_key
        ? `<button class="play" onclick="_scan()">Scanner le disque</button>`
        : `Commencez par renseigner votre clé TMDB dans les <a href="#/reglages" style="color:var(--accent)">Réglages</a>.`}</p>
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
    (a.title_fr || "").localeCompare(b.title_fr || "", "fr"));

  const nbReview = items.filter(m => m.needs_review).length;

  app.innerHTML = `
  <div class="filters">
    <div class="tabs">
      <button class="${isFilms ? "active" : ""}" onclick="_setTab('films')">Films (${(state.lib.movies || []).length})</button>
      <button class="${!isFilms ? "active" : ""}" onclick="_setTab('series')">Séries (${(state.lib.series || []).length})</button>
    </div>
    <select onchange="_setFilter('genre', this.value)">
      <option value="">Tous les genres</option>
      ${genres.map(g => `<option ${state.filters.genre === g ? "selected" : ""}>${esc(g)}</option>`).join("")}
    </select>
    <select onchange="_setFilter('lang', this.value)">
      <option value="">Toutes langues</option>
      ${["MULTI", "VF", "VOSTFR", "VO"].map(l => `<option ${state.filters.lang === l ? "selected" : ""}>${l}</option>`).join("")}
    </select>
    <select onchange="_setFilter('decade', this.value)">
      <option value="">Toutes décennies</option>
      ${decades.map(d => `<option value="${d}" ${state.filters.decade == d ? "selected" : ""}>${d}s</option>`).join("")}
    </select>
    <select onchange="_setFilter('sort', this.value)">
      <option value="titre" ${sort === "titre" ? "selected" : ""}>Tri : titre</option>
      <option value="annee" ${sort === "annee" ? "selected" : ""}>Tri : année</option>
      <option value="note" ${sort === "note" ? "selected" : ""}>Tri : note</option>
    </select>
    <span class="count">${filtered.length} ${isFilms ? "films" : "séries"}${nbReview ? ` · <a href="#/reglages" style="color:var(--accent2)">${nbReview} à vérifier</a>` : ""}</span>
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
    <div class="flags">${ageBadge(s.age_rating)}${s.misplaced ? `<span class="badge warn">égarée</span>` : ""}</div>
    ${posterImg(s.poster)}
    <div class="title">${esc(s.title_fr)}</div>
    <div class="sub">${(s.seasons || []).length} saison(s) · ${nbEp} ép.</div>
  </div>`;
}

window._setTab = t => { state.tab = t; state.filters = { genre: "", lang: "", decade: "", sort: "titre" }; render(); };
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
    <span class="fname">${esc(v.raw_name)}${(v.parts || []).length ? ` (+${v.parts.length} partie(s))` : ""}</span>
    <span class="size">${fmtSize(v.size_bytes)}</span>
    ${subs.length ? `<select title="Sous-titres">
        <option value="">Sous-titres : aucun</option>
        ${subs.map(s => `<option value="${esc(s.path)}">${esc(s.lang ? s.lang.toUpperCase() + " — " : "")}${esc(s.path.split("/").pop())}</option>`).join("")}
      </select>` : ""}
    <button class="play" onclick="_playVersion(this)">▶ Lire</button>
  </div>`;
}

function renderMovie(id) {
  const m = (state.lib.movies || []).find(x => x.id === id);
  if (!m) { app.innerHTML = `<div class="empty">Film introuvable.</div>`; return; }
  app.innerHTML = `
  <a class="back" href="#/">← Bibliothèque</a>
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
          ${m.needs_review ? `<span class="badge warn">match incertain — corrigez dans Réglages</span>` : ""}
        </div>
        <p class="overview">${esc(m.overview_fr || "Pas de synopsis.")}</p>
        ${state.lib.kid_mode ? "" : `<div style="margin-top:14px">${ratingPicker("movie", m.tmdb_id, m.age_rating)}</div>`}
      </div>
    </div>
  </div>
  <div class="block">
    <h2>Fichier${(m.versions || []).length > 1 ? "s (" + m.versions.length + " versions — voir Doublons)" : ""}</h2>
    ${(m.versions || []).map(versionRow).join("")}
  </div>
  ${state.lib.kid_mode ? "" : fixBlock(m.title_fr, "movie", m.versions?.[0]?.path)}`;
}

/* bloc repliable « mauvaise fiche ? » présent sur chaque page de détail */
function fixBlock(title, type, path) {
  if (!path) return "";
  return `<details class="block">
    <summary style="cursor:pointer;color:var(--muted)">Mauvaise fiche ? Corriger le film/la série associé(e)…</summary>
    <div class="review-item" style="margin-top:12px">
      <div class="field">
        <input type="text" placeholder="Chercher le bon titre sur TMDB…" value="${esc(title)}"
          onkeydown="if(event.key==='Enter')_fixSearch(this, '${type}')">
        <button class="ghost" onclick="_fixSearch(this.previousElementSibling, '${type}')">Chercher</button>
      </div>
      <div class="fixresults" data-path="${esc(path)}"></div>
    </div>
  </details>`;
}

/* ---------- détail série ---------- */

function renderSeries(id) {
  const s = (state.lib.series || []).find(x => x.id === id);
  if (!s) { app.innerHTML = `<div class="empty">Série introuvable.</div>`; return; }
  app.innerHTML = `
  <a class="back" href="#/">← Bibliothèque</a>
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
          ${s.misplaced ? `<span class="badge warn">des fichiers sont dans [Films]</span>` : ""}
        </div>
        <p class="overview">${esc(s.overview_fr || "Pas de synopsis.")}</p>
        ${state.lib.kid_mode ? "" : `<div style="margin-top:14px">${ratingPicker("tv", s.tmdb_id, s.age_rating)}</div>`}
      </div>
    </div>
  </div>
  ${(s.seasons || []).map((se, i) => `
  <details class="season block" ${i === 0 ? "open" : ""}>
    <summary>${esc(se.name_fr || "Saison " + se.number)} <span class="size">${(se.episodes || []).length} épisode(s)</span></summary>
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
      <div class="epname">${esc(ep.name_fr || "Épisode " + ep.episode)} ${badgeHTML(f.lang_badge)} ${dup ? `<span class="badge warn">×${ep.files.length}</span>` : ""}</div>
      ${ep.overview_fr ? `<div class="epover">${esc(ep.overview_fr)}</div>` : ""}
    </div>
    ${(f.external_subs || []).length ? `<select title="Sous-titres">
        <option value="">ST : aucun</option>
        ${f.external_subs.map(s => `<option value="${esc(s.path)}">${esc((s.lang || "?").toUpperCase())} ${esc(s.path.split("/").pop().slice(0, 40))}</option>`).join("")}
      </select>` : ""}
    <button class="play" onclick="_playVersion(this)">▶</button>
  </div>`;
}

/* ---------- doublons ---------- */

async function renderDupes() {
  app.innerHTML = `<div class="empty">Chargement…</div>`;
  let d;
  try { d = await api("/api/duplicates"); } catch (e) { app.innerHTML = `<div class="empty">${esc(e.message)}</div>`; return; }
  const trash = await api("/api/trash").catch(() => []);

  const movieGroups = d.movies || [];
  const epDups = d.episodes || [];
  const misplaced = d.misplaced_series || [];

  app.innerHTML = `
  <h1 style="margin-bottom:18px">Doublons</h1>
  ${state.lib.readonly ? `<div class="block" style="border-left:3px solid var(--accent2)">Disque en lecture seule : la suppression est désactivée.</div>` : ""}
  ${!movieGroups.length && !epDups.length && !misplaced.length ? `<div class="empty"><div class="big">✨</div>Aucun doublon détecté.</div>` : ""}

  ${movieGroups.map(g => `
  <div class="block">
    <h2>${esc(g.title)} ${g.year ? `(${g.year})` : ""} — ${g.versions.length} versions</h2>
    <table>
      <tr><th></th><th>Fichier</th><th>Résolution</th><th>Langue</th><th>ST</th><th>Taille</th><th></th></tr>
      ${g.versions.map(v => {
        const keep = v.path === g.keep_path;
        return `<tr class="${keep ? "keep" : "del"}">
          <td class="reco ${keep ? "keep" : "del"}">${keep ? "✓ garder" : "✗ supprimer ?"}</td>
          <td style="word-break:break-all;font-size:12.5px">${esc(v.path)}</td>
          <td>${esc(v.resolution_tag || "?")}${v.is_iso ? " ISO" : ""}</td>
          <td>${badgeHTML(v.lang_badge) || "?"}</td>
          <td>${(v.external_subs || []).length || ""}</td>
          <td>${fmtSize(v.size_bytes)}</td>
          <td>${keep ? "" : `<button class="ghost" ${state.lib.readonly ? "disabled" : ""} onclick='_confirmTrash(${JSON.stringify(JSON.stringify([v.path]))})'>🗑 Corbeille</button>`}</td>
        </tr>`;
      }).join("")}
    </table>
  </div>`).join("")}

  ${epDups.length ? `<div class="block">
    <h2>Épisodes en double</h2>
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
    <h2>Séries égarées dans [Films]</h2>
    <p class="help">Ces séries ont des fichiers rangés dans le dossier [Films]. Elles apparaissent quand même dans l'onglet Séries ; à déplacer à la main si vous voulez ranger le disque.</p>
    <ul style="margin:10px 0 0 20px">${misplaced.map(t => `<li>${esc(t)}</li>`).join("")}</ul>
  </div>` : ""}

  ${trash.length ? `<div class="block">
    <h2>Corbeille (${trash.length})</h2>
    <table>
      ${trash.map(t => `<tr>
        <td style="word-break:break-all;font-size:12.5px">${esc(t.original_path)}</td>
        <td>${fmtSize(t.size_bytes)}</td>
        <td>${new Date(t.deleted_at).toLocaleDateString("fr-FR")}</td>
        <td><button class="ghost" onclick='_restore(${JSON.stringify(JSON.stringify(t.trash_path))})'>↩ Restaurer</button></td>
      </tr>`).join("")}
    </table>
    <p class="help" style="margin-top:8px">Les fichiers sont déplacés dans <code>.absolute_cinema/corbeille/</code> sur le disque — rien n'est effacé définitivement. Videz ce dossier à la main quand vous êtes sûr.</p>
  </div>` : ""}`;
}

window._confirmTrash = (pathsJSON) => {
  const paths = JSON.parse(pathsJSON);
  const root = $("#modal-root");
  root.innerHTML = `<div class="modal-back" onclick="if(event.target===this)this.remove()">
    <div class="modal">
      <h3>Mettre à la corbeille ?</h3>
      <p class="help">Le fichier sera déplacé vers <code>.absolute_cinema/corbeille/</code> (restaurable) :</p>
      <ul>${paths.map(p => `<li>${esc(p)}</li>`).join("")}</ul>
      <div class="actions">
        <button class="ghost" onclick="this.closest('.modal-back').remove()">Annuler</button>
        <button class="danger" onclick='_doTrash(${JSON.stringify(pathsJSON)})'>Mettre à la corbeille</button>
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
    if ((res.failed || []).length) toast("Échec : " + res.failed.join(" ; "), 7000);
    else toast("Déplacé en corbeille ✓");
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
    toast("Restauré ✓ — relancez un scan pour le réintégrer");
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

  app.innerHTML = `
  <h1 style="margin-bottom:18px">Réglages</h1>

  <div class="block">
    <h2>Clé API TMDB</h2>
    <p class="help">Les affiches, synopsis et dates viennent de
      <a href="https://www.themoviedb.org" target="_blank">The Movie Database</a>.
      Créez un compte gratuit puis une clé API (v3) dans
      <a href="https://www.themoviedb.org/settings/api" target="_blank">Paramètres → API</a>.
      La clé et toutes les données sont stockées sur le disque dur.</p>
    <div class="field">
      <input type="password" id="tmdbkey" placeholder="${cfg.has_key ? "Clé enregistrée (" + esc(cfg.tmdb_key_masked) + ") — coller pour remplacer" : "Collez votre clé API TMDB ici"}">
      <button class="play" onclick="_saveKey()">Enregistrer</button>
    </div>
  </div>

  <div class="block">
    <h2>Scan de la bibliothèque</h2>
    <p class="help">
      Médiathèque : <code>${esc(cfg.media_root || "")}</code><br>
      Données : <code>${esc(cfg.data_dir || "")}</code>${cfg.readonly ? ` <span class="badge warn">lecture seule</span>` : ""}<br>
      Dernier scan : ${state.lib.scanned_at && state.lib.scanned_at !== "0001-01-01T00:00:00Z" ? new Date(state.lib.scanned_at).toLocaleString("fr-FR") : "jamais"}
    </p>
    <div class="field">
      <button class="play" onclick="_scan()">${state.lib.movies?.length ? "Re-scanner le disque" : "Scanner le disque"}</button>
    </div>
  </div>

  ${kidSettingsBlock()}

  ${reviewMovies.length || reviewSeries.length ? `<div class="block">
    <h2>À vérifier (${reviewMovies.length + reviewSeries.length})</h2>
    <p class="help">Le match TMDB est incertain pour ces fiches. Cherchez le bon titre et cliquez dessus pour corriger.</p>
    ${reviewMovies.map(m => reviewItem(m, "movie", m.versions?.[0]?.path)).join("")}
    ${reviewSeries.map(s => reviewItem(s, "tv", s.folders?.[0])).join("")}
  </div>` : ""}

  ${unmatched.length ? `<div class="block">
    <h2>Non identifiés (${unmatched.length})</h2>
    <p class="help">Aucune fiche TMDB trouvée pour ces fichiers. Cherchez manuellement, ou ignorez (concerts, bonus, documents…).</p>
    ${unmatched.map(u => `
    <div class="review-item">
      <div><strong>${esc(u.parsed_title || "(sans titre)")}</strong> ${u.year ? `(${u.year})` : ""} <span class="size">${fmtSize(u.size_bytes)}</span></div>
      <div class="path">${esc(u.path)}</div>
      <div class="field">
        <input type="text" placeholder="Chercher sur TMDB…" value="${esc(u.parsed_title || "")}"
          onkeydown="if(event.key==='Enter')_fixSearch(this, 'movie')">
        <button class="ghost" onclick="_fixSearch(this.previousElementSibling, 'movie')">Film…</button>
        <button class="ghost" onclick="_fixSearch(this.previousElementSibling.previousElementSibling, 'tv')">Série…</button>
      </div>
      <div class="fixresults" data-path="${esc(u.path)}"></div>
    </div>`).join("")}
  </div>` : ""}

  <div class="block">
    <h2>À propos</h2>
    <p class="help">
      <strong>absolute cinema</strong> — version ${APP_VERSION}. Votre vidéothèque de poche :
      la bibliothèque vit sur le disque et fonctionne hors ligne après le premier scan.<br><br>
      Ce logiciel organise des fichiers que vous possédez déjà ; il n'inclut, ne fournit
      et ne télécharge aucun film.<br><br>
      Métadonnées et affiches fournies par
      <a href="https://www.themoviedb.org" target="_blank">The Movie Database (TMDB)</a>.
      Ce produit utilise l'API TMDB sans être approuvé ni certifié par TMDB.
      Lecture vidéo assurée par <a href="https://www.videolan.org" target="_blank">VLC</a>
      (logiciel indépendant, à installer séparément).
    </p>
  </div>`;
}

function kidSettingsBlock() {
  const unratedM = (state.lib.movies || []).filter(m => m.age_rating < 0);
  const unratedS = (state.lib.series || []).filter(s => s.age_rating < 0);
  const nbUnrated = unratedM.length + unratedS.length;
  const ages = { 0: "tous publics", 10: "10 ans", 12: "12 ans", 16: "16 ans" };
  return `<div class="block">
    <h2>🧸 Mode enfant</h2>
    <p class="help">Le mode enfant ne montre (et ne lit) que les titres classés jusqu'à l'âge choisi
      — classification officielle TMDB (visas français, puis américains/britanniques).
      Les titres <strong>sans classification sont cachés</strong> par prudence.
      La sortie du mode enfant demande un code PIN.
      ${state.lib.has_pin ? "Un code PIN est défini." : "Le code PIN sera défini à la première activation."}</p>
    <div class="field">
      <button class="play" onclick="_kidToggle()">Activer le mode enfant…</button>
    </div>
    ${nbUnrated ? `
    <details style="margin-top:10px">
      <summary style="cursor:pointer;color:var(--accent)">${nbUnrated} titre(s) sans classification — classer manuellement</summary>
      <p class="help" style="margin:10px 0">Cliquez sur un âge pour classer le titre (mémorisé, survit aux re-scans). Sans classement, le titre reste invisible en mode enfant.</p>
      ${[...unratedM.map(m => ({ t: "movie", x: m })), ...unratedS.map(s => ({ t: "tv", x: s }))].map(({ t, x }) => `
      <div class="vrow">
        <span style="flex:1">${esc(x.title_fr)} ${x.year ? `(${x.year})` : ""} <span class="badge res">${t === "tv" ? "série" : "film"}</span></span>
        ${ratingPicker(t, x.tmdb_id, x.age_rating)}
      </div>`).join("")}
    </details>` : `<p class="help">Tous les titres de la bibliothèque ont une classification ✓</p>`}
  </div>`;
}

function reviewItem(m, type, path) {
  return `
  <div class="review-item">
    <div><strong>${esc(m.title_fr)}</strong> ${m.year ? `(${m.year})` : ""}
      <span class="size">confiance ${(m.match_confidence * 100).toFixed(0)} %</span>
      <span class="badge res">${type === "tv" ? "série" : "film"}</span></div>
    <div class="path">${esc(path || "")}</div>
    <div class="field">
      <input type="text" placeholder="Chercher le bon titre…" value="${esc(m.title_fr)}"
        onkeydown="if(event.key==='Enter')_fixSearch(this, '${type}')">
      <button class="ghost" onclick="_fixSearch(this.previousElementSibling, '${type}')">Chercher</button>
    </div>
    <div class="fixresults" data-path="${esc(path || "")}"></div>
  </div>`;
}

window._fixSearch = async (input, type) => {
  const item = input.closest(".review-item");
  const box = $(".fixresults", item);
  box.innerHTML = `<p class="help">Recherche…</p>`;
  try {
    const rs = await api(`/api/tmdb/search?type=${type}&q=` + encodeURIComponent(input.value));
    if (!rs.length) { box.innerHTML = `<p class="help">Aucun résultat.</p>`; return; }
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
    toast("Correction enregistrée — re-scan en cours…");
    pollScan();
  } catch (e) { toast(e.message, 6000); }
};

window._saveKey = async () => {
  const key = $("#tmdbkey").value.trim();
  if (!key) { toast("Collez d'abord votre clé."); return; }
  try {
    await api("/api/config", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tmdb_api_key: key }),
    });
    toast("Clé validée et enregistrée ✓");
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
  app.innerHTML = `<div class="empty">Erreur de chargement : ${esc(e.message)}</div>`;
});
