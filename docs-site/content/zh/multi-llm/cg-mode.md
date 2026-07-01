---
title: CG 模式（Claude + GLM）
weight: 20
draft: false
---

## CG 模式是什么？

CG（Claude + GLM）模式是一种混合模式，其中**读者使用 Claude API**，**工作者使用 GLM API**。通过 tmux 会话级别的环境变量隔离来实现。

## 架构

```
运行 moai cg
    │
    ├── 1. 将 GLM 设置注入 tmux 会话环境变量
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* 变量)
    │
    ├── 2. 从 settings.local.json 中移除 GLM 环境变量
    │      → 读者 pane 使用 Claude API
    │
    ├── 3. 设置 CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
    │      → 工作者在新 pane 中继承 GLM 环境变量
    │
    └── 4. 运行 Claude Code（替换当前进程）
```

```
┌─────────────────────────────────────────────────────────────┐
│  读者（当前 tmux pane，Claude API）                          │
│  - 执行 /moai --team 时进行工作流编排                         │
│  - 处理 plan、quality、sync 阶段                              │
│  - 无 GLM 环境变量 → 使用 Claude API                         │
└──────────────────────┬──────────────────────────────────────┘
                       │ Agent Teams（新 tmux pane）
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  团队成员（新 tmux pane，GLM API）                            │
│  - 继承 tmux 会话环境变量 → 使用 GLM API                    │
│  - 在 run 阶段执行实现工作                                    │
│  - 通过 SendMessage 与读者通信                                │
└─────────────────────────────────────────────────────────────┘
```

## 使用方法

### 第 1 步：保存 GLM API 密钥（仅第一次）

```bash
moai glm sk-your-glm-api-key
```

密钥安全地保存在 `~/.moai/.env.glm` 中。

### 第 2 步：确认 tmux 环境

如果您已经在使用 tmux，无需创建新会话。

```bash
# 如果不使用 tmux：
tmux new -s moai
```

> **提示**：将 VS Code 终端默认设置为 tmux，可以完全跳过此步骤。

### 第 3 步：运行 CG 模式

```bash
moai cg
```

`moai cg` 会在当前 pane 中自动运行 Claude Code。无需单独运行 `claude` 命令。

### 第 4 步：运行团队工作流

```bash
/moai --team "实现用户认证功能"
```

## 重要事项

| 项目 | 描述 |
|------|------|
| **tmux 环境** | 如果已在使用 tmux，无需新会话。将 VS Code 终端默认设置为 tmux 很方便 |
| **自动运行** | `moai cg` 在当前 pane 中自动运行 Claude Code。无需单独 `claude` 命令 |
| **会话结束** | session_end 挂钩自动清理 tmux 会话环境变量 → 下一会话使用 Claude |
| **团队通信** | 通过 SendMessage 工具进行读者 ↔ 工作者通信 |
| **模式切换** | 从 `moai glm` 切换时 `moai cg` 自动初始化 GLM 设置 — 无需中间 `moai cc` |

## tmux 环境变量注入安全模型 {#tmux-env-security}

从 v2.20.0-rc1 开始，`moai cg` 向 tmux 会话环境变量注入 GLM 令牌（`ANTHROPIC_AUTH_TOKEN`）时，使用**源文件通道**（`tmux source-file <tmp>`）而不是**argv 通道**（`tmux set-environment <KEY> <VALUE>`）。令牌不再以明文形式在 `ps auxe`、`/proc/<pid>/cmdline`、auditd 日志、sysmon 跟踪、崩溃转储中公开（CWE-214）。

### 注入流程

1. 在 `~/.moai/run/` 下用 `mkstemp` 创建临时文件（强制 mode `0o600`）
2. 记录 `set-environment -t <session> <KEY> <VALUE>` 一行
3. `tmux source-file <tmp>` 让 tmux 读取该文件并注入环境
4. 注入后立即用 `os.Remove` 删除

argv 仅公开临时文件路径，令牌本身不公开。

### 非敏感值保持 argv

`CLAUDE_CONFIG_DIR`、`ANTHROPIC_BASE_URL`、`ANTHROPIC_DEFAULT_*_MODEL` 等非令牌值保持现有 argv 路径（无安全威胁）。

### 用户责任

`~/.moai/.env.glm` 源文件必须在用户环境中保持 `0o600` 权限。这由 `moai glm` 命令自动设置：

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### 自我检查

在 CG 模式运行中检查令牌是否在 argv 中公开：

```bash
# 运行 moai cg 后，在新 tmux 会话中
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期望值：0 matches（令牌不在 argv 中）
```

有关详细的威胁模型、失败行为（`ErrTmuxSensitiveInjectFailed` sentinel）和其他检查步骤，请参阅[安全说明 — CWE-214](/zh/advanced/security-notes/#cwe-214)。

## 显示模式

Agent Teams 支持两种显示模式：

| 模式 | 描述 | 通信 | 读者/工作者分离 |
|------|------|------|----------------|
| `in-process` | 默认模式，所有终端 | ✅ SendMessage | ❌ 相同环境 |
| `tmux` | 分割屏幕显示 | ✅ SendMessage | ✅ 会话环境变量隔离 |

> **CG 模式仅在 `tmux` 显示模式中可以实现读者/工作者 API 分离。**

## 模式比较

| 命令 | 读者 | 工作者 | 需要 tmux | 成本节省 | 用途 |
|------|------|--------|----------|--------|------|
| `moai cc` | Claude | Claude | 否 | - | 复杂任务，最高质量 |
| `moai glm` | GLM | GLM | 推荐 | ~70% | 成本优化 |
| `moai cg` | Claude | GLM | **是** | **~60%** | 质量 + 成本平衡 |

### 何时使用 CG 模式？

**适合 CG 模式：**
- 实现为中心的 SPEC 执行（run 阶段）
- 代码生成任务
- 编写测试
- 文档生成

**适合 Claude 专用（cc）：**
- 架构设计/规划（需要 Opus 推理）
- 安全审计（需要 Claude 的安全训练）
- 复杂调试（需要高级推理）

## 故障排除

| 问题 | 原因 | 解决方案 |
|------|------|--------|
| 工作者使用 Claude API | tmux 会话环境变量未设置 | 在 tmux 中重新运行 `moai cg` |
| `moai cg` 后 Claude Code 未运行 | 在 tmux 外运行 | 运行 `tmux new -s moai` 后重试 |
| 会话结束后 GLM 环境变量残留 | session_end 挂钩失败 | 用 `moai cc` 手动清理 |

## 后续步骤

- [模型策略](/zh/multi-llm/model-policy) — 按代理分配模型
- [双执行模式](/zh/getting-started/faq) — Sub-Agent vs Agent Teams
- [CLI 参考](/zh/getting-started/cli) — moai cc、moai glm、moai cg 详细说明
