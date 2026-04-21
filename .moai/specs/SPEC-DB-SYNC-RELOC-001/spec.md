---
id: SPEC-DB-SYNC-RELOC-001
version: 1.0.0
status: completed
created_at: 2026-04-21
updated_at: 2026-04-21
author: moai-adk-go
priority: medium
labels: [db, hook, sync, performance, architecture, relocation]
issue_number: null
depends_on: [SPEC-DB-SYNC-001, SPEC-DB-SYNC-HARDEN-001, SPEC-DB-CMD-001, SPEC-SKILL-GATE-001]
related_specs: []
---

# SPEC-DB-SYNC-RELOC-001: PostToolUse DB 훅 → `/moai sync` Phase 이관

## HISTORY

- 2026-04-21 v1.0.0: 설계 확정 및 구현. 사용자 지시("db sync는 매번 hooks으로 db와 무관한 내용도 매번 처리를 하려고 체크를 하지 않는가? sync때 db 스키마 문서 업데이트가 필요한지 체크해서 업데이트 하도록 하자") 반영. PostToolUse 훅 등록을 제거하고 `/moai sync`의 Phase 0.08로 DB 문서 갱신 로직을 이관. 기존 Go 구현(`internal/hook/dbsync/`), `moai hook db-schema-sync` CLI 서브커맨드, `/moai db refresh` 워크플로우, `moai-domain-db-docs` 스킬은 전부 유지 — 트리거만 이동.

## Background

SPEC-DB-SYNC-001은 `handle-db-schema-change.sh`를 `Write|Edit` 매처의 PostToolUse 훅으로 등록했다. 이는 **모든 Write/Edit 이벤트마다** bash spawn + `moai hook db-schema-sync` 호출을 발생시키는 설계다.

### 실측 비용

- 1 이벤트당 비용: bash 프로세스 기동 + Go binary 로드 + `filepath.Clean` + `matchGlob` 5 패턴 검사 + `IsExcluded` 3 패턴 검사 + 디바운스 상태 파일 읽기 → **~30-60ms**
- 일반 코딩 세션의 Write/Edit 이벤트 수: 수십~수백 회
- 누적 오버헤드: **5-30초 순손실** (대부분 비-DB 파일 편집)

### 실패한 대안

SPEC-DB-SYNC-HARDEN-001의 5개 견고화(크기 가드, 동시성 O_EXCL, 경로 traversal 차단 등)는 정확성 문제는 해결했으나, **근본적인 "매번 실행" 아키텍처 비용**은 건드리지 않았다. 모든 hardening은 "한 번의 훅 호출이 얼마나 올바른가"를 개선했을 뿐, "훅이 왜 매번 호출되어야 하는가"를 되묻지 않았다.

### 올바른 트리거

DB 문서(`.moai/project/db/schema.md`, `erd.mmd`, `migrations.md`)는 **milestone / PR boundary**에서 한 번 최신화하면 충분한 파생 문서이다. 이 지점은 이미 `/moai sync`가 담당한다. 따라서:

1. 훅 등록을 제거하여 Write/Edit마다 발생하던 process spawn 비용 0화
2. `/moai sync`의 pre-quality phase(Phase 0.08)에 git diff 기반 감지를 추가
3. 마이그레이션 파일 변경 감지 시 `moai-domain-db-docs`에 위임하여 한 번 배치 갱신

### Scope Boundary

- Go 내부 구현(`internal/hook/dbsync/`): 전부 유지. `moai hook db-schema-sync` CLI는 manual 호출 및 향후 `/moai db refresh` 내부 재사용을 위해 보존.
- `/moai db refresh`·`/moai db verify` 워크플로우(`workflows/db.md`): 변경 없음.
- `moai-domain-db-docs` 스킬: 변경 없음.
- `.moai/project/db/` 템플릿 7개 파일: 변경 없음.
- `.moai/config/sections/db.yaml`: `auto_sync` 키의 시맨틱 변경 (훅 발동 여부 → sync phase 발동 여부) — 키 자체는 유지.

