// MoAI Web Console — minimal vanilla-JS progressive enhancement.
//
// The form works without JavaScript (plain HTML <form> POST round-trip). This
// script only adds five conveniences:
//   1. Toggle the custom-segments group based on the selected statusline preset.
//   2. Auto-submit the profile selector so switching profiles reloads the form.
//   3. Light/dark theme toggle persisted client-side in localStorage
//      (SPEC-WEB-CONSOLE-004 / REQ-WC4-006). No server round-trip, no config
//      field — the theme is a machine-local UI preference only.
//   4. Interface-language picker (SPEC-WEB-CONSOLE-005 / REQ-WC5-004/005). The
//      langpick switches the UI chrome language by replacing [data-i18n] element
//      text from the embedded i18n.js dictionary and updates <html lang> (which
//      activates the CJK webfont for ja/zh). It persists ONLY in
//      localStorage("moai-console-lang") — it is NOT a form field, submits
//      nothing, and never touches a server-validated content-language setting.
//      Interface language ≠ content language (the cohort core invariant).
//   5. In-page server shutdown button (#serverShutdown). A confirm dialog →
//      same-origin POST /__shutdown__ → the server reuses its existing
//      signal.NotifyContext drain path (no parallel shutdown). The page shows a
//      "shutting down" overlay and disables interactive controls; a fetch
//      rejection (connection reset mid-drain) is treated as expected success.
//
// No build toolchain, no framework, no network fetch of dependencies (REQ-WC-005).
// FOUC is prevented by an inline <head> snippet that applies the persisted theme
// + interface language before first paint; this script wires the interactive
// toggles and applies the active-locale translations on load.
(function () {
  "use strict";

  var THEME_KEY = "moai-console-theme";
  var LANG_KEY = "moai-console-lang";
  var LOCALES = ["en", "ko", "ja", "zh"];

  function syncSegmentsVisibility() {
    var preset = document.querySelector('select[name="statusline_preset"]');
    var segments = document.getElementById("custom-segments");
    if (!preset || !segments) {
      return;
    }
    segments.style.display = preset.value === "custom" ? "" : "none";
  }

  function wireProfileSwitch() {
    var sel = document.querySelector('select[name="__profile_select"]');
    if (!sel) {
      return;
    }
    sel.addEventListener("change", function () {
      window.location.search = "?profile=" + encodeURIComponent(sel.value);
    });
  }

  function currentTheme() {
    return document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light";
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute("data-theme", theme);
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch (e) {
      /* localStorage unavailable — toggle still works for this page view */
    }
  }

  function wireThemeToggle() {
    var btn = document.getElementById("themeToggle");
    if (!btn) {
      return;
    }
    btn.addEventListener("click", function () {
      applyTheme(currentTheme() === "dark" ? "light" : "dark");
    });
  }

  // ── Server shutdown button (in-page graceful stop) ──

  // i18n key for the confirm-dialog text. Looks up the active interface locale's
  // string from window.MOAI_I18N; falls back to an English sentence when the
  // dictionary or key is unavailable so the dialog is never blank.
  function shutdownConfirmText() {
    var locale = readPersistedLang();
    var dict = (window.MOAI_I18N && window.MOAI_I18N[locale]) || null;
    if (dict) {
      var s = dict["appbar.shutdown.confirm"];
      if (typeof s === "string" && s.length > 0) {
        return s;
      }
    }
    return "Shut down the server? The console will stop and this tab will go offline.";
  }

  // wireShutdownButton wires the in-page power button: confirm → POST /__shutdown__
  // → show a "shutting down" overlay and disable the form/buttons. The fetch is a
  // plain same-origin POST (REQ-WC-005 — no external fetch). The server responds
  // 200 then triggers its existing signal/drain path in a goroutine; the page may
  // lose connectivity mid-drain, so a fetch rejection is treated as success too.
  function wireShutdownButton() {
    var btn = document.getElementById("serverShutdown");
    if (!btn) {
      return;
    }
    btn.addEventListener("click", function () {
      if (!window.confirm(shutdownConfirmText())) {
        return;
      }
      // Disable further clicks immediately (idempotent on the server side via the
      // signal.NotifyContext cancel, but this avoids duplicate dialogs).
      btn.disabled = true;
      showShutdownOverlay();
      disableInteractiveControls();

      fetch("/__shutdown__", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" }
      }).then(
        function () {
          /* server acknowledged; it is now draining. Overlay already shown. */
        },
        function () {
          /* Connection reset / closed mid-drain is expected — the server is
             shutting down. The overlay stays; the tab is going offline. */
        }
      );
    });
  }

  // showShutdownOverlay surfaces a full-page "shutting down" notice so the user
  // understands the tab is going offline. Uses minimal inline DOM (no framework).
  function showShutdownOverlay() {
    if (document.getElementById("moai-shutdown-overlay")) {
      return;
    }
    var overlay = document.createElement("div");
    overlay.id = "moai-shutdown-overlay";
    overlay.setAttribute("role", "status");
    overlay.setAttribute("aria-live", "polite");
    overlay.style.position = "fixed";
    overlay.style.inset = "0";
    overlay.style.display = "flex";
    overlay.style.alignItems = "center";
    overlay.style.justifyContent = "center";
    overlay.style.background = "rgba(0,0,0,0.55)";
    overlay.style.color = "#fff";
    overlay.style.fontFamily = "system-ui, sans-serif";
    overlay.style.fontSize = "1.1rem";
    overlay.style.zIndex = "9999";
    overlay.textContent = "Server is shutting down…";
    document.body.appendChild(overlay);
  }

  // disableInteractiveControls disables the form submit and all appbar buttons so
  // a half-drained page cannot initiate further writes.
  function disableInteractiveControls() {
    var form = document.querySelector("form.form");
    if (form) {
      form.style.pointerEvents = "none";
      form.style.opacity = "0.5";
    }
    var btns = document.querySelectorAll(".appbar .iconbtn, .actions .btn");
    for (var i = 0; i < btns.length; i++) {
      btns[i].disabled = true;
    }
  }

  // ── Interface i18n (REQ-WC5-004/005) ──

  // normalizeLocale returns one of the 4 valid locales, defaulting to "en" for
  // an absent/invalid value (REQ-WC5-005).
  function normalizeLocale(loc) {
    return LOCALES.indexOf(loc) >= 0 ? loc : "en";
  }

  // applyI18n replaces the text of every [data-i18n] element with the active
  // locale's string and updates <html lang> (which activates the CJK font stack
  // for ja/zh). A key absent from the dictionary leaves the element's existing
  // (English baseline) text intact — it is never blanked (R6 / EC-3).
  function applyI18n(locale) {
    locale = normalizeLocale(locale);
    document.documentElement.setAttribute("lang", locale);
    // The dictionary is the embedded /static/i18n.js (window.MOAI_I18N). When it
    // is unavailable (e.g. blocked), applyI18n is a no-op beyond the lang attr.
    var dict = (window.MOAI_I18N && window.MOAI_I18N[locale]) || null;
    if (!dict) {
      return;
    }
    var nodes = document.querySelectorAll("[data-i18n]");
    for (var i = 0; i < nodes.length; i++) {
      var key = nodes[i].getAttribute("data-i18n");
      var str = dict[key];
      // Missing key → keep the existing baseline text (do not blank the element).
      if (typeof str === "string" && str.length > 0) {
        nodes[i].textContent = str;
      }
    }
  }

  function persistLang(locale) {
    try {
      localStorage.setItem(LANG_KEY, locale);
    } catch (e) {
      /* localStorage unavailable — the switch still applies for this page view */
    }
  }

  function readPersistedLang() {
    try {
      return normalizeLocale(localStorage.getItem(LANG_KEY));
    } catch (e) {
      return "en";
    }
  }

  function wireLangpick() {
    var sel = document.getElementById("uiLangSelect");
    if (!sel) {
      return;
    }
    // Reflect the persisted locale in the picker, then apply it on load.
    var current = readPersistedLang();
    sel.value = current;
    applyI18n(current);
    // On change: apply + persist client-side only. No form submit, no fetch.
    sel.addEventListener("change", function () {
      var locale = normalizeLocale(sel.value);
      applyI18n(locale);
      persistLang(locale);
    });
  }

  document.addEventListener("DOMContentLoaded", function () {
    syncSegmentsVisibility();
    var preset = document.querySelector('select[name="statusline_preset"]');
    if (preset) {
      preset.addEventListener("change", syncSegmentsVisibility);
    }
    wireProfileSwitch();
    wireThemeToggle();
    wireLangpick();
    wireShutdownButton();
  });
})();
