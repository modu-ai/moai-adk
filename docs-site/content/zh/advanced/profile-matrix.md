---
title: 配置矩阵
weight: 4
draft: false
---

MoAI-ADK 通过单一的 **配置矩阵**将每个保留代理映射到 `{model, effort}` 对。活动的 **配置文件**(`max` / `medium` / `low`)选择矩阵的一列(column)，该列的值应用于所有子代理 spawn。这个单一的 3 列配置文件轴取代了此前的 `plan_type × tier` 60 格矩阵(SPEC-MODEL-PROFILE-MATRIX-001)。

## 配置文件轴

配置文件有三个值:

- `max` — 最高质量列。在推理节点放置 Fable，在设计·线束·E2E 放置 Opus。
- `medium` (默认) — 平衡列。在推理与执行放置 Opus/high。值为空或缺失时解释为 `medium`。
- `low` — 经济列。以低 effort 放置 Opus，将机械性工作转给 Sonnet。

配置文件不是与 `performance_tier` 分离的字段，而是同一个轴 —— `llm.profile` 优先，缺失时 legacy `performance_tier` 作为别名读取(`high` → `max` 归一化，`max`/`medium`/`low` 保持不变)。解析器读取此有效配置文件以确定每个代理的格。

## 设置配置文件

```bash
moai init . --profile max              # 初始化时设置
moai update --profile low              # 事后切换
```

当前值可在 `.moai/config/sections/llm.yaml` 的 `llm.profile` 字段查看。在 `moai init` 交互式向导中，`high` 回答会归一化为 `max`。

## 配置矩阵

10 个分组代理从下面的矩阵获取 `{model, effort}`。`Explore` 和用户自定义代理没有分组，因此解释为 `inherit`(继承父会话模型)，不是 model 注入对象。矩阵中任何位置都没有 Haiku。

| 代理 (分组) | max | medium (默认) | low |
|---|---|---|---|
| manager-spec (spec_auditors) | fable / medium | opus / high | opus / low |
| plan-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| sync-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| manager-develop (develop) | fable / low | opus / high | opus / medium |
| super-advisor (advisor) | fable / medium | fable / low | opus / high |
| manager-design (design_harness_e2e) | opus / high | opus / medium | opus / low |
| builder-harness (design_harness_e2e) | opus / high | opus / medium | opus / low |
| e2e-tester (design_harness_e2e) | opus / high | opus / medium | opus / low |
| manager-docs (docs) | sonnet / medium | sonnet / medium | sonnet / medium |
| manager-git (git) | sonnet / low | sonnet / low | sonnet / low |
| Explore (—) | inherit | inherit | inherit |

`docs` 和 `git` 行与配置文件无关，固定不变(分别为 sonnet/medium、sonnet/low) —— 机械性工作即使配置文件改变也不会提升模型类别。

## 代理分组

矩阵不是以代理名称为单位，而是以 6 个 **分组**为单位定义。分组 → 代理成员关系如下:

| 分组 | 代理 |
|---|---|
| `spec_auditors` | manager-spec, plan-auditor, sync-auditor |
| `develop` | manager-develop |
| `advisor` | super-advisor |
| `design_harness_e2e` | manager-design, builder-harness, e2e-tester |
| `docs` | manager-docs |
| `git` | manager-git |

`Explore` 和用户添加的代理没有成员关系，因此解释为 `inherit`。

## 解析器优先级

每个代理的有效 `{model, effort}` 按以下顺序确定:

1. 若存在 `llm.agent_overrides[agent]`，则它胜出。
2. 若不存在，使用活动配置文件的分组格(config `llm.profiles`)。
3. 若 config 中没有该格，使用 Go 默认矩阵(`template.DefaultProfileMatrix`)的分组格。
4. 若没有分组成员关系，则为 `inherit`(不注入)。

`agent_overrides` 以规范代理名称为键，并针对目录 + enum 进行校验:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: xhigh }
```

**model** 与 **effort** 的消费路径不同。解析后的 **model** 是编排器在 spawn 时作为 `Agent(model: <alias>)` 运行时参数注入的值(`[1m]`-safe，与 frontmatter `model:` 字段分离)。代理 `.md` frontmatter 保持为 `model: inherit`，init/update/web 保存不会改变它。解析后的 **effort** 是对 NAMED 子代理的 *文档化意图* —— Agent/Task 工具不为 named 子代理接受 per-spawn effort 参数，因此 effort 仅通过 (a) 代理 frontmatter effort 默认值、(b) GLM effort 叠加、(c) Workflow / `Agent(general-purpose)` 提示级 steering 消费。

## moai model profile

以活动配置文件解析出的每个代理的 model+effort，可通过只读访问器确认:

```bash
moai model profile          # 供人阅读的表
moai model profile --json   # 供机器解析
```

该命令不改变任何东西 —— 它原样暴露编排器在 spawn 时将注入的值。

## GLM 后端 effort 叠加

{{< icon warning warn >}} **诚实性声明**: GLM 后端 effort 叠加处于 **已实现 + 已接线**状态，但 wire 有效性(实时有效性)尚待实证 —— 不将其描述为"保证工作"。

在 GLM 后端(`moai glm` / `moai cg` GLM 面板)中，配置矩阵之上会应用叠加:

- 模型插槽映射: `fable` → `glm-5.2` (Fable 插槽, `ANTHROPIC_DEFAULT_FABLE_MODEL`)
- 将 Claude 的 5 步 effort 折叠为 z.ai 可达的 3-state:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max` → **reasoning-max**
  - (无法识别的值 → reasoning-max，防止推理不足)
- coding-max override: `manager-develop` 无论折叠结果如何，强制 **reasoning-max**
- `manager-git` 以 low effort → **thinking-off**

z.ai 是否通过 Anthropic-compat shim 实际消费 `ANTHROPIC_REASONING_EFFORT` 值，是需要实时 GLM 会话出站观测的实证课题。运行时 SSOT 为 `internal/template/glm_effort_overlay.go`。

## 下一步

- [3 层代理架构](/zh/advanced/no-haiku-3tier/) — DeepSWE 排行榜依据与 3 层定义
- [代币经济概览](/zh/advanced/tokenomics-overview/) — 4 层代币经济结构的 B 层路由
- [模型策略](/zh/multi-llm/model-policy/) — performance_tier 别名与 GLM 后端详情
