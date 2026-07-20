---
title: moai loop フィードバックループ
weight: 76
draft: false
---

`moai loop` は SPEC ライフサイクルの Ralph フィードバックループコントローラーを管理します。1 つの SPEC に対してツール診断が検出した作業を反復処理する状態機械を制御します。

> この CLI コマンドは Claude Code 対話画面の `/moai loop` スキルとは別物です — CLI はループコントローラーの状態を操作し、`/moai loop` スキルはオーケストレーターが実際の反復修正を実行します。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai loop start <SPEC-ID>` | SPEC に対するフィードバックループを開始 |
| `moai loop status` | 現在のループ状態を表示 |
| `moai loop pause` | 実行中のループを一時停止 |
| `moai loop resume <SPEC-ID>` | 一時停止したループを再開 |
| `moai loop cancel` | 実行中のループをキャンセル |

## 例

```bash
# SPEC に対するループを開始
moai loop start SPEC-AUTH-001

# 現在の状態を確認
moai loop status

# 一時停止して後で再開
moai loop pause
moai loop resume SPEC-AUTH-001

# ループをキャンセル
moai loop cancel
```

## 関連ドキュメント

- [自律継続ループ](/ja/advanced/autonomous-loops) — Ralph エンジンとゴールベースループ
- [moai goal](/ja/cli-reference/goal)
- [CLI 概要](/ja/getting-started/cli)
