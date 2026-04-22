---
title: 마이그레이션 패턴
description: 16개 프로그래밍 언어의 기본 마이그레이션 경로 및 설정
weight: 30
draft: false
---

## 지원하는 언어와 마이그레이션 도구

MoAI는 16개 프로그래밍 언어의 기본 마이그레이션 경로를 지원합니다. 각 언어는 업계 표준 도구를 사용합니다.

| 언어 | 마이그레이션 도구 | 기본 경로 패턴 |
|------|-----------------|--------------|
| Go | golang-migrate | `db/migrations/*.sql` 또는 `migrations/*.sql` |
| Python | Alembic | `alembic/versions/*.py` |
| TypeScript | Prisma Migrate | `prisma/migrations/**/*.sql` |
| JavaScript | Knex.js | `migrations/*.js` 또는 `knexfile migrations/` |
| Rust | SQLx | `migrations/*.sql` |
| Java | Flyway | `src/main/resources/db/migration/V*.sql` |
| Kotlin | Flyway | `src/main/resources/db/migration/V*.sql` |
| C# | EF Core Migrations | `Migrations/*.cs` |
| Ruby | Rails ActiveRecord | `db/migrate/*.rb` |
| PHP | Laravel Migrations | `database/migrations/*.php` |
| Elixir | Ecto | `priv/repo/migrations/*.exs` |
| C++ | 표준 없음 (관례) | `db/migrations/*.sql` |
| Scala | Slick / Flyway | `src/main/resources/db/migration/V*.sql` |
| R | 표준 없음 (관례) | `migrations/*.sql` |
| Flutter | Drift | `assets/migrations/*.sql` |
| Swift | GRDB | `Resources/Migrations/*.sql` |

## 자동 언어 감지

MoAI는 다음 방법으로 프로젝트 언어를 자동 감지합니다:

1. `.moai/config/sections/language.yaml`의 `project_markers` 확인
2. 프로젝트 루트의 언어별 마커 파일 스캔:
   - Go: `go.mod`
   - Python: `pyproject.toml`, `setup.py`
   - TypeScript/JavaScript: `package.json`
   - Rust: `Cargo.toml`
   - Ruby: `Gemfile`
   - PHP: `composer.json`
   - Java/Kotlin: `pom.xml`, `build.gradle`
   - C#: `*.csproj`
   - Elixir: `mix.exs`

## 커스텀 마이그레이션 경로 설정

기본 경로가 프로젝트와 맞지 않으면 `.moai/config/sections/db.yaml`에서 수동으로 지정할 수 있습니다:

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

## 예제: 각 언어별 마이그레이션 파일 구조

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

## 멀티 언어 프로젝트 설정

마이크로서비스나 모놀리식 구조에서 여러 언어의 마이그레이션을 관리하는 경우:

```yaml
db:
  migration_patterns:
    # 백엔드 (Go)
    - path: "services/api/db/migrations"
      file_pattern: "*.sql"
      language: "go"
    
    # 데이터 파이프라인 (Python)
    - path: "services/analytics/alembic/versions"
      file_pattern: "*.py"
      language: "python"
    
    # 웹 애플리케이션 (TypeScript)
    - path: "apps/web/prisma/migrations"
      file_pattern: "*.sql"
      language: "typescript"
```

## 마이그레이션 도구 선택 가이드

### Prisma (TypeScript/JavaScript)

장점:
- 간단한 문법
- 자동 타입 생성
- 직관적인 관계 정의

단점:
- Prisma 생태계에 의존
- 복잡한 마이그레이션 제한

### Alembic (Python)

장점:
- 자동 마이그레이션 생성 기능
- 유연한 커스터마이징
- SQLAlchemy 완전 통합

단점:
- 학습 곡선
- 초기 설정 복잡

### Flyway (Java/Kotlin)

장점:
- 언어별 마이그레이션 지원
- 강력한 검증
- 워터마크 시스템

단점:
- 설정 복잡도
- 성능 오버헤드

### golang-migrate (Go)

장점:
- 가볍고 빠름
- Up/Down 명확한 구분
- 순수 SQL 사용

단점:
- 도움 기능 없음
- 자동 생성 불가

## 마이그레이션 파일 명명 규칙

각 도구별 권장 명명 규칙:

| 도구 | 규칙 | 예제 |
|------|------|------|
| golang-migrate | `YYYYMMDDHHMMSS_description.up.sql` | `20240101120000_create_users.up.sql` |
| Alembic | `rev_<hash>_description.py` | `rev_a001b002_add_email.py` |
| Prisma | 타임스탐프 폴더 | `20240101120000_init` |
| Flyway | `V<version>__description.sql` | `V1__Create_users.sql` |
| Rails | `YYYYMMDDHHMMSS_description.rb` | `20240101120000_create_users.rb` |
| Laravel | `YYYY_MM_DD_HHMMSS_description.php` | `2024_01_01_120000_create_users.php` |

## 문제 해결

### 마이그레이션 파일이 스캔되지 않음

1. 마이그레이션 경로 확인:

```bash
ls -la $(path/to/migrations)
```

2. 파일 패턴 확인 — 파일 확장자가 예상된 패턴과 일치하는지 확인

3. 언어 감지 확인:

```bash
cat .moai/config/sections/language.yaml
```

4. 커스텀 경로 설정:

```yaml
db:
  migration_patterns:
    - path: "your/custom/path"
      file_pattern: "*.sql"
```

### 여러 언어 설정이 충돌함

각 서비스별 경로를 명확히 분리합니다:

```yaml
db:
  migration_patterns:
    - path: "services/api/**"          # 백엔드 전용
    - path: "apps/web/**"              # 웹앱 전용
```
