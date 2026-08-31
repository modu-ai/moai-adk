---
id: SPEC-INTEGRATION-LOCK-ATOMIC-001
title: "Acceptance criteria — integration lock mutation atomicity (card t336)"
version: "0.1.2"
created: 2026-08-29
updated: 2026-08-30
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "integration-lock, atomicity, kanban, factory, t336"
tier: M
---

# Acceptance Criteria — SPEC-INTEGRATION-LOCK-ATOMIC-001

Every criterion below is decided by the observable output of a NAMED command. None may be
satisfied by reading code. Commands are scoped to the touched packages; `go test ./...` appears
nowhere and is prohibited locally (`spec.md` §D) — CI is the full-suite judge.

Every criterion that spawns child processes carries the cleanup obligation: children are bounded
by `exec.CommandContext` with a deadline AND a kill registered with `t.Cleanup`. A trailing `kill`
at the end of a test body is explicitly NOT acceptable — every early-return path skips it.

## §D AC Matrix

| AC | Requirement(s) | Claim | Deciding command | Observable that decides it |
|---|---|---|---|---|
| AC-ILA-001 | REQ-ILA-001, 002 | RED — the criterion asserts that on the unrepaired mutation path, two concurrent acquires from two SEPARATE OS processes **shall both be** told they hold the window, and that the double-hold **shall be attributable** to the read-modify-write window rather than to a stale takeover | `go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1 -v` at **HEAD with M2's critical section disabled** (the AC-ILA-006 one-line revert; before M2 exists, HEAD itself is that state) — NOT at `15453140a`, where neither the test nor the interleaving hook exists | The test log reports a round with `successes=2` AND `REPLACED=none` for BOTH children AND the two children's reported session ids DIFFER (`SESSION=<id>` on each outcome line, asserted by the test, not merely arranged by its setup) AND the first winner's record reading non-`Stale()` at the second acquire. That positive report is the pass condition; the test's non-zero exit is a by-product, not the signal |
| AC-ILA-002 | REQ-ILA-001, 002, 003 | GREEN — with the critical section restored, the same deterministic interleaving yields exactly one holder and the loser is refused with `ErrIntegrationLockHeld` | `go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1` | PASS reporting `RESULT=acquired REPLACED=none` for one child and `RESULT=held REPLACED=none` for the other, with the two children's `SESSION=<id>` values asserted DIFFERENT; the on-disk record names the winner. A `RESULT=busy` is a misconfigured harness (stall-release timeout ≥ the mutation-lock budget) and FAILS the criterion; a `REPLACED=<session>` on either child means a takeover was observed instead of serialization and also FAILS |
| AC-ILA-003 | REQ-ILA-005 | Positive control — a NON-CONFLICTING concurrent pair (both children acquire as the SAME session id) BOTH succeed | `go test ./internal/kanban/... -run TestIntegrationLockAcquire_ConcurrencyPositiveControl -count=1` | PASS with both children reporting `RESULT=acquired`; refusal is therefore conditional on contention, not on concurrency itself |
| AC-ILA-004 | REQ-ILA-004 | Mutation-lock contention is distinguishable from window contention | `go test ./internal/kanban/... -run TestIntegrationLockBusy_IsNotHeld -count=1` | PASS: the busy sentinel satisfies its own predicate and `IsIntegrationLockHeld(busy)` is false |
| AC-ILA-005 | REQ-ILA-005, 009 | Single-threaded semantics unchanged (re-acquire by holder, stale takeover, `--force`, foreign-release refusal, empty release) | `go test ./internal/kanban/... -run IntegrationLock -count=1` and `git diff --stat 15453140a -- internal/kanban/integration_lock_test.go` | Tests PASS and the diff is EMPTY — the pre-existing criteria are met without editing them |
| AC-ILA-006 | REQ-ILA-001 (guard) | Mutation guard — the GREEN criterion is not vacuous | Insert `return fn()` as the first statement of `withIntegrationLockMutation` (the documented one-line revert), run `go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1`; restore the line and re-run | Disabled: the attributed double-hold of AC-ILA-001 is reported (`successes=2`, `REPLACED=none` on both). Restored: `successes=1 refusals=1`. Both outputs recorded verbatim in `progress.md` §E.2. Same escalation bound as AC-ILA-001: three attempts with no attributed double-hold in the disabled state is a blocker, not a retry loop |
| AC-ILA-007 | REQ-ILA-006 | Windows substrate: a killed mutation-lock holder does not wedge the record, and `ClearStaleBoardLock` does not regress | (a) `GOOS=windows go vet ./internal/kanban/...` locally — compilation only; (b) CI's windows job running BOTH the new mutation-lock clear test AND the board's pre-existing `board_lock_clear_windows_test.go` criteria | (a) exit 0 with no output. (b) CI windows job green on the PR head, with both test files' criteria in the run. (b) is judged by CI; no darwin-lane command compiles this windows-tagged code, so no local command is offered as evidence for it |
| AC-ILA-008 | REQ-ILA-010 | Unique staging path — no fixed shared `.tmp` name remains, and a write leaves no residue | `grep -n 'path + "\.tmp"' internal/kanban/integration_lock.go` and `go test ./internal/kanban/... -run TestWriteIntegrationLock_UniqueStagingPath -count=1` | grep prints NOTHING (base measurement: it prints line 257 today) and the test PASSES with no leftover staging file in the state dir |
| AC-ILA-009 | §D cleanup obligation | No child process outlives the scoped test run | `PAT=kanban; pgrep -f "[${PAT:0:1}]${PAT:1}\.test" \| grep -v "^$$\$" \| wc -l` immediately before and immediately after `go test ./internal/kanban/... -count=1` — the bracketed first character keeps the measuring pipeline's own command line from matching, and `$$` excludes the measuring shell | Both readings are `0`, counting no PID other than the measuring pipeline itself. A non-zero AFTER against a zero BEFORE is a cleanup defect and fails this criterion; a `1`/`1` reading means the self-match exclusion is not working and the measurement is broken, NOT that the run was clean |

