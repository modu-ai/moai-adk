---
id: SPEC-V3R6-RULES-CATALOG-SCRUB-001
title: "Archived-agent catalog scrub across .claude/rules"
version: "0.1.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai + internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "rules, catalog-scrub, archived-agents, template-mirror, consistency"
era: V3R6
depends_on: []
related_specs: [SPEC-V3R6-AGENT-TEAM-REBUILD-001, SPEC-V3R6-CATALOG-SSOT-001]
tier: M
---

# SPEC-V3R6-RULES-CATALOG-SCRUB-001 — Archived-agent catalog scrub across .claude/rules

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-19 | manager-spec | Plan-phase draft. SPEC 2 of 4-SPEC Sprint 16 "rules-improvement" cohort. Owns the archived-agent catalog scrub. Defect inventory derived from grep sweep (25 files matched) + manual classification (live-spawn/example/hook-action defects vs intentional enumeration preserves). 14 defect groups across 18 files. |

## §A Context

The 2026-05-25 Anthropic consolidation (SPEC-V3R6-AGENT-TEAM-REBUILD-001) reduced the MoAI agent catalog from 17 entries to **8 retained agents** (7 MoAI-custom: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness` + Anthropic built-in `Explore`). **12 agents were archived**: `manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide` (the MoAI-custom one — distinct from the Anthropic built-in helper of the same name), `researcher`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`.

The orchestrator rejects any spawn of an archived agent (`ARCHIVED_AGENT_REJECTED`) per `.claude/rules/moai/workflow/archived-agent-rejection.md`. Despite this, a grep sweep found archived-agent names surviving in 25 rule files. Not all are defects: the **intentional enumerations** in `archived-agent-rejection.md` (the migration table) and `NOTICE.md` (the archive provenance record) MUST be preserved — they ARE the canonical reference lists. The defects are the residual references that present an archived agent as a **live spawn target, a live normative example, or a live hook action** — these silently mislead the orchestrator and contradict the file's own consolidation narrative.

This SPEC owns the mechanical scrub of those live-reference defects across `.claude/rules/`, plus the identical edits in the `internal/template/templates/.claude/rules/` mirror (Template-First).

## §B Scope

### In-scope defect inventory (verified file:line)

Each defect group below was verified against the live tree during plan-phase. The canonical replacement follows the `archived-agent-rejection.md` §C migration table: `Agent(general-purpose)` with a domain whitelist for implementation work, `Agent(Explore)` for read-only investigation, or a retained agent / role_profile.

| # | File | Line(s) | Defect | Canonical replacement |
|---|------|---------|--------|----------------------|
| D1 | `core/agent-hooks.md` | ~52-55 | Archived `expert-backend` / `expert-frontend` / `expert-devops` / `manager-quality` rows in the live `{agent}-{phase}` hook-action table | Remove the 4 archived rows OR mark the table "archived — non-functional, historical mapping only" |
| D2 | `core/agent-common-protocol.md` | 345, 401 | `expert-*` in pre-spawn normative example (L345); `manager-quality in diagnostic mode` in read-only exemption (L401). (L37 `phantom manager-quality / expert-security` is intentional prose — PRESERVE) | L345 → `general-purpose with domain whitelist`; L401 → `Explore` (read-only) |
| D3 | `core/zone-registry.md` | 561 | CONST entry `[HARD] expert-frontend MUST implement design tokens` — archived name frozen in clause text (mirrors design/constitution D14) | Resolve jointly with D14 (carve-out note OR retained rename); at minimum cross-reference migration table. (L240 `researcher` is a role_profile — PRESERVE) |
| D4 | `workflow/ci-autofix-protocol.md` | 83, 101, 107, 122 | **P0** — `manager-quality` invoked as LIVE diagnostic subagent (would trigger `ARCHIVED_AGENT_REJECTED`) | Stop hook `sync-phase-quality-gate.sh` and/or per-spawn `Agent(general-purpose)` diagnostic scope |
| D5 | `workflow/ci-watch-protocol.md` | 79, 113 | **P0** — `manager-quality` as live Wave-3 handoff target | Same as D4 |
| D6 | `workflow/agent-teams-pattern.md` | 3 | Frontmatter `paths:` references non-existent `.claude/agents/moai/manager-strategy.md` (dead path token) | Remove the dead path token only (full file consolidation belongs to SSOT-DEDUP-001) |
| D7 | `workflow/spec-workflow.md` | 220, 368 | Phantom team agents `team-validator` / `team-tester` (L220), `backend-dev` / `frontend-dev` (L368) in Team Mode examples | Rewrite to role_profiles (implementer / tester / reviewer) spawned as `general-purpose` |
| D8 | `workflow/worktree-integration.md` | 167, 253 | Archived `expert-backend` / `expert-frontend` / `expert-refactoring` + `researcher` named as isolation-requiring agents (L167 HARD list); stale "5 agents declare isolation" count (L253) that no longer holds under the 8-agent catalog | Recount against the 8-agent catalog; replace archived names with retained equivalents / role_profiles |
| D9 | `workflow/worktree-state-guard.md` | 62 | MoAI-custom `claude-code-guide` (archived) as escalation target. (The Anthropic built-in `claude-code-guide` is a separate VALID agent — clarify which is meant) | `Agent(Explore)` read-only investigation |
| D10 | `development/agent-authoring.md` | ~128-149 | **P0** — §Agent Categories ships the OLD 17-agent catalog (Manager Agents (7) + Expert Agents (6) + Builder), self-contradicting the file's own later 8-agent ceiling section | Replace with the 7 retained MoAI-custom agents + `Explore` |
| D11 | `development/orchestrator-templates.md` | 102-104, 175 | Spawn examples use invalid `subagent_type: "analyzer"` / `"designer"` / `"implementer"` / `"reviewer"` | Rewrite to `general-purpose` + role profile |
| D12 | `development/model-policy.md` | ~84-90 | Model Policy Tiers table references archived agent names (`strategy`, `researcher`, `quality`) in Opus/Sonnet/Haiku columns | Update to 8-agent catalog OR note tiers are role-profile-based |
| D13 | `languages/{cpp,csharp,elixir,flutter,javascript,r,ruby,typescript}.md` | (boilerplate) | 8 files carry the identical line `delegate to \`manager-quality\` agent for AI-powered debugging`. The other 8 language files already omit it | Remove the line (restores 16-file consistency) |
| D14 | `design/constitution.md` | 24, 72, 103, 122, 285, 289, 315, 323, 328, 339 | **HIGHER-RISK / SHOULD** — archived `expert-frontend` named as a LIVE design-pipeline agent across HARD clauses + phase diagram + Sprint Contract. Known design-pipeline tension predating consolidation | Carve-out note (pipeline `expert-frontend` resolves to `Agent(general-purpose)` with frontend whitelist per archived-agent-rejection §C) OR retained rename — see plan §F.D14 decision |
| D15 | `development/skill-authoring.md` | 102, 173 | NEW (not in original audit) — live `expert-backend` in YAML frontmatter `agents:` examples (`agents: ["expert-backend"]`, `agents: ["manager-spec", "expert-backend"]`) | Replace `expert-backend` with a retained agent in the example (e.g. `manager-develop`) |

### Out of Scope — Exclusions (What NOT to Build)

[HARD] The following are explicitly OUT OF SCOPE for this SPEC:

- **PRESERVE — canonical reference lists**: `workflow/archived-agent-rejection.md` (the migration table) and `NOTICE.md` (the archive provenance record) MUST NOT be touched. These ARE the authoritative archived-agent enumerations.
- **PRESERVE — intentional enumerations**: `development/agent-patterns.md:269,273` (the `manager-strategy → manager-develop` deprecation doc — it documents WHY the chain is impossible), `development/spec-frontmatter-schema.md:64` (the "archived agent names MUST NOT appear as owners" enumeration). These name archived agents on purpose, to document the prohibition.
- **PRESERVE — role_profiles ≠ archived agents**: `researcher`, `analyst`, `reviewer`, `implementer`, `tester`, `designer`, `architect` as `role_profile` values (e.g. `core/zone-registry.md:240`, `workflow/worktree-integration.md:83,166,210,236`) are valid `workflow.yaml` role profiles, NOT archived agent spawns. Do not scrub role_profile usage.
- **No team-file consolidation/deletion**: `agent-teams-pattern.md` full merge/deletion belongs to the cohort's SSOT-DEDUP-001 SPEC. Here, only the dead frontmatter `paths:` token (D6) is fixed.
- **No new lint rule / CI guard authoring**: this SPEC does not add a mechanical archived-name detector. Detection is by the grep-verifiable ACs in `acceptance.md`. (A future enforcement SPEC may add a lint rule.)
- **No Go code changes**: the orchestrator's `ARCHIVED_AGENT_REJECTED` logic in `internal/` is unchanged. This is a rules-content scrub only.
- **No `settings-management.md:32` MCP-usage rewrite**: that line documents the `pencil` MCP tool's historical consumers (`expert-frontend` sub-agent mode, `team-designer` team mode). Decision deferred to plan §F.D2-adjacent note — treated as a documentation-of-tooling reference, lower priority; if scrubbed, it is a SHOULD, not a HARD.

## §C GEARS Requirements

### Live-reference removal (HARD core)

- **REQ-RCS-001** (Ubiquitous): The rules catalog shall not present any of the 12 archived agents (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`-MoAI-custom, `researcher`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`) as a live spawn target, a live normative example, or a live hook action, in any file under `.claude/rules/` other than the two canonical reference files.

- **REQ-RCS-002** (Ubiquitous): The scrub shall preserve, verbatim and untouched, the archived-agent enumerations in `workflow/archived-agent-rejection.md` and `NOTICE.md`, and the intentional prohibition enumerations in `development/agent-patterns.md` and `development/spec-frontmatter-schema.md`.

- **REQ-RCS-003** (Ubiquitous): The scrub shall not alter any `role_profile` token (`researcher`, `analyst`, `reviewer`, `implementer`, `tester`, `designer`, `architect`); a `role_profile` is a valid `workflow.yaml` profile, not an archived agent spawn.

### Replacement fidelity

- **REQ-RCS-004** (Ubiquitous): Where a live archived reference is removed, the scrub shall substitute the canonical replacement per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C migration table — `Agent(general-purpose)` with a domain whitelist for implementation work, `Agent(Explore)` for read-only investigation, or a retained agent / role_profile.

- **REQ-RCS-005** (Event-driven): When the `agent-authoring.md` §Agent Categories section is rewritten (D10), the rewrite shall list exactly the 7 retained MoAI-custom agents plus the Anthropic built-in `Explore`, and shall be internally consistent with the file's own 8-agent ceiling section.

- **REQ-RCS-006** (Event-driven): When a recount of isolation-requiring agents is performed (D8), the count shall be re-derived against the 8-agent catalog and the resulting agent list shall name only retained agents and/or role_profiles.

### P0 CI-protocol live spawns

- **REQ-RCS-007** (Ubiquitous): The `ci-autofix-protocol.md` and `ci-watch-protocol.md` files shall not invoke `manager-quality` as a live diagnostic subagent; the diagnostic role shall be expressed as the Stop hook `sync-phase-quality-gate.sh` and/or a per-spawn `Agent(general-purpose)` diagnostic scope.

### Template-First mirror

- **REQ-RCS-008** (Ubiquitous): Every edit applied to a file under `.claude/rules/moai/` shall be applied identically to its mirror under `internal/template/templates/.claude/rules/moai/`, and `make build` shall be run to regenerate the embedded templates.

- **REQ-RCS-009** (Ubiquitous): The template-mirror edits shall preserve template neutrality per CLAUDE.local.md §25 — no internal SPEC IDs, internal dates, commit SHAs, REQ/AC tokens, or archive/memory paths shall be introduced into template content.

### Design-pipeline carve-out (SHOULD — higher risk)

- **REQ-RCS-010** (Where, capability gate): Where the `design/constitution.md` design pipeline (D14) and the mirrored `zone-registry.md` CONST entry (D3) name `expert-frontend` as a live design-pipeline agent, the scrub shall at minimum add a carve-out note that pipeline `expert-frontend` resolves to `Agent(general-purpose)` with a frontend whitelist per archived-agent-rejection §C, so the archived name is not silently load-bearing. A full retained-rename is permitted but not required; this requirement is SHOULD pending the §F.D14 design decision.

## §D Acceptance Criteria Matrix (summary — see acceptance.md)

| AC group | Binds REQ | Verification mode |
|----------|-----------|-------------------|
| AC-RCS-001..012, 015 | REQ-RCS-001/004/005/006/007 | Per-defect grep (archived name absent at the live-reference site) — covers D1-D13, D15 (per-defect MUST scrubs) |
| AC-RCS-013, AC-RCS-014 | REQ-RCS-010 | design/constitution (AC-RCS-013, D14) + zone-registry CONST-V3R2-064 (AC-RCS-014, D3) carry carve-out note OR retained rename (SHOULD) |
| AC-RCS-016 | REQ-RCS-002 | Grep confirms archived enumerations intact in the 4 PRESERVE files |
| AC-RCS-017 | REQ-RCS-003 | Grep confirms role_profile tokens unchanged |
| AC-RCS-018 | REQ-RCS-008 | Per-file mirror parity by enrollment split (explicit `diff -q` for non-enrolled byte-identical files; Go test for enrolled; leak-test + change-confinement for §25-sanitized) |
| AC-RCS-019 | REQ-RCS-009 | Template neutrality check passes (`TestTemplateNeutralityAudit`) |
| AC-RCS-020 | REQ-RCS-008 | `make build` + `go build ./...` succeed (template tree embedded via `//go:embed all:templates`; no `embedded.go`) |
| AC-RCS-021 | REQ-RCS-001 (global) | Catalog-wide: archived names remain ONLY in the 4 PRESERVE files + role_profile contexts |

## §E Cross-References

- `.claude/rules/moai/workflow/archived-agent-rejection.md` — §C migration table (canonical replacement source)
- `.claude/rules/moai/NOTICE.md` — archive provenance (PRESERVE)
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/` — the 17→8 consolidation origin
- `CLAUDE.local.md` §25 (template internal-content isolation) + §2 (Template-First) + §15 (language neutrality)
- `.moai/docs/template-internal-isolation-doctrine.md` — neutrality content-class catalogue
