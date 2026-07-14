---
paths: "**/.claude/skills/**/SKILL.md"
---

# Skill Authoring

Guidelines for creating MoAI skills following the Agent Skills open standard (agentskills.io).

## YAML Frontmatter Schema

MoAI skills follow the Agent Skills standard with MoAI-specific extensions.

### Standard Fields (agentskills.io)

Required fields:
- description: Purpose description using YAML folded scalar (>). The official cap is 1,536 characters combined across `description` + `when_to_use` (the skill-listing entry).

Optional standard fields:
- name: Skill identifier, lowercase with hyphens, max 64 characters. Optional — when omitted it defaults to the skill directory name. The `moai-{category}-{name}` form (system) / `custom-{name}` form (user) remains the recommended convention for MoAI skills.
- license: SPDX license identifier (default: Apache-2.0)
- compatibility: Target platform description, max 500 characters (default: Designed for Claude Code)
- allowed-tools: Comma-separated string of tool names the skill can use (experimental)
- user-invocable: Boolean to control slash command menu visibility (default: true, set to false to hide from / menu)
- disable-model-invocation: Boolean, when true only user can invoke (not Claude). Use for workflows with side effects (default: false)
- effort: Session effort override: low, medium, high, xhigh, max (xhigh/max require Opus 4.7+)
- model: Model override when skill is active (sonnet, opus, haiku)
- shell: Shell for command injection: bash (default) or powershell
- context: Set to "fork" to run skill in forked subagent context (isolated execution)
- agent: Subagent type when context is fork. Built-in: Explore, Plan, general-purpose, or custom agent name
- hooks: Hook definitions scoped to skill lifecycle
- paths: Glob patterns limiting auto-invocation to matching files (comma-separated or YAML array)
- when_to_use: Additional trigger context (trigger phrases, example requests) appended to description in the skill listing; counts toward the 1,536-character listing cap
- argument-hint: Autocomplete hint for expected arguments, e.g. `[issue-number]` or `[filename] [format]`. Top-level field — NOT a metadata key
- arguments: Named positional arguments for `$name` substitution in skill content; space-separated string or YAML list, mapped to argument positions in order
- disallowed-tools: Tools removed from Claude's available pool while this skill is active (comma/space-separated string or YAML list). Use for autonomous skills that must never call certain tools (e.g. AskUserQuestion in a background loop); the restriction clears on the next user message

### metadata Map

Key-value pairs where both keys and values MUST be strings. Used for simple custom properties.

Common metadata keys:
- version: Semantic version as string (e.g., "1.0.0")
- category: foundation, workflow, domain, language, platform, library, tool
- status: active, experimental, deprecated
- updated: ISO date as string (e.g., "2026-01-28")
- modularized: Whether content is split into modules ("true" or "false")
- tags: Comma-separated tag list as single string
- author: Skill author name
- related-skills: Comma-separated related skill names
- aliases: Comma-separated alternative names

Note: `argument-hint`, `arguments`, `context`, and `agent` are **top-level frontmatter fields** (see Optional standard fields above), NOT metadata keys. Do not nest them under `metadata:`.

### MoAI Extension Fields

Complex structured fields kept at top level with standardized comments.

progressive_disclosure: Token optimization configuration
- enabled: boolean
- level1_tokens: approximate tokens for metadata level
- level2_tokens: approximate tokens for body level

triggers: Loading trigger conditions
- keywords: list of trigger keywords
- agents: list of agent names that load this skill
- phases: list of workflow phases
- languages: list of programming languages

### Schema Example

```yaml
---
name: moai-example-skill
description: >
  Brief description of what this skill does, within the 1,536-character listing cap.
  Use YAML folded scalar (>) for multi-line descriptions.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
effort: low
shell: bash
model: sonnet
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-01-28"
  modularized: "false"
  tags: "example, demo, template"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["example", "demo"]
  agents: ["manager-develop"]
  phases: ["run"]
---
```

### Key Format Rules

allowed-tools format: [ZONE:Evolvable] [HARD] Comma-separated string ONLY. Space-separated values are PROHIBITED.
- Correct: `allowed-tools: Read, Grep, Glob, Bash`
- WRONG: `allowed-tools: Read Grep Glob Bash` (YAML parses as single string scalar, silently breaks tool permissions)
- YAML arrays also supported since v2.1.0 but CSV is the MoAI convention

