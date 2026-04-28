# Acceptance Criteria: SPEC-V3R2-EXT-001 — Typed Memory Taxonomy

- SPEC: SPEC-V3R2-EXT-001
- Plan ref: `.moai/specs/SPEC-V3R2-EXT-001/plan.md`
- Tasks ref: `.moai/specs/SPEC-V3R2-EXT-001/tasks.md`
- 총 AC: 15 (spec.md §6 13개 + AC-01b/AC-04 분할로 추가) + Edge/DoD 확장
- Traceability: 모든 REQ(17개)에 최소 1개 AC 매핑. 역매핑은 §3 Traceability Matrix 참조.
- v1.1.0 변경: AC-01b(REQ-002 name/description 명시 커버), AC-04 를 AC-04a(automated regex)/AC-04b(optional human review)로 분할.

---

## 1. Given–When–Then Acceptance Scenarios

### AC-EXT001-01 — Missing type warning

- Maps: REQ-EXT001-001, REQ-EXT001-007
- **Given** 파일 `.claude/agent-memory/manager-spec/foo.md` 의 frontmatter에 `type` 키가 없다
- **When** PostToolUse 훅이 이 파일을 대상으로 실행된다
- **Then** stderr에 `MEMORY_MISSING_TYPE` 경고가 출력되고, 훅 exit code는 0 (non-blocking)
- Test path: `internal/hook/memo/taxonomy/audit_test.go::TestAuditFile_MissingType`, `internal/hook/post_tool_test.go::TestPostTool_MemoryMissingType`

### AC-EXT001-01b — Missing name/description warning

- Maps: REQ-EXT001-002
- **Given** 파일 `.claude/agent-memory/manager-spec/foo.md` 의 frontmatter에 `type`은 있지만 `name` 또는 `description` 키가 누락되어 있다 (세 가지 케이스: name 누락, description 누락, 둘 다 누락)
- **When** PostToolUse 훅이 해당 파일을 감사한다
- **Then** `MEMORY_MISSING_FRONTMATTER` 또는 `MEMORY_MISSING_TYPE`과 동등한 non-blocking warning이 발생하고, detail에 어떤 키가 누락되었는지(`name` / `description`) 명시된다
- Test path: `internal/hook/memo/taxonomy/taxonomy_test.go::TestParseFile_MissingName`, `TestParseFile_MissingDescription`, `TestParseFile_MissingBoth`, `internal/hook/memo/taxonomy/audit_test.go::TestAuditFile_MissingFrontmatterKeys`

### AC-EXT001-02 — Unknown type rejected by loader

- Maps: REQ-EXT001-001, REQ-EXT001-013
- **Given** 파일 frontmatter가 `type: unknown`
- **When** `memo.ParseFile` 또는 후속 SPEC-V3R2-EXT-002 loader가 호출된다
- **Then** `memo.ValidateType` 가 오류를 반환하고, 본 SPEC 범위에서는 `AuditFile`이 `MEMORY_MISSING_TYPE`에 준하는 오류 경로를 탄다
- Test path: `internal/hook/memo/taxonomy_test.go::TestValidateType_Unknown`, `TestParseFile_InvalidType`

### AC-EXT001-03 — MEMORY.md line cap overflow

- Maps: REQ-EXT001-003, REQ-EXT001-008
- **Given** `MEMORY.md` 파일이 250줄이다
- **When** PostToolUse 훅이 해당 파일 수정 이벤트를 받는다
- **Then** `MEMORY_INDEX_OVERFLOW` 경고가 출력되고 `lineCap=200`이 로그에 포함된다
- Test path: `internal/hook/memo/audit_test.go::TestAuditIndex_Overflow`

### AC-EXT001-04a — Feedback memory body structure (automated regex check)

- Maps: REQ-EXT001-010
- **Given** `type: feedback` 메모리 파일의 body
- **When** 감사 루틴이 정규식 `(?s).+?\*\*Why:\*\*.+?\*\*How to apply:\*\*` 로 body를 검사한다 (rule 텍스트 이후 `**Why:**`와 `**How to apply:**` 마커 순서 존재 확인)
- **Then** 정규식이 매치되어야 PASS (binary-testable). 미매치 시 `MEMORY_BODY_STRUCTURE_MISSING` non-blocking warning 발생
- Test path: `internal/hook/memo/taxonomy/audit_test.go::TestAuditFile_FeedbackBodyMarkers` (fixture: `feedback_testing.md` PASS, `feedback_missing_why.md` FAIL)
- Unblock criterion: 자동 regex 통과가 문서 레벨 PASS의 필요 조건 (객관적 기준)

### AC-EXT001-04b — Feedback memory prose quality (optional human review)

