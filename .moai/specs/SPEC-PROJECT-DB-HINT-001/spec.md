---
id: SPEC-PROJECT-DB-HINT-001
version: 0.2.1
status: draft
created_at: 2026-04-20
updated_at: 2026-04-21
author: moai-adk-go
priority: Low
labels: [project, db-detection, workflow-extension, next-steps, hint, multi-language]
issue_number: null
depends_on: [SPEC-DB-CMD-001]
related_specs: [SPEC-DB-TEMPLATES-001, SPEC-DB-SYNC-001]
---

# SPEC-PROJECT-DB-HINT-001: `/moai project` Phase 4 DB 감지 → `/moai db init` Next Step 힌트

## HISTORY

- 2026-04-21 v0.2.1: plan-auditor iteration 2 FAIL 후속 1-line 수정. frontmatter 필드명 `created`/`updated` → `created_at`/`updated_at` (MoAI 표준 엄격 준수). MP-3 PASS 확보.
- 2026-04-21 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. MoAI 표준 frontmatter 적용(version/labels/updated/issue_number 추가), 16개 언어 의존성 파일/ORM 키워드 전면 확장(REQ-003/004 + Detection Keywords Reference), REQ별 1-line Rationale 추가, 누락 REQ(REQ-001~004, REQ-010, REQ-011)에 대한 AC 6개 신설, REQ-008/011/012에서 구현 세부사항(정확한 copy 문자열·파일 경로·툴 플래그)을 Phase 4.1a Skeleton으로 이동, DB 엔진 목록에 CockroachDB/Cassandra/ScyllaDB/Riak 등 추가, AC를 EARS-style 또는 shell-verifiable 형식으로 통일, REQ-010 EARS 'then' 연결사 추가, REQ-006의 "optional second tier" 표현을 구체적 위치 기반으로 재작성.
- 2026-04-20 v0.1.0: SPEC 최초 작성.

## Background

`/moai project` 워크플로우는 현재 Phase 4.2(Next Steps)에서 사용자에게 3가지 다음 단계 선택지(Create SPEC / Review / Done)를 제안한다. 하지만 프로젝트가 DB 기술을 포함하고 있을 경우, 사용자는 별도로 `/moai db init` 명령의 존재를 학습하고 수동으로 실행해야 하는 불편함이 있다.

본 SPEC은 `/moai project` 마무리 단계에서 자연스럽게 `/moai db init`를 소개하는 힌트를 추가한다. 자동 실행은 절대 금지하며, 사용자 선택지에 추가만 하는 비침습적 UX 확장이다. 기존 Phase 4.2 Next Steps의 구조를 유지하면서 DB 감지 결과에 따라 선택지 하나를 조건부로 삽입한다.

핵심 설계 원칙:

- **비침습적**: 기존 3가지 옵션은 그대로 유지, 감지 결과에 따라서만 추가 옵션 노출
- **발견 가능성**: 사용자가 `/moai db` 명령의 존재를 알지 못해도 자연스럽게 안내
- **사용자 주도**: 자동 실행 금지, 사용자가 직접 다음 턴에 `/moai db init`를 입력
- **멱등성**: `.moai/project/db/`가 이미 존재하면 `init` 대신 `refresh` 제안
- **언어 중립성**: MoAI가 지원하는 16개 언어를 동등하게 취급하며, 특정 언어 우대 금지

## Requirements (EARS)

### REQ-001: Phase 4.1a 진입 조건

WHEN `/moai project` Phase 3.7 (Development Methodology Auto-Configuration) completes, the workflow SHALL proceed to Phase 4.1a (DB Detection) before Phase 4.2 (Next Steps).

**Rationale**: DB 감지는 tech.md가 확정된 이후(Phase 3.7) Phase 4.2 Next Steps에서 소비되어야 하므로 순서 고정이 필요하다. 순서가 뒤바뀌면 감지 결과를 Next Steps가 읽지 못한다.

### REQ-002: tech.md 키워드 스캔

WHEN Phase 4.1a runs, it SHALL read `.moai/project/tech.md` and search for DB keywords defined in the Detection Keywords Reference (DB Engines section) case-insensitively.

**Rationale**: tech.md는 Phase 3에서 생성된 사용자 프로젝트 기술 스택 캐논 문서이며, DB 기술의 1차 선언 소스이다. 이를 누락하면 핵심 신호 하나를 잃는다.

### REQ-003: 16개 언어 의존성 파일 스캔

