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


## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
