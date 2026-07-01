---
title: ビルダーエージェントと Harness v4
weight: 40
draft: false
---

MoAI-ADK 拡張のための Harness v4 Builder を詳細に解説します。

{{< callout type="info" >}}
  **一言でいうと**: Harness v4 Builder は自然言語リクエストでプロジェクト固有の専門家チームを動的に生成します。4 段階ワークフロー（ANALYZE → PLAN → GENERATE → ACTIVATE）と manifest ベース Runner で構成されます。
{{< /callout >}}

## Harness v4 Builder とは？

Harness v4 Builder は `/moai:harness <自然言語リクエスト>` を通じて**プロジェクト固有の専門家チームを動的に生成**します。

### 前バージョンとの違い

| 区分 | 前バージョン（v3/静的モデル） | 現在（v4 Builder） |
|------|-----|-----------|
| 生成方式 | 3 つのビルダーエージェント（ビルダー-スキル、ビルダー-エージェント、ビルダー-プラグイン） | 単一 Harness v4 Builder（動的生成） |
| ワークフロー | ユーザー定義構造 | 4-phase ANALYZE → PLAN → GENERATE → ACTIVATE |
| 実行方式 | それぞれ独立的 | Manifest ベース Runner（選択的ワークツリー格離） |
| 拡張性 | 限定的 | プロジェクトコンテキスト自動検知 |

## Harness v4 Builder 4-Phase ワークフロー

### 1. ANALYZE（分析段階）

現在のプロジェクトを分析し、必要な専門性を把握します。

- ソースコード構造分析
- 使用言語とフレームワーク検知
- 既存エージェント/スキルインベントリ調査
- プロジェクト規模推定

### 2. PLAN（計画段階）

必要な専門家チームの構成と役割を定義します。

- チーム規模決定（3～5 チームメンバー）
- 各チームメンバーの役割プロファイル定義
- ワークツリー格離必要性判断
- Manifest スキーマ設計

### 3. GENERATE（生成段階）

実際のエージェント定義と設定を生成します。

- `.claude/agents/harness/` 配下エージェントファイル生成
- `.moai/harness/manifest.json` 生成（Runner 設定）
- 役割別システムプロンプト作成
- スキル事前ロードリスト定義

### 4. ACTIVATE（活性化段階）

生成されたハネスを即座に使用可能にします。

- エージェント登録および検証
- Manifest Runner 初期化
- 選択的ワークツリー生成および格離設定
- チームメンバー自動委任ルール活性化

## Manifest ベース Runner

Harness v4 は**Manifest ベース Runner**を使用して生成されたチームを運用します。

### manifest.json 構造

```json
{
  "spec_id": "HARNESS-PROJECT-001",
  "name": "My Project Custom Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "researcher",
          "model": "haiku",
          "mode": "plan",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "implementer",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        }
      ]
    }
  ],
  "worktree_isolation": "L1_optional"
}
```

### Runner 動作

1. **Phase 進入**: manifest の phase シーケンスに従って進行
2. **Teammate Spawn**: 各 phase の teammates を動的に生成
3. **Isolation 適用**: 条件付きワークツリー格離適用
4. **Result Aggregation**: 各 teammate の結果を統合

## Harness ライフサイクル コマンド

Harness v4 Builder で生成されたハネスは `/harness:<name>` コマンドで管理されます。

### 使用可能なコマンド

```bash
# 生成されたハネス一覧表示
/harness list

# 特定ハネスの状態確認
/harness:my-project-team status

# ハネス設定編集
/harness:my-project-team edit

# ハネス削除
/harness:my-project-team remove

# Harness v4 Builder で新規ハネス生成
/moai:harness <自然言語リクエスト>
```

## 自然言語リクエストでハネス生成

### 基本使用法

```bash
> 私たちのバックエンドプロジェクトに合った専門家チームを作成してください。
> API 設計、DB スキーマ、テストを担当するチームが必要です。
```

### Builder の動作フロー

1. ANALYZE: プロジェクト構造（Go、PostgreSQL、REST API）を分析
2. PLAN: 3 人チーム（API Designer、DB Specialist、Test Engineer）決定
3. GENERATE: 各エージェント定義と manifest.json 生成
4. ACTIVATE: チーム活性化および `/harness:backend-team` コマンド登録

### 生成結果の場所

- エージェント定義: `.claude/agents/harness/api-designer.md`、`db-specialist.md`、...
- Manifest: `.moai/harness/manifest.json`
- 選択的ワークツリー: `~/.moai/worktrees/<project>/`（ユーザーオプトイン時）

## ワークツリー格離（選択的）

Harness v4 は条件付きワークツリー格離をサポートします。

### L1 格離（Optional）

Claude Code ランタイムがエージェント当たり L1 ワークツリーを生成します。

- **使用時期**: 並列チームメンバーが同じファイルを編集するとき
- **格離範囲**: 各チームメンバーのファイル書き込みが独立したワークツリーで発生
- **コスト**: 追加メモリ + 並列利益相殺

### 無効化

manifest の `"worktree_isolation": "none"` で L1 格離スキップ。

## 関連ドキュメント

- [Harness v4 Builder 詳細ガイド](/ja/advanced/harness-v4-builder) - Builder 4-phase 詳細および manifest スキーマ
- [エージェントガイド](/ja/advanced/agent-guide) - 8 つのコアエージェントカタログ
- [ダイナミックワークフロー](/ja/advanced/ultracode-workflows) - `/effort ultracode` 並列実行

{{< callout type="info" >}}
**ヒント**: Harness v4 Builder はプロジェクトごとに**カスタムチームを一度だけ生成**すれば、その後すべての作業で自動的にそのチームが委任されます。初回生成後は `/harness:team-name` でいつでも再利用できます。
{{< /callout >}}
