---
title: /moai harness
weight: 55
draft: false
---

プロジェクト固有の動的専門家チーム (ハーネス) を生成し、ハーネス学習ライフサイクルを管理します。

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai:harness <自然言語リクエスト>` と入力すると、このコマンドをすぐに実行できます。
{{< /callout >}}

## 概要

`/moai:harness` は MoAI-ADK の **Harness v4 Builder** を実行し、プロジェクトの要件に合わせた動的専門家チームを自動生成します。

v3 の第 3 の柱である **エージェンティックハーネス** をそのまま体感できるコマンドです — ハーネスがハーネスを作る再帰構造です。汎用エージェントカタログでは足りないプロジェクト固有の領域 (例: 特定の DB マイグレーション手順、社内 API 規約) があれば、自然言語のひと言でその領域の専門家チームをスキャフォールドできます。生成されたハーネスは **再帰的自己学習** サブシステムにつながります — 使用の観察が蓄積されるとハーネスが自ら改善提案を作り、ユーザー承認ゲートを経て指針が進化します。

### Harness v4 Builder とは?

Harness v4 Builder は Socratic インタビューに基づく 4-phase ワークフロー (ANALYZE → PLAN → GENERATE → ACTIVATE) でチームを構成します。

| 段階 | 説明 |
|------|------|
| ANALYZE | プロジェクト構造、使用言語、既存エージェントのインベントリを分析 |
| PLAN | 必要なチーム規模 (3~5 名)、各メンバーの役割、worktree 分離の有無を決定 |
| GENERATE | `.claude/agents/harness/` エージェントファイル、`.moai/harness/manifest.json` を生成 |
| ACTIVATE | チーム登録と `/harness:<name>` コマンドの有効化 |

## 使用方法

### ステップ 1: 自然言語でチーム生成をリクエスト

```bash
> /moai:harness <自然言語リクエスト>
```

**例:**
```
うちの Go バックエンドプロジェクトに合う専門家チームを作って。
DB マイグレーション、REST API エンドポイント、ユニットテストをそれぞれ担当するチームが必要。
```

### ステップ 2: Builder の自動処理

Builder が 4-phase を自動実行します:

1. **ANALYZE**: Go、PostgreSQL、REST API の技術スタックを検出
2. **PLAN**: DB Engineer、API Developer、Test Engineer の 3 名チーム構成を決定
3. **GENERATE**:
   - `.claude/agents/harness/db-engineer.md`
   - `.claude/agents/harness/api-developer.md`
   - `.claude/agents/harness/test-engineer.md`
   - `.moai/harness/manifest.json` を生成
4. **ACTIVATE**: `/harness:backend-team` コマンドを登録

### ステップ 3: 生成されたチームの活用

生成後、すべての作業でチームを自動活用:

```bash
/moai run SPEC-BACKEND-001
/moai run --team SPEC-BACKEND-001    # チームモードを強制
```

MoAI が SPEC の複雑度を分析し、manifest の phase 順にメンバーへ自動委任します。

## Harness 管理コマンド

### harness list

生成されたすべてのハーネス一覧を照会:

```bash
/harness list
```

### harness:<name> status

特定ハーネスの詳細情報:

```bash
/harness:backend-team status
```

出力情報:
- メンバー一覧と役割
- 使用モデル (inherit, haiku, sonnet, opus)
- オプションの worktree 分離設定
- Manifest のバージョンと作成日

### harness:<name> edit

manifest.json とエージェント定義の編集:

```bash
/harness:backend-team edit
```

修正可能な項目:
- メンバーの追加/削除
- スキルの事前ロード一覧
- Worktree 分離ポリシー
- 役割ごとのプロンプト

### harness:<name> remove

ハーネスと関連ファイルの削除:

```bash
/harness:backend-team remove
```

削除される項目:
- `.claude/agents/harness/` エージェント定義
- `.moai/harness/manifest.json`
- 登録済み `/harness:<name>` コマンド
- ワークツリー分離ポリシー

## ハーネス学習ライフサイクル — 再帰的自己学習

ハーネスは生成して終わりの静的な成果物ではありません。`/moai harness` サブコマンドで **学習サブシステム** のライフサイクルを管理します。

| コマンド | 説明 |
|--------|------|
| `moai harness status` | 学習状態の確認 (観察数、パターン、提案) |
| `moai harness apply` | 提案の適用 (ユーザー承認ゲートの通過が必要) |
| `moai harness rollback` | 直前の適用をロールバック |
| `moai harness disable` | 学習の無効化 |
| `moai harness list` (v4) | すべての学習ルールの一覧 |
| `moai harness edit` (v4) | ルールの直接編集 |
| `moai harness remove` (v4) | ルールの削除 |
| `moai harness doctor` (v4) | 学習システムの診断 |

**4 層学習ラダー** — 観察が蓄積されるほど学習段階が上がります:

| Tier | 観察数 | 動作 |
|------|---------|------|
| TierObservation | ≥1 | 単純な記録 |
| TierHeuristic | ≥3 | パターン認識 |
| TierRule | ≥5 | ルール形成 |
| TierAutoUpdate | ≥10 | 自動更新 (ユーザー承認必須) |

**成果物**: `.moai/harness/` ディレクトリ (usage-log.jsonl, learned-rules.yaml)

{{< callout type="warning" >}}
自動進化は常に **ユーザー承認ゲート** の下でのみ適用されます。評価者と承認権限は進化ループの外にあり、いつでも `moai harness rollback` で復元できます。
{{< /callout >}}

## Manifest 構造

Harness v4 は **manifest.json** でチーム構成を定義します。

### manifest.json の例

```json
{
  "spec_id": "HARNESS-BACKEND-001",
  "name": "Backend Development Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "worktree_isolation": "L1_optional",
  
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "architect",
          "role": "API アーキテクチャ専門家",
          "model": "inherit",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "DB 設計とマイグレーション",
          "model": "inherit"
        },
        {
          "name": "api-developer",
          "role": "REST API エンドポイント",
          "model": "inherit"
        },
        {
          "name": "test-engineer",
          "role": "ユニットテスト",
          "model": "haiku"
        }
      ]
    }
  ]
}
```

### Phase フィールド

| フィールド | 説明 |
|------|------|
| `name` | 段階名 (`plan`, `run`, `sync`) |
| `teammates` | この段階に参加するメンバーの配列 |

### Teammate フィールド

| フィールド | デフォルト | 説明 |
|------|--------|------|
| `name` | 必須 | メンバーの一意識別子 |
| `role` | 必須 | メンバーの役割説明 |
| `model` | `inherit` | モデル選択 (`inherit`, `haiku`, `sonnet`, `opus`) |
| `skills` | `[]` | 事前ロードするスキルの一覧 |

メンバーごとにモデルを変えられること (`model` フィールド) はトークノミクス設計の延長です — アーキテクチャ決定のように推論が重い役割と、反復的なテスト作成のように軽い役割に同じモデルを使う理由はありません。

## Worktree 分離

Harness v4 はオプションの worktree 分離をサポートします。

### L1_optional (デフォルト)

```json
"worktree_isolation": "L1_optional"
```

Claude Code が並列メンバー間の衝突を検知すると、自動的に L1 ワークツリーを作成します。

- **オプション**: 衝突時のみ分離を適用
- **自動**: ランタイムが衝突検知後に自動作成
- **コスト**: ワークツリー分離時にメモリが増加

### none

```json
"worktree_isolation": "none"
```

すべてのメンバーがプロジェクトルートで作業します (最小メモリ使用)。

## チーム委任ワークフロー

Harness が有効化されると、MoAI はそのチームを自動的に活用します。

### SPEC 実行時のチーム委任

```bash
> /moai run SPEC-BACKEND-001
```

**MoAI の自動判断:**
1. SPEC の複雑度推定 (ファイル数、コード行数)
2. 適切なハーネスの選択
3. manifest の phase 順にメンバーを逐次/並列で委任

### Phase ベースの委任例

```
PLAN Phase:
  → architect メンバーがアーキテクチャ設計を担当

