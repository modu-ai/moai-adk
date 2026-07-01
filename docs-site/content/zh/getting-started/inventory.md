---
title: moai inventory 命令
weight: 25
draft: false
---

查询项目的活跃会话、worktree、harness 的 `moai inventory` 命令指南。

{{< callout type="info" >}}
**一句话总结**: `moai inventory` 是查询当前项目所有活跃资源（会话、worktree、harness）的读取专用命令。
{{< /callout >}}

## 概述

`moai inventory` 是读取专用命令，为当前项目状态提供**集成库存**。

### 查询对象

| 资源 | 描述 | 位置 |
|------|------|------|
| **活跃会话** | 当前运行的 Claude Code 会话 | `.moai/state/active-sessions.json` |
| **Worktrees** | 项目用 L2/L3 隔离分支 | `~/.moai/worktrees/<project>/` |
| **Harnesses** | 生成的动态代理团队 | `.moai/harness/manifest.json` |
| **SPEC 进度** | 活跃 SPEC 的进度状态 | `.moai/specs/SPEC-*/progress.md` |

## 命令格式

```bash
moai inventory [options]
```

### 基础使用

```bash
moai inventory
```

以基础文本格式输出库存。

### JSON 格式输出

```bash
moai inventory --json
```

以结构化 JSON 输出用于自动分析。

### 筛选

仅查询特定资源类型:

```bash
moai inventory --type sessions
moai inventory --type worktrees
moai inventory --type harnesses
moai inventory --type specs
```

### 详细信息

包含每个资源的附加信息:

```bash
moai inventory --verbose
moai inventory --verbose --json
```

## 文本格式输出

### 基础输出示例

```
MOAI Inventory for moai-adk-go
Project Root: /Users/goos/MoAI/moai-adk-go
Updated: 2026-07-01T10:15:00Z

========== 活跃会话 ==========
Session ID                              分支        SPEC ID            状态
edc25996-04cb-4139-b2f6-c2968e7337db    main        SPEC-DOCS-001      进行中
a1b2c3d4-e5f6-7890-1234-567890abcdef    feat/auth   SPEC-AUTH-002      运行阶段

========== WORKTREES ==========
名称                    分支              创建日期       状态
SPEC-DOCS-001          docs/rebuild      2026-07-01    活跃
SPEC-AUTH-002          feat/auth         2026-07-01    活跃

========== HARNESSES ==========
名称                    版本      团队成员    Worktree 隔离    状态
backend-team            1.0.0     3         L1_optional      活跃
frontend-team           1.0.0     2         无              活跃

========== 活跃 SPECS ==========
SPEC ID                 状态          阶段      拥有者           进度
SPEC-DOCS-001          进行中        运行      manager-develop  M3/6
SPEC-AUTH-002          进行中        运行      manager-develop  M2/5
```

### 详细信息 (`--verbose`)

```
========== 活跃会话 (详细) ==========

会话: edc25996-04cb-4139-b2f6-c2968e7337db
  创建:     2026-06-29T14:30:00Z
  最后更新: 2026-07-01T10:15:00Z
  分支:      main
  SPEC ID:   SPEC-DOCS-001
  状态:      进行中 (运行中 M3)
  上下文:    ~145K / 200K 令牌 (73%)
  模型:      claude-haiku-4-5
  继续:      可用 (.moai/specs/SPEC-DOCS-001/progress.md)

========== WORKTREES (详细) ==========

Worktree: SPEC-DOCS-001
  路径:           ~/.moai/worktrees/moai-adk-go/SPEC-DOCS-001
  基础分支:       main (origin/main)
  创建:           2026-07-01T08:00:00Z
  会话:           edc25996-04cb-4139-b2f6-c2968e7337db
  修改文件:       7
  创建文件:       4
  提交:           2
```

## JSON 格式输出

### 架构

```json
{
  "inventory": {
    "project_root": "/Users/goos/MoAI/moai-adk-go",
    "timestamp": "2026-07-01T10:15:00Z",
    "sessions": [...],
    "worktrees": [...],
    "harnesses": [...],
    "specs": [...]
  }
}
```

### 会话对象

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
  "title": "文档 v3 重建",
  "status": "in-progress",
  "phase": "run",
  "current_milestone": 3,
  "total_milestones": 6,
  "owner": "manager-develop",
  "progress_file": ".moai/specs/SPEC-DOCS-001/progress.md",
  "created_at": "2026-06-20T09:00:00Z"
}
```

## 实用使用示例

### 1. 多会话竞争检测

```bash
moai inventory --type sessions

# 输出中检测到处理同一 SPEC 的会话 > 1 → 竞争风险
```

### 2. Worktree 清理检查

```bash
moai inventory --type worktrees --verbose

# 确认旧 worktree 后进行清理
moai worktree remove <name>
```

### 3. Harness 团队列表查询

```bash
moai inventory --type harnesses --json | jq '.inventory.harnesses[] | {name, teammates, status}'

# 预期输出:
# {
#   "name": "backend-team",
#   "teammates": 3,
#   "status": "active"
# }
```

### 4. 活跃 SPEC 进度追踪

```bash
moai inventory --type specs | grep 进行中

# 查看所有进行中的 SPEC
```

### 5. 自动化脚本中使用

```bash
#!/bin/bash
# Worktree 自动清理脚本

moai inventory --type worktrees --json | jq -r '.inventory.worktrees[] | select(.status == "stale") | .name' | while read name; do
  echo "删除陈旧 worktree: $name"
  moai worktree remove "$name"
done
```

## 输出解释

### Status 字段

| 状态 | 含义 |
|------|------|
| `active` | 当前使用中 |
| `idle` | 暂停 (会话明确暂停状态) |
| `stale` | 未使用 (7 天以上未访问) |
| `error` | 错误状态 (需要检查) |

### Phase 字段

| 阶段 | 描述 |
|------|------|
| `plan` | 计划阶段运行中 |
| `run` | 运行阶段运行中 |
| `sync` | 同步阶段运行中 |
| `completed` | 完成状态 |

## 相关文档

- [SPEC 驱动开发](/workflow-commands/moai-plan) - SPEC 生命周期
- [Worktree 管理](/getting-started/worktree) - Worktree 隔离和生命周期
- [Harness v4 Builder](/advanced/builder-agents) - 动态团队管理
- [CLI 参考](/getting-started/cli) - 其他 CLI 命令

{{< callout type="info" >}}
**提示**: `moai inventory` 可用于自动清理脚本和监控仪表板。使用 JSON 格式进行自动分析，始终掌握项目状态。
{{< /callout >}}
