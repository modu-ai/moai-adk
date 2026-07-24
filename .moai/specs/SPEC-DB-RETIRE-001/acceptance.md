---
id: SPEC-DB-RETIRE-001
title: "DB 문서화 서브시스템 전면 제거 — 인수 기준"
version: "0.1.0"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
---

# SPEC-DB-RETIRE-001 — 인수 기준 (acceptance.md)

모든 AC는 기계 검증 가능(grep / 파일 존재 / 빌드·테스트 종료코드). run-phase 실행 후 아래 명령을 실측해 verbatim 출력을 §E.2 증거로 인용한다. 명령의 grep 스코프 제외는 §D.2 참조.

## §D. 인수 기준 매트릭스 (AC ↔ REQ)

| AC | REQ | 검증 명령 / 조건 | 기대 |
|----|-----|------------------|------|
| AC-DBR-001 | REQ-DBR-001 | `go run ./cmd/moai hook --help` (또는 `hookCmd.Commands()` 열거) 에 `db-schema-sync` 부재 | 미표시 |
| AC-DBR-002 | REQ-DBR-002 | `test -d internal/hook/dbsync && echo PRESENT || echo ABSENT` | `ABSENT` |
| AC-DBR-003 | REQ-DBR-003 | `grep -n "dbsync\|runDBSchemaSync\|loadMigrationPatterns\|defaultMigrationPatterns\|func splitLines\|func trimSpace" internal/cli/hook.go` | 0건 |
| AC-DBR-004 | REQ-DBR-004 | `grep -n "db-schema-sync" internal/cli/hook_e2e_test.go` | 0건 |
| AC-DBR-005 | REQ-DBR-005 | (comment scope) `grep -n "loadMigrationPatterns" internal/defs/dirs.go` = 0 **AND** dirs.go 주석에 db.yaml=live-config-read-by-loadMigrationPatterns 서술 부재. **주의**: bare `db.yaml` 은 grep=0 대상 아님 — REQ-DBR-022가 `Path: ".moai/config/sections/db.yaml"` DeprecatedPaths 엔트리를 추가하므로 dirs.go에 db.yaml 리터럴이 **정당하게** 잔존. count 39→40 검증은 AC-DBR-023 소관 | 통과 |
| AC-DBR-006 | REQ-DBR-006 | `grep -n "db\.yaml\|db-schema-sync" internal/cli/v2_detection.go` | 0건 |
| AC-DBR-007 | REQ-DBR-007 | `grep -n '"db":' internal/config/audit_registry.go` = 0 **AND** `grep -n '"db",' internal/config/audit_loader_completeness_test.go` = 0 (acknowledgedUnloadedSections) **AND** `go test ./internal/config/... -run 'TestAuditParity\|TestAuditLoaderCompleteness'` 종료 0 | 통과 |
| AC-DBR-008 | REQ-DBR-008 | `grep -n "internal/hook/dbsync" internal/template/internal_content_leak_test.go` = 0 **AND** 교체된 경로가 실존(`test -f <new-path>`) | 통과 |
| AC-DBR-009 | REQ-DBR-009 | `grep -n "db-schema-sync\|internal/hook/dbsync" internal/template/settings_test.go` | 0건 |
| AC-DBR-010 | REQ-DBR-010 | `test -d internal/template/templates/.moai/project/db && echo PRESENT || echo ABSENT` | `ABSENT` |
| AC-DBR-011 | REQ-DBR-011 | `test -f internal/template/templates/.moai/config/sections/db.yaml && echo PRESENT || echo ABSENT` | `ABSENT` |
| AC-DBR-012 | REQ-DBR-012 | `make build` 후 `go run ./cmd/moai init /tmp/dbr-check && test -e /tmp/dbr-check/.moai/project/db` | db 스캐폴드 미배포(`ABSENT`) |
| AC-DBR-013 | REQ-DBR-013 | `grep -rn "db-schema-sync\|db\.enabled\|detected_db\|DB Detection" .claude/skills/moai/workflows/project/doc-generation.md internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md` | 0건(양쪽) |
| AC-DBR-014 | REQ-DBR-014 | `grep -rn "db-schema-sync\|db\.enabled\|DB Schema Doc" .claude/skills/moai/workflows/sync/quality-gates-context.md internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-context.md` | 0건(양쪽) |
| AC-DBR-015 | REQ-DBR-015 | `diff -q <local> <template>` (두 지침 파일) = IDENTICAL **AND** `grep -rn "SPEC-DB-RETIRE-001" internal/template/templates/.claude/skills/moai/workflows/` = 0 | 통과 |
| AC-DBR-016 | REQ-DBR-016 | `git ls-files .moai/project/db/ .moai/config/sections/db.yaml` | 0건(untracked) |
| AC-DBR-017 | REQ-DBR-017 | (D4 narrowed) `grep -inE "db\.yaml\|\.moai/project/db\|db-schema-sync" CHANGELOG.md` 에 (a) `db.yaml` update 자동 삭제 + (b) `.moai/project/db/` 수동 삭제 두 문구 존재 **AND** `.moai/project` preserve-root 로직 무변경 | 두 문구 존재 |
| AC-DBR-018 | REQ-DBR-018 | `go build ./...` | 종료 0 |
| AC-DBR-019 | REQ-DBR-019 | `go test ./...` | 종료 0 |
| AC-DBR-020 | REQ-DBR-020 | `make build` | clean(종료 0) |
| AC-DBR-021 | REQ-DBR-021 | (사전) `grep -rn '"github.com/modu-ai/moai-adk/internal/hook/dbsync"' internal/ cmd/ pkg/` | `internal/cli/hook.go` 1곳만(삭제 전) |
| AC-DBR-022 | REQ-DBR-001..016 (closure) | (D1 fixed) `grep -rn "db-schema-sync\|dbsync\|db\.enabled" internal/ cmd/ \| grep -v "internal/settings/"` | 0건 (post-removal) |
| AC-DBR-023 | REQ-DBR-022 | (a) `grep -n '\.moai/config/sections/db\.yaml' internal/defs/dirs.go` 에 `DeprecatedPaths` 엔트리 존재; (b) clean-reinstall 테스트가 seed된 `.moai/config/sections/db.yaml` 제거 확인(`go test ./internal/cli/... -run 'TestUpdate\|CleanReinstall'` 종료 0, 신규/확장 케이스); (c) `go test ./internal/defs/... -run TestDeprecatedPaths` 종료 0 (count 39→40 원자 갱신) | 통과 |
| AC-DBR-024 | REQ-DBR-022 (collision guard) | `go test ./internal/cli/... -run TestDeprecatedPaths_NoTemplateCollision` 종료 0 (템플릿 db.yaml 삭제 후 DeprecatedPaths ∩ 템플릿 = ∅) | 통과 |

