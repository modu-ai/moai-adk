# Forensics — SPEC-CI-FLAKE-SERIES-001 (t278 M1)

Evidence fixed at run-phase tree `.claude/worktrees/t278` @ `d1289c5db` (branch `WT-ci-flake-series`, base `175d63f3f` = origin/main 2026-08-26). All log excerpts below were re-captured in this run-phase (2026-08-27) with `gh run view <id> --attempt <n> --log-failed` and are quoted verbatim; the number after the colon is the 1-based line offset within that command's output.

## 1. Flake #1 — `TestConcurrentSendPoll` (internal/sessionmsg)

### 1.1 Primary flare-up — run 32774108273 attempt 1 (fast job, required merge gate)

```console
$ gh run view 32774108273 --attempt 1 --log-failed | grep -n 'ConcurrentSendPoll\|store_test.go'
112:Test (ubuntu-latest)	Run tests with coverage (fast — no race detector)	2026-08-24T20:33:19.8564953Z --- FAIL: TestConcurrentSendPoll (0.07s)
113:Test (ubuntu-latest)	Run tests with coverage (fast — no race detector)	2026-08-24T20:33:19.8566145Z     store_test.go:385: received 97 messages, want 100 (loss or duplication)
```

- PR #1601 (docs-only, `WT-project-pipeline`). **Required gate job** — blocked merge.
- Duration 0.07s: unrelated to the 60s poll deadline (`store_test.go:351`). Zero `poll:` / `send` error lines anywhere in the failed log — the above 2 lines are the test's only output.
- Mechanism (spec §2.1, confirmed): stop-rule TOCTOU at `store_test.go:361-370` — the poller observes 0/0 pending at listing time (1), then confirms `sendersDone` closed at select time (3); the last 3 Sends' pending renames can land between the two instants, so both pollers exit with 3 messages forever unreceived.

### 1.2 Recurrence — run 32777242100 attempt 1 (race job, advisory)

```console
$ gh run view 32777242100 --log-failed | grep -n 'ConcurrentSendPoll'
3893:Race Test	Run race detector across all packages	2026-08-24T21:07:31.3326286Z --- FAIL: TestConcurrentSendPoll (0.27s)
```

Same day, 35 minutes later, different job (advisory race). Both job kinds are affected.

## 2. Flake #2 — `TestAssertPairedHealthyEndToEnd` (internal/timing)

### 2.1 Post-#1591 flare-up — run 32779472351 attempt 1 (race job)

```console
$ gh run view 32779472351 --attempt 1 --log-failed | grep -n 'AssertPaired\|timing.go'
754:Race Test	UNKNOWN STEP	2026-08-24T21:35:42.6609240Z --- FAIL: TestAssertPairedHealthyEndToEnd (0.13s)
755:Race Test	UNKNOWN STEP	2026-08-24T21:35:42.6611086Z     timing.go:233: paired-cpu-1x: n=20 median=1.502ms p95=4.479ms worst=5.355ms avg=2.812ms | refUnit=1.375ms ratio=2.47x (maxUnits=2.00x, steadyCeiling=10s, budget=30s)
756:Race Test	UNKNOWN STEP	2026-08-24T21:35:42.6615423Z     timing.go:237: paired-cpu-1x: measured latency is 2.47x the reference unit (median 1.502491ms), above the 2.00x calibrated bound — the operation now costs more machine-cost units than its design (e.g. an added subprocess or per-write fsync); this is a code regression, not machine load (load inflates the reference equally)
```

- PR #1600 (`WT-server-version`, internal/cli only). ref and fn are byte-identical `cpuUnit(2_000_000)` (`paired_test.go:68`).
- **Discriminator**: median(fn)=1.502ms vs median(ref)=refUnit=1.375ms → ratio-of-medians = **1.09x (healthy)**, while the per-round paired-ratio median reads 2.47x. Noise phase-locked to the round alternation (spec §2.2 candidate (i)).

### 2.2 Pre-#1591 occurrences (NEW this sweep — NOT current-defect baseline)

The attempt-aware sweep of the full window found two additional `--- FAIL: TestAssertPairedHealthyEndToEnd` occurrences in run **32429213275** (PR head `ecf9e337`, branch `WT-release-notes`), **both before PR #1591 merged**:

