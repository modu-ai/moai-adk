---
title: Statusline System and PR Segment
weight: 78
draft: false
---

A **custom statusline system** for integrating Claude Code with moai-adk-go. Starting from v2.1.145, you can display GitHub PR information in the statusline.

> The MoAI workflow is PR-centric. Every SPEC generates plan-PR → run-PR → sync-PR, so displaying the current PR status in the statusline improves development efficiency.

## Overview

### Why a Custom Statusline?

Claude Code's default statusline is optimized for general usage patterns. However, MoAI-ADK users need specialized information:

- **PR-Centric Workflow**: Display the current PR number and review status (approved/pending/changes_requested)
- **Multi-Pane Development**: Show current SPEC status when using worktree-based parallel development
- **Cost Tracking**: Real-time cost monitoring when using GLM environments
- **Context Management**: Current session token usage and cumulative costs

The custom statusline displays this information through the `.moai/status_line.sh` renderer.

### Statusline Architecture

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData parsing)
    ↓
internal/statusline/builder.go (segment composition)
    ↓
internal/statusline/renderer.go (color coding and rendering)
    ↓
.moai/status_line.sh (template-based final rendering)
```

## Configuration

### Basic Structure

Configure the statusline in `.moai/config/sections/statusline.yaml`:

```yaml
statusline:
  mode: default              # default | compact | verbose
  theme: catppuccin-mocha    # Color theme selection
  preset: full               # full | minimal | custom
  segments:
    model: true              # Display Claude model
    context: true            # Display context usage
    directory: true          # Display working directory
    git_status: true         # Display Git status
    git_branch: true         # Display Git branch
    worktree: false          # Display worktree info (optional)
    effort_thinking: false   # Display effort/thinking state (optional)
    pr: false                # Display PR info (optional, v2.1.145+)
```

### Segment Options

| Segment | Default | Purpose | Description |
|---------|---------|---------|-------------|
| `model` | true | Current model | Display Claude model version |
| `context` | true | Context usage | Current session token usage |
| `directory` | true | Working path | Current working directory |
| `git_status` | true | Git status | Modified files count, stash status |
| `git_branch` | true | Current branch | Branch name and divergence from remote |
| `worktree` | false | Worktree info | Display current worktree (parallel dev) |
| `effort_thinking` | false | Thinking mode | effort and thinking state |
| `pr` | false | PR info | GitHub PR number and review status (NEW v2.1.145+) |

## Available Segments

### Always-Active Segments (4)

**model** — Claude model
- Display current model (Claude 3.5 Sonnet, Claude 3.7 Opus, etc.)
- Example: `Claude 3.5 Sonnet`

**context** — Context usage
- Display current session token usage
- Format: `150K/200K` (used / total)
- Warning color when usage exceeds 75%

**directory** — Working directory
- Display relative path of current working directory
- Shows location relative to project root

**git_status** — Git status
- Modified files count: `M5` (5 files modified)
- Stash status: `S2` (2 stashes)
- Example: `M5 S2`

**git_branch** — Current branch
- Branch name
- Commit difference from remote
- Example: `feat/SPEC-001 +3 -1`

### Optional Segments (7)

**worktree** — Worktree information (optional)
- Display when using L2 worktree
- Show current SPEC name
- Enable: `segments.worktree: true`

**effort_thinking** — Effort/thinking state (optional)
- Claude 4.7 thinking mode status
- effort level (high/xhigh/max)
- Enable: `segments.effort_thinking: true`

**output_style** — Output style (optional)
- Current output style setting
- Enable: `segments.output_style: true`

**claude_version** — Claude version (optional)
- Claude Code version
- Enable: `segments.claude_version: true`

**moai_version** — moai version (optional)
- MoAI-ADK version
- Enable: `segments.moai_version: true`

**session_time** — Session elapsed time (optional)
- Time elapsed since current session start
- Enable: `segments.session_time: true`

**usage_5h** — 5-hour cumulative cost (optional)
- Cost tracking for the past 5 hours
- Useful in GLM environments
- Enable: `segments.usage_5h: true`

**usage_7d** — 7-day cumulative cost (optional)
- Cost tracking for the past 7 days
- Enable: `segments.usage_7d: true`

**task** — Active SPEC workflow info (optional)
- Output format: `📋 [<command> <SPEC-ID>-<stage>]` (e.g. `📋 [/moai run SPEC-V3R5-DOCS-SECURITY-001-M3]`)
- Data source: `~/.moai/state/last-session-state.json` `active_task` field (auto-set by the SessionStart hook)
- Inactive task renders nothing (segment hidden — graceful no-output)
- Enable: `segments.task: true` (default `true` as of v2.20.0-rc1, opt-out via `false`)

**repo** — Repository information (optional, v2.1.145+)
- Display current GitHub repository owner/name
- Example: `modu-ai/moai-adk`
- Enable: `segments.repo: true`

**long_context** — Long context warning (optional, v2.1.139+)
- Display warning marker when 200K tokens exceeded
- Example: `⚠️ 200K+ exceeded`
- Enable: `segments.long_context: true`

**handoff_guide** — Handoff threshold guide (optional, v2.1.146+)
- Display current model's context window size and recommended handoff threshold
- 1M model: 50% threshold (≈500K tokens)
- 200K model: 90% threshold (≈180K tokens)
- Example: `[1M: 50% | 200K: 90%]`
- Enable: `segments.handoff_guide: true`

## NEW v2.1.145: PR Segment

### Overview

Starting from Claude Code v2.1.145, the statusline stdin JSON includes GitHub PR information. MoAI-ADK leverages this to display the current PR's review status in the statusline.

**Enable**: Default `true` as of v2.20.0-rc1 (default-on). Set `segments.pr: false` to opt out. Graceful no-output: when no PR info is present, the segment is hidden automatically.

### PR Segment Display Format

The PR segment is displayed in the following format:

```
#1023 ⌥approved
```

- `#1023`: PR number
- `⌥`: PR status indicator symbol
- `approved`: Review status (color coded)

