---
id: SPEC-DB-TEMPLATES-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-21
author: moai-adk-go
priority: High
labels: [db, templates, schema-docs, erd, migrations, rls, template-first]
issue_number: null
depends_on: []
related_specs: [SPEC-DB-CMD-001, SPEC-DB-SYNC-001]
---

# SPEC-DB-TEMPLATES-001: .moai/project/db/ 7-File Template Set and db.yaml Config Section

## HISTORY

- 2026-04-21 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정.
  - MP-3 critical: frontmatter에 `version`, `labels` 필드 추가, `created` → `created_at` 리네임.
  - CN-1 major: REQ-002 "6 files" → "7 files"로 수정하여 AC-1과 일치.
  - CN-1 major: REQ-010을 `db.yaml` 실제 구조(8 keys)와 일치시키고 `interview_fields` 의미 명확화.
  - Traceability major: REQ-005~REQ-009, REQ-014에 대응하는 AC-10~AC-15 추가.
  - Testability minor: AC-7 `"false"` → 비인용 `false` (yq native boolean 출력) + yq 구현체 명시.
  - Clarity minor: README.md의 `_TBD_` 포함 의무 명시하여 AC-4 경계 확정.
  - Format minor: 각 AC에 해당 REQ 참조 추가하여 traceability matrix 강화.
  - Rationale 추가: 모든 REQ(1~14)에 한 줄 근거 명시.
- 2026-04-20 v0.1.0: SPEC 최초 작성.

## Background

`/moai db init` 워크플로우가 렌더링할 6개 데이터베이스 문서 템플릿 파일 + 1개 README.md와 `.moai/config/sections/db.yaml` 설정 섹션을 정의합니다. 기존 `.moai/project/brand/` 디렉토리 패턴을 참조하여, 모든 `.md` 템플릿은 `_TBD_` 마커를 포함한 스켈레톤 형태로 제공되고 `/moai db init` 인터뷰를 통해 사용자 맞춤으로 채워집니다.

이 SPEC은 "정적 자산 정의"에 한정됩니다. 실제 PostToolUse 자동 동기화 훅 로직, `/moai db` 커맨드 구현, 마이그레이션 파일 파서는 각각 SPEC-DB-SYNC-001과 SPEC-DB-CMD-001의 범위입니다.

기존 컨텍스트:

- `moai-domain-database` 스킬: DB 구현 (PostgreSQL, MongoDB, Redis, Oracle) 담당 — 별개 관심사
- 신규 `moai-domain-db-docs` 스킬 (SPEC-DB-SYNC-001): DB 문서 생성 담당 — 본 SPEC의 템플릿을 스켈레톤으로 사용
- `.moai/project/brand/` 3 파일 패턴: `_TBD_` 마커 + 인터뷰 초기화 방식의 선례

## Scope

### IN SCOPE

- 7개 파일 세트 (README.md + 6개 콘텐츠 템플릿: schema.md, erd.mmd, migrations.md, rls-policies.md, queries.md, seed-data.md)
- `.moai/config/sections/db.yaml` 섹션 파일 (8 키 구조: 5개 고정 키 + 3개 인터뷰 입력 키)
- 사용자 수정 보호 동작 (REQ-013)
- Template-First HARD rule 준수: 템플릿은 `internal/template/templates/` 하위에만 추가

### OUT OF SCOPE (Exclusions — What NOT to Build)

- 실제 PostToolUse 훅 핸들러 구현 (→ SPEC-DB-SYNC-001)
- `moai-domain-db-docs` 스킬 본문 작성 (→ SPEC-DB-SYNC-001)
- `/moai db` 슬래시 커맨드 동작 정의 (→ SPEC-DB-CMD-001)
- 마이그레이션 파일 파서 (Prisma AST, Alembic AST 등) (→ SPEC-DB-SYNC-001)
- erd.mmd 자동 재생성 로직 (→ SPEC-DB-SYNC-001)
- `moai-domain-database` 스킬 수정 (별개 관심사)
- 실제 DB 엔진 연결 또는 스키마 introspection

## Requirements (EARS)

각 REQ는 한 줄 Rationale을 포함하며, 최소 1개 이상의 Acceptance Criterion으로 검증됩니다. Traceability Matrix는 본 섹션 말미에 제공됩니다.

### Event-Driven Requirements

1. **REQ-001**: WHEN moai installs templates via embedded.go, it SHALL include `internal/template/templates/.moai/project/db/README.md` as a new template source.
   - **Rationale**: README.md가 embedded.go에 포함되어야 `moai init` 시 사용자 프로젝트에 동기 배포된다.

