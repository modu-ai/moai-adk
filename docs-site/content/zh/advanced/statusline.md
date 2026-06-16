---
title: Statusline 系统 — 3-line 布局完整指南
weight: 78
draft: false
---

Claude Code 与 moai-adk-go 集成的 **自定义 statusline 系统**。从 Claude Code v2.1.139 开始 effort/thinking,v2.1.145 开始 workspace.repo + pr 字段已添加到 stdin JSON,可以显示丰富的会话上下文。

> MoAI 工作流是 PR 中心的。每个 SPEC 都会生成 plan-PR → run-PR → sync-PR 周期,因此在 statusline 中即时显示当前 PR 号码、审查状态、上下文使用率和 handoff 建议可显著提高开发效率。

## 概述

### 最终布局 (3-line v3)

```
🤖 Opus 4.7 │ 🧠 xhigh·t │ 🔅 v2.1.146 │ 🗿 v2.20.0-rc1 │ ⏳ 4h 52m │ 💬 MoAI
🪫 CW: ███████░░░ 72% (⚠️/clear) │ 🔋 5H: █████░░░░░ 56% (46m) │ 🔋 7D: █░░░░░░░░░ 13% (May 28)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk (🅱️ main ↑5 +2) │ 💾 +0 M1 ?1 │ 💌 PR #1234 (⌥approved)
```

- **Line 1 (Info)**: 模型 · effort/thinking · Claude Code 版本 · MoAI 版本 · 会话时间 · output style
- **Line 2 (Usage bars)**: CW (context window) · 5H (rolling) · 7D (rolling) — 每个 bar 显示 emoji + label + bar + % + reset 信息
- **Line 3 (Git/PR)**: 目录 · 仓库+分支组合 · git status · 活动 SPEC task · PR 信息

### 数据流

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData 解析)
    ↓
internal/statusline/builder.go (CollectMemory, CollectMetrics, etc.)
    ↓
internal/statusline/renderer.go (3-line v3 布局)
    ↓
