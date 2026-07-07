---
paths: ".claude/**/*.md,.claude/**/*.yaml,.moai/**/*.yaml,CLAUDE.md"
---

# Coding Standards

MoAI-specific coding standards. General coding conventions are not included as Claude already knows them.

## Language Policy

All instruction documents must be in English:
- CLAUDE.md
- Agent definitions (.claude/agents/**/*.md)
- Slash commands (.claude/commands/**/*.md)
- Skill definitions (.claude/skills/**/*.md)
- Hook scripts (.claude/hooks/**/*.py, *.sh)
- Configuration files (.moai/config/**/*.yaml)

User-facing documentation may use multiple languages:
- README.md, CHANGELOG.md
- User guides, API documentation

## File Size Limits

CLAUDE.md should stay under 40,000 characters. This is a MoAI CI-enforceable heuristic; the official Claude Code spec instead targets "under 200 lines per CLAUDE.md" and loads the file in full regardless of length. Any project-local instruction file that also loads in full at every session launch follows the same size discipline.

When approaching the limit, reduce launch-time context (priority order):
- Move detailed content to path-scoped rules (.claude/rules/ with `paths:` frontmatter) so it loads only when matching files are touched
- Move stable doctrine to .moai/docs/ and reference it with a plain prose pointer ("See: .moai/docs/<file>.md")
- Trim content not needed in every session
- Keep only core identity and hard rules inline

Note: `@import` (`@path/to/file`) does NOT reduce context — imported files are expanded and loaded in full at launch (see `.claude/skills/moai-foundation-cc/reference/claude-code-memory-official.md`). Use it for organization, never for size reduction.

## Content Restrictions

Prohibited in instruction documents:
- Code examples for conceptual explanations
- Flow control as code syntax
- Decision trees as code structures
- Emoji characters (except output styles)
- Time estimates or duration predictions

## Footer Convention

Rule files follow a consistent-by-absence footer policy: a `Version` / `Status` / `Classification` footer is OPTIONAL, not required. SSOT-owning canonical-reference rules (files that declare "Single Source of Truth" or carry a `Classification: Canonical Reference` line) SHOULD include a footer stating version, status, and classification. Short path-scoped rules MAY omit a footer entirely — absence is a valid consistent state, not a gap. Do not bulk-add footers to rules that currently lack one; the policy statement (this section) is the deliverable, not uniform footer insertion.

## Duplicate Prevention

Single source of truth principle:
- Each piece of information exists in exactly one location
- Use references (@file) instead of copying content
- Update source file, not copies

## Thin Command Pattern

All slash command files MUST be thin routing wrappers (under 20 LOC body).

Rules:
- Commands route to skills via `Skill("moai")` -- they never contain workflow logic
- All workflow logic belongs in `.claude/skills/moai/workflows/` or skill body
- YAML frontmatter must include: description, argument-hint, allowed-tools (CSV string)
- Root commands may contain router tables but no implementation logic
- Custom commands and skills are merged into one namespace: a `.claude/commands/X.md` and a `.claude/skills/X/SKILL.md` with the same name are the same invocation, and the skill form wins when both exist. Author the workflow once as a skill rather than duplicating it across a command and a skill.

Template:
```
---
description: [One-sentence action description]
argument-hint: "[Optional arg]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: [subcommand] $ARGUMENTS
```

Enforcement: `internal/template/commands_audit_test.go` verifies this pattern on every `go test`.


## Claude Code Version Compatibility

Settings fields introduced by specific Claude Code versions:

| Field | Version | Notes |
|-------|---------|-------|
| `effortLevel` | v2.1.110 | Sets CLAUDE_CODE_EFFORT_LEVEL; values: low/medium/high/xhigh/max |
| `disableBypassPermissionsMode` | v2.1.111 | Prevents agents from using bypassPermissions mode when true |
| `Bash(timeout=N)` | v2.1.110 | Per-command Bash timeout in ms; max 600,000ms |

When adding new settings fields, update `internal/template/templates/.claude/settings.json.tmpl`
and this compatibility table.

## Paths Frontmatter

Use paths frontmatter for conditional rule loading:

```yaml
---
paths: "**/*.py,**/pyproject.toml"
---
```

This ensures rules load only when working with matching files.

## Bash Risk-Amplifier Doctrine

