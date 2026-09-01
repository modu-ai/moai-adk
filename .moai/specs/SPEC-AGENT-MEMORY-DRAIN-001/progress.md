# SPEC-AGENT-MEMORY-DRAIN-001 — Progress

> Tier M · card t223 · status: draft (plan-phase v0.2.0; plan-audit iter-1 PASS-WITH-DEBT 0.94, fixes applied — awaiting kickoff)

## §E.1 Plan-phase Audit-Ready Signal

- **Claim**: plan-phase artifact set complete for SPEC-AGENT-MEMORY-DRAIN-001 (Tier M:
  spec.md + plan.md + acceptance.md + progress.md) at v0.2.0, status `draft`, authored
  2026-09-02 on card tree `c0c36c421` (branch `WT-agent-memory-drain`). Plan-audit
  iteration 1 returned **PASS-WITH-DEBT 0.94** (Tier M threshold 0.80 met); audit
  defects D1-D3, D5, D7, D8 folded in at v0.2.0.
- **Evidence**:
  - SPEC ID pre-write check: `SPEC-AGENT-MEMORY-DRAIN-001` → `PASS` (executed Bash).
  - Uniqueness: `ls .moai/specs | grep -c AGENT-MEMORY` → 0 pre-existing.
  - RED-now cells (single-invocation carrier form, verification-completeness §2.1):
    `moai memory drain` → `Unknown command "drain" for "moai memory".` exit 1;
    `comm -12 /tmp/t223-wt-topics.txt /tmp/t223-primary-topics.txt` → empty stdout,
    exit 0 (EV-002); `grep -n "mirror" …/internal/hook/post_tool.go` → empty stdout,
    exit 1 (EV-005); per-worktree native `memory/` dirs → 0 (baseline AC-AM-009).
  - Scale re-measured: 186 worktrees / 88 agent-memory trees / 26 file-bearing / 70
    files at the 2026-09-02 snapshot (30 per-agent `MEMORY.md` indices + 40 topics;
    breakdown manager-develop 12, manager-spec 9, plan-auditor 6, sync-auditor 1,
    manager-lead 1, manager-docs 1). Drifting population — re-count minutes later read
    73 (31 + 42). Stale 5-of-156 superseded.
- **Baseline-attribution**: all commands run 2026-09-02 from
  `.claude/worktrees/t223` at `c0c36c421`, outputs observed verbatim in this session.
- **Gaps**: mirror runtime behavior and drain execution are design-only at plan phase
  (green-path cells are run-phase fixture tests per verification-completeness §2);
  `phase: "v3.1.5 target"` assumes v3.1.4 closes with pending PR #1685 — operator
  confirmation at the kickoff gate; the t209 disposal path is inference (tree absent,
  reaper never fired — which unhookable path killed it is unobservable post hoc).
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
