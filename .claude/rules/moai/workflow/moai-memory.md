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

---

## Agent Memory Taxonomy (SPEC-V3R2-EXT-001)

All memory files written to `.claude/agent-memory/<agent-name>/` MUST conform to the 4-type taxonomy.

### Required Frontmatter

Every memory file MUST have a YAML frontmatter block with all three required fields:

```markdown
---
name: <short descriptive name>
description: <one-line description used for relevance matching>
type: <user | feedback | project | reference>
---

<memory body>
```

### The 4 Types

| Type | When to Use | Body Structure |
|------|-------------|----------------|
| `user` | User role, preferences, knowledge, working style | Free prose |
| `feedback` | Corrections, validated approaches, behavioral guidance | Lead with rule; add **Why:** and **How to apply:** sub-lines |
| `project` | Ongoing work, goals, decisions, incidents | Lead with fact/decision; add **Why:** and **How to apply:** sub-lines |
| `reference` | Pointers to external resources (Linear, Grafana, etc.) | Free prose with URL |

The type set is immutable — no types beyond these four are accepted.

### Body Structure Requirements

`feedback` and `project` memory files MUST include:

```markdown
<rule or fact statement>

**Why:** <reason — often a past incident or constraint>
**How to apply:** <when/where this guidance kicks in>
```

Files of type `user` and `reference` do not require this structure.

### MEMORY.md Line Cap

`MEMORY.md` (the index file) must not exceed **200 lines**. Lines 201+ are silently truncated by the Claude Code memory loader. Keep each entry to a single line under 150 characters.

### Excluded Categories

The following content MUST NOT be stored in memory files:

| Category | Examples |
|----------|----------|
| Code patterns / conventions | Architecture diagrams, file path conventions already in the codebase |
| Git history | `git log` output, who changed what |
| Debug recipes | Step-by-step fix instructions already captured in the fix commit |
| CLAUDE.md mirrors | Anything already documented in CLAUDE.md or `.claude/rules/` |
| Ephemeral state | In-progress task lists, current session context |

Use the `MOAI_MEMORY_AUDIT=0` environment variable to temporarily disable taxonomy enforcement during bulk memory migrations.

### Staleness Caveat

Memory files older than **24 hours** (mtime) are considered potentially stale. At session start, the SessionStart hook wraps stale files in a `<system-reminder>` block to signal that the content should be verified before acting on it.

When 10 or more stale files are detected simultaneously, a single aggregated warning is emitted instead of per-file wrappers to avoid token bloat.

### Audit Warnings

The PostToolUse hook audits memory files on Write/Edit operations and emits non-blocking warnings to stderr for:

| Code | Condition |
|------|-----------|
| `MEMORY_MISSING_TYPE` | File has no `type` field in frontmatter |
| `MEMORY_MISSING_FRONTMATTER` | File missing `name` or `description` |
| `MEMORY_BODY_STRUCTURE_MISSING` | feedback/project file missing **Why:** or **How to apply:** |
| `MEMORY_EXCLUDED_CATEGORY` | Body matches an excluded-category keyword pattern |
| `MEMORY_INDEX_OVERFLOW` | MEMORY.md exceeds 200 lines |
| `MEMORY_DUPLICATE` | Two files share the same description |

All warnings are observation-only. Hook exit code is always 0 (non-blocking).