WHEN Phase 4.1a runs, it SHALL scan project dependency manifest files for all 16 MoAI-supported languages (Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift) as enumerated in the Detection Keywords Reference (Dependency Files section), plus standalone SQL migration directories.

**Rationale**: MoAI는 16개 언어를 동등 지원하는 것이 헌법적 원칙(CLAUDE.local.md §15)이다. 일부 언어만 스캔하면 동등성 원칙이 깨지고 비-Node/Python 프로젝트 사용자에게 DB 힌트가 제공되지 않는 편향이 발생한다.

### REQ-004: 16개 언어 ORM/ODM 키워드 스캔

WHEN Phase 4.1a finds a dependency manifest file of size ≤ file_size_limit, it SHALL search the manifest content for ORM/ODM keywords for the corresponding language as enumerated in the Detection Keywords Reference (ORMs / ODMs section), matching case-insensitively.

**Rationale**: ORM/ODM 선언은 단순 DB 이름 언급보다 강한 신호이며(Risk R-1 완화), 언어별로 dominant ORM이 다르므로(Java의 Hibernate, Ruby의 ActiveRecord, Swift의 Core Data 등) 16개 언어 각각에 대해 매칭 사전이 필요하다.

### REQ-005: 감지 결과 상태 기록

WHEN any DB or ORM keyword is found, Phase 4.1a SHALL set `detected_db=true` and record the matched keyword(s) and source file(s) to a state artifact for Phase 4.2 consumption.

**Rationale**: Phase 4.1a(감지)와 Phase 4.2(UI 분기)는 관심사가 분리되어 있으므로 중간 매개체가 필요하다. 매칭된 키워드와 출처를 함께 기록해야 사용자에게 왜 DB 힌트가 등장했는지 투명하게 설명 가능하다.

### REQ-006: DB 감지 시 Next Steps 확장 (신규 프로젝트)

WHEN `detected_db` is true AND `.moai/project/db/` does NOT already exist, Phase 4.2 Next Steps AskUserQuestion SHALL include an additional option "Initialize DB documentation (`/moai db init`)" marked as Recommended and placed as the first option, shifting the existing "Create SPEC" option to the second position.

**Rationale**: 신규 DB 프로젝트에서는 SPEC 작성 전에 DB 스키마/문서를 먼저 구축하는 것이 의존성 순서상 유리하며, Recommended 플래그가 사용자를 올바른 순서로 유도한다.

### REQ-007: DB 감지 시 Next Steps 확장 (기존 프로젝트)

WHEN `detected_db` is true AND `.moai/project/db/` already exists, Phase 4.2 SHALL NOT add the `/moai db init` option, and SHALL instead append "Refresh DB documentation (`/moai db refresh`)" as the fourth option (appearing after the existing three options) without altering the order or Recommended flag of the existing options.

**Rationale**: 이미 DB 문서가 초기화된 프로젝트에서 `init`을 다시 제안하면 멱등성이 깨진다. `refresh`는 별도 유스케이스(tech.md 업데이트 반영)이므로 기본 3옵션 하단에 추가 옵션으로만 노출한다.

### REQ-008: DB 미감지 시 기존 동작 유지

WHEN `detected_db` is false, Phase 4.2 SHALL present the existing three options (Create SPEC / Review / Done) unchanged, with no `/moai db` hint injected.

**Rationale**: DB를 사용하지 않는 CLI/라이브러리 프로젝트에서 DB 힌트는 노이즈이다. 미감지 경로는 기존 UX를 완전 보존해야 회귀 불만이 없다.

### REQ-009: 선택 시 가이드 메시지 표시

WHEN the user selects the `/moai db init` option from Next Steps, the orchestrator SHALL display a guidance message describing the upcoming interview rounds and the artifact directory, AND SHALL NOT auto-execute the `/moai db init` command.

**Rationale**: 사용자는 다음에 무슨 일이 일어날지 알 권리가 있고, 자동 실행 금지 원칙(사용자 주도)을 지키면서도 발견 가능성을 확보해야 한다. 구체적 copy 문자열은 WHAT이 아닌 HOW이므로 Phase 4.1a Skeleton에서 정의한다.

### REQ-010: 자동 실행 금지

WHEN the user selects the `/moai db init` option, the orchestrator SHALL end `/moai project` and return control to the user so that the user invokes `/moai db init` themselves in a subsequent turn.

**Rationale**: 사용자 주도 원칙은 본 SPEC의 constitutional constraint이다. 자동 chaining은 투명성을 훼손하고 권한 경계를 모호하게 만든다.

