---
title: "模型策略"
weight: 30
draft: false
---

## 什么是模型策略?

MoAI-ADK为8个保留Agent（7 MoAI自定义 + Anthropic内置Explore）各自分配最优AI模型。根据Claude Code订阅套餐最大化质量，同时防止速率限制错误。

## 3级策略概览

| 策略 | 套餐 | 🟣 Opus | 🔵 Sonnet | 🟡 Haiku | 适用场景 |
|------|------|---------|-----------|----------|----------|
| **High** | Max $200/月 | 5 | 1 | 1 | 最高质量，最大吞吐 |
| **Medium** | Max $100/月 | 2 | 3 | 2 | 质量与成本平衡 |
| **Low** | Plus $20/月 | 0 | 4 | 3 | 低预算，无Opus |

> **为什么重要?** Plus $20套餐无法访问Opus。设置`Low`策略后，所有Agent仅使用Sonnet和Haiku，防止速率限制错误。高级套餐将Opus分配给核心Agent（安全、策略、架构），日常任务使用Sonnet/Haiku。

## 按Agent分配的模型表

### Manager Agent (4个)

| Agent | High | Medium | Low |
|-------|------|--------|-----|
| manager-spec | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-develop | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-docs | 🔵 sonnet | 🟡 haiku | 🟡 haiku |
| manager-git | 🟡 haiku | 🟡 haiku | 🟡 haiku |

### 评估 & Builder Agent (3个)

| Agent | High | Medium | Low |
|-------|------|--------|-----|
| plan-auditor | 🟣 opus | 🟣 opus | 🔵 sonnet |
| sync-auditor | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| builder-harness | 🟣 opus | 🔵 sonnet | 🟡 haiku |

> Team模式的角色（researcher, analyst, architect, implementer, tester, designer, reviewer）不是静态Agent，而是通过`workflow.yaml`的role profile经由`Agent(general-purpose)`动态生成。

## 分配原则

- **始终Opus**: 计划审计（plan-auditor）、SPEC编写（manager-spec） — 需要高推理能力
- **始终Haiku**: 文档（manager-docs）、Git（manager-git） — 轻量快速任务
- **按套餐变动**: 实现（manager-develop, cycle_type=tdd/ddd） — 高级套餐使用Opus

## 配置方法

### 项目初始化时

```bash
moai init my-project
# 交互式向导包含模型策略选择
```

### 重新配置现有项目

```bash
moai update
# 交互式提示:
# - Reset model policy? (y/n) — 重置模型策略
# - Update GLM settings? (y/n) — 配置GLM环境变量
```

> 默认策略为`High`。GLM设置隔离在`settings.local.json`中，不会提交到Git。

## 下一步

- [CG模式](/zh/multi-llm/cg-mode) — 通过Claude + GLM混合降低成本
- [Agent指南](/zh/advanced/agent-guide) — 自定义Agent
- [CLI参考](/zh/getting-started/cli) — moai init, moai update详情
