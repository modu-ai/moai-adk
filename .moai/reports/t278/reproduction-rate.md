# Reproduction Rate — SPEC-CI-FLAKE-SERIES-001 (t278)

Baseline measured 2026-08-27 at run-phase tree `d1289c5db` (branch `WT-ci-flake-series`). Post-merge section will accrue from M3 onward.

## §1 Sweep recipe (attempt-aware, REQ-CFS-009)

- Window: **2026-08-10T00:09:07Z → 2026-08-26T14:55:28Z** (workflow `ci.yml` runs only — the sole go_code-gated workflow that runs the affected jobs).
- Run enumeration: `gh run list --workflow ci.yml -L 600 --json databaseId,conclusion,createdAt,event` → filter `createdAt >= 2026-08-10` → **537 runs** (439 `pull_request` + 98 `push`).
- Attempt enumeration per run: `run_attempt` field (highest attempt number) from `gh api repos/modu-ai/moai-adk/actions/runs/<id>`; per-attempt conclusion via `gh run view <id> --attempt <n> --json conclusion`. (The `.../attempts` REST path returns 404 on this repo — see forensics.md §4.)
- Failed attempts inspected with `gh run view <id> --attempt <n> --log-failed | grep '--- FAIL: <test>'` for the 3 test names.
- Sweep scripts (committed, re-runnable): `.moai/reports/t278/sweep-attempts.sh` (v2) + `.moai/reports/t278/refetch-jobs.sh` (rate-limit repair for 16 jobs files).

## §2 Baseline occurrence table (Claim)

