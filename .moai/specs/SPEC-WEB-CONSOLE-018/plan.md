# SPEC-WEB-CONSOLE-018 — Implementation Plan

Card t1079 (Class C: plan→run→sync). Run-phase `cycle_type=tdd`. Tier M.

## §A Context

`internal/web` measures 67.9% statement coverage against the 85% package target (`.moai/config/sections/quality.yaml` `test_coverage_target: 85`). The dominant uncovered mass is GENERATED templ codegen output (2,815 of 3,023 zero-covered statements across 8 `*_templ.go` files). Line-by-line coverage of generated markup is exactly what card t1079 [HARD-3] forbids; the plan therefore shapes all templ-directed work as **behavioral render assertions** — render component C under state S, assert distinguishing markup present or absent — extending the render-assertion idioms the package already has (`primary_surface_test.go`, `host_gate_test.go`, `i18n_chrome_wiring_test.go`, `schema_render_test.go`, `widget_states_test.go`, `viewmodel_degradation_test.go`).

## §B Known Issues

- The card text names the surface "render/monitor"; no `render.go`/`monitor.go` files exist (verified: `git ls-tree 9d4a20eae -- internal/web/`). The real surface is the `*_templ.go` family plus the Monitor component at `screens_templ.go:2293` (zero-covered blocks verified in the 2296–2336 region) and `screens.go handleMonitor`. Spec §A.3 records this divergence; run-phase targets measured files.
- Baseline provenance: the 67.9% figure circulating from the t1051 sync-audit F2 is HISTORICAL (tree `WT-save-observability @ 9d4a20eae`). The baseline of record is the 2026-09-22 remeasurement on `WT-web-coverage @ 0314801c2` (`go test ./internal/web/... -cover` → `coverage: 67.9% of statements`). The numeric coincidence is recorded, not relied on.

## §C Pre-flight (M1 gate inputs)

1. Confirm starting tree: `git rev-parse --short HEAD` recorded alongside the baseline measurement.
2. Baseline command: `go test ./internal/web/ -coverprofile=<scratch>/baseline.out` (single invocation — `-coverprofile` still prints the `coverage:` summary line; `internal/web` has no subpackages) — verbatim output into progress.md §E.2.
3. Reference profile: `.moai/state/verify/t1079/cover-baseline.out` (9,414 stmts / 3,023 zero / 67.9%).

## §D Constraints

- Tests only: every landed file is `_test.go`. Product code, `.templ` sources, and injection seams are untouched (spec §E Out of Scope).
- Verification scoped to `./internal/web/...` (no full-suite local runs).
- [HARD-3] enforced at review: each test group carries its one-line defect answer; a test that cannot name its regression class is discarded (spec REQ-003).

## §E Self-Verification (run-phase §E evidence duties)

- E1: `go test ./internal/web/... -cover` verbatim output (baseline M1 + verdict M8).
- E5: `golangci-lint run ./internal/web/...` clean on the new test files.
- E2: n/a (no cross-platform build surface — tests only).
- Evidence path: `.moai/specs/SPEC-WEB-CONSOLE-018/progress.md` §E.2.

## §F Milestones

Ordered: baseline assertion first, then test groups by regression-class value (monitor — the card-named surface — first), then the coverage verdict. No time estimates.

