---
title: "三层智能体架构 (No-Haiku)"
weight: 3
draft: false
---

MoAI-ADK 把 Haiku 从路由模型集合中移除，并把剩下的模型按任务性质分成三档来用。这一页写得能让你讲给朋友听 —— *"把便宜模型塞进繁忙的位置看似省成本，但在长周期的任务上恰恰相反。MoAI 用数据确认了'更常用贵模型反而缩小账单'这一点，并在此基础上定下了按任务种类分三档的分配规则。"*

这里，**智能体** (agent，自主判断并工作的 AI 助手) 完成任务所用的模型与 **effort**（推理深度，决定模型为一次回答思考多深的五档 `low` → `medium` → `high` → `xhigh` → `max`）是决定成本的两根轴。本页讲为什么这两根轴这样分配、为什么干脆移除 Haiku，以及这个"级别"与自主级别是不同的概念。

## 直觉与现实分岔的地方

常见的假设是这样的："把单价便宜的模型（比如 Sonnet · Haiku）派去繁忙的位置，贵模型 (Opus) 只省着用在重要位置，总成本就会降。" 只看代币 (token，模型读写文字的计费单位) 的单价表，这个假设似乎成立。

但这个假设要成立，需要一个前提 —— *弱模型最终也能在同样的次数内完成同样的任务*。在长周期的智能体任务（多次调用工具、自行修正计划、走到头才算完成的任务）上，这个前提崩塌。弱模型收敛失败，失败多少就多转多少步、多喷多少输出代币，任务还是完不成。代币单价虽便宜，*烧掉的代币量*却大得多，每个任务的账单反而更厚。

这是整页的出发点。**决定账单的不是单价，而是完赛效率。**

```mermaid
flowchart TD
    A["完成一个智能体任务"] --> B{"模型足够强吗?"}
    B -- "强 (Opus)" --> C["少量步数收敛"]
    C --> D["输出代币少"]
    D --> E["每任务成本低"]
    B -- "弱 (Sonnet·Haiku)" --> F["收敛失败"]
    F --> G["反而烧掉更多步数与代币"]
    G --> H["每任务成本更高"]
    E --> I["'便宜模型 = 便宜账单'\n命题被翻转"]
    H --> I
```

## 数据怎么说 —— DeepSWE 排行榜

验证上述直觉的出处是 DeepSWE 排行榜的 **"All effort levels"** 视图（113 tasks / 91 repos / 5 languages，mini-swe-agent 线束）。由于 effort 按档分别报告，我们不是盯着一个运行点，而是能从每个模型的成本 · 得分曲线形状推导级别。

| 模型 | effort | 得分 | $/任务 | 输出代币 | 步数 |
|---|---|---|---|---|---|
| Opus 5 | low | 58% | $1.66 | 20k | 36 |
| Opus 5 | medium | 69% | $3.29 | 37k | 52 |
| Opus 5 | high | 73% | $6.08 | 64k | 73 |
| Opus 5 | xhigh | 73% | $9.07 | 92k | 89 |
| Opus 5 | max | 74% | $11.84 | 118k | 99 |
| Sonnet 5 | low | 31% | $2.19 | 36k | 77 |
| Sonnet 5 | medium | 40% | $4.08 | 57k | 108 |
| Sonnet 5 | high | 48% | $7.43 | 87k | 147 |
| Sonnet 5 | xhigh | 50% | $11.89 | 121k | 186 |
| Sonnet 5 | max | 54% | $26.40 | 214k | 268 |
| Fable 5 | high | 69% | $9.18 | 57k | 59 |
| Fable 5 | max | 70% | $21.63 | 119k | 88 |

每 MTok 原价（输入/输出）： Opus 5 $5/$25 · Sonnet 5 $2/$10（导入价，至 2026-08-31，此后 $3/$15） · Fable 5 $10/$50。

{{< icon warning warn >}} **单价倒挂**： Sonnet 的代币单价*低于* Opus，但在所有可比位置上每任务成本反而更高 —— Opus 5 `low` 花 $1.66 拿 58%，Sonnet 5 `max` 花 $26.40 只拿 54%。"改用便宜模型就能省额度"的通念在长周期智能体工作中不成立，因为决定账单的不是单价而是完赛效率。

从数据里读出的结论有四条。

