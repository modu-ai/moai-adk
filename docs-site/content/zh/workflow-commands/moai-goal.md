---
title: "/moai goal"
weight: 35
draft: false
new: true
added_in: "v3.1"
---

{{< new-badge v3.1 >}}

这是 **条件声明式自治循环**：你只声明结束条件，会话就会在该条件成立之前持续把回合接力下去。每个回合结束时，评估器都会检查条件；条件一旦满足，循环自行停止。你不必在每一步按下"继续"。

{{< callout type="info" >}}
{{< icon flash primary >}} <strong>v3.1 新增</strong>：可无限持续的目标（<code>--max-turns 0</code>）、半自治模式的流程，以及块上限的自动调整，都在本次发布中引入。
{{< /callout >}}

## 本页涵盖的内容

`/moai goal` 向会话注册并武装（arm）一个 **完成条件**。条件语句会被解析成两类谓词：

- **机械条件** (mechanical)：以 shell 命令的退出码判定真假。例如："`go test ./...` exits 0"。
- **模型条件** (model)：以对话记录中是否出现特定行来判定。例如："所有 AC 行都已记录为 PASS"。

注册的条件由 `moai hook stop-goal` Stop-hook 评估器在每个回合结束时读入。条件不成立时，评估器会阻塞（block）当前回合、把下一个回合接上来；条件成立时，循环结束。因此，即便用户少敲一次命令，长周期任务也不会停下，而是自行滚动下去。

## 为什么需要它

在长周期任务 —— 多里程碑 run、大规模重构、TDD 循环 —— 里，最昂贵的开销是 **每个回合都要用户按下"继续"的往返**。一次 `/moai run SPEC-X` 会唤来实现代理，但当这个代理返回时，谁来把下一个回合接上？通常，用户得再输入一遍 prompt。

`/moai goal` 消除了这种往返。只要声明一次条件，评估器就会在每个回合结束时判定"条件尚未满足"，自动开启下一个回合。如果任务的结束条件可以机械验证，用户用一条命令就能把整个循环一路推到底。于是 harness 设计中"最小化用户介入"的原则，靠这一条命令就落到了实处。

不过，**它并非无条件地自行运转**。`/moai goal` 是 arm-only，也就是"只负责注册条件并把回合接上"。条件注册后，真正启动工作的命令（如 `/moai run SPEC-X`）必须与之配对出现。如果只武装了一个孤零零的条件、却不启动工作，每个回合结束时条件都不会满足，回合就只能在原地打转，沦为 idle loop。

## 用法

```bash
# 注册条件 + 武装
> /moai goal "go test ./... exits 0 && lint is clean, or stop after 20 turns"

# 查看状态
> /moai goal status

# 查看所有会话的 goal 状态
> /moai goal status --all

# 中断循环
> /moai goal clear
```

条件语句用双引号包裹。机械条件由评估器在每个回合结束时实际执行，因此快速且确定的命令（`go test -run <pattern>`）比缓慢的全量套件（`go test ./...`）在每个回合上的开销更小。

### 写出好条件的三条原则

1. **一个可测量的终态** —— 测试结果、构建退出码、文件数量、队列是否清空。抽象目标（"代码变好了"）无论对机器还是对模型，都无法判定。
2. **明确测量方式** —— "`go test ./...` exits 0"、"`git status` is clean"。只写"测试通过"，评估器无从知道该执行什么。
3. **把约束也写进去** —— "此时不碰其他测试文件"。只看终态，中途可能改动意料之外的东西。

## 何时使用

`/moai goal` 在以下四种情形下最能发挥价值：

- **T1 —— 长 run-phase / 多里程碑 SPEC（Tier M/L）**：run-phase 自治性接线（`run.md` 中的 `ac_converge` 块）负责这一情形。它会把 run phase 一路接到 SPEC 的所有验收标准（AC）都显示为 PASS 为止。
- **T2 —— 大规模迁移 / 涉及多个调用点的重构**：让循环跑到所有调用点都能编译、测试都通过为止。前提是调用点清单已经在对话中浮现之后再武装。
- **T3 —— TDD 循环 / AC 收敛**：转 RED-GREEN-REFACTOR 循环，把回合一路接到目标测试套件变 green、lint 干净为止。
- **T4 —— `/moai loop` 的替代**：这是"修工具诊断指出的问题"（`/moai loop`）与"一路滚到所声明的终态成真"（`/moai goal`）之间的选择。如果结束点明确，`/moai goal` 更合适。

## 无限持续目标 (v3.1)

