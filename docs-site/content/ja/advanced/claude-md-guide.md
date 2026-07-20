---
title: CLAUDE.md ガイド
weight: 80
draft: false
---

Claude Code の核心指針ファイル体系を詳しく案内します。`CLAUDE.md` は毎セッション常にロードされるファイルなので、このファイルの 1 行 1 行がそのまま常時コンテキストのコストです — 指針体系の設計はハーネス設計であると同時にトークノミクスでもあります。

{{< callout type="info" >}}
**一行要約**: `CLAUDE.md` はプロジェクトの **憲法** です。Claude Code がプロジェクトをどう理解し、どんなルールに従い、どのエージェントを呼び出すか、すべてこのファイルで決まります。
{{< /callout >}}

## CLAUDE.md とは?

`CLAUDE.md` は Claude Code がセッションを開始するとき **最初に読む指針ファイル** です。このファイルにプロジェクトのルール、エージェント構造、ワークフロー、品質基準などが定義されています。

人が新しい会社に入社すると社員ハンドブックを読むように、Claude Code はセッションを開始するとき `CLAUDE.md` を読んでプロジェクトの文脈を把握します。

## ファイル構造

MoAI-ADK は 2 つの指針ファイルとルールディレクトリを使います。

```mermaid
flowchart TD
    subgraph MAIN["CLAUDE.md (プロジェクトレベル)"]
        M1["核心アイデンティティ"]
        M2["リクエスト処理パイプライン"]
        M3["コマンドリファレンス"]
        M4["エージェントカタログ"]
        M5["SPEC ワークフロー"]
        M6["品質ゲート"]
    end

    subgraph LOCAL["CLAUDE.local.md (個人レベル)"]
        L1["個人ルール"]
        L2["MDX ガイドライン"]
        L3["プロジェクトメモ"]
    end

    subgraph RULES[".claude/rules/ (条件付きルール)"]
        R1["core/ 核心原則"]
        R2["development/ 開発標準"]
        R3["workflow/ ワークフロー"]
        R4["languages/ 言語別ルール"]
    end

    MAIN --> LOCAL
    MAIN --> RULES

```

| ファイル/ディレクトリ | 用途 | Git 追跡 | アップデート時 |
|---------------|------|----------|-------------|
| `CLAUDE.md` | MoAI-ADK 核心指針 | はい | 上書き |
| `CLAUDE.local.md` | 個人カスタム指針 | いいえ | 保存 |
| `.claude/rules/moai/` | 条件付き詳細ルール | はい | 上書き |
| `.claude/rules/local/` | 個人カスタムルール | いいえ | 保存 |

## MoAI CLAUDE.md の主要セクション

### 1. 核心アイデンティティ

MoAI オーケストレーターの役割と HARD ルールを定義します。

```markdown
## 1. 核心アイデンティティ

MoAI は Claude Code の戦略的オーケストレーターです。

### HARD ルール (必須)
- [HARD] 言語認識応答: ユーザーの conversation_language で応答
- [HARD] 並列実行: 独立したツール呼び出しは並列実行
- [HARD] XML タグ非表示: ユーザー対面応答に XML を非表示
- [HARD] Markdown 出力: すべてのコミュニケーションに Markdown を使用
```

### 2. リクエスト処理パイプライン (Analyze-First)

すべてのリクエストは入力言語と無関係に 1 つの順序化されたパイプラインを経ます。v3.0 の核心は **意図分析が常に先** という点です — 英語のキーワードマッチングではなく言語独立的な意味分類でルーティングします。

| ステップ | 説明 |
|------|------|
| 1. 意図分析 | リクエストの意図を言語独立的に分類 (Analyze-First) |
| 2. コンテキスト十分性の点検 | 不足なら実行前に Socratic インタビューで確認 |
| 3. 実行計画の構成 | スキル/エージェント/ワークフローのチェーン + オーケストレーションモードの選択 |
| 4. 承認ゲート | 実装着手承認 (plan→run ヒューマンゲート) を含む |
| 5. 実行 → 検証 → 反復 | 受け入れ基準に対する検証、goal が設定されると goal 評価者が終了判定 |

### 3. コマンドリファレンス

`/moai` がすべての MoAI 開発ワークフローの単一の進入点です。

| タイプ | コマンド | 用途 |
|------|--------|------|
| SPEC パイプライン | `/moai plan`, `/moai run`, `/moai sync` | 3-phase 開発ワークフロー |
| ループ/修正 | `/moai goal`, `/moai loop`, `/moai fix` | 条件宣言ループ、反復修正、単発修正 |
| プロジェクト/ハーネス | `/moai project`, `/moai harness` | プロジェクトドキュメント + ハーネス生成/管理 |
| 品質/ユーティリティ | `/moai review`, `/moai gate`, `/moai clean`, `/moai mx`, `/moai codemaps`, `/moai feedback` | レビュー、ゲート、整理、注釈、ドキュメント、報告 |
| (自然言語) | `/moai "リクエスト"` | Analyze-First ルーティング → 自律パイプライン |

### 4. エージェントカタログ

