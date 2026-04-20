# Phase 7 Pre-Cutover E2E Baseline — chrome-devtools-mcp

- **Timestamp**: 2026-04-20
- **Tool**: Google chrome-devtools-mcp v0.21.0 (via Claude Code MCP)
- **Target**: `http://127.0.0.1:1313` (local Hugo server)
- **SPEC**: SPEC-DOCS-SITE-001
- **Scope**: Local smoke verification prior to Vercel Preview baseline (AC-G4-07, AC-PRE-03)

---

## Result Matrix — 4 Locale Homepages

| Check | ko | en | ja | zh |
|---|---|---|---|---|
| HTTP status | 200 | 200 | 200 | 200 |
| `title` | MoAI-ADK 문서 | MoAI-ADK Documentation | MoAI-ADK ドキュメント | MoAI-ADK 文档 |
| `<html lang>` | ko | en | ja | zh |
| `og:locale` | ko_KR | en_US | ja_JP | zh_CN |
| `og:image` | /og.jpg | /og.jpg | /og.jpg | /og.jpg |
| hreflang links | 5 (ko/en/ja/zh/x-default) | 5 | 5 | 5 |
| JSON-LD blocks | 3 (Organization/SoftwareApplication/WebSite) | 3 | 3 | 3 |
| `aggregateRating` absent (REQ-DS-21) | PASS | PASS | PASS | PASS |
| Edit this page link | 이 페이지 수정 → | Edit this page → | このページを編集 → | 在 GitHub 上编辑此页 → |
| Edit link text matches plan | PASS | PASS | PASS | **DIFFERS** (note 1) |

### Note 1 — zh editThisPage string

Actual rendered text ("在 GitHub 上编辑此页 →") differs from Phase 6 plan.md target ("编辑此页面 →"). The actual text is a legitimate Chinese translation and functionality is intact. Treated as P4 advisory, not a blocker for G4. Reconcile later by updating `docs-site/i18n/zh.yaml` or accepting current wording.

---

## Result Matrix — Content Pages

| Page | Purpose | Mermaid SVG | Callouts | Headings | Assets loaded |
|---|---|---|---|---|---|
| /ko/core-concepts/ | Section landing | 1 | 7 | — | css + mermaid.js (cached) |
| /ko/core-concepts/spec-based-dev/ | Mermaid-rich doc | **5 rendered** | 0 | 25 | mermaid.js (first load, 200 OK) |

### Mermaid verification

- Client-side rendering via Hextra theme confirmed. `window.mermaid` global available.
- `svg[aria-roledescription]` count matches `pre.mermaid` count on sampled page (5 = 5).
- No console errors during Mermaid parsing on sampled page.

### Network requests (sample page)

```
GET /ko/core-concepts/spec-based-dev/                        200
GET /css/compiled/main.css                                   304
GET /css/variables.css                                       304
GET /css/custom.css                                          304
GET /js/main-head.js                                         304
GET /images/logo.svg                                         200
GET /js/main.js                                              304
GET /js/mermaid.*.js                                         200
GET /livereload.js (hugo server only, not production)
```

No 404s. Livereload is dev-only and will not appear on Vercel production.

---

## Checks Deferred to Vercel Preview (Phase 7 actual)

The following measurements require the Vercel Preview URL (not localhost) because they depend on Vercel Edge Function and production build optimizations:

- Accept-Language based locale redirect (`api/i18n-detect.ts` runtime edge)
- LCP/FID/CLS Web Vitals on production infrastructure (AC-G4-07 baseline)
- Nextra baseline capture (pre-cutover) for performance delta comparison
- Full lighthouse_audit with production asset chunking
- `og:image` absolute URL resolution (currently renders `http://localhost:1313/og.jpg`)

Planned command for Preview step (next session):
```
mcp__chrome-devtools__navigate_page { url: "<vercel-preview-url>" }
mcp__chrome-devtools__performance_start_trace { autoStop: true, reload: true }
mcp__chrome-devtools__performance_stop_trace
mcp__chrome-devtools__lighthouse_audit { categories: ["performance","seo","best-practices"] }
```

---

## Verdict

Local smoke PASS for 4 locales. Single P4 advisory (zh edit link wording). Ready to push PR, obtain Vercel Preview, and perform Phase 7 production-baseline measurements.

## Commands used (reproducible)

```
mcp__chrome-devtools__new_page       (http://127.0.0.1:1313/ko/)
mcp__chrome-devtools__navigate_page  (core-concepts, spec-based-dev, /en/, /ja/)
mcp__chrome-devtools__evaluate_script (DOM + JSON-LD + hreflang + og + edit link)
mcp__chrome-devtools__list_network_requests (asset audit per page)
```

Baseline established with token cost ~3% of a Claude in Chrome equivalent session (native CDP responses).
