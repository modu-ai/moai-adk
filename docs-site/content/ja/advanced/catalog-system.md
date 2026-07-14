---
title: カタログシステム
weight: 80
draft: false
---

トークノミクスはトークンだけに適用される原則ではありません。プロジェクトに配布されるテンプレートファイル一つひとつも、結局はセッションがロードすることになるコンテキスト候補です。カタログシステムは「必要なものだけを配布する」という原則で、このコストを初期化ステップの段階から減らします。

## 概要

MoAI-ADK のカタログシステムは、すべてのエージェント、スキル、ルールを **3 階層マニフェスト** (`catalog.yaml`) で管理します。デフォルト配布は **slim モード** でコアテンプレート (core) のみを配布するため、初期化が速く、プロジェクトに残るファイルも軽くなります。全配布が必要なら `--all` フラグを使います。

## 3 階層マニフェスト

すべての配布対象は 3 つの階層のいずれかに属します。

| 階層 | catalog.yaml キー | 説明 | 配布基準 |
|------|-----------------|------|----------|
| **Core** | `catalog.core` | コアインフラ — オーケストレーター、品質ゲート、基本スキル/エージェント | 常に配布 (slim モードのデフォルト) |
| **Optional Packs** | `catalog.optional_packs` | ドメイン拡張 — backend, frontend, design, devops, deployment, testing パック | `--all` フラグ時に配布 |
| **Harness-generated** | `catalog.harness_generated` | ハーネスが動的生成したエージェント/スキル | `--all` フラグ時に配布 |

## カタログファイル

カタログマニフェストは `internal/template/catalog.yaml` に YAML 形式で定義されます。

```yaml
catalog:
  core:                        # 常に配布 (slim モードのデフォルト)
    skills:
      - name: moai-workflow-tdd
        tier: core
        path: templates/.claude/skills/moai-workflow-tdd/
        hash: 6f89fb72...      # コンテンツハッシュ (整合性検証)
        version: 1.0.0
    agents:
      - name: manager-spec
        tier: core
        path: templates/.claude/agents/moai/manager-spec.md
        hash: a1b2c3d4...
        version: 1.0.0
  optional_packs:              # --all フラグ時に配布
    backend:
      - name: moai-domain-backend
        tier: optional-pack:backend
        path: templates/.claude/skills/moai-domain-backend/
        hash: ...
    frontend:
      - name: moai-domain-frontend
        tier: optional-pack:frontend
        path: templates/.claude/skills/moai-domain-frontend/
        hash: ...
  harness_generated:           # --all フラグ時に配布
    skills: []
    agents:
      - name: builder-harness
        tier: harness-generated
        path: templates/.claude/agents/moai/builder-harness.md
        hash: ...
```

各エントリは `name`、`tier`、`path`、`hash`、`version` フィールドを持ちます。`hash` フィールドがコンテンツハッシュを保持しているため、配布されたファイルが破損したり任意に変わったりしていないかをローダーが検証できます。スキルディレクトリ内部のエントリポイントファイルは `SKILL.md` です (小文字の `skill.md` ではありません)。

## Slim モードと --all フラグ

デフォルト配布は **slim モード** で `catalog.core` のみを配布します。全配布が必要なら `--all` フラグまたは `MOAI_DISTRIBUTE_ALL=1` 環境変数を使います。

```bash
# Slim インストール (デフォルト — core のみ)
moai init my-project

# 全インストール (core + optional_packs + harness_generated)
moai init --all my-project

# 環境変数で全インストール
MOAI_DISTRIBUTE_ALL=1 moai init my-project
```

### 配布ロジック

配布は 2 段階で動作します。

1. `catalog.core` (skills + agents) は常に含まれる — slim モードのデフォルト
2. `--all` フラグまたは `MOAI_DISTRIBUTE_ALL=1` 環境変数が設定されている場合、`catalog.optional_packs` と `catalog.harness_generated` を追加配布

## Typed Loader

`LoadCatalog()` 関数がマニフェストを型安全にロードします。文字列パースに依存せず構造体単位で検証するため、マニフェストのエラーは配布前にふるい落とされます。

- 3 階層分類の検証
- ハッシュ整合性検査 (Hash Sentinel)
- 欠落フィールドの検出
- 100% テストカバレッジ

## カタログの活用

### プロジェクト初期化

```bash
# デフォルト初期化 — core のみ配布 (slim モード)
moai init my-project

# 全初期化 — core + optional_packs + harness_generated
moai init --all my-project
```

### アップデート

`moai update` は同じカタログを基準に動作します。slim で初期化したプロジェクトは core のみ、`--all` で初期化したプロジェクトは全体をアップデートします。

```bash
# カタログベースのアップデート
moai update                  # 初期化モードに応じて自動決定
```

## 関連ドキュメント

- [インストール](/ja/getting-started/installation) — インストールガイド
- [初期設定](/ja/getting-started/init-wizard) — init ウィザード
- [アップデート](/ja/getting-started/update) — アップデートガイド
- [スキルガイド](/ja/advanced/skill-guide) — スキル作成ガイド
