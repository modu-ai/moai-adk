# Sync Audit — SPEC-CI-FLAKE-SERIES-001 (t278)

Independent post-implementation audit (sync-auditor). Audit tree: `.claude/worktrees/t278` @
`f0ecc5001` (branch `WT-ci-flake-series`, clean working tree, verified this run). Judged against
`acceptance.md` §D (10 ACs). Run-phase evidence re-observed or traced to attributable artifacts;
progress.md claims were treated as suspect until re-observed.

## 1. Verdict

**PASS-WITH-DEBT** — aggregate harmonic mean **0.942**

9/10 AC GREEN + AC-CFS-007 **PENDING** (observation window arithmetic, not a defect). The PENDING
is the documented debt: post-merge window requires ≥7 days from 2026-08-26T18:05:57Z **AND**
N≥40 `go_code=true` runs; at audit time ~1 day elapsed and N=1/40 (measured this run — see
AC-CFS-007 row). plan.md §D-M4 explicitly permits PENDING with remaining window stated. Declaring
(b) GREEN today would be a false completion claim.

## 2. AC Matrix

| AC | Verdict | Evidence (command → observed, this run unless attributed) | Tree |
|----|---------|-----------------------------------------------------------|------|
| AC-CFS-001 stop-rule TOCTOU | **GREEN** | `go test -race -count=3 -run 'TestPollerStopRule\|TestConcurrentSendPoll' -v ./internal/sessionmsg/` → both tests PASS ×3, `ok 2.553s`. Code read: `pollUntilDrained` re-poll-after-`sendersDone` rule + `beforeStopCheck` handshake gate force the CI interleaving deterministically (stoprule_test.go:20-67, 111-116). Mutant RED 97/100: attributed (progress.md §E.2 E8 verbatim, `324883ebb`; post-`324883ebb` commits are docs-only — verified via `git log --oneline` this run) | `f0ecc5001` (tests) / `324883ebb` (mutant, attributed) |
| AC-CFS-002 sessionmsg continuity | **GREEN** | Attributed: `-race -count=20` → `ok 57.477s` @ `324883ebb`; re-observed partially this run as -count=3 (above, same package, all PASS). Code unchanged between the two trees (docs-only commits). Note per acceptance.md: local green is a necessary condition only — CI-side judgment rides AC-CFS-007 | `f0ecc5001` / `324883ebb` (attributed) |
| AC-CFS-003 alternation asymmetry passes | **GREEN** | `go test -race -count=5 ./internal/timing/` → `ok 2.542s`. `TestCalibratedGateSurvivesAlternationLockedAsymmetry` (t.Parallel — synthetic pure-judgment, no measurement) asserts `CheckRatioAnd` returns 0 errors on the 1.00x-true distribution **and** the falsifier arm asserts old `CheckRatio` DOES trip (paired_asym_test.go:62-72 — code verified). Pre-existing pin `TestCalibratedRatioSurvivesOffsetLoadStep` untouched (absent from diff) and passes | `f0ecc5001` |
| AC-CFS-004 homogeneous 3x detection preserved | **GREEN** | Same command; `TestPairedAndGateStillCatchesHomogeneousRegression` asserts `CheckRatioAnd` fires exactly 1 calibrated error on uniform 3x under varying load (paired_asym_test.go:79-99). `CheckRatioAnd` AND-logic verified at timing.go:366-376 | `f0ecc5001` |
| AC-CFS-005 statistic decision record | **GREEN** | `timing-statistic-decision.md` exists; carries (a) AND-gate vs fallback comparison, (b) measured basis — run 32779472351 a1 form (2.47x/1.09x), local 10×0.99–1.01x, channel gap stated, (c) caller census re-run this audit: `grep -rn "AssertPaired(" --include="*_test.go" internal/` → exactly 3 sites (paired_test.go:61, observer_test.go:251, pre_tool_branch_guard_integration_test.go:207) — matches the report's table | `f0ecc5001` |
| AC-CFS-006 p95 re-build | **GREEN** | `go test -race -count=5 -run 'TestConfigChange' ./internal/hook/` → `ok 4.089s`. Code read: 20 fresh-handler samples, `WaitForAsync` drainage between samples, nearest-rank p95 = 19th of 20, assert ≤100ms (config_change_test.go:46-94); single wall-clock-maximum form ABSENT (acceptance FAIL condition not met); functional invariants first-sample-only as documented; `BenchmarkConfigChange_AsyncReturn` untouched. Mutant RED (150ms→151.16ms p95): attributed (progress.md §E.2 E8, `324883ebb`) | `f0ecc5001` (tests+code) / `324883ebb` (mutant, attributed) |
| AC-CFS-007 reproduction-rate verdict | **PENDING** | (a) **GREEN** — reproduction-rate.md §1-§4: attempt-aware sweep, window 08-10→08-26, 537 runs / 556 attempts, 4 current-defect occurrences, pooled p̂=4/166≈0.0241 with per-test split; scripts committed and re-runnable. (c) **GREEN** — §4 power arithmetic ((1-p̂)^40≈0.377 — N≥40 named a floor, not proof) + §6 Gaps/Residual-risk. (b) **PENDING — cannot be judged today**: window opened 2026-08-26T18:05:57Z (PR #1666 squash `379b310a6`); this audit measured `gh run list --workflow ci.yml` (2026-08-27) → N=1 run in window (`32997835484`, still `in_progress`, conclusion not yet accrued), elapsed ~1 day of ≥7. Remaining: ≥6 days AND N 39 more `go_code=true` runs with 0 recurrences of the 3 tests | `f0ecc5001` (measurement) |
| AC-CFS-008 series report | **GREEN** | `series-analysis.md` exists; §1 series table (3 cases × occurrence counts × forms, pooled p̂ consistent with reproduction-rate.md §4), §2 common factor + per-case single-observation/window/contract-statistic mapping, §3 reusable authoring rule (positive condition · contract statistic · mutant proof) verbatim as acceptance requires. Cross-checked against progress.md §E.2 — no divergence found | `f0ecc5001` |
| AC-CFS-009 scope discipline | **GREEN** | `git diff --stat 175d63f3f..HEAD` → code files: `internal/sessionmsg/{stoprule_test.go(new), store_test.go}`, `internal/timing/{timing.go, paired_asym_test.go(new)}`, `internal/hook/config_change_test.go` — exactly the named set (acceptance names "timing 테스트 2종" = the 2 new tests inside paired_asym_test.go); remainder is `.moai/reports/t278/*` + SPEC 4종. `store.go`, CI workflows, unrelated tests: **0** entries | `f0ecc5001` |
| AC-CFS-010 non-parallel markers | **GREEN** | func-boundary scan re-run this audit (awk, per-file): marker-owning functions = `TestMeasurePairedEqualSampleCounts`, `TestAssertPairedHealthyEndToEnd`, `TestMeasureCalibratedRatioHealthy`, `TestMeasureCalibratedRatioTripsAt4x`, `TestAssertHealthyEndToEnd` (5/5, matching progress.md). t.Parallel-owning functions enumerated separately; **intersection empty → VIOLATION 0**. New timing tests are pure synthetic-distribution judgments (no measurement — code verified: constructed `time.Duration` slices into pure helpers) so t.Parallel is sound and marker-joining does not apply; hook p95 test is outside REQ-CFS-008's timing-package marker set, with p95-under-concurrent-load itself the contract | `f0ecc5001` |

## 3. Dimension Scores

| Dimension | Weight | Score | Verdict | Evidence |
|-----------|--------|-------|---------|----------|
| Functionality | 40% | **0.95** | PASS | 9/10 AC GREEN re-observed this run; the 10th is time-gated, not broken. All 5 judging commands re-executed green on `f0ecc5001` |
| Security | 25% | **0.95** | PASS | Diff is tests + one internal library file; `git diff 175d63f3f..HEAD -- internal/ \| grep -inE "password\|secret\|token\|api[_-]?key\|..."` (non-test) → 0 matches. No new input surface, no network, no filesystem writes beyond `GITHUB_STEP_SUMMARY` append (pre-existing, failure-ignored by design). `t.TempDir()` isolation throughout new tests |
| Craft | 20% | **0.92** | PASS | RED-first with mutant probes for all 3 fixes (E8 verbatim records); falsifier arm built into the property test; pre-existing pins untouched and passing; `golangci-lint run --timeout=2m` → `0 issues` (re-measured this run); `go vet` ×3 packages → VET_OK; cross-platform `go build` (darwin + windows/amd64) → exit 0 both |
| Consistency | 15% | **0.95** | PASS | Matches package idiom (error-message form, comment density, `publish` pattern); shared `checkAbsolute` extraction instead of duplication; Conventional Commits with SPEC-ID + card id; artifacts exactly at plan §F paths; scope 0-violation |

**Aggregate (harmonic mean): 4 / (1/0.95 + 1/0.95 + 1/0.92 + 1/0.95) ≈ 0.942**

Must-pass firewall (Functionality + Security): both independently above threshold — firewall
satisfied.

## 4. Five-Section Evidence Report

**Claim** — The implementation satisfies 9 of 10 acceptance criteria with verifiable evidence;
AC-CFS-007(b) cannot complete before 2026-09-02T18:05:57Z + N≥40, and is honestly PENDING.

**Evidence** — All AC rows in §2 carry the judging command plus observed output, executed this
audit run at tree `f0ecc5001` except where explicitly marked attributed to `324883ebb` (final
code tree; subsequent commits `ead7130ed`, `c0613a815`, `30f3eddd0`, `f0ecc5001` verified
docs-only via `git log --oneline -10` + diff scope).

**Baseline-attribution** — Every re-executed command was run in this worktree, against
`f0ecc5001`, on 2026-08-27. Attributed items cite their source tree and the verbatim record
location (progress.md §E.2 E8 blocks; §E battery). The attributed-vs-re-executed boundary is
marked per row.

**Gaps** (explicitly NOT observed this audit):
1. AC-CFS-007(b) post-merge recurrence-free window — time-gated; N=1/40, ~1/7 days at audit.
2. Mutant-probe RED outputs (AC-CFS-001, AC-CFS-006) — accepted as attributed records; mutants
   were NOT re-injected in this audit (would have mutated the audited tree).
3. sessionmsg `-count=20` full pass and hook full-suite `-count=1` — attributed at `324883ebb`;
   this audit re-ran the scoped forms (-count=3 stop-rule pair; TestConfigChange -count=5).
4. CI green ratio distribution — mechanically uncollectable (channel gap, documented in
   reproduction-rate.md §3); local reference distribution stands.
5. Coverage percentages (`go test -cover`) — not re-measured (load discipline; the executed test
   commands serve as the functional proxy).

**Residual-risk**:
1. Statistical power: even a future 0/40 window leaves ≈37.7% probability the flake simply did
   not fire (pooled p̂). The final verdict must carry this arithmetic — the report already
   mandates it (reproduction-rate.md §4).
2. First in-window run `32997835484` has been `in_progress` for >24h (observed this audit). A
   stuck run neither accrues N nor rules out a recurrence; its conclusion must be resolved before
   the ledger count is meaningful.
3. p̂ rests on 4 occurrences — one additional event moves pooled p̂ by ~25% relative.
4. AND-gate blind spot: a true regression shaped as "median(ref) inflated, most per-round ratios
   healthy" would pass the calibrated arm; detection rides the untouched SteadyCeiling/Budget
   arms (timing-statistic-decision.md Residual risk). No such form observed in-window.
5. Out-of-scope flake series exist (run 32687843472 a1: `TestBranchGuard_Latency` 1.82x,
   `TestReadCardStatus_DoesNotSearchBranchSet`) — recorded in M1, not fixed by this SPEC by
   design.

## 5. Recommended Close Path

**Sync-phase may proceed with AC-CFS-007 PENDING (documented).** Conditions:

1. progress.md / final sync record states the PENDING explicitly: window closes earliest
   2026-09-02T18:05:57Z AND N≥40 `go_code=true` runs with 0 recurrences of the 3 tests, ledger at
   reproduction-rate.md §5 (accrue via the committed sweep scripts).
2. The final AC-CFS-007 verdict (when the window completes) must carry the power arithmetic per
   acceptance.md §D.1 — a bare "0 recurrences" is insufficient by the SPEC's own rule.
3. A recurrence ≥1 blocks GREEN for that test's fix and reopens its mutant chain (acceptance.md
   §D.1 AC-CFS-007 GREEN condition, "재발 1건이면 GREEN 불가").
4. Watch item for the window: resolve run `32997835484`'s terminal state so N accounting starts
   from a decided denominator.

No blocker to sync-phase. The implementation diff is exactly scoped, all re-runnable judging
commands pass on the audited tree, and the sole open criterion is an observation window whose
clock started at merge.
