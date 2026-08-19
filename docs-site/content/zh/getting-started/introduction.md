---
title: 介绍
weight: 20
draft: false
---

MoAI-ADK 是用 **成本**（代币经济学）· **自我改进**（智能体循环工程）· **质量把控**（智能体线束）这三件事把 Claude Code 包起来的 Agentic Development Kit。同等质量的代码用更少的代币产出。只要声明完成条件，循环就会自行工作，过程中积累的观察则成为线束学习的原料。"完成"由 SPEC 三阶段与 TRUST 5 闸门以证据判定。模型选择、推理深度、上下文用量都由系统管理。它用 Go 写成单一二进制，无依赖即可直接运行。

本页把 MoAI-ADK 是什么、为什么长成这个样子放在一条线上介绍：三个核心各自回答什么问题，SPEC · TRUST 5 · CG 模式这些术语在其中站在哪里，以及第一次上手该往哪走。安装步骤与第一个项目的运行交给[安装](/zh/getting-started/installation)和[快速开始](/zh/getting-started/quickstart)页面，这里专注于"为什么"。


## 记法说明

本文代码块的前缀表示执行环境：

- 在 **Claude Code** 对话窗输入的命令
  ```bash
  > /moai plan "功能描述"
  ```

- 在**终端** (Terminal) 输入的命令
  ```bash
  $ moai init my-project
  ```

## 三大核心价值

MoAI-ADK 是用**三根轴**包住 Claude Code 的 Agentic Development Kit —— 成本 · 自我改进 · 质量把控。只推一根轴，其余的会塌。只压成本，质量会荒芜；只立质量闸门，同样的失误每个会话重复；只跑自主循环，一次计费就能烧光额度。三根轴互相支撑。

### 成本 —— 代币经济学

同样的质量，更少的代币。决定成本的不是单价而是**模型分配** —— DeepSWE 基准测试中，Opus 最低推理的得分高于 Sonnet 最高推理，成本却是其十六分之一。三档模型策略 · CG 模式 · 提示缓存 · Token Circuit Breaker 让预算由系统管理。

### 自我改进 —— 智能体循环工程

线束越跑越聪明。声明完成条件后循环自行工作（`/moai goal` · `/moai loop`），观察积累成规则，让下一个会话不再重复同样的失误。

### 质量把控 —— 智能体线束

用证据判定"完成"。SPEC 三阶段生命周期 + TRUST 5 闸门 + worktree 隔离挡住返工（最大的代币浪费），并把计划与审计分开，不让制作的人自己检查。

各核心的详细内容在[核心概念](/zh/core-concepts/)一节讲述。

## v3.1 更方便了的地方

- **`/moai goal`** —— 一行声明完成条件，会话自主推进。
- **看板模式** —— 同时运行多个会话。
- **BAS Navigator** —— 自动同步三段代码地图。
- **manager-lead** —— 协调大规模工作：SPEC 内的 Tier L 里程碑扇出，加上看板与工厂领导会话调度。
- **multi-model audit** —— 用多模型交叉验证抓偏差。
- **autonomy tier** —— 调节自主档位，安全地跑。
- **profile matrix** —— 以 12 个智能体 × 3 个配置文件分配模型。

## 核心概念

MoAI-ADK 遵循 **SPEC 基础的 TDD/DDD** 方法论，并以 **TRUST 5** 质量框架保障代码质量。

### 什么是 SPEC？（轻松理解）

**SPEC** (Specification) 就是"把与 AI 的对话作为文档留存下来"。

**氛围编程** (Vibe Coding) 最大的问题是**上下文丢失**：

- 与 AI 讨论 1 小时的内容，会话一断就**消失**
- 第二天想接着工作，得**从头重新说明**
- 功能越复杂，越会出现**与意图不同的结果**

**SPEC 解决这个问题：**

- 把需求**保存为文件**，永久保留
- 即使会话中断，只要读 SPEC 就能**接着工作**
- 用 EARS 格式**无歧义地**明确定义
- 不必重复相同的说明，因此**也省代币**

{{< callout type="info" >}}
**一句话概括**： 昨天与 AI 讨论的"JWT 认证 + 1 小时过期 + 刷新令牌"，今天无需重新说明，用 `/moai run SPEC-AUTH-001` 一行就能立即开始实现!
{{< /callout >}}

### 方法论与质量标准

实现方式依项目状态自动指派为两者之一，产出用共同的质量标准验证。

| 名称 | 何时适用 | 详细 |
|------|-----------|--------|
| **TDD** (Test-Driven Development) | 新项目或测试覆盖率 10% 以上（默认） | [SPEC 基础开发](/zh/core-concepts/spec-based-dev) |
| **DDD** (Domain-Driven Development) | 测试覆盖率不足 10% 的既有项目 | [DDD](/zh/core-concepts/ddd) |
| **TRUST 5** | 与方法论无关，适用于所有代码变更 | [TRUST 5](/zh/core-concepts/trust-5) |

