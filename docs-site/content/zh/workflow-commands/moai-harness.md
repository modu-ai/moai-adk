---
title: /moai harness
weight: 55
draft: false
---

运维 V3R4 Self-Evolving Harness 学习子系统的命令。介绍 4 阶段进化阶梯（observer → heuristic → rule → frozen-zone）以及 5 层安全流水线（frozen-guard → canary → contradiction → rate-limit → human oversight）。

{{< callout type="info" >}}
**斜杠命令**: 在 Claude Code 中输入 `/moai harness` 即可直接执行该命令。
{{< /callout >}}

## 概述

`/moai harness` 提供 4 个 verb（`status`、`apply`、`rollback`、`disable`），用于安全运维 MoAI-ADK 的自演化学习子系统。PostToolUse 钩子会将每次工具调用以 append-only 方式写入 `.moai/harness/usage-log.jsonl`，提案沿 4-tier 进化阶梯进行分类。任何 Tier-4 的 frozen-zone 变更都必须通过 `AskUserQuestion` 获得用户的明确批准后才会应用。

核心概念:

- **Observer**: PostToolUse 钩子把每次工具调用追加到 `.moai/harness/usage-log.jsonl`。
- **4-Tier Evolution Ladder**: observation → heuristic → rule → frozen-zone 提案的 4 阶段。
- **5-Layer Safety Pipeline**: 每条进化提案都必须通过 5 层安全验证才能应用。
- **CLI Retirement**: 自 V3R4 起,所有 verb 仅在 workflow body 中通过文件系统操作执行,Go 二进制不再暴露 `moai harness` 子命令。

## 命令格式

```bash
/moai harness {status | apply | rollback <YYYY-MM-DD> | disable}
```

- 参数为空时输出帮助。
- 所有 verb 均在 orchestrator 主上下文中执行。

## verbs 详解

### status

输出当前 harness 学习状态、待处理 Tier-4 提案数量,以及 7 天 rate-limit 窗口使用量。

- **只读**: 不修改任何文件。
- **输出包含**:
  - `.moai/config/sections/harness.yaml` 中 `learning.enabled` 的值
  - `.moai/harness/proposals/` 中待处理 Tier-4 提案数
  - `.moai/harness/learning-history/applied/` 目录,过去 7 天内的应用次数
  - 最近的 tier 提升事件(`tier-promotions.jsonl`)
  - Frozen Guard 违规日志(`frozen-guard-violations.jsonl`)

### apply

将最久未处理的 Tier-4 提案送入 5-Layer Safety 流水线进行应用。应用前必须由 orchestrator 执行一次 `AskUserQuestion` 轮次,获得用户的明确批准。

- **前置条件**:
  - 7 天窗口内应用次数小于 1 次(REQ-HRN-FND-012 rate-limit floor)。
  - 提案载荷完整性校验通过。
- **用户选项(推荐 / Modify / Defer / Reject)**: 首选项带 `(推荐)` 标识。选择 Apply 时,事前快照将保存到 `.moai/harness/learning-history/snapshots/<ISO-DATE>/`。

### rollback `<YYYY-MM-DD>`

使用指定日期的快照回滚最近一次应用。若期间已累积其他进化,将输出冲突报告并请求用户再次批准。

- **参数**: ISO-8601 日期(YYYY-MM-DD)。格式错误将报错。
- **效果**: `.moai/harness/learning-history/applied/<DATE>.json` 会被移动至 `rolled-back/`,相关文件恢复到快照状态。

### disable

暂停 harness 学习(`learning.enabled: false`)。PostToolUse 观察继续运行,但 4-tier 分类器与提案生成器停止工作。

- **使用场景**: 当进化提案可疑或正在进行外部审计时。
- **重新启用**: 在 `.moai/config/sections/harness.yaml` 中将 `learning.enabled` 设回 `true`。

## 4-Tier Evolution Ladder

| Tier | 分类 | 自动应用 | 备注 |
|------|------|----------|------|
| Tier-1 | Observation | n/a(人工复核) | 仅被动累积日志 |
| Tier-2 | Heuristic | 仅提示 | orchestrator 向用户建议 |
| Tier-3 | Rule | 非 frozen 区域可自动应用 | 必须通过 canary |
| Tier-4 | Frozen-zone | **必须用户批准** | 必须完成 5-Layer Safety |

Frozen 区域由 `.claude/rules/moai/design/constitution.md` §2 与 `.claude/rules/moai/core/zone-registry.md` 定义。

## 5-Layer Safety Pipeline

1. **L1 Frozen Guard**: 阻止对 frozen 区域的修改尝试。
2. **L2 Canary**: 在隔离沙箱中模拟变更影响。
3. **L3 Contradiction**: 检测与其他生效规则的冲突。
4. **L4 Rate Limit**: 7 天窗口内最多应用 1 次(REQ-HRN-FND-012)。
5. **L5 Human Oversight**: 由 orchestrator 主导的 `AskUserQuestion` 批准轮次。

任意一层拒绝,`apply` 即中止,提案保持 `pending` 状态。

## 使用示例

```bash
# 1) 查看当前状态
/moai harness status

# 2) 复核并应用待处理的 Tier-4 提案
/moai harness apply

# 3) 用昨日快照回滚最近一次应用
/moai harness rollback 2026-05-21

# 4) 暂停学习
/moai harness disable
```

## 相关资料

- [`.claude/skills/moai/workflows/harness.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`SPEC-V3R4-HARNESS-001`](https://github.com/modu-ai/moai-adk) — V3R4 foundation SPEC(合并三个 V3R3 harness SPEC)
- [`/moai plan`](/zh/workflow-commands/moai-plan) — SPEC 文档创建
- [`/moai run`](/zh/workflow-commands/moai-run) — DDD/TDD 实现
- [`/moai sync`](/zh/workflow-commands/moai-sync) — 文档同步 + PR