MoAI-ADK は **11 個の保存エージェント** (10 個 MoAI-custom + 1 個 Anthropic built-in) で構成されます。アーキテクチャの単純化により manager-strategy, manager-quality, manager-brain, manager-project など 12 個の archived エージェントは特定ドメインに対する per-spawn `Agent(general-purpose)` delegation に置き換えられました。

| 分類 | エージェント | 役割 |
|------|----------|------|
| Manager (5) | manager-spec, manager-develop, manager-docs, manager-git, manager-design | 核心ライフサイクルのステップ別専門家 |
| Evaluator (2) | plan-auditor, sync-auditor | 計画/完了ステップの独立した品質評価 |
| Builder (1) | builder-harness | 動的なプロジェクト別ハーネス生成 |
| Advisor (1) | super-advisor | 高推論の助言 (E1-E4 エスカレーション) |
| Specialist (1) | e2e-specialist | Web/モバイル/デスクトップの E2E テスト実行 (`/moai e2e`) |
| Built-in (1) | Explore (Anthropic) | 読み取り専用のコードベース探索 |

### 5. SPEC ワークフロー

3 段階の SPEC ベース開発ワークフローを定義します。

```bash
# Plan: SPEC ドキュメント生成 (30K トークン)
> /moai plan "機能の説明"

# Run: DDD 実装 (180K トークン)
> /moai run SPEC-XXX

# Sync: ドキュメント同期 (40K トークン)
> /moai sync SPEC-XXX

# E2E: Web/モバイル/デスクトップの E2E テスト実行
> /moai e2e
```

### 6. 品質ゲート

TRUST 5 フレームワークと LSP 品質ゲートを定義します。

| 品質基準 | 要件 |
|-----------|----------|
| Tested | 85%+ カバレッジ、LSP 型エラー 0 |
| Readable | 明確な名前、LSP リントエラー 0 |
| Unified | 一貫したスタイル、LSP 警告 10 以下 |
| Secured | OWASP 準拠、LSP セキュリティ警告 0 |
| Trackable | 明確なコミット、LSP 状態の追跡 |

### 7. ユーザー相互作用アーキテクチャ

下位エージェントはユーザーと直接対話できません。ユーザー接点は MoAI 一つに固定されます。

```mermaid
flowchart TD
    USER["ユーザー"] --> MOAI["MoAI"]
    MOAI -->|"1. 情報収集"| USER
    MOAI -->|"2. 作業委任"| AGENT["下位エージェント"]
    AGENT -->|"3. 結果を返す"| MOAI
    MOAI -->|"4. 結果報告"| USER

    AGENT -.-x|"直接対話不可"| USER

```

### 8. 構成リファレンス

言語設定、ユーザー設定、プロジェクトルールを参照します。

```yaml
language:
  conversation_language: ko           # ユーザー応答言語
  agent_prompt_language: en           # エージェント内部言語
  git_commit_messages: en             # Git コミットメッセージ
  code_comments: en                   # コードコメント
  documentation: en                   # ドキュメントファイル
```

## CLAUDE.local.md の活用法

`CLAUDE.local.md` は個人的なルールとメモを書くファイルです。MoAI-ADK アップデートと無関係に保存されます。

### 記述例

```markdown
# プロジェクトローカル設定

## ドキュメント作成ガイドライン

### MDX レンダリングエラーの防止
- 強調表示と括弧の間に必ず空白

### Mermaid ダイアグラムの方向
- すべてのダイアグラムは縦方向 (flowchart TD)

## 個人メモ
- DB マイグレーション前にバックアップ必須
- API エンドポイントの命名: kebab-case を使用
```

### 活用のヒント

| 用途 | 内容の例 |
|------|-----------|
| コーディングルール | 「変数名は camelCase、ファイル名は kebab-case」 |
| プロジェクトメモ | 「認証は JWT、有効期限 24 時間、更新 7 日」 |
| 禁止事項 | 「console.log をプロダクションコードに残さないこと」 |
| 好みのパターン | 「React コンポーネントは関数型のみ使用」 |
| MDX ルール | 「強調と括弧の間に空白必須」 |

## .claude/rules/ システム

`.claude/rules/` ディレクトリには **条件付きでロードされる詳細ルール** が保存されます。すべてのルールを CLAUDE.md に入れず条件付きファイルに分離する理由はただ 1 つ — 使わないルールがコンテキストを占めないようにするためです。

### ディレクトリ構造

```
.claude/rules/moai/
├── core/                          # 核心原則
│   └── moai-constitution.md       # TRUST 5、核心ルール
├── development/                   # 開発標準
│   ├── skill-authoring.md         # スキル作成ガイド
│   └── coding-standards.md        # コーディング標準
├── workflow/                      # ワークフロー
│   └── spec-workflow.md           # SPEC ワークフロー (Plan/Run/Sync 定義)
└── languages/                     # 言語別ルール (16 個)
    ├── python.md
    ├── typescript.md
    ├── javascript.md
    └── ...
```

### 条件付きロード (paths frontmatter)

ルールファイルは `paths` フロントマターを通じて **特定のファイル作業時にのみロード** されます。

