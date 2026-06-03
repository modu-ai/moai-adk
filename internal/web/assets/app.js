// MoAI Web Console — minimal vanilla-JS progressive enhancement.
//
// The form works without JavaScript (plain HTML <form> POST round-trip). This
// script only adds two conveniences:
//   1. Toggle the custom-segments fieldset based on the selected statusline preset.
//   2. Auto-submit the profile selector so switching profiles reloads the form.
//
// No build toolchain, no framework, no network fetch of dependencies (REQ-WC-005).
(function () {
  "use strict";

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

  document.addEventListener("DOMContentLoaded", function () {
    syncSegmentsVisibility();
    var preset = document.querySelector('select[name="statusline_preset"]');
    if (preset) {
      preset.addEventListener("change", syncSegmentsVisibility);
    }
    wireProfileSwitch();
  });
})();
