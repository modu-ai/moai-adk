# Acceptance Criteria — SPEC-V3R6-CI-FLAKY-STABILIZE-002

> Tier S LEAN. Each AC has a PASS evidence command. All commands are read-only verification (except the AC-CFS2-002 temporary-edit-then-revert proof, which must leave the tree byte-identical afterward). FLAKY-1 = `internal/lsp/subprocess`; FLAKY-2 = `internal/hook/quality`.

## §A — Definition of Done

- [ ] M1 (FLAKY-1 comma-ok + skip), M2 (FLAKY-2 timeout 5s→60s), M3 (green-gate + no-prod-diff) complete.
- [ ] All 7 AC below PASS.
- [ ] `git diff` touches ONLY the two `*_test.go` files; `supervisor.go` / `client.go` / `gate.go` byte-identical to HEAD.
- [ ] Deterministic production-behavior assertions preserved (FLAKY-1 exit-0 stub still FAILs "want 1"; FLAKY-2 routing `passed == true` preserved).
- [ ] `status: draft` at plan-phase; transitions to `in-progress` on first run-phase commit (manager-develop), `implemented` on sync (manager-docs).

## §B — Severity Map

| Severity | AC |
|----------|----|
| MUST-PASS (blocking) | AC-CFS2-001, AC-CFS2-002, AC-CFS2-003, AC-CFS2-004, AC-CFS2-006 |
| SHOULD-PASS | AC-CFS2-005, AC-CFS2-007 |

## §C — Given-When-Then Scenarios

### Scenario 1 — FLAKY-1 stabilized under CPU contention
**Given** a CPU-saturated machine (16× `yes` burners + concurrent `go build ./...`),
**When** `TestSupervisor_NonZeroExit` / `TestSupervisor_NormalExit` run repeatedly,
**Then** any instance where the 5s ctx deadline fires before the stub exits surfaces as a SKIP (with the load-attribution message), NEVER as a FAIL with "ExitCode = 0, want 1".

### Scenario 2 — FLAKY-1 still catches a genuine wrong exit code
**Given** the stub is temporarily changed to `exit 0` (so a real `ExitEvent{ExitCode:0}` is delivered, `ok == true`),
**When** `TestSupervisor_NonZeroExit` runs,
**Then** it FAILs deterministically with "ExitCode = 0, want 1" (the comma-ok guard skips only the closed-empty/timeout case, not a delivered wrong code).

### Scenario 3 — FLAKY-2 stabilized under CPU contention
**Given** a CPU-saturated machine (32× `yes` + `GOMAXPROCS=1` + concurrent broad suite),
**When** `TestQualityGate_RunsDotnetFormatWhenCSharpStaged` runs repeatedly,
**Then** the 60s budget (>> fake-binary cost even under saturation) yields 0 FAIL, while the routing assertion `passed == true` is still exercised.

## §D — AC Matrix (PASS evidence commands)

### AC-CFS2-001 — FLAKY-1 stress repro neutralized (MUST-PASS) — REQ-CFS2-001/002/003

Under CPU saturation, the two Supervisor tests yield 0 FAIL; timeout instances surface as SKIP, never FAIL.

```bash
# Start saturation in the background, then run the targeted tests under -count=30.
( for i in $(seq 1 16); do yes >/dev/null & done; \
  go build ./... >/dev/null 2>&1 & )
go test -run 'TestSupervisor_(NonZeroExit|NormalExit)' -count=30 ./internal/lsp/subprocess/ -v 2>&1 \
  | tee /tmp/cfs2_flaky1.log
# stop burners
pkill -x yes 2>/dev/null || true
# PASS: no "--- FAIL" lines; any timeout shows "--- SKIP" with the load-attribution message.
grep -c -- '--- FAIL' /tmp/cfs2_flaky1.log   # expect 0
```

PASS criterion (one-sided, deterministic): `go test` exit 0 AND `grep -c -- '--- FAIL'` returns `0`. The pass test is the **absence of FAIL** — NOT the presence of a SKIP line.

