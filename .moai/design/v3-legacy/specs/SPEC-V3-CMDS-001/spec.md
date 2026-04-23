---
id: SPEC-V3-CMDS-001
title: "Command Frontmatter Extensions"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 3 Agent Runtime v2"
module: "internal/template/frontmatter_parser.go, internal/command/loader.go"
dependencies:
  - SPEC-V3-SCH-001
  - SPEC-V3-SKL-001
related_gap:
  - gm#87
  - gm#88
  - gm#89
  - gm#90
  - gm#91
  - gm#102
related_theme: "Theme 10: Command Extension Parity"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "command, frontmatter, v3, context-fork, paths"
---

# SPEC-V3-CMDS-001: Command Frontmatter Extensions

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

moai's slash commands use a minimal frontmatter surface (description, argument-hint, allowed-tools). Claude Code's command/skill frontmatter schema is richer, supporting per-command `model` selection, `disable-model-invocation` (hide from SkillTool), structured `argument-hint` arrays (progressive typeahead), `context: inline | fork` for sub-agent-isolated execution, `paths:` conditional activation globs, and `skills:` array for preloading skill context. Adopting this schema brings moai into parity with CC's command surface and unlocks cost-efficient command dispatch (e.g., Haiku-per-command, fork-isolated linters, path-conditional formatters).

Per CLAUDE.local.md §2 "Thin Command Pattern", moai's 13 commands remain 20-LOC routers — this SPEC expands what the frontmatter CAN declare, not what the body must contain.

## 2. Scope (범위)

In-scope:
- `model: inherit | opus | sonnet | haiku | <specific-id>` per-command model override.
- `disable-model-invocation: boolean` hides the command from the SkillTool routing layer (command is only user-invocable via slash).
- `argument-hint: string | string[]` — accept array form for progressive typeahead (`["source-file", "target-file"]` shows gray hint as user types).
- `context: inline | fork` — inline expands body into current conversation (default, v2 behavior); fork runs command as sub-agent with fresh context.
- `agent: <agent-type>` — sub-agent type when `context: fork` (e.g., `general-purpose`, `expert-frontend`).
- `paths: string | string[]` — conditional activation globs (e.g., `**/*.py`, `src/*.{ts,tsx}` with brace expansion).
- `skills: string[]` — YAML array (existing partial moai support) for preloading skill context (upstream CC accepts comma-string AND YAML array; moai standardizes on YAML array for validator-friendliness per CLAUDE.local.md §12).
- `effort: low | medium | high | xhigh | max | <integer>` per-command Opus 4.7 effort override.
- Validator/v10 schema at `internal/config/schema/command_schema.go`.
- Frontmatter parser upgrade to handle brace expansion in `paths` and YAML array/string dual form for `allowed-tools`.

