---
title: moai harness ハーネス
weight: 80
draft: false
---

`moai harness` は SPEC 複雑度ルーティングとハーネス学習サブシステムを管理する統合コマンドツリーです。ルーティング、検証、ライフサイクル、提案管理、v4 ハーネスライフサイクル、観測台帳 (ledger) のサブコマンドを提供します。

共通フラグとして `--project-root <path>` (デフォルト: 現在のディレクトリ) を受け取ります。

## ルーティング verb

| コマンド | 説明 |
|--------|------|
| `moai harness route --spec <id>` | SPEC を minimal/standard/thorough ハーネスレベルへルーティング |
| `moai harness validate` | harness.yaml をスキーマ・不変条件基準で検証 |

`route` は `--json` (JSON 出力)、`--path <harness.yaml>`、`--base-dir <dir>` フラグを受け取ります。

## ライフサイクル verb

| コマンド | 説明 |
|--------|------|
| `moai harness status` | 観測/ティア/進化サマリーを表示 |
| `moai harness apply` | 保留中の提案をオーケストレーターへ返却 (または `--execute` で Go apply 経路を実行) |
| `moai harness rollback <date>` | 指定日付のスナップショットを復元 |
| `moai harness disable` | 学習サブシステムを無効化 (`learning.enabled: false`) |

## 提案管理 verb

| コマンド | 説明 |
|--------|------|
| `moai harness mute` | 提案カテゴリをミュート (workflow.yaml) |
| `moai harness mute-list` | 現在ミュート中のカテゴリを出力 |
| `moai harness unmute` | カテゴリをミュートリストから除去 |
| `moai harness verify` | ハーネスの決定性を検証 |

## v4 ハーネスライフサイクル verb

| コマンド | 説明 |
|--------|------|
| `moai harness list` | すべての v4 ハーネスを一覧表示 (名前 + ドメイン + 起動コマンド) |
| `moai harness edit <name>` | v4 ハーネスの manifest + specialist 編集パスを表示 |
| `moai harness remove <name>` | v4 ハーネスの原子的削除 (コマンド + workflow + specialist + skill + manifest) |
| `moai harness doctor` | ハーネスのインストール状態を診断 |

`list`、`edit` は `--json` フラグを受け取ります。

## 観測台帳 (ledger)

`moai harness ledger` はルーティング観測台帳を管理します。

| コマンド | 説明 |
|--------|------|
| `moai harness ledger record` | dispatch 時点のルーティング決定を記録 (pending row) |
| `moai harness ledger evidence` | pending row に機械証拠 ref (または委任項目) を追加 |
| `moai harness ledger list` | フィルタで最終台帳 row を一覧表示 |

> `ledger record` と `ledger evidence` は結果 (outcome) を捏造できないよう `--outcome` フラグを公開していません。結果は機械証拠から導出されます。

## 関連ドキュメント

- [ハーネス自己進化](/ja/advanced/self-evolving)
- [Harness v4 Builder 詳細ガイド](/ja/advanced/harness-v4-builder)
- [ハーネスプロファイルと評価](/ja/advanced/harness-profiles)
- [CLI 概要](/ja/getting-started/cli)
