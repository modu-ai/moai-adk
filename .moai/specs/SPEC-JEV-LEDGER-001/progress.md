# progress — SPEC-JEV-LEDGER-001

status: draft

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-26T10:21:34Z
tier: M
artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md
tree_measured: cb30e05b3e5f64774b8dabd87e80c24332f4551d
card: t1762
note: scope exceeds the ~10 files / ~400 lines card guidance (15 files, ~1080 lines est. at iter-3); carried as one card, serial M1 -> M4, one commit per milestone (plan.md §E; decision-index Q6)

iter2_repair:
  trigger: plan-audit FAIL 0.6875 at audited_sha d1aa4ed656af948c95c2567718c4f9a7be54e191 (.moai/reports/t1762/jev-ledger-plan-audit.md)
  repaired_at: 2026-10-26T10:41:41Z
  defects: D1 margin 1 s / deadline 4 s (Q2 filled, plan.md clarification marker removed); D2 REQ-016 aggregates valid rows + skipped N; D3 one card serial, AC partial/final closure map, split rejected; D4 AC-004 four modes x 3 events x 2 harnesses, AC-006 call-site mutant; D5 per-path preconditions (answered / write-failure / timeout / skipped); D6 per-consumer tolerant yaml.Node decode + loader-level AC-007; D7 RED cells re-collected with in-invocation exit capture on d1aa4ed656; D8 REQ-002/006/007 restated with named subjects
  open_gap: hook-handler elapsed time unmeasured (M3 measures)

iter3_repair:
  trigger: plan-audit FAIL 0.81 at audited_sha 82aff00a6d3b9d1b3dd6f9c2a6fc68ebf5b8e0d0 (.moai/reports/t1762/plan-audit-verdict.md)
  version: 0.1.2 (spec, plan, acceptance)
  defects: N1 cost unit renamed to estimated input tokens, with an "estimated" report label and the response usage.input_tokens out of scope; N2 ledger open with nofollow and nonblock plus an opened-handle regular-file check (REQ-011/012, plan §F), and AC-012 FIFO and symlink cases with cleanup guarantee; N3 AC-001 asserts builder calls = 0 and Ask calls = 0, the gate-skip mutant now goes red, and the vacuous M2 partial was removed; N4 AC-013 intermediate-elapsed case (250/50 ms, 150 ms elapsed, deadline = start + 200 ms, T = 50 ms) with the entry-anchor mutant; N5 M3 measures the worst-case wrapper-start-to-exit wall clock and blocks when p99 > 5 s - S (S = 0.5 s, decision); N6 AC-013 compares the exit code to the gate-off baseline, with an exit-2 fault-fixture case
  optional: O1 file count 15 / ~1080 lines; O3 AC-007 scalar consumers case; O6 report label strings fixed; O7 hookEventSubcommands
  red_recollected: E1', E3', E6', E10 on 82aff00a6d (acceptance §L)
  open_gap: S = 0.5 s and the 1 s margin are decisions until M3 measures; Codex wrapper command line not read

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
