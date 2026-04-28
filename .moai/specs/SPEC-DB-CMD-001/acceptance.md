---
spec_id: SPEC-DB-CMD-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  REVERSE-DOC: 본 acceptance.md는 SPEC 작성 이후 구현이 완료된 상태에서 SDD 아티팩트
  공백을 메우기 위해 역공학 방식으로 생성되었다. plan-auditor 2026-04-24 감사에서
  `acceptance.md` 미존재가 지적되어 backfill됨. AC는 실제 구현된 코드 동작으로부터
  파생되었으며, spec.md는 수정하지 않았다 (frozen).

  Implementation 근거:
  - commands/moai/db.md (7 LOC thin wrapper, 2026-04-21 커밋)
  - skills/moai/workflows/db.md (9 Phases: Router + Phase 1~8, 2026-04-21 커밋)
  - .moai/config/sections/db.yaml (8 keys — 5 system + 3 interview, 2026-04-21)
  - internal/template/commands_audit_test.go (thin wrapper 제약 강제)
---

# Acceptance Criteria — SPEC-DB-CMD-001

> `/moai db` Thin Wrapper + 서브커맨드 라우터 스킬의 실제 구현 동작으로부터 역도출된
> 검증 기준이다. 모든 AC는 파일:라인 또는 `go test -run` 참조로 역추적 가능하다.

## Traceability Matrix

| REQ ID | AC ID | 검증 수단 (파일:라인 또는 테스트) |
|---|---|---|
| REQ-001 (Thin Wrapper 위임) | AC-001 | `internal/template/templates/.claude/commands/moai/db.md:7` — `Use Skill("moai") with arguments: db $ARGUMENTS` |
| REQ-002 (LOC 제약) | AC-002 | `wc -l` < 20; `internal/template/commands_audit_test.go` (TestCommandsAudit) |
| REQ-003 (Frontmatter `argument-hint`) | AC-003 | `internal/template/templates/.claude/commands/moai/db.md:3` |
| REQ-004 (서브커맨드 디스패치) | AC-004 | `internal/template/templates/.claude/skills/moai/workflows/db.md:43-52` Router Table |
| REQ-005 (빈 인수 AskUserQuestion) | AC-005 | `workflows/db.md:54-66` — Empty first token 섹션 |
| REQ-006 (알 수 없는 서브커맨드) | AC-006 | `workflows/db.md:68-84` — Unknown first token 섹션 |
| REQ-007, REQ-008 (init 전제조건) | AC-007 | `workflows/db.md:88-109` — Phase 1: Preflight |
| REQ-009 (refresh 언어 중립 스캔) | AC-008 | `workflows/db.md:257-283` — Appendix: 16 Language Migration Path Mapping |
| REQ-010 (verify drift 정의) | AC-009 | `workflows/db.md:206-229` — Phase 7: Drift Detection |
| REQ-010a (verify 스키마 부재) | AC-010 | `workflows/db.md:225-227` — Fallback inline 로직 |
| REQ-011 (list Markdown 테이블) | AC-011 | `workflows/db.md:234-254` — Phase 8: List Tables |
| REQ-011a (list 스키마 부재) | AC-012 | `workflows/db.md:240-242` |
| REQ-012 (작업 추적 TaskCreate) | AC-013 | `workflows/db.md:131-144, 170-182` |
| REQ-013 (4-질문 인터뷰) | AC-014 | `workflows/db.md:113-124` — Phase 2: Interview |
| REQ-014 (allowed-tools 9개) | AC-015 | `workflows/db.md:9` frontmatter |
| REQ-015 (9 Phase 구조) | AC-016 | `grep -c "^## Phase" workflows/db.md` → 9 |
| REQ-016 (Thin Wrapper 용어) | AC-017 | `workflows/db.md:6, 32` |

---

## AC-001 — Thin Wrapper가 Skill에만 위임

**Given** `/moai db` 커맨드 파일이 렌더링된 상태
**When** 사용자가 `/moai db init` 을 호출할 때
**Then** 커맨드 파일 본문은 `Skill("moai")`로 위임하는 단 한 줄만 포함해야 하며, 비즈니스 로직이 존재하지 않아야 한다.
**Verification**:
- `internal/template/templates/.claude/commands/moai/db.md:7` = `Use Skill("moai") with arguments: db $ARGUMENTS`
- 본문에 조건문·루프·파일 접근 로직 부재 (`wc -l` = 7)

---

## AC-002 — LOC 제약 (20 라인 미만)

**Given** `commands/moai/db.md` 파일이 존재하는 상태
**When** `wc -l` 로 파일 라인 수를 측정할 때
**Then** 총 라인 수는 20 미만이어야 한다.
**Verification**:
- `wc -l /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/commands/moai/db.md` → 7
- `go test ./internal/template/... -run TestCommandsAudit` 통과 (thin wrapper 제약 자동 강제)

