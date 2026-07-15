---
title: moai inventory 命令
weight: 10
draft: false
---

介绍一目了然地查询当前项目活跃会话、worktree、harness 的 `moai inventory` 命令。

{{< callout type="info" >}}
**一句话概括**:`moai inventory` 以只读方式查询当前项目的活跃资源(会话、worktree、harness)。可用 `--json` 获取结构化输出以用于脚本。
{{< /callout >}}

## 概述

`moai inventory` 是只读命令,在同时运营多个并行会话与 worktree 时,提供一次性确认"现在有什么在运转?"的综合视图。

### 查询对象

| 资源 | 说明 | 数据来源 |
|------|------|------------|
| **Sessions** | 活跃的 Claude Code 会话 | `.moai/state/active-sessions.json` |
| **Worktrees** | 项目用 Git worktree | Git worktree 列表 |
| **Harnesses** | 已注册的 harness | `.moai/harness/` manifest |

## 命令格式

```bash
moai inventory [OPTIONS]
```

### 标志

| 标志 | 说明 |
|------|------|
| `--json` | 结构化 JSON 输出(机器可读) |
| `--project-root <path>` | 项目根路径(默认: 当前目录) |

该命令仅支持以上两个标志。没有过滤或详细模式标志 —— 需要的加工通过 `--json` 输出交给 `jq` 等处理。

## 基本用法

```bash
moai inventory
```

以文本形式输出会话·worktree·harness 摘要。

## JSON 格式输出

```bash
moai inventory --json
```

以结构化 JSON 输出,可用于自动分析或 CI 脚本。

## JSON schema

`--json` 输出的顶层结构由三个部分组成。

```json
{
  "sessions": { ... },
  "worktrees": { ... },
  "harnesses": { ... }
}
```

每个部分都有 `count`、`entries`,以及可选的 `error` 字段。

### Session 条目

```json
{
  "session_id": "edc25996",
  "spec_id": "SPEC-DOCS-001",
  "phase": "run"
}
```

| 字段 | 说明 |
|------|------|
| `session_id` | 会话 ID(短形式,前 8 位) |
| `spec_id` | 关联的 SPEC ID |
| `phase` | 当前 Phase(`plan`, `run`, `sync`, `mx`) |

### Worktree 条目

```json
{
  "branch": "feat/auth",
  "path": "/home/user/.moai/worktrees/project/SPEC-AUTH-001",
  "head": "a1b2c3d4"
}
```

| 字段 | 说明 |
|------|------|
| `branch` | worktree 分支名 |
| `path` | worktree 文件系统路径 |
| `head` | HEAD commit 哈希(短形式,前 8 位) |

### Harness 条目

```json
{
  "name": "backend-team",
  "domain": "backend",
  "manifest_missing": false
}
```

| 字段 | 说明 |
|------|------|
| `name` | harness 名称 |
| `domain` | harness 领域 |
| `manifest_missing` | manifest 文件是否缺失(为 `true` 则设置不完整) |

### 完整输出示例

```json
{
  "sessions": {
    "count": 2,
    "entries": [
      { "session_id": "edc25996", "spec_id": "SPEC-DOCS-001", "phase": "run" },
      { "session_id": "a1b2c3d4", "spec_id": "SPEC-AUTH-002", "phase": "plan" }
    ]
  },
  "worktrees": {
    "count": 1,
    "entries": [
      { "branch": "feat/auth", "path": "/home/user/.moai/worktrees/project/SPEC-AUTH-001", "head": "a1b2c3d4" }
    ]
  },
  "harnesses": {
    "count": 1,
    "entries": [
      { "name": "backend-team", "domain": "backend", "manifest_missing": false }
    ]
  }
}
```

## 实用使用示例

### 1. 检测多会话竞争

同一 SPEC 有 2 个以上会话在处理时存在竞争风险。

```bash
moai inventory --json | jq '[.sessions.entries[] | .spec_id] | group_by(.) | map(select(length > 1))'
```

### 2. 活跃 worktree 分支列表

```bash
moai inventory --json | jq -r '.worktrees.entries[].branch'
```

### 3. 查找缺失 manifest 的 harness

`manifest_missing: true` 的 harness 处于设置不完整状态。

```bash
moai inventory --json | jq '.harnesses.entries[] | select(.manifest_missing)'
```

### 4. 当前进行中的 Phase 分布

```bash
moai inventory --json | jq '[.sessions.entries[].phase] | group_by(.) | map({phase: .[0], count: length})'
```

## 相关文档

- [CLI 参考](./cli) —— 全部 CLI 命令
- [项目状态](./status) —— `moai status` 命令
- [基于 SPEC 的开发](/zh/workflow-commands/moai-plan) —— SPEC 生命周期

{{< callout type="info" >}}
**提示**:`moai inventory --json` 可用于监控仪表盘与 CI 脚本。它是只读命令,可安全地自动化。
{{< /callout >}}