| # | Test | Run | Attempt | Job | Timestamp (UTC) | Form |
|---|------|-----|---------|-----|-----------------|------|
| 1 | TestConcurrentSendPoll | 32774108273 | 1 | Test (ubuntu-latest) — **required gate** | 2026-08-24T20:33:19Z | 97/100 loss, 0 errors, 0.07s |
| 2 | TestConcurrentSendPoll | 32777242100 | 1 | Race Test (advisory) | 2026-08-24T21:07:31Z | recurrence, 0.27s |
| 3 | TestAssertPairedHealthyEndToEnd | 32779472351 | 1 | Race Test (advisory) | 2026-08-24T21:35:42Z | per-round 2.47x / ratio-of-medians 1.09x |
| 4 | TestConfigChange_RT005ReloadIntegration | 32815411885 | 1 | Race Test (advisory) | 2026-08-25T06:12:45Z | single-sample 123.195919ms vs ≤100ms |
| — | *(pre-#1591, excluded from current-defect baseline)* | 32429213275 | 1 | Test (ubuntu-latest) | 2026-08-20T23:39:29Z | bilateral 2.72x (defect fixed by #1591, merged 2026-08-21T04:20:27Z) |
| — | *(pre-#1591, excluded)* | 32429213275 | 2 | Test (ubuntu-latest) | 2026-08-21T02:10:38Z | rerun of same run also failed |

Rerun-green concealment re-measured on this window: `gh run list --status failure` style views key on the LATEST attempt; the sweep inspected every attempt and found **19 multi-attempt runs / 556 total attempts** across the 537-run window.

## §3 Green ratio distribution (statistic-decision input)

- **CI channel gap (Gaps)**: `publish()` writes `GITHUB_STEP_SUMMARY`, a web-UI surface with no `gh` CLI read path; green runs' `t.Log` lines are discarded by non-verbose `go test` (both CI jobs run non-verbose). No CI green `paired-cpu-1x` line is mechanically extractable — ≥10 green summary lines could NOT be collected locally in M1. Per plan.md §D-M1 item 3 fallback, the local reference distribution was measured and the gap stands recorded here.
- **Local reference distribution** (command: `go test -race -count=10 -run TestAssertPairedHealthyEndToEnd -v ./internal/timing/` on darwin, tree `d1289c5db`, 2026-08-27, all PASS, exit 0):

| run | median | refUnit | ratio |
|-----|--------|---------|-------|
| 1 | 514µs | 515µs | 1.00x |
| 2 | 503µs | 503µs | 0.99x |
| 3 | 514µs | 507µs | 1.01x |
| 4 | 515µs | 513µs | 1.00x |
| 5 | 514µs | 511µs | 1.00x |
| 6 | 512µs | 511µs | 1.00x |
| 7 | 507µs | 507µs | 1.00x |
| 8 | 495µs | 496µs | 1.00x |
| 9 | 514µs | 507µs | 1.00x |
| 10 | 515µs | 514µs | 1.00x |

Healthy-form evidence: on an unloaded machine the two estimators coincide (0.99x–1.01x), i.e. the divergence seen in occurrence #3 (1.09x vs 2.47x) is a CI-load artifact, not a code property. CI green figures accrue in the post-merge window via the job-summary channel (manual visibility) — recorded as they surface.

## §4 Denominator arithmetic (for REQ-CFS-010 `(1-p̂)^N` confidence)

Observation unit N = count of `go_code=true` workflow runs (spec v0.2.0 D2: run count, NOT job-instance count; both affected jobs enumerated per run).

| Quantity | Value | Derivation |
|----------|-------|------------|
| ci.yml runs in window | 537 | run list filter |
| — with Test job executed (go_code=true) | 535 | jobs API, Test job present under either name era (union 535; `Test (ubuntu-latest)` 529 runs ∩ unexpanded `Test (${{ matrix.os }})` rendering 535 runs — names overlap, not a partition; 0 skipped) |
| — cancelled before jobs started | 2 | runs 31870129254, 32805423009 (empty jobs list) |
| Race Test job present | 386 | jobs API (condition `!startsWith(head_ref,'release/')` + workflow evolution) |
| Runs with >1 attempt | 19 | `run_attempt > 1` |
| Total attempts inspected | 556 | 518 single-attempt + 38 across multi-attempt runs |
| Failed attempts grepped for the 3 names | 52 | `--log-failed` + grep |
| True `--- FAIL:` matches | 6 | 4 current-defect + 2 pre-#1591 |
| **Post-#1591 runs (current-defect exposure)** | **166** | createdAt ≥ 2026-08-21T04:20:27Z (#1591 merge) |
| **Current-defect baseline p̂ (3 tests pooled)** | **4/166 ≈ 0.0241** | occurrences 1–4 over post-#1591 runs |

Per-test (post-#1591): TestConcurrentSendPoll 2/166 ≈ 0.0120 · TestAssertPairedHealthyEndToEnd 1/166 ≈ 0.0060 · TestConfigChange_RT005ReloadIntegration 1/166 ≈ 0.0060.

**Power arithmetic for the future verdict** (0 recurrences over N post-merge runs):

| N | P(0 events \| fix ineffective) pooled p̂=0.0241 | per-test p̂=0.0060 |
|---|-----|------|
| 40 (spec minimum) | 0.9759^40 ≈ **0.377** | 0.9940^40 ≈ 0.786 |
| 100 | ≈ 0.087 | ≈ 0.549 |
| 166 | ≈ 0.017 | ≈ 0.368 |

Reading: a 0/40 window leaves a 37.7% chance the flake simply did not fire — the spec's N≥40 is a floor, not proof; the ≥7-day calendar bound and run-ID ledger are the honest instruments. This number goes verbatim into the M4 verdict's Residual-risk (spec R3).

## §5 Post-merge ledger (accruing from M3)

Window opened **2026-08-26T18:05:57Z** — PR #1666 squash merge `379b310a6` landed on `main` (operator-approved via AskUserQuestion after CI read: required 5/5 pass, windows-latest Integration 13m41s pass; non-required `graph-freshness` fail attributed to orphan stamp `0d15864ae90b`, unrelated to this diff, main HEAD green).

Inclusion boundary: ci.yml runs with `createdAt >= 2026-08-26T18:05:57Z` (the merge-commit push run included); `go_code=true` = Test job present via jobs API (same name-union as §4). Attempt-aware per §1 recipe. Window close per REQ-CFS-010: N ≥ 40 `go_code=true` runs AND ≥ 7 days.

| Date | Run | Attempt | Job | Test | Result |
|------|-----|---------|-----|------|--------|
| 2026-08-26T18:06:04Z | 32997835484 | 1 | `Test (ubuntu-latest)` success 18:14:18Z · `Race Test` success 18:16:19Z | 3 tests | **0 recurrences** — run conclusion `cancelled` (concurrency: superseded by main push run 32999196269 at 18:20:53Z; only `Integration Tests (windows-latest)` job was cancelled). Test/Race — the jobs hosting the 3 flaky tests — completed green before cancellation, so the observation is valid; `go_code=true`, counted (N=1) |
| 2026-08-26T18:20:53Z | 32999196269 | 1 | all jobs success (incl. windows-latest Integration) | 3 tests | **0 recurrences** — `go_code=true`, counted (N=2) |

Window status after first accrual: N=2/40, elapsed 0/7 days (earliest close 2026-09-02T18:05:57Z). Run 33000104346 (18:30:58Z, push main) in progress at ledger update — accrues on completion. Note the sync-audit watch item applied: conclusion state confirmed before accrual (cancelled ≠ unobserved).

## §6 Gaps / Residual-risk (baseline section)

- **Gap**: CI green-run ratio distribution not mechanically collected (channel gap, §3) — local darwin reference only; CI-side figures rely on the web job summary.
- **Gap**: the sweep window starts 2026-08-10; any flare-up before that is outside the denominator (log retention 90d would allow deeper, but the card window is 08-10+).
- **Residual risk**: p̂ estimates rest on 4 occurrences — a single additional event moves pooled p̂ by ~25% relative. The verdict must carry the power arithmetic above, not a bare "0 recurrences".
- **Residual risk**: `cancelled` runs (2) and `Test (macos-latest)`-era job-name variance were classified by jobs-API presence; a workflow rename inside a future window would need the name-union recomputed.
