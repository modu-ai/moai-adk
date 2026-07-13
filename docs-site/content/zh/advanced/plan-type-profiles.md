---
title: plan_type 层级配置
weight: 4
draft: false
---

MoAI-ADK 认识到即使是相同的工作流，API 按量付费和订阅计划的最优分配也不同。`plan_type` 轴按计费模型分别应用 Tier × Phase 模型/effort 矩阵。本页正式文档化由 SPEC-MODEL-TIER-PLANTYPE-001 (CLOSED)实现的 60 格配置矩阵。

## plan_type 轴

`plan_type` 有两个值:

- `api` — 按量付费。美元是唯一约束。目标是每任务成本优化。
- `subscription` — 订阅计划。周 token 配额 + Opus 加权扣减是约束。目标是最大化每配额已解决任务数。

在订阅计划中，Opus 时间另行加权扣减(Max 5x: Sonnet 140-280h vs Opus 15-35h，约 1/8)。因此，在订阅中，Opus 仅分配给推理，执行在丰富的 Sonnet 时间上运行 — opusplan 结构是最优的。

## 设置 plan_type

```bash
moai init . --plan-type api           # 初始化时设置
moai update --plan-type subscription  # 事后切换
```

可以在 `llm.yaml` 的 `llm.plan_type` 字段查看当前值。

## 60 格配置矩阵

10 代理 × 3 层级 × 2 plan_type = 60 格。下表是 SPEC-MODEL-TIER-PLANTYPE-001 的 ApplyTierProfile 实现。

### Plan A — API 按量付费 (rev2)

在 API 中，美元是唯一约束。rev2 修改: Sonnet 名义单价是 Opus 的一半，但每任务成本反转(Opus $13.22 < Sonnet $26.40)。因此，API 在执行中也使用 Opus。推理 = Fable(质量第一)、执行 = Opus(每任务成本第一)、机械 = Sonnet low。

| 代理 (角色) | A-max (质量) | A-medium (推荐) | A-low (成本) |
|---|---|---|---|
| manager-spec (推理) | fable / high | fable / high | opus / high |
| plan-auditor (推理) | fable / high | fable / high | opus / high |
| sync-auditor (推理) | fable / high | opus / high | opus / medium |
| manager-design (推理) | fable / high | fable / high | opus / high |
| super-advisor (最高推理) | fable / xhigh | fable / high | opus / high |
| manager-develop (执行) | fable / high | opus / high | opus / medium |
| builder-harness (执行) | opus / high | opus / medium | opus / medium |
| manager-docs (机械) | sonnet / medium | sonnet / low | sonnet / low |
| manager-git (机械) | sonnet / low | sonnet / low | sonnet / low |
| Explore (探索) | inherit / medium | inherit / low | inherit / low |

### Plan B — 订阅 (可用性优先)

订阅的约束不是美元而是周 token 配额 + Opus 加权扣减。目标是最大化"每配额已解决任务数"= 排除重试循环 + Opus 仅分配给推理。这是 Anthropic 官方 opusplan 模式("Opus 用于规划，Sonnet 用于执行")的精确版本。

| 代理 (角色) | B-max (推荐) | B-medium | B-low (Pro) |
|---|---|---|---|
| manager-spec (推理) | opus / high | opus / high | opus / medium |
| plan-auditor (推理) | opus / high | opus / medium | sonnet / high |
| sync-auditor (推理) | opus / high | opus / medium | sonnet / high |
| manager-design (推理) | opus / high | opus / medium | sonnet / high |
| super-advisor (最高推理) | opus / xhigh | opus / high | opus / medium |
| manager-develop (执行) | sonnet / high | sonnet / high | sonnet / high |
| builder-harness (执行) | sonnet / high | sonnet / medium | sonnet / medium |
| manager-docs (机械) | sonnet / low | sonnet / low | sonnet / low |
| manager-git (机械) | sonnet / low | sonnet / low | sonnet / low |
| Explore (探索) | inherit / medium | inherit / low | inherit / low |

## ApplyTierProfile 机制

`ApplyTierProfile` 替换代理 frontmatter 中的 `model` 和 `effort`(replace-both)。由于所有代理都有 `effort:` 字段，"保留"模式无效，因此始终以 replace-both 运作。

此机制在 SPEC-MODEL-TIER-PLANTYPE-001(运行阶段完成，CLOSED)中实现。上表所有格子均作为实时行为验证完毕。

## GLM 后端 effort 叠加

{{< icon warning warn >}} **诚实性说明 (REQ-DA-060)**: GLM 后端 effort 叠加的 wire 有效性是需要实时 GLM 会话出站观测的实证课题。

GLM 后端(`moai glm` / `moai cg` GLM 面板)将 Claude 的 5 级 effort(max / xhigh / high / medium / low)折叠为 GLM 的 3 级 reasoning_effort(high / max)应用。实现内容:

- `IsGLMBackend` 检测识别 GLM 会话
- 5 级 → 3 级折叠映射(max/xhigh → max, high → high, medium/low → GLM 不支持)
- coding 作业时 max override

**已实现 + 已接线，wire 有效性待验证** — z.ai 的 Anthropic-compat shim 是否实际消费 `ANTHROPIC_REASONING_EFFORT` 环境变量值，是需要实时 GLM 会话出站观测的运行阶段实证课题。本页不以"保证工作"描述，而是以"已实现 + 已接线，wire 有效性待验证"记载。

## 模型策略看板 (moai web)

`moai web` 的 `/model-policy` 看板可视化地确认和设置 plan_type 与层级配置。此看板是 SPEC-WEB-CONSOLE-013 的批准例外，允许 plan_type 写入。

## 路线图

{{< icon clock >}} **spawn-time 36 格路由** (SPEC-MODEL-TIER-ROUTING-PROFILES-001) — 目前 ApplyTierProfile 以代理为单位路由。spawn-time 结合 phase 和 SPEC Tier 的 36 格精确路由是已取消范围的后续 SPEC。目前以代理 frontmatter 的 model/effort 被 ApplyTierProfile replace-both 的结构运作。

## 下一步

- [三层代理架构](/zh/advanced/no-haiku-3tier/) — DeepSWE 排行榜依据和三层定义
- [代币经济学概述](/zh/advanced/tokenomics-overview/) — 四层代币经济学结构的 Layer B 路由
