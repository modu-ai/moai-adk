---
title: Harness v4 Builder 深掘りガイド
weight: 45
draft: false
description: "Harness v4 Builderの4フェーズワークフロー(ANALYZE/PLAN/GENERATE/ACTIVATE)、Manifestスキーマ、Runnerプリミティブの動作ルール。"
---

[ビルダーエージェントガイド](/ja/advanced/builder-agents)が Harness v4 Builder の概要だったとすれば、このドキュメントは設計図です — 4-phase ワークフローの各段階の成果物、Manifest スキーマ全体、Runner プリミティブの動作ルールを扱います。

{{< callout type="info" >}}
**ひと言要約**: Harness v4 Builder は Socratic インタビューで必要な専門性を把握し、manifest ベースの Runner で動的チームを運用します。どのメンバーがどのモデルで働くかは、コードではなく manifest の宣言で決まります。
{{< /callout >}}

## 4-Phase Workflow 詳細

### Phase 1: ANALYZE (分析)

現在のプロジェクトの技術スタックと要求事項を分析します。このフェーズの目標は「このプロジェクトにどの専門性が不足しているか」をデータで答えることです。

#### 分析対象

- **プロジェクト構造**: ディレクトリ階層、コアパッケージの識別
- **使用言語**: Go、Python、TypeScript、Java などを検出
- **フレームワーク**: REST API、gRPC、FastAPI、Django などを認識
- **既存エージェント**: `.claude/agents/` の既存定義カタログ
- **プロジェクト規模**: ファイル数、コード行数ベースの推定
- **依存関係**: `go.mod`、`package.json`、`pyproject.toml` を分析

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

### Phase 2: PLAN (計画)

ANALYZE の結果をもとにチーム構成を設計します。チーム規模から役割別モデル割り当てまで、コストに影響する決定はすべてこのフェーズで下されます。

#### 計画の決定事項

| 項目 | 決定方式 | 例 |
|------|---------|------|
| **チーム規模** | プロジェクト複雑度 × 必要な専門性 | 3〜5 名 |
| **役割プロファイル** | Anthropic role_profiles (researcher/architect/implementer/tester/designer/reviewer) | architect, implementer, tester |
| **Worktree 分離** | 並列メンバーの衝突可能性 | L1_optional (選択的分離) |
| **モデル選択** | 役割別の推論複雑度 | architect: inherit, tester: haiku |
| **スキル事前ロード** | 役割の専門性に必要なスキル | moai-foundation-core, moai-domain-backend |

役割別のモデル選択がトークノミクスの核心です — 設計は深い推論が必要なモデルに、反復的なテスト作成は安価なモデルに割り当てます。

#### 計画の検証

生成前にユーザーへ確認します。承認ゲートなしでファイルが生成されることはありません。

```
計画されたチーム構成:
- チーム名: Backend Development Team
- メンバー 3 名:
  ① architect (model: inherit)
  ② implementer (model: inherit)
  ③ tester (model: haiku)
- Worktree 分離: L1_optional
- Manifest: .moai/harness/manifest.json

この構成で進めますか?
```

### Phase 3: GENERATE (生成)

PLAN 承認後、実際のエージェントファイルと manifest を生成します。

#### 生成物

**1. エージェント定義ファイル**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

各ファイルは YAML プロンプトで定義されます。

```yaml
---
name: architect
description: API アーキテクチャ設計の専門家
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

あなたはこのプロジェクトの API アーキテクチャ専門家です。
[役割別の詳細指針]
```

**2. Manifest ファイル**

```
.moai/harness/manifest.json
```

Phase と Teammate の定義を含む JSON です (スキーマは § Manifest スキーマ参照)。

#### 生成の検証

生成直後にファイルの存在と定義の正確性を直接確認できます。

```bash
ls .claude/agents/harness/
# architect.md, implementer.md, tester.md を確認

ls .moai/harness/
# manifest.json を確認

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# phase 定義が正確か確認
```

### Phase 4: ACTIVATE (有効化)

生成されたハーネスを登録し、すぐに使用可能にします。

#### 有効化の手順

1. **エージェント検証**: 各エージェントファイルの文法確認
2. **Manifest 検証**: JSON スキーマとフィールドの検証
3. **コマンド登録**: `/harness:backend-team` コマンドの有効化
4. **Runner 初期化**: Manifest ベースの Runner 開始準備
5. **Worktree 作成** (選択的): L1 分離の有効化条件を設定