.moai/status_line.sh → 终端显示
```

## Line 1 — Info (7 个段)

### 🤖 Model

- **格式**: `🤖 <model display name>`
- **数据源**: stdin `model.display_name` (或 string shorthand)
- **示例**: `🤖 Opus 4.7`, `🤖 Sonnet 4.6`, `🤖 Haiku 4.5`
- **隐藏条件**: `model` 字段缺失或 `data.Metrics.Model == ""`
- **段键**: `model`

### 🧠 Effort / Thinking

- **格式**: `🧠 <level>[·t]`
- **数据源**: stdin `effort.level` + `thinking.enabled` (Claude Code v2.1.139+)
- **Level 值**: `low` / `medium` / `high` / `xhigh` / `max`
- **`·t` 后缀**: `thinking.enabled == true` 时添加 (extended reasoning 启用)
- **示例**:
  - `🧠 xhigh·t` (xhigh effort + thinking 启用)
  - `🧠 high` (high effort,无 thinking)
  - `·t` (effort 缺失 + 仅 thinking 启用)
- **隐藏条件**: `effort` + `thinking` 都缺失 (包括 effort.level 空字符串)
- **段键**: `effort_thinking`

### 🔅 Claude Code 版本

- **格式**: `🔅 v<version>` (default) 或 `🔅 cc v<version>` (full mode)
- **数据源**: stdin `version` 字符串
- **示例**: `🔅 v2.1.146`
- **隐藏条件**: `version` 空字符串
- **段键**: `claude_version`

### 🗿 MoAI 版本

- **格式**: `🗿 v<current>` 或更新可用时 `🗿 v<current> -> 🗿 v<latest>`
- **数据源**: `.moai/config/sections/system.yaml` `moai.version` + 后台更新检查器
- **示例**:
  - `🗿 v2.20.0-rc1` (最新)
  - `🗿 v2.18.0 -> 🗿 v2.20.0-rc1` (建议更新)
- **段键**: `moai_version`

### ⏳ 会话时间

- **格式**: `⏳ <X>h <Y>m` (≥1h) / `⏳ <X>m` (<1h) / `⏳ <X>d <Y>h` (≥24h)
- **数据源**: stdin `cost.total_duration_ms`
- **示例**: `⏳ 4h 52m`, `⏳ 35m`, `⏳ 1d 3h`
- **段键**: `session_time`

### 💬 Output Style

- **格式**: `💬 <style name>`
- **数据源**: stdin `output_style.name`
- **示例**: `💬 MoAI`, `💬 R2-D2`, `💬 default`
- **隐藏条件**: `output_style.name` 空字符串
- **段键**: `output_style`

## Line 2 — Usage Bars (3 个段)

### 🪫/🔋 CW (Context Window)

- **格式**: `<icon> CW: <bar> <pct>% [(⚠️/clear)]`
- **数据源**:
  - bar: `context_window.context_window_size` × auto-compact threshold (default 85%) → scaled budget
  - 百分比: `context_window.used_percentage` (预计算) 或 `current_usage` tokens 总和
  - (⚠️/clear) 启用条件: `shouldShowHandoffGuide(data) == true`
- **emoji**:
  - 🔋 (正常, <50% scaled)
  - 🪫 (警告, 50-79% scaled)
  - 🪫 (危险, ≥80% scaled, 颜色添加)
- **(⚠️/clear) handoff 后缀**:
  - 1M context 模型 (Opus 4.7): used_percentage ≥50% (基于 raw context_window_size)
  - 200K context 模型 (Sonnet/Haiku): used_percentage ≥90%
  - 含义: 下一 turn 开始前建议 `/clear` + 利用 paste-ready resume message
- **示例**: `🪫 CW: ███████░░░ 72% (⚠️/clear)`
- **段键**: `context`

### 🔋 5H (5小时 rolling rate limit)

- **格式**: `🔋 5H: <bar> <pct>% [(<reset>)]`
- **数据源**: stdin `rate_limits.five_hour.{used_percentage, resets_at}`
- **Reset 格式**:
  - <60 分: `(Nm)` (例: `(47m)`)
  - <24 小时: `(Nh Nm)` (例: `(2h 15m)`)
  - ≥24 小时: `(Mon DD)` (例: `(May 28)`)
- **示例**: `🔋 5H: █████░░░░░ 56% (47m)`
- **数据缺失**: `rate_limits.five_hour == null` → bar 0%, reset `(rolling)`
- **段键**: `usage_5h`

### 🔋 7D (7天 rolling rate limit)

- **格式**: `🔋 7D: <bar> <pct>% [(<reset>)]`
- **数据源**: stdin `rate_limits.seven_day.{used_percentage, resets_at}`
- **Reset 格式**: `(Mon DD)` (绝对日期)
- **示例**: `🔋 7D: █░░░░░░░░░ 13% (May 28)`
- **段键**: `usage_7d`

## Line 3 — Git / PR (5 个段)

### 📁 Directory

- **格式**: `📁 <directory name>`
- **数据源**: stdin `workspace.project_dir` (basename) 或 `cwd`
- **示例**: `📁 moai-adk-go`, `📁 my-project`
- **隐藏条件**: `data.Directory` 空字符串
- **段键**: `directory`

### 🔀 Repo + Branch (组合段)

- **格式**: `🔀 <owner>/<name> (🅱️ <branch>[ ↑N][ ↓N][ +N])`
- **数据源**:
  - `🔀 owner/name`: stdin `workspace.repo.{host, owner, name}` (Claude Code v2.1.145+)
  - `🅱️ branch`: 本地 git `branch --show-current`
  - `↑N`: ahead 计数 (对比 origin/<branch>)
  - `↓N`: behind 计数
  - `+N`: dirty 计数 = Modified + Staged + Untracked
- **示例**:
  - `🔀 modu-ai/moai-adk (🅱️ main ↑3 +2)` (repo + branch + ahead + dirty)
  - `🔀 modu-ai/moai-adk (🅱️ main)` (clean branch, no ahead)
  - `🔀 (🅱️ feat/auth ↑2 ↓1 +6)` (repo 信息缺失 fallback)
- **隐藏条件**:
  - branch 空字符串 → 整段隐藏
  - repo nil 时 fallback (仅在括号内显示 branch)
- **Worktree 模式**: `worktree` 段启用时 branch 加 `[WT] ` prefix
- **段键**: `git_branch` (combined)

### 💾 Git Status

- **格式**: `💾 +<staged> M<modified> ?<untracked>`
- **数据源**: 本地 git `git status --porcelain` 解析
- **示例**: `💾 +0 M1 ?1` (staged 0, modified 1, untracked 1)
- **隐藏条件**: git 不可用
- **注意**: 之前的 mailbox 四种 emoji (📬/📫/📪/📭) 已弃用,统一使用 💾
- **段键**: `git_status`

### 📋 Task (活动 SPEC workflow)

- **格式**: `📋 [<command> <SPEC-ID>-<stage>]`
- **数据源**: `~/.moai/state/last-session-state.json` `active_task` 字段 (仅在该文件创建时显示)
- **示例**: `📋 [/moai run SPEC-V3R5-STATUSLINE-001-implement]`
- **隐藏条件**: 文件缺失或 `active_task` nil → 段隐藏
- **段键**: `task` (opt-in default off)

### 💌 PR (活动 GitHub Pull Request)

- **格式**: `💌 PR #<number> (⌥<review_state>)` (state 存在时) / `💌 PR #<number>` (state 空字符串)
- **数据源**: stdin `pr.{number, url, review_state}` (Claude Code v2.1.146+)
- **Review state 值**: `approved` / `pending` / `changes_requested` / `draft` / 其他 (raw passthrough)
- **颜色编码** (review_state 部分):
  - `approved`: 绿色 (Success)
  - `pending`: 黄色 (Warning)
  - `changes_requested`: 红色 (Error)
  - `draft`: 灰色 (Muted)
  - 其他: 无颜色 (raw passthrough)
