/* MoAI Docs — Claude Design handoff behaviors
 * (1) TOC scroll-spy: marks the .cw-toc link of the last h2/h3[id] above the fold.
 * Self-contained vanilla JS, no dependencies. Light-only theme.
 */
(function () {
  'use strict';

  /* ---------- (1) TOC scroll-spy ---------- */
  function initScrollSpy() {
    var toc = document.querySelector('.cw-toc');
    var article = document.querySelector('.gdoc-markdown');
    if (!toc || !article) return;
    var links = toc.querySelectorAll('a[href^="#"]');
    if (!links.length) return;
    var heads = article.querySelectorAll('h2[id], h3[id]');
    if (!heads.length) return;

    var ticking = false;
    function update() {
      ticking = false;
      var cur = null;
      heads.forEach(function (h) {
        if (h.getBoundingClientRect().top < 120) cur = h.id;
      });
      if (!cur) cur = heads[0].id;
      links.forEach(function (a) {
        var hash = decodeURIComponent(a.getAttribute('href') || '').slice(1);
        a.classList.toggle('active', hash === cur);
      });
    }
    window.addEventListener('scroll', function () {
      if (!ticking) { ticking = true; window.requestAnimationFrame(update); }
    }, { passive: true });
    update();
  }

  function init() {
    initScrollSpy();
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
