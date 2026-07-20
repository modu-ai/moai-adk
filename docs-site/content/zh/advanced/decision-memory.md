---
title: 决策记忆系统
weight: 50
draft: false
---

智能体循环工程的起点是观察 — 循环每转一圈，观察就会累积，累积的观察成为学习的原料。决策记忆是把这一观察对象从代码扩展到 **用户的选择** 的层。

{{< callout type="info" >}}
**一句话总结**：决策记忆会记住用户的选择，并在今后类似的情境中提供个性化推荐。
{{< /callout >}}

## 系统概述

决策记忆 (Decision Memory) 是 MoAI-ADK 的 **长期学习层**。它在 AskUserQuestion 轮次中观察用户的选择，并在今后同类决策点上基于统计多数的选择提供自适应推荐。

关键在于方向。不是把系统想推的默认值包装成 `(推荐)`，而是让 **用户实际反复选择过的选项** 成为推荐。

### 核心原则

| 原则 | 说明 |
|------|------|
| **基于观察** | 学习用户选择的统计多数（而非策略默认值） |
| **透明性** | 始终注明推荐依据（包括 cold-start 状态） |
| **自主性** | 用户随时可以拒绝推荐 |
| **自适应强度** | 根据熟练度自动调整推荐强度 |

## 4 个组成部分

### 1. 3-Tier Memory Layer（记忆层）

决策记忆由 3 个层级组成。越往下留存越久。

#### L0: Immediate（即时记忆）
- **范围**：当前会话内
- **用途**：引用用户刚刚选择的选项
- **持久性**：会话结束即消失

#### L1: Session Span（会话跨度记忆）
- **范围**：同一项目最近 3 个会话
- **用途**：基于近期偏好的推荐
- **持久性**：`.claude/projects/{hash}/memory/` 自动记忆

#### L2: Long-term（长期记忆）
- **范围**：所有会话（无限制）
- **用途**：统计多数学习、长期趋势
- **持久性**：MEMORY.md + topic 文件（用户管理）

### 2. Adaptive Recommendation Placement（自适应推荐放置）

推荐放置由 5 项原则构成（SSOT：`.claude/rules/moai/core/askuser-protocol.md` § Recommendation Placement Principles）。

#### 原则 1 — 发起时机（信息增益对齐）

编排器估算即将到来决策的不确定性 p 时，在 p ≈ 0.5（Fisher information I = p(1−p) 最大的决策边界）通过 AskUserQuestion 发起该问题。当 p 接近 0 或 1（几乎确定）时，自动用统计多数选项解决并省略提问。

#### 原则 2 — 问题排序（信息增益降序）

在一次 AskUserQuestion 调用中放置多个问题时，把信息增益最高的问题放在最前。让用户先完成核心决策，较低价值的问题稍后再遇到。

#### 原则 3 — 推荐选项（统计多数的理性默认值）

推荐（第一个选项的 `(推荐)` 标签）以 **观察到的统计多数** 为依据。不是系统想推的策略默认值，而是用户实际反复选择的选项成为推荐。根据观察量在三种状态间切换。

##### Cold-Start（初始状态）
- **观察 < N**：缺乏足够的观察数据
- **推荐放置**：静态默认值（显式公开）
- **显示方式**：`based on static default, N observations needed for personalization`

##### Warm State（学习中）
- **观察 = N~M**：部分学习
- **推荐放置**：观察到的多数 + 置信度信号
- **置信度**：观察数 × 选择一致性

##### Mature State（稳定化）
- **观察 > M**：学习充分
- **推荐放置**：强多数确信（统计显著）
- **置信度**：最高（≥95% 置信度）

#### 原则 4 — 明示前提条件

推荐选项的说明必须明示该推荐成立的前提条件。为便于用户在前提被违反时立即拒绝，以 `"Recommended when <前提条件>"` 形式给出。未明示前提的推荐是设计缺陷。

#### 原则 5 — 基于熟练度的自适应强度

同一条推荐，面对不同的人强度也不同。对专家的强推荐会侵蚀自主性，对新手的弱推荐只会增加决策疲劳。

- **专家（会话 > 50）**：弱推荐强度（自主性优先，仅公开 inferred preference）
- **新手（会话 < 10）**：强推荐强度（`(推荐)` 标签 + 注明理由）
- **中级（10 ≤ 会话 ≤ 50）**：中等强度（视情况调整）

