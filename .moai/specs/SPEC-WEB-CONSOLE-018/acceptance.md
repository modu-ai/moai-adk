# SPEC-WEB-CONSOLE-018 — Acceptance Criteria

Verification layer. Every AC is binary-testable. GEARS obligations live in spec.md §B; this file owns the Given-When-Then evidence.

## §D AC Matrix

| AC | Milestone | Severity | Verification |
|---|---|---|---|
| AC-001 | M1 | must | command output present in progress.md §E.2 |
| AC-010 | M2 | must | test run + defect-answer grep |
| AC-011 | M3 | must | test run + defect-answer grep |
| AC-012 | M4 | must | test run + defect-answer grep |
| AC-013 | M5 | should | test run + defect-answer grep |
| AC-014 | M6 | should | test run + defect-answer grep |
| AC-015 | M7 | should | test run + defect-answer grep |
| AC-020 | M2–M7 | must | manual + grep audit of landed test files (regression-class answers + idiom reuse, no new harness) |
| AC-030 | M8 | must | verdict output in §E.2 |
| AC-031 | M8 | must | figure ≥85% OR documented shortfall |
| AC-040 | M8 | must | `git diff --name-only` filter |
| AC-041 | M1–M8 | must | exclusions present in spec §E and unviolated |

## §D.1 AC-001 — Baseline re-pinned at run start

**Given** the run-phase lane is entered on tree `<run-start SHA>`
**When** the lane executes `go test ./internal/web/ -coverprofile=<scratch>/baseline.out` (single invocation; `-coverprofile` still prints the `coverage:` summary line, and `internal/web` has no subpackages so the `/...` form adds nothing)
**Then** the verbatim output (coverage figure + tree SHA + command) is recorded in progress.md §E.2 before any `_test.go` file is created; and the recorded figure is within 1 percentage point of 67.9% or a re-baseline note explains the divergence.

## §D.2 AC-010 — TG-1 Monitor state matrix lands

**Given** the package test suite is green on the baseline tree
**When** the TG-1 test file(s) land and `go test ./internal/web/... -run <TG1-pattern> -count=1` executes
**Then** the run PASSES; and the rendered-state assertions cover at minimum: populated session table (row content per session), empty list (empty-state row present), degraded viewmodel state (distinguishing markup); and the file header carries the group defect answer ("catches monitor table rendering wrong row content / omitting empty-state / dropping live-i18n wiring").

## §D.3 AC-011 — TG-2 Screens state matrices land

**Given** the package test suite is green
**When** the TG-2 test file(s) land and run
**Then** the run PASSES; per-state branch assertions exist for Overview/Kanban/Todo/Specs including `todoStateBadge` badge states, `closeDebtPanel`/`mustFixPanel` panel content, `specDetail`/`specRowLink` rows; and the defect answer is in the file header.

## §D.4 AC-012 — TG-3 Fieldsets state matrix lands

**Given** the package test suite is green
**When** the TG-3 test file(s) land and run
**Then** the run PASSES; assertions exist for error-state (`fieldErr`/`fieldErrorMsg` present on error input, absent otherwise), selected-option marking (`langOptionTags`/`optOptionTags`), toggle state attribute, banner/state-mark components matching inputs, and the codex family (`fieldsetCodex`, `codexMirrorRowView`, `codexProbeBlock`); defect answer in file header.

## §D.5 AC-013 — TG-4 Shell chrome + widgets land

**Given** the package test suite is green
**When** the TG-4 test file(s) land and run
**Then** the run PASSES; assertions exist for active-tab rail/nav marking, `liveState` output matching input state, `saveCluster` carrying the save-failure banner hook; defect answer in file header.

## §D.6 AC-014 — TG-5 Page/icons/root land

**Given** the package test suite is green
**When** the TG-5 test file(s) land and run
**Then** the run PASSES; assertions exist for profile action gating (each of delete/create/rename rendered only when mutability permits) and icon resolution (known icon renders its glyph; unknown name does NOT render a wrong glyph silently); defect answer in file header.

## §D.7 AC-015 — TG-6 viewmodel edge branches land

**Given** the package test suite is green
**When** the TG-6 test file(s) land and run
**Then** the run PASSES; unit assertions cover the degradation/empty-input branches of `viewmodel_ops.go` following the `viewmodel_degradation_test.go` idiom; defect answer in file header.

## §D.8 AC-020 — No coverage-filling tests

**Given** all groups M2–M7 have landed
**When** each new test file is read against its header defect answer
**Then** every test maps to a named regression class; a test whose failure could not be attributed to that class does not exist; no assertion-free render loops or assertion-free snapshots exist; and the landed files reuse the helpers of the named idiom files (spec REQ-004: `primary_surface_test.go`, `host_gate_test.go`, `i18n_chrome_wiring_test.go`, `schema_render_test.go`, `widget_states_test.go`) — no new test harness introduced. Binary probe: for each new test, the question "what product regression does this fail on?" has a one-line answer in the file.

## §D.9 AC-030 — Coverage verdict measured with the baseline command

**Given** M2–M7 are complete
**When** `go test ./internal/web/ -coverprofile=<scratch>/verdict.out` executes (single invocation — same form as AC-001)
**Then** the figure and a per-file residual table (file / stmts / zero / coverage — same shape as spec §A.2) are recorded in progress.md §E.2.

## §D.10 AC-031 — 85% met, or documented shortfall without filler

**Given** the AC-030 verdict is recorded
**When** the figure is read against the 85% target
**Then** either the figure is ≥85.0% and AC-020 re-verification (REQ-013) passed, or the figure is <85.0% and the shortfall report (residual table + uncovered regression classes + no filler confirmation) is recorded — and the run does NOT add filler tests to force the number.

## §D.11 AC-040 — Product code untouched

**Given** all milestones are complete and the spec.md §C tests-only constraint (deliverable is `_test.go` files exclusively) is in force
**When** `git diff --name-only <base>..HEAD -- internal/ | grep -v _test.go` executes
**Then** it outputs nothing (all changes under `internal/web` are `_test.go` files).

## §D.12 AC-041 — Exclusions recorded and unviolated

**Given** spec.md §E records the browser.go/assets.go exclusions with rationale
**When** the landed test files are inspected
**Then** no test launches a browser, no injection seam was added to `browser.go` or `assets.go`, and the verdict report states their residual arithmetic impact (6 of 3,023 zero-covered statements).

## §D.13 Indirect verification

- Coverage movement per group is indirect evidence: each group's file(s) must show a non-zero zero-statement reduction in the M8 verdict profile vs the M1 baseline profile (`go tool cover` diff by file). A group whose file shows zero movement is investigated before the verdict is accepted.

## §D.14 Closure gates

- Definition of Done: AC-001 + AC-010..AC-015 (must-severity subset) + AC-020 + AC-030 + AC-031 + AC-040 + AC-041 all hold; progress.md §E.2 carries the evidence trail (baseline output, verdict output, per-file residual table).
- Quality gate: `go vet ./internal/web/...`, `golangci-lint run ./internal/web/...`, scoped test run — all clean.

## §D.15 Forward-looking checks

- The per-group defect-answer header comments remain the maintenance contract: future test additions to these files inherit the answer-first rule (card t1079 [HARD-3] persists past this SPEC's close).
