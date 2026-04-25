---
id: SPEC-DB-SYNC-002
version: 0.1.0
status: draft
created_at: 2026-04-24
updated_at: 2026-04-24
author: manager-spec (follow-up to plan-auditor 2026-04-24)
priority: High
labels: [db, parser, migration, prisma, alembic, rails, raw-sql, follow-up]
issue_number: null
depends_on: [SPEC-DB-SYNC-001]
related_specs: [SPEC-DB-SYNC-RELOC-001, SPEC-DB-SYNC-HARDEN-001, SPEC-DB-CMD-001]
---

# SPEC-DB-SYNC-002: Migration File Parser Backends for DB Schema Sync

## HISTORY

- 2026-04-24 v0.1.0: 최초 작성. 2026-04-24 plan-auditor 리포트에서 SPEC-DB-SYNC-001의 parser-stub gap을 독립 follow-up SPEC으로 분리. PostToolUse 훅 복원은 SPEC-DB-SYNC-RELOC-001에서 의도적으로 제거되었으므로 본 SPEC의 범위 외(scope-out)로 명시. `internal/db/parser/` 패키지 신설과 4개 포맷(Prisma, Alembic, Rails, raw SQL) 파서 백엔드 구현에 한정.

## Background

2026-04-24 plan-auditor 감사 결과 SPEC-DB-SYNC-001에 대해 다음 gap이 확인되었다.

> Go hook handler (`internal/hook/dbsync/db_schema_sync.go`)와 CLI 서브커맨드 `moai hook db-schema-sync`는 27개 테스트 함수와 함께 완전 구현되어 있다. 그러나 실제 마이그레이션 파서(`internal/db/parser/`)는 스텁 상태(디렉터리 자체가 존재하지 않음)이므로 `/moai db verify` drift 검출과 `/moai db refresh` full-rebuild가 Prisma/Alembic/Rails 파일에서 스키마를 실제로 추출할 수 없다. 영향: `/moai db refresh`와 `/moai db verify` 명령은 구조적으로 존재하나 파서 부재로 인해 빈/trivial 결과만 반환한다.

현재 상태 검증(본 SPEC 작성 시점):

- `internal/db/` 디렉터리: **존재하지 않음**
- `internal/db/parser/` 디렉터리: **존재하지 않음**
- `internal/hook/dbsync/db_schema_sync.go`: 존재(20,312 bytes, 두 테스트 파일 포함)
- `/moai db refresh`·`/moai db verify`: Claude Code slash command 레벨에서 존재(skill-based), Go CLI 서브커맨드는 별도 존재하지 않음

본 SPEC은 위 gap 중 **파서 백엔드 구현**에만 초점을 맞춘다. 감사가 부차적으로 언급한 "bash wrapper + settings.json.tmpl PostToolUse matcher" 복원은 SPEC-DB-SYNC-RELOC-001이 **의도적으로** 제거하고 `/moai sync` Phase 0.08로 호출 지점을 이동시킨 후속 결정이므로 본 SPEC의 범위 **외**로 명시한다.

### Scope Boundary

- **SPEC-DB-SYNC-001과의 경계**: SPEC-DB-SYNC-001은 훅 핸들러 엔트리포인트와 debounce/state 관리를 정의한다. 본 SPEC은 그 핸들러가 호출하는 `Schema` 추출 로직(파서 백엔드)을 신설한다.
- **SPEC-DB-SYNC-RELOC-001과의 경계**: RELOC-001은 훅 호출 지점을 PostToolUse에서 `/moai sync` Phase 0.08로 이전했다. 본 SPEC은 이 호출 경로 변경과 무관하며, 파서 패키지는 호출 지점이 어디든 동일하게 작동한다.
- **SPEC-DB-SYNC-HARDEN-001과의 경계**: HARDEN-001은 기존 Go 훅 핸들러의 보안·안정성 보강(v0.3.0 H3 supersede)을 다룬다. 본 SPEC은 신규 파서 패키지의 초기 구현이며, HARDEN-001의 하드닝 원칙(안전한 파일 I/O, 크기 제한, 에러 전파)을 계승한다.
- **SPEC-DB-CMD-001과의 경계**: SPEC-DB-CMD-001은 `/moai db` 명령 표면과 서브커맨드 라우팅(스텁)을 정의한다. 본 SPEC이 제공하는 파서 백엔드는 `refresh`/`verify`가 호출하는 내부 API로 소비된다.

