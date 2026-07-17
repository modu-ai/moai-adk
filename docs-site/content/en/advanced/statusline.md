---
title: Statusline System — Complete 3-line Layout Guide
weight: 78
draft: false
---

A **custom statusline system** for the Claude Code and moai-adk-go integration. Tokenomics starts with measurement — context usage (CW%), prompt cache hit rate, and rate-limit burn rates are displayed permanently at the bottom of the terminal, letting you read your token posture at a glance. From Claude Code v2.1.139 effort/thinking, and from v2.1.145 the workspace.repo + pr fields, are added to the stdin JSON, enabling a rich context display.

> The MoAI workflow is PR-centric. Every SPEC produces a plan-PR → run-PR → sync-PR cycle, so surfacing the current PR number + review state + context usage + handoff advice directly in the statusline greatly improves development efficiency.

## Overview

### Final Layout (3-line v3)

```
🤖 Opus 4.7 │ 🧠 xhigh·t │ 💾 67% │ 🔅 v2.1.146 │ 🗿 v3.0.0 │ ⏳ 4h 52m │ 💬 MoAI
🪫 CW: ███████░░░ 72% (⚠️/clear) │ 🔋 5H: █████░░░░░ 56% (46m) │ 🔋 7D: █░░░░░░░░░ 13% (May 28)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk (🅱️ main ↑5 +2) │ 💾 +0 M1 ?1 │ 💌 PR #1234 (⌥approved)
```

- **Line 1 (Info)**: model · effort/thinking · cache hit rate · Claude Code version · MoAI version · session time · output style
- **Line 2 (Usage bars)**: CW (context window) · 5H (rolling) · 7D (rolling) — each bar is emoji + label + bar + % + reset info
- **Line 3 (Git/PR)**: directory · combined repository+branch · git status · active SPEC task · PR info

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

### Model

- **Format**: `🤖 <model display name>`
- **Data source**: stdin `model.display_name` (or string shorthand)
- **Examples**: `🤖 Opus 4.7`, `🤖 Sonnet 4.6`, `🤖 Haiku 4.5`
- **Hidden when**: the `model` field is absent or `data.Metrics.Model == ""`
- **Segment key**: `model`

### Effort / Thinking

- **Format**: `🧠 <level>[·t]`
- **Data source**: stdin `effort.level` + `thinking.enabled` (Claude Code v2.1.139+)
- **Level values**: `low` / `medium` / `high` / `xhigh` / `max`
- **The `·t` suffix**: added when `thinking.enabled == true` (extended reasoning active)
- **Examples**:
  - `🧠 xhigh·t` (xhigh effort + thinking active)
  - `🧠 high` (high effort, no thinking)
  - `·t` (effort absent + thinking only)
- **Hidden when**: both `effort` and `thinking` are absent (including an empty effort.level string)
- **Segment key**: `effort_thinking`

In always showing the reasoning depth the current session runs at, this segment is also a window for verifying that the model policy is actually being applied.

### Cache Hit Rate

- **Format**: `💾 <N>%` (N = cache_read / (cache_read + cache_creation) × 100, truncated)
- **Data source**: stdin `current_usage.cache_read_tokens` + `current_usage.cache_creation_tokens`
- **Example**: `💾 28%` (cache_read 2000, cache_creation 5000 → 2000/7000)
- **Hidden when**: `current_usage` absent · `cache_creation == 0` (no fresh cache write) · both 0 — silently omitted rather than fabricating a value (graceful degradation)
- **Toggle**: `cache_hit: false` in statusline.yaml → hidden (default-on)
- **Segment key**: `cache_hit`
- **Note**: the same `💾` emoji is also used in Line 3 Git Status (`💾 +N M? ?`) — this segment sits on Line 1 and is distinguished by the percentage format. Prompt-cache reuse monitoring (SPEC-TOKEN-EFFICIENCY-001 P0-2)

The cache hit rate is the effect meter of the context diet — trim the always-loaded instructions and you immediately see this number rise.

### Claude Code Version

- **Format**: `🔅 v<version>` (default) or `🔅 cc v<version>` (full mode)
- **Data source**: the stdin `version` string
- **Example**: `🔅 v2.1.146`
- **Hidden when**: `version` is an empty string
- **Segment key**: `claude_version`

### MoAI Version

- **Format**: `🗿 v<current>` or, when an update is available, `🗿 v<current> -> 🗿 v<latest>`
- **Data source**: `.moai/config/sections/system.yaml` `moai.version` + the background update checker result
- **Examples**:
  - `🗿 v3.0.0` (up to date)
  - `🗿 v2.18.0 -> 🗿 v3.0.0` (update advised)