- Maps: REQ-EXT001-010 (보조)
- **Given** AC-04a를 통과한 `type: feedback` 메모리 파일
- **When** PR 리뷰어가 prose 품질을 검토한다 (Why/How to apply의 서술이 `moai-memory.md` 예시 톤과 유사한가)
- **Then** reviewer 판단에 따라 선택적 개선 의견 제시 (unblock criterion: AC-04a PASS이면 merge 차단 금지)
- Test path: 문서 레벨 AC — human review only. Automation은 AC-04a에서 이미 수행.

### AC-EXT001-05 — User memory content scope

- Maps: REQ-EXT001-009
- **Given** `type: user` 메모리 파일
- **When** 내용을 검토한다
- **Then** 개발자 역할/선호/책임만 기술되고, 일회성 태스크 상태(예: 현재 브랜치 작업 내역)가 포함되지 않는다
- Test path: 문서 레벨 AC (리뷰 체크리스트) + 예시 `fixtures/user_role.md`

### AC-EXT001-06 — Rule file writing guidelines

- Maps: REQ-EXT001-005
- **Given** `.claude/rules/moai/workflow/moai-memory.md`
- **When** 파일을 연다
- **Then** 4-type 각각에 대해 `when to save / how to use / body structure` 세 가지가 모두 기술되어 있다
- Test path: 문서 레벨 AC + (선택) `internal/template/rules_audit_test.go`에서 regex 존재 검증

### AC-EXT001-07 — SessionStart stale wrap

- Maps: REQ-EXT001-006
- **Given** `.claude/agent-memory/expert-backend/note.md`의 mtime이 25시간 전
- **When** SessionStart 훅이 실행된다
- **Then** 훅이 에이전트 컨텍스트로 전달하는 memory 블록이 `<system-reminder>...verify against current state before acting on it.</system-reminder>` 로 감싸져 있다
- Test path: `internal/hook/session_start_test.go::TestSessionStart_MemoryStaleWrap`

### AC-EXT001-08 — Aggregated stale warning

- Maps: REQ-EXT001-017
- **Given** 12개 memory 파일이 모두 mtime > 24h
- **When** SessionStart 훅이 실행된다
- **Then** 개별 12 warning 대신 단일 aggregated 경고 1줄이 출력되고, 포함된 파일 수 `12`가 메시지에 표기된다
- Test path: `internal/hook/session_start_extra_coverage_test.go::TestSessionStart_MemoryStaleAggregated`

### AC-EXT001-09 — Adding 5th type is rejected

- Maps: REQ-EXT001-004
- **Given** PR이 `ValidTypes`에 5번째 값을 추가하거나 `type: lesson` 같은 값을 허용하도록 변경
- **When** CI에서 `go test ./internal/hook/memo/...` 가 실행된다
- **Then** enum 불변성 테스트가 실패하고 실패 메시지에 `REQ-EXT001-004` 가 인용된다
- Test path: `internal/hook/memo/taxonomy_test.go::TestValidTypes_ImmutableSet`

### AC-EXT001-10 — Excluded category warning

- Maps: REQ-EXT001-015
- **Given** memory 파일 내용이 CLAUDE.md에서 복사한 텍스트(예: TRUST 5 섹션) 또는 git log / debug fix recipe 를 담고 있다
- **When** `AuditFile` 또는 PostToolUse 훅이 실행된다
- **Then** `MEMORY_EXCLUDED_CATEGORY` 경고가 출력되고 매칭된 카테고리 이름(`claude_md_mirror` / `git_history` / `debug_recipe` / `ephemeral_state` / `code_pattern`)이 detail에 포함된다
- Test path: `internal/hook/memo/audit_test.go::TestAuditFile_ExcludedCategory`

### AC-EXT001-11 — Duplicate description recommend merge

- Maps: REQ-EXT001-016
- **Given** 두 memory 파일이 동일한 frontmatter `description` 값을 가진다
- **When** `AuditDuplicates` 가 해당 디렉터리를 스캔한다
- **Then** `MEMORY_DUPLICATE` 경고가 출력되고 두 파일 경로가 detail에 포함된다. auto-merge 동작은 수행되지 않는다
- Test path: `internal/hook/memo/audit_test.go::TestAuditDuplicates`

### AC-EXT001-12 — Reference memory points externally

- Maps: REQ-EXT001-012
- **Given** `type: reference` memory 파일
- **When** 내용을 검토한다
- **Then** 외부 시스템(Linear/Grafana/Slack 등) URL 또는 project/channel 이름만 기술되며, 해당 외부 시스템의 상세 내용을 복제하여 저장하지 않는다
- Test path: 문서 레벨 AC + 예시 `fixtures/reference_grafana.md` + REQ-EXT001-005 가이드 검증