metadata values: All values must be quoted strings.
- Correct: `version: "1.0.0"`
- Wrong: `version: 1.0.0`

description format: Use YAML folded scalar (>) for readability.
- Correct: `description: >\n  Multi-line description here.`
- Wrong: `description: "Long description in quotes"`

## Progressive Disclosure

Three-level system for token efficiency:

Level 1 (Metadata):
- Tokens: ~100
- Content: name, description, version, triggers
- Loading: The description is **always listed** so Claude (and the user's `/` menu) knows the skill exists — this is the skill-listing step, distinct from loading the body.

Level 2 (Body):
- Tokens: ~5000
- Content: Full documentation, code examples
- Loading: The body is loaded **on invocation** (when the skill is invoked, whether by Claude auto-matching the description or by the user) and then stays in context across turns until compaction. "Listing the description" (Level 1) and "loading the body on invocation" (Level 2) are separate events — the description being listed does not by itself load the body.

Level 3 (Bundled):
- Tokens: Variable
- Content: reference.md (single root file), references/ (multi-file directory), reference/ (single directory, Claude Code Pattern 2), modules/, examples/, INDEX.md (module index)
- Loading: On-demand by Claude

### Skill Directory Layout

MoAI skills follow the Agent Skills standard with MoAI-specific extensions. The bundled directory layout (Level 3) accepts any of the following optional supporting locations, chosen by skill size and content shape:

| Path | Form | Status | Use |
|------|------|--------|-----|
| `reference.md` | root file | official SSOT | single external-links / API-docs file (simple skills) |
| `references/` | multi-file directory | MoAI official extension | multiple reference files grouped together (large skills, e.g. a `claude-code-*-official.md` series) |
| `reference/` | single directory | Claude Code Pattern 2 (Anthropic `skills.md` bigquery example) | domain-specific reference organization |
| `modules/` | multi-file directory | MoAI official (`modular-system.md`) | topic-focused deep-dive content, self-contained |
| `examples.md` | root file | official | copy-paste-ready working examples |
| `scripts/` | directory | official | executable utility scripts |
| `templates/` | directory | official | reusable file templates |
| `INDEX.md` | root file (inside `modules/`) | MoAI convention | optional human-readable index of the modules in a `modules/` directory |

A skill MAY use `reference.md`, `references/`, and `reference/` independently or together based on its content volume. `modular-system.md` (inside the `moai-foundation-core` skill) is the canonical deep-dive on this layout.

### Skill Listing Budget and Compaction (Claude Code runtime)

MoAI's 3-level disclosure sits on top of two runtime budgets the Claude Code host applies. CLAUDE.md § Progressive Disclosure System cross-references this section as canonical:

- **Listing budget**: skill descriptions are loaded so Claude knows what is available; the budget scales at ~1% of the model context window. On overflow, the least-used skills' descriptions are dropped first (names are always kept). Raise it with the `skillListingBudgetFraction` setting (e.g. `0.02` = 2%) or the `SLASH_COMMAND_TOOL_CHAR_BUDGET` env var; each entry's combined `description` + `when_to_use` text is capped at 1,536 characters (`maxSkillDescriptionChars`). Run `/doctor` to detect overflow.
- **Compaction budget**: an invoked skill's rendered content stays in context across turns. After auto-compaction, Claude Code re-attaches the most recent invocation of each skill keeping its first ~5,000 tokens, sharing a combined ~25,000-token budget filled from the most-recently-invoked skill. Older skills can be dropped entirely; re-invoke a skill after compaction to restore its full content.

## Tool Permissions by Category

Foundation Skills:
- Allowed: Read, Grep, Glob, WebFetch
- Never: Bash, Agent

Workflow Skills:
- Allowed: Read, Write, Edit, Grep, Glob, Bash, TodoWrite
- Conditional: AskUserQuestion (MoAI only), Agent (managers only)

Domain Skills:
- Allowed: Read, Grep, Glob, Bash
- Conditional: Write, Edit (implementation tasks only)
- Never: AskUserQuestion, Agent

Language Skills:
- Allowed: Read, Grep, Glob, Bash, WebFetch
- Conditional: Write, Edit (implementation tasks only)
- Never: AskUserQuestion, Agent

## Trigger Configuration

[HARD] The `triggers:` block is OPTIONAL metadata, NOT a machine matcher. Claude Code does not literally match these keywords to route a skill — skill invocation is model-side **semantic matching** of the `description` / `when_to_use` fields. The `keywords` / `agents` / `phases` entries document intent for human readers and tooling; they are not a literal-match gate. Prefer a precise `description` over exhaustive `triggers` keywords.

```yaml
triggers:
  keywords: ["api", "database", "authentication"]
  agents: ["manager-spec", "manager-develop"]
  phases: ["plan", "run"]
  languages: ["python", "typescript"]
```

## Agent Initialization

### initialPrompt

Agents can specify an initial prompt that auto-submits when the agent starts. This enables agents to begin work immediately without waiting for user input. Available since Claude Code v2.1.83+.

The initialPrompt field is only applicable to agent definitions (.claude/agents/), not skills.

Example:
```yaml
---
name: my-agent
initialPrompt: "Analyze the following code for performance issues: @.src/"
---
```

## Built-in Variables

Variables available inside skill SKILL.md content:

| Variable | Description | Available Since |
|----------|-------------|-----------------|
| `${CLAUDE_SKILL_DIR}` | Absolute path to the skill's own directory | v2.1.69 |
| `${CLAUDE_SESSION_ID}` | Current session identifier | v2.1.9 |
| `${CLAUDE_PLUGIN_ROOT}` | Plugin root directory (plugin skills only) | v2.0.12 |
| `$ARGUMENTS` | All arguments passed when invoking skill | v2.1.0 |
| `$ARGUMENTS[N]` | Specific argument by 0-based index (e.g., `$ARGUMENTS[0]`) | v2.1.0 |
| `$N` | Shorthand for `$ARGUMENTS[N]` (e.g., `$0`, `$1`) | v2.1.0 |

Use `${CLAUDE_SKILL_DIR}` for referencing files within the skill directory instead of relative paths. This is more reliable across different invocation contexts.

If skill content does not reference `$ARGUMENTS`, Claude Code automatically appends `ARGUMENTS: <value>` at the end of the skill content.

## Skill Invocation Control

Three invocation modes control how skills appear and load:

| Setting | User invokes | Claude invokes | Description loaded | Use case |
|---------|-------------|---------------|-------------------|----------|
| (default / user-invocable: true) | Yes | Yes | Always | Standard skills |
| disable-model-invocation: true | Yes | No | No | Workflows with side effects |
| user-invocable: false | No | Yes | Always | Background knowledge |

When `disable-model-invocation: true` is set, the skill is NOT loaded into Claude's context, so Claude cannot auto-invoke it. Use for skills that perform destructive actions.

When `user-invocable: false` is set, the skill is hidden from the `/` menu but Claude can still invoke it as background knowledge. Use for reference material.

### skillOverrides setting

The `skillOverrides` settings key (a map of skill name to override state) lets settings.json adjust how an individual skill is exposed without editing its frontmatter. Each entry takes one of four states:

| State | Effect |
|-------|--------|
| `on` | Skill is fully available (user-invocable AND Claude-invocable) |
| `name-only` | The skill name is listed, but its description is not loaded into context |
| `user-invocable-only` | Only the user can invoke it; Claude cannot auto-invoke |
| `off` | Skill is disabled entirely |

`skillOverrides` applies to personal and project skills only — plugin skills are not affected by it.

## Shell Command Injection

Skills support dynamic context via shell command injection. Commands run BEFORE skill content is sent to Claude, and their output replaces the placeholder in the skill content.

Inline syntax: Use exclamation-backtick notation to inject a single command's output inline.

Multi-line syntax: Use a triple-backtick fence with exclamation mark as the language identifier to run multiple commands sequentially.

Policy control: Disable with `disableSkillShellExecution: true` in settings.json. Each command placeholder is replaced with `[shell command execution disabled by policy]`.

## Skill Scope and Priority

Skills can exist at multiple levels. When the same name exists across levels, higher priority wins:

| Priority | Location | Path | Scope |
|----------|----------|------|-------|
| 1 (highest) | Enterprise | Per managed settings | All org users |
| 2 | Personal | ~/.claude/skills/name/SKILL.md | All projects |
| 3 | Project | .claude/skills/name/SKILL.md | This project |
| 4 (lowest) | Plugin | plugin/skills/name/SKILL.md | Where enabled (uses plugin-name:skill-name namespace) |

### Discovery (nested / monorepo / --add-dir)

Project skill discovery walks the directory tree: Claude Code finds `.claude/skills/` not only at the project root but also in nested subdirectories (parent-walk), so a monorepo can place package-local skills in each package's own `.claude/skills/` directory. When you are working inside a nested directory that contains its own `.claude/skills/`, the skills in that nested directory are loaded alongside the root-level skills for the duration of the work in that subtree. Directories added at launch via the `--add-dir` flag are an exception — their skills are NOT auto-loaded for skill discovery (use `permissions.additionalDirectories` in settings.json when an added directory's skills should participate in discovery rather than `--add-dir`, which grants file access only).

