---
id: SPEC-INTEGRATION-LOCK-ATOMIC-001
title: "Implementation plan — integration lock mutation atomicity (card t336)"
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

# Implementation Plan — SPEC-INTEGRATION-LOCK-ATOMIC-001

Sections are ordered so the decisions most likely to change are read first: the lock-shape
decision (§B), then the contention contract it implies (§C), then the milestones (§F), which
end with the mechanical steps.

## §A Context

Card t336. Baseline tree `15453140a`, branch `WT-integration-lock-atomic` (`0 0` against
`origin/develop` at pre-flight time). The measured problem, the layer judgment, and the explicit
gaps are in `.moai/reports/t336/preflight.md`; `spec.md` §A/§B restate what this plan depends on.
The race is a code-path hypothesis until M1 observes it — no milestone here may cite it as a
measured defect before then.

## §B Design decision — the mutation lock's shape (RESOLVED: option b)

The pre-flight offered two shapes. This plan picks **(b): a dedicated short-lived mutation lock
beside the record, reusing the existing platform substrate**, and reuses at the substrate level so
the simplicity ladder's step 2 ("does a helper already exist here — reuse it") is still honoured.

**What is reused rather than written** (all measured at `15453140a`):

- `acquireBoardLockImpl(lockPath string)` — `board_lock_unix.go:36` / `board_lock_windows.go:55`.
  Already path-parameterized; the flock and atomic-create substrates are taken as they are.
- `newLockOwnerRecord()` — `board_lock.go:124`, the pid+created_at identity written into the
  artifact.
- `boardLockRetryWait` / `boardLockWaitBudget` — `board_store.go:114-151`, the existing bounded,
  jittered contention policy.
- The Windows dead-owner clear's shape — `board_lock_clear_windows.go:75` (`ClearStaleBoardLock`),
  including its pre-removal re-read.

So (b) adds **no new primitive**: it adds a second *scope* over the same substrate.

**Why not (a) — `AcquireBoardLock` as the mutation mutex.** (a) is genuinely cheaper in artifact
count. The deciding argument is item 1 below; items 2 and 3 are real but secondary, and item 3 is
stated at its true (limited) weight rather than inflated.

1. **Scope correctness is the thing being repaired — this is the deciding argument.** The card's
   whole subject is a lock whose scope did not match what it claimed to protect. Answering it by
   borrowing a lock whose scope is the WHOLE BOARD repeats the category error one level up: the
   integration window's mutations would be serialized by an artifact that says nothing about the
   integration window, and the next reader would have to re-derive why a board lock guards a
   release record. (b) makes the guarded thing and the guard name the same scope.
2. **Error classification and cross-scope latency.** Under (a) the contention sentinel reaching an
   integration verb would be `ErrBoardLockHeld` — a board error on `moai integration acquire` —
   and every board card mutation would serialize against every integration-window mutation. Stated
   accurately: this is remediable by wrapping the sentinel, and it is NOT a spurious failure. The
   board's mutation path is `acquireBoardLockSerialized` (`board_store.go:156`), which retries
   `AcquireBoardLock` under `boardLockWaitBudget` (`board_store.go:117` — 10 writers × 33 ms × 5
   headroom = 1.65 s), so an option-(a) implementation using that same entry point would WAIT, not
   fail. The residual cost is therefore misclassification plus added latency, not a broken verb.
   (An earlier draft of this plan claimed a `moai todo` write in flight would make
   `moai integration acquire` fail; that inference does not follow from the cited lines and has
   been corrected.)
3. **A side effect on projects that have no board.** `AcquireBoardLock` calls
   `os.MkdirAll(BoardDir(root))` (`board_lock.go:76-79`), so the first integration acquire in a
   project that never used the kanban board would create a board directory. Minor, but a real
   observable change for a project that never opted into the board.

**Windows substrate obligation is first-class, not an afterthought.** On Windows the artifact IS
the lock (`board_lock_windows.go:3-8`), so a process killed inside the critical section leaves a
file that blocks every subsequent mutation. Because the mutation lock is SHORT-LIVED — one CLI
invocation — the wedge is both more likely to be noticed and cheap to resolve: a contender that
exhausts the bounded budget consults the recorded owner and, only when that pid is positively
observed absent, re-reads the identity and clears, then retries (REQ-ILA-006). On Unix nothing is
added: the kernel drops flock on process exit (`board_lock_unix.go:4-7`), which is why
`ClearStaleBoardLock` is already a documented no-op there (`board_lock_clear_unix.go:19-27`).

