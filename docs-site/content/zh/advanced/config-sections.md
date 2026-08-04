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

## llm.yaml — 后端·配置矩阵

定义配置文件、配置矩阵、每个代理的 override 以及 GLM 模型映射。

```yaml
llm:
  profile: "medium"            # high | medium | low (活动矩阵列; max 读作 high)
  performance_tier: "medium"   # legacy 别名 (profile 缺失时读取; 同一套词汇)
  profiles:                    # 配置文件列 → 11 个代理 → {model, effort}
    high: { ... }              # 详表: 配置矩阵页面
    medium: { ... }
    low: { ... }
  agent_overrides: {}          # 每个代理的 {model, effort} override (可选)
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.2"          # 1M context — Opus 插槽
      medium: "glm-4.7"        # 202K context — Sonnet 插槽
      low: "glm-4.5-air"       # 128K context — 轻量插槽
      fable: "glm-5.2"
```

| 键 | 说明 |
|----|------|
| `profile` | 活动配置矩阵列 (`high`/`medium`/`low`; 旧的 `max` 被读作 `high` 的别名)。为空时解释为 `medium`。所有子代理 spawn 的 model+effort 来源 |
| `performance_tier` | legacy 别名字段。仅当 `profile` 缺失时读取; 与 `profile` 共享同一套 `high`/`medium`/`low` 词汇，因此不需要归一化步骤 |
| `profiles` | 每个配置文件列的 per-agent → `{model, effort}` 矩阵 (11 个代理 × 3 列 = 33 格)。Go 默认值(`template.DefaultProfileMatrix`)是缺失格的权威 fallback |
| `agent_overrides` | 每个规范代理名称的 `{model, effort}` override。优先于活动配置文件的代理格 (目录+enum 校验) |
| `glm.base_url` | Z.AI Anthropic兼容代理端点 |
| `glm.models` | 每个插槽的 GLM 模型映射。GLM将Claude的5步effort折叠为3个推理状态 (thinking-off / reasoning-high / reasoning-max) |

相关: [配置矩阵](/zh/advanced/profile-matrix), [3层代理架构](/zh/advanced/no-haiku-3tier).

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
| `segments` | 16个段单独切换 (唯一的运行时控制杆)。全部默认on，非活动状态正常处理为无输出 |

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

## ralph.yaml — Ralph Engine

控制 `/moai loop` 的基于诊断的迭代修复循环(Ralph Engine)行为。

```yaml
ralph:
  max_iterations: 5         # 迭代上限 (默认 5; CLI --max 优先)
  auto_converge: true       # 停滞检测时自动收敛
  human_review: true        # 在审查阶段中断以引入人工
  lint_as_instruction: true # LSP diagnostic 注入为下一轮指令
  warn_as_instruction: false # 仅在没有 error 时注入 warning
```

| 键 | 说明 |
|----|------|
| `max_iterations` | 迭代上限(默认 5)。优先级:CLI `--max` 标志 > `ralph.max_iterations` > `workflow.yaml loop_prevention.max_iterations` |
| `auto_converge` | 连续 N 轮无进展时自动判定收敛 |
| `human_review` | 在审查阶段中断,让人类介入 |
| `lint_as_instruction` | 把 LSP diagnostic 作为 `systemMessage` 注入,让 AI 在下一个 prompt 中收到(默认 true) |
| `warn_as_instruction` | 在没有 error 时也注入 warning(默认 false) |

相关: [/moai loop](/zh/utility-commands/moai-loop), [自主持续循环](/zh/advanced/autonomous-loops).

## harness.yaml — harness 深度·评估

定义 harness 质量管道深度(minimal/standard/thorough)与自动检测、评估者 memory scope、升级机制。

```yaml
harness:
  default_profile: "default"  # default | strict | lenient | frontend
  evaluator:
    memory_scope: per_iteration  # FROZEN — 不可变更 (design-constitution §11.4.1)
  mode_defaults:
    solo: auto
    team: auto
    cg: thorough                 # CG 模式恒为 thorough
  auto_detection:
    enabled: true
    rules:
      minimal:  # file_count <= 3 AND single_domain, ...
      standard: # file_count > 3 OR multi_domain, ...
      thorough: # security/payment 关键字, critical 优先级, ...
  escalation:
    enabled: true
```

| 块 | 说明 |
|------|------|
| `default_profile` | SPEC 中没有 `evaluator_profile` 时使用的基础评估 profile |
| `evaluator.memory_scope` | 评估者内存范围。固定为 `per_iteration`(FROZEN) |
| `mode_defaults` | 按执行模式(solo/team/cg)的默认深度 |
| `auto_detection.rules` | Complexity Estimator 自动分类为 minimal/standard/thorough 的条件 |
| `escalation` | 失败时升级到更高深度 |

相关: [harness profile 与评估](/zh/advanced/harness-profiles), [moai harness](/zh/cli-reference/harness).

## quality.yaml — 质量门禁·开发方法论

控制开发模式(DDD/TDD)、覆盖率目标、LSP 质量门禁阈值、质量门禁是否强制执行。

```yaml
constitution:
  development_mode: tdd        # tdd | ddd | off
  enforce_quality: true
  test_coverage_target: 85
  lsp_quality_gates:
    enabled: true
    plan: { require_baseline: true }
    run:  { max_errors: 0, max_type_errors: 0, max_lint_errors: 0, allow_regression: false }
    sync: { max_errors: 0, max_warnings: 10, require_clean_lsp: true }
```