---

## AC-003 — Frontmatter `argument-hint` 필드

**Given** `commands/moai/db.md` frontmatter가 존재하는 상태
**When** frontmatter를 파싱할 때
**Then** `argument-hint` 필드 값은 정확히 `"init|refresh|verify|list [args]"` 이어야 하며, `description` 필드는 단일 문장이고 `allowed-tools: Skill` 이어야 한다.
**Verification**:
- `internal/template/templates/.claude/commands/moai/db.md:3` = `argument-hint: "init|refresh|verify|list [args]"`
- `internal/template/templates/.claude/commands/moai/db.md:2` = `description: Manage project database metadata (schema, migrations, seeds) via init/refresh/verify/list`
- `internal/template/templates/.claude/commands/moai/db.md:4` = `allowed-tools: Skill`

---

## AC-004 — 서브커맨드 디스패치 (Router Table)

**Given** `workflows/db.md` 스킬이 `$ARGUMENTS[0]` 을 수신한 상태
**When** 첫 토큰이 `init`, `refresh`, `verify`, `list` 중 하나일 때
**Then** 라우터는 해당 서브커맨드의 Phase 시퀀스로 분기해야 한다.
- `init` → Phase 1 (Preflight) → Phase 2 (Interview) → Phase 3 (Template Render) → Phase 4 (Hook Registration)
- `refresh` → Phase 5 (Scan) → Phase 6 (Regenerate)
- `verify` → Phase 7 (Drift Detection)
- `list` → Phase 8 (List Tables)

**Verification**:
- `workflows/db.md:43-52` Router Table이 위 4개 경로를 열거
- Phase 섹션 (L88, L113, L127, L152, L167, L188, L207, L234) 실제 존재

---

## AC-005 — 빈 인수 AskUserQuestion 폴백

**Given** 사용자가 `/moai db` 를 서브커맨드 없이 호출한 상태
**When** `$ARGUMENTS[0]` 이 비어있을 때
**Then** 스킬은 AskUserQuestion을 호출하여 4개 옵션(`init`, `refresh`, `verify`, `list`)을 제시해야 하며, 첫 번째 옵션은 `init (권장)` 이어야 한다.
**Verification**:
- `workflows/db.md:54-66` `### Empty first token` 섹션
- L60 `1. init (권장) — Initialize database metadata...`
- L61 ~ L63 나머지 3개 옵션

---

## AC-006 — 알 수 없는 서브커맨드 처리

**Given** 사용자가 `/moai db xyz` 와 같이 알려지지 않은 서브커맨드를 호출한 상태
**When** `$ARGUMENTS[0]` 이 `init`/`refresh`/`verify`/`list` 중 어느 것도 아닐 때
**Then** 스킬은 유효한 4개 서브커맨드 이름을 나열한 구조화된 오류를 출력하고 exit 1로 종료해야 하며, 어떠한 Phase도 실행하지 않아야 한다.
**Verification**:
- `workflows/db.md:68-84` `### Unknown first token` 섹션
- L72 `Error: Unknown subcommand "$ARGUMENTS[0]"`
- L75-79 4개 유효 서브커맨드 나열
- L84 `Exit with non-zero status (exit 1).`

---

## AC-007 — init 전제조건 검사

**Given** `/moai db init` 이 호출된 상태
**When** Phase 1 (Preflight) 이 실행될 때
**Then** 스킬은 `.moai/project/product.md` 와 `.moai/project/tech.md` 의 존재를 Glob으로 검증해야 하며, 어느 하나라도 없으면 `"Run /moai project first to generate product.md and tech.md"` 메시지와 함께 중단하고 어떠한 파일도 생성하지 않아야 한다.
**Verification**:
- `workflows/db.md:88-109` Phase 1: Preflight 섹션
- L94-96 필수 파일 목록 (product.md, tech.md)
- L103-107 누락 시 오류 메시지 (`"Run /moai project first..."`)

---

## AC-008 — refresh 스캔의 16개 언어 매핑

**Given** `/moai db refresh` 가 호출된 상태
**When** Phase 5 (Scan) 이 실행될 때
**Then** 스킬은 Appendix의 Language Migration Path Mapping 표에서 16개 MoAI 지원 언어(`go`, `python`, `typescript`, `javascript`, `rust`, `java`, `kotlin`, `csharp`, `ruby`, `php`, `elixir`, `cpp`, `scala`, `r`, `flutter`, `swift`)의 canonical 마이그레이션 경로를 참조해야 한다. 하드코딩된 언어별 Glob이 본문에 존재해서는 안 된다.
**Verification**:
- `workflows/db.md:264-281` Appendix 표 — 16개 언어 각 1행
- `grep -cE "^\| \`(go|python|typescript|javascript|rust|java|kotlin|csharp|ruby|php|elixir|cpp|scala|r|flutter|swift)\`" workflows/db.md` → 16

