---
title: /moai harness コマンド
weight: 55
draft: false
---

Harness v4 Builder でプロジェクト固有の動的専門家チームを生成・管理します。

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai:harness <自然言語リクエスト>` を入力するとこのコマンドを直接実行できます。
{{< /callout >}}

## 概要

`/moai:harness` は MoAI-ADK の **Harness v4 Builder** を実行してプロジェクト要求に合わせた動的専門家チームを自動生成します。

### Harness v4 Builder とは？

Harness v4 Builder は Socratic インタビューベースの 4-phase ワークフロー（ANALYZE → PLAN → GENERATE → ACTIVATE）でチームを構成します。

| 段階 | 説明 |
|------|------|
| ANALYZE | プロジェクト構造、使用言語、既存エージェントインベントリ分析 |
| PLAN | 必要なチーム規模（3～5 名）、各チームメンバーの役割、worktree 隔離の有無決定 |
| GENERATE | `.claude/agents/harness/` エージェントファイル、`.moai/harness/manifest.json` 生成 |
| ACTIVATE | チーム登録および `/harness:<name>` コマンド活性化 |

## 使用方法

### 1段階：自然言語でチーム生成リクエスト

```bash
> /moai:harness <自然言語リクエスト>
```

**例示:**
```
私たちの Go バックエンドプロジェクトに合わせた専門家チームを作成してください。
DB マイグレーション、REST API エンドポイント、単位テストをそれぞれ担当するチームが必要です。
```

### 2段階：Builder の自動処理

Builder が 4-phase を自動実行します:

1. **ANALYZE**: Go、PostgreSQL、REST API 技術スタック検知
2. **PLAN**: DB Engineer、API Developer、Test Engineer 3 人チーム構成決定
3. **GENERATE**: 
   - `.claude/agents/harness/db-engineer.md`
   - `.claude/agents/harness/api-developer.md`
   - `.claude/agents/harness/test-engineer.md`
   - `.moai/harness/manifest.json` 生成
4. **ACTIVATE**: `/harness:backend-team` コマンド登録

### 3段階：生成されたチーム活用

生成後すべての作業でチームを自動活用:

```bash
/moai run SPEC-BACKEND-001
/moai run --team SPEC-BACKEND-001    # チームモード強制
```

MoAI が SPEC 複雑度を分析して manifest の phase 順序通りにチームメンバーを自動委任します。

## Harness 管理コマンド

### harness list

生成されたすべてのハネス一覧表示:

```bash
/harness list
```

### harness:<name> status

特定ハネスの詳細情報:

```bash
/harness:backend-team status
```

出力情報:
- チームメンバーリストと役割
- 使用モデル（inherit、haiku、sonnet、opus）
- 選択的 worktree 隔離設定
- Manifest バージョンおよび生成日

### harness:<name> edit

manifest.json とエージェント定義編集:

```bash
/harness:backend-team edit
```

修正可能な項目:
- チームメンバー追加/削除
- スキル事前ロードリスト
- Worktree 隔離ポリシー
- 役割別プロンプト

### harness:<name> remove

ハネス及び関連ファイル削除:

```bash
/harness:backend-team remove
```

削除対象:
- `.claude/agents/harness/` エージェント定義
- `.moai/harness/manifest.json` ファイル
- 登録された `/harness:<name>` コマンド
- Worktree 隔離ポリシー

## Manifest 構造

Harness v4 は **manifest.json** でチーム構成を定義します。

### manifest.json 例

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
          "role": "DB 設計及びマイグレーション",
          "model": "inherit"
        },
        {
          "name": "api-developer",
          "role": "REST API エンドポイント",
          "model": "inherit"
        },
        {
          "name": "test-engineer",
          "role": "単位テスト",
          "model": "haiku"
        }
      ]
    }
  ]
}
```

### Phase フィールド

| フィールド | 説明 |
|----------|------|
| `name` | 段階名 (`plan`, `run`, `sync`) |
| `teammates` | この段階に参加するチームメンバー配列 |

### Teammate フィールド

| フィールド | デフォルト | 説明 |
|----------|--------|------|
| `name` | 必須 | チームメンバー固有識別子 |
| `role` | 必須 | チームメンバーの役割説明 |
| `model` | `inherit` | モデル選択 (`inherit`, `haiku`, `sonnet`, `opus`) |
| `skills` | `[]` | 事前ロードするスキル一覧 |

## Worktree 隔離

Harness v4 は選択的 worktree 隔離をサポートします。

### L1_optional (デフォルト)

```json
"worktree_isolation": "L1_optional"
```

Claude Code が並列チームメンバー間の競合を検出した場合、自動的に L1 ワークツリーを生成します。

- **選択的**: 競合時のみ隔離適用
- **自動**: ランタイムが競合検出後自動生成
- **コスト**: ワークツリー隔離時のメモリ増加

### none

```json
"worktree_isolation": "none"
```

すべてのチームメンバーがプロジェクトルートで作業します（最小メモリ使用）。

## チーム委任ワークフロー

Harness が活性化されると MoAI はそのチームを自動的に活用します。

### SPEC 実行時のチーム委任

```bash
> /moai run SPEC-BACKEND-001
```

**MoAI の自動判断:**
1. SPEC 複雑度推定（ファイル数、コード行数）
2. 適切なハネス選択
3. manifest phase 順序通りにチームメンバーを順次/並列委任

### Phase 基準の委任例

```
PLAN Phase:
  → architect チームメンバーがアーキテクチャ設計担当

RUN Phase:
  → db-engineer、api-developer 並列委任
  → test-engineer 順次委任（テスト）

SYNC Phase:
  → ドキュメント生成及び PR 作成（デフォルト manager-docs）
```

## 自然言語リクエストの力

Harness v4 Builder は Socratic インタビュー方式で要件を把握します。

### 効果的なリクエスト例

```
私たちのチーム は Python FastAPI バックエンドを開発中です。
API エンドポイント、データ検証、エラーハンドリングが得意なチームが必要です。
```

Builder が自動的に:
- Python、FastAPI、asyncio 技術スタック検知
- 3～5 名チーム規模決定
- 各チームメンバーの特化領域設定
- 必要なスキル事前ロード

### 不明確なリクエストは Builder が質問します

```
チームが必要です。

→ Builder: プロジェクトの主要技術は？（言語、フレームワーク）
→ Builder: チームが集中する領域は？（バックエンド、フロントエンド、全体）
→ Builder: 特別に必要な専門性は？
```

## 関連ドキュメント

- [Harness v4 Builder ガイド](/advanced/builder-agents) - Builder 4-phase 詳細
- [エージェントガイド](/advanced/agent-guide) - 8 つのコアエージェント理解
- [SPEC ベース開発](/workflow-commands/moai-plan) - SPEC ワークフロー概要

{{< callout type="info" >}}
**ヒント**: Harness を一度生成すると、すべての後続作業でそのチームが自動的に活用されます。`/harness:team-name` コマンドでいつでも再利用できます。
{{< /callout >}}