### REQ-011: tech.md 미존재 엣지 케이스

IF `.moai/project/tech.md` does not exist (edge case: Phase 3 failed or skipped), THEN Phase 4.1a SHALL skip gracefully without error AND SHALL set `detected_db=false`.

**Rationale**: Phase 3 실패가 Phase 4를 크래시시키면 사용자는 프로젝트 초기화 자체를 재시도해야 한다. 우아한 fallback이 robustness에 필수다.

### REQ-012: 스캔 제한 및 대소문자 무시

WHILE Phase 4.1a scans dependency manifest files, it SHALL skip files exceeding the configured size limit AND SHALL perform keyword matching case-insensitively.

**Rationale**: 대용량 generated lockfile(package-lock.json 등)은 스캔 비용을 수 배로 늘리며, 대소문자 구분 매칭은 `postgresql` vs `PostgreSQL` 같은 자연스러운 표기 차이에서 false negative를 유발한다. 정확한 임계값과 매칭 툴은 HOW이므로 Skeleton으로 이동.

### REQ-013: 상태 지속성 및 무효화

WHEN Phase 4.1a writes its detection result, it SHALL persist a state record that (a) allows Phase 4.2 to read detection results deterministically and (b) enables stale-detection via a hash derived from tech.md content.

**Rationale**: tech.md가 변경되었는데 감지 결과가 재생성되지 않으면 오래된 DB 힌트가 노출된다. 내용 해시를 기록해야 후속 실행에서 무효화 판단이 가능하다. 구체적 파일 경로(`.moai/state/db-detection.json`)와 키 목록은 HOW이므로 Skeleton으로 이동.

## Acceptance Criteria

아래 AC는 EARS-style(WHEN/THEN) 또는 shell-verifiable 명령으로 기술한다. 각 AC는 하나 이상의 REQ를 추적한다.

- **AC-1 (covers REQ-001)**: WHEN implementation is complete, THEN `.claude/skills/moai/workflows/project.md` SHALL contain a section titled `### Phase 4.1a: DB Detection` positioned between the Phase 3.7 section and the Phase 4.2 section. Verifiable by: `grep -n "### Phase 4.1a: DB Detection" .claude/skills/moai/workflows/project.md` returning a line number strictly between the Phase 3.7 and Phase 4.2 line numbers.
- **AC-2 (covers REQ-002)**: WHEN Phase 4.1a executes in a project whose `.moai/project/tech.md` contains "PostgreSQL", THEN `detected_db` SHALL be set to true with at least the matched keyword "postgresql" recorded in state.
- **AC-3 (covers REQ-003)**: WHEN implementation is complete, THEN the Detection Keywords Reference SHALL enumerate dependency manifest markers for all 16 MoAI-supported languages, plus standalone SQL migration directory pattern, verified by cross-check against `.moai/project/tech.md` canonical language list.
- **AC-4 (covers REQ-004)**: WHEN Phase 4.1a runs in a Kotlin project whose `build.gradle.kts` declares `org.jetbrains.exposed:exposed-core`, THEN `detected_db` SHALL be true with "exposed" recorded as matched keyword (verifying non-Node/Python language coverage).
- **AC-5 (covers REQ-005)**: WHEN any DB/ORM keyword match occurs, THEN the state artifact SHALL include both the matched keyword(s) and the source file path that produced the match.
- **AC-6 (covers REQ-006)**: WHEN `/moai project` runs in a Next.js+Prisma project where `package.json` declares a `prisma` dependency and `.moai/project/db/` is absent, THEN Phase 4.2 Next Steps SHALL present "Initialize DB documentation (`/moai db init`)" as the first option marked Recommended, and "Create SPEC" SHALL appear as the second option.
- **AC-7 (covers REQ-007)**: WHEN `/moai project` runs in a project where `tech.md` mentions PostgreSQL AND `.moai/project/db/` directory already exists, THEN Phase 4.2 SHALL present the existing three options (Create SPEC/Review/Done) unchanged in order and Recommended flag, AND SHALL append "Refresh DB documentation (`/moai db refresh`)" as a fourth option.
- **AC-8 (covers REQ-008)**: WHEN `/moai project` runs in a CLI-tool project with no DB keywords in `tech.md` and no ORM in dependency files, THEN Phase 4.2 SHALL present exactly the original three options (Create SPEC/Review/Done) with no `/moai db` related option injected.
- **AC-9 (covers REQ-009, REQ-010)**: WHEN the user selects the `/moai db init` option from Next Steps, THEN the orchestrator SHALL emit a guidance message describing the interview rounds and artifact location, AND SHALL terminate `/moai project` without invoking `/moai db init` automatically.
- **AC-10 (covers REQ-011)**: WHEN `/moai project` runs in an environment where `.moai/project/tech.md` is absent (simulating Phase 3 failure), THEN Phase 4.1a SHALL complete without raising an error, `detected_db` SHALL be false, AND Phase 4.2 SHALL present the original three options unchanged.
- **AC-11 (covers REQ-012)**: WHEN Phase 4.1a encounters a dependency manifest larger than the configured size limit, THEN the file SHALL be skipped (not scanned), AND keyword matching for all scanned files SHALL be case-insensitive (verifiable by `lowercase/UPPERCASE/MixedCase` keyword variants in a fixture all triggering detection).
- **AC-12 (covers REQ-013)**: WHEN Phase 4.1a completes, THEN a state record SHALL exist containing (a) detection flag, (b) matched keywords, (c) scan timestamp, and (d) a content hash of the scanned `tech.md`, enabling Phase 4.2 to detect stale detections when tech.md changes.
- **AC-13 (Quality Gate)**: WHEN implementation is complete, THEN `go test ./internal/template/...` SHALL pass with no new failures introduced.

