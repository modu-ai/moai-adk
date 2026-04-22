---
title: 始める
description: /moai db init でプロジェクトのデータベースメタデータを初期化
weight: 10
draft: false
---

## 前提条件

データベースワークフローを開始する前に、以下が必要です:

1. `/moai project` で生成された `.moai/project/product.md` と `.moai/project/tech.md` ファイル
2. サポートされるデータベースエンジン (PostgreSQL、MySQL、SQLite、MongoDB など)
3. ORM またはクエリビルダー (GORM、sqlc、Prisma、SQLAlchemy、ActiveRecord など)
4. マイグレーションツール (golang-migrate、Flyway、Liquibase、Alembic など)

## ステップバイステップ初期化ガイド

### ステップ 1: プロジェクトメタデータを確認

まず必要なファイルが存在するか確認します:

```bash
ls -la .moai/project/
# これらのファイルが存在する必要があります:
# - product.md
# - tech.md
# - structure.md
```

これらのファイルが存在しない場合は、まず `/moai project` を実行してください。

### ステップ 2: データベースメタデータを初期化

次に `/moai db init` コマンドを実行します:

```bash
/moai db init
```

### ステップ 3: インタビュー質問に回答

MoAI は 4つの対話型質問をします:

1. **データベースエンジン** — 使用中のデータベース (PostgreSQL、MySQL、SQLite、MongoDB など)
2. **ORM/クエリビルダー** — データアクセスレイヤーツール
3. **マルチテナンシー戦略** — シングルスキーマ、テナント単位スキーマ、テナント単位DB、またはなし
4. **マイグレーションツール** — スキーマ変更管理ツール

各質問に対して適切なオプションを選択します。

### ステップ 4: 生成されたファイルを確認

初期化後、`.moai/project/db/` ディレクトリに以下のファイルが作成されます:

```
.moai/project/db/
├── README.md              # DB セクション概要
├── schema.md              # 自動生成テーブルレジストリ
├── erd.mmd                # エンティティ関係図
├── migrations.md          # マイグレーションファイルインデックス
├── rls-policies.md        # Row-level security ルール (Supabase/Postgres)
├── queries.md             # 共通クエリライブラリ
└── seed-data.md           # シードデータパターン
```

ファイルの説明:

- `schema.md` — すべてのテーブル、列、データタイプ、制約を自動的に文書化
- `erd.mmd` — Mermaid 構文でテーブル関係を可視化
- `migrations.md` — 適用されたマイグレーションのタイムライン
- `queries.md` — AI エージェントが参照する共通クエリ例

### ステップ 5: 最初のマイグレーションを作成して同期

プロジェクトに新しいマイグレーションファイルを追加します。例えば、Go/golang-migrate の場合:

```bash
# db/migrations/ ディレクトリにマイグレーションファイルを作成
touch db/migrations/001_create_users_table.sql
```

マイグレーションを作成してから、スキーマドキュメントをリフレッシュします:

```bash
/moai db refresh
```

このコマンド:
- すべてのマイグレーションファイルをスキャン
- schema.md に新しいテーブル情報を追加
- erd.mmd ダイアグラムを更新
- migrations.md タイムラインをリフレッシュ

### ステップ 6: ドリフトを確認する (オプション)

ドリフトの有無を確認します:

```bash
/moai db verify
```

結果:

- `スキーマドキュメントが同期されています` — マイグレーションとドキュメントが一致
- ドリフトレポート出力 — 差異の詳細を表示 (exit code: 1)

## トラブルシューティング

### 「Missing prerequisite files」エラー

`.moai/project/product.md` と `.moai/project/tech.md` が存在しない場合:

```bash
/moai project
```

このコマンドを実行してプロジェクトメタデータを生成してください。

### マイグレーションファイルが認識されない

プロジェクトの言語とマイグレーションツールが正しく検出されているか確認します:

```bash
cat .moai/config/sections/language.yaml
```

`language` フィールドを確認してください。必要に応じて `.moai/config/sections/db.yaml` で `migration_patterns` を手動指定できます。

### 自動同期が機能しない

PostToolUse フックが正しく登録されているか確認します:

```bash
grep -A5 "PostToolUse" .claude/settings.json
```

フックが見つからない場合は、`/moai db init` を再度実行するか、`.claude/settings.json` に手動登録してください。
