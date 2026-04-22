---
id: SPEC-DB-SYNC-RELOC-001
document: plan
version: 1.0.0
created_at: 2026-04-21
updated_at: 2026-04-21
---

# SPEC-DB-SYNC-RELOC-001: Implementation Plan

## Technical Approach

- 문서/설정 전용 이관. Go 소스 코드 변경 없음.
- 훅 레지스트레이션(`settings.json.tmpl`의 PostToolUse 블록) 제거 → sync 워크플로우(`sync.md`) Phase 0.08 추가.
- Template source → 로컬 `.claude/` 동기화 (Template-First 규칙).
- Go 공개 API 불변이므로 `internal/hook/dbsync/` 테스트는 그대로 green.

## Milestone Decomposition

### M1 — 훅 번들 제거 (High)

- `internal/template/templates/.claude/settings.json.tmpl`의 PostToolUse `Write|Edit` db-schema-change 엔트리(플랫폼 분기 포함 약 10줄) 제거
- `internal/template/templates/.claude/hooks/moai/handle-db-schema-change.sh` 파일 삭제

### M2 — sync Phase 0.08 추가 (High)

- `internal/template/templates/.claude/skills/moai/workflows/sync.md`에 `### Phase 0.08: DB Schema Doc Check` 섹션 추가. Phase 0(Gate) 완료 직후, Phase 0.1 진입 전에 위치.
- 섹션 내용:
  - Purpose: DB 문서 자동 갱신 조건 평가
  - Condition: `db.yaml` 존재 + `enabled: true` + `auto_sync: true` + git diff에 migration_patterns 매칭 파일 ≥ 1
  - Action: 조건 만족 시 `moai-domain-db-docs` 스킬 refresh 호출 / 불만족 시 advisory 또는 skip
  - Output: sync 리포트에 반영, 같은 commit에 포함
- `.claude/skills/moai/workflows/sync.md` (로컬) 동일 내용 동기화.

### M3 — config 주석 및 CHANGELOG (Medium)

- `internal/template/templates/.moai/config/sections/db.yaml`의 `auto_sync` 키 주석에 새 시맨틱 반영(있을 경우)
- `CHANGELOG.md`에 Unreleased 항목 추가

### M4 — 검증 및 커밋 (High)

- `go test -race ./...` green
- `make build` 성공
- grep 기반 AC 전수 통과
- 단일 `refactor(hook)` 또는 `feat(sync)` 커밋 (Conventional Commits)

## Risks and Mitigations

- R-1 기존 프로젝트 회귀: `moai update` 사용 안내를 CHANGELOG에 명시.
- R-2 sync phase 시점 오차: `/moai db refresh` 수동 호출이 escape hatch로 존재 — 문서화.
- R-3 CLI 고아 함수: `moai hook db-schema-sync`는 향후 `/moai db refresh`가 내부 재사용할 수 있도록 유지.

## Commit Strategy

단일 커밋 권장:

```
refactor(hook/dbsync): SPEC-DB-SYNC-RELOC-001 — PostToolUse 훅을 /moai sync Phase 0.08로 이관

PostToolUse 훅은 모든 Write/Edit 이벤트에서 bash + moai 프로세스를 spawn하여
~30-60ms/event 비용을 발생시킴. 일반 세션 누적 5-30초 순손실.
DB 문서는 milestone/PR boundary에서 한 번 최신화로 충분한 파생 문서이므로
/moai sync Phase 0.08로 이관하여 per-edit 비용 0화.

변경:
- settings.json.tmpl: PostToolUse db-schema-change 엔트리 제거
- handle-db-schema-change.sh: 파일 삭제
- sync.md (template + local): Phase 0.08 DB Schema Doc Check 추가
- CHANGELOG.md: Unreleased 항목

보존:
- internal/hook/dbsync/: Go 공개 API 불변
- moai hook db-schema-sync CLI: 수동 호출 및 향후 재사용 가능
- /moai db refresh, moai-domain-db-docs, .moai/project/db/ 템플릿: 불변

🗿 MoAI <email@mo.ai.kr>
```
