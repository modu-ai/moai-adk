/* doc-detail.js — round3 docs-detail behaviors for single pages.
 * SPEC-DESIGN-DOCSV2-001 M3c-2. Vanilla JS, no dependencies. Light-only theme.
 *
 * (1) Read-progress: #rp bar width tracks scroll percentage of the page.
 * (2) TOC scroll-spy: highlights the .docs-toc-list anchor of the h2 in view.
 *
 * Selector note: §C/§D draft referenced ".toc-list", but the M3a CSS renamed
 * .toc → .docs-toc (moai-docs-layout.css line 23, ".docs-toc-list"). This file
 * aligns to the M3a CSS SSOT (".docs-toc-list"); the HTML emits the same class.
 */
(function () {
  'use strict';

  function init() {
    /* ---------- (1) Read-progress bar (#rp) ---------- */
    var rp = document.getElementById('rp');
    if (rp) {
      var ticking = false;
      function updateProgress() {
        ticking = false;
        var max = document.documentElement.scrollHeight - window.innerHeight;
        var pct = max > 0 ? (window.scrollY / max) * 100 : 0;
        rp.style.width = Math.max(0, Math.min(100, pct)) + '%';
      }
      window.addEventListener('scroll', function () {
        if (!ticking) { ticking = true; window.requestAnimationFrame(updateProgress); }
      }, { passive: true });
      updateProgress();
    }

    /* ---------- (2) TOC scroll-spy (.docs-toc-list) ---------- */
    var heads = document.querySelectorAll('.doc-body h2[id]');
    if (heads.length) {
      var links = document.querySelectorAll('.docs-toc-list a');
      if (links.length) {
        var obs = new IntersectionObserver(function (entries) {
          entries.forEach(function (e) {
            if (e.isIntersecting) {
              /* href is URL-encoded for non-ASCII ids (ko/ja/zh headings), so
                 decode before comparing to the raw element.id — matches the
                 decodeURIComponent pattern in moai-docs-theme.js. */
              var id = e.target.id;
              links.forEach(function (a) {
                var hash = decodeURIComponent((a.getAttribute('href') || '').slice(1));
                a.classList.toggle('active', hash === id);
              });
            }
          });
        }, { rootMargin: '-20% 0px -70% 0px' });
        heads.forEach(function (h) { obs.observe(h); });
      }
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