## Goals and Non-Goals

### Goals (In Scope)

- `internal/db/parser/` 패키지 신설 및 공통 `Parser` 인터페이스·레지스트리 정의
- 4개 마이그레이션 포맷 파서 백엔드 구현:
  1. Prisma schema (`schema.prisma`)
  2. Alembic version files (`alembic/versions/**/*.py`)
  3. Rails migration (`db/migrate/**/*.rb`)
  4. Raw SQL (`migrations/**/*.sql`, `supabase/migrations/**/*.sql`, `sql/migrations/**/*.sql`)
- 공통 `Schema` 표현(Table/Column/Index/ForeignKey/Check)과 YAML·Mermaid ERD 직렬화
- 파서 실패 시 partial schema 반환과 경고 로깅(panic 금지)
- 파일 크기 상한(`parser.MaxFileSize`, 기본 1MB) 기반 거부(`ErrFileTooLarge`)

### Non-Goals (Out of Scope)

- **PostToolUse 훅 복원**: RELOC-001이 의도적으로 제거한 bash wrapper와 `settings.json.tmpl` matcher는 본 SPEC의 범위가 아니다.
- **신규 `/moai db` 서브커맨드 추가**: 기존 `/moai db refresh`·`/moai db verify`가 소비자이며, 신규 명령은 도입하지 않는다.
- **실제 데이터베이스 드라이버 연결**: 본 SPEC은 **파일 파싱만** 담당한다. 런타임 DB 쿼리, 실 DB 스키마 내성(introspection), 마이그레이션 실행은 모두 범위 외다.
- **스키마 diff 알고리즘 확장**: 현재 `/moai db verify`가 사용하는 diff 로직 수준을 유지한다. 신규 diff 규칙(예: 컬럼 순서 변경 감지, 데이터 호환성 분석)은 도입하지 않는다.
- **SQL 방언 완전 지원**: PostgreSQL을 주력으로, MySQL/SQLite를 보조로 지원한다. Oracle/MSSQL은 best-effort로만 처리하며 보장하지 않는다.
- **Python/Ruby 동적 코드 실행**: Alembic/Rails 파일 내 동적 분기·외부 함수 호출은 정적 분석만 수행하고 분석 불가 구간은 `TODO_UNPARSED`로 마킹한다.

## Scope

### Package Layout

신설 패키지 및 하위 서브패키지(모두 `internal/db/parser/` 하위):

```
internal/db/parser/
├── parser.go          # Parser 인터페이스, Registry, MaxFileSize, Errors
├── schema.go          # Schema/Table/Column/Index/ForeignKey/Check 타입
├── serialize.go       # YAML + Mermaid ERD 직렬화
├── registry.go        # 파일 경로 glob → Parser 매핑
├── prisma/
│   ├── prisma.go      # Prisma 파서 구현
│   └── prisma_test.go
├── alembic/
│   ├── alembic.go     # Alembic 파서 구현(정적 AST 기반)
│   └── alembic_test.go
├── rails/
│   ├── rails.go       # Rails 파서 구현(정규식 + DSL 매칭)
│   └── rails_test.go
├── sql/
│   ├── sql.go         # Raw SQL 파서 구현(정규식 우선 + 선택적 AST)
│   └── sql_test.go
└── testdata/
    ├── prisma/        # 5+ fixtures
    ├── alembic/       # 5+ fixtures
    ├── rails/         # 5+ fixtures
    └── sql/           # 5+ fixtures
```

### Consumers

