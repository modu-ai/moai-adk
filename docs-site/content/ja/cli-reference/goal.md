---
title: moai goal ゴールループ
weight: 72
draft: false
---

`moai goal` は現在のセッションに対して条件を宣言したエージェンティックゴールループを arm/照会/解除します。MoAI goal エンジンは、宣言した条件が満たされるか、ターン上限に達するまで、セッションがターンを越えて作業を継続できるようにします。

これはネイティブ `/goal` (ユーザー専用 TUI コマンド) のプログラマティックな MoAI 対応物であり、オーケストレーターが人間による `/goal` 行の入力なしにゴールを登録・arm できるようにします。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai goal arm "<condition>"` | アクティブセッションにゴールを登録 + arm (`moai goal "<condition>"` も arm のエイリアス)。arm 専用 — これ自体は作業を開始しない |
| `moai goal status` | アクティブセッションのゴール状態を出力 (`--all` で全セッションを一覧) |
| `moai goal clear` | アクティブセッションのゴールを解除 |
| `moai goal render` | アクティブセッションのゴールダッシュボードを self-contained HTML ファイルにレンダリング (`.moai/state/goal/` 配下に保存)。arm された goal がない場合は非 0 の終了コードで終了。v3.1(PR #1388)からダッシュボードに判定セクション (天井 exit 時にサイドカーからロード) と再武装条件付きビューが表示されます — 詳細は [/moai goal ダッシュボードセクション](/ja/utility-commands/moai-goal#ゴールダッシュボード) を参照 |

## 共通フラグ

| フラグ | 説明 |
|--------|------|
| `--session <id>` | セッション id を上書き (デフォルト: `moai session current` で解決) |
| `--json` | 機械可読 JSON 出力 |
| `--all` | (`status` 専用) アクティブセッションだけでなく全セッションのゴールを一覧表示 |

## arm フラグ

| フラグ | 説明 |
|--------|------|
| `--max-turns <N>` | ターン上限。`0` = 無限 (SPEC-INFINITE-GOAL-001); 省略時はデフォルト `30` (完全な後方互換)。**`0` (無限) は `--max-duration <sec>` を必須とする** (arm 時点の fail-closed)。 |
| `--max-duration <sec>` | 実時間上限 (arm 時点からの秒数)。**無限 goal (`--max-turns 0`) の実際の壁時間上限** — 無限 goal はこのフラグなしでは arm できない。 |
| `--cost-cap <value>` | 呼び出し回数上限として**記録専用 (recorded only)** — 現在 enforce ロジックがないため実際の bound ではない。`--max-turns 0` に対する実際の bound 要件を満たさないため、cost-cap 単独では拒否される。 |

## 状態と評価

ゴール状態は `.moai/state/goal/<session-id>.json` (セッションごとに 1 ファイル) に保存されます。Stop フック `moai hook stop-goal` が各ターン終了時にゴールを評価します。

**条件のパース**:

- 実行可能なシェルコマンド (任意で `exits <N>` 接尾辞) は **機械的 (mechanical) 条件** になります。
- 対話 transcript を参照する主張は、オーケストレーターが評価する **モデル (model) 条件** になります。

## 例

```bash
# テストスイートが通過するまで作業を継続
moai goal arm "go test ./... exits 0"

# 現在のゴール状態を確認
moai goal status

# ゴールを解除
moai goal clear
```

## 関連ドキュメント

- [自律継続ループ](/ja/advanced/autonomous-loops) — `/goal` と `/moai loop` の比較
- [moai loop](/ja/cli-reference/loop)
- [CLI 概要](/ja/getting-started/cli)
