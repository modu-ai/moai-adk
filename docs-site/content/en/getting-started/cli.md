---
title: CLI Reference
weight: 90
draft: false
---

A reference for every command and flag of `moai` (the Go binary) that you run in the terminal. It is a completely different tool from `/moai` (the slash subcommand) that you type in the Claude Code chat — this page covers only the terminal CLI.


## Command tree

```bash
moai --help
```

The `moai` CLI is divided into three groups.

| Group | Commands | Description |
|------|--------|------|
| **Launch** | `moai cc` · `moai cg` · `moai glm` | Start a Claude Code session (choose the backend) |
| **Project** | `moai init` · `moai update` · `moai doctor` · `moai status` | Project initialization, update, diagnostics, status |
| **Tools** | `moai profile` · `moai inventory` · `moai hook` · `moai worktree` · `moai spec` · `moai harness` · ... | Configuration, inventory, hooks, worktrees, and other tools |

Use `moai version` to check the currently installed version.

```bash
moai version
```

```text
╭────────────────────────╮
│                        │
│  moai-adk v3.0.0-rc11  │
│                        │
│                        │
╰────────────────────────╯
 v3.0.0-rc11   none   built unknown
```

The line below the box banner shows `<version>   <commit hash>   built <build time>` in order. When built without ldflags (e.g. via `go install`), the commit shows `none` and the build time shows `unknown`.

---

## moai init

Initializes a project. An interactive wizard configures language, Git automation, model policy, harness profile, and more.

```bash
moai init [project-name] [OPTIONS]
```

### Flags

| Flag | Description |
|--------|------|
| `--non-interactive` | Skip the interactive wizard (use flags and defaults) |
| `--force` | Force re-initialization of an existing project (backs up the current `.moai/`) |
| `--no-hooks` | Skip Git hook installation |
| `--all` | Deploy all catalog items (core + optional packs + harness-generated) |
| `--standard` | Show Phase 1 questions (project mode, harness profile, LSP, quality gates, design) |
| `--advanced` | Show Phase 1 + Phase 2 questions (includes `--standard`; Phase 2 only when prerequisites are met) |
| `--mode <ddd\|tdd>` | Development methodology (default: tdd) |
| `--language <lang>` | Primary programming language |
| `--framework <name>` | Framework name (default: auto-detect or "none") |
| `--name <name>` | Project name (default: directory name) |
| `--root <path>` | Project root directory (default: current directory) |
| `--git-mode <manual\|personal\|team>` | Git workflow mode (default: manual) |
| `--git-provider <github\|gitlab>` | Git provider |
| `--project-mode <personal\|team>` | Project mode (default: personal) |
| `--harness-profile <profile>` | Harness evaluator profile (default, strict, lenient, frontend) |
| `--enable-lsp` | Enable LSP integration (default: false) |
| `--enforce-quality` | Enforce quality gates (default: true) |
| `--enable-design` | Enable the design workflow (default: true) |
| `--model-policy <max\|medium\|low>` | Performance tier — stored in `llm.yaml` `performance_tier` |
| `--plan-type <api\|subscription>` | Pricing plan type — stored in `llm.yaml` `plan_type` |
| `--high` | **Deprecated** alias for `--model-policy max` |

### Examples

```bash
# Initialize a new project (interactive wizard)
moai init my-project

# Install into an existing folder
cd my-existing-project
moai init

# Non-interactive (CI/CD)
moai init --non-interactive --project-mode personal --model-policy medium

# Show up to Phase 1 questions
moai init my-project --standard
```

For detailed wizard steps, see the [Initial Setup](./init-wizard) page.

---

## moai update

Updates MoAI-ADK to the latest version. Run without flags, it refreshes both the binary and the templates, and your custom assets are preserved automatically.

```bash
moai update [OPTIONS]
```

### Flags

