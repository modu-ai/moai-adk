# Plan-Audit iter-1 — SPEC-CI-FLAKE-SERIES-001 (card t278)

- Auditor: plan-auditor (independent, fresh context)
- Date: 2026-08-26
- Tree: `.claude/worktrees/t278` @ `175d63f3f` (branch `WT-ci-flake-series`)
- Audit mode: single-backend (claude-only) — `audit_model` not configured in `.moai/config/` (grep 0 hits), so no `audit_multi` fan-out per the cross-model-audit skill's entry condition.
- Verdict: **PASS-WITH-DEBT** — overall 0.86 (Tier M threshold 0.80), 0 BLOCKING / 2 SHOULD-FIX / 3 MINOR.

---

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk (audit-level)

**Claim**: The SPEC's evidence chain (3 flakes, 2 premise overturns) and its completion plan are sound and executable.

**Evidence** (all commands run in this audit, this tree):

| # | What was checked | Command | Observed |
|---|---|---|---|
| V1 | Flake #1 CI evidence | `gh run view 32774108273 --attempt 1 --log-failed` | `store_test.go:385: received 97 messages, want 100 (loss or duplication)` (0.07s) — byte-identical to spec.md §1.1 |
| V2 | Flake #1 race-job recurrence | `gh run view 32777242100 --log-failed` | `--- FAIL: TestConcurrentSendPoll (0.27s)` Race Test — matches 관측 2 |
| V3 | Flake #2 evidence | `gh run view 32779472351 --attempt 1 --log-failed` | `timing.go:233: paired-cpu-1x: n=20 median=1.502ms ... refUnit=1.375ms ratio=2.47x (maxUnits=2.00x...)` — byte-identical; 1.502/1.375 = 1.092x confirms the discriminator arithmetic |
| V4 | Flake #3 evidence | `gh run view 32815411885 --attempt 1 --log-failed` | `config_change_test.go:51: synchronous return took 123.195919ms, want ≤ 100ms (REQ-HAE-002)` — byte-identical |
| V5 | Flake #1 TOCTOU code | Read `store_test.go:352-371` | Exit rule at 361-370: 0/0 poll → `select` on `sendersDone` (later time) → immediate exit. Deadline at :351, assertion at :385 — all cited lines exact |
| V6 | Store exclusion | Read `store.go:262-264, 309-412`, `lock.go:69-114` | Send/Poll both under `withAgentLock` on the same mailbox (in-process mutex + flock) — serialized; claim order writeJSONAtomic(claimed) → removeIfExists(pending), no absence window; TTL sweep 24h/10m (`defaults.go:401-404`) cannot fire in 0.07s |
| V7 | Flake #3 category error | Read `config_change_test.go:24-63`, `config_change.go:54-82` | RT005 has `t.Parallel()` (:25); assertion :50-52 is single-sample wall clock; Handle sync path = slog.Info + WithTimeout + `go func()` (µs-scale); godoc + `BenchmarkConfigChange_AsyncReturn` (:271-313, p95 check :308-311) confirm the contract is p95 (AC-HAE-003); owner SPEC `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` exists with REQ-HAE-002 |
| V8 | publish() overturn | `git log -S "func publish(" -- internal/timing/timing.go` | Introduced by `7a531fe86` "test(timing): interleave the calibration reference with the measured runs (t162)" (= PR #1591); `publish()` at timing.go:254-265 called unconditionally in `report()` (:234); `publish_test.go` exists. Card premise "ratio only on failure" is genuinely overturned |
| V9 | Flake #2 structure | Read `paired_test.go`, `paired_step_test.go`, `timing_test.go` | ref/fn byte-identical `cpuUnit(2_000_000)` (:67-68); alternation at timing.go:342-348; load-step pin + falsifier arm (CheckRatio direct call :56, comment :61-63); historical 2.32x/2.72x/4.64x + 1.82x @ run 32687843472 a1 recorded in test comments — all as cited |
| V10 | attempt-1 hiding | `gh run list --workflow ci.yml -L 30 --json databaseId,conclusion` | 0 `failure` conclusions among the last 30 runs (08-24→08-26) while ≥4 attempt-1 flare-ups exist in that window (32779472351 and 32815411885 now read `success`). Stronger than the SPEC's own 3-vs-4 measurement |
| V11 | AC-CFS-007 executability | same run list | ~30 runs / 2.5 days ≈ 12 runs/day cadence → ≥40 go_code=true observations over ≥7 days is concretely achievable (job-count reading: ×2) |
| V12 | CI job structure | Read `.github/workflows/ci.yml:85-240` | `test` job: no -race, merge gate; `test-race`: advisory; both gated on `needs.detect.outputs.go_code == 'true'` (docs-only PRs excluded from denominator — correctly handled by AC wording "go_code=true") |
| V13 | Line-reference spot checks | Reads across 6 files | ~15 cited `file:line` references checked; all accurate (incl. `paired_step_test.go:56` CheckRatio, `timing_test.go:14` HARD no-parallel rule, `config_change.go:56-84`) |
| V14 | AssertPaired blast radius | `grep -rn "AssertPaired(" --include="*_test.go" internal/` | 3 call sites: paired_test.go:61, internal/harness/observer_test.go:251, internal/hook/pre_tool_branch_guard_integration_test.go:207 (TestBranchGuard_Latency) — R1 concern real, M1 item 5 covers enumeration |
| V15 | New-file collisions | `ls internal/sessionmsg/*_test.go internal/timing/*_test.go` | `stoprule_test.go` / `paired_asym_test.go` do not exist — 신규 as planned; `grep -c "publish(" timing.go` = 2, exactly matching §C expectation |

**Baseline-attribution**: all observations above were produced in this audit run, against this worktree (`175d63f3f`) and the live `gh` view of `modu-ai/moai-adk` ci.yml runs on 2026-08-26.

**Gaps** (explicitly NOT observed by this auditor):
- Local `-race -count=20 ./internal/sessionmsg/` green (AC-CFS-002 RED-now cell) was NOT re-run — machine-load discipline; §C pre-flight re-measures it at run-phase entry.
- Flake #2's mechanism (i) vs (ii) remains unproven — the SPEC itself defers the statistic decision to M1 data (REQ-CFS-006); this auditor verified the *discriminator arithmetic* and the *structural facts*, not the physical cause.
- No `go test` of any kind was executed in this audit (read-only + `gh` evidence only).
- `moai spec audit` was not run (completion-time target per spec.md §7, not a plan-phase gate).

**Residual-risk**:
- Store-exclusion is code-reading + log-elimination based. A store-side silent-loss path no one has read yet cannot be *fully* excluded until REQ-CFS-002's deterministic repro exists; §G already routes a premise reversal to scope revision.
- The AND-gate's blind spot (a true regression that inflates fn in only half the rounds) is recorded as R2 residual — acceptable and honestly stated.
- Post-card-cycle observation-window ownership rests on "무인 sweep(스크립트)" (§G) — the script shape is specified but its operator after the lane session ends is implicit (lead/M4-resumer). Not a defect; noted.

---

## Per-dimension scores

| Dimension | Score | Basis |
|---|---|---|
| Clarity | 0.88 | Measured evidence with run IDs; ~15 line refs all exact; RED/GREEN cells explicit. Docked for the undefined observation unit N (D2). |
| Completeness | 0.85 | Two-cell adoption on all testable ACs; mutant probes for all three fixes (rule-revert / cpuUnit(6M) / sleep-150ms); estimator falsifier arm; PENDING window handling. Docked for D1 (unmapped REQ) and D5. |
| Testability | 0.90 | Every AC carries a 판정 명령; AC-CFS-007's sweep script is concrete and its ≥40/≥7-day target verified achievable against measured cadence (V11); RED-now values are actually-observed, not asserted (V1-V4). |
| Traceability | 0.80 | REQ→AC matrix + evidence index §E (independently re-verified); tier ceilings respected (12 REQ ≤ 16, 9 AC ≤ 16). Docked for D1 (REQ-CFS-008 unmapped) and D3 (dangling related_spec). |
| **Overall** | **0.86** | Simple mean; ≥ 0.80 Tier M PASS threshold. |

## Defect list

| ID | Severity | Dimension | Defect | Minimal revision |
|---|---|---|---|---|
| D1 | SHOULD-FIX | Traceability | **REQ-CFS-008 has no AC.** The AC matrix rows map REQ-001..007, 009..012 (+ scope axis); REQ-CFS-008 (no `t.Parallel()` on cpuUnit measuring tests) is unmapped and carries no observable judgment command — under the SPEC's own verification-completeness axis an unmeasurable REQ is an unfinished REQ. | Add an AC row (e.g. AC-CFS-010) with a grep-based 판정: `grep -L "t.Parallel" internal/timing/paired_test.go internal/timing/paired_step_test.go ...` → all listed; or fold the check into AC-CFS-003's 판정 명령 and note the mapping in the 요구사항 column. |
| D2 | SHOULD-FIX | Clarity / Testability | **Observation unit N is undefined.** REQ-CFS-010 / AC-CFS-007 say "≥40 affected-job 관측" — ambiguous between 40 workflow-runs and 40 job-instances (2 jobs per run). The `(1-p̂)^N` confidence arithmetic (M4.2, REQ-CFS-010) requires a pinned N; the two readings differ 2× in claimed confidence. | Define once: recommend N = count of go_code=true workflow-run observations (both jobs enumerated per run), and use the same N in REQ-CFS-010, AC-CFS-007(b), and the M4 confidence arithmetic. |
| D3 | MINOR | Traceability | `related_specs` entry `SPEC-V3R6-CI-PR-SPEEDUP-001` never existed as a SPEC directory in git (`git log --all --diff-filter=A` on the path is empty); it survives only as a ci.yml comment citation. A reader following the frontmatter reference finds nothing. | Drop the entry or annotate it (e.g. "(ci.yml comment provenance only, no SPEC doc)"). Keep SPEC-CI-FLAKY-STABILIZE-001 (exists, verified). |
| D4 | MINOR | Clarity | spec.md §2.2: "20개 라운드 비 중 ≥10개가 2.47 이상어야 하고" — for the median of 20 sorted values to equal 2.47 you need ≥11 values above it (or exactly 10 with the 10th at the boundary). The load-bearing conclusion (systematic half-round skew; random noise cannot produce it) is unaffected — this auditor re-derived it. | Reword to "≥11개 (또는 경계값 포함 ≥10개)". |
| D5 | MINOR | Completeness | plan.md §B.3: "핸들러 매회 신규 생성 — wg 추적 그대로" under-specifies how all 20 handlers' wait groups are awaited — the current test awaits a single `h.waitGroup()`; 20 handlers need aggregation (wait each, or a shared wg). | One clause: "20개 핸들러의 wg를 각각(또는 공유 wg로) WaitForAsync" — run-phase resolvable as-is, pinned for clarity. |

## Why the two premise overturns are SOUND (audit question (a))

- **Overturn #1 (store → test-side TOCTOU)**: independently re-verified — (i) both `Send` and `Poll` serialize on the same per-mailbox lock (`lock.go:69-114`), so a claim/Remaining bookkeeping race between them is structurally impossible; (ii) the claim loop writes claimed before deleting pending (`store.go:405-408`) — no absence window; (iii) the failing logs contain zero `poll:`/`send:` errors while the test's error paths always log; (iv) the 361-370 select-after-observe window explains 97/100 + 0 errors + 0.07s exactly. The chain does not leap.
- **Overturn #2 (ratio observable on green)**: `publish()` lands in `7a531fe86` (t162 / PR #1591), is called unconditionally from `report()` on every run, and has its own test — the card's "failure-only exposure" premise is factually dead on this tree.

## Audit questions (b)(c)(d)

- **(b) AC-CFS-007 executable?** Yes — measured cadence ~12 ci.yml runs/day (V11) makes ≥40 go_code observations + ≥7 days concrete; WHO/WHEN is assigned (M3 lane: daily sweep + post-merge; M4 PENDING fallback forbids false completion; §G unmanned sweep). Only soft spot: post-lane ownership (residual, not a defect). D2's N-unit must be pinned for the confidence number to mean anything.
- **(c) Milestones investigation-first + RED-first?** Yes — M1 produces 0 code modifications (forensics, baseline sweep, green-ratio collection, statistic decision, caller census); M2 orders #1→#3→#2 with each fix preceded by an observed RED (deterministic repro 97/100; 150ms mutant; falsifier-arm property test on the current estimator).
- **(d) Scope creep?** None found — §4 안/밖 is explicit, AC-CFS-009 enforces it by diff, and the out-of-scope list protects store.go / CI workflow / benchmark / other timing arms / other callers' test bodies. REQ-CFS-008 codifies an existing HARD rule (timing_test.go:14) with a compliance check — in-family for the series, not creep.

## Minimal revisions to clear PASS-WITH-DEBT → PASS

1. D1: add the REQ-CFS-008 AC row (or fold into AC-CFS-003's command) — a 2-line edit.
2. D2: pin the observation unit N in REQ-CFS-010 + AC-CFS-007(b) + M4.2 — a 3-place wording edit.
3. (Optional, recommended) D3 frontmatter cleanup; D4/D5 one-clause rewording.

All are documentation-level; none block Implementation Kickoff Approval mechanics, but D1/D2 should land before run-phase entry so the completion verdict's arithmetic is unambiguous.
