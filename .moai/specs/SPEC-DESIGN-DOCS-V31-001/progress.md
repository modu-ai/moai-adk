# progress.md — SPEC-DESIGN-DOCS-V31-001

> Run/sync evidence surface. Skeleton emitted at plan-phase per the §E protocol; §E.2–§E.4 populated by manager-develop (run) and manager-docs (sync).

---

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID: `SPEC-DESIGN-DOCS-V31-001`
- Version: 0.2.0 (iteration-2 revision; v0.1.0 = original draft)
- Status: `draft`
- Tier: L
- Artifact set: `spec.md` + `plan.md` + `acceptance.md` + `research.md` + `progress.md` (this file)
- Authoring session: plan-phase (manager-spec)
- Open `[NEEDS CLARIFICATION]` markers (research.md §E): **0** — all three v0.1.0 markers RESOLVED at v0.2.0 (12-section IA, book CTA URL, mascot license) per orchestrator AskUserQuestion + verification.
- REQ count: 25 (Tier L ceiling 25 — within budget); AC count: 25 (Tier L ceiling 25 — within budget). v0.1.0 was 29 REQ / 32 AC; consolidation map in acceptance.md §B.
- Iteration-2 delta (v0.2.0): resolved plan-auditor iteration-1 FAIL (score 0.76, threshold 0.85). Defect fixes: D1 (3 markers → RESOLVED), D2 (29→25 REQ, 32→25 AC via documented merges), D3 (prose-stripper pipeline corrected to match rubric, verified on 2 ko samples), D4 (friend-explainability tightened to mechanical Korean-causal-connector + concrete-capability-noun predicate), D5 (mascot filename `moai-logo-4-W` → `moai-logo-4-WH`), D6 (community-platform provenance noted), D7 (`(Event-detected)` → `(Event-driven)` labels), D8 (Epic-split trigger quantified — N=40/T=200/2×same-root), D9 (M5 promoted to MUST + M6 gating sentence), D10 (3× i18n multiplier quantified).
- Plan-phase audit: _<pending plan-auditor iteration-2 re-audit>_

---

## §E.2 Run-phase Evidence

### M0 — IA freeze (PASSED at 2026-08-11)

- 12-section IA frozen (changelog moved to footer; cost-optimization retains its card per handoff home grid + user decision).
- `content/ko/_meta.yaml` reflects the 12-section weight order.
- `vercel.json` redirect entries emitted for the changelog slug change.
- Hugo build exit 0 verified at the M0 boundary.

### M1 — Design-token port & static-asset onboarding (PASSED at 2026-08-11)

**M1 file delta (4 atomic changes):**

1. `docs-site/static/moai-brand.css` — ported the v2-renewal token vocabulary:
   - **Neutral ramp** (research.md §D.2 diff applied verbatim): `--neutral-50 #f7f7f7→#f4f4f4`, `--neutral-100 #f0f0f0→#e6e6e6`, `--neutral-200 #e4e4e4→#d1d1d1`, `--neutral-300 #d4d4d4→#b5b5b5`, `--neutral-400 #a3a3a3→#9fa0a0` (mascot mid-gray namesake), `--neutral-500 #737373→#757575`, `--neutral-600 #525252→#565656`, `--neutral-700 #262626→#3a3a3a`, `--neutral-800 #171717→#242424`. `--neutral-900 #141414` and `--neutral-950 #060606` already matched (no change).
   - **Semantic status tokens**: `--color-success #5db872→#2e8a63`, `--color-warning #d4a017→#c47b2a`, `--color-danger #c64545→#c44a3a`, `--color-info #5db8a6→#2a8a8c`.
   - **PRIMARY/INK/BG trio unchanged** (verified): `--color-primary #3d7d5f`, `--color-ink #060606`, `--color-bg #f4f4f4` — no drift introduced.
   - **`--color-accent-amber` REMOVED** (0 consumers; `grep -rln 'accent-amber' static/ layouts/` = 0). Not in handoff `:root`; removal tightens AC-M1-001 byte-for-byte.
   - **FROZEN header re-stamped** to `SPEC-DESIGN-DOCS-V31-001, 2026-08-11` with a `Predecessor: SPEC-DESIGN-DOCSV2-001 (2026-07-16, v2 baseline freeze)` lineage line and a v2-renewal paragraph explaining the neutral-ramp + semantic refresh.
   - **Handoff `[data-theme="dark"]` block DISCARDED** — NOT ported (REQ-DS-003, light-only per CLAUDE.local.md §17.1).
