---
title: 模型策略
weight: 30
draft: false
---

## 什么是模型策略？

模型策略是 MoAI-ADK 代币经济学的骨架。它不是"所有任务都用最强模型"，
而是为每个代理 —— 计划、审计等推理繁重的工作与文档化、Git 等轻量工作 ——
声明式地分配合适的模型。它配合 Claude Code 订阅计划最大化质量，同时避免
速率限制错误。

MoAI-ADK v3.0 的代理目录共 **11 个**（MoAI 自定义 10 个 + Anthropic 内置
`Explore`）。在 **No-Haiku 策略** 下，Haiku 不出现在任何位置。Opus 承担所有
多轮代理式行，Sonnet 被限定在单次完成、以输入为主的行；策略级别控制的是每个
代理落在 Opus effort 阶梯的哪一级，而不是它获得哪个模型等级。

## 3 级策略概览

| 策略（配置文件） | CLI 标志 | Opus 格数 | Sonnet 格数 | 适合用途 |
|------|------|---------|-----------|-----------|
| **high** | `--model-policy high` | 11 中的 9 | 11 中的 2 | 最高质量；两个调用频率最低的行使用 `max` effort |
| **medium**（默认） | `--model-policy medium` | 11 中的 9 | 11 中的 2 | 质量与成本的平衡；成本/得分曲线的拐点 |
| **low** | `--model-policy low` | 11 中的 7 | 11 中的 4 | 每任务成本最低；代理式行降到 Opus `low` |

> **名称对应**：`llm.yaml` 的 `profile` 字段、legacy `performance_tier` 别名与
> CLI 标志 `--model-policy` 都使用同样的 `high`/`medium`/`low` 三个值并 1:1
> 对应（无需单独转换）。默认值是 `medium`。旧的顶层名称 `max` 仍会被**读取**
> 为 `high` 的别名，因此既有配置仍能解析，但保存时始终写入 `high` —— 无需任何
> 迁移操作。`performance_tier` 仅在 `profile` 缺失时读取。用户名等信息单独保存
> 在 `user.yaml` 中。

> **为什么重要？** 降低策略不再意味着切换到更弱的模型等级。在长周期代理式任务上，
> Opus 以 `low` effort 运行时比任何 effort 的 Sonnet 得分更高**且**每任务成本更低，
> 因为账单由模型完成任务所花的步数决定 —— 而不是单 token 费率。因此 `low` 策略是
> 在 Opus *内部*通过降低推理深度来节省，只在多步完成失败不适用的单次完成行上才
> 动用 Sonnet。

## 各代理的模型分配表

下面的 33 个格子就是配置矩阵（11 个代理 × 3 个配置文件）。每个格子是解析器在
spawn 时注入的 `{model, effort}` 对。（编排器主会话不是被 spawn 的代理，因此
不在表中。）

### Manager Agents（5 个）

| 代理 | high | medium | low |
|---------|------|--------|-----|
| manager-spec | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| manager-design | opus / high | opus / medium | opus / low |

### Evaluator · Advisor · Builder · Specialist Agents（5 个）

| 代理 | high | medium | low |
|---------|------|--------|-----|
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |

### 内置代理（1 个）

| 代理 | high | medium | low |
|---------|------|--------|-----|
| Explore | sonnet / low | sonnet / low | sonnet / low |

> `Explore` 在磁盘上没有代理文件，因此它的 effort 无法固定在 frontmatter 中
> —— 矩阵将 `sonnet / low` 记录为调用时的默认值，并在 spawn 提示词中说明。
> Agent Teams 静态层（静态 role profile）在 v3.0 中已退役，并行工作由
> sub-agent 并行执行和动态工作流取代。`moai cg` 的 teammate 运行时
> （tmux pane）保持不变。

> **Haiku 移除（v3.0）**：原先的 Haiku 槽位（文档化、MX 标注、Git 流程）被替换为
> 更低的推理深度，而非更低的模型等级 —— 成本靠 effort 分级削减，而不是靠替换模型。

## 分配原则

- **所有代理式行都用 Opus**：`manager-spec`、`manager-develop`、`plan-auditor`、`sync-auditor`、`manager-design`、`builder-harness`、`manager-docs`、`e2e-tester` —— 所有多轮工作都留在 Opus 上，因为 Opus 在 `low` 下比任何 effort 的 Sonnet 得分更高，而每任务成本更低
- **Sonnet 仅用于单次完成行**：`manager-git` 的机械性操作与 `Explore` 的检索都在一次以输入为主的通过中完成，多步完成失败不适用，因此 Sonnet 更低的输入价格才是决定性因素。这两行在三个配置文件下固定不变
- **`max` 仅限两格**：`manager-develop` 与 `super-advisor`，且仅在 `high` 配置文件下 —— 调用频率最低的行，那里单个决策承载不成比例的下游成本
- **`xhigh` 在任何地方都不使用**：在 Opus 上它与 `high` 得分相同，成本却高 49%
- **`low` 降的是 effort，不是模型等级**：代理式行移到 Opus `low`；只有 `manager-docs` 与 `e2e-tester` 会额外回退到 Sonnet

为了避免"制定计划的代理自己做审计"，plan-auditor 和 sync-auditor 的分配与
manager-spec 保持独立 —— 偏差防止是目录的结构性属性，而不是格子取值的属性。

## v3.0 扩展：Tier×Phase 声明矩阵

v3.0 在代理级分配之上，新增了**工作阶段 (phase) 与 SPEC 规模 (Tier)** 轴。
`internal/config/model_routing.go` 以声明方式管理 Tier×Phase →
{model, effort} 矩阵：

- **model**: inherit / sonnet / opus / glm / fable
- **effort**（推理深度）: low / medium / high / xhigh / max
- **tier**（SPEC 规模）: S / M / L
- **phase**（工作阶段）: plan / run / sync / mx

每个代理的 model+effort 分配由单一配置矩阵负责。活动配置文件(`profile` ——
`high`/`medium`/`low`)选择矩阵的一列，`profile` 缺失时 legacy `performance_tier`
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
moai init my-project --model-policy high    # 最高质量 (2 行使用 max effort)
moai init my-project --model-policy medium  # 平衡 (默认值)
moai init my-project --model-policy low     # 每任务成本最低
```

`--model-policy` 接受 `high`/`medium`/`low` 三个值,并持久化到 `llm.yaml` 的
`performance_tier` 字段。旧的顶层名称 `max` 仍可作为输入被接受,并按 `high` 的
别名处理。

> 默认策略为 `medium`(llm.yaml `performance_tier: "medium"`,对应 CLI `--model-policy medium` —— 无值时解释为 `medium`)。GLM 配置隔离在 `settings.local.json` 中,不会提交到 Git。

## 下一步

- [CG 模式](/zh/multi-llm/cg-mode) —— 用 Claude + GLM 混合节省成本
- [代理指南](/zh/advanced/agent-guide) —— 代理自定义
- [CLI 参考](/zh/getting-started/cli) —— moai init, moai update 详解
