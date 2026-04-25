---
spec_id: SPEC-DB-SYNC-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  REVERSE-DOC (PARTIAL): 본 acceptance.md는 SPEC 작성 이후 Go hook handler 구현이
  완료된 상태에서 SDD 아티팩트 공백을 메우기 위해 역공학 방식으로 생성되었다.
  plan-auditor 2026-04-24 감사에서 `acceptance.md` 미존재가 지적되어 backfill됨.

  PARTIAL 상태:
  - [완료] internal/hook/dbsync/db_schema_sync.go — Go hook handler (debounce, path
    guard, proposal.json 생성)
  - [완료] internal/cli/hook.go L90-99 — `moai hook db-schema-sync --file` subcommand
  - [완료] .claude/skills/moai-domain-db-docs/SKILL.md — 스킬 frontmatter + 본문
  - [미완료] .claude/hooks/moai/handle-db-schema-change.sh — bash wrapper 미생성
  - [미완료] .claude/settings.json.tmpl PostToolUse matcher — db-schema-sync 항목 없음
  - [미완료] /moai db refresh, /moai db verify 실제 drift 파싱 로직 — 스킬 위임 skeleton 수준

  따라서 AC-01, AC-02 는 "PARTIAL — pending implementation" 으로 표기하고,
  AC-03 ~ AC-10 중 hook handler 레벨은 테스트로 검증, UI/orchestration 레벨은
  수동 세션 검증이 필요함.
---

# Acceptance Criteria — SPEC-DB-SYNC-001

> PostToolUse 훅 + `moai-domain-db-docs` 스킬 + `/moai db verify`/`refresh` 의 실제
> 구현 동작으로부터 역도출된 검증 기준이다. PARTIAL 구현이므로 일부 AC는 hook handler
> Go 레벨에서만 검증되고, UI/orchestration 레벨은 미구현이다.

## Traceability Matrix

| REQ ID | AC ID | 검증 수단 (파일:라인 또는 테스트) |
|---|---|---|
| REQ-001 (settings.json.tmpl matcher) | AC-001 | **PARTIAL — pending**: `settings.json.tmpl` 에 `handle-db-schema-change.sh` 매처 항목 부재 |
| REQ-002 (stdin JSON 읽기) | AC-002 | **PARTIAL — pending**: bash wrapper `handle-db-schema-change.sh` 부재 |
| REQ-003, REQ-004 (excluded/non-migration exit 0) | AC-003 | `db_schema_sync.go:122-132` + `TestHandleDBSchemaSync_ExcludedPattern`, `TestHandleDBSchemaSync_NonMigrationFile` |
| REQ-005 (moai hook db-schema-sync 호출) | AC-004 | `internal/cli/hook.go:90-99` cobra subcommand |
| REQ-006, REQ-007 (debounce) | AC-005 | `db_schema_sync.go:309-444` + `TestDebounce_*` (4건) + `TestHandleDBSchemaSync_Debounced` |
| REQ-008, REQ-009 (parse + proposal.json) | AC-006 | `db_schema_sync.go:150-178` + `TestHandleDBSchemaSync_WritesProposal` |
| REQ-010 (decision=ask-user stdout) | AC-007 | `db_schema_sync.go:41-42, 180-182` + `internal/cli/hook.go:274-281` |
| REQ-011 (parse 실패 non-blocking) | AC-008 | `db_schema_sync.go:152-157` + `TestHandleDBSchemaSync_NonBlockingOnParseError` |
| REQ-012~015 (user approval 3-option) | AC-009 | **PARTIAL — orchestrator 레벨, 자동 테스트 불가** |
| REQ-016~018 (moai-domain-db-docs 갱신) | AC-010 | `.claude/skills/moai-domain-db-docs/SKILL.md` frontmatter + 본문 |
| REQ-019 (재귀 차단 이중 보호) | AC-011 | `db_schema_sync.go:20-25` DefaultExcludedPatterns + SKILL.md 내 excluded 섹션 |
| REQ-020~022 (/moai db verify drift) | AC-012 | **PARTIAL — pending**: 파서 모듈(`internal/db/parser/`) 미구현 |
| REQ-023, REQ-024 (/moai db refresh full rebuild) | AC-013 | **PARTIAL — pending**: moai-domain-db-docs Phase D 미구현 |

---

## AC-001 — PostToolUse 매처 등록 (PARTIAL)

