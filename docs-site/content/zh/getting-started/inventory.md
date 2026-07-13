---
title: moai inventory 命令
weight: 25
draft: false
---

本文介绍用于查询项目活跃会话、worktree 与挽具的 `moai inventory` 命令。

{{< callout type="info" >}}
**一句话总结**：`moai inventory` 一目了然地查询当前项目的所有活跃资源（会话、worktree、挽具）。
{{< /callout >}}

## 概述

`moai inventory` 是只读命令，提供当前项目状态的**统一清单**。同时运行多个并行会话和 worktree 时，你需要一个能一次性回答"现在到底有什么在跑？"的地方 — 这个命令就是答案。

### 查询对象

| 资源 | 说明 | 位置 |
|------|------|------|
| **Active Sessions** | 当前正在运行的 Claude Code 会话 | `.moai/state/active-sessions.json` |
| **Worktrees** | 项目用 L2/L3 隔离分支 | `~/.moai/worktrees/<project>/` |
| **Harnesses** | 已生成的动态智能体团队 | `.moai/harness/manifest.json` |
| **SPEC Progress** | 活跃 SPEC 的进度状态 | `.moai/specs/SPEC-*/progress.md` |

## 命令格式

```bash
moai inventory [options]
```

### 基本用法

```bash
moai inventory
```

以默认文本格式输出清单。

### JSON 格式输出

```bash
moai inventory --json
```

以结构化 JSON 输出，可用于自动分析。

### 过滤

只查询特定资源类型：

```bash
moai inventory --type sessions
moai inventory --type worktrees
moai inventory --type harnesses
moai inventory --type specs
```

### 详细信息

包含各资源的附加信息：

```bash
moai inventory --verbose
moai inventory --verbose --json
```

## 文本格式输出

### 基本输出示例

```
MOAI Inventory for moai-adk-go
Project Root: /path/to/your-project
Updated: 2026-07-01T10:15:00Z

========== ACTIVE SESSIONS ==========
Session ID                              Branch        SPEC ID            Status
edc25996-04cb-4139-b2f6-c2968e7337db    main          SPEC-DOCS-001      in-progress
a1b2c3d4-e5f6-7890-1234-567890abcdef    feat/auth     SPEC-AUTH-002      run-phase

========== WORKTREES ==========
Name                    Branch              Created        Status
SPEC-DOCS-001          docs/rebuild        2026-07-01     active
SPEC-AUTH-002          feat/auth            2026-07-01     active

========== HARNESSES ==========
Name                    Version    Teammates    Worktree Isolation    Status
backend-team            1.0.0      3            L1_optional           active
frontend-team           1.0.0      2            none                  active

========== ACTIVE SPECS ==========
SPEC ID                 Status          Phase      Owner           Progress
SPEC-DOCS-001          in-progress     run        manager-develop  M3/6
SPEC-AUTH-002          in-progress     run        manager-develop  M2/5
```

### 详细信息（`--verbose`）

```
========== ACTIVE SESSIONS (VERBOSE) ==========

Session: edc25996-04cb-4139-b2f6-c2968e7337db
  Created:     2026-06-29T14:30:00Z
  Last Update: 2026-07-01T10:15:00Z
  Branch:      main
  SPEC ID:     SPEC-DOCS-001
  Status:      in-progress (running M3)
  Context:     ~145K / 200K tokens (73%)
  Model:       claude-haiku-4-5
  Resume:      available (.moai/specs/SPEC-DOCS-001/progress.md)

========== WORKTREES (VERBOSE) ==========

Worktree: SPEC-DOCS-001
  Path:         ~/.moai/worktrees/moai-adk-go/SPEC-DOCS-001
  Base Branch:  main (origin/main)
  Created:      2026-07-01T08:00:00Z
  Session:      edc25996-04cb-4139-b2f6-c2968e7337db
  Files Modified: 7
  Files Created:  4
  Commits:       2
```

## JSON 格式输出

### Schema

