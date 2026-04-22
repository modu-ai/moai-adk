---
title: Migration Patterns
description: Default migration paths and configuration for 16 programming languages
weight: 30
draft: false
---

## Supported Languages and Tools

MoAI supports default migration paths for 16 programming languages, each using industry-standard tools.

| Language | Migration Tool | Default Path Pattern |
|----------|---------------|---------------------|
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

## Automatic Language Detection

MoAI auto-detects your project language using:

1. Check `project_markers` in `.moai/config/sections/language.yaml`
2. Scan for language-specific marker files at project root:
   - Go: `go.mod`
   - Python: `pyproject.toml`, `setup.py`
   - TypeScript/JavaScript: `package.json`
   - Rust: `Cargo.toml`
   - Ruby: `Gemfile`
   - PHP: `composer.json`
   - Java/Kotlin: `pom.xml`, `build.gradle`
   - C#: `*.csproj`
   - Elixir: `mix.exs`

## Custom Migration Path Configuration

If the default path doesn't match your project, manually specify it in `.moai/config/sections/db.yaml`:

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

## Example: Migration Structure by Language

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

## Multi-Language Project Configuration

For microservices or monolithic structures with multiple languages:

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
    
    # Web app (TypeScript)
    - path: "apps/web/prisma/migrations"
      file_pattern: "*.sql"
      language: "typescript"
```

## Migration Tool Selection Guide

### Prisma (TypeScript/JavaScript)

Pros:
- Simple syntax
- Automatic type generation
- Intuitive relationship definitions

Cons:
- Prisma ecosystem dependency
- Limited for complex migrations

### Alembic (Python)

Pros:
- Auto-migration generation
- Flexible customization
- Complete SQLAlchemy integration

Cons:
- Learning curve
- Complex initial setup

### Flyway (Java/Kotlin)

Pros:
- Multi-language migration support
- Strong validation
- Watermark system

Cons:
- Complex configuration
- Performance overhead

### golang-migrate (Go)

Pros:
- Lightweight and fast
- Clear Up/Down separation
- Pure SQL usage

Cons:
- No helper functions
- No auto-generation

## Migration File Naming Conventions

Recommended naming patterns per tool:

| Tool | Pattern | Example |
|------|---------|---------|
| golang-migrate | `YYYYMMDDHHMMSS_description.up.sql` | `20240101120000_create_users.up.sql` |
| Alembic | `rev_<hash>_description.py` | `rev_a001b002_add_email.py` |
| Prisma | Timestamp folder | `20240101120000_init` |
| Flyway | `V<version>__description.sql` | `V1__Create_users.sql` |
| Rails | `YYYYMMDDHHMMSS_description.rb` | `20240101120000_create_users.rb` |
| Laravel | `YYYY_MM_DD_HHMMSS_description.php` | `2024_01_01_120000_create_users.php` |

## Troubleshooting

### Migration Files Not Scanned

1. Check migration path:

```bash
ls -la $(path/to/migrations)
```

2. Verify file pattern — confirm file extensions match expected pattern

3. Check language detection:

```bash
cat .moai/config/sections/language.yaml
```

4. Set custom path:

```yaml
db:
  migration_patterns:
    - path: "your/custom/path"
      file_pattern: "*.sql"
```

### Multiple Language Conflicts

Separate each service with explicit paths:

```yaml
db:
  migration_patterns:
    - path: "services/api/**"          # Backend only
    - path: "apps/web/**"              # Web app only
```
