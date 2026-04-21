---
id: SPEC-DB-SYNC-RELOC-001
document: spec-compact
version: 1.0.0
source: spec.md (Requirements + AC + Exclusions)
---

# SPEC-DB-SYNC-RELOC-001: Compact View

## Requirements (EARS)

### R1 — 훅 제거
- REQ-RELOC-HOOK-001: `settings.json.tmpl`에 `handle-db-schema-change` 관련 엔트리 0건
- REQ-RELOC-HOOK-002: `handle-db-schema-change.sh` 템플릿 파일 부재
- REQ-RELOC-HOOK-003: `moai init` 결과물에도 흔적 없음

### R2 — sync Phase 0.08
- REQ-RELOC-SYNC-001: `sync.md` (template + local)에 `### Phase 0.08: DB Schema Doc Check` 각 1회
- REQ-RELOC-SYNC-002: Phase 0.08은 db.yaml → git diff → migration_patterns 필터 → auto_sync 분기
- REQ-RELOC-SYNC-003: 갱신 문서는 sync 기존 커밋에 포함

### R3 — config 시맨틱
- REQ-RELOC-CONFIG-001: `db.auto_sync` 의미만 변경 (훅→sync phase 발동), 키/기본값 불변
- REQ-RELOC-CONFIG-002: db.yaml 인라인 주석 업데이트

### R4 — 하위 호환
- REQ-RELOC-COMPAT-001: `moai hook db-schema-sync` CLI 유지
- REQ-RELOC-COMPAT-002: 기존 프로젝트 `.moai/project/db/` 데이터 보존
- REQ-RELOC-COMPAT-003: Go 공개 API 5개 시그니처 불변

## Acceptance Criteria

- AC-1: `grep 'handle-db-schema-change' settings.json.tmpl` = 0 라인
- AC-2: `handle-db-schema-change.sh` 파일 부재
- AC-3: `### Phase 0.08` template + local 각 1회
- AC-4: `moai hook db-schema-sync --help` 정상 동작
- AC-5: `go test -race ./internal/hook/dbsync/...` PASS
- AC-6: `make build` 성공
- AC-7: `grep 'db-schema-change' templates/` = 0 라인
- AC-8: Go 공개 API 5개 시그니처 불변

## Exclusions

- Go 소스 수정 없음
- `/moai db refresh`·`/moai db verify` 워크플로우 변경 없음
- `moai-domain-db-docs` 스킬 변경 없음
- `.moai/project/db/` 템플릿 7개 변경 없음
- SPEC-DB-SYNC-001 status 유지
