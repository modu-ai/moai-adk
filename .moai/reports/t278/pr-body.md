# CI flake series fix — SPEC-CI-FLAKE-SERIES-001 (card t278, absorbs t270 · t271)

Three CI-dependent flaky tests, each fixed at the mechanism level (test-side unless noted).
All three were reproduced from CI run logs, fixed RED-first, and validated with mutant probes
(RED output observed verbatim before GREEN; see "Verification" below).

## The three mechanisms

### 1. `TestConcurrentSendPoll` — poller exit-rule TOCTOU (internal/sessionmsg)

The test's drain-loop exit rule treated a **pre-completion empty observation as terminal**: the
poller could observe 0 pending messages on a still-open channel (senders not yet done) and exit
early, losing messages sent in the race window. CI observed `received 97 messages, want 100`
(run 32774108273, attempt 1) — reproduced locally at exactly 97/100.

Fix (test-side, REQ-CFS-001): the exit rule is extracted into a `pollUntilDrained` helper shared
by both `TestConcurrentSendPoll` and a new reproduction test `TestPollerStopRule` (single source
of the rule, so the mutant probe exercises the real test's rule). On observing a closed channel,
the poller re-polls **once**; only a 0/0 on that re-poll is terminal; received counts are summed.
A `beforeStopCheck` handshake hook holds the poller deterministically inside the CI race window.

- RED: `--- FAIL: TestPollerStopRule (0.17s): received 97 messages, want 100` (identical to CI)
- Mutant probe (AC-CFS-001): reverting the rule to the old select-immediate-exit re-fails with
  97/100; restored, `go test -race -count=3` passes 3/3.

### 2. `TestAssertPairedHealthyEndToEnd` — alternation-fragile per-round estimator (internal/timing)

The calibrated arm of `AssertPaired` used a single statistic (per-round median ratio). That
estimator is alternation-fragile: alternation-locked asymmetry (true ratio 1.00x, but half the
rounds fn-dominant at 2.50x and half ref-dominant at 0.40x) trips it on a healthy workload. The
post-#1591 CI firing had exactly this shape (2.47x / 1.09x); local green runs sit at ~1.00x both
sides.

Fix (production-side, REQ-CFS-003/004): `reportPaired` now computes and logs **both** statistics
(per-round median + ratio-of-medians); a new pure helper `CheckRatioAnd` fails the calibrated arm
only when **both** exceed MaxUnits (AND-gate). `CheckRatio`'s signature and semantics are
unchanged — the existing pins at `paired_step_test.go:56/:79` pass unmodified. The p95/worst arm
is shared via `checkAbsolute`.

- RED: the new property test's expectation ("AND-gate survives alternation-locked asymmetry")
  observed against the old single-statistic estimator →
  `--- FAIL: TestCalibratedGateSurvivesAlternationLockedAsymmetry` (2.50x median, calibrated
  bound 2.00x), while the homogeneous-3x detection arm still PASSes (power reference).
- Property tests (synthetic distributions — same class as the existing pins):
  - (a) alternation-locked asymmetry: `CheckRatioAnd` passes with 0 errors **plus a falsifier
    arm** — the test permanently asserts the old `CheckRatio` (per-round alone) errors on the
    same distribution, so the estimator difference stays pinned.
  - (b) homogeneous 3x under load variation: `CheckRatioAnd` still errors (detection power
    preserved, REQ-CFS-004).
- Callers verified green unchanged: `TestAssertPairedHealthyEndToEnd` (`1.00x per-round / 1.12x
  medians`), `internal/harness TestRecordEvent100Sequential` (`0.99x / 0.98x`),
  `internal/hook TestBranchGuard_Latency` (`1.00x / 1.01x`, maxUnits 1.50x).

### 3. `TestConfigChange_RT005ReloadIntegration` — single-sample wall-clock assertion (internal/hook)

The test asserted one wall-clock sample of synchronous-return latency ≤ 100ms — a single-sample
bound under CI load noise. Fix (test-side only; zero production change, `git diff --stat`
confirmed clean): collect **20 samples** (fresh handler per sample, `testutil.WaitForAsync`
multiple between samples) and assert nearest-rank p95 (19th of 20 = second-largest) ≤ 100ms
(REQ-HAE-002). Functional invariants (Continue/SystemMessage) checked once on the first sample.

- Mutant RED (AC-CFS-006): a 150ms sleep injected into the synchronous path →
  `config_change_test.go:93: synchronous return p95 151.163958ms over 20 samples, want ≤ 100ms` /
  `--- FAIL: (3.44s)`.

## Evidence

- `.moai/reports/t278/forensics.md` — failure-log excerpts per flake, pinned with run IDs and
  output line numbers, re-verified from the run-phase tree.
- `.moai/reports/t278/reproduction-rate.md` — baseline window 2026-08-10~08-26: 537 runs (535
  `go_code=true` observed + 2 cancelled), 19 multi-attempt runs / 556 attempts checked, 4
  current-defect firings post-#1591 (p̂ = 4/166 ≈ 0.0241). Explicitly records that
  (1−p̂)^40 ≈ 0.377 — a 40-run green window alone is not proof, which is why the observation
  window below is run-count **AND** time-bounded.
- `.moai/reports/t278/timing-statistic-decision.md` — the AND-gate adoption decision with the
  arithmetic rejection of the trimmed-mean fallback (≥11 of 20 rounds skewed survives up to 20%
  trimming; only 45%+ trimming moves it, which destroys detection power).
- `.moai/specs/SPEC-CI-FLAKE-SERIES-001/` — spec.md / plan.md / acceptance.md / progress.md
  (full §E evidence battery with verbatim RED/GREEN/mutant output).

## Local verification (final tree `324883ebb`, all observed this run)

| Command | Result |
|---|---|
| `go test -race -count=20 ./internal/sessionmsg/` | `ok 57.477s` (AC-CFS-002) |
| `go test -race -count=5 ./internal/timing/` | `ok 3.146s` |
| `go test -race -count=1 ./internal/hook/` | `ok 168.898s` |
| `go test -race -count=5 -run 'TestConfigChange' ./internal/hook/` | `ok 4.216s` |
| `go build ./...` | exit 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| `golangci-lint run --timeout=2m` | `0 issues` (baseline 0 → NEW 0) |
| `go vet ./internal/{sessionmsg,hook,timing}/` | VET_OK each |

Local darwin passes are necessary-not-sufficient; this PR's CI jobs (Test ubuntu-latest, Race
Test) are the first post-fix observations.

## Post-merge observation-window plan (AC-CFS-007)

The formal flake-absence window starts at **merge**: ≥40 `go_code` runs **AND** ≥7 days, swept
attempt-aware via `.moai/reports/t278/sweep-attempts.sh` (v2 — re-runnable judgment command using
the `run_attempt` field + `gh run view --attempt N --json`, replacing the 404-ing
`.../runs/<id>/attempts` REST path measured in this repo). CI green-ratio channel gap is
recorded in reproduction-rate.md; web-summary readings will be logged during the window.

Closes card t278 (absorbs t270, t271).

🗿 MoAI
