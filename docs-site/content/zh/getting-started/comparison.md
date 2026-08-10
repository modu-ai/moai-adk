---
title: 有什么不同 — 三种使用形式比较
weight: 35
draft: false
---
# 有什么不同 — 三种使用形式比较

将 Claude Code 用于代理性开发主要有三种方式。(1)直接使用 Claude Code,(2)用通用包装器(wrapper)包裹,(3)像 MoAI-ADK 这样用代理性 Harness 包裹。这个页面一 目了然地整理三种形式在哪里相同,在哪里分化。评价标准是 README 强调的三个核心 — 成本(tokenomics)、自我改进(循环工程)、质量控制(代理性 harness)。

## 三种使用形式

| | Claude Code 单独 | 通用包装器 | MoAI-ADK (代理性 Harness) |
|---|---|---|---|
| **本质** | 官方模型·工具本身 | 包裹模型调用的薄层 | 包裹模型的执行环境(harness) |
| **分发单位** | 官方安装 | 各工具各自不同 | 单个 Go 二进制(不需要 Python·运行时) |
| **成本控制** | 模型自己用多少算多少 | 通常没有 | 每个任务分配模型·推理深度 + 预算守卫 |
| **质量关卡** | 用户每次确认 | 各工具各自不同 | SPEC 3-阶段 + TRUST 5,自动验证 |
| **学习循环** | 每个会话重新开始 | 通常没有 | 将观察累积成规则的自我进化 |
| **会话连续性** | 每次 `/clear` 就断开 | 各工具各自不同 | paste-ready 历史 + 自动注入 |
| **代理编组** | 单个会话 | 单个会话 | 12-代理目录 + 3-阶段工作流 |

Claude Code 单独并不坏。相反,MoAI-ADK 不替代 Claude Code,而是**包裹(wrap)**它并在上面添加结构。模型路由和质量关卡、成本控制、学习循环、会话连续性 — 这些 Claude Code 留给用户的部分,由 harness 系统负责。

## 为什么是三个核心

只推一个就会掉进陷阱。README 一起设立三个核心的原因。

{{< callout type="warning" >}}
- **只优化成本**质量就会偷偷崩溃。后续返工和调试循环是最昂贵的令牌支出。
- **只立质量关卡而没有学习循环**每个会话都会重复同样的错误。
- **只跑自律循环而没有成本上限**一次过度运行就会吞掉整个配额。
{{< /callout >}}

三个核心互相支撑。成本因为质量防止返工而经济,质量因为循环抓住什么有效而可强制,循环因为成本守卫在超标前停止而可承担。MoAI-ADK 的所有设计决定都服从这三个之一。

## 成本 — tokenomics

单价 3 年内下降了 98%(Linux Foundation),但同期企业 AI 支出增长了 320%。用量增加覆盖了单价下降。代理完成一个任务要跑几十到几百个步骤,按比例消耗令牌。

{{< icon target >}} **Claude Code 单独 / 通用包装器** — 模型自己定步骤,用户监控成本。单价低但步骤多,账单还是会很大。

{{< icon target primary >}} **MoAI-ADK** — 区分成本的不是单价,而是**分配**。在 DeepSWE 基准中,Opus 5 的最低推理比 Sonnet 5 的最高推理得分更高,每个任务成本是 1/16。重试循环消耗账单,不是令牌单价。所以为每个任务分配合适的模型和推理深度,减少上下文,在预算超标前停止。`moai cg` 的 Claude+GLM 混合模式在实现为主的工作中带来 60-70% 成本节省。

[Tokenomics 概述](/zh/advanced/tokenomics-overview/) 和 [成本优化](/zh/cost-optimization/) 中详细讨论。

## 自我改进 — 代理性循环工程

最便宜的会话是不重复上次会话错误的会话。

{{< icon rotate >}} **Claude Code 单独 / 通用包装器** — 会话结束时观察也一起消失。下次会话每次都从零开始。

{{< icon rotate primary >}} **MoAI-ADK** — 每次执行成为下次执行的材料。记录路由决策和关卡证据,将重复模式变成规则,推送声明的目标(`/moai goal`)直到满足条件。观察到的失败模式作为规则变更建议上报,不会偷偷适用,而是获得批准。不是模型权重,而是**递归改进包裹模型的 harness** 是现实的短期自我改进路径。

[自我进化系统](/zh/advanced/self-evolving/)、[自律连续循环](/zh/advanced/autonomous-loops/)、[决策记忆](/zh/advanced/decision-memory/) 中详细讨论。

## 质量控制 — 代理性 Harness

返工是最大的令牌浪费。已经出去的 bug 再回来,比所有路由优化加起来还贵。

{{< icon package >}} **Claude Code 单独** — "完成了"需要用户每次直接确认。

{{< icon package >}} **通用包装器** — 每个工具的质量标准各自不同,或者根本没有。

{{< icon package primary >}} **MoAI-ADK** — 将"完成"变成*验证过的完成*。SPEC 3-阶段(plan → run → sync)和 TRUST 5 关卡(已测试·可读·统一·安全·可追踪)适用于每次变更。关卡审判的是验证,不是代理。12-代理目录从一开始就分离计划和审计,让编写的人不能为自己的工作打分。[验证主张完整性](/zh/core-concepts/verification-claim-integrity/)规则防止未经观察的"通过"作为空白通过。

[Harness 工程](/zh/core-concepts/harness-engineering/)、[TRUST 5 质量](/zh/core-concepts/trust-5/)、[基于 SPEC 的开发](/zh/core-concepts/spec-based-dev/) 中详细讨论。

## 选择哪种形式

- **Claude Code 单独**就足够时 — 探索性编码、一两个文件的简单修改、短会话中直接监控成本和质量。
- **通用包装器**合适时 — 只想自动化一个特定工作流,不需要更多结构时。
- **需要 MoAI-ADK**时 — 需要系统为每个会话的成本和质量负责,代理并行工作时不互相踩踏,上次会话的学习延续到下次会话时。

三种形式不互斥。MoAI-ADK 不替代 Claude Code,而是包裹它。单个 Go 二进制在 macOS·Linux·Windows 上无额外依赖运行。

## 下一步

- [安装](/zh/getting-started/installation/) — 单个二进制安装
- [MoAI-ADK 是什么?](/zh/core-concepts/what-is-moai-adk/) — 本质和哲学
- [Harness 工程](/zh/core-concepts/harness-engineering/) — 三个核心相遇的地方