```console
$ gh run view 32429213275 --attempt 1 --log-failed | grep -A2 'FAIL: TestAssertPairedHealthyEndToEnd'
Test (ubuntu-latest)	UNKNOWN STEP	2026-08-20T23:39:29.8987941Z --- FAIL: TestAssertPairedHealthyEndToEnd (0.16s)
Test (ubuntu-latest)	UNKNOWN STEP	2026-08-20T23:39:29.8989093Z     timing.go:209: paired-cpu-1x: n=20 median=3.536ms p95=7.187ms worst=9.187ms avg=3.704ms | refUnit=1.3ms ratio=2.72x (maxUnits=2.00x, steadyCeiling=10s, budget=30s)
Test (ubuntu-latest)	UNKNOWN STEP	2026-08-20T23:39:29.8991243Z     timing.go:213: paired-cpu-1x: median latency 3.536237ms is 2.72x the reference unit 1.300194ms, above the 2.00x calibrated bound — the operation now costs more machine-cost units than its design (e.g. an added subprocess or per-write fsync); this is a code regression, not machine load (load inflates the reference equally)

$ gh run view 32429213275 --attempt 2 --log-failed | grep 'FAIL: TestAssertPairedHealthyEndToEnd'
Test (ubuntu-latest)	Run tests with coverage (fast — no race detector)	2026-08-21T02:10:38.6191376Z --- FAIL: TestAssertPairedHealthyEndToEnd (0.15s)
```

Timeline (measured):

| Event | Timestamp |
|---|---|
| run 32429213275 attempt 1 FAIL | 2026-08-20T23:39:29Z |
| run 32429213275 attempt 2 FAIL (rerun also failed) | 2026-08-21T02:10:38Z |
| **PR #1591 merged** (`test(timing): stop the measuring tests from measuring each other`, `bd463f55e`) | **2026-08-21T04:20:27Z** |

Classification: these two are the **pre-#1591 defect** (measuring tests overlapping each other's windows), already recorded as "2.32x / 2.72x / 4.64x" in `timing_test.go:21-22` — the 2.72x figure matches exactly. Their form is *bilateral* (ratio-of-medians ≈ per-round ≈ 2.72x: 3.536/1.3), unlike the post-#1591 form (1.09x vs 2.47x). They are counted in the sweep totals but **excluded from the current-defect baseline** (the code under test changed at #1591). Design consequence for the statistic decision: an AND-gate would still have failed these (both figures over bound) — correct behavior, since that flake form was real and was fixed by #1591.

## 3. Flake #3 — `TestConfigChange_RT005ReloadIntegration` (internal/hook)

```console
$ gh run view 32815411885 --attempt 1 --log-failed | grep -n 'RT005ReloadIntegration\|config_change_test'
205:Race Test	Run race detector across all packages	2026-08-25T06:12:45.1059931Z --- FAIL: TestConfigChange_RT005ReloadIntegration (0.15s)
206:Race Test	Run race detector across all packages	2026-08-25T06:12:45.1060749Z     config_change_test.go:51: synchronous return took 123.195919ms, want ≤ 100ms (REQ-HAE-002)
```

- PR #1653 (docs-only). The failing assertion is the single-sample wall-clock check (`config_change_test.go:50-52`), not `WaitForAsync` (2s budget — a timeout there would show ≥2s). Sync path is µs-scale (`config_change.go:56-84`: one slog.Info + WithTimeout + goroutine spawn); 123ms is a scheduler preemption inside the measurement window (t.Parallel test + package-wide -race + 2-vCPU runner). Contract is p95 (REQ-HAE-002 / AC-HAE-003, measured by `BenchmarkConfigChange_AsyncReturn`) — category error in enforcement (spec §2.3).

## 4. Channel notes (measurement caveats fixed during this run)

- The REST path `actions/runs/<id>/attempts` referenced by plan.md §D-M1 item 2 returns **404 for all 537/537 runs** on this repo (measured 2026-08-27). Attempts were enumerated instead from each run's `run_attempt` field (highest attempt number) with per-attempt conclusions from `gh run view --attempt N --json conclusion`. Both commands verified on run 32774108273: run_attempt=2, attempt 1 = `failure`, attempt 2 = `success`.
- `grep` on bare test names matches slog INFO lines too: 7 additional runs matched `TestConfigChange_RT005ReloadIntegration` only via its INFO output (`config file changed session_id=sess-rt005 ...`) with no `--- FAIL:` line — the runs failed for other reasons and RT005 itself passed. Occurrence counting therefore keys on `--- FAIL: <test>`.
- CI green-run ratio distribution is NOT extractable via `gh` CLI: `publish()` writes to `GITHUB_STEP_SUMMARY` (web-UI-only surface), and a passing package's `t.Log` is discarded by non-verbose `go test` (both CI jobs). Local reference distribution was measured instead (see reproduction-rate.md §3) and the channel gap is recorded there.
