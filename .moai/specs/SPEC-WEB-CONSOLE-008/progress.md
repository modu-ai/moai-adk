# SPEC-WEB-CONSOLE-008 — Progress

> statusline config redesign (honest hybrid). Tier M, cycle_type=tdd.
> Ground-truth: `.moai/reports/web-console-statusline-gitconvention-audit.md` (statusline 17-defect inventory + Redesign Proposal).

## §A Plan-phase

- plan_complete_at: 2026-06-07
- plan_status: audit-ready
- tier: M
- cycle_type: tdd
- plan-auditor: iter-1 0.84 PASS-WITH-DEBT → 6 orchestrator-direct patches (D1/D2/D3/D5/D6/D7) → iter-2 0.87 PASS-WITH-DEBT (monotonic +0.03, MP-1..4 PASS); D-r1 MINOR (AC-WC8-010 decorative clause) cleaned orchestrator-direct.
- GATE-2: user-approved run-phase entry (2026-06-07).

## §E — Phase 0.95 Mode Selection

Input parameters:
- tier: M
- scope (file count): ~12-16 (internal/cli, internal/profile, internal/statusline, internal/web, internal/config, pkg/models, internal/template/templates)
- domain count: 1 (statusline config — single domain, multi-package)
- file language mix: Go + Templ + YAML
- concurrency benefit: LOW (coding-heavy; sequential milestone dependencies M1→M9, Finding A4 caveat)
- Agent Teams prereqs status: NOT met (harness level not `thorough` / team env unconfirmed)

Mode evaluation:

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-package semantic change, not a typo/single-line |
| 2 background | no | Write/Edit required (CONST-V3R2-020 — background auto-denies writes) |
| 3 agent-team | no | REQ-ATR-013 prereqs not all met AND single-domain (<3 domains) |
| 4 parallel | no | coding-heavy, not research-heavy (Finding A4 caveat) |
| 5 sub-agent | YES | coding-heavy default; sequential M1→M9 dependency chain |
| 6 workflow | no | <30 files AND not a single uniform mechanical transform (multi-rule semantic) |

Decision: sub-agent (Mode 5)

Justification: SPEC-WEB-CONSOLE-008 is coding-heavy, single-domain (statusline config), Tier M, ~12-16 files with a sequential milestone dependency chain (M1 template schema → M3 struct Mode removal → M4 preset write-effective → M9 symmetry guard LAST). Per Finding A4 ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path is the safe default. GATE-2 (plan→run HUMAN GATE) passed before this selection — Mode 5 is strictly downstream of GATE-2.

## §E.2 — Run-phase Evidence / Audit-Ready Signal

_(to be populated by manager-develop during run-phase: per-milestone commits, AC PASS/FAIL matrix, run_commit_sha)_

## §E.3 — Run-phase Audit-Ready Signal

_(to be populated at run-phase completion)_
