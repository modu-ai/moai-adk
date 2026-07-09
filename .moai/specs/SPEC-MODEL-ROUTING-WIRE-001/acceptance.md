---
id: SPEC-MODEL-ROUTING-WIRE-001
title: "Wire the Tier×Phase Model Routing Matrix into Spawn Paths — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "model-routing, tier-phase-matrix, moai-route, spawn-wiring, model-policy, haiku-inherit, workflow-reflex, acceptance"
---

# SPEC-MODEL-ROUTING-WIRE-001 — Acceptance Criteria

> Observable, testable assertions derived from spec.md §Requirements (GEARS). Each AC traces to REQs and an audit finding (R1/R2/R4/R5).

## §D AC Matrix

| AC ID | REQ trace | Finding | Severity | Description |
|-------|-----------|---------|----------|-------------|
| AC-MRW-001 | REQ-MRW-002 | R1 | MUST-PASS | `moai route M run` resolves the matrix entry (sonnet/xhigh per current workflow.yaml) and prints model+effort |
| AC-MRW-002 | REQ-MRW-002 | R1 | MUST-PASS | `moai route --json` output carries model, effort, fallback_applied fields |
| AC-MRW-003 | REQ-MRW-003 | R1 | MUST-PASS | Invalid tier/phase → non-zero exit + stderr error naming valid enums `{S,M,L}` / `{plan,run,sync,mx}` |
| AC-MRW-004 | REQ-MRW-002 | R1 | MUST-PASS | `RouteModelFor` has ≥1 non-test call site outside `internal/config` (zero-call-site defect closed) |
| AC-MRW-005 | REQ-MRW-001 | R1/R2 | MUST-PASS | run.md Phase 0.95 pre-spawn step instructs consulting model_routing (via `moai route`) and passing per-spawn model/effort args, yielding to explicit override |
| AC-MRW-006 | REQ-MRW-001 | R1/R2 | MUST-PASS | orchestration-mode-selection.md §B.1 lists the routed model/effort as a pre-spawn input signal |
| AC-MRW-007 | REQ-MRW-004 | R2 | MUST-PASS | §A row 4 + §C.2 carry Mode 4 worker-model guidance (sonnet default / haiku pure-extraction / effort per workflow_agents) |
| AC-MRW-008 | REQ-MRW-005 | R4 | MUST-PASS | haiku↔inherit resolved in ONE direction across all 5 surface groups (agent frontmatter ×2 + mirrors, model-policy.md prose, builder-harness.md table) — zero remaining contradiction |
| AC-MRW-009 | REQ-MRW-006 | R5 | MUST-PASS | `team.default_model: inherit` in workflow.yaml (live + template mirror); no `[1m]` suffix remains in that key |
| AC-MRW-010 | REQ-MRW-007 | R1 | MUST-PASS | model_routing comment describes the implemented mechanism; the aspirational unqualified spawn-time-consultation claim is gone |
| AC-MRW-011 | REQ-MRW-008 | R2 | MUST-PASS | mode-orchestration.md team-composition line reconciled with role_profiles SSOT (blanket "(inherit)" contradiction removed) |
| AC-MRW-012 | REQ-MRW-009 | — | MUST-PASS | All edited doc/config surfaces mirrored template-first; `make build` green |
| AC-MRW-013 | (constraint) | — | MUST-PASS | `moai route` static guard: no AskUserQuestion/prompting (grep-based test per internal/cli convention) |
| AC-MRW-014 | (gate) | — | MUST-PASS | Full suite + cross-platform build green; new CLI file coverage ≥85% |

## §D.1 Severity Classification

All 14 ACs are MUST-PASS. AC-MRW-008 additionally carries a **decision-gate precondition**: it can only be evaluated after the D1 direction (plan.md §D) is user-confirmed at Implementation Kickoff Approval; running M3 without the recorded decision is itself a FAIL condition.

## §D.2 Given-When-Then Scenarios

### AC-MRW-001 — matrix resolution

**Given** a project root whose workflow.yaml carries the 12-entry model_routing matrix
**When** `moai route M run` executes
**Then** stdout carries the resolved entry matching the matrix (`model=sonnet effort=xhigh`, exact format per D3) and exit code is 0; **and** `moai route S sync` resolves `haiku/low` correspondingly.

### AC-MRW-003 — invalid enum rejection

