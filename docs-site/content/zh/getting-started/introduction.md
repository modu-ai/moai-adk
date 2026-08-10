---
title: 介绍
weight: 20
draft: false
---

MoAI-ADK 是用 **成本**（代币经济学）· **自我改进**（智能体循环工程）· **品质把控**（智能体 Harness）三大核心价值把 Claude Code 包起来的 Agentic Development Kit。同等质量的代码用更少的 token 产出，声明完成条件后循环自行工作并把积累的观察作为 harness 学习的原料，SPEC 3-phase 与 TRUST 5 门禁用证据判定「完成」—— 模型选择、推理深度、上下文用量都由系统管理。它是用 Go 编写的单一二进制，无需依赖即可运行。


## 表记法说明

本文中代码块的前缀表示执行环境:

- 在 **Claude Code** 对话窗输入的命令
  ```bash
  > /moai plan "功能描述"
  ```

- 在 **终端**(Terminal)输入的命令
  ```bash
  moai init my-project
  ```

## 三大核心价值

MoAI-ADK v3.0 的价值可归纳为三大核心价值。

| 核心价值 | 一句话说明 | 代表工具 |
|------|-----------|----------|
| **代币经济学** | 最大化性价比的智能资源分配 | 3 层模型策略 · CG 模式 · Token Circuit Breaker |
| **智能体循环工程** | 循环自行工作,观察积累使 harness 学习 | `/moai goal` · `/moai loop` · Analyze-First 路由 |
| **智能体 harness** | 不亲自写代码,而是设计智能体工作的环境 | 11 个智能体 · SPEC 3-phase · TRUST 5 |

各核心价值的详细内容在[核心概念](/zh/core-concepts/)一节讲解。本文只看开始所需的部分。

## 核心概念

MoAI-ADK 以 **基于 SPEC 的 TDD/DDD** 方法论为基础,并通过 **TRUST 5** 质量框架保障代码质量。

### 什么是 SPEC?(轻松理解)

**SPEC**(Specification)就是"把与 AI 的对话作为文档留存下来"。

**氛围编程**(Vibe Coding)最大的问题是 **上下文丢失**:

- 与 AI 讨论 1 小时的内容,会话一断就 **消失**
- 第二天想接着工作,得 **从头重新说明**
- 功能越复杂,越会出现 **与意图不同的结果**

**SPEC 解决这个问题:**

- 将需求 **保存为文件** 永久保留
- 即使会话中断,只要读 SPEC 就能 **接着工作**
- 用 EARS 格式 **无歧义地** 明确定义
- 无需重复相同说明,因此 **也节省 token**

{{< callout type="info" >}}
**一句话概括:** 昨天与 AI 讨论的"JWT 认证 + 1 小时过期 + 刷新令牌",今天无需重新说明,用 `/moai run SPEC-AUTH-001` 一行就能立即开始实现!
{{< /callout >}}

### 方法论与质量标准

实现方式会依项目状态自动指派为两者之一,产出则用共同的质量标准来验证。

| 名称 | 何时适用 | 详细 |
|------|----------|--------|
| **TDD** (Test-Driven Development) | 新项目或测试覆盖率 10% 以上(默认) | [SPEC 驱动开发](/zh/core-concepts/spec-based-dev) |
| **DDD** (Domain-Driven Development) | 测试覆盖率不足 10% 的既有项目 | [DDD](/zh/core-concepts/ddd) |
| **TRUST 5** | 与方法论无关,适用于所有代码变更 | [TRUST 5](/zh/core-concepts/trust-5) |

{{< callout type="info" >}}
MoAI-ADK v2.5.0+ 采用二元方法论选择(仅 TDD 或 DDD)。为明确性与一致性,hybrid 模式已被移除。方法论在 `moai init` 时自动选择,可在 `.moai/config/sections/quality.yaml` 的 `development_mode` 中更改。
{{< /callout >}}

## Go Edition 特点

MoAI-ADK 将 Python Edition 完全用 Go 重写,以最大化性能与效率。

| 项目 | Python Edition | Go Edition |
|------|---------------|------------|
| 分发 | pip + venv + 依赖 | **单一二进制**,无依赖 |
| 启动时间 | ~800ms 解释器启动 | **~5ms** 原生执行 |
| 并发 | asyncio / threading | **原生 goroutines** |
| 类型安全 | 运行时(mypy 可选) | **编译期强制** |
| 跨平台 | 需要 Python 运行时 | **预构建二进制**(macOS, Linux, Windows) |

### 核心数字(以 v3.0 为准)

- **11 个** 智能体目录(10 个 MoAI 自定义 + 1 个 Anthropic 内置 `Explore`)
- **31 个** 技能(template-managed)
- **36 个** CLI 命令 · **16 种** `/moai` 子命令
- **16 种** 编程语言支持
- 基于 **543 个** SPEC 文档开发的代码库

