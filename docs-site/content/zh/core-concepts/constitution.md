---
title: Constitution 系统
weight: 35
draft: false
---

这是管理 MoAI-ADK 不变规则 (FROZEN) 与可演化规则 (Evolvable) 的宪法式约束系统。

## 概述

正如在[挽具工程](/zh/core-concepts/harness-engineering)中所见，MoAI-ADK 的挽具会依据循环积累的观察自行演化指令。那么，是什么在管控这种演化？答案就是 **Constitution (宪法)** 系统。

Constitution 区分了 AI 智能体不得随意更改的不变约束 (FROZEN Zone) 与可通过学习改进的可演化约束 (Evolvable Zone)。把评估标准和安全规则放在演化循环**之外** — 这正是自我演化挽具不会失控的原因，也是挽具工程的核心安全机制。

## FROZEN vs Evolvable

### FROZEN Zone (不变)

AI 智能体绝对不能修改的规则，只有人类开发者可以更改。

**代表条目**：

| 条目 | 说明 | 来源 |
|------|------|------|
| TRUST 5 | 5 项质量标准 | moai-constitution.md |
| SPEC + EARS | 规格书格式 | spec-workflow.md |
| AskUserQuestion 独占 | 用户提问通道 | agent-common-protocol.md |
| 4 个评估维度 | Functionality/Security/Craft/Consistency | harness/scorer.go |
| 评分锚点 4 级 | 0.25/0.50/0.75/1.00 | harness/rubric.go |
| 通过阈值下限 | 最低 0.60（不可下调） | design-constitution.md |
| 设计流水线顺序 | manager-spec 最先，sync-auditor 最后 | design-constitution.md |

### Evolvable Zone (可演化)

可以通过学习 (lessons) 与研究 (research) 提出改进建议的规则。

**代表条目**：

| 条目 | 说明 |
|------|------|
| 技能正文内容 | moai-domain-* 技能的细节内容 |
| 流水线权重 | design.yaml 的 phase_weights |
| 迭代上限 | design.yaml 的 iteration_limits |
| 智能体行为规则 | Surface Assumptions、Enforce Simplicity 等 |

## Zone Registry

枚举所有 HARD 条款的**单一事实来源**(Single Source of Truth)。

### ID 分配规则

```
CONST-V3R2-NNN (3자리 이상 zero-padding)

001-050: 기존 HARD 조항
051-099: design constitution 미러 엔트리
100-149: design overflow (자동 확장)
150+: 신규 추가
```

### Canary Gate

FROZEN 条款带有 `canary_gate: true`。修改前必须通过 canary 验证。

```yaml
# Zone Registry 엔트리 예시
- id: CONST-V3R2-154
  zone: Frozen
  file: internal/harness/scorer.go
  anchor: "#dimension-enum"
  clause: "Dimension enum FROZEN at 4 values"
  canary_gate: true
```

## 安全架构（5 层）

Constitution 系统受 5 层安全架构保护。无论挽具积累了多少学习，任何变更都必须依次通过以下五道关口：

### Layer 1: Frozen Guard

在写入操作前确认目标文件不在 FROZEN zone 内。违规时阻断写入 + 记录日志 + 通知用户。

### Layer 2: Canary Check

将提议的变更应用到内存中，并重新评估最近 3 个项目。若得分下降超过 0.10 则拒绝变更。

### Layer 3: Contradiction Detector

当新学习与既有规则冲突时，向用户同时呈现双方。绝不会发生自动覆盖。

### Layer 4: Rate Limiter

限制演化速度：

| 参数 | 默认值 | 说明 |
|-----------|--------|------|
| `max_evolution_rate_per_week` | 3 | 每周最大演化次数 |
| `cooldown_hours` | 24 | 演化间最短等待时间 |
| `max_active_learnings` | 50 | 活跃学习条目最大数 |

### Layer 5: Human Oversight

当 `require_approval: true` 时，所有演化提案都需要用户批准。

## 在 CLI 中使用

```bash
# 전체 registry 조회
moai constitution list

# Frozen zone 필터
moai constitution list --zone frozen

# 특정 파일 조항만 조회
moai constitution list --file internal/harness/scorer.go

# JSON 형식 출력
moai constitution list --format json
```

## 相关文档

- [TRUST 5 质量](/zh/core-concepts/trust-5) — 5 项质量标准
- [挽具工程](/zh/core-concepts/harness-engineering) — 挽具概念概览
- [基于 SPEC 的开发](/zh/core-concepts/spec-based-dev) — SPEC 工作流