**요구사항→AC 커버리지: 22/22 REQ 커버(각 REQ ≥1 AC). 총 24 AC. (REQ-DBR-022 → AC-DBR-023+024; REQ-DBR-007 → AC-DBR-007 [양쪽 audit map])**

## §D.2 grep 스코프 제외 (오탐 방지, D1 반영)

closure grep(AC-DBR-022)은 `internal/settings/` **전체**를 제외한다(종전 `internal/settings/testdata/`만 제외 → D1: `schema_sections_test.go:292 "db.enabled"` false-fail). 제외 대상:
- `internal/settings/testdata/sections/db.yaml` — settings 테스트 픽스처(§C 보존).
- `internal/settings/schema_sections_test.go:292` — `"db.enabled", "db.orm", ...` OOS settings key-path 테스트(보존). 실측 확인: `internal/settings/` 안의 db 매칭은 이 두 곳뿐, 둘 다 제거 대상 아님.
- `.moai/specs/SPEC-DB-RETIRE-001/` — 이 SPEC 문서 자체.
- **패턴 불일치라 매칭 안 됨(제거 대상 아님)**: `backend-db`(Phase 11 MCP), "database schema changes" / "migrations"(Phase 3 배포 준비), `PhaseDB = "db"`(session, D5 OOS) — 세 패턴(`db-schema-sync` / `dbsync` / `db.enabled`) 어디에도 매칭 안 됨(실측 확인).

## §D.3 Given-When-Then 시나리오

### 시나리오 1 — dbsync 서브커맨드 휴면 제거 (정상 경로)
- **Given** dbsync 패키지와 `db-schema-sync` CLI 등록이 존재하는 트리에서,
- **When** run-phase가 dbsync 패키지 삭제 + hook.go import/핸들러/로더/죽은 헬퍼 제거 + `make build` 를 수행하면,
- **Then** `moai hook --help` 에 `db-schema-sync` 가 사라지고, `go build ./...` / `go test ./...` 가 종료 0이며, `grep -rn "dbsync" internal/ cmd/` 가 0건이다.