> **Proxy note (D-1)**: On a fast / lightly-loaded machine the 5s deadline may never fire, so a SKIP line may not appear; this is expected and is NOT a failure of the AC. "Absence of FAIL under `-count=30`" is the accepted deterministic proxy for stress-neutralization. Requiring a SKIP line to appear would make this AC contention-dependent (non-deterministic across machines), so it is deliberately NOT required. The correctness of the skip-branch itself (the `ok == false` closed-empty path) is independently proven by the deterministic AC-CFS2-002 (`ok == true` delivered-wrong-code path still FAILs "ExitCode = 0, want 1") — together AC-CFS2-001 (no spurious FAIL) and AC-CFS2-002 (genuine FAIL preserved) bound both sides without a contention-dependent assertion.

### AC-CFS2-002 — FLAKY-1 deterministic prod assertion preserved (MUST-PASS) — REQ-CFS2-004

Temporarily editing the stub to `exit 0` makes `TestSupervisor_NonZeroExit` FAIL with "ExitCode = 0, want 1" — proving the comma-ok guard skips only the closed-empty/timeout case, not a delivered wrong code. The tree MUST be reverted to byte-identical afterward.

```bash
# Temporarily flip the NonZeroExit stub "exit 1" → "exit 0" (manual edit or sed in a scratch copy),
# run the test, confirm a deterministic FAIL, then revert.
go test -run 'TestSupervisor_NonZeroExit' ./internal/lsp/subprocess/ -v 2>&1 | tee /tmp/cfs2_flaky1_neg.log
grep -q 'ExitCode = 0, want 1' /tmp/cfs2_flaky1_neg.log && echo "PASS: deterministic FAIL on wrong code"
# REVERT the stub edit — verify clean afterward:
git diff --quiet internal/lsp/subprocess/supervisor_test.go && echo "reverted clean" || echo "NEEDS REVERT"
```

PASS criterion: with stub flipped to `exit 0` the run FAILs carrying "ExitCode = 0, want 1"; after revert `git diff --quiet` confirms the file matches the committed comma-ok version (no stray test edit remains).

### AC-CFS2-003 — FLAKY-1 race-clean (MUST-PASS) — REQ-CFS2-005

`supervisor.go` unchanged; the package is race-clean under repeated runs.

```bash
go test -race -count=20 ./internal/lsp/subprocess/ 2>&1 | tail -5
```

PASS criterion: exit 0, no `--- FAIL`, no `WARNING: DATA RACE`.

### AC-CFS2-004 — FLAKY-2 stress repro neutralized (MUST-PASS) — REQ-CFS2-006/007

Under the same-class harness, the dotnet routing test yields 0 FAIL (60s >> fake-binary cost even under saturation).

```bash
( for i in $(seq 1 32); do yes >/dev/null & done & )
GOMAXPROCS=1 go test ./internal/hook/quality/ -run Dotnet -count=40 -v 2>&1 \
  | tee /tmp/cfs2_flaky2.log &
go test ./internal/... >/dev/null 2>&1 &   # concurrent broad suite to amplify contention
wait
pkill -x yes 2>/dev/null || true
grep -c -- '--- FAIL' /tmp/cfs2_flaky2.log   # expect 0
```

PASS criterion (one-sided, deterministic): `grep -c -- '--- FAIL'` returns `0`; the routing test passes 40/40.

> **Proxy note (D-1)**: As with AC-CFS2-001, the pass test is the **absence of FAIL** under `-count=40` — the one-sided `grep -c -- '--- FAIL'` == 0 criterion. On a fast / lightly-loaded machine the 5s→60s change has no observable effect (the fake exit-0 subprocess always completes well under either bound), so this AC cannot certify the headroom is *exercised* on every machine; "absence of FAIL" is the accepted deterministic proxy for stress-neutralization. The 60s headroom is independently justified by the production-SLA mirror (`DefaultGateConfig.LintTimeout = 60s`, `gate.go:46`) — the test timeout now matches the production budget for this step rather than an arbitrary under-spec bound. The routing assertion `passed == true` (preserved per AC-CFS2-005 / REQ-CFS2-009) bounds the other side: a genuine routing regression still FAILs deterministically regardless of timeout.

