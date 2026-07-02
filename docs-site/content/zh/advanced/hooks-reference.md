---
title: Hooks事件参考
weight: 60
draft: false
---

截至MoAI-ADK v2.10.1，Claude Code的钩子系统支持 **29个事件类型**、**5种钩子类型**、**按事件匹配器** 和 **智能行为**。

> 有关钩子的基本概念和设置说明，请参阅 [Hooks指南](/zh/advanced/hooks-guide)。本页面是完整的事件参考。

## 钩子类型

**提供5种钩子类型：**

| 类型 | 说明 | 示例 |
|-----|------|------|
| **command** | Shell脚本执行 | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM评估 | LLM执行提示文本并返回结果 |
| **agent** | 子代理验证 | 代理验证任务并返回结果 |
| **http** | Webhook端点 | HTTP POST请求到远程端点 |
| **mcp_tool** | MCP工具调用 | 远程调用MCP服务器工具 |

## 完整事件参考 (29个)

### 生命周期事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `SessionStart` | 会话开始 | — |
| `SessionEnd` | 会话结束 | — |
| `PostSession` | 会话结束后运行 (self-hosted runner 生命周期事件，CC 2.1.169+)。在会话完全拆除后触发，晚于 `SessionEnd`。MoAI-ADK 目前不接入此钩子；作为面向需要会话后清理/遥测的 self-hosted 部署的可用选项进行文档化。 | — |
| `Stop` | 代理停止 | — |
| `SubagentStop` | 子代理停止 | — |
| `SubagentStart` | 子代理启动 | — |
| `StopFailure` | 停止失败 | `errorType` |
| `Setup` | 初始设置 | — |

### 工具事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `PreToolUse` | 工具执行前 | `toolName` |
| `PostToolUse` | 工具执行后 | `toolName` |
| `PostToolUseFailure` | 工具执行失败 | `toolName`, `errorType` |
| `PostToolBatch` | 并行工具调用批处理后 (v2.1.89+) | — |

### 上下文事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `PreCompact` | 上下文压缩前 | — |
| `PostCompact` | 上下文压缩后 | — |
| `InstructionsLoaded` | 指令加载完成 | — |

### 输入事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `UserPromptSubmit` | 用户提示提交 | — |
| `UserPromptExpansion` | 斜线命令扩展为提示 (v2.1.90+) | — |
| `Elicitation` | Elicitation开始 | — |
| `ElicitationResult` | Elicitation完成 | — |

### 安全事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `PermissionRequest` | 权限请求 | `toolName` |
| `PermissionDenied` | 权限拒绝 | `toolName` |

### 团队事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `TeammateIdle` | 队友转为空闲状态 | — |
| `TaskCompleted` | 标记任务完成 | — |
| `TaskCreated` | 任务创建 | — |

### 工作树事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `WorktreeCreate` | 创建工作树 | — |
| `WorktreeRemove` | 删除工作树 | — |

### 环境事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `ConfigChange` | 配置变更 | `configSource` |
| `CwdChanged` | 工作目录变更 | — |
| `FileChanged` | 文件变更 | — |

### UI事件

| 事件 | 说明 | 匹配器 |
|-----|------|--------|
| `Notification` | 用户通知 | — |

## 智能行为 (Smart Behaviors)

MoAI-ADK的钩子不仅处理简单事件，还会执行智能化行为：

### PermissionDenied自动重试

当只读工具（Read、Grep、Glob）的权限被拒绝时，钩子会自动触发重试。这缓解了后台代理中权限提示不显示的问题。

### StopFailure错误类型响应

代理停止失败时，会根据错误类型提供差异化响应，确保长时间运行会话的稳定性。

### PostCompact会话备忘录恢复

上下文压缩后，会自动恢复重要的会话备忘录（进度状态、SPEC引用）。这可防止上下文压缩时核心信息丢失。

### SubagentStart上下文注入

子代理启动时会自动注入所需上下文（项目规则、MX标签、进度状态）。

## 匹配器 (Matchers)

使用匹配器可以过滤，使钩子仅在特定条件下执行：

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
| `toolName` | PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, PermissionDenied | 按工具名称过滤 |
| `errorType` | StopFailure, PostToolUseFailure | 按错误类型过滤 |
| `configSource` | ConfigChange | 按配置来源过滤 |

## CLAUDE_ENV_FILE

通过 `CwdChanged` 和 `FileChanged` 钩子，可以持续管理环境变量：

```bash
# .claude/hooks/moai/handle-cwd-changed.sh
# 通过 CLAUDE_ENV_FILE 持久化环境变量
echo "MOAI_PROJECT_DIR=$(pwd)" >> "$CLAUDE_ENV_FILE"
```

借此可以在会话间保留环境变量，并在目录变更时自动重新设置环境。

## MoAI-ADK使用的主要钩子

| 事件 | MoAI处理器 | 作用 |
|--------|-----------|------|
| `SessionStart` | `handle-session-start.sh` | 初始化状态栏，启动指标会话 |
| `PostToolUse` | `handle-post-tool.sh` | 记录任务指标 |
| `TeammateIdle` | `handle-teammate-idle.sh` | 验证LSP质量门禁 |
| `TaskCompleted` | `handle-task-completed.sh` | 确认SPEC文档存在 |
| `WorktreeCreate` | (无 — MoAI 默认未注册) | 使用 Claude Code 默认 worktree 行为 (供 `isolation: worktree` 代理). 注册时必须实现 active creator 契约 (创建目录 + 将绝对路径 echo 到 stdout). |
| `WorktreeRemove` | (无 — MoAI 默认未注册) | 使用 Claude Code 默认 worktree 清理. 注册时仅为 observer-only 契约 (无需 stdout). |
| `UserPromptSubmit` | `handle-user-prompt.sh` | 自动执行质量门禁 |

## 下一步

- [Hooks指南](/zh/advanced/hooks-guide) — 钩子基本概念与设置方法
- [settings.json指南](/zh/advanced/settings-json) — settings.json完整参考
- [CLI参考](/zh/getting-started/cli) — `moai hook` 命令详解
