---
id: SPEC-CONTEXT-CLEANUP-001
title: Remove Git-Based Context Memory and Enhance SPEC Lifecycle
version: "1.0.0"
status: Draft
phase: "Plan"
module: "template system (skills, workflows, agents, rules)"
dependencies: []
lifecycle: spec-anchored
---

# SPEC-CONTEXT-CLEANUP-001: Remove Git-Based Context Memory and Enhance SPEC Lifecycle

## 1. Environment

### 1.1 Project Context

moai-adk-go is a Go-based ADK (Agent Development Kit) that generates Claude Code project templates. Template source files live in `internal/template/templates/` and are deployed to user projects via `moai init` / `moai update`.

### 1.2 Problem Statement

The Git-Based Context Memory system (`/moai context`, `/moai memory`) was designed to provide cross-session continuity by parsing structured `## Context` sections from git commit messages. However:

1. **Severe overlap with SPEC**: 90%+ of context memory categories (Decision, Constraint, Pattern, Risk) already exist in SPEC documents
2. **Write-only system**: manager-git writes `## Context` sections into every commit, but no evidence exists of `/moai memory` ever being invoked to read them back
3. **Fragile architecture**: Relies on regex parsing of git commit messages with 80% similarity deduplication — error-prone compared to directly reading SPEC files
4. **Token waste**: ~10,000+ tokens of system overhead (workflow definitions, agent instructions, config files) for zero actual usage
5. **Single Source of Truth violation**: Same decisions recorded in both SPEC spec.md and git commit messages

### 1.3 Target State

- SPEC documents become the **sole** cross-session context mechanism
- SPEC spec.md gains an **Implementation Log** section for runtime discoveries (Gotchas, decisions made during implementation)
- All git-based context memory infrastructure is removed
- Token budget management (moai-foundation-context) is preserved — it serves a different purpose
- Commit messages return to clean conventional format without `## Context` bloat

## 2. Assumptions

### 2.1 Technical Assumptions

- A1: moai-foundation-context skill is primarily about token budget management, not git-based context memory. It will be preserved with minor cleanup.
- A2: All template changes require `make build` to regenerate `embedded.go`.
- A3: Both local (`.claude/`) and template source (`internal/template/templates/.claude/`) must be modified in sync.
- A4: Session boundary git tags (`moai/SPEC-{ID}/{phase}-complete`) are lightweight and useful for phase tracking — they will be PRESERVED.
- A5: The 8 agents referencing moai-foundation-context in their skills list do NOT need modification (the skill remains).

### 2.2 Business Assumptions

- B1: No users depend on `/moai context` or `/moai memory` commands (zero usage evidence).
- B2: Simplifying the system reduces onboarding friction and maintenance burden.
- B3: SPEC documents are already the primary mechanism users rely on for cross-session work.

## 3. Requirements

### REQ-CTX-001: Delete Context Memory Workflow

**When** the context memory system is removed,
**the system shall** delete the following files:
- `.claude/skills/moai/workflows/context.md`
- `.moai/config/sections/context.yaml`

Both local and template source copies.

**Acceptance**: Files no longer exist in project or template directory.

### REQ-CTX-002: Remove Context Routing from SKILL.md

**When** the `/moai` skill processes user input,
**the system shall** no longer recognize `context`, `ctx`, or `memory` as valid subcommands.

**Acceptance**: The context/ctx/memory alias, routing logic, and workflow reference are removed from `.claude/skills/moai/SKILL.md`.

### REQ-CTX-003: Remove Context Search Protocol from CLAUDE.md

**When** CLAUDE.md is loaded,
**the system shall** no longer contain Section 16 (Context Search Protocol).

**Acceptance**: Section 16 is removed. Subsequent sections are renumbered.

### REQ-CTX-004: Remove Context Memory Generation from Sync Workflow

**When** the sync workflow executes,
**the system shall** no longer generate `## Context (AI-Developer Memory)` sections in commit messages (Step 3.1.1).

**Acceptance**: sync.md Step 3.1.1 is removed. Commit messages use clean conventional format.

### REQ-CTX-005: Simplify Manager-Git Commit Format

**When** manager-git creates commits,
**the system shall** use standard conventional commit format without `## Context` sections or structured `- Decision:` / `- Constraint:` lines.

**Acceptance**: manager-git.md no longer contains context memory commit format instructions.

### REQ-CTX-006: Clean Up Cross-References