2. **REQ-002**: WHEN `/moai db init` renders templates, it SHALL produce 7 files in `.moai/project/db/` with exact filenames: README.md, schema.md, erd.mmd, migrations.md, rls-policies.md, queries.md, seed-data.md.
   - **Rationale**: 7개 파일(README 1 + 콘텐츠 6)이 DB 문서 세트의 최소 단위이며, Scope의 IN SCOPE 항목과 1:1 대응한다.

3. **REQ-003**: WHEN each rendered `.md` template (including README.md) is opened, it SHALL contain at least one `_TBD_` marker in a user-fillable section, following the brand-voice.md convention.
   - **Rationale**: `_TBD_` 마커는 인터뷰 대상 필드를 시각적으로 드러내어 사용자 입력을 유도하며, README.md도 `Last reviewed: _TBD_` 형식으로 포함해야 AC-4의 count ≥ 6 보장된다.

4. **REQ-004**: WHEN erd.mmd is rendered, it SHALL start with the line `erDiagram` as the first non-comment line, producing a valid Mermaid diagram header.
   - **Rationale**: Mermaid ER 다이어그램 문법에서 `erDiagram`은 필수 선언으로, 누락 시 렌더링이 완전히 실패한다.

5. **REQ-005**: WHEN schema.md is rendered, it SHALL include level-2 headings: `## Tables` (or `## Collections` for NoSQL), `## Relationships`, `## Indexes`, `## Constraints`.
   - **Rationale**: 스키마 문서의 권위 있는 구조를 강제하여 downstream consumer(`moai-domain-db-docs`, erd.mmd 재생성)가 안정적으로 섹션을 파싱할 수 있다.

6. **REQ-006**: WHEN migrations.md is rendered, it SHALL include level-2 headings: `## Applied Migrations`, `## Pending Migrations`, `## Rollback Notes`.
   - **Rationale**: 마이그레이션 이력의 3단 구조(적용/대기/롤백)는 팀 협업 시 배포 상태 추적의 최소 요건이다.

7. **REQ-007**: WHEN rls-policies.md is rendered, it SHALL include a skeleton with at least one of the following level-2 headings: `## Supabase RLS Policies` or `## PostgreSQL Policies`, plus commented-out policy examples.
   - **Rationale**: RLS는 보안 경계의 문서화가 누락되면 프로덕션 취약점으로 직결되므로 스켈레톤 강제가 필요하다.

8. **REQ-008**: WHEN queries.md is rendered, it SHALL include level-2 headings: `## Common Queries`, `## Aggregations`, `## Reports`.
   - **Rationale**: 쿼리 문서를 목적별(일반/집계/리포트)로 분류하여 재사용성과 리뷰 효율을 높인다.

9. **REQ-009**: WHEN seed-data.md is rendered, it SHALL include level-2 headings: `## Seed Strategy`, `## Fixture Locations`, `## Dev vs Prod Data`.
   - **Rationale**: 시드 데이터의 전략/위치/환경 구분은 프로덕션 데이터 오염 사고를 예방하는 최소 안전 장치이다.

10. **REQ-010**: WHEN db.yaml section file is rendered, it SHALL contain exactly 8 keys directly under the top-level `db:` node — 5 fixed keys (`enabled`, `dir`, `auto_sync`, `migration_patterns`, `engine`) AND 3 interview-input keys (`orm`, `multi_tenant`, `migration_tool`). The 3 interview-input keys SHALL default to the literal string `"_TBD_"` (or `"none"` for `multi_tenant`) until `/moai db init` updates them.
    - **Rationale**: 5개 고정 키는 시스템 동작(훅, 경로, 패턴)을 결정하고 3개 인터뷰 키는 프로젝트 정체성을 결정한다. 8 키 구조가 db.yaml 참조 예시와 일치하여 내부 모순을 제거한다.

11. **REQ-011**: WHEN db.yaml is rendered, `auto_sync.excluded_patterns` SHALL include `".moai/project/db/**"` verbatim to prevent hook recursion.
    - **Rationale**: PostToolUse 훅이 자신이 쓰는 파일을 다시 트리거하면 무한 루프가 발생한다. 이 패턴은 재귀 차단의 단일 방어선이다.

12. **REQ-012**: WHEN db.yaml is rendered, `migration_patterns` SHALL include at least 6 patterns covering Prisma, Alembic, Rails, raw SQL, Supabase, generic SQL.
    - **Rationale**: 6개 주요 마이그레이션 도구를 기본 커버하여 `moai init` 직후에도 주요 생태계에서 즉시 hook이 활성화된다.

