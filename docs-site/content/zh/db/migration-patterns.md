---
title: 迁移模式
description: 16 种编程语言的默认迁移路径和配置
weight: 30
draft: false
---

## 支持的语言和工具

MoAI 支持 16 种编程语言的默认迁移路径，各语言都使用行业标准工具。

| 语言 | 迁移工具 | 默认路径模式 |
|------|--------|-----------|
| Go | golang-migrate | `db/migrations/*.sql` 或 `migrations/*.sql` |
| Python | Alembic | `alembic/versions/*.py` |
| TypeScript | Prisma Migrate | `prisma/migrations/**/*.sql` |
| JavaScript | Knex.js | `migrations/*.js` 或 `knexfile migrations/` |
| Rust | SQLx | `migrations/*.sql` |
| Java | Flyway | `src/main/resources/db/migration/V*.sql` |
| Kotlin | Flyway | `src/main/resources/db/migration/V*.sql` |
| C# | EF Core Migrations | `Migrations/*.cs` |
| Ruby | Rails ActiveRecord | `db/migrate/*.rb` |
| PHP | Laravel Migrations | `database/migrations/*.php` |
| Elixir | Ecto | `priv/repo/migrations/*.exs` |
| C++ | 无标准 (约定) | `db/migrations/*.sql` |
| Scala | Slick / Flyway | `src/main/resources/db/migration/V*.sql` |
| R | 无标准 (约定) | `migrations/*.sql` |
| Flutter | Drift | `assets/migrations/*.sql` |
| Swift | GRDB | `Resources/Migrations/*.sql` |

## 自动语言检测

MoAI 使用以下方法自动检测你的项目语言:

1. 检查 `.moai/config/sections/language.yaml` 中的 `project_markers`
2. 扫描项目根的特定于语言的标记文件:
   - Go: `go.mod`
   - Python: `pyproject.toml`、`setup.py`
   - TypeScript/JavaScript: `package.json`
   - Rust: `Cargo.toml`
   - Ruby: `Gemfile`
   - PHP: `composer.json`
   - Java/Kotlin: `pom.xml`、`build.gradle`
   - C#: `*.csproj`
   - Elixir: `mix.exs`

## 自定义迁移路径配置

如果默认路径与你的项目不匹配，在 `.moai/config/sections/db.yaml` 中手动指定:

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

## 示例: 按语言的迁移结构

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

## 多语言项目配置

对于有多种语言的微服务或单体结构:

```yaml
db:
  migration_patterns:
    # 后端 (Go)
    - path: "services/api/db/migrations"
      file_pattern: "*.sql"
      language: "go"
    
    # 数据管道 (Python)
    - path: "services/analytics/alembic/versions"
      file_pattern: "*.py"
      language: "python"
    
    # Web 应用 (TypeScript)
    - path: "apps/web/prisma/migrations"
      file_pattern: "*.sql"
      language: "typescript"
```

## 迁移工具选择指南

### Prisma (TypeScript/JavaScript)

优点:
- 简洁语法
- 自动类型生成
- 直观的关系定义

缺点:
- Prisma 生态系统依赖
- 复杂迁移有限制

### Alembic (Python)

优点:
- 自动迁移生成
- 灵活定制
- 完整 SQLAlchemy 集成

缺点:
- 学习曲线
- 复杂的初始设置

### Flyway (Java/Kotlin)

优点:
- 多语言迁移支持
- 强大的验证
- 水位线系统

缺点:
- 配置复杂
- 性能开销

### golang-migrate (Go)

优点:
- 轻量且快速
- Up/Down 分明
- 纯 SQL 使用

缺点:
- 无辅助函数
- 无自动生成

## 迁移文件命名约定

每个工具的推荐命名模式:

| 工具 | 模式 | 示例 |
|------|------|------|
| golang-migrate | `YYYYMMDDHHMMSS_description.up.sql` | `20240101120000_create_users.up.sql` |
| Alembic | `rev_<hash>_description.py` | `rev_a001b002_add_email.py` |
| Prisma | 时间戳文件夹 | `20240101120000_init` |
| Flyway | `V<version>__description.sql` | `V1__Create_users.sql` |
| Rails | `YYYYMMDDHHMMSS_description.rb` | `20240101120000_create_users.rb` |
| Laravel | `YYYY_MM_DD_HHMMSS_description.php` | `2024_01_01_120000_create_users.php` |

## 故障排除

### 迁移文件未被扫描

1. 检查迁移路径:

```bash
ls -la $(path/to/migrations)
```

2. 验证文件模式 — 确认文件扩展名与预期模式匹配

3. 检查语言检测:

```bash
cat .moai/config/sections/language.yaml
```

4. 设置自定义路径:

```yaml
db:
  migration_patterns:
    - path: "your/custom/path"
      file_pattern: "*.sql"
```

### 多种语言配置冲突

分开每个服务的路径:

```yaml
db:
  migration_patterns:
    - path: "services/api/**"          # 仅后端
    - path: "apps/web/**"              # 仅 Web 应用
```