**When** any skill, rule, or documentation file references context memory,
**the system shall** remove or update the reference.

Files to update:
- `.claude/rules/moai/workflow/moai-memory.md` — simplify to SPEC-as-context only
- `.claude/skills/moai/workflows/plan.md` — remove Context Loading section
- `.claude/skills/moai/workflows/run.md` — remove Context Loading + Context Propagation
- `.claude/skills/moai-foundation-context/SKILL.md` — remove "memory" from trigger keywords
- `.claude/skills/moai-foundation-core/SKILL.md` — remove context redirect reference
- `.claude/skills/moai-foundation-core/modules/agents-reference.md` — update skill list description
- `.claude/skills/moai-foundation-core/modules/commands-reference.md` — remove context command
- `.claude/skills/moai-foundation-core/modules/execution-rules.md` — update reference
- `README.md` / `README.ko.md` — remove Context Memory System section

**Acceptance**: `grep -ri "context memory\|moai-workflow-context\|/moai context\|/moai memory\|/moai ctx\|Context Search Protocol\|## Context" --include="*.md" --include="*.yaml"` returns zero results in workflow/skill/agent/rule files. (Excluding this SPEC itself and CHANGELOG.)

### REQ-CTX-007: Add Implementation Log to SPEC Template

**When** a SPEC document is created via `/moai plan`,
**the system shall** include a `## 5. Implementation Log` section in spec.md.

Format:
```markdown
## 5. Implementation Log

<!-- Updated during /moai run phase. Records runtime discoveries. -->

### Decisions
<!-- Key technical decisions made during implementation -->

### Gotchas
<!-- Pitfalls and warnings discovered during implementation -->

### Risks
<!-- Open risks and deferred items -->
```

**Acceptance**: New SPECs created via plan workflow include the Implementation Log section.

### REQ-CTX-008: Update Run Workflow for SPEC Implementation Log

**When** the run workflow discovers important implementation details (decisions, gotchas, risks),
**the system shall** record them in the SPEC's Implementation Log section.

**Acceptance**: run.md includes guidance to update `spec.md` Implementation Log during implementation.

### REQ-CTX-009: Preserve Token Budget Management

**When** the context memory system is removed,
**the system shall** preserve moai-foundation-context skill intact (except removing "memory" from trigger keywords).

**Acceptance**: moai-foundation-context SKILL.md continues to function for token budget management. All 8 agent skill references remain valid.

### REQ-CTX-010: Preserve Session Boundary Git Tags

**When** the context memory system is removed,
**the system shall** preserve session boundary git tag functionality (`moai/SPEC-{ID}/{phase}-complete`).

**Acceptance**: sync.md retains git tag creation for phase transitions. Only the `## Context` commit message content is removed.

## 4. Implementation Plan

### Task 1: Delete context memory files
- Delete `context.md` workflow (local + template)
- Delete `context.yaml` config (local + template)

### Task 2: Clean SKILL.md routing
- Remove context/ctx/memory alias from SKILL.md (local + template)

### Task 3: Clean CLAUDE.md
- Remove Section 16 (Context Search Protocol)
- Renumber subsequent sections (local + template)

### Task 4: Clean sync.md
- Remove Step 3.1.1 (Context Memory Generation)
- Preserve git tag functionality (local + template)

### Task 5: Clean manager-git.md
- Remove `## Context` commit format instructions (local + template)

### Task 6: Clean cross-references
- Update all files listed in REQ-CTX-006 (local + template)

### Task 7: Enhance SPEC template
- Add Implementation Log section to plan.md workflow output (local + template)
- Update manager-spec agent guidance

### Task 8: Update run.md
- Add Implementation Log recording guidance (local + template)

### Task 9: Update moai-memory.md rule
- Simplify to SPEC-only context rules (local + template)

### Task 10: Update documentation
- README.md / README.ko.md context memory sections (local only — not in template)

### Task 11: Build and verify
- Run `make build` to regenerate embedded.go
- Run `go test ./...` to verify no breakage
- Run grep verification per REQ-CTX-006 acceptance criteria

## 5. Implementation Log

<!-- Updated during /moai run phase. Records runtime discoveries. -->

### Decisions
<!-- Key technical decisions made during implementation -->

### Gotchas
<!-- Pitfalls and warnings discovered during implementation -->

### Risks
<!-- Open risks and deferred items -->
