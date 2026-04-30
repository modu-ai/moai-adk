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

### MCP Configuration

`.mcp.json` - MCP server definitions:

- mcpServers: Server command and arguments
- Environment variables for servers

Standard MCP servers in MoAI-ADK:

- context7: Library documentation lookup
- sequential-thinking: Complex problem analysis
- pencil: .pen file design editing. Used by expert-frontend (sub-agent mode) and team-designer (team mode).
- claude-in-chrome: Browser automation

**`alwaysLoad` field (Claude Code v2.1.119+)**

Claude Code v2.1.119에서 `.mcp.json`의 MCP 서버 항목에 `"alwaysLoad": true` 필드가 추가되었다.
이 필드가 `true`로 설정된 서버의 툴 스키마는 세션 시작 시 즉시 로드된다(기존 지연 로드 방식 대비).

MoAI-ADK 기본 설정:
- `context7`: `"alwaysLoad": true` — 매 세션 문서 조회가 빈번하므로 즉시 로드
- `sequential-thinking`: `"alwaysLoad": true` — DeepThink 워크플로우에서 첫 호출 지연 제거
- `moai-lsp`: `alwaysLoad` 미설정 — 프로젝트에 따라 LSP가 필요 없는 경우도 있으므로 지연 로드 유지

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

Source: SPEC-CC2122-MCP-001 (2026-04-30)

MCP tools are deferred and must be loaded before use:

1. Use ToolSearch to find and load the tool
2. Then call the loaded tool directly

Example flow:
- ToolSearch("context7 docs") loads mcp__context7__* tools
- mcp__context7__resolve-library-id is then available

MCP rules:
- Always use ToolSearch before calling MCP tools
- Prefer MCP tools over manual alternatives
- Authenticated URLs require specialized MCP tools

Example `.mcp.json` configuration:

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@context7/mcp"]
    }
  }
}
```

**Context7 Usage** - For up-to-date library documentation:

1. resolve-library-id: Find library identifier
2. get-library-docs: Retrieve documentation

**Sequential Thinking Usage** - For complex analysis requiring step-by-step reasoning:

- Breaking down multi-step problems
- Architecture decisions
- Technology trade-off analysis

Activate with `--deepthink` flag for enhanced analysis.

### MoAI Configuration

`.moai/config/` - MoAI-specific settings:

- config.yaml: Main configuration
- sections/quality.yaml: Quality gates, coverage targets
- sections/language.yaml: Language preferences
- sections/user.yaml: User information

## Hooks Configuration

Hooks support environment variables and must be quoted to handle spaces:

```json
{
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"",
      "timeout": 5
    }],
    "PreToolUse": [{
      "matcher": "Write|Edit|Bash",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh\"",
      "timeout": 5
    }]
  }
}
```

**Important**: Quote the entire path: `"\"$CLAUDE_PROJECT_DIR/path\""` not `"$CLAUDE_PROJECT_DIR/path"`

Hook timeout unit is **seconds** (not milliseconds, despite some external docs). Default is 5s for most hooks. Recommended ceilings:

| Hook | Recommended timeout | Rationale |
|------|--------------------|-----------|
| SessionStart | 30s | MCP server startup latency |
| PreToolUse | 5s | Fast pre-flight checks only |
| PostToolUse | **10s + `async: true`** (was 60s synchronous before v2.16.0) | LSP/AST/MX validations run in background; results delivered via systemMessage on next turn. 10s is the per-run upper bound, not a blocking wait |
| Stop / SubagentStop | 5s | Lightweight teardown |
| TeammateIdle / TaskCompleted | 10s | Quality validation may run lint/test |

For very long validations (full test suites, deployments), prefer `"async": true` over high timeout — the hook runs in background and results arrive on the next turn (see hooks-system.md §Async Command Hooks).

### Freeze Diagnosis Checklist

If a session appears to freeze mid-conversation, check in this order (cheapest to most invasive):

1. **MCP authentication failures** — most common cause. Run `claude mcp list` and remove servers showing `oauth_required` / `connection_failed`. Each unauthenticated MCP can add 5-30s retry latency on tool calls.
2. **Hook timeout** — run `claude --debug "hooks"` to see per-hook latency. If a hook exceeds its timeout, the response stalls until timeout expires. moai hook handlers (post-tool, stop, subagent-stop) typically complete in <50ms; persistent slowness usually points to LSP server hangs.
3. **Context window pressure** — see `.claude/rules/moai/workflow/context-window-management.md`. SSE streams stall when prompts approach 75% of the window.
4. **Terminal I/O saturation** — high write ratio (>90% writes in `tmux info`) can make output appear delayed. This is rendering only, not a true freeze.

Profile a hook directly:
```bash
echo '{"hook_event_name":"PostToolUse","tool_name":"Write","tool_response":{"success":true},"session_id":"test"}' | time moai hook post-tool
```
Healthy result: under 100ms. Persistent slowness → check LSP / disk I/O / MX validation cost.

## StatusLine Configuration

StatusLine does NOT support environment variables. Use relative paths from project root:

```json
{
  "statusLine": {
    "type": "command",
    "command": ".moai/status_line.sh"
  }
}
```

Reference: GitHub Issue #7925 - statusline does not expand environment variables.

## Permission Management

Tool permissions in settings.json:

- Read, Write, Edit: File operations
- Bash: Shell command execution
- Agent: Sub-agent delegation
- AskUserQuestion: User interaction

## Quality Configuration

Quality gates in quality.yaml:

- development_mode: ddd or tdd
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
Two styles ship with MoAI-ADK: **MoAI** (`moai.md`) and **Einstein** (`einstein.md`).

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

