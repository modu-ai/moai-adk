---
title: Migration Patterns
description: Default migration paths and configuration for 16 programming languages
weight: 30
draft: false
---

MoAI-ADK supports 16 programming languages equally. The database workflow is
no exception — it knows the default paths of each language's industry-standard
migration tool, so most projects scan out of the box with no extra
configuration.

## Supported languages and migration tools

| Language | Migration tool | Default path pattern |
|------|-----------------|--------------|
| Go | golang-migrate | `db/migrations/*.sql` or `migrations/*.sql` |
| Python | Alembic | `alembic/versions/*.py` |
| TypeScript | Prisma Migrate | `prisma/migrations/**/*.sql` |
| JavaScript | Knex.js | `migrations/*.js` or `knexfile migrations/` |
| Rust | SQLx | `migrations/*.sql` |
| Java | Flyway | `src/main/resources/db/migration/V*.sql` |
| Kotlin | Flyway | `src/main/resources/db/migration/V*.sql` |
| C# | EF Core Migrations | `Migrations/*.cs` |
| Ruby | Rails ActiveRecord | `db/migrate/*.rb` |
| PHP | Laravel Migrations | `database/migrations/*.php` |
| Elixir | Ecto | `priv/repo/migrations/*.exs` |
| C++ | No standard (convention) | `db/migrations/*.sql` |
| Scala | Slick / Flyway | `src/main/resources/db/migration/V*.sql` |
| R | No standard (convention) | `migrations/*.sql` |
| Flutter | Drift | `assets/migrations/*.sql` |
| Swift | GRDB | `Resources/Migrations/*.sql` |

## Automatic language detection

MoAI auto-detects the project language as follows:

1. Check `project_markers` in `.moai/config/sections/language.yaml`
2. Scan the project root for language-specific marker files:
   - Go: `go.mod`
   - Python: `pyproject.toml`, `setup.py`
   - TypeScript/JavaScript: `package.json`
   - Rust: `Cargo.toml`
   - Ruby: `Gemfile`
   - PHP: `composer.json`
   - Java/Kotlin: `pom.xml`, `build.gradle`
   - C#: `*.csproj`
   - Elixir: `mix.exs`

## Custom migration path configuration

If the default path does not match your project, specify it manually in
`.moai/config/sections/db.yaml`:

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

## Examples: migration file structure per language

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

## Multi-language project configuration

For microservices or a monorepo managing migrations across several languages:

```yaml
db:
  migration_patterns:
    # Backend (Go)
    - path: "services/api/db/migrations"
      file_pattern: "*.sql"
      language: "go"

    # Data pipeline (Python)
    - path: "services/analytics/alembic/versions"
      file_pattern: "*.py"
      language: "python"

    # Web application (TypeScript)
    - path: "apps/web/prisma/migrations"
      file_pattern: "*.sql"
      language: "typescript"
```

## Migration tool selection guide

### Prisma (TypeScript/JavaScript)

Pros:
- Simple syntax
- Automatic type generation
- Intuitive relationship definitions

Cons:
- Depends on the Prisma ecosystem
- Limited for complex migrations

### Alembic (Python)

Pros:
- Auto-generation of migrations
- Flexible customization
- Full SQLAlchemy integration

Cons:
- Learning curve
- Complex initial setup

### Flyway (Java/Kotlin)

Pros:
- Language-specific migration support
- Strong validation
- Watermark system

Cons:
- Configuration complexity
- Performance overhead

### golang-migrate (Go)

Pros:
- Lightweight and fast
- Clear Up/Down separation
- Pure SQL

Cons:
- No helper features
- No auto-generation

## Migration file naming conventions

Recommended naming conventions per tool:

| Tool | Convention | Example |
|------|------|------|
| golang-migrate | `YYYYMMDDHHMMSS_description.up.sql` | `20240101120000_create_users.up.sql` |
| Alembic | `rev_<hash>_description.py` | `rev_a001b002_add_email.py` |
| Prisma | Timestamped folder | `20240101120000_init` |
| Flyway | `V<version>__description.sql` | `V1__Create_users.sql` |
| Rails | `YYYYMMDDHHMMSS_description.rb` | `20240101120000_create_users.rb` |
| Laravel | `YYYY_MM_DD_HHMMSS_description.php` | `2024_01_01_120000_create_users.php` |

## Troubleshooting

### Migration files are not scanned

1. Check the migration path:

```bash
ls -la $(path/to/migrations)
```

2. Check the file pattern — verify file extensions match the expected pattern

3. Check language detection:

```bash
cat .moai/config/sections/language.yaml
```

4. Configure a custom path:

```yaml
db:
  migration_patterns:
    - path: "your/custom/path"
      file_pattern: "*.sql"
```

### Multiple language configurations conflict

Clearly separate paths per service:

```yaml
db:
  migration_patterns:
    - path: "services/api/**"          # backend only
    - path: "apps/web/**"              # web app only
```