```json
{
  "inventory": {
    "project_root": "/path/to/your-project",
    "timestamp": "2026-07-01T10:15:00Z",
    "sessions": [...],
    "worktrees": [...],
    "harnesses": [...],
    "specs": [...]
  }
}
```

### Session 对象

```json
{
  "session_id": "edc25996-04cb-4139-b2f6-c2968e7337db",
  "created_at": "2026-06-29T14:30:00Z",
  "branch": "main",
  "spec_id": "SPEC-DOCS-001",
  "status": "in-progress",
  "context_usage": {
    "current": 145000,
    "total": 200000,
    "percentage": 72.5
  },
  "model": "claude-haiku-4-5",
  "resume_available": true
}
```

### Worktree 对象

```json
{
  "name": "SPEC-DOCS-001",
  "path": "~/.moai/worktrees/moai-adk-go/SPEC-DOCS-001",
  "base_branch": "main",
  "created_at": "2026-07-01T08:00:00Z",
  "session_id": "edc25996-04cb-4139-b2f6-c2968e7337db",
  "status": "active",
  "files_modified": 7,
  "files_created": 4,
  "commits": 2
}
```

### Harness 对象

```json
{
  "name": "backend-team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "teammates": 3,
  "worktree_isolation": "L1_optional",
  "status": "active",
  "manifest_path": ".moai/harness/manifest.json"
}
```

### SPEC 对象

```json
{
  "spec_id": "SPEC-DOCS-001",
  "title": "Documentation v3 Rebuild",
  "status": "in-progress",
  "phase": "run",
  "current_milestone": 3,
  "total_milestones": 6,
  "owner": "manager-develop",
  "progress_file": ".moai/specs/SPEC-DOCS-001/progress.md",
  "created_at": "2026-06-20T09:00:00Z"
}
```

## 实用示例

### 1. 检测多会话竞争

```bash
moai inventory --type sessions

# 출력에서 같은 SPEC을 다루는 세션 > 1개 감지 → 경합 위험
```

### 2. 确认 worktree 清理

```bash
moai inventory --type worktrees --verbose

# 오래된 worktree 확인 후 정리
moai worktree remove <name>
```

### 3. 查询 Harness 团队列表

```bash
moai inventory --type harnesses --json | jq '.inventory.harnesses[] | {name, teammates, status}'

# 예상 출력:
# {
#   "name": "backend-team",
#   "teammates": 3,
#   "status": "active"
# }
```

### 4. 追踪活跃 SPEC 进度

```bash
moai inventory --type specs | grep in-progress

# 현재 진행 중인 모든 SPEC 확인
```

### 5. 在自动化脚本中使用

```bash
#!/bin/bash
# Worktree 자동 정리 스크립트

moai inventory --type worktrees --json | jq -r '.inventory.worktrees[] | select(.status == "stale") | .name' | while read name; do
  echo "Removing stale worktree: $name"
  moai worktree remove "$name"
done
```

## 输出解读

### Status 字段

| Status | 含义 |
|--------|------|
| `active` | 当前使用中 |
| `idle` | 暂停中（会话被显式置为暂停状态） |
| `stale` | 未使用（7 天以上未访问） |
| `error` | 错误状态（需要确认） |

### Phase 字段

| Phase | 说明 |
|-------|------|
| `plan` | Plan 阶段执行中 |
| `run` | Run 阶段执行中 |
| `sync` | Sync 阶段执行中 |
| `completed` | 已完成 |

## 相关文档

- [基于 SPEC 的开发](/zh/workflow-commands/moai-plan) - SPEC 生命周期
- [Worktree 管理](/zh/getting-started/worktree) - Worktree 隔离与生命周期
- [Harness v4 Builder](/zh/advanced/builder-agents) - 动态团队管理
- [CLI 参考](/zh/getting-started/cli) - 其他 CLI 命令

{{< callout type="info" >}}
**提示**：`moai inventory` 可用于自动清理脚本与监控看板。以 JSON 格式自动分析，就能随时掌握项目状态。
{{< /callout >}}
