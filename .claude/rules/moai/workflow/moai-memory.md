---
paths: "**/.moai/specs/**"
---

# MoAI Memory and Context

Rules for managing persistent context across sessions.

## Memory Hierarchy

Claude Code supports multiple memory levels (highest priority first):

1. Managed Policy: Organization-level rules (read-only)
2. Project Instructions: CLAUDE.md (checked into repo)
3. Project Rules: .claude/rules/**/*.md (auto-discovered, conditional via paths)
4. User Instructions: ~/.claude/CLAUDE.md (personal global)
5. Local Instructions: CLAUDE.local.md (personal project, not committed)
6. Auto Memory: ~/.claude/projects/{hash}/memory/ (AI-managed)

## SPEC Context Persistence

SPEC documents serve as persistent context for multi-session work:

- SPEC document: `.moai/specs/SPEC-XXX/spec.md` (requirements and design)
- Research artifact: `.moai/specs/SPEC-XXX/research.md` (codebase analysis)
- Progress tracking: Task list state via TaskCreate/TaskUpdate

## Session Continuity

When resuming work across sessions:
- Reference SPEC documents for requirements context
- Check git log for recent changes
- Read task list if team mode was active
- Use /clear between major phase transitions to free context

## Rules

- SPEC documents are the primary cross-session context mechanism
- Auto memory should store stable patterns, not session-specific state
- Maximum 5,000 tokens for injected context from previous sessions
- Prefer referencing files over copying content into context
