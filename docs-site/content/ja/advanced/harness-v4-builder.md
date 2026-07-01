---
title: Harness v4 Builder 深掘りガイド
weight: 45
draft: false
---

Harness v4 Builder の4フェーズワークフロー、Manifest スキーマ、Runner プリミティブについて詳しくガイドします。

{{< callout type="info" >}}
**一行要約**: Harness v4 Builder はSocratic インタビューで必要な専門性を把握し、manifest ベースの Runner で動的チームを運用します。
{{< /callout >}}

## 4-Phase Workflow 詳細

### Phase 1: ANALYZE（分析）

現在のプロジェクトの技術スタックと要件を分析します。

#### 分析対象

- **プロジェクト構造**: ディレクトリ階層、コアパッケージ識別
- **使用言語**: Go、Python、TypeScript、Java など検出
- **フレームワーク**: REST API、gRPC、FastAPI、Django など認識
- **既存エージェント**: `.claude/agents/` 既存定義カタログ
- **プロジェクト規模**: ファイル数、コード行数ベースの推定
- **依存関係**: `go.mod`、`package.json`、`pyproject.toml` 分析

#### 成果物

```yaml
analysis_result:
  languages:
    - go (primary)
    - shell (build scripts)
  frameworks:
    - REST API (net/http)
    - PostgreSQL ORM (sqlc)
  scale: "100~300 files, ~50K LOC"
  existing_agents: 0
  expertise_gaps:
    - Database schema design
    - API error handling patterns
    - Test coverage automation
```

### Phase 2: PLAN（計画）

ANALYZE 結果に基づいてチーム構成を設計します。

#### 計画決定事項

| 項目 | 決定方式 | 例示 |
|------|---------|------|
| **チームサイズ** | プロジェクト複雑度 × 必要専門性 | 3~5名 |
| **役割プロファイル** | Anthropic role_profiles（researcher/architect/implementer/tester/designer/reviewer） | architect、implementer、tester |
| **Worktree 隔離** | 並行チームメンバー衝突の可能性 | L1_optional（オプション隔離） |
| **モデル選択** | 役割別推論複雑度 | architect: inherit、tester: haiku |
| **スキル事前ロード** | 役割専門性必要スキル | moai-foundation-core、moai-domain-backend |

#### 計画検証

生成前にユーザーに確認:

```
計画されたチーム構成:
- チーム名: Backend Development Team
- チームメンバー 3名:
  ① architect（model: inherit）
  ② implementer（model: inherit）
  ③ tester（model: haiku）
- Worktree 隔離: L1_optional
- Manifest: .moai/harness/manifest.json

この構成で進めますか?
```

### Phase 3: GENERATE（生成）

PLAN 承認後に実際のエージェントファイルと manifest を生成します。

#### 生成成果物

**1. エージェント定義ファイル**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

各ファイルはYAML プロンプトで定義:

```yaml
---
name: architect
description: API アーキテクチャ設計専門家
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

あなたはこのプロジェクトの API アーキテクチャ専門家です。
[役割別詳細指針]
```

**2. Manifest ファイル**

```
.moai/harness/manifest.json
```

Phase と Teammate 定義を含む JSON（スキーマは § Manifest スキーマ参照）。

#### 生成検証

```bash
ls .claude/agents/harness/
# architect.md、implementer.md、tester.md を確認

ls .moai/harness/
# manifest.json を確認

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# phase 定義が正確か確認
```

### Phase 4: ACTIVATE（活性化）

生成されたハーネスを登録し、即座に使用可能にします。

#### 活性化ステップ

1. **エージェント検証**: 各エージェントファイル構文確認
2. **Manifest 検証**: JSON スキーマおよびフィールド検証
3. **コマンド登録**: `/harness:backend-team` コマンド活性化
4. **Runner 初期化**: Manifest ベースの Runner 開始準備
5. **Worktree 生成**（オプション）: L1 隔離活性化条件設定

#### 活性化確認

```bash
/harness list
# backend-team が表示される

/harness:backend-team status
# チームメンバー3名、モデル、ステータス確認
```

## Manifest スキーマ

### 最上位フィールド

| フィールド | 型 | 必須 | 説明 |
|-----------|------|------|------|
| `spec_id` | string | はい | `HARNESS-{DOMAIN}-{NUM}` 形式 |
| `name` | string | はい | チーム表示名 |
| `version` | string | はい | Semantic versioning `X.Y.Z` |
| `created_at` | string | はい | ISO 8601 タイムスタンプ |
| `worktree_isolation` | enum | はい | `L1_optional` \| `none` |
| `phases` | array | はい | Phase オブジェクト配列 |

