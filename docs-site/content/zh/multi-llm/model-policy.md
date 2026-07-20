---
title: 模型策略
weight: 30
draft: false
---

## 什么是模型策略？

模型策略是 MoAI-ADK 托克诺米克斯的骨架。它不是"所有任务都用最强模型"，
而是为每个代理 —— 计划、审计等推理繁重的工作与文档化、Git 等轻量工作 ——
声明式地分配合适的模型。它配合 Claude Code 订阅计划最大化质量，同时避免
速率限制错误。

MoAI-ADK v3.0 的代理目录共 **11 个**（MoAI 自定义 10 个 + Anthropic 内置
`Explore`），下面的分配表覆盖其中由模型策略直接分配的核心 7 个代理。

## 3 级策略概览

| 策略 | 计划 | Opus | Sonnet | Haiku | 适合用途 |
|------|------|---------|-----------|----------|-----------|
| **High** | Max $200/月 | 5 | 1 | 1 | 最高质量、最大吞吐量 |
| **Medium** | Max $100/月 | 2 | 3 | 2 | 质量与成本的平衡 |
| **Low** | Plus $20/月 | 0 | 4 | 3 | 低预算、不含 Opus |

> **为什么重要？** Plus $20 计划无法访问 Opus。设置 `Low` 策略后，所有代理只使用 Sonnet 和 Haiku，从而避免速率限制错误。更高级别的计划为核心代理（计划、审计）分配 Opus，日常任务使用 Sonnet/Haiku。

## 各代理的模型分配表

### Manager Agents（4 个）

| 代理 | High | Medium | Low |
|---------|------|--------|-----|
| manager-spec | opus | opus | sonnet |
| manager-develop | opus | sonnet | sonnet |
| manager-docs | sonnet | haiku | haiku |
| manager-git | haiku | haiku | haiku |

### Evaluator & Builder Agents（3 个）

| 代理 | High | Medium | Low |
|---------|------|--------|-----|
| plan-auditor | opus | opus | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| builder-harness | opus | sonnet | haiku |

> Anthropic 内置的 `Explore` 是只读探索代理，无需单独分配即可运行。
> Agent Teams 静态层（静态 role profile）在 v3.0 中已退役，并行工作由
> sub-agent 并行执行和动态工作流取代。`moai cg` 的 teammate 运行时
> （tmux pane）保持不变。

## 分配原则

- **始终 Opus**：计划审计 (plan-auditor)、SPEC 撰写 (manager-spec) —— 需要高推理能力
- **始终 Haiku**：Git (manager-git) —— 轻量快速的工作
- **随计划变动**：实现 (manager-develop, cycle_type=tdd/ddd) —— 计划越高越倾向 Opus

为了避免"制定计划的代理自己做审计"，plan-auditor 和 sync-auditor 保持
独立分配 —— 这张表同时设计了成本轴与质量轴（偏差防止）。

## v3.0 扩展：Tier×Phase 声明轴

v3.0 在代理级分配之上，新增了**工作阶段 (phase) 与 SPEC 规模 (Tier)** 轴。
`internal/config/model_routing.go` 以声明方式管理 Tier×Phase →
{model, effort} 矩阵：

- **model**: inherit / sonnet / opus / glm / fable
- **effort**（推理深度）: low / medium / high / xhigh / max
- **tier**（SPEC 规模）: S / M / L
- **phase**（工作阶段）: plan / run / sync / mx

每个代理的 model+effort 分配由单一配置矩阵负责。活动配置文件(`profile` ——
`max`/`medium`/`low`)选择矩阵的一列，`profile` 缺失时 legacy `performance_tier`
作为别名读取，再缺失则解释为 `medium`。详细的每个代理映射请参阅
[配置矩阵](/zh/advanced/profile-matrix/)页面。

## 配置方法

### 项目初始化时

```bash
moai init my-project
# 交互式向导中包含模型策略选择
```

### 重置现有项目

```bash
moai update
# 交互式提示:
# - Reset model policy? (y/n) — 重置模型策略
# - Update GLM settings? (y/n) — 配置 GLM 环境变量
```

### 用 CLI 标志直接设置

```bash
moai init my-project --model-policy max     # 最高质量 (以 Opus 为中心)
moai init my-project --model-policy medium  # 平衡 (默认值)
moai init my-project --model-policy low     # 仅 Sonnet, 不使用 Opus
```

`--model-policy` 接受 `max`/`medium`/`low` 三个值,并原样保存到 `llm.yaml` 的
`performance_tier` 字段。已弃用的 `--high` 标志是 `--model-policy max` 的别名。

> 默认策略为 `medium`(llm.yaml `performance_tier: "medium"`,对应 CLI `--model-policy medium` —— 无值时解释为 `medium`)。GLM 配置隔离在 `settings.local.json` 中,不会提交到 Git。

## 下一步

- [CG 模式](/zh/multi-llm/cg-mode) —— 用 Claude + GLM 混合节省成本
- [代理指南](/zh/advanced/agent-guide) —— 代理自定义
- [CLI 参考](/zh/getting-started/cli) —— moai init, moai update 详解