## Requirements (EARS)

### R1 — PostToolUse 훅 등록 제거

- **REQ-RELOC-HOOK-001** (Ubiquitous): `internal/template/templates/.claude/settings.json.tmpl`은 `handle-db-schema-change.sh`를 참조하는 PostToolUse 엔트리를 포함하지 않는다.
- **REQ-RELOC-HOOK-002** (Ubiquitous): `internal/template/templates/.claude/hooks/moai/handle-db-schema-change.sh` 파일은 존재하지 않는다(삭제). 훅 래퍼가 불필요하므로 번들에서 제거한다.
- **REQ-RELOC-HOOK-003** (Event-driven): WHEN `moai init`으로 신규 프로젝트가 생성될 때, THEN 생성된 `.claude/settings.json`에 `db-schema-change` 관련 엔트리가 없고, `.claude/hooks/moai/handle-db-schema-change.sh`도 존재하지 않는다.

### R2 — `/moai sync` Phase 0.08 DB Schema Doc Check 추가

- **REQ-RELOC-SYNC-001** (Ubiquitous): `.claude/skills/moai/workflows/sync.md`와 `internal/template/templates/.claude/skills/moai/workflows/sync.md` 두 파일에 `### Phase 0.08: DB Schema Doc Check` 섹션이 정확히 1회 등장한다.
- **REQ-RELOC-SYNC-002** (Event-driven): WHEN Phase 0.08이 실행될 때, THEN 다음 단계가 수행된다:
  1. `.moai/config/sections/db.yaml`을 읽어 `db.enabled`가 true인지 확인. false 또는 파일 부재 시 phase 스킵.
  2. `git diff --name-only <base>..HEAD`로 변경 파일 목록을 얻고 `db.migration_patterns`의 글롭 패턴으로 필터링.
  3. 매칭 파일 ≥ 1 AND `db.auto_sync: true`이면 `moai-domain-db-docs` 스킬의 refresh 모드를 호출.
  4. 매칭 파일 ≥ 1 AND `db.auto_sync: false`이면 sync 리포트에 advisory 한 줄 추가 후 phase 스킵.
  5. 매칭 파일 0 이면 phase 스킵.
- **REQ-RELOC-SYNC-003** (Ubiquitous): Phase 0.08로 갱신된 `.moai/project/db/**` 문서는 sync의 기존 커밋(`manager-git`)에 함께 포함된다 — 별도 커밋 없음.

### R3 — Config 시맨틱 전이

- **REQ-RELOC-CONFIG-001** (Ubiquitous): `.moai/config/sections/db.yaml`의 `auto_sync` 키는 의미가 "PostToolUse 훅 발동 여부"에서 "sync phase에서 자동 refresh 여부"로 변경된다. 키 이름·기본값(true)·YAML 구조는 불변.
- **REQ-RELOC-CONFIG-002** (Ubiquitous): `internal/template/templates/.moai/config/sections/db.yaml`의 인라인 주석(있다면)이 새 시맨틱을 반영한다.

### R4 — 하위 호환성 보존

- **REQ-RELOC-COMPAT-001** (Ubiquitous): `moai hook db-schema-sync --file <path>` CLI 서브커맨드는 계속 작동한다(Go 내부 구현 불변). 수동 호출과 향후 `/moai db refresh` 내부 재사용을 위해 보존.
- **REQ-RELOC-COMPAT-002** (Ubiquitous): 기존 프로젝트(`.moai/project/db/`를 가진)는 `moai update`로 마이그레이션 시 훅만 제거되며 데이터(`schema.md`, `erd.mmd` 등)는 건드리지 않는다.
- **REQ-RELOC-COMPAT-003** (Ubiquitous): `internal/hook/dbsync/` Go 패키지의 공개 API(`HandleDBSchemaSync`, `CheckDebounce`, `MatchesMigrationPattern`, `IsExcluded`, `BuildProposal`)는 시그니처 불변.