- **Segment key**: `moai_version`

### Session Time

- **Format**: `⏳ <X>h <Y>m` (≥1h) / `⏳ <X>m` (<1h) / `⏳ <X>d <Y>h` (≥24h)
- **Data source**: stdin `cost.total_duration_ms`
- **Examples**: `⏳ 4h 52m`, `⏳ 35m`, `⏳ 1d 3h`
- **Segment key**: `session_time`

### Output Style

- **Format**: `💬 <style name>`
- **Data source**: stdin `output_style.name`
- **Examples**: `💬 MoAI`, `💬 R2-D2`, `💬 default`
- **Hidden when**: `output_style.name` is an empty string
- **Segment key**: `output_style`

## Line 2 — Usage Bars (3 segments)

### CW (Context Window)

- **Format**: `<icon> CW: <bar> <pct>% [(⚠️/clear)]`
- **Data sources**:
  - bar: `context_window.context_window_size` × auto-compact threshold (default 85%) → scaled budget
  - percentage: `context_window.used_percentage` (precomputed) or the sum of `current_usage` tokens
  - `(⚠️/clear)` activation condition: `shouldShowHandoffGuide(data) == true`
- **Emoji**:
  - `🔋` (normal, <50% scaled)
  - `🪫` (warning, 50-79% scaled)
  - `🪫` (danger, ≥80% scaled, with color)
- **The `(⚠️/clear)` handoff suffix**:
  - 1M-context models (Opus 4.8, GLM-5.2): used_percentage ≥50% (based on raw context_window_size)
  - 200K-context models (Sonnet/Haiku): used_percentage ≥90%
  - Meaning: recommend `/clear` before the next turn + use the paste-ready resume message
- **Example**: `🪫 CW: ███████░░░ 72% (⚠️/clear)`
- **Segment key**: `context`

### 5H (5-hour rolling rate limit)

- **Format**: `🔋 5H: <bar> <pct>% [(<reset>)]`
- **Data source**: stdin `rate_limits.five_hour.{used_percentage, resets_at}`
- **Reset formats**:
  - <60 minutes: `(Nm)` (e.g. `(47m)`)
  - <24 hours: `(Nh Nm)` (e.g. `(2h 15m)`)
  - ≥24 hours: `(Mon DD)` (e.g. `(May 28)`)
- **Example**: `🔋 5H: █████░░░░░ 56% (47m)`
- **Data absent**: `rate_limits.five_hour == null` → bar 0%, reset `(rolling)`
- **Segment key**: `usage_5h`

### 7D (7-day rolling rate limit)

- **Format**: `🔋 7D: <bar> <pct>% [(<reset>)]`
- **Data source**: stdin `rate_limits.seven_day.{used_percentage, resets_at}`
- **Reset format**: `(Mon DD)` (absolute date)
- **Example**: `🔋 7D: █░░░░░░░░░ 13% (May 28)`
- **Segment key**: `usage_7d`

For subscription-plan users, the 5H/7D bars are effectively budget gauges — you can look at these two bars to decide whether to schedule heavy work before the rate limit runs out, or hand it off to GLM workers in CG mode.

## Line 3 — Git / PR (5 segments)

### Directory

- **Format**: `📁 <directory name>`
- **Data source**: stdin `workspace.project_dir` (basename) or `cwd`
- **Examples**: `📁 moai-adk-go`, `📁 my-project`
- **Hidden when**: `data.Directory` is an empty string
- **Segment key**: `directory`

### Repo + Branch (combined segment)

- **Format**: `🔀 <owner>/<name> (🅱️ <branch>[ ↑N][ ↓N][ +N])`
- **Data sources**:
  - `🔀 owner/name`: stdin `workspace.repo.{host, owner, name}` (Claude Code v2.1.145+)
  - `🅱️ branch`: local git `branch --show-current`
  - `↑N`: ahead count (relative to origin/<branch>)
  - `↓N`: behind count
  - `+N`: dirty count = Modified + Staged + Untracked
- **Examples**:
  - `🔀 modu-ai/moai-adk (🅱️ main ↑3 +2)` (repo + branch + ahead + dirty)
  - `🔀 modu-ai/moai-adk (🅱️ main)` (clean branch, no ahead)
  - `🔀 (🅱️ feat/auth ↑2 ↓1 +6)` (fallback when repo info absent)
- **Hidden when**:
  - branch is an empty string → whole segment hidden
  - repo nil → fallback (only the branch in parentheses shown)
- **Worktree mode**: with the `worktree` segment active, the branch gets a `[WT] ` prefix
- **Segment key**: `git_branch` (combined)

