---
title: moai inventory コマンド
weight: 25
draft: false
---

プロジェクトのアクティブセッション、ワークツリー、ハーネスを照会する `moai inventory` コマンドを案内します。

{{< callout type="info" >}}
**一言要約**: `moai inventory` は、現在のプロジェクトのすべてのアクティブな資源(セッション、ワークツリー、ハーネス)を一目で照会します。
{{< /callout >}}

## 概要

`moai inventory` は読み取り専用のコマンドで、現在のプロジェクト状態の **統合インベントリ** を提供します。並列セッションや worktree を複数走らせていると「今、何が動いているんだっけ?」を一度に確認できる場所が必要になりますが、その答えがこのコマンドです。

### 照会対象

| 資源 | 説明 | 場所 |
|------|------|------|
| **Active Sessions** | 現在実行中の Claude Code セッション | `.moai/state/active-sessions.json` |
| **Worktrees** | プロジェクト用の L2/L3 分離ブランチ | `~/.moai/worktrees/<project>/` |
| **Harnesses** | 生成された動的エージェントチーム | `.moai/harness/manifest.json` |
| **SPEC Progress** | アクティブな SPEC の進行状態 | `.moai/specs/SPEC-*/progress.md` |

## コマンド形式

```bash
moai inventory [options]
```

### 基本の使い方

```bash
moai inventory
```

既定のテキスト形式でインベントリを出力します。

### JSON 形式での出力

```bash
moai inventory --json
```

構造化された JSON で出力し、自動分析に活用できます。

### フィルタリング

特定の資源タイプのみ照会:

```bash
moai inventory --type sessions
moai inventory --type worktrees
moai inventory --type harnesses
moai inventory --type specs
```

### 詳細情報

各資源の追加情報を含める:

```bash
moai inventory --verbose
moai inventory --verbose --json
```

## テキスト形式の出力

### 基本出力の例

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

### 詳細情報 (`--verbose`)

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

## JSON 形式の出力

### スキーマ

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

### Session オブジェクト

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

### Worktree オブジェクト

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

### Harness オブジェクト

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

### SPEC オブジェクト

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

## 実用的な使用例

### 1. マルチセッション競合の検知

```bash
moai inventory --type sessions

# 出力で同じ SPEC を扱うセッションが 1 つより多い場合 → 競合リスク
```

### 2. Worktree の整理確認

```bash
moai inventory --type worktrees --verbose

# 古い worktree を確認して整理
moai worktree remove <name>
```

### 3. Harness チームの一覧照会

```bash
moai inventory --type harnesses --json | jq '.inventory.harnesses[] | {name, teammates, status}'

# 期待される出力:
# {
#   "name": "backend-team",
#   "teammates": 3,
#   "status": "active"
# }
```

### 4. アクティブ SPEC の進捗追跡

```bash
moai inventory --type specs | grep in-progress

# 現在進行中のすべての SPEC を確認
```

### 5. 自動化スクリプトでの使用

```bash
#!/bin/bash
# Worktree 自動整理スクリプト

moai inventory --type worktrees --json | jq -r '.inventory.worktrees[] | select(.status == "stale") | .name' | while read name; do
  echo "Removing stale worktree: $name"
  moai worktree remove "$name"
done
```

## 出力の読み方

### Status フィールド

| Status | 意味 |
|--------|------|
| `active` | 現在使用中 |
| `idle` | 一時停止 (セッションが明示的に一時停止状態) |
| `stale` | 未使用 (7日以上アクセスなし) |
| `error` | エラー状態 (確認が必要) |

### Phase フィールド

| Phase | 説明 |
|-------|------|
| `plan` | Plan フェーズ実行中 |
| `run` | Run フェーズ実行中 |
| `sync` | Sync フェーズ実行中 |
| `completed` | 完了状態 |

## 関連ドキュメント

- [SPEC ベース開発](/workflow-commands/moai-plan) - SPEC のライフサイクル
- [Worktree 管理](/getting-started/worktree) - Worktree の分離とライフサイクル
- [Harness v4 Builder](/advanced/builder-agents) - 動的チーム管理
- [CLI リファレンス](/getting-started/cli) - その他の CLI コマンド

{{< callout type="info" >}}
**ヒント**: `moai inventory` は自動整理スクリプトや監視ダッシュボードに活用できます。JSON 形式で自動分析すれば、プロジェクト状態を常に把握できます。
{{< /callout >}}
