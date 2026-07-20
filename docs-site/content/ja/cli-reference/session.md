---
title: moai session セッションレジストリ
weight: 65
draft: false
---

`moai session` は `.moai/state/active-sessions.json` の多重セッション調整レジストリを管理します。複数の Claude Code セッションが同じプロジェクトで同時に作業する際に発生する競合 (race) を緩和するためのツールです。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai session register <session_id> <spec_id> <phase>` | 新しいアクティブセッションを登録 |
| `moai session heartbeat <session_id>` | 既存セッションの last_heartbeat を更新 (idempotent) |
| `moai session deregister <session_id>` | セッションを除去 (idempotent) |
| `moai session list` | アクティブセッションを一覧表示 (`--filter-spec` でフィルタ可能) |
| `moai session purge` | stale なエントリを除去 (デフォルト: 最終 heartbeat から 30 分超過) |
| `moai session current` | このオーケストレーターのセッション UUID を出力 |
| `moai session doctor` | レジストリが空である原因を診断 |

ほとんどのサブコマンドは `--json` フラグで機械可読出力をサポートします。

## moai session list

```bash
moai session list
moai session list --filter-spec SPEC-AUTH-001
```

| フラグ | 説明 |
|--------|------|
| `--json` | 機械可読 JSON 出力 (オーケストレーターの pre-spawn 検査形式) |
| `--filter-spec <id>` | 該当する spec_id と一致するエントリのみ返却 |

## moai session purge

```bash
moai session purge
```

| フラグ | 説明 |
|--------|------|
| `--json` | JSON 出力 |
| `--threshold-minutes <n>` | stale heartbeat のカットオフ (分、デフォルト 30) |

## moai session current

```bash
moai session current
```

オーケストレーター自身のセッション UUID を出力します。ランタイムがセッション ID を公開しない場合、canonical fallback 文字列を返します。

| フラグ | 説明 |
|--------|------|
| `--json` | JSON 出力 |
| `--show-fallback` | canonical fallback 文字列のみ出力 (paste-ready resume 生成用) |

## moai session doctor

```bash
moai session doctor
```

多重セッション調整レジストリが空である理由を診断します (write-path 診断)。

| フラグ | 説明 |
|--------|------|
| `--json` | JSON 出力 |

## 使用コンテキスト

このレジストリは、オーケストレーターが実装エージェントを spawn する前に同時セッションの競合を検出するために使用されます。`moai session list --json --filter-spec <SPEC-ID>` が他のセッションのエントリを返した場合、オーケストレーターは進行を止めてユーザーに確認します。

## 関連ドキュメント

- [moai inventory](/ja/cli-reference/inventory) — セッション・ワークツリー・ハーネスの統合照会
- [CLI 概要](/ja/getting-started/cli)
