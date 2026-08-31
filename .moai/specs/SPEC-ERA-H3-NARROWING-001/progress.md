# SPEC-ERA-H3-NARROWING-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- tier: S
- artifacts: spec.md + plan.md (+ progress.md) — AC는 spec.md §3에 인라인
- baseline_tree: f72c0bf0f (worktree `.claude/worktrees/t382`, branch `WT-era-plan-phase`) — re-measured after the plan artifacts landed, because this SPEC is itself an instance of the defect (V3R5 23 -> 24, grandfathered 285 -> 286)
- superseded_baseline_tree: 9328a5242 (pre-artifact; background only)
- measurement_tool: `./bin/moai` (this tree, `make build` rc=0). No PATH binary.
- decision_evidence: `.moai/reports/t382/red-evidence.md` (R1-R4 — command + verbatim stdout + exit code + tree SHA per verification-completeness.md 2.1). Probe bodies: `red_probe.py` (R1), `drift_probe.py` (R4). Per-item attribution: R1-R3 at `1f10f5e8d`, R4 at `f967089ba`
- background_evidence: `.moai/reports/t382/measurements-9328a5242.md` (M1-M13), `v3r5-population.txt`, `drift-before-9328a5242.txt` — no verbatim stdout or exit codes, so NOT cited as decision basis
- red_state_at_plan_close: R1 exit 1 (`misclassified: 23` of 24 swept); R4 exit 1 (`unearned exemption: 22` of 23 era-exempt rows)
- plan_audit: iter 1/1 (Tier S ceiling), verdict PASS-WITH-DEBT, score 0.825 vs Tier S baseline 0.75. Report `.moai/reports/t382/plan-audit-iter1.md` (auditor-owned, not modified by this card)
- audit_debt_repaid: D1-D4 (blocking) + D5-D10 (optional) all closed at v0.4.0. Auditor findings independently re-verified before adoption: D1 via `drift_probe.py`, D2 via `d2_check.py`, D7 via `d7_check.py`, D6/D9/D10 via git and grep

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