## Acceptance Criteria

- **AC-1**: `grep -n 'handle-db-schema-change' internal/template/templates/.claude/settings.json.tmpl` 결과 0 라인.
- **AC-2**: `internal/template/templates/.claude/hooks/moai/handle-db-schema-change.sh` 파일 부재(`test ! -f`).
- **AC-3**: `grep -n '### Phase 0.08' internal/template/templates/.claude/skills/moai/workflows/sync.md .claude/skills/moai/workflows/sync.md`가 양쪽에서 정확히 1 라인씩 출력.
- **AC-4**: `moai hook db-schema-sync --help` 여전히 동작 (backward compat verify).
- **AC-5**: `go test -race ./...` 전체 green (dbsync 패키지 포함 — Go API 불변 증명).
- **AC-6**: `make build` 성공, `internal/template/embedded.go` 재생성 확인.
- **AC-7**: `grep -rn 'db-schema-change\|PostToolUse.*handle-db-schema' internal/template/templates/` 결과 0 라인 (훅 레지스트레이션 흔적 전수 제거).
- **AC-8**: `internal/hook/dbsync/db_schema_sync.go`의 exported 함수 5개(`HandleDBSchemaSync`, `BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce`) 시그니처 불변 — `git diff HEAD`로 확인.

## Target Files

최대 5개:

1. `internal/template/templates/.claude/settings.json.tmpl` — PostToolUse `db-schema-change` 엔트리 제거
2. `internal/template/templates/.claude/hooks/moai/handle-db-schema-change.sh` — **파일 삭제**
3. `internal/template/templates/.claude/skills/moai/workflows/sync.md` — Phase 0.08 추가
4. `.claude/skills/moai/workflows/sync.md` — 로컬 동기화 (template 변경과 동일)
5. (선택) `internal/template/templates/.moai/config/sections/db.yaml` — `auto_sync` 주석 업데이트

## Exclusions

- Go 소스 수정(`internal/hook/dbsync/**`) 일체 없음. 공개 API 보존.
- `workflows/db.md` 변경 없음 (`/moai db refresh`·`/moai db verify`는 기존 그대로).
- `moai-domain-db-docs` 스킬 변경 없음.
- `.moai/project/db/` 템플릿 7개 파일 변경 없음.
- SPEC-DB-SYNC-001 자체는 status 유지 (이 SPEC이 cross-reference amendment 역할).

## Risks

- **R-1 (기존 프로젝트 회귀)**: 이미 `moai init`으로 설치되어 PostToolUse 훅이 `.claude/settings.json`에 등록된 프로젝트는 `moai update`를 통해서만 훅이 제거됨. **완화**: `moai update --force` 권장 안내 + CHANGELOG 명시.
- **R-2 (sync phase 시점 오차)**: 사용자가 `/moai sync` 전에 수동으로 마이그레이션 파일을 DB에 반영하면 실제 DB와 문서가 일시적으로 불일치. **완화**: `/moai db refresh`를 수동으로 언제든 호출 가능 — 긴급 시 escape hatch 존재.
- **R-3 (CLI 고아 함수)**: `moai hook db-schema-sync`가 이제 훅에서 호출되지 않으므로 내부 테스트 외 실사용 감소. **완화**: `/moai db refresh` 구현에서 재사용 가능 — 죽은 코드 아님.

## Implementation Notes

v1.0.0 구현 절차:
1. `handle-db-schema-change.sh` 템플릿 파일 삭제 + `settings.json.tmpl`의 PostToolUse `Write|Edit` 번째 엔트리(db-schema-change 블록) 제거
2. `sync.md` (template + local) Phase 0의 gate 이후 위치에 Phase 0.08 추가
3. `make build`로 embedded.go 재생성
4. `go test -race ./...` 전체 green 확인
5. 커밋 + push
