---
title: プロファイルマトリクス
weight: 4
draft: false
---

MoAI-ADK は、維持されるエージェント 11 個を 1 つの **プロファイルマトリクス** を通じてそれぞれの `{model, effort}` ペアにマッピングします。アクティブな **プロファイル**（`high` / `medium` / `low`）がマトリクスの 1 列（column）を選択し、その列の値がすべてのサブエージェント spawn に適用されます。マトリクスはエージェント名単位の **33 セル**（エージェント 11 個 × プロファイル 3 個）であり、以前のグループ抽象化と `plan_type × tier` 軸の両方を置き換えます。

## プロファイル軸

プロファイルは 3 つの値を持ちます:

- `high` — 品質優先の列。マルチターンのエージェンティック行はすべて Opus 5 が担当し、`max` は呼び出し頻度が最も低い 2 行（`manager-develop`、`super-advisor`）に限定されます。`xhigh` はどのセルにも現れません — Opus 5 では `high` と同じスコアでコストだけが明確に高くなるためです。
- `medium`（デフォルト） — バランス列であり、マトリクスの残りが派生する基準点（アンカー）です。`manager-develop` はコスト/スコア曲線の変曲点である Opus 5 `medium` に配置されます。値が無いか空の場合は `medium` として解釈されます。
- `low` — 経済列。Opus 5 の `low` は、どの effort の Sonnet 5 よりもスコアが高く **かつ** 課題あたりコストが安いため、エージェンティック行はすべて Opus のまま維持されます。Sonnet は単発・入力支配の行にのみ現れます。

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
| manager-spec | opus / high | opus / medium | opus / low |
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| manager-design | opus / high | opus / medium | opus / low |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / low | sonnet / low | sonnet / low |

33 セル全体のモデル分布は Opus 25 / Sonnet 8 です。Fable はどのセルにも現れず、`xhigh` を使うセルもありません。

`manager-git` と `Explore` の行はプロファイルと無関係に `sonnet / low` で固定されます — 機械的作業と読み取り専用の探索は、プロファイルが上がってもモデルクラスを上げません。

各行は単調（monotone）です: `high` ≥ `medium` ≥ `low`。プロファイルを下げても、どのエージェントも以前より強い組み合わせを受け取ることはありません。

### セルの根拠

セルはトークン単価からではなく、**すべての effort 段階について** スコア・課題あたりコスト・出力トークン・エージェントステップを報告する長期ホライズンのコーディングエージェントベンチマークから導出されています。配置を決める実測は 3 つです:

- **Opus はすべての effort で Sonnet を支配します。** Opus 5 の `low`（58%、$1.66/課題、36 ステップ）は、`max` の Sonnet 5（54%、$26.40/課題、268 ステップ）を含むどの段階の Sonnet 5 よりもスコアが高く、課題あたりコストも安くなります。課題あたりコストを決めるのは完了効率 — 課題を終えるまでに費やすステップと出力トークン — であり、トークン単価ではありません。したがって Sonnet が残るのは、マルチステップの完了が当てはまらない場所、つまり単発・入力支配の行（`Explore` の検索、`manager-git` の機械的作業）だけであり、そこでは Sonnet の低い入力単価が支配的要因になります。
- **`xhigh` は Opus 上で厳密に劣位です。** `high` は $6.08 で 73%、`xhigh` は $9.07 で同じ 73% — 利得なし、コスト +49%、ステップ +22%。マトリクスから退役しました（6 セル → 0）。`max` は呼び出し頻度が最も低い 2 セルにのみ残ります。
- **`medium` は曲線の変曲点です。** これを超えると 1 ポイントあたりの限界コストが数倍に上がります: `low`→`medium` は 1 ポイントあたり $0.15、`medium`→`high` は 1 ポイントあたり $0.70（4.7 倍）。これが `manager-develop` の `medium` がデフォルト列のアンカーになる理由です。

{{< icon warning warn >}} **根拠の適用範囲**: このベンチマークが測定しているのは *コーディング* エージェントです。ドキュメント作成、監査判断、SPEC 作成の品質は **直接測定されていません** — それらの行の配置は、マルチターンのエージェンティック作業との類似性推論に基づきます。どの行も `llm.agent_overrides` でエージェント単位に元へ戻せます。

Anthropic 組み込みの `Explore` はもはや `inherit` ではなく、自身のセル（`sonnet / low`）として解釈されます。`inherit` センチネルは、ユーザーが追加したエージェントにのみ残ります。

## ハーネススペシャリストの model + effort

`/moai:harness` が生成するスペシャリストは **モデルが `opus` に統一** され、**effort のみで差別化** されます。ハーネスエージェントはユーザー所有の永続的なスペシャリストであり、それらを分ける軸はモデルティアではなく推論の深さだからです。すべての非 Haiku モデルが 1M コンテキストを持つようになったため、モデルを固定してもコンテキストの損失はありません。

effort は、各目的クラスが対応する維持エージェントの行から借用します:

| 目的クラス | effort の由来行 | high | medium | low |
|---|---|---|---|---|
| `read-only-extract` | Explore | opus / low | opus / low | opus / low |
| `mechanical-transform` | manager-git | opus / low | opus / low | opus / low |
| `synthesize` | manager-docs | opus / medium | opus / low | opus / low |
| `research` | plan-auditor | opus / high | opus / medium | opus / low |
| `verify-judge` | sync-auditor | opus / high | opus / medium | opus / low |
| `implement` | manager-develop | opus / max | opus / medium | opus / low |
| `design-architecture` | manager-design | opus / high | opus / medium | opus / low |

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
    manager-develop: { model: opus, effort: high }
```

enum は依然としてモデルとして `fable`、effort として `xhigh` を受け付けます — デフォルトマトリクスから外れただけで語彙から削除されたわけではないため、オーバーライドではどちらも選択できます。

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

- モデルスロットのマッピング: `fable` → `glm-5.3`（Fable スロット、`ANTHROPIC_DEFAULT_FABLE_MODEL`）。このスロットは GLM 環境のバインディングであり、プロファイルマトリクスとは独立です — マトリクスのどのセルも Fable を選択しませんが、配線は維持されます。
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
