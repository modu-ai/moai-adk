# progress.md — SPEC-MEMORY-FOLD-BUDGET-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-04
- revision: 0.2.0 (plan delta for plan-audit iteration 1; the previous audited revision was 0.1.0 at `73c4ab646`)
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set) + progress.md + decision-index.md (conditional, `interview.decision_gate: on`) + fixtures/store-A/ (synthetic, read-only acceptance input, unchanged in this revision)
- req_count: 14
- ac_count: 14
- baseline_tree: 2f492df19
- certain_file_count: 20 (Tier M band 5-15; flagged for the leader, plan.md §A and §I)

### Plan-phase lint (tool provenance: judging build vs measured tree)

- judging build: `./bin/moai`, built from this tree and invoked by path, commit stamped `2f492df19` through `-ldflags` (`./bin/moai version` prints `v3.1.3   2f492df19   built 2026-10-04`); measured tree HEAD for the 0.2.0 run: `73c4ab646` (parent `2f492df19`; no Go source differs). The Go VCS stamp of a plain `go build` in this worktree named another checkout (`c8f245c2c9a5`, not an ancestor of HEAD), so it was not trusted.
- 0.2.0 run: `./bin/moai spec lint SPEC-MEMORY-FOLD-BUDGET-001` → stdout `✓ No findings — all SPEC documents are valid`, exit 0.
- 0.2.0 run: `./bin/moai spec lint --strict SPEC-MEMORY-FOLD-BUDGET-001` → same stdout, exit 0.
- 0.1.0 run (kept for the record): same two commands, same output, exit 0; positive control (a deliberately bad scratch spec, `phase: plan`, uppercase conditional keywords, no AC, no exclusions) → exit 1, `1 error(s), 4 warning(s)` — the lint can fail.
- `moai spec audit` (MCP, filter this SPEC, `project_root` this worktree), 0.1.0 run → era `V3R6`, one INFO `EraAutoDetected`, `modern_era_clean: 1`, no drift. Not repeated for 0.2.0.

### Plan-audit history

- iteration 1 (0.1.0, audited sha `73c4ab646880c68743966b9d6e8c1b98e3112632`): FAIL 0.69, 12 blocking (D1-D12), 8 optional (D13-D20). Report: `.moai/reports/t1502/plan-audit.md` (gitignored card evidence). Leader disposition: `.moai/reports/t1502/leader-disposition.md`.
- iteration 2 input (0.2.0): this revision. Disposition of every finding: `plan.md` §L for D13-D20 and the plan-delta report for D1-D12. A confirming audit has not run.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