| Flag | Description |
|--------|------|
| `--check` | Only check whether a new version exists (no update) |
| `-c, --config` | Re-run the configuration wizard (no template sync) |
| `--force` | Force update (skip version match, force backup+merge, overwrite archive drift) |
| `--yes` | Auto-approve all confirmations (CI/CD mode) |
| `--templates-only` | Skip the binary update and sync templates only |
| `--binary` | Skip template sync and update the binary only |
| `--dry-run` | Show planned actions only, with no filesystem changes |
| `--no-hooks` | Skip Git hook installation |
| `--verbose` | Show all warnings (diagnostic mode) |
| `--shell-env` | Configure shell environment variables for Claude Code |
| `--plan-type <api\|subscription>` | Override the pricing plan type (re-applies `llm.yaml` `plan_type` and the tier profile) |

### Examples

```bash
# Default update (binary + templates)
moai update

# Only check whether a new version exists
moai update --check

# Re-run the configuration wizard
moai update -c

# Sync templates only
moai update --templates-only
```

For the detailed update procedure, see the [Update](./update) page.

---

## moai doctor

Runs system diagnostics. It checks Git, the project structure, configuration files, and language-specific development tools.

```bash
moai doctor [OPTIONS]
```

### Flags

| Flag | Description |
|--------|------|
| `-v, --verbose` | Show detailed tool versions and language-detection results |
| `--fix` | Suggest fixes for missing tools |
| `--export <path>` | Export diagnostic results to a JSON file |
| `--check <tool>` | Check a specific tool only (e.g., git, go, config) |

### Subcommands

| Command | Description |
|--------|------|
| `moai doctor sandbox` | Diagnose sandbox backend availability |
| `moai doctor permission` | Diagnose permission resolution |
| `moai doctor hook` | Show the 27-hook-event coverage table |
| `moai doctor config dump` | Dump the merged configuration with provenance |
| `moai doctor config diff <tier-a> <tier-b>` | Compare two config tiers |

### Examples

```bash
# Full diagnostics
moai doctor

# Detailed diagnostics
moai doctor --verbose

# Export diagnostic results
moai doctor --export diagnostics.json
```

---

## moai status

Shows the project status at a glance. It displays whether the project is initialized, the SPEC count, and the number of configuration files.

```bash
moai status
```

It is a read-only command with no flags. For detailed output, see the [Project Status](./status) page.

---

## moai inventory

A read-only command that shows a unified view of active sessions, worktrees, and harnesses.

```bash
moai inventory [OPTIONS]
```

### Flags

| Flag | Description |
|--------|------|
| `--json` | Structured JSON output |
| `--project-root <path>` | Project root path (default: current directory) |

For the detailed JSON schema and usage examples, see the [moai inventory](./inventory) page.

---

## moai profile

Manages Claude Code configuration profiles. Each profile keeps independent model, language, and display settings.

```bash
moai profile [COMMAND]
```

### Subcommands

| Command | Description |
|--------|------|
| `moai profile list` | Show all available profiles |
| `moai profile setup` | Run the interactive setup wizard |
| `moai profile current` | Show the currently active profile |
| `moai profile delete <name>` | Delete the specified profile |

Specify a profile at launch with the `-p` flag:

```bash
moai cc -p work       # Run Claude with the work profile
moai glm -p cost-save # Run GLM with the cost-save profile
moai cg -p team       # Run CG mode with the team profile
```

For more details, see the [Profile Management](./profile) page.

---

## moai hook

A dispatcher that handles Claude Code hook events. It is called in the form `moai hook <event>` from the hook configuration in `settings.json`.

```bash
moai hook <event>
```

### Supported subcommands (~38)

The `moai hook` dispatcher provides about 38 subcommands, combining the standard Claude Code hook events and MoAI-specific internal actions. All names are kebab-case. Below are the representative events.