### 3. PostToolUse Capture Hook（决策捕获）

AskUserQuestion 的回答到达时，PostToolUse Hook 会自动捕获决策。用户无需另行记录。

#### 捕获的数据

```json
{
  "decision_id": "moai-ask-001",
  "timestamp": "2026-07-01T10:00:00Z",
  "question": "请选择下一步",
  "user_choice": "Option A (推荐)",
  "all_options": ["Option A", "Option B", "Option C"],
  "context": {
    "spec_id": "SPEC-XXX-001",
    "phase": "run",
    "workflow": "/moai run"
  }
}
```

#### 存储位置

preference 记忆把在编排器用户提问渠道中捕获的决策持久化到 `~/.claude/projects/{slug}/memory/user_decisions/`（SPEC-V3R6-ASKUSER-DECISION-MEMORY-001）。

**3-tier 层级：**

- **core**：最高优先级 hot 缓存。命中 core 时不再访问 recall/archival
- **recall**（`recall.jsonl`）：最近会话事实
- **archival**（`archival/`）：全量检索对象

### 4. Decay Policy（衰减策略）

过去的选择并不原样代表今天的偏好。transient 条目会受 **power-law 衰减 + 28 天 TTL**（REQ-ADM-011、REQ-ADM-012）作用，权重逐渐衰减，TTL 过后被逐出。core-tier 逐出采用从最低权重项开始降级的方式。

管理命令：

```bash
moai preference decay-scan   # 后台衰减遍历一次 (每天最多一次, timestamp gate)
moai preference toggle       # 会话级个性化 on/off (非持久, 每会话重置)
```

{{< callout type="info" >}}
**参考**：精确的衰减指数·半衰期常数等细节参数属于运行时实现细节（仅固定 power-law + 28 天 TTL 契约），本页只在契约层面叙述。
{{< /callout >}}

## 决策类别

记忆跟踪的主要决策类型如下。

| 类别 | 示例 |
|----------|------|
| **Tier Selection** | 选择 Tier S/M/L |
| **Cycle Type** | DDD vs TDD 模式 |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR-based |
| **Team Mode** | Solo vs Agent Teams |
| **Model Selection** | 按任务选择模型 |
| **Effort Level** | Effort 级别（low/medium/high/xhigh） |

值得注意的是 Model Selection 与 Effort Level 也包含在内 — 决策记忆学到的偏好最终会体现在模型与推理深度的分配上，因此这个系统同时也是代币经济学的个性化层。

## 统计多数学习的示例

### 场景 1: Tier Selection

假设用户做过 10 次 Tier 选择：

```
Tier S: 选择 3 次
Tier M: 选择 6 次  ← 统计多数 (60%)
Tier L: 选择 1 次

学习结果: Tier M 显示为 (推荐)
置信度: 中上 (6/10 = 60%, N=10)
推荐文案: "Tier M (推荐) — 基于最近选择 60%"
```

### 场景 2: Cycle Type

```
DDD: 4 次
TDD: 选择 5 次  ← 统计多数
其他: 1 次

学习结果: TDD 为 (推荐)
置信度: 中 (5/10 = 50%, N=10)
推荐文案: "TDD (推荐) — 基于观察"
```

## Cold-Start 透明性

观察不足时，不隐瞒这一事实，而是显式公开。

```
选项 1: Tier M (推荐) — based on static default, 5 observations needed for personalization
选项 2: Tier L
选项 3: Tier S
```

用户可以清楚地意识到系统仍处于学习状态。

## 基于熟练度的强度调整示例

### 新手用户（会话 < 10）
```
Tier M (推荐) — 基于最近选择提示
(强推荐强度)
```

### 专家用户（会话 > 50）
```
选项:
- Tier M (最近选择 60%)
- Tier L
- Tier S
(弱推荐强度, 仅公开 inferred preference)
```

## 相关文档

- [智能体指南](/zh/advanced/agent-guide) - AskUserQuestion 推荐放置规则 (HARD)
- [Harness v4 Builder 深入指南](/zh/advanced/harness-v4-builder) - Tier 选择与决策
- [记忆系统](/zh/getting-started/memory) - 用户偏好管理

{{< callout type="info" >}}
**提示**：决策记忆自动运行，无需显式配置 — 每当你做出决策，系统都会静静地学习。
{{< /callout >}}
