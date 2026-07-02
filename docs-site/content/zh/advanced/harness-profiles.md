---
title: Harness 配置与评估系统
weight: 75
draft: false
---
# Harness 配置与评估系统


通过3级Harness等级和4维度评估配置实现的自适应质量验证系统。

## 概述

MoAI-ADK的Harness（工具链）是一个**3级自适应质量验证系统**。会根据SPEC的复杂度
自动调整验证深度。sync-auditor代理通过4维度评分执行独立且严谨的质量评估。

## 3级Harness等级

| 等级 | 说明 | 适用时机 | sync-auditor |
|------|------|----------|-----------------|
| **minimal** | 快速验证 | 简单变更（typo、配置修改） | 可省略 |
| **standard** | 基本质量验证 | 大多数任务 | 可选 |
| **thorough** | 完整验证 + TRUST 5 | 复杂SPEC、大规模变更 | 必需 |

Harness等级由基于SPEC范围的**复杂度评估器**（Complexity Estimator）自动
决定。

## 4维度评分

sync-auditor按4个维度进行评分：

| 维度 | 说明 | 默认Must-Pass |
|------|------|---------------|
| **Functionality** | 功能完成度 — 是否达成预期目的 | 是 |
| **Security** | 安全性 — OWASP、认证、权限、输入验证 | 是 |
| **Craft** | 代码质量 — 可读性、结构、测试覆盖率 | 否 |
| **Consistency** | 一致性 — 是否遵守项目规则与代码风格 | 否 |

### 分数范围

每个维度的分数范围为0.0~1.0。

### 评分标准锚点

所有评估标准均具有4阶段评分标准锚点：

| 分数 | 等级 | 含义 |
|------|------|------|
| 0.25 | 未达标 | 未满足基本要求 |
| 0.50 | 部分达标 | 部分满足，需要改进 |
| 0.75 | 达标 | 大部分满足，小幅改进即可 |
| 1.00 | 优秀 | 完全满足所有标准 |

## 评估配置

`.moai/config/evaluator-profiles/` 中提供4种配置：

| 配置 | 说明 | 适用场景 |
|--------|------|------------|
| `default.md` | 均衡的默认配置 | 大多数任务 |
| `strict.md` | 严格标准 | 安全关键任务 |
| `lenient.md` | 宽松标准 | 原型开发 |
| `frontend.md` | 前端专用 | UI/UX任务 |

## 评估者偏见防范（5种机制）

为防止评估者过度宽容，共有5种机制在起作用：

| # | 机制 | 说明 |
|---|---------|------|
| 1 | **评分标准锚定** | 评分必须有评分标准依据 |
| 2 | **回归基线** | 检测相对于历史项目的分数过度上升 |
| 3 | **Must-Pass防火墙** | 必需标准不可被其他维度分数弥补 |
| 4 | **独立复审** | 每第5次执行独立复审（偏差 > 0.10时重新校准） |
| 5 | **反模式交叉检查** | 检测到已知反模式时，将该维度分数限制在0.50以内 |

## Evaluator Memory Scope

评估者的判断记忆是**按迭代临时保存**的。在GAN Loop的每次迭代中，sync-auditor都会
以全新上下文重新启动，上一次迭代的判断依据不会包含在新提示中。
只有Sprint Contract状态会在各次迭代之间保留。

## 配置

在 `.moai/config/sections/harness.yaml` 中进行配置：

```yaml
harness:
  level: auto              # auto | minimal | standard | thorough
  evaluator:
    memory_scope: per_iteration   # FROZEN — 不可更改
    profiles:
      default: .moai/config/evaluator-profiles/default.md
      strict: .moai/config/evaluator-profiles/strict.md
    aggregation: min              # min | mean
    must_pass_dimensions:
      - Functionality
      - Security
```

## 相关文档

- [Harness工程](/zh/core-concepts/harness-engineering) — Harness概念概述
- [TRUST 5质量](/zh/core-concepts/trust-5) — 5项质量标准
- [Constitution系统](/zh/core-concepts/constitution) — FROZEN/Evolvable规则
- GAN Loop — 用于设计质量验证的迭代循环（GAN Loop是一种基于对抗性评估者-判别器循环的迭代验证模式，用于提升质量）
