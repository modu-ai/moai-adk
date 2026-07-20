---
title: "@MX TAGシステム"
weight: 61
draft: false
---

@MX TAGはコードレベルのアノテーションであり、AIエージェントが開発セッションをまたいで **コンテキスト・不変量・危険ゾーン** を伝達する標準手段です。プロンプトは無視される可能性がありますが、コードに刻まれたコメントはコードと共に生き残り、次のエージェントがコードを初めて読む瞬間に意図と制約を即座に把握できます。

> @MX TAGの運用 (スキャン · 追加 · 問い合わせ) は `/moai mx` コマンドで行います。このページはタグシステム自体のプロトコルとライフサイクルを扱います。

## タグ構文

```go
// @MX:TAG_TYPE: [説明]
// @MX:SUB_KEY: [サブ値]
```

タグはインラインのソースコメントであり、独立したJSON台帳ではありません。`grep` または `moai mx query` で収集されます。

## タグタイプ

| タグ | 用途 | 必須サブライン |
|------|------|----------------|
| `@MX:NOTE` | コンテキストと意図の伝達 | — |
| `@MX:WARN` | 危険ゾーンの標識 | `@MX:REASON` |
| `@MX:ANCHOR` | 不変契約 (高fan_in) | `@MX:REASON` |
| `@MX:TODO` | 未完了作業 | — |
| `@MX:DEBT` | 意図的単純化 (動作するコード) | `@MX:CEILING` + `@MX:UPGRADE` |

## サブライン

`@MX:SPEC` · `@MX:LEGACY` · `@MX:REASON` · `@MX:TEST` · `@MX:PRIORITY` · `@MX:CEILING` · `@MX:UPGRADE`

- `@MX:REASON` は WARN · ANCHOR に **必須** です。
- `[AUTO]` 接頭辞はエージェント生成タグに **必須** です。

## 追加タイミング

**@MX:NOTE** — マジック定数、100行超のexported関数にgodocがない場合、説明のないビジネスルール。

**@MX:WARN** — `context.Context` のないgoroutine/channel、循環的複雑度15以上、グローバル状態変更、if分岐8個以上。

**@MX:ANCHOR** — fan_in 3以上、公開API境界、外部システム統合ポイント。

**@MX:TODO** — テストファイルのない公開関数、未実装のSPEC要件、処理なしで返されるエラー。

**@MX:DEBT** — 意図的単純化を採用し、明示された限界 (`@MX:CEILING`) 内で正確であり、再訪トリガー (`@MX:UPGRADE`) が存在する場合。

## DEBT — 動作する単純化の明示的限界

`@MX:DEBT` は未完了作業の標識ではありません。コードは **すでに完成し正確に動作** していますが、明示された限界内での意図的単純化であることを記録します。2つのサブラインが続きます。

```go
// @MX:DEBT: in-memory map cache, no eviction
// @MX:CEILING: < 10k entries
// @MX:UPGRADE: switch to LRU when entry count exceeds 10k
```

`@MX:UPGRADE` のないDEBTは終了条件がなく、**静かに腐敗** (rot) します。`moai mx query --kind DEBT --json` はこれを `"rotRisk": "no-trigger"` として表示します。腐敗シグナルは `@MX:UPGRADE` の不在であり、`@MX:CEILING` の不在は品質メモに過ぎず腐敗の基準ではありません。

> `@MX:TODO` はGREENステップで解決される未完了作業 (コードはまだ未完成) を、`@MX:DEBT` は完成し正確に動作するが明示的限界を持つ単純化 (コードは完成) を標識します。DEBTは複数のGREENステップにわたって正常に維持でき、TODOの「3回未解決でWARN昇格」ルールは適用されません。

## 更新・削除タイミング

- **ANCHOR** — fan_in変化またはSPEC更新時に更新。自動削除禁止、レポートでNOTE降格。
- **NOTE** — 関数シグネチャ変更時に再レビュー。
- **WARN** — 危険構造の改善時に削除。
- **TODO** — 解決時 (テスト通過または実装完了) に削除。3回反復未解決でWARNへ昇格。
- **DEBT** — 限界またはトリガー変化時に更新。`@MX:UPGRADE` トリガー発火で単純化が置換される時に削除され、他の作業完了とは無関係です。自動昇格なし。

## ライフサイクル要約

```text
TODO     RED/ANALYZEで生成 → GREEN/IMPROVEで解決 (削除) → 3回未解決でWARN昇格
ANCHOR   fan_in ≥ 3で生成 → 呼び出し数·SPEC変化で更新 → fan_in < 3でNOTE降格 (レポート) → 自動削除なし
WARN     危険検出で生成 → 構造的なら持続 → 解決で削除
NOTE     コンテキスト必要で生成 → シグネチャ変更後に更新 → コード削除で廃棄
DEBT     意図的単純化で生成 → UPGRADEトリガー発火で解決 (単純化を置換) → 自動昇格なし
```

## 言語別コメント構文

| 言語 | 接頭辞 | 例 |
|------|--------|------|
| Go · Java · TS · Rust · C/C++ · Swift · Kotlin · Dart · Zig · Scala | `//` | `// @MX:NOTE:` |
| Python · Ruby · Elixir | `#` | `# @MX:WARN:` |
| Haskell | `--` | `-- @MX:ANCHOR:` |

## 設定 (`.moai/config/sections/mx.yaml`)

- **thresholds** — `fan_in_anchor`, `complexity_warn`, `branch_warn`
- **limits** — `anchor_per_file` (デフォルト 3), `warn_per_file` (デフォルト 5)。超過時、ANCHORは最低fan_inから降格、WARNはP1–P5優先のみ保持。
- **exclude** — `**/*_generated.go`, `**/vendor/**`, `**/mock_*.go` などのタグ付け除外パターン
- **require_reason_for** — REASONが必須のタグタイプ

## タグ言語

タグ説明と `@MX:REASON` は `.moai/config/sections/language.yaml` の `code_comments` 設定に従います (デフォルト `en`)。韓国語プロジェクトなら `code_comments: ko` と設定すればタグは韓国語で記述されます。

## 次のステップ

- [Hooksガイド](/ja/advanced/hooks-guide) — フックと共にコードコンテキストを扱う基盤
- [SPECベース開発](/ja/core-concepts/spec-based-dev) — SPECライフサイクルと@MX TAG連携
- [TRUST 5品質フレームワーク](/ja/core-concepts/trust-5) — Readable原則と@MX:NOTE
