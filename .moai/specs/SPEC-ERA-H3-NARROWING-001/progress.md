# SPEC-ERA-H3-NARROWING-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- tier: S
- artifacts: spec.md + plan.md (+ progress.md) — AC는 spec.md §3에 인라인
- baseline_tree: f72c0bf0f (worktree `.claude/worktrees/t382`, branch `WT-era-plan-phase`) — re-measured after the plan artifacts landed, because this SPEC is itself an instance of the defect (V3R5 23 -> 24, grandfathered 285 -> 286)
- superseded_baseline_tree: 9328a5242 (pre-artifact; background only)
- measurement_tool: `./bin/moai` (this tree, `make build` rc=0). No PATH binary.
- decision_evidence: `.moai/reports/t382/red-evidence.md` (R1-R3 — command + verbatim stdout + exit code + tree SHA per verification-completeness.md 2.1); probe body `.moai/reports/t382/red_probe.py`
- background_evidence: `.moai/reports/t382/measurements-9328a5242.md` (M1-M13), `v3r5-population.txt`, `drift-before-9328a5242.txt` — no verbatim stdout or exit codes, so NOT cited as decision basis
- red_state_at_plan_close: R1 exit 1, `misclassified: 23` of 24 swept

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