## Phase 4.1a Skeleton (Implementation Reference)

구현 시 `.claude/skills/moai/workflows/project.md`에 다음 골격을 삽입한다. 아래는 REQ가 금지하지 않는 구체적 HOW(파일 경로, 임계값, 툴 선택, copy 문자열)이며 REQ가 정의하는 WHAT이 아니다.

```
### Phase 4.1a: DB Detection

Purpose: Detect database technology from generated documentation and dependency
files to conditionally propose /moai db init in Next Steps.

[HARD] This phase runs automatically without user interaction.

Steps:
1. Check .moai/project/tech.md exists. If not: set detected_db=false and skip.
2. Grep tech.md for DB engine keywords (case-insensitive).
3. Glob for dependency manifests across all 16 supported languages
   (see Detection Keywords Reference → Dependency Files section).
4. For each found file ≤ 1 MB: grep for ORM/ODM keywords relevant to that language.
5. Aggregate matches {detected, matched_keywords[], source_files[], scanned_at, tech_md_hash}.
6. Write state artifact at .moai/state/db-detection.json.
7. Proceed to Phase 4.2 with detected_db flag.

Guidance message on user selection (REQ-009):
  "/moai db init will run 4 interview rounds (engine selection, connection config,
   schema survey, migration strategy) and create .moai/project/db/ templates. Run
   it in your next turn."

File size limit: 1 MB (skip larger manifests to avoid scanning generated lockfiles).

Tool choice: Grep with -i (case-insensitive) for keyword matching; Glob for
manifest discovery.
```

## Detection Keywords Reference

Phase 4.1a가 참조하는 키워드 목록은 다음과 같으며, 대소문자를 구분하지 않고 매칭한다. ORM/ODM 매칭은 `tech.md`에만 언급된 DB 이름보다 강한 신호로 취급한다(Risk R-1 완화 근거).

### DB Engines

Relational / SQL:
- PostgreSQL
- MySQL
- MariaDB
- SQLite
- Oracle
- SQL Server / MSSQL
- CockroachDB
- Supabase
- Neon
- Planetscale

NoSQL Document:
- MongoDB
- Firestore
- Firebase
- Couchbase

NoSQL Key-Value / Wide-column:
- Redis
- DynamoDB
- Cassandra
- ScyllaDB
- Riak

Search / Analytics:
- Elasticsearch
- ClickHouse
- Snowflake
- InfluxDB

### Dependency Files (16 MoAI-supported languages + SQL standalone)

| 언어 (canonical name) | 의존성 매니페스트 파일 |
|---|---|
| go | `go.mod`, `go.sum` |
| python | `requirements.txt`, `pyproject.toml`, `Pipfile`, `setup.py` |
| typescript | `package.json`, `tsconfig.json` |
| javascript | `package.json` |
| rust | `Cargo.toml`, `Cargo.lock` |
| java | `pom.xml`, `build.gradle` |
| kotlin | `build.gradle.kts`, `build.gradle` |
| csharp | `*.csproj`, `packages.config`, `Directory.Packages.props` |
| ruby | `Gemfile`, `Gemfile.lock`, `*.gemspec` |
| php | `composer.json`, `composer.lock` |
| elixir | `mix.exs`, `mix.lock` |
| cpp | `CMakeLists.txt`, `conanfile.txt`, `conanfile.py`, `vcpkg.json` |
| scala | `build.sbt`, `project/plugins.sbt` |
| r | `DESCRIPTION`, `renv.lock` |
| flutter | `pubspec.yaml`, `pubspec.lock` |
| swift | `Package.swift`, `Podfile`, `Podfile.lock` |
| sql-standalone | `migrations/**/*.sql`, `db/migrate/**/*.sql`, `schema.sql` |

