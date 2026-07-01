---
title: 决策记忆系统
weight: 50
draft: false
---

# 决策记忆系统

指导 MoAI 用户偏好学习和自适应推荐系统。

{{< callout type="info" >}}
**一行总结**: 决策记忆记住用户的选择，并在将来类似的情况下提供个性化推荐。
{{< /callout >}}

## 系统概述

决策记忆(Decision Memory)是 MoAI-ADK 的**长期学习层**。在 AskUserQuestion 轮次中观察用户选择，并基于未来相同决策点的统计多数选择提供自适应推荐。

### 核心原则

| 原则 | 说明 |
|------|------|
| **观察基础** | 学习用户选择的统计多数（非政策默认值） |
| **透明度** | 始终明示推荐依据（包括冷启动状态） |
| **自主性** | 用户可随时拒绝推荐 |
| **自适应强度** | 根据熟练度自动调整推荐强度 |

## 5 个组成部分

### 1. 3-Tier Memory Layer（内存层）

决策记忆由 3 个层组成。

#### L0: Immediate（即时记忆）
- **范围**: 当前会话内
- **用途**: 参考用户刚选择的选项
- **持久性**: 会话结束时消失

#### L1: Session Span（会话范围记忆）
- **范围**: 同一项目的最近 3 个会话
- **用途**: 基于最近偏好的推荐
- **持久性**: `.claude/projects/{hash}/memory/` 自动记忆

#### L2: Long-term（长期记忆）
- **范围**: 所有会话（无限制）
- **用途**: 统计多数学习、长期趋势
- **持久性**: MEMORY.md + topic 文件（用户管理）

### 2. Adaptive Recommendation Placement（自适应推荐配置）

推荐（第一个选项的`（推荐）`标签）基于**观察到的统计多数**。

#### 冷启动（初始状态）
- **观察 < N**: 数据不足
- **推荐配置**: 静态默认值（明确公开）
- **显示方式**: `based on static default, N observations needed for personalization`

#### 热状态（学习中）
- **观察 = N~M**: 部分学习
- **推荐配置**: 观察到的多数 + 信心信号
- **信心度**: 观察数 × 选择一致性

#### 成熟状态（稳定化）
- **观察 > M**: 充分学习
- **推荐配置**: 强多数信心（统计显著）
- **信心度**: 最高（≥95% 信心度）

#### 熟练度基础自适应强度
- **专家（会话 > 50）**: 弱推荐强度（自主性优先，仅推断偏好公开）
- **初学者（会话 < 10）**: 强推荐强度（`（推荐）`标签 + 理由明确）
- **中级（10 ≤ 会话 ≤ 50）**: 中等强度（视情况调整）

### 3. PostToolUse Capture Hook（决策捕捉）

当 AskUserQuestion 响应到达时，PostToolUse 钩子自动捕捉决策。

#### 捕捉数据

```json
{
  "decision_id": "moai-ask-001",
  "timestamp": "2026-07-01T10:00:00Z",
  "question": "请选择下一步",
  "user_choice": "选项 A（推荐）",
  "all_options": ["选项 A", "选项 B", "选项 C"],
  "context": {
    "spec_id": "SPEC-XXX-001",
    "phase": "run",
    "workflow": "/moai run"
  }
}
```

#### 保存位置

- **会话期间**: `.moai/state/decisions/`（临时 JSON）
- **会话结束**: `~/.claude/projects/{hash}/memory/decisions.jsonl`（自动记忆）

### 4. Decay Policy（衰减策略）

逐步降低旧决策的权重。

#### 衰减函数

```
weight(t) = initial_weight × exp(-decay_rate × days_ago)
```

#### 默认值
- **初始权重**: 1.0
- **衰减率**: 0.1（每 7 天约 50% 衰减）
- **保留期**: 90 天（之后自动存档）

#### 示例

```
昨天选择: weight = 0.95
7 天前选择: weight = 0.50
30 天前选择: weight = 0.04
90 天以上: 存档（排除推荐反映）
```

### 5. Recovery Controls（恢复控制）

管理决策记忆的错误恢复和重置。

#### 内存初始化

用户可以重置学习的偏好:

```bash
/moai memory reset
```

#### 偏好编辑

修改特定决策类别的推荐:

```bash
/moai memory set <category> <preferred-option>
```

#### 偏好查询

检查当前学习的偏好:

```bash
/moai memory list
```

## 决策类别

内存追踪的主要决策类型:

| 类别 | 示例 |
|------|------|
| **Tier Selection** | Tier S/M/L 选择 |
| **Cycle Type** | DDD vs TDD 模式 |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | 直接提交 vs 基于 PR |
| **Team Mode** | Solo vs Agent Teams |
| **Model Selection** | 按任务选择模型 |
| **Effort Level** | Effort 级别（low/medium/high/xhigh） |

## 统计多数学习示例

### 场景 1: Tier Selection

如果用户进行了 10 次 Tier 选择:

```
Tier S: 3 次选择
Tier M: 6 次选择  ← 统计多数（60%）
Tier L: 1 次选择

学习结果: Tier M 显示为（推荐）
信心度: 中等（6/10 = 60%, N=10）
推荐显示: "Tier M（推荐）— 基于最近选择 60%"
```

### 场景 2: Cycle Type

```
DDD: 4 次
TDD: 5 次选择  ← 统计多数
其他: 1 次

学习结果: TDD 显示为（推荐）
信心度: 中等（5/10 = 50%, N=10）
推荐显示: "TDD（推荐）— 基于观察"
```

## 冷启动透明度

在观察不足时明确公开:

```
选项 1: Tier M（推荐）— based on static default, 5 observations needed for personalization
选项 2: Tier L
选项 3: Tier S
```

用户清楚了解处于学习阶段。

## 熟练度基础强度调整示例

### 初学用户（会话 < 10）
```
Tier M（推荐）— 基于最近选择提示
（强推荐强度）
```

### 专家用户（会话 > 50）
```
选项:
- Tier M（最近选择 60%）
- Tier L
- Tier S
（弱推荐强度，仅推断偏好公开）
```

## 相关文档

- [AskUserQuestion 协议](/advanced/agent-guide) - 推荐配置规则（HARD）
- [工作流选择](/advanced/harness-v4-builder) - Tier 选择和决策
- [内存系统](/getting-started/memory) - 用户偏好管理

{{< callout type="info" >}}
**提示**: 决策记忆自动运行。不需要明确配置。用户在做出决策时会自动学习。
{{< /callout >}}
