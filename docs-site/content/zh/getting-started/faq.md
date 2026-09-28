---
title: 常见问题
weight: 100
draft: false
---

以下是使用 MoAI-ADK 时经常被问到的问题与解答。


---

## Q: `moai` 和 `/moai` 有什么区别？

它们是完全不同的两样东西。这是最常见的混淆，先说清楚。

| | `moai`（终端 CLI） | `/moai`（斜杠子命令） |
|---|---|---|
| **执行位置** | 终端 shell | Claude Code 对话窗口 |
| **本质** | Go 二进制 | Claude Code 技能调用 |
| **用途** | 项目设置、模板部署 | AI 智能体开发工作流 |
| **示例** | `moai init my-project` | `/moai plan "认证功能"` |

- 在终端运行 `moai plan` 不会生效 — `/moai plan` 只在 Claude Code 内有效。
- 在 Claude Code 中输入 `/moai init` 也不会生效 — `moai init` 是终端命令。

---

## Q: statusline 中的版本显示是什么意思？

MoAI statusline 同时显示版本信息与更新提醒：

```text
🗿 v3.1.2 -> 🗿 v3.1.3
```

- **`🗿 v3.1.2`**：当前已安装的版本
- **`-> 🗿 v3.1.3`**：可更新的新版本（用 ASCII 箭头 `->` 相连）

使用最新版本时只显示版本号：

```text
🗿 v3.1.3
```

**更新方法**：执行 `moai update` 后更新提醒会消失。

{{< callout type="info" >}}
**提示**：这与 Claude Code 内置的版本显示（`🔅 v2.1.38`）不同。MoAI 显示追踪的是 MoAI-ADK 版本，Claude Code 会单独显示自己的版本。
{{< /callout >}}

---

## Q: 如何自定义 statusline 显示的分段？

statusline 按段落逐个开关。把各个段落分别打开或关闭，只留下你想看的信息。没有显示预设 —— 配置只有主题和段落两样。

可在 `moai init` 或 `moai update -c` 向导中设置，或直接编辑 `.moai/config/sections/statusline.yaml`：

```yaml
statusline:
  segments:
    model: true
    context: true
    output_style: false
    directory: false
    git_status: true
    claude_version: false
    moai_version: false
    git_branch: true
```

没有 `segments:` 块时，默认启用全部段落。

{{< callout type="info" >}}
详细内容请参考 [SPEC-STATUSLINE-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-STATUSLINE-001/spec.md)。
{{< /callout >}}

---

## Q: 如何选择模型策略？

v3.2 起，逐智能体的模型策略已经没有可选项了。**子代理沿用主会话的模型与推理深度** —— 生成子代理时不传 `model` 也不传 `effort`，MoAI 智能体定义对两者都不作声明。剩下的是会话层面的一个选择：`moai profile setup` 的**会话模型策略**，决定以此配置启动的 Claude 会话的默认推理强度（未单独选择推理强度时生效）。

### 会话模型策略对比

| 值 | 含义 |
|------|------|
| **high** | 会话 effort 回退 `high` |
| **medium**（默认） | 会话 effort 回退 `medium` —— 成本/评分曲线的拐点 |
| **low** | 会话 effort 回退 `low` —— 同一个模型内部的节省用法 |

{{< callout type="warning" >}}
**为什么重要？** 降低 effort 主要降低的是*推理深度*，而不是模型级别。在长时程 agentic 任务中，Opus 的 `low` effort 比任何 effort（包括 `max`）的 Sonnet 评分更高、每任务成本更低 —— 账单由模型完成任务所花的步数决定，而不是按 token 的单价。这份节省的落点现在是会话 effort，旧版本的逐智能体分配表已经退役。
{{< /callout >}}

### 逐智能体分配时代之后的变化

到 v3.1 为止，MoAI-ADK 通过配置矩阵给 13 个目录智能体各自分配 `{model, effort}`，策略层级选出矩阵的一列。这套装置在 SPEC-AGENT-MODEL-INHERIT-001 中退役了 —— 实测显示带 model 参数的调用不足 1%，分配的位置因此移到了会话本身。旧的 `--model-policy`、`--profile`、`--high`、`--medium-alias`、`--low` 旗标以只打印警告、没有任何效果的弃用桩形式保留。