1. **Opus 5 在所有 effort 上帕累托支配 Sonnet 5。** Opus `low`（58%，$1.66）在得分与成本两轴上同时压过 Sonnet 的全部五个点，Sonnet `max`（54%，$26.40）也不例外。"繁忙的智能体派给便宜模型"这条路由命题在长周期智能体工作中被证伪。
2. **原因不是单价而是完赛效率。** Sonnet 完成同一组任务约多花 2.7 倍步数。让任务变贵的不是每代币费率，而是这些多出来的步数与输出代币。
3. **`xhigh` 在 Opus 上是纯损失。** `high` 与 `xhigh` 同为 73%，`xhigh` 成本多 49%、步数多 22%。Fable 在同样的位置也出现天花板变平。越过拐点的 effort 买到的是代币，不是分数。
4. **`medium` 是拐点。** 每得 1 分的边际成本： `low` → `medium` $0.15、`medium` → `high` $0.70（4.7 倍）、`xhigh` → `max` $2.77（18.6 倍）。默认配置把核心实现智能体锚定在 `medium`，原因正在这里。

## 为什么干脆移除 Haiku

连 Sonnet 在长周期任务上都比 Opus 贵，Haiku 比 Sonnet 更弱。把 Haiku 放进路由，能力不会增加，只是步数浪费增加 —— Sonnet 上已经观测到的完赛失败模式，在 Haiku 上只会更陡峭。

所以 MoAI 把 Haiku 从路由模型集合中完全排除（No-Haiku 策略，SPEC-AGENT-ARCH-V2-001 §D）。Haiku 在模型 enum 中仍是合法值，因此会出现在文档 · 示例 YAML 里，但不会进入实际智能体分配矩阵的任何一格。移除 Haiku 之后，降低成本的轴仍在 —— 不改模型等级，而是按档调节 effort（推理深度）。这就是三层结构的起点。

## 三档分配规则

把剩下的模型（Opus、Sonnet）与 effort 按任务性质分成三档。这里的"级别"指*按任务种类分配模型 · effort 的档位*。

```mermaid
flowchart TD
    START["任务进入智能体"] --> Q{"这一行做的是什么?"}
    Q -- "机械处理<br/>或只读探索" --> T1
    Q -- "生产出某种东西" --> T2
    Q -- "判断别人产出的东西<br/>或协调多个行" --> T3

    T1["Tier 1 — 机械 · 探索<br/>Sonnet low<br/>manager-docs · manager-git · Explore"]
    T2["Tier 2 — 生产<br/>Opus，逐行档位不同<br/>manager-spec · manager-develop<br/>builder-harness · e2e-tester"]
    T3["Tier 3 — 判断 · 协调<br/>Opus，以 high 为主<br/>plan-auditor · sync-auditor · manager-design<br/>manager-lead · super-advisor · mission-governor"]

    T1 --> NOTE["三个配置下全部固定"]
    T2 --> NOTE2["两行在三列都固定在 medium<br/>只有两行随配置下降"]
    T3 --> NOTE3["super-advisor · mission-governor<br/>在经济列也保持 high"]
```

### Tier 1 — 机械 · 探索

{{< icon database >}} 按既定步骤照做，或者只读取就结束的工作。成本由输入而非迭代左右，让弱模型变贵的原因——多步完赛失败——在这里不出现。于是 Sonnet 更低的输入单价成为实质变量，用 `low` effort 把步数压到最少。负责的智能体是 `manager-docs`（文档整理）、`manager-git`（提交与 PR 的机械作业）、`Explore`（只读探索）三个，三行在三个配置（经济 · 默认 · 质量）下都固定为 `sonnet / low` —— 提高配置也不会提高它们的模型级别。

### Tier 2 — 生产

{{< icon flash >}} 写规格、实现代码、生成线束、跑 E2E 场景 —— **产出东西**的行。它们是多轮的，完赛效率决定账单；Opus `low` 的得分已经高于任何 effort 的 Sonnet、每任务成本更低，所以默认由 Opus 承担。

配置**并不以同样方式**移动这四行。逐行不同。

| 行 | 质量列 | 默认列 | 经济列 |
|---|---|---|---|
| `manager-spec` | `opus / medium` | `opus / medium` | `opus / medium` |
| `manager-develop` | `opus / medium` | `opus / medium` | `opus / medium` |
| `builder-harness` | `opus / high` | `opus / medium` | `opus / low` |
| `e2e-tester` | `opus / medium` | `opus / low` | `sonnet / low` |

撰写与实现的行 `manager-spec` 和 `manager-develop` **在三列都停在 `medium`** —— 不把开支进一步推向生产端，正是这张矩阵的决定。完整走完三级的只有 `builder-harness` 一行，而 `e2e-tester` 在经济列连模型都降到 Sonnet。

### Tier 3 — 判断 · 协调

{{< icon sparkles >}} **判断**别人产出的东西，或**协调**多个行的位置。这张矩阵的原理一句话就在这里 —— **开支给判断的行，而不是生产的行**，因为一次判断会大幅左右后续成本。

