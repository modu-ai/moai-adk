---
title: 状态栏系统与PR段
weight: 78
draft: false
---

这是用于集成 Claude Code 和 moai-adk-go 的**自定义状态栏系统**。从 v2.1.145 开始，您可以在状态栏中显示 GitHub PR 信息。

> MoAI 工作流是以 PR 为中心的。所有 SPEC 都会生成 plan-PR → run-PR → sync-PR，因此在状态栏中显示当前 PR 状态可以提高开发效率。

## 概述

### 为什么需要自定义状态栏

Claude Code 的默认状态栏针对通用使用模式进行了优化。但是 MoAI-ADK 用户需要以下特殊信息：

- **PR 中心工作流**：当前 PR 编号和审核状态（approved/pending/changes_requested）
- **多窗格开发**：使用工作树进行并行开发时显示当前 SPEC 状态
- **成本跟踪**：使用 GLM 环境时的实时成本监控
- **上下文管理**：当前会话的令牌使用率和累计成本

自定义状态栏通过 `.moai/status_line.sh` 渲染器显示这些信息。

### 状态栏架构

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData 解析)
    ↓
internal/statusline/builder.go (段构建)
    ↓
internal/statusline/renderer.go (颜色编码和渲染)
    ↓
.moai/status_line.sh (基于模板的最终渲染)
```

## 配置

### 基本结构

在 `.moai/config/sections/statusline.yaml` 中配置状态栏：

```yaml
statusline:
  mode: default              # default | compact | verbose
  theme: catppuccin-mocha    # 选择颜色主题
  preset: full               # full | minimal | custom
  segments:
    model: true              # 显示 Claude 模型
    context: true            # 显示上下文使用率
    directory: true          # 显示工作目录
    git_status: true         # 显示 Git 状态
    git_branch: true         # 显示 Git 分支
    worktree: false          # 显示工作树信息（可选）
    effort_thinking: false   # 显示 Effort/thinking 状态（可选）
    pr: false                # 显示 PR 信息（可选，v2.1.145+）
```

### 段选项

| 段 | 默认值 | 用途 | 说明 |
|------|--------|------|------|
| `model` | true | 当前模型 | 显示 Claude 模型版本 |
| `context` | true | 上下文使用率 | 显示当前会话的令牌使用率 |
| `directory` | true | 工作路径 | 显示当前工作目录 |
| `git_status` | true | Git 状态 | 修改的文件数、贮藏状态 |
| `git_branch` | true | 当前分支 | 分支名称和远程差异 |
| `worktree` | false | 工作树信息 | 显示当前工作树（并行开发） |
| `effort_thinking` | false | 思考模式 | effort 和 thinking 状态 |
| `pr` | false | PR 信息 | GitHub PR 编号和审核状态（新 v2.1.145+） |

## 可用的段

### 始终启用的段（4个）

**model** — Claude 模型
- 显示当前模型（Claude 3.5 Sonnet、Claude 3.7 Opus 等）
- 示例：`Claude 3.5 Sonnet`

**context** — 上下文使用率
- 显示当前会话的令牌使用率
- 格式：`150K/200K`（使用中 / 总计）
- 75% 以上时以警告颜色显示

**directory** — 工作目录
- 当前工作目录的相对路径
- 显示相对于项目根目录的位置

**git_status** — Git 状态
- 修改的文件数：`M5`（5个文件修改）
- 贮藏状态：`S2`（2个贮藏）
- 示例：`M5 S2`

**git_branch** — 当前分支
- 分支名称
- 与远程的提交差异显示
- 示例：`feat/SPEC-001 +3 -1`

### 可选段（7个）

**worktree** — 工作树信息（可选）
- 使用 L2 工作树时显示
- 显示当前 SPEC 名称
- 启用：`segments.worktree: true`

**effort_thinking** — Effort/thinking 状态（可选）
- Claude 4.7 思考模式启用状态
- effort 级别（high/xhigh/max）
- 启用：`segments.effort_thinking: true`

**output_style** — 输出样式（可选）
- 当前输出样式设置
- 启用：`segments.output_style: true`

**claude_version** — Claude 版本（可选）
- Claude Code 版本
- 启用：`segments.claude_version: true`

**moai_version** — moai 版本（可选）
- MoAI-ADK 版本
- 启用：`segments.moai_version: true`

**session_time** — 会话经过时间（可选）
- 当前会话开始以来的经过时间
- 启用：`segments.session_time: true`

**usage_5h** — 5小时累计成本（可选）
- 过去 5 小时的成本跟踪
- 在 GLM 环境中有用
- 启用：`segments.usage_5h: true`

**usage_7d** — 7天累计成本（可选）
- 过去 7 天的成本跟踪
- 启用：`segments.usage_7d: true`

**task** — 活动 SPEC 工作流信息（可选）
- 输出格式：`📋 [<command> <SPEC-ID>-<stage>]`（例如：`📋 [/moai run SPEC-V3R5-DOCS-SECURITY-001-M3]`）
- 数据源：`~/.moai/state/last-session-state.json` 的 `active_task` 字段（由 SessionStart 钩子自动设置）
- 非活动状态下 segment 不显示（graceful no-output）
- 启用：`segments.task: true`（自 v2.20.0-rc1 起默认 true — default-on，可通过 `false` 显式禁用）

## 新增 v2.1.145：PR 段

### 概述

从 Claude Code v2.1.145 开始，状态栏 stdin JSON 包含 GitHub PR 信息。MoAI-ADK 利用这些信息在状态栏中显示当前 PR 的审核状态。

**启用**：自 v2.20.0-rc1 起默认 `true`（default-on）。设置 `segments.pr: false` 可显式禁用。graceful no-output 模式：当无 PR 信息时 segment 自动隐藏。

### PR 段显示格式

PR 段以以下格式显示：

```
#1023 ⌥approved
```

- `#1023`：PR 编号
- `⌥`：PR 状态显示符号
- `approved`：审核状态（颜色编码）

