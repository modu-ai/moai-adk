---
title: 配置章节参考
weight: 71
draft: false
description: ".moai/config/sections/ 的主要配置文件键参考 (handoff/delegation/llm/statusline/security)。"
---

MoAI-ADK的项目设置分为 `.moai/config/sections/` 下的多个YAML文件。[settings.json指南](/zh/advanced/settings-json)涵盖Claude Code运行时设置，而本页面记录控制MoAI-ADK自身行为的主要section文件的键。

{{< callout type="info" >}}
**一句话总结**: `settings.json` 定义允许Claude Code做什么，`.moai/config/sections/*.yaml` 定义MoAI-ADK如何编排。
{{< /callout >}}

## handoff.yaml — 自动恢复交接

控制跨会话边界时如何处理已保存的交接。

```yaml
handoff:
    mode: manual   # manual | auto
    guide: false
```

| 键 | 值 | 说明 |
|----|-----|------|
| `mode` | `manual` (默认) | 不自动注入已保存的交接 (opt-in基准UX) |
| `mode` | `auto` | 在 `/clear` 时将已保存的交接注入会话上下文，然后移动到audit-trail副本 |
| `guide` | `false` (默认) | 为 `true` 时，在非-`/clear` 会话开始(startup/resume/compact)时发出关于有待处理交接的best-effort stderr提示。仅信息性，不阻止会话 |

相关: [自主持续循环](/zh/advanced/autonomous-loops), [moai handoff](/zh/cli-reference/handoff).

## delegation.yaml — 代理路由SSOT

这是 `/moai` 子命令的默认技能/代理分配映射。当编排器构建执行计划(Analyze-First)时，它读取此映射以决定spawn哪些代理以及注入哪些技能。

```yaml
delegation:
    version: 1
    learning:
        observe: routing-ledger
        propose_via: harness-tier-ladder
        auto_apply: false          # Tier-4网关 — 需要用户批准
    subcommands:
        plan:
            agents: [manager-spec, plan-auditor, Explore]
            skills: [moai-workflow-spec, moai-foundation-thinking]
        # run / sync / project / fix / loop / ...
    domain_skills:
        backend:  [moai-ref-api-patterns, moai-domain-backend]
        security: [moai-ref-owasp-checklist, moai-ref-llm-security, ...]
    agents:
        manager-spec: [moai-workflow-spec, moai-foundation-thinking]
```

| 块 | 说明 |
|------|------|
| `learning` | 将路由使用管理为仅追加账本 (`.moai/state/routing-ledger.jsonl`, opt-in·fail-open)，harness学习子系统通过4-tier阶梯提出更新。`auto_apply: false` — Tier-4变更需要 `AskUserQuestion` 用户批准 |
| `subcommands` | 每个子命令的 `agents` (要spawn的11个retained代理) + `skills` (spawn时注入的workflow技能)。0个分配也有效 (编排器直接执行) |
| `domain_skills` | 按任务域注入的技能 (每次spawn 0-3个)。与域信号匹配 |
| `agents` | 每个代理的条件技能 (触发时on-demand加载) |

相关: [代理指南](/zh/advanced/agent-guide), [技能指南](/zh/advanced/skill-guide).

## llm.yaml — 后端·模型层级

定义性能层级、计费计划和Claude/GLM模型映射。

```yaml
llm:
  performance_tier: "medium"   # high | medium | low
  plan_type: "subscription"    # api | subscription
  claude_models:
    high: "opus"
    medium: "sonnet"
    low: "sonnet"
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.2"          # 1M context — Opus 插槽
      medium: "glm-4.7"        # 202K context — Sonnet 插槽
      low: "glm-4.5-air"       # 128K context — Haiku 插槽
      fable: "glm-5.2"
```

| 键 | 说明 |
|----|------|
| `performance_tier` | 控制所有子代理·团队代理的模型选择 (high=复杂推理, medium=平衡, low=快速/低成本) |
| `plan_type` | 计费计划类型。`api`=按任务成本优化，`subscription`=每周配额优化 (空值解释为subscription) |
| `claude_models` | 层级到Claude模型的映射。Harness级别映射到effort (thorough→xhigh, standard→high, minimal→medium) |
| `glm.base_url` | Z.AI Anthropic兼容代理端点 |
| `glm.models` | 层级到GLM模型的映射。GLM将Claude的5步effort折叠为3个推理状态 (thinking-off / reasoning-high / reasoning-max) |

相关: [plan_type层级配置](/zh/advanced/plan-type-profiles), [3层代理架构](/zh/advanced/no-haiku-3tier).

## statusline.yaml — 状态栏

控制statusline主题和16个段切换。

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  segments:
    model: true
    context: true
    # ... 共16个段 (全部默认on)
    task: true
    pr: true
```

| 键 | 说明 |
|----|------|
| `theme` | 恰好存在2个主题: `catppuccin-mocha` (默认) 或 `catppuccin-latte` |
| `segments` | 16个段单独切换 (唯一的运行时控制杆)。全部默认on，非活动状态优雅处理为无输出 |

段放置在3行中 — 行1(模型·版本·会话元数据)，行2(上下文窗口·API使用量栏)，行3(目录·git·工作流·PR)。

相关: [Statusline系统 & PR段](/zh/advanced/statusline).

## security.yaml — 安全加固

扩展(非替换)内置 `DefaultSecurityPolicy` 模式的额外安全设置。遵循SOLID的开闭原则 — 无core修改的config扩展。

```yaml
security:
  extra_dangerous_bash_patterns:
    - 'curl\s+.*\|\s*(ba)?sh'
    - 'rm\s+-rf\s+/[^.]'
  extra_deny_patterns: []
  extra_ask_patterns: []
  permission:
    strict_mode: true
    session_rules: []
  sandbox:
    required: false
    network_allowlist: []
    env_scrub_extra: []
    docker_image: "alpine:latest"
```

| 键 | 说明 |
|----|------|
| `extra_dangerous_bash_patterns` | **添加到**内置deny模式的危险Bash命令正则表达式 (不区分大小写) |
| `extra_deny_patterns` / `extra_ask_patterns` | 额外文件deny/ask模式 |
| `permission.strict_mode` | 为 `true` 时，拒绝bypassPermissions模式的代理spawn |
| `sandbox.required` | 为 `true` 时，拒绝无 `sandbox.justification` 的 `sandbox: none` 代理 (默认false) |
| `sandbox.network_allowlist` | **添加到**默认8个主机的允许网络主机 |
| `sandbox.env_scrub_extra` | **添加到**默认scrub列表的环境变量名 (AWS_*, GITHUB_TOKEN等) |
| `sandbox.docker_image` | docker后端默认镜像 |

相关: [安全说明](/zh/advanced/security-notes), [settings.json指南](/zh/advanced/settings-json).

## 相关文档

- [settings.json指南](/zh/advanced/settings-json) — Claude Code运行时设置
- [Harness配置与评估](/zh/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/zh/cli-reference/doctor) — `moai doctor config` 用于合并配置检查
