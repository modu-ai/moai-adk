auditor-model: glm-5.3-flash

# Cold Sync-Audit Verdict — card t1630 (SPEC-USER-SETTINGS-PROTECT-001)

verdict: FAIL
audited_sha: 6f7ddb13078b5432270df03966058caa5afb02ed

Cold fallback binding auditor after the sync-audit-4dim workflow FAIL (round 0). This document records re-entry #2, the FINAL re-audit per operator decision d-20261014T055241Z-1bc6 option (a) ("카드 내 잠금 + 재감사" — in-card lock + one more sync-audit), at HEAD 6f7ddb1307 (the merge absorbing local main 842160157f; card work tip 9b9fc68069 F2' lock + a28cb7dc2e evidence). History: round 0 FAIL at 25df1da08c (rcpt-ce0152d51eac3787c2b2a0d7), re-entry #1 FAIL at 486fc3747d (rcpt-9a7cb95694999d93a622b157).

## Re-entry #2 verdict: FAIL

F1 is **CLOSED** by the recorded leader ruling, F8 is **FIXED**, and the F2'/F2/F3 primary scenarios are **verified by this auditor's own runs** (below). The verdict stays FAIL because the required codex gate re-ran at 6f7ddb1307 and returned **fail** with four NEW P1 findings (F9–F12) on the cross-process surface the F2' repair opened — two of which this auditor's independent code read corroborates in mechanism (the takeover/release ownership gap was flagged by this auditor before the fold; the USER-span and pre-lock body-read positions are visible in the code). Fold rule: any required FAIL → overall fail (case 2). The iteration ceiling is reached: F9–F12 route to the leader for disposition.

Weighted dimension scores (flat, for the trail; the verdict is fold-driven): Functionality 88, Security 85, Craft 85, Consistency 90 → 87.0/100. The required-gate fail forces overall FAIL regardless of the numeric score.

## Convergence digest — round 2 (for the orchestrator's check)

