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

该循环是 v3 的第二根支柱 **智能体循环工程** 的通用引擎。goal 状态按会话保存到 `.moai/state/goal/<session-id>.json`(非共享文件),**回合上限(默认 30)** 使循环有界。达到上限时评估器给出 5 段判定(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)并停止阻断。`--max-turns 0` 可启用一种跨越压缩边界持续运行的无限 goal,其实际上限由 `--max-duration`(运行时间)和停滞防护担任,而非回合数。仅 arm `--max-turns 0` 而不带任何实上限会在 arm 时被拒绝(fail-closed)。

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

## `--auto` 任务模式

```bash
moai goal --auto --session <session-id> "在已批准范围内实现并验证功能"
```

`--auto` 不是 goal 条件或 `progression_mode=autonomous` 的别名。任务文本不会作为 shell 执行，也不会交给条件解析器，而是以 `mission_mode=auto`、`state=draft` 单独保存。创建消息明确显示 `approval required`，所以这一步只是**创建任务草案**，并不批准自主执行。

```bash
moai goal approve --scope <路径> --action publish --action commit --completion-evidence <证据> --max-operations 20
moai goal run --action publish --target <gtd-id> --recommend
moai goal run --supervise --card-worktree <WT-路径> --develop-worktree <develop-路径> --governor-receipt <决策.json> --audit-receipt <审计.json> --completion-receipt <完成.json>
moai goal status
moai goal revoke
moai goal resume
```

`approve` 会一次封存目标、范围、允许行为、完成证据和资源上限。之后的 workflow loop 会在每个范围内操作前重新检查当前 snapshot 与 receipt，但不会反复询问同一批准。`status` 读取持久状态；`revoke` 阻止新效果，同时保留进行中效果的协调状态；`resume` 只允许在同一合同下恢复已保存的**已批准、因策略而 blocked**任务，不会批准新范围。

`--recommend` 只是兼容语法，不授予任何权限。真实操作必须同时持有仓库内权限为 `0600` 的 mission-governor 决策 receipt，以及独立审计 PASS receipt；两者绑定任务、合同、snapshot、行为、目标、有效期、签发者、HEAD 与判定为 true 的 typed evidence。`run --supervise` 按 `publish → pick → 带 lease 的磁盘调度 → commit → local develop --no-ff merge` 有界执行封存计划。受监督的 Git 效果必须分别提供 `--card-worktree` 与 `--develop-worktree`；只使用旧 `--repo` 时会以零效果拒绝。完成需要含有合并 ancestry 的 `0600` 完成 receipt，不能仅因动作列表耗尽而完成。遇到 blocked 或 completed 即停止；重放已完成任务的效果数为 0。

批准后，确定性代码仍须在每次操作前检查封存的目标、完成证据、范围、允许行为、资源上限和停止条件。`mission-governor` 只读并提出建议。若继续推进需要扩大范围或增加权限，系统会停止副作用并记录 `blocked`，而不会暗中放宽批准。

`super-advisor` 的意见不具约束力，只读的 `mission-governor` 生成结构化决策。只有确定性 validator 与归属角色 adapter 才能执行效果。提交需要当前 HEAD 的测试 receipt，本地合并需要 manager-git 角色、基准 SHA 与 lease。若持久 runtime 能力尚未证明，则采用 `active-session-only`。远程 batch push、release branch、release PR 与 main merge 的 provider 尚未配置，因此会以 `provider_unsupported` 停止，不会伪装成功。GTD 边界详见 [`/moai gtd`](/zh/utility-commands/moai-gtd)。

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

- [/moai loop - 反复修复循环](/zh/utility-commands/moai-loop)
- [/moai fix - 一次性自动修复](/zh/utility-commands/moai-fix)
- [/moai - 完全自主自动化](/zh/utility-commands/moai)