## 系统要求

| 平台 | 支持环境 | 备注 |
|--------|----------|------|
| macOS | Terminal, iTerm2 | 完全支持 |
| Linux | Bash, Zsh | 完全支持 |
| Windows | **WSL(推荐)**, PowerShell 7.x+ | 不支持原生 cmd.exe |

**必要条件:**
- 所有平台都必须安装 **Git**
- **Windows 用户**:为获得最佳体验,推荐使用 WSL(Windows Subsystem for Linux)

## 主要功能

### 智能体目录(11 个)

MoAI 编排器不亲自实现,而是把工作委派给 11 个专业智能体。计划与审计是分离的 —— 制作者不检查。

| 类别 | 数量 | 主要智能体 |
|----------|------|--------------|
| **Manager** | 5 个 | manager-spec, manager-develop, manager-docs, manager-git, manager-design |
| **Evaluator** | 2 个 | plan-auditor, sync-auditor |
| **Builder** | 1 个 | builder-harness |
| **Advisor** | 1 个 | super-advisor(高推理咨询) |
| **Specialist** | 1 个 | e2e-tester(执行 Web/移动/桌面 E2E 测试) |
| **内置** | 1 个 | Explore(Anthropic 内置,只读代码分析) |

### 模型策略(代币经济学)

MoAI-ADK 依 Claude Code 订阅套餐为智能体分配最优 AI 模型。在套餐的用量限制内最大化质量 —— 为计划·审计这类推理繁重的阶段分配上位模型,为重复作业分配轻量模型。

| 层级 | 特点 |
|------|------|
| **high** | 最高质量 —— 对调用频率最低的两个智能体使用 `max` 推理深度 |
| **medium**(默认) | 质量与成本的平衡 |
| **low** | 每任务成本最低 —— agentic 智能体降到 Opus `low` effort,Sonnet 仅用于单次调用的行 |

{{< callout type="info" >}}
默认层级是 **medium**。层级调整的是每个智能体在 Opus 推理深度阶梯上的位置,而不是换成更弱的模型级别 —— `low` 让所有 agentic 行保持 Opus 并使用 `low` effort,仅在单次调用的行上回退到 Sonnet;`high` 则把调用频率最低的两个智能体提升到 `max` effort。通过 `--model-policy` 标志或初始化向导设置。
{{< /callout >}}

### 执行模式与编排

自然语言请求会经过 **Analyze-First** 路由 —— 无论用哪种语言请求,都先分析意图再连接到合适的工作流。编排器会依作业复杂度在顺序子智能体(默认)、并行子智能体扇出、动态工作流之间选择其一。

```bash
/moai run SPEC-AUTH-001           # 基于复杂度自动选择
/moai run SPEC-AUTH-001 --solo    # 强制顺序子智能体
```

{{< callout type="info" >}}
**v3.0 变更**:过去的 Agent Teams 静态编排层已退役。即使强制 `--team` 也会回退到子智能体模式。Claude Code 的原生 teammate 运行时 —— `moai cg` 的 tmux 分割窗口 —— 原样保留。
{{< /callout >}}

### SPEC-First 工作流

MoAI-ADK 遵循 3 阶段开发工作流。Run 阶段的方法论依项目状态自动选择:

```mermaid
flowchart TD
    A["Phase 1: SPEC<br/>/moai plan"] -->|"用 EARS 格式定义需求"| B{"方法论选择"}
    B -->|"新项目 (TDD)"| C["Phase 2: TDD<br/>/moai run"]
    B -->|"既有项目 (DDD)"| D["Phase 2: DDD<br/>/moai run"]
    C -->|"RED → GREEN → REFACTOR"| E["Phase 3: Docs<br/>/moai sync"]
    D -->|"ANALYZE → PRESERVE → IMPROVE"| E
    E -->|"文档化与部署"| F["完成"]

    style C fill:#4CAF50,color:#fff
    style D fill:#2196F3,color:#fff
```

### 智能体循环

声明完成条件后,循环会自行工作:

```text
/moai goal "直到所有测试通过且 lint 干净"     # 条件声明型循环
/moai loop                                    # 基于诊断的反复修复(loop_prevention 默认 100 次)
/moai fix                                     # 单次 pass 自动修复
```

`/moai loop` 是 goal 引擎之上的预设 —— 反复修复直到清空诊断工具找到的问题队列。

迭代上限由分管不同层级的两个设置各自决定。`workflow.loop_prevention.max_iterations`(默认 **100**)是单个操作的诊断修复循环上限,`workflow.agentic_loop.max_iterations`(默认 **10**)是流水线整体的完成循环上限。两者是彼此独立的设置,取值不同并不矛盾。