Out-of-scope:
- `$ARGUMENTS[N]` and named-arg `$name` substitution → SPEC-V3-CMDS-002.
- `!`cmd` inline bash execution → SPEC-V3-CMDS-003.
- Dynamic skill discovery (walk-up from touched files) → SPEC-V3-SKL-002.
- `hooks:` in command frontmatter → covered by SPEC-V3-HOOKS-006 source tier "skillHook" (deferred to v3.2).
- `isSensitive` arg redaction, `immediate: true` bypass queue, `version:` — deferred beyond v3.0.

## 3. Environment (환경)

Current moai-adk state:
- Command frontmatter fields currently enforced: `description`, `argument-hint` (string only), `allowed-tools` (CSV string). Per CLAUDE.local.md §3 and `internal/template/commands_audit_test.go` (Thin Command Pattern), commands are thin routers.
- No `model` field; model is inherited from session settings (findings-wave1-moai-current.md §8 — command catalog).
- No `context: fork` primitive; every command expands inline.
- No `paths:` conditional activation for commands.
- `skills:` accepted in agent frontmatter (partial) but not command frontmatter.

Claude Code reference:
- `utils/frontmatterParser.ts:10-59` — full frontmatter schema (findings-wave1-hooks-commands.md §7).
- `skills/loadSkillsDir.ts:185-264` — command/skill loader consumes schema (findings-wave1-hooks-commands.md §7).
- `types/command.ts:25-57` — PromptCommand fields including hooks/skillRoot/context/agent/paths/effort (findings-wave1-hooks-commands.md §11).
- `utils/frontmatterParser.ts:85-121` — YAML special-char auto-quoting for glob patterns (findings-wave1-hooks-commands.md §7).
- `utils/frontmatterParser.ts:189-232` — `paths` split + brace expansion (findings-wave1-hooks-commands.md §7).
- `skills/loadSkillsDir.ts:997-1040` — conditional skill activation via paths (findings-wave1-hooks-commands.md §14.5).

Affected modules:
- `internal/template/frontmatter_parser.go` — expand parser with validator/v10 struct tags.
- `internal/command/loader.go` — new file: command loader with conditional activation and context:fork dispatch.
- `internal/config/schema/command_schema.go` — validator schema.
- `internal/template/commands_audit_test.go` — extend to validate new fields.

## 4. Assumptions (가정)

- Brace expansion semantics match CC (`src/*.{ts,tsx}` → two patterns `src/*.ts`, `src/*.tsx`).
- Glob matching uses a gitignore-style filter (`doublestar` library or equivalent; no new dep if stdlib `filepath.Match` suffices for simple cases with user-opt-in for advanced patterns).
- `context: fork` dispatcher reuses existing Agent() subagent spawn machinery; `agent:` field maps to `subagent_type` parameter.
- `skills:` YAML array is authoritative moai form; CC's comma-string form is accepted for compatibility but converted to array at load time.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-CMDS-001-001: The command frontmatter schema SHALL accept optional fields: `model`, `disable-model-invocation`, `argument-hint` (string or array), `context`, `agent`, `paths`, `skills`, `effort`.
- REQ-CMDS-001-002: The `model` field SHALL accept values in the set `{inherit, opus, sonnet, haiku}` OR a specific model ID string.
- REQ-CMDS-001-003: The `context` field SHALL default to `inline` when omitted.
- REQ-CMDS-001-004: The `agent` field SHALL be required when `context: fork` and SHALL be rejected by the loader when present with `context: inline`.
- REQ-CMDS-001-005: The `paths` field SHALL accept a string, comma-separated list, or YAML array and SHALL support brace expansion (`{ts,tsx}` expands to two patterns).
- REQ-CMDS-001-006: The `skills` field SHALL be a YAML array (moai convention per CLAUDE.local.md §12); loader SHALL accept legacy comma-string form and convert at load time.
- REQ-CMDS-001-007: The `effort` field SHALL accept values in the set `{low, medium, high, xhigh, max}` or a non-negative integer.
- REQ-CMDS-001-008: The frontmatter parser SHALL auto-quote YAML-special-character values (e.g., glob patterns containing `*`, `{`, `}`) to preserve parseability.

### 5.2 Event-driven Requirements

- REQ-CMDS-001-010: WHEN a command is user-invoked via slash input, the loader SHALL apply the resolved `model` and `effort` fields to the underlying Claude Code invocation.
- REQ-CMDS-001-011: WHEN `disable-model-invocation: true`, the command SHALL be hidden from the SkillTool (model-routable) list while remaining accessible via slash.
- REQ-CMDS-001-012: WHEN `context: fork` is set and the command is invoked, the dispatcher SHALL spawn a sub-agent of type `agent` with fresh context and route the command body into the sub-agent's initial prompt.
- REQ-CMDS-001-013: WHEN `paths:` is set, the command SHALL only be surfaced to the SkillTool when the current session has touched at least one file matching a glob.
- REQ-CMDS-001-014: WHEN the frontmatter contains an unknown field, the parser SHALL warn (not error) and preserve backward compatibility.

### 5.3 State-driven Requirements

- REQ-CMDS-001-020: WHILE `argument-hint` is an array of N strings and the user has typed K < N arguments, the typeahead SHALL display `<hint[K]> <hint[K+1]> ...` progressively.
- REQ-CMDS-001-021: WHILE `paths` is set and the session has touched zero matching files, the command SHALL NOT appear in the SkillTool list but SHALL remain user-invocable.

### 5.4 Optional Features

- REQ-CMDS-001-030: WHERE the configuration key `command.paths.strict_glob: true` is set, the loader SHALL refuse to load commands whose `paths` contains patterns not parseable by the glob engine.
- REQ-CMDS-001-031: WHERE a command has both `context: fork` and `skills: [...]`, the sub-agent SHALL preload the listed skills into its initial context.

### 5.5 Complex Requirements

- REQ-CMDS-001-040: IF `model: inherit` AND the session's active model is unavailable, THEN the loader SHALL emit a warn log and fall back to the session default; ELSE the explicit model is used verbatim.
- REQ-CMDS-001-041: IF `context: fork` AND `effort: <override>` are both set, THEN the fork sub-agent SHALL inherit the command's `effort` value (not the parent session's); ELSE the session default applies.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-CMDS-001-01: Given a command with `model: haiku`, When invoked, Then the Claude Code call uses Haiku-class model. (maps REQ-CMDS-001-002, REQ-CMDS-001-010)
- AC-CMDS-001-02: Given a command with `disable-model-invocation: true`, When the SkillTool enumerates available commands, Then this command is omitted. (maps REQ-CMDS-001-011)
- AC-CMDS-001-03: Given a command with `argument-hint: ["source", "target"]` and user has typed `/cmd src.go `, When typeahead renders, Then the gray hint shows `target`. (maps REQ-CMDS-001-020)
- AC-CMDS-001-04: Given a command with `context: fork`, `agent: general-purpose`, When invoked, Then a sub-agent of type `general-purpose` is spawned with fresh context and the command body is routed as initial prompt. (maps REQ-CMDS-001-012)
- AC-CMDS-001-05: Given a command with `context: inline` AND `agent: anything`, When loaded, Then the loader rejects the command with `CommandInvalidContextAgent`. (maps REQ-CMDS-001-004)
- AC-CMDS-001-06: Given a command with `paths: "**/*.py"` and session has touched zero `.py` files, When SkillTool enumerates, Then the command is omitted from the list but the user can still invoke it via slash. (maps REQ-CMDS-001-013, REQ-CMDS-001-021)
- AC-CMDS-001-07: Given a command with `paths: "src/*.{ts,tsx}"`, When the loader parses, Then the paths list expands to `["src/*.ts", "src/*.tsx"]`. (maps REQ-CMDS-001-005, REQ-CMDS-001-008)
- AC-CMDS-001-08: Given a command with `skills: "moai-foundation-core, moai-workflow-spec"` (legacy comma form), When loaded, Then the value is converted to `["moai-foundation-core", "moai-workflow-spec"]` array. (maps REQ-CMDS-001-006)
- AC-CMDS-001-09: Given a command with `effort: 42`, When loaded, Then the integer value is accepted and passed through to the invocation. (maps REQ-CMDS-001-007)
- AC-CMDS-001-10: Given a command with an unknown field `foo: bar`, When loaded, Then a warn log is emitted and the command still loads. (maps REQ-CMDS-001-014)