| 行 | 质量列 | 默认列 | 经济列 |
|---|---|---|---|
| `plan-auditor` · `sync-auditor` | `opus / high` | `opus / high` | `opus / medium` |
| `manager-design` · `manager-lead` | `opus / high` | `opus / high` | `opus / medium` |
| `super-advisor` · `mission-governor` | `opus / high` | `opus / high` | `opus / high` |

只有 `super-advisor`（升级通道）与 `mission-governor`（密封任务的判定）**在经济列也保持 `high`**。因为最值得在便宜的一列里保持稳健的，恰恰是这两个位置。

`mission-governor` 说明了这条轴为什么好过"是不是多轮"。它读一次、返回一个决定，是**单发**的行，按多轮标准本该落在 Sonnet 一侧；实际上它三列都是 `opus / high`。**因为它是判断的行。**

`max` **没有任何一行拿到**。它作为 `high` 之上唯一的档位留在词汇里，但当前持有它的格子是 0 个。`xhigh` 也哪里都不用 —— 在 Opus 上得分与 `high` 相同，成本却多 49%。

## 模型级别与自主级别是两回事

本页的"级别"是关于*哪个模型以哪个推理深度分配给哪类工作*的档位。名字相似容易混淆，但 MoAI 里还有一个与它**不同**的"级别"。

{{< callout type="info" >}}
**名字相同、对象不同的两个级别**
- **模型级别**（本页）—— 按任务种类决定智能体用*哪个模型 · 哪个 effort* 工作。对象是**成本 · 质量**。
- **自主级别** (`MOAI_AUTONOMY_TIER`) —— 智能体*在没有人工批准的情况下能自主行动到哪一步*的等级。对象是**权限 · 控制**。
{{< /callout >}}

两者正交。一个智能体用贵模型、高 effort 工作（模型级别高），不代表它能不经人工批准自主行动，反之亦然。自主级别由独立的环境变量与三档模式选择处理，详见[自主级别](/zh/advanced/autonomy-tier/)页面。

## 分开阅读设计意图与已实现的行为

{{< icon warning warn >}} **诚实性区分**： 本页明确区分设计阶段的意图与实际实现的行为。

**设计阶段** (`.moai/reports/agent-architecture-redesign-v2-20260709.html`) —— v2 架构的设计意图。提出三层模型策略的原则与 DeepSWE 依据。

**已实现的行为** —— 实际路由由单一配置矩阵执行。活动配置（`high` / `medium` / `low`）选出矩阵的一列，解析器定下各智能体的 `{model, effort}`，在 spawn 时把 model 作为运行时参数注入。详细矩阵请看[配置矩阵](/zh/advanced/profile-matrix/)页面。

阅读侧同样要把设计意图（本页的 DeepSWE 依据）与已实现的行为（单一配置矩阵）分开看。

## 这个基准测不到的东西

{{< icon info >}} **局限声明**： 这个基准测量的是**编码**智能体。文档撰写、审计判断、SPEC（需求规格书）撰写质量未被直接测量，这些行的安排不是观测，而是建立在"与多轮智能体工作相似"的推断上。置信区间也要一起看 —— `medium`（69%±1）与 `high`（73%±2）不重叠，但 `max`（74%±4）与 `high` 重叠。这正是不把 `max` 分配给任何一格的原因 —— 那等于为重叠的区间多付钱。所有默认值都可以用 `llm.agent_overrides` 按智能体逐个回退。

{{< icon info >}} **关于 Fable 5**： Fable 在编码工作上被全面压制。Fable `high`（69%，$9.18）与 Opus `medium`（69%，$3.29）得分相同，成本近 3 倍。所以没有放进任何矩阵格。它在模型 enum 中仍是合法值，GLM 后端的 Fable 槽位接线也原样保留 —— 变的只是默认值。

## 与线束自我进化的连接

三档分配是线束自我变好的循环的地基。观察 → 反思 → 晋升这条进化循环要尽到本分，观察阶段的路由决策本身就得让合适的模型以合适的 effort 进行。错误的路由会污染观察本身，被污染的观察在晋升阶段产出错误的规则。详见[线束自我进化](/zh/advanced/self-evolving/)页面。

## 下一步

- [配置矩阵](/zh/advanced/profile-matrix/) —— 单一 3 列 per-agent 配置矩阵（13 个智能体 × 3 个配置 = 39 格）
- [自主级别](/zh/advanced/autonomy-tier/) —— 与模型级别正交、以权限 · 控制为对象的自主等级
- [代币经济学概述](/zh/advanced/tokenomics-overview/) —— 四层代币经济学结构的路由层
