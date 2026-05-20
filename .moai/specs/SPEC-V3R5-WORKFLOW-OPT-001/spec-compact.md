---
id: SPEC-V3R5-WORKFLOW-OPT-001
title: "Workflow Optimization — Compact Summary"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: ".claude/rules/moai + .moai/config/sections/workflow.yaml + internal/harness/capture + .claude/agents/moai/plan-auditor.md"
lifecycle: spec-anchored
tags: "workflow, optimization, compact, executive-summary"
---

# SPEC-V3R5-WORKFLOW-OPT-001 — Compact Summary

## Problem

W3 HARNESS-AUTONOMY-001 run-phase took **91 minutes** vs target **30 minutes (+200%)**. Meta-analysis identified 4 defect classes that caused 3 manager-develop re-delegations + 1 orchestrator direct-fix; verification/wait time matched implementation time 1:1.

## Solution — 8-Layer Workflow Optimization

| Layer | Domain                | Outcome                                                                                |
|-------|-----------------------|----------------------------------------------------------------------------------------|
| A     | Rule (manager-develop prompt) | 5-section template (Context / Known Issues B1–B8 / Pre-flight / Constraints / Self-Verify) |
| B     | Config (Agent Teams)  | `workflow.yaml team.role_profiles` (5 implementer + tester + reviewer)                  |
| C     | Rule (CI watch)       | `gh pr checks --watch` + `run_in_background: true` standardization                      |
| D     | Rule (verification batch) | 7-item single-turn multi-Bash pattern                                              |
| E     | Rule (Phase Transitions) | Plan-PR overlap with run-start; Plan Audit Gate skip on PASS ≥ 0.90                 |
| F     | Code (defect detector) | `internal/harness/capture/defect_detector.go` heuristic classifier (B1–B8 + confidence) |
| G     | Agent (plan-auditor)  | D7 (Cross-SPEC Reconciliation) + D8 (Cross-Platform Discipline) dimensions             |
| H     | Rule (tool patterns)  | `gh pr checks --json \| jq`, ToolSearch per-turn, single-command idioms                |

## Architecture (4 domains × 6 milestones)

| Milestone | Layers     | Domain              | Risk    |
|-----------|------------|---------------------|---------|
| M1        | A, C, D, E, H | R (rule)         | Low     |
| M2        | B          | C (config)          | Medium  |
| M3        | F          | F (Go code)         | High    |
| M4        | G          | G (agent prompt)    | Medium  |
| M5        | (dogfooding) | Integration       | Low     |
| M6        | (cleanup)  | Docs                | Low     |

M2, M3, M4 are independent and may run in parallel after M1.

## Targets

| Metric                                | Baseline (W3) | Target          |
|---------------------------------------|---------------|-----------------|
| Run-phase wall-time                   | 91 min        | ≤ 30 min        |
| manager-develop delegation count (1-pass) | 3 (33%)   | ≤ 1 (≥ 80%)     |
| CI serial wait                        | 15 min        | ≤ 3 min         |
| Verification serial                   | 10 min        | ≤ 3 min         |

## Requirements (EARS, 14 total)

- **Ubiquitous**: REQ-WO-001 (5-section template), REQ-WO-002 (parallel verify), REQ-WO-003 (no idle on CI wait), REQ-WO-004 (mirror invariant)
- **State-driven**: REQ-WO-010 (team spawn), REQ-WO-011 (audit skip), REQ-WO-012 (D7), REQ-WO-013 (D8)
- **Event-driven**: REQ-WO-020 (defect classify), REQ-WO-021 (plan-PR scan), REQ-WO-022 (syscall check), REQ-WO-023 (auto-prepend lessons)
- **Optional**: REQ-WO-030 (evaluator-active auto-spawn)
- **Unwanted**: REQ-WO-040 (Section B missing → halt), REQ-WO-041 (mirror missing → CI fail)

## Acceptance (14 binary ACs)

All ACs include verification commands producing exit-0 PASS / exit-≥1 FAIL signals. Highlights:

- AC-WO-001: Self-dogfooding ≤ 30 min
- AC-WO-002: B1–B8 markers present in delegation prompt
- AC-WO-003: Template mirror byte-parity + `make build` regen
- AC-WO-005: defect_detector.go ≥ 90% coverage
- AC-WO-014: zero NEW spec-lint / golangci-lint / test findings (delta-only)

## Exclusions (8)

- EXCL-WO-001: W4 PROJECT-MEGA harness self-improvement
- EXCL-WO-002: PR #1024 W3 merge (user natural area)
- EXCL-WO-003: observer.go path bug (separate SPEC)
- EXCL-WO-004: Lint baseline cleanup (separate SPEC)
- EXCL-WO-005: GitHub webhooks
- EXCL-WO-006: Other-domain meta-analysis
- EXCL-WO-007: No syscall introduction by this SPEC
- EXCL-WO-008: No stray working-tree files in commits

## Cross-SPEC Reconciliation

Add-only relationship to W3 (capture/), V3R2-WF-003/WF-004 (workflow.yaml mode dispatch), plan-auditor (D1–D6). No retirement, no reversal.

## Risks (7)

R-WO-01 Agent Teams instability (mitigated by solo fallback), R-WO-02 false-positive defects (heuristic confidence ≥ 0.7), R-WO-03 D7/D8 over-strict (iter-1 user review), R-WO-04 W3 PR #1024 dependency (rebase strategy), R-WO-05 mirror drift (CI guard), R-WO-06 background watch blocking (notification pattern), R-WO-07 dogfooding misses target (partial improvement still acceptable).

## Definition of Done

All 14 ACs PASS + 6 edge cases PASS + quality gates green + run-PR + sync-PR merged + lessons #22 archived.
