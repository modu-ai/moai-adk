---
title: 配置章节参考
weight: 71
draft: false
description: ".moai/config/sections/ 的主要配置文件键参考 (handoff/delegation/llm/statusline/security/workflow/crosssession)。"
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
      high: "glm-5.3"          # 1M context — Opus 插槽
      medium: "glm-5.3"        # 1M context   — Sonnet 插槽
      low: "glm-5.3"          # 1M context   — 轻量插槽
      fable: "glm-5.3"
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

控制statusline主题、`github` 段所统计的托管服务，以及16个段切换。

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  # forge: gitlab             # github | gitlab | none（不填则按 origin 主机判定）
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
| `forge` | 取 `github` · `gitlab` · `none` 之一，决定 `github` 段在哪个托管服务上统计未完成的工作。留空则按 origin 远程的主机判定——`github.com` 选 `gh`，`gitlab.com` 选 `glab`。自建实例的名字里没有线索，必须显式指定。无法识别的取值不会退回判定，而是什么都不渲染——于是拼写错误表现为该段缺失，而不是错误的数字 |
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

## workflow.yaml — branch_guard

守护主检出（primary checkout）分支状态的可选加入式防护。当多个会话共用一个检出时，其中一方执行的 `git switch` · `git checkout` · `git reset --hard` · `git stash` · `git rebase` 会在毫无提示的情况下改变另一个会话的工作树。该防护只在主检出中拒绝这些命令。

```yaml
workflow:
    branch_guard:
        enabled: false   # 发布默认值
```

| 键 | 值 | 说明 |
|----|-----|------|
| `enabled` | `false`（默认） | 防护完全处于惰性状态。连用于判定的 `git rev-parse` 都不会执行，因此没有附带开销 |
| `enabled` | `true` | 在主检出中拒绝改变分支状态的命令。在工作树内部照常允许 |

**默认关闭的原因。** 该防护所应对的风险只在多个会话共用一个检出时才会出现。单人使用的仓库不会遇到这个问题，因此发布版本以关闭状态出厂。同时运行多个会话的仓库维护者自行写入上面的键将其开启。

**作用范围。** 防护会区分主检出与工作树，不会阻拦工作树内部的分支操作。`git status` · `git log` · `git diff` · `git fetch` 这类只读命令，以及 `git stash list` · `git merge-base`，即使在开启时也会通过。

**例外与失败方向。** 需要创建分支的 git 负责代理按身份获得豁免，也可以通过 `MOAI_BRANCH_GUARD_EXEMPT=1` 环境变量绕过。判定不确定时（不是仓库、`git rev-parse` 执行失败等）不做拒绝而是放行，只写入审计日志 — 只有在证据确凿时才拒绝。

需要切换分支的工作，正统做法是移到工作树而不是让它被拒绝。具体步骤请参阅 [moai worktree](/zh/cli-reference/worktree/)。

## workflow.yaml — audit

指定跨模型审计后端（`codex_audit` · `glm_audit` · `audit_multi`）实际以哪个模型和 effort 运行。每个后端取一组 `{model, effort}`，发布默认值全部为空。

```yaml
workflow:
    audit:
        codex:
            model: ""   # 如 gpt-5.6-sol — codex 可提供的模型 id
            effort: ""  # 如 high — low | medium | high | xhigh | max
        glm:
            model: ""   # 如 glm-5.3
            effort: ""  # 如 max — low | high | max（z.ai 的 reasoning 状态名）
```

| 键 | 说明 |
|----|------|
| `audit.codex.{model, effort}` | codex 审计在会话开启与评审轮次上携带的一组。model 为空或 codex 无法提供时，丢弃该固定值并回到原有 SSOT 解析 |
| `audit.glm.{model, effort}` | GLM 审计在 z.ai 请求上携带的一组。effort 只接受 z.ai 的 reasoning 状态名 `low` · `high` · `max`；其他值会去掉 reasoning 指令，仅应用模型固定值 |
| 两组均为空 | 解析行为与该键出现之前完全一致 — 从未写入固定值的项目不会有任何变化 |