### 推荐的工作流链

**新功能开发:**
```
/moai plan → /moai run SPEC-XXX → /moai sync SPEC-XXX
```

**Bug 修复:**
```
/moai fix (或 /moai loop) → /moai review → /moai sync
```

**重构:**
```
/moai plan → /moai clean → /moai run SPEC-XXX → /moai review → /moai codemaps
```

**文档更新:**
```
/moai codemaps → /moai sync
```

## 多语言支持

MoAI-ADK 支持以下 4 种语言:

- **韩语**(Korean)
- **英语**(English)
- **日语**(Japanese)
- **中文**(Chinese)

在安装向导中选择偏好语言,或在配置文件中直接更改。

## LSP 集成

**LSP**(Language Server Protocol)是代码编辑器与语言工具之间的标准通信协议。它实时检测代码错误、类型错误、lint 结果以提供即时反馈。

**Ralph-Loop Style** 是把 LSP 诊断结果用作反馈循环的自主工作流。检测到质量问题时自动调用修复智能体,反复直到达成质量标准。

MoAI-ADK 通过 Ralph-Loop Style LSP 集成提供自主工作流:

- **基于 LSP 的完成自动检测**:实时监控代码质量状态
- **实时回归探测**:立即检测变更对既有功能的影响
- **自动完成条件**:达成 0 错误、0 类型错误、85% 覆盖率时自动判定完成

{{< callout type="info" >}}
Ralph-Loop Style LSP 集成使开发工作流的质量门禁自动化,让你无需手动介入也能保持高代码质量。
{{< /callout >}}

## CG 模式节省 token(50~70%)

{{< callout type="info" >}}
**成本（代币经济学）的实战工具:** z.ai GLM 是与 Claude Code 完全兼容的 AI 后端。在 **CG 模式**（`moai cg`,需 tmux）下,Claude 领导承担编排·架构决策·代码评审,GLM 队友并行处理实现·测试·文档化,实现为主的作业可 **节省 50~70% token**。架构设计或安全评审这类需要深度推理的场合则使用 Claude 专用（`moai cc`）。

```bash
moai cc            # Claude 专用
moai glm           # GLM 专用
moai cg            # CG 混合(Claude 领导 + GLM 队友,需 tmux)
```

若没有 GLM 账户,请前往 [z.ai 注册(额外 10% 折扣)](https://z.ai/subscribe?ic=1NDV03BGWU)注册。通过注册链接获得的奖励将用于 **MoAI 开源开发**。详细架构与模型策略请参阅[多 LLM](/zh/multi-llm/)一节。
{{< /callout >}}

## 自我改进 —— 循环自行工作,harness 从中学习

{{< callout type="info" >}}
**自我改进（智能体循环工程）的实战工具:** 声明完成条件后,会话会自行工作直到条件满足。`/moai goal "<条件>"` 是条件声明型自主循环;`/moai loop` 反复修改直到清空 LSP 诊断·AST-grep·linter 找到的问题队列（流水线完成循环默认 10 次 — `agentic_loop.max_iterations`）;`/moai fix` 是单次 pass 自动修复。循环留下的观察（用户纠正、失败模式、路由决策）沿 4 层学习阶梯（观察 → 启发式 → 规则 → 自动更新,在用户批准门禁之下）沉淀为 harness 指令 —— 下一轮会话不会重复上一轮的失误。
{{< /callout >}}

## 开始

要开始 MoAI-ADK 之旅,请遵循以下步骤:

1. **[安装](/zh/getting-started/installation)** - 在系统中安装 MoAI-ADK
2. **[初始设置](/zh/getting-started/init-wizard)** - 运行交互式设置向导
3. **[快速开始](/zh/getting-started/quickstart)** - 创建第一个项目
4. **[核心概念](/zh/core-concepts/what-is-moai-adk)** - 深入理解 MoAI-ADK

## 核心优势

| 优势 | 说明 |
|------|------|
| **质量保障** | 用 TRUST 5 框架保持一致的质量 |
| **token 效率** | 用模型策略 + CG 模式 + Token Circuit Breaker 让系统管理成本 |
| **生产力提升** | 用 AI 智能体自动化缩短开发时间 |
| **可扩展** | 用模块化架构与 harness 构建器灵活扩展 |
| **多语言** | 支持 4 种语言 |

## 附加资源

- [GitHub 仓库](https://github.com/modu-ai/moai-adk)
- [文档站点](https://adk.mo.ai.kr)
- [社区论坛](https://github.com/modu-ai/moai-adk/discussions)

---

## 下一步

在[安装指南](./installation)中了解 MoAI-ADK 的安装方法。
