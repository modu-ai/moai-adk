---
title: Statusline 系统 — 3-line 布局完全指南
weight: 78
draft: false
---

这是用于 Claude Code 与 moai-adk-go 集成的 **定制 statusline 系统**。代币经济学始于度量 — 把上下文使用率 (CW%)、提示词缓存命中率、rate limit 消耗率常驻显示在终端底部，让代币运用状态一目了然。自 Claude Code v2.1.139 起 stdin JSON 增加了 effort/thinking 字段，自 v2.1.145 起增加了 workspace.repo + pr 字段，可以显示更丰富的上下文。

> MoAI 工作流以 PR 为中心。每个 SPEC 都会产生 plan-PR → run-PR → sync-PR 循环，因此在 statusline 中即时展示当前 PR 编号 + 评审状态 + 上下文使用率 + handoff 建议，可以大幅提升开发效率。

## 概述

### 最终布局 (3-line v3)

```
🤖 Opus 4.7 │ 🧠 xhigh·t │ 💾 67% │ 🔅 v2.1.146 │ 🗿 v3.0.0 │ ⏳ 4h 52m │ 💬 MoAI
🪫 CW: ███████░░░ 72% (⚠️/clear) │ 🔋 5H: █████░░░░░ 56% (46m) │ 🔋 7D: █░░░░░░░░░ 13% (May 28)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk (🅱️ main ↑5 +2) │ 💾 +0 M1 ?1 │ 💌 PR #1234 (⌥approved)
```

- **Line 1 (Info)**：模型 · effort/thinking · 缓存命中率 · Claude Code 版本 · MoAI 版本 · 会话时长 · output style
- **Line 2 (Usage bars)**：CW (context window) · 5H (rolling) · 7D (rolling) — 每条 bar 为 emoji + label + bar + % + reset 信息
- **Line 3 (Git/PR)**：目录 · 仓库+分支合并段 · git status · 活动 SPEC task · PR 信息

### 数据流

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (解析 StdinData)
    ↓
internal/statusline/builder.go (CollectMemory, CollectMetrics, etc.)
    ↓
internal/statusline/renderer.go (3-line v3 layout)
    ↓
