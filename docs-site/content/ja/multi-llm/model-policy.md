---
title: モデルポリシー
weight: 30
draft: false
---

## モデルポリシーとは?

モデルポリシーは MoAI-ADK トークノミクスの骨格です。「すべての作業に最高モデル」
ではなく、エージェントごとに — 計画・監査のように推論が重い仕事と、ドキュメント化・Git のように
軽い仕事ごとに — 適切なモデルを宣言的に割り当てます。Claude Code サブスクリプションプランに
合わせて品質を最大化しながら、レート制限エラーを防ぎます。

MoAI-ADK v3.0 のエージェントカタログは **11 個** (MoAI カスタム 10 個 + Anthropic
内蔵 `Explore`) で、下記の割り当て表はそのうちモデルポリシーが直接割り当てるコア 7 個の
エージェントを扱います。

## 3 段階ポリシー概要

| ポリシー (performance_tier) | CLI フラグ | プラン | Opus | Sonnet | 適した用途 |
|------------------------|-----------|------|------|--------|-----------|
| **max** | `--model-policy max` | Max $200/月 | 5 | 2 | 最高品質、最大スループット |
| **medium** (デフォルト) | `--model-policy medium` | Max $100/月 | 2 | 5 | 品質とコストのバランス |
| **low** | `--model-policy low` | Plus $20/月 | 0 | 7 | 低予算、Opus 非含 |

> **名前の軸**: `llm.yaml` の `performance_tier` フィールドと CLI フラグ `--model-policy` は
> 同じく `max`/`medium`/`low` の 3 値を使い、1:1 でマッピングされます (別途変換
> なし)。デフォルト値は `medium` です。`--high` フラグは `--model-policy max` の、もう
> 使われない別名です (1 サイクル後方互換、`--low` も同様)。
> `performance_tier` は `profile`(プロファイルマトリクス列)の legacy エイリアスフィールドで、
> `profile` がない場合のみ読み込まれ `high`→`max` に正規化されます。両フィールドは同じ
> `max`/`medium`/`low` 軸です。ユーザー名などは `user.yaml` に別途保管されます。

> **なぜ重要ですか?** Plus $20 プランは Opus にアクセスできません。`low` ポリシーを設定すると、すべてのエージェントが Sonnet のみを使ってレート制限エラーを防ぎます。上位プランはコアエージェント (計画、監査) に Opus を割り当て、日常作業には Sonnet を使います。

## エージェント別モデル割り当て表

### Manager Agents (4 個)

| エージェント | max | medium | low |
|---------|-----|--------|-----|
| manager-spec | opus | opus | sonnet |
| manager-develop | opus | sonnet | sonnet |
| manager-docs | sonnet | sonnet | sonnet |
| manager-git | sonnet | sonnet | sonnet |

### Evaluator & Builder Agents (3 個)

| エージェント | max | medium | low |
|---------|-----|--------|-----|
| plan-auditor | opus | opus | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| builder-harness | opus | sonnet | sonnet |

> Anthropic 内蔵の `Explore` は読み取り専用の探索エージェントで、別途割り当てなしで
> 動作します。Agent Teams 静的階層 (静的 role profile) は v3.0 で
> 引退し、並列作業は sub-agent 並列実行と動的ワークフローが
> 代替します。`moai cg` の teammate ランタイム (tmux pane) はそのまま維持されます。

> **Haiku 除去 (v3.0)**: かつての Haiku スロットは `sonnet`/`effort:low` に
> 置き換えられました。`manager-git` と `manager-docs` の軽い作業がこれに
> 該当し、モデルは Sonnet ですが推論深度を下げてコストを削減します。

## 割り当て原則

- **常に Opus**: 計画監査 (plan-auditor)、SPEC 作成 (manager-spec) — 高い推論能力が必要
- **常に Sonnet/effort:low**: Git (manager-git) — 軽くて速い作業
- **プランに応じて変動**: 実装 (manager-develop, cycle_type=tdd/ddd) — プランが高いほど Opus

計画を作ったエージェントが監査しないように、plan-auditor と sync-auditor は独立した
割り当てを維持します — コスト軸と品質軸 (バイアス防止) が一緒に設計された表です。

## v3.0 拡張: Tier×Phase 宣言軸

v3.0 ではエージェント単位の割り当ての上に **作業ステップ (phase) と SPEC サイズ (Tier)**
軸が加わりました。`internal/config/model_routing.go` が Tier×Phase →
{model, effort} マトリックスを宣言的に管理します:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (推論深度): low / medium / high / xhigh / max
- **tier** (SPEC サイズ): S / M / L
- **phase** (作業ステップ): plan / run / sync / mx

エージェント別の model+effort 割り当ては単一のプロファイルマトリクスが担当します。アクティブ
プロファイル(`profile` — `max`/`medium`/`low`)がマトリクスの 1 列を選択し、
`profile` がなければ legacy `performance_tier` がエイリアスとして読み込まれ、それもなければ
`medium` として解釈されます。詳細なエージェント別マッピングは
[プロファイルマトリクス](/ja/advanced/profile-matrix/) ページを参照してください。

## 設定方法

### プロジェクト初期化時

```bash
moai init my-project
# 対話型ウィザードでモデルポリシー選択を含む
```

### 既存プロジェクトの再設定

```bash
moai update
# 対話型プロンプト:
# - Reset model policy? (y/n) — モデルポリシー再設定
# - Update GLM settings? (y/n) — GLM 環境変数設定
```

### CLI フラグで直接設定

```bash
moai init my-project --model-policy max     # 最高品質 (Opus 中心)
moai init my-project --model-policy medium  # バランス (デフォルト値)
moai init my-project --model-policy low     # Sonnet のみ、Opus 非使用
```

`--model-policy` は `max`/`medium`/`low` の 3 値を受け付け、`llm.yaml` の
`performance_tier` フィールドにそのまま保存されます。もう使われない `--high`
フラグは `--model-policy max` の別名です。

> デフォルトポリシーは `medium` です (llm.yaml `performance_tier: "medium"`、CLI `--model-policy medium` に該当 — 値がなければ `medium` として解釈)。GLM 設定は `settings.local.json` に隔離され、Git にコミットされません。

## 次のステップ

- [CG モード](/ja/multi-llm/cg-mode) — Claude + GLM ハイブリッドでコスト削減
- [エージェントガイド](/ja/advanced/agent-guide) — エージェントのカスタマイズ
- [CLI リファレンス](/ja/getting-started/cli) — moai init、moai update 詳細
