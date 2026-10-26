/* Native Windows host: video is decoded by LibVLC into a child HWND.
   Only the surrounding catalogue and controls run in WebView2. */
(() => {
  if (!window.acNativeAvailable) return;
  const token = new URLSearchParams(location.search).get("native") || sessionStorage.getItem("acNativeSession");
  if (!token) return;
  sessionStorage.setItem("acNativeSession", token);
  // Do not propagate the session capability through copied links/history.
  history.replaceState(null, "", location.pathname + location.hash);
  const invoke = (name, ...args) => window[name](token, ...args);
  const command = (action, value = 0) => invoke("_acNativeCommand", action, value);
  let active = null;
  let sequence = 0;
  const oldClose = window.closePlayer;
  const time = seconds => {
    seconds = Math.max(0, Math.floor(seconds || 0));
    const h = Math.floor(seconds / 3600), m = Math.floor(seconds / 60) % 60, s = seconds % 60;
    return `${h ? h + ":" : ""}${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  };
  const error = value => toast(t("Lecture impossible :") + " " + String(value), 7000);
  // A page reload must stop and hide any surface from the previous document.
  command("stop").catch(error);
  const persist = session => {
    const s = session?.status;
    if (!s || s.duration <= 0 || s.position < 0) return;
    const last = s.part + 1 === s.parts;
    saveProgress(session.version.path, s.position, s.duration,
      last && (s.state === 6 || s.position / s.duration > .92), s.part);
  };
  const bounds = () => {
    const box = document.querySelector("#native-surface");
    if (!active || !box || active.failed) return;
    const rect = box.getBoundingClientRect(), scale = window.devicePixelRatio || 1;
    invoke("_acNativeBounds", ...[rect.x, rect.y, rect.width, rect.height].map(v => Math.max(0, Math.round(v * scale)))).catch(error);
  };
  const revealControls = () => {
    const session = active;
    if (!session?.fullscreen || session.failed) return;
    clearTimeout(session.hideTimer);
    document.querySelector(".native-player")?.classList.remove("controls-hidden");
    bounds();
    const hide = () => {
      if (active !== session || !session.fullscreen || session.failed) return;
      if (session.status?.state === 4 || document.activeElement?.tagName === "SELECT") {
        session.hideTimer = setTimeout(hide, 500);
        return;
      }
      document.querySelector(".native-player")?.classList.add("controls-hidden");
      requestAnimationFrame(bounds);
    };
    session.hideTimer = setTimeout(hide, 2400);
  };
  window._acNativePointer = revealControls;
  const playbackFailed = session => {
    if (active !== session || session.failed) return;
    session.failed = true;
    clearTimeout(session.hideTimer);
    clearInterval(session.timer);
    invoke("_acNativeBounds", 0, 0, 0, 0).catch(error);
    const root = document.querySelector(".native-player");
    root?.classList.remove("controls-hidden");
    root?.classList.add("playback-failed");
    document.querySelector("#native-status").textContent = t("Fichier illisible");
    document.querySelector("#native-surface").innerHTML = `<div class="native-error" role="alert"><strong>${t("Impossible de lire ce fichier")}</strong><p>${t("Le fichier est endommagé, incomplet ou son format n’est pas reconnu.")}</p><button class="ghost" onclick="closePlayer()">${t("Fermer")}</button></div>`;
  };
  const tracks = (select, list, selected, external = []) => {
    if (!select || document.activeElement === select) return;
    const signature = JSON.stringify([list, external]);
    if (select.dataset.signature !== signature) {
      select.dataset.signature = signature;
      select.replaceChildren();
      for (const item of list) select.add(new Option(item.name || String(item.id), String(item.id)));
      for (const sub of external) select.add(new Option(`${(sub.lang || "ST").toUpperCase()} — ${sub.path.split("/").pop()}`, "external:" + sub.path));
    }
    if (!select.value.startsWith("external:")) select.value = String(selected);
  };
  async function poll(session) {
    if (active !== session || session.polling) return;
    session.polling = true;
    try {
      const s = await invoke("_acNativeStatus");
      if (active !== session) return;
      const previousState = session.status?.state;
      session.status = s;
      session.hadVideo ||= s.video_width > 0 && s.video_height > 0;
      if (s.state === 7 || (!session.hadVideo && [0, 5, 6].includes(s.state) && Date.now() - session.opened > 1500)) {
        playbackFailed(session);
        return;
      }
      if (previousState !== s.state) revealControls();
      const pause = document.querySelector("#native-pause");
      pause.textContent = s.state === 4 ? "▶" : "❚❚";
      pause.setAttribute("aria-label", s.state === 4 ? t("Reprendre") : "Pause");
      document.querySelector("#native-time").textContent = `${time(s.position)} / ${time(s.duration)}${s.parts > 1 ? ` · ${s.part + 1}/${s.parts}` : ""}`;
      document.querySelector("#native-status").textContent = s.state === 1 || s.state === 2 ? t("Chargement…") : "LibVLC";
      const seek = document.querySelector("#native-seek");
      if (!session.seeking) {
        seek.max = String(Math.max(1, s.duration));
        seek.value = String(Math.max(0, s.position));
      }
      const volume = document.querySelector("#native-volume");
      if (document.activeElement !== volume && s.volume >= 0) volume.value = String(s.volume);
      tracks(document.querySelector("#native-audio"), s.audio || [], s.audio_track);
      tracks(document.querySelector("#native-subs"), s.subtitles || [], s.subtitle_track, session.version.external_subs || []);
      if (Date.now() - session.saved > 10000) { persist(session); session.saved = Date.now(); }
      if (s.state === 6 && session.hadVideo && s.part + 1 === s.parts) { persist(session); await closeNative(); }
    } catch (e) { if (active === session && !session.failed) { session.failed = true; error(e); } }
    finally { session.polling = false; }
  }
  async function closeNative() {
    const session = active;
    if (!session) return;
    active = null;
    clearInterval(session.timer);
    clearTimeout(session.hideTimer);
    session.observer.disconnect();
    persist(session);
    try { await command("stop"); } catch (e) { error(e); }
    if (session.fullscreen) { try { await invoke("_acNativeFullscreen"); } catch {} }
    if (!active) { document.querySelector("#player-root").replaceChildren(); document.body.classList.remove("native-playing"); render(); }
  }
  window.closePlayer = () => active ? closeNative() : oldClose();
  window.openNativePlayer = async (version, subPath, title) => {
    const request = ++sequence;
    await closeNative();
    if (request !== sequence) return;
    oldClose();
    document.body.classList.add("native-playing");
    const progress = progFor(version.path);
    const resume = progress && !progress.watched && progress.position > 30 ? progress.position : 0;
    document.querySelector("#player-root").innerHTML = `
      <div class="player-back native-player" role="dialog" aria-label="${esc(title || version.raw_name)}">
        <div class="native-header"><span class="player-title">${esc(title || version.raw_name)}</span>
          <span id="native-status" role="status">${t("Chargement…")}</span>
          <button id="native-close" class="ghost" aria-label="${t("Fermer")}">✕</button></div>
        <div id="native-surface"></div>
        <div class="native-controls">
          <input id="native-seek" type="range" min="0" max="1" step="1" value="0" aria-label="${t("Position")}">
          <div class="native-control-row">
            <button id="native-pause" class="ghost" aria-label="Pause" disabled>❚❚</button>
            <span id="native-time">00:00 / 00:00</span>
            <label>${t("Volume")} <input id="native-volume" type="range" min="0" max="100" value="100"></label>
            <label>Audio <select id="native-audio" aria-label="${t("Piste audio")}"></select></label>
            <label>${t("Sous-titres :")} <select id="native-subs" aria-label="${t("Sous-titres :")}"></select></label>
            <button id="native-fullscreen" class="ghost" aria-label="${t("Plein écran")}">⛶</button>
          </div>
        </div>
      </div>`;
    const session = active = { version, status: null, saved: Date.now(), opened: Date.now(), polling: false, fullscreen: false };
    const fail = promise => promise.catch(error);
    document.querySelector("#native-close").onclick = () => closeNative();
    document.querySelector("#native-pause").onclick = () => fail(command("pause", session.status?.state === 4 ? 0 : 1));
    document.querySelector("#native-volume").oninput = event => fail(command("volume", Number(event.target.value)));
    const seek = document.querySelector("#native-seek");
    seek.oninput = () => { session.seeking = true; };
    seek.onchange = () => { session.seeking = false; fail(command("seek", Number(seek.value))); };
    seek.onblur = () => { session.seeking = false; };
    document.querySelector("#native-audio").onchange = event => fail(command("audio", Number(event.target.value)));
    document.querySelector("#native-subs").onchange = event => {
      const value = event.target.value;
      fail(value.startsWith("external:") ? invoke("_acNativeSubtitle", value.slice(9)) : command("subtitle", Number(value)));
    };
    document.querySelector("#native-fullscreen").onclick = async () => {
      try {
        session.fullscreen = await invoke("_acNativeFullscreen");
        const root = document.querySelector(".native-player");
        root.classList.toggle("cinema-mode", session.fullscreen);
        root.classList.remove("controls-hidden");
        clearTimeout(session.hideTimer);
        document.querySelector("#native-fullscreen").blur();
        revealControls();
        requestAnimationFrame(bounds);
      } catch (e) { error(e); }
    };
    session.observer = new ResizeObserver(bounds);
    session.observer.observe(document.querySelector("#native-surface"));
    try {
      const deadline = Date.now() + 60000;
      while (active === session && !(await invoke("_acNativeReady"))) {
        if (Date.now() > deadline) throw new Error(t("Le lecteur met trop de temps à démarrer. Réessayez dans un instant."));
        await new Promise(resolve => setTimeout(resolve, 200));
      }
      if (active !== session) return;
      await invoke("_acNativePlay", version.path, subPath || "", resume, resume ? (progress.part || 0) : 0);
      if (active !== session) return;
      document.querySelector("#native-pause").disabled = false;
      bounds();
      session.timer = setInterval(() => poll(session), 500);
      poll(session);
    } catch (e) { if (active === session) { error(e); await closeNative(); } }
  };
  window.addEventListener("resize", bounds);
  window.addEventListener("mousemove", revealControls);
  window.addEventListener("keydown", event => {
    if (!active) return;
    if (event.key === "Escape") {
      event.preventDefault();
      if (active.fullscreen) document.querySelector("#native-fullscreen").click();
      else closeNative();
    } else if (event.key === "F11" || (event.key.toLowerCase() === "f" && !/INPUT|SELECT|TEXTAREA/.test(event.target.tagName))) {
      event.preventDefault(); document.querySelector("#native-fullscreen").click();
    } else if (/INPUT|SELECT|TEXTAREA/.test(event.target.tagName)) return;
    else if (event.code === "Space") { event.preventDefault(); document.querySelector("#native-pause").click(); }
    else if (event.key === "ArrowRight" || event.key === "ArrowLeft") {
      event.preventDefault(); command("seek", Math.max(0, (active.status?.position || 0) + (event.key === "ArrowRight" ? 10 : -10))).catch(error);
    }
  });
  document.addEventListener("click", event => {
    const link = event.target.closest("a[href]");
    if (!link) return;
    const target = new URL(link.href, location.href);
    if (target.origin !== location.origin && /^https?:$/.test(target.protocol)) {
      event.preventDefault(); invoke("_acNativeExternal", target.href).catch(error);
    }
  });
})();