- **示例**:
  - `💌 PR #1234 (⌥approved)` (绿色)
  - `💌 PR #1023 (⌥pending)` (黄色)
  - `💌 PR #7 (⌥changes_requested)` (红色)
  - `💌 PR #99 (⌥draft)` (灰色)
  - `💌 PR #100` (无 state)
- **隐藏条件**:
  - `pr` 字段缺失 (无 PR 或 v2.1.145 以下)
  - `pr.number == 0`
  - `SegmentPR` config 明确 false
- **段键**: `pr` (default on per v2.20.0-rc1)

## 配置

### 基本结构

`.moai/config/sections/statusline.yaml` 中管理段启用:

```yaml
statusline:
  mode: default              # default | full
  theme: catppuccin-mocha    # 颜色主题
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

### 段启用矩阵

| 段 | 行 | 默认启用 | stdin field |
|---------|------|----------|-------------|
| `model` | L1 | ✅ | `model.display_name` |
| `effort_thinking` | L1 | ✅ | `effort.level` + `thinking.enabled` |
| `claude_version` | L1 | ✅ | `version` |
| `moai_version` | L1 | ✅ | (本地 config) |
| `session_time` | L1 | ✅ | `cost.total_duration_ms` |
| `output_style` | L1 | ✅ | `output_style.name` |
| `context` | L2 | ✅ | `context_window.*` |
| `usage_5h` | L2 | ✅ | `rate_limits.five_hour.*` |
| `usage_7d` | L2 | ✅ | `rate_limits.seven_day.*` |
| `directory` | L3 | ✅ | `workspace.project_dir` |
| `git_branch` (combined) | L3 | ✅ | `workspace.repo.*` + 本地 git |
| `git_status` | L3 | ✅ | 本地 git |
| `task` | L3 | ⚠️ opt-in | `~/.moai/state/last-session-state.json` |
| `pr` | L3 | ✅ (v2.20.0-rc1+) | `pr.*` (Claude Code v2.1.146+) |
| `worktree` | L3 | ❌ opt-in | `workspace.git_worktree` |

## Handoff Guide — (⚠️/clear) 建议标准

CW bar 的 `(⚠️/clear)` 后缀在上下文使用量超过模型特定阈值时启用。这是预防 SSE stall 风险并建议使用 paste-ready resume message 的视觉标记。

| 模型类别 | Context Window | 阈值 | 建议时机 |
|------------|----------------|------|----------|
| **1M context** (Opus 4.7) | 1,000,000 tokens | **≥50%** | ~500K tokens 使用 |
| **200K context** (Sonnet, Haiku) | 200,000 tokens | **≥90%** | ~180K tokens 使用 |
| 其他 / 未知 | — | 不显示 | (安全 default) |

> 阈值由 `internal/statusline/renderer.go shouldShowHandoffGuide()` 函数强制执行。这些阈值与 `.claude/rules/moai/workflow/context-window-management.md` HARD rule 一致。

启用时用户流程:
1. `(⚠️/clear)` marker 显示
2. 保存进行中的工作到 `progress.md` 等
3. orchestrator 生成 paste-ready resume message (session-handoff.md 6-block 格式)
4. 执行 `/clear` 后粘贴 resume message
5. 在新会话中继续工作

## stdin JSON 模式参考

Claude Code 传递给 statusline 脚本的 stdin JSON 完整字段列表请参考 [官方 docs Available data](https://code.claude.com/docs/en/statusline#available-data)。moai-adk-go 使用以下字段:

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

## 版本历史

- **v2.20.0-rc1 layout v3** (2026-05-22): 3-line 布局重新设计 — repo+branch 组合段、directory L3 head、`🪫 CW:` emoji 前移、`(⚠️/clear)` handoff 后缀、`💾` git status 统一、`💌 PR #N (⌥state)` 格式
- **v2.20.0-rc1 STATUSLINE-STDINFIELDS-001** (2026-05-21): 添加 `workspace.repo` + `exceeds_200k_tokens` + `pr` stdin 字段映射、1M context handoff threshold 75% → 50%
- **v2.20.0-rc1 STATUSLINE-V2145-001** (2026-05-20): 添加 PR segment (v2.1.145+ stdin)、4-locale docs 同步
- **v2.1.139** (Claude Code): `effort.level` + `thinking.enabled` 添加到 stdin JSON
- **v2.1.146** (Claude Code): `workspace.repo` + `pr` 添加到 stdin JSON

