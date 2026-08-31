---
paths: "**/.moai/specs/**,**/.claude/agent-memory/**"
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
| Index loading | `MEMORY.md` is loaded at the start of every session. A large index is truncated at some point, but **the cut's shape is not settled** — re-measure rather than quoting a figure. See § MEMORY.md Index Budget |
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
- All auto-memory files (the `MEMORY.md` index AND every topic file) MUST be authored in English only — never the conversation language. CJK costs ~3× the UTF-8 bytes per character vs English for the same information, wasting the context window; LLMs process English with higher adherence and reasoning fidelity; memory is read by the model, not humans. User-facing chat replies still use the conversation language; only the stored memory file is English.

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

### MEMORY.md Index Budget

`MEMORY.md` (the auto-memory index) is loaded at the start of every session, and a sufficiently
large index is truncated. **The cut's shape is not asserted here.** The loader is not part of this
repository, so its behaviour cannot be read from source, and a figure written into doctrine cannot
be re-measured by the reader who later relies on it. Measure your own store instead:

```bash
# `moai memory doctor` reports every candidate store it resolved, and whether each exists.
# Pass the resolved path back with --dir, and read a count only when the JSON says exists: true.
moai memory doctor
moai memory doctor --json --dir "<the store path the line above printed>"
```

Keep each index entry to a single line, short enough to scan.

#### Compressing the index means making entries shorter — never fewer

[HARD] When the index needs to shrink, rewrite entries to be **shorter**. Do **not** remove an
entry, and do not fold several entries into one grouped line to save space.

Removing an index entry does not archive its topic file; it makes that file **unreadable to every
future session**. Topic files are loaded on demand and are discovered *only* through the index, so
an unindexed file is not "still there for later" — it is unreachable. A shorter entry costs a few
characters of detail; a removed entry costs the whole memory.

Two consequences worth stating because both have been observed:

- **A dead index line is worse than a long one.** An entry pointing at a file that is not in the
  store gives a session nothing at all, not a degraded result. Repair the direction that restores
  reachability — put the file back — never the one that deletes the line.
- **Entry-count metrics do not see grouped lines.** A count of `^- \[`-anchored lines misses link
  targets carried on other line shapes, so folding entries into a grouped line can delete live
  references while the count reports no change. When measuring before and after an index edit,
  count unique link targets file-wide as well as anchored entry lines, and treat a decrease in
  **either** as a failure rather than a saving.

#### Two stores, and only one of them is loaded

[HARD] More than one memory store can exist for the same project, and a session loads exactly one
of them. The other is invisible from the index alone: its files are absent from the loaded store,
so every index line naming one reads as a broken link with no indication that the file exists
elsewhere and could simply be copied back.

`moai memory doctor` is the tool that surfaces this — it enumerates **every** candidate store, not
just the active one, and reports `exists`, the topic-file count, and the index-line count for each.

[HARD] Two rules bind every reading taken from it:

- **Always pass `--dir`.** A bare invocation resolves the store path for itself, and the resolved
  path is observably sensitive to where the command is run from: invoked from a worktree or another
  subdirectory, it can resolve a store that does not exist. Run the bare form once to see which
  paths it resolved, then pass the one you mean back with `--dir`. (Exactly how the path is derived
  is not settled — this file's own description of the derivation and the observed behaviour do not
  agree, so rely on what the tool reports rather than on either account.)
- **Check `exists: true` and a non-zero index-line count before reading any finding count.** An
  absent store reports zero findings. That zero means *nothing was measured*, not that nothing is
  wrong, and it will otherwise satisfy a check written to look for zero.

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

**MEMORY.md index is the discovery surface.** Every active topic file SHOULD have a one-line entry in `MEMORY.md` (`- [title](file) — short-summary`), because topic files are loaded on demand — an index entry is how a future session discovers the file exists. When the index grows long, shorten entries rather than dropping them (§ Compressing the index means making entries shorter — never fewer).

**MEMORY.md diet procedure (when the index grows long).** Rewrite each entry as `[title](file) — <topic description>` (single line, ≤~150 chars), drawing the summary from the topic file's `description` field. The body of detail belongs in the topic file, NOT the index entry. This collapses oversized index entries without losing information — which is the point: the procedure shortens entries and keeps every one of them.

**Stale completed records move to `_archive/`, never deleted.** When a topic is a completed/superseded one-time incident with no ongoing relevance, move the file into a `memory/_archive/` subdirectory (reversible) and drop its index entry. Do NOT delete — archive preserves the audit trail. Conservatively keep anything that holds an enduring lesson.