### AC-EXT001-13 — Project memory body structure

- Maps: REQ-EXT001-011
- **Given** `type: project` memory 파일
- **When** 내용을 검토한다
- **Then** body가 `사실/결정 → **Why:** → **How to apply:**` 순서로 기술된다
- Test path: 문서 레벨 AC + 예시 fixtures

---

## 2. Edge Cases

| Edge | 기대 동작 | Test |
|------|-----------|------|
| mtime == 정확히 24h | stale로 판정 (inclusive) | `TestDetectStale_Boundary` |
| memory dir 자체가 없음 | 오류 없이 skip, warning 없음 | `TestDetectStale_EmptyDir` |
| MEMORY.md 가 정확히 200줄 | overflow 없음 | `TestAuditIndex_EdgeExactly200` |
| 201줄 | overflow 경고 | `TestAuditIndex_EdgeOverflow` |
| frontmatter 자체가 없는 파일 | `ErrNoFrontmatter` + `MEMORY_MISSING_TYPE` | `TestParseFile_NoFrontmatter` |
| 파일 크기 0 byte | skip + warning 없음 | `TestParseFile_Empty` |
| description 은 같지만 type 이 다른 2 파일 | `MEMORY_DUPLICATE` (type 무관, description 기준) | `TestAuditDuplicates_SameDescDifferentType` |
| 10 stale 파일 (임계치 경계) | aggregation 발동 | `TestAggregateWarning_ExactlyTen` |
| 9 stale 파일 | 개별 warning 유지 | `TestAggregateWarning_Nine` |
| `MOAI_MEMORY_AUDIT=0` 설정 | SessionStart/PostToolUse 모두 audit skip | `TestSessionStart_AuditDisabled`, `TestPostTool_AuditDisabled` |
| Windows mtime precision | `time.Time` 주입 기반으로 동작 안정 | `TestDetectStale_InjectedNow` |
| symlink memory 파일 | follow 하지 않음, skip + log | `TestParseFile_Symlink` |

---

## 3. Traceability Matrix (REQ ↔ AC)

| REQ | 커버 AC | Test 경로 |
|-----|---------|-----------|
| REQ-EXT001-001 | AC-01, AC-02 | `taxonomy/taxonomy_test.go`, `taxonomy/audit_test.go` |
| REQ-EXT001-002 | AC-01b (name/description 명시 커버) | `taxonomy/taxonomy_test.go::TestParseFile_MissingName`, `TestParseFile_MissingDescription`, `TestParseFile_MissingBoth`, `taxonomy/audit_test.go::TestAuditFile_MissingFrontmatterKeys` |
| REQ-EXT001-003 | AC-03 | `taxonomy/audit_test.go` |
| REQ-EXT001-004 | AC-09 | `taxonomy/taxonomy_test.go::TestValidTypes_ImmutableSet` |
| REQ-EXT001-005 | AC-06 | 문서 + (선택) `rules_audit_test.go` |
| REQ-EXT001-006 | AC-07 | `session_start_test.go` |
| REQ-EXT001-007 | AC-01, AC-01b | `post_tool_test.go`, `taxonomy/audit_test.go` |
| REQ-EXT001-008 | AC-03 | `taxonomy/audit_test.go`, `post_tool_test.go` |
| REQ-EXT001-009 | AC-05 | 문서 + fixtures |
| REQ-EXT001-010 | AC-04a (automated), AC-04b (optional review) | `taxonomy/audit_test.go::TestAuditFile_FeedbackBodyMarkers` + 문서 |
| REQ-EXT001-011 | AC-13 | 문서 + fixtures |
| REQ-EXT001-012 | AC-12 | 문서 + fixtures |
| REQ-EXT001-013 | AC-02 (enum 소스 확정, 실제 loader 구현은 EXT-002에서 검증) | `taxonomy/taxonomy_test.go` + EXT-002 referral |
| REQ-EXT001-014 | (Optional, 30일 5 세션 미독) | Wave 2 이후 별도 SPEC 또는 후속 Run — 본 Wave 1 범위 외. 문서에 "may" 명시. |
| REQ-EXT001-015 | AC-10 | `taxonomy/audit_test.go::TestAuditFile_ExcludedCategory` |
| REQ-EXT001-016 | AC-11 | `taxonomy/audit_test.go::TestAuditDuplicates` |
| REQ-EXT001-017 | AC-08 | `session_start_extra_coverage_test.go` |

REQ-EXT001-014(Optional "may")는 Wave 1 deliverable에 포함되지 않음. spec.md §5.4 "may" 어구와 일치. acceptance.md는 "범위 외" 명시.

---

