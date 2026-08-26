# Timing Statistic Decision — SPEC-CI-FLAKE-SERIES-001 (t278 M1, REQ-CFS-006)

Decided 2026-08-27 at run-phase tree `d1289c5db`. Verdict: **ADOPT the AND-gate** — the calibrated arm of `AssertPaired` shall fail only when BOTH the per-round paired-ratio median AND the ratio-of-medians exceed `MaxUnits`. The fallback (trimmed per-round median + more iterations) is REJECTED, on measured arithmetic below.

## (a) Option comparison

| Property | AND-gate (adopted) | Fallback: trimmed per-round + Iterations↑ (rejected) |
|---|---|---|
| Rule | fail iff `median(perRoundRatios) > MaxUnits` **AND** `median(fn)/median(ref) > MaxUnits` | drop top-k% of per-round ratios, raise Iterations |
| Alternation-locked asymmetry (post-#1591 observed form: per-round 2.47x, medians 1.09x) | **PASS** — medians healthy ⇒ no failure | trimmed median stays 2.47x ⇒ **still FAILS** (arithmetic below) |
| Offset load-step (pinned by `TestCalibratedRatioSurvivesOffsetLoadStep`: per-round 1.00x, medians 1.89x) | **PASS** — per-round healthy | PASS (per-round family is step-robust by construction) |
| Homogeneous true regression (4x every round, pinned by `TestPairedRatioStillCatchesRealCostGrowth`) | **FAIL** — both figures 4x ⇒ detection preserved | FAIL — detection preserved |
| Healthy code (local 10×0.99–1.01x; CI healthy ≈1.0x expected) | PASS with ~2x headroom under a 2.00x bound | PASS |
| blast radius | `CheckRatio` semantics extended for the paired path; `Assert`/`Check` (unpaired) untouched | touches sample-count economics of every paired caller (Iterations is a Bound field callers set) |

Implementation note for M2: `measurePaired` already returns `refSt` (`timing.go:332-356`), so both figures exist at the `report()` call site (`timing.go:215-218`); enforce the AND there (or in a new pure helper beside `CheckRatio`, `timing.go:289-308`) while keeping `CheckRatio`'s signature available — `paired_step_test.go:56` and `paired_step_test.go:79` call `CheckRatio`/`Check` directly and must keep compiling and passing unchanged (REQ-CFS-005).

## (b) Measured basis

1. **The only post-#1591 flare-up is alternation-locked asymmetry** (forensics.md §2.1, run 32779472351 a1): `n=20 median=1.502ms … refUnit=1.375ms ratio=2.47x` — ratio-of-medians = 1.502/1.375 = **1.09x (healthy)** while the enforced per-round median reads 2.47x. The defect being fixed is exactly "per-round alone over-trips on this form"; the AND-gate passes it.
2. **Local green distribution** (10 runs, darwin, `-race -count=10 -run TestAssertPairedHealthyEndToEnd`, tree `d1289c5db`, all PASS): ratios 0.99x–1.01x, median ≈ refUnit ≈ 503–515µs — on healthy unloaded code the two estimators coincide, so the AND conjunction costs nothing there. CI green figures are not mechanically extractable (channel gap — reproduction-rate.md §3).
3. **Pre-#1591 bilateral form** (run 32429213275 a1/a2, 8/20–21, both before #1591 merged 8/21 04:20): median=3.536ms, refUnit=1.3ms — both estimators ≈ 2.72x. That flake form (measuring tests overlapping windows) was real and was fixed by #1591; the AND-gate **correctly still fails it** — the gate does not re-open a fixed hole.
4. **Out-of-scope caller flake, same window** (run 32687843472 a1, 8/24, fast job): `TestBranchGuard_Latency` failed with median=3.431ms, refUnit=1.884ms — again bilateral 1.82x. AND-gate behavior unchanged for it; recorded here as evidence the calibrated arm's bilateral class keeps its detection across callers (see (c)).

Why ≥11 rounds must be inflated for the observed form: `medianFloat` returns `sorted[len/2]` (`timing.go:359-366`) = the 11th of 20 sorted values; a 2.47x median requires at least 11 of 20 rounds ≥ 2.47x while `median(fn)/median(ref)` stays 1.09x — only a systematic, alternation-locked asymmetry produces that (spec §2.2 candidate (i); candidate (ii), a CFS freeze of one in-flight sample, lands on random rounds and cannot hold 11 of 20 above bound while keeping both medians equal — no observation in this window matches (ii)).

## (c) Full AssertPaired caller census (command: `grep -rn "AssertPaired(" --include="*_test.go" internal/` — 3 call sites)

| Caller | Site | Bound | Per-caller impact of AND-gate |
|---|---|---|---|
| `TestAssertPairedHealthyEndToEnd` | `internal/timing/paired_test.go:61` | MaxUnits 2.0, It 20, W 2, ref/fn byte-identical `cpuUnit(2_000_000)` | The flake itself: observed 2.47x/1.09x form now passes; healthy 1.00x passes; no downside measured |
| `TestRecordEvent100Sequential` (RecordEvent append cycle) | `internal/harness/observer_test.go:251` | MaxUnits 2.0, It 100, W 5, ref mirrors full append mix | Positive/neutral: alternation-locked false positives suppressed; an added fsync/spawn regression inflates both estimators ⇒ still caught |
| `TestBranchGuard_Latency` (checkBranchState spawn cycle) | `internal/hook/pre_tool_branch_guard_integration_test.go:207` | MaxUnits 1.5, It 100, W 3, ref = same rev-parse spawn | Positive/neutral, same argument; its 8/24 observed failure was bilateral (1.82x) and remains detectable |

No non-test callers (`AssertPaired` is test-only by design; `timing` exports are consumed exclusively from `*_test.go`).

## (d) Rejection rationale (fallback)

- **Trimming cannot recover the observed form.** With 20 rounds and ≥11 at ≥2.47x: a 10% top-trim removes 2 rounds (→18 kept, median = 9th of 18 — still ≥2.47x while ≥9 inflated rounds remain); a 20% trim removes 4 (→16 kept, median = 8th — still ≥2.47x with ≥7 inflated above it); only trimming ≥45% (9 rounds, past the 11th) would move the median, and a 45%-trim also discards exactly the outlier rounds a true regression (4x-every-round distributions under noisy load) needs for detection — it buys immunity by blinding the arm. The observed form is *half the distribution inflated*, not a tail.
- **More iterations do not remove a systematic bias.** The asymmetry is phase-locked to the ref-first/fn-first alternation (`timing.go:342-348`); doubling Iterations samples more of the same phase pattern, converging the median *more tightly onto* the inflated value.
- **The conjunction is information-preserving in both directions**: each estimator is individually robust to one noise class (per-round: load-step; ratio-of-medians: alternation asymmetry) and the AND requires *both* to agree before declaring regression — the classes measured in this codebase (steps, alternation locks) each leave one estimator healthy, while every true-regression class examined (uniform k× cost) moves both.

## Residual risk (recorded, spec R2)

A true regression shaped as "median(ref) inflated but most per-round ratios healthy" — e.g. a load-step coinciding with a genuine cost increase in ref's window only — would pass the gate. No such form has been observed (this window: 0 occurrences), and it is indistinguishable in principle from a load step without more signal than a single paired run provides. Detection for it rides the SteadyCeiling/Budget arms (whole-distribution and worst-sample), which the AND-gate does not touch.
