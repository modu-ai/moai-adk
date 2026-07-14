---
title: CLI 参考
weight: 90
draft: false
---

在此查阅在终端执行的 `moai`(Go 二进制)的所有命令与标志。它与在 Claude Code 对话窗输入的 `/moai`(斜杠子命令)是完全不同的工具 —— 本页只讲终端 CLI。

## 命令树

```bash
moai --help
```

`moai` CLI 分为三组。

| 组 | 命令 | 说明 |
|------|--------|------|
| **Launch** | `moai cc` · `moai cg` · `moai glm` | 启动 Claude Code 会话(选择后端) |
| **Project** | `moai init` · `moai update` · `moai doctor` · `moai status` | 项目初始化、更新、诊断、状态查询 |
| **Tools** | `moai profile` · `moai inventory` · `moai hook` · `moai worktree` · `moai spec` · `moai harness` · ... | 配置、清单、钩子、工作树等工具 |

用 `moai version` 确认当前安装的版本。

```bash
moai version
# 输出示例: moai <版本> (commit: <哈希>, built: <构建日期>)
```

---

## moai init

初始化项目。交互式向导会设置语言、Git 自动化、模型策略、harness 配置文件等。

```bash
moai init [project-name] [OPTIONS]
```

### 标志

| 标志 | 说明 |
|--------|------|
| `--non-interactive` | 跳过交互式向导(使用标志与默认值) |
| `--force` | 强制重新初始化既有项目(会备份当前 `.moai/`) |
| `--no-hooks` | 跳过 Git 钩子安装 |
| `--all` | 部署 catalog 全部条目(core + optional packs + harness-generated) |
| `--standard` | 显示 Phase 1 提问(project mode、harness profile、LSP、quality gates、design) |
| `--advanced` | 显示 Phase 1 + Phase 2 提问(含 `--standard`;Phase 2 仅在满足前置条件时) |
| `--project-mode <personal\|team>` | 项目模式(默认: personal) |
| `--harness-profile <profile>` | harness 评估器配置文件(default, strict, lenient, frontend) |
| `--enable-lsp` | 启用 LSP 联动(默认: false) |
| `--enforce-quality` | 强制质量门禁(默认: true) |
| `--enable-design` | 启用 design 工作流(默认: true) |
| `--model-policy <max\|medium\|low>` | 性能层级 —— 保存到 `llm.yaml` `performance_tier` |
| `--plan-type <api\|subscription>` | 计费套餐类型 —— 保存到 `llm.yaml` `plan_type` |
| `--high` | **将被删除** `--model-policy max` 的别名 |

### 示例

```bash
# 初始化新项目(交互式向导)
moai init my-project

# 安装到既有文件夹
cd my-existing-project
moai init

# 非交互(CI/CD)
moai init --non-interactive --project-mode personal --model-policy medium

# 显示到 Phase 1 提问
moai init my-project --standard
```

详细的向导步骤请参阅[初始设置](./init-wizard)页面。

---

## moai update

将 MoAI-ADK 更新到最新版本。不带标志运行时会一并更新二进制与模板,用户自定义资产会自动保留。

```bash
moai update [OPTIONS]
```

### 标志

| 标志 | 说明 |
|--------|------|
| `--check` | 仅确认是否有新版本(不更新) |
| `-c, --config` | 重新运行设置向导(不同步模板) |
| `--force` | 强制更新(跳过版本一致检查、强制 备份+合并、覆盖归档 drift) |
| `--yes` | 自动批准所有确认(CI/CD 模式) |
| `--templates-only` | 跳过二进制更新,仅同步模板 |
| `--binary` | 跳过模板同步,仅更新二进制 |
| `--dry-run` | 不改动文件系统,仅显示计划的操作 |
| `--no-hooks` | 跳过 Git 钩子安装 |
| `--verbose` | 显示所有警告(诊断模式) |
| `--shell-env` | 为 Claude Code 配置 shell 环境变量 |
| `--plan-type <api\|subscription>` | 覆盖计费套餐类型(重新应用 `llm.yaml` `plan_type` 及层级配置) |

