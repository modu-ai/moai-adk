---
title: Harness 配置档案与评估系统
weight: 75
draft: false
---

如果对所有变更都套用同等深度的验证，就会浪费代币；如果把验证统一压到浅层，质量又会漏掉。MoAI-ADK 的答案是 **自适应验证** — 根据 SPEC 的复杂度自动调节验证深度，并把评估交给独立评估者而非实现方。

## 概述

MoAI-ADK 的 Harness 是一套 **3 层自适应质量验证系统**。它根据 SPEC 的复杂度自动调节验证深度，并由 sync-auditor 智能体以 4 维评分执行独立、持怀疑立场的质量评估。完成与否不是靠"感觉差不多了"，而是靠分数与依据来判定。

## 3 层 Harness 级别

| 级别 | 说明 | 适用时机 | sync-auditor |
|------|------|----------|-----------------|
| **minimal** | 快速验证 | 简单变更（typo、配置修改） | 可省略 |
| **standard** | 基础质量验证 | 大多数任务 | 可选 |
| **thorough** | 完整验证 + TRUST 5 | 复杂 SPEC、大规模变更 | 必需 |

Harness 级别由 **复杂度估算器** (Complexity Estimator) 基于 SPEC scope 自动决定。不对 typo 修复跑 thorough 验证 — 这本身就是代币经济学。

## 4 维评分

sync-auditor 从 4 个维度打分。

| 维度 | 说明 | 默认 Must-Pass |
|------|------|---------------|
| **Functionality** | 功能完成度 — 是否达成了预期目的 | 是 |
| **Security** | 安全 — OWASP、认证、权限、输入校验 | 是 |
| **Craft** | 代码质量 — 可读性、结构、测试覆盖率 | 否 |
| **Consistency** | 一致性 — 项目规则、代码风格遵循 | 否 |

### 分数范围

每个维度得到 0.0 ~ 1.0 的分数。

### 评分锚点

为了避免分数随评估者的"心情"波动，所有评估标准都带有 4 级评分锚点 (rubric anchor)。

| 分数 | 水平 | 含义 |
|------|------|------|
| 0.25 | 未达标 | 未满足基本要求 |
| 0.50 | 部分 | 部分满足，需要改进 |
| 0.75 | 达标 | 大部分满足，小幅改进即可 |
| 1.00 | 优秀 | 完美满足所有标准 |

## 评估配置档案

`.moai/config/evaluator-profiles/` 提供 4 个配置档案。可以根据任务性质切换评估标准的严格程度。

| 档案 | 说明 | 适用场景 |
|--------|------|------------|
| `default.md` | 均衡的默认档案 | 大多数任务 |
| `strict.md` | 严格标准 | 安全关键任务 |
| `lenient.md` | 宽松标准 | 原型开发 |
| `frontend.md` | 前端特化 | UI/UX 任务 |

## 评估者偏差防范（5 种机制）

LLM 评估者若放任不管就会趋于宽容。为了从结构上抑制这一点，5 种机制协同工作。

| # | 机制 | 说明 |
|---|---------|------|
| 1 | **锚点约束** | 打分必须附带对应锚点的论证 |
| 2 | **回归基线** | 检测相对以往项目的异常分数上升 |
| 3 | **Must-Pass 防火墙** | 必过标准不能用其他维度的分数来补偿 |
| 4 | **独立复评** | 每第 5 次进行独立复评（偏差 > 0.10 时重新校准） |
| 5 | **反模式交叉检查** | 发现已知反模式时，对应维度分数上限 0.50 |

## Evaluator Memory Scope

评估者的判断记忆是 **按迭代临时存在** 的。在 GAN Loop 的每次迭代中，sync-auditor 都以全新上下文重启，上一轮迭代的判断依据不会进入新提示词。只有 Sprint Contract 状态在迭代之间保留。这一设计是为了防止评估者锚定在自己之前的判断上、惯性打分。

## 配置

在 `.moai/config/sections/harness.yaml` 中配置。

```yaml
harness:
  level: auto              # auto | minimal | standard | thorough
  evaluator:
    memory_scope: per_iteration   # FROZEN — 不可修改
    profiles:
      default: .moai/config/evaluator-profiles/default.md
      strict: .moai/config/evaluator-profiles/strict.md
    aggregation: min              # min | mean
    must_pass_dimensions:
      - Functionality
      - Security
```

## 相关文档

- [Harness 工程](/zh/core-concepts/harness-engineering) — Harness 概念概览
- [TRUST 5 质量](/zh/core-concepts/trust-5) — 5 项质量标准
- [Constitution 系统](/zh/core-concepts/constitution) — FROZEN/Evolvable 规则
- GAN Loop — 设计质量验证迭代（GAN Loop 是一种对抗式评估者-判别者循环，用于以迭代验证驱动质量改进的模式）
