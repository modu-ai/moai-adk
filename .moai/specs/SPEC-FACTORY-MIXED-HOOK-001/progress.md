---
id: SPEC-FACTORY-MIXED-HOOK-001
document: progress
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Progress — SPEC-FACTORY-MIXED-HOOK-001

## §A Status

- Current status: `in-progress`.
- Card: `t1074` (`picked`).
- Worktree: `WT-factory-mixed-hook`.
- Plan baseline: `758314007` (local develop matched when authoring began).
- Implementation baseline: `8d8e101bf`; M1 implementation commit: `cb099897a`.
- Next gate: finish M2/M3/M4 without absorbing the t1082 worktree lifecycle, then independent sync audit.

## §B Plan artifacts

| Artifact | Status |
|---|---|
| spec.md | authored |
| plan.md | authored |
| acceptance.md | authored |
| research.md | authored |
| design.md | authored |
| progress.md | authored |

## §C Plan audit

### Iteration 1

- Verdict: FAIL, merge-blocking.
- Findings: non-canonical GEARS clauses, forbidden sibling status fields, abbreviated REQ references, missing executable RED ledger, live `NOT_RUN` false-pass, and missing exclusion structure.
- Resolution: all findings corrected; strict lint and mutant gates added.

### Iteration 2 — final

- Independent auditor verdict: `PASS`.
- Score: `0.94 / 1.00` (Tier L threshold `0.85`).
- Findings: empty; merge-blocking findings: none.
- Auditor-observed evidence:
  - `go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json` → `[]`, exit 0 (with isolated `GOCACHE`).
  - `git diff --check` → stdout empty, exit 0.
  - Fifteen RED-now named-test presence probes → stdout empty, exit 1 at tree `758314007d8c696ff1af377dc8cdc46d76368314`.
  - Synthetic live events: positive passes, `skip` and `NOT_RUN` fail.

## §D Run-phase evidence

- M1 committed as `cb099897a` (`feat(t1074): M1 bind canonical factory runs and peers`).
- Observed scoped tests at that commit: `internal/cli` → `ok ... 4.491s`; `internal/factorymsg` → `ok ... 1.310s`.
- M2/M3 committed as `6bde8412c` (`feat(t1074): deliver durable mixed factory messaging`).
- AC-FMH-001 through AC-FMH-009 exact JSON/JQ gates passed; logs are `../../reports/t1074/ac01.jsonl` through `ac09.jsonl`.
- M4 live harnesses executed. AC-FMH-010 through AC-FMH-012 failed because the local process-start probe was `indeterminate`; AC-FMH-013 and AC-FMH-014 failed because the Claude subscription OAuth token was revoked. These rows remain FAIL, not skipped or passed.
- AC-FMH-015 executed the full hook matrix. On the final real-worktree run the fixed empty-inbox p95 target failed (`370.799917ms`), and every matrix cell exceeded the 200ms inspection target. The threshold was not raised.

## §D.1 Codex cwd redesign decision

- Latest official source/docs and a local Codex `0.155.1` probe establish that interactive `/cd` can keep the visible conversation flow while rotating the physical session/thread UUID.
- `t1074` therefore owns the stable logical lane/current-endpoint broker seam only.
- `t1082` owns card worktree creation and `/cd` or headless `cwd` handoff through `BOUND`.
- `t1075` depends on t1082 and may wake only the current bound endpoint.

## §E Audit-ready signals

- Plan-phase: audit-ready; independent `PASS` at score `0.94`.
- Run-phase: M1/M2/M3 implementation complete; M4 harness implemented and executed with FAIL evidence. Independent sync audit remains pending.

## §E.2 Run-phase evidence

| Criterion / invariant | Actual output | Status |
|---|---|---|
| AC-FMH-001..009 | Exact named Go JSON gates each produced final `true`; `ac01.jsonl`..`ac09.jsonl`. | PASS |
| AC-FMH-010 | `Codex owner fingerprint unavailable: state=indeterminate`. | FAIL |
| AC-FMH-011 | `Codex owner fingerprint unavailable: state=indeterminate`. | FAIL |
| AC-FMH-012 | `Codex owner fingerprint unavailable: state=indeterminate`. | FAIL |
| AC-FMH-013 | Claude CLI reached subscription transport, then returned HTTP 401 revoked OAuth. | FAIL |
| AC-FMH-014 | Pending-until-next-turn state was created; live Claude boundary turn returned HTTP 401 revoked OAuth before receipt. | FAIL |
| AC-FMH-015 | Matrix executed for 0/16/1000 queues and 1/10 sessions; empty p95 `370.799917ms` exceeded 50ms and every real-worktree cell exceeded 200ms. | FAIL |
| Canonical/legacy isolation | Actual linked-worktree fixture converged; foreign run/project list/claim/read/receipt failed; legacy sentinel bytes unchanged. | PASS |
| Template parity | Project and embedded Stop hook entries are synchronous and parity test passed. | PASS |
| Repository-wide verdict | Not run; integration-branch CI owns this verdict. A broad local package attempt hit pre-existing live/environment-sensitive failures and `internal/hook` timeout. | PENDING |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: null
run_commit_sha: pending-m4-evidence-commit
run_status: fail
ac_pass_count: 9
ac_fail_count: 6
preserve_list_post_run_count: 1
l44_pre_commit_fetch: not-run
l44_post_push_fetch: not-applicable-no-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-run
  reason: "M4 live blockers and benchmark regression leave the run incomplete"
total_run_phase_files: 25
m1_to_mN_commit_strategy: "M1 cb099897a; M2/M3 6bde8412c; M4 harness/evidence pending"
```
