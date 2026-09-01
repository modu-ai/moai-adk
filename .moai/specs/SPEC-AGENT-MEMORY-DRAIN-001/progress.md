# SPEC-AGENT-MEMORY-DRAIN-001 — Progress

> Tier M · card t223 · status: draft (plan-phase complete, awaiting plan-audit + kickoff)

## §E.1 Plan-phase Audit-Ready Signal

- **Claim**: plan-phase artifact set complete for SPEC-AGENT-MEMORY-DRAIN-001 (Tier M:
  spec.md + plan.md + acceptance.md + progress.md) at v0.1.0, status `draft`, authored
  2026-09-02 on card tree `c0c36c421` (branch `WT-agent-memory-drain`).
- **Evidence**:
  - SPEC ID pre-write check: `SPEC-AGENT-MEMORY-DRAIN-001` → `PASS` (executed Bash).
  - Uniqueness: `ls .moai/specs | grep -c AGENT-MEMORY` → 0 pre-existing.
  - RED-now cells measured in this run on this tree: `moai memory drain` →
    `Unknown command "drain" for "moai memory".` exit 1; worktree topic files 40, primary
    overlap 0; `grep -n "mirror" internal/hook/post_tool.go` → no match; per-worktree
    native `memory/` dirs → 0 (baseline AC-AM-009).
  - Scale re-measured: 186 worktrees / 88 agent-memory trees / 26 file-bearing / 70
    files (stale 5-of-156 superseded).
- **Baseline-attribution**: all commands run 2026-09-02 from
  `.claude/worktrees/t223` at `c0c36c421`, outputs observed verbatim in this session.
- **Gaps**: mirror runtime behavior and drain execution are design-only at plan phase
  (green-path cells are run-phase fixture tests per verification-completeness §2);
  `phase: "v3.1.5 target"` assumes v3.1.4 closes with pending PR #1685 — operator
  confirmation at the kickoff gate.
- **Residual-risk**: the write-time mirror sees Write/Edit tool calls only (Bash-written
  memory bypasses it — accepted blind spot, spec.md §G); the index-append race window is
  tolerated and doctor-detectable; the t209 concrete instance is already lost and this
  SPEC prevents recurrence, not that loss.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