### Closest-wins on name collision (nested `.claude/`)

When the same skill name appears in more than one `.claude/skills/` directory along the nested chain, the **closest-directory-wins** rule resolves the collision: the `.claude/skills/` nearest to the current working directory shadows the one further up the tree. This mirrors the precedence that already applies to agents, workflows, and output-styles under nested `.claude/` directories — the innermost `.claude/` wins. Authoring implication: a package-local skill that intentionally overrides a root skill MUST keep the same name; renaming it would create a second skill rather than an override.

### `disableBundledSkills` toggle

`disableBundledSkills` (settings.json boolean, or its environment-variable form) hides the Claude Code bundled skills and workflows — e.g. `/deep-research`, built-in slash-command skills — from discovery, leaving only enterprise + personal + project + plugin skills visible. Use it when shipping a curated, bundle-free skill surface. MoAI-ADK does not emit this toggle from its own generators; it is documented here as an available option. See `.claude/rules/moai/core/settings-management.md` § Claude Code Settings for the companion `--safe-mode` launch flag.

## Best Practices

- Use minimum required permissions
- Prefer Read before Write/Edit operations
- Prefer Edit over Bash for file modifications
- Include 5-10 keywords per skill for accurate triggering
- Overestimate token usage by 10-20% for safety
- Use YAML folded scalar (>) for description field
- Keep all metadata values as quoted strings
- Use comma-separated format for allowed-tools (YAML arrays also supported since v2.1.0)
- Mark MoAI extension fields with standardized comments
- Use `${CLAUDE_SKILL_DIR}` for self-referencing paths within skill content
- Keep skill descriptions concise for menu display, within the 1,536-character listing cap (combined `description` + `when_to_use`)

