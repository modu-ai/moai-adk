---
title: Statusline System — Complete 3-line Layout Guide
weight: 78
draft: false
---

A **custom statusline system** for Claude Code ↔ moai-adk-go integration. Since Claude Code v2.1.139 (effort/thinking) and v2.1.145 (workspace.repo + pr), stdin JSON exposes rich session context that this statusline visualizes.

> MoAI workflows are PR-centric. Every SPEC produces a plan-PR → run-PR → sync-PR cycle, so surfacing the current PR number, review state, context usage, and handoff hint directly in the statusline dramatically improves development efficiency.

## Overview

### Final Layout (3-line v3)

```
🤖 Opus 4.7 │ 🧠 xhigh·t │ 🔅 v2.1.146 │ 🗿 v2.20.0-rc1 │ ⏳ 4h 52m │ 💬 MoAI
🪫 CW: ███████░░░ 72% (⚠️/clear) │ 🔋 5H: █████░░░░░ 56% (46m) │ 🔋 7D: █░░░░░░░░░ 13% (May 28)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk (🅱️ main ↑5 +2) │ 💾 +0 M1 ?1 │ 💌 PR #1234 (⌥approved)
```

- **Line 1 (Info)**: model · effort/thinking · Claude Code version · MoAI version · session time · output style
- **Line 2 (Usage bars)**: CW (context window) · 5H (rolling) · 7D (rolling) — each bar shows emoji + label + bar + % + reset info
- **Line 3 (Git/PR)**: directory · combined repo+branch · git status · active SPEC task · PR info

### Data Flow

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData parsing)
    ↓
internal/statusline/builder.go (CollectMemory, CollectMetrics, etc.)
    ↓
internal/statusline/renderer.go (3-line v3 layout)
    ↓