#### 有効化の確認

```bash
/harness list
# backend-team が表示される

/harness:backend-team status
# メンバー 3 名、モデル、状態を確認
```

## Manifest スキーマ

### トップレベルフィールド

| フィールド | 型 | 必須 | 説明 |
|------|------|------|------|
| `spec_id` | string | はい | `HARNESS-{DOMAIN}-{NUM}` 形式 |
| `name` | string | はい | チーム表示名 |
| `version` | string | はい | Semantic versioning `X.Y.Z` |
| `created_at` | string | はい | ISO 8601 タイムスタンプ |
| `worktree_isolation` | enum | はい | `L1_optional` \| `none` |
| `phases` | array | はい | Phase オブジェクトの配列 |

### Phase オブジェクト

```json
{
  "name": "run",
  "description": "実装フェーズ",
  "teammates": [...]
}
```

| フィールド | 型 | 説明 |
|------|------|------|
| `name` | string | `plan` \| `run` \| `sync` |
| `description` | string | Phase 目標の説明 |
| `teammates` | array | Teammate オブジェクトの配列 |

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
|------|--------|------|
| `name` | 必須 | メンバー ID (ハイフン使用、空白なし) |
| `role` | 必須 | 役割の説明 (自由テキスト) |
| `model` | `inherit` | `inherit`, `haiku`, `sonnet`, `opus` |
| `mode` | `acceptEdits` | 権限モード (`acceptEdits`, `default`, `bypassPermissions`) |
| `skills` | `[]` | 事前ロードスキルの配列 (例: `["moai-foundation-core"]`) |
| `isolation` | なし | `worktree_optional` (worktree 分離の条件付き有効化) |

### 全体の例

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

Manifest ベースの Runner が生成されたチームを実行します。

### Runner のライフサイクル

```
Team Spawn
  ↓
[Phase 1: plan]
  → Teammate(architect) を生成して委任
  → 結果を収集
  ↓
[Phase 2: run]
  → Teammate(db-engineer) を並列生成
  → Teammate(api-developer) を並列生成
  → Teammate(test-engineer) を順次生成
  → 結果の収集と統合
  ↓
[Phase 3: sync]
  → デフォルトの manager-docs を実行
  ↓
Team Teardown
```

### Runner 設定

Runner の動作は manifest のフィールドで制御されます。

| 設定 | 意味 |
|------|------|
| `worktree_isolation: "L1_optional"` | 衝突検出時に自動分離を適用 |
| `worktree_isolation: "none"` | 分離を無効化 |
| `model: "inherit"` | 親セッションのモデルを継承 |
| `model: "haiku"` | Haiku モデルを強制 (コスト最適) |
| `skills: ["..."]` | 事前ロードスキル |

## Worktree 分離ルール

### L1_optional の動作

```
Runner 生成時:
├── メンバー 1: メインプロジェクトルート
├── メンバー 2: メインプロジェクトルート
└── 衝突検出時
    ├── メンバー 2 → L1 ワークツリーへ切り替え
    └── メンバー 1 はメイン維持 (またはメンバー 1 も切り替え)

結果:
└── ファイル衝突の回避 ✓
```

### 分離条件

次のいずれかが真であれば分離が有効化されます。

1. **同一ファイルの並列編集**: 2 人のメンバーが同じファイルを同時に修正
2. **再帰的なディレクトリ書き込み**: メンバーたちが同じディレクトリに複数ファイルを生成
3. **依存関係の競合**: メンバー A の出力がメンバー B の入力 (順序が重要)

### 非分離 (none) 選択時

```
すべてのメンバーがメインプロジェクトで作業
利点: 最小メモリ、高速な並列
欠点: 衝突の可能性
```

## 関連ドキュメント

- [Harness v4 Builder 使用ガイド](/ja/workflow-commands/moai-harness) - コマンドリファレンス
- [エージェントガイド](/ja/advanced/agent-guide) - エージェント定義フォーマット
- [SPEC ベース開発](/ja/workflow-commands/moai-plan) - Harness と SPEC の統合

{{< callout type="info" >}}
**ヒント**: Manifest は生成後 `/harness:team-name edit` でいつでも修正できます。メンバー追加、スキル変更、分離ポリシーの調整がすべて可能です。
{{< /callout >}}