Source: `github.com/wquguru/harness-books` book1 ch04 (Bash as 风险放大器 / risk amplifier) + appendix A.3 litmus. book1 ch04 names Bash a **risk amplifier**: a compound Bash command (`a && b | c; rm x`) multiplies blast radius — one failed or mis-targeted segment can destroy more than a single failed command. The appendix A.3 litmus states verbatim: *"if your Bash and ReadTool are governed the same way, your risk model is under-developed."* Prior to this doctrine, moai-adk's PreToolUse hook fired uniformly on `Write|Edit|Bash` with no Bash-specific risk tier — Bash (write/irreversible by default) was treated identically to Read/Glob/Grep (read). The appendix A.3 litmus was failing.

### (1) Bash risk-tier classification

The Bash tool is classified as risk-tier **"write/irreversible" by default**, DISTINCT from the Read/Glob/Grep risk-tier **"read"**. Per book1 ch04, the more general a capability, the denser its governance must be; Bash is the most general tool in the harness and therefore the most heavily governed. This is the risk-amplifier (风险放大器) thesis: a single Bash invocation can fan out into file-system mutation, network calls, git history rewrite, and process spawning in one compound expression.

### (2) Compound-command subcommand soft cap

**`BASH_SUBCOMMAND_SOFT_CAP = 5`** is a named, tunable threshold constant (NOT prose like "around five subcommands"). While a Bash command is compound — contains one or more of: pipe `|`, logical-and `&&`, logical-or `||`, sequence `;`, backtick command substitution, or `$(...)` POSIX substitution — the orchestrator/agent SHALL treat a subcommand count exceeding `BASH_SUBCOMMAND_SOFT_CAP` (5) as a **warn condition** that MUST be surfaced either by **splitting into a script file** OR by **explicit delegation** to a sub-agent. "Soft" means **warn-only** — the cap MUST NOT hard-fail or break existing workflows. The `handle-pre-tool.sh` PreToolUse hook emits a warn signal on this condition but exits 0 (fail-open); see (4) below.

### (3) Destructive primitives require explicit confirmation

When a Bash command contains a destructive primitive, the orchestrator/agent SHALL require explicit confirmation EVEN in `bypassPermissions` mode. The destructive-primitive set (non-exhaustive, extend conservatively):

- `rm -rf` — recursive forced deletion
- `git push --force` (and `git push -f`) — force-push rewrites remote history
- `git push --no-verify` — bypasses pre-push hooks (the harness safety net)
- `git reset --hard` — discards uncommitted working-tree state
- SQL `DROP TABLE` / `TRUNCATE` — irreversible data loss
- `chmod -R 777` — recursive world-writable (security hole)

This cross-references the **Implementation Kickoff Approval** human-gate pattern: irreversible / shared-system / hard-to-reverse actions require explicit user confirmation regardless of permission mode. The warn-only hook signal (§4) does NOT enforce this confirmation — confirmation is an orchestrator/agent obligation (doctrine-level), not a hook block.

### (4) Warn-only, fail-open hook signal

The `handle-pre-tool.sh` PreToolUse hook, on detecting a Bash command whose subcommand count exceeds `BASH_SUBCOMMAND_SOFT_CAP`, emits a warn-only signal (a stderr line AND a structured `[moai:bash-risk] WARN` marker) and then exits **0** (fail-open). The hook MUST NOT block, exit non-zero, or otherwise prevent the Bash call from proceeding. A blocking hook here would itself instantiate the death-spiral hazard book1 ch06 warns about. This fail-open constraint is **non-negotiable**. The counter is a **heuristic** — shell-metacharacter counting over-counts in edge cases (e.g., `echo "a && b && c && d && e && f"` inside a quoted string would over-count); the counter is NOT a parser and MUST NOT be the basis of a hard block.

### (5) Additive-only — no regression to CLAUDE.md safeguards

This doctrine is **strictly additive**. It layers a Bash-specific risk tier on top of the existing uniform PreToolUse hook behavior. It does NOT contradict, weaken, or supersede `CLAUDE.md §7 Safe Development Protocol` or `CLAUDE.md §14 Parallel Execution Safeguards`. Both sections remain the authoritative compatibility targets; this doctrine only adds a Bash-specific risk-amplifier lens that was previously absent.
