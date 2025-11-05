---
title: MoAI-ADK (智能开发工具包)
description: AI驱动的测试驱动开发框架，通过SPEC → TDD → 代码 → 文档的自然衔接，提供完整的AI协作开发工作流程
---

# MoAI-ADK (智能开发工具包)

[简体中文](index.md) | [English](../../index.md) | [한국어](../../ko/index.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![Coverage](https://img.shields.io/badge/coverage-97.7%25-brightgreen)](https://github.com/modu-ai/moai-adk)

> **MoAI-ADK 提供自然衔接 AI 与 SPEC（规格说明）→ TDD（测试驱动开发）→ 代码 → 文档的开发工作流程。**

---

## 1. MoAI-ADK 一览

MoAI-ADK 通过三个核心原则革新 AI 协作开发。通过下方导航，快速定位适合您情况的章节。

如果您是**第一次接触 MoAI-ADK**，请从"什么是 MoAI-ADK"开始。想要**快速开始**，可以直接跳转到"5分钟快速入门"。已经**安装完成并理解概念**，推荐"核心概念轻松理解"。

| 问题                                      | 快速导航                                          |
| ----------------------------------------- | ------------------------------------------------- |
| 初次接触，这是什么？                      | [什么是 MoAI-ADK？](#什么是-moai-adk)             |
| 如何开始？                                | [5分钟快速入门](#5分钟快速入门)                   |
| 想了解基本流程？                          | [基本工作流程 (0 → 3)](#基本工作流程-0--3)        |
| Plan / Run / Sync 命令的作用是什么？       | [核心命令摘要](#核心命令摘要)                     |
| SPEC·TDD·TAG 是什么？                     | [核心概念轻松理解](#核心概念轻松理解)             |
| 对代理/Skills 感兴趣？                     | [子代理 & Skills 概述](#子代理--skills-概述)       |
| 对 Claude Code Hooks 感兴趣？              | [Claude Hooks 指南](#claude-hooks-指南)           |
| 想要深入学习？                            | [更多资源](#更多资源)                             |

---

## 什么是 MoAI-ADK？

### 问题：AI 开发的信任危机

如今，许多开发者希望获得 Claude 或 ChatGPT 的帮助，但无法摆脱一个根本性的疑虑：**"我真能相信 AI 生成的代码吗？"**

现实情况是：当我们让 AI"创建登录功能"时，虽然会得到语法完美的代码，但会反复出现以下问题：

- **需求不明确**："到底要创建什么？"这个基本问题没有得到回答。邮箱/密码登录？OAuth？双重认证（2FA）？一切都基于推测。
- **测试缺失**：大多数 AI 只测试"正常路径"。密码错误怎么办？网络错误怎么办？3 个月后生产环境会爆发 bug。
- **文档不一致**：代码修改后，文档保持原样。"这段代码为什么在这里？"的问题反复出现。
- **上下文丢失**：即使在同一项目中，每次都必须从头开始解释。项目结构、决策原因、之前的尝试都没有被记录。
- **变更影响无法掌握**：需求变更时，无法追踪哪些代码会受到影响。

### 解决方案：SPEC-First TDD with Alfred 超级代理

**MoAI-ADK**（MoAI 智能开发工具包）是一个旨在**系统性地解决**这些问题的开源框架。

核心原理简单但强大：

> **"没有代码就没有测试，没有测试就没有 SPEC"**

更准确的说法是逆向：

> **"SPEC 优先出现。没有 SPEC 就没有测试。没有测试和代码，文档就不完整。"**

遵循这个顺序，您将体验到永不失败的代理式编码：

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span> 明确的需求**
使用 `/alfred:1-plan` 命令首先编写 SPEC。"登录功能"这个模糊请求转换为"当提供有效凭证时，必须发放 JWT 令牌"的**明确需求**。Alfred 的 spec-builder 使用 EARS 语法，短短 3 分钟就能创建专业的 SPEC。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span> 测试保证**
在 `/alfred:2-run` 中自动进行测试驱动开发（TDD）。按照 RED（失败的测试）→ GREEN（最小实现）→ REFACTOR（代码整理）的顺序进行，**保证测试覆盖率在 85% 以上**。不再有"稍后测试"，测试引领代码编写。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span> 文档自动同步**
只需一个 `/alfred:3-sync` 命令，代码、测试、文档全部**保持最新状态同步**。README、CHANGELOG、API 文档，甚至动态文档都会自动更新。6 个月后，代码和文档仍然保持一致。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_4</span> @TAG 系统追踪**
所有代码、测试、文档都加上 `@TAG:ID`。以后需求变更时，`rg "@SPEC:EX-AUTH-001"` 一个命令就能**找到所有相关**的测试、实现、文档。重构时充满信心。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_5</span> Alfred 记住上下文**
AI 代理协作，记住项目的**所有结构、决策原因、工作历史**。无需重复相同的问题。

### MoAI-ADK 的三个核心承诺

为了让初学者也能记住，MoAI-ADK 的价值简化为三条：

**第一，SPEC 优先于代码**
先明确定义要创建什么。编写 SPEC 的过程中，能在实现前发现问题。大幅减少与团队成员的沟通成本。

**第二，测试引领代码（TDD）**
实现前先编写测试（RED）。通过最小实现让测试通过（GREEN）。然后整理代码（REFACTOR）。结果：bug 更少，重构更有信心，代码更易理解。

**第三，文档与代码始终保持一致**
只需一个 `/alfred:3-sync` 命令，所有文档自动更新。README、CHANGELOG、API 文档、动态文档始终与代码同步。修改 6 个月前的代码时不再绝望。

---

## 为什么需要它？

### AI 开发的现实挑战

现代 AI 协作开发面临多种挑战。MoAI-ADK **系统性地解决**所有这些问题：

| 担忧                     | 传统方法的问题                  | MoAI-ADK 的解决方案                               |
| ------------------------ | ------------------------------- | ------------------------------------------------ |
| "无法信任 AI 代码"       | 无测试实现，验证方法不明确      | 强制 SPEC → TEST → CODE 顺序，保证覆盖率 85%+    |
| "重复相同解释"           | 上下文丢失，项目历史未记录      | Alfred 记住所有信息，19 个 AI 团队协作           |
| "编写提示词困难"         | 不知道如何编写好的提示词        | `/alfred` 命令自动提供标准化提示词               |
| "文档总是过时"           | 代码修改后忘记更新文档          | `/alfred:3-sync` 一个命令自动同步                 |
| "不知道在哪里修改"       | 代码搜索困难，意图不明确        | @TAG 链接 SPEC → TEST → CODE → DOC                |
| "团队入职时间长"         | 新团队成员无法掌握代码上下文    | 阅读 SPEC 即可立即理解意图                       |

### 立即可体验的收益

引入 MoAI-ADK 的瞬间，您将感受到：

- **开发速度提升**：明确的 SPEC 减少反复说明时间
- **Bug 减少**：基于 SPEC 的测试提前发现问题
- **代码理解度提升**：通过 @TAG 和 SPEC 立即把握意图
- **维护成本降低**：代码与文档始终一致
- **团队协作效率**：通过 SPEC 和 TAG 实现明确沟通

---

## ⚡ 3 分钟极速入门

通过 MoAI-ADK **三个步骤**开始第一个项目。初学者也能在 5 分钟内完成。

### 步骤 <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span>：安装（约 1 分钟）

#### UV 安装命令

```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows (PowerShell)
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### 实际输出（示例）

```bash
# UV 版本确认
uv --version
✓ uv 0.5.1 is already installed

$ uv --version
uv 0.5.1
```

#### 下一步：安装 MoAI-ADK

```bash
uv tool install moai-adk

# 结果: <span class="material-icons">check_circle</span> Installed moai-adk
```

**验证**：

```bash
moai-adk --version
# 输出: MoAI-ADK v1.0.0
```

---

### 步骤 <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span>：创建第一个项目（约 2 分钟）

#### 命令

```bash
moai-adk init hello-world
cd hello-world
```

#### 实际创建的内容

```
hello-world/
├── .moai/              <span class="material-icons">check_circle</span> Alfred 设置
├── .claude/            <span class="material-icons">check_circle</span> Claude Code 自动化
└── CLAUDE.md           <span class="material-icons">check_circle</span> 项目指南
```

#### 验证：检查核心文件

```bash
# 检查核心设置文件
ls -la .moai/config.json  # <span class="material-icons">check_circle</span> 是否存在？
ls -la .claude/commands/  # <span class="material-icons">check_circle</span> 是否有命令？

# 或一次检查
moai-adk doctor
```

**输出示例**：

```
<span class="material-icons">check_circle</span> Python 3.13.0
<span class="material-icons">check_circle</span> uv 0.5.1
<span class="material-icons">check_circle</span> .moai/ directory initialized
<span class="material-icons">check_circle</span> .claude/ directory ready
<span class="material-icons">check_circle</span> 16 agents configured
<span class="material-icons">check_circle</span> 74 skills loaded
```

---

### 步骤 3️�：启动 Alfred（约 1-2 分钟）

#### 运行 Claude Code

```bash
claude
```

#### 在 Claude Code 中输入以下内容

```
/alfred:0-project
```

#### Alfred 会询问的内容

```
Q1: 项目名称？
A: hello-world

Q2: 项目目标？
A: 学习 MoAI-ADK

Q3: 主要开发语言？
A: python

Q4: 模式？
A: personal（用于本地开发）
```

#### 结果：项目准备完成！<span class="material-icons">check_circle</span>

```
<span class="material-icons">check_circle</span> 项目初始化完成
<span class="material-icons">check_circle</span> 设置保存到 .moai/config.json
<span class="material-icons">check_circle</span> 在 .moai/project/ 中创建文档
<span class="material-icons">check_circle</span> Alfred 完成技能推荐

下一步: /alfred:1-plan "第一个功能说明"
```

---

## <span class="material-icons">rocket_launch</span> 下一步：10 分钟内完成第一个功能

现在来实际**创建功能并自动生成文档**！

> **→ 转到下一节：["第一次 10 分钟实践：Hello World API"](#第一次-10-分钟实践hello-world-api)**

本节包括：

- <span class="material-icons">check_circle</span> 用 SPEC 定义简单 API
- <span class="material-icons">check_circle</span> 完全体验 TDD（RED → GREEN → REFACTOR）
- <span class="material-icons">check_circle</span> 体验自动文档生成
- <span class="material-icons">check_circle</span> 理解 @TAG 系统

---

## <span class="material-icons">auto_stories</span> 安装和项目设置完整指南

快速入门后如需更详细说明，请参考下文。

### 安装详细指南

**uv 安装后额外确认**：

```bash
# PATH 设置确认（如需要）
export PATH="$HOME/.cargo/bin:$PATH"

# 再次确认
uv --version
```

**MoAI-ADK 安装后也可使用其他命令**：

```bash
moai-adk init          # 项目初始化
moai-adk doctor        # 系统诊断
moai-adk update        # 更新到最新版本
```

### MCP（模型上下文协议）设置指南

MoAI-ADK 自动安装和配置遵循 Microsoft MCP 标准的 4 个核心 MCP 服务器。

#### <span class="material-icons">settings</span> MCP 服务器类型和用途

| MCP 服务    | 主要功能                     | 目标代理                     | 安装方式             |
|-----------|----------------------------|----------------------------|--------------------|
| **Context7** | 最新库文档搜索                 | 所有专家代理                   | NPX 自动安装         |
| **Figma**    | 设计系统和组件规格              | ui-ux-expert               | Claude Code 官方远程服务器 |
| **Playwright** | Web E2E 测试自动化           | frontend-expert, tdd-implementer, quality-gate | NPX 自动安装 |
| **Sequential Thinking** | 复杂推理和逻辑分析        | spec-builder, implementation-planner, security-expert | NPX 自动安装 |

#### <span class="material-icons">rocket_launch</span> 自动 MCP 设置（moai-adk init）

**运行 moai-adk init 时自动安装 MCP 服务器**：

```bash
# 包含 MCP 服务器的项目初始化
moai-adk init my-project --with-mcp

# 或为现有项目添加 MCP
cd your-project
moai-adk init . --with-mcp
```

**自动生成的 MCP 设置文件 (.claude/mcp.json)**：

```json
{
  "servers": {
    "context7": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@upstash/context7-mcp"
      ],
      "env": {}
    },
    "figma": {
      "type": "http",
      "url": "https://mcp.figma.com/mcp",
      "headers": {
        "Authorization": "Bearer ${FIGMA_ACCESS_TOKEN}"
      }
    },
    "playwright": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@playwright/mcp"
      ],
      "env": {}
    },
    "sequential-thinking": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sequential-thinking"
      ],
      "env": {}
    }
  }
}
```

#### <span class="material-icons">settings</span> Figma Access Token 设置

**Claude Code 官方 Figma MCP 使用远程服务器**：

1. **生成 Figma Access Token**
   - 访问：https://www.figma.com/developers/api#access-tokens
   - 用适当权限创建新的 Access Token

2. **设置令牌**（选择其中一种方法）

   **环境变量方法（推荐）**：
   ```bash
   # 添加到 shell 配置文件 (~/.zshrc 或 ~/.bashrc)
   export FIGMA_ACCESS_TOKEN="your_figma_token_here"

   # 立即生效
   source ~/.zshrc  # 或 source ~/.bashrc
   ```

   **Claude Code 设置方法**：
   ```bash
   claude-code settings
   ```

3. **重启 Claude Code** 激活令牌

**注意**：Figma MCP 使用 Claude Code 官方远程服务器(https://mcp.figma.com/mcp)，无需单独本地安装。

#### <span class="material-icons">target</span> 遵循 Microsoft MCP 标准

**设置标准**：
- **文件**：`.claude/mcp.json`（Microsoft MCP 标准）
- **格式**：在 `servers` 对象中明确各服务器 `type: "stdio"` 或 `type: "http"`
- **命令**：所有本地服务器使用 `npx` 和 `-y` 标志自动安装
- **向后兼容**：在 `.claude/settings.json` 中也复制设置以支持旧版

#### <span class="material-icons">check_circle</span> 代理功能扩展

MCP 服务器正常安装后，以下代理会自动扩展功能：

- **ui-ux-expert**：Figma 设计系统集成（官方远程服务器）
- **spec-builder**：Sequential Thinking 支持复杂 SPEC 编写
- **implementation-planner**：多步骤计划制定时强化推理
- **backend-expert**：架构设计时支持系统思考
- **database-expert**：模式设计时添加逻辑分析
- **security-expert**：威胁分析时支持分步思考过程
- **frontend-expert**：Context7 查阅最新文档
- **tdd-implementer**：Playwright 自动生成测试
- **quality-gate**：Web 质量验证自动化
- **所有专家**：Context7 实时文档查询

#### <span class="material-icons">search</span> 问题解决

**MCP 服务器不可见时**：
1. 重启 Claude Code
2. 检查 `.claude/mcp.json` 文件语法
3. 用 `claude-code --version` 检查 Claude Code 版本（需要 v1.5.0+）
4. 用 `npx --version` 检查 npm/npx 版本
5. 确认 Figma Access Token 正确设置

**详细 MCP 设置指南**：
- [Claude Code MCP Documentation (English)](https://docs.claude.com/en/docs/claude-code/mcp)
- [Microsoft MCP Standard](https://modelcontextprotocol.io)

### 项目创建详细指南

**创建新项目**：

```bash
moai-adk init my-project
cd my-project
```

**添加到现有项目**：

```bash
cd your-existing-project
moai-adk init .
```

创建的完整结构：

```
my-project/
├── .moai/                          # MoAI-ADK 项目设置
│   ├── config.json                 # 项目设置（语言、模式、所有者）
│   ├── project/                    # 项目信息
│   │   ├── product.md              # 产品愿景和目标
│   │   ├── structure.md            # 目录结构
│   │   └── tech.md                 # 技术栈和架构
│   ├── memory/                     # Alfred 的知识库（8个文件）
│   ├── specs/                      # SPEC 文件
│   └── reports/                    # 分析报告
├── .claude/                        # Claude Code 自动化
│   ├── agents/                     # 16个子代理（包括专家）
│   ├── commands/                   # 4个 Alfred 命令
│   ├── skills/                     # 74个 Claude Skills
│   ├── hooks/                      # 5个事件自动化钩子
│   └── settings.json               # Claude Code 设置
└── CLAUDE.md                       # Alfred 的核心指令
```

---

## 核心概念：3步循环

设置完成后，所有功能开发都重复这 3 个步骤：

| 步骤        | 命令                       | 执行工作                     | 时间 |
| ----------- | -------------------------- | ---------------------------- | ---- |
| 📋 **PLAN** | `/alfred:1-plan "功能说明"` | SPEC 编写（EARS 格式）         | 2分钟 |
| 💻 **RUN**  | `/alfred:2-run SPEC-ID`    | TDD 实现（RED→GREEN→REFACTOR） | 5分钟 |
| <span class="material-icons">menu_book</span> **SYNC** | `/alfred:3-sync`           | 文档自动同步                 | 1分钟 |

**一个循环 ≈ 8分钟** → **一天可完成 7-8 个功能** ⚡

---

## 📦 保持 MoAI-ADK 最新版本

### 版本确认

```bash
# 检查当前安装版本
moai-adk --version

# 检查 PyPI 最新版本
uv tool list  # 检查 moai-adk 当前版本
```

### 升级

MoAI-ADK 提供**两种更新机制**：

1. **`moai-adk update`**：包版本 + 模板同步（推荐）
2. **`uv tool upgrade`**：标准 uv 工具升级（选择）

#### 方法 1：moai-adk 自身更新命令（推荐 - 最完整）

此方法同时更新包版本并自动同步本地模板。

```bash
# 步骤 1：MoAI-ADK 包更新（+ 模板同步）
moai-adk update
```

**更新了什么？**

- <span class="material-icons">check_circle</span> `moai-adk` 包本身（PyPI 最新版本）
- <span class="material-icons">check_circle</span> 16个子代理模板
- <span class="material-icons">check_circle</span> 74个 Claude Skills
- <span class="material-icons">check_circle</span> 5个 Claude Code Hooks
- <span class="material-icons">check_circle</span> 4个 Alfred 命令定义

---

## 核心命令摘要

| 命令                        | 作用是什么？                                                   | 主要输出                                                       |
| --------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------- |
| `/alfred:0-project`         | 项目初始化：设置收集、文档创建、技能推荐                        | `.moai/config.json`、`.moai/project/*`、初始报告              |
| `/alfred:0-project setting` | 修改现有设置：语言、昵称、GitHub 设置、报告生成选项更改          | 更新的 `.moai/config.json`                                    |
| `/alfred:0-project update`  | 模板优化：`moai-adk update` 后保留用户自定义                     | 合并的 `.claude/`、`.moai/` 模板文件                          |
| `/alfred:1-plan <说明>`     | 需求分析、SPEC 草案、计划板编写                               | `.moai/specs/SPEC-*/spec.md`、计划/验收文档、功能分支          |
| `/alfred:2-run <SPEC-ID>`   | TDD 执行、测试/实现/重构、质量验证                              | `tests/`、`src/` 实现、质量报告、TAG 链接                     |
| `/alfred:3-sync`            | 文档/README/CHANGELOG 同步、TAG/PR 状态整理                      | `docs/`、`.moai/reports/sync-report.md`、Ready PR             |
| `/alfred:9-feedback`        | MoAI-ADK 改进反馈 GitHub Issue 创建（类型 → 标题 → 说明 → 优先级） | GitHub Issue + 自动标签 + 优先级 + URL                       |

> ❗ 所有命令都保持 **Phase 0(可选) → Phase 1 → Phase 2 → Phase 3** 循环结构。Alfred 会自动报告运行状态和下一步建议。

---

## 核心概念轻松理解

MoAI-ADK 由 5 个核心概念组成。每个概念相互连接，共同作用时创建强大的开发系统。

### 核心概念 1：SPEC-First（需求优先）

**比喻**：就像没有建筑师不能建房子一样，没有设计图就不能编码。

**核心**：实现前明确**"要创建什么"**。这不是简单文档，而是团队和 AI 能共同理解的**可执行规格**。

**EARS 语法的 5 种模式**：

1. **Ubiquitous**（基本功能）："系统必须提供基于 JWT 的认证"
2. **Event-driven**（条件）："**当**提供有效凭证时，系统必须发放令牌"
3. **State-driven**（状态）："**当**用户处于认证状态时，系统必须允许访问受保护资源"
4. **Optional**（可选）："**如果**有刷新令牌，系统可以发放新令牌"
5. **Constraints**（约束）："令牌过期时间不得超过 15 分钟"

**如何实现？** `/alfred:1-plan` 命令自动以 EARS 格式创建专业 SPEC。

**收获**：

- <span class="material-icons">check_circle</span> 团队所有人都理解的明确需求
- <span class="material-icons">check_circle</span> 基于 SPEC 的测试用例（已经定义要测试什么）
- <span class="material-icons">check_circle</span> 需求变更时通过 `@SPEC:ID` TAG 追踪所有受影响的代码

---

### 核心概念 2：TDD（测试驱动开发）

**比喻**：就像确定目的地后找路一样，用测试确定目标再编写代码。

**核心**：**实现**前先编写**测试**。这就像做饭前确认食材一样，在实现前明确需求是什么。

**3步循环**：

1. **🔴 RED**：先编写失败的测试

   - SPEC 的每个需求成为测试用例
   - 还没有实现，所以必然失败
   - Git 提交：`test(AUTH-001): add failing test`

2. **🟢 GREEN**：最小实现让测试通过

   - 最简单方法让测试通过
   - 通过优先于完美
   - Git 提交：`feat(AUTH-001): implement minimal solution`

3. **<span class="material-icons">recycling</span> REFACTOR**：整理和改进代码
   - 应用 TRUST 5原则
   - 消除重复，提高可读性
   - 测试必须仍然通过
   - Git 提交：`refactor(AUTH-001): improve code quality`

**如何实现？** `/alfred:2-run` 命令自动执行这 3 个步骤。

**收获**：

- <span class="material-icons">check_circle</span> 保证覆盖率 85% 以上（没有无测试的代码）
- <span class="material-icons">check_circle</span> 重构信心（随时可以通过测试验证）
- <span class="material-icons">check_circle</span> 明确的 Git 历史（追踪 RED → GREEN → REFACTOR 过程）

---

### 核心概念 3：@TAG 系统

**比喻**：就像快递运单一样，必须能追踪代码的旅程。

**核心**：所有 SPEC、测试、代码、文档都加上 `@TAG:ID` 建立**一对一对应**。

**TAG 链**：

```
@SPEC:EX-AUTH-001 (需求)
    ↓
@TEST:EX-AUTH-001 (测试)
    ↓
@CODE:EX-AUTH-001 (实现)
    ↓
@DOC:EX-AUTH-001 (文档)
```

**TAG ID 规则**：`<领域>-<3位数字>`

- AUTH-001, AUTH-002, AUTH-003...
- USER-001, USER-002...
- 一旦分配**绝不更改**

**如何使用？** 需求变更时：

```bash
# 查找与 AUTH-001 相关的所有内容
rg '@TAG:AUTH-001' -n

# 结果：SPEC、TEST、CODE、DOC 全部一次显示
# → 明确在哪里修改
```

**如何实现？** `/alfred:3-sync` 命令验证 TAG 链，检测孤立 TAG（无对应 TAG）。

**收获**：

- <span class="material-icons">check_circle</span> 所有代码意图明确（读 SPEC 就明白为什么有这段代码）
- <span class="material-icons">check_circle</span> 重构时立即掌握所有受影响的代码
- <span class="material-icons">check_circle</span> 3 个月后仍能理解代码（TAG → SPEC 追踪）

---

### 核心概念 4：TRUST 5原则

**比喻**：就像健康身体一样，好代码必须满足所有 5 个要素。

**核心**：所有代码必须遵守以下 5 个原则。`/alfred:3-sync` 自动验证。

1. **🧪 Test First**（测试优先）

   - 测试覆盖率 ≥ 85%
   - 所有代码受测试保护
   - 功能添加 = 测试添加

2. **<span class="material-icons">auto_stories</span> Readable**（可读代码）

   - 函数 ≤ 50行，文件 ≤ 300行
   - 变量名体现意图
   - 通过 linter（ESLint/ruff/clippy）

3. **<span class="material-icons">target</span> Unified**（一致结构）

   - 保持基于 SPEC 的架构
   - 相同模式重复（学习曲线降低）
   - 类型安全或运行时验证

4. **<span class="material-icons">lock</span> Secured**（安全）

   - 输入验证（防御 XSS、SQL Injection）
   - 密码哈希（bcrypt、Argon2）
   - 敏感信息保护（环境变量）

5. **<span class="material-icons">link</span> Trackable**（可追踪）

   - 使用 @TAG 系统
   - Git 提交包含 TAG
   - 所有决策文档化

**如何实现？** `/alfred:3-sync` 命令自动执行 TRUST 验证。

**收获**：

- <span class="material-icons">check_circle</span> 保证生产级代码质量
- <span class="material-icons">check_circle</span> 团队按相同标准开发
- <span class="material-icons">check_circle</span> 减少 bug，预防安全漏洞

---

### 核心概念 5：Alfred 超级代理

**比喻**：就像个人助理一样，Alfred 处理所有复杂工作。

**核心**：AI 代理协作自动化整个开发过程：

**代理构成**：

- **Alfred 超级代理**：整体编排
- **核心子代理**：SPEC 编写、TDD 实现、文档同步等专业工作
- **零项目专家**：项目初始化、语言检测等
- **内置代理**：一般问题、代码库搜索

**Claude Skills**：

- **基础**：TRUST/TAG/SPEC/Git/EARS 原则
- **核心**：调试、性能、重构、代码审查
- **Alfred**：工作流程自动化
- **领域**：后端、前端、安全等
- **语言**：Python、JavaScript、Go、Rust 等
- **运维**：Claude Code 会话管理

**如何实现？** `/alfred:*` 命令自动激活所需专家团队。

**收获**：

- <span class="material-icons">check_circle</span> 无需编写提示词（使用标准化命令）
- <span class="material-icons">check_circle</span> 自动记忆项目上下文（不重复相同问题）
- <span class="material-icons">check_circle</span> 自动配置最佳专家团队（按情况激活相应子代理）

> **想深入了解？** 在 `.moai/memory/development-guide.md` 查看详细规则。

---

## 第一次 10 分钟实践：Hello World API

**目标**：10 分钟内体验 MoAI-ADK 的完整工作流程
**学习内容**：SPEC 编写、TDD 实现、文档自动化、@TAG 系统

> 如果已完成 3 分钟极速入门，可以从本节开始！

### 事前准备

- <span class="material-icons">check_circle</span> MoAI-ADK 安装完成
- <span class="material-icons">check_circle</span> 项目创建完成（`moai-adk init hello-world`）
- <span class="material-icons">check_circle</span> Claude Code 运行中

---

### Step <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span>：编写 SPEC（2分钟）

#### 命令

```bash
/alfred:1-plan "GET /hello 端点 - 接收查询参数 name 返回问候语"
```

#### Alfred 自动生成

```
<span class="material-icons">check_circle</span> SPEC ID: HELLO-001
<span class="material-icons">check_circle</span> 文件: .moai/specs/SPEC-HELLO-001/spec.md
<span class="material-icons">check_circle</span> 分支: feature/SPEC-HELLO-001
```

#### 确认生成的 SPEC

```bash
cat .moai/specs/SPEC-HELLO-001/spec.md
```

**内容示例**：

```yaml
---
id: HELLO-001
version: 0.0.1
status: draft
priority: high
---
# `@SPEC:EX-HELLO-001: Hello World API

## Ubiquitous Requirements
- 系统必须提供 HTTP GET /hello 端点

## Event-driven Requirements
- 当提供查询参数 name 时，必须返回 "Hello, {name}!"
- 当没有 name 时，必须返回 "Hello, World!"

## Constraints
- name 必须限制在最多 50 字符
- 响应必须是 JSON 格式
```

<span class="material-icons">check_circle</span> **验证**：`ls .moai/specs/SPEC-HELLO-001/`

---

### Step <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span>：TDD 实现（5分钟）

#### 命令

```bash
/alfred:2-run HELLO-001
```

#### 🔴 RED 阶段：编写失败测试

**生成文件**：`tests/test_hello.py`

```python
# `@TEST:EX-HELLO-002 | SPEC: SPEC-HELLO-001.md

import pytest
from fastapi.testclient import TestClient
from src.hello.api import app

client = TestClient(app)

def test_hello_with_name_should_return_personalized_greeting():
    """当提供 name 时，必须返回 "Hello, {name}!" """
    response = client.get("/hello?name=张三")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, 张三!"}

def test_hello_without_name_should_return_default_greeting():
    """当没有 name 时，必须返回 "Hello, World!" """
    response = client.get("/hello")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, World!"}
```

**执行**：

```bash
pytest tests/test_hello.py -v
```

**结果**：<span class="material-icons">cancel</span> FAILED（没有 app - 正常！）

**Git 提交**：

```bash
git add tests/test_hello.py
git commit -m "🔴 test(HELLO-001): add failing hello API tests"
```

---

#### 🟢 GREEN 阶段：最小实现

**生成文件**：`src/hello/api.py`

```python
# `@CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI

app = FastAPI()

@app.get("/hello")
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - Hello 端点"""
    return {"message": f"Hello, {name}!"}
```

**执行**：

```bash
pytest tests/test_hello.py -v
```

**结果**：<span class="material-icons">check_circle</span> PASSED（全部通过！）

**Git 提交**：

```bash
git add src/hello/api.py
git commit -m "🟢 feat(HELLO-001): implement hello API"
```

---

#### <span class="material-icons">recycling</span> REFACTOR 阶段：代码改进

**添加验证逻辑**：

```python
from fastapi import FastAPI, HTTPException

app = FastAPI()

@app.get("/hello")
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - 带验证的 Hello 端点"""
    if len(name) > 50:
        raise HTTPException(status_code=400, detail="Name too long (max 50 chars)")
    return {"message": f"Hello, {name}!"}
```

**添加测试**：

```python
def test_hello_with_long_name_should_return_400():
    """name 超过 50 字符时必须返回 400 错误"""
    long_name = "a" * 51
    response = client.get(f"/hello?name={long_name}")
    assert response.status_code == 400
```

**执行**：

```bash
pytest tests/test_hello.py -v
```

**结果**：<span class="material-icons">check_circle</span> PASSED（全部通过！）

**Git 提交**：

```bash
git add tests/test_hello.py src/hello/api.py
git commit -m "<span class="material-icons">recycling</span> refactor(HELLO-001): add name length validation"
```

---

### Step <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span>：文档同步（2分钟）

#### 命令

```bash
/alfred:3-sync
```

#### Alfred 自动处理

```
<span class="material-icons">check_circle</span> docs/api/hello.md - API 文档生成
<span class="material-icons">check_circle</span> README.md - API 使用方法添加
<span class="material-icons">check_circle</span> CHANGELOG.md - v0.1.0 发布说明添加
<span class="material-icons">check_circle</span> TAG 链验证 - 所有 @TAG 确认
```

#### 确认生成的 API 文档

```bash
cat docs/api/hello.md
```

**内容示例**：

````markdown
# Hello API 文档

## GET /hello

### 说明

接收名字并返回个性化问候语。

### 参数

- `name` (query, 可选): 名字（默认值: "World", 最多 50 字符）

### 响应

- **200**: 成功
  ```json
  { "message": "Hello, 张三!" }
  ```

- **400**: 名字过长

### 示例

```bash
curl "http://localhost:8000/hello?name=张三"
# → {"message": "Hello, 张三!"}

curl "http://localhost:8000/hello"
# → {"message": "Hello, World!"}
```

### 可追踪性

- `@SPEC:EX-HELLO-001` - 需求
- `@TEST:EX-HELLO-002` - 测试
- `@CODE:EX-HELLO-001:API` - 实现
````

---

### Step <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_4</span>：TAG 链验证（1分钟）

#### 命令
```bash
rg '@(SPEC|TEST|CODE|DOC):HELLO-001' -n
````

#### 输出（完全可追踪）

```
.moai/specs/SPEC-HELLO-001/spec.md:7:# `@SPEC:EX-HELLO-001: Hello World API
tests/test_hello.py:3:# `@TEST:EX-HELLO-002 | SPEC: SPEC-HELLO-001.md
src/hello/api.py:3:# `@CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md
docs/api/hello.md:24:- `@SPEC:EX-HELLO-001`
```

<span class="material-icons">check_circle</span> **意义**：需求 → 测试 → 实现 → 文档完美连接！

---

### 🎉 10 分钟后：您获得的成果

#### 生成的文件

```
hello-world/
├── .moai/specs/SPEC-HELLO-001/
│   ├── spec.md              ← 需求文档
│   └── plan.md              ← 计划
├── tests/test_hello.py      ← 测试（100% 覆盖率）
├── src/hello/
│   ├── api.py               ← API 实现
│   └── __init__.py
├── docs/api/hello.md        ← API 文档
├── README.md                ← 已更新
└── CHANGELOG.md             ← v0.1.0 发布说明
```

#### Git 历史

```bash
git log --oneline | head -4
```

**输出**：

```
c1d2e3f <span class="material-icons">recycling</span> refactor(HELLO-001): add name length validation
b2c3d4e 🟢 feat(HELLO-001): implement hello API
a3b4c5d 🔴 test(HELLO-001): add failing hello API tests
d4e5f6g Merge branch 'develop' (initial project commit)
```

#### 学习总结

- <span class="material-icons">check_circle</span> **SPEC**：用 EARS 格式明确定义需求
- <span class="material-icons">check_circle</span> **TDD**：体验 RED → GREEN → REFACTOR 循环
- <span class="material-icons">check_circle</span> **自动化**：文档与代码一起自动生成
- <span class="material-icons">check_circle</span> **可追踪性**：@TAG 系统连接所有步骤
- <span class="material-icons">check_circle</span> **质量**：测试 100%、明确实现、自动文档化

---

## <span class="material-icons">rocket_launch</span> 下一步

现在创建更复杂的功能：

```bash
# 开始下一个功能
/alfred:1-plan "用户数据库查询 API"
```

或需要深入示例，请参考下文。

---

## 子代理 & Skills 概述

Alfred 结合多个专业代理和 Claude Skills 进行工作。

### 核心子代理（Plan → Run → Sync）

| 子代理               | 模型   | 职责                                           |
| ------------------- | ------ | ---------------------------------------------- |
| project-manager 📋  | Sonnet | 项目初始化、元数据访谈                         |
| spec-builder <span class="material-icons">construction</span>     | Sonnet | 计划板、EARS SPEC 编写、推荐专家咨询         |
| code-builder 💎     | Sonnet | 用 `implementation-planner` + `tdd-implementer` 执行完整 TDD |
| doc-syncer <span class="material-icons">auto_stories</span>       | Haiku  | 动态文档、README、CHANGELOG 同步              |
| tag-agent <span class="material-icons">label</span>        | Haiku  | TAG 清单、孤立检测、@EXPERT TAG 验证          |
| git-manager <span class="material-icons">rocket_launch</span>      | Haiku  | GitFlow、Draft/Ready、自动合并                |
| debug-helper <span class="material-icons">search</span>     | Sonnet | 失败分析、forward-fix 策略                   |
| trust-checker <span class="material-icons">check_circle</span>    | Haiku  | TRUST 5 质量门禁                               |
| quality-gate <span class="material-icons">shield</span>     | Haiku  | 覆盖率变更和发布阻止条件审查                  |
| cc-manager <span class="material-icons">build</span>       | Sonnet | Claude Code 会话优化、Skill 部署              |
| skill-factory 🏭   | Sonnet | Skills 创建和管理、69个 Skills 生态系统维护 |

### 专家代理（根据 SPEC 关键字自动激活）

专家代理在 `implementation-planner` 从 SPEC 文档检测到领域特定关键字时自动激活。每个专家提供自己领域的架构指南、技术推荐、风险分析。

| 专家代理            | 模型   | 专业领域                             | 自动激活关键字                                                     |
| ------------------- | ------ | ------------------------------------- | ------------------------------------------------------------------ |
| backend-expert <span class="material-icons">settings</span>   | Sonnet | 后端架构、API 设计、DB               | 'backend', 'api', 'server', 'database', 'deployment', 'authentication' |
| frontend-expert 💻  | Sonnet | 前端架构、组件、状态管理              | 'frontend', 'ui', 'page', 'component', 'client-side', 'web interface'  |
| devops-expert <span class="material-icons">rocket_launch</span>    | Sonnet | DevOps、CI/CD、部署、容器             | 'deployment', 'docker', 'kubernetes', 'ci/cd', 'pipeline', 'aws'       |
| ui-ux-expert <span class="material-icons">palette</span>     | Sonnet | UI/UX 设计、可访问性、设计系统        | 'design', 'ux', 'accessibility', 'a11y', 'figma', 'design system'      |

**工作原理**：

- `/alfred:2-run` 开始时，`implementation-planner` 扫描 SPEC 内容
- 匹配关键字自动激活对应专家代理
- 每个专家提供领域特定架构指南
- 所有专家咨询用 `@EXPERT:DOMAIN` 标签标记以保持可追踪性

---

## Claude Hooks 指南

MoAI-ADK 提供 5 个与开发流程无缝集成的 Claude Code Hooks。这些 Hook 在会话开始/结束、工具执行前后、提示提交时自动运行，透明地处理检查点、JIT 上下文加载、会话管理等。

### Hook 是什么？

Hook 是响应 Claude Code 会话特定事件的事件驱动脚本。在不干扰用户流程的情况下，在后台提供安全保护和生产力提升。

### 安装的 Hooks（5个）

| Hook          | 状态   | 功能                                                         |
| ------------- | ------ | ------------------------------------------------------------ |
| SessionStart  | <span class="material-icons">check_circle</span> 激活 | 语言/Git/SPEC 进度/检查点等 项目状态摘要                     |
| PreToolUse    | <span class="material-icons">check_circle</span> 激活 | 风险检测 + 自动检查点(删除/合并/批量编辑/重要文件) + **TAG Guard**（检测缺失的 @TAG） |
| UserPromptSubmit | <span class="material-icons">check_circle</span> 激活 | JIT 上下文加载（自动加载@SPEC·测试·代码·文档）                |
| PostToolUse   | <span class="material-icons">check_circle</span> 激活 | 代码更改后自动测试（Python/TS/JS/Go/Rust/Java 等）          |
| SessionEnd    | <span class="material-icons">check_circle</span> 激活 | 会话清理和状态保存                                           |

---

## <span class="material-icons">settings</span> 初学者问题解决

MoAI-ADK 开始时常见错误和解决方法。

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span> uv 未安装

**症状**：

```bash
$ uv --version
bash: uv: command not found
```

**原因**：uv 未安装或未添加到 PATH

**解决**：

**macOS/Linux**：

```bash
# 安装
curl -LsSf https://astral.sh/uv/install.sh | sh

# 重启 shell
source ~/.bashrc  # 或 ~/.zshrc

# 验证
uv --version
```

**Windows (PowerShell)**：

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# 验证
uv --version
```

**仍然失败时**：

```bash
# 手动添加 PATH (macOS/Linux)
export PATH="$HOME/.cargo/bin:$PATH"

# 再次确认
uv --version
```

---

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span> Python 版本不匹配

**症状**：

```
Python 3.8 found, but 3.13+ required
```

**原因**：Python 版本低于 3.13

**解决**：

**选项 A: 使用 pyenv（推荐）**：

```bash
# 安装 pyenv
curl https://pyenv.run | bash

# 安装 Python 3.13
pyenv install 3.13
pyenv global 3.13

# 验证
python --version  # Python 3.13.x
```

**选项 B: 用 uv 自动管理 Python**：

```bash
# uv 自动下载 Python 3.13
uv python install 3.13
uv python pin 3.13

# 验证
python --version
```

---

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span> Git 未安装

**症状**：

```
✗ Git (runtime): not found
```

**原因**：系统未安装 Git

**解决**：

**macOS**：

```bash
# 用 Homebrew 安装
brew install git

# 或 Xcode Command Line Tools
xcode-select --install
```

**Ubuntu/Debian**：

```bash
sudo apt update
sudo apt install git -y
```

**Windows**：

```powershell
# 用 winget 安装
winget install Git.Git

# 或手动下载
# https://git-scm.com/download/win
```

**验证**：

```bash
git --version  # git version 2.x.x
```

---

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_4</span> Claude Code 无法识别 .moai/ 文件夹

**症状**：

```
"项目未初始化"
/alfred:0-project 命令不工作
```

**原因**：`.moai/` 或 `.claude/` 文件夹不存在或损坏

**解决**：

```bash
# 1. 确认当前目录
pwd  # /path/to/your-project

# 2. 检查 .moai/ 文件夹
ls -la .moai/config.json

# 3. 没有则重新初始化
moai-adk init .

# 4. 重启 Claude Code
exit  # 退出 Claude Code
claude  # 重新启动 Claude Code
```

**验证**：

```bash
moai-adk doctor
# 所有项目应该显示 <span class="material-icons">check_circle</span>
```

---

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_5</span> 测试运行时找不到模块

**症状**：

```
FAILED tests/test_hello.py - ModuleNotFoundError: No module named 'fastapi'
```

**原因**：未安装必要包

**解决**：

```bash
# 在项目根目录安装依赖
uv sync

# 只需要安装特定包
uv add fastapi pytest

# 激活虚拟环境后再次运行
source .venv/bin/activate  # macOS/Linux
.venv\Scripts\activate     # Windows

pytest tests/ -v
```

---

### 6️⃣ /alfred 命令不工作

**症状**：

```
Unknown command: /alfred:1-plan
```

**原因**：Claude Code 版本问题或 `.claude/` 文件夹损坏

**解决**：

```bash
# 1. 检查 Claude Code 版本（最低 v1.5.0+）
claude --version

# 2. 检查 .claude/ 文件夹
ls -la .claude/commands/

# 3. 必要时重新初始化
moai-adk init .

# 4. 重启 Claude Code
exit
claude
```

---

### 7️⃣ TAG 链损坏

**症状**：

```
<span class="material-icons">check_circle</span> Fixed: @TEST:EX-HELLO-002 (TAG ID updated for consistency)
```

**原因**：SPEC 被删除或 TAG 不一致

**解决**：

```bash
# 1. 验证 TAG 链
rg '@(SPEC|TEST|CODE):HELLO-001' -n

# 2. 检查缺失的 TAG
rg '@SPEC:EX-HELLO-001' -n .moai/specs/

# 3. 没有 SPEC 则重新生成
/alfred:1-plan "功能说明"

# 或修改测试的 TAG
# 在 tests/test_hello.py 中修改为 @TEST:EX-HELLO-002

# 4. 同步
/alfred:3-sync
```

---

### 8️⃣ 常用调试命令

**系统状态确认**：

```bash
moai-adk doctor
```

**输出**：所有依赖检查 + 建议

**项目结构确认**：

```bash
tree -L 2 .moai/
```

**TAG 链完整性验证**：

```bash
rg '@(SPEC|TEST|CODE|DOC):' -n | wc -l
```

**输出**：总 TAG 数量

**Git 状态确认**：

```bash
git status
git log --oneline -5
```

---

### 💡 常用调试顺序

1. **阅读**：完整阅读并复制错误消息
2. **搜索**：用错误消息搜索 GitHub Issues
3. **验证**：运行 `moai-adk doctor`
4. **重启**：重启 Claude Code
5. **提问**：在 GitHub Discussions 提问

```bash
# 快速诊断（详细信息）
moai-adk doctor --verbose
```

---

### 🆘 仍然无法解决？

- **GitHub Issues**：搜索是否有类似问题
- **GitHub Discussions**：提问
- **Discord 社区**：实时提问

**报告时应包含的信息**：

1. `moai-adk doctor --verbose` 输出
2. 完整错误消息（截图或复制）
3. 复现方法（执行了什么命令？）
4. 操作系统和版本

---

## 常见问题解答（FAQ）

- **Q. 可以安装在现有项目吗？**
  - A. 可以。运行 `moai-adk init .` 只添加 `.moai/` 结构，不修改现有代码。
- **Q. 如何运行测试？**
  - A. 先运行 `/alfred:2-run`，必要时再次运行 `pytest`、`pnpm test` 等语言特定命令。
- **Q. 如何确认文档总是最新？**
  - A. `/alfred:3-sync` 生成同步报告。在 Pull Request 中确认报告。
- **Q. 可以手动进行吗？**
  - A. 可以，但必须保持 SPEC → TEST → CODE → DOC 顺序，并务必留下 TAG。

---

## 最新更新

| 版本        | 主要功能                                                                         | 日期       |
| ----------- | ------------------------------------------------------------------------------- | ---------- |
| **v0.17.0** | 🌍 **多语言 Lint/Format 架构**（Python、JS、TS、Go、Rust、Java、Ruby、PHP）- 自动语言检测 + 非阻塞错误 | 2025-11-04 |
| **v0.16.x** | <span class="material-icons">check_circle</span> 4个 Alfred 命令 100% 命令式指南完成 + Hook 架构稳定化                          | 2025-11-03 |
| **v0.8.2**  | <span class="material-icons">auto_stories</span> EARS 术语更新："Constraints" → "Unwanted Behaviors"（提高清晰度）              | 2025-10-29 |
| **v0.8.1**  | 🔄 命令更改：`/alfred:9-help` → `/alfred:9-feedback` + 用户反馈工作流程改进         | 2025-10-28 |
| **v0.8.0**  | <span class="material-icons">label</span> @DOC TAG 自动生成系统 + SessionStart 版本检查强化                            | 2025-10-27 |
| **v0.7.0**  | 🌍 完整多语言支持系统（英语、韩语、日语、中文、西班牙语）                        | 2025-10-26 |
| **v0.6.3**  | ⚡ 3步更新工作流程：并行操作提升 70-80% 性能                                   | 2025-10-25 |

> 📦 **立即安装**：`uv tool install moai-adk` 或 `pip install moai-adk`

### <span class="material-icons">target</span> v0.17.0 主要功能

#### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span> 多语言 Lint/Format 自动化（11种语言）

现在无论用什么语言编写项目，都会**自动执行 linting 和格式化**。

**支持语言**：
- Python (ruff + mypy)
- JavaScript (eslint + prettier)
- TypeScript (tsc + eslint + prettier)
- Go (golangci-lint + gofmt)
- Rust (clippy + rustfmt)
- Java (checkstyle + spotless)
- Ruby (rubocop)
- PHP (phpstan + php-cs-fixer)
- C# (dotnet)
- Kotlin (ktlint)
- SQL（迁移检测）

**主要特性**：
- <span class="material-icons">check_circle</span> **自动语言检测**：基于项目配置文件（pyproject.toml、package.json、go.mod 等）
- <span class="material-icons">check_circle</span> **非阻塞错误处理**：工具缺失/错误时仍继续开发流程
- <span class="material-icons">check_circle</span> **部署前错误检测**：Write/Edit 后自动运行 linting 检查
- <span class="material-icons">check_circle</span> **自动代码格式化**：文件修改时自动应用格式化

---

## 更多资源

| 目的                      | 资源                                                                        |
| ------------------------- | --------------------------------------------------------------------------- |
| <span class="material-icons">menu_book</span> 多语言 Lint 指南       | `.claude/hooks/alfred/core/MULTILINGUAL_LINTING_GUIDE.md`（完整 API）        |
| <span class="material-icons">auto_stories</span> 多语言安装指南         | `.claude/hooks/alfred/core/INSTALLATION_GUIDE.md`（各语言安装）             |
| 🧪 测试报告              | `.moai/reports/MULTILINGUAL_LINTING_TEST_REPORT.md`（103/103 测试通过）       |
| <span class="material-icons">target</span> 实现摘要              | `.moai/reports/MULTILINGUAL_LINTING_IMPLEMENTATION_SUMMARY.md`               |
| Skills 详细结构          | `.claude/skills/` 目录（74个 Skills）                                        |
| 子代理详细信息            | `.claude/agents/alfred/` 目录（16个代理 + 4个命令）                         |
| 工作流程指南             | `.claude/commands/alfred/`（4个命令：0-project ~ 3-sync）                    |
| Alfred 命令命令式指南     | `.claude/commands/alfred/`（0-project ~ 3-sync，100% 命令式）                |
| 发布说明                 | GitHub Releases: https://github.com/modu-ai/moai-adk/releases               |

---

## 社区 & 支持

| 渠道                      | 链接                                           |
| ------------------------- | ---------------------------------------------- |
| **GitHub Repository**     | https://github.com/modu-ai/moai-adk            |
| **Issues & Discussions**  | https://github.com/modu-ai/moai-adk/issues     |
| **PyPI Package**          | https://pypi.org/project/moai-adk/             |
| **Latest Release**        | https://github.com/modu-ai/moai-adk/releases   |
| **Documentation**         | 参考项目内 `.moai/`、`.claude/`、`docs/`       |

---

## <span class="material-icons">rocket_launch</span> MoAI-ADK 的理念

> **"没有 SPEC 就没有 CODE"**

MoAI-ADK 不是简单的代码生成工具。Alfred 超级代理和 19 人团队、56 个 Claude Skills 共同保证：

- <span class="material-icons">check_circle</span> **规格说明（SPEC）→ 测试（TDD）→ 代码（CODE）→ 文档（DOC）一致性**
- <span class="material-icons">check_circle</span> **@TAG 系统追踪完整历史**
- <span class="material-icons">check_circle</span> **保证覆盖率 87.84% 以上**
- <span class="material-icons">check_circle</span> **4步工作流程（0-project → 1-plan → 2-run → 3-sync）循环开发**
- <span class="material-icons">check_circle</span> **与 AI 协作但保持透明、可追踪的开发文化**

与 Alfred 一起开始**可信 AI 开发**的全新体验！🤖

---

**MoAI-ADK** — SPEC-First TDD with AI SuperAgent & Complete Skills + TAG Guard

- 📦 PyPI: https://pypi.org/project/moai-adk/
- 🏠 GitHub: https://github.com/modu-ai/moai-adk
- <span class="material-icons">description</span> License: MIT
- ⭐ Skills: 73+ 生产就绪指南（多语言 linting 等）
- <span class="material-icons">check_circle</span> Tests: 570+ 通过（89%+ 覆盖率 - v0.17.0 新增 103 个测试）
- <span class="material-icons">label</span> TAG Guard: PreToolUse Hook 中自动 @TAG 验证