{{< callout type="info" >}}
MoAI-ADK v2.5.0+ 只在 TDD 与 DDD 中选一个。为明确性与一致性，hybrid 模式已移除。方法论在 `moai init` 时自动决定，可在 `.moai/config/sections/quality.yaml` 的 `development_mode` 中更改。
{{< /callout >}}

## Go Edition 特点

MoAI-ADK 把 Python Edition 用 Go 完全重写，把性能与效率拉到最高。

| 项目 | Python Edition | Go Edition |
|------|---------------|------------|
| 分发 | pip + venv + 依赖 | **单一二进制**，无依赖 |
| 启动时间 | ~800ms 解释器启动 | **~5ms** 原生执行 |
| 并发 | asyncio / threading | **原生 goroutines** |
| 类型安全 | 运行时（mypy 可选） | **编译期强制** |
| 跨平台 | 需要 Python 运行时 | **预构建二进制**（macOS, Linux, Windows） |

### 核心数字（以 v3.0 为准）

- **11 个**智能体目录（10 个 MoAI 自定义 + 1 个 Anthropic 内置 `Explore`）
- **31 个**技能 (template-managed)
- **36 个**终端 CLI 命令 · **16 种** `/moai` 斜杠子命令
- **16 种**编程语言支持
- 基于 **543 个** SPEC 文档开发的代码库

## 系统要求

| 平台 | 支持环境 | 备注 |
|--------|----------|------|
| macOS | Terminal, iTerm2 | 完全支持 |
| Linux | Bash, Zsh | 完全支持 |
| Windows | **WSL（推荐）**， PowerShell 7.x+ | 不支持原生 cmd.exe |

**必要条件：**
- 所有平台都必须安装 **Git**
- **Windows 用户**： 为获得最顺畅的体验，推荐使用 WSL (Windows Subsystem for Linux)

## 主要功能

### 智能体目录（11 个）

MoAI 编排器不亲自实现，而是把工作委派给 11 个专业智能体。计划与审计是分开的，制作的人不自己检查。

| 类别 | 数量 | 主要智能体 |
|----------|------|--------------|
| **Manager** | 5 个 | manager-spec, manager-develop, manager-docs, manager-git, manager-design |
| **Evaluator** | 2 个 | plan-auditor, sync-auditor |
| **Builder** | 1 个 | builder-harness |
| **Advisor** | 1 个 | super-advisor（高推理咨询） |
| **Specialist** | 1 个 | e2e-tester（执行 Web/移动/桌面 E2E 测试） |
| **内置** | 1 个 | Explore（Anthropic 内置，只读代码分析） |

### 模型策略（代币经济学）

MoAI-ADK 为每个智能体分配最优的模型与推理深度。目标是在套餐的用量额度内把质量尽量拉高。因此不去换更弱的模型等级，而是在同一个 Opus 内部只调节各智能体的推理深度。因为在长周期的智能体工作里，弱模型会消耗更多步数，每任务成本反而更高。

| 档位 | 特点 |
|------|------|
| **high** | 最高质量 —— 调用频率最低的两个智能体使用 `max` 推理深度 |
| **medium**（默认） | 质量与成本的平衡 |
| **low** | 每任务成本最低 —— 智能体行降到 Opus `low` effort，Sonnet 只用于单发行 |

{{< callout type="info" >}}
默认档位是 **medium**。调整档位也不会换模型等级，只改变各智能体的 Opus 推理深度。`low` 让所有智能体行保持 Opus `low` effort，只在单发行上用 Sonnet；`high` 把调用频率最低的两个智能体提升到 `max` effort。通过 `--model-policy` 标志或初始化向导设置。
{{< /callout >}}

### 执行模式与编排

自然语言请求经过 **Analyze-First** 路由。无论用哪种语言请求，都先分析意图，再接上合适的工作流。编排器依任务复杂度在顺序子智能体（默认）、并行子智能体扇出、动态工作流中选择。

```bash
/moai run SPEC-AUTH-001           # 基于复杂度自动选择
/moai run SPEC-AUTH-001 --solo    # 强制顺序子智能体
```

{{< callout type="info" >}}
**v3.0 变更**： 过去的 Agent Teams 静态编排层已废止。强制 `--team` 也会回退到子智能体模式。Claude Code 的原生 teammate 运行时（`moai cg` 的 tmux 分屏）保留不变。
{{< /callout >}}

### SPEC-First 工作流

MoAI-ADK 遵循三阶段开发工作流。Run 阶段的方法论依项目状态自动选择：

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

声明完成条件后，循环会自行工作：

```text
/moai goal "直到所有测试通过且 lint 干净"     # 条件声明型循环
/moai loop                                    # 基于诊断的反复修复 (loop_prevention 默认 100 次)
/moai fix                                     # 单 pass 自动修复
```

`/moai loop` 是 goal 引擎之上的预设 —— 反复修复，直到清空诊断工具找到的问题队列。