- `internal/hook/dbsync/db_schema_sync.go` — `Parser.Parse`를 호출하여 변경된 마이그레이션 파일의 스키마 표현을 획득
- `/moai db refresh` skill — 디렉터리 전체를 순회하여 `Parser.ParseAll`로 통합 스키마 생성 → `.moai/project/db/schema.md` 기록
- `/moai db verify` skill — 현재 파일 파싱 결과 vs `.moai/project/db/schema.md` 스냅샷 diff

## Requirements (EARS)

### Package Foundation

1. **REQ-001 (Ubiquitous)**: The package `internal/db/parser/` SHALL exist and expose an interface `Parser { Parse(path string) (Schema, error) }`, a concrete `Schema` struct (see REQ-008), and a registry that maps file-path globs to registered `Parser` implementations.
   - **Rationale**: 공통 인터페이스가 없으면 각 포맷 파서의 호출부가 switch 문 분기로 흩어져 유지보수가 불가능하다. Registry 패턴은 신규 포맷(예: Liquibase) 추가 시 최소 변경을 보장한다.

### Prisma Parser

2. **REQ-002 (Event-driven)**: WHEN `Parser.Parse` is invoked with a file that matches glob `prisma/schema.prisma` (or any file ending in `.prisma`), the Prisma parser SHALL extract `model` declarations (each with name, fields, attributes, indexes, and relations) and populate `Schema.Tables` with one `Table` per model.
   - **Rationale**: Prisma schema는 선언적 DSL이며 단일 파일에 모든 스키마가 포함되므로 파일 파싱만으로 완전한 스키마 표현 추출이 가능하다.

### Alembic Parser

3. **REQ-003 (Event-driven)**: WHEN `Parser.Parse` is invoked with a file matching glob `alembic/versions/**/*.py`, the Alembic parser SHALL extract static calls to `op.create_table`, `op.add_column`, `op.drop_table`, `op.drop_column`, `op.alter_column`, `op.create_index`, `op.drop_index`, and `op.create_foreign_key` from the `upgrade()` function and represent them as schema operations (creates, drops, alters) carried on the returned `Schema`.
   - **Rationale**: Alembic 버전 파일은 정방향 마이그레이션 델타를 기술한다. 전체 스키마가 아닌 델타로서 소비되어야 consumer(`refresh`)가 여러 버전을 시간순으로 축적하여 누적 스키마를 재구성할 수 있다.

### Rails Parser

4. **REQ-004 (Event-driven)**: WHEN `Parser.Parse` is invoked with a file matching glob `db/migrate/**/*.rb`, the Rails parser SHALL extract standard `ActiveRecord::Migration` DSL calls (`create_table`, `drop_table`, `add_column`, `remove_column`, `rename_column`, `add_index`, `remove_index`, `add_foreign_key`) from `change`, `up`, and `down` methods and represent them as schema operations on the returned `Schema`.
   - **Rationale**: Rails 마이그레이션은 Ruby DSL이며 표준 메서드만 정적으로 매칭한다. 커스텀 Ruby 블록은 REQ-006에 의해 `TODO_UNPARSED`로 마킹된다.

### Raw SQL Parser

5. **REQ-005 (Event-driven)**: WHEN `Parser.Parse` is invoked with a file matching glob `migrations/**/*.sql`, `supabase/migrations/**/*.sql`, or `sql/migrations/**/*.sql`, the SQL parser SHALL parse `CREATE TABLE`, `ALTER TABLE ... ADD COLUMN`, `ALTER TABLE ... DROP COLUMN`, `DROP TABLE`, `CREATE INDEX`, `DROP INDEX`, `CREATE UNIQUE INDEX`, and `ALTER TABLE ... ADD CONSTRAINT` statements using a regex-first approach, with a documented fallback to an AST-based library (`github.com/xwb1989/sqlparser` or equivalent) when the regex path cannot unambiguously resolve a statement.
   - **Rationale**: Raw SQL은 방언이 다양하고 공식 AST 라이브러리는 Go 호환성·PostgreSQL DDL 커버리지가 불완전하다. 정규식 우선 전략은 80% 이상의 표준 DDL 케이스를 빠르고 결정적으로 처리하며, AST 라이브러리는 복잡한 표현식·중첩 CHECK 제약에 한정하여 호출된다.