.moai/status_line.sh → 终端显示
```

## Line 1 — Info（7 个段）

### Model

- **格式**：`🤖 <model display name>`
- **数据来源**：stdin `model.display_name`（或 string shorthand）
- **示例**：`🤖 Opus 4.7`、`🤖 Sonnet 4.6`、`🤖 Haiku 4.5`
- **隐藏条件**：无 `model` 字段或 `data.Metrics.Model == ""`
- **段键**：`model`

### Effort / Thinking

- **格式**：`🧠 <level>[·t]`
- **数据来源**：stdin `effort.level` + `thinking.enabled`（Claude Code v2.1.139+）
- **Level 值**：`low` / `medium` / `high` / `xhigh` / `max`
- **`·t` 后缀**：`thinking.enabled == true` 时添加（extended reasoning 激活）
- **示例**：
  - `🧠 xhigh·t`（xhigh effort + thinking 激活）
  - `🧠 high`（high effort，无 thinking）
  - `·t`（无 effort + 仅 thinking 激活）
- **隐藏条件**：`effort` + `thinking` 均缺失（含 effort.level 空字符串）
- **段键**：`effort_thinking`

能够常驻确认当前会话以怎样的推理深度运行，这个段同时也是验证模型策略是否真正生效的窗口。

### 缓存命中率

- **格式**：`💾 <N>%`（N = cache_read / (cache_read + cache_creation) × 100，向下取整）
- **数据来源**：stdin `current_usage.cache_read_tokens` + `current_usage.cache_creation_tokens`
- **示例**：`💾 28%`（cache_read 2000、cache_creation 5000 → 2000/7000）
- **隐藏条件**：无 `current_usage` · `cache_creation == 0`（无 fresh cache write）· 两者均为 0 — 不编造数值而是静默省略 (graceful degradation)
- **开关**：statusline.yaml 中 `cache_hit: false` → 隐藏（默认开启）
- **段键**：`cache_hit`
- **备注**：同一 `💾` emoji 也用于 Line 3 Git Status（`💾 +N M? ?`）— 本段位于 Line 1，以百分比格式区分。用于监控 prompt-cache 复用率 (SPEC-TOKEN-EFFICIENCY-001 P0-2)

缓存命中率是上下文瘦身的效果测量仪 — 缩减常驻加载指令后，可以立即看到这个数字上升。

### Claude Code 版本

- **格式**：`🔅 v<version>`（默认）或 `🔅 cc v<version>`（full 模式）
- **数据来源**：stdin `version` 字符串
- **示例**：`🔅 v2.1.146`
- **隐藏条件**：`version` 为空字符串
- **段键**：`claude_version`

### MoAI 版本

- **格式**：`🗿 v<current>`，或可更新时 `🗿 v<current> -> 🗿 v<latest>`
- **数据来源**：`.moai/config/sections/system.yaml` `moai.version` + 后台 update checker 结果
- **示例**：
  - `🗿 v2.20.0-rc1`（最新）
  - `🗿 v2.18.0 -> 🗿 v2.20.0-rc1`（建议更新）
- **段键**：`moai_version`

### 会话时长

- **格式**：`⏳ <X>h <Y>m`（≥1h）/ `⏳ <X>m`（<1h）/ `⏳ <X>d <Y>h`（≥24h）
- **数据来源**：stdin `cost.total_duration_ms`
- **示例**：`⏳ 4h 52m`、`⏳ 35m`、`⏳ 1d 3h`
- **段键**：`session_time`

### Output Style

- **格式**：`💬 <style name>`
- **数据来源**：stdin `output_style.name`
- **示例**：`💬 MoAI`、`💬 R2-D2`、`💬 default`
- **隐藏条件**：`output_style.name` 为空字符串
- **段键**：`output_style`

## Line 2 — Usage Bars（3 个段）

### CW (Context Window)

- **格式**：`<icon> CW: <bar> <pct>% [(⚠️/clear)]`
- **数据来源**：
  - bar：`context_window.context_window_size` × auto-compact 阈值（默认 85%）→ scaled budget
  - 百分比：`context_window.used_percentage`（预先计算）或 `current_usage` tokens 合计
  - `(⚠️/clear)` 激活条件：`shouldShowHandoffGuide(data) == true`
- **Emoji**：
  - `🔋`（正常，<50% scaled）
  - `🪫`（警告，50-79% scaled）
  - `🪫`（危险，≥80% scaled，附加颜色）
- **`(⚠️/clear)` handoff 后缀**：
  - 1M context 模型 (Opus 4.8, GLM-5.2)：used_percentage ≥50%（基于 raw context_window_size）
  - 200K context 模型 (Sonnet/Haiku)：used_percentage ≥90%
  - 含义：建议在下一个 turn 开始前 `/clear` + 使用 paste-ready resume message
- **示例**：`🪫 CW: ███████░░░ 72% (⚠️/clear)`
- **段键**：`context`

### 5H（5 小时 rolling rate limit）

- **格式**：`🔋 5H: <bar> <pct>% [(<reset>)]`
- **数据来源**：stdin `rate_limits.five_hour.{used_percentage, resets_at}`
- **Reset 格式**：
  - <60 分钟：`(Nm)`（例：`(47m)`）
  - <24 小时：`(Nh Nm)`（例：`(2h 15m)`）
  - ≥24 小时：`(Mon DD)`（例：`(May 28)`）
- **示例**：`🔋 5H: █████░░░░░ 56% (47m)`
- **数据缺失**：`rate_limits.five_hour == null` → bar 0%，reset `(rolling)`
- **段键**：`usage_5h`

### 7D（7 天 rolling rate limit）

- **格式**：`🔋 7D: <bar> <pct>% [(<reset>)]`
- **数据来源**：stdin `rate_limits.seven_day.{used_percentage, resets_at}`
- **Reset 格式**：`(Mon DD)`（绝对日期）
- **示例**：`🔋 7D: █░░░░░░░░░ 13% (May 28)`
- **段键**：`usage_7d`

对订阅套餐用户来说，5H/7D bar 实际上就是预算仪表 — 看着这两条 bar，就能判断在 rate limit 耗尽前是安排重活，还是切到 CG 模式交给 GLM 工作者。

## Line 3 — Git / PR（5 个段）

### Directory

- **格式**：`📁 <directory name>`
- **数据来源**：stdin `workspace.project_dir`（basename）或 `cwd`
- **示例**：`📁 moai-adk-go`、`📁 my-project`
- **隐藏条件**：`data.Directory` 为空字符串
- **段键**：`directory`

### Repo + Branch（合并段）

- **格式**：`🔀 <owner>/<name> (🅱️ <branch>[ ↑N][ ↓N][ +N])`
- **数据来源**：
  - `🔀 owner/name`：stdin `workspace.repo.{host, owner, name}`（Claude Code v2.1.145+）
  - `🅱️ branch`：本地 git `branch --show-current`
  - `↑N`：ahead 计数（相对 origin/<branch>）
  - `↓N`：behind 计数
  - `+N`：dirty 计数 = Modified + Staged + Untracked
- **示例**：
  - `🔀 modu-ai/moai-adk (🅱️ main ↑3 +2)`（repo + branch + ahead + dirty）
  - `🔀 modu-ai/moai-adk (🅱️ main)`（clean 分支，无 ahead）
  - `🔀 (🅱️ feat/auth ↑2 ↓1 +6)`（无 repo 信息的 fallback）
- **隐藏条件**：
  - branch 为空字符串 → 隐藏整个段
  - repo 为 nil 时 fallback（只显示括号内 branch）
- **Worktree 模式**：`worktree` 段激活时 branch 前加 `[WT] ` 前缀
- **段键**：`git_branch`（combined）

### Git Status

- **格式**：`💾 +<staged> M<modified> ?<untracked>`
- **数据来源**：解析本地 git `git status --porcelain`
- **示例**：`💾 +0 M1 ?1`（staged 0、modified 1、untracked 1）
- **隐藏条件**：git 不可用
- **备注**：废弃了先前的 4 种 mailbox emoji（`📬`/`📫`/`📪`/`📭`），统一使用 `💾`
- **段键**：`git_status`

### Task（活动 SPEC workflow）

- **格式**：`📋 [<command> <SPEC-ID>-<stage>]`
- **数据来源**：`~/.moai/state/last-session-state.json` 的 `active_task` 字段（仅在该文件被写入时显示）
- **示例**：`📋 [/moai run SPEC-V3R5-STATUSLINE-001-implement]`
- **隐藏条件**：文件缺失或 `active_task` 为 nil → 隐藏该段
- **段键**：`task`（opt-in，默认关闭）

### PR（活动 GitHub Pull Request）

- **格式**：`💌 PR #<number> (⌥<review_state>)`（有 state 时）/ `💌 PR #<number>`（state 为空字符串）
- **数据来源**：stdin `pr.{number, url, review_state}`（Claude Code v2.1.146+）
- **Review state 值**：`approved` / `pending` / `changes_requested` / `draft` / 其他（raw passthrough）
- **颜色编码**（review_state 部分）：
  - `approved`：绿色 (Success)
  - `pending`：黄色 (Warning)
  - `changes_requested`：红色 (Error)
  - `draft`：灰色 (Muted)
  - 其他：无颜色（raw passthrough）
