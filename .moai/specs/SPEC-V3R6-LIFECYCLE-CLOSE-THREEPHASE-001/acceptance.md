# Acceptance — SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001

> Acceptance criteria are binary-testable Given-When-Then scenarios. The GEARS obligation lives in `spec.md` §C (REQ-LSG-*); this file is the verification layer.

## §A. Severity & Traceability

| AC | REQ | Severity | Milestone |
|----|-----|----------|-----------|
| AC-LSG-001 | REQ-LSG-001 | MUST-PASS | M1 |
| AC-LSG-002 | REQ-LSG-002 | MUST-PASS | M1 |
| AC-LSG-003 | REQ-LSG-003 | MUST-PASS | M2 (scope guard) |
| AC-LSG-004 | REQ-LSG-004, REQ-LSG-005 | MUST-PASS | M1 |
| AC-LSG-005 | REQ-LSG-007 | MUST-PASS | M2 |
| AC-LSG-006 | REQ-LSG-006 | MUST-PASS | M1 (TDD RED→GREEN) |
| AC-LSG-007 | REQ-LSG-008 | DEFERRED | post-merge |

## §B. Indirect verification & scope guards

- **AC-LSG-003 (scope guard)** is verified by a baseline-delta grep, not by a behavioral test: the count of `§E.5` matches in `internal/spec/era.go` MUST be unchanged across the run-phase diff, and `closeInfix3Phase` + `closeInfix4Phase` MUST both remain present in `transitions.go`. The relaxation does not touch those files.
- **AC-LSG-005 (no regression)** is verified by the full `go test ./...` suite — the relaxation's blast radius is contained because `era.go`, `audit.go`, and `transitions.go` are out of scope.

## §C. Given-When-Then Scenarios

### AC-LSG-001: 3-phase progress.md passes Precondition 2 (§E.5 absent)

**Given** a `progress.md` carrying `## §E.4 Sync-phase Audit-Ready Signal` with a non-empty `sync_commit_sha` field, and NO `§E.5` section
**When** `Close(specID, CloseOptions{...})` evaluates `validatePreconditions`
**Then** Precondition 2 does NOT append `"missing §E.5 Mx-phase audit-ready signal"` to `PreconditionsFailed`
**And** when all other preconditions are met, `Close` proceeds (does not return `ErrPreconditionMissing`)

**Verification**: `go test ./internal/spec/ -run TestClose_ThreePhasePreconditionPass -v` exits 0.

### AC-LSG-002: Legacy §E.5 progress.md still passes Precondition 2 (backward-compat)

**Given** a legacy `progress.md` carrying `## §E.5 Mx-phase Audit-Ready Signal` (grandfather-era SPEC)
**When** `Close(specID, CloseOptions{...})` evaluates `validatePreconditions`
**Then** Precondition 2 does NOT append any `"missing §E.5"` failure
**And** the close proceeds when all other preconditions are met — backward-compat preserved

**Verification**: `go test ./internal/spec/ -run TestClose_LegacyMxSectionStillAccepted -v` exits 0.

### AC-LSG-003: era.go §E.5 recognition unchanged (scope guard)

**Given** the run-phase diff touches only `internal/spec/closer.go` and `internal/spec/closer_test.go`
**When** `grep -c '§E.5' internal/spec/era.go` is run before and after the diff
**Then** the match count is unchanged
**And** `grep -n 'closeInfix3Phase\|closeInfix4Phase' internal/spec/transitions.go` still shows both constants present

**Verification**: baseline-delta grep (M2 pre-flight vs post-edit). No behavioral test — the guard is a diff-scope assertion.

### AC-LSG-004: Commit-subject template + doc comments canonicalized to 3-phase

**Given** the run-phase diff is applied
**When** `grep -n '4-phase close\|4-phase precondition\|4-phase close transition' internal/spec/closer.go` is run
**Then** zero matches
**And** `grep -n '3-phase close' internal/spec/closer.go` returns ≥1 match at the subject-template site (L350)
**And** `grep -n '3-phase' internal/spec/closer.go` returns matches at the L107 and L683 doc-comment sites

**Verification**: post-edit grep. Covers REQ-LSG-004 (subject template) + REQ-LSG-005 (doc comments).

