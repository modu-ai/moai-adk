---
title: 代币预算管理与优雅停止
weight: 2
draft: false
---

深入讨论代币经济学四层结构中的 D 层 — 预算防御(Budget defense)。涵盖当代理到达上下文窗口极限时会话如何无损停止、保存进度使下一个会话可以接续的优雅中止(graceful abort)机制。

## 预算防御的必要性

Anthropic SSE 流在上下文窗口天花板附近会间歇性停顿(`stream_idle_partial`)。这是概率性的但在阈值以上可预测。停顿发生时代理调用可能在流中途失败，从而丢失进度。

预算防御主动解决这个问题。在上下文使用量到达阈值前，系统执行优雅中止，确保会话无损地转移到下一步。

## 各模型的上下文阈值

操作阈值因模型而异。更大的窗口允许更高的百分比利用率；更小的窗口在百分比上较晚到达操作天花板但绝对余量更少。

| 模型类别 | 窗口 | 交接阈值 | 绝对天花板 |
|---------|------|---------|-----------|
| Opus 4.8 (1M) | 1,000,000 代币 | 50% | ~500,000 代币 |
| GLM-5.2 (1M) | 1,000,000 代币 | 50% | ~500,000 代币 |
| Opus / Fable (256K) | 256,000 代币 | 90% | ~230,000 代币 |
| Sonnet / Opus 标准 (200K) | 200,000 代币 | 90% | ~180,000 代币 |
| Haiku (200K) | 200,000 代币 | 90% | ~180,000 代币 |

GLM-5.2(通过 `moai glm` / `moai cg` GLM 面板)是 1M 上下文模型，以 50% 阈值运作。Claude Code 根据 Claude 插槽(Opus=1M, Sonnet/Haiku=200K)报告 `context_window_size`，因此 GLM 会话中原始 telemetry 可能显示 ~180K；MoAI 将其校正为 1M。请信任 statusline 的 CW% 表盘。

## 两阶段交接标记

statusline 分两阶段在上下文栏中追加 `/clear` 提示。

- {{< icon warning warn >}} **软标记** — statusline 上下文栏中显示带有警告色的 `/clear` 提示。在频带的软阈值出现，是允许用户决定是否运行 `/clear` 的建议信号。
- {{< icon warning danger >}} **硬标记** — statusline 上下文栏中显示带有强烈警告色的 `/clear` 提示。在 auto-compact 感知天花板出现，下一个操作必须是 `/clear`。

硬天花板设置在 auto-compact 阈值附近，因此运行时 auto-compact 通常先发，硬标记实际上很少触发。这是 auto-compact 感知公式的有意权衡。

## 优雅中止步骤

由 SPEC-TOKEN-BUDGET-STOP-001 实现的优雅中止机制按以下步骤工作。

1. **检测** — `Tracker.IsAtHardLimit(agentName)` 返回 true(累计使用量 ≥ hard_clear_threshold，默认 0.90)
2. **状态保存** — 将进行中的工作状态持久化到 `progress.md`
3. **发出交接** — 生成可粘贴的 resume 消息(6 块结构)
4. **建议回合结束** — 向用户建议 `/clear`(HARD: 绝不自动 `/clear`)
5. **证据持久化** — 将验证证据持久化到 `.moai/state/verify/`

`/clear` 绝不会被自动执行。系统只建议用户运行 `/clear`；由用户决定。

## 可粘贴 Resume 的 6 块结构

会话交接消息遵循以下 6 块结构。每块的设计使下一个会话能以最少信息继续工作。

```text
✂──── 从此处复制 ────✂

ultrathink. <SPEC-ID> <phase> entering.
applied lessons: <memory-file-1>, <memory-file-2>

Preconditions:
1) <可验证的前提 1>
2) <可验证的前提 2>

Run: <命令或动作>

After merge: <下一个动作或 SPEC>

✂──── 复制到此处 ────✂
```

各块的作用:

- **块 1** — `ultrathink.` 开场白设置 effort:xhigh，声明进入的 phase 和 SPEC-ID
- **块 2** — `applied lessons:` 引用从之前会话学到的 memory 文件(最多 4 个)
- **块 3** — `Preconditions:` 下一个会话在开始前须确认的可验证前提(最多 4 个，每个 ≤200 字符)
- **块 4** — 个别前提项
- **块 5** — `Run:` 单一主要动作(通常是 `/moai <subcommand>`)
- **块 6** — `After merge:` 下一个动作或 SPEC ID

## 验证节食 (verify-diet)

将验证命令的长输出重定向到磁盘并在上下文中只留摘要的文件重定向契约(file-redirect contract)。

规则: 当验证命令的 verbatim 输出超过 **bounded-tail ceiling**(默认 50 行或 2KB，取较小者)时，输出被重定向到文件，上下文中只显示 exit code + bounded tail。

```bash
go test ./... > /tmp/moai-verify/1-go-test.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/1-go-test.log
```

此契约"将 verbatim 证据留在磁盘上，上下文只携带 exit code + bounded tail"。它去除的是双重消耗(内联输出 + 横幅再引用)，不是证据本身。

## 验证证据持久化义务

文件重定向契约写入 `/tmp` 的证据会被操作系统定期清除(macOS 重启、Linux tmpfs 重挂载、systemd-tmpfiles)。当引用的路径不再解析为文件时，审计时无法到达证据。

持久化义务解决这个问题。验证证据必须持久化到 `.moai/state/verify/<session>/` 下。此目录是与 `context-usage.json` 和 `active-sessions.json` 相同的 gitignored 运行时状态区域。

确切的持久化机制(直接写入或 `/tmp` 写入后复制)是实现细节。契约陈述义务: 证据必须在 `/tmp` 清除后仍然存在于可引用、审计时可到达的路径上。

## 下一步

- [代币经济学概述](/zh/advanced/tokenomics-overview/) — 四层结构整体概览
- [三层代理架构](/zh/advanced/no-haiku-3tier/) — Layer B 路由的模型策略基础