迭代上限由分管不同层面的两个设置各自决定。`workflow.loop_prevention.max_iterations`（默认 **100**）是单个任务的诊断修复循环上限，`workflow.agentic_loop.max_iterations`（默认 **10**）是整条流水线的完成循环上限。两者是彼此独立的设置，取值不同并不矛盾。

### 推荐的工作流链

**新功能开发：**
```
/moai plan → /moai run SPEC-XXX → /moai sync SPEC-XXX
```

**Bug 修复：**
```
/moai fix (或 /moai loop) → /moai review → /moai sync
```

**重构：**
```
/moai plan → /moai clean → /moai run SPEC-XXX → /moai review → /moai codemaps
```

**文档更新：**
```
/moai codemaps → /moai sync
```

## 多语言支持

MoAI-ADK 支持以下 4 种语言：

- **韩语** (Korean)
- **英语** (English)
- **日语** (Japanese)
- **中文** (Chinese)

在安装向导中选择偏好语言，也可以在配置文件中直接更改。

## LSP 集成

**LSP** (Language Server Protocol) 是代码编辑器与语言工具之间的标准通信协议。它实时检测代码错误、类型错误、lint 结果并立即反馈。

**Ralph-Loop Style** 是把 LSP 诊断结果用作反馈循环的自主工作流。检测到质量问题时自动调用修复智能体，反复执行直到达成质量标准。

MoAI-ADK 的 Ralph-Loop Style LSP 集成按如下方式工作：

- **基于 LSP 的完成自动检测**： 实时监控代码质量状态
- **实时回归探测**： 立即检测变更对既有功能的影响
- **自动完成条件**： 达成 0 错误、0 类型错误、85% 覆盖率时自动判定完成

{{< callout type="info" >}}
Ralph-Loop Style LSP 集成把开发工作流的质量闸门自动化，让人不亲手介入也能保持高代码质量。
{{< /callout >}}

## 用 CG 模式省代币（50~70%）

{{< callout type="info" >}}
**成本（代币经济学）的实战工具**： z.ai GLM 是与 Claude Code 完全兼容的 AI 后端。在 **CG 模式**（`moai cg`，需 tmux）下，Claude 领队负责编排 · 架构决策 · 代码审查，GLM 工作者并行处理实现 · 测试 · 文档化，在实现为主的工作上**节省 50~70% 代币**。架构设计或安全审查这类需要深度推理的场合，则使用 Claude 专用（`moai cc`）。

```bash
moai cc            # Claude 专用
moai glm           # GLM 专用
moai cg            # CG 混合 (Claude 领队 + GLM 工作者, 需 tmux)
```

没有 GLM 账户的话，请到 [z.ai 注册（额外 10% 折扣）](https://z.ai/subscribe?ic=1NDV03BGWU)注册。通过注册链接获得的奖励用于 **MoAI 开源开发**。详细架构与模型策略请参阅[多 LLM](/zh/multi-llm/)一节。
{{< /callout >}}

## 自我改进 —— 循环自行工作，线束从中学习

{{< callout type="info" >}}
**自我改进（智能体循环工程）的实战工具**： 声明完成条件后，会话会自行工作直到条件满足。`/moai goal "<条件>"` 是条件声明型自主循环；`/moai loop` 反复修复，直到清空 LSP 诊断 · AST-grep · linter 找到的问题队列（流水线完成循环默认 10 次 —— `agentic_loop.max_iterations`）；`/moai fix` 是单 pass 自动修复。循环留下的观察（用户纠正、失败模式、路由决策）沿 4 层学习阶梯（观察 → 启发式 → 规则 → 自动更新，在用户批准闸门之下）沉淀为线束指令。所以下一个会话不会重复上一个会话的失误。
{{< /callout >}}

## 开始

要开始使用 MoAI-ADK，请按以下顺序进行：

1. **[安装](/zh/getting-started/installation)** - 在系统中安装 MoAI-ADK
2. **[初始设置](/zh/getting-started/init-wizard)** - 运行交互式设置向导
3. **[快速开始](/zh/getting-started/quickstart)** - 创建第一个项目
4. **[核心概念](/zh/core-concepts/what-is-moai-adk)** - 深入理解 MoAI-ADK

## 核心优势

| 优势 | 说明 |
|------|------|
| **质量保障** | 用 TRUST 5 框架保持一致的质量 |
| **代币效率** | 模型策略 + CG 模式 + Token Circuit Breaker 让系统管理成本 |
| **生产力提升** | AI 智能体自动化缩短开发时间 |
| **可扩展** | 模块化架构与线束构建器灵活扩展 |
| **多语言** | 支持 4 种语言 |

## 附加资源

- [GitHub 仓库](https://github.com/modu-ai/moai-adk)
- [文档站点](https://adk.mo.ai.kr)
- [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)

---

## 下一步

在[安装指南](/zh/getting-started/installation)中了解 MoAI-ADK 的安装方法。
