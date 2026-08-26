---
title: 配置矩阵
weight: 4
draft: false
---

MoAI-ADK 通过单一的 **配置矩阵**将保留的 12 个代理各自映射到 `{model, effort}` 对。活动的 **配置文件** (`high` / `medium` / `low`)选择矩阵的一列(column)，该列的值应用于所有子代理 spawn。矩阵以代理名为单位共 **36 格** (12 个代理 × 3 个配置文件)，同时取代了此前的分组抽象与 `plan_type × tier` 轴。

## 配置文件轴

配置文件有三个取值:

- `high` — 质量优先列。开销流向"做判断的行"而非"做产出的行": 审计·顾问行(`plan-auditor`、`sync-auditor`、`super-advisor`)与协调行(`manager-design`、`manager-lead`)保持 `high`，而撰写·实现行(`manager-spec`、`manager-develop`)在三列中都停在 `medium`。没有任何行取 `max`。`xhigh` 不出现在任何格子中: 在 Opus 5 上它与 `high` 得分相同，成本却明显更高。
- `medium`(默认) — 平衡列。与 `high` 列恰好只在两行上不同: `builder-harness` 降到 `medium`、`e2e-tester` 降到 `low`。取值缺失或为空时按 `medium` 解释。
- `low` — 经济列。Opus 5 在 `low` 下比任何 effort 的 Sonnet 5 得分更高**且**每任务成本更低，因此所有代理式行都保留 Opus；大多数 Opus 行落在 `medium`，唯独 `super-advisor` 保持 `high` —— 升级路径正是便宜列里最值得保持健全的位置。Sonnet 只出现在单次完成、以输入为主的行上。

`max` 是 `high` 的**只读别名**。既有配置中的 `profile: max` 仍解析为 `high`，保存时始终写入规范名 `high`。无需任何迁移操作。

配置文件与 `performance_tier` 并非两个字段，而是同一个轴 — `llm.profile` 优先，缺失时读取 legacy `performance_tier` 作为别名。两个字段共享 `high`/`medium`/`low` 词汇。解析器读取该有效配置文件来确定每个代理的格子。

## 设置配置文件

```bash
moai init . --profile high             # 初始化时设置
moai update --profile low              # 事后切换
```

允许值为 `high` / `medium` / `low`，legacy 的 `max` 也可作为输入并规范化为 `high`。当前值可在 `.moai/config/sections/llm.yaml` 的 `llm.profile` 字段中查看。

## 配置矩阵

保留的 12 个代理直接从下表矩阵中获得各自的 `{model, effort}`。只有用户自行添加的代理才解析为 `inherit`(继承父会话模型)并被排除在 model 注入之外。矩阵中任何位置都没有 Haiku。

| 代理 | high | medium(默认) | low |
|---|---|---|---|
| manager-spec | opus / medium | opus / medium | opus / medium |
| plan-auditor | opus / high | opus / high | opus / medium |
| sync-auditor | opus / high | opus / high | opus / medium |
| manager-develop | opus / medium | opus / medium | opus / medium |
| super-advisor | opus / high | opus / high | opus / high |
| manager-design | opus / high | opus / high | opus / medium |
| manager-lead | opus / high | opus / high | opus / medium |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |
| manager-docs | sonnet / low | sonnet / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / low | sonnet / low | sonnet / low |

36 个格子的模型分布为 Opus 26 / Sonnet 10。Fable 不出现在任何格子中，也没有任何格子使用 `xhigh` 或 `max`。

`manager-docs`、`manager-git` 与 `Explore` 行与配置文件无关，固定为 `sonnet / low` — 文档整理、机械性工作与只读探索不会因配置文件上升而提升模型等级。

每一行都是单调(monotone)的: `high` ≥ `medium` ≥ `low`。降低配置文件时，任何代理都不会获得比之前更强的组合。

### 这些格子的依据

这些格子不是从成本/得分曲线推导出来的，而是**敲定的运营者判断**(settled operator input)。布局原理只有一条: 开销流向"做判断的行"，不流向"做产出的行"。审计·顾问行(`plan-auditor`、`sync-auditor`、`super-advisor`)与协调行(`manager-design`、`manager-lead`)保持 `high`，撰写·实现行(`manager-spec`、`manager-develop`)在三列中都停在 `medium`，`manager-docs` 降到 `sonnet / low`，没有任何行取 `max`。试图从成本曲线重新推导这些格子会悄悄把产出行的开销推回去 —— 需要改值时，按运营者判断的更新处理，而不是按重新计算处理。

决定模型等级的两条规则以实测为根:

