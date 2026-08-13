---
id: SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001
title: "Close-Tool 3-Phase Precondition Relaxation — close the last §E.5 gap in the 3-phase lifecycle"
version: "0.1.0"
status: completed
created: 2026-08-13
updated: 2026-08-13
author: manager-spec
priority: P0
phase: "v3.1.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "lifecycle, close, 3-phase, precondition, closer, backward-compat"
era: V3R6
tier: M
related_specs: [SPEC-V3R6-LIFECYCLE-REDESIGN-001, SPEC-V3R6-LIFECYCLE-SYNC-GATE-001]
---

# SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001

## §A. Problem Statement

The SPEC-V3R6-LIFECYCLE-REDESIGN-001 redesign migrated `internal/spec/audit.go` and `internal/spec/migrate_3phase.go` to the canonical 3-marker predicate (`§E.2` run-evidence start + `§E.4` sync marker + `sync_commit_sha`), retiring the `§E.5` Mx-phase section and folding it into `§E.4`. The schema (`spec-frontmatter-schema.md`) now **forbids** authoring new `§E.5` sections.

**One file was missed: `internal/spec/closer.go`.** Its Precondition 2 still requires a `§E.5` Mx section to be present in `progress.md` for a SPEC to close:

```go
// internal/spec/closer.go:696-699
if !state.HasMxSection {
    failed = append(failed, "missing §E.5 Mx-phase audit-ready signal in progress.md")
}
```

The result is a structural contradiction: the schema forbids writing `§E.5`, but the closer requires `§E.5` present. Any 3-phase SPEC (no `§E.5`) is therefore **blocked from closing**. This is the last hole in the 3-phase lifecycle redesign.

Real-world impact: 4 SPECs in the mo.ai.kr checkout are stuck at `status: implemented`, blocked from close by this defect (SPEC-DB-BACKLOG-001, SPEC-DB-CLEANUP-001, SPEC-DESIGN-HANDOFF-NEWS-001, SPEC-ID-SEC-MEDIUM-001). `moai spec close SPEC-DB-CLEANUP-001 --dry-run` reports `"missing §E.5 Mx-phase audit-ready signal"`.

## §B. Context

### §B.1 The Contradiction (verified this session)

- `internal/spec/closer.go:696-699` — Precondition 2 requires `state.HasMxSection` (a `§E.5` section).
- `internal/spec/closer.go:571,611` — `HasMxSection` is populated by `hasProgressMarker(body, "§E.5")`.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — `§E.5` is **RETIRED** ("do NOT author new §E.5 sections; retained in the parser for backward-compat classification of legacy SPECs"; "folded into §E.4").
- Net: a 3-phase SPEC (which honors the schema by omitting `§E.5`) cannot pass Precondition 2.

### §B.2 The Relaxation Pattern (verified reference inside the same package)

`internal/spec/transitions.go:79-92` already implements the exact dual-acceptance idea this SPEC extends — `closeInfixMatch` accepts BOTH `closeInfix3Phase` ("3-phase close", the new canonical infix per REQ-LR-020 of the parent redesign) AND `closeInfix4Phase` ("4-phase close", legacy retained so the drift walker still recognizes historical close commits in git history). The same dual-acceptance shape applies cleanly to Precondition 2:

- `§E.5` present → legacy path OK (backward-compat preserved for grandfather-era SPECs that still carry `§E.5`).
- `§E.5` absent BUT 3-marker present (`§E.4` sync marker + non-empty `sync_commit_sha`) → 3-phase path OK.

This is a **relaxation, not a removal**. The legacy `§E.5` path stays accepted.

### §B.3 Files Already 3-Phase-Ready (DO NOT TOUCH — out of scope)

- `internal/spec/audit.go:359-384` — already migrated to the 3-marker predicate (owned by REDESIGN-001; the `§E.5`/`mx_commit_sha`-keyed findings are RETIRED).
- `internal/spec/era.go:135-167` — `§E.5` literal-heading recognition is **INTENTIONALLY PRESERVED** for legacy era classification (the H-4-legacy migration-window dual predicate: `§E.5 + mx_commit_sha`). Removing `§E.5` from `era.go` would silently break era classification of legacy SPECs. See `spec-frontmatter-schema.md:97-105` for the parser-load-bearing explanation.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — already documents the `§E.5` retirement; no schema change needed.
- `internal/spec/transitions.go` — already dual-infix; no change needed.

### §B.4 Subject-Template / Comment Drift (additional canonicalization targets)

