---
title: マイグレーションパターン
description: 16 プログラミング言語のデフォルトマイグレーションパスと設定
weight: 30
draft: false
---

MoAI-ADK は 16 のプログラミング言語を同等にサポートします。データベース
ワークフローも同様です — 言語ごとの業界標準マイグレーションツールのデフォルト
パスを把握しているため、ほとんどのプロジェクトは追加設定なしですぐにスキャンされます。

## サポートする言語とマイグレーションツール

| 言語 | マイグレーションツール | デフォルトパスパターン |
|------|-----------------|--------------|
| Go | golang-migrate | `db/migrations/*.sql` または `migrations/*.sql` |
| Python | Alembic | `alembic/versions/*.py` |
| TypeScript | Prisma Migrate | `prisma/migrations/**/*.sql` |
| JavaScript | Knex.js | `migrations/*.js` または `knexfile migrations/` |
| Rust | SQLx | `migrations/*.sql` |
| Java | Flyway | `src/main/resources/db/migration/V*.sql` |
| Kotlin | Flyway | `src/main/resources/db/migration/V*.sql` |
| C# | EF Core Migrations | `Migrations/*.cs` |
| Ruby | Rails ActiveRecord | `db/migrate/*.rb` |
| PHP | Laravel Migrations | `database/migrations/*.php` |
| Elixir | Ecto | `priv/repo/migrations/*.exs` |
| C++ | 標準なし (慣例) | `db/migrations/*.sql` |
| Scala | Slick / Flyway | `src/main/resources/db/migration/V*.sql` |
| R | 標準なし (慣例) | `migrations/*.sql` |
| Flutter | Drift | `assets/migrations/*.sql` |
| Swift | GRDB | `Resources/Migrations/*.sql` |

## 自動言語検出

MoAI は次の方法でプロジェクトの言語を自動検出します:

1. `.moai/config/sections/language.yaml` の `project_markers` を確認
2. プロジェクトルートの言語別マーカーファイルをスキャン:
   - Go: `go.mod`
   - Python: `pyproject.toml`, `setup.py`
   - TypeScript/JavaScript: `package.json`
   - Rust: `Cargo.toml`
   - Ruby: `Gemfile`
   - PHP: `composer.json`
   - Java/Kotlin: `pom.xml`, `build.gradle`
   - C#: `*.csproj`
   - Elixir: `mix.exs`

## カスタムマイグレーションパスの設定

デフォルトのパスがプロジェクトに合わない場合は、`.moai/config/sections/db.yaml` で手動指定できます:

```yaml
db:
  migration_patterns:
    - path: "custom/db/migrations"
      file_pattern: "*.sql"
      language: "go"
    - path: "backend/alembic/versions"
      file_pattern: "*.py"
      language: "python"
```

## 例: 各言語のマイグレーションファイル構造

### Go (golang-migrate)

```
project/
├── db/
│   ├── migrations/
│   │   ├── 001_create_users.up.sql
│   │   ├── 001_create_users.down.sql
│   │   ├── 002_add_email.up.sql
│   │   └── 002_add_email.down.sql
│   └── sqlc/
│       └── queries.sql
└── go.mod
```

### Python (Alembic)

```
project/
├── alembic/
│   ├── versions/
│   │   ├── 001_create_users.py
│   │   └── 002_add_email.py
│   ├── env.py
│   └── alembic.ini
└── pyproject.toml
```

### TypeScript (Prisma)

```
project/
├── prisma/
│   ├── migrations/
│   │   ├── 20240101120000_init/
│   │   │   └── migration.sql
│   │   └── 20240115143000_add_email/
│   │       └── migration.sql
│   └── schema.prisma
└── package.json
```

### Ruby (Rails)

```
project/
├── db/
│   ├── migrate/
│   │   ├── 20240101120000_create_users.rb
│   │   └── 20240115143000_add_email_to_users.rb
│   └── schema.rb
└── Gemfile
```

## マルチ言語プロジェクトの設定

マイクロサービスやモノレポ構成で複数言語のマイグレーションをまとめて管理する場合:

```yaml
db:
  migration_patterns:
    # バックエンド (Go)
    - path: "services/api/db/migrations"
      file_pattern: "*.sql"
      language: "go"

    # データパイプライン (Python)
    - path: "services/analytics/alembic/versions"
      file_pattern: "*.py"
      language: "python"

    # Web アプリケーション (TypeScript)
    - path: "apps/web/prisma/migrations"
      file_pattern: "*.sql"
      language: "typescript"
```

## マイグレーションツール選択ガイド

### Prisma (TypeScript/JavaScript)

長所:
- シンプルな文法
- 自動型生成
- 直感的なリレーション定義

短所:
- Prisma エコシステムに依存
- 複雑なマイグレーションに制限

### Alembic (Python)

長所:
- 自動マイグレーション生成機能
- 柔軟なカスタマイズ
- SQLAlchemy との完全統合

短所:
- 学習曲線
- 初期設定が複雑

### Flyway (Java/Kotlin)

長所:
- 言語別マイグレーションのサポート
- 強力な検証
- ウォーターマークシステム

短所:
- 設定の複雑さ
- パフォーマンスオーバーヘッド

### golang-migrate (Go)

長所:
- 軽量で高速
- Up/Down の明確な区別
- 純粋な SQL を使用

短所:
- 補助機能なし
- 自動生成不可

## マイグレーションファイルの命名規則

ツールごとの推奨命名規則:

| ツール | 規則 | 例 |
|------|------|------|
| golang-migrate | `YYYYMMDDHHMMSS_description.up.sql` | `20240101120000_create_users.up.sql` |
| Alembic | `rev_<hash>_description.py` | `rev_a001b002_add_email.py` |
| Prisma | タイムスタンプフォルダ | `20240101120000_init` |
| Flyway | `V<version>__description.sql` | `V1__Create_users.sql` |
| Rails | `YYYYMMDDHHMMSS_description.rb` | `20240101120000_create_users.rb` |
| Laravel | `YYYY_MM_DD_HHMMSS_description.php` | `2024_01_01_120000_create_users.php` |

## トラブルシューティング

### マイグレーションファイルがスキャンされない

1. マイグレーションパスの確認:

```bash
ls -la $(path/to/migrations)
```

2. ファイルパターンの確認 — ファイル拡張子が期待されるパターンと一致するか確認

3. 言語検出の確認:

```bash
cat .moai/config/sections/language.yaml
```

4. カスタムパスの設定:

```yaml
db:
  migration_patterns:
    - path: "your/custom/path"
      file_pattern: "*.sql"
```

### 複数言語の設定が衝突する

サービスごとのパスを明確に分離します:

```yaml
db:
  migration_patterns:
    - path: "services/api/**"          # バックエンド専用
    - path: "apps/web/**"              # Web アプリ専用
```