---

## AC-009 — verify drift 계산 (대칭 차)

**Given** `.moai/project/db/schema.md` 에 등록된 테이블 집합 S와 마이그레이션 파일에서 추출된 테이블 집합 M이 존재하는 상태
**When** `/moai db verify` 가 호출될 때
**Then** 스킬은 S와 M의 대칭 차(symmetric difference)를 drift로 계산해야 하며, 비어있지 않으면 unified diff를 stdout에 출력하고 exit 1로 종료한다. drift가 없으면 `"Schema documentation is in sync"` 를 출력하고 exit 0으로 종료한다.
**Verification**:
- `workflows/db.md:207-230` Phase 7: Drift Detection 섹션
- L219-221 delegated flow (moai-domain-db-docs Phase C3: no drift → exit 0, drift → exit 1)
- L228 inline fallback: `Extract table names... compute symmetric difference`

---

## AC-010 — verify 스키마 부재 처리

**Given** `.moai/project/db/schema.md` 가 존재하지 않는 디렉터리
**When** `/moai db verify` 가 호출될 때
**Then** 스킬은 `"Schema not found. Run /moai db init to initialize db schema."` 메시지와 함께 exit 2로 중단해야 하며, 어떠한 파일도 생성·수정하지 않아야 한다.
**Verification**:
- `workflows/db.md:225-227` Fallback inline 로직 — Glob 부재 검사 + 메시지 + exit 2

---

## AC-011 — list Markdown 정렬 테이블 출력

**Given** `.moai/project/db/schema.md` 에 테이블이 등록된 상태
**When** `/moai db list` 가 호출될 때
**Then** 스킬은 `table_name | column_count | primary_key | last_migration_file` 4개 열을 가진 Markdown 정렬 테이블을 stdout에 출력해야 하며, 파일 시스템에 어떠한 쓰기도 수행하지 않아야 한다.
**Verification**:
- `workflows/db.md:234-254` Phase 8: List Tables 섹션
- L237 `Read-only table listing. This phase MUST NOT create or modify any files.`
- L246-250 Markdown 테이블 예시 (정확한 4개 열 명시)
- L253 `exit 0. Do NOT write or modify any files.`

---

## AC-012 — list 스키마 부재 처리

**Given** `.moai/project/db/schema.md` 가 존재하지 않는 디렉터리
**When** `/moai db list` 가 호출될 때
**Then** 스킬은 `"Schema not found. Run /moai db init to initialize db schema."` 메시지와 함께 exit 2로 중단해야 하며, 어떠한 파일도 생성·수정하지 않아야 한다.
**Verification**:
- `workflows/db.md:240-242` — Glob 부재 검사 + 메시지 + exit 2

---

## AC-013 — TaskCreate/TaskUpdate 진행 추적

**Given** 다단계 서브커맨드 (`init`, `refresh`) 가 실행 중인 상태
**When** 2개 이상의 Phase를 가로지르는 작업을 수행할 때
**Then** 스킬은 Phase 3 (Template Render) 및 Phase 5 (Scan) 시작 시 `TaskCreate` 를, Phase 완료 시 `TaskUpdate` 를 호출해야 한다.
**Verification**:
- `workflows/db.md:131-144` Phase 3: `TaskCreate: "Render DB metadata templates"`, `TaskUpdate: mark complete`
- `workflows/db.md:170-182` Phase 5: `TaskCreate: "Scan migration files"`, `TaskUpdate: mark complete with file count`

---

## AC-014 — 4-질문 단일 AskUserQuestion 인터뷰

**Given** `/moai db init` 이 Phase 1 (Preflight) 를 통과한 상태
**When** Phase 2 (Interview) 에 진입할 때
**Then** 스킬은 **단일 `AskUserQuestion` 호출**로 정확히 4개 질문을 제시해야 한다: (1) Database engine, (2) ORM/query builder, (3) Multi-tenant strategy, (4) Migration tool.
**Verification**:
- `workflows/db.md:113-124` Phase 2: Interview 섹션
- L117 `Call AskUserQuestion once with exactly 4 questions (Claude Code limit: max 4 questions per call)`
- L118-L121 4개 주제 열거 (engine, ORM, multi-tenant, migration tool)

---

## AC-015 — 스킬 allowed-tools 9개 도구

