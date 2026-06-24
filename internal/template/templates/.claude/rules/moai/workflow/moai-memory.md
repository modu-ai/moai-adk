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
5. Optional local instructions file (e.g., a project-local override document if your team maintains one; not committed)
6. Auto Memory: ~/.claude/projects/{hash}/memory/ (AI-managed)

## Official Claude Code Auto-Memory Feature

Auto memory (level 6 above) is a native Claude Code feature (requires v2.1.59 or later). MoAI layers its taxonomy and Lessons Protocol on top of this native feature; the bullets below are the underlying Claude Code behavior.

| Aspect | Behavior |
|--------|----------|
| Default | ON. Disable via the `/memory` toggle, `autoMemoryEnabled: false` in settings.json (any scope), or env `CLAUDE_CODE_DISABLE_AUTO_MEMORY=1` |
| Storage | `~/.claude/projects/<project>/memory/`. The `<project>` path is derived from the **git repository root**, so all worktrees and subdirectories of the same repo share ONE memory directory. Outside a git repo, the project root is used |
| Override | `autoMemoryDirectory` in settings.json (absolute or `~/` path; honored only after the workspace trust dialog) |
| Index loading | `MEMORY.md` is loaded every session — the first **200 lines OR 25KB, whichever comes first**. Content beyond either limit is not loaded at session start |
| Topic files | `debugging.md`, `api-conventions.md`, etc. are NOT loaded at startup; Claude reads them on demand. They are plain markdown with **no mandated frontmatter schema** |
| Subagents | Subagents can maintain their own auto memory (see the Claude Code sub-agents documentation) |
| Inspect | `/memory` lists the loaded CLAUDE.md and rules files, toggles auto memory, and links to the auto-memory folder |

Full reference: `.claude/skills/moai-foundation-cc/reference/claude-code-memory-official.md`.

## Two Memory Locations (do not conflate)

MoAI deals with two distinct memory directories. Keep their rules separate:

| Directory | Scope | Frontmatter schema | Enforcement |
|-----------|-------|--------------------|-------------|
| `.claude/agent-memory/<agent-name>/` | Per-agent memory | MoAI 4-type taxonomy (below) is REQUIRED | PostToolUse audit hook (warnings) |
| `~/.claude/projects/<hash>/memory/` | Project/session auto-memory (the native feature above); `MEMORY.md` index lives here | None mandated by Claude Code; MoAI applies the 4-type convention to its own authored project entries only | Loader line/byte cap |

The 4-type taxonomy below is a MoAI convention. The `### MEMORY.md Line Cap` rule applies to the auto-memory index in `~/.claude/projects/<hash>/memory/`, not to agent-memory files.

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

## Agent Memory Taxonomy

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

`MEMORY.md` (the auto-memory index) is loaded each session up to the first **200 lines OR 25KB, whichever the Claude Code loader reaches first**. Content beyond either limit is silently truncated at session start — entries past the cap are NOT loaded. Keep each entry to a single line under 150 characters, and archive older entries so the index stays within both the 200-line and 25KB limits.

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

The PostToolUse hook audits memory files on Write/Edit operations within `.claude/agent-memory/<agent-name>/` and emits non-blocking warnings to stderr. The hook is scoped to agent-memory files only — it does NOT observe the auto-memory index at `~/.claude/projects/<hash>/memory/MEMORY.md`, which Claude Code writes through its native subsystem rather than the Write/Edit tools the hook sees.

Wired codes (emitted by the PostToolUse hook today):

| Code | Condition |
|------|-----------|
| `MEMORY_MISSING_TYPE` | File has no `type` field in frontmatter |
| `MEMORY_MISSING_FRONTMATTER` | File missing `name` or `description` |
| `MEMORY_BODY_STRUCTURE_MISSING` | feedback/project file missing **Why:** or **How to apply:** |
| `MEMORY_EXCLUDED_CATEGORY` | Body matches an excluded-category keyword pattern |

Available checks, NOT yet wired into the PostToolUse hook — the index-overflow and duplicate-description checks exist in the codebase but currently have no production caller, so do NOT rely on these firing automatically (run them manually or wait for a future SPEC to wire them in):

| Code | Condition |
|------|-----------|
| `MEMORY_INDEX_OVERFLOW` | MEMORY.md exceeds 200 lines |
| `MEMORY_DUPLICATE` | Two files share the same description |

All warnings are observation-only. Hook exit code is always 0 (non-blocking).

### Memory Hygiene (operating discipline)

These operating rules prevent the memory store from degrading into an unread, oversized, mis-classified dump. Apply them whenever you write or consolidate memory entries.

**type field is top-level SSOT.** The canonical `type:` field lives at the frontmatter top level (per Required Frontmatter above). The host's native auto-memory MAY additionally emit a `metadata:` block containing `type:` — when both exist, the top-level `type:` is the source of truth and the two MUST carry the same value. Do not rely on `metadata.type` alone; if a file has `metadata.type` but no top-level `type:`, add the top-level one.

**Filename prefix determines type.** Use the naming convention as the default classifier and keep filenames honest:
- `project_*` → `project`
- `feedback_*` / `lesson*` → `feedback` (lessons are experience-derived feedback)
- `reference_*` → `reference`
- Non-standard filenames require a content-based type decision.

**Topic-file `description` is a one-line recall summary.** Keep each `description:` under ~150 characters. It is the string the recall layer matches against — a long description defeats relevance matching; an empty one defeats discovery. If a topic needs more detail, put it in the body, not the description.

**MEMORY.md index is the discovery surface.** Every active topic file SHOULD have a one-line entry in `MEMORY.md` (`- [title](file) — short-summary`), because topic files are loaded on demand — an index entry is how a future session discovers the file exists. When the index approaches the 200-line / 25KB ceiling, prefer promoting the most-recent active topics over retaining stale ones.

**MEMORY.md diet procedure (when the index exceeds limits).** Rewrite each entry as `[title](file) — <topic description>` (single line, ≤~150 chars), drawing the summary from the topic file's `description` field. The body of detail belongs in the topic file, NOT the index entry. This collapses oversized index entries without losing information.

**Stale completed records move to `_archive/`, never deleted.** When a topic is a completed/superseded one-time incident with no ongoing relevance, move the file into a `memory/_archive/` subdirectory (reversible) and drop its index entry. Do NOT delete — archive preserves the audit trail. Conservatively keep anything that holds an enduring lesson.