**The two lifetimes stay separate** (`integration_lock.go:13-20`): the mutation lock represents a
critical section and dies with the process; the WINDOW stays a record whose validity is decided by
the recorded holder's liveness, spanning many invocations. Nothing in this plan reads the mutation
artifact to decide who holds the window.

## §C Contention contract implied by §B

- Contention on the mutation lock is transient by construction; the contender retries within
  `boardLockWaitBudget` using `boardLockRetryWait`'s jittered backoff.
- Budget exhaustion returns a NEW sentinel (working name `ErrIntegrationLockBusy`) that is NOT
  `ErrIntegrationLockHeld` and does not satisfy `IsIntegrationLockHeld` (REQ-ILA-004). The CLI
  surfaces it as a transient retry-me condition.
- After entering the critical section, the mutation re-reads the record (REQ-ILA-003). The read at
  `integration_lock.go:186` moves inside the lock rather than being duplicated outside it.
- **Budget vs the test stall-release timeout (load-bearing ordering).** The cross-process test
  releases its stalled child on a bounded timeout (M1). That timeout MUST be shorter than the
  mutation-lock wait budget, so the second child's outcome is `ErrIntegrationLockHeld` (it waited,
  then found a live record) and not the busy sentinel (it gave up waiting). A busy outcome in
  AC-ILA-002 is a misconfigured harness, not a passing criterion, and the AC says so.

## §D Constraints carried into implementation

Restated from `spec.md` §D only where a milestone would otherwise be free to violate them:
`internal/hook/**` and `internal/config/**` untouched; t320's release strings byte-identical;
scoped verification only (`go test ./internal/kanban/...`), never `go test ./...`; child processes
bounded by `t.Cleanup`-registered kills or an external deadline; `t.TempDir()` isolation.

## §E Open questions

None. The single open design choice named in the card is resolved in §B. No
`[NEEDS CLARIFICATION]` marker is carried by this plan.

## §F Milestones

Ordering is fixed by the lead's requirement: the failing observation exists BEFORE the fix.

### M1 — RED: observe the double-hold across processes, deterministically and attributed (Priority: High)

Build ONE cross-process test and the interleaving mechanism it needs, then run it against the
unrepaired path.

Files touched:
- `internal/kanban/integration_lock.go` — add the nil-by-default test-only interleaving hook
  permitted by REQ-ILA-005's amendment: a package-level
  `var integrationLockMutationTestHook func()`, invoked once at `:204` (after the decision, before
  the write). Nil in every production path; no production caller sets it.
- `internal/kanban/kanban_helper_test.go` — add ONE helper op (working name
  `integration-acquire`) reading `HELPER_ROOT`, `HELPER_SESSION`, `HELPER_OWNER_PID`, and an
  optional `HELPER_STALL_FLAG`. When the stall flag is set, the child installs the hook so it
  blocks (bounded poll) until the parent touches the proceed flag. It calls
  `AcquireIntegrationLock` and prints an outcome line carrying the result AND both
  discriminators: `RESULT=<acquired|held|busy|error> REPLACED=<none|session-id> SESSION=<own id>`.
  The `SESSION=` field is what lets the parent ASSERT the two children's ids differ rather than
  assume it from its own setup (v0.1.2, audit finding N2).
- `internal/kanban/integration_lock_cross_test.go` — NEW. One test function,
  `TestIntegrationLockAcquire_SerializedAcrossProcesses`, driving the deterministic interleaving
  described below. This is the ONLY cross-process test this card ships; there is no separate
  RED-named function (see "One test, two tree states").

**[HARD] The PID the helper records (D1 — the criteria hinge on this).** The helper MUST set
`want.PID` to the PARENT test process's pid, passed in `HELPER_OWNER_PID`, together with
`PIDSource: PIDSourceSessionOwner` — mirroring the production CLI, which records
`session.ResolveOwnerPID()` and not its own pid (`internal/cli/integration.go:184`). It MUST NOT
use `os.Getpid()`. Reason: `AcquireIntegrationLock` writes `want.PID` verbatim (`:205`) and
`Stale()` (`:113-121`) probes `FactoryProcessAlive`. A child recording its own pid exits
immediately after writing, its record then reads STALE, and the second child takes it over
**legitimately** — producing `successes=2` even with a correct mutation lock. That would make the
RED a misattributed observation (a stale takeover, not the read-write race), which is worse than
no observation because it looks like success. The parent's pid is alive for the whole round by
construction. (`PID: 0` is the acceptable alternative — `Stale()` treats it as live by design —
but it discards the marker parity with production, so the parent pid is preferred.)

