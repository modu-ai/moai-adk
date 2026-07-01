---
id: SPEC-HAIKU-EFFORT-INERT-001
title: "Progress — remove inert effort field from haiku-tier agents"
version: "0.1.1"
status: completed
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tier: S
tags: "agent, effort, haiku, frontmatter, template"
---

# Progress — SPEC-HAIKU-EFFORT-INERT-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored (spec.md + plan.md + acceptance.md + progress.md), Tier S (`tier: S` frontmatter set). Root cause traced to a stray hand-authored `effort: medium` frontmatter line in the two `model: haiku` agents (`manager-docs`, `manager-git`) — NOT emitted by `agentEffortMap` (already correct). Fix = remove the field from both trees + `make build` + regression guard test (scans both template + local). SPEC ID regex self-check: PASS. plan-auditor FAIL (0.83) addressed at v0.1.1 (D1 §25 non-parity correction, D2 tier field, D3 both-tree guard, D4 GEARS subject) — root cause survived independent verification unchanged. Awaiting Implementation Kickoff Approval (resolution option selection) before run-phase.

## §E.2 Run-phase Evidence

Implementation Kickoff Approval granted (user selected run-phase entry). Resolution: Option 1 (remove the stray `effort` line). cycle_type=tdd.

- **M2 (RED first)**: guard test `internal/template/haiku_effort_guard_test.go` (`TestHaikuAgentsHaveNoEffort`, sentinel `HAIKU_EFFORT_INERT`) added — scans BOTH template source and local mirror, keyed on the `model` field (not agent names). Initial run RED: 4 findings (`manager-docs.md` + `manager-git.md` × 2 trees each declaring `model: haiku` AND `effort:`).
- **M1 (GREEN)**: removed `effort: medium` from all 4 files (template source ×2 + local mirror ×2) + `make build` (embedded template FS recompiled via `//go:embed all:templates`; `catalog.yaml` hashes regenerated). Guard test re-run: GREEN.
- **Files changed (6)**: `.claude/agents/moai/manager-docs.md`, `.claude/agents/moai/manager-git.md`, `internal/template/templates/.claude/agents/moai/manager-docs.md`, `internal/template/templates/.claude/agents/moai/manager-git.md`, `internal/template/catalog.yaml`, `internal/template/haiku_effort_guard_test.go` (+110/−6).
- **Commit provenance**: manager-develop implemented in an L1 worktree (`agent-a81e131d8f87ab080`, commit `bf0e527fa`); push blocked by a parallel-session race (DOCS-V3-REBUILD advanced origin by 2 docs-only commits). Orchestrator reconciled via `cherry-pick -x bf0e527fa` onto main → `54e8f76b3`; fast-forward push `eb8ce68fb..54e8f76b3`. The race was clean (docs-site vs SPEC files, empty overlap).

**run_commit_sha: 54e8f76b3**

## §E.3 Run-phase Audit-Ready Signal

All 9 acceptance criteria PASS, verified twice (manager-develop §E matrix in the worktree + orchestrator independent re-verification on main after cherry-pick):

- AC-HEI-001..002: template source `effort:` count 0, `model: haiku` intact (both agents).
- AC-HEI-003: local mirror `effort:` count 0; `^model:`/`^effort:` line parity `diff` empty for both agents (whole-frontmatter byte-identity NOT asserted — §25 non-parity per `rule_template_mirror_test.go:102-103`).
- AC-HEI-004: `TestHaikuAgentsHaveNoEffort` GREEN (RED→GREEN transition captured); scans both trees.
- AC-HEI-005: `agentEffortMap` = exactly 5 `model: inherit` entries, no haiku key; `TestGetAgentEffort` PASS. No logic change.
- AC-HEI-006: 5 effort-supporting agents retain effort (xhigh × 4 + high × 1), unchanged.
- AC-HEI-007: `make build` exit 0; `catalog.yaml` regenerated (idempotent — re-run on main produced no diff).
- AC-HEI-008: `TestApplyEffortPolicy/no_op_for_agent_not_in_effort_map` PASS.
- AC-HEI-009: `go test ./internal/template/...` — no NEW failure introduced (pre-existing `TestTemplateNeutralityAudit/C5-claude-local-ref` baseline RED confirmed via `git stash` proof, out of this SPEC's scope).

Independent orchestrator batch (on main): guard test GREEN, `go build ./...` exit 0, `make build` idempotent (catalog no diff), `golangci-lint` 0 issues.

## §E.4 Sync-phase Audit-Ready Signal

**sync_commit_sha: 54e8f76b3**

Sync executed orchestrator-direct (Tier S small consolidated lifecycle — worktree-reconcile risk avoidance per memory lesson `feedback_glm_orchestrator_direct_sync_mx` + `feedback_small_spec_consolidated_lifecycle`).

- Status transition: draft → completed (Tier S consolidated close; run implementation already committed + pushed at `54e8f76b3`).
- Era: V3R6 (H-4 — §E.2 + §E.4 + `sync_commit_sha` present). No `era:` override field needed (auto-detection reliable).
- CHANGELOG.md `[Unreleased] > Fixed` entry added (duplicate-check `grep -c 'SPEC-HAIKU-EFFORT-INERT-001' CHANGELOG.md` = 0 before emission).
- specific-path-discipline: only `.moai/specs/SPEC-HAIKU-EFFORT-INERT-001/` + `CHANGELOG.md` staged; unrelated `llm.yaml` (M) isolated.
- sync-auditor: skipped (Tier S doc-only status/CHANGELOG close; no code delta beyond the already-verified run commit).

**Residual debt (plan-auditor iter-2 MINORs, non-blocking, deferred to a later manager-spec touch-up — SPEC body ownership):**
- R1: spec.md §A / plan.md test-citation slightly over-attributes `manager-docs.md` to `rule_template_mirror_test.go:102-103` (that line names `manager-git.md`/`manager-spec.md`; `manager-docs.md` is §25-non-parity but not at that specific line). Underlying claim correct; citation precision only.
- R2: acceptance.md Scenario 2 illustrative Given-When-Then still says "template only" (lags the both-trees normative REQ-HEI-006 / AC-HEI-004 / M2). Normative cells govern; narrative is stale.