### 示例

```bash
# 默认更新(二进制 + 模板)
moai update

# 仅确认是否有新版本
moai update --check

# 重新运行设置向导
moai update -c

# 仅同步模板
moai update --templates-only
```

详细的更新流程请参阅[更新](./update)页面。

---

## moai doctor

执行系统诊断。检查 Git、项目结构、配置文件、各语言的开发工具。

```bash
moai doctor [OPTIONS]
```

### 标志

| 标志 | 说明 |
|--------|------|
| `-v, --verbose` | 显示详细的工具版本与语言检测结果 |
| `--fix` | 提出缺失工具的修复建议 |
| `--export <path>` | 将诊断结果导出为 JSON 文件 |
| `--check <tool>` | 仅确认特定工具(例: git, go, config) |

### 子命令

| 命令 | 说明 |
|--------|------|
| `moai doctor sandbox` | 沙箱环境诊断 |
| `moai doctor permission` | 权限设置诊断 |
| `moai doctor hook` | 钩子加载问题诊断 |
| `moai doctor config dump` | 将当前配置转储为 JSON |
| `moai doctor config diff` | 对比本地配置与模板默认值 |

### 示例

```bash
# 完整诊断
moai doctor

# 详细诊断
moai doctor --verbose

# 导出诊断结果
moai doctor --export diagnostics.json
```

---

## moai status

一目了然地查询项目状态。显示是否已初始化、SPEC 数量、配置文件数。

```bash
moai status
```

这是无标志的只读命令。详细的输出内容请参阅[项目状态](./status)页面。

---

## moai inventory

综合查询活跃会话、工作树、harness 的只读命令。

```bash
moai inventory [OPTIONS]
```

### 标志

| 标志 | 说明 |
|--------|------|
| `--json` | 结构化 JSON 输出 |
| `--project-root <path>` | 项目根路径(默认: 当前目录) |

详细的 JSON schema 与使用示例请参阅[moai inventory](./inventory)页面。

---

## moai profile

管理 Claude Code 设置配置文件。可为每个配置文件维护独立的模型、语言、显示设置。

```bash
moai profile [COMMAND]
```

### 子命令

| 命令 | 说明 |
|--------|------|
| `moai profile list` | 显示所有可用配置文件 |
| `moai profile setup` | 运行交互式设置向导 |
| `moai profile current` | 显示当前活跃的配置文件 |
| `moai profile delete <name>` | 删除指定的配置文件 |

用 `-p` 标志指定运行时的配置文件:

```bash
moai cc -p work       # 用 work 配置文件运行 Claude
moai glm -p cost-save # 用 cost-save 配置文件运行 GLM
moai cg -p team       # 用 team 配置文件运行 CG 模式
```

详情请参阅[配置文件管理](./profile)页面。

---

## moai hook

处理 Claude Code 钩子事件的调度器。由 `settings.json` 的钩子设置以 `moai hook <event>` 形式调用。

```bash
moai hook <event>
```

### 支持的事件(26 个)

所有事件名均为 kebab-case。

| 事件 | 说明 |
|-------|------|
| `session-start` | 会话开始 |
| `session-end` | 会话结束 |
| `pre-tool` | 工具执行前(PreToolUse) |
| `post-tool` | 工具执行后(PostToolUse) |
| `post-tool-failure` | 工具执行失败后 |
| `stop` | 会话停止 |
| `stop-failure` | 停止失败 |
| `compact` | 上下文压缩前(PreCompact) |
| `post-compact` | 上下文压缩后 |
| `notification` | 系统通知 |
| `subagent-start` | 子智能体开始 |
| `subagent-stop` | 子智能体结束 |
| `user-prompt-submit` | 用户提示提交 |
| `permission-request` | 权限请求 |
| `permission-denied` | 权限拒绝 |
| `teammate-idle` | 队友空闲状态 |
| `task-completed` | 任务完成 |
| `task-created` | 任务创建 |
| `worktree-create` | 工作树创建 |
| `worktree-remove` | 工作树移除 |
| `instructions-loaded` | 指令加载完成 |
| `config-change` | 配置变更 |
| `cwd-changed` | 工作目录变更 |
| `file-changed` | 文件变更 |
| `elicitation` | MCP elicitation 请求 |
| `elicitation-result` | MCP elicitation 结果 |

