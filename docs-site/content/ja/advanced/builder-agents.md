---
title: ビルダーエージェントとハーネス v4
weight: 40
draft: false
---

エージェンティック・ハーネスの最後のピースは再帰です — ハーネスがハーネスを作ります。Harness v4 Builder は、自然言語のリクエストひと言でプロジェクト固有の専門家チームを生成する、その再帰構造の入り口です。

{{< callout type="info" >}}
**ひと言要約**: Harness v4 Builder は自然言語のリクエストでプロジェクト固有の専門家チームを動的に生成します。4 フェーズワークフロー (ANALYZE → PLAN → GENERATE → ACTIVATE) と manifest ベースの Runner で構成されます。
{{< /callout >}}

## Harness v4 Builder とは?

Harness v4 Builder は `/moai:harness <自然言語リクエスト>` を通じて **プロジェクト固有の専門家チームを動的に生成** します。

汎用エージェントカタログ (10 個) がすべてのプロジェクトに共通だとすれば、Builder が作るハーネスはあなたのプロジェクトにだけ存在するカスタムチームです。

### 以前のバージョンとの違い

| 区分 | 以前 (v3/静的モデル) | 現在 (v4 Builder) |
|------|-----|-----------|
| 生成方式 | 3 種類のビルダーエージェント (ビルダー・スキル、ビルダー・エージェント、ビルダー・プラグイン) | 単一の Harness v4 Builder (動的生成) |
| ワークフロー | ユーザー定義構造 | 4-phase ANALYZE → PLAN → GENERATE → ACTIVATE |
| 実行方式 | それぞれ独立 | Manifest ベースの Runner (選択的な worktree 分離) |
| 拡張性 | 限定的 | プロジェクトコンテキストの自動検出 |

## Harness v4 Builder 4-Phase Workflow

### 1. ANALYZE (分析フェーズ)

現在のプロジェクトを分析し、必要な専門性を把握します。

- ソースコード構造の分析
- 使用言語とフレームワークの検出
- 既存エージェント/スキルのインベントリ調査
- プロジェクト規模の推定

### 2. PLAN (計画フェーズ)

必要な専門家チームの構成と役割を定義します。

- チーム規模の決定 (3〜5 メンバー)
- 各メンバーの役割プロファイル定義
- worktree 分離の必要性判断
- Manifest スキーマ設計

### 3. GENERATE (生成フェーズ)

実際のエージェント定義と設定を生成します。

- `.claude/agents/harness/` 配下にエージェントファイルを生成
- `.moai/harness/manifest.json` を生成 (Runner 設定)
- 役割別システムプロンプトの作成
- スキル事前ロードリストの定義

### 4. ACTIVATE (有効化フェーズ)

生成されたハーネスをすぐに利用できるよう有効化します。

- エージェントの登録と検証
- Manifest Runner の初期化
- 選択的な worktree 作成と分離設定
- チームメンバーの自動委任ルールの有効化

## Manifest ベースの Runner

Harness v4 は **Manifest ベースの Runner** を使って生成されたチームを運用します。どの phase にどのメンバーが、どのモデルと権限モードで投入されるかが manifest 1 ファイルに宣言されます — モデル割り当てを宣言で管理するトークノミクス原則がここにも適用されます。

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

### Runner の動作

1. **Phase 進入**: manifest の phase シーケンスに沿って進行
2. **Teammate Spawn**: 各 phase の teammates を動的に生成
3. **Isolation 適用**: 条件付き worktree 分離を適用
4. **Result Aggregation**: 各 teammate の結果を統合

## Harness Lifecycle Commands

Harness v4 Builder で生成されたハーネスは `/harness:<name>` コマンドで管理されます。

### 利用可能なコマンド

```bash
# 生成されたハーネスの一覧表示
/harness list

# 特定ハーネスの状態確認
/harness:my-project-team status

# ハーネス設定の編集
/harness:my-project-team edit

# ハーネスの削除
/harness:my-project-team remove

# Harness v4 Builder で新しいハーネスを生成
/moai:harness <自然言語リクエスト>
```

## 自然言語リクエストでハーネスを生成

### 基本的な使い方

```bash
> うちのバックエンドプロジェクトに合った専門家チームを作って。
> API 設計、DB スキーマ、テストを担当するチームが必要。
```

### Builder の動作フロー

1. ANALYZE: プロジェクト構造 (Go、PostgreSQL、REST API) を分析
2. PLAN: 3 人チーム (API Designer、DB Specialist、Test Engineer) を決定
3. GENERATE: 各エージェント定義と manifest.json を生成
4. ACTIVATE: チームを有効化し `/harness:backend-team` コマンドを登録

### 生成結果の場所

- エージェント定義: `.claude/agents/harness/api-designer.md`, `db-specialist.md`, ...
- Manifest: `.moai/harness/manifest.json`
- 選択的なワークツリー: `~/.moai/worktrees/<project>/` (ユーザー opt-in 時)

## Worktree 分離 (選択的)

Harness v4 は条件付き worktree 分離をサポートします。

### L1 分離 (Optional)

Claude Code ランタイムがエージェントごとに L1 ワークツリーを作成します。

- **使用タイミング**: 並列メンバーが同じファイルを編集するとき
- **分離範囲**: 各メンバーのファイル書き込みが独立したワークツリーで発生
- **コスト**: 追加メモリ + 並列メリットの相殺

### 無効化

manifest の `"worktree_isolation": "none"` に設定すると L1 分離を省略します。

## 関連ドキュメント

- [Harness v4 Builder 深掘りガイド](/ja/advanced/harness-v4-builder) - Builder 4-phase 詳細と manifest スキーマ
- [エージェントガイド](/ja/advanced/agent-guide) - 10 個のコアエージェントカタログ
- [動的ワークフロー](/ja/advanced/ultracode-workflows) - `/effort ultracode` 並列実行

{{< callout type="info" >}}
**ヒント**: Harness v4 Builder でプロジェクトごとに **カスタムチームを一度だけ生成** すれば、以降のすべての作業でそのチームが自動的に委任されます。初回生成後は `/harness:team-name` でいつでも再利用できます。
{{< /callout >}}