固定值只作用于审计入口。任务委派路径（`codex_task` · `glm_task`）的模型解析不受影响，同样的字段也可以在网页控制台的 Audit 面板中直接编辑。

## workflow.yaml — todo

关掉待办队列(todo)**引导表面**的开关。会话启动时的队列摘要行、状态栏的 TODO 段、把自然语言请求推理路由到 todo 工作流 —— 这三样在这一个键下一起安静下来。

```yaml
workflow:
    todo:
        enabled: false   # 显式关闭 —— 键不存在时读作开启
```

| 键 | 值 | 含义 |
|----|-----|------|
| `todo.enabled` | (键不存在,默认) | 开启。发行模板里根本没有这个块,大多数项目就处在这个状态 |
| `todo.enabled` | `false` | 会话启动摘要、状态栏 TODO 段、技能的自动路由全部关闭 |

**关掉并不删除命令。** `moai todo` CLI 保持注册、所有动词照常工作,按名字直接调用的 `/moai todo` 也照常执行。这条边界是有意为之 —— 开关只让"没要队列的人看到队列"的那些表面安静下来,不碰真正在用队列的人的路径。读取不到配置文件的失败路径也解析为开启(fail-open)。

在小型一次性项目里,每次会话都弹出的待办摘要读起来像噪音时,就用这个键。队列本身的用法在 [moai todo](/zh/utility-commands/moai-todo/) 页面。

## crosssession.yaml — 会话间消息

决定本会话如何对待来自你其他 Claude Code 会话的消息。`moai cc` · `moai glm` · `moai cg` 启动器会在启动时把这些取值写入一个临时的 `--settings` 文件，Web 控制台则通过设置 seam 编辑本文件。不经启动器、直接用 `claude` 起的会话不会读取本文件。

```yaml
crosssession:
  inbound: ""             # "" | accept | hold | refuse
  isolate_machines: false # 默认——消息离开本机前不要求批准
  dialog_expiry: ""       # "" | 60s | 5m | 10m | never
```

| 键 | 取值 | 说明 |
|----|------|------|
| `inbound` | `""`（默认） | 由 Claude Code 依据两个会话的权限模式类别逐条判断 |
| `inbound` | `accept` | 直接投递消息。要让无人值守的工作会话能收到消息，就需要这个取值 |
| `inbound` | `hold` | 每条消息都扣下等待批准。这样扣下的消息不会过期，等到后来适用 `accept` 时才投递 |
| `inbound` | `refuse` | 丢弃消息 |
| `isolate_machines` | `false`（默认） | **消息发往本机之外的会话时不要求批准。** 同一台机器上的会话之间无论取值如何都不会离开本机，但发往机外会话的消息要经过 Anthropic 服务器——默认就是在无批准的情况下允许这条路径，是否保留默认值请自行判断 |
| `isolate_machines` | `true` | 消息离开本机前必须获得你的明确批准（即使在 `bypassPermissions` 模式下）。任何一个设置作用域写了 `true` 即生效，因此签入仓库的项目文件能打开这项要求却关不掉——要关掉必须把写了 `true` 的每个作用域都清除 |
| `dialog_expiry` | `""`（默认） | 沿用 Claude Code 的五分钟默认值 |
| `dialog_expiry` | `60s` · `5m` · `10m` · `never` | **按默认判断**扣下的消息，其批准对话框的期限；`never` 会一直扣到会话结束。对以显式 `inbound: hold` 扣下的消息不适用 |

## 相关文档

- [settings.json指南](/zh/advanced/settings-json) — Claude Code运行时设置
- [Harness配置与评估](/zh/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/zh/cli-reference/doctor) — `moai doctor config` 用于合并配置检查