### AC-CFS2-005 — Isolation regression + routing/immunity preserved (SHOULD-PASS) — REQ-CFS2-008/009

The two dotnet routing tests (Runs + the immune Skips sibling) are 100% green in isolation, and the immune sibling is unchanged.

```bash
go test ./internal/hook/quality/ -run 'TestQualityGate_(Runs|Skips)DotnetFormat' -count=20 2>&1 | tail -3
go test ./internal/lsp/subprocess/ -count=20 2>&1 | tail -3
# Confirm the immune sibling was NOT edited:
git diff internal/hook/quality/gate_test.go | grep -A3 'SkipsDotnetFormatWhenNoCSharpStaged' || echo "immune sibling untouched"
```

PASS criterion: both `-count=20` runs exit 0 with no `--- FAIL`; the immune sibling shows no diff hunk.

### AC-CFS2-006 — No production diff (MUST-PASS) — REQ-CFS2-005/010/011

`git diff --stat` shows ONLY the two `*_test.go` files; production files byte-identical to HEAD.

```bash
git diff --stat
# Expect exactly two paths:
#   internal/lsp/subprocess/supervisor_test.go
#   internal/hook/quality/gate_test.go
# Byte-identity of the 3 PRESERVE production files (all mechanical/executable):
git diff --quiet HEAD -- internal/lsp/subprocess/supervisor.go && echo "supervisor.go identical"
git diff --quiet HEAD -- internal/lsp/core/client.go && echo "client.go identical"
git diff --quiet HEAD -- internal/hook/quality/gate.go && echo "gate.go identical"
```

`internal/lsp/core/client.go` is the sole production `Watch` caller (interface contract at client.go:32, call site at client.go:359) — confirmed at plan-audit, so the byte-identity check is fully mechanical (no run-phase path resolution needed).

PASS criterion: `git diff --stat` lists exactly the two `*_test.go` paths; each of the three `git diff --quiet HEAD -- <prod-file>` checks returns exit 0 (no diff).

### AC-CFS2-007 — Full-suite stability (SHOULD-PASS) — REQ-CFS2-012

`go test ./...` passes; or at minimum the two affected packages under `-count=5`.

```bash
go test ./... 2>&1 | tail -15
# Minimum gate if full-suite is impractical locally:
go test ./internal/lsp/subprocess/ ./internal/hook/quality/ -count=5 2>&1 | tail -5
```

PASS criterion: the two affected packages pass under `-count=5` with no `--- FAIL`; full-suite `ok`/no-FAIL where runnable.

## §E — Traceability

| REQ | AC | Milestone |
|-----|----|-----------|
| REQ-CFS2-001 / -002 / -003 | AC-CFS2-001 | M1 |
| REQ-CFS2-004 | AC-CFS2-002 | M1 |
| REQ-CFS2-005 | AC-CFS2-003, AC-CFS2-006 | M1, M3 |
| REQ-CFS2-006 / -007 | AC-CFS2-004 | M2 |
| REQ-CFS2-008 / -009 | AC-CFS2-005 | M2 |
| REQ-CFS2-010 / -011 | AC-CFS2-006 | M3 |
| REQ-CFS2-012 | AC-CFS2-007 | M3 |

## §F — Quality Gate Summary

- All MUST-PASS AC (001, 002, 003, 004, 006) green → blocking gate satisfied.
- SHOULD-PASS AC (005, 007) green → no debt.
- Production diff = 0 (AC-CFS2-006) is the cardinal invariant: a non-empty production diff is an automatic blocker regardless of other AC.