**Given** `workflows/db.md` 스킬 frontmatter가 로드된 상태
**When** Claude Code가 frontmatter를 파싱할 때
**Then** `allowed-tools` 필드는 정확히 9개 도구를 포함해야 한다: `AskUserQuestion, Read, Write, Edit, Bash, TaskCreate, TaskUpdate, Glob, Grep`.
**Verification**:
- `workflows/db.md:9` = `allowed-tools: AskUserQuestion, Read, Write, Edit, Bash, TaskCreate, TaskUpdate, Glob, Grep`
- CSV 형식 (공백 구분 아님), 9개 토큰

---

## AC-016 — 9개 Phase 구조 완전성

**Given** `workflows/db.md` 파일이 렌더링된 상태
**When** 본문에서 `^## Phase` 헤딩을 `grep -c` 로 세었을 때
**Then** 정확히 9개의 Phase 헤딩이 존재해야 한다: Phase 0 (Router), Phase 1 (Preflight), Phase 2 (Interview), Phase 3 (Template Render), Phase 4 (Hook Registration), Phase 5 (Scan), Phase 6 (Regenerate), Phase 7 (Drift Detection), Phase 8 (List Tables).
**Verification**:
- `grep -c "^## Phase" internal/template/templates/.claude/skills/moai/workflows/db.md` → 9
- 각 Phase의 시작 라인: L39, L88, L113, L127, L152, L167, L188, L207, L234

---

## AC-017 — "Thin Wrapper" 용어 일관성

**Given** `workflows/db.md` 본문이 존재하는 상태
**When** 본문 내 용어 사용을 grep으로 검사할 때
**Then** "Thin Wrapper" 용어가 사용되어야 하며, `"Single-File Command"` 표기는 본문(HISTORY 제외)에 존재하지 않아야 한다.
**Verification**:
- `workflows/db.md:6` = `through four subcommands: init ... Thin Wrapper entry point delegates...`
- `workflows/db.md:32` = `This Thin Wrapper router dispatches $ARGUMENTS[0]...`
- `grep -c '"Single-File Command"' workflows/db.md` → 0

---

## Edge Cases and Defensive Behaviors

- **EC-01**: `/moai db` (no subcommand, no AskUserQuestion support) → Router Table (L52) 지정된 fallback. Manual session 검증 필요.
- **EC-02**: `/moai db init` 실행 시 `.moai/project/db/` 디렉터리가 이미 존재하고 사용자가 편집한 파일을 포함하는 경우 → SPEC-DB-TEMPLATES-001 REQ-013 (사용자 수정 보호) 위임. 본 SPEC 범위 밖.
- **EC-03**: Phase 3/4/5 의 실제 렌더링·훅 등록·파싱 로직은 skeleton 수준이며 SPEC-DB-TEMPLATES-001 / SPEC-DB-SYNC-001 이 구현. 본 SPEC은 Router + Phase 뼈대만 검증.

---

## Definition of Done

- [x] `commands/moai/db.md` (7 LOC, < 20 제약 준수)
- [x] `workflows/db.md` (9 Phases, 라우터 + 서브커맨드 skeleton)
- [x] `commands_audit_test.go` 에 의한 thin wrapper 자동 검증
- [x] 16개 언어 매핑 표 (Appendix)
- [x] allowed-tools 9 도구 CSV
- [x] AskUserQuestion 4-질문 단일 호출
- [x] 모든 Phase 섹션 `^## Phase` 헤딩으로 식별 가능
- [x] `go test ./internal/template/...` 통과

---

## Quality Gate Alignment

| Gate | Criterion | Evidence |
|---|---|---|
| Tested | thin wrapper 제약 자동 테스트 | `commands_audit_test.go` (TestCommandsAudit) |
| Readable | 용어 통일 ("Thin Wrapper") | REQ-016 / AC-017 |
| Unified | CSV allowed-tools + Phase 네이밍 규칙 | AC-015 / AC-016 |
| Secured | 방어적 실패 (empty arg, unknown, missing prereq) | AC-005 / AC-006 / AC-007 |
| Trackable | TaskCreate/TaskUpdate 진행 추적 | AC-013 |

---

## Reverse-Doc Divergence Notes

- 구현과 SPEC이 **일치**함. 모든 REQ가 workflows/db.md 에 해당 Phase로 매핑됨.
- Phase 3 (Template Render), Phase 4 (Hook Registration) 는 skeleton 수준이며 SPEC-DB-TEMPLATES-001, SPEC-DB-SYNC-001 에 상세 위임됨 — 이는 SPEC의 명시된 Scope 경계와 일치.
- 본 SPEC의 AC-5 (AskUserQuestion 시각 확인) 는 수동 Claude Code 세션 검증이 필요하며 자동화되지 않음 (plan-audit review-2 N1 참조).
