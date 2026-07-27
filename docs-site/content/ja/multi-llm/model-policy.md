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
内蔵 `Explore`) です。**No-Haiku ポリシー** の下で Haiku はどこにも現れません。
マルチターンのエージェンティック行はすべて Opus が担当し、Sonnet は単発・入力支配の
行に限定されます。ポリシーティアが制御するのは、各エージェントが Opus の effort
ラダー上のどこに位置するかであり、どのモデルクラスを受け取るかではありません。

## 3 段階ポリシー概要

| ポリシー (profile) | CLI フラグ | Opus セル | Sonnet セル | 適した用途 |
|------------------|-----------|----------|------------|-----------|
| **high** | `--model-policy high` | 11 中 9 | 11 中 2 | 最高品質。呼び出し頻度が最も低い 2 行に `max` effort |
| **medium** (デフォルト) | `--model-policy medium` | 11 中 9 | 11 中 2 | 品質とコストのバランス。コスト/スコア曲線の変曲点 |
| **low** | `--model-policy low` | 11 中 7 | 11 中 4 | 課題あたりコスト最小。エージェンティック行は Opus `low` に下がる |

> **名前の軸**: `llm.yaml` の `profile` フィールド、legacy の `performance_tier`
> エイリアス、CLI フラグ `--model-policy` はすべて同じく `high`/`medium`/`low` の
> 3 値を使い、1:1 でマッピングされます (別途変換なし)。デフォルト値は `medium` です。
> 旧最上位ティア名の `max` は既存設定が解決を続けられるよう依然として `high` の
> エイリアスとして **読み込まれ** ますが、保存時は常に `high` が書き込まれます —
> マイグレーション作業は不要です。`performance_tier` は `profile` がない場合のみ
> 読み込まれます。ユーザー名などは `user.yaml` に別途保管されます。

> **なぜ重要ですか?** ポリシーを下げることは、もはや弱いモデルクラスへの切り替えを
> 意味しません。長期ホライズンのエージェンティック課題では、Opus の `low` effort は
> どの effort の Sonnet よりもスコアが高く **かつ** 課題あたりコストが安くなります。
> 請求額を決めるのはトークン単価ではなく、モデルが完了までに費やすステップ数だからです。
> したがって `low` ポリシーは推論深度を下げて Opus の *内側* で節約し、Sonnet に
> 手を伸ばすのは、マルチステップの完了失敗が当てはまらない単発の行だけです。

## エージェント別モデル割り当て表

以下の 33 セルがプロファイルマトリクス (11 エージェント × 3 プロファイル) です。各セルは
リゾルバが spawn 時点で注入する `{model, effort}` ペアです。(オーケストレーターの
メインセッションは spawn されるエージェントではないため、表には含まれません。)

### Manager Agents (5 個)

| エージェント | high | medium | low |
|---------|------|--------|-----|
| manager-spec | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| manager-design | opus / high | opus / medium | opus / low |

### Evaluator · Advisor · Builder · Specialist Agents (5 個)

| エージェント | high | medium | low |
|---------|------|--------|-----|
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |

### ビルトインエージェント (1 個)

| エージェント | high | medium | low |
|---------|------|--------|-----|
| Explore | sonnet / low | sonnet / low | sonnet / low |

> `Explore` はディスク上にエージェントファイルが無いため frontmatter で effort を
> 固定できません — マトリクスは呼び出し時のデフォルトとして `sonnet / low` を記録し、
> spawn プロンプトで指定します。Agent Teams 静的階層 (静的 role profile) は v3.0 で
> 引退し、並列作業は sub-agent 並列実行と動的ワークフローが
> 代替します。`moai cg` の teammate ランタイム (tmux pane) はそのまま維持されます。

> **Haiku 除去 (v3.0)**: かつての Haiku スロット (ドキュメント、MX タグ付け、Git
> 手続き) は、低いモデルクラスではなく低い推論深度に置き換えられました — コストは
> モデルの差し替えではなく effort のティア分けで削減します。

## 割り当て原則

- **エージェンティック行はすべて Opus**: `manager-spec`, `manager-develop`, `plan-auditor`, `sync-auditor`, `manager-design`, `builder-harness`, `manager-docs`, `e2e-tester` — Opus の `low` がどの effort の Sonnet よりも高スコアかつ課題あたり低コストであるため、マルチターン作業はすべて Opus に留まります
- **Sonnet は単発の行のみ**: `manager-git` の機械的作業と `Explore` の検索は入力支配の 1 パスで完了するため、マルチステップの完了失敗が当てはまらず、Sonnet の低い入力単価が支配的要因になります。この 2 行は 3 つのプロファイル全体で固定されます
- **`max` は 2 セルに限定**: `manager-develop` と `super-advisor`、しかも `high` プロファイルのみ — 呼び出し頻度が最も低く、1 つの判断が不均衡に大きな下流コストを持つ行です
- **`xhigh` はどこにも使いません**: Opus 上では 49% 高いコストで `high` と同スコアです
- **`low` はモデルクラスではなく effort を下げます**: エージェンティック行は Opus `low` に移り、Sonnet にフォールバックするのは `manager-docs` と `e2e-tester` だけです

計画を作ったエージェントが監査しないように、`plan-auditor` と `sync-auditor` は
`manager-spec` から独立した割り当てを維持します — バイアス防止はセルの値ではなく
カタログの構造的性質です。

## v3.0 拡張: Tier×Phase 宣言軸

v3.0 ではエージェント単位の割り当ての上に **作業ステップ (phase) と SPEC サイズ (Tier)**
軸が加わりました。`internal/config/model_routing.go` が Tier×Phase →
{model, effort} マトリックスを宣言的に管理します:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (推論深度): low / medium / high / xhigh / max
- **tier** (SPEC サイズ): S / M / L
- **phase** (作業ステップ): plan / run / sync / mx

エージェント別の model+effort 割り当ては単一のプロファイルマトリクスが担当します。アクティブ
プロファイル(`profile` — `high`/`medium`/`low`)がマトリクスの 1 列を選択し、
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
moai init my-project --model-policy high    # 最高品質 (2 行に max effort)
moai init my-project --model-policy medium  # バランス (デフォルト値)
moai init my-project --model-policy low     # 課題あたりコスト最小
```

`--model-policy` は `high`/`medium`/`low` の 3 値を受け付け、`llm.yaml` の
`performance_tier` フィールドに永続化されます。旧最上位ティア名の `max` も入力として
受け付けられ、`high` のエイリアスとして扱われます。

> デフォルトポリシーは `medium` です (llm.yaml `performance_tier: "medium"`、CLI `--model-policy medium` に該当 — 値がなければ `medium` として解釈)。GLM 設定は `settings.local.json` に隔離され、Git にコミットされません。

## 次のステップ

- [CG モード](/ja/multi-llm/cg-mode) — Claude + GLM ハイブリッドでコスト削減
- [エージェントガイド](/ja/advanced/agent-guide) — エージェントのカスタマイズ
- [CLI リファレンス](/ja/getting-started/cli) — moai init、moai update 詳細
