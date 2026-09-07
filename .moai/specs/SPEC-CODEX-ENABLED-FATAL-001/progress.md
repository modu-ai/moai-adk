# SPEC-CODEX-ENABLED-FATAL-001 — Progress

card: t508 · branch `WT-codex-enabled-guard` · base `0b1e27877` (= origin/develop)

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts**: `spec.md`, `plan.md`, `acceptance.md`, `research.md`, `progress.md` — written
  2026-09-07 by manager-spec. Tier S, with `acceptance.md` and `research.md` emitted deliberately
  (the criteria matrix and the evidence derivation do not compress into the spec body).
- **SPEC ID check**: executed as Bash, verbatim output `PASS`.
- **ID uniqueness**: confirmed against `.moai/specs/` — 24 existing `SPEC-CODEX-*` directories,
  none named `SPEC-CODEX-ENABLED-FATAL-001`.
- **Requirements**: 12 (REQ-CEF-001..012), GEARS notation.
- **Acceptance criteria**: 13 (AC-CEF-001..013), each Given-When-Then with a named verify command.
  Four are mandatory controls (AC-CEF-005, 006, 007, 009).
- **Evidence basis**: `.moai/reports/t508/codex-enabled-lab.md` and
  `.moai/reports/t508/red-baseline.md`. Structural claims independently re-read from source in the
  plan-phase session; codex-behaviour claims carried with attribution to the lab's run.
- **Template-First verdict**: no template change expected — `grep -rl 'skills\.config'
  internal/template/templates/` exits 1 (zero matches) against a non-vacuous control
  (`grep -rl 'moai'` → 396 files); no `.go` files under templates.
- **Operator decisions**: B.1 (scope: both fatal shapes, reversing the quoted-value leniency) and
  B.2 (severity: `uikit.CheckFail`, doctor exits 1) recorded with attribution to the 2026-09-07
  lane-3 session. Not re-opened.
- **Open clarifications**: none. Unmeasured items are recorded as Gaps (acceptance.md §D.4), not as
  `[NEEDS CLARIFICATION]` markers — none of them changes what gets built.

_Status: draft. Awaiting plan-audit and Implementation Kickoff Approval._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