```yaml
---
paths:
  - "**/*.py"
  - "**/pyproject.toml"
---

# Python コーディングルール
- ruff フォーマッターを使用
- type hints 必須
- docstring は Google スタイル
```

このルールは Python ファイルを修正するときだけロードされて **トークンを節約** します。

### ルールファイルの種類

| ディレクトリ | ファイル | ロード条件 |
|----------|------|-----------|
| `core/` | `moai-constitution.md` | 常にロード |
| `development/` | `skill-authoring.md` | スキル関連の作業時 |
| `development/` | `coding-standards.md` | コード作業時 |
| `workflow/` | `spec-workflow.md` | ワークフローコマンド・SPEC 関連の作業時 |
| `languages/` | `python.md` など | 該当言語ファイルの修正時 |

## サイズ制限

`CLAUDE.md` は **40,000 文字以下** を維持する必要があります。MoAI-ADK 自体も v3 期間中に CLAUDE.md を継続的にダイエットしてきました — 常時ロード指針は短いほどすべてのセッションが安くなります。

### サイズ超過時の対応方法

```mermaid
flowchart TD
    CHECK{"CLAUDE.md<br>40,000 文字超過?"}

    CHECK -->|はい| MOVE["詳細内容を<br>.claude/rules/ へ移動"]
    CHECK -->|いいえ| OK["正常維持"]

    MOVE --> REF["CLAUDE.md には<br>参照のみ残す"]
    REF --> SLIM["核心ルールのみ<br>CLAUDE.md に維持"]
```

**対応戦略:**

1. **詳細内容の移動**: 長い説明は `.claude/rules/` ファイルに分離
2. **参照の使用**: `CLAUDE.md` で `@ファイルパス` で参照
3. **核心のみ維持**: アイデンティティ、HARD ルール、エージェントカタログのみ維持
4. **スキルへの転換**: 長いパターンの説明はスキルに変換

## 実践例: CLAUDE.local.md のカスタムルール

### フロントエンドプロジェクト

```markdown
# プロジェクトローカル設定

## React ルール
- コンポーネントは必ず関数型で記述
- Props インターフェースはコンポーネントファイルの上部に定義
- 状態管理は Zustand を使用
- CSS は Tailwind CSS のみ使用

## 命名ルール
- コンポーネント: PascalCase (UserProfile.tsx)
- ユーティリティ: camelCase (formatDate.ts)
- 定数: UPPER_SNAKE_CASE (MAX_RETRY_COUNT)
- API エンドポイント: kebab-case (/api/user-profiles)

## 禁止事項
- any 型の使用禁止
- console.log をプロダクションコードに禁止
- default export 禁止 (named export のみ使用)
```

### バックエンドプロジェクト

```markdown
# プロジェクトローカル設定

## Python ルール
- FastAPI を使用
- 非同期関数を優先 (async/await)
- Pydantic v2 モデルを使用
- SQLAlchemy 2.0 スタイル

## データベースルール
- マイグレーション前に必ずバックアップ
- インデックスはクエリパターン分析後に追加
- soft delete パターンを使用 (is_deleted フラグ)

## API ルール
- RESTful エンドポイントの命名
- 応答形式の統一: {"data": ..., "message": ...}
- エラーコードの標準化
```

## CLAUDE.md, rules, skills の関係

指針体系は 4 階層に分かれ、下の階層に行くほどロード条件が狭くなります。

```mermaid
flowchart TD
    subgraph HIERARCHY["指針体系の階層"]
        CLAUDE["CLAUDE.md<br>最上位指針 (常にロード)"]
        RULES[".claude/rules/<br>条件付きルール (paths マッチ時)"]
        SKILLS[".claude/skills/<br>専門知識 (トリガーマッチ時)"]
        AGENTS[".claude/agents/<br>エージェント定義 (委任時)"]
    end

    CLAUDE --> RULES
    RULES --> SKILLS
    SKILLS --> AGENTS

    CLAUDE -.->|"参照"| RULES
    AGENTS -.->|"スキル使用"| SKILLS

```

| 階層 | ファイル | ロードタイミング | 役割 |
|------|------|-----------|------|
| 1. CLAUDE.md | `CLAUDE.md` | 常に | プロジェクトのアイデンティティ、核心ルール |
| 2. Rules | `.claude/rules/*.md` | ファイルパターンのマッチ時 | 条件付き詳細ルール |
| 3. Skills | `.claude/skills/*/skill.md` | トリガーマッチ時 | 専門知識、パターン |
| 4. Agents | `.claude/agents/*.md` | 委任時 | 専門家の役割定義 |

## 関連ドキュメント

- [スキルガイド](/ja/advanced/skill-guide) - スキルシステムの詳細
- [エージェントガイド](/ja/advanced/agent-guide) - エージェントシステムの詳細
- [settings.json ガイド](/ja/advanced/settings-json) - 設定ファイルの管理
- [Hooks ガイド](/ja/advanced/hooks-guide) - イベント自動化

{{< callout type="info" >}}
**ヒント**: `CLAUDE.md` を直接修正するより `CLAUDE.local.md` に個人ルールを追加することを推奨します。MoAI-ADK アップデート時にも個人ルールが安全に保存されます。
{{< /callout >}}