### Git Status

- **Format**: `💾 +<staged> M<modified> ?<untracked>`
- **Data source**: parsed from local git `git status --porcelain`
- **Example**: `💾 +0 M1 ?1` (staged 0, modified 1, untracked 1)
- **Hidden when**: git is unavailable
- **Note**: the previous 4-emoji mailbox set (`📬`/`📫`/`📪`/`📭`) is retired; the unified `💾` is used
- **Segment key**: `git_status`

### Task (active SPEC workflow)

- **Format**: `📋 [<command> <SPEC-ID>-<stage>]`
- **Data source**: the `active_task` field of `~/.moai/state/last-session-state.json` (shown only when that file is written)
- **Example**: `📋 [/moai run SPEC-V3R5-STATUSLINE-001-implement]`
- **Hidden when**: file absent or `active_task` nil → segment hidden
- **Segment key**: `task` (opt-in default off)

### PR (active GitHub Pull Request)

- **Format**: `💌 PR #<number> (⌥<review_state>)` (with state) / `💌 PR #<number>` (state empty)
- **Data source**: stdin `pr.{number, url, review_state}` (Claude Code v2.1.146+)
- **Review state values**: `approved` / `pending` / `changes_requested` / `draft` / other (raw passthrough)
- **Color coding** (the review_state portion):
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
  - the `pr` field is absent (no PR, or v2.1.145 and below)
  - `pr.number == 0`
  - `SegmentPR` config explicitly false
- **Segment key**: `pr` (default on per v2.20.0-rc1)

## Configuration

### Basic Structure

Segment activation is managed in `.moai/config/sections/statusline.yaml`.

```yaml
statusline:
  theme: catppuccin-mocha    # color theme
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

### Refresh Interval

The statusline's refresh interval is set via `statusLine.refreshInterval` in `settings.json` (unit: **seconds**, default `10`). This is a Claude Code runtime setting, not `.moai/config/sections/statusline.yaml`. Too low a value increases CPU usage; too high a value delays how quickly context-usage changes are reflected.

```json
{
  "statusLine": {
    "type": "command",
    "command": "$CLAUDE_PROJECT_DIR/.moai/status_line.sh",
    "refreshInterval": 10
  }
}
```

### Segment Activation Matrix

| Segment | Line | Default on | stdin field |
|---------|------|----------|-------------|
| `model` | L1 | ✓ | `model.display_name` |
| `effort_thinking` | L1 | ✓ | `effort.level` + `thinking.enabled` |
| `claude_version` | L1 | ✓ | `version` |
| `moai_version` | L1 | ✓ | (local config) |
| `session_time` | L1 | ✓ | `cost.total_duration_ms` |
| `output_style` | L1 | ✓ | `output_style.name` |
| `context` | L2 | ✓ | `context_window.*` |
| `usage_5h` | L2 | ✓ | `rate_limits.five_hour.*` |
| `usage_7d` | L2 | ✓ | `rate_limits.seven_day.*` |
| `directory` | L3 | ✓ | `workspace.project_dir` |
| `git_branch` (combined) | L3 | ✓ | `workspace.repo.*` + local git |
| `git_status` | L3 | ✓ | local git |
| `task` | L3 | opt-in | `~/.moai/state/last-session-state.json` |
| `pr` | L3 | ✓ (v2.20.0-rc1+) | `pr.*` (Claude Code v2.1.146+) |
| `worktree` | L3 | ✗ opt-in | `workspace.git_worktree` |

## Handoff Guide — the `(⚠️/clear)` Recommendation Criteria

The handoff suffix on the CW bar activates when context usage crosses the model-specific threshold. It is a visual marker that preempts SSE-stall risk and encourages use of the paste-ready resume message, and it operates in **two stages**.

- **soft stage** `(⚠️/clear)`: on reaching the band's soft threshold
- **hard stage** `(🛑/clear!)`: on reaching the auto-compact-aware ceiling (`min(cap, auto-compact-threshold + margin)`) (`internal/statusline/renderer.go`). Because the runtime's auto-compact often pre-empts this ceiling, the hard stage is a rarely-fired upper signal in practice.

| Model class | Context Window | Threshold | Recommended at |
|------------|----------------|--------|----------|
| **1M context** (Opus 4.8) | 1,000,000 tokens | **≥50%** | ~500K tokens used |
| **256K context** (Fable) | 256,000 tokens | **≥90%** | ~230K tokens used |
| **200K context** (Sonnet, Haiku) | 200,000 tokens | **≥90%** | ~180K tokens used |
| Other / unknown | — | Not shown | (safe default) |

> The thresholds are enforced in the handoff-stage decision in `internal/statusline/renderer.go`. They match the HARD rule in `.claude/rules/moai/workflow/context-window-management.md`.

### GLM Context Gauge Correction (Issue #653)

GLM-5.2 is a genuine 1M-context model, but Claude Code reports `context_window_size` based on the Claude slot regardless of provider, so raw telemetry (`effectiveWindow`) can be misreported as ~180K in a GLM session. MoAI corrects this with `ResolveGLMContextWindow` (`internal/statusline/memory.go`) — resolving it from the `MOAI_STATUSLINE_CONTEXT_SIZE` environment variable (explicit override) or the `glm.context_windows` table in `llm.yaml` (glm-5.2 → 1,000,000). In a GLM session, trust the MoAI statusline's CW%, not the raw `effectiveWindow`.

The user flow when activated is as follows.

1. The `(⚠️/clear)` marker appears
2. Save in-progress work to `progress.md` or the like
3. The orchestrator generates a paste-ready resume message (the session-handoff.md 6-block format)
4. Run `/clear` and paste the resume message
5. Continue the work in the new session

## stdin JSON Schema Reference

For the full list of stdin JSON fields Claude Code passes to the statusline script, see [the official docs Available data](https://code.claude.com/docs/en/statusline#available-data). moai-adk-go uses the following fields.

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

- **v2.20.0-rc1 layout v3** (2026-05-22): 3-line layout redesign — combined repo+branch segment, directory at the L3 head, `🪫 CW:` emoji moved forward, `(⚠️/clear)` handoff suffix, unified `💾` git status, `💌 PR #N (⌥state)` format
- **v2.20.0-rc1 STATUSLINE-STDINFIELDS-001** (2026-05-21): added `workspace.repo` + `exceeds_200k_tokens` + `pr` stdin field mappings, 1M-context handoff threshold 75% → 50%
- **v2.20.0-rc1 STATUSLINE-V2145-001** (2026-05-20): PR segment added (v2.1.145+ stdin), 4-locale docs sync
- **v2.1.139** (Claude Code): `effort.level` + `thinking.enabled` added to stdin JSON
- **v2.1.146** (Claude Code): `workspace.repo` + `pr` added to stdin JSON