**Given** the same project root
**When** `moai route X run` or `moai route M deploy` executes
**Then** exit code is non-zero and stderr names the valid tier set `{S, M, L}` and phase set `{plan, run, sync, mx}`; stdout carries no resolved value.

### AC-MRW-005 — pre-spawn instruction present

**Given** the post-M2 run.md
**When** grepping the Phase 0.95 / pre-spawn region for the routing instruction
**Then** the text instructs: consult `model_routing[<TIER>-<phase>]` (via `moai route <tier> <phase>`), pass the resolved model/effort as per-spawn runtime args, and yield to an explicit caller override — all three elements present.

### AC-MRW-008 — single-direction reconciliation

**Given** the D1 decision recorded (assume Option A: haiku)
**When** grepping all five surface groups
**Then** manager-docs.md + manager-git.md (+ template mirrors) frontmatter carry the decided value, model-policy.md exception prose names the SAME value for the same agents, builder-harness.md table is consistent, and a cross-surface grep finds zero remaining haiku↔inherit disagreement for these two agents. (If Option B was decided, the same check with values swapped and the policy prose rewritten.)

### AC-MRW-010 — truthful comment

**Given** the post-M3 workflow.yaml
**When** reading the model_routing comment block
**Then** it describes the prompt-layer instruction + `moai route` CLI as the consultation mechanism, and no longer asserts unqualified automatic spawn-time consultation by the orchestrator.

## §D.3 Edge Cases

- **EC-1**: absent model_routing block in workflow.yaml → `moai route` surfaces the accessor's documented fallback (`fallback_applied: true`), exit 0 — mirrors RouteModelFor's contract.
- **EC-2**: lowercase input (`moai route m run`) → run-phase decision: normalize or reject; MUST be deterministic and tested either way.
- **EC-3**: `moai route` with missing args → cobra usage error, non-zero exit (no partial resolution).
- **EC-4**: template/live tree divergence pre-existing on an edited file → run-phase halts that surface and reports (no silent overwrite of unmirrored local deltas).

## §D.4 Verification Commands (indicative)

```bash
go test -run TestRoute ./internal/cli/...                                    # AC-MRW-001..003, EC-1..3
grep -rn "RouteModelFor" --include='*.go' internal/ cmd/ | grep -v _test | grep -v 'internal/config/'  # AC-MRW-004 (expect ≥1)
grep -n "model_routing\|moai route" .claude/skills/moai/workflows/run.md      # AC-MRW-005
grep -n "moai route\|model_routing" .claude/rules/moai/workflow/orchestration-mode-selection.md  # AC-MRW-006/007
grep -n '^model:' .claude/agents/moai/manager-docs.md .claude/agents/moai/manager-git.md \
  internal/template/templates/.claude/agents/moai/manager-docs.md internal/template/templates/.claude/agents/moai/manager-git.md  # AC-MRW-008
grep -n "manager-docs\|manager-git" .claude/rules/moai/development/model-policy.md | head  # AC-MRW-008
grep -n "default_model" .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml  # AC-MRW-009
grep -n "consults RouteModelFor" .moai/config/sections/workflow.yaml          # AC-MRW-010 (expect 0 for the unqualified claim)
grep -n "inherit" .claude/skills/moai/workflows/run/mode-orchestration.md     # AC-MRW-011
make build && go test ./... && GOOS=windows GOARCH=amd64 go build ./...       # AC-MRW-012/014
```

## §D.5 Quality Gate Criteria

- TRUST 5: Tested (table-driven CLI tests over 12 matrix entries + fallback), Readable (help text disambiguates from `moai harness route`), Unified (lint clean vs baseline), Secured (no secrets; no prompting), Trackable (Conventional Commits; D1 decision recorded in commit body).
- Doc parity: every edited live-tree doc has an identical template mirror hunk (Template-First rule).

## §D.6 Definition of Done

1. All 14 ACs PASS with verbatim evidence in run-phase §E self-verification.
2. D1 decision recorded (user-confirmed direction + rationale) in progress.md before M3 commits.
3. `make build` green after template edits; template-neutrality CI guard unaffected (no internal SPEC IDs leaked into templates).
4. SPEC-ADVISOR-RUNG-001 prerequisite note satisfied: `moai route` + pre-spawn instruction landed and cited in run.md.
