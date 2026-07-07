---
paths: "**/.moai/config/**,**/.mcp.json,**/.claude/settings.json,**/.claude/settings.local.json"
---

# Settings Management

Claude Code and MoAI configuration management rules.

## Configuration Files

### Claude Code Settings

`.claude/settings.json` - Project-level settings:

- allowedTools: Permitted tool list
- hooks: Hook script definitions
- permissions: Access control
- statusLine: Statusline configuration
- attribution: Commit/PR attribution block. Sub-keys: `commit` (attribution text appended to git commits, including trailers; empty string hides attribution), `pr` (attribution text for PR bodies), and `sessionUrl` (Claude Code v2.1.183+; boolean, default `true`) which controls whether the claude.ai session link is appended to commits and PRs created from web or Remote Control sessions — set `false` to omit the Claude-Session trailer and PR-body link. The MoAI template pins `sessionUrl: false` so its own `🗿 MoAI` attribution trailers are not accompanied by a session link. The boolean type was confirmed against the bundled Claude Code v2.1.183 settings schema; verify the type against your own Claude Code instance before pinning a non-default value, since the published machine-readable schemastore entry may lag the release.
- disableBundledSkills: Hide bundled skills/workflows (e.g. `/deep-research`) from discovery. Set `true` to suppress the Claude Code bundled skill catalog so only project + user skills remain visible. An equivalent environment variable form is also supported. MoAI-ADK does not emit this toggle — it is documented here as a Claude Code option that exists for projects that want to ship a curated, bundle-free skill surface.
- `--safe-mode` CLI flag: Launch Claude Code with bundled skills and workflows disabled (equivalent runtime effect to `disableBundledSkills: true`, but applied at launch time rather than via settings). Useful for locked-down environments or when debugging whether a behavior originates from a bundled skill. MoAI-ADK does not pass this flag automatically; it is documented as an available launch option.

#### `/config` command (Claude Code v2.1.178+)