## Troubleshooting

### PR Not Appearing in the Statusline

- Check the Claude Code version: `🔅 v2.1.146` or later required (v2.1.145 does not include the `pr` field in stdin)
- Confirm the current branch has an OPEN PR: `gh pr view`
- Check whether `pr: false` is explicitly set in `statusline.yaml`

### `(⚠️/clear)` Not Showing

- 1M-context models: used_percentage under 50% → normal (threshold not yet reached)
- 200K-context models: used_percentage under 90% → normal
- Over threshold but not showing: check the `MemoryData.ContextWindowSize` mapping in the `shouldShowHandoffGuide` function (possible boundary defect)

### Colors Not Displaying

- Check that the terminal supports ANSI 256 colors
- Check that `theme: catppuccin-mocha` suits the environment
- Check whether the `NO_COLOR=1` environment variable is set

### Verification Command

```bash
# Verify actual statusline output with a stdin fixture
NOW=$(date +%s)
echo '{"session_id":"test","model":{"display_name":"Opus 4.7"},"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}},"version":"2.1.146","output_style":{"name":"MoAI"},"context_window":{"used_percentage":62,"context_window_size":1000000},"exceeds_200k_tokens":true,"effort":{"level":"xhigh"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":56,"resets_at":'$((NOW + 2820))'},"seven_day":{"used_percentage":13,"resets_at":'$((NOW + 518400))'}},"cost":{"total_duration_ms":17520000},"pr":{"number":1234,"url":"https://github.com/modu-ai/moai-adk/pull/1234","review_state":"approved"}}' | moai statusline
```

## `/cd` Cache-Preserving Directory Switch (CC 2.1.169+)

Claude Code 2.1.169+ provides the `/cd <path>` command, which changes the session's working directory **while preserving the prompt cache** — the statusline's `cwd` field updates to reflect the new directory, but the in-flight reasoning context is not rebuilt. It is the cache-preserving alternative to opening a new terminal session: `/cd` keeps the accumulated context, while a new terminal cold-starts from scratch. When the statusline shows a `cwd` you want to leave without losing context (e.g. switching to an L2 worktree mid-session), `/cd` is the lower-friction path. For resume-pattern integration, see [Session Handoff](/en/workflow-commands/moai-sync).

## Related Documents

- [Settings JSON](/en/advanced/settings-json) — configuring the Claude Code `statusLine` field