## 故障排除

### Statusline 中 PR 未显示

- 验证 Claude Code 版本: 需要 `🔅 v2.1.146` 以上 (v2.1.145 stdin 不包含 `pr` 字段)
- 验证当前 branch 有 OPEN PR: `gh pr view`
- 检查 `statusline.yaml` 中是否明确 `pr: false`

### (⚠️/clear) 未出现

- 1M context 模型: used_percentage 低于 50% → 正常 (尚未达到阈值)
- 200K context 模型: used_percentage 低于 90% → 正常
- 高于阈值但未显示: 验证 `shouldShowHandoffGuide` 函数的 `MemoryData.ContextWindowSize` 映射 (可能的 boundary defect)

### 颜色未显示

- 验证终端是否支持 ANSI 256-color
- 确认 `theme: catppuccin-mocha` 是否适合环境
- 检查是否设置了 `NO_COLOR=1` 环境变量

### 验证命令

```bash
# 使用 stdin fixture 验证 statusline 实际输出
NOW=$(date +%s)
echo '{"session_id":"test","model":{"display_name":"Opus 4.7"},"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}},"version":"2.1.146","output_style":{"name":"MoAI"},"context_window":{"used_percentage":62,"context_window_size":1000000},"exceeds_200k_tokens":true,"effort":{"level":"xhigh"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":56,"resets_at":'$((NOW + 2820))'},"seven_day":{"used_percentage":13,"resets_at":'$((NOW + 518400))'}},"cost":{"total_duration_ms":17520000},"pr":{"number":1234,"url":"https://github.com/modu-ai/moai-adk/pull/1234","review_state":"approved"}}' | moai statusline
```

## `/cd` 缓存保留目录切换 (CC 2.1.169+)

Claude Code 2.1.169+ 提供了 `/cd <path>` 命令，可在 **保留提示缓存的同时** 更改会话的工作目录 — statusline 的 `cwd` 字段更新以反映新目录，但进行中的推理上下文不会重建。这是相对于打开新终端会话的缓存保留替代方案：`/cd` 保留累积的上下文，而新终端从零开始冷启动。当 statusline 显示您想在不丢失上下文的情况下离开的 `cwd` 时（例如会话中切换到 L2 worktree），`/cd` 是更低摩擦的路径。恢复模式集成请参阅 [会话交接](/zh/workflow-commands/moai-sync)。

## 相关文档

- [Settings JSON](/advanced/settings-json) — Claude Code `statusLine` 字段配置
