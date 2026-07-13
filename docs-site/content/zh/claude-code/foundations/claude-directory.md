---
title: .claude 目录
weight: 60
draft: false
description: ".claude 目录是 Claude Code 按项目读取 CLAUDE.md、settings.json、技能、子智能体和 hook 的配置根。本文梳理其结构与作用域。"
---

# .claude 目录

`.claude` 目录是 Claude Code 为每个项目读取指令、配置与扩展功能的单一配置根。

{{< callout type="info" >}}
**一句话总结**：`.claude` 是 Claude Code 在每次会话启动时查看的项目专属"操作面板"，其中大部分提交到 git 与团队共享，只有个人用文件单独隔离。
{{< /callout >}}

对大多数用户来说，只编辑 `CLAUDE.md` 和 `settings.json` 两个文件就足够了。其余的技能、rules、子智能体在需要时逐一添加即可。

## .claude 目录的角色

Claude Code 从两处读取配置。一处是当前项目的 `.claude/` 目录，另一处是主目录下的 `~/.claude/`。项目内的文件提交到 git 与团队共享，`~/.claude/` 中的文件则作为适用于所有项目的个人配置保留。

- **传递项目上下文**：像 `CLAUDE.md` 这样 Claude "阅读并遵循"的指引
- **强制行为**：像 `settings.json` 的权限 (permissions) 与 hook 这样、无论 Claude 是否遵守都会被"强制执行"的设置
- **存放扩展功能**：技能、子智能体、动态工作流等可复用资产

这里的关键区分是**指引** (guidance) 与**配置** (configuration)。`CLAUDE.md` 或 rules 是供 Claude 参考的说明，不能保证始终被遵守；而 hook 与 permissions 由运行时直接强制执行，是确定性的。需要确定性行为时，应当用 hook 或 permissions 而非指引来实现。这一区分正是挽具工程的第一个设计决策 —— MoAI-ADK 也是通过一次 `moai init` 向该目录部署编排器指引 (CLAUDE.md)、质量门禁 hook、智能体·技能资产，构成项目专属挽具。

## 项目 .claude/ 目录结构

| 项目 | 位置 | 提交 | 角色 |
| --- | --- | --- | --- |
| `CLAUDE.md` | 项目根或 `.claude/` | ✓ | 每次会话作为上下文加载的项目指引 |
| `settings.json` | `.claude/` | ✓ | 权限、hook、环境变量、默认模型等被强制执行的设置 |
| `settings.local.json` | `.claude/` | - | 个人用设置覆盖（自动 gitignore） |
| `rules/` | `.claude/` | ✓ | 按主题拆分的指引，可按文件路径条件加载 |
| `skills/` | `.claude/` | ✓ | 用 `/name` 调用或由 Claude 自动调用的技能 |
| `commands/` | `.claude/` | ✓ | 单文件提示词（与技能同一机制） |
| `agents/` | `.claude/` | ✓ | 拥有独立上下文窗口的子智能体定义 |
| `workflows/` | `.claude/` | ✓ | 协调多个子智能体的动态工作流脚本 |
| `hooks/` | `.claude/` | ✓ | hook 执行的脚本（在 settings.json 中注册） |
| `agent-memory/` | `.claude/` | ✓ | 子智能体专用持久记忆 |
| `.mcp.json` | 项目根 | ✓ | 团队共享的 MCP 服务器配置 |
| `.worktreeinclude` | 项目根 | ✓ | 创建 worktree 时要复制的 gitignore 模式 |

### 指引文件（Claude 阅读的内容）

**`CLAUDE.md`**：承载项目规则、常用命令与架构脉络。每次会话全文加载为上下文，因此建议保持在 200 行以内，变长时拆分为 rules。

**`rules/*.md`**：没有 `paths:` frontmatter 时在会话启动时加载；有 `paths:` glob 时仅当对应文件进入上下文时才加载。当 `CLAUDE.md` 接近 200 行时，按主题拆成 rule 是最佳实践。

