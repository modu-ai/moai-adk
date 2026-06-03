// MoAI Web Console — minimal vanilla-JS progressive enhancement.
//
// The form works without JavaScript (plain HTML <form> POST round-trip). This
// script only adds four conveniences:
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
  });
})();
