---
title: MoAI-ADK 文档
weight: 99
draft: false
---

MoAI-ADK (Agentic Development Kit) 是 Claude Code 的战略编排框架。

> **当前版本:** {{< version >}} — 版本信息来自 `hugo.toml` 中的 `params.version` 单一可信源 (SSOT)。

![MoAI-ADK](/og.jpg)

![文档结构图](/images/sections/doc-map-zh.png)

## v3.2 新功能 —— 工厂模式 {{< new-badge v3.2 >}}

一个会话只有一个上下文窗口，长 SPEC 会把它填满，之后的每一步都背着前面的全部内容前进。工厂模式把工作分到**多个终端**：主导会话负责分配选定的卡片，每条 lane 是各自独立的会话，把一张卡片从 `plan`、`run` 到 `sync` 整张接下来。评审判定不是单独的阶段，由 sync 关卡吸收。上限并没有消失，但没有任何会话需要背负超过一张卡片的历史，因此同样的预算能走得远得多。

每条 lane 都可以用不同的后端。主导会话和各条 lane 可以分别在 `cc`、`glm`、`codex` 中选择。

{{< terminal title="factory mode" raw="true" >}}
moai cc -f                    # 主导：开启工厂运行
moai cc -l                    # lane，各开一个终端：作为下一条空闲 lane 加入
moai glm -l                   # GLM 后端的 lane
moai codex -l                 # Codex 的 lane
{{< /terminal >}}

委派通道是队列。`backlog` 刻意不设归属会话，只有用 [`/moai todo`](/zh/utility-commands/moai-todo) 主动放入时，工作才会进来。主导只依据自己从卡片 `progress.md` 中读到的证据推进卡片，不依据 lane 的回复。

启动 `moai web`，即可在工厂页面并排查看工厂 lane 与 SPEC 流水线。

![moai web 控制台 Overview 页面 —— SPEC 统计、进行中 SPEC 列表、会话注册表](/images/profile/web-console-v31-overview.png)

详见：[工厂模式](/zh/advanced/factory-mode) · [manager-lead 主导协调者](/zh/advanced/manager-lead) · [`/moai todo`](/zh/utility-commands/moai-todo) · [moai web 控制台](/zh/advanced/moai-web-console)

## MoAI 3.1 的三大核心价值

- {{< icon database primary >}} **代币经济学** — 通过上下文瘦身和提示缓存将推理成本降低60-70%。参见[多 LLM](/zh/multi-llm)、[成本优化](/zh/cost-optimization)和[高级/代币经济学概述](/zh/advanced/tokenomics-overview)。

- {{< icon rotate primary >}} **智能体循环工程** — 通过决策记忆和自主智能体系统实现自主改进循环。参见[自进化系统](/zh/advanced/self-evolving)、[自主循环](/zh/advanced/autonomous-loops)和[决策记忆](/zh/advanced/decision-memory)。

- {{< icon package primary >}} **代理型线束** — 通过技能、钩子和 MCP 提供可组合的执行环境，实现可扩展的智能体编排。参见[核心概念](/zh/core-concepts)、[工作流命令](/zh/workflow-commands)和[智能体指南](/zh/advanced/agent-guide)。

## 主要特性

- **MoAI 编排器**: 通过专业代理进行战略任务委派
- **基于 SPEC 的 TDD/DDD**: 自适应方法论 — 新项目用TDD，遗留代码用DDD
- **TRUST 5 框架**: 5条质量原则：已测试、可读、统一、安全、可追踪
- **渐进式披露**: 3级技能加载，减少67%令牌消耗

## 入门指南

要开始使用 MoAI-ADK,请参阅[入门指南](/zh/getting-started)部分。

## 文档结构

- [入门指南](/zh/getting-started) - 安装、基本设置、快速开始
- [核心概念](/zh/core-concepts) - SPEC 格式、代理、工作流
- [高级](/zh/advanced) - 高级模式、技能使用、性能优化
- [Git Worktree](/zh/worktree) - 完整的 Git Worktree CLI 指南
