---
id: SPEC-HAIKU-EFFORT-INERT-001
title: "Remove inert effort field from haiku-tier agents (manager-docs, manager-git)"
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

# SPEC-HAIKU-EFFORT-INERT-001 — Remove inert `effort` field from haiku-tier agents

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-01 | manager-spec | Initial plan-phase draft (Tier S). Root-cause traced: stray hand-authored `effort: medium` in haiku agent frontmatter, NOT emitted by `agentEffortMap`. |
| 0.1.1 | 2026-07-01 | manager-spec | plan-auditor FAIL (0.83) revision: D1 corrected false "byte-identical mirror" premise → §25 non-parity, model/effort-line-only parity (REQ/AC-HEI-003); D2 added `tier: S` frontmatter; D3 guard scans both trees (REQ-HEI-006, M2, AC-HEI-004); D4 REQ-HEI-007 canonical GEARS subject. Root cause unchanged. |

## §A Context / Problem

Two of the 8 retained MoAI agents declare a reasoning-`effort` level that is **silently inert** because their model does not support effort:

- `.claude/agents/moai/manager-docs.md` — `model: haiku`, `effort: medium`
- `.claude/agents/moai/manager-git.md` — `model: haiku`, `effort: medium`

### External evidence (verified against official Claude Code docs during investigation)

