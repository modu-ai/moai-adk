---
id: SPEC-CLI-WORKTREE-FLAG-RACE-001
title: "Acceptance criteria — the four-sibling seam race"
version: "0.1.2"
created: 2026-09-03
author: manager-spec (card t464)
---

# Acceptance Criteria — SPEC-CLI-WORKTREE-FLAG-RACE-001

## §D AC matrix

| AC | Asserts | Requirement | Severity |
|---|---|---|---|
| AC-WFR-001a | RED is observed on the pre-repair tree, on the exact command the GREEN uses | REQ-WFR-002, REQ-WFR-006 | MUST |
| AC-WFR-001b | GREEN under `-race` repetition after the repair | REQ-WFR-001, REQ-WFR-002 | MUST |
| AC-WFR-002 | All four siblings survive with their assertions intact | REQ-WFR-004 | MUST |
| AC-WFR-003 | The diff stays inside the blast-radius fence | REQ-WFR-005 | MUST |
| AC-WFR-004 | The whole `internal/cli` package is green under `-race` (non-regression) | REQ-WFR-003 | MUST |
| AC-WFR-005 | Cross-platform compile of the package's tests | — (compile-only; see §D.3 note) | SHOULD |
| AC-WFR-006 | The option choice and its criteria are recorded | REQ-WFR-006 | MUST |

---

## §D.1 Criteria in Given-When-Then form

### AC-WFR-001a — RED is observed before it is fixed

**Given** the tree at `d592b0551` (or the current head, with a matching fresh capture),
**When** the implementer runs

```bash
go test ./internal/cli/ -run 'TestResolveWorktreeExistingBranch' -count=20 -race
```

**Then** the run exits `1` and its output contains at least one `WARNING: DATA RACE` **or** the
`Log in goroutine after … has completed` panic described below, and the verbatim output is
persisted under `.moai/reports/t464/red-race-<sha>.txt`.

Baseline attribution for this run is already satisfied for `d592b0551`:
`.moai/reports/t464/red-race-d592b0551.txt`, exit `1`, `WARNING: DATA RACE` **≥ 23**, 1369 lines.
Citing that path discharges AC-WFR-001a **only while HEAD is `d592b0551`**; on a moved head, a
fresh capture is required, because a baseline attributed to a different tree is not a baseline.

**The RED run was truncated — 23 is a floor.** It panicked at report line 1344
(`panic: Log in goroutine after TestResolveWorktreeExistingBranch_NoFlagIsNoop has completed: materialize must not run without --branch`)
and aborted with `FAIL … 1.543s` before exhausting `-count=20`. Do not cite 23 as the count of a
completed run; cite it as `≥ 23` from a truncated one.

**The panic is a second manifestation of the same root cause** (`spec.md` §A.2) and is therefore
part of what this AC's RED records — not an unrelated failure to be filtered out.

Why this AC exists: without an observed RED on this exact command, a subsequent GREEN cannot
distinguish "the race is gone" from "the command never detected it".

### AC-WFR-001b — GREEN under the same repetition shape

**Given** the repair has landed in the working tree,
**When** the implementer runs the **same** command as AC-WFR-001a:

```bash
go test ./internal/cli/ -run 'TestResolveWorktreeExistingBranch' -count=20 -race
```

**Then** the run exits `0`, and the count of `WARNING: DATA RACE` occurrences in its output is
exactly `0`.

Both halves are binary and both are required. The exit code alone is insufficient — `go test` can
exit 0 on a run whose output still carries race warnings under some configurations, so the warning
count is checked independently:

```bash
go test ./internal/cli/ -run 'TestResolveWorktreeExistingBranch' -count=20 -race \
  > .moai/reports/t464/green-race-<sha>.txt 2>&1; echo "exit=$?"
grep -c 'WARNING: DATA RACE' .moai/reports/t464/green-race-<sha>.txt   # expect 0
grep -c 'Log in goroutine after'  .moai/reports/t464/green-race-<sha>.txt   # expect 0
```

`grep -c` exits `1` when the count is `0`, so read the printed count, not the exit status of the
grep. Expected observation: `exit=0` and a printed count of `0` from **both** greps.

The second grep is not redundant. The RED's dominant terminal symptom was the panic, not a race
warning (`spec.md` §A.2), and a repair that silenced the warnings while leaving a completed test's
stub reachable would still panic. Checking only the race count would let that pass.

Because the RED aborted early, the GREEN run executes **more** iterations than the RED reached.
That asymmetry runs in the safe direction and is not a defect in the comparison.

[HARD] A green run of the full suite, or of this package at `-count=1`, does **NOT** satisfy this
AC and MUST NOT be substituted for it. Manifestation is schedule-dependent; a single pass is
consistent with the defect still being present.

### AC-WFR-002 — the four siblings survive intact

**Given** the repaired file,
**When** the implementer lists the test functions and their assertion bodies:

```bash
grep -n 'func TestResolveWorktreeExistingBranch_' internal/cli/worktree_branch_flag_test.go
git diff -- internal/cli/worktree_branch_flag_test.go
```