## 7. Constraints (제약)

- Technical: Go 1.22+. Frontmatter parsed via existing `gopkg.in/yaml.v3`; validator/v10 from SPEC-V3-SCH-001.
- Backward compat: All new fields optional; existing 13 moai commands load unchanged. Thin Command Pattern enforcement (`internal/template/commands_audit_test.go`) preserved — new fields may appear in frontmatter but command body remains ≤ 20 LOC.
- Platform: Glob matching via stdlib `path/filepath.Match` with brace expansion handled at parse time.
- Template-first: New fields go to `internal/template/templates/.claude/commands/` first, then `make build`, then deploy (CLAUDE.local.md §2).
- Language-neutral: Templates apply equally to all 16 supported languages (CLAUDE.local.md §15).

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Users misuse `context: fork` for commands that need current conversation state | M | M | Docs-site migration guide explains fork vs inline; `moai doctor command --validate` flags suspicious fork commands (e.g., those referencing unscoped $ARGUMENTS). |
| `paths` glob DoS with catastrophic patterns | L | M | `command.paths.strict_glob: true` rejects unparseable patterns; stdlib `filepath.Match` has bounded complexity (no regex). |
| `skills: [...]` preload increases cost for simple commands | L | M | Docs-site emphasizes preload sparingly; `moai doctor command --cost-estimate` surfaces preload impact. |
| `effort` field drift between commands and agents | L | L | Same enum used for SPEC-V3-AGT-001 agent frontmatter; validator shares the `EffortLevel` type. |
| `disable-model-invocation` causes silent user confusion when slash-typing lists commands | L | L | Slash-list unchanged (command still appears); only SkillTool hides. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-SCH-001 (validator/v10 for frontmatter validation).
- SPEC-V3-SKL-001 (shared `paths:` and `effort:` definitions between skills and commands).

### 9.2 Blocks

- SPEC-V3-CMDS-002 (argument substitution consumes `argument-hint` array form defined here).
- SPEC-V3-CMDS-003 (inline bash execution runs inside commands loaded with this schema).

### 9.3 Related

- SPEC-V3-AGT-001 (agent frontmatter shares `model`, `effort`, `skills` field definitions).
- SPEC-V3-PLG-001 (plugin-shipped commands adhere to this schema).

## 10. Traceability (추적성)

- Theme: master-v3 §8.4 Skill SPECs (command+skill frontmatter share the v2 bundle).
- Gap rows: gm#87 (High — context: inline|fork), gm#88 (High — agent: frontmatter), gm#89 (High — paths conditional activation), gm#90 (Medium — effort per-skill), gm#91 (Low — disableModelInvocation), gm#102 (Low — progressive argument hint).
- BC-ID: None (purely additive).
- Wave 1 sources: findings-wave1-hooks-commands.md §7 (frontmatter schema), §11 (PromptCommand fields), §14.5 (command feature gap table), §12 (source references `utils/frontmatterParser.ts:10-59, 85-121, 189-232`).
- Priority: P1 High (command schema parity with CC; ships Phase 3).