| 块 | 说明 |
|------|------|
| `development_mode` | `tdd` / `ddd` / `off`。决定 `/moai run` 的 cycle_type 默认值 |
| `enforce_quality` | 为 true 时,质量门禁违反会让 run-phase 无法进入 GREEN |
| `test_coverage_target` | 包级覆盖率最低目标(%)。关键包(cli/template/hook)建议 90%+ |
| `lsp_quality_gates.plan` | plan 阶段:是否捕获 LSP baseline |
| `lsp_quality_gates.run` | run 阶段:errors/type errors/lint errors 为 0,禁止回退 |
| `lsp_quality_gates.sync` | sync 阶段:errors 0、warning ≤ 10、要求 clean LSP |

相关: [SPEC 工作流](/zh/advanced/spec-workflow), [LSP 门禁](/zh/advanced/lsp-gates).

## workflow.yaml — 工作流状态

控制工作流阈值、branch-state guard opt-in、会话 worktree opt-in、agentic 循环上限。

```yaml
workflow:
  agentic_loop:
    max_iterations: 10        # 管道级 completion-loop 上限
  branch_guard:
    enabled: false            # 分布式默认 — hook 处于 INERT (SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001)
  session_worktree:
    enabled: false            # 分布式默认 — 自动隔离 INERT
  loop_prevention:
    max_iterations: 5         # per-operation 诊断 fix-loop 上限 (与 ralph.max_iterations 不同轴)
```

| 块 | 说明 |
|------|------|
| `agentic_loop.max_iterations` | 管道级 completion-loop 上限(`AgenticLoopConfig`) |
| `branch_guard.enabled` | Main-Checkout Branch-State Guard 是否启用。**分布式默认 `false`** —— 非共享 checkout 的 1 人开发环境中 hook 不评估。共享多会话 checkout 的运维者 opt-in |
| `session_worktree.enabled` | `moai init` / `moai profile` / `moai web` 时自动启用 worktree 隔离。**分布式默认 `false`** (SPEC-SESSION-WORKTREE-001)。可通过 `MOAI_SESSION_WORKTREE` 环境变量重定义 |
| `loop_prevention.max_iterations` | per-operation 诊断 fix-loop 上限。与 `ralph.max_iterations` 不同轴 —— ralph.yaml 优先 |

> branch_guard 与 session_worktree 的分布式默认值均为 `false`。在单用户仓库中它们为 `false` 是预期行为,不是缺陷。

相关: [Main-Checkout Branch Guard](/zh/advanced/branch-guard), [Worktree 集成](/zh/advanced/worktree).

## mx.yaml — @MX 标签系统

定义 `@MX` 代码注释标签系统的标签种类、按语言检测和校验规则。

```yaml
mx:
  version: "2.1"
  languages:
    go:      { enabled: auto, patterns: ["*.go"], exclude: ["*_generated.go", "vendor/**"] }
    python:  { enabled: auto, patterns: ["*.py"], exclude: ["**/__pycache__/**"] }
    # ... 16 个语言平级 (go, python, typescript, javascript, rust, java, kotlin,
    #     csharp, ruby, php, elixir, cpp, scala, r, flutter, swift)
  tags:
    # 各标签的元数据与启用与否
  reason_required:  # @MX:REASON 必填标签列表
    - WARN
    - DEBT
```

| 块 | 说明 |
|------|------|
| `version` | MX 标签系统 schema 版本 |
| `languages` | 16 个语言平级列出。每个语言基于项目 marker(`go.mod`、`pyproject.toml`、`Cargo.toml` 等)自动检测(`enabled: auto`) |
| `tags` | `@MX:NOTE` / `@MX:WARN` / `@MX:ANCHOR` / `@MX:TODO` / `@MX:SPEC` / `@MX:DEBT` / `@MX:LEGACY` 等标签的元数据。`@MX:SPEC` 记录 SPEC 关联关系(SPEC-MX-ASSOCIATION-001) |
| `reason_required` | `@MX:REASON` 必填标签列表(默认:WARN、DEBT) |

> 不要把某个语言降级为 "PRIMARY",也不要只把一部分设为 "enabled" —— 16 个语言全部平级。

相关: [@MX 标签协议](/zh/advanced/mx-tag-protocol), [moai mx](/zh/cli-reference/mx).

## 环境变量覆盖

部分 YAML section 值可以用环境变量覆盖。环境变量优先于文件值(`internal/config/manager.go` 的 `applyEnvOverrides`)。

| 环境变量 | 目标 | 说明 |
|----------|------|------|
| `MOAI_DEVELOPMENT_MODE` | `constitution.development_mode` | 强制为 `tdd`/`ddd`/`off` |
| `MOAI_LOG_LEVEL` | 日志级别 | `debug`/`info`/`warn`/`error` |
| `MOAI_LOG_FORMAT` | 日志格式 | `text`/`json` |
| `MOAI_NO_COLOR` | 颜色输出 | 设为 `1`/`true` 强制关闭颜色 |
| `MOAI_CONFIG_DIR` | config 目录位置 | 用其他路径替代 `.moai/config/` |

> 以上 5 个就是 config manager 实际会读取的环境变量覆盖的全部。常量定义在 `internal/config/envkeys.go`。`MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG` 目前未实现 —— 名称与对话语言只从 `user.yaml`/`language.yaml` 读取。

## 相关文档

- [settings.json指南](/zh/advanced/settings-json) — Claude Code运行时设置
- [Harness配置与评估](/zh/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/zh/cli-reference/doctor) — `moai doctor config` 用于合并配置检查