### Error Handling

6. **REQ-006 (State-driven)**: WHILE a parser encounters syntax it cannot statically resolve (dynamic Python code, unknown Ruby method, non-standard SQL dialect, custom Prisma generator), the parser SHALL log a warning with file path, line number, and reason, append an entry to `Schema.UnparsedSegments` with tag `TODO_UNPARSED`, continue processing remaining statements, and return a partial `Schema` with no panic or uncaught error.
   - **Rationale**: 마이그레이션 파일은 프로젝트마다 관행이 다르며 부분적으로 파싱 가능한 경우가 많다. 첫 실패에서 panic하면 `/moai db refresh` 전체가 중단되어 사용자가 수동 개입해야 하므로 graceful degradation이 필수다.

### Serialization

7. **REQ-007 (Ubiquitous)**: The `Schema` struct SHALL be serializable both to YAML (for `.moai/project/db/schema.md` body) via `schema.ToYAML(w io.Writer) error` and to Mermaid ERD syntax (for `.moai/project/db/erd.mmd`) via `schema.ToMermaid(w io.Writer) error`, where the Mermaid output uses `erDiagram` with relationships derived from `ForeignKey` entries.
   - **Rationale**: 두 개의 별도 출력 포맷이 `.moai/project/db/` 문서 생성의 primary consumer다. YAML은 diff-friendly·기계 판독용, Mermaid ERD는 사람 친화적 시각화를 담당하며, 동일 `Schema`에서 유도되어야 두 문서의 드리프트가 원천 차단된다.

### Common Schema Representation

8. **REQ-008 (Ubiquitous)**: The package SHALL define `Schema` as a struct with the following public fields: `Tables []Table`, `UnparsedSegments []UnparsedSegment`, `SourcePath string`, `Dialect string`, `GeneratedAt time.Time`. Each `Table` SHALL contain `Name string`, `Columns []Column`, `Indexes []Index`, `ForeignKeys []ForeignKey`, `Checks []CheckConstraint`, `Comment string`. Each `Column` SHALL contain `Name string`, `Type string`, `Nullable bool`, `Default *string`, `IsPrimaryKey bool`, `Comment string`.
   - **Rationale**: Prisma/Alembic/Rails/SQL 네 포맷의 공통분모를 통일된 표현으로 노출해야 consumer가 포맷별 분기 없이 소비할 수 있다. 포맷 고유 메타데이터(예: Prisma `@relation`)는 `Comment` 또는 `Attributes` 확장 필드로 흘려 보내되 공통 필드는 최소 집합으로 고정한다.

### File Size Limit

9. **REQ-009 (Unwanted Behavior)**: IF `Parser.Parse` is invoked with a file whose byte size exceeds `parser.MaxFileSize` (default 1,048,576 bytes = 1MiB, configurable via package-level variable), THEN the parser SHALL return `ErrFileTooLarge` without reading the full file into memory, log a warning with file path and size, and leave any previous `.moai/project/db/schema.md` unchanged.
   - **Rationale**: 비정상적으로 큰 마이그레이션 덤프(예: seed 데이터 포함 SQL 수십 MB)를 파싱 시도하면 훅이 지연·OOM을 일으킨다. 상한 초과는 명시적 거부로 처리하여 사용자가 `db.yaml` 설정으로 해결하도록 유도한다.

### Registry Lookup

10. **REQ-010 (Ubiquitous)**: The package SHALL expose `registry.ForPath(path string) (Parser, error)` which resolves a parser based on the first matching glob in a deterministic registration order, and SHALL return `ErrNoParser` when no registered parser matches the path.
    - **Rationale**: 호출자(훅 핸들러, `/moai db refresh`)는 파일 경로만 알기 때문에 레지스트리 기반 dispatch가 필요하며, 등록 순서가 비결정적이면 동일 경로가 서로 다른 파서로 해석되는 회귀가 발생한다.

