---
title: プロファイルマトリクス
weight: 4
draft: false
---

MoAI-ADK は、維持されるエージェント 11 個を 1 つの **プロファイルマトリクス** を通じてそれぞれの `{model, effort}` ペアにマッピングします。アクティブな **プロファイル**（`high` / `medium` / `low`）がマトリクスの 1 列（column）を選択し、その列の値がすべてのサブエージェント spawn に適用されます。マトリクスはエージェント名単位の **33 セル**（エージェント 11 個 × プロファイル 3 個）であり、以前のグループ抽象化と `plan_type × tier` 軸の両方を置き換えます。

## プロファイル軸

プロファイルは 3 つの値を持ちます:

- `high` — 品質優先の列。推論・監査の行に Fable 5 を配置し、コーディングには Opus 5 を `xhigh` で配置します（ベンダーがコーディング・エージェンティック作業に推奨する開始点）。
- `medium`（デフォルト） — バランス列。Opus 5 をベンダー API のデフォルト effort である `high` で配置するため、最も予測可能な運用点です。値が無いか空の場合は `medium` として解釈されます。
- `low` — 経済列。Opus 5 の `low`/`medium` effort を第一のトークンコストレバーとし、その次に Sonnet 5 へ下げます。

`max` は `high` の **読み取り専用エイリアス** です。既存設定の `profile: max` はそのまま `high` として解釈され、保存時には常に正規名 `high` で記録されます。マイグレーション作業は不要です。

プロファイルは `performance_tier` とは別のフィールドではなく同一の軸です — `llm.profile` が優先され、無い場合は legacy `performance_tier` がエイリアスとして読まれます。両フィールドとも `high`/`medium`/`low` の語彙を共有します。リゾルバはこの有効プロファイルを読んで各エージェントのセルを決定します。

## プロファイルの設定

```bash
moai init . --profile high             # 初期化時に設定
moai update --profile low              # 事後の切り替え
```

許容値は `high` / `medium` / `low` であり、legacy の `max` も入力として受け付け `high` に正規化します。現在の値は `.moai/config/sections/llm.yaml` の `llm.profile` フィールドで確認できます。

## プロファイルマトリクス

維持されるエージェント 11 個が、以下のマトリクスからそれぞれの `{model, effort}` を直接受け取ります。ユーザーが追加したエージェントのみが `inherit`（親セッションモデルの継承）として解釈され、model 注入の対象から外れます。マトリクスのどこにも Haiku はありません。

| エージェント | high | medium（デフォルト） | low |
|---|---|---|---|
| manager-spec | fable / xhigh | opus / high | opus / low |
| plan-auditor | fable / xhigh | opus / high | opus / low |
| sync-auditor | fable / xhigh | opus / high | opus / low |
| manager-develop | opus / xhigh | opus / high | sonnet / medium |
| super-advisor | opus / xhigh | opus / high | opus / medium |
| manager-design | fable / high | opus / medium | sonnet / medium |
| builder-harness | opus / xhigh | opus / medium | sonnet / medium |
| e2e-tester | fable / high | opus / medium | sonnet / medium |
| manager-docs | sonnet / high | sonnet / medium | sonnet / medium |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / low | sonnet / low | sonnet / low |

`manager-git` と `Explore` の行はプロファイルと無関係に `sonnet / low` で固定されます — 機械的作業と読み取り専用の探索は、プロファイルが上がってもモデルクラスを上げません。

各行は単調（monotone）です: `high` ≥ `medium` ≥ `low`。プロファイルを下げても、どのエージェントも以前より強い組み合わせを受け取ることはありません。

Anthropic 組み込みの `Explore` はもはや `inherit` ではなく、自身のセル（`sonnet / low`）として解釈されます。`inherit` センチネルは、ユーザーが追加したエージェントにのみ残ります。

## ハーネススペシャリストの model + effort

`/moai:harness` が生成するスペシャリストは **モデルが `opus` に統一** され、**effort のみで差別化** されます。ハーネスエージェントはユーザー所有の永続的なスペシャリストであり、それらを分ける軸はモデルティアではなく推論の深さだからです。すべての非 Haiku モデルが 1M コンテキストを持つようになったため、モデルを固定してもコンテキストの損失はありません。