.moai/status_line.sh → terminal display
```

## Line 1 — Info (7 segments)

### 🤖 Model

- **Format**: `🤖 <model display name>`
- **Data source**: stdin `model.display_name` (or string shorthand)
- **Examples**: `🤖 Opus 4.7`, `🤖 Sonnet 4.6`, `🤖 Haiku 4.5`
- **Hidden when**: `model` field absent or `data.Metrics.Model == ""`
- **Segment key**: `model`

### 🧠 Effort / Thinking

- **Format**: `🧠 <level>[·t]`
- **Data source**: stdin `effort.level` + `thinking.enabled` (Claude Code v2.1.139+)
- **Level values**: `low` / `medium` / `high` / `xhigh` / `max`
- **`·t` suffix**: appended when `thinking.enabled == true` (extended reasoning active)
- **Examples**:
  - `🧠 xhigh·t` (xhigh effort + thinking active)
  - `🧠 high` (high effort, no thinking)
  - `·t` (effort absent + only thinking active)
- **Hidden when**: both `effort` + `thinking` absent (including empty effort.level)
- **Segment key**: `effort_thinking`

### 🔅 Claude Code Version

- **Format**: `🔅 v<version>` (default mode) or `🔅 cc v<version>` (full mode)
- **Data source**: stdin `version` string
- **Example**: `🔅 v2.1.146`
- **Hidden when**: `version` empty string
- **Segment key**: `claude_version`

### 🗿 MoAI Version

- **Format**: `🗿 v<current>` or `🗿 v<current> -> 🗿 v<latest>` when update available
- **Data source**: `.moai/config/sections/system.yaml` `moai.version` + background update checker
- **Examples**:
  - `🗿 v2.20.0-rc1` (latest)
  - `🗿 v2.18.0 -> 🗿 v2.20.0-rc1` (update recommended)
- **Segment key**: `moai_version`

### ⏳ Session Time

- **Format**: `⏳ <X>h <Y>m` (≥1h) / `⏳ <X>m` (<1h) / `⏳ <X>d <Y>h` (≥24h)
- **Data source**: stdin `cost.total_duration_ms`
- **Examples**: `⏳ 4h 52m`, `⏳ 35m`, `⏳ 1d 3h`
- **Segment key**: `session_time`

### 💬 Output Style

- **Format**: `💬 <style name>`
- **Data source**: stdin `output_style.name`
- **Examples**: `💬 MoAI`, `💬 R2-D2`, `💬 default`
- **Hidden when**: `output_style.name` empty string
- **Segment key**: `output_style`

## Line 2 — Usage Bars (3 segments)

### 🪫/🔋 CW (Context Window)

- **Format**: `<icon> CW: <bar> <pct>% [(⚠️/clear)]`
- **Data source**:
  - bar: `context_window.context_window_size` × auto-compact threshold (default 85%) → scaled budget
  - percentage: `context_window.used_percentage` (pre-calculated) or sum of `current_usage` tokens
  - (⚠️/clear) gated by: `shouldShowHandoffGuide(data) == true`
- **Icons**:
  - 🔋 (normal, <50% scaled)
  - 🪫 (warning, 50-79% scaled)
  - 🪫 (critical, ≥80% scaled, color added)
- **(⚠️/clear) handoff suffix**:
  - 1M context model (Opus 4.7): used_percentage ≥50% (against raw context_window_size)
  - 200K context model (Sonnet/Haiku): used_percentage ≥90%
  - Meaning: recommend running `/clear` before next turn + use paste-ready resume message
- **Example**: `🪫 CW: ███████░░░ 72% (⚠️/clear)`
- **Segment key**: `context`

### 🔋 5H (5-hour rolling rate limit)

- **Format**: `🔋 5H: <bar> <pct>% [(<reset>)]`
- **Data source**: stdin `rate_limits.five_hour.{used_percentage, resets_at}`
- **Reset format**:
  - <60 min: `(Nm)` (e.g., `(47m)`)
  - <24 hours: `(Nh Nm)` (e.g., `(2h 15m)`)
  - ≥24 hours: `(Mon DD)` (e.g., `(May 28)`)
- **Example**: `🔋 5H: █████░░░░░ 56% (47m)`
- **No data**: `rate_limits.five_hour == null` → bar 0%, reset `(rolling)`
- **Segment key**: `usage_5h`

### 🔋 7D (7-day rolling rate limit)

- **Format**: `🔋 7D: <bar> <pct>% [(<reset>)]`
- **Data source**: stdin `rate_limits.seven_day.{used_percentage, resets_at}`
- **Reset format**: `(Mon DD)` (absolute date)
- **Example**: `🔋 7D: █░░░░░░░░░ 13% (May 28)`
- **Segment key**: `usage_7d`

## Line 3 — Git / PR (5 segments)

### 📁 Directory

- **Format**: `📁 <directory name>`
- **Data source**: stdin `workspace.project_dir` (basename) or `cwd`
- **Examples**: `📁 moai-adk-go`, `📁 my-project`
- **Hidden when**: `data.Directory` empty string
- **Segment key**: `directory`

### 🔀 Repo + Branch (combined segment)

- **Format**: `🔀 <owner>/<name> (🅱️ <branch>[ ↑N][ ↓N][ +N])`
- **Data source**:
  - `🔀 owner/name`: stdin `workspace.repo.{host, owner, name}` (Claude Code v2.1.145+)
  - `🅱️ branch`: local git `branch --show-current`
  - `↑N`: ahead count (vs origin/<branch>)
  - `↓N`: behind count
  - `+N`: dirty count = Modified + Staged + Untracked
- **Examples**:
  - `🔀 modu-ai/moai-adk (🅱️ main ↑3 +2)` (repo + branch + ahead + dirty)
  - `🔀 modu-ai/moai-adk (🅱️ main)` (clean branch, no ahead)
  - `🔀 (🅱️ feat/auth ↑2 ↓1 +6)` (fallback: no repo info)
- **Hidden when**:
  - branch empty string → whole segment hidden
  - repo nil → fallback (only branch shown in parens)
- **Worktree mode**: if `worktree` segment enabled, `[WT] ` prefix prepended to branch
- **Segment key**: `git_branch` (combined)

### 💾 Git Status

- **Format**: `💾 +<staged> M<modified> ?<untracked>`
- **Data source**: local git `git status --porcelain` parsing
- **Example**: `💾 +0 M1 ?1` (staged 0, modified 1, untracked 1)
- **Hidden when**: git unavailable
- **Note**: previous mailbox quartet emoji (📬/📫/📪/📭) retired in favor of unified 💾
- **Segment key**: `git_status`

### 📋 Task (Active SPEC workflow)

- **Format**: `📋 [<command> <SPEC-ID>-<stage>]`
- **Data source**: `~/.moai/state/last-session-state.json` `active_task` field (populated only when that file exists)
- **Example**: `📋 [/moai run SPEC-V3R5-STATUSLINE-001-implement]`
- **Hidden when**: file absent or `active_task` nil → segment hidden
- **Segment key**: `task` (opt-in default off)

### 💌 PR (Active GitHub Pull Request)

- **Format**: `💌 PR #<number> (⌥<review_state>)` (when state present) / `💌 PR #<number>` (when state empty)
- **Data source**: stdin `pr.{number, url, review_state}` (Claude Code v2.1.146+)
- **Review state values**: `approved` / `pending` / `changes_requested` / `draft` / others (raw passthrough)
- **Color coding** (review_state portion):
  - `approved`: green (Success)
  - `pending`: yellow (Warning)
  - `changes_requested`: red (Error)
  - `draft`: gray (Muted)
  - other: no color (raw passthrough)