| Event | Description |
|-------|------|
| `session-start` | Session start |
| `session-end` | Session end |
| `pre-tool` | Before tool execution (PreToolUse) |
| `post-tool` | After tool execution (PostToolUse) |
| `post-tool-failure` | After a tool execution failure |
| `stop` | Session stop |
| `stop-failure` | Stop failure |
| `compact` | Before context compaction (PreCompact) |
| `post-compact` | After context compaction |
| `notification` | System notification |
| `subagent-start` | Subagent start |
| `subagent-stop` | Subagent stop |
| `user-prompt-submit` | User prompt submitted |
| `permission-request` | Permission request |
| `permission-denied` | Permission denied |
| `teammate-idle` | Teammate idle |
| `task-completed` | Task completed |
| `task-created` | Task created |
| `worktree-create` | Worktree created |
| `worktree-remove` | Worktree removed |
| `instructions-loaded` | Instructions loaded |
| `config-change` | Configuration change |
| `cwd-changed` | Working directory changed |
| `file-changed` | File changed |
| `elicitation` | MCP elicitation request |
| `elicitation-result` | MCP elicitation result |

MoAI-specific subcommands are also included.

| Subcommand | Description |
|-------|------|
| `stop-goal` | Evaluate the active session goal at turn end |
| `pre-push` | Validate commit messages against the convention |
| `spec-status` | Auto-update SPEC status on git commit |
| `harness-classify` | Run the harness classifier and record tier promotions |
| `harness-observe` · `harness-observe-stop` · `harness-observe-subagent-stop` · `harness-observe-user-prompt-submit` | Record harness usage logs |
| `db-schema-sync` | Detect DB schema changes in the PostToolUse hook |

You do not run hooks directly — Claude Code's `settings.json` calls them automatically.

---

## moai worktree

Manages Git worktrees for parallel SPEC development.

```bash
moai worktree <COMMAND> [ARGS]...
```

### Subcommands

| Command | Description |
|--------|------|
| `moai worktree new [branch-name]` | Create a new worktree |
| `moai worktree list` | List active worktrees |
| `moai worktree go [branch-name]` | **Print** the worktree path (for shell cd) |
| `moai worktree switch [branch-name]` | Switch to a worktree |
| `moai worktree done [branch-name]` | Complete and clean up a worktree |
| `moai worktree sync [branch-name]` | Sync a worktree with the base branch |
| `moai worktree remove [path]` | Remove a worktree |
| `moai worktree config [key] [value]` | Query/change worktree config |
| `moai worktree status` | Query worktree status |
| `moai worktree clean` | Clean up stale worktree references |
| `moai worktree recover` | Recover the worktree registry |
| `moai worktree snapshot` | Capture a working-tree state snapshot |
| `moai worktree restore` | Restore the working tree to the snapshot HEAD state |
| `moai worktree verify` | Verify the working-tree state against the snapshot |

`moai worktree go` does not change the directory, it only prints the path. To actually move, wrap it in the shell like this:

```bash
cd "$(moai worktree go my-branch)"
```

---

## moai cc / moai cg / moai glm

Launch commands that start Claude Code while choosing the backend. All three support the `-p <profile>` flag to specify a profile. Passing arguments after `--` straight through to Claude Code is supported only by `moai cc` and `moai glm` (`moai cg` does not support it).

```bash
moai cc [-p profile] [-- claude-args...]
moai glm [-p profile] [-- claude-args...]
moai cg [-p profile]
```

| Command | Leader | Workers | tmux required | Use case |
|--------|------|------|-----------|------|
| `moai cc` | Claude | Claude | No | Highest quality (single backend) |
| `moai glm` | GLM | GLM | No | Cost optimization (GLM only) |
| `moai cg` | Claude | GLM | Required | Quality + cost balance (hybrid) |

`moai cg` activates CG mode (a Claude leader + GLM teammates). It must be run inside a tmux session, and it injects the GLM environment variables into the tmux session while the leader pane uses the Claude API. `moai cg` starts Claude Code directly in the current pane after setup, so there is no separate `claude` launch step.

```bash
# 1. Save your GLM API key (once)
moai glm setup sk-your-glm-api-key

# 2. Activate CG mode (run inside tmux — Claude Code starts directly in the current pane)
moai cg
```

