---
title: プロファイルマトリクス
weight: 4
draft: false
---

MoAI-ADK は、維持される各エージェントを 1 つの **プロファイルマトリクス** を通じて `{model, effort}` ペアにマッピングします。アクティブな **プロファイル**（`max` / `medium` / `low`）がマトリクスの 1 列（column）を選択し、その列の値がすべてのサブエージェント spawn に適用されます。この単一の 3 列プロファイル軸は、以前の `plan_type × tier` 60 セルマトリクスを置き換えます（SPEC-MODEL-PROFILE-MATRIX-001）。

## プロファイル軸

プロファイルは 3 つの値を持ちます:

- `max` — 最高品質の列。推論ポイントに Fable を、設計・ハーネス・E2E に Opus を配置します。
- `medium`（デフォルト）— バランス列。推論と実行に Opus/high を配置します。値がない、または空の場合は `medium` として解釈されます。
- `low` — 経済列。Opus を低い effort で配置し、機械的な作業を Sonnet に回します。

プロファイルは `performance_tier` とは別のフィールドではなく同一の軸です — `llm.profile` が優先され、なければ legacy の `performance_tier` がエイリアスとして読み込まれます（`high` → `max` 正規化、`max`/`medium`/`low` はそのまま）。リゾルバはこの有効プロファイルを読み取り、各エージェントのセルを決定します。

## プロファイルの設定

```bash
moai init . --profile max              # 初期化時に設定
moai update --profile low              # 事後切替
```

現在の値は `.moai/config/sections/llm.yaml` の `llm.profile` フィールドで確認できます。`moai init` 対話型ウィザードでは `high` の回答が `max` に正規化されます。

## プロファイルマトリクス

10 個のグループ化されたエージェントが、下のマトリクスから `{model, effort}` を受け取ります。`Explore` とユーザー定義エージェントはグループがないため `inherit`（親セッションモデルを継承）として解釈され、model 注入の対象ではありません。マトリクスのどこにも Haiku はありません。

| エージェント (グループ) | max | medium (デフォルト) | low |
|---|---|---|---|
| manager-spec (spec_auditors) | fable / medium | opus / high | opus / low |
| plan-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| sync-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| manager-develop (develop) | fable / low | opus / high | opus / medium |
| super-advisor (advisor) | fable / medium | fable / low | opus / high |
| manager-design (design_harness_e2e) | opus / high | opus / medium | opus / low |
| builder-harness (design_harness_e2e) | opus / high | opus / medium | opus / low |
| e2e-tester (design_harness_e2e) | opus / high | opus / medium | opus / low |
| manager-docs (docs) | sonnet / medium | sonnet / medium | sonnet / medium |
| manager-git (git) | sonnet / low | sonnet / low | sonnet / low |
| Explore (—) | inherit | inherit | inherit |

`docs` と `git` の行はプロファイルに関係なく固定されます（それぞれ sonnet/medium、sonnet/low）— 機械的な作業はプロファイルが変わってもモデルクラスを上げません。

## エージェントグループ

マトリクスはエージェント名ではなく、6 つの **グループ** 単位で定義されます。グループ → エージェントのメンバーシップは次のとおりです:

| グループ | エージェント |
|---|---|
| `spec_auditors` | manager-spec, plan-auditor, sync-auditor |
| `develop` | manager-develop |
| `advisor` | super-advisor |
| `design_harness_e2e` | manager-design, builder-harness, e2e-tester |
| `docs` | manager-docs |
| `git` | manager-git |

`Explore` とユーザーが追加したエージェントはメンバーシップがないため `inherit` として解釈されます。

## リゾルバの優先順位

各エージェントの有効な `{model, effort}` は、次の順序で決定されます:

1. `llm.agent_overrides[agent]` があればそれが勝ちます。
2. なければアクティブプロファイルのグループセル（config `llm.profiles`）を使用します。
3. config にセルがなければ Go デフォルトマトリクス（`template.DefaultProfileMatrix`）のグループセルを使用します。
4. グループメンバーシップがなければ `inherit`（注入なし）です。

`agent_overrides` は正規のエージェント名をキーとし、カタログ + enum に対して検証されます:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: xhigh }
```

**model** と **effort** の消費経路は異なります。リゾルブされた **model** は、オーケストレーターが spawn 時点で `Agent(model: <alias>)` ランタイム引数として注入する値です（`[1m]`-safe、frontmatter の `model:` フィールドとは別）。エージェント `.md` frontmatter は `model: inherit` のまま維持され、init/update/web の保存がこれを変更しません。リゾルブされた **effort** は NAMED サブエージェントに対する *文書化された意図* です — Agent/Task ツールは named サブエージェントに per-spawn の effort 引数を受け取らないため、effort は (a) エージェント frontmatter の effort デフォルト、(b) GLM effort オーバーレイ、(c) Workflow / `Agent(general-purpose)` プロンプトレベルの steering を通じてのみ消費されます。

## moai model profile

アクティブプロファイルでリゾルブされたエージェントごとの model+effort は、読み取り専用のアクセサで確認します:

```bash
moai model profile          # 人間用の表
moai model profile --json   # 機械可読用
```

このコマンドは何も変更しません — オーケストレーターが spawn 時に注入する値をそのまま公開します。

## GLM バックエンド effort オーバーレイ

{{< icon warning warn >}} **正直性注記**: GLM バックエンド effort オーバーレイは **実装 + 配線完了** の状態ですが、wire 有効性（ライブ有効性）は実証予定です — 「動作保証」とは記載しません。

GLM バックエンド（`moai glm` / `moai cg` GLM パネル）では、プロファイルマトリクスの上にオーバーレイが適用されます:

- モデルスロットマッピング: `fable` → `glm-5.2`（Fable スロット、`ANTHROPIC_DEFAULT_FABLE_MODEL`）
- Claude の 5 段 effort を z.ai が到達可能な 3-state に collapse:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max` → **reasoning-max**
  - （認識不可の値 → reasoning-max、過小推論の防止）
- coding-max override: `manager-develop` は collapse 結果に関係なく **reasoning-max** を強制
- `manager-git` は low effort → **thinking-off**

z.ai が Anthropic-compat shim で `ANTHROPIC_REASONING_EFFORT` の値を実際に消費するかは、ライブ GLM セッションのアウトバウンド観測が必要な実証課題です。ランタイムの SSOT は `internal/template/glm_effort_overlay.go` です。

## 次のステップ

- [3-ティアエージェントアーキテクチャ](/ja/advanced/no-haiku-3tier/) — DeepSWE リーダーボードの根拠と 3-ティア定義
- [トークノミクス概論](/ja/advanced/tokenomics-overview/) — 4 層トークノミクス構造の Layer B ルーティング
- [モデルポリシー](/ja/multi-llm/model-policy/) — performance_tier のエイリアスと GLM バックエンドの詳細
