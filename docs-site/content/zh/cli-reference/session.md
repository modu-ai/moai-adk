---
title: moai session 会话注册表
weight: 65
draft: false
---

`moai session` 管理 `.moai/state/active-sessions.json` 中的多会话协调注册表。它用于缓解多个 Claude Code 会话同时在同一项目上工作时产生的竞争(race)。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai session register <session_id> <spec_id> <phase>` | 注册新的活动会话 |
| `moai session heartbeat <session_id>` | 更新既有会话的 last_heartbeat(idempotent) |
| `moai session deregister <session_id>` | 移除会话(idempotent) |
| `moai session list` | 列出活动会话(可用 `--filter-spec` 过滤) |
| `moai session purge` | 移除 stale 项(默认:最后一次 heartbeat 后超过 30 分钟) |
| `moai session current` | 输出此编排器的会话 UUID |
| `moai session doctor` | 诊断注册表为空的原因 |

大多数子命令支持 `--json` 标志以输出机器可读结果。

## moai session list

```bash
moai session list
moai session list --filter-spec SPEC-AUTH-001
```

| 标志 | 说明 |
|--------|------|
| `--json` | 机器可读 JSON 输出(编排器 pre-spawn 检查格式) |
| `--filter-spec <id>` | 仅返回与该 spec_id 匹配的项 |

## moai session purge

```bash
moai session purge
```

| 标志 | 说明 |
|--------|------|
| `--json` | JSON 输出 |
| `--threshold-minutes <n>` | stale heartbeat 截止(分钟,默认 30) |

## moai session current

```bash
moai session current
```

输出编排器自身的会话 UUID。若运行时未暴露会话 ID,则返回 canonical fallback 字符串。

| 标志 | 说明 |
|--------|------|
| `--json` | JSON 输出 |
| `--show-fallback` | 仅输出 canonical fallback 字符串(用于生成 paste-ready resume) |

## moai session doctor

```bash
moai session doctor
```

诊断多会话协调注册表为空的原因(write-path 诊断)。

| 标志 | 说明 |
|--------|------|
| `--json` | JSON 输出 |

## 使用场景

此注册表用于编排器在 spawn 实现 agent 之前检测并发会话竞争。若 `moai session list --json --filter-spec <SPEC-ID>` 返回其他会话的项,编排器会停止推进并向用户确认。

## 相关文档

- [moai inventory](/zh/cli-reference/inventory) —— 会话·工作树·harness 统合查询
- [CLI 概览](/zh/getting-started/cli)