Beyond the L696-699 precondition, three more sites in `closer.go` carry stale "4-phase" / `§E.5` wording that must be canonicalized to the 3-phase convention in the same run-phase pass:

- `internal/spec/closer.go:350` — commit-subject template `fmt.Sprintf("chore(%s): Mx-phase audit-ready signal + 4-phase close", specID)` — canonical is `"3-phase close"` per REQ-LR-020.
- `internal/spec/closer.go:107` — doc comment `"Close orchestrates the atomic 4-phase close transition for a SPEC."`.
- `internal/spec/closer.go:683` — doc comment `"validatePreconditions checks the 4-phase precondition matrix per AC-LSG-006."`.
- `internal/spec/closer.go:90` — sentinel-error comment referencing `§E.5`.

### §B.5 Parent Contract

This SPEC extends `SPEC-V3R6-LIFECYCLE-REDESIGN-001` (REQ-LR-020/021, AC-LR-012, D4 reconciliation). The parent owns the close-infix convention and the 3-marker predicate contract for `audit.go` / `era.go` / `transitions.go`. This SPEC's scope is strictly local to `closer.go`'s precondition and subject/comment canonicalization — it does NOT modify the parent's owned files.

## §C. Requirements (GEARS)

### REQ-LSG-001 (Event-detected) — 3-phase Precondition 2 path

**When** a 3-phase `progress.md` is evaluated by the closer (i.e. `§E.5` absent AND `§E.4` sync marker present AND `sync_commit_sha` non-empty), the closer **shall** pass Precondition 2 without appending any `"missing §E.5"` failure to `PreconditionsFailed`.

### REQ-LSG-002 (State-driven) — Legacy backward-compat path retained

**While** a legacy `progress.md` carries a `§E.5` section (grandfather-era SPECs that still author `§E.5`), the closer **shall** continue to pass Precondition 2 for that SPEC — the relaxation MUST NOT remove the legacy `§E.5` acceptance path.

### REQ-LSG-003 (Unwanted) — era.go §E.5 recognition preserved

The closer relaxation **shall not** modify `internal/spec/era.go` — the `§E.5` literal-heading token recognized by `hasProgressMarker(content, "§E.5")` and the H-4-legacy predicate (`§E.2 + §E.5 + both commit_sha`) in `era.go:135-167` MUST remain intact so legacy SPEC era classification does not shift.

### REQ-LSG-004 (Event-driven) — Commit-subject template canonicalization

**When** the closer emits its close commit (`closer.go:350`), the subject template **shall** use the canonical `"3-phase close"` infix (`fmt.Sprintf("chore(%s): Mx-phase audit-ready signal + 3-phase close", specID)`), consistent with `closeInfix3Phase` and REQ-LR-020 of the parent redesign.

### REQ-LSG-005 (Ubiquitous) — Doc-comment canonicalization

The closer doc comments at `closer.go:107` ("4-phase close transition") and `closer.go:683` ("4-phase precondition matrix"), plus the `closer.go:90` sentinel-error comment referencing `§E.5`, **shall** be canonicalized to the 3-phase convention so the file's prose no longer contradicts the schema's `§E.5` retirement.

### REQ-LSG-006 (Event-detected) — TDD test conversion

**When** `TestClose_PreconditionMissingMx` (`closer_test.go:47-86`) is converted, the `§E.5`-absent case **shall** become a PASS assertion (3-phase SPEC closes), and a new test asserting that a 3-phase `progress.md` (`§E.4` + `sync_commit_sha`, `§E.5` absent) passes Precondition 2 **shall** be added. The legacy `§E.5`-present PASS case **shall** remain covered by a sibling test so backward-compat is regression-guarded.

### REQ-LSG-007 (Unwanted) — No regression

The relaxation **shall not** regress `go build ./...`, `go test ./internal/spec/...`, or the full `go test ./...` suite. The legacy `§E.5`-present close path, the atomic-rollback invariant (AC-LSG-014), the no-op invariant (AC-LSG-018), and the drift walker's close recognition (`transitions.go closeInfixMatch`) MUST remain green and behaviorally unchanged.

### REQ-LSG-008 (State-driven) — Deferred real-world validation recorded

**While** this SPEC's run-phase scope is limited to the `moai-adk-go` codebase, the acceptance evidence for the 4 blocked mo.ai.kr SPECs **shall** be recorded in `acceptance.md` as a DEFERRED validation (command + expected output) — not executed during run phase — because it requires a `make build` reinstall of `~/go/bin/moai` and a checkout switch to mo.ai.kr.

## §D. Constraints

