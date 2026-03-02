---
paths: "**/.moai/specs/**"
---

# MoAI SPEC Context and Session Continuity

Rules for managing cross-session context through SPEC documents and Claude Code memory hierarchy.

## Memory Hierarchy

Claude Code supports multiple memory levels (highest priority first):

1. Managed Policy: Organization-level rules (read-only)
2. Project Instructions: CLAUDE.md (checked into repo)
3. Project Rules: .claude/rules/**/*.md (auto-discovered, conditional via paths)
4. User Instructions: ~/.claude/CLAUDE.md (personal global)
5. Local Instructions: CLAUDE.local.md (personal project, not committed)
6. Auto Memory: ~/.claude/projects/{hash}/memory/ (AI-managed)

## SPEC as Primary Cross-Session Context

SPEC documents are the primary mechanism for maintaining context across sessions:

- SPEC document: `.moai/specs/SPEC-XXX/spec.md` (requirements and design)
- Research artifact: `.moai/specs/SPEC-XXX/research.md` (codebase analysis)
- Progress tracking: Task list state via TaskCreate/TaskUpdate

SPEC documents persist between sessions and provide complete context for resuming work without relying on conversation history.

## Session Continuity via SPEC Documents

When resuming work across sessions:
- Reference the relevant SPEC document for requirements context
- Read research.md for codebase analysis from the plan phase
- Check git log for recent implementation changes
- Read task list if team mode was active
- Use /clear between major phase transitions to free context

## Rules

- SPEC documents are the single source of truth for cross-session context
- All requirements, design decisions, and acceptance criteria live in SPEC documents
- Auto memory should store stable patterns, not session-specific state
- Maximum 5,000 tokens for injected context from previous sessions
- Prefer referencing SPEC files over copying content into context
- When context is insufficient, re-read the SPEC document rather than relying on conversation history