| # | Priority | Milestone | Scope | Done when |
|---|----------|-----------|-------|-----------|
| M1 | High | **Baseline re-pin** — re-measure on the run-start tree; pin figure + SHA + command in progress.md §E.2; re-baseline if divergent >1pt from 67.9% @ `0314801c2` | no code | verbatim baseline recorded |
| M2 | High | **TG-1 Monitor state matrix** — behavioral render assertions over `Monitor` (screens_templ.go:2293+) across session-list/queue states: populated rows, empty-state, degradation; extends `screen_chrome_test.go`/`session_telemetry_cells_test.go` idioms | new `_test.go` | group passes; defect answer in file header comment |
| M3 | High | **TG-2 Screens state matrices** — Overview/Kanban/Todo/Specs + `todoStateBadge`, `closeDebtPanel`, `mustFixPanel`, `specDetail`, `specRowLink` per-state branch assertions | new `_test.go` | group passes; per-state assertion present |
| M4 | High | **TG-3 Fieldsets state matrix** — error/select/toggle/banner states across `fieldsets_templ.go` components + codex family (`fieldsetCodex`, `codexMirrorRowView`, `codexProbeBlock`); extends `schema_sections_test.go`/`fieldsets`-adjacent idioms | new `_test.go` | group passes; error-state and selected-option assertions present |
| M5 | Medium | **TG-4 Shell chrome + widgets** — `Shell`/`rail`/`navRow`/`topbar`/`liveState`/`saveCluster` + `widgets_templ.go` state matrices; active-tab marking and save-failure banner hook asserted (extends `i18n_chrome_wiring_test.go`, `save_observability_test.go` contracts) | new `_test.go` | group passes |
| M6 | Medium | **TG-5 Page/icons/root** — profile action gating per mutability state; icon family resolution vs silent-missing fallback; `root_templ.go` branches | new `_test.go` | group passes |
| M7 | Medium | **TG-6 viewmodel_ops edge branches** — unit tests over the 89 zero-covered statements (degradation/empty-input paths); extends `viewmodel_degradation_test.go` idiom | new `_test.go` | group passes |
| M8 | High | **Coverage verdict** — same command as M1; per-file residual table vs 85% in §E.2; on shortfall: report residuals + uncovered regression classes, no filler (REQ-012); on success: re-verify REQ-002 conformance (REQ-013) | no code | verdict recorded |

### Coverage arithmetic adopted (measured, this tree)

- Package: 9,414 statements, 3,023 zero-covered, 67.9%. Target 85% ⇒ covered ≥ 8,002 ⇒ **must convert +1,611 statements (≈53% of the current zero mass)**.
- Mass addressed by groups (zero-covered statements): TG-2/TG-1 screens_templ 888; TG-3 fieldsets_templ 892 + fieldsets_codex_templ 84; TG-4 shell_templ 257 + widgets_templ 305; TG-5 page_templ 231 + root_templ 98 + icons_templ 60; TG-6 viewmodel_ops 89. Total addressable: **2,904**.
- Excluded and immaterial: browser.go 5 + assets.go 1 (+ 113 residual zero statements in already-mostly-≥85% non-templ files, addressable only by number-filling — not attempted).
- Honesty clause: whether behavioral state matrices convert ≥1,611 of the 2,904 is an EMPIRICAL question this plan cannot decide at plan phase. A full-state render of a component typically covers most of its codegen blocks along the exercised path, which is why the mass is addressable — but the verdict is M8's measurement, not this table. M8 follows REQ-012: a measured shortfall is reported with residuals; the gap is never closed with filler.

## §G Anti-Patterns (forbidden in run-phase)

- Rendering a component N times in a loop without per-state distinguishing assertions.
- Snapshotting generated markup with no failure-attributable assertion.
- Testing `browser.go openDefaultBrowser` by launching a browser, or adding an injection seam to it.
- Any `_test.go`-adjacent product change (helper moved into non-test file, seam var, `.templ` edit) smuggled in to make a test possible.
- Re-running `go test ./...` locally to "confirm" the verdict — CI owns the full suite.

## §H Cross-References

- Card t1079 (this SPEC's origin; [HARD-1..3] constraints carried in spec §A.1/§B.2).
- SPEC-WEB-CONSOLE-017 (t1051, closed) — source of the HISTORICAL audit figure; not modified by this SPEC.
- Baseline evidence: `.moai/state/verify/t1079/cover-baseline.out` (regenerable per §C.2).
- `.moai/config/sections/quality.yaml` — `test_coverage_target: 85`.
