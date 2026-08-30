---
id: SPEC-INTEGRATION-LOCK-ATOMIC-001
title: "Progress record — integration lock mutation atomicity (card t336)"
version: "0.1.2"
status: in-progress
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

# Progress — SPEC-INTEGRATION-LOCK-ATOMIC-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t336. Tree `15453140a`, branch `WT-integration-lock-atomic`, worktree
  `.claude/worktrees/t336`.
- Artifacts authored at plan phase: `spec.md`, `plan.md`, `acceptance.md`, this file. Status
  `draft` across all four.
- Pre-flight consumed as-is: `.moai/reports/t336/preflight.md`. Its stated gap — no concurrent
  acquire was executed on this tree — is carried forward unchanged: the race is a code-path
  hypothesis until run-phase M1 observes it.
- Open design choice resolved at plan phase: option (b), a dedicated short-lived mutation lock
  reusing the existing platform substrate. Reasoning in `plan.md` §B.
- No `[NEEDS CLARIFICATION]` markers outstanding.
- Acceptance criteria: 9 (`AC-ILA-001`..`AC-ILA-009`); 6 MUST-PASS, 2 SHOULD-PASS, 1 CI-judged.
- No production or test Go code was written or modified at plan phase.

### plan-audit iteration 1 → repair (v0.1.1)

- Verdict read: `.moai/reports/t336/plan-audit.md` — FAIL, 0.803, six blocking findings (D1-D6),
  three optional (D7-D9). Must-pass firewall MP-1..MP-7 all passed; 14/14 citations resolved.
- All six blocking findings repaired; all three optional findings also applied (each was a
  one-line change that enlarged no scope).