### 시나리오 2 — 지침 미러 byte-parity 보존 (경계)
- **Given** `doc-generation.md` / `quality-gates-context.md` 가 로컬-템플릿 byte-identical 인 상태에서,
- **When** DB Detection phase / DB Schema Doc Check phase 를 양쪽에 동일 편집으로 제거하면,
- **Then** `diff -q` 가 IDENTICAL 을 반환하고, 템플릿에 `SPEC-DB-RETIRE-001` 인용이 0건이다.

### 시나리오 3 — 무관 기능 오제거 방지 (엣지)
- **Given** Phase 11 MCP `backend-db` / Phase 3 deployment migration check / settings testdata `db.yaml` 이 존재하는 상태에서,
- **When** db 문서화 서브시스템만 제거하면,
- **Then** 위 3개는 그대로 존재하고 관련 테스트가 통과한다(`go test ./internal/settings/...` 종료 0).

### 시나리오 4 — audit parity 순서 의존 (엣지)
- **Given** 로컬 `db.yaml` 과 `audit_registry.go` "db" 예외가 공존하는 상태에서,
- **When** 두 참조를 함께 제거하면(순서 무관하되 최종 상태 일치),
- **Then** `TestAuditParity` 가 orphan-yaml/orphan-exception 오류 없이 통과한다. (한쪽만 제거해 db.yaml 이 disk 잔존 + 예외 삭제 시 orphan 실패 — 회피)

### 시나리오 5 — db.yaml 능동 삭제 (clarification RESOLVED, 정상 경로)
- **Given** 템플릿 `db.yaml` 이 삭제되고(REQ-011) `.moai/config/sections/db.yaml` 이 `DeprecatedPaths` 에 등록된 상태에서,
- **When** 기존 배포 사용자 프로젝트(orphaned `db.yaml` 보유)에서 `moai update`(clean-reinstall)를 실행하면,
- **Then** `scanDeprecatedPaths` 가 `.moai/config/sections/db.yaml` 을 제거하고, 충돌 가드(`TestDeprecatedPaths_NoTemplateCollision`)는 통과한다(템플릿이 더 이상 db.yaml 을 배포하지 않으므로 ∩ = ∅). **주의**: `.moai/project/db/` 는 preserve 루트라 update 가 삭제하지 않는다(시나리오 3 분리 확인).

## §D.4 엣지 케이스

- 병렬 세션이 동일 파일을 편집 중 → run-phase pre-spawn sync 점검(오케스트레이터). plan-phase는 SPEC 디렉터리 밖 무접촉.
- `make build` 미실행 시 embed FS 에 구 db 자산 잔존 → AC-DBR-012/020 로 포착.
- `splitLines`/`trimSpace` 미삭제 → `go build` 통과하나 golangci-lint `unused` 실패 → 품질 게이트에서 포착.

## §D.5 Definition of Done

- [ ] AC-DBR-001..024 전부 PASS (verbatim 명령 출력 인용).
- [ ] `go build ./...` 0 / `go test ./...` 0 / `make build` clean.
- [ ] 두 지침 파일 로컬-템플릿 byte-parity(`diff -q` IDENTICAL).
- [ ] 템플릿에 `SPEC-DB-RETIRE-001` / 내부 토큰 인용 0건(§25). (`internal/defs/dirs.go` DeprecatedPaths 엔트리는 템플릿 아님 → SPEC-ID 인용 허용.)
- [ ] `.moai/config/sections/db.yaml` DeprecatedPaths 등록 + clean-reinstall 제거 검증(AC-023) + collision guard 통과(AC-024).
- [ ] out-of-scope 5항목(MCP backend-db, deployment migration check, settings testdata db.yaml, `.moai/project/db/` 문서 preserve-root 잔존, PhaseDB) 무변경.
- [ ] CHANGELOG 두 문구 존재: (a) `db.yaml` update 자동 삭제(수동 불필요), (b) `.moai/project/db/` 수동 삭제 안내.
- [ ] 미해소 clarification 마커(콜론-정규형) 0건(RESOLVED → REQ-DBR-022 능동 삭제).

## §D.6 품질 게이트 (TRUST 5)

- **Tested**: 삭제 후 전체 스위트 green — 제거가 무관 패키지를 깨지 않음을 증명(제거 안전성).
- **Readable**: 참조 해제는 정밀 편집으로 주변 문맥 보존.
- **Unified**: `gofmt` / `golangci-lint` 통과(죽은 헬퍼 제거 포함).
- **Secured**: 해당 없음(순수 제거, 새 입력 경로 없음).
- **Trackable**: 커밋 제목 `feat(SPEC-DB-RETIRE-001): ...`(plan-phase는 feat 접두사), CHANGELOG 항목.
