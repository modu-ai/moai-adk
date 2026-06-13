# SPEC-SEC-HARDEN-001 — Progress

## §A — Phase 0.95 Mode Selection

**Input parameters**
- tier: L
- scope (file count): ~10-14 (M1 stack.go, M2 conflict.go + conflict_test.go, M3 tmux_integration.go + session.go + tmux_integration_test.go, M4 tracker.go, M5 circuit.go + per-milestone test files)
- domain count: 1 (Go source — security/concurrency internal packages: permission, tmux, lsp, resilience)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, behavior-preserving security/concurrency fixes — Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: not met (harness not `thorough` / `workflow.team.enabled` not set / env flag unset)

**Mode evaluation**

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-file semantic security/concurrency changes, not a typo |
| 2 background | no | writes code (not read-only) |
| 3 agent-team | no | Agent Teams prereqs unmet + coding-heavy |
| 4 parallel | no | coding-heavy, not research (Anthropic coding-task parallelism caveat) |
| 5 sub-agent | **YES** | Tier L coding-heavy default; sequential per-milestone |
| 6 workflow | no | not mechanical-uniform (semantic security fixes, inter-milestone behavior reasoning) |

**Decision: sub-agent (Mode 5)** — sequential per-milestone implementation.

**Justification**: Tier L, coding-heavy, behavior-preserving security/concurrency work. Per Anthropic's coding-task parallelism caveat, the sequential sub-agent path (Mode 5) is the safe default; Mode 6 is excluded because the milestones are semantic fixes, not a single uniform mechanical transform. A single `manager-develop` (cycle_type=tdd) delegation implements M1-M5 sequentially with per-milestone commits; the orchestrator independently verifies the batch afterward. L1 `isolation: worktree` is used to isolate run-phase commits from the active parallel SPEC-MERGE-METHOD-CONFIG-001 session that shares the main working tree (Race Absorbed: disjoint scope, no file overlap).

**Plan Audit (Phase 0.5)**: plan-auditor PASS 0.91 (Tier L threshold 0.85, margin +0.06); 3 MINOR defects (D1 M2 deny-precedence dual-phrasing, D2 stale baseline SHA, D3 M2 log-path seam run-phase deferral), none blocking. Report: `.moai/reports/plan-audit/SPEC-SEC-HARDEN-001-2026-06-13.md`.

**GATE-2**: user-approved run-phase entry (구현 착수 승인).