### ORMs / ODMs by Language

Go:
- GORM
- SQLc
- Ent
- mongo-go-driver
- sqlx

Python:
- SQLAlchemy
- Django ORM (django.db)
- Tortoise ORM
- Peewee
- python-oracledb
- motor (Mongo async)
- pymongo

TypeScript / JavaScript:
- Prisma
- TypeORM
- Drizzle
- Sequelize
- Mongoose
- Objection
- Kysely
- MikroORM

Rust:
- Diesel
- SQLx
- SeaORM
- mongodb (crate)
- tokio-postgres

Java:
- Hibernate
- JPA / jakarta.persistence
- Spring Data
- MyBatis
- jOOQ

Kotlin:
- Exposed
- Ktorm
- Hibernate (via JVM)
- JPA (via JVM)

C#:
- Entity Framework (EF Core)
- Dapper
- NHibernate
- LINQ to DB

Ruby:
- ActiveRecord
- Sequel
- Mongoid
- ROM-rb

PHP:
- Eloquent (Laravel)
- Doctrine
- Phinx
- CakePHP ORM

Elixir:
- Ecto

C++:
- SOCI
- ODB
- SQLite (direct, via CMake/conan)
- mongocxx

Scala:
- Slick
- Doobie
- Quill
- ScalikeJDBC

R:
- DBI
- dplyr (dbplyr backend)
- RPostgres
- RSQLite
- RMariaDB

Flutter / Dart:
- Drift (formerly Moor)
- sqflite
- hive
- isar
- objectbox

Swift:
- Core Data
- GRDB
- Realm
- SQLite.swift
- FluentKit (Vapor)

## Scope

### IN SCOPE

- Phase 4.1a DB 감지 로직 (tech.md + 16개 언어 의존성 매니페스트 스캔)
- Phase 4.2 Next Steps 조건부 옵션 추가 (DB 미초기화 / 이미 초기화 / 미감지 3-way 분기)
- 감지 결과 상태 지속화 및 tech.md 내용 해시 기반 무효화 메커니즘
- `/moai db init` 선택 시 가이드 메시지 표시 및 `/moai project` 종료
- Detection Keywords Reference의 16개 언어 동등 취급

### OUT OF SCOPE (Exclusions — What NOT to Build)

- `/moai db init` 명령 자체의 구현 (SPEC-DB-CMD-001 범위)
- DB 템플릿 파일 생성 (SPEC-DB-TEMPLATES-001 범위)
- DB 문서 자동 동기화 훅 (SPEC-DB-SYNC-001 범위)
- `/moai project` 기존 Phase 0~3.7 로직 변경 (본 SPEC은 Phase 4.1a만 신설, 그 외 Phase는 불변)
- `/moai db init` 자동 실행 (사용자 주도 원칙 고수)
- DB 이외 도메인(예: 메시지 큐, 캐시 전용 Redis 단독 사용, CDN, 검색 엔진 단독)의 Next Step 힌트 확장 (추후 별도 SPEC)
- 감지된 DB 종류에 따른 SPEC 템플릿 자동 선택 (본 SPEC은 Next Step 힌트에만 국한)
- 의존성 매니페스트 파싱(AST-level). 본 SPEC은 키워드 grep만 수행

## Risks

- **R-1 (False Positive)**: 주석이나 문서에만 DB 이름이 언급된 프로젝트에서 오탐 발생 가능. 완화: 의존성 파일 매칭을 tech.md-only 매칭보다 강한 신호로 간주하며, ORM 키워드를 우선 신호로 사용.
- **R-2 (Stale State)**: `tech.md`가 변경되었으나 상태 record가 구 해시를 가진 경우 오래된 감지 결과 재사용 위험. 완화: REQ-013의 content hash 필드로 무효화 처리.
- **R-3 (UX Reordering)**: 기존 "Create SPEC (Recommended)" 위치가 변경되어 기존 사용자 혼동 가능. 완화: CHANGELOG에 명시적으로 변경 내용 기록, REQ-006/007/008 분기로 DB 미감지 시 기존 UX 완전 보존.
- **R-4 (언어 커버리지 편향)**: 16개 언어 중 일부의 ORM 사전이 불완전할 경우 해당 언어 사용자에게 힌트가 노출되지 않는 2차 편향 발생. 완화: Detection Keywords Reference를 SPEC-DB-CMD-001의 권위 있는 목록과 sync하며, 누락 ORM 발견 시 본 SPEC의 follow-up으로 추가.
- **R-5 (Lockfile 스캔 비용)**: `package-lock.json`, `Cargo.lock`, `poetry.lock` 등 대용량 generated 파일 스캔이 Phase 4.1a 지연을 유발할 위험. 완화: REQ-012의 파일 크기 제한으로 스캔 제외.

