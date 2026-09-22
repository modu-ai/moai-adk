---
id: SPEC-WEB-CONSOLE-018
title: "internal/web test coverage reinforcement — behavioral render assertions to the 85% package target"
version: "0.1.0"
status: draft
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, test-coverage, tdd, render-assertion, templ, regression-class
era: V3R6
tier: M
related_specs: [SPEC-WEB-CONSOLE-017, SPEC-DESIGN-MOAIWEBV2-001]
card: t1079
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-22 | manager-spec | Initial plan-phase draft (card t1079, Class C) |

---

## §A Background

Card t1079 asks for test-coverage reinforcement of `internal/web` from its measured baseline to the 85% package target (`.moai/config/sections/quality.yaml` `test_coverage_target: 85`), run-phase `cycle_type=tdd` — the deliverable IS tests.

### §A.1 Baseline attribution (HARD-1 / HARD-2)

Two figures exist and only one is the baseline of record:

- **HISTORICAL figure — cited for provenance only.** 67.9%, originating from the SPEC-WEB-CONSOLE-017 (t1051) sync-audit F2 finding, measured on tree `WT-save-observability @ 9d4a20eae` (evidence: `.moai/reports/t1051/sync-audit.md` §F2, local-only).
- **BASELINE OF RECORD — remeasured.** 2026-09-22 on tree `WT-web-coverage @ 0314801c2` (develop-aligned, this SPEC's starting tree). Command: `go test ./internal/web/... -cover`. Verbatim output:

  ```
  ok  	github.com/modu-ai/moai-adk/internal/web	26.621s	coverage: 67.9% of statements
  ```

The two figures coincide numerically (both 67.9%). The coincidence is recorded; only the remeasurement is cited as the baseline of record. Full per-statement profile: `.moai/state/verify/t1079/cover-baseline.out` (regenerable: `go test ./internal/web/ -coverprofile=/tmp/x.out && go tool cover -func=/tmp/x.out`).

### §A.2 Measured mass distribution (this tree, computed from the profile)

Package: 9,414 statements, 3,023 zero-covered → 67.9% covered. The uncovered mass is dominated by generated templ codegen output:

| File | Statements | Zero-covered | File coverage |
|---|---|---|---|
| fieldsets_templ.go | 2,821 | 892 | 68.4% |
| screens_templ.go | 1,940 | 888 | 54.2% |
| widgets_templ.go | 665 | 305 | 54.1% |
| shell_templ.go | 771 | 257 | 66.7% |
| page_templ.go | 791 | 231 | 70.8% |
| root_templ.go | 329 | 98 | 70.2% |
| viewmodel_ops.go | 289 | 89 | 69.2% |
| fieldsets_codex_templ.go | 268 | 84 | 68.7% |
| icons_templ.go | 170 | 60 | 64.7% |
| browser.go | 12 | 5 | 58.3% |
| assets.go | 4 | 1 | 75.0% |
| all other non-templ files | 1,354 | 113 | mostly ≥85% |

### §A.3 Surface naming divergence (recorded, premise alive)

The card names the surface as "render/monitor". Verified against `git ls-tree 9d4a20eae -- internal/web/`: no files named `render.go` or `monitor.go` exist — only `render_helpers.go` (85.4% covered) and `schema_render_test.go`. The card's naming is CONCEPT-named, not file-named. The actual surface it maps to:

- the `*_templ.go` render-template family (the dominant uncovered mass, §A.2), plus
- the monitor screen: `screens.go:158 handleMonitor` (75.0%), `viewmodel_ops.go:772 buildMonitor` (100.0%), and the `Monitor` component at `screens_templ.go:2293` (inside screens_templ.go's 54.2%; its region 2296–2336 carries verified zero-covered blocks).

Premise alive (the monitor/render surface is real and under-tested); naming imprecise (recorded here so run-phase targets measured files, not the card's file names).

---

## §B Requirements (GEARS)

### §B.1 Baseline integrity

**REQ-001** — **When** run-phase begins on any tree, the lane shall re-measure `go test ./internal/web/ -coverprofile=<run-scratch>/baseline.out` and record the verbatim output before any test lands; **When** the re-measured figure diverges from the pinned baseline of record (67.9% @ `0314801c2`) by more than 1 percentage point, the lane shall re-pin the baseline in `.moai/specs/SPEC-WEB-CONSOLE-018/progress.md` §E.2 with the new figure, tree SHA, and command before writing tests.

### §B.2 Test quality — no coverage-filling

**REQ-002** — The suite shall consist exclusively of behavioral tests in which every test (or test group targeting one regression class) is answerable in one line: "this test catches \<specific defect\>".

**REQ-003** — The suite shall not contain any test whose only purpose is raising the measured coverage number; **When** a proposed test cannot name its defect class in one line, the lane shall discard it rather than land it.

**REQ-004** — The suite shall extend the existing render-assertion idioms already established in the package (`primary_surface_test.go`, `host_gate_test.go`, `i18n_chrome_wiring_test.go`, `schema_render_test.go`, `widget_states_test.go`) rather than inventing a new harness.

### §B.3 Planned test groups

Each group below carries its one-line defect answer (inherited verbatim by run-phase; §D acceptance verifies presence).

**REQ-005 (TG-1, Monitor screen state matrix)** — **When** the monitor screen is rendered across session-list and queue states (populated, empty, degraded), the rendered markup shall match the viewmodel state: correct table row content per session, an empty-state row when the list is empty, and `data-live`/`data-i18n` attributes present where the viewmodel marks them. *Defect class this group catches: monitor table rendering wrong row content, omitting the empty-state row, or dropping live/i18n wiring when viewmodel state changes (screens_templ.go Monitor region 2293+, `screens.go handleMonitor`, `viewmodel_ops.go buildMonitor` degradation paths).*

**REQ-006 (TG-2, Screens state matrices: Overview / Kanban / Todo / Specs)** — **While** the board, todo, and spec views carry non-uniform item states (active/done cards, badge states, close-debt and must-fix panels, spec detail), each screen component shall render the branch its state selects, including state badges (`todoStateBadge`), debt panels (`closeDebtPanel`, `mustFixPanel`), and `specDetail` rows. *Defect class: wrong conditional branch rendered for an item state — e.g. a done card rendering the active-badge branch, a must-fix panel omitting entries, `specRowLink` losing its target.*

**REQ-007 (TG-3, Fieldsets state matrix)** — **When** a fieldset is rendered with field error, invalid-value, default-selected, and banner states, the emitted markup shall carry the state-specific attributes and content: `fieldErr`/`fieldErrorMsg` present on error, `langOptionTags`/`optOptionTags` marking the selected option, `toggle` emitting its state attribute, `noteBanner`/`stateMark`/`backendBadge`/`stageMark`/`gauge`/`laneUnresolved`/`badge`/`spark`/`saveAction`/`bannerClass` matching their inputs, and the codex fieldset family (`fieldsetCodex`, `codexMirrorRowView`, `codexProbeBlock`) rendering probe/mirror rows per state. *Defect class: field-rendering regressions — error state rendering without its error message, a select losing its selected option, a toggle emitting the wrong state, a broken i18n key emitted, unescaped interpolation reaching markup.*

**REQ-008 (TG-4, Shell chrome + widgets)** — **When** the shell is rendered across navigation and live-state states, the rail/nav shall mark the active tab, `topbar`/`liveState` shall emit the current live state, `saveCluster` shall carry the save-failure banner hook, and widget components shall render their state matrices. *Defect class: navigation losing active-tab marking, live-state cluster emitting stale state, save cluster missing the failure banner surface (extends the SPEC-WEB-CONSOLE-017 observability contract).*

**REQ-009 (TG-5, Page header / profile actions / icons / root)** — **Where** profile mutability gates the page actions, `settingsPageHeader`/`profileModifiable`/`profileDeleteAction`/`profileCreateAction`/`profileRenameAction` shall render only the actions the mutability state permits, and the icon family (`iconSVG`, `icon`, `iconAt`, `iconPath`, `brandMark`) shall resolve each named icon rather than falling back silently to a missing glyph. *Defect class: a profile action rendering for the wrong permission state; an unknown icon name silently rendering an empty or wrong glyph.*

**REQ-010 (TG-6, viewmodel_ops uncovered branches)** — **When** viewmodel construction hits its uncovered edge branches (degradation paths, empty/absent-input branches — 89 zero-covered statements), the returned viewmodel shall carry the field values the screens distinguish. *Defect class: a viewmodel returning a degradation or edge value that the screen layer does not distinguish — wrong VM value flowing into render.* These are unit-level tests over real logic (not codegen) and follow the `viewmodel_degradation_test.go` idiom.

### §B.4 Coverage verdict

**REQ-011** — **When** all planned test groups land, the lane shall re-measure with the same command form as the baseline (`go test ./internal/web/... -cover` plus a `-coverprofile` run for the per-file table) and record the figure against the 85% target with a per-file residual table in progress.md §E.2.

**REQ-012** — **When** the measured figure falls short of 85% after all groups land, the lane shall report the shortfall with the per-file residual table and the regression classes not covered — and shall not add coverage-filling tests to close the gap; the shortfall path (report vs re-plan) is the orchestrator's decision, surfaced with the measured evidence.

**REQ-013** — **When** the measured figure meets or exceeds 85%, the lane shall verify REQ-002 conformance once more before declaring the target met: each landed test file still carries its per-group defect answer.

### §B.5 Exclusions recorded in-spec

**REQ-014** — The SPEC shall record `browser.go openDefaultBrowser` and `assets.go staticFS` as coverage-excluded with rationale (§E), and the verdict arithmetic shall state their residual impact explicitly (6 of 3,023 zero-covered statements — immaterial).

---

## §C Constraints

- Run-phase `cycle_type=tdd`; the deliverable is tests only. No product (non-`_test.go`) code changes — the package's injection seams are NOT created by this SPEC (see §E exclusions).
- All tests run scoped to `./internal/web/...`; no full-suite local runs (repo test-discipline rule; CI owns the full verdict).
- Test file naming follows the package's existing `snake_case_test.go` convention; test bodies in English.
- Baseline evidence lives at `.moai/state/verify/t1079/cover-baseline.out` (machine-local scratch, regenerable); the run-phase evidence path is `.moai/specs/SPEC-WEB-CONSOLE-018/progress.md` §E.2.

---

## §D Acceptance Criteria (summary)

Full Given-When-Then enumeration in `acceptance.md`. Summary:

| AC | Subject | Verdict form |
|---|---|---|
| AC-001 | Baseline re-pinned at run start | verbatim command output in progress.md §E.2 |
| AC-010..AC-015 | TG-1..TG-6 tests exist, pass, and carry per-group defect answers | `go test ./internal/web/...` PASS + grep for defect-class comment per test group |
| AC-020 | No coverage-filling tests; idiom conformance (REQ-004) | every landed test maps to a named regression class; idiom files' helpers reused, no new harness |
| AC-030 | Coverage verdict measured with the baseline command | figure + per-file residual table in progress.md §E.2 |
| AC-031 | 85% met, or documented shortfall without filler | ≥85% figure, OR shortfall report + zero unnamed tests |
| AC-040 | Product code untouched | `git diff --name-only` shows `_test.go` files only |

---

## §E Out of Scope and Exclusions

### Out of Scope — Product code changes

- No changes to any non-`_test.go` file in `internal/web` or elsewhere; no new injection seams (see browser.go exclusion below); no template source (`.templ`) edits.

### Out of Scope — browser.go / assets.go coverage

- `browser.go openDefaultBrowser` (5 zero-covered statements, 0.05% of package) is an OS browser-exec seam. Covering it requires either a product-code injection seam (forbidden by this SPEC's tests-only scope) or launching a real browser in CI (environment-dependent flake). Excluded with this rationale; residual impact on the 85% arithmetic: immaterial (5 statements of 9,414).
- `assets.go staticFS` (1 zero-covered statement) is an `embed.FS` wrapper with no branch logic to assert behaviorally. Excluded; residual impact: 1 statement.

### Out of Scope — Non-web packages

- Only `internal/web` is in scope; no coverage work in any other package, even where the toolchain reports adjacent gaps.

### Out of Scope — E2E / browser automation

- No headless-browser or screenshot-based verification; assertions are Go-level markup assertions against rendered component output, per the existing in-package idiom.

### Out of Scope — Coverage-filling under any name

- Snapshot tests of generated markup with no distinguishing assertion, loops that render a component N times with no per-state assertion, and any test whose failure cannot be attributed to a named regression class are out of scope regardless of their effect on the number.