For detailed CG mode guidance, see [Introduction — Save tokens with GLM](./introduction#save-tokens-with-glm-5070).

### Launch flags

Flags common to all three launch commands.

| Flag | Description |
|--------|------|
| `-p, --profile <name>` | Use a named Claude profile |
| `--permission-mode <mode>` | Permission mode (default, acceptEdits, plan, auto, bypassPermissions, dontAsk) |
| `-b, --bypass` | Shortcut for `--permission-mode bypassPermissions` |

`moai cc` additionally supports these flags.

| Flag | Description |
|--------|------|
| `-c, --continue` | Continue the previous session |
| `-m, --model <model>` | Override the model selection |
| `--chrome` / `--no-chrome` | Toggle the Chrome MCP |

> The `auto` permission mode is not available on GLM (a third-party provider) — it is supported only in `moai cc` or `moai cg`.

### moai glm subcommands

| Command | Description |
|--------|------|
| `moai glm setup <api-key>` | Save the GLM API key |
| `moai glm status` | Show the current GLM credential status |
| `moai glm tools` | Manage Z.AI MCP server tools (enable/disable) |

---

## moai goal

Registers, queries, and clears a condition-based autonomous goal loop for the current session. It is evaluated at the end of each turn until the condition is met.

```bash
moai goal <COMMAND>
```

| Command | Description |
|--------|------|
| `moai goal arm <condition>` | Register and arm a goal on the active session |
| `moai goal status` | Print the active session's goal status |
| `moai goal clear` | Clear the active session's goal |

---

## moai handoff

Manages the auto-resume handoff pending record for continuing a session across the `/clear` boundary.

```bash
moai handoff <COMMAND>
```

| Command | Description |
|--------|------|
| `moai handoff save` | Save the paste-ready resume body as a pending record |
| `moai handoff clear` | Remove the pending handoff record |

---

## moai session

Manages the active-session coordination registry for multi-session race mitigation.

```bash
moai session <COMMAND>
```

| Command | Description |
|--------|------|
| `moai session current` | Print the current orchestrator session UUID |
| `moai session list` | List active sessions (filterable with `--filter-spec`) |
| `moai session register <session_id> <spec_id> <phase>` | Register a session in the registry |
| `moai session deregister <session_id>` | Remove a session from the registry (idempotent) |
| `moai session heartbeat <session_id>` | Update the session's last_heartbeat |
| `moai session purge` | Remove stale entries (default: last heartbeat older than 30 minutes) |
| `moai session doctor` | Diagnose why the session registry is empty |

---

## moai web

Launches the MoAI Web Console, a browser-based configuration editor.

```bash
moai web [OPTIONS]
```

| Flag | Description |
|--------|------|
| `--port <N>` | TCP port to bind on 127.0.0.1 (default: 3041) |
| `--no-open` | Do not open the browser automatically |
| `--no-reuse` | Do not reclaim the port from a stale moai instance |

---

## moai version

Shows the version, commit hash, and build date.

```bash
moai version
moai --version    # identical
```

---

## Model policy (performance tier)

MoAI-ADK provides a performance-tier system that assigns the optimal AI model to each agent — the starting point of Tokenomics. It is set via the `performance_tier` field in `llm.yaml`, chosen with the `--model-policy` flag or the initialization wizard.

| Tier | Characteristics |
|------|------|
| **max** | Highest quality — Opus assigned to planning and auditing, maximum reasoning depth |
| **medium** (default) | Balance of quality and cost |
| **low** | Economical — Sonnet-centric allocation |

```bash
# Set at initialization
moai init my-project --model-policy max

# Reconfigure an existing project
moai update -c
```

The pricing plan type (`plan_type`: api or subscription) is set separately, so even at the same tier, model assignment differs by billing method. For the detailed model-tier mapping, see the [Model Policy](/en/multi-llm/model-policy) page.

---

## See also

- [Quick Start](./quickstart)
- [Installation](./installation)
- [Update](./update)
- [Initial Setup](./init-wizard)
- [Profile Management](./profile)
- [Project Status](./status)
