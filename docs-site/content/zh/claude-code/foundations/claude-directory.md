---
title: .claude 目录
weight: 60
draft: false
description: ".claude 目录是 Claude Code 按项目读取 CLAUDE.md、settings.json、技能、子智能体和 hook 的配置根。本文梳理其结构和作用域。"
---

# .claude 目录

`.claude` 目录是 Claude Code 按项目读取指令、配置和扩展功能的单一配置根目录。

{{< callout type="info" >}}
**一句话总结**：`.claude` 是 Claude Code 在每次会话开始时查看的项目专属"操作面板"。其中大部分内容会提交到 git 与团队共享，仅个人文件单独隔离。
{{< /callout >}}

对大多数用户来说，编辑 `CLAUDE.md` 和 `settings.json` 这两个文件就足够了。其余的技能、rules、子智能体可在需要时逐个添加。

## .claude 目录的作用

Claude Code 从两个地方读取配置。一个是正在处理的项目的 `.claude/` 目录，另一个是主目录下的 `~/.claude/`。项目内的文件会提交到 git 与团队共享，`~/.claude/` 中的文件则保留为适用于所有项目的个人配置。

- **传递项目上下文**：像 `CLAUDE.md` 这样 Claude "读取并遵循"的指令
- **强制行为**：像 `settings.json` 的权限 (permissions) 和 hook 这样与 Claude 是否遵守无关、被"执行"的配置
- **保管扩展功能**：技能、子智能体、动态工作流等可复用的资产

这里的关键区分是**指引 (guidance)** 与**配置 (configuration)**。`CLAUDE.md` 和 rules 是 Claude 参考的说明文档，因此不保证始终被遵守。而 hook 和 permissions 由运行时直接执行，因此是确定性的。确实需要确定行为时，应该用 hook 或 permissions 来实现，而不是指引。

## 项目 .claude/ 目录结构

| 项目 | 位置 | 提交 | 作用 |
| --- | --- | --- | --- |
| `CLAUDE.md` | 项目根或 `.claude/` | ✓ | 每次会话作为上下文加载的项目指令 |
| `settings.json` | `.claude/` | ✓ | 权限、hook、环境变量、默认模型等被执行的配置 |
| `settings.local.json` | `.claude/` | - | 个人配置覆盖 (自动 gitignore) |
| `rules/` | `.claude/` | ✓ | 按主题拆分的指令，可按文件路径条件加载 |
| `skills/` | `.claude/` | ✓ | 用 `/name` 调用或由 Claude 自动调用的技能 |
| `commands/` | `.claude/` | ✓ | 单文件提示词 (与技能机制相同) |
| `agents/` | `.claude/` | ✓ | 拥有独立上下文窗口的子智能体定义 |
| `workflows/` | `.claude/` | ✓ | 协调多个子智能体的动态工作流脚本 |
| `hooks/` | `.claude/` | ✓ | hook 执行的脚本 (在 settings.json 中注册) |
| `agent-memory/` | `.claude/` | ✓ | 子智能体专用的持久内存 |
| `.mcp.json` | 项目根 | ✓ | 团队共享的 MCP 服务器配置 |
| `.worktreeinclude` | 项目根 | ✓ | worktree 生成时复制的 gitignore 模式 |

### 指引文件 (Claude 读取的内容)

**`CLAUDE.md`**：包含项目的规则、常用命令和架构背景。由于每次会话整个文件都作为上下文加载，建议控制在 200 行以内。过长时拆分到 rules。

**`rules/*.md`**：没有 `paths:` frontmatter 时在会话开始时加载，存在 `paths:` glob 时仅在对应文件进入上下文时加载。当 `CLAUDE.md` 接近 200 行时，按主题拆分为 rule 是最佳实践。

### 执行配置 (Claude Code 强制的内容)

**`settings.json`**：包含 `permissions` (工具·命令的允许/拒绝)、`hooks` (在事件时点执行脚本)、`statusLine`、`model`、`env`、`outputStyle` 等键。

**`settings.local.json`**：模式相同但为个人用，不提交。当需要与团队默认值不同的权限时使用。

### 扩展资产

**`skills/<name>/SKILL.md`**：以文件夹为单位的技能，可将参考文档、模板、脚本一起打包。

**`commands/*.md`**：单文件提示词。官方上与技能机制相同，建议将新工作流编写为技能。

**`agents/*.md`**：拥有自己的系统提示词和工具访问权限的子智能体。在新的上下文窗口中运行，保持主对话整洁。

**`workflows/*.js`**：生成并协调多个子智能体的动态工作流脚本。

## 全局 ~/.claude/ 目录结构

| 项目 | 位置 | 作用 |
| --- | --- | --- |
| `CLAUDE.md` | `~/.claude/` | 适用于所有项目的个人指令 |
| `settings.json` | `~/.claude/` | 所有项目的默认配置 (被项目设置覆盖) |
| `keybindings.json` | `~/.claude/` | 自定义键盘快捷键 |
| `skills/` | `~/.claude/` | 所有项目可用的个人技能 |
| `commands/` | `~/.claude/` | 所有项目可用的个人命令 |
| `agents/` | `~/.claude/` | 所有项目可用的个人子智能体 |
| `workflows/` | `~/.claude/` | 所有项目可用的个人工作流 |
| `output-styles/` | `~/.claude/` | 个人输出风格 |
| `projects/` | `~/.claude/` | 按项目保存的会话记录、对话转录、自动内存 |

## 配置作用域与优先级

同一配置可能存在于多个位置，更具体的作用域优先级越高。作用域分为企业、用户、项目三个层级。

| 作用域 | 位置 | 适用范围 |
| --- | --- | --- |
| 企业 | `managed-settings.json` (各 OS 系统路径) | 整个组织 (用户无法覆盖，最优先) |
| 用户 (全局) | `~/.claude/` | 所有项目 (个人默认值) |
| 项目 | `.claude/` | 当前项目 (团队共享) |
| 项目本地 | `.claude/settings.local.json` | 当前项目、个人 (用户编辑文件中最优先) |

**数组型配置** (`permissions.allow` 等) 会**合并**所有作用域的值。**标量型配置** (`model` 等) 会使用最具体作用域的**单个值**。

## 纳入版本管理 vs 排除

| 文件 | 提交 | 原因 |
| --- | --- | --- |
| `CLAUDE.md`、`rules/`、`settings.json` | ✓ | 团队共享的上下文和策略 |
| `skills/`、`commands/`、`agents/`、`workflows/` | ✓ | 团队共享的扩展资产 |
| `.mcp.json` | ✓ | 团队共享的 MCP 服务器配置 |
| `settings.local.json` | - | 个人覆盖 (自动 gitignore) |
| `~/.claude/` 全部 | - | 适用于所有项目的个人配置 |
| `CLAUDE.local.md` | - | 按项目的个人指令 (手动创建后添加到 `.gitignore`) |

Claude Code 首次创建 `settings.local.json` 时，会自动将其添加到 `.gitignore`。

## 关联文档

- [settings.json 指南](/advanced/settings-json)
- [CLAUDE.md 指南](/advanced/claude-md-guide)
- [Statusline 系统](/advanced/statusline)

## 参考资料

- [Explore the .claude directory (Claude Code 官方文档)](https://code.claude.com/docs/en/claude-directory)

{{< callout type="tip" >}}
如果是新项目，先只填写 `CLAUDE.md` 和 `settings.json` 这两个文件。把团队权限和 hook 放在项目的 `settings.json` 中，把只有自己使用的权限放在 `settings.local.json` 中，就能在没有 git 冲突的情况下干净地起步。
{{< /callout >}}