effort は、各目的クラスが対応する維持エージェントの行から借用します:

| 目的クラス | effort の由来行 | high | medium | low |
|---|---|---|---|---|
| `read-only-extract` | Explore | opus / low | opus / low | opus / low |
| `mechanical-transform` | manager-git | opus / low | opus / low | opus / low |
| `synthesize` | manager-docs | opus / high | opus / medium | opus / medium |
| `research` | plan-auditor | opus / xhigh | opus / high | opus / low |
| `verify-judge` | sync-auditor | opus / xhigh | opus / high | opus / low |
| `implement` | manager-develop | opus / xhigh | opus / high | opus / medium |
| `design-architecture` | manager-design | opus / high | opus / medium | opus / medium |

`llm.harness_agents[プロファイル][クラス].effort` でクラスごとの effort を上書きできます。モデルはどの経路でも変わりません。認識されないクラスは `implement` にフォールバックします。

## リゾルバの優先順位

各エージェントの有効な `{model, effort}` は次の順序で決定されます:

1. `llm.agent_overrides[agent]` があればそれが勝ちます。
2. 無ければアクティブプロファイルのエージェントセル（config `llm.profiles`）を使用します。
3. config にセルが無ければ Go デフォルトマトリクス（`template.DefaultProfileMatrix`）のエージェントセルを使用します。
4. マトリクスに無いエージェント（ユーザー追加）は `inherit`（注入しない）です。

`agent_overrides` は正規エージェント名をキーとし、カタログ + enum に対して検証されます:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: xhigh }
```

**model** と **effort** の消費経路は異なります。リゾルブされた **model** は、オーケストレーターが spawn 時点で `Agent(model: <alias>)` ランタイム引数として注入する値です（`[1m]`-safe、frontmatter の `model:` フィールドとは別）。エージェント `.md` の frontmatter は `model: inherit` のまま維持され、init/update/web の保存はこれを変更しません。リゾルブされた **effort** は NAMED サブエージェントに対する *文書化された意図* です — Agent/Task ツールは named サブエージェントに per-spawn の effort 引数を受け取らないため、effort は (a) エージェント frontmatter の effort デフォルト、(b) GLM effort オーバーレイ、(c) Workflow / `Agent(general-purpose)` のプロンプトレベル steering を通じてのみ消費されます。

## moai model profile

アクティブプロファイルでリゾルブされたエージェント別の model+effort は、読み取り専用のアクセサで確認します:

```bash
moai model profile          # 人間向けの表
moai model profile --json   # 機械可読
```

このコマンドは何も変更しません — オーケストレーターが spawn 時に注入する値をそのまま公開します。

## GLM バックエンド effort オーバーレイ

{{< icon warning warn >}} **正直性の告知**: GLM バックエンドの effort オーバーレイは **実装 + 配線完了** の状態ですが、wire の有効性（ライブ有効性）は実証予定です — 「動作保証」としては記述しません。

GLM バックエンド（`moai glm` / `moai cg` の GLM ペイン）では、プロファイルマトリクスの上にオーバーレイが適用されます:

- モデルスロットのマッピング: `fable` → `glm-5.2`（Fable スロット、`ANTHROPIC_DEFAULT_FABLE_MODEL`）
- Claude の 5 段 effort を z.ai が到達可能な 3-state に collapse:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max`（legacy の effort 値） → **reasoning-max**
  - （認識不能な値 → reasoning-max、過少推論の防止）
- coding-max override: `manager-develop` は collapse 結果と無関係に **reasoning-max** を強制
- `manager-git` は low effort → **thinking-off**

z.ai が Anthropic-compat shim で `ANTHROPIC_REASONING_EFFORT` の値を実際に消費するかは、ライブ GLM セッションのアウトバウンド観測が必要な実証課題です。ランタイムの SSOT は `internal/template/glm_effort_overlay.go` です。

## 次のステップ

- [3-ティアエージェントアーキテクチャ](/ja/advanced/no-haiku-3tier/) — DeepSWE リーダーボードの根拠と 3-ティア定義
- [トークノミクス概論](/ja/advanced/tokenomics-overview/) — 4 層トークノミクス構造の Layer B ルーティング
- [モデルポリシー](/ja/multi-llm/model-policy/) — performance_tier のエイリアスと GLM バックエンドの詳細