**Given** `moai init` 또는 `moai update` 가 `settings.json.tmpl` 을 배포한 상태
**When** 렌더링된 `.claude/settings.json` 을 검사할 때
**Then** PostToolUse 매처 엔트리가 tool types `Write` 과 `Edit` 을 참조하고 command path가 `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 이며 `timeout: 30000` 을 포함해야 한다.

**Status**: **PARTIAL — pending implementation**
**Verification Gap**:
- `grep -c "handle-db-schema-change" internal/template/templates/.claude/settings.json.tmpl` → 0 (매처 미존재)
- 현재 PostToolUse 매처는 `handle-post-tool.sh` 하나뿐 (settings.json.tmpl:67-82)
**Remediation**: 별도 HARDEN SPEC(`SPEC-DB-SYNC-HARDEN-001` 후속)에서 처리 예정.

---

## AC-002 — Bash wrapper stdin JSON 소비 (PARTIAL)

**Given** PostToolUse 매처가 스크립트를 트리거한 상태
**When** bash wrapper `handle-db-schema-change.sh` 가 실행될 때
**Then** wrapper는 stdin JSON을 읽어 `.tool_input.file_path` 필드를 추출하고, `moai hook db-schema-sync --file <path>` 호출 후 exit 0 으로 종료해야 한다. 스크립트 크기는 30 라인 미만이고 실행 권한(`+x`)을 가진다.

**Status**: **PARTIAL — pending implementation**
**Verification Gap**:
- `ls internal/template/templates/.claude/hooks/moai/handle-db-schema-change.sh` → 파일 없음
- 기존 hook wrapper 28개는 존재하나 db-schema-change.sh 만 미구현
**Remediation**: 후속 SPEC에서 bash skeleton (스펙 본문 L169-L176) 을 배포.

---

## AC-003 — 제외 패턴 및 비-마이그레이션 파일 silent skip

**Given** PostToolUse 훅이 `moai hook db-schema-sync --file <path>` 로 호출된 상태
**When** 파일 경로가 `.moai/project/db/**`, `.moai/cache/**`, `.moai/logs/**` 중 하나거나 `migration_patterns` 에 매치되지 않을 때
**Then** handler는 exit 0 을 반환하고 `proposal.json` 을 생성하지 않아야 한다.
**Verification**:
- `internal/hook/dbsync/db_schema_sync.go:20-25` DefaultExcludedPatterns (3 패턴)
- `internal/hook/dbsync/db_schema_sync.go:122-132` IsExcluded / MatchesMigrationPattern 체크
- 테스트:
  - `go test ./internal/hook/dbsync/ -run TestHandleDBSchemaSync_ExcludedPattern`
  - `go test ./internal/hook/dbsync/ -run TestHandleDBSchemaSync_NonMigrationFile`
  - `go test ./internal/hook/dbsync/ -run TestIsExcluded`

---

## AC-004 — CLI subcommand `moai hook db-schema-sync`

**Given** `moai` 바이너리가 설치된 상태
**When** `moai hook db-schema-sync --file <path>` 이 호출될 때
**Then** cobra subcommand가 `internal/cli/hook.go` 의 `runDBSchemaSync` 함수로 위임하고, stdout에 decision JSON을 출력한 뒤 exit 0 으로 종료한다.
**Verification**:
- `internal/cli/hook.go:90-99` `dbSchemaSyncCmd := &cobra.Command{Use: "db-schema-sync", RunE: runDBSchemaSync}`
- `internal/cli/hook.go:248-285` runDBSchemaSync 구현
- 테스트: `internal/cli/hook_test.go:63` (subcommand 수 31개로 확장 검증)
- 테스트: `internal/cli/hook_e2e_test.go:345` (`db-schema-sync` 명시 허용 목록)

---

## AC-005 — Debounce 10초 윈도우

**Given** 동일 `file_path` 가 `DebounceWindow` (기본 10초) 이내에 두 번 보고된 상태
**When** 두 번째 호출이 발생할 때
**Then** handler는 `proposal.json` 생성 없이 `Decision: DecisionDebounced` 를 반환하고 exit 0으로 종료한다. 상태 파일 `.moai/cache/db-sync/last-seen.json` 은 원자적으로 갱신된다 (temp file + os.Rename).
**Verification**:
- `db_schema_sync.go:134-148` HandleDBSchemaSync 내 debounce 체크
- `db_schema_sync.go:345-444` CheckDebounce / checkDebounceWithLog (O_EXCL 락파일 + temp file + Rename)
- `db_schema_sync.go:86-89` DebounceState JSON 구조 (`file_path`, `timestamp`)
- 테스트:
  - `go test ./internal/hook/dbsync/ -run TestDebounce_FirstCallNotDebounced`
  - `go test ./internal/hook/dbsync/ -run TestDebounce_SameFileWithinWindow`
  - `go test ./internal/hook/dbsync/ -run TestDebounce_DifferentFileNotDebounced`
  - `go test ./internal/hook/dbsync/ -run TestDebounce_ExpiredWindowNotDebounced`
  - `go test ./internal/hook/dbsync/ -run TestHandleDBSchemaSync_Debounced`
  - `go test ./internal/hook/dbsync/ -run TestCheckDebounceConcurrency` (동시성 winner-loser 검증)

---

## AC-006 — proposal.json 생성

**Given** 파일이 migration pattern에 매치되고 debounce 를 통과한 상태
**When** handler가 파일을 파싱하는 데 성공할 때
**Then** `.moai/cache/db-sync/proposal.json` 에 `{"file_path", "parsed_content", "decision": "ask-user", "timestamp"}` 4개 필드를 가진 JSON이 기록되어야 한다.
**Verification**:
- `db_schema_sync.go:78-84` Proposal 구조체 정의
- `db_schema_sync.go:159-178` WriteFile(cfg.ProposalFile, proposalJSON, 0o644)
- `db_schema_sync.go:198-205` BuildProposal — Decision = DecisionAskUser, Timestamp = RFC3339 UTC
- 테스트:
  - `go test ./internal/hook/dbsync/ -run TestHandleDBSchemaSync_WritesProposal`
  - `go test ./internal/hook/dbsync/ -run TestBuildProposal_ValidJSON`

---

## AC-007 — decision=ask-user stdout 시그널

**Given** proposal.json 이 기록된 상태
**When** handler가 결과를 반환할 때
**Then** CLI wrapper `runDBSchemaSync` 는 `{"decision": "ask-user"}` JSON 메시지를 stdout에 출력하여 orchestrator가 다음 턴에 AskUserQuestion을 제시하도록 시그널해야 한다.
**Verification**:
- `db_schema_sync.go:41-42` `const DecisionAskUser = "ask-user"`
- `db_schema_sync.go:182` `return Result{ExitCode: 0, Decision: DecisionAskUser}`
- `internal/cli/hook.go:274-281` Result를 stdout JSON으로 인코드

---

## AC-008 — 파싱 실패 non-blocking

**Given** 마이그레이션 파일이 손상되었거나 파싱 불가능한 상태
**When** `parseMigrationStub` 이 에러를 반환할 때
**Then** handler는 `.moai/logs/db-sync-errors.log` 에 에러를 기록하고 exit 0 (non-blocking) 으로 종료해야 한다. 사용자의 Write/Edit 흐름이 차단되지 않는다.
**Verification**:
- `db_schema_sync.go:152-157` parseErr 핸들링 — `logError(cfg.ErrorLogFile, ...)` + `return Result{ExitCode: 0, ...}`
- `db_schema_sync.go:488-503` logError 구현 (APPEND + CREATE 모드, RFC3339 타임스탬프)
- 테스트:
  - `go test ./internal/hook/dbsync/ -run TestHandleDBSchemaSync_NonBlockingOnParseError`
  - `go test ./internal/hook/dbsync/ -run TestHandleDBSchemaSync_OversizedFile` (1 MiB 크기 가드)
  - `go test ./internal/hook/dbsync/ -run TestParseMigrationStub_OversizedFile`

---

## AC-009 — 사용자 승인 3-옵션 플로우 (PARTIAL)

**Given** orchestrator 가 `decision: ask-user` 시그널을 감지한 상태
**When** 다음 턴이 시작될 때
**Then** orchestrator는 AskUserQuestion을 호출하여 3개 옵션(`Apply proposed schema update`, `Review diff first`, `Skip this time`) 을 제시해야 한다. `Apply` 선택 시 `moai-domain-db-docs` 스킬이 호출되어 schema.md / erd.mmd / migrations.md 를 갱신한다. `Skip` 선택 시 `proposal.json` 이 삭제된다.

**Status**: **PARTIAL — orchestrator layer, manual session verification required**
**Verification**:
- `.claude/skills/moai-domain-db-docs/SKILL.md` 에 승인 플로우 서술이 존재하나 자동화된 테스트는 없음
- AskUserQuestion 호출은 orchestrator (MoAI) 의 책임이며 Go 테스트 범위 밖
**Remediation**: 통합 테스트 또는 수동 Claude Code 세션으로 검증.

---

## AC-010 — moai-domain-db-docs 스킬 frontmatter

**Given** `moai init` 이 템플릿을 배포한 상태
**When** `.claude/skills/moai-domain-db-docs/SKILL.md` 를 읽을 때
**Then** frontmatter는 Skill Frontmatter Template(spec.md L142-L165) 과 일치해야 한다:
- 필수 최상위 필드: `name`, `description`, `license`, `compatibility`, `allowed-tools`, `user-invocable`
- `metadata.*` 키: `version`, `category`, `status`, `updated`, `tags`
- `triggers.*` 키: `keywords`, `agents`, `phases`

**Verification**:
- `.claude/skills/moai-domain-db-docs/SKILL.md:1-23` frontmatter 전체
- 필수 필드 모두 존재, `name: moai-domain-db-docs`, `user-invocable: false`, `allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TaskCreate, TaskUpdate`
- metadata: version=1.0.0, category=domain, status=active, updated=2026-04-20, tags=db,schema,migration,...
- triggers: keywords=[db docs, schema sync, ...], agents=[expert-backend], phases=[run, sync]
- SKILL.md 본문 243 lines

---

## AC-011 — 재귀 차단 이중 보호 (훅 + 스킬)

**Given** moai-domain-db-docs 스킬이 `.moai/project/db/schema.md` 등을 갱신한 상태
**When** 갱신 이벤트가 PostToolUse 훅으로 다시 전파될 때
**Then** `.moai/project/db/**` 패턴이 Excluded Patterns 에 포함되어 있어 훅 레벨(REQ-004) 에서 1차 차단되고, 스킬 레벨(REQ-019) 에서 2차 인식되어 재귀 호출이 발생하지 않아야 한다.
**Verification**:
- `db_schema_sync.go:20-25` DefaultExcludedPatterns 에 3개 패턴 모두 존재:
  - `.moai/project/db/**`
  - `.moai/cache/**`
  - `.moai/logs/**`
- `db_schema_sync.go:240-247` IsExcluded 함수 — `matchGlob` 기반 재귀 차단
- SKILL.md 본문에 `.moai/project/db/**` excluded 섹션 명시

---

## AC-012 — /moai db verify drift 검출 (PARTIAL)

**Given** `/moai db verify` 가 호출된 상태
**When** 현재 `schema.md` 를 마이그레이션 파일에서 기댓값과 비교할 때
**Then** drift 검출 시 exit 1 + unified diff 를 stdout에 출력, drift 부재 시 exit 0 + `"Schema documentation is in sync"` 를 출력해야 한다.

**Status**: **PARTIAL — pending full parser implementation**
**Verification Gap**:
- `workflows/db.md:207-230` Phase 7: Drift Detection 스켈레톤 존재
- 실제 마이그레이션 파일 파서 (`internal/db/parser/`) 는 **미구현** (본 SPEC의 OUT OF SCOPE, 후속 작업)
- Inline fallback (symmetric difference of table names from schema.md) 은 workflows/db.md:228 에 서술되나 자동 테스트는 없음
**Remediation**: `internal/db/parser/` 모듈 구현 후 통합 테스트 추가.

---

## AC-013 — /moai db refresh full rebuild (PARTIAL)

**Given** `/moai db refresh` 가 호출된 상태
**When** 사용자 승인(`Apply`) 을 받은 후
**Then** 스킬은 모든 마이그레이션 파일을 재스캔하고 schema.md / erd.mmd / migrations.md 3개 문서를 처음부터 재구성해야 한다.

**Status**: **PARTIAL — pending full parser + rebuild logic**
**Verification Gap**:
- `workflows/db.md:188-203` Phase 6: Regenerate 스켈레톤 — moai-domain-db-docs 에 위임
- moai-domain-db-docs Phase D1/D2/D3 (AskUserQuestion 확인 + 풀 스캔 + 재구성) 은 SKILL.md 본문에 서술되나 실제 파싱·렌더링 구현체 부재
**Remediation**: 파서 모듈 구현 후 통합 테스트 추가.

---

## Edge Cases and Defensive Behaviors

- **EC-01**: 경로 탈출 공격 (`migrations/../../../etc/passwd.sql`) → `db_schema_sync.go:108-115` filepath.Clean + prefix 체크로 거부. 검증: `TestHandleDBSchemaSync_PathTraversalRejected`.
- **EC-02**: 1 MiB 초과 마이그레이션 파일 → `parseMigrationStub` 의 `maxMigrationFileSize` 가드에 의해 ParsedContent=빈 문자열 + Truncated=true. 검증: `TestParseMigrationStub_OversizedFile`.
- **EC-03**: 손상된 `last-seen.json` → `isWithinDebounceWindow` 가 JSON decode 실패를 "no fresh state" 로 처리하여 새 윈도우 수립. 검증: `TestCheckDebounce_CorruptStateRecovery`.
- **EC-04**: state directory 쓰기 권한 없음 → `mkdir` 실패 시 logError + exit 0 (non-blocking). 검증: `TestHandleDBSchemaSync_ReadonlyStateDir`, `TestHandleDBSchemaSync_ReadonlyProposalDir`.
- **EC-05**: 동시에 두 프로세스가 동일 stateFile을 갱신 시도 → O_EXCL 락파일로 정확히 1명의 winner가 debounced=false 를 반환. 검증: `TestHandleDBSchemaSync_WinnerLoserOrdering`, `TestCheckDebounceConcurrency`.
- **EC-06**: Windows 경로 (`migrations\001.sql`) → `filepath.ToSlash` 정규화로 forward-slash glob 매칭 가능. 검증: `TestMatchGlob_TableDrivenEdges`.

---

## Definition of Done

| 항목 | 상태 | 근거 |
|---|---|---|
| Go hook handler (HandleDBSchemaSync) | [x] DONE | `internal/hook/dbsync/db_schema_sync.go` 504 LOC |
| CLI subcommand (`moai hook db-schema-sync`) | [x] DONE | `internal/cli/hook.go:90-99` |
| Debounce with lock file (atomic rename) | [x] DONE | `db_schema_sync.go:345-444` |
| Path traversal guard | [x] DONE | `db_schema_sync.go:108-115` |
| File size guard (1 MiB) | [x] DONE | `db_schema_sync.go:33, 469-479` |
| Test coverage | [x] DONE | 27 테스트 함수 (internal + e2e) |
| moai-domain-db-docs SKILL.md frontmatter | [x] DONE | 23 라인 frontmatter, 243 LOC 본문 |
| Bash wrapper (`handle-db-schema-change.sh`) | [ ] PARTIAL | **미구현** — 후속 SPEC |
| PostToolUse 매처 (settings.json.tmpl) | [ ] PARTIAL | **미구현** — 후속 SPEC |
| 실제 마이그레이션 파서 (internal/db/parser/) | [ ] PARTIAL | **미구현** — SPEC 자체 OUT OF SCOPE |
| /moai db verify drift 자동 테스트 | [ ] PARTIAL | 파서 구현 후 가능 |
| /moai db refresh full rebuild 자동 테스트 | [ ] PARTIAL | 파서 구현 후 가능 |

---

## Quality Gate Alignment

| Gate | Criterion | Evidence |
|---|---|---|
| Tested | Go hook handler 27개 테스트 통과 | `go test -race ./internal/hook/dbsync/...` |
| Readable | 공개 함수 godoc + @MX:NOTE/ANCHOR 완비 | `db_schema_sync.go:91-99, 185-205, 207-225, 227-246` |
| Unified | Config struct 단일 진입점, stub parser 분리 | `Config`, `Result`, `Proposal` 구조체 |
| Secured | Path traversal guard + O_EXCL 락파일 + 크기 가드 | EC-01, EC-02, EC-05 |
| Trackable | REQ 주석 (`REQ coverage:` 블록) + SPEC ID | `db_schema_sync.go:6` |

---

## Reverse-Doc Divergence Notes

1. **스펙은 완전하나 구현은 부분적**: spec.md 는 3개 레이어(훅 인프라 + 스킬 + 서브커맨드) 를 모두 정의하지만, 구현은 Go handler + 스킬 frontmatter 까지만 완료됨.
2. **훅 인프라 연결 누락**: `handle-db-schema-change.sh` 와 `settings.json.tmpl` PostToolUse 매처가 없어 실제 Claude Code 환경에서 훅이 발화되지 않음. Go handler는 존재하나 외부에서 호출될 경로가 없음.
3. **파서 부재로 인한 기능 공백**: `/moai db verify` 의 drift 계산과 `/moai db refresh` 의 전체 재구성은 마이그레이션 파싱 결과를 요구하나, 파서가 stub (`parseMigrationStub` 는 파일 내용을 그대로 반환) 상태.
4. **권고**: 후속 SPEC 에서 (a) bash wrapper 추가, (b) settings.json.tmpl 매처 추가, (c) Prisma/Alembic/Rails 파서 구현 3개를 완료하면 본 SPEC의 REQ-001, REQ-002, REQ-020, REQ-023 이 FULLY IMPLEMENTED 상태로 전환됨.
