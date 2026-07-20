---
title: Hooks 事件参考
weight: 60
draft: false
---

Claude Code 的 Hook 系统支持 **29 种事件类型**、**5 种 Hook 类型**、**按事件的匹配器** 以及 **智能行为**。Hook 是智能体 Harness 中唯一保证"一定会执行"的确定性 (deterministic) 控制点 — 提示词可能被忽略，但 Hook 不会。

> Hook 的基本概念与配置方法请参阅 [Hooks 指南](/zh/advanced/hooks-guide)。本页是完整的事件参考。

## Hook 类型

可用的 Hook 类型有五种。

| 类型 | 说明 | 示例 |
|------|------|------|
| **command** | 执行 Shell 脚本 | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM 评估 | 由 LLM 执行提示词文本并返回结果 |
| **agent** | 子智能体校验 | 由智能体校验工作并返回结果 |
| **http** | Webhook 端点 | 通过 HTTP POST 请求传递事件 |
| **mcp_tool** | 执行 MCP 工具 | 远程调用 MCP 服务器工具 |

## 完整事件参考（29 个）

### 生命周期事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `SessionStart` | 会话开始 | — |
| `SessionEnd` | 会话结束 | — |
| `PostSession` | 会话结束后执行（self-hosted runner 生命周期事件，CC 2.1.169+）。在会话完全释放后、晚于 `SessionEnd` 触发。MoAI-ADK 目前未接线此 Hook。此处将其记录为需要会话后清理/遥测的 self-hosted 部署的可用选项。 | — |
| `Stop` | 智能体停止 | — |
| `SubagentStop` | 子智能体停止 | — |
| `SubagentStart` | 子智能体启动 | — |
| `StopFailure` | 停止失败 | `errorType` |
| `Setup` | 初始设置 | — |

### 工具事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `PreToolUse` | 工具执行前 | `toolName` |
| `PostToolUse` | 工具执行后 | `toolName` |
| `PostToolUseFailure` | 工具执行失败 | `toolName`, `errorType` |
| `PostToolBatch` | 并行工具批次执行后 (v2.1.89+) | — |

### 上下文事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `PreCompact` | 上下文压缩前 | — |
| `PostCompact` | 上下文压缩后 | — |
| `InstructionsLoaded` | 指令加载完成 | — |

### 输入事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `UserPromptSubmit` | 用户提交提示词 | — |
| `UserPromptExpansion` | 斜杠命令提示词展开 (v2.1.90+) | — |
| `Elicitation` | Elicitation 开始 | — |
| `ElicitationResult` | Elicitation 完成 | — |

### 安全事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `PermissionRequest` | 权限请求 | `toolName` |
| `PermissionDenied` | 权限被拒 | `toolName` |

### 团队事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `TeammateIdle` | 团队成员转入空闲 | — |
| `TaskCompleted` | 任务标记完成 | — |
| `TaskCreated` | 任务创建 | — |

### Worktree 事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `WorktreeCreate` | 创建 worktree | — |
| `WorktreeRemove` | 删除 worktree | — |

### 环境事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `ConfigChange` | 配置变更 | `configSource` |
| `CwdChanged` | 工作目录变更 | — |
| `FileChanged` | 文件变更 | — |

### UI 事件

| 事件 | 说明 | 匹配器 |
|--------|------|------|
| `Notification` | 用户通知 | — |

## 智能行为 (Smart Behaviors)

MoAI-ADK 的 Hook 超越简单的事件处理，执行智能化的行为。

### PermissionDenied 自动重试

当只读工具（Read、Grep、Glob）的权限被拒时，Hook 会自动触发重试。这缓解了后台智能体中权限提示不显示的问题。

### StopFailure 错误类型响应

智能体停止失败时，按错误类型提供差异化响应。保障长时间运行会话的稳定性。

### PostCompact 会话备忘恢复

上下文压缩后自动恢复重要的会话备忘（进度状态、SPEC 引用）。上下文压缩是一笔用信息换代币的交易，而这个 Hook 在损失中守住了核心信息。

### SubagentStart 上下文注入

子智能体启动时自动注入所需上下文（项目规则、MX 标签、进度状态）。

## 匹配器 (Matchers)

使用匹配器可以过滤 Hook，使其只在特定条件下执行。给所有事件都挂 Hook 会相应增加执行成本，因此用匹配器收窄范围是基本做法。

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": { "toolName": "Bash" },
      "hooks": [{
        "type": "command",
        "command": "echo 'Bash tool detected'",
        "timeout": 5
      }]
    }]
  }
}
```

### 可用的匹配器字段

| 匹配器字段 | 适用事件 | 说明 |
|----------|-----------|------|
| `toolName` | PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, PermissionDenied | 按工具名过滤 |
| `errorType` | StopFailure, PostToolUseFailure | 按错误类型过滤 |
| `configSource` | ConfigChange | 按配置来源过滤 |

## CLAUDE_ENV_FILE

通过 `CwdChanged` 与 `FileChanged` Hook 可以持续管理环境变量。

```bash
# .claude/hooks/moai/handle-cwd-changed.sh
# 通过 CLAUDE_ENV_FILE 持久化环境变量
echo "MOAI_PROJECT_DIR=$(pwd)" >> "$CLAUDE_ENV_FILE"
```

由此可以在会话之间保持环境变量，并在目录变更时自动重置环境。

## MoAI-ADK 使用的主要 Hook

| 事件 | MoAI 处理器 | 角色 |
|--------|-----------|------|
| `SessionStart` | `handle-session-start.sh` | 初始化 Statusline、开始指标会话 |
| `PostToolUse` | `handle-post-tool.sh` | 记录 Task 指标 |
| `TeammateIdle` | `handle-teammate-idle.sh` | LSP 质量门禁校验 |
| `TaskCompleted` | `handle-task-completed.sh` | 确认 SPEC 文档存在 |
| `WorktreeCreate` | （无 — MoAI 默认不注册） | 使用 Claude Code 默认 worktree 行为（供 `isolation: worktree` 智能体使用）。若注册，则须履行 active creator 契约（创建目录 + 向 stdout 回显路径）。 |
| `WorktreeRemove` | （无 — MoAI 默认不注册） | 使用 Claude Code 默认 worktree 清理行为。若注册，则为 observer-only 契约（无需输出）。 |
| `UserPromptSubmit` | `handle-user-prompt.sh` | 自动执行质量门禁 |

## 下一步

- [Hooks 指南](/zh/advanced/hooks-guide) — Hook 基本概念与配置方法
- [settings.json 指南](/zh/advanced/settings-json) — settings.json 完整参考
- [CLI 参考](/zh/getting-started/cli) — `moai hook` 命令详解