## Acceptance Criteria

각 AC는 `internal/db/parser/` 하위 테스트 파일로 traceable하다. 테이블 주도(table-driven) 테스트에 각 포맷당 최소 5개 fixture 필요.

| AC   | Requirement    | Description                                                                                                                                               | Test File                                                 |
| ---- | -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| AC-1 | REQ-001        | `Parser` 인터페이스와 `Registry`가 export되며 `go vet`/`go build` 통과. 레지스트리에 4개 파서 등록 확인.                                                  | `internal/db/parser/parser_test.go`                       |
| AC-2 | REQ-002        | Prisma `schema.prisma` fixture 5개(단순 모델, 관계, 복합 인덱스, enum, @@map)를 파싱하여 기대 `Schema.Tables` 구조 일치.                                  | `internal/db/parser/prisma/prisma_test.go`                |
| AC-3 | REQ-003        | Alembic 버전 파일 5개(create_table only, add_column, drop_column, alter_column, create_foreign_key)를 파싱하여 기대 delta 배열 생성.                      | `internal/db/parser/alembic/alembic_test.go`              |
| AC-4 | REQ-004        | Rails 마이그레이션 5개(change method, up/down method, add_index, add_foreign_key, remove_column)를 파싱하여 기대 delta 배열 생성.                        | `internal/db/parser/rails/rails_test.go`                  |
| AC-5 | REQ-005        | Raw SQL fixture 5개(PostgreSQL CREATE TABLE, MySQL ALTER, SQLite INDEX, complex CHECK, multi-statement)를 정규식 경로 우선으로 파싱하여 기대 구조 생성.  | `internal/db/parser/sql/sql_test.go`                      |
| AC-6 | REQ-006        | 각 포맷당 최소 1개의 "unparsable" fixture(예: 동적 Python 분기)가 panic 없이 partial Schema + `UnparsedSegments` 엔트리 포함하여 반환.                    | 각 포맷 `*_test.go`의 `TestUnparsable*` 서브테스트        |
| AC-7 | REQ-007        | 임의 `Schema`를 `ToYAML`/`ToMermaid`로 직렬화 후 재파싱(YAML) 또는 Mermaid syntax 검증(정규식 기반)으로 round-trip 또는 구조 유효성 확인.               | `internal/db/parser/serialize_test.go`                    |
| AC-8 | REQ-008        | `Schema`/`Table`/`Column` 구조체의 public field 집합이 REQ-008 명세와 일치(reflect 기반 구조 assertion 테스트).                                          | `internal/db/parser/schema_test.go`                       |
| AC-9 | REQ-009        | 1MiB + 1 byte 크기의 임시 파일에 대해 `ErrFileTooLarge` 반환 확인. `t.TempDir()` 사용. 파일은 stream 방식으로만 읽혀 OOM 없이 반환.                      | `internal/db/parser/parser_test.go::TestFileTooLarge`     |
| AC-10| REQ-010        | `registry.ForPath`에 4개 포맷 각 매칭 경로·매칭 실패 경로(`README.md`) 주입 시 올바른 파서 또는 `ErrNoParser` 반환.                                      | `internal/db/parser/registry_test.go`                     |

## Implementation Notes

### Library Recommendations

| Format  | Primary Approach    | Library Candidate                                     | Go Version Caveat                                    |
| ------- | ------------------- | ----------------------------------------------------- | ---------------------------------------------------- |
| Prisma  | 정규식 + 상태 머신 | (순수 구현) 또는 `github.com/prisma/prisma-client-go` | prisma-client-go는 runtime client가 주목적이라 파서만 추출 시 대용량 의존성 유입 위험. 순수 구현 권장. |
| Alembic | Python AST 정적 해석 | `github.com/go-python/gpython` 또는 정규식 fallback  | gpython은 Python 3 호환성 제한. `op.*` 호출 패턴만 정규식으로 매칭하는 접근이 현실적. |
| Rails   | 정규식 + DSL 매처   | (순수 구현)                                           | Ruby AST 파서는 Go 생태계에 부재. 표준 메서드 화이트리스트 기반 정규식이 유일한 실현 경로. |
| SQL     | 정규식 우선, AST fallback | `github.com/xwb1989/sqlparser` (MySQL 방언)        | PostgreSQL DDL은 sqlparser 커버리지 낮음. 정규식 우선 + `pg_query_go` 선택적 의존성 고려. |