## Language Guidance Lives in Rules, Not Skills

<!-- @MX:ANCHOR: Language-as-rules canonical decision; cross-referenced by all skill authors. Changes here affect every future language-related decision. -->
<!-- @MX:REASON: This section is the single source of truth for language vs skill classification; consulted by every skill author and plan-auditor on every language-related decision. -->

The 16 supported languages live as **rules** under
`.claude/rules/moai/languages/*.md`, never as skills.

- **No `moai-lang-<name>` skill** may be created. Any PR adding such a
  skill directory triggers `LANG_AS_SKILL_FORBIDDEN` in CI.
- **Canonical location**: `.claude/rules/moai/languages/<name>.md` for all
  16 supported languages: `cpp`, `csharp`, `elixir`, `flutter`, `go`,
  `java`, `javascript`, `kotlin`, `php`, `python`, `r`, `ruby`, `rust`,
  `scala`, `swift`, `typescript`. Canonical Dart name is `flutter` per
  `.claude/rules/moai/development/coding-standards.md` § Language Policy
  (16-language neutrality contract).
- **Loading mechanism**: each language rule uses `paths:` frontmatter for
  conditional loading (e.g., `paths: "**/*.py,**/pyproject.toml"`).
  Path-based loading is the structurally correct primary mechanism for
  language-scoped guidance; keyword-based skill activation is the wrong
  abstraction for files-on-disk language detection.
- **Adding a 17th language**: create a new `.md` file under
  `.claude/rules/moai/languages/` with a `paths:` frontmatter; never a new
  skill. A reversal of this decision requires a new SPEC with an atomic
  migration plan covering all languages (no partial adoption).
- **Cross-language abstraction**: when guidance applies across languages
  (general API design, security checklists), use the `moai-ref-*` skills
  (`moai-ref-api-patterns`, `moai-ref-owasp-checklist`) — not a
  `moai-lang-*` composite.
- **CI guard**: `internal/template/lang_boundary_audit_test.go` enforces
  this principle.

See `.claude/rules/moai/languages/*.md` (16 files) for the canonical
per-language guidance, and `.claude/rules/moai/development/coding-standards.md`
§ Language Policy for the 16-language neutrality contract.