钩子不由用户直接执行 —— Claude Code 的 `settings.json` 会自动调用。

---

## moai worktree

管理 Git worktree 以进行并行 SPEC 开发。

```bash
moai worktree <COMMAND> [ARGS]...
```

### 子命令

| 命令 | 说明 |
|--------|------|
| `moai worktree new <SPEC_ID>` | 创建新 worktree |
| `moai worktree list` | 活跃 worktree 列表 |
| `moai worktree go <SPEC_ID>` | 移动到 worktree 目录 |
| `moai worktree remove <SPEC_ID>` | 移除 worktree |
| `moai worktree clean` | 清理过期 worktree |
| `moai worktree recover` | 从既有目录恢复 |
| `moai worktree status` | 查询 worktree 状态 |

---

## moai cc / moai cg / moai glm

在启动 Claude Code 时选择后端的启动命令。三条命令都能用 `-p <profile>` 标志指定配置文件,并将 `--` 之后的参数原样传给 Claude Code。

```bash
moai cc [-p profile] [-- claude-args...]
moai cg [-p profile] [-- claude-args...]
moai glm [-p profile] [-- claude-args...]
```

| 命令 | 领导 | Worker | 需要 tmux | 用途 |
|--------|------|------|-----------|------|
| `moai cc` | Claude | Claude | 否 | 最高质量(单一后端) |
| `moai glm` | GLM | GLM | 推荐 | 成本优化(GLM 单独) |
| `moai cg` | Claude | GLM | 必需 | 质量 + 成本平衡(混合) |

`moai cg` 激活 CG 模式(Claude 领导 + GLM 队友)。必须在 tmux 会话内运行,它会把 GLM 环境变量注入 tmux 会话,而领导窗口使用 Claude API。

```bash
# 1. 保存 GLM API 密钥(首次一次)
moai glm sk-your-glm-api-key

# 2. 激活 CG 模式(在 tmux 内运行)
moai cg

# 3. 在同一窗口启动 Claude Code
claude
```

详细的 CG 模式指南请参阅[简介 — 用 GLM 节省 token](./introduction#用-glm-节省-token5070)。

---

## moai version

显示版本、commit 哈希、构建日期。

```bash
moai version
moai --version    # 相同
```

---

## 模型策略(性能层级)

MoAI-ADK 提供为智能体分配最优 AI 模型的性能层级系统 —— 这是代币经济学的起点。通过 `llm.yaml` 的 `performance_tier` 字段设置,用 `--model-policy` 标志或初始化向导选择。

| 层级 | 特点 |
|------|------|
| **max** | 最高质量 —— 计划·审计分配 Opus,最大推理深度 |
| **medium**(默认) | 质量与成本的平衡 |
| **low** | 经济 —— 以 Sonnet 为中心分配 |

```bash
# 初始化时设置
moai init my-project --model-policy max

# 在既有项目中重新设置
moai update -c
```

计费套餐类型(`plan_type`: api 或 subscription)单独设置,即使层级相同,也会因计费方式不同而使模型分配不同。详细的模型-层级映射请参阅[模型策略](/zh/multi-llm/model-policy)页面。

---

## 参考

- [快速开始](./quickstart)
- [安装](./installation)
- [更新](./update)
- [初始设置](./init-wizard)
- [配置文件管理](./profile)
- [项目状态](./status)