### Unwanted Behavior Requirements

13. **REQ-013**: IF `/moai db init` runs in a project where `.moai/project/db/` already contains user-modified files, THEN the system SHALL NOT overwrite them AND SHALL log a warning listing preserved files.
    - **Rationale**: 사용자가 편집한 스키마/쿼리/정책 문서가 재초기화로 소실되면 신뢰성 붕괴. 보존 + 경고 로그가 최소 사용자 안전 계약이다.

### Event-Driven Requirements (Deployment)

14. **REQ-014**: WHEN `/moai init` (or `moai init`) is run, the template system SHALL deploy db.yaml to `.moai/config/sections/db.yaml` alongside other section files (language.yaml, quality.yaml, user.yaml, workflow.yaml).
    - **Rationale**: db.yaml이 다른 섹션 파일과 동일 위치에 배포되어야 설정 로더가 단일 glob 패턴으로 일관되게 로드한다.

### Traceability Matrix (REQ → AC)

| REQ | AC(s) | Coverage |
|-----|-------|----------|
| REQ-001 | AC-1, AC-8 | File existence + embedded.go regeneration |
| REQ-002 | AC-1 | 7-file enumeration verified |
| REQ-003 | AC-4 | `_TBD_` count ≥ 6 across 6 `.md` files (README 포함) |
| REQ-004 | AC-3 | `erDiagram` first non-comment line |
| REQ-005 | AC-10 | schema.md section headings |
| REQ-006 | AC-11 | migrations.md section headings |
| REQ-007 | AC-12 | rls-policies.md skeleton |
| REQ-008 | AC-13 | queries.md section headings |
| REQ-009 | AC-14 | seed-data.md section headings |
| REQ-010 | AC-2, AC-16 | YAML validity + 8-key structure |
| REQ-011 | AC-5 | Recursion guard pattern present |
| REQ-012 | AC-6 | migration_patterns length ≥ 6 |
| REQ-013 | AC-9 | Idempotent user protection |
| REQ-014 | AC-15 | Runtime deployment destination |

## Target Folder Structure

```
.moai/project/db/
├── README.md            # folder rules, auto-sync policy, update workflow
├── schema.md            # tables/collections definitions (authoritative)
├── erd.mmd              # Mermaid erDiagram (auto-regenerated)
├── migrations.md        # applied + pending migration history
├── rls-policies.md      # Row-level security / access control
├── queries.md           # common query patterns, aggregations, reports
└── seed-data.md         # seed strategy, fixture locations
```

## Target db.yaml Section File

`internal/template/templates/.moai/config/sections/db.yaml`:

```yaml
# Database Documentation Configuration
# Controls /moai db workflow, auto-sync hook, and schema documentation policy.
#
# Structure: 8 keys directly under `db:` (REQ-010)
#   - 5 system-fixed keys: enabled, dir, auto_sync, migration_patterns, engine
#   - 3 interview-input keys: orm, multi_tenant, migration_tool (filled by `/moai db init`)

db:
  # [System] Enabled when /moai db init has run. Default false until user opts in.
  enabled: false

  # [System] Documentation folder (relative to project root)
  dir: ".moai/project/db"

  # [System] Automatic schema sync on migration file changes
  auto_sync:
    enabled: true
    debounce_seconds: 10
    require_user_approval: true
    excluded_patterns:
      - ".moai/project/db/**"
      - ".moai/cache/**"
      - "**/*.lock"

  # [System] Migration file detection globs (PostToolUse matcher references these)
  migration_patterns:
    - "prisma/schema.prisma"
    - "alembic/versions/**/*.py"
    - "db/migrate/**/*.rb"
    - "migrations/**/*.sql"
    - "supabase/migrations/**/*.sql"
    - "sql/migrations/**/*.sql"

  # [System] Primary database engine (set during /moai db init interview)
  engine: "_TBD_"

  # [Interview] ORM/ODM (set during /moai db init interview)
  orm: "_TBD_"

  # [Interview] Multi-tenant strategy (set during /moai db init interview)
  multi_tenant: "none"

  # [Interview] Migration tool (set during /moai db init interview)
  migration_tool: "_TBD_"
```

## Template Content Skeletons

각 템플릿 파일의 필수 구조(outline)를 정의합니다. 실제 파일 작성은 구현 단계에서 수행되며, 이 섹션은 검증 기준입니다.

### README.md 아웃라인