### Regex-First Rationale

외부 파서 라이브러리는 (1) 방언 커버리지가 불완전하고 (2) Go 모듈 의존성을 증가시키며 (3) 향후 업데이트 중단 위험이 있다. 정규식 우선 전략은:

- 표준 DDL의 80% 이상을 즉시 처리
- 실패 시 `UnparsedSegments`로 graceful degrade(REQ-006)
- AST 기반 백엔드는 pluggable interface로 후속 SPEC에서 추가 가능

### Dialect Strategy

- **Primary**: PostgreSQL(Supabase, Prisma PostgreSQL provider)
- **Secondary**: MySQL, SQLite
- **Best-effort**: Oracle, MSSQL — 파싱 실패 시 `UnparsedSegments`로 로깅

### Parser Registration Order (Deterministic)

```
1. prisma/schema.prisma         → prisma.Parser
2. alembic/versions/**/*.py     → alembic.Parser
3. db/migrate/**/*.rb           → rails.Parser
4. migrations/**/*.sql          → sql.Parser
5. supabase/migrations/**/*.sql → sql.Parser
6. sql/migrations/**/*.sql      → sql.Parser
```

glob 매칭은 `path.Match`·`doublestar` 라이브러리 기반. 첫 번째 매칭이 승리하며 등록 순서는 `registry.go`에 하드코딩(결정적 동작 보장).

## Risks and Mitigations

| 리스크                                                           | 영향도 | 완화 전략                                                                                              |
| ---------------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------ |
| Prisma DSL 문법 진화(새로운 `@` attribute, preview feature)      | Medium | 테스트 fixture를 Prisma 특정 major version(예: v5)에 고정하고 `prisma-version.txt`로 버전 명시. 신버전 도입 시 별도 SPEC. |
| Alembic 동적 Python(다른 함수 호출, 조건부 `op`)                 | High   | 정적 `op.*` 호출만 파싱하고 비정적 구간은 `TODO_UNPARSED`로 마킹. Consumer는 이를 `schema.md`에 명시. |
| Rails custom Ruby 블록                                           | Medium | 표준 DSL 메서드 화이트리스트만 처리. 미지의 메서드는 warning + `UnparsedSegments`. |
| SQL 방언 변동(특히 Oracle `CREATE TABLE ... AS SELECT`)          | Medium | Primary=PostgreSQL 선언. 방언별 regex pack 분리 가능한 구조로 설계. Oracle/MSSQL 이슈는 best-effort로만 해결. |
| `sqlparser`/`pg_query_go` 의존성 유입 시 바이너리 크기 증가      | Low    | AST fallback을 build tag 또는 optional package로 격리. 기본 빌드는 정규식만 사용.                     |
| 1MiB 초과 seed 데이터 SQL이 일상적으로 커밋되는 프로젝트         | Low    | `MaxFileSize`를 `db.yaml`에서 오버라이드 가능하도록 노출(후속). 본 SPEC에서는 package-level var로만 노출. |

## Test Strategy

### Table-Driven Test Pattern

각 포맷당 fixture 디렉터리(`testdata/<format>/`)에 5+ 케이스를 배치하고 공통 테스트 러너가 `*.input` ↔ `*.golden.yaml` round-trip을 검증.

```
internal/db/parser/prisma/testdata/
├── 01-simple-model/
│   ├── schema.prisma
│   └── expected.yaml
├── 02-relations/
│   ├── schema.prisma
│   └── expected.yaml
├── 03-composite-index/
│   ├── schema.prisma
│   └── expected.yaml
├── 04-enum-types/
│   ├── schema.prisma
│   └── expected.yaml
├── 05-map-attribute/
│   ├── schema.prisma
│   └── expected.yaml
└── 99-unparsable/
    ├── schema.prisma
    └── expected-partial.yaml
```