### AC-LSG-005: No regression — build + full test suite green

**Given** the run-phase diff is applied
**When** `go build ./...` and `go test ./...` are run
**Then** both exit 0
**And** the load-bearing existing tests remain green: `TestClose_BackfillOnly_CompletedProductionVariantsAreNoOp` (AC-LSG-018 no-op invariant), `TestClose_FullClose_AtomicRollbackOnFailure` (AC-LSG-014 atomic-rollback invariant), `TestClose_FullClose_ProducesCommit` (full-close commit emission), and `closeInfixMatch` recognition in `transitions_test.go`

**Verification**: `go build ./... && go test ./...` exit 0.

### AC-LSG-006: TDD RED→GREEN pivot — TestClose_PreconditionMissingMx converted

**Given** `closer_test.go:47-86` `TestClose_PreconditionMissingMx` currently asserts the `§E.5`-absent case FAILS (returns `ErrPreconditionMissing` with a `§E.5`-mentioning failure)
**When** the TDD cycle is executed
**Then** at RED (before the Precondition 2 edit): the new 3-phase PASS test FAILS against current `closer.go` (because Precondition 2 still hard-requires `§E.5`) — RED evidence recorded in plan.md §E.1
**And** at GREEN (after the Precondition 2 edit): the converted test asserts the `§E.5`-absent case PASSES (3-phase SPEC closes), and a sibling legacy-`§E.5`-present PASS test is added for backward-compat regression cover

**Verification**:
- RED: `go test ./internal/spec/ -run TestClose_ThreePhasePreconditionPass -v` against unmodified `closer.go` exits non-zero with a `missing §E.5` failure message.
- GREEN: same command against the relaxed `closer.go` exits 0; `go test ./internal/spec/ -run TestClose_LegacyMxSectionStillAccepted -v` also exits 0.

### AC-LSG-007 (DEFERRED): mo.ai.kr close dry-run passes

**Given** the run-phase fix is merged and `~/go/bin/moai` is reinstalled via `make build`
**When** in the mo.ai.kr checkout: `moai spec close SPEC-DB-CLEANUP-001 --dry-run` is run
**Then** the dry-run reports precondition PASS (no `"missing §E.5 Mx-phase audit-ready signal"` failure)
**And** the 4 blocked SPECs (SPEC-DB-BACKLOG-001, SPEC-DB-CLEANUP-001, SPEC-DESIGN-HANDOFF-NEWS-001, SPEC-ID-SEC-MEDIUM-001) become closeable

**Verification**: DEFERRED — recorded as command + expected output only; NOT executed during run phase (requires `make build` reinstall + checkout switch). Executed post-merge by the user.

## §D. Edge Cases

- **EC-1**: `progress.md` has BOTH `§E.5` AND `§E.4 + sync_commit_sha` (a SPEC mid-migration). Precondition 2 MUST pass (either path accepts). Covered by AC-LSG-001 OR AC-LSG-002 logic — the predicate is an OR.
- **EC-2**: `progress.md` has `§E.4` but `sync_commit_sha` is empty. Precondition 2 MUST fail (3-phase path requires both); the legacy `§E.5` path applies only if `§E.5` is present. The 3-phase path is `§E.4 AND sync_commit_sha != ""`.
- **EC-3**: `progress.md` is absent entirely (V2.x SPEC). Unchanged behavior — other preconditions fail first (Precondition 1 `§E.2`). Not affected by this relaxation.

## §E. Definition of Done

- [ ] All MUST-PASS ACs (AC-LSG-001 through AC-LSG-006) verified PASS with mechanical evidence (test output, grep output).
- [ ] AC-LSG-007 (DEFERRED) recorded as command + expected output; execution deferred to post-merge user action.
- [ ] `go build ./...` and `go test ./...` green.
- [ ] `era.go` `§E.5` match count unchanged (scope guard).
- [ ] `transitions.go` dual-infix intact (scope guard).
- [ ] TDD RED evidence captured in `plan.md §E.1`.
- [ ] `closer.go` has zero residual `4-phase close` / `4-phase precondition` / `4-phase close transition` matches (AC-LSG-004).
