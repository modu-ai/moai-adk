// MoAI Web Console — minimal vanilla-JS progressive enhancement.
//
// The form works without JavaScript (plain HTML <form> POST round-trip). This
// script only adds three conveniences:
//   1. Toggle the custom-segments group based on the selected statusline preset.
//   2. Auto-submit the profile selector so switching profiles reloads the form.
//   3. Light/dark theme toggle persisted client-side in localStorage
//      (SPEC-WEB-CONSOLE-004 / REQ-WC4-006). No server round-trip, no config
//      field — the theme is a machine-local UI preference only.
//
// No build toolchain, no framework, no network fetch of dependencies (REQ-WC-005).
// FOUC is prevented by an inline <head> snippet that applies the persisted theme
// before first paint; this script only wires the interactive toggle.
(function () {
  "use strict";

  var THEME_KEY = "moai-console-theme";

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

  document.addEventListener("DOMContentLoaded", function () {
    syncSegmentsVisibility();
    var preset = document.querySelector('select[name="statusline_preset"]');
    if (preset) {
      preset.addEventListener("change", syncSegmentsVisibility);
    }
    wireProfileSwitch();
    wireThemeToggle();
  });
})();