**Then** all four functions are present —
`_NoFlagIsNoop`, `_WiresAndStrips`, `_RejectsBadUsage`, `_MaterializeErrorPropagates` — none is
`t.Skip`-ped, none is merged into another, and the diff removes no assertion (`if …; t.Fatal` /
`t.Error` bodies). Under Option A the only removals are the four `t.Parallel()` lines; under
Option B the stub-installation lines are replaced by per-call argument passing, and the assertions
are untouched either way.

### AC-WFR-003 — the diff stays inside the fence

**Given** the repair is staged,
**When** the implementer runs

```bash
git diff --stat d592b0551 -- internal/cli/
```

**Then** the changed-file list under `internal/cli/` is a subset of exactly:

- `internal/cli/worktree_branch_flag_test.go`
- `internal/cli/worktree_branch_flag.go` (Option B only; absent under Option A)

and **no other** file in `internal/cli/` appears — in particular, none of the 22 other test files
that assign `findProjectRootFn`.

Positive control (mandatory, so an empty result cannot pass as a clean fence): the same command
without the pathspec must show a non-empty diff. An empty diff on both means nothing was measured,
not that the fence held.

### AC-WFR-004 — package non-regression

**Given** the repair has landed,
**When** the implementer runs

```bash
go test ./internal/cli/... -race -timeout 600s
```

**Then** the run exits `0`.

This is a **non-regression** check, not repair evidence: a single package pass is
schedule-dependent for this defect class. It establishes that the repair broke nothing else in the
package, and nothing more. The repair evidence is AC-WFR-001b.

Scope note: `go test ./...` is prohibited locally (repository rule); the full-suite verdict belongs
to CI on the pushed head.

### AC-WFR-005 — cross-platform compile

**Given** the repair has landed,
**When** the implementer runs

```bash
GOOS=linux  GOARCH=amd64 go vet ./internal/cli/...
GOOS=windows GOARCH=amd64 go vet ./internal/cli/...
```

**Then** both exit `0`.

Limitation recorded rather than hidden: `go vet` establishes that the package and its tests
compile on those targets. It does **not** run them, so it says nothing about whether the race
reproduces or stays fixed on linux/amd64 — that gap is `spec.md` §F2 and remains open.

### AC-WFR-006 — the option choice is recorded

**Given** the run-phase implementer has chosen between Option A and Option B,
**When** a reader opens `progress.md` §E.2,
**Then** it names the chosen option and states which of the five criteria in `spec.md` §D.3
decided it, and the entry was written before the first code edit.

---

## §D.2 Severity

- **MUST** (blocks close): AC-WFR-001a, AC-WFR-001b, AC-WFR-002, AC-WFR-003, AC-WFR-004, AC-WFR-006.
- **SHOULD** (a failure is a recorded gap, not a block): AC-WFR-005.

## §D.3 Traceability

| Requirement | Covered by |
|---|---|
| REQ-WFR-001 | AC-WFR-001b |
| REQ-WFR-002 | AC-WFR-001a, AC-WFR-001b |
| REQ-WFR-003 | AC-WFR-001b, AC-WFR-002, AC-WFR-004 |
| REQ-WFR-004 | AC-WFR-002 |
| REQ-WFR-005 | AC-WFR-003 |
| REQ-WFR-006 | AC-WFR-001a, AC-WFR-006 |

**Why REQ-WFR-003 does not map to AC-WFR-005.** REQ-WFR-003 asserts behaviour preservation across
three named behaviours (`--branch` token stripping, error propagation, the no-flag no-op path).
AC-WFR-005 runs `go vet` cross-compiles and **executes nothing**, so it can establish only that the
code still builds — it cannot observe any of the three behaviours. The ACs that actually exercise
them are AC-WFR-001b (the four siblings run, and each of the three behaviours is one sibling's
subject) and AC-WFR-002 (their assertions are still present to be exercised). AC-WFR-004 adds the
rest of the package as a non-regression net.

This matters most under **Option B**, where a production signature change makes behaviour
preservation the primary risk of the card. Mapping that risk onto a compile-only check would have
left the card's largest hazard unverified while the traceability table read as complete.

AC-WFR-005 is retained as a SHOULD with no requirement mapping: it guards cross-platform
compilation, which is worth checking and which no requirement in this SPEC asserts.

## §D.4 Definition of Done

All MUST criteria observed with verbatim output persisted under `.moai/reports/t464/`; `spec.md`
§F risks re-read and any that changed state recorded in `progress.md`; the SPEC status advanced by
the owning agent (not by this one).

## §D.5 What this AC set does NOT establish

- That the other 22 `internal/cli` test files assigning `findProjectRootFn` are race-free in
  general. The package-wide scans establish only that none combines `t.Parallel()` with a direct
  global write; no AC here runs them under `-race` beyond AC-WFR-004's package pass.
- That the race is fixed on linux/amd64. Only compilation is checked there (`spec.md` §F2).
- That `-count=20` is a minimal or sufficient reproduction bound. It is the shape that was measured,
  and the RED never completed all 20 iterations (`spec.md` §F3).
- That the other 22 `internal/cli` test files are free of an **indirect** shared-global write — one
  routed through a helper or an alias. The package-wide scans that cleared them match a syntactic
  pattern only (`spec.md` §F1).
- That CI exercises the `-race` guard on this package. Unverified (`spec.md` §F5).