- 제목: `# .moai/project/db/`
- 목적 (Purpose): 1문단 — DB 스키마 문서의 권위 있는 소스, 마이그레이션 파일과 자동 동기화
- 자동 동기화 정책 (Auto-sync Policy): PostToolUse 훅 트리거, 10초 디바운스
- 업데이트 워크플로우 (Update Workflow): 수동 편집 보존, 충돌 시 AskUserQuestion 호출
- 파일 책임 테이블 (File Responsibilities): schema.md, erd.mmd, migrations.md, rls-policies.md, queries.md, seed-data.md 각각 역할
- 제외 패턴 (Excluded Patterns): 자동 동기화에서 제외되는 대상
- 메타 필드: `Last reviewed: _TBD_` (REQ-003 준수, AC-4 count 보장)

### schema.md 아웃라인

- 프론트매터: engine, orm, last_synced_at, manifest_hash
- `## Tables` — 테이블별: name, description, columns (name/type/nullable/default/index), PK/FK
- `## Relationships` — cardinality, FK 방향
- `## Indexes` — standalone + composite
- `## Constraints` — UNIQUE, CHECK, EXCLUSION
- NoSQL 프로젝트는 `## Tables` 대신 `## Collections` 사용

### erd.mmd 아웃라인

- 주석 헤더: `%% Mermaid ER diagram — auto-generated by moai-domain-db-docs`
- `erDiagram` 라인 (첫 번째 비-주석 라인, REQ-004 강제)
- 플레이스홀더 엔티티: `USER { int id PK string email }`
- 플레이스홀더 관계 주석

### migrations.md 아웃라인

- `## Applied Migrations` 테이블 — filename, applied_at, checksum, up_summary 컬럼
- `## Pending Migrations` — 아직 적용되지 않은 마이그레이션
- `## Rollback Notes` — 마이그레이션별 롤백 단계

### rls-policies.md 아웃라인

- `## Supabase RLS Policies` — 주석 처리된 예시
- `## PostgreSQL Policies` — 주석 처리된 예시
- `## Access Control Matrix` — roles × operations 매트릭스 테이블

### queries.md 아웃라인

- `## Common Queries` — query name, SQL/aggregation, purpose
- `## Aggregations` — 리포트성 쿼리
- `## Reports` — 대시보드 쿼리

### seed-data.md 아웃라인

- `## Seed Strategy` — factory vs fixture vs script
- `## Fixture Locations` — dev/, test/, staging/
- `## Dev vs Prod Data` — seed 대상과 절대 seed 안 되는 데이터

## Acceptance Criteria

검증 도구: Mike Farah `yq` (v4+, Go 구현체). `grep` 및 `yq` 명령은 `internal/template/templates/` 경로에서 실행.

- **AC-1** (REQ-001, REQ-002): `internal/template/templates/.moai/project/db/` contains exactly 7 files (README.md + 6 content templates).
- **AC-2** (REQ-010): `internal/template/templates/.moai/config/sections/db.yaml` exists and parses as valid YAML (`yq eval '.' db.yaml` exits 0).
- **AC-3** (REQ-004): `grep -n "^erDiagram" internal/template/templates/.moai/project/db/erd.mmd` returns exactly 1 match, and the match line number is the first non-comment (non-`%%`) line.
- **AC-4** (REQ-003): `grep -c "_TBD_" internal/template/templates/.moai/project/db/*.md` returns at least 6 total (each of 6 `.md` files — README.md, schema.md, migrations.md, rls-policies.md, queries.md, seed-data.md — contains ≥1 `_TBD_`).
- **AC-5** (REQ-011): `yq '.db.auto_sync.excluded_patterns[]' internal/template/templates/.moai/config/sections/db.yaml` output includes the exact line `.moai/project/db/**`.
- **AC-6** (REQ-012): `yq '.db.migration_patterns | length' internal/template/templates/.moai/config/sections/db.yaml` returns an integer ≥ 6.
- **AC-7** (REQ-010): `yq '.db.enabled' internal/template/templates/.moai/config/sections/db.yaml` returns the unquoted YAML boolean `false` (Mike Farah yq native output; no surrounding quotes).
- **AC-8** (REQ-001): `go test ./internal/template/...` passes after `make build` regenerates embedded.go.
- **AC-9** (REQ-013): Running `/moai db init` twice in the same directory preserves the second run's user-modified files (idempotent with user protection per REQ-013); the second run logs a warning listing preserved files.
- **AC-10** (REQ-005): `grep -cE "^## (Tables|Collections|Relationships|Indexes|Constraints)$" internal/template/templates/.moai/project/db/schema.md` returns ≥ 4 (one optional alternative between Tables/Collections; remaining 3 mandatory).
- **AC-11** (REQ-006): `grep -cE "^## (Applied Migrations|Pending Migrations|Rollback Notes)$" internal/template/templates/.moai/project/db/migrations.md` returns exactly 3.
- **AC-12** (REQ-007): `grep -cE "^## (Supabase RLS Policies|PostgreSQL Policies)$" internal/template/templates/.moai/project/db/rls-policies.md` returns ≥ 1, AND the file contains at least one commented-out policy example (line starting with `<!--` or `--` or `#`).
- **AC-13** (REQ-008): `grep -cE "^## (Common Queries|Aggregations|Reports)$" internal/template/templates/.moai/project/db/queries.md` returns exactly 3.
- **AC-14** (REQ-009): `grep -cE "^## (Seed Strategy|Fixture Locations|Dev vs Prod Data)$" internal/template/templates/.moai/project/db/seed-data.md` returns exactly 3.
- **AC-15** (REQ-014): After running `moai init` in a fresh temp directory, `.moai/config/sections/db.yaml` exists at the deployed path and its content equals the template source (`diff` returns empty). Verified alongside the existence of `.moai/config/sections/language.yaml` and `.moai/config/sections/quality.yaml`.
- **AC-16** (REQ-010): `yq '.db | keys | length' internal/template/templates/.moai/config/sections/db.yaml` returns exactly 8, AND `yq '.db | keys' db.yaml` output includes all of: `enabled`, `dir`, `auto_sync`, `migration_patterns`, `engine`, `orm`, `multi_tenant`, `migration_tool`.