## 4. Test File Paths (Summary)

> 패키지 경로 주의: OPEN-3 해소에 따라 신규 코드는 `internal/hook/memo/taxonomy/` 서브패키지에 배치된다(plan.md §3.1 참조). 기존 `internal/hook/memo/` (session memo read/write)는 건드리지 않는다.

단위 (Go):

- `internal/hook/memo/taxonomy/taxonomy_test.go`
- `internal/hook/memo/taxonomy/staleness_test.go`
- `internal/hook/memo/taxonomy/audit_test.go`

통합 (Hook):

- `internal/hook/session_start_test.go` (기존 파일에 추가)
- `internal/hook/session_start_extra_coverage_test.go` (기존 파일에 추가)
- `internal/hook/post_tool_test.go` (기존 파일에 추가)

문서/Fixture:

- `internal/hook/memo/taxonomy/fixtures/user_role.md`
- `internal/hook/memo/taxonomy/fixtures/feedback_testing.md`
- `internal/hook/memo/taxonomy/fixtures/feedback_missing_why.md` (AC-04a negative)
- `internal/hook/memo/taxonomy/fixtures/project_migration.md`
- `internal/hook/memo/taxonomy/fixtures/reference_grafana.md`
- `internal/hook/memo/taxonomy/fixtures/missing_type.md`
- `internal/hook/memo/taxonomy/fixtures/missing_name.md` (AC-01b)
- `internal/hook/memo/taxonomy/fixtures/missing_description.md` (AC-01b)
- `internal/hook/memo/taxonomy/fixtures/unknown_type.md`
- `internal/hook/memo/taxonomy/fixtures/overflow_index.md`
- `internal/hook/memo/taxonomy/fixtures/excluded_claude_md.md`

---

## 5. Definition of Done Checklist

### 5.1 기능 (Functional)

- [ ] 17개 REQ 중 16개(REQ-014 제외) AC 매핑 완료, 테스트 녹색
- [ ] 모든 AC의 Test path 존재 및 `go test -count=1 -race ./...` 통과
- [ ] `.claude/rules/moai/workflow/moai-memory.md` 와 template mirror 내용 일치
- [ ] `MOAI_MEMORY_AUDIT=0` 경로가 SessionStart/PostToolUse 모두 skip 처리

### 5.2 비기능 (Quality Gates — TRUST 5)

- [ ] **Tested**: `internal/hook/memo/taxonomy/**` 커버리지 90%+, 기존 `internal/hook/memo/` (session memo) 및 `internal/hook/` 전반 회귀 없음
- [ ] **Readable**: 패키지명 `memo`, 공개 API godoc 영문, lint 경고 0
- [ ] **Unified**: `gofmt`/`goimports` 적용, `golangci-lint run` zero issues
- [ ] **Secured**: memory 파일은 read-only로만 접근, path traversal 방어, symlink follow 금지 (`TestParseFile_Symlink`)
- [ ] **Trackable**: Conventional Commits + SPEC-ID 커밋 트레일러(`Refs: SPEC-V3R2-EXT-001`)

### 5.3 워크플로우 (Process)

- [ ] `make build && make install` 수행, `moai version` commit hash = `git rev-parse HEAD`
- [ ] `go vet ./...` zero warnings
- [ ] `go test ./... -race -count=1` zero failures
- [ ] 템플릿/live YAML diff 없음 (workflow.yaml)
- [ ] 이모지 없음, 산문형 사용자 질문 없음 (AskUserQuestion 전용)
- [ ] 하드코딩 금지 — 24/200/10 리터럴이 코드 본문에 나타나지 않음 (단일 원천 상수 참조)

### 5.4 롤백 준비

- [ ] `MOAI_MEMORY_AUDIT=0` 수동 테스트 1회 성공 (SessionStart와 PostToolUse 모두 warning 억제 확인)
- [ ] Phase별 단일 커밋 유지 (revert 가능성 확보)

### 5.5 Open Question 해소

- [ ] OPEN-1(config key 이름) 최종 결정 기록
- [x] **OPEN-2(excluded category 판정 방식) — RESOLVED 2026-04-25**: static-keyword list v1 (plan.md §8 category→keyword 매핑 테이블). LLM 기반은 v2 연기.
- [x] **OPEN-3(`internal/hook/memo/` 기존 구조) — RESOLVED 2026-04-25**: Strategy B (adapted), 신규 코드는 `internal/hook/memo/taxonomy/` 서브패키지. 기존 session memo 패키지 비수정.
- [ ] OPEN-4(`rules_audit_test.go` 여부) 결정 기록
- [ ] 200-line cap 수치 공식 문서 근거 확인 (T0 precheck에서 수행)

---

End of Acceptance.
