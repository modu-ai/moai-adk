---
id: SPEC-DB-SYNC-RELOC-001
document: acceptance
version: 1.0.0
created_at: 2026-04-21
updated_at: 2026-04-21
---

# SPEC-DB-SYNC-RELOC-001: Acceptance Criteria

## Definition of Done

1. spec.md의 R1 ~ R4 (9 REQ) 전수 구현
2. AC-1 ~ AC-8 전수 통과
3. `go test -race ./...` 전체 green
4. `make build` 성공

---

## AC-1: settings.json.tmpl 훅 엔트리 0건

**When**: `grep -n 'handle-db-schema-change' internal/template/templates/.claude/settings.json.tmpl`

**Then**: 0 라인 출력

**AC Coverage**: REQ-RELOC-HOOK-001

## AC-2: 훅 래퍼 스크립트 부재

**When**: `test -f internal/template/templates/.claude/hooks/moai/handle-db-schema-change.sh`

**Then**: exit code 1 (파일 부재)

**AC Coverage**: REQ-RELOC-HOOK-002

## AC-3: Phase 0.08 섹션 존재

**When**: `grep -n '### Phase 0.08' internal/template/templates/.claude/skills/moai/workflows/sync.md .claude/skills/moai/workflows/sync.md`

**Then**: template 1 라인 + local 1 라인 = 총 2 라인 (각 파일당 1회)

**AC Coverage**: REQ-RELOC-SYNC-001

## AC-4: CLI backward compat

**When**: `./bin/moai hook db-schema-sync --help`

**Then**: exit 0, usage 출력 정상

**AC Coverage**: REQ-RELOC-COMPAT-001

## AC-5: Go 테스트 회귀 없음

**When**: `go test -race -count=1 ./internal/hook/dbsync/...`

**Then**: PASS (기존 12+ 테스트 + H1~H5 신규 테스트 유지)

**AC Coverage**: REQ-RELOC-COMPAT-003

## AC-6: 빌드 성공

**When**: `make build`

**Then**: exit 0, `bin/moai` 생성, `internal/template/embedded.go` 최신 반영

**AC Coverage**: REQ-RELOC-HOOK-003 (간접)

## AC-7: 훅 흔적 전수 제거

**When**: `grep -rn 'db-schema-change' internal/template/templates/`

**Then**: 0 라인 (훅 래퍼 파일·settings 엔트리·문서 내 언급 전수 제거)

**AC Coverage**: REQ-RELOC-HOOK-001, REQ-RELOC-HOOK-002

## AC-8: Go 공개 API 시그니처 불변

**When**: `git diff HEAD -- internal/hook/dbsync/db_schema_sync.go` 검토

**Then**: exported 5개 함수(`HandleDBSchemaSync`, `BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce`) 시그니처 라인에 변경 없음

**AC Coverage**: REQ-RELOC-COMPAT-003