### 按审核状态的颜色

根据 PR 的审核状态以不同颜色显示：

| 状态 | 颜色 | 含义 |
|------|------|------|
| `approved` | 绿色 | PR 已批准 |
| `pending` | 黄色 | 等待审核中 |
| `changes_requested` | 红色 | 已请求更改 |
| `draft` | 灰色 | 草稿状态 |
| （其他/空） | 默认 | 未应用样式 |

### 启用方法

1. 编辑 `.moai/config/sections/statusline.yaml` 文件

```yaml
statusline:
  segments:
    pr: true   # 启用 PR 段
```

2. 重启 Claude Code 会话

现在状态栏将显示当前 PR 的编号和审核状态。

### JSON 输入架构（v2.1.145+）

Claude Code v2.1.145+ 将以下格式的 JSON 传递给状态栏 stdin：

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

- **pr.number**：PR 编号（必需）
- **pr.url**：PR 网址（可选）
- **pr.review_state**：审核状态（可选，默认：空）
- **workspace.repo.host**：Git 主机（github.com）
- **workspace.repo.owner**：存储库所有者
- **workspace.repo.name**：存储库名称

### 相关信息

- **SPEC 参考**：[SPEC-V3R5-STATUSLINE-V2145-001](/zh/advanced/statusline#参考)
- **最低版本**：需要 Claude Code v2.1.145 或更高版本
- **可选功能**：默认为 `false`，需要明确启用
- **向后兼容性**：早期版本的 Claude Code 不提供 PR 信息（段不显示）

## 故障排除：状态栏消失问题

### 症状

- 状态栏间歇性不显示
- Claude Code UI 中的状态栏区域为空
- `.moai/cache/statusline_debug.log` 文件持续增长

### 原因分析（v2.1.145 M1 修复前）

状态栏渲染器必须遵守 Claude Code 的**300ms 防抖契约**。违反此合同将取消正在进行的执行。

以前代码的问题：

```bash
# 问题：DEBUG_STATUSLINE 的默认值为 1（始终启用）
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-1}

# 这导致每次渲染时：
# 1. python3 -m json.tool 进程分叉（50-250ms）
# 2. 写入 ~/.moai/cache/statusline_debug.log（~10ms）
# 合计：60-260ms → 超过 300ms 防抖边界
# → Claude Code 取消正在进行的状态栏渲染
# → 结果：状态栏不显示
```

### 解决方案（v2.1.145 M1 中已修复）

从 v3.5.0 开始，`DEBUG_STATUSLINE` 的默认值为 **0**：

```bash
# 已修复：默认值 0（禁用）
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0}

# 仅在需要调试时明确启用：
export DEBUG_STATUSLINE=1
```

### 填充调整

以前使用 `echo ""` 来调整状态栏周围的空白。这现在不再推荐。

**改为**在 `.claude/settings.json` 中设置：

```json
{
  "statusLine": {
    "padding": 1
  }
}
```

- `padding: 0`：无填充
- `padding: 1`：上下 1 行填充（默认）
- `padding: 2`：上下 2 行填充

### 检查清单

解决状态栏显示问题的步骤：

1. ✓ 检查 `DEBUG_STATUSLINE` 环境变量
   ```bash
   echo $DEBUG_STATUSLINE  # 默认应为 unset 或 0
   ```

2. ✓ 检查 `.moai/status_line.sh` 文件
   ```bash
   grep "DEBUG_STATUSLINE=" ~/.moai/status_line.sh
   # 结果应为：DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0}
   ```

3. ✓ 明确启用调试（仅在需要时）
   ```bash
   export DEBUG_STATUSLINE=1
   # 现在将记录调试信息
   ```

4. ✓ 设置填充
   ```json
   {
     "statusLine": {
       "padding": 1
     }
   }
   ```

5. ✓ 重启 Claude Code 会话

## 参考

### 官方文档

- [Claude Code 状态栏官方文档](https://code.claude.com/docs/en/statusline) — Claude Code 的状态栏契约和 JSON 架构

### moai-adk-go 内部

- **包**：`internal/statusline/`
  - `types.go`：StdinData、PRInfo、RepoInfo 结构体定义
  - `builder.go`：段构建逻辑
  - `renderer.go`：颜色编码和最终渲染

- **模板**：`.moai/status_line.sh.tmpl`
  - 渲染器调用和执行逻辑

- **配置**：`.moai/config/sections/statusline.yaml`
  - 段启用/禁用设置

### 相关 SPEC

- **[SPEC-V3R5-STATUSLINE-V2145-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-V3R5-STATUSLINE-V2145-001/spec.md)**
  - M1：修复状态栏消失问题
  - M2：添加 v2.1.145 PR 段
  - M3：文档化（当前页面）