### 强制执行配置（Claude Code 强制的内容）

**`settings.json`**：包含 `permissions`（工具·命令允许/拦截）、`hooks`（在事件时点执行脚本）、`statusLine`、`model`、`env`、`outputStyle` 键。

**`settings.local.json`**：结构相同但供个人使用，不提交。需要与团队默认值不同的权限时使用。

### 扩展资产

**`skills/<name>/SKILL.md`**：以文件夹为单位的技能，可把参考文档·模板·脚本一并打包。

**`commands/*.md`**：单文件提示词。官方定义为与技能同一机制，新工作流推荐写成技能。

**`agents/*.md`**：拥有自身系统提示与工具访问权限的子智能体。在新的上下文窗口中运行，保持主对话干净。

**`workflows/*.js`**：孵化并协调多个子智能体的动态工作流脚本。

## 全局 ~/.claude/ 目录结构

| 项目 | 位置 | 角色 |
| --- | --- | --- |
| `CLAUDE.md` | `~/.claude/` | 适用于所有项目的个人指引 |
| `settings.json` | `~/.claude/` | 所有项目的默认设置（被项目设置覆盖） |
| `keybindings.json` | `~/.claude/` | 自定义键盘快捷键 |
| `skills/` | `~/.claude/` | 所有项目可用的个人技能 |
| `commands/` | `~/.claude/` | 所有项目可用的个人命令 |
| `agents/` | `~/.claude/` | 所有项目可用的个人子智能体 |
| `workflows/` | `~/.claude/` | 所有项目可用的个人工作流 |
| `output-styles/` | `~/.claude/` | 个人输出样式 |
| `projects/` | `~/.claude/` | 各项目的会话记录、对话转录、自动记忆 |

## 配置作用域与优先级

同一设置可以存在于多个位置，更具体的作用域优先。作用域分为企业、用户、项目三级。

| 作用域 | 位置 | 适用范围 |
| --- | --- | --- |
| 企业 | `managed-settings.json`（按 OS 的系统路径） | 整个组织（用户不可覆盖，最优先） |
| 用户（全局） | `~/.claude/` | 所有项目（个人默认值） |
| 项目 | `.claude/` | 当前项目（团队共享） |
| 项目本地 | `.claude/settings.local.json` | 当前项目、个人（用户可编辑文件中最优先） |

**数组类设置**（`permissions.allow` 等）会把所有作用域的值**合并**。**标量类设置**（`model` 等）则**使用最具体作用域的单一值**。

## 纳入版本管理 vs 排除

| 文件 | 提交 | 原因 |
| --- | --- | --- |
| `CLAUDE.md`、`rules/`、`settings.json` | ✓ | 团队共享的上下文与策略 |
| `skills/`、`commands/`、`agents/`、`workflows/` | ✓ | 团队共享的扩展资产 |
| `.mcp.json` | ✓ | 团队共享的 MCP 服务器配置 |
| `settings.local.json` | - | 个人覆盖（自动 gitignore） |
| 整个 `~/.claude/` | - | 适用于所有项目的个人配置 |
| `CLAUDE.local.md` | - | 项目级个人指引（手动创建后加入 `.gitignore`） |

Claude Code 在首次创建 `settings.local.json` 时会自动把它加入 `.gitignore`。

## 相关文档

- [settings.json 指南](/advanced/settings-json)
- [CLAUDE.md 指南](/advanced/claude-md-guide)
- [Statusline 系统](/advanced/statusline)

## 参考资料

- [Explore the .claude directory（Claude Code 官方文档）](https://code.claude.com/docs/en/claude-directory)

{{< callout type="tip" >}}
新项目只需先填好 `CLAUDE.md` 和 `settings.json` 两个文件：团队权限·hook 放进项目 `settings.json`，只有自己用的权限放进 `settings.local.json`，就能不带 git 冲突地干净起步。
{{< /callout >}}