用 `--max-turns 0` 关掉回合上限后，循环在触发挂钟时间 / 停滞 guard 之前不会停止。这种形态适合长时间的无人值守 run。

```bash
# 4 小时挂钟上限，回合不限
> /moai goal "<condition>" --max-turns 0 --max-duration 14400
```

不过 Claude Code 运行时的连续阻塞上限（`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`，默认 8）会先于回合上限把循环截断。武装无限 goal 时，这个上限也要一并调高。

```bash
# 把阻塞上限调到 200，让无限循环真正持续下去
$ CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200
> /moai goal "<condition>" --max-turns 0 --max-duration 14400
```

从 v3.1 起，用 `moai goal arm --max-turns 0` 武装、或进入工厂模式时，启动器会自动把这个上限注入为 200。因此用户不必亲自动手改环境变量，4 小时长的链也不会在中途被截断。

## arm-only 与安全边界

`/moai goal` 是 **arm-only**。它只注册条件并把回合接上，本身不会启动工作。因此，始终要与启动工作的命令配对使用。

```bash
# 错误形态 —— 只武装 goal，不启动工作（idle loop）
> /moai goal "<condition>"

# 正确形态 —— 与 run 命令在同一回合武装
> /moai run SPEC-X     # 启动工作
> /moai goal "<condition>"   # 把回合接到 run 结束为止
```

武装好的 goal **不会放松安全边界**：

- **实现启动批准**（plan→run 的人工 gate）仍然是必需的。武装了 goal，并不等于预先批准了进入 run-phase。
- **难以撤销的 / 涉及共享系统的动作**（创建 PR、破坏性运算）的确认边界也照旧。评估器只决定"是否把回合接下去"，不会预先批准破坏性运算。

## 评估成本

每个回合结束时，评估器都会执行机械条件命令。因此每个回合的开销就等于 **条件里那条命令的开销**。用快速且确定的命令，回合循环就更紧凑。Stop-hook 超时是 120 秒，但更快的命令让循环更敏捷。模型条件判定的是已存在的对话记录，不产生额外开销。

## 对比 —— 三种自治连续方式

| 方式 | 由谁来开启下一个回合 | 停止条件 |
|------|------------------|------------|
| **`/moai goal`** | 上一个回合结束后，评估器判定条件未满足 | 条件成立、回合上限、停滞 guard、`/moai goal clear` |
| `/loop`（Claude Code 内置） | 固定时间间隔到了就重新执行 prompt/命令 | 用户取消 |
| `/moai loop` | 诊断扫描生成一个有限的 issue 队列，goal 引擎在每个回合结束时评估"队列已空且诊断干净" | 队列清空且诊断干净，或触达上限 |

`/moai goal` 与 `/moai loop` 是互补关系。`/moai loop` 适合"把工具指出的问题全部修好"，`/moai goal` 适合"一路滚到所声明的终态被证明为真"。两者都跑在同一个 goal 引擎上，区别在于由谁来决定什么是"结束"。

## 与其他命令的关系

- **`/moai loop`** —— 诊断主导的确定性循环。区别在于由工具判定要修什么。[`/moai loop`](/zh/utility-commands/moai-loop) 位于 utility 命令一节。
- **`/moai run`** —— 启动工作的命令。与 goal 配对使用。[`/moai run`](./moai-run)。
- **工厂模式** —— 把 `/moai goal` 的无限持续 goal 串成 `plan → run → verify → sync` 链的进入开关。详见 [工厂模式](/zh/advanced/factory-mode)。

## 这条命令不做的事（范围边界）

- **不会启动工作** —— 它是 arm-only。只武装条件却不启动工作，就会变成 idle loop。
- **不会越过人工 gate** —— 实现启动批准、创建 PR 等难以撤销的决定，仍是用户独有的权限。
- **hooks 被禁用时不生效** —— 一旦 `disableAllHooks` 或 `allowManagedHooksOnly` 打开，Stop-hook 评估器本身就不跑了，会降级为标准的每回合手动流程（优雅降级）。
- **不会与原生 `/goal` 冲突** —— 运行时发出原生 `/goal` 的活跃信号时，MoAI 评估器会让位（避免双重阻塞）。

## 相关文档

- [自治连续循环](/zh/advanced/autonomous-loops) —— goal 引擎的停滞 guard 与上限语义
- [工厂模式](/zh/advanced/factory-mode) —— 由 `factory_chain` goal 预设串起来的四阶段链
- [`/moai loop`](/zh/utility-commands/moai-loop) —— 诊断主导的确定性循环（兄弟命令）
- [Harness 工程](/zh/core-concepts/harness-engineering) —— 循环与观察流向 harness 学习的路径