## §D.1 Severity

MUST-PASS (any failure blocks the card): AC-ILA-001, AC-ILA-002, AC-ILA-003, AC-ILA-005,
AC-ILA-006, AC-ILA-009.

SHOULD-PASS (failure is reported and judged, not silently accepted): AC-ILA-004, AC-ILA-007(a),
AC-ILA-008.

CI-JUDGED (not claimable from a local run): AC-ILA-007(b).

## §D.2 Given-When-Then scenarios

**AC-ILA-001 — RED, the double-hold**

- **Given** the record layer with NO exclusion across `integration_lock.go:186..205` — HEAD with
  M2's critical section disabled (before M2 exists, HEAD itself is that state) — a fresh
  `t.TempDir()` with an empty `.moai/state/`, and both children recording the PARENT test
  process's pid (`HELPER_OWNER_PID`, `PIDSource: PIDSourceSessionOwner`), never `os.Getpid()`
- **When** child A is held inside the deterministic interleaving hook between its decision and its
  write, child B runs its entire acquire with a DIFFERENT session id, and A is then released (on
  B's completion or on the bounded stall-release timeout, whichever comes first)
- **Then** the run reports `successes=2` with `REPLACED=none` for BOTH children, the two
  children's reported `SESSION=<id>` values DIFFER, and the winner's record reads non-`Stale()` at
  the moment of the second acquire. All four parts are required: `successes=2` alone is also what a
  stale takeover produces; `REPLACED=none` on both is what attributes the double-hold to the
  unserialized read-modify-write; and the differing session ids are what exclude the SAME-SESSION
  path, which produces `successes=2`, `REPLACED=none` on both, and a non-`Stale()` winner while
  exercising the legal re-acquire of `integration_lock.go:156-159` rather than the race. The
  distinctness is ASSERTED by the test from the children's own reported ids, never assumed from the
  setup: a harness bug passing one id to both children would otherwise yield a false RED that
  survives every other check. A round where either child reports `REPLACED=<session>`, or where the
  two ids are equal, is a harness fault to fix, never the RED observation

**AC-ILA-002 — GREEN, exactly one holder**

- **Given** the repaired record layer with the mutation lock in place, and the same PID discipline
- **When** the identical deterministic interleaving runs
- **Then** one child reports `RESULT=acquired REPLACED=none`, the other `RESULT=held REPLACED=none`
  (its error satisfying `IsIntegrationLockHeld`), the two children's reported `SESSION=<id>` values
  are asserted DIFFERENT (same reason as AC-ILA-001: the same-session path is a legal re-acquire,
  not a refusal, so an undetected id collision would silently change what this criterion measures),
  and the persisted record names the winner. A
  `RESULT=busy` fails the criterion: it means the stall-release timeout was not shorter than the
  mutation-lock wait budget, so the harness measured its own configuration rather than the lock

**AC-ILA-003 — positive control**

- **Given** the repaired record layer and a fresh `t.TempDir()`
- **When** two child processes concurrently acquire using the SAME session id (a lane re-entering
  its own window after a `/clear` — the case `integration_lock.go:156-159` documents as legal)
- **Then** both report `RESULT=acquired` and the record names that session; the lock refuses
  contention, not concurrency

**AC-ILA-004 — contention is not possession**

- **Given** the new transient sentinel returned on mutation-lock budget exhaustion
- **When** a caller inspects it with the package predicates
- **Then** its own predicate reports true and `IsIntegrationLockHeld` reports false, so a lane is
  never told another session owns a window that is free

**AC-ILA-005 — preserved single-threaded semantics**

- **Given** the pre-existing criteria in `internal/kanban/integration_lock_test.go`, unmodified
- **When** the scoped `IntegrationLock` test selection runs against the repaired tree
- **Then** all pass AND `git diff --stat` against `15453140a` for that file is empty — the repair
  changed when mutations interleave, not what a single-threaded call decides

**AC-ILA-006 — mutation guard**

- **Given** a passing AC-ILA-002
- **When** `return fn()` is inserted as the first statement of `withIntegrationLockMutation` —
  one line, `fn` stays used so the package compiles, and it subtracts exactly mutual exclusion
  (REQ-ILA-003's in-section re-read lives inside `fn` and survives) — and AC-ILA-002's command is
  re-run, then the line is removed and the command run once more
- **Then** the disabled run reports the attributed double-hold (AC-ILA-001's three-part observable)
  and the restored run reports `successes=1 refusals=1` — so the criterion is measuring the lock
  and not something else. Zero attributed double-holds across three disabled-state attempts is a
  blocker to escalate, not a signal to retry

**AC-ILA-007 — Windows substrate**

- **Given** the windows-tagged mutation-lock clear and the board's pre-existing
  `board_lock_clear_windows_test.go` criteria
- **When** `GOOS=windows go vet ./internal/kanban/...` runs locally, and CI's windows job runs BOTH
  test files — the new one, in which a mutation artifact whose recorded owner is dead is cleared
  and the retry succeeds while a live-owner artifact is left alone, and the board's existing one,
  which must not regress
- **Then** vet exits 0 locally (compilation only), and the CI windows job is green on the PR head
  with both files' criteria in the run. The behavioral half is CI's verdict: this code is
  `//go:build windows`, so no darwin-lane command compiles it and none may be cited for it

**AC-ILA-008 — unique staging path**

- **Given** `integration_lock.go:257`'s fixed `path + ".tmp"` at baseline (the grep prints that
  line today — this is the zero-baseline the criterion is measured against)
- **When** the staging path is made unique per call
- **Then** the grep prints nothing and a write leaves no residual staging file. Honest scope: under
  the mutation lock two writers cannot reach this path, so this observes the PROPERTY (unique
  staging, no residue), never a torn record

**AC-ILA-009 — cleanup**

- **Given** a self-excluding count of `0` before the run — the pattern's first character is
  bracketed so the measuring pipeline's own command line cannot match it, and `$$` is filtered so
  the measuring shell is not counted
- **When** the scoped package test completes (including any failing or early-returning test)
- **Then** the same self-excluding count is `0` afterwards — no PID other than the measuring
  pipeline. A `1`/`1` reading is a broken measurement (the exclusion failed), not a clean run

## §D.3 Traceability

| Requirement | Covered by |
|---|---|
| REQ-ILA-001 | AC-ILA-001, AC-ILA-002, AC-ILA-006 |
| REQ-ILA-002 | AC-ILA-001, AC-ILA-002 |
| REQ-ILA-003 | AC-ILA-002 |
| REQ-ILA-004 | AC-ILA-004 |
| REQ-ILA-005 | AC-ILA-003, AC-ILA-005 |
| REQ-ILA-006 | AC-ILA-007 |
| REQ-ILA-007 | AC-ILA-005 (the window record's semantics are unchanged); no artifact is persisted as the window |
| REQ-ILA-008 | AC-ILA-002 (the criteria set the flag nowhere; the repaired behavior holds with `Workflow.IntegrationLock.Enabled` at its default `false`) |
| REQ-ILA-009 | AC-ILA-005 (`--force` and foreign-release refusal still behave as before) |
| REQ-ILA-010 | AC-ILA-008 |
| REQ-ILA-011 | Closure gate below (header truth), verified by diff review, not by a test |

## §D.4 Indirect verification (stated, not hidden)

- **REQ-ILA-011** is a documentation truth condition. No command can decide whether prose matches a
  mechanism, so it is a closure gate reviewed on the diff rather than an AC row.
- **REQ-ILA-007** (lifetime separation) is verified negatively: no read path consults the mutation
  artifact to answer who holds the window. The observable is the absence of such a call site in the
  diff, which is review, not measurement — stated here rather than dressed as an AC.

## §D.5 Closure gates

1. Every MUST-PASS criterion observed, with the command and its verbatim output recorded in
   `progress.md` §E.2.
2. AC-ILA-006's two directions both recorded (FAIL with the section disabled, PASS restored).
3. t320 boundary: `git diff 15453140a -- internal/cli/integration.go` shows no change to the
   release verb's strings or error classification.
4. Layer boundary: `git diff --stat 15453140a -- internal/hook internal/config` is EMPTY.
5. `integration_lock.go`'s header no longer claims a mechanism that does not exist.
6. The interleaving hook has no production setter: `grep -rn 'integrationLockMutationTestHook\s*=' internal/ | grep -v '_test\.go'` prints nothing (the declaration itself is `var … func()` with no assignment). REQ-ILA-005's carve-out permits a nil-by-default hook and nothing more.

## §D.6 Definition of Done

The record layer serializes its own mutations across processes; the double-hold was OBSERVED
before the repair and is refused after it; a non-conflicting concurrent pair still both succeed;
single-threaded semantics and t320's surface are byte-unchanged; the Windows substrate cannot be
wedged by a killed short-lived holder; and every claim above is attributed to a command run on
this tree.

## §D.7 Forward-looking checks (not gates for this card)

- The mutation lock's wait budget is inherited from the board's derivation
  (`board_store.go:114-127`, sized for board mutations). If integration-window mutations turn out
  to be slower in practice, the budget is re-derived in a later card — not re-tuned by guess here.
- Multi-host coordination remains out of scope (`spec.md` § Exclusions); a future host-qualified
  record would need its own criterion.