- **示例**：
  - `💌 PR #1234 (⌥approved)`（绿色）
  - `💌 PR #1023 (⌥pending)`（黄色）
  - `💌 PR #7 (⌥changes_requested)`（红色）
  - `💌 PR #99 (⌥draft)`（灰色）
  - `💌 PR #100`（无 state）
- **隐藏条件**：
  - 无 `pr` 字段（无 PR 或 v2.1.145 以下）
  - `pr.number == 0`
  - `SegmentPR` 配置显式为 false
- **段键**：`pr`（v2.20.0-rc1 起默认开启）

## 配置

### 基本结构

在 `.moai/config/sections/statusline.yaml` 中管理段的启用。

```yaml
statusline:
  theme: catppuccin-mocha    # 颜色主题
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

### 刷新周期

Statusline 的刷新周期由 `settings.json` 的 `statusLine.refreshInterval` 设置(单位:**秒**,默认值 `10`)。它属于 Claude Code 运行时设置,而非 `.moai/config/sections/statusline.yaml`。值太低会增加 CPU 占用,太高则上下文使用率变化反映得慢。

```json
{
  "statusLine": {
    "type": "command",
    "command": "$CLAUDE_PROJECT_DIR/.moai/status_line.sh",
    "refreshInterval": 10
  }
}
```

### 段启用矩阵

| 段 | 行 | 默认启用 | stdin field |
|---------|------|----------|-------------|
| `model` | L1 | ✓ | `model.display_name` |
| `effort_thinking` | L1 | ✓ | `effort.level` + `thinking.enabled` |
| `claude_version` | L1 | ✓ | `version` |
| `moai_version` | L1 | ✓ | （本地 config） |
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

## Handoff Guide — `(⚠️/clear)` 建议标准

CW bar 的 handoff 后缀在上下文使用量超过按模型的阈值时激活。这是提前防范 SSE stall 风险、建议使用 paste-ready resume message 的可视化标记，并**分两个阶段**动作。

- **soft 阶段** `(⚠️/clear)`：到达 band 的 soft 阈值时
- **hard 阶段** `(🛑/clear!)`：到达 auto-compact-aware ceiling(`min(cap, auto-compact-threshold + margin)`)时（`internal/statusline/renderer.go`）。由于运行时 auto-compact 常常抢先占用该 ceiling，hard 阶段实际上是很少发火的上位信号。

| 模型类别 | Context Window | 阈值 | 建议时点 |
|------------|----------------|--------|----------|
| **1M context** (Opus 4.8) | 1,000,000 tokens | **≥50%** | 使用约 500K 代币时 |
| **256K context** (Fable) | 256,000 tokens | **≥90%** | 使用约 230K 代币时 |
| **200K context** (Sonnet, Haiku) | 200,000 tokens | **≥90%** | 使用约 180K 代币时 |
| 其他 / 未知 | — | 不显示 | （安全默认） |

> 阈值由 `internal/statusline/renderer.go` 的 handoff 阶段判定强制。该阈值与 `.claude/rules/moai/workflow/context-window-management.md` HARD 规则一致。

### GLM 上下文仪表校正 (Issue #653)

GLM-5.2 是真正的 1M 上下文模型，但 Claude Code 与 provider 无关地按 Claude 槽位报告 `context_window_size`，因此在 GLM 会话中 raw telemetry(`effectiveWindow`)可能被错误显示为 ~180K。MoAI 用 `ResolveGLMContextWindow`(`internal/statusline/memory.go`)对此进行校正 —— 从 `MOAI_STATUSLINE_CONTEXT_SIZE` 环境变量（显式覆盖）或 `llm.yaml` 的 `glm.context_windows` 表(glm-5.2 → 1,000,000)解析。在 GLM 会话中，请信任 MoAI statusline 的 CW%，而非 raw `effectiveWindow`。

激活时的用户流程如下。

1. 显示 `(⚠️/clear)` 标记
2. 把进行中的工作保存到 `progress.md` 等
3. orchestrator 生成 paste-ready resume message（session-handoff.md 6-block 格式）
4. 执行 `/clear` 后粘贴 resume message
5. 在新会话中继续工作

## stdin JSON 模式参考

Claude Code 传给 statusline 脚本的 stdin JSON 完整字段列表见 [官方文档 Available data](https://code.claude.com/docs/en/statusline#available-data)。moai-adk-go 使用以下字段。

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

- **v2.20.0-rc1 layout v3**（2026-05-22）：3-line 布局重设计 — repo+branch 合并段、directory 移到 L3 首位、`🪫 CW:` emoji 前置、`(⚠️/clear)` handoff 后缀、`💾` git status 统一、`💌 PR #N (⌥state)` 格式
- **v2.20.0-rc1 STATUSLINE-STDINFIELDS-001**（2026-05-21）：新增 `workspace.repo` + `exceeds_200k_tokens` + `pr` stdin 字段映射，1M context handoff 阈值 75% → 50%
- **v2.20.0-rc1 STATUSLINE-V2145-001**（2026-05-20）：新增 PR 段（v2.1.145+ stdin），4 语言文档同步
- **v2.1.139**（Claude Code）：stdin JSON 新增 `effort.level` + `thinking.enabled`
- **v2.1.146**（Claude Code）：stdin JSON 新增 `workspace.repo` + `pr`