### Phase オブジェクト

```json
{
  "name": "run",
  "description": "実装フェーズ",
  "teammates": [...]
}
```

| フィールド | 型 | 説明 |
|-----------|------|------|
| `name` | string | `plan` \| `run` \| `sync` |
| `description` | string | Phase 目標説明 |
| `teammates` | array | Teammate オブジェクト配列 |

### Teammate オブジェクト

```json
{
  "name": "api-developer",
  "role": "REST API エンドポイント開発",
  "model": "inherit",
  "mode": "acceptEdits",
  "skills": ["moai-foundation-core"],
  "isolation": "worktree_optional"
}
```

| フィールド | デフォルト | 説明 |
|-----------|--------|------|
| `name` | 必須 | チームメンバー ID（ハイフン使用、スペースなし） |
| `role` | 必須 | 役割説明（自由テキスト） |
| `model` | `inherit` | `inherit`、`haiku`、`sonnet`、`opus` |
| `mode` | `acceptEdits` | 権限モード（`acceptEdits`、`default`、`bypassPermissions`） |
| `skills` | `[]` | 事前ロードスキル配列（例: `["moai-foundation-core"]`） |
| `isolation` | なし | `worktree_optional`（worktree 隔離条件付き活性化） |

### 完全な例示

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
      "description": "アーキテクチャ設計と SPEC 作成",
      "teammates": [
        {
          "name": "architect",
          "role": "API アーキテクチャ専門家",
          "model": "inherit",
          "mode": "acceptEdits",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "description": "実際の実装",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "DB 設計とマイグレーション",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "api-developer",
          "role": "REST API エンドポイント実装",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "test-engineer",
          "role": "単体テストと統合テスト",
          "model": "haiku",
          "mode": "acceptEdits"
        }
      ]
    }
  ]
}
```

## Runner プリミティブ

Manifest ベースの Runner は生成されたチームを実行します。

### Runner ライフサイクル

```
Team Spawn
  ↓
[Phase 1: plan]
  → Teammate(architect) 生成と委譲
  → 結果収集
  ↓
[Phase 2: run]
  → Teammate(db-engineer) 並行生成
  → Teammate(api-developer) 並行生成
  → Teammate(test-engineer) 順序生成
  → 結果収集と統合
  ↓
[Phase 3: sync]
  → 基本 manager-docs 実行
  ↓
Team Teardown
```

### Runner 設定

Runner の動作は manifest のフィールドで制御されます:

| 設定 | 意味 |
|------|------|
| `worktree_isolation: "L1_optional"` | 衝突検出時に自動隔離適用 |
| `worktree_isolation: "none"` | 隔離無効化 |
| `model: "inherit"` | 親セッションモデル継承 |
| `model: "haiku"` | Haiku モデル強制（コスト最適） |
| `skills: ["..."]` | 事前ロードスキル |

## Worktree 隔離ルール

### L1_optional 動作

```
Runner 生成時:
├── チームメンバー 1: メインプロジェクトルート
├── チームメンバー 2: メインプロジェクトルート
└── 衝突検出時
    ├── チームメンバー 2 → L1 ワークツリーに転換
    └── チームメンバー 1 はメイン維持（または両方転換）

結果:
└── ファイル衝突回避 ✓
```

### 隔離条件

次のいずれかに該当する場合、隔離を活性化:

1. **同一ファイル並行編集**: 2つのチームメンバーが同じファイルを同時に修正
2. **再帰的ディレクトリ書き込み**: チームメンバーが同じディレクトリに複数ファイル作成
3. **依存関係競合**: チームメンバー A の出力がチームメンバー B の入力（順序重要）

### 非隔離（none）選択時

```
すべてのチームメンバーがメインプロジェクトで作業
利点: 最小メモリ、高速並行
欠点: 衝突の可能性
```

## 関連ドキュメント

- [Harness v4 Builder 使用ガイド](/workflow-commands/moai-harness) - コマンドリファレンス
- [エージェントガイド](/advanced/agent-guide) - エージェント定義形式
- [SPEC ベース開発](/workflow-commands/moai-plan) - Harness と SPEC 統合

{{< callout type="info" >}}
**ヒント**: Manifest は生成後に `/harness:team-name edit` でいつでも修正できます。チームメンバー追加、スキル変更、隔離ポリシー調整がすべて可能です。
{{< /callout >}}