**[HARD] Attribution discriminator (two fields, both asserted).** Outcome collection records
`REPLACED` and `SESSION` per child. A double-hold counts as the read-write race ONLY when BOTH
children report `REPLACED=none` AND the two `SESSION` values DIFFER; any round where either child
reports a takeover is a stale-reclaim, and any round where the two ids are equal is the legal
same-session re-acquire (`integration_lock.go:156-159`) — both are harness faults, never RED.
AC-ILA-001 and AC-ILA-002 both assert on BOTH fields.

The session-id half is the v0.1.2 repair of audit finding N2: the distinctness previously lived
only in `acceptance.md` §D.2's setup prose, so a harness bug passing one id to both children would
have produced `successes=2`, `REPLACED=none` on both, and a non-`Stale()` winner — passing every
stated check while exercising the re-acquire path instead of the race. The test now reads the ids
back off the children's own outcome lines and fails when they match.

**Deterministic interleaving (D2), and why it needs a timeout.** Child A is started with the stall
flag and blocks inside the hook between its decision and its write. The parent waits for A's
`STALLED` marker, then runs child B to completion. Child A is released when B finishes **or** when
a bounded stall-release timeout elapses — the timeout is not optional: under the repair B blocks on
the mutation lock until A releases it, so waiting only for B would deadlock. The timeout is a
liveness bound, not a race window, and per §C it MUST be shorter than the mutation-lock wait budget
so B's outcome is `held` rather than `busy`.

**One test, two tree states (D3 + D4).** The shipped test asserts the GREEN invariant. It is run in
two states, which is what produces both directions:
- unrepaired path (before M2, or after M2 with the critical section disabled per M5's one-line
  revert) → both children acquire, `REPLACED=none` on both: the attributed RED observation;
- repaired path → exactly one acquires, the other reports `held`.
No test is deleted, renamed, or "flipped" at M2, and no criterion names a function that is absent
from the delivered tree.

Cleanup: children are bounded by `exec.CommandContext` with a deadline AND a
`t.Cleanup`-registered kill. A trailing `kill` is not accepted.

Escalation: three consecutive attempts observing no attributed double-hold is a blocker — stop and
escalate with the outputs; do not widen and retry, and do not proceed to M2 (spec.md §G).

Exit evidence: the command run and its verbatim output showing the attributed double-hold,
recorded in `progress.md` §E.2 during run phase.

### M2 — Repair: serialize the record's read-modify-write (Priority: High)

Files touched:
- `internal/kanban/integration_lock_mutation.go` — NEW. `withIntegrationLockMutation(projectRoot,
  fn)`: acquire the dedicated artifact via `acquireBoardLockImpl` at
  `.moai/state/integration-mutation.lock`, retry per §C, run `fn`, release on every path
  (`defer`). The artifact deliberately does NOT share the record's `integration-lock` stem, so
  `.moai/state/integration-lock*` cannot glob the two lifetimes back together (REQ-ILA-007).
- `internal/kanban/integration_lock.go` — wrap `AcquireIntegrationLock`'s L186..L205 and
  `ReleaseIntegrationLock`'s L216..L226 in the critical section; move the record read inside it.
  Add the `ErrIntegrationLockBusy` sentinel and its `Is…` predicate. No change to the existing
  sentinels' text or to any decision branch.

M1's test is unchanged by M2 — the same function now observes the GREEN direction (exactly one
success, one `ErrIntegrationLockHeld`). The positive control (AC-ILA-003) is added here.

### M3 — Windows substrate: a killed holder must not wedge the mutation lock (Priority: High)

Files touched:
- `internal/kanban/board_lock_clear_windows.go` — extract the path-keyed core of
  `ClearStaleBoardLock` into a path-parameterized helper; `ClearStaleBoardLock` keeps its exact
  signature and behavior and becomes a caller of it.
- `internal/kanban/integration_lock_mutation_windows.go` / `_unix.go` — NEW, build-tagged. Windows:
  on budget exhaustion, clear only a positively-dead owner (pre-removal re-read, abort on
  mismatch) and retry once. Unix: no-op, mirroring `board_lock_clear_unix.go`'s stated reason.
- `internal/kanban/integration_lock_mutation_windows_test.go` — NEW, windows-tagged: a dead-owner
  artifact is cleared and the subsequent mutation succeeds; a live-owner artifact is not cleared.

