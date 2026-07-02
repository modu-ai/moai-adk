---
title: カタログシステム
weight: 80
draft: false
---

3階層カタログマニフェストとslim initでプロジェクト初期化を最適化します。

## 概要

MoAI-ADK v2.15+のカタログシステムは、すべてのエージェント、スキル、プラグイン、ルールを **3階層
マニフェスト** で管理します。`moai init --slim`により、プロジェクトに必要な最小限のテンプレートのみを
デプロイし、初期化時間を短縮します。

## 3階層マニフェスト

| 階層 | 説明 | 配布基準 |
|------|------|----------|
| **Tier 1 (Core)** | 中核インフラ — オーケストレーター、品質ゲート、基本スキル | 常に配布 |
| **Tier 2 (Standard)** | 標準拡張 — 言語別ルール、フレームワークスキル | プロジェクトの言語/フレームワークを検出した場合 |
| **Tier 3 (Optional)** | 任意 — ドメインスキル、プラットフォーム別設定 | 明示的なリクエストまたはプロジェクト設定時 |

## カタログファイル

カタログマニフェストはYAML形式で定義されます:

```yaml
# カタログエントリ例
- id: moai-workflow-tdd
  tier: 1                    # 1=Core, 2=Standard, 3=Optional
  type: skill
  path: .claude/skills/moai/workflows/tdd.md
  languages: []              # 空の配列 = すべての言語
  frameworks: []
  hash: abc123...             # コンテンツハッシュ (整合性検証)
```

## SlimFSフィルタ

`moai init --slim`はSlimFSフィルタを通じて配布ファイルを制限します:

```bash
# フルインストール (すべての階層)
moai init my-project

# Slimインストール (Tier 1 + 検出されたTier 2のみ)
moai init --slim my-project
```

### フィルタロジック

1. Tier 1は常に含まれる
2. プロジェクトの言語を検出 (Go、Python、TypeScriptなど)
3. 検出された言語に対応するTier 2項目のみを含める
4. Tier 3は除外

## Typed Loader

`LoadCatalog()`関数がマニフェストを型安全にロードします:

- 3階層分類の検証
- ハッシュ整合性チェック (Hash Sentinel)
- 欠落フィールドの検出
- 100%のテストカバレッジ

## カタログの活用

### プロジェクト初期化

```bash
# 通常初期化 — すべてのテンプレートを配布
moai init my-project

# Slim初期化 — 最小限のテンプレートのみを配布
moai init --slim my-project
```

### アップデート

```bash
# カタログベースのアップデート
moai update                  # すべての階層を更新
moai update --slim           # slimモードで更新
```

## 関連ドキュメント

- [インストール](/ja/getting-started/installation) — インストールガイド
- [初期設定](/ja/getting-started/init-wizard) — initウィザード
- [アップデート](/ja/getting-started/update) — アップデートガイド
- [スキルガイド](/ja/advanced/skill-guide) — スキル作成ガイド
