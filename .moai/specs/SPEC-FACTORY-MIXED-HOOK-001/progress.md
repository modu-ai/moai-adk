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
- Next gate: rerun AC-FMH-011 through AC-FMH-014 after Claude subscription capacity recovers, then independent sync audit.

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
- M4 repair added a native Darwin process-start probe, one-pass linked-worktree canonicalization, explicit Codex MCP factory env forwarding, Codex launcher attribution scrubbing/backend export, tmux factory-env propagation, and distinct live owner identities.
- AC-FMH-010 now passes with separate real Codex model contexts, unique owner PIDs, no injected Codex session UUID, and two explicit acknowledgements. `ac10.jsonl` records `acknowledged=2` and the exact JQ gate printed `true`.
- AC-FMH-011 was rerun after the Codex repair: the Codex lead send succeeded, then the Claude worker failed with `OAuth session expired and could not be refreshed`. The project-managed Claude profile was separately probed and authenticated far enough to return HTTP 429 weekly-limit exhaustion. AC-FMH-012 through AC-FMH-014 were not rerun because they require the same unavailable Claude subscription turn; all four rows remain FAIL, never skipped or passed.
- AC-FMH-015 now passes without changing its thresholds. The final matrix measured empty full-hook p50 `31.216917ms`, p95 `33.143167ms`, every inspection cell below 200ms, and bounded contention at `57.022917ms` with truthful `SQLITE_BUSY` degradation. The exact JQ gate printed `true`.

## §D.1 Codex cwd redesign decision

- Latest official source/docs and a local Codex `0.155.1` probe establish that interactive `/cd` can keep the visible conversation flow while rotating the physical session/thread UUID.
- `t1074` therefore owns the stable logical lane/current-endpoint broker seam only.
- `t1082` owns card worktree creation and `/cd` or headless `cwd` handoff through `BOUND`.
- `t1075` depends on t1082 and may wake only the current bound endpoint.

## §E Audit-ready signals

- Plan-phase: audit-ready; independent `PASS` at score `0.94`.
- Run-phase: M1/M2/M3 complete; M4 Codex and performance rows pass, while four Claude-dependent live rows remain blocked by current subscription capacity. Independent sync audit remains pending.

## §E.2 Run-phase evidence

| Criterion / invariant | Actual output | Status |
|---|---|---|
| AC-FMH-001..009 | Exact named Go JSON gates each produced final `true`; `ac01.jsonl`..`ac09.jsonl`. | PASS |
| AC-FMH-010 | Exact live gate printed `true`; separate Codex contexts completed nonce round trip with `acknowledged=2`. | PASS |
| AC-FMH-011 | Codex lead send returned `ok`; Claude worker returned `OAuth session expired and could not be refreshed`. | FAIL |
| AC-FMH-012 | Not rerun after repair because the required Claude turn is unavailable under the same exhausted subscription. | FAIL |
| AC-FMH-013 | Not rerun after repair because the required Claude turns are unavailable under the same exhausted subscription. | FAIL |
| AC-FMH-014 | Not rerun after repair because the required Claude turn is unavailable under the same exhausted subscription. | FAIL |
| AC-FMH-015 | Exact benchmark gate printed `true`; empty p95 `33.143167ms`, contention `57.022917ms`, all inspection cells below 200ms. | PASS |
| Canonical/legacy isolation | Actual linked-worktree fixture converged; foreign run/project list/claim/read/receipt failed; legacy sentinel bytes unchanged. | PASS |
| Template parity | Project and embedded Stop hook entries are synchronous and parity test passed. | PASS |
| Repository-wide verdict | Scoped packages outside `hook`/`cli` passed. Separate full `hook` and `cli` attempts both reached the explicit 180-second local timeout; this is a GAP, not a pass or an attributed regression. | PENDING |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: null
run_commit_sha: pending-m4-repair-commit
run_status: fail
ac_pass_count: 11
ac_fail_count: 4
preserve_list_post_run_count: 1
l44_pre_commit_fetch: not-run
l44_post_push_fetch: not-applicable-no-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-run
  reason: "M4 live blockers and benchmark regression leave the run incomplete"
total_run_phase_files: 28
m1_to_mN_commit_strategy: "M1 cb099897a; M2/M3 6bde8412c; M4 harness 45285bf1b; M4 repair/evidence pending"
```
