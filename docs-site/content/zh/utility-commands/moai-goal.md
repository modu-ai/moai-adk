---
title: /moai goal
weight: 25
draft: false
---

声明完成条件后,会话会自主工作直到该条件满足的 **条件声明型自主循环** 命令。用 `/moai goal "<条件>"` arm 完成条件后,每个回合结束时 `stop-goal` Stop 钩子评估条件是否满足,直到满足为止自动开始下一个回合。

{{< callout type="info" >}}
**一句话概括**:`/moai goal` 是"声明终态的通用循环"。若说 `/moai loop` 是把"直到消除诊断工具找到的全部问题为止"这一条件预先设定好的预设,那么 `/moai goal` 就是 **直接声明** 完成条件的通用引擎。
{{< /callout >}}

{{< callout type="info" >}}
**程序化命令**:原生 Claude Code 的 `/goal` 是只有用户能输入的(HUMAN-ONLY)TUI 命令。`/moai goal` 是把相同语义 **在流水线中程序化** 实现的 MoAI 自有命令,通过 `moai` 技能路由和 `moai goal` CLI 进入。
{{< /callout >}}

## 概述

当你想让智能体"在此条件满足之前自行持续工作"时使用。条件可以混用两种。

- **机械条件(mechanical)**:由 shell 命令验证的条件。例:`go test ./... exits 0`。执行命令并观察退出码。
- **模型评估条件(model-evaluated)**:由对 transcript 的判断验证的条件。例:`所有 AC 行记录为 PASS`。基于会话至今留下的内容进行评估。

该循环是 v3 的第二根支柱 **智能体循环工程** 的通用引擎。goal 状态按会话保存到 `.moai/state/goal/<session-id>.json`(非共享文件),**回合上限(默认 30)** 使循环有界。达到上限时评估器给出 5 段判定(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)并停止阻断。

## 动词(verbs)

### `/moai goal "<条件>"` — 注册 + arm

注册条件文本并对活跃会话 arm goal。条件被解析为 `conditions[]` 数组 —— 纯 shell 命令字符串是机械条件,引用 transcript 的主张是模型条件。arm 后会原子地(temp+rename)写入 `.moai/state/goal/<session-id>.json`,`stop-goal` Stop 钩子会在下一回合结束时读取它并开始评估。

```bash
> /moai goal "go test ./... exits 0;所有 AC 记录为 PASS,或 30 回合后停止"
```

### `/moai goal status [--all]`

输出活跃会话的 goal(或用 `--all` 输出所有会话的 goal)—— 条件文本、conditions 数组、已用回合数 vs 上限、进度日志、生命周期状态(`armed` / `satisfied` / `ceiling-exit` / `cleared`)。

### `/moai goal clear`

解除活跃会话的 goal(删除状态文件)。Stop 钩子见到没有 arm 的 goal 便停止阻断。这是编排器判定模型条件已满足后结束循环的方法。

{{< callout type="info" >}}
**不提供 `resume` 动词。** 以前讨论过的 `resume`(从归档恢复已解除的 goal)动词目前不在 CLI 中 —— `moai goal --help` 只列出 `arm` / `status` / `clear`。因为 `clear` 会 **删除** 状态文件(而非归档为 tombstone),所以不留下可恢复的原件。
{{< /callout >}}

## 进行模式(自主 / 半自主)

编排器执行实施启动批准(plan→run 边界的 `AskUserQuestion`)时,会让用户在与批准/拒绝决定 **相区分的独立轴** 上选择 **自主 vs 半自主** 进行模式。所选模式保存在 goal 状态的 `progression_mode` 字段中(用户不选则默认 `autonomous`)。

| 模式 | 行为 |
|------|------|
| **自主(autonomous,默认)** | 评估器在条件满足或达到上限之前每回合阻断,不会每回合询问用户。与既有 Stop 钩子行为相同。 |
| **半自主(semi-autonomous)** | `stop-goal` 钩子在每个回合边界发出 **检查点信号** 块 JSON,编排器读取它并进行 `AskUserQuestion` 确认轮次(继续 / 解除 goal / 转为自主)。钩子本身绝不调用 `AskUserQuestion`(钩子·子智能体边界 —— 只发出结构化 JSON)。 |

{{< callout type="warning" >}}
**两种模式下批准都是必需的。** 进行模式轴只选择门禁通过 **之后** 做什么 —— 它不是门禁绕过,也不是实施启动批准的放宽。arm 的 goal 在任何模式下都不批准进入 run 阶段、不创建 PR、不执行破坏性操作。
{{< /callout >}}

## 安全不变式

1. **实施启动批准两种模式都必需** —— 进行模式是批准之后的进行选择,而非门禁放宽,且与分数无关地保持。
2. **arm 的 goal 不绕过门禁** —— 不自动创建 PR,不执行破坏性操作。评估器只决定是否继续回合,不预先批准不可逆的操作。
3. **`stop-goal` 钩子不调用 `AskUserQuestion`** —— 只发出结构化 JSON(钩子·子智能体边界)。
4. **停滞守卫(stagnation guard)** —— 检测到连续 N 次无进展的迭代时停止循环,并给出附带 E1/E3 升级说明的 5 段判定。

## goal 条件应当快

评估器在每回合结束时执行。相比完整套件更倾向 `go test -run <pattern>`,相比耗时命令更倾向确定性命令 —— `stop-goal` 的 Stop 钩子超时为 120 秒,但快命令能让回合循环更紧凑。

## 与 /moai loop 的关系

`/moai loop` 是 **goal 引擎之上的预设**。若说 `/moai goal` 是用户直接声明完成条件的通用循环,那么 `/moai loop` 就是把"直到清空诊断工具找到的问题队列为止"这一条件预填好的预设。

| 引擎 | 目标 | 完成条件 |
|------|------|----------|
| `/moai goal` | 条件声明型通用循环 | 满足用户定义的条件式 |
| `/moai loop` | 诊断修复循环(预设) | 清空问题队列 + 诊断干净(0 错误 / 测试通过 / 覆盖率) |

若终态能用条件式表达则用 `/moai goal`,若是"把工具找到的问题全部消除"则 `/moai loop` 更合适。

## 相关文档

- [/moai loop - 反复修复循环](/utility-commands/moai-loop)
- [/moai fix - 一次性自动修复](/utility-commands/moai-fix)
- [/moai - 完全自主自动化](/utility-commands/moai)
