/* MoAI Docs — Claude Design handoff behaviors
 * (1) TOC scroll-spy: marks the .cw-toc link of the last h2/h3[id] above the fold.
 * (2) Accent-color toolbox: fixed FAB + 6 palette swatches, persisted to localStorage.
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

  /* ---------- (2) Accent-color toolbox ---------- */
  var PALETTES = [
    { id: 'clay',   names: { ko: '클레이', en: 'Clay',   ja: 'クレイ',   zh: '陶土' }, c: '#D97757', h: '#C4633F', p: '#A44E2F', s: '#FCEEE7' },
    { id: 'amber',  names: { ko: '앰버',   en: 'Amber',  ja: 'アンバー', zh: '琥珀' }, c: '#B5720A', h: '#9A6008', p: '#7F4E06', s: '#EADFC5' },
    { id: 'sage',   names: { ko: '세이지', en: 'Sage',   ja: 'セージ',   zh: '鼠尾草' }, c: '#3F7A34', h: '#34662B', p: '#2A5322', s: '#DFE3C6' },
    { id: 'teal',   names: { ko: '틸',     en: 'Teal',   ja: 'ティール', zh: '青绿' }, c: '#1E7A78', h: '#196663', p: '#145350', s: '#CFE0D8' },
    { id: 'indigo', names: { ko: '인디고', en: 'Indigo', ja: 'インディゴ', zh: '靛蓝' }, c: '#3A57B0', h: '#2F4894', p: '#263B78', s: '#D7DCE8' },
    { id: 'plum',   names: { ko: '플럼',   en: 'Plum',   ja: 'プラム',   zh: '梅紫' }, c: '#8A3E82', h: '#74336D', p: '#5E2A58', s: '#E4D3E0' }
  ];
  var TITLES = { ko: '테마 색상', en: 'Theme color', ja: 'テーマカラー', zh: '主题颜色' };
  var KEY = 'moai-docs-accent';

  function locale() {
    var lang = (document.documentElement.lang || 'ko').slice(0, 2);
    return ['ko', 'en', 'ja', 'zh'].indexOf(lang) >= 0 ? lang : 'en';
  }

  function applyPalette(pl, wrap) {
    var r = document.documentElement.style;
    r.setProperty('--accent', pl.c);
    r.setProperty('--accent-hover', pl.h);
    r.setProperty('--accent-press', pl.p);
    r.setProperty('--accent-soft', pl.s);
    r.setProperty('--text-link', pl.c);
    r.setProperty('--border-clay', 'color-mix(in srgb, ' + pl.c + ' 45%, transparent)');
    /* keep the FROZEN brand-variable chain in sync */
    r.setProperty('--color-primary', pl.c);
    r.setProperty('--color-primary-hover', pl.h);
    r.setProperty('--color-primary-active', pl.p);
    if (wrap) {
      wrap.querySelectorAll('.md-tk-sw').forEach(function (b) {
        b.classList.toggle('on', b.getAttribute('data-id') === pl.id);
      });
    }
    try { localStorage.setItem(KEY, pl.id); } catch (e) { /* private mode */ }
  }

  function initToolbox() {
    if (document.querySelector('.md-toolbox')) return;
    var lang = locale();
    var wrap = document.createElement('div');
    wrap.className = 'md-toolbox';
    wrap.innerHTML =
      '<button type="button" class="md-tk-fab" aria-label="' + TITLES[lang] + '">' +
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">' +
      '<circle cx="13.5" cy="6.5" r="2.5"/><circle cx="17.5" cy="10.5" r="2.5"/><circle cx="8.5" cy="7.5" r="2.5"/><circle cx="6.5" cy="12.5" r="2.5"/>' +
      '<path d="M12 2a10 10 0 1 0 0 20 2.5 2.5 0 0 0 2-4 2.5 2.5 0 0 1 2-4h1.5a4.5 4.5 0 0 0 4.5-4.5C24 5.6 18.6 2 12 2Z"/></svg></button>' +
      '<div class="md-tk-pop"><div class="md-tk-title">' + TITLES[lang] + '</div><div class="md-tk-grid"></div></div>';
    document.body.appendChild(wrap);

    var grid = wrap.querySelector('.md-tk-grid');
    PALETTES.forEach(function (pl) {
      var b = document.createElement('button');
      b.type = 'button';
      b.className = 'md-tk-sw';
      b.setAttribute('data-id', pl.id);
      b.title = pl.names[lang];
      b.innerHTML = '<span class="md-tk-dot" style="background:' + pl.c + '"></span><span class="md-tk-nm">' + pl.names[lang] + '</span>';
      b.addEventListener('click', function () { applyPalette(pl, wrap); });
      grid.appendChild(b);
    });

    var fab = wrap.querySelector('.md-tk-fab');
    fab.addEventListener('click', function (e) { e.stopPropagation(); wrap.classList.toggle('open'); });
    document.addEventListener('click', function (e) { if (!wrap.contains(e.target)) wrap.classList.remove('open'); });

    var saved = null;
    try { saved = localStorage.getItem(KEY); } catch (e) { /* private mode */ }
    if (saved && saved !== 'clay') {
      var pl = null;
      PALETTES.forEach(function (p) { if (p.id === saved) pl = p; });
      if (pl) applyPalette(pl, wrap);
    } else if (saved === 'clay') {
      wrap.querySelector('.md-tk-sw[data-id="clay"]').classList.add('on');
    } else {
      wrap.querySelector('.md-tk-sw[data-id="clay"]').classList.add('on');
    }
  }

  function init() {
    initScrollSpy();
    initToolbox();
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
