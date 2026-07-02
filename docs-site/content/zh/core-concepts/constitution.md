---
title: Constitution 系统
weight: 35
draft: false
---

用于管理MoAI-ADK不可变规则（FROZEN）与可演化规则（Evolvable）的宪法性约束系统。

## 概述

MoAI-ADK通过 **Constitution（宪法）** 系统，区分AI代理绝不能任意变更的
不可变约束（FROZEN Zone）与可通过学习不断改进的可演化约束（Evolvable Zone）。
这是Harness工程的核心安全机制。

## FROZEN vs Evolvable

### FROZEN Zone（不可变）

AI代理绝对不能修改的规则。仅人类开发者可以变更。

**代表性条目**：

| 条目 | 说明 | 来源 |
|------|------|------|
| TRUST 5 | 5项质量标准 | moai-constitution.md |
| SPEC + EARS | 规格说明书格式 | spec-workflow.md |
| AskUserQuestion独占 | 用户提问渠道 | agent-common-protocol.md |
| 4个评估维度 | Functionality/Security/Craft/Consistency | harness/scorer.go |
| 4阶段评分标准锚点 | 0.25/0.50/0.75/1.00 | harness/rubric.go |
| 通过阈值下限 | 最低0.60（不可下调） | design-constitution.md |
| 设计流水线顺序 | manager-spec先行，sync-auditor最后 | design-constitution.md |

### Evolvable Zone（可演化）

可通过学习（lessons）与研究（research）提出改进建议的规则。

**代表性条目**：

| 条目 | 说明 |
|------|------|
| 技能正文内容 | moai-domain-*技能的具体内容 |
| 流水线权重 | design.yaml的phase_weights |
| 迭代限制 | design.yaml的iteration_limits |
| 代理行为规则 | Surface Assumptions、Enforce Simplicity等 |

## Zone Registry

列举所有HARD条款的**单一事实来源**（Single Source of Truth）。

### ID分配规则

```
CONST-V3R2-NNN（3位以上zero-padding）

001-050: 既有HARD条款
051-099: design constitution镜像条目
100-149: design overflow（自动扩展）
150+: 新增条目
```

### Canary Gate

FROZEN条款具有 `canary_gate: true`。变更前必须进行canary验证。

```yaml
# Zone Registry条目示例
- id: CONST-V3R2-154
  zone: Frozen
  file: internal/harness/scorer.go
  anchor: "#dimension-enum"
  clause: "Dimension enum FROZEN at 4 values"
  canary_gate: true
```

## 安全架构（5层）

Constitution系统由5层安全架构保护：

### Layer 1: Frozen Guard

写入操作前确认目标文件不属于FROZEN zone。违反时阻止写入 + 记录日志 +
通知用户。

### Layer 2: Canary Check

在内存中应用拟议变更并重新评估最近3个项目。若分数下降
超过0.10，则拒绝该变更。

### Layer 3: Contradiction Detector

当新学习内容与既有规则冲突时，会将两者都呈现给用户。绝不会
自动覆盖。

### Layer 4: Rate Limiter

限制演化速度：

| 参数 | 默认值 | 说明 |
|-----------|--------|------|
| `max_evolution_rate_per_week` | 3 | 每周最大演化次数 |
| `cooldown_hours` | 24 | 演化之间的最短等待时间 |
| `max_active_learnings` | 50 | 活跃学习条目的最大数量 |

### Layer 5: Human Oversight

当 `require_approval: true` 时，所有演化提案均需用户批准。

## 在CLI中使用

```bash
# 查询完整registry
moai constitution list

# 筛选Frozen zone
moai constitution list --zone frozen

# 仅查询特定文件的条款
moai constitution list --file internal/harness/scorer.go

# 以JSON格式输出
moai constitution list --format json
```

## 相关文档

- [TRUST 5质量](/zh/core-concepts/trust-5) — 5项质量标准
- [Harness工程](/zh/core-concepts/harness-engineering) — Harness概念概述
- [基于SPEC的开发](/zh/core-concepts/spec-based-dev) — SPEC工作流