### Review State Colors

The PR's review status is displayed in different colors:

| State | Color | Meaning |
|-------|-------|---------|
| `approved` | Green | PR is approved |
| `pending` | Yellow | Waiting for review |
| `changes_requested` | Red | Changes requested |
| `draft` | Gray | Draft state |
| (other / empty) | Default | No style applied |

### How to Enable

1. Edit `.moai/config/sections/statusline.yaml`

```yaml
statusline:
  segments:
    pr: true   # Enable PR segment
```

2. Restart Claude Code session

Now the statusline will display the current PR number and review status.

### JSON Input Schema (v2.1.145+)

Claude Code v2.1.145+ passes the following JSON format to the statusline stdin:

```json
{
  "pr": {
    "number": 1023,
    "url": "https://github.com/modu-ai/moai-adk/pull/1023",
    "review_state": "pending"
  },
  "workspace": {
    "repo": {
      "host": "github.com",
      "owner": "modu-ai",
      "name": "moai-adk"
    }
  }
}
```

- **pr.number**: PR number (required)
- **pr.url**: PR URL (optional)
- **pr.review_state**: Review state (optional, default: empty)
- **workspace.repo.host**: Git host (github.com)
- **workspace.repo.owner**: Repository owner
- **workspace.repo.name**: Repository name

### Related Information

- **SPEC Reference**: [SPEC-V3R5-STATUSLINE-V2145-001](/en/advanced/statusline#reference)
- **Minimum Version**: Claude Code v2.1.145 or later required
- **Optional Feature**: Default is `false`, must be explicitly enabled
- **Backward Compatibility**: Earlier versions of Claude Code do not provide PR information (segment will not display)

## Troubleshooting: Statusline Disappearing

### Symptoms

- Statusline intermittently does not display
- Statusline area is blank in Claude Code UI
- `.moai/cache/statusline_debug.log` file keeps growing

### Root Cause Analysis (Before v2.1.145 M1 fix)

The statusline renderer must comply with Claude Code's **300ms debounce contract**. Violations cause in-flight executions to be cancelled.

Issues in previous code:

```bash
# Problem: DEBUG_STATUSLINE defaults to 1 (always enabled)
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-1}

# This caused every render to:
# 1. Fork python3 -m json.tool process (50-250ms)
# 2. Write to ~/.moai/cache/statusline_debug.log (~10ms)
# Total: 60-260ms → exceeds 300ms debounce boundary
# → Claude Code cancels in-flight statusline rendering
# → Result: statusline does not display
```

### Solution (Fixed in v2.1.145 M1)

From v3.5.0 onwards, `DEBUG_STATUSLINE` defaults to **0**:

```bash
# Fixed: default 0 (disabled)
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0}

# Only enable explicitly when debugging:
export DEBUG_STATUSLINE=1
```

### Padding Adjustment

Previously, `echo ""` was used to adjust spacing around the statusline. This is no longer recommended.

**Instead**, configure it in `.claude/settings.json`:

```json
{
  "statusLine": {
    "padding": 1
  }
}
```

- `padding: 0`: No padding
- `padding: 1`: 1 line padding above and below (default)
- `padding: 2`: 2 lines padding above and below

### Troubleshooting Checklist

Steps to resolve statusline display issues:

1. ✓ Check `DEBUG_STATUSLINE` environment variable
   ```bash
   echo $DEBUG_STATUSLINE  # Should be unset or 0 by default
   ```

2. ✓ Verify `.moai/status_line.sh` file
   ```bash
   grep "DEBUG_STATUSLINE=" ~/.moai/status_line.sh
   # Should show: DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0}
   ```

3. ✓ Enable debugging explicitly (only if needed)
   ```bash
   export DEBUG_STATUSLINE=1
   # Debug information will now be recorded
   ```

4. ✓ Configure padding
   ```json
   {
     "statusLine": {
       "padding": 1
     }
   }
   ```

5. ✓ Restart Claude Code session

## Reference

### Official Documentation

- [Claude Code Statusline Official Documentation](https://code.claude.com/docs/en/statusline) — Claude Code's statusline contract and JSON schema

### moai-adk-go Internals

- **Package**: `internal/statusline/`
  - `types.go`: StdinData, PRInfo, RepoInfo struct definitions
  - `builder.go`: Segment creation logic
  - `renderer.go`: Color coding and final rendering

- **Template**: `.moai/status_line.sh.tmpl`
  - Renderer invocation and execution logic

- **Configuration**: `.moai/config/sections/statusline.yaml`
  - Segment enable/disable settings

### Related SPEC

- **[SPEC-V3R5-STATUSLINE-V2145-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-V3R5-STATUSLINE-V2145-001/spec.md)**
  - M1: Fix statusline disappearing issue
  - M2: Add v2.1.145 PR segment
  - M3: Documentation (this page)
