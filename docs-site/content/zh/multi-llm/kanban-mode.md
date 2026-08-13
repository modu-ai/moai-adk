---
title: 看板模式
weight: 30
draft: false
---

{{< callout type="info" >}}
看板模式的完整概览和 Origin-Trail Chain 设计方向请看 [看板模式](/zh/advanced/kanban-mode)。本页只讲多会话（主导 + 同伴）的运用流程。
{{< /callout >}}

## 什么是看板模式？

看板模式让一个**主导**会话驱动 `plan -> run -> verify -> sync` 链，
同时四个**同伴**会话加入同一次运行来并行分担工作。运行中的每个会话
（无论主导还是同伴）都获得提升后的 Stop-hook 连续块上限，使会话
中途设置的目标能越过默认连续块上限继续运行。

主导负责播种链，同伴不播种。每个同伴同时携带看板成员标志（`-k`）
和角色标签（`--name <role>-<run-id>`），使调度器正确分类、
SessionStart 钩子宣告其成员身份。

## 入口开关

### 主导入口

```bash
moai cc -k                     # Claude 后端主导
moai cc -k SPEC-AUTH-001       # 绑定到 SPEC 的主导
moai glm -k                    # GLM 后端主导
```

主导会话：

{{< icon check-circle ok >}} 设置 `MOAI_KANBAN` + `MOAI_KANBAN_ID`（链种子）。
{{< icon check-circle ok >}} 在 SessionStart 输出运行 id 和四个同伴启动命令。
{{< icon x-circle danger >}} 不设置 `MOAI_KANBAN_LABEL`（那是同伴的信号）。

### 同伴入口

```bash
moai cc -k --name plan-abc123    # plan 同伴
moai cc -k --name run-abc123     # run 同伴
moai cc -k --name review-abc123  # review 同伴
moai cc -k --name sync-abc123    # sync 同伴
moai glm -k --name run-abc123    # 在 GLM 后端上相同
```

`<run-id>` 是主导在启动时宣告的标识符，四个角色是
`plan`、`run`、`review`、`sync`。

同伴会话：

{{< icon check-circle ok >}} 设置 `MOAI_KANBAN_LABEL`（成员身份 + 角色标签）。
{{< icon check-circle ok >}} 获得与主导相同的提升后 Stop-hook 块上限。
{{< icon x-circle danger >}} 不设置 `MOAI_KANBAN` — 不播种链。

### 无操作（不变的会话）

```bash
moai cc --name mysession         # 无 -k，无看板成员身份
moai cc --name run-abc123        # 同伴形 --name 但无 -k → 无操作
```

没有 `-k` 时，无论 `--name` 形状如何，调度器都不做任何操作。
`--name` 标志原样传递给 Claude。

## 多会话引导流程

```
终端 1（主导）             终端 2-5（同伴）
─────────────────          ────────────────────────
moai cc -k                 moai cc -k --name plan-<run-id>
                           moai cc -k --name run-<run-id>
                           moai cc -k --name review-<run-id>
                           moai cc -k --name sync-<run-id>
```

引导是手动的：一个会话无法启动另一个会话。主导的 SessionStart 通知
打印要复制的四条命令，每个新终端一条。在任何同伴上用 `moai glm` 替换
`moai cc` 即可在 GLM 后端运行。

## 跨会话消息

会话间通信使用 Claude Code 的跨会话消息（`ListAgents` / `SendMessage`）。
`crossSessionInbound` 设置字段控制入站消息是被接受、保留还是拒绝。

看板模式自动接受入站消息：启动器写入一个包含
`{"crossSessionInbound": "accept"}` 的临时设置文件，并通过 `--settings`
传给后端。该文件是会话私有的（退出时清理），不修改你的持久设置。

### 操作者提供的 `--settings`

如果你在命令行传入 `--settings <file>`，启动器不会注入自己的设置文件。
请确认你的文件包含：

```json
{
  "crossSessionInbound": "accept"
}
```

主导的 SessionStart 通知会在启动器未注入时打印提醒。

## SessionStart 通知

主导通知宣告运行 id、四条同伴启动命令、主导套接字路径和入站自动化状态。
同伴通知是确认加入的无角色单行。两者都不会提示，都是信息性的 stdout。
