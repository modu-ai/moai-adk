---
title: 配置矩阵
weight: 4
draft: false
---

MoAI-ADK 通过单一的 **配置矩阵**将保留的 11 个代理各自映射到 `{model, effort}` 对。活动的 **配置文件**(`high` / `medium` / `low`)选择矩阵的一列(column)，该列的值应用于所有子代理 spawn。矩阵以代理名为单位共 **33 格**(11 个代理 × 3 个配置文件)，同时取代了此前的分组抽象与 `plan_type × tier` 轴。

## 配置文件轴

配置文件有三个取值:

- `high` — 质量优先列。推理与审计行由 Fable 5 承担，编码由 Opus 5 以 `xhigh` 承担(厂商针对编码与代理式工作推荐的起点)。
- `medium`(默认) — 平衡列。Opus 5 运行在厂商 API 默认 effort `high` 上，因此是最可预测的运行点。取值缺失或为空时按 `medium` 解释。
- `low` — 经济列。Opus 5 的 `low`/`medium` effort 是首要的 token 成本杠杆，之后才降到 Sonnet 5。

`max` 是 `high` 的**只读别名**。既有配置中的 `profile: max` 仍解析为 `high`，保存时始终写入规范名 `high`。无需任何迁移操作。

配置文件与 `performance_tier` 并非两个字段，而是同一个轴 — `llm.profile` 优先，缺失时读取 legacy `performance_tier` 作为别名。两个字段共享 `high`/`medium`/`low` 词汇。解析器读取该有效配置文件来确定每个代理的格子。

## 设置配置文件

```bash
moai init . --profile high             # 初始化时设置
moai update --profile low              # 事后切换
```

允许值为 `high` / `medium` / `low`，legacy 的 `max` 也可作为输入并规范化为 `high`。当前值可在 `.moai/config/sections/llm.yaml` 的 `llm.profile` 字段中查看。

## 配置矩阵

保留的 11 个代理直接从下表矩阵中获得各自的 `{model, effort}`。只有用户自行添加的代理才解析为 `inherit`(继承父会话模型)并被排除在 model 注入之外。矩阵中任何位置都没有 Haiku。

| 代理 | high | medium(默认) | low |
|---|---|---|---|
| manager-spec | fable / xhigh | opus / high | opus / low |
| plan-auditor | fable / xhigh | opus / high | opus / low |
| sync-auditor | fable / xhigh | opus / high | opus / low |
| manager-develop | opus / xhigh | opus / high | sonnet / medium |
| super-advisor | opus / xhigh | opus / high | opus / medium |
| manager-design | fable / high | opus / medium | sonnet / medium |
| builder-harness | opus / xhigh | opus / medium | sonnet / medium |
| e2e-tester | fable / high | opus / medium | sonnet / medium |
| manager-docs | sonnet / high | sonnet / medium | sonnet / medium |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / low | sonnet / low | sonnet / low |

`manager-git` 与 `Explore` 行与配置文件无关，固定为 `sonnet / low` — 机械性工作与只读探索不会因配置文件上升而提升模型等级。

每一行都是单调(monotone)的: `high` ≥ `medium` ≥ `low`。降低配置文件时，任何代理都不会获得比之前更强的组合。

Anthropic 内置的 `Explore` 不再解析为 `inherit`，而是解析为自身的格子(`sonnet / low`)。`inherit` 哨兵值现在只保留给用户添加的代理。

## 骨架专家的 model + effort

`/moai:harness` 生成的专家**模型统一为 `opus`**，**仅以 effort 区分**。骨架代理是用户所有的持久化专家，区分它们的轴是推理深度而非模型等级。由于所有非 Haiku 模型现已具备 1M 上下文，固定模型不会损失上下文。

effort 从各目的类别对应的保留代理行借用:

| 目的类别 | effort 来源行 | high | medium | low |
|---|---|---|---|---|
| `read-only-extract` | Explore | opus / low | opus / low | opus / low |
| `mechanical-transform` | manager-git | opus / low | opus / low | opus / low |
| `synthesize` | manager-docs | opus / high | opus / medium | opus / medium |
| `research` | plan-auditor | opus / xhigh | opus / high | opus / low |
| `verify-judge` | sync-auditor | opus / xhigh | opus / high | opus / low |
| `implement` | manager-develop | opus / xhigh | opus / high | opus / medium |
| `design-architecture` | manager-design | opus / high | opus / medium | opus / medium |

可通过 `llm.harness_agents[配置文件][类别].effort` 覆盖各类别的 effort。模型在任何路径下都不会改变。无法识别的类别回退到 `implement`。

## 解析器优先级

每个代理的有效 `{model, effort}` 按以下顺序确定:

1. 若存在 `llm.agent_overrides[agent]`，则其胜出。
2. 否则使用活动配置文件的代理格子(config `llm.profiles`)。
3. 若 config 中没有该格子，则使用 Go 默认矩阵(`template.DefaultProfileMatrix`)的代理格子。
4. 矩阵中不存在的代理(用户添加)为 `inherit`(不注入)。

`agent_overrides` 以规范代理名为键，并针对目录 + enum 进行校验:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: xhigh }
```

**model** 与 **effort** 的消费路径不同。解析出的 **model** 是编排器在 spawn 时作为 `Agent(model: <alias>)` 运行时参数注入的值(`[1m]`-safe，与 frontmatter 的 `model:` 字段无关)。代理 `.md` 的 frontmatter 保持 `model: inherit`，init/update/web 保存不会改变它。解析出的 **effort** 是对 NAMED 子代理的*文档化意图* — Agent/Task 工具对 named 子代理不接受 per-spawn 的 effort 参数，因此 effort 仅通过 (a) 代理 frontmatter 的 effort 默认值、(b) GLM effort 覆盖层、(c) Workflow / `Agent(general-purpose)` 的提示词级 steering 被消费。

## moai model profile

活动配置文件下解析出的各代理 model+effort，可通过只读访问器查看:

```bash
moai model profile          # 人类可读表格
moai model profile --json   # 机器可读
```

该命令不改变任何内容 — 它原样暴露编排器在 spawn 时将注入的值。

## GLM 后端 effort 覆盖层

{{< icon warning warn >}} **诚实声明**: GLM 后端的 effort 覆盖层处于**已实现 + 已接线**状态，但 wire 有效性(实时有效性)尚待实证 — 不表述为"行为保证"。

在 GLM 后端(`moai glm` / `moai cg` 的 GLM 面板)上，会在配置矩阵之上应用一层覆盖:

- 模型槽位映射: `fable` → `glm-5.2`(Fable 槽位，`ANTHROPIC_DEFAULT_FABLE_MODEL`)
- 将 Claude 的 5 级 effort collapse 到 z.ai 可达的 3-state:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max`(legacy effort 值) → **reasoning-max**
  - (无法识别的值 → reasoning-max，防止推理不足)
- coding-max override: `manager-develop` 无论 collapse 结果如何都强制为 **reasoning-max**
- `manager-git` 在 low effort 下 → **thinking-off**

z.ai 是否通过 Anthropic-compat shim 实际消费 `ANTHROPIC_REASONING_EFFORT` 值，是需要实时 GLM 会话出站观测的实证课题。运行时 SSOT 为 `internal/template/glm_effort_overlay.go`。

## 下一步

- [3 层代理架构](/zh/advanced/no-haiku-3tier/) — DeepSWE 排行榜依据与 3 层定义
- [代币经济概览](/zh/advanced/tokenomics-overview/) — 4 层代币经济结构的 B 层路由
- [模型策略](/zh/multi-llm/model-policy/) — performance_tier 别名与 GLM 后端详情