## Test Scenarios (Given-When-Then)

### TS-1: Next.js + Prisma 프로젝트 신규 초기화

- **Given**: `package.json`에 `"prisma": "^5.0.0"` 의존성 포함, `.moai/project/db/` 미존재
- **When**: `/moai project` 완료 후 Phase 4.1a 및 4.2 진입
- **Then**: Next Steps 첫 번째 옵션으로 "Initialize DB documentation (`/moai db init`) (Recommended)"가 표시되고, "Create SPEC"은 두 번째로 이동

### TS-2: Kotlin + Exposed 프로젝트 (비-Node/Python 언어 커버리지 검증)

- **Given**: `build.gradle.kts`에 `org.jetbrains.exposed:exposed-core` 의존성 포함, `.moai/project/db/` 미존재
- **When**: `/moai project` 완료 후 Phase 4.1a 진입
- **Then**: `detected_db=true`로 "exposed" 키워드 매칭 기록, Phase 4.2는 DB init 옵션을 첫 번째 Recommended로 노출

### TS-3: CLI 도구 프로젝트 (DB 없음)

- **Given**: `go.mod`에 DB/ORM 의존성 없음, `tech.md`에 DB 키워드 없음
- **When**: `/moai project` 완료 후 Phase 4.1a 진입
- **Then**: `detected_db=false`로 기록, Phase 4.2는 기존 3개 옵션(Create SPEC / Review / Done)만 표시

### TS-4: 이미 초기화된 DB 프로젝트

- **Given**: `tech.md`에 PostgreSQL 언급, `.moai/project/db/` 디렉토리 이미 존재
- **When**: `/moai project` 재실행
- **Then**: Next Steps 상위 3개는 기존 순서/Recommended 그대로, 네 번째 옵션으로 "Refresh DB documentation (`/moai db refresh`)" 추가

### TS-5: tech.md 부재 엣지 케이스

- **Given**: Phase 3가 실패하여 `.moai/project/tech.md`가 생성되지 않음
- **When**: Phase 4.1a 진입
- **Then**: 에러 없이 skip, `detected_db=false`로 Phase 4.2 진행

### TS-6: 대소문자 혼합 매칭

- **Given**: `tech.md`에 "PostgreSQL", `package.json`에 `"@prisma/client"` (혼합 대소문자)
- **When**: Phase 4.1a 스캔
- **Then**: 두 키워드 모두 매칭되어 state에 기록(대소문자 무시 검증)

### TS-7: 대용량 lockfile 스킵

- **Given**: `package-lock.json`이 1 MB 초과(generated)
- **When**: Phase 4.1a 스캔
- **Then**: lockfile은 스캔하지 않고 skip 로그 기록, `package.json`만 스캔 대상으로 포함

## Definition of Done

- [ ] `.claude/skills/moai/workflows/project.md`에 Phase 4.1a 섹션 추가 (AC-1)
- [ ] Phase 4.2 Next Steps 3-way 조건부 분기 로직 문서화 (AC-6/AC-7/AC-8)
- [ ] 감지 상태 record 스키마 Skeleton에 명시 (AC-12)
- [ ] 16개 언어 Dependency Files + ORMs 사전 완비 (AC-3)
- [ ] TS-1 ~ TS-7 일곱 가지 테스트 시나리오 수동 검증 완료
- [ ] `go test ./internal/template/...` 통과 (AC-13)
- [ ] CHANGELOG에 Next Steps UX 변경 및 16개 언어 커버리지 내용 기록
- [ ] SPEC-DB-CMD-001와의 의존 관계 확인 (선행 머지 필요)
- [ ] REQ ↔ AC traceability 매트릭스 후속 PR 설명에 포함 (REQ-001..REQ-013 전부 최소 1개 AC로 커버 확인)
