---
title: moai epic 史诗进度
weight: 100
draft: false
description: "从磁盘计算跨多个 SPEC 的史诗里程碑进度的 moai epic status 命令。"
---

`moai epic status` **只读取磁盘上已有的内容**来算出一个史诗进行到了哪里。它不另设进度存储，每次都遍历 SPEC 文档重新构建里程碑映射。它是只读的，不会修改任何文件。

一个史诗被拆成多个 SPEC 之后，「已经完成了几个」很快就说不清了。与其逐个打开 SPEC 查看 status，不如让这条命令用一行给出答案。

## 概述

```bash
moai epic status <prefix> [OPTIONS]
```

`<prefix>` 是必填参数，即标识该史诗的 SPEC-ID 前缀。例如传入 `KANBAN` 时，作用对象是 `.moai/specs/SPEC-KANBAN-*/spec.md`。

## 读取什么

里程碑映射由三部分信息构成。

1. **SPEC frontmatter** — 各个 `spec.md` 中的 `status` 取值，是判断里程碑是否完成的依据。
2. **标题中的里程碑标记** — 写在 SPEC 标题里的 `(TOKEN Mx)` 形式的标记，把某个 SPEC 与某个里程碑关联起来。
3. **设计报告（可选）** — 若存在，则作为规范里程碑清单的来源。它会被自动发现，也可用 `--design-report` 显式指定。

没有标记的 SPEC 不会被丢弃，而是单独作为 `untracked_specs` 报告 — 悄悄略去会被读作「那个 SPEC 不存在」。

## 标志

| 标志 | 说明 |
|------|------|
| `--json` | 向 stdout 输出固定形状的 JSON |
| `--design-report <path>` | 用显式路径替代设计报告的自动发现 |
| `--marker <token>` | 显式指定推断出的史诗令牌（例如 `BAS`） |
| `--base-dir <path>` | 项目根目录（默认值：当前工作目录） |

## 示例

默认输出是供人阅读的进度看板。

```bash
$ moai epic status KANBAN
🎯 KANBAN ▓▓▓▓▓░░░░░ 2/4 (50%)
Epic progress:   KANBAN
  🟢 M0 M0                            SPEC-KANBAN-RENAME-001 (completed)
  ⬜ M1 M1                             SPEC-KANBAN-BOOTSTRAP-001 (draft)
  ⬜ M2 M2                             SPEC-KANBAN-WORKTREE-001 (draft)
  🟢 M3 M3                            SPEC-KANBAN-BOARD-001 (completed)
```

若一个标记都没有，它会照实写明，并改为列出匹配到的 SPEC。

```bash
$ moai epic status DESIGN-DOCS
🎯 DESIGN-DOCS — 2 SPEC(s) matched, none carrying a (TOKEN Mx) milestone marker
Epic progress:   DESIGN-DOCS
untracked_specs: SPEC-DESIGN-DOCS-001, SPEC-DESIGN-DOCS-V31-001
```

`--json` 输出脚本可以依赖的固定形状。

```bash
$ moai epic status KANBAN --json
{
  "epic": "KANBAN",
  "epic_token": "KANBAN",
  "milestones": [
    {
      "id": "M0",
      "label": "M0",
      "status": "done",
      "covered": true,
      "spec_id": "SPEC-KANBAN-RENAME-001",
      "spec_status": "completed",
      "sync_commit_sha": "144573336d07da19f4b8a50aa26c38db2704afb5"
    }
  ],
  "done": 2,
  "total": 4,
  "pct": 50,
  "extra_mx": [],
  "untracked_specs": ["SPEC-KANBAN-TODO-CLI-001"],
  "baseline_attribution": "3b9b3bf9959669c4bfc43da313e25bca61f910a2"
}
```

## JSON 字段

| 字段 | 说明 |
|------|------|
| `epic` | 作为参数传入的前缀 |
| `epic_token` | 从标题标记中找到的史诗令牌；未找到时为空字符串 |
| `milestones` | 里程碑数组。每一项包含 id、标签、状态、负责的 SPEC、该 SPEC 的 status、sync 提交 SHA |
| `done` / `total` / `pct` | 完成数、总数、百分比 |
| `extra_mx` | 带有标记但不在规范清单中的里程碑 |
| `untracked_specs` | 匹配了前缀但没有里程碑标记的 SPEC |
| `baseline_attribution` | 计算时刻的 git 提交 SHA |

之所以一并给出 `baseline_attribution`，是为了留下这个结果**读取的是哪一时刻的代码树**。只记下进度数字，就无从知道它是什么时候测的。

## 只读

这条命令只做观测。它不改变任何 SPEC 的 status，不把进度保存到任何地方，每次运行都从磁盘重新计算。正因如此，它的结果不会与实际文件产生偏差。

计算结果同样会作为史诗面板出现在网页控制台的监视区域。请参阅 [MoAI Web Console](/zh/advanced/moai-web-console/)。

---

相关: [moai spec 文档管理](/zh/cli-reference/spec/) · [MoAI Web Console](/zh/advanced/moai-web-console/)