- **Examples**:
  - `💌 PR #1234 (⌥approved)` (green)
  - `💌 PR #1023 (⌥pending)` (yellow)
  - `💌 PR #7 (⌥changes_requested)` (red)
  - `💌 PR #99 (⌥draft)` (gray)
  - `💌 PR #100` (no state)
- **Hidden when**:
  - `pr` field absent (no PR or Claude Code below v2.1.145)
  - `pr.number == 0`
  - `SegmentPR` config explicitly false
- **Segment key**: `pr` (default on per v2.20.0-rc1)

## Configuration

### Basic Structure

Segment activation is managed in `.moai/config/sections/statusline.yaml`:

```yaml
statusline:
  mode: default              # default | full
  theme: catppuccin-mocha    # color theme
  preset: custom             # full | minimal | custom
  segments:
    # Line 1
    model: true
    effort_thinking: true
    claude_version: true
    moai_version: true
    session_time: true
    output_style: true

    # Line 2
    context: true
    usage_5h: true
    usage_7d: true

    # Line 3
    directory: true
    git_branch: true       # combined repo+branch
    git_status: true
    task: true             # opt-in default off in older versions
    pr: true               # default on per v2.20.0-rc1
    worktree: false
```

### Segment Activation Matrix

| Segment | Line | Default Active | stdin field |
|---------|------|---------------|-------------|
| `model` | L1 | ✅ | `model.display_name` |
| `effort_thinking` | L1 | ✅ | `effort.level` + `thinking.enabled` |
| `claude_version` | L1 | ✅ | `version` |
| `moai_version` | L1 | ✅ | (local config) |
| `session_time` | L1 | ✅ | `cost.total_duration_ms` |
| `output_style` | L1 | ✅ | `output_style.name` |
| `context` | L2 | ✅ | `context_window.*` |
| `usage_5h` | L2 | ✅ | `rate_limits.five_hour.*` |
| `usage_7d` | L2 | ✅ | `rate_limits.seven_day.*` |
| `directory` | L3 | ✅ | `workspace.project_dir` |
| `git_branch` (combined) | L3 | ✅ | `workspace.repo.*` + local git |
| `git_status` | L3 | ✅ | local git |
| `task` | L3 | ⚠️ opt-in | `~/.moai/state/last-session-state.json` |
| `pr` | L3 | ✅ (v2.20.0-rc1+) | `pr.*` (Claude Code v2.1.146+) |
| `worktree` | L3 | ❌ opt-in | `workspace.git_worktree` |

## Handoff Guide — (⚠️/clear) Threshold

The CW bar's `(⚠️/clear)` suffix activates when context usage crosses model-specific thresholds. It is a visual marker that helps prevent SSE stall risk and prompts the user to leverage a paste-ready resume message.

| Model Class | Context Window | Threshold | When Recommended |
|-------------|----------------|-----------|------------------|
| **1M context** (Opus 4.7) | 1,000,000 tokens | **≥50%** | ~500K tokens used |
| **200K context** (Sonnet, Haiku) | 200,000 tokens | **≥90%** | ~180K tokens used |
| Other / unknown | — | not shown | (safety default) |

> Thresholds are enforced by `internal/statusline/renderer.go shouldShowHandoffGuide()`. These match the HARD rule in `.claude/rules/moai/workflow/context-window-management.md`.