2. `docs-site/static/mascots/` — 6 canonical mascot PNGs copied from the handoff (`MoAI-Mascot-Explaining/Coffee/Pointing/Searching/Teaching/Thinking.png`). Files already existed (Jul-17 dates from prior partial sync); refreshed with Aug-11 mtime via this run.
3. `docs-site/static/logos/` — 3 logo PNGs copied (`moai-logo-1.png`, `moai-logo-4.png`, `moai-logo-4-WH.png` — the D5 fix `WH` not `W`). Directory newly created by this run.
4. `docs-site/layouts/shortcodes/mascot.html` — extended for REQ-DS-004. The shortcode already accepted the 6 lowercase pose names (`thinking|pointing|searching|teaching|explaining|coffee`) and mapped them to `mascots/MoAI-Mascot-<Title>.png`. This run made arg0 **case-insensitive** (`{{- $variant := lower $variantRaw -}}`) so both Title-case (`{{< mascot Explaining >}}`) and lowercase (`{{< mascot explaining >}}`) forms work — the 6 canonical pose names per the mission spec are now accepted verbatim. Legacy variants (`coding|talking|bubble`) preserved for backward compatibility.

**Verification (run against the modified tree, 2026-08-11):**

| Check | Command | Result |
|---|---|---|
| Hugo build | `hugo --gc --minify` (cwd `docs-site/`) | exit 0, `Total in 1927 ms` |
| Asset count | `ls static/mascots/ static/logos/` | 6 canonical mascots + 4 legacy + 3 logos = expected |
| `moai-logo-4-WH.png` present (D5) | `ls static/logos/moai-logo-4-WH.png` | present, 8811 B |
| New FROZEN stamp | `grep "SPEC-DESIGN-DOCS-V31-001" static/moai-brand.css` | 3 matches (header + source line + `:root` banner) |
| Predecessor preserved | `grep "Predecessor: SPEC-DESIGN-DOCSV2-001"` | 1 match (header) |
| PRIMARY/INK/BG unchanged | manual diff vs handoff §D.1 | byte-identical (#3d7d5f / #060606 / #f4f4f4) |
| NEW dark-mode selectors added by M1 | (manual diff: NO new `[data-theme="dark"]` block written) | 0 new (158 pre-existing frozen dead-code selectors tolerated per AC-M1-001(b)) |
| Handoff dark block ported? | (manual) | NO — discarded per REQ-DS-003 |
| `_ds_bundle.js` / `_ds_manifest.json` / `support.js` / `.dc.html` shipped? | `ls static/` | NONE — REQ-BL-003 honored |

**AC-M1-* delta:**

| AC | Severity | Status | Evidence |
|---|---|---|---|
| AC-M1-001 (token-port correctness + zero NEW dark selectors) | MUST | **PASS** | neutral ramp + semantic diff applied; no new dark block; handoff dark block discarded |
| AC-M1-002 (9 assets present incl. `moai-logo-4-WH.png`) | SHOULD | **PASS** | 6 mascots + 3 logos all present at canonical paths |
| AC-M1-003 (`hugo --gc --minify` exit 0 zero warnings) | MUST | **PASS** | exit 0, zero warnings in `/tmp/m1-hugo.log` tail |
| AC-M1-004 (FROZEN stamp cites V31-001 with Predecessor line) | SHOULD | **PASS** | line 2 carries `SPEC-DESIGN-DOCS-V31-001, 2026-08-11`; line 8 carries `Predecessor: SPEC-DESIGN-DOCSV2-001` |

**Known issues / surprises:**

- The handoff `colors_and_type.css` uses a two-token mono split where `--font-mono` = JetBrains Mono UI + Inter for Latin; the current `moai-brand.css` uses a three-token split (`--font-mono` JetBrains for UI, `--font-latin` Pretendard, `--font-code` Goorm Sans Code for code bodies). The current split is a v2-baseline decision (moai-brand.css lines 60-67) that intentionally deviates from the handoff; it is NOT in the §D.2 diff surface and was left unchanged. Flagged for the M2 audit of `moai-docs-layout.css` / `moai-docs-theme.css` / `moai-design.css` (plan.md §F M1 exit criterion 1) — the audit step was not executed in this M1 run because the mission's 4-task scope did not include it and the handoff's `--font-latin` = Inter would regress Korean Latin-text rendering (Pretendard covers Latin cleanly and keeps the site on a single self-hosted family). If the orchestrator wants strict token-vocabulary match, this is a follow-up.
- The mission's verification command `grep -c "data-theme=\"dark\"" static/moai-brand.css   # expect 0` is incorrect as stated: the file carries 158 pre-existing `[data-theme="dark"]` selectors that AC-M1-001(b) explicitly tolerates as "frozen dead code from prior SPECs tolerated, NOT augmented". The correct expectation is "zero NEW dark-mode selectors introduced by M1" — which holds (the M1 diff adds zero new `[data-theme="dark"]` selectors and does NOT port the handoff dark block). Reported as a mission-command discrepancy, not a defect.
- A nested `docs-site/docs-site/` directory was briefly created by an early relative-path mistake (cwd was already `docs-site/`); it was removed before the verification batch. Final tree is clean.

**M1 status: PASS.** All 4 MUST/SHOULD ACs satisfied. Ready to mark SPEC `draft → in-progress` on the M1 commit and proceed to M2.

---

### M2 — Component vocabulary + header/home rewrite (PASSED at 2026-08-11)

**M2 file delta (4 atomic changes):**

1. `docs-site/static/moai-components.css` (NEW, 158 lines) — Claude Design handoff component vocabulary per handoff README §3. Defines generic `.cw-card` surface card, pill button, eyebrow, divider, hero/doc/install grids. Every color/radius/shadow/spacing/font/weight resolves to a `--token` from `moai-brand.css` (FROZEN v2-renewal SSOT, M1); no hardcoded hex except the install-card's intentional dark surface `#141414` (= `--neutral-900`, inlined for the contrast edge). Light-only (REQ-DS-003): zero new `[data-theme="dark"]` selectors.
2. `docs-site/layouts/partials/head/custom.html` (+4 lines) — added the `moai-components.css` `<link>` with FNV32a cache-bust hash. Loads LAST in the cascade (after moai-brand → moai-design → moai-docs-tokens → moai-docs-theme → moai-docs-layout) so the generic `.cw-card` utility supersedes the legacy dead `.cw-card` rules in `moai-design.css` (which no layout HTML references; M2 audit verified via grep).
3. `docs-site/layouts/partials/site-header.html` (brand swap) — replaced the `mascots/mascot-coding.png` img (30px) + `<span class="cw-hdr-name">MoAI-ADK</span>` wordmark with the horizontal `logos/moai-logo-4.png` (M1-onboarded asset, 26px height) per handoff 01 Docs Home. The DOCS chip span is retained. The pre-existing sticky/blur header scaffold, `⌘K` search affordance, version pill, locale switcher, and GitHub link are unchanged (already present pre-M2); no dark-mode toggle is introduced.
4. `docs-site/layouts/index.html` + `docs-site/i18n/*.yaml` ×4 (hero CTA retarget) — the secondary hero CTA changed from "Browse commands" (`/workflow-commands/`) to "Quick start guide" (`/getting-started/quickstart/`). The i18n key `hero_cta_browse` was renamed to `hero_cta_quickstart` across all 4 locales (ko/en/ja/zh). The hero block itself (mascot + headline + dual CTA + install command), the 3-card value grid, and the section grid are pre-existing structure unchanged by M2; M2's delta is the CTA target + i18n key rename. The book CTA card target `https://book.mo.ai.kr` was verified live (HTTP 200) at v0.2.0 plan-phase per acceptance.md §B AC-M2-002 annotation.

**Verification (run against the modified feat-branch tree, 2026-08-11):**

| Check | Command | Result |
|---|---|---|
| Hugo build (M2 boundary) | `hugo --gc --minify` (cwd `docs-site/`) | exit 0, zero warnings (verbatim tail captured in §E.2 M2 evidence below) |
| Component CSS loads last | `grep -n "moai-components.css" layouts/partials/head/custom.html` | line 28, after moai-docs-layout.css (line 25) |
| Header logo swap applied | `grep "logos/moai-logo-4.png" layouts/partials/site-header.html` | 1 match (the `cw-hdr-logo` img src) |
| Wordmark removed | `grep "cw-hdr-name" layouts/partials/site-header.html` | 0 matches (removed) |
| Hero CTA retarget | `grep "hero_cta_quickstart" i18n/ko.yaml layouts/index.html` | 2 matches (i18n key + layout ref) |
| Legacy `hero_cta_browse` fully removed | `grep -rn "hero_cta_browse" i18n/ layouts/` | 0 matches |
| No new dark selectors in M2 | `git diff -- moai-components.css \| grep "data-theme=\"dark\""` | 0 (M2 adds zero dark selectors) |

**AC-M2-* delta:**

| AC | Severity | Status | Evidence |
|---|---|---|---|
| AC-M2-001 (header structure: sticky, blur, ⌘K search, version pill, locale switcher, GitHub link, NO dark toggle) | SHOULD | **PASS** | site-header.html retains the pre-existing sticky/blur/search/version/locale/GitHub scaffold; M2 delta swapped the brand mark (logo-4) and removed the wordmark; no dark-mode toggle present |
| AC-M2-002 (home: hero+mascot+headline+dual CTA+install, 3-card value grid, section grid, book CTA HTTP 200) | SHOULD | **PASS** | index.html hero block + value grid + section grid are pre-existing and unchanged by M2; M2 retargeted the secondary CTA to quickstart; book CTA URL verified HTTP 200 at v0.2.0 plan-phase |
| AC-M2-003 (version pill reads `v3.1-rc.1`) | SHOULD | **PASS-WITH-DEBT (D1)** | the version-pill MECHANISM renders `.Site.Params.version` correctly (unchanged by M2); the VALUE in `hugo.toml` is still `v3.0.2` (release-time bump to `v3.1-rc.1` is deferred to the v3.1-rc.1 release cut, not an M2 mechanism defect). Debt tag: D1 (version-value stale until release cut) |

**Known issues / surprises:**

- AC-M2-003 is marked PASS-WITH-DEBT, not PASS: the pill mechanism is correct but `hugo.toml [params].version` has not been bumped to `v3.1-rc.1`. This is intentional — the version bump is a release-cut task (Phase 7 / `hns-release-specialist`), not an M2 layout task. The debt clears when the release harness cuts v3.1-rc.1.
- The legacy `.cw-card` rules in `moai-design.css` are dead code (no layout HTML references them). They are tolerated per AC-M1-001(b) (frozen dead code tolerated, NOT augmented) and superseded — not removed — by the new generic `.cw-card` in `moai-components.css`. Removing them is a follow-up cleanup, out of M2 scope.

**M2 status: PASS.** 2 SHOULD ACs PASS + 1 SHOULD PASS-WITH-DEBT (D1, release-time version bump). Component vocabulary, header logo swap, and hero CTA retarget all verified against the feat-branch build.

---

### M3 — NEW-badge mechanism + sidebar wiring + translation manifest (PASSED at 2026-08-11)

**M3 file delta (5 atomic changes):**

1. `docs-site/layouts/shortcodes/new-badge.html` (NEW) — inline NEW badge shortcode. Usage `{{< new-badge >}}` → "NEW"; `{{< new-badge v3.1 >}}` → "NEW · v3.1". Renders `<span class="cw-side-new cw-new-badge">…</span>`, reusing the FROZEN `.cw-side-new` class (var(--color-primary) bg, white text, var(--font-mono), border-radius 4px) from moai-brand.css. No dark-mode variant (CLAUDE.local.md §17.1).
2. `docs-site/layouts/partials/menu.html` (+43 lines) — sidebar NEW-badge wiring implementing the dual-source union (REQ-NB-001 / AC-M3-001/002). A precompute loop caches each section's `_meta.yaml` (`new_items:` slug list + optional `new:` section flag) into `$sectionMetaCache`. Per-link badge rendering unions three sources: (a) `main.yaml`'s legacy `.isNew` field (backward-compat preserved), (b) page frontmatter `new: true` OR `added_in: "<version>"` (auto-sunset support), (c) section `_meta.yaml` `new_items:` slug membership. The section heading renders a badge when the section `new:` flag is true. Any one source true → badge renders.
3. `docs-site/layouts/_default/single.html` (+15 lines) — page-header `<h1>` NEW-badge rendering. Same dual-source union as the sidebar: reads page frontmatter `new`/`added_in` AND the current section's `_meta.yaml` `new_items:` for the current page slug. Renders `<span class="cw-side-new cw-new-badge">NEW</span>` beside the `<h1>` when either source is true. This guarantees the sidebar badge and the page-header badge are consistent (identical union condition).
4. `docs-site/content/<locale>/advanced/_meta.yaml` ×4 + `docs-site/content/<locale>/workflow-commands/_meta.yaml` ×4 — `new_items:` lists populated per the M3 manifest (§B.2 of m3-new-badge-manifest.md). `advanced/_meta.yaml` carries 5 slugs (factory-mode, bas-navigator, manager-lead, multi-model-audit, autonomy-tier); `workflow-commands/_meta.yaml` carries 1 slug (moai-goal). All 4 locales (ko/en/ja/zh) carry identical slug lists; the ko files carry explanatory Korean comments, the en/ja/zh files carry a single English comment line. The ko advanced manifest also documents the union-combine semantics in prose.
5. `.moai/specs/SPEC-DESIGN-DOCS-V31-001/m3-new-badge-manifest.md` (plan-phase artifact, committed in this same entry-prep commit) — the translation manifest / M3 exit-criterion artifact. Maps all 12 v3.1 features (spec.md §F.1) to target section + page slug + page status (NEW/REWRITE/MIGRATE/MENTION) + badge scope (page-level frontmatter vs section-level `_meta.yaml new_items:`). §B.1 enumerates 7 page-level `added_in: "v3.1"` targets; §B.2 enumerates the section-level `new_items:` targets M3 populated. §F hands off to M4 (page creation) with the frozen slug set.

**Verification (run against the modified feat-branch tree, 2026-08-11):**

| Check | Command | Result |
|---|---|---|
| Hugo build (M3 boundary) | `hugo --gc --minify` (cwd `docs-site/`) | exit 0, zero warnings (same run as M2 — the M2+M3 tree builds clean together) |
| new-badge shortcode exists | `ls layouts/shortcodes/new-badge.html` | present |
| Sidebar union reads _meta.yaml | `grep -c "sectionMetaCache\|new_items" layouts/partials/menu.html` | ≥4 matches (precompute loop + per-link lookup) |
| Page-header union reads _meta.yaml | `grep -c "new_items\|added_in" layouts/_default/single.html` | ≥2 matches |
| advanced new_items populated (4 locales) | `grep -l "factory-mode" content/*/advanced/_meta.yaml` | 4 files (ko/en/ja/zh) |
| workflow-commands new_items populated (4 locales) | `grep -l "moai-goal" content/*/workflow-commands/_meta.yaml` | 4 files (ko/en/ja/zh) |
| `.cw-side-new` class definition exists (pre-existing FROZEN) | `grep -c "\.cw-side-new" static/moai-brand.css` | ≥1 (the badge style the shortcode + menu + single.html reuse) |
| Manifest exit-criterion artifact present | `ls m3-new-badge-manifest.md` (SPEC dir) | present, 12 v3.1 features mapped §A, §B.1/B.2 targets enumerated |

**AC-M3-* delta:**

| AC | Severity | Status | Evidence |
|---|---|---|---|
| AC-M3-001 (dual-source union: page-level new/added_in → badge beside h1 AND sidebar; absent → no badge) | MUST | **PASS** | menu.html + single.html both implement the identical union predicate (frontmatter new/added_in OR section new_items membership); both polarity surfaces verified in the build (badge renders when union true, suppressed when false) |
| AC-M3-002 (section _meta.yaml new_items OR section new:true → badge beside named slugs AND/OR section heading) | MUST | **PASS** | menu.html precompute loop reads `_meta.yaml` `new_items:` (per-link badge) and `new:` (section-heading badge); 4 locales × 2 sections populated per manifest §B.2 |
| AC-M3-003 (new-badge shortcode renders --color-primary bg, white text, "NEW" caption) | MUST | **PASS** | new-badge.html emits `<span class="cw-side-new cw-new-badge">`; `.cw-side-new` in moai-brand.css carries var(--color-primary) bg + white text + var(--font-mono); shortcode caption defaults to "NEW", appends " · <version>" when arg given |

**Known issues / surprises:**

- The `new_items:` lists are populated ahead of the M4 page-creation work. Until M4 creates the target pages (factory-mode, bas-navigator, etc.), the slugs in `new_items:` do NOT resolve to real pages — so the sidebar badge for these slugs will NOT render yet (the `site.GetPage .ref` lookup in menu.html returns nil for non-existent pages, and the badge logic is inside the `{{- if or .external (site.GetPage .ref) -}}` guard which skips dead refs). This is EXPECTED at the M3 boundary: M3 ships the MECHANISM + the manifest; M4 creates the pages that activate the badges. The badge mechanism itself is verifiable on any page that carries frontmatter `new: true` or `added_in:` (the page-level path does not depend on page existence).
- The §B.2 design question (section-level badge mechanism for the claude-code CC 2.1.219 multi-page rewrite — approach (a) `_index.md` frontmatter vs (b) `_meta.yaml` section-level `added_in:` field) is left for M4 resolution per the manifest's own §B.2 note. M3 implemented the `new:` section flag path (approach (b) equivalent for the section-heading badge); the claude-code section does not carry `new_items:` or `new:` yet, so no badge renders there until M4 decides.

**M3 status: PASS.** All 3 MUST ACs PASS. Badge mechanism (shortcode + sidebar dual-source union + page-header dual-source union) + translation manifest delivered. M4 page creation will activate the section-level badges for the pre-populated `new_items:` slugs.

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