Local evidence is `GOOS=windows go vet ./internal/kanban/...` — compilation only. Behavior,
including non-regression of `ClearStaleBoardLock` itself, is judged by CI's windows job running
BOTH the board's existing criteria (`board_lock_clear_windows_test.go`) and the new ones. No
darwin-lane command can observe this code at all, so none is offered as a mitigation
(`spec.md` §G risk 3).

### M4 — Secondary: unique staging path in `writeIntegrationLock` (Priority: Medium)

File touched: `internal/kanban/integration_lock.go` — replace the fixed `path + ".tmp"` (`:257`)
with `os.CreateTemp(filepath.Dir(path), ...)` + `Rename`, removing the temp file on every failure
path. Honest framing (REQ-ILA-010): under M2's lock two writers cannot meet here, so this is
defense-in-depth against a future caller writing outside the lock. No criterion claims to observe
a torn record.

### M5 — Mutation guard, header truth, and boundary check (Priority: Medium)

- Mutation guard: a documented, runnable recipe that disables the critical section by inserting
  `return fn()` as the first statement of `withIntegrationLockMutation` — one line, `fn` stays
  used so the package still compiles, and it subtracts EXACTLY mutual exclusion (REQ-ILA-003's
  in-section re-read lives inside `fn` and survives the revert). Re-run M1's test and observe the
  attributed double-hold return; restore and observe it refuse again. Both directions recorded in
  `progress.md` §E.2 with verbatim output. Same 3-attempt escalation bound as AC-ILA-001.
- `internal/kanban/integration_lock.go:13-26` — the header's flock claim is made true or rewritten
  to describe what exists (REQ-ILA-011).
- t320 boundary check: `git diff 15453140a -- internal/cli/integration.go` shows no change to the
  release verb's strings or error classification.

### M6 — Scoped verification and evidence (Priority: Medium)

`go test ./internal/kanban/...` (add `./internal/cli/...` only if a CLI test was touched),
`go vet ./internal/kanban/...`, `GOOS=windows go vet ./internal/kanban/...`, and the surviving-child
check. Full-suite judgment is CI's; no local `go test ./...`.

## §G Anti-Patterns (named, so the run phase does not drift into them)

- **Widening into t320.** Touching release's error text or classification because the same function
  is being edited. The lock wraps the mutation; the message stays byte-identical.
- **Widening into the deny layer.** Editing `internal/hook/**` or adding a config flag "while we
  are here". REQ-ILA-008 forbids gating; the layer boundary forbids the edit.
- **Reporting mutation-lock contention as `ErrIntegrationLockHeld`.** It tells a lane a false thing
  about the board (REQ-ILA-004).
- **A one-sided test.** GREEN alone cannot be distinguished from the lock being disabled; M1's RED
  and M5's mutation guard are what make the criterion mean anything.
- **Accepting `successes=2` without its attribution.** A double-hold where either child reports
  `REPLACED=<session>` is a stale takeover, not the race — reporting it as RED is the one failure
  mode with no later signal to correct it (D1).
- **A helper that records `os.Getpid()`.** It manufactures exactly that stale takeover. The
  parent's pid (or `0`) is the only acceptable recorded PID.
- **Widening the round count until RED appears.** Three attempts without an attributed
  double-hold is a blocker to escalate, not a knob to turn.
- **Setting the interleaving hook from any production path.** It is nil-by-default and test-only;
  a production setter would void REQ-ILA-005's carve-out.
- **Trailing `kill` as cleanup.** Every early return skips it; use `t.Cleanup` or an external
  deadline.
- **Claiming the race as measured before M1.** The pre-flight ran no concurrent acquire.
- **Persisting the mutation artifact as the window.** Two lifetimes, never conflated.

## §H Cross-References

- `.moai/reports/t336/preflight.md` — the measured pre-flight (file:line at `15453140a`, layer
  judgment, harness inventory, stated gaps).
- `.moai/specs/SPEC-INTEGRATION-LOCK-LIVENESS-001/` — t298; §G carries this card's residue.
- `internal/kanban/board_lock.go`, `board_lock_unix.go`, `board_lock_windows.go`,
  `board_lock_clear_windows.go`, `board_store.go:114-183` — the reused substrate.
- `internal/kanban/kanban_helper_test.go`, `board_lock_cross_test.go` — the harness M1 extends.
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Integration into the release branch is
  self-served — the doctrine this mechanism backs.