## Risks

- **R-1**: 예약된 파일명이 사용자 파일과 충돌할 수 있음 → REQ-013의 사용자 보호 메커니즘으로 완화 (AC-9 검증).
- **R-2**: erd.mmd Mermaid 문법 오류가 문서 렌더링을 깨뜨릴 위험 → AC-3이 헤더를 강제; 전체 문법 검증은 SPEC-DB-SYNC-001 단계에서 수행.
- **R-3**: db.yaml의 opt-in (`enabled: false`) 기본값이 자동 활성화를 기대한 사용자에게 혼란을 줄 수 있음 → `/moai db init`이 명시적으로 `enabled: true`로 설정; README.md에 동작 방식 명시.
- **R-4**: `_TBD_` 문자열이 사용자 실제 데이터와 충돌할 이론적 가능성 → `_TBD_`는 brand-voice.md에서 이미 사용하는 확립된 sentinel이며, 사용자가 의도적으로 이 문자열을 입력할 가능성은 낮다. 필요 시 SPEC-DB-SYNC-001에서 escape 메커니즘 도입.

## Dependencies

- **depends_on**: 없음 (순수 정적 자산 추가)
- **related_specs**:
  - SPEC-DB-CMD-001 — `/moai db` 슬래시 커맨드 구현 (본 SPEC의 템플릿 소비자)
  - SPEC-DB-SYNC-001 — auto-sync 훅 + `moai-domain-db-docs` 스킬 (본 SPEC의 템플릿 기반 동작)

## Validation Strategy

- **정적 검증**: AC-1~AC-7, AC-10~AC-14, AC-16은 파일 존재성 + YAML 파싱 + grep/yq 쿼리로 자동 검증
- **빌드 검증**: AC-8은 `make build` 후 `go test ./internal/template/...`로 embedded.go 회귀 방지
- **통합 검증**: AC-15는 `moai init`을 fresh temp directory에서 실행하여 배포 대상 경로를 확인; 통합 테스트(`internal/cli/init_test.go` 또는 전용 테스트)로 자동화 권장
- **동작 검증**: AC-9는 `/moai db init` 수동 실행 2회로 idempotency 확인; 구현 완료 시 통합 테스트로 자동화 권장 (SPEC-DB-CMD-001에서 처리)

## Notes on AC Format Convention

본 SPEC의 AC는 "EARS 자연어" 대신 "binary-testable shell command + expected output" 형식을 채택합니다. 이는 MoAI-ADK 코드베이스 전반의 관례(예: SPEC-THIN-CMDS-001의 audit test 패턴)와 일치하며, CI 자동화 친화성을 위해 의도적으로 선택된 형식입니다. 각 AC는 단일 검증 명령어와 기대 결과로 구성되어 외부 모호성 없이 pass/fail 판정이 가능합니다.

EARS 자연어 변환이 필요한 경우 다음 패턴으로 재해석 가능:

> "WHEN the template source tree is inspected after `make build`, the system SHALL present files and YAML structure such that [verification command] returns [expected output]."