- **TDD cycle** (`cycle_type=tdd`): RED first (write/convert the `§E.5`-absent close test, confirm it FAILS on current code) → GREEN (relax Precondition 2) → REFACTOR. RED evidence recorded in `plan.md`.
- **Backward-compat preserved absolutely**: the legacy `§E.5`-present path MUST stay accepted. This is a relaxation (OR-in a new path), not a replacement.
- **Scope discipline**: the relaxation is local to `closer.go`'s Precondition 2 (+ the L107/L350/L683/L90 canonicalization). It is NOT a lifecycle-wide `§E.5` purge. `era.go`, `audit.go`, `transitions.go`, and `spec-frontmatter-schema.md` are out of scope.
- **Code style**: match existing `closer.go` conventions — `fmt.Errorf("...: %w", err)` wrapping, English comments (`code_comments: en`), naming density, no new exported symbols unless required.
- **Plan phase only**: this artifact set carries `status: draft`. No code, no commits. Code + commits belong to run phase.

## §E. AC Matrix (summary — full Given-When-Then in acceptance.md)

| AC | REQ | Severity | Phase |
|----|-----|----------|-------|
| AC-LSG-001 | REQ-LSG-001 | MUST-PASS | M1 (TDD pivot) |
| AC-LSG-002 | REQ-LSG-002 | MUST-PASS | M1 |
| AC-LSG-003 | REQ-LSG-003 | MUST-PASS | M2 (scope guard) |
| AC-LSG-004 | REQ-LSG-004, REQ-LSG-005 | MUST-PASS | M1 |
| AC-LSG-005 | REQ-LSG-007 | MUST-PASS | M2 |
| AC-LSG-006 | REQ-LSG-006 | MUST-PASS | M1 (TDD RED→GREEN) |
| AC-LSG-007 | REQ-LSG-008 | DEFERRED | post-merge |

## §F. Scope (files this SPEC MAY modify in run phase)

1. `internal/spec/closer.go` — relax Precondition 2 to dual-acceptance; make `HasMxSection` logic 3-phase-aware (add a sibling 3-marker check OR widen Precondition 2's predicate); canonicalize the L350 subject template to `"3-phase close"`; update the L90, L107, L683 comments.
2. `internal/spec/closer_test.go` — convert `TestClose_PreconditionMissingMx` (L47-86): the `§E.5`-absent case must now PASS (3-phase); add a sibling test asserting a legacy `§E.5`-present SPEC still closes (backward-compat guard); add a new test asserting a 3-phase `progress.md` (`§E.4` + `sync_commit_sha`, `§E.5` absent) passes Precondition 2.

## §G. Out of Scope

### Out of Scope — Files already 3-phase-ready (owned by the parent redesign)

- `internal/spec/audit.go` — already migrated to the 3-marker predicate (REDESIGN-001). No change.
- `internal/spec/era.go` — `§E.5` recognition INTENTIONALLY PRESERVED for legacy era classification (H-4-legacy dual predicate). Removing it would silently misclassify legacy SPECs. No change.
- `internal/spec/transitions.go` — already dual-infix (`closeInfix3Phase` + `closeInfix4Phase`). No change.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — already documents the `§E.5` retirement. No schema change.

### Out of Scope — Real-world mo.ai.kr close execution

- The actual `moai spec close SPEC-DB-CLEANUP-001 --dry-run` execution against the mo.ai.kr checkout is DEFERRED validation (AC-LSG-007). It requires a `make build` reinstall and a checkout switch, and is recorded as command+expected-output only — not executed during run phase.

### Out of Scope — Lifecycle-wide §E.5 purge

- This SPEC does NOT purge `§E.5` from the codebase. `§E.5` remains a load-bearing legacy token in `era.go`, `transitions.go` (via `closeInfixMx` / legacy close commits), and the parser. The relaxation is strictly local to `closer.go`'s Precondition 2 predicate and the L90/L107/L350/L683 prose canonicalization.

## §H. History

- 2026-08-13: plan-phase v0.1.0 authored. Defect location, contradiction, relaxation pattern, parent contract, and scope boundaries verified by direct source inspection of `closer.go` (L90, L107, L350, L571, L611, L683, L696-699), `transitions.go` (L79-92), `audit.go` (L359-384), `era.go` (L135-167), and `spec-frontmatter-schema.md` (§E.5 retirement). SPEC ID `SPEC-V3R6-LIFECYCLE-CLOSE-3PHASE-001` was rejected by the pre-write regex self-check (segment `3PHASE` starts with a digit); user-confirmed canonical form `SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001` adopted.
