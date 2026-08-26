# Series Analysis — SPEC-CI-FLAKE-SERIES-001 (t278 M4)

REQ-CFS-012 deliverable. Finalizes the common factor drafted in spec.md §3, with the three
cases as measured instances and the reusable authoring rule named. Fixes landed as PR #1666
(squash `379b310a6`, merged 2026-08-26T18:05:57Z): commits `def99739d` (#1), `5aedc1cd3` (#3),
`324883ebb` (#2).

## §1 The series

Three CI-only flakes, all observed on ubuntu 2-vCPU shared runners under full-suite parallel
load (local darwin green throughout — M1 §2.5), all failing in seconds-to-sub-second durations
unrelated to any timeout budget:

| # | Test | Package | Baseline occurrences (post-#1591) | Form |
|---|------|---------|-----------------------------------|------|
| 1 | `TestConcurrentSendPoll` | internal/sessionmsg | 2/166 runs | 97/100 loss, 0 errors, 0.07s |
| 2 | `TestAssertPairedHealthyEndToEnd` | internal/timing | 1/166 runs | per-round 2.47x vs ratio-of-medians 1.09x |
| 3 | `TestConfigChange_RT005ReloadIntegration` | internal/hook | 1/166 runs | single-sample 123.196ms vs ≤100ms |

Pooled baseline p̂ = 4/166 ≈ 0.0241 (reproduction-rate.md §4).

## §2 Common factor (final — spec §3 confirmed against M1–M3 evidence)

> **A test that decides a time/scheduling-dependent predicate from a single observation
> cannot separate scheduling nondeterminism on a contended 2-vCPU shared runner from the
> signal it claims to check.**

(Original, spec.md §3: 시간·스케줄링 의존 판정을 단일 관측으로 내리는 테스트가, 경합 중인
2-vCPU 공유 러너의 스케줄링 비결정성을 신호와 구분하지 못한다.)

The three cases instantiate the same defect shape at three different layers — liveness rule,
estimator, and assertion statistic:

| Case | The single observation | The window noise entered through | What the contract actually requires |
|------|------------------------|----------------------------------|-------------------------------------|
| #1 | "this poll returned 0/0" → treated as drain complete | observation time ↔ stop-confirmation time (TOCTOU) | **positive condition**: a 0/0 observed by a poll *started after all senders completed* |
| #2 | one per-round paired-ratio median | alternation phase locked across the whole measurement window | a load-robust ratio statistic (noise dichotomy: step ↔ alternating asymmetry) |
| #3 | wall-clock of one Handle invocation | start ↔ elapsed measurement span | p95 (multi-sample) ≤ 100ms |

Corroborating discriminators measured in M1 (not inferred): in #2 the two estimators split
(1.09x vs 2.47x) proving load-phase noise, while the pre-#1591 defect form was bilateral
(2.72x ≈ 2.72x) proving a real cost regression the same instrument caught; in #3 the failing
assertion was not the 2s-budget async wait but the µs-scale sync path measured at 123ms — a
scheduler preemption inside the span, not a code cost.

## §3 The reusable authoring rule

> **Do not decide a time- or scheduling-dependent predicate from a single observation.
> Wait for a positive condition, assert with the contract's own statistic, and prove the
> repair with a mutant.**

(원문: 시간·스케줄링 의존 판정은 단일 관측으로 내리지 않는다 — 긍정 조건·계약 통계·변이 증명.)

Applied per case (fix → M2 evidence, all at tree `324883ebb` final / RED-first per
progress.md §E.2):

1. **Positive condition over absence-of-work** — #1's stop rule now requires a re-poll after
   `sendersDone` closes; only that re-poll's 0/0 is terminal, and received-message totals are
   summed across both polls (`def99739d`). The reproduction test `TestPollerStopRule` forces
   the CI race window deterministically via a `beforeStopCheck` handshake gate; the mutant
   probe (old select-and-exit rule re-inserted) reproduces the exact 97/100 loss.
2. **The assertion's statistic equals the contract's statistic** — #3 rebuilt as a 20-sample
   nearest-rank p95 (second-largest of 20) with `WaitForAsync` drainage between samples
   (`5aedc1cd3`); a 150ms injected sleep makes it fail at 151.164ms over 20 samples. #2's
   calibrated arm now requires **both** per-round median and ratio-of-medians over bound
   (AND-gate, `CheckRatioAnd`, `324883ebb`): alternation-locked asymmetry (true ratio 1.00x)
   passes with the falsifier arm still proving the old single estimator would have tripped;
   homogeneous 3x growth still trips (detection power preserved).
3. **Mutant proof** — every fix carried a mutant probe that re-observed the RED form before
   GREEN was re-confirmed (progress.md §E.2 E8 blocks). The property tests pair a detection
   arm and a survival arm so the rule, not the sample, is what's tested.

Applicability bound: the rule addresses *authoring* of timing/scheduling-sensitive tests. It
does not cover CI-infrastructure flake classes (runner-network, cache poisoning, tool-chain
drift), which exhibit different signatures (multi-minute timeouts, error text rather than
statistical drift).

## §4 Verification state

- Local: full §E battery green at final tree `324883ebb` (progress.md §E.2 — sessionmsg
  -count=20, timing -count=5, hook full + TestConfigChange -count=5, cross-platform build,
  lint 0/0, vet).
- CI: PR #1666 all checks green including `Integration Tests (windows-latest)` 13m41s;
  non-required `graph-freshness` fail attributed to an orphan codemaps stamp unrelated to
  this diff (main HEAD green same hour).
- Reproduction-rate verdict: **PENDING** — post-merge window opened 2026-08-26T18:05:57Z;
  REQ-CFS-010 requires N ≥ 40 `go_code=true` runs AND ≥ 7 days. Accrues in
  reproduction-rate.md §5; the M4 verdict must carry the power arithmetic (§4), not a bare
  "0 recurrences".

## §5 Cross-references

- forensics.md — verbatim failure logs, pre/post-#1591 classification, channel caveats
- reproduction-rate.md — baseline table, p̂ arithmetic, post-merge ledger, power numbers
- timing-statistic-decision.md — the M1 estimator decision (AND-gate) this analysis finalizes
- progress.md §E.2 — M1/M2 evidence chain (RED/GREEN/mutant per fix)