### Coverage Target

- 패키지 전체 85% 이상(`.claude/rules/moai/development/coding-standards.md` 기준)
- 각 파서 서브패키지 90% 이상(핵심 로직)
- `serialize.go` 95% 이상(직렬화는 단순 구조, 높은 커버리지 용이)

### Integration Test

- `internal/hook/dbsync/db_schema_sync.go` 기존 테스트는 파서 패키지 도입 후에도 통과해야 함(회귀 방지).
- `testdata/integration/` 디렉터리에 4개 포맷 혼합 프로젝트 샘플을 두고 `/moai db refresh` simulation 테스트(`cli_integration_test.go`) 추가.

### Benchmark

- `testdata/benchmark/` 디렉터리에 100KB·500KB 파일 배치. `go test -bench=. -benchmem` 기준 500KB 파일 파싱 < 500ms 목표(비강제, 관찰용).

## Traceability

### REQ ↔ AC Matrix

| Requirement | Acceptance Criteria | Consumer                                        |
| ----------- | ------------------- | ----------------------------------------------- |
| REQ-001     | AC-1                | 훅 핸들러, `/moai db refresh`                   |
| REQ-002     | AC-2                | Prisma 프로젝트                                 |
| REQ-003     | AC-3                | Alembic/SQLAlchemy 프로젝트                     |
| REQ-004     | AC-4                | Rails 프로젝트                                  |
| REQ-005     | AC-5                | Supabase, 순수 SQL 프로젝트                     |
| REQ-006     | AC-6                | 모든 포맷 — graceful degradation                |
| REQ-007     | AC-7                | `.moai/project/db/schema.md`, `erd.mmd` 생성기 |
| REQ-008     | AC-8                | 모든 consumer                                   |
| REQ-009     | AC-9                | 훅 핸들러 방어 계층                             |
| REQ-010     | AC-10               | 훅 핸들러, `/moai db refresh`                   |

### Parent SPEC Traceability

- SPEC-DB-SYNC-001 REQ-008(파서 모듈 구현 상세 out-of-scope 명시) → 본 SPEC 전체가 해당 구현을 담당
- SPEC-DB-CMD-001 `/moai db refresh`·`/moai db verify` 스텁 동작 → 본 SPEC 파서 패키지를 소비하여 실동작으로 승격

### Dependency Graph

```
SPEC-DB-CMD-001 (stubs)
        │
        ▼
SPEC-DB-SYNC-001 (hook + orchestration)  ──── SPEC-DB-SYNC-RELOC-001 (call-site relocation)
        │                                              │
        │                                              │ (no conflict — parser package is call-site agnostic)
        ▼                                              │
SPEC-DB-SYNC-002 (this SPEC — parser backends) ◄──────┘
        │
        ▼
(Future) SPEC-DB-SYNC-00X (AST-based backends, diff rules, etc.)
```

## Open Questions

1. **AST library adoption**: 정규식으로 처리 불가능한 엣지 케이스 비율이 실 프로젝트에서 몇 %인지 파일럿으로 측정해야 한다. 20% 이상이면 AST 백엔드를 후속 SPEC에서 도입.
2. **`db.yaml` MaxFileSize override**: 현재 package-level var만 노출. 후속 SPEC에서 `db.yaml` → config binding 추가 여부 결정.
3. **Mermaid ERD의 관계 표기**: 일대다/다대다 구분을 ForeignKey 단일 필드로 충분히 표현 가능한지, 아니면 relation cardinality 명시 필드가 필요한지 구현 중 재검토.

---

**Version**: 0.1.0
**Status**: draft
**Expected Next Step**: `/moai plan SPEC-DB-SYNC-002` 실행 → `plan.md`·`acceptance.md` 생성(본 세션 범위 외)