## 故障排查

### Statusline 不显示 PR

- 确认 Claude Code 版本：需要 `🔅 v2.1.146` 以上（v2.1.145 的 stdin 不含 `pr` 字段）
- 确认当前分支是否有 OPEN PR：`gh pr view`
- 确认 `statusline.yaml` 中是否显式设为 `pr: false`

### `(⚠️/clear)` 不显示

- 1M context 模型：used_percentage 低于 50% → 正常（尚未达到阈值）
- 200K context 模型：used_percentage 低于 90% → 正常
- 超过阈值仍不显示：检查 `shouldShowHandoffGuide` 函数的 `MemoryData.ContextWindowSize` 映射（可能存在 boundary defect）

### 颜色不显示

- 确认终端是否支持 ANSI 256-color
- 确认 `theme: catppuccin-mocha` 是否适合当前环境
- 确认是否设置了 `NO_COLOR=1` 环境变量

### 验证命令

```bash
# 用 stdin fixture 确认 statusline 实际输出
NOW=$(date +%s)
echo '{"session_id":"test","model":{"display_name":"Opus 4.7"},"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}},"version":"2.1.146","output_style":{"name":"MoAI"},"context_window":{"used_percentage":62,"context_window_size":1000000},"exceeds_200k_tokens":true,"effort":{"level":"xhigh"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":56,"resets_at":'$((NOW + 2820))'},"seven_day":{"used_percentage":13,"resets_at":'$((NOW + 518400))'}},"cost":{"total_duration_ms":17520000},"pr":{"number":1234,"url":"https://github.com/modu-ai/moai-adk/pull/1234","review_state":"approved"}}' | moai statusline
```

## `/cd` 保留缓存的目录切换 (CC 2.1.169+)

Claude Code 2.1.169+ 提供 `/cd <path>` 命令，可在 **保留提示词缓存** 的同时变更会话的工作目录 — statusline 的 `cwd` 字段会更新为新目录，但进行中的推理上下文不会重建。这是打开新终端会话的缓存保留替代方案：`/cd` 保留累积的上下文，而新终端会从零 cold-start。当 statusline 显示的 `cwd` 是你想在不丢上下文的前提下离开的位置（例如会话中切换到 L2 worktree）时，`/cd` 是摩擦最小的路径。resume-pattern 集成见 [会话交接](/zh/workflow-commands/moai-sync)。

## 相关文档

- [Settings JSON](/zh/advanced/settings-json) — Claude Code `statusLine` 字段配置
