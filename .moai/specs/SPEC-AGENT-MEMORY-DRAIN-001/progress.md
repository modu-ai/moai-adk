# SPEC-AGENT-MEMORY-DRAIN-001 — Progress

> Tier M · card t223 · status: draft (plan-phase v0.2.0; plan-audit iter-1 PASS-WITH-DEBT 0.94, fixes applied — awaiting kickoff)

## §E.1 Plan-phase Audit-Ready Signal

- **Claim**: plan-phase artifact set complete for SPEC-AGENT-MEMORY-DRAIN-001 (Tier M:
  spec.md + plan.md + acceptance.md + progress.md) at v0.2.1, status `draft`, authored
  2026-09-02 on card tree `c0c36c421` (branch `WT-agent-memory-drain`). Plan-audit
  iteration 1 returned **PASS-WITH-DEBT 0.94** (Tier M threshold 0.80 met); audit
  defects D1-D3, D5, D7, D8 folded in at v0.2.0. Kickoff approved 2026-09-02 (design (c)
  adopted, AUTONOMOUS progression); v0.2.1 carries the non-transition `phase` correction
  to `"v3.1.4"` (card joins the v3.1.4 close, PR #1685).
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
  (green-path cells are run-phase fixture tests per verification-completeness §2); the
  former `phase`-target assumption was resolved at kickoff — `phase: "v3.1.4"` (operator
  decision 2026-09-02, card joins the v3.1.4 close, PR #1685); the t209 disposal path is
  inference (tree absent, reaper never fired — which unhookable path killed it is
  unobservable post hoc).
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

## §F Phase 4 Mode Selection

- Recorded: 2026-09-02, lane session (card t223; lead dispatch + operator kickoff decisions relayed 2026-09-02).
- Implementation Kickoff Approval: GRANTED by the operator (2026-09-02, relayed via the lead) — run-phase entry approved, progression mode **AUTONOMOUS** (run→sync continuous; no inter-milestone approval pauses). Operator decisions recorded: (1) design (c) write-time mirror + backfill ADOPTED; (2) `phase` corrected to `"v3.1.4"` (card joins the v3.1.4 close, PR #1685) — applied at `f14f0c569`; (3) M3 docs surface = moai-memory.md WITH the Template-First mirror duty (template-source edit + `make build`); (4) backfill auto-runs immediately after M1 lands (the operator's literal "M1 착지 후" — M1 delivers `moai memory drain`, plan M1 step 4 already schedules the backfill with `--json` evidence archived).
- Plan Audit Gate: SKIP taken — final verdict PASS 1.00 (iter-2 delta re-check at `f671d6f6b`; threshold Tier M 0.80). Artifact hash: plan artifacts changed AFTER iter-1 (the fix commits `f671d6f6b` + `f14f0c569`) but the iter-2 delta re-check ran ON the fixed artifacts and passed — the most-recent verdict (PASS 1.00) is on the current hash, so the three skip conditions hold.
- Input parameters: tier M · scope ~2 Go packages (internal/cli new `memory drain` subcommand + internal/hook PostToolUse mirror) + docs (moai-memory.md + template mirror) · domains 2 (Go source, markdown docs+template) · language mix Go+markdown · concurrency benefit LOW (sequential dependency: M1 reconciliation core → M2 mirror shares the path predicate → M3 docs) · agent-teams prereqs: not requested.
- Mode evaluation: `direct` — not selected (multi-file Go implementation with AC discipline); `serial` — SELECTED; `fanout` — not selected (2 domains, coding-heavy, M2 reuses M1's shared predicate — sequential dependency, write-capable parallel fan-out not sanctioned); `sweep` — not selected (authored code, not a mechanical-uniform transform).
- Decision: serial
- Justification: coding-heavy Go work with in-milestone sequential dependencies (M2's mirror anchor is M1's shared path-predicate constant; M3 documents both). Per Anthropic's coding-task caveat, one writer via a single sequential manager-develop delegation covering M1→M3. AUTONOMOUS progression honored by continuous in-session execution; the goal engine is deliberately NOT armed (worktree goal-keying friction on record) — progression is managed by the lane session across the run→sync boundary.
