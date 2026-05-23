---
id: SPEC-V3R6-SEQ-THINKING-RETIRE-001
title: "Sequential-Thinking MCP Retirement + Ultrathink Consolidation"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.6.0"
module: ".claude/agents,.claude/skills,.claude/rules,internal/template/templates,.mcp.json,settings.json"
lifecycle: spec-anchored
tags: "v3r6, mcp, retirement, ultrathink, adaptive-thinking, sequential-thinking, sprint-2, namespace-cleanup, ci-guard"
tier: M
---

# SPEC-V3R6-SEQ-THINKING-RETIRE-001 — Sequential-Thinking MCP Retirement + Ultrathink Consolidation

## §1. Purpose and Policy Origin

User policy 2026-05-23 (GOOS): *"agent/skills에서 sequential-thinking mcp를 사용하지말고 ultrathink를 사용해서 깊은 추론이 필요할 때 호출을 하고 sequential-thinking는 모두 다 제거를 하자."*

This SPEC executes complete retirement of `sequential-thinking` MCP server from the MoAI-ADK project namespace and consolidates all deep-reasoning paths onto the `ultrathink` keyword (which triggers Adaptive Thinking on Opus 4.7+ via Claude Code's `effort: max` mechanism).

`ultrathink` is a **keyword in user/orchestrator prompts** — not an MCP tool — and therefore needs no `allowed-tools:` registration in agent or skill frontmatter. Sequential-thinking MCP, by contrast, ships as a separate external server (`@modelcontextprotocol/server-sequential-thinking`) and currently appears in 48 source files plus 3 config/template files.

## §2. Background and Rationale

| Dimension | Sequential-Thinking MCP | Ultrathink (Adaptive Thinking) |
|-----------|-------------------------|-------------------------------|
| Mechanism | External MCP server, `sequentialthinking` tool call per thought | Built-in Opus 4.7 reasoning, dynamic token allocation |
| Latency | Multi-round tool calls (round-trip per thought) | Single inference pass, internal |
| GLM compatibility | YES (any model) | YES (Opus 4.7+) — degrades to extended-thinking on older models |
| Cost | npx install + per-call MCP overhead | Built-in, no extra startup |
| Determinism | Structured branching/revision tools | Model-internal, less inspectable |
| Tool count | +1 deferred tool in 14 agents and 6 skills | 0 — keyword only |

Operating both paths produces:
- Redundant frontmatter entries (`mcp__sequential-thinking__sequentialthinking` in 28 agent files)
- Cognitive load (when to use sequential-thinking vs ultrathink unclear)
- Maintenance burden (separate MCP server lifecycle, allow-list in settings, enabledMcpjsonServers)
- Token waste (deferred tool schema preload + tool call round-trips)

The user policy directs single-path operation: ultrathink for deep reasoning, sequential-thinking removed from the project namespace. Users who want sequential-thinking for personal use install it independently via `~/.claude/settings.json` (user-local scope).

## §3. EARS Requirements

### REQ-STR-001 [Ubiquitous] — Namespace cleanliness

The MoAI-ADK project namespace **shall not** contain any reference to `sequential-thinking` or `sequentialthinking` tokens in source files under `.claude/`, `internal/template/templates/`, `.mcp.json`, `.claude/settings.json`, or root documentation files (`CLAUDE.md`, `CLAUDE.local.md`, `README.md`, `README.ko.md`).

### REQ-STR-002 [Ubiquitous] — Ultrathink as canonical deep-reasoning path

Agent prompts, orchestrator delegation, and skill body guidance **shall** use the `ultrathink` keyword (Adaptive Thinking trigger on Opus 4.7+) as the canonical mechanism for invoking deep reasoning. References to sequential-thinking as an alternative path **shall** be removed or marked retired (BC marker).

### REQ-STR-003 [Ubiquitous] — `moai-foundation-thinking` skill redesign

The `moai-foundation-thinking` skill body **shall** describe only the three creative frameworks (Critical Evaluation, Diverge-Converge, Deep Questioning), the absorbed First Principles content (philosopher), and Adaptive Thinking + ultrathink invocation patterns. The "Sequential Thinking MCP (absorbed from moai-workflow-thinking)" subsection **shall** be marked retired with a BC marker noting that ultrathink supersedes sequential-thinking for deep reasoning.

### REQ-STR-004 [Event-Driven] — CI guard activation

**When** the CI test suite runs the template audit batch, a new guard test (`TestSeqThinkingRetired`) **shall** scan `.claude/` and `internal/template/templates/.claude/` plus the configuration files listed in REQ-STR-001 and **shall** fail with sentinel `SEQ_THINKING_REINTRODUCED` if any `sequential-thinking` or `sequentialthinking` token is detected.

### REQ-STR-005 [Unwanted] — Reintroduction prevention

**If** a future commit reintroduces any `mcp__sequential-thinking__*` tool reference into agent frontmatter, skill frontmatter, settings allow-list, `.mcp.json` server entry, `enabledMcpjsonServers` array, or rule text within the project namespace, **then** the CI guard from REQ-STR-004 **shall** fail and block merge.

### REQ-STR-006 [State-Driven] — In-flight preservation

**While** the retirement run-phase (manager-develop M2–M6) is in progress, existing `ultrathink` references in agent prompts, skill bodies, and rule documents **shall** be preserved unchanged. New `ultrathink` references **shall** be added only at the specific replacement sites where sequential-thinking was previously invoked, and only when the surrounding prose semantically requires a deep-reasoning directive.

### REQ-STR-007 [Optional] — User-local opt-in (out of project scope)

**Where** a user explicitly requires Sequential Thinking MCP for personal workflow, they **may** install it via their own `~/.claude/settings.json` (user-local scope) without re-introducing it to the project namespace. This SPEC does not document or guide such installation — it only ensures the project namespace remains clean.

### REQ-STR-008 [Ubiquitous] — Template-First Rule compliance

Every change to `.claude/` source files **shall** have an equivalent change to the corresponding `internal/template/templates/.claude/` mirror file, per CLAUDE.local.md §2 [HARD] Template-First Rule. The pairing **shall** be byte-equivalent except for paths that are intentionally divergent (e.g., template variables like `{{.GoBinPath}}` are preserved).

## §4. Acceptance Criteria Summary

12 binary acceptance criteria verify the requirements above. Each criterion is verifiable by a single shell command that produces a deterministic exit code. See `acceptance.md` for the complete Given/When/Then matrix and per-AC verification commands.

Header inventory:

- AC-STR-001: Project-namespace grep for sequential-thinking literals returns zero matches across agents, skills, rules, templates
- AC-STR-002: `.mcp.json` and `.mcp.json.tmpl` no longer contain sequential-thinking server entry
- AC-STR-003: `settings.json` and `settings.json.tmpl` no longer contain `mcp__sequential-thinking__*` allow-list or `enabledMcpjsonServers` array entry
- AC-STR-004: `CLAUDE.md` is updated with retirement note in §12, `CLAUDE.local.md` §22.2 mention is corrected
- AC-STR-005: `moai-foundation-thinking/SKILL.md` body contains zero `sequential-thinking` literal occurrences and one or more `ultrathink` or `Adaptive Thinking` references
- AC-STR-006: New CI guard test `TestSeqThinkingRetired` passes with sentinel `SEQ_THINKING_REINTRODUCED` defined
- AC-STR-007: `make build` exits 0; `go test ./internal/template/...` produces zero new regressions over the pre-SPEC baseline
- AC-STR-008: `catalog.yaml` hashes are regenerated for affected entries (14 agents + 6 skills)
- AC-STR-009: All 14 agent frontmatter `tools:` fields contain zero `mcp__sequential-thinking__sequentialthinking` tokens
- AC-STR-010: All 6 skill frontmatter `allowed-tools:` fields contain zero `mcp__sequential-thinking__sequentialthinking` tokens
- AC-STR-011: Cross-platform builds pass for darwin/amd64, linux/amd64, windows/amd64
- AC-STR-012: `CLAUDE.md` §12 contains a retirement statement (Sequential Thinking MCP retired; ultrathink is the canonical deep-reasoning path) and identifies the SPEC ID

Full Given/When/Then traceability for each AC: see `acceptance.md`.

## §5. Constraints

- **Tier**: M (3-artifact spec.md + plan.md + acceptance.md). 48 source files affected + 3 config files. Manageable in a single run-phase with manager-develop cycle_type=ddd.
- **Frontmatter**: Canonical 12-field schema per `.claude/rules/moai/development/spec-frontmatter-schema.md`. `tier: M` is an optional field permitted by the schema.
- **SPEC ID**: `SPEC-V3R6-SEQ-THINKING-RETIRE-001` (matches `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`).
- **plan-auditor PASS threshold**: 0.80 (Tier M baseline).
- **Template-First Rule**: All source changes have template mirror changes. Per REQ-STR-008.
- **manager-develop-prompt-template Section A–E REQUIRED**: Tier M Section A–E discipline (Section D safe-execution rules including B9 "no autonomous git pull/fetch/rebase" from Wave 3 lesson, Section E sentinel-key audit).
- **`moai-foundation-thinking` redesign**: Not a delete — a redesign. SKILL body's "Sequential Thinking MCP (absorbed from moai-workflow-thinking)" subsection is retired with BC marker preserving historical context but redirecting deep reasoning to ultrathink/Adaptive Thinking.
- **No new MCP servers**: ultrathink is a keyword, not an MCP. `allowed-tools:` registration is unnecessary and forbidden.
- **Agent prompt examples**: Where previously the prompt instructed "Use mcp__sequential-thinking__sequentialthinking for complex problems", the replacement guidance reads "When deep reasoning is required, the orchestrator may prepend `ultrathink` to the user's prompt (Opus 4.7+ Adaptive Thinking)".
- **Out of scope for run-phase**: Documenting user-local Sequential Thinking installation, adding new Adaptive Thinking features, retiring other MCP servers (context7, chrome-devtools, claude-in-chrome, zai-mcp-server are preserved).
- **No GitHub Issue required**: Internal infrastructure cleanup. May be created post-merge for audit-trail purposes; not blocking plan-phase.
- **Conventional Commits**: Run-phase commits use `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M<N> — <description>` format per `.moai/config/sections/git-convention.yaml`.

## §6. Affected Surface Inventory (validated 2026-05-23)

Inventory was confirmed by grep audit immediately before plan-phase authoring. Counts may shift by ±1–2 if files change between plan and run; verification commands in `plan.md` Milestone M1 produce authoritative counts at run-phase start.

| Surface | Files (local) | Files (template) | Total |
|---------|---------------|------------------|-------|
| Agents (frontmatter `tools:` field) | 14 | 14 | 28 |
| Skills (frontmatter `allowed-tools:` + body) | 6 | 6 | 12 |
| Rules (body text) | 1 | 1 | 2 |
| Configuration files | `.claude/settings.json`, `.claude/settings.local.json`, `.mcp.json` | `settings.json.tmpl`, `.mcp.json.tmpl` | 5 |
| Root documentation | `CLAUDE.local.md` §22.2, `CLAUDE.md` §12 (add retirement note) | n/a (CLAUDE.md is not templated) | 2 |
| New CI guard test | n/a | `internal/template/seq_thinking_retire_audit_test.go` (new file) | 1 |
| Total | 21 + ~7 supporting | 20 | **50** |

Agent file list (14): `manager-{develop,docs,project,quality,spec,strategy}`, `expert-{backend,devops,frontend,refactoring,security}`, `builder-harness`, `evaluator-active`, `plan-auditor`.

Skill file list (6): `moai-foundation-cc/reference/claude-code-settings-official.md`, `moai-foundation-cc/reference/skill-formatting-guide.md`, `moai-foundation-core/modules/agents-reference.md`, `moai-foundation-thinking/SKILL.md`, `moai/workflows/plan/clarity-interview.md`, `moai/workflows/run/context-loading.md`.

Rule file: `.claude/rules/moai/core/settings-management.md`.

## §7. Exclusions (What NOT to Build)

### §7.1 Out of Scope

- **User-local Sequential Thinking installation guide** — users who want it install themselves via `~/.claude/settings.json`. Not documented in this SPEC.
- **New Adaptive Thinking features** — ultrathink keyword behavior is unchanged. This SPEC is namespace cleanup + consolidation, not capability extension.
- **Retirement of other MCP servers** — `context7`, `chrome-devtools`, `claude-in-chrome`, `zai-mcp-server` are preserved. Each has its own use case unaffected by this SPEC.
- **Run-phase implementation** — manager-develop executes the cleanup in a separate commit/PR using the plan and acceptance criteria authored here. Plan-phase delivers only documentation artifacts.
- **`moai-foundation-thinking` deletion** — the skill remains active. Only the sequential-thinking subsection is retired; creative frameworks and First Principles content survive.
- **`moai-workflow-thinking` resurrection** — the absorbed origin skill stays absorbed. Sequential thinking content is retired from `moai-foundation-thinking`'s body, not relocated.
- **Backporting to v3.5 or earlier** — this is a v3.6 forward-only change. Past release branches retain sequential-thinking references.

## §8. Cross-References

- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field schema (REQ-SPC-003-006).
- `.claude/rules/moai/core/agent-common-protocol.md` — Subagent prohibitions and Parallel Execution patterns.
- `.claude/rules/moai/core/askuser-protocol.md` — Deferred tool preload reference for ToolSearch.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A–E Tier M discipline + B9 "no autonomous git pull/fetch/rebase" rule.
- `CLAUDE.md` §12 MCP Servers & Deep Analysis Modes — current canonical description of UltraThink, Adaptive Thinking, Sequential Thinking; needs retirement note added.
- `CLAUDE.local.md` §22.2 `enableAllProjectMcpServers` — current mention of `mcp__sequential-thinking` as alwaysLoad; needs correction.
- chore commit `d0782a365` (2026-05-23) — Round 1 namespace_protection_audit_test.go CI guard pattern (NAMESPACE_LEAK_* sentinel precedent; reused for SEQ_THINKING_REINTRODUCED sentinel).
- chore commit `4f1135684` (2026-05-23) — Template pollution cleanup precedent.

## §9. Risk and Mitigation Summary

Detailed risk register lives in `plan.md` §Risks. Headline risks:

- **R1 — Sibling SPEC overlap (V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 candidate)**: separate audit SPEC may also touch the same files for unrelated drift cleanup. Mitigation: this SPEC limits scope to sequential-thinking literals only; runs serial after any template-neutrality SPEC if scheduled.
- **R2 — Manager-develop autonomous git pull**: Wave 3 lesson `feedback_b9_no_autonomous_git_pull` documents recurrence pattern. Mitigation: B9 prohibition is now in `.claude/rules/moai/development/manager-develop-prompt-template.md` and is also recited verbatim in this SPEC's plan.md run-phase delegation prompt.
- **R3 — catalog.yaml hash drift between local and template**: 14 agents + 6 skills hash entries must be regenerated in M7 after all source/template edits land. Mitigation: M7 runs `gen-catalog-hashes.go --all` and verifies hash parity.
- **R4 — User session interrupted with `enabledMcpjsonServers: [..., "sequential-thinking", ...]` still present**: existing user sessions may have cached the old MCP server registration. Mitigation: settings.json change takes effect on next Claude Code session restart; no breaking effect during in-flight sessions.

## §10. Lifecycle Position

- **Sprint**: V3R6 Sprint 2 candidate (alongside AGENT-MODEL-ROUTING / PROMPT-CACHE / HOOK-OBSERVE-OPT-IN / HOOK-ASYNC-EXPAND per project_v3r6_sprint1_lane_a_complete memory).
- **Dependencies**: None blocking. Independent of V3R6 Sprint 1 Lane A (RULES-PATH-SCOPE / RULES-COMPRESS / SKILL-CONSOLIDATE / SKILL-COMPRESS — all already merged into main by `fa658d927`).
- **Blockers**: None.
- **Blocks**: None (consolidation-only cleanup, no downstream SPEC depends on retirement completion).
- **Lifecycle**: spec-anchored (SPEC maintained alongside implementation; quarterly review confirms namespace remains clean).

---

**Plan-Phase Deliverable Status**: spec.md authored. See `plan.md` for milestone breakdown and `acceptance.md` for Given/When/Then verification matrix.