### 设置方法

```bash
# 设置会话模型策略（会话模型策略问题）
moai profile setup

# 既有项目重新设置
moai update -c                # 重新运行设置向导
```

{{< callout type="info" >}}
默认 effort 回退为 `medium`。可在 `moai profile setup` 中修改，或用 `/effort`、`ultrathink` 随时调整会话 effort —— 之后的全部子代理生成都跟随该值。
{{< /callout >}}

---

## Q: 出现了 "Allow external CLAUDE.md file imports?" 警告

打开项目时，Claude Code 可能针对外部文件 import 显示安全提示：

```
External imports:
  /Users/<user>/.moai/config/sections/quality.yaml
  /Users/<user>/.moai/config/sections/user.yaml
  /Users/<user>/.moai/config/sections/language.yaml
```

{{< callout type="info" >}}
**推荐操作：** 请选择 **"No, disable external imports"**。
{{< /callout >}}

**原因：**
- 项目的 `.moai/config/sections/` 中已存在这些文件
- 项目级设置优先于全局设置
- 必需的设置已包含在 CLAUDE.md 文本中
- 禁用外部 import 更安全，且不影响功能

**文件说明：**
- `quality.yaml`：TRUST 5 框架及开发方法论设置
- `language.yaml`：语言设置（对话、注释、提交）
- `user.yaml`：用户名（可选，用于 Co-Authored-By 显示）

---

## Q: TDD 与 DDD 方法论有什么区别？

MoAI-ADK v2.5.0+ 在方法论上**只能二选一**（仅 TDD 或 DDD）。为了清晰与一致性，hybrid 模式已被移除。

TDD 是先写测试、再让测试通过的顺序，因此适合新开发；DDD 则是先用特性测试把现有行为固定下来、再逐步改动，因此适合几乎没有测试的代码。两个循环的分阶段流程请见 [SPEC 驱动开发](/zh/core-concepts/spec-based-dev)与 [DDD](/zh/core-concepts/ddd)。

### 方法论选择表

| 项目状态 | 测试覆盖率 | 推荐方法论 | 理由 |
|--------------|---------------|-------------|------|
| 新项目 | N/A | TDD | 测试优先开发 |
| 现有项目 | 50%+ | TDD | 已有测试基础 |
| 现有项目 | 10-49% | TDD | 可扩展测试 |
| 现有项目 | < 10% | DDD | 需要渐进式特性测试 |

### 设置方法

```bash
# 项目初始化时自动检测
moai init my-project          # 可用 --mode <ddd|tdd> 标志指定

# 手动设置
# 编辑 .moai/config/sections/quality.yaml
development_mode: tdd         # 或 ddd
```


---

## Q: 为什么我的代码里没有 @MX 标签？

这是**完全正常**的。@MX 标签系统的设计目标，是只标记 AI 需要优先关注的最危险、最重要的代码。

| 问题 | 回答 |
|------|------|
| 没有标签是问题吗？ | **不是。** 大部分代码不需要标签。 |
| 标签什么时候添加？ | 仅在**高 fan_in**（调用者 >= 3）、**复杂逻辑**（复杂度 >= 15）、**危险模式**（无 context 的 goroutine）时添加。 |
| 所有项目都类似吗？ | **是。** 所有项目中大部分代码都没有标签。 |

### 标签优先级

| 优先级 | 条件 | 标签类型 |
|---------|------|----------|
| **P1（致命）** | fan_in >= 3 | `@MX:ANCHOR` |
| **P2（危险）** | goroutine、复杂度 >= 15 | `@MX:WARN` |
| **P3（上下文）** | 魔法常量、无 godoc | `@MX:NOTE` |
| **P4（缺失）** | 无测试文件 | `@MX:TODO` |

用 @MX 标签扫描代码库：

```bash
/moai mx --all        # 全部扫描
/moai mx --dry        # 预览
/moai mx --priority P1  # 仅致命项
```

---

## 还有更多问题？

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) — 提问、想法、反馈
- [Issues](https://github.com/modu-ai/moai-adk/issues) — Bug 报告、功能请求
- [Discord 社区](https://discord.gg/Z7E7Mdc5aN) — 实时交流、分享技巧
