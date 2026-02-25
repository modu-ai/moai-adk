---
name: moai-workflow-context
description: >
  Git-based context memory system. Extracts AI-developer interaction context
  from git commit messages, reconstructs session history, and injects relevant
  context into new sessions for seamless continuity.
  Use when resuming work, checking previous decisions, or loading SPEC history.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-25"
  tags: "context, memory, git, session, resume, history"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4000

# MoAI Extension: Triggers
triggers:
  keywords: ["context", "memory", "resume", "history", "previous session", "what happened"]
  agents: ["manager-git"]
  phases: ["context"]
---

# Workflow: Context - Git-Based Context Memory

Purpose: Extract and reconstruct AI-developer interaction context from git commit history. Enables seamless session continuity by loading previous decisions, constraints, patterns, and risks.

Flow: Parse Arguments -> Extract Commits -> Parse Context -> Categorize -> Inject/Display

## Supported Flags

- --spec SPEC-XXX: Filter context to specific SPEC (default: auto-detect from current branch)
- --days N: Limit to commits within N days (default: 30)
- --category CAT: Filter by category (Decision, Constraint, Gotcha, Pattern, Risk, UserPref)
- --summary: Display compressed summary only (default: full context)
- --inject: Inject context into current session (default: display only)

## Phase 1: Commit Extraction

Extract commits with structured context from git log:

Git Query Strategy:
- If --spec flag: `git log --grep="SPEC: SPEC-XXX" --format="%H%n%s%n%b" --since="N days ago"`
- If on feature branch: Auto-detect SPEC from branch name (feature/SPEC-XXX)
- If no SPEC detected: Extract from all recent commits with Context sections

Commit Filtering:
- Include only commits with `## Context` section in body
- Parse SPEC, Phase, and Session metadata fields
- Sort by date (newest first)

## Phase 2: Context Parsing

Parse structured context from each commit message body:

Parsing Rules:
- Extract lines starting with `- Decision:`, `- Constraint:`, `- Gotcha:`, `- Pattern:`, `- Risk:`, `- UserPref:`
- Extract `## MX Tags Changed` section for tag history
- Extract `SPEC:` and `Phase:` metadata
- Associate each context item with its commit SHA and date

Output: Structured context map per category.

## Phase 3: Categorization and Deduplication

Organize extracted context:

Categories:
- **Decisions**: Technical choices with rationale (most recent takes precedence if conflicting)
- **Constraints**: Active constraints (accumulate, do not deduplicate)
- **Gotchas**: Pitfalls and warnings (accumulate with deduplication)
- **Patterns**: Reference implementations and patterns used (deduplicate by file path)
- **Risks**: Known risks and deferred items (track resolution status)
- **UserPrefs**: User preferences captured during sessions (most recent takes precedence)

Deduplication:
- Compare new items against existing items using string similarity
- If similarity > 80%: Keep the most recent version
- Track item evolution across commits (first seen -> last updated)

## Phase 4: Token Budget Management

Compress context to fit within token budget:

Budget Allocation (from context.yaml):
- max_injection_tokens: 5000 (default)
- skip_if_usage_above: 150000

Compression Strategy:
- Priority 1: Decisions and Constraints (most critical for continuity)
- Priority 2: Gotchas and Risks (prevent repeated mistakes)
- Priority 3: Patterns and UserPrefs (nice to have)
- If over budget: Truncate Priority 3 first, then summarize Priority 2

## Phase 5: Display or Inject

If --inject flag:
- Format context as structured markdown
- Present to user for confirmation before injection
- Inject approved context into session via AskUserQuestion

If display mode (default):
- Show formatted context report in user's conversation_language

Display Format:

```markdown
## Context Memory Report - SPEC-XXX

### Session Timeline
- 2026-02-20: Plan phase completed (3 decisions, 2 constraints)
- 2026-02-21: Run phase RED-GREEN (5 decisions, 3 gotchas)
- 2026-02-22: Run phase REFACTOR (2 patterns applied)

### Active Decisions (N)
| # | Decision | Rationale | Commit | Date |
|---|----------|-----------|--------|------|
| 1 | EdDSA over RSA256 | Performance priority (user request) | abc123 | 2026-02-21 |

### Active Constraints (N)
- API v1 backward compatibility required
- PostgreSQL 15+ only (no MySQL support)

### Known Gotchas (N)
- Redis TTL unreliable for RefreshToken storage
- Concurrent map access in Go requires sync.Mutex

### Applied Patterns (N)
- middleware chain pattern (auth.go:45)
- table-driven tests (auth_test.go)

### Open Risks (N)
- Token rotation deferred to Phase 2
- Rate limiting not yet implemented

### User Preferences
- Prefers functional style over OOP
- Wants verbose error messages

### MX Tag History
- 3 ANCHOR tags added, 2 TODO tags resolved
```

## Session Boundary Tags

When used with /moai sync or phase transitions, create git tags for session boundaries:

Tag Format: `moai/SPEC-{ID}/{phase}-complete`

Tag Message:
```
{Phase} phase completed
Decisions: N, Constraints: N, Risks: N
Next: /moai {next-phase} SPEC-{ID}
```

## Execution Summary

1. Parse arguments (extract flags: --spec, --days, --category, --summary, --inject)
2. Detect SPEC from branch name or flag
3. Extract commits with Context sections from git log
4. Parse structured context from commit bodies
5. Categorize and deduplicate context items
6. Apply token budget compression
7. Display report or inject into session
8. If phase transition: Create session boundary git tag

---

Version: 1.0.0
Source: Git-Based Context Memory System design. Inspired by claude-mem, memory-mcp, and claude-brain architectures.
