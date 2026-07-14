---
title: moai inventory コマンド
weight: 25
draft: false
---

現在のプロジェクトのアクティブセッション、ワークツリー、ハーネスを一目で照会する `moai inventory` コマンドを案内します。

{{< callout type="info" >}}
**一行要約**: `moai inventory` は現在のプロジェクトのアクティブリソース (セッション、ワークツリー、ハーネス) を読み取り専用で照会します。`--json` で構造化された出力を受け取ってスクリプトに活用できます。
{{< /callout >}}

## 概要

`moai inventory` は読み取り専用コマンドで、並列セッションとワークツリーを複数運用するときに「今、何が動いているか?」を一度に確認する統合ビューを提供します。

### 照会対象

| リソース | 説明 | データソース |
|------|------|------------|
| **Sessions** | アクティブな Claude Code セッション | `.moai/state/active-sessions.json` |
| **Worktrees** | プロジェクト用の Git ワークツリー | Git worktree 一覧 |
| **Harnesses** | 登録されたハーネス | `.moai/harness/` マニフェスト |

## コマンド形式

```bash
moai inventory [OPTIONS]
```

### フラグ

| フラグ | 説明 |
|------|------|
| `--json` | 構造化された JSON 出力 (マシンリーダブル) |
| `--project-root <path>` | プロジェクトルートパス (デフォルト値: 現在のディレクトリ) |

このコマンドは上記 2 つのフラグのみをサポートします。フィルタリングや詳細モードのフラグはありません — 必要な加工は `--json` 出力を `jq` などで処理します。

## 基本的な使用

```bash
moai inventory
```

テキスト形式でセッション・ワークツリー・ハーネスの要約を出力します。

## JSON 形式出力

```bash
moai inventory --json
```

構造化された JSON で出力し、自動分析や CI スクリプトに活用できます。

## JSON スキーマ

`--json` 出力のトップレベル構造は 3 つのセクションで構成されます。

```json
{
  "sessions": { ... },
  "worktrees": { ... },
  "harnesses": { ... }
}
```

各セクションは `count`、`entries`、そしてオプションの `error` フィールドを持ちます。

### Session エントリ

```json
{
  "session_id": "edc25996",
  "spec_id": "SPEC-DOCS-001",
  "phase": "run"
}
```

| フィールド | 説明 |
|------|------|
| `session_id` | セッション ID (短縮形、最初の 8 文字) |
| `spec_id` | 連携された SPEC ID |
| `phase` | 現在の Phase (`plan`, `run`, `sync`, `mx`) |

### Worktree エントリ

```json
{
  "branch": "feat/auth",
  "path": "/home/user/.moai/worktrees/project/SPEC-AUTH-001",
  "head": "a1b2c3d4"
}
```

| フィールド | 説明 |
|------|------|
| `branch` | ワークツリーのブランチ名 |
| `path` | ワークツリーのファイルシステムパス |
| `head` | HEAD コミットハッシュ (短縮形、最初の 8 文字) |

### Harness エントリ

```json
{
  "name": "backend-team",
  "domain": "backend",
  "manifest_missing": false
}
```

| フィールド | 説明 |
|------|------|
| `name` | ハーネス名 |
| `domain` | ハーネスドメイン |
| `manifest_missing` | マニフェストファイルの欠落有無 (`true` なら設定が不完全) |

### 全出力例

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

## 実用的な使用例

### 1. 多重セッション競合の検出

同じ SPEC を扱うセッションが 2 つ以上あれば競合のリスクがあります。

```bash
moai inventory --json | jq '[.sessions.entries[] | .spec_id] | group_by(.) | map(select(length > 1))'
```

### 2. アクティブなワークツリーブランチ一覧

```bash
moai inventory --json | jq -r '.worktrees.entries[].branch'
```

### 3. マニフェスト欠落ハーネスを探す

`manifest_missing: true` のハーネスは設定が不完全な状態です。

```bash
moai inventory --json | jq '.harnesses.entries[] | select(.manifest_missing)'
```

### 4. 現在進行中の Phase 分布

```bash
moai inventory --json | jq '[.sessions.entries[].phase] | group_by(.) | map({phase: .[0], count: length})'
```

## 関連ドキュメント

- [CLI リファレンス](./cli) — すべての CLI コマンド
- [プロジェクト状態](./status) — `moai status` コマンド
- [SPEC ベース開発](/ja/workflow-commands/moai-plan) — SPEC ライフサイクル

{{< callout type="info" >}}
**ヒント**: `moai inventory --json` はモニタリングダッシュボードと CI スクリプトに活用できます。読み取り専用コマンドなので安全に自動化できます。
{{< /callout >}}
