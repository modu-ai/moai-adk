# エージェントシステムリファレンス

MoAI-ADKの19人のAIエージェントチーム構造を理解してください。

## 概要

Alfredは**19人の専門家チーム**を調整するスーパーエージェントです。各エージェントは特定のドメインまたはタスクに最適化されており、必要に応じて自動的にアクティベーションされます。

## 19人のチーム構造

### 組織図

```
┌─────────────────────────────────────────┐
│          Alfred（スーパーエージェント）    │
│         SPEC → TDD → Syncオーケストレーション │
└──────────────────┬──────────────────────┘
                   │
        ┌──────────┼──────────┐
        │          │          │
   ┌────▼─────┐  ┌▼────────┐  ┌▼────────────┐
   │  コア    │  │ エキスパート│  │ ビルトイン   │
   │ エージェント│  │エージェント│  │ エージェント │
   │  (10人)   │  │ (6人)    │  │ (2人)      │
   └──────────┘  └────────┘  └────────────┘
```

## エージェント分類

### 1️⃣ コアサブエージェント（10人）

プロジェクト全体のライフサイクルを管理します：

| エージェント                   | 役割                     | アクティベーション条件          |
| ----------------------------- | ------------------------ | ----------------------------- |
| **project-manager**           | プロジェクト初期化および設定 | `/alfred:0-project`  |
| **spec-builder**              | SPEC作成（EARS構文）     | `/alfred:1-plan`     |
| **implementation-planner**    | アーキテクチャおよび実装計画 | `/alfred:2-run`開始 |
| **tdd-implementer**          | RED→GREEN→REFACTOR実行   | `/alfred:2-run`中   |
| **doc-syncer**                | ドキュメント自動生成および同期 | `/alfred:3-sync`     |
| **tag-agent**                 | TAG検証および追跡可能性管理 | `/alfred:3-sync`     |
| **git-manager**               | Gitワークフロー自動化    | すべての段階            |
| **trust-checker**             | TRUST 5原則検証          | `/alfred:2-run`完了 |
| **quality-gate**              | リリース準備状態確認      | `/alfred:3-sync`     |
| **debug-helper**              | エラー分析および解決      | 必要時に自動アクティベーション |

### 2️⃣ エキスパートエージェント（6人）

ドメイン特化タスクをサポートします：

| エージェント            | ドメイン                        | アクティベーション条件              |
| ----------------------- | ----------------------------- | --------------------------------- |
| **backend-expert**      | API、サーバー、DBアーキテクチャ | SPECにサーバー/APIキーワード   |
| **frontend-expert**     | UI、状態管理、パフォーマンス    | SPECにフロントエンドキーワード |
| **devops-expert**       | デプロイ、CI/CD、インフラ       | SPECにデプロイキーワード       |
| **ui-ux-expert**        | デザインシステム、アクセシビリティ | SPECにデザインキーワード     |
| **security-expert**     | セキュリティ分析、脆弱性診断    | SPECにセキュリティキーワード |
| **database-expert**     | DB設計、最適化、マイグレーション | SPECにDBキーワード         |

### 3️⃣ ビルトインClaudeエージェント（2人）

複雑な推論が必要な場合：

- **Claude Opus/Sonnet**: 複雑な推論、深い分析
- **Claude Haiku**: 軽量タスク、高速処理

## ハイブリッドパターン

### Lead-Specialistパターン

専門化されたドメインエキスパートがリードエージェントをサポートします：

```
ユーザーリクエスト
    ↓
Alfred（リード）
    ├─→ フロントエンドキーワード検出
    │   └─→ frontend-expertアクティベーション
    ├─→ データベースキーワード検出
    │   └─→ database-expertアクティベーション
    └─→ セキュリティキーワード検出
        └─→ security-expertアクティベーション
```

**使用ケース**：

- UIコンポーネント設計が必要 → UI/UXエキスパート
- DBパフォーマンス最適化 → データベースエキスパート
- セキュリティレビュー → セキュリティエキスパート

### Master-Cloneパターン

大規模タスクはAlfredクローンが並列で処理します：

```
大規模タスク（100+ファイル、5+ステップ）
    ↓
マスターAlfred（調整）
    ├─→ クローン1: モジュールAリファクタリング
    ├─→ クローン2: モジュールBリファクタリング
    └─→ クローン3: モジュールCリファクタリング
    ↓
結果マージおよび統合
```

**使用ケース**：

- 大規模マイグレーション（v1.0 → v2.0）
- 全体アーキテクチャリファクタリング
- 複数ドメイン同時作業

## エージェント協力方法

### 順次協力（Sequential）

```
Alfred → spec-builder → implementation-planner → tdd-implementer → doc-syncer
```

### 並列協力（Parallel）

```
Alfred
├─→ backend-expert（API設計）
├─→ frontend-expert（UI設計）
└─→ database-expert（DBスキーマ）
    ↓
すべて完了後、tdd-implementer実行
```

### 条件付き協力（Conditional）

```
Alfred
└─→ SPEC分析
    ├─ セキュリティキーワード？ → security-expertアクティベーション
    ├─ デプロイ必要？ → devops-expertアクティベーション
    └─ パフォーマンス最適化？ → debug-helperアクティベーション
```

## エージェントアクティベーションダイアグラム

```
User Request
    ↓
┌─────────────────────────────────┐
│   Alfred（意図検出）              │
├─────────────────────────────────┤
│ 1. リクエスト分類                 │
│ 2. ドメイン検出                   │
│ 3. 必要エージェント判断            │
└──────────────┬──────────────────┘
               │
        ┌──────┴──────┐
        │             │
    コアエージェント？  ドメインエキスパート？
        │             │
        ▼             ▼
    ┌────────┐    ┌──────────────┐
    │project │    │backend-expert│
    │manager │    │frontend-expert
    └────────┘    └──────────────┘
        ↓             ↓
    ┌────────────────────┐
    │  TDD実行           │
    │ (tdd-implementer)  │
    └────────────────────┘
        ↓
    ┌────────────────────┐
    │  検証               │
    │ (trust-checker)    │
    └────────────────────┘
        ↓
    ┌────────────────────┐
    │  ドキュメント同期   │
    │ (doc-syncer)       │
    └────────────────────┘
        ↓
    完了
```

## エージェント選択アルゴリズム

Alfredがどのエージェントをアクティベーションするか決定する方法：

```python
# 意思決定ツリー
def select_agents(user_request):
    # 1. ドメイン検出
    if "database" or "schema" or "query" in request:
        activate(database_expert)

    if "frontend" or "ui" or "component" in request:
        activate(frontend_expert)

    if "security" or "auth" or "encryption" in request:
        activate(security_expert)

    if "deploy" or "ci/cd" or "docker" in request:
        activate(devops_expert)

    # 2. 規模判断
    if file_count > 100 or steps > 5:
        use_master_clone_pattern()
    else:
        use_lead_specialist_pattern()

    # 3. コアエージェントは常にアクティベーション
    activate(core_agents)
```

## <span class="material-icons">library_books</span> 詳細ガイド

- **[コアサブエージェント](core.md)** - 10人のエージェント詳細
- **[エキスパートエージェント](experts.md)** - 6人のエキスパート詳細

## 関連ドキュメント

- [Alfredスーパーエージェント](guides/alfred/index.md) - Alfred概念とワークフロー
- [スキルシステム](skills/index.md) - 55個以上のClaudeスキル
- [アーキテクチャ説明](advanced/architecture.md) - 4層スタック構造

______________________________________________________________________

**次**: [コアサブエージェント](core.md)または[エキスパートエージェント](experts.md)