## Skills Namespace Policy

[ZONE:Evolvable] [HARD] The skill namespace is split into "general distribution" vs "user-generated", and the prefix determines the namespace.

| Prefix | Scope | Source of Truth | `moai update` behavior |
|--------|-------|-----------------|------------------------|
| `moai-foundation-*` / `moai-workflow-*` / `moai-domain-*` / `moai-ref-*` / `moai-meta-*` | core framework + workflow + domain + reference | template | **delete then reinstall** (overwrite) |
| `moai-harness-*` | **harness builder/lifecycle** (currently only `moai-harness-learner`; `moai-meta-harness` is a deprecated legacy-redirect) | template | **delete then reinstall** (overwrite) |
| **`hns-*`** | **user-generated** — created by the v4 harness Builder (`builder-harness` agent via `/moai harness`) | user project | **NEVER delete/modify + preserve backup** (Go enforcement recognizes canonical `hns-*` plus the legacy `harness-*` and `my-harness-*` generations — tri-generation recognition) |

### Deprecated Skill Slots (split into three independent harnesses)

The following dev-only skill slots were retired and their workflows live as three INDEPENDENT dev-maintainer harnesses under the user-owned harness namespace (`.claude/agents/harness/hns-{release-update,github,release}-specialist.md` + `.claude/commands/harness/{release-update,github,release}.md`; only release-update carries a Runner + `.claude/commands/harness/release-update/manifest.json`; see agent-authoring.md § Agent Directory Convention). Each thin command routes directly to its matching specialist. These three workflows were first consolidated into a single unified entry by the harness-consolidation effort, then split into three independent harnesses by the harness-split effort (which reverses the unified-entry decision). The earlier intermediate migration into `.claude/agents/local/*-specialist.md` (with `/97-release-update`, `/98-github` thin wrappers) and the standalone `/99-release` command were all removed during the consolidation step.

| Retired Skill / Command | Current Target | Entry Point |
|---------------|------------------|-------------|
| `.claude/skills/moai/workflows/release-update.md` (97 series) | `.claude/agents/harness/hns-release-update-specialist.md` | `/harness:release-update` |
| `.claude/skills/moai/workflows/github.md` (98 series) | `.claude/agents/harness/hns-github-specialist.md` | `/harness:github` |
| `.claude/skills/moai/workflows/release.md` (99 series) | `.claude/agents/harness/hns-release-specialist.md` | `/harness:release` |

Each split harness preserves the structural fidelity of its workflow body. Routing shifted from `Skill("moai/workflows/<name>")` / `Use the <name>-specialist subagent` to `/harness:<name>` direct dispatch. The harness artifacts live in the user-owned namespace (`moai update` preserves them); they are dev-only and never distributed to user projects.

### Rules

- [HARD] The `moai-*` namespace (all prefixes) is template-distributed. If the user modifies it directly, the next `moai update` overwrites it — user customizations are lost.
- [HARD] The `hns-*` namespace is user-owned, as are the legacy `harness-*` and `my-harness-*` generations. `moai update` MUST NOT delete, modify, or sync skills in these namespaces. Backup is mandatory.
- [HARD] Per-project domain skills emitted by the v4 harness Builder (`builder-harness` agent via `/moai harness`) MUST use the **`hns-*` prefix only**. Emitting under `moai-harness-*` or any other `moai-*` prefix violates the contract. (The legacy `moai-meta-harness` skill is deprecated and only redirects to the v4 Builder.)
- [HARD] Do NOT mistake the `moai-harness-*` namespace for per-project artifacts — this namespace is framework-builder-only and currently comprises `moai-harness-learner` (with the deprecated legacy-redirect `moai-meta-harness`).
- [HARD] Distinguish `hns-*` / legacy `harness-*` (user-owned) vs `moai-harness-*` (template builder) as substrings: prefix matching MUST use an exact startsWith comparison; the `*harness-*` substring pattern is prohibited due to false-positive risk.
- [HARD] CI guard: a leak of `internal/template/templates/.claude/skills/hns-*` or `internal/template/templates/.claude/skills/harness-*` MUST fail lint (the namespace-leak sentinel detects both the `hns-` and `harness-` patterns).

### Cross-References

- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy (this section — canonical skill namespace SSOT)
- `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation (canonical generator contract)
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention (agent counterpart — includes `.claude/agents/local/` for the migrated maintainer specialists)