The genuine Claude Code `/config` slash command (distinct from MoAI's `.moai`-prefixed config filesystem paths) edits `settings.json` interactively or directly:

- Direct-set form: `/config key=value` writes a single setting without opening the selector (e.g. `/config theme=dark`). `/config <key>=<value>` is the general syntax.
- Help listing: `/config --help` lists the available shorthand keys the command accepts.
- Toggle-key behavior (within the `/config` settings selector): Enter AND Space both change the currently-selected setting, and Esc now saves-and-closes the selector (it no longer reverts unsaved changes).

### MCP Configuration

`.mcp.json` - MCP server definitions:

- mcpServers: Server command and arguments
- Environment variables for servers

Standard MCP servers in MoAI-ADK:

- context7: Library documentation lookup
- pencil: .pen file design editing. Used by a per-spawn `Agent(general-purpose)` frontend specialist (sub-agent mode) and the designer role_profile (team mode).
- claude-in-chrome: Browser automation
- z.ai MCP servers (optional, GLM backend): three separate servers registered via `moai glm tools enable [vision|websearch|webreader|all]` — `zai-mcp-server` (npx stdio, GLM-4.6V vision tools), `web_search_prime` (HTTP, `webSearchPrime`), `web_reader` (HTTP, `webReader`). Under `moai glm` / `moai cg` GLM panes these replace the built-in `WebSearch` / `WebFetch` / `Read`-on-image per `.claude/rules/moai/core/glm-web-tooling.md`.

> Sequential Thinking MCP was retired in an earlier deep-reasoning consolidation. Use the `ultrathink` keyword (Adaptive Thinking on Opus 4.7+ / 4.8) for deep reasoning.

**`alwaysLoad` field (Claude Code v2.1.119+)**

Claude Code v2.1.119 added the `"alwaysLoad": true` field to MCP server entries in `.mcp.json`.
When this field is set to `true`, the server's tool schema is loaded immediately at session start (instead of the deferred-load default).

MoAI-ADK default configuration:
- `context7`: `"alwaysLoad": true` — loaded eagerly because documentation lookups occur frequently every session

```json
{
  "mcpServers": {
    "context7": {
      "$comment": "Up-to-date documentation and code examples via Context7",
      "alwaysLoad": true,
      "command": "/bin/bash",
      "args": ["-l", "-c", "exec npx -y @upstash/context7-mcp@latest"]
    }
  }
}
```


MCP tools are deferred by default and must be loaded before use. Exception: servers with `alwaysLoad: true` are loaded at session start automatically.

1. Use ToolSearch to find and load the tool
2. Then call the loaded tool directly

Example flow:
- ToolSearch("context7 docs") loads mcp__context7__* tools
- mcp__context7__resolve-library-id is then available
- With `alwaysLoad: true`, this step is unnecessary for context7

MCP rules:
- Always use ToolSearch before calling MCP tools (unless server has alwaysLoad: true)
- Prefer MCP tools over manual alternatives
- Authenticated URLs require specialized MCP tools

Example `.mcp.json` configuration:

```json
{
  "mcpServers": {
    "context7": {
      "alwaysLoad": true,
      "command": "npx",
      "args": ["-y", "@context7/mcp"]
    }
  }
}
```

**MCP `alwaysLoad` field** (introduced v2.1.119; the v2.1.121 change below is the separate `updatedToolOutput` extension, NOT a second introduction): Setting `alwaysLoad: true` on a server entry forces its tool schemas to load at session start, bypassing tool-search auto-mode deferral. MoAI-ADK sets this for `context7` to ensure Context7 documentation lookup is available immediately without ToolSearch preload.

**Claude Code v2.1.119-121 Hook Changes**:

| Version | Change | Impact |
|---------|--------|--------|
| v2.1.119 | PostToolUse / PostToolUseFailure stdin JSON now includes `duration_ms` field | MoAI records slow hooks (>5000ms) to `.moai/observability/hook-metrics.jsonl` when observability dir exists |
| v2.1.119 | `claude --print` mode honors agent `tools:` / `disallowedTools:` frontmatter | CG Mode regression risk — verify `disallowedTools` in agent frontmatter is intentional |
| v2.1.121 | PostToolUse `hookSpecificOutput.updatedToolOutput` extended from MCP-only to all tools | `MOAI_HOOK_OUTPUT_TRANSFORM=1` env var activates output transform scaffold |

**Settings intentionally unset by MoAI-ADK**: Several Claude Code settings exist that MoAI-ADK deliberately does NOT set in `settings.json.tmpl`. Their absence is by design — a future template audit should treat it as intentional, not a gap:

| Setting | Version | Scope | Why MoAI leaves it unset |
|---------|---------|-------|--------------------------|
| `agent` | v2.1.157+ | User/Project/Local (not Managed) | The top-level `agent` key (example `"code-reviewer"`) runs the main thread as a named subagent and sets the default agent for sessions dispatched from `claude agents`, applying that subagent's system prompt, tool restrictions, and model. MoAI invokes its retained agent catalog via explicit delegation, not a session-wide default agent (orchestrator-is-main-thread model). |
| `requiredMinimumVersion` | v2.1.163+ | Managed | Hard version-gate — Claude Code refuses to start when its version is below the floor. An org/admin decision, parallel to the `disableWorkflows` stance. Distinct from the older advisory `minimumVersion`. |
| `requiredMaximumVersion` | v2.1.163+ | Managed | Hard version-ceiling — refuses to start above the cap. Likewise an org/admin decision. |

Reference: https://code.claude.com/docs/en/settings.

**Context7 Usage** - For up-to-date library documentation:

1. resolve-library-id: Find library identifier
2. get-library-docs: Retrieve documentation

**Adaptive Thinking Usage** - For complex analysis requiring deeper reasoning:

- Breaking down multi-step problems
- Architecture decisions
- Technology trade-off analysis

Use the `ultrathink` keyword in user prompts to activate Adaptive Thinking (Opus 4.7+ / 4.8). This is the canonical deep-reasoning path; Sequential Thinking MCP was retired in an earlier consolidation.

### MoAI Configuration

`.moai/config/` - MoAI-specific settings:

- config.yaml: Main configuration
- sections/quality.yaml: Quality gates, coverage targets
- sections/language.yaml: Language preferences
- sections/user.yaml: User information

#### MoAI Configuration — Section Loaders

Configuration sections are loaded via two mechanisms:

**1. `Loader.Load()` chain** (`internal/config/loader.go:31-74`):
Loads the following 15 sections in fixed order. All return defaults on absent file.

| YAML file | loadedSections key | Go field |
|---|---|---|
| user.yaml | `user` | `cfg.User` |
| language.yaml | `language` | `cfg.Language` |
| quality.yaml | `quality` | `cfg.Quality` |
| git-convention.yaml | `git_convention` | `cfg.GitConvention` |
| git-strategy.yaml | `git_strategy` | `cfg.GitStrategy` |
| llm.yaml | `llm` | `cfg.LLM` |
| ralph.yaml | `ralph` | `cfg.Ralph` |
| state.yaml | `state` | `cfg.State` |
| workflow.yaml | `workflow` | `cfg.Workflow` |
| statusline.yaml | `statusline` | `cfg.Statusline` |
| research.yaml | `research` | `cfg.Research` |
| constitution.yaml | `constitution` | `cfg.Constitution` |
| context.yaml | `context_search` | `cfg.ContextSearch` |
| interview.yaml | `interview` | `cfg.Interview` |
| design.yaml | `design` | `cfg.Design` |

**2. Dedicated entry-points** (outside `Loader.Load()` by design):

| Section | Loader | Package | Rationale |
|---|---|---|---|
| harness.yaml | `LoadHarnessConfig(path)` | `internal/config` | FROZEN validation (HRN-001); returns error on absent file (not defaults) |

**MIG-003 new loaders** (`internal/config/loader_{constitution,context,interview,design}.go`):

- `LoadConstitutionConfig(path)` — constitution.yaml; exposes `ForbiddenPatterns` (ForbiddenLibraries alias) policy enforcement.
- `LoadContextConfig(path)` — context.yaml; provides `TokenBudget.MaxInjectionTokens` and `Search.DateRangeDays` for CLAUDE.md §16 Context Search.
- `LoadInterviewConfig(path)` — interview.yaml; provides `ClarityThreshold`, `Plan.MaxRounds`, `SkipConditions`.
- `LoadDesignConfig(path)` — design.yaml; provides `GanLoop.PassThreshold` (FROZEN floor 0.60), `GanLoop.SprintContract.Enabled`, `Adaptation.IterationLimits` for GAN loop runtime.

**SunsetConfig** (`internal/config/types.go`): DORMANT — struct defined but no runtime hot path enforces sunset conditions. `LoadSunsetConfig` must NOT be added until an activation SPEC is filed (REQ-MIG003-006).

**CI Guards** (run on every `go test ./internal/config/...`):

- `YAML_SECTION_NO_LOADER` (`audit_loader_completeness_test.go:TestAuditLoaderCompleteness`): fails if a new `.moai/config/sections/*.yaml` file has no loader and is not in the acknowledged allowlist.
- `CONFIG_STRUCT_YAML_MISMATCH` (`audit_struct_yaml_symmetry_test.go:TestStructYAMLSymmetry_*`): fails if a Go struct field lacks a matching YAML key or vice versa.

**Adding a new YAML section** (5-step procedure):
1. Add `<name>.yaml` to `internal/template/templates/.moai/config/sections/`
2. Add `XxxConfig` struct + sub-types + `xxxFileWrapper` to `internal/config/types.go`
3. Add `defaultXxxConfig()` helper to `internal/config/defaults.go` and wire into `NewDefaultConfig()`
4. Create `internal/config/loader_<name>.go` with `LoadXxxConfig(path)` + `loadXxxSection(dir, cfg *Config)`
5. Wire `l.loadXxxSection(sectionsDir, cfg)` into `Loader.Load()` AND add the struct to `audit_struct_yaml_symmetry_test.go` symmetryCases

## Hooks Configuration

> Canonical: see `.claude/rules/moai/core/hooks-system.md` § Hook Configuration (the hook JSON config block + `$CLAUDE_PROJECT_DIR` path-quoting rule) and § Timeout Configuration (the per-hook timeout table, including the PostToolUse 10s+`async:true` exception vs the 5s synchronous-default). This file owns only the StatusLine-no-env-var delta (below).

Hook timeout unit is **seconds** (not milliseconds). The canonical per-hook timeout policy lives in `hooks-system.md` § Timeout Configuration — the PostToolUse 10s+`async:true` exception (background LSP/AST/MX validation) and the 5s synchronous-default for SessionStart/PreToolUse are stated there. For very long validations, prefer `"async": true` over high timeout.

### Freeze Diagnosis Checklist

If a session appears to freeze mid-conversation, check in this order (cheapest to most invasive):

1. **MCP authentication failures** — most common cause. Run `claude mcp list` and remove servers showing `oauth_required` / `connection_failed`. Each unauthenticated MCP can add 5-30s retry latency on tool calls.
2. **Hook timeout** — run `claude --debug "hooks"` to see per-hook latency. If a hook exceeds its timeout, the response stalls until timeout expires. moai hook handlers (post-tool, stop, subagent-stop) typically complete in <50ms; persistent slowness usually points to LSP server hangs.
3. **Context window pressure** — see `.claude/rules/moai/workflow/context-window-management.md`. SSE streams stall when prompts approach the model-specific threshold (1M context = 50%, 200K context = 90%).
4. **Terminal I/O saturation** — high write ratio (>90% writes in `tmux info`) can make output appear delayed. This is rendering only, not a true freeze.

Profile a hook directly:
```bash
echo '{"hook_event_name":"PostToolUse","tool_name":"Write","tool_response":{"success":true},"session_id":"test"}' | time moai hook post-tool
```
Healthy result: under 100ms. Persistent slowness → check LSP / disk I/O / MX validation cost.

## StatusLine Configuration

The statusLine command runs with the same environment variables as hooks, including the Claude Code built-in `$CLAUDE_PROJECT_DIR`. Anchor paths to it so they resolve regardless of the current working directory (e.g. after `/cd` changes cwd away from the project root):

```json
{
  "statusLine": {
    "type": "command",
    "command": "$CLAUDE_PROJECT_DIR/.moai/status_line.sh"
  }
}
```

Note: GitHub Issue #7925 ("statusline does not expand environment variables") refers to generic shell env-var interpolation and user-defined `env` values, NOT the Claude Code built-in `CLAUDE_PROJECT_DIR` token. The built-in token is exported into the command's environment before it shells out. Reference: https://code.claude.com/docs/en/statusline ("The status line command runs with the same environment variables as hooks, including `CLAUDE_PROJECT_DIR`.").

## Permission Management

Tool permissions in settings.json:

- Read, Write, Edit: File operations
- Bash: Shell command execution
- Agent: Sub-agent delegation
- AskUserQuestion: User interaction

### Permission Rule Syntax

Claude Code permission rules support two forms:

- `Tool(specifier)` — scope a tool by a specifier (e.g. `Bash(npm test:)` allows only `npm test` Bash commands; `Read(//tmp/**)` allows reads under `/tmp`).
- `Tool(param:value)` — param-scoped wildcard (e.g. `WebFetch(domain:example.com)` allows WebFetch only against that domain; `Bash(cmd:git status)` matches the `git status` command). The `*` wildcard is accepted inside the value to broaden a match (`WebFetch(domain:*.example.com)`, `Bash(cmd:git *)`).

Both forms compose with `allow` / `deny` / `ask` in `permissions`. MoAI-ADK does not currently emit param-scoped rules from its own settings generators; the `Tool(param:value)` syntax is documented here as an available option for projects that need fine-grained, parameter-level permission control beyond the plain `Tool(specifier)` form.

Example:

```json
{
  "permissions": {
    "allow": [
      "Bash(npm test:)",
      "Bash(cmd:git status)",
      "WebFetch(domain:*.moai.kr)"
    ]
  }
}
```

## Quality Configuration

Quality gates in quality.yaml:

- constitution.development_mode: ddd or tdd (nested under the top-level `constitution:` block in quality.yaml — NOT a top-level `quality.development_mode` key)
- test_coverage_target: Minimum coverage percentage
- lsp_quality_gates: LSP-based validation

## Language Settings

Language preferences in language.yaml:

- conversation_language: User response language
- agent_prompt_language: Internal communication
- code_comments: Code comment language

## Agent Teams Settings

Agent Teams require both an environment variable and workflow configuration.

### Environment Variable

Enable in `.claude/settings.json`:

```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

This env var must be set for Claude Code to expose the Teams API.

### Workflow Configuration

Team behavior is controlled by the `workflow.team` section in `.moai/config/sections/workflow.yaml`:

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| team.enabled | boolean | true | Master switch for team mode |
| team.max_teammates | integer | 10 | Maximum teammates per team (2-10 recommended) |
| team.default_model | string | inherit | Default model for teammates (inherit/haiku/sonnet/opus) |
| team.require_plan_approval | boolean | true | Require plan approval before implementing |
| team.delegate_mode | boolean | true | Team lead coordination-only mode (no direct implementation) |

### Auto-Selection Thresholds

When `workflow.execution_mode` is `auto`, these thresholds determine when team mode activates:

| Setting | Default | Description |
|---------|---------|-------------|
| team.auto_selection.min_domains_for_team | 3 | Minimum distinct domains to trigger team mode |
| team.auto_selection.min_files_for_team | 10 | Minimum affected files to trigger team mode |
| team.auto_selection.min_complexity_score | 7 | Minimum complexity score (1-10) to trigger team mode |

## Output Style Configuration

Output styles are Markdown files in `.claude/output-styles/moai/` that control how MoAI formats responses.
Three styles ship with MoAI-ADK: **MoAI** (`moai.md`), **MoAI-Easy** (`moai-easy.md`), and **Einstein** (`einstein.md`).

### Precedence

When `outputStyle` is set in multiple places, the first match wins:

| Priority | Source | Key | Example |
|----------|--------|-----|---------|
| 1 (highest) | `.claude/settings.json` (project) | `outputStyle` | `"outputStyle": "Einstein"` |
| 2 | `~/.claude/settings.json` (user) | `outputStyle` | `"outputStyle": "MoAI"` |
| 3 (lowest) | Hardcoded default | — | `"MoAI"` |

**Example 1 — project overrides user:**

```json
// ~/.claude/settings.json
{ "outputStyle": "MoAI" }

// .claude/settings.json (project)
{ "outputStyle": "Einstein" }
```

Result: **Einstein** loads (project wins over user, REQ-WF006-006).

**Example 2 — user setting applies when project is absent:**

```json
// ~/.claude/settings.json
{ "outputStyle": "Einstein" }

// .claude/settings.json (project) — outputStyle key not present
```

Result: **Einstein** loads (user setting applies, REQ-WF006-015).

**Example 3 — third-party style at project level:**

```json
// .claude/settings.json (project)
{ "outputStyle": "ThirdStyle" }
```

Result: **ThirdStyle** loads if the file `output-styles/moai/thirdstyle.md` exists (REQ-WF006-011).
If the file does not exist, see Fallback Policy below.

### Fallback Policy

When the requested style name cannot be resolved to a file in `output-styles/moai/`, MoAI falls back
to the built-in **MoAI** style and emits the following warning to **stderr**:

```
OUTPUT_STYLE_UNKNOWN: <name> not found; falling back to MoAI
```

`<name>` is replaced by the exact string from the `outputStyle` setting (e.g., `NonExistent`).
This warning is emitted to stderr only — it does not appear in the AI response body.

### Frontmatter Schema Contract

Every output style file MUST have a YAML frontmatter block with exactly these required keys:

| Key | Type | Description |
|-----|------|-------------|
| `name` | string | Human-readable style name (e.g., `"MoAI"`) |
| `description` | string | One-sentence description of the style |
| `keep-coding-instructions` | boolean | `true` = preserve coding directives; `false` = suppress |

`keep-coding-instructions` MUST be a raw boolean literal (`true` or `false`) — quoted strings
(`"true"`, `"false"`), capitalized forms (`True`, `False`), or other values are schema errors.

Additional frontmatter keys beyond these three are tolerated and ignored.

### Breaking Change Policy

Adding a new output style requires:
1. Adding the `.md` file to `internal/template/templates/.claude/output-styles/moai/` (Template-First).
2. Running `make build` to regenerate the embedded template.
3. Mirroring the file to `.claude/output-styles/moai/` in the project.
4. Updating `TestOutputStylesExactlyTwo` in `internal/template/output_styles_audit_test.go` to reflect
   the new expected count and add the new file to the allowed set.

Removing a built-in style is a breaking change and requires a major version bump.

## Rules

- Never commit secrets to settings files
- Use environment variables for sensitive data
- Keep settings minimal and focused
- Hook paths must be quoted when using environment variables
- StatusLine uses relative paths only (no env var expansion)
- Template sources (.tmpl files) belong in `internal/template/templates/` only
- Local projects should contain rendered results, not template sources