User flow when activated:
1. `(⚠️/clear)` marker appears
2. Save in-flight work to `progress.md` or equivalent
3. Orchestrator generates paste-ready resume message (session-handoff.md 6-block format)
4. Run `/clear`, then paste the resume message
5. Continue work in the fresh session

## stdin JSON Schema Reference

For the complete list of stdin JSON fields Claude Code passes to the statusline script, see the [official docs Available data](https://code.claude.com/docs/en/statusline#available-data). moai-adk-go consumes the following fields:

```json
{
  "session_id": "abc...",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/path/to/cwd",
  "model": {"id": "claude-opus-4-7", "display_name": "Opus 4.7"},
  "workspace": {
    "current_dir": "...",
    "project_dir": "...",
    "git_worktree": "feature-xyz",
    "repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"}
  },
  "version": "2.1.146",
  "output_style": {"name": "MoAI"},
  "cost": {
    "total_cost_usd": 1.234,
    "total_duration_ms": 17520000,
    "total_lines_added": 156,
    "total_lines_removed": 23
  },
  "context_window": {
    "used_percentage": 62,
    "context_window_size": 1000000,
    "total_input_tokens": 620000,
    "total_output_tokens": 0,
    "current_usage": {
      "input_tokens": 8500,
      "output_tokens": 1200,
      "cache_creation_input_tokens": 5000,
      "cache_read_input_tokens": 605300
    }
  },
  "exceeds_200k_tokens": true,
  "effort": {"level": "xhigh"},
  "thinking": {"enabled": true},
  "rate_limits": {
    "five_hour": {"used_percentage": 56, "resets_at": 1779286800},
    "seven_day": {"used_percentage": 13, "resets_at": 1779832400}
  },
  "pr": {
    "number": 1234,
    "url": "https://github.com/modu-ai/moai-adk/pull/1234",
    "review_state": "approved"
  }
}
```

## Version History

- **v2.20.0-rc1 layout v3** (2026-05-22): 3-line layout redesign — combined repo+branch segment, directory on L3 head, `🪫 CW:` emoji moved to front, `(⚠️/clear)` handoff suffix, unified `💾` git status, `💌 PR #N (⌥state)` format
- **v2.20.0-rc1 STATUSLINE-STDINFIELDS-001** (2026-05-21): added `workspace.repo` + `exceeds_200k_tokens` + `pr` stdin field mappings, tightened 1M context handoff threshold 75% → 50%
- **v2.20.0-rc1 STATUSLINE-V2145-001** (2026-05-20): PR segment added (v2.1.145+ stdin), 4-locale docs sync
- **v2.1.139** (Claude Code): `effort.level` + `thinking.enabled` added to stdin JSON
- **v2.1.146** (Claude Code): `workspace.repo` + `pr` added to stdin JSON

## Troubleshooting

### PR not showing in statusline

- Verify Claude Code version: needs `🔅 v2.1.146` or higher (v2.1.145 does not include `pr` in stdin)
- Verify current branch has an open PR: `gh pr view`
- Check `statusline.yaml` for explicit `pr: false`

### (⚠️/clear) not appearing

- 1M context model: used_percentage below 50% → expected (threshold not yet reached)
- 200K context model: used_percentage below 90% → expected
- Above threshold but no marker: verify `MemoryData.ContextWindowSize` mapping in `shouldShowHandoffGuide` (possible boundary defect)

### No colors displayed

- Verify terminal supports ANSI 256-color
- Confirm `theme: catppuccin-mocha` is appropriate for your environment
- Check whether `NO_COLOR=1` env var is set

### Verification Command

```bash
# Verify actual statusline output with stdin fixture
NOW=$(date +%s)
echo '{"session_id":"test","model":{"display_name":"Opus 4.7"},"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}},"version":"2.1.146","output_style":{"name":"MoAI"},"context_window":{"used_percentage":62,"context_window_size":1000000},"exceeds_200k_tokens":true,"effort":{"level":"xhigh"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":56,"resets_at":'$((NOW + 2820))'},"seven_day":{"used_percentage":13,"resets_at":'$((NOW + 518400))'}},"cost":{"total_duration_ms":17520000},"pr":{"number":1234,"url":"https://github.com/modu-ai/moai-adk/pull/1234","review_state":"approved"}}' | moai statusline
```

## Related Documentation

- [Settings JSON](/advanced/settings-json) — Claude Code `statusLine` field configuration
