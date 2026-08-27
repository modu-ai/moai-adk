# Progress — SPEC-CLI-TEST-CWD-ISOLATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-28
- Artifacts: spec.md / plan.md / acceptance.md (**Tier M, 3 artifacts** — rationale in
  spec.md §C: measured scope is S-shaped; M stands on the canonical 3-file artifact set,
  the iter-1-FAIL retry-slot reality, and the guard+per-test verification ceremony;
  downgrade path documented for independent audit judgment) + this progress skeleton.
- Budget: REQ 5 / AC 5 — Tier M ceilings (16 / 16) respected.
- Evidence base: **measured RED reproducer** (lead session, this worktree, base `d34a789a4`,
  2026-08-28; probe record `.moai/reports/t334/red-probe.md`): env-scrubbed
  `go test ./internal/cli -run 'Kanban|Factory' -count=1` → rc=0, 0.869 s, recreates
  `state/todo/leads.json` + `state/factory/workers.json`. Negative probes: internal/config,
  internal/kanban, `Config|Cache|Settings` — all clean (config-cache reclassified
  historical; internal/hook not probed, no claim). Base-tree fact: 4 committed `.moai` dirs
  on `d34a789a4` → tree judgments are baseline deltas, never emptiness. Plus t317 §G-1
  (`0ad4b52ba`) and primary-checkout residue re-verified 2026-08-28.
- **plan-audit iter1 (2026-08-28): FAIL 0.775 vs Tier M 0.80** — report
  `.moai/reports/t334/plan-audit-iter1.md`. v0.2.1 revision = D1 (baseline-delta scan) +
  D2 (tier M restored, acceptance.md retained) + D3 (attribution aligned, red-probe.md
  cited) + D4 (internal/hook struck) + D5/D7 (wording). iter2 delta re-audit pending.

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop populates; the re-established RED (frozen reproducer,
verbatim output + exit code + HEAD SHA) lands here first>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates; carries sync_commit_sha on close>_