- Official `code.claude.com/docs/en/sub-agents` frontmatter table: `effort` is a REAL field honored by Claude Code — "Effort level when this subagent is active. Overrides the session effort level. ... Options: low, medium, high, xhigh, max; **available levels depend on the model**."
- Official `code.claude.com/docs/en/model-config`: "The available effort levels depend on the model. **Models not listed here do not support effort.**" The effort-supported list is Fable 5, Sonnet 5, Opus 4.8/4.7/4.6, Sonnet 4.6. **Haiku is NOT listed → Haiku does not support effort.**
- Conclusion: for the 5 agents with `model: inherit` (which resolve to the session's Opus/Sonnet), their `effort` frontmatter IS honored. For the 2 haiku agents, `effort: medium` is silently inert — a misleading configuration.

### Root-cause finding (in-repo, verified)

The frontmatter `effort` values are written at `moai init` / `moai update` by `ApplyEffortPolicy` reading `agentEffortMap` in `internal/template/model_policy.go`. Investigation of that file establishes:

- `agentEffortMap` (model_policy.go ~L164-170) contains **exactly 5 entries**: `manager-spec`, `plan-auditor`, `sync-auditor` (xhigh), `manager-develop` (xhigh), `builder-harness` (high) — all `model: inherit` agents whose resolved model supports effort.
- `manager-docs` and `manager-git` are **NOT** in `agentEffortMap`. `GetAgentEffort` returns `""` for them, so `ApplyEffortPolicy` **skips** them (no injection). This is already exercised by the existing test `TestApplyEffortPolicy/no_op_for_agent_not_in_effort_map` (model_policy_test.go ~L672).
- Therefore the `effort: medium` on the two haiku agents is **hand-authored directly in the template frontmatter** — it is NOT emitted by the generator. **The map is already correct** and needs no change to stop emitting effort for haiku agents.

### Propagation path (why the fix must target the SOURCE)

Per CLAUDE.local.md §2 Template-First Rule, `.claude/agents/moai/*.md` are template-managed. The authoritative source is `internal/template/templates/.claude/agents/moai/`; the local `.claude/agents/moai/` is a mirror. **These two agents are §25-sanitized NON-PARITY files** (CLAUDE.local.md §25 internal-content isolation): their local (working-copy) `description:` block and body carry internal SPEC-IDs and dates (e.g. `SPEC-V3R6-LIFECYCLE-REDESIGN-001`, `2026-05-25`) that are STRIPPED from the distributed template mirror, so full-frontmatter byte-identity is doctrinally **prohibited** and is explicitly excluded from the mirror-parity test (`internal/template/rule_template_mirror_test.go:102-103` enrolls `manager-git.md` / `manager-spec.md` as sanitized non-parity files, routing cleanliness to `TestTemplateNoInternalContentLeak` instead of byte-parity). Verified by live `diff`: both files DIFFER between the two trees, but the divergence is confined to the `description:` block and body prose — the `^model:` and `^effort:` lines are identical across the two trees today. The stray `effort:` line exists in BOTH trees and must be removed from BOTH (followed by `make build`). The **only** parity this fix requires or touches is `^model:` / `^effort:` line parity — NOT whole-frontmatter byte-identity. `ApplyModelPolicy` / `ApplyEffortPolicy` only run on user projects at `moai init` / `moai update`, not on the dev tree.

## §B Requirements (GEARS)

- **REQ-HEI-001** (Ubiquitous): The `manager-docs` and `manager-git` agent definitions **shall not** declare an `effort` frontmatter field, because their `model: haiku` does not support effort levels (effort is silently inert on Haiku).
- **REQ-HEI-002** (Capability gate — Where): **Where** an agent definition declares `model: haiku`, the agent definition **shall not** declare an `effort` field. This is the invariant the fix establishes and the guard test enforces.
- **REQ-HEI-003** (Ubiquitous): After the fix, the `^model:` and `^effort:` frontmatter lines of the two haiku-tier agents **shall** be identical between the template source (`internal/template/templates/.claude/agents/moai/`) and the local mirror (`.claude/agents/moai/`). Whole-frontmatter byte-identity is NOT required and NOT asserted — these are §25-sanitized non-parity files (`rule_template_mirror_test.go:102-103`); only the `model` / `effort` line parity is in scope.
- **REQ-HEI-004** (Ubiquitous): The `agentEffortMap` in `internal/template/model_policy.go` **shall** map an effort value only to effort-supporting (`model: inherit`) agents; it **shall not** be extended to emit an effort value for any haiku-tier agent. (No generator logic change is required — this REQ records the invariant, not an edit.)
- **REQ-HEI-005** (Event-driven — When): **When** `make build` runs after the frontmatter edit, the embedded template FS **shall** reflect the effort-field removal for the two haiku agents (the recompiled binary carries the corrected frontmatter).
- **REQ-HEI-006** (Capability gate — Where): **Where** a guard test scans the agent frontmatter in BOTH the template source (`internal/template/templates/.claude/agents/moai/`) AND the local mirror (`.claude/agents/moai/`), the test **shall** fail if any agent declaring `model: haiku` also declares an `effort` field (keyed on the model field, not agent names), and **shall** pass once the two stray lines are removed from both trees.
- **REQ-HEI-007** (Ubiquitous): The `manager-spec`, `plan-auditor`, `sync-auditor`, `manager-develop`, and `builder-harness` agent definitions **shall** retain their existing `effort` field (xhigh × 4, high × 1) unchanged.

## §C Exclusions (Out of Scope)

### Out of Scope — model tier changes
- Do NOT change `manager-docs` or `manager-git` off `model: haiku`. Sync/git work is lightweight; haiku is the intended, correct model (the fix is to remove the meaningless effort field, not to make effort meaningful by upgrading the model).
- Do NOT add an effort-supporting model tier to either agent.

### Out of Scope — agentEffortMap logic rewrite
- No change to `agentEffortMap`, `GetAgentEffort`, `ApplyEffortPolicy`, or `insertEffortInFrontmatter` logic. The map already excludes haiku agents; the generator is already correct. This SPEC only removes stray hand-authored frontmatter and adds a guard.

### Out of Scope — user-project migration tooling
- No bespoke migration script for existing user projects. `moai update` overwrites `.claude/agents/moai/` (template-managed, overwrite-on-sync) with the corrected template, so downstream projects self-heal on their next update.

### Out of Scope — other agents' effort fields
- The 5 effort-supporting agents (`model: inherit`) keep their `effort` fields (xhigh × 4, high × 1) unchanged.

### Out of Scope — implementation details
- Exact test placement (a new `internal/template/*_test.go` file vs. an added function in `model_policy_test.go`) is a Run-phase decision, deferred to `manager-develop`.

## §D Acceptance Summary

See `acceptance.md` for the full AC matrix and Given-When-Then scenarios. In brief: both haiku agents lose their `effort` line (both trees), `make build` succeeds, a guard test enforces the `model: haiku ⇒ no effort` invariant, the 5 effort-supporting agents are untouched, and `go test ./internal/template/...` passes.

## §E Cross-References

- `internal/template/model_policy.go` — `agentEffortMap`, `ApplyEffortPolicy`, `GetAgentEffort` (generator; no logic change).
- `internal/template/model_policy_test.go` — existing effort tests, including `no_op_for_agent_not_in_effort_map` (evidence the generator does not re-add haiku effort).
- `.claude/rules/moai/development/model-policy.md` § Inherit-by-Default Convention — records that `manager-docs` / `manager-git` are the two `model: haiku` exceptions.
- `.claude/rules/moai/development/agent-authoring.md` § Supported Frontmatter Fields — `effort: ... (xhigh/max require Opus 4.7+)`.
- CLAUDE.local.md §2 Template-First Rule — template source + `make build` + local mirror discipline.