- **Opus 在每个 effort 上都压制 Sonnet。** Opus 5 在 `low`(58%、$1.66/任务、36 步)下得分高于任何级别的 Sonnet 5，每任务成本也更低，包括 `max` 下的 Sonnet 5(54%、$26.40/任务、268 步)。决定每任务成本的是完成效率 — 即完成任务所花的步数与输出 token — 而非单 token 价格。因此 Sonnet 仅保留在多步完成不适用之处: 单次完成、以输入为主的行(`Explore` 检索、`manager-git` 机械性操作)，那里更低的输入价格才是决定性因素。这也是所有多轮代理式行都用 Opus 的原因。
- **`xhigh` 在 Opus 上被严格压制。** `high` 以 $6.08 取得 73%，而 `xhigh` 以 $9.07 取得同样的 73% — 没有收益，成本 +49%，步数 +22%。它已从矩阵中退役(6 格 → 0 格)。`max` 仍作为 `high` 之上唯一的级别留在词汇表中，但当前没有行取它。

{{< icon warning warn >}} **证据的适用范围**: 该基准测试衡量的是*编码*代理。文档撰写、审计判断与 SPEC 撰写质量**并未**被直接测量 — 这些行的位置基于与多轮代理式工作的相似性推断。任何行都可通过 `llm.agent_overrides` 按代理逐一回退。

Anthropic 内置的 `Explore` 不再解析为 `inherit`，而是解析为自身的格子(`sonnet / low`)。`inherit` 哨兵值现在只保留给用户添加的代理。

## 骨架专家的 model + effort

`/moai:harness` 生成的专家**模型统一为 `opus`**，**仅以 effort 区分**。骨架代理是用户所有的持久化专家，区分它们的轴是推理深度而非模型等级。由于所有非 Haiku 模型现已具备 1M 上下文，固定模型不会损失上下文。

effort 从各目的类别对应的保留代理行借用:

| 目的类别 | effort 来源行 | high | medium | low |
|---|---|---|---|---|
| `read-only-extract` | Explore | opus / low | opus / low | opus / low |
| `mechanical-transform` | manager-git | opus / low | opus / low | opus / low |
| `synthesize` | manager-docs | opus / low | opus / low | opus / low |
| `research` | plan-auditor | opus / high | opus / high | opus / medium |
| `verify-judge` | sync-auditor | opus / high | opus / high | opus / medium |
| `implement` | manager-develop | opus / medium | opus / medium | opus / medium |
| `design-architecture` | manager-design | opus / high | opus / high | opus / medium |

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
    manager-develop: { model: opus, effort: high }
```

enum 仍接受 `fable` 作为 model、`xhigh` 作为 effort — 它们只是不出现在默认矩阵中，并未从词汇表中移除，因此覆盖项仍可选择其中任一。

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

- 模型槽位映射: `fable` → `glm-5.3-flash`(Fable 槽位，`ANTHROPIC_DEFAULT_FABLE_MODEL`)。该槽位是 GLM 环境绑定，与配置矩阵无关 — 即使没有任何矩阵格子选择 Fable，它仍保持接线状态。
- Claude 的 5 级 effort collapse 到 z.ai 的 reasoning 上限上。GLM-5.3 **始终推理** — 不支持关闭 reasoning，请求关闭会直接失败 — 所以调节轴只有三档 `reasoning_effort`(low / high / max):
  - `low` → **reasoning-low**
  - `medium` / `high` / `xhigh` / `max` → **reasoning-max**
  - (无法识别的值 → reasoning-max，全称条款: 绝不推理不足)
  - reasoning-high 仍是有效的 wire 值，但没有任何 Claude effort collapse 到它上面
  - 没有显式覆盖的 GLM 会话默认以 **reasoning-max** 运行
- flash 例外: 在 `glm-5.3-flash`（默认模型）上，上面的 collapse 规则不适用 — 包括 `low` 在内的**所有** Claude effort 都固定为 **reasoning-max**，因为 flash 只接受 `reasoning_effort: max`。low/high/max 三档是 glm-5.3 及更早模型的体系。
- coding-max override: `manager-develop` 无论 collapse 结果如何都强制为 **reasoning-max**(z.ai 的"编码任务用 reasoning max"建议)
- `manager-git` 在三个配置文件中都是 `low` effort，占据 reasoning-low 的位置

运行时 SSOT 为 `internal/template/glm_effort_overlay.go`。

## 下一步

- [3 层代理架构](/zh/advanced/no-haiku-3tier/) — DeepSWE 排行榜依据与 3 层定义
- [代币经济概览](/zh/advanced/tokenomics-overview/) — 4 层代币经济结构的 B 层路由
- [模型策略](/zh/multi-llm/model-policy/) — performance_tier 别名与 GLM 后端详情
