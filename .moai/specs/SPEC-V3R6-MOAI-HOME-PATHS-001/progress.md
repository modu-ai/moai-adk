# Progress — SPEC-V3R6-MOAI-HOME-PATHS-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-08-16T22:18:29+09:00
- plan_status: audit-ready
- plan_audit_verdict: PASS — iteration 2, score 0.81 (Tier M threshold 0.80; report: `.moai/reports/plan-audit/SPEC-V3R6-MOAI-HOME-PATHS-001-review-2.md`)
- audit_history: iter1 FAIL 0.75 (D1–D6) → remediation 2026-08-16 (orchestrator-direct, user-approved) → iter2 PASS 0.81; residual R1 (M4 milestone stale label) fixed in same session

## Decision Point 2 (2026-08-16, post-iteration-2)

Audit iteration 1 FAIL (0.75) remediated via AskUserQuestion round — 3 decisions, all recommended options selected: (1) 8-accessor surface confirmed for D1, (2) orchestrator-direct remediation edit, (3) full D1-D6 scope. Iteration 2 delta re-audit: PASS 0.81, no blocking defects; residual R1 fixed immediately (plan.md M4 row stale label). An independent plan-auditor verdict now exists — spawn re-measurement post-PR #1574 (context-window fix) succeeded: iteration 1 + iteration 2 both completed with artifacts.

## Decision Point 1 (2026-08-16T22:18+09:00)

User APPROVED via AskUserQuestion (option "승인 (권장)"). Audit basis: orchestrator self-check in lieu of independent plan-auditor (3 subagent spawns died on context-window limits: manager-spec ×2, plan-auditor ×1; no independent verdict exists — no audit_score recorded to the routing ledger for this reason). Self-check caught and fixed 1 AC defect (un-anchored predicate, 218→17) and 1 coverage gap (AC-015). Backlog card t66 registered per same approval round: model-profile simplification (GLM session-model inheritance + web effort-map exposure).

## Authoring Record

- 2026-08-16 research.md persisted (workflow run wf_ebb53184-4f0, 4 lenses + synthesizer, 5 agents, 0 errors).
- 2026-08-16 spec.md / plan.md / acceptance.md authored **orchestrator-direct** per user approval (AskUserQuestion, option "제가 직접 작성") after two `manager-spec` subagent delegations failed with context-window-limit API errors (13:04:24Z, 13:05:59Z; both produced zero artifacts). Disputed-site classification (4 sites) resolved by direct line reads; membership finalized at 17 (predicate-hidden site migrate_agency.go:184 added).
- Pending: plan-audit attempt (spawn; fallback per user-approved degradation: orchestrator self-check + user review) → Decision Point 1.

## §F Phase 4 Mode Selection

Input parameters:
- tier: M
- scope: ~21 files (17 migration sites across 14 files + new `internal/paths` package + envkeys.go registration + homedir.go/preference collapse)
- domain count: 1 (Go home-resolution refactor — mechanical migration spanning 7 packages, single semantic domain)
- file language mix: 100% Go (+ test files)
- concurrency benefit: LOW (coding-heavy; ordered milestone dependencies)
- Agent Teams prereqs: N/A (Mode 3 retired)

Mode evaluation:
1. trivial — not selected: multi-file semantic refactor, no single-line scope
2. background — not selected: write-capable implementation work
3. agent-team — RETIRED (never selected)
4. parallel — not selected: coding-heavy per Anthropic's coding-task parallelism caveat; no ≥3-domain research split
5. sub-agent — **selected**: single manager-develop delegation covering M1→M6 with per-milestone Conventional Commits; orchestrator verifies at commit boundaries
6. workflow — not selected: <~30 files; inter-milestone dependencies (M3-M5 all import the M1 package); not a dependency-free uniform transform

Decision: sub-agent
Justification: Tier M coding-heavy refactor with strictly ordered milestones (package creation → env registration → 17-site migration → helper collapse → full verification). Sequential sub-agent delegation keeps every milestone gate verifiable by the orchestrator and satisfies the Tier M Section A-E delegation-template requirement. Implementation Kickoff Approval passed (user selected "승인 · 표준 진행", 2026-08-16) before this spawn.