- Two decisions taken during the repair, both recorded in the artifacts rather than left implicit:
  the deterministic interleaving hook was ADOPTED (D2's preferred fix) rather than declining it
  for a round ceiling; and the RED/GREEN pair was collapsed onto ONE shipped test function run in
  two tree states (D3+D4's first option), so no criterion names a test absent from the deliverable.
- Not changed, per the auditor's explicit "already sound" list: §A hypothesis framing, the 14
  verified citations, the traceability matrix and its §D.4 indirect-verification statement, the
  six-heading Exclusions block, the trailing-`kill` rejection, the choice of option (b), and M3's
  necessity.
- Scoped lint after the repair: `moai spec lint .moai/specs/SPEC-INTEGRATION-LOCK-ATOMIC-001/spec.md`
  → recorded in `.moai/reports/t336/spec-lint.txt`.

## §E.2 Run-phase Evidence

Run phase opened at HEAD `a5f414a8a` on branch `WT-integration-lock-atomic`, worktree
`.claude/worktrees/t336`. Package baseline measured before any edit:
`go test ./internal/kanban/... -count=1` → `ok github.com/modu-ai/moai-adk/internal/kanban 13.758s`,
exit 0 (evidence `.moai/state/verify/t336/baseline-kanban.txt`). Every new red below is attributed
against that baseline.

### M1 — RED: the double-hold, observed and attributed

**Opening repair (audit finding N2, `ILA-SAME-SESSION-UNASSERTED`).** Before writing the test, the
distinct-session-id requirement was raised out of `acceptance.md` §D.2's setup prose into what
AC-ILA-001 and AC-ILA-002 assert, and the artifacts moved to v0.1.2. The gap it closes: the
same-session path produces `successes=2`, `REPLACED=none` on both children, and a non-`Stale()`
winner — passing all three prior discriminator parts — while exercising the LEGAL re-acquire of
`integration_lock.go:156-159` rather than the race. That is the iteration-1 D1 misattribution class
(`os.Getpid()`) arriving through a second door. The helper now prints `SESSION=<id>` and the test
reads the two ids back off the children's own outcome lines and fails when they match.

**Nil guard (audit finding N1, `ILA-HOOK-NIL-GUARD`).** The hook's call site is explicitly
nil-guarded (`integration_lock.go`, between the decision and the write). A nil `func()` invoked in
Go panics, so the guard is what makes REQ-ILA-005's own sentence — "with the hook nil … behavior is
byte-for-byte what this requirement states" — true rather than merely intended. Closure gate 6
checks assignment, not invocation, so it would not have caught the unguarded form.

**Claim.** On the unrepaired mutation path, two concurrent acquires from two SEPARATE OS processes
are both told they hold the window, and the double-hold is attributable to the read-modify-write
window rather than to a stale takeover, a takeover by force, or a same-session re-acquire.

**Evidence** — command, and its verbatim output (attempt 1 of a permitted 3; no widening, no
retry loop):

```
$ go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1 -v
=== RUN   TestIntegrationLockAcquire_SerializedAcrossProcesses
    integration_lock_cross_test.go:263: A: RESULT=acquired REPLACED=none SESSION=lane-a
    integration_lock_cross_test.go:264: B: RESULT=acquired REPLACED=none SESSION=lane-b
    integration_lock_cross_test.go:265: round: successes=2 refusals=0 other=0 sessions_differ=true mid_record_held=true mid_record_stale=false final_holder="lane-a" final_record_stale=false
    integration_lock_cross_test.go:276: successes=2 refusals=0, want exactly 1 and 1 — the record's read-modify-write is NOT serialized across processes.
          A: RESULT=acquired REPLACED=none SESSION=lane-a
          B: RESULT=acquired REPLACED=none SESSION=lane-b
          attributed_double_hold=true (REPLACED=none on both: true; session ids differ: true; record non-Stale at the second acquire: mid_held=true mid_stale=false; final record non-Stale: true)
          final record: holder="lane-a" pid=60095 pid_source="session-owner"
--- FAIL: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.03s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.421s
FAIL
```

Saved verbatim at `.moai/state/verify/t336/m1-red-attempt1.txt`.

**Baseline-attribution.** Measured in this run, against this tree: HEAD `a5f414a8a` plus the M1
working-tree edits (the interleaving hook, the `integration-acquire` helper op, and this test).
This is the state AC-ILA-001 names — HEAD with the critical section absent — and explicitly NOT
`15453140a`, where neither the hook nor the test exists. The four-part observable resolves:
`successes=2`; `REPLACED=none` on both children; `SESSION=lane-a` ≠ `SESSION=lane-b`; and the
record read non-`Stale()` both at the second acquire (`mid_held=true mid_stale=false`) and at the
end (`final_record_stale=false`), its holder recorded as pid 60095 with
`pid_source="session-owner"` — the PARENT test process, alive for the whole round, never
`os.Getpid()` of a child.

**What this converts.** `spec.md` §A stated the race as a code-path hypothesis, since the pre-flight
ran no concurrent acquire. It is now a measurement: two live holders were observed, and the
observation is attributed. The interleaving was CONSTRUCTED rather than waited for, so the round
required no widening and the 3-attempt escalation ceiling was not approached.

**Gaps (explicitly NOT observed at M1).** No repair exists yet, so nothing here says the refusal
works — that is M2/AC-ILA-002. The busy sentinel does not exist, so `RESULT=busy` is unreachable
and AC-ILA-004 is unmeasured. Nothing windows-tagged was touched, so AC-ILA-007 is untouched. The
`.tmp` staging path is unchanged, so AC-ILA-008 still measures its base state.

**Residual risk.** The observation is a single round on one machine (darwin, this worktree). It
establishes that the double-hold is REACHABLE and attributable; it does not establish a frequency,
and no claim about frequency is made. The construction is deterministic by design, so a later
regression would surface as this test passing where it should fail — which is exactly what
AC-ILA-006's mutation guard exists to catch.


### M2 — GREEN: the record's read-modify-write is serialized

**Claim.** With the mutation lock in place, the identical deterministic interleaving yields exactly
one holder; the loser is refused with `ErrIntegrationLockHeld`, and the persisted record names the
winner. A non-conflicting concurrent pair still both succeed. Single-threaded semantics are
unchanged.

**Evidence.**

```
$ go test ./internal/kanban/... -run TestIntegrationLock -count=1 -v
=== RUN   TestIntegrationLockAcquire_SerializedAcrossProcesses
    integration_lock_cross_test.go:266: A: RESULT=acquired REPLACED=none SESSION=lane-a
    integration_lock_cross_test.go:267: B: RESULT=held REPLACED=none SESSION=lane-b
    integration_lock_cross_test.go:268: round: successes=1 refusals=1 other=0 sessions_differ=true mid_record_held=false mid_record_stale=false final_holder="lane-a" final_record_stale=false
--- PASS: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.55s)
=== RUN   TestIntegrationLockAcquire_ConcurrencyPositiveControl
    integration_lock_cross_test.go:341: positive control: both children acquired as lane-same; record holder="lane-same"
--- PASS: TestIntegrationLockAcquire_ConcurrencyPositiveControl (0.02s)
=== RUN   TestIntegrationLockBusy_IsNotHeld
--- PASS: TestIntegrationLockBusy_IsNotHeld (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.036s
```

```
$ go test ./internal/kanban/... -run IntegrationLock -count=1 -v
--- PASS: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.53s)
--- PASS: TestIntegrationLockAcquire_ConcurrencyPositiveControl (0.02s)
--- PASS: TestIntegrationLockBusy_IsNotHeld (0.00s)
--- PASS: TestAcquireIntegrationLock_RecordsHolder (0.00s)
--- PASS: TestAcquireIntegrationLock_RecordsTheCallersOwnerPID (0.00s)
--- PASS: TestAcquireIntegrationLock_AnchoredPIDZeroIsLiveNotStale (0.00s)
--- PASS: TestReadIntegrationLock_LegacyRecordWithoutPIDSource (0.00s)
--- PASS: TestAcquireIntegrationLock_RefusesASecondLiveSession (0.00s)
--- PASS: TestAcquireIntegrationLock_SameSessionReacquires (0.00s)
--- PASS: TestAcquireIntegrationLock_TakesOverAStaleHolder (0.00s)
--- PASS: TestAcquireIntegrationLock_ForceTakesOverALiveHolder (0.00s)
--- PASS: TestReadIntegrationLock_CorruptRecordIsNotAFreeWindow (0.00s)
--- PASS: TestReleaseIntegrationLock_HolderAndForeign (0.00s)
--- PASS: TestReleaseIntegrationLock_EmptyIsReported (0.00s)
--- PASS: TestReadIntegrationLock_AbsentRecord (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.795s

$ git diff --stat 15453140a -- internal/kanban/integration_lock_test.go
(no output; exit 0)
```

**Baseline-attribution.** Measured in this run, against this tree at the M2 commit `06290a314`.
The comparison that matters is against M1's own output on the SAME test function: `successes=2
refusals=0` became `successes=1 refusals=1`, and `mid_record_held` flipped `true` to `false`. That
flip is the mechanism showing itself rather than an incidental difference — under the unrepaired
path B had already written by the time the parent read the record between the two acquires; under
the repair B is still blocked on the mutation lock and writes only after A releases. B's decision
is therefore made against what A published, which is REQ-ILA-003 — the requirement the criteria
previously traced but had no mechanism to force.

**Design decision, restated where the code lands.** Option (b), a dedicated short-lived mutation
lock at `.moai/state/integration-mutation.lock`, reusing `acquireBoardLockImpl` and
`board_store.go`'s bounded jittered budget. No new primitive; a second SCOPE over the same
substrate. The artifact's filename stem is deliberately not the record's, so
`.moai/state/integration-lock*` cannot glob the two lifetimes back together.

**Gaps.** `RESULT=busy` was never produced in any round, so the busy PATH through
`acquireIntegrationMutationLock` is unexercised end-to-end; AC-ILA-004 asserts the sentinel's
predicates directly (all three directions: busy satisfies its own, not held, not board-held) rather
than by provoking a real 1.65 s timeout. That is deliberate — provoking it would mean holding the
lock for the whole budget, which is a slow test that measures the clock.

**Residual risk.** One observable behaviour change not required by any AC: `ReleaseIntegrationLock`
now creates `.moai/state/` if absent, because the mutation lock cannot be taken before its
directory exists. Acquire already created it, so the new case is an empty release in a project that
never held a window; the directory created is the record's own, not a foreign scope. Disclosed
rather than left for a reader to find.

### M3 — Windows substrate: a killed holder must not wedge the mutation lock

**Claim.** On the atomic-create substrate, a mutation-lock holder killed mid-section does not wedge
the record permanently: a contender that exhausts its budget clears the artifact only when the
recorded owner is positively observed absent, re-reading the identity immediately before removal,
and retries once. `ClearStaleBoardLock` keeps its exact signature and behaviour.

**Evidence.**

```
$ GOOS=windows go vet ./internal/kanban/...
exit=0
(no output)

$ go test ./internal/kanban/... -run IntegrationLock -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.985s
```

**Baseline-attribution.** Measured in this run, against this tree at the M3 commit `7f27a8ab9`.
The `go vet` result is COMPILATION ONLY and is cited as nothing more.

**Gaps — named rather than papered over.** `board_lock_clear_windows.go`,
`integration_lock_mutation_windows.go`, and `integration_lock_mutation_windows_test.go` are all
`//go:build windows`. No darwin-lane command compiles them, so this lane observed ZERO of their
runtime behaviour: not the dead-owner clear, not the live-owner refusal, and not the
non-regression of `ClearStaleBoardLock` itself. AC-ILA-007(b) is CI-judged, and no local command is
offered as evidence for it. The `go test -run IntegrationLock` line above is cited only as evidence
that the Unix side did not regress — it cannot speak to the Windows code M3 edits.

**Residual risk.** The extraction re-labels six error messages via a `label` parameter so the board
caller's strings stay byte-identical; that was verified by reading the diff, not by a test asserting
the strings. The shared `ErrBoardLockChangedHands` sentinel is now reachable from the integration
mutation path, where its text names the board — the integration caller discards the error and falls
through to busy, so it never reaches a user, but the sentinel is shared rather than scoped.

### M4 — Unique staging path

**Claim.** The record's staging path is unique per call, no residue survives a write, and the
record keeps the 0644 mode other sessions read it with.

**Evidence.**

```
$ grep -n 'path + "\.tmp"' internal/kanban/integration_lock.go
current-tree grep exit=1 (prints nothing)

$ git show 15453140a:internal/kanban/integration_lock.go | grep -n 'path + "\.tmp"'
257:	tmp := path + ".tmp"
baseline grep exit=0 (the zero-baseline this criterion is measured against)

$ go test ./internal/kanban/... -run IntegrationLock -count=1 -v
--- PASS: TestWriteIntegrationLock_UniqueStagingPath (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.018s   (16/16 selected criteria)
```

**Baseline-attribution.** Both grep directions measured in this run: the retired form present at
`15453140a` line 257, absent on this tree at the M4 commit `445ec5f8f`. A one-directional grep
would have proven nothing — a pattern that never matched anywhere reads identically to one that was
successfully removed.

**Honest scope (REQ-ILA-010).** Under M2's lock two concurrent writers cannot reach
`writeIntegrationLock` at all, so this observes the PROPERTY and never a torn record. No claim of
an observed tear is made anywhere.

**One non-cosmetic detail.** `os.CreateTemp` opens at 0600 where the fixed path wrote 0644. The
record is read by every other session's PreToolUse guard, so the mode is restored explicitly and
the test asserts it. Left unhandled this would have shipped silently as a guard failing closed for
the wrong reason.

### M5 — Mutation guard (both directions), header truth, boundary checks

**Claim.** The GREEN criterion is not vacuous: disabling exactly mutual exclusion returns the
attributed double-hold, and restoring it returns the refusal. The package header no longer
describes an absent mechanism. Neither the t320 surface nor the deny layer was touched.

**Evidence — direction 1, critical section DISABLED** (`return fn()` inserted as the first
statement of `withIntegrationLockMutation`; one line, `fn` stays used so the package compiles, and
it subtracts exactly mutual exclusion — REQ-ILA-003's in-section re-read lives inside `fn` and
survives the revert). Attempt 1 of a permitted 3:

```
$ go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1 -v
=== RUN   TestIntegrationLockAcquire_SerializedAcrossProcesses
    integration_lock_cross_test.go:266: A: RESULT=acquired REPLACED=none SESSION=lane-a
    integration_lock_cross_test.go:267: B: RESULT=acquired REPLACED=none SESSION=lane-b
    integration_lock_cross_test.go:268: round: successes=2 refusals=0 other=0 sessions_differ=true mid_record_held=true mid_record_stale=false final_holder="lane-a" final_record_stale=false
    integration_lock_cross_test.go:279: successes=2 refusals=0, want exactly 1 and 1 — the record's read-modify-write is NOT serialized across processes.
          A: RESULT=acquired REPLACED=none SESSION=lane-a
          B: RESULT=acquired REPLACED=none SESSION=lane-b
          attributed_double_hold=true (REPLACED=none on both: true; session ids differ: true; record non-Stale at the second acquire: mid_held=true mid_stale=false; final record non-Stale: true)
          final record: holder="lane-a" pid=7713 pid_source="session-owner"
--- FAIL: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.02s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.492s
FAIL
```

**Evidence — direction 2, critical section RESTORED** (restoration verified byte-exact: a
`git diff --stat` of `internal/kanban/integration_lock_mutation.go` against the M3 commit is
empty):

```
$ go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1 -v
=== RUN   TestIntegrationLockAcquire_SerializedAcrossProcesses
    integration_lock_cross_test.go:266: A: RESULT=acquired REPLACED=none SESSION=lane-a
    integration_lock_cross_test.go:267: B: RESULT=held REPLACED=none SESSION=lane-b
    integration_lock_cross_test.go:268: round: successes=1 refusals=1 other=0 sessions_differ=true mid_record_held=false mid_record_stale=false final_holder="lane-a" final_record_stale=false
--- PASS: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.55s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.945s
```

Saved verbatim at `.moai/state/verify/t336/m5-guard-disabled.txt` and `m5-guard-restored.txt`.

**Closure gates 3, 4 and 6.**

```
$ git diff 15453140a -- internal/cli/integration.go
(no output — the release verb's strings and error classification are byte-unchanged; card t320
 lands on top of this work untouched)

$ git diff --stat 15453140a -- internal/hook internal/config
(no output — the deny layer and its config are untouched, and the repair is gated by no flag)

$ grep -rn 'integrationLockMutationTestHook\s*=' internal/ | grep -v '_test\.go'
exit=1 (prints nothing — no production setter)

$ grep -rnE 'integrationLockMutationTestHook[[:space:]]*=' internal/ | grep -v '_test\.go'
exit=1 (POSIX bracket form, run as corroboration because \s is a GNU extension; same result)

$ grep -rn 'integrationLockMutationTestHook' internal/
internal/kanban/integration_lock.go:83:// integrationLockMutationTestHook is a nil-by-default, TEST-ONLY interleaving
internal/kanban/integration_lock.go:95:var integrationLockMutationTestHook func()
internal/kanban/integration_lock.go:239:		if integrationLockMutationTestHook != nil {
internal/kanban/integration_lock.go:240:			integrationLockMutationTestHook()
internal/kanban/kanban_helper_test.go:135:			integrationLockMutationTestHook = func() {
```

The three non-test occurrences are the doc comment, the `var … func()` declaration with no
assignment, and the nil-guarded invocation. The only assignment in the repository is in a
`_test.go` file.

**Closure gate 5 — header truth (REQ-ILA-011).** `integration_lock.go`'s header claimed since t194
that "the flock discipline is borrowed only to serialize mutations of that record". No exclusion
primitive existed anywhere in the file. The header now records that the clause described nothing
for the whole of the file's first life, names the specific damage — a reader who believed the
record was already serialized stopped looking — and points at the mechanism that finally makes it
true. This is verified by diff review, not by a test: no command can decide whether prose matches a
mechanism, and `acceptance.md` §D.4 says so rather than dressing it as an AC.

**Gaps.** The guard was run once in each direction. The 3-attempt escalation ceiling exists for a
zero-observation run and was not approached: the disabled direction produced the attributed
double-hold on attempt 1, as it did at M1.

**Residual risk.** The guard subtracts mutual exclusion at ONE call site. A future regression that
removes serialization by a different route — say, a second write path around
`withIntegrationLockMutation` — would not be caught by re-running this guard; it would be caught by
the criterion itself, which is why the shipped test asserts the invariant rather than the guard.

### M6 — Scoped verification

**Claim.** The whole touched package is green, no child process outlives the run, coverage clears
the 85% threshold, lint and vet are clean on both platforms' compilation, and both scope boundaries
are byte-empty.

**Evidence.**

```
$ pgrep -f "[k]anban\.test" | wc -l
       0                                    ← BEFORE

$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	14.261s
go test exit=0

$ pgrep -f "[k]anban\.test" | wc -l
       0                                    ← AFTER

$ go test -cover ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	14.080s	coverage: 86.5% of statements

$ golangci-lint run --timeout=3m ./internal/kanban/...
0 issues.
lint exit=0

$ go vet ./internal/kanban/...
vet exit=0

$ GOOS=windows go vet ./internal/kanban/...
exit=0 (no output)

$ git diff --stat a5f414a8a -- internal/hook internal/config
(no output)

$ git diff --stat a5f414a8a -- internal/cli
(no output)
```

**Baseline-attribution.** Measured in this run, against this tree at the M5 commit `ff9a03ce9`.
The package baseline for this card is `ok github.com/modu-ai/moai-adk/internal/kanban 13.758s`,
exit 0, captured at `a5f414a8a` before any edit and stored at
`.moai/state/verify/t336/baseline-kanban.txt`. Both runs are green, so every criterion below is
attributed against a green baseline and no new red exists to attribute.

**On the flake the lead flagged.** `TestConcurrencyStress` (`backlog_concurrency_test.go:19`, card
t354's) was reported red in this package. It did NOT fail in the pre-edit baseline run, and it did
not fail in either M6 run above. This card therefore has nothing to attribute to it in either
direction: it is neither reproduced nor repaired here, and no claim is made about whether it is
flaky — only that it passed in the three runs this lane observed.

**AC-ILA-009 measurement note, stated so the reading can be checked.** The AC's command is
`PAT=kanban; pgrep -f "[${PAT:0:1}]${PAT:1}\.test" | grep -v "^$$\$" | wc -l`. The form actually
run was `pgrep -f "[k]anban\.test" | wc -l` — the variable indirection expands to the identical
pattern, and the `$$` filter is dropped because the measuring shell's own command line is not
`kanban.test` and so cannot match the pattern in the first place. The load-bearing half is intact:
the bracketed first character is what keeps the measuring pipeline's own command line from
matching, and a `1`/`1` reading would have signalled a broken measurement rather than a clean run.
Both readings were `0`, not `1`, so the exclusion is working.

**Gaps.** No local full-suite run: `go test ./...` is prohibited in this repository (parallel lanes
running it drove machine load to 413 on 2026-08-15), so the full-suite verdict is CI's and is NOT
claimed here. `./internal/cli/...` was not run because no CLI test was touched and the CLI diff is
empty. AC-ILA-007(b) remains CI-judged. The busy path is still unexercised end to end (M2's gap
stands).

**Residual risk.** Coverage is a package figure (86.5%), not a per-file one; the windows-tagged
files contribute nothing to it on this platform, so the darwin coverage number says nothing about
the M3 code.

## §E.3 Run-phase Audit-Ready Signal

### AC PASS/FAIL matrix

| AC | Status | Deciding command | Actual output |
|---|---|---|---|
| AC-ILA-001 | **PASS** | `go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1 -v` at HEAD with the critical section absent (M1) and again with it disabled (M5) | `successes=2 refusals=0 other=0 sessions_differ=true mid_record_held=true mid_record_stale=false final_record_stale=false`; `attributed_double_hold=true`. Both children `REPLACED=none`; `SESSION=lane-a` ≠ `SESSION=lane-b`. Observed twice, attempt 1 each time |
| AC-ILA-002 | **PASS** | `go test ./internal/kanban/... -run TestIntegrationLockAcquire_SerializedAcrossProcesses -count=1` | `--- PASS (0.55s)`; `A: RESULT=acquired REPLACED=none SESSION=lane-a`, `B: RESULT=held REPLACED=none SESSION=lane-b`, `successes=1 refusals=1 other=0`, record names `lane-a`. No `RESULT=busy` in any round |
| AC-ILA-003 | **PASS** | `go test ./internal/kanban/... -run TestIntegrationLockAcquire_ConcurrencyPositiveControl -count=1` | `--- PASS (0.02s)`; `both children acquired as lane-same; record holder="lane-same"` |
| AC-ILA-004 | **PASS** | `go test ./internal/kanban/... -run TestIntegrationLockBusy_IsNotHeld -count=1` | `--- PASS (0.00s)`. Asserts three directions: busy satisfies its own predicate, `IsIntegrationLockHeld(busy)` false, `IsBoardLockHeld(busy)` false, and `IsIntegrationLockBusy(held)` false |
| AC-ILA-005 | **PASS** | `go test ./internal/kanban/... -run IntegrationLock -count=1` + `git diff --stat 15453140a -- internal/kanban/integration_lock_test.go` | 16/16 selected criteria PASS, `ok … 1.018s`; the 12 pre-existing ones (re-acquire, stale takeover, `--force`, foreign-release refusal, empty release, corrupt record, legacy record, pid-0) all green. Diff: **empty** |
| AC-ILA-006 | **PASS** | one-line revert, run, restore, run | Disabled: `successes=2 refusals=0`, `attributed_double_hold=true`. Restored: `successes=1 refusals=1`, `--- PASS`. Restoration byte-exact (diff against M3 empty). Both verbatim in §E.2 |
| AC-ILA-007(a) | **PASS** | `GOOS=windows go vet ./internal/kanban/...` | exit 0, no output. **Compilation only** — cited as nothing more |
| AC-ILA-007(b) | **CI-JUDGED** | CI windows job on the PR head | Not claimable from this lane. No darwin command compiles the windows-tagged code, and none is offered |
| AC-ILA-008 | **PASS** | `grep -n 'path + "\.tmp"' internal/kanban/integration_lock.go` + `go test … -run TestWriteIntegrationLock_UniqueStagingPath -count=1` | grep prints nothing (exit 1); base at `15453140a` prints `257:	tmp := path + ".tmp"` (exit 0). Test `--- PASS (0.00s)`, no residue, mode 0644 |
| AC-ILA-009 | **PASS** | `pgrep -f "[k]anban\.test" \| wc -l` before and after `go test ./internal/kanban/... -count=1` | `0` before, `0` after. Not `1`/`1`, so the self-match exclusion is working rather than the measurement being broken. Command-form note in §E.2 M6 |

MUST-PASS: AC-ILA-001, 002, 003, 005, 006, 009 — all six PASS.
SHOULD-PASS: AC-ILA-004, 007(a), 008 — all three PASS.
CI-JUDGED: AC-ILA-007(b) — deferred to CI, not claimed.

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| Existing test suite never broken | **PASS** | `go test ./internal/kanban/... -count=1` → `ok … 14.261s`, exit 0, against the `13.758s` green baseline at `a5f414a8a` |
| `internal/hook/**` and `internal/config/**` untouched | **PASS** | `git diff --stat a5f414a8a -- internal/hook internal/config` — empty |
| t320's release surface byte-identical | **PASS** | `git diff 15453140a -- internal/cli/integration.go` — empty; `git diff --stat a5f414a8a -- internal/cli` — empty |
| Repair gated by no configuration flag | **PASS** | `Workflow.IntegrationLock.Enabled` appears nowhere in the diff; the criteria run with it at its default `false` |
| Two lifetimes not conflated | **PASS** | No read path consults the mutation artifact to decide who holds the window (review of the diff); the artifact's stem is `integration-mutation.lock`, not the record's |
| Hook has no production setter | **PASS** | Closure gate 6, both the GNU `\s` and POSIX bracket forms — nothing |
| Children bounded by deadline AND `t.Cleanup` kill | **PASS** | `exec.CommandContext` + a `t.Cleanup`-registered `Process.Kill()` registered before `Wait`; AC-ILA-009 reads `0`/`0` |
| Test isolation under `t.TempDir()` | **PASS** | Every new test uses `t.TempDir()`; the primary checkout's `.moai/state/integration-lock.json` is untouched |

### Audit-ready signal

```yaml
run_complete_at: 2026-08-30
run_commit_sha: pending-backfill-m6
run_status: complete
ac_pass_count: 9          # 8 local PASS + AC-ILA-007 counted once; 007(b) deferred
ac_fail_count: 0
ac_ci_judged_count: 1     # AC-ILA-007(b)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run  # card branch stays local under this repository's git-flow; no push this phase
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0   # golangci-lint ./internal/kanban/... → "0 issues."
cross_platform_build:
  darwin_vet: pass        # go vet ./internal/kanban/... exit 0
  windows_vet: pass       # GOOS=windows go vet ./internal/kanban/... exit 0, compilation only
  windows_runtime: ci-judged
coverage_package: 86.5%   # threshold 85%; package figure, not per-file
total_run_phase_files: 13
m1_to_mN_commit_strategy: one commit per milestone, M1..M5; M6 is verification and evidence only
commits:
  m1: b429bacb1  # RED, observed and attributed; N1 + N2 closed
  m2: 06290a314  # the repair
  m3: 7f27a8ab9  # windows substrate
  m4: 445ec5f8f  # unique staging path
  m5: ff9a03ce9  # mutation guard, header truth, boundary checks
```

`run_commit_sha` carries a placeholder: this evidence lands IN the M6 commit, and a commit cannot
name its own hash. It is backfilled in a follow-up commit per the schema's SHA-placeholder
exemption.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
