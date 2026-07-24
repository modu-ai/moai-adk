// MoAI Web Console — minimal vanilla-JS progressive enhancement.
//
// The form works without JavaScript (plain HTML <form> POST round-trip). This
// script only adds four conveniences:
//   1. Toggle the custom-segments group based on the selected statusline preset.
//   2. Auto-submit the profile selector so switching profiles reloads the form.
//   3. Interface-language picker (SPEC-WEB-CONSOLE-005 / REQ-WC5-004/005). The
//      langpick switches the UI chrome language by replacing [data-i18n] element
//      text from the embedded i18n.js dictionary and updates <html lang> (which
//      activates the CJK webfont for ja/zh). It persists ONLY in
//      localStorage("moai-console-lang") — it is NOT a form field, submits
//      nothing, and never touches a server-validated content-language setting.
//      Interface language ≠ content language (the cohort core invariant).
//   4. In-page server shutdown button (#serverShutdown). A confirm dialog →
//      same-origin POST /__shutdown__ → the server reuses its existing
//      signal.NotifyContext drain path (no parallel shutdown). The page shows a
//      "shutting down" overlay and disables interactive controls; a fetch
//      rejection (connection reset mid-drain) is treated as expected success.
//
// (The former light/dark theme toggle — SPEC-WEB-CONSOLE-004 / REQ-WC4-006 —
// is retired by SPEC-DESIGN-MOAIWEBV2-002: the console is light-only. A stale
// theme-preference localStorage key from a prior version is simply
// ignored: never read, never written.)
//
// No build toolchain, no framework, no network fetch of dependencies (REQ-WC-005).
// FOUC is prevented by an inline <head> snippet that applies the persisted
// interface language before first paint; this script wires the interactive
// controls and applies the active-locale translations on load.
//
// 초기화는 DOMContentLoaded 와 htmx:afterSettle 양쪽에서 모두 실행된다. 폼이
// hx-boost="true" (root.templ) 라 POST /save 시 htmx 가 전체 body 를 AJAX 로
// 교체(swap)하는데, 이때 document.readyState 가 이미 "complete" 이므로
// DOMContentLoaded 가 재발생하지 않는다. afterSettle 없으면 swap 직후 새 body 의
// [data-i18n] 요소에 한국어가 재적용되지 않고(영어 서버 기본이 잔류),
// uiLangSelect 가 리스너를 잃어 언어 변경이 먹통이 된다.
// boost swap 은 body 전체 교체이므로 새 요소는 리스너가 없어 중복 등록 우려도 없다.
// JS/htmx 가 비활성된 환경에서는 폼이 일반 POST(전체 새로고침)로 동작하므로
// DOMContentLoaded 경로가 정상 작동한다(afterSettle 리스너는 htmx 로드 시에만 의미).
(function () {
  "use strict";

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
    // Enum <option> labels stay ENGLISH in every locale: keys containing ".opt."
    // (f.permission_mode.opt.*, f.model.opt.*, f.effort_level.opt.*, ...) carry
    // runtime enum tokens ("Bypass permissions", "High quality", "High") that are
    // more intuitive untranslated. Field titles, descriptions and tooltips still
    // follow the active locale — and so do the placeholder options, whose keys
    // (opt.project_default / opt.unset / opt.runtime_default) have no ".opt."
    // substring, which is why the guard is key-scoped rather than tag-scoped.
    var enDict = (window.MOAI_I18N && window.MOAI_I18N.en) || dict;
    var nodes = document.querySelectorAll("[data-i18n]");
    for (var i = 0; i < nodes.length; i++) {
      var key = nodes[i].getAttribute("data-i18n");
      var str = (key.indexOf(".opt.") >= 0 ? enDict : dict)[key];
      // Missing key → keep the existing baseline text (do not blank the element).
      if (typeof str === "string" && str.length > 0) {
        nodes[i].textContent = str;
      }
    }
    // SPEC-WEBCONF-SIMPLIFY-001 M4 (REQ-WC-015): resolve data-i18n-title into the
    // native title attribute — used by per-option <option> descriptions (design.md
    // §H.3, zero custom interaction JS — native hover tooltip) + field descriptions.
    var titled = document.querySelectorAll("[data-i18n-title]");
    for (var j = 0; j < titled.length; j++) {
      var tkey = titled[j].getAttribute("data-i18n-title");
      var tstr = dict[tkey];
      if (typeof tstr === "string" && tstr.length > 0) {
        titled[j].setAttribute("title", tstr);
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

  // ── M5-b D1 — Tab navigation (CSS show/hide) ──

  // wireTabs 배선: 탭 버튼 클릭 시 .is-active 클래스를 토글한다. 패널은 DOM
  // 에서 제거되지 않고 CSS display:none 만 토글된다 — 비활성 패널의 폼 필드도
  // 그대로 제출된다 (atomic Save contract). 첫 탭이 기본 활성 탭이다(서버가
  // is-active 클래스를 렌더 시 부여한다).
  function wireTabs() {
    var tabBtns = document.querySelectorAll(".tabs .tab[data-tab]");
    if (tabBtns.length === 0) {
      return;
    }
    for (var i = 0; i < tabBtns.length; i++) {
      tabBtns[i].addEventListener("click", function () {
        var tabId = this.getAttribute("data-tab");
        // 모든 탭/패널에서 is-active 제거.
        var allBtns = document.querySelectorAll(".tabs .tab[data-tab]");
        var allPanels = document.querySelectorAll(".tabpanel[data-panel]");
        for (var j = 0; j < allBtns.length; j++) {
          allBtns[j].classList.remove("is-active");
          allBtns[j].setAttribute("aria-selected", "false");
        }
        for (var k = 0; k < allPanels.length; k++) {
          allPanels[k].classList.remove("is-active");
        }
        // 클릭한 탭과 대응 패널에 is-active 추가.
        this.classList.add("is-active");
        this.setAttribute("aria-selected", "true");
        var panel = document.querySelector('.tabpanel[data-panel="' + tabId + '"]');
        if (panel) {
          panel.classList.add("is-active");
        }
      });
    }
  }

  // (구 wireAgentFMSubtabs 는 제거됐다: agentfm 섹션의 subagents/harness sub-tab
  // 이 사라지고 .claude/agents/moai/ 행만 단일 그리드로 렌더된다. data-agentfm-*
  // 속성을 방출하는 마크업이 더 이상 없다.)

  // wireProfileMatrix 는 성능 티어 라디오(max/medium/low) 변경 시 각 에이전트의
  // model/effort select 를 그 티어의 프로파일 매트릭스 셀로 즉시 재설정한다 (G3-3 —
  // 클라이언트 전용, 서버 왕복 없음). #moai-profile-matrix JSON blob 에서
  // matrix[tier][agent] = {model, effort} 를 읽는다. 사용자가 이번 세션에서 직접
  // 편집한 select(dirty)는 보존하며, 셀을 직접 편집하면 즉시 Custom 라디오로 전환한다.
  function wireProfileMatrix() {
    var el = document.getElementById("moai-profile-matrix");
    if (!el) return;
    var matrix;
    try {
      matrix = JSON.parse(el.textContent);
    } catch (e) {
      return;
    }
    var selects = document.querySelectorAll('select[name^="agentfm."]');
    var custom = document.getElementById("performance_tier--custom");
    var dirty = {};
    for (var s = 0; s < selects.length; s++) {
      (function (sel) {
        sel.addEventListener("change", function () {
          dirty[sel.name] = true;
          if (custom) custom.checked = true; // 직접 편집 → Custom 상태로 전환
        });
      })(selects[s]);
    }
    var radios = document.querySelectorAll('input[name="performance_tier"]');
    for (var r = 0; r < radios.length; r++) {
      radios[r].addEventListener("change", function () {
        if (!this.checked) return;
        var cells = matrix[this.value]; // Custom(=undefined)이면 재설정하지 않음
        if (!cells) return;
        for (var i = 0; i < selects.length; i++) {
          var sel = selects[i];
          if (dirty[sel.name]) continue; // 미저장 직접 편집 보존
          var mm = sel.name.match(/^agentfm\.(.+)\.(model|effort)$/);
          if (!mm) continue;
          var cell = cells[mm[1]];
          if (!cell) continue;
          var val = mm[2] === "model" ? cell.model : cell.effort;
          for (var o = 0; o < sel.options.length; o++) {
            if (sel.options[o].value === val) {
              sel.value = val;
              break;
            }
          }
        }
        // 티어 재설정이 model 셀을 haiku 로 또는 haiku 에서 바꿨을 수 있으므로
        // 각 행의 effort 잠금 규칙을 재적용한다 (repopulation 은 model select 에
        // change 이벤트를 발생시키지 않으므로 명시적 재적용이 필요하다).
        reapplyHaikuLocks();
      });
    }
  }

  // ── Haiku → effort-select lock ──

  // applyHaikuEffortLock 은 한 model select 값이 "haiku" 이면 같은 행(.agentfm-row)
  // 의 effort select 를 비활성화하고, "Effort N/A for Haiku" 힌트를 노출한다. haiku 가
  // 아니면 되돌린다. 행 단위 페어링은 name 매칭이 아니라 공유 컨테이너(.agentfm-row)
  // 로 스코프하므로 다른 에이전트 행에 새지 않는다. disabled select 는 제출되지
  // 않으며(서버 save 경로가 미제출 effort 를 resolved 값으로 backfill), 프로그램적
  // 값 설정(wireProfileMatrix)에는 영향이 없다.
  function applyHaikuEffortLock(modelSel) {
    var row = modelSel.closest ? modelSel.closest(".agentfm-row") : null;
    if (!row) return;
    var effort = row.querySelector('select[name$=".effort"]');
    var hint = row.querySelector(".agentfm-haiku-hint");
    var isHaiku = modelSel.value === "haiku";
    if (effort) effort.disabled = isHaiku;
    if (hint) {
      if (isHaiku) hint.classList.remove("is-hidden");
      else hint.classList.add("is-hidden");
    }
  }

  // reapplyHaikuLocks 는 리스너를 건드리지 않고 현재 값 기준으로 모든 행의 잠금을
  // 다시 적용한다 (초기 로드 + 티어 repopulation 이후 호출).
  function reapplyHaikuLocks() {
    var models = document.querySelectorAll('select[name^="agentfm."][name$=".model"]');
    for (var i = 0; i < models.length; i++) applyHaikuEffortLock(models[i]);
  }

  // wireHaikuEffortLock 는 각 model select 의 change 에 잠금 규칙을 배선하고 초기
  // 상태를 적용한다. boost body swap 후 재호출돼도 새 요소는 리스너가 없어 중복
  // 등록 우려가 없다(다른 wire* 함수와 동일).
  function wireHaikuEffortLock() {
    var models = document.querySelectorAll('select[name^="agentfm."][name$=".model"]');
    for (var i = 0; i < models.length; i++) {
      (function (sel) {
        sel.addEventListener("change", function () {
          applyHaikuEffortLock(sel);
        });
      })(models[i]);
    }
    reapplyHaikuLocks(); // 초기 상태(서버 렌더가 이미 disabled 를 방출하지만 재확인).
  }

  // initConsole 는 모든 콘솔 초기화를 한 곳에서 수행한다 — DOMContentLoaded(첫
  // 로드 / htmx 비활성 전체 새로고침) 와 htmx:afterSettle(boost body swap 직후)
  // 양쪽에서 호출된다. boost swap 은 body 전체를 교체하므로 새 요소는 리스너가
  // 없어 중복 등록 우려가 없다. persisted i18n 을 매번 재적용하여 swap 후에도
  // 한국어 인터페이스가 유지된다(파일 헤더 주석의 버그 A 수정 참조).
  function initConsole() {
    syncSegmentsVisibility();
    var preset = document.querySelector('select[name="statusline_preset"]');
    if (preset) {
      preset.addEventListener("change", syncSegmentsVisibility);
    }
    wireProfileSwitch();
    // wireLangpick 내부의 applyI18n(readPersistedLang()) 가 새 body 의 [data-i18n]
    // 요소에 persisted 언어(예: 한국어)를 재적용한다.
    wireLangpick();
    wireShutdownButton();
    // M5-b D1: 탭 nav 배선. CSS show/hide 만으로 동작 — 패널은 DOM 에 상주한다
    // (atomic Save contract: 비활성 패널의 필드도 제출됨).
    wireTabs();
    // G3-3: 성능 티어 → 에이전트 매트릭스 셀 즉시 재설정.
    wireProfileMatrix();
    // haiku model → effort select 비활성 잠금(초기 + change + 티어 repopulation).
    wireHaikuEffortLock();
  }

  document.addEventListener("DOMContentLoaded", initConsole);
  // htmx boost 가 body 를 swap 한 직후 document 에서 발생한다. afterSettle 없으면
  // swap 이후 DOMContentLoaded 가 재발생하지 않아 초기화가 누락된다.
  document.addEventListener("htmx:afterSettle", initConsole);

  // SPEC-WEB-CONSOLE-011 M5 — SPEC 보드의 remediation 명령 복사 버튼.
  // document 레벨 위임 리스너를 IIFE 최상위에서 "한 번만" 등록한다(initConsole
  // 내부가 아님 — htmx afterSettle 재호출 시 중복 등록을 피하기 위함). data-copy 를
  // 가진 버튼 클릭 시 그 값을 클립보드에 복사한다. 서버로의 요청·명령 실행은 전혀
  // 없다 — 순수 클라이언트 텍스트 복사다(REQ-WC11-043/044). 보드가 아닌 페이지에는
  // data-copy 요소가 없으므로 이 리스너는 no-op 이다.
  function copyToClipboard(text, btn) {
    function flash() {
      if (!btn) {
        return;
      }
      var label = btn.querySelector("[data-i18n]") || btn;
      var prev = label.textContent;
      label.textContent = "✓";
      setTimeout(function () {
        label.textContent = prev;
      }, 1200);
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(flash, function () {
        legacyCopy(text);
        flash();
      });
    } else {
      legacyCopy(text);
      flash();
    }
  }
  function legacyCopy(text) {
    // navigator.clipboard 불가(비-보안 컨텍스트 등) 시 textarea + execCommand 폴백.
    try {
      var ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "absolute";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
    } catch (e) {
      /* 복사 불가 — 사용자는 <code> 텍스트를 직접 선택해 복사할 수 있다 */
    }
  }
  document.addEventListener("click", function (e) {
    var btn = e.target && e.target.closest ? e.target.closest("[data-copy]") : null;
    if (!btn) {
      return;
    }
    var text = btn.getAttribute("data-copy");
    if (!text) {
      return;
    }
    e.preventDefault();
    copyToClipboard(text, btn);
  });
})();