RUN Phase:
  → db-engineer, api-developer を並列委任
  → test-engineer を逐次委任 (テスト)

SYNC Phase:
  → ドキュメント生成と PR 作成 (デフォルト manager-docs)
```

## 自然言語リクエストの力

Harness v4 Builder は Socratic インタビュー方式で要件を把握します。

### 効果的なリクエスト例

```
うちのチームは Python FastAPI バックエンドを開発中です。
API エンドポイント、データ検証、エラーハンドリングが得意なチームが必要です。
```

Builder が自動で:
- Python、FastAPI、asyncio の技術スタックを検出
- 3~5 名のチーム規模を決定
- 各メンバーの特化領域を設定
- 必要なスキルを事前ロード

### 不明確なリクエストは Builder が質問します

```
チームが必要です。

→ Builder: プロジェクトの主要技術は? (言語、フレームワーク)
→ Builder: チームが集中する領域は? (バックエンド、フロントエンド、全体)
→ Builder: 特に必要な専門性は?
```

## 関連ドキュメント

- [Harness v4 Builder ガイド](/advanced/builder-agents) - Builder 4-phase の詳細
- [エージェントガイド](/advanced/agent-guide) - 10 エージェントカタログの理解
- [SPEC ベース開発](/workflow-commands/moai-plan) - SPEC ワークフローの概要

{{< callout type="info" >}}
**ヒント**: Harness を一度生成すれば、すべての後続作業でそのチームが自動的に活用されます。`/harness:team-name` コマンドでいつでも再利用できます。
{{< /callout >}}
