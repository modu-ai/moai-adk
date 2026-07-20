---
title: moai handoff 交接记录
weight: 68
draft: false
---

`moai handoff` 管理 auto-resume 交接 pending 记录。它保存或清除跨会话边界(`/clear`)继续工作所需的 paste-ready resume 正文。保存的记录在设置 `handoff.mode: auto` 时会在下一次会话开始时自动注入。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai handoff save` | 将 paste-ready resume 正文保存为 pending 记录 |
| `moai handoff clear` | 移除 pending 记录 |

公共标志接受 `--project-dir <path>`(项目根目录,默认:当前目录)。

## moai handoff save

```bash
moai handoff save --stdin --spec SPEC-AUTH-001 --phase run < resume.txt
```

| 标志 | 说明 |
|--------|------|
| `--body <text>` | resume 正文(verbatim 6-block paste-ready) |
| `--stdin` | 从 stdin 读取正文以替代 `--body` |
| `--spec <id>` | 此交接所恢复的 SPEC id |
| `--phase <plan\|run\|sync>` | 阶段 |
| `--session <uuid>` | saved_by_session uuid(归属) |
| `--lang <lang>` | conversation_language 快照 |
| `--ultrathink` | 记录 ultrathink 指令(用于恢复引导) |
| `--ultracode` | 记录 ultracode 指令(用于恢复引导) |
| `--goal <condition>` | 记录 `/goal` 条件(用于恢复引导) |

## moai handoff clear

```bash
moai handoff clear
```

移除 pending 交接记录。

## Fail-open 保证

即使 `moai` CLI 不在 PATH 中,或 `moai handoff save` 以 non-zero 退出,编排器的 paste-ready 输出也保持不变。保存失败绝不会阻止交接的发出,手动 paste 路径无需保存也完全可用 —— 保存只是附加的持久化步骤,不是门禁。

## 相关文档

- [自主连续循环](/zh/advanced/autonomous-loops)
- [moai goal](/zh/cli-reference/goal)
- [CLI 概览](/zh/getting-started/cli)
