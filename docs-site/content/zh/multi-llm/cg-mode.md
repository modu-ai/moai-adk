---
title: CG 模式（Claude + GLM）
weight: 20
draft: false
---

## 什么是 CG 模式？

CG (Claude + GLM) 模式是领队使用 **Claude API**、工作者使用 **GLM API** 的
混合模式。它通过 tmux 会话级环境变量隔离来实现，把"计划由 Claude 深入做、
实现由 GLM 便宜做"这一托克诺米克斯分配在同一个会话内落地。以实现为主的
工作可节省约 60-70% 成本。

## 架构

```
运行 moai cg
    │
    ├── 1. 将 GLM 配置注入 tmux 会话环境变量
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* 变量)
    │
    ├── 2. 从 settings.local.json 移除 GLM 环境变量
    │      → 领队 pane 使用 Claude API
    │
    ├── 3. 设置 CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
    │      → 工作者在新 pane 中继承 GLM 环境变量
    │
    └── 4. 运行 Claude Code（替换当前进程）
```

```
┌─────────────────────────────────────────────────────────────┐
│  领队（当前 tmux pane, Claude API）                          │
│  - 工作流编排                                                 │
│  - 处理 plan, quality, sync 阶段                             │
│  - 无 GLM 环境变量 → 使用 Claude API                         │
└──────────────────────┬──────────────────────────────────────┘
                       │ spawn 队友（新 tmux pane）
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  队友（新 tmux pane, GLM API）                               │
│  - 继承 tmux 会话环境变量 → 使用 GLM API                      │
│  - 在 run 阶段执行实现工作                                    │
│  - 通过 SendMessage 与领队通信                                │
└─────────────────────────────────────────────────────────────┘
```

## 使用方法

### 第 1 步：保存 GLM API 密钥（仅首次）

```bash
moai glm sk-your-glm-api-key
```

密钥被安全地保存在 `~/.moai/.env.glm` 中。

### 第 2 步：确认 tmux 环境

如果已经在使用 tmux，就不需要创建新会话。

```bash
# 如果尚未使用 tmux:
tmux new -s moai
```

> **提示**：将 VS Code 终端默认值设置为 tmux，即可完全跳过这一步。

### 第 3 步：运行 CG 模式

```bash
moai cg
```

`moai cg` 会在当前 pane 中自动运行 Claude Code。无需单独运行 `claude`。

### 第 4 步：运行工作流

```bash
/moai "实现用户认证功能"
```

之后与平常一样。编排器（领队, Claude）负责计划、质量、同步，实现工作量大
的任务被委派给新 tmux pane 中的 GLM 队友。

> **注意**：过去的 `--team` 标志（Agent Teams 静态编排层）在 v3.0 中已
> 退役。即使强制指定也会回退到 sub-agent 模式。CG 模式的领队/工作者分离
> 由 Claude Code 内置的 teammate 运行时（tmux pane）驱动，该运行时保持
> 不变。

## 重要事项

| 项目 | 说明 |
|------|------|
| **tmux 环境** | 已在使用 tmux 时不需要新会话。将 VS Code 终端默认值设置为 tmux 会很方便 |
| **自动运行** | `moai cg` 在当前 pane 自动运行 Claude Code。无需单独的 `claude` 命令 |
| **会话结束** | session_end 钩子自动清理 tmux 会话环境变量 → 下一个会话使用 Claude |
| **团队通信** | 通过 SendMessage 工具进行领队↔工作者之间的通信 |
| **模式切换** | 从 `moai glm` 切换时 `moai cg` 会自动重置 GLM 配置 —— 中间无需 `moai cc` |

## tmux 环境变量注入安全模型 {#tmux-env-security}

自 v2.20.0-rc1 起，`moai cg` 将 GLM token (`ANTHROPIC_AUTH_TOKEN`) 注入 tmux 会话环境变量时，使用 **source-file 通道** (`tmux source-file <tmp>`) 而非 **argv 通道** (`tmux set-environment <KEY> <VALUE>`)。token 不再以明文暴露在 `ps auxe`、`/proc/<pid>/cmdline`、auditd 日志、sysmon 跟踪、崩溃转储中 (CWE-214)。

### 注入流程

1. 在 `~/.moai/run/` 下用 `mkstemp` 创建临时文件（强制 mode `0o600`）
2. 写入一行 `set-environment -t <session> <KEY> <VALUE>`
3. 通过 `tmux source-file <tmp>` 让 tmux 读取该文件并注入环境
4. 注入后立即用 `os.Remove` unlink

argv 中只暴露临时文件路径，token 本身不会暴露。

### 非敏感值维持 argv

`CLAUDE_CONFIG_DIR`、`ANTHROPIC_BASE_URL`、`ANTHROPIC_DEFAULT_*_MODEL` 等非 token 值维持原有 argv 路径（无安全威胁）。

### 用户责任

`~/.moai/.env.glm` source 文件必须在用户环境中保持 `0o600` 权限。这由 `moai glm` 命令自动设置：

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### 自检

确认 CG 模式运行期间 token 是否暴露在 argv 中：

```bash
# 运行 moai cg 后在新 tmux 会话内
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期望值: 0 matches（token 不在 argv 中）
```

详细威胁模型、失败时行为（`ErrTmuxSensitiveInjectFailed` sentinel）、附加检查步骤请参考[安全说明 — CWE-214](/zh/advanced/security-notes/#cwe-214)。

## 显示模式

teammate 运行时支持两种显示模式：

| 模式 | 说明 | 通信 | 领队/工作者分离 |
|------|------|------|--------------|
| `in-process` | 默认模式，所有终端 | 支持 SendMessage | 无分离（相同环境） |
| `tmux` | 分屏显示 | 支持 SendMessage | 会话环境变量隔离 |

> **CG 模式仅在 `tmux` 显示模式下才能实现领队/工作者 API 分离。**

## 模式对比

| 命令 | 领队 | 工作者 | 需要 tmux | 成本节省 | 用途 |
|--------|------|------|----------|----------|------|
| `moai cc` | Claude | Claude | 否 | - | 复杂任务、最高质量 |
| `moai glm` | GLM | GLM | 推荐 | ~70% | 成本优化 |
| `moai cg` | Claude | GLM | **必需** | **~60%** | 质量 + 成本平衡 |

### 什么时候该用 CG 模式？

**适合 CG 模式：**
- 以实现为主的 SPEC 执行（run 阶段）
- 代码生成任务
- 编写测试
- 生成文档

**适合 Claude 专用 (cc)：**
- 架构设计/计划（需要 Opus 推理）
- 安全审查（需要 Claude 的安全训练）
- 复杂调试（需要高级推理）

## 问题排查

| 问题 | 原因 | 解决 |
|------|------|------|
| 工作者使用 Claude API | 未设置 tmux 会话环境变量 | 在 tmux 内重新运行 `moai cg` |
| `moai cg` 后 Claude Code 未运行 | 在 tmux 之外运行 | `tmux new -s moai` 后重新运行 |
| 会话结束后 GLM 环境变量残留 | session_end 钩子失败 | 用 `moai cc` 手动清理 |

## 下一步

- [模型策略](/zh/multi-llm/model-policy) —— 各代理的模型分配
- [常见问题](/zh/getting-started/faq) —— 运行模式相关 FAQ
- [CLI 参考](/zh/getting-started/cli) —— moai cc, moai glm, moai cg 详解