- overall_verdict: **fail**
- gate_unmet: absent from the result JSON; `residual_risk_note: "required-backend FAIL: codex"`
- plan_source: **config**
- per_backend_verdicts: codex / gate **required** / verdict **fail** (sole participant; participant_count 1, disagreement_flag null, fail_open_backends [])
- audit_receipt (round 2): **rcpt-2f0145b43a32bd96b19103d7** (round 1: rcpt-9a7cb95694999d93a622b157; round 0: rcpt-ce0152d51eac3787c2b2a0d7)
- build_lag note: the installed moai binary is built from ec193e4777 (ancestor); no moai-CLI tree measurement is cited as evidence in this verdict.
- Codex's own gate observations at the same HEAD: the package `-race` tests pass; its four overlay reproductions are reported in codex's summary ("재현이 실패했다" — the repro-success semantics of that phrase are ambiguous in codex's text); `moai verify codex-review` = **inconclusive** (`codex stdout closed before response to id=1`, same child-EOF shape as round 1) — named non-evidence.
- This auditor is read-only: the orchestrator writes `.moai/state/audit-plan-result.json` from these digest members and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`. **This verdict is not final until that check passes.**

## Auditor-model observation (process — leader disposition, not a code finding)

Operator decision d-20261014T055241Z-1bc6 (a) and the a52a/role-table chain (d-20261014T081435Z-64ee) expect this re-audit on opus/high. The model actually serving this audit is **glm-5.3-flash** (recorded verbatim in line 1); the d-64ee ruling itself marks GLM-lane model reach unmeasured (t1728 step 3). Recorded so the leader can disposition the model axis. It changes no outcome this round: the verdict is FAIL on the fold either way.

## Per-finding status (final, F1–F12)

- **F1 [P1] [CLOSED — verified this round]** full-bundle path vs Q1. Ruling **d-20261014T052723Z-3eef** verified on the board via `moai decision read --scope card:t1630` (scope=card:t1630, kind=ruling, decided_by factory-leader 14399b4f): "closed by routing, not by an in-card fix: the full-bundle path belongs to card t1666, and this SPEC records its non-normative scoping of REQ-003 (consistent with rulings 10fe/56ff/6892 and the AC-003 B self-measurement contract)". Closes F1 per its own terms.
- **F2 [P1] [FIXED — held at this HEAD]** in-process record inversion; `recordProtocolMu` + inter-process lock around both writer spans. Verified re-entry #1 and re-run this round (Claim 1).
- **F3 [P2] [FIXED — held at this HEAD]** failed-write rollback; verified this round (Claim 1).
- **F2' [P1] [PARTIALLY FIXED — primary scenario closed, surface still open]** The primary scenario — two live concurrent healthy renders of one project — is closed: the lock arbitrates it correctly (mutex-first ordering, O_EXCL create, bounded retry, pid probe with EPERM-treated-alive, windows age-bound; verified by this auditor's two-process `-race` run + code read). It stays open on four NEW P1 defects the fold found on the same surface:
  - **F9 [P1] [blocking-grade, leader disposition]** `tier_render.go:236-244` — the age branch of `takeoverStaleRecordLock` bypasses the pid probe entirely: a lockfile older than 30s is removed **even when its holder is alive and mid-render** (a stalled/slow render loses mutual exclusion). Codex reproduces a child acquiring while a live parent holds. Recommendation: an OS-level lock that releases on process exit (flock) instead of age-only takeover.
  - **F10 [P1] [blocking-grade, leader disposition]** `tier_render.go:231-244` (takeover Stat→Remove is not atomic) + `:203-210` (release removes by path with no ownership check) — a late remover can delete a subsequent owner's fresh lockfile, leaving two processes both believing they hold the lock. This auditor's pre-fold code read flagged the same TOCTOU class; codex grades it P1.
  - **F11 [P1] [blocking-grade, leader disposition]** `codegen.go:~215` before `:246` — BuildInto reads/splices the existing settings body (`extractPermissions(body)`) BEFORE acquiring the protocol lock; a concurrent write between read and lock is clobbered by a stale-based output (codex: a `plan` write reverted to stale `bypassPermissions`). Recommendation: move the read inside the span.
  - **F12 [P1] [blocking-grade, leader disposition]** USER-scope writes carry no cross-process serialization at all: `RenderTierPermissions` renders the USER file at `tier_render.go:78`, before the lock is taken at `:99`, and the `:471` existing-mode-preservation path has no lock either (codex: A's absence-check, B's `plan`, A's `auto` overwrite). Recommendation: serialize the USER read-check-write span.
- **F4 [minor] [optional — unchanged]** `WriteUserDefaultMode` fail-open (untouched).
- **F5 [minor] [optional — unchanged]** stale defaultMode comments in profile_setup files.
- **F6 [minor] [optional — unchanged]** `mergeProjectAllow` inline union duplication.
- **F7 [minor] [optional — unchanged]** `ac001ReadPerms` helper name.
- **F8 [minor] [FIXED — verified this round]** `managedAllowRecord` doc comment (`tier_render.go:374-383`) now states: "Across a failed settings write the record is rolled back to its pre-render content (rollbackManagedAllowRecord) — a failed render leaves no trace." Matches the F3 semantics.

## Verification evidence — re-entry #2 (five-section format)

**Claim 1 — the F2'/F2/F3 primary scenarios hold under the race detector at the new HEAD.**
- Evidence (this run, verbatim): `go test -race -count=1 -v -run '^(TestInterProcessRecordLock_TwoProcessRepro|TestConcurrentRendersKeepRecordConsistentWithSettings|TestFailedSettingsWriteRollsRecordBack_UserAddedEntrySurvivesRevocation)$' ./internal/config/toolpolicy/` → `--- PASS: TestInterProcessRecordLock_TwoProcessRepro (13.65s)` with all three subtests RUN (`blocks-until-child-releases-and-converges`, `takes-over-a-dead-pid-lockfile`, `errors-after-budget-on-a-live-holder`), `--- PASS: TestConcurrentRendersKeepRecordConsistentWithSettings (5.95s)`, `--- PASS: TestFailedSettingsWriteRollsRecordBack_UserAddedEntrySurvivesRevocation (0.02s)`, `PASS`, `ok github.com/modu-ai/moai-adk/internal/config/toolpolicy 20.928s`, DECIDING-EXIT=0, no `WARNING: DATA RACE`. The two-process subtest is a genuine second OS process (`exec.Command(os.Args[0])` reexec child holding the same production lockfile path) — this satisfies the two-process dimension with this auditor's own probe.
- Baseline-attribution: this run, HEAD 6f7ddb13078b5432270df03966058caa5afb02ed, worktree `.moai/worktrees/t1630`, env-scrubbed invocation.
- Gaps: the F9/F10 takeover-corner interleavings are NOT covered by these tests; the RED run at 486fc3747d recorded in progress.md cannot be re-witnessed from the commit graph (test and fix share commit 9b9fc68069 — §2.3 ordering attribution; recorded as a Gap, GREEN re-observed here).
- Residual-risk: F9–F12 (codex-reproduced, mechanism-corroborated by this auditor's read).

**Claim 2 — Craft gates hold on the repaired + absorbed tree.**
- Evidence: full `-race -count=1` suites → `ok internal/config/toolpolicy 18.781s`, `ok internal/core/project 12.728s` (the second package moved from `internal/cli/project` to `internal/core/project` in the absorbed main — this auditor's first probe at the old path failed with `directory not found` and was re-run at the new path); `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test -count=1 -cover` → toolpolicy `coverage: 88.0% of statements`, core/project `coverage: 89.0% of statements` (≥ 85% threshold; the dip from 90.3% is the lock's uncovered error branches); `gofmt -l` on both packages → empty; `go vet` clean; `golangci-lint run` on both packages → `0 issues.`
- Baseline-attribution: this run, HEAD 6f7ddb1307.
- Gaps: AC selectors not individually re-run (delta-scoped; the full package suites ran twice, once under `-race`).
- Residual-risk: none observed for Craft beyond the F9/F10 ownership gaps.

**Claim 3 — F1 ruling recorded on the board.**
- Evidence: `moai decision read --scope card:t1630` → `board=ok unparseable=0 records=83`, record `id=d-20261014T052723Z-3eef scope=card:t1630 kind=ruling` with the routing body quoted above.
- Baseline-attribution: this run, this tree's board state.
- Gaps: none for this claim.
- Residual-risk: none.

**Claim 4 — F8 comment fixed.**
- Evidence: read of `internal/config/toolpolicy/tier_render.go:374-383` — the rollback sentence quoted in the F8 row above.
- Baseline-attribution: this run, HEAD 6f7ddb1307.
- Gaps: none.
- Residual-risk: none.

## Carried items (Gaps, not failures — leader-owned)

- **AC-013 gate**: NOT re-measured this round (delta scope; classified merge-ready HOLD, leader-owned, pending t1619 + t1591 landing — do not fail on it per the re-entry #2 dispatch).
- init byte-diff ≠ 0 (`ensureGlobalSettingsEnv`) → t1703; `TestRuleDateProvenance` date leak → t1704; staging-window persistence-direction residual → t1703.
- §E.2 named residuals (bounded-wait + pid-reuse misjudged-takeover class) — now joined by the F9–F12 findings, which are concrete instances and extensions of that class.

## Iteration history

1. Plan-audit iter1 FAIL 0.69 → revision 2 split (d-20261010T091713Z-10fe) → iter2 FAIL → B1-B3 repair fb672d172 → iter3 PASS 0.81 conditional → operator ratification d-20261010T113615Z-6892.
2. Run phase: M1 057db17c41, M2+M3 e581bdbb20; P-3 RED re-observation; E1-E6 green; AC-013 divergence measured and contained.
3. Card-review rounds 1-3, closed at the 2/2 ceiling; sync close cb4b0b6233 + D3 backfill 25df1da08c.
4. sync-audit-4dim: FAIL harmonic 0.8364 (Consistency 0.75, contested major).
5. Round 0 cold audit (25df1da08c): FAIL, folded from codex (rcpt-ce0152d51eac3787c2b2a0d7); F1-F3 blocking, F4-F7 optional.
6. Re-entry #1 (486fc3747d): FAIL — F2/F3 verified fixed; F1 ruling pending; F2' new/open (rcpt-9a7cb95694999d93a622b157).
7. **Re-entry #2 (6f7ddb1307): FAIL — F1 CLOSED (ruling 3eef verified), F8 FIXED, F2'/F2/F3 primary path verified by own two-process `-race` run; required codex gate FAIL with four new P1s (F9–F12) on the cross-process surface; receipt rcpt-2f0145b43a32bd96b19103d7. Iteration ceiling reached (final re-audit per operator decision (a)) — F9–F12 route to the leader for disposition.**

## Disposition summary for the leader

- F9–F12 are blocking-grade P1s on the F2' mechanism and its adjacent unserialized spans. The concrete fix directions codex named: OS-level process-exit-released locking (flock) for takeover/release ownership binding (F9/F10), move BuildInto's body read inside the span (F11), serialize the USER-file read-check-write span (F12). Scope decision (in-card vs carry-over card) is the leader's — the operator's decision (a) authorized one repair + one re-audit, both now consumed.
- The auditor-model axis (glm-5.3-flash vs the opus/high expectation) is recorded above for the leader's disposition; it does not affect this FAIL.
