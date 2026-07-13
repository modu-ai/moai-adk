---
title: はじめる
description: /moai db init でプロジェクトのデータベースメタデータを初期化
weight: 10
draft: false
---

`/moai db init` を一度実行すれば、プロジェクトのデータベースメタデータが
整います。以降はマイグレーションファイルを追加するたびにスキーマドキュメントが
追随するため、エージェントもチームメンバーも常に最新のスキーマを同じ場所で確認できます。

## 前提条件

データベースワークフローを始める前に、次のものが必要です:

1. `/moai project` コマンドで生成された `.moai/project/product.md` と `.moai/project/tech.md` ファイル
2. サポートされているデータベースエンジン (PostgreSQL、MySQL、SQLite、MongoDB など)
3. ORM またはクエリビルダー (GORM、sqlc、Prisma、SQLAlchemy、ActiveRecord など)
4. マイグレーションツール (golang-migrate、Flyway、Liquibase、Alembic など)

## ステップバイステップ初期化ガイド

### ステップ 1: プロジェクトメタデータの確認

まず `/moai project` が作成した必須ファイルが存在するか確認します:

```bash
ls -la .moai/project/
# 次のファイルが存在する必要があります:
# - product.md
# - tech.md
# - structure.md
```

これらのファイルがない場合は、先に `/moai project` を実行してください。

### ステップ 2: データベースメタデータの初期化

次に `/moai db init` コマンドを実行します:

```bash
/moai db init
```

### ステップ 3: インタビュー質問への回答

MoAI は次の 4 項目について対話形式で質問します:

1. **データベースエンジン** — 使用中のデータベース (PostgreSQL、MySQL、SQLite、MongoDB など)
2. **ORM/クエリビルダー** — データアクセス層のツール
3. **マルチテナント戦略** — シングルスキーマ、テナントごとのスキーマ、テナントごとの DB、またはなし
4. **マイグレーションツール** — スキーマ変更管理ツール

各質問に対して適切なオプションを選択します。

### ステップ 4: 生成されたファイルのレビュー

初期化後、`.moai/project/db/` ディレクトリに次のファイルが生成されます:

```
.moai/project/db/
├── README.md              # DB セクションの概要
├── schema.md              # 自動生成されるテーブルレジストリ
├── erd.mmd                # エンティティ関係図
├── migrations.md          # マイグレーションファイルのインデックス
├── rls-policies.md        # Row-level security ルール (Supabase/Postgres)
├── queries.md             # 一般的なクエリライブラリ
└── seed-data.md           # シードデータパターン
```

各ファイルの役割:

- `schema.md` — すべてのテーブル、カラム、データ型、制約を自動でドキュメント化
- `erd.mmd` — Mermaid 記法でテーブル関係を可視化
- `migrations.md` — 適用済みマイグレーションファイルのタイムライン
- `queries.md` — AI エージェントが参照する一般的なクエリ例のコレクション

### ステップ 5: 最初のマイグレーションの作成と同期

新しいマイグレーションファイルをプロジェクトに追加します。例えば Go/golang-migrate の場合:

```bash
# db/migrations/ ディレクトリにマイグレーションファイルを作成
touch db/migrations/001_create_users_table.sql
```

マイグレーションファイルを作成したら、次のコマンドでスキーマドキュメントを更新します:

```bash
/moai db refresh
```

このコマンドは:
- すべてのマイグレーションファイルをスキャン
- schema.md に新しいテーブル情報を追加
- erd.mmd ダイアグラムを更新
- migrations.md タイムラインを更新

### ステップ 6: ドリフト検証 (任意)

ドリフトがあるかどうかを確認するには:

```bash
/moai db verify
```

結果:

- `スキーマドキュメントは同期されています` — マイグレーションとドキュメントが一致
- ドリフトレポートの出力 — 差分の詳細を表示 (exit code: 1)

## トラブルシューティング

### "Missing prerequisite files" エラー

`.moai/project/product.md` と `.moai/project/tech.md` がない場合:

```bash
/moai project
```

上記コマンドを先に実行してプロジェクトメタデータを生成してください。

### マイグレーションファイルが認識されない

プロジェクトの言語とマイグレーションツールが正しく検出されているか確認します:

```bash
cat .moai/config/sections/language.yaml
```

`language` フィールドを確認し、必要であれば `.moai/config/sections/db.yaml` で `migration_patterns` を手動で指定できます。

### 自動同期が動作しない

PostToolUse フックが正しく登録されているか確認します:

```bash
grep -A5 "PostToolUse" .claude/settings.json
```

フックがない場合は `/moai db init` を再実行するか、`.claude/settings.json` に手動で登録してください。
