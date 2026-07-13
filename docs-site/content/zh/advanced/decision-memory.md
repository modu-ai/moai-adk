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

## 5 个组成部分

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

推荐（第一个选项的 `(推荐)` 标签）以 **观察到的统计多数** 为依据。根据观察量在三种状态间切换。

#### Cold-Start（初始状态）
- **观察 < N**：缺乏足够的观察数据
- **推荐放置**：静态默认值（显式公开）
- **显示方式**：`based on static default, N observations needed for personalization`

#### Warm State（学习中）
- **观察 = N~M**：部分学习
- **推荐放置**：观察到的多数 + 置信度信号
- **置信度**：观察数 × 选择一致性

#### Mature State（稳定化）
- **观察 > M**：学习充分
- **推荐放置**：强多数确信（统计显著）
- **置信度**：最高（≥95% 置信度）

#### 基于熟练度的自适应强度

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

- **会话中**：`.moai/state/decisions/`（临时 JSON）
- **会话结束**：`~/.claude/projects/{hash}/memory/decisions.jsonl`（自动记忆）

### 4. Decay Policy（衰减策略）

3 个月前的选择并不代表今天的偏好。旧决策的权重会逐渐衰减。

#### 衰减函数

```
weight(t) = initial_weight × exp(-decay_rate × days_ago)
```

#### 默认值
- **Initial weight**: 1.0
- **Decay rate**: 0.1（每 7 天约衰减 50%）
- **Retention period**: 90 天（此后自动归档）

#### 示例

```
昨天的选择: weight = 0.95
7 天前的选择: weight = 0.50
30 天前的选择: weight = 0.04
90 天以上: 归档 (不再纳入推荐)
```

### 5. Recovery Controls（恢复控制）

为学习方向固化出错的情况提供纠错与重置手段。

#### 重置记忆

用户可以重置已学习的偏好。

```bash
/moai memory reset
```

#### 编辑偏好

修改特定决策类别的推荐。

```bash
/moai memory set <category> <preferred-option>
```

#### 查询偏好

查看当前已学习的偏好。

```bash
/moai memory list
```

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
