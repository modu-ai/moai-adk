---
id: SPEC-DB-CMD-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-21
author: moai-adk-go
priority: High
labels: [db, command, cli, thin-wrapper, subcommand-routing]
issue_number: null
depends_on: []
related_specs: [SPEC-DB-TEMPLATES-001, SPEC-DB-SYNC-001, SPEC-PROJECT-DB-HINT-001]
---

# SPEC-DB-CMD-001: /moai db Thin Wrapper with Subcommand Router Skill

## HISTORY

- 2026-04-21 v0.2.0 — plan-auditor iteration 1 FAIL 후속 수정. Frontmatter MoAI 표준(`version`, `created_at`, `labels`, `priority: High`, `issue_number`, `depends_on`) 반영. REQ ↔ AC traceability 100% 복구(모든 AC에 `[REQ-XXX]` 태그, REQ-014/REQ-015/REQ-016 신설). REQ당 1줄 **Rationale** 보강. 용어 "Single-File Command" → "Thin Wrapper"로 통일. "drift" 정의 정밀화. "가독성 있는 형식" 구체 출력 스펙으로 대체. "4 라운드" 애매성 해소. 언어 중립성 확보(16개 언어 자동 탐지 전략).
- 2026-04-20 v0.1.0 — SPEC 최초 작성.

## Background

`/moai db` 커맨드는 프로젝트의 데이터베이스 메타데이터(스키마, 마이그레이션, 시드)를 관리하는 4개 서브커맨드(`init`, `refresh`, `verify`, `list`)를 제공한다. 본 SPEC은 해당 커맨드 엔트리포인트의 **파일 구조 결정**(단일 파일 thin wrapper + 라우터 스킬)만을 정의하며, 실제 템플릿 렌더링·훅 연동·다른 커맨드와의 통합은 명시적으로 범위 밖이다.

### 사용자 결정: 단일 파일 패턴 (Single-File Thin Wrapper)

사용자는 다중 파일 대안(`db/init.md`, `db/refresh.md`, `db/verify.md`, `db/list.md` — 4개 파일)이 아닌 **단일 `db.md` 파일**을 명시적으로 선택하였다. 선택 근거는 다음과 같다.

- **파일 관리 부담 감소**: 4개 파일 대신 1개 파일만 유지
- **서브커맨드 추가 확장성**: 새로운 서브커맨드는 스킬 본체(`workflows/db.md`)만 수정하면 되며 커맨드 파일은 건드리지 않음
- **라우터 테이블 단일 진실 공급원**: 중앙 라우터가 모든 서브커맨드 분기를 한곳에서 문서화
- **기존 `/moai` 루트 패턴과의 일관성**: `$ARGUMENTS` 첫 토큰으로 서브커맨드에 라우팅하는 기존 관례를 그대로 재사용

### 기존 패턴 참조

`.claude/commands/moai/plan.md`는 8 LOC thin wrapper이다.

```
---
description: Create SPEC document with EARS format requirements and acceptance criteria
argument-hint: "\"description\" [--team] [--worktree] [--branch] [--resume SPEC-XXX]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: plan $ARGUMENTS
```

호출 흐름: `/moai plan "my feature"` → `$ARGUMENTS = "my feature"` → `Skill("moai")`에 `plan my feature` 전달 → 스킬이 첫 토큰으로 라우팅.

`/moai db init`도 동일 패턴을 2-level subcommand 구조로 확장한다.

- `commands/moai/db.md` (단일 파일, ~8 LOC)
- `Skill("moai")`에 `db $ARGUMENTS` 전달
- `Skill("moai")`이 `workflows/db.md`로 라우팅
- `workflows/db.md` 내부 라우터가 `$ARGUMENTS[0]`(`init`/`refresh`/`verify`/`list`)으로 분기

### 헌법적 제약: Thin Command Pattern

`.claude/rules/moai/development/coding-standards.md`의 Thin Command Pattern 규칙:

- 커맨드 파일 body 20 LOC 미만
- 커맨드는 `Skill("moai")`를 통해 라우팅
- 모든 로직은 `.claude/skills/moai/workflows/`에 위치
- `internal/template/commands_audit_test.go`가 `go test` 실행 시 자동 검증

### 언어 중립성 원칙

대상 파일은 `internal/template/templates/.claude/**` 하위이므로 `CLAUDE.local.md` §15에 따라 **16개 MoAI 지원 언어를 동등 취급**한다. 특정 언어(예: TypeScript/Python/Ruby)만 하드코딩된 Glob 패턴을 명시하는 대신, 프로젝트의 `project_markers` 기반 자동 탐지 전략을 채택한다(REQ-009).

## Target Files

- **NEW**: `internal/template/templates/.claude/commands/moai/db.md` (thin wrapper)
- **NEW**: `internal/template/templates/.claude/skills/moai/workflows/db.md` (라우터 + 서브커맨드 skeleton 포함 워크플로우 스킬)

## Requirements (EARS)

### REQ-001: Thin Wrapper 위임

WHEN `/moai db`가 호출될 때, thin wrapper `commands/moai/db.md`는 `Skill("moai")`에 `"db $ARGUMENTS"` 형태의 arguments를 전달하여 위임해야 하며, 어떠한 비즈니스 로직도 포함해서는 안 된다.

**Rationale:** Thin Command Pattern 준수; 모든 로직은 스킬 본체에 단일 공급원으로 집중.

### REQ-002: LOC 제약

WHEN `commands/moai/db.md`가 렌더링될 때, 해당 파일은 frontmatter와 body를 포함하여 총 20 라인 미만이어야 한다.

**Rationale:** `commands_audit_test.go`가 강제하는 헌법적 제약; 커맨드 파일 비대화 방지.

### REQ-003: Frontmatter 필드

WHEN `commands/moai/db.md`의 frontmatter가 파싱될 때, `description` 필드는 단일 문장이어야 하며, `argument-hint` 필드는 4개 서브커맨드를 모두 명시해야 한다: `"init|refresh|verify|list [args]"`.

**Rationale:** Claude Code UI가 argument-hint로 사용자 자동완성/가이드를 제공; 서브커맨드 인지도 향상.

### REQ-004: 서브커맨드 디스패치

WHEN 스킬 `workflows/db.md`가 `$ARGUMENTS`의 첫 토큰으로 [`init`, `refresh`, `verify`, `list`] 중 하나를 수신할 때, 해당 서브커맨드 단계(Phase)로 디스패치해야 한다.

**Rationale:** 단일 스킬 파일이 4개 서브커맨드의 라우팅 허브 역할을 수행하는 핵심 동작.

### REQ-005: 빈 인수 처리

WHEN 스킬이 빈 첫 토큰의 `$ARGUMENTS`를 수신할 때, `AskUserQuestion`을 제시하여 사용자가 서브커맨드를 선택하도록 요구해야 하며, 첫 번째 옵션은 `init`(권장)이어야 한다.

**Rationale:** 사용자가 서브커맨드를 생략했을 때 UX 폴백; MoAI AskUserQuestion HARD 규칙 준수.

### REQ-006: 알 수 없는 서브커맨드 처리

WHEN 스킬이 알 수 없는 첫 토큰을 수신할 때, 유효한 서브커맨드(`init`, `refresh`, `verify`, `list`)를 나열한 구조화된 오류를 반환해야 하며, 어떠한 Phase도 실행하지 않고 종료해야 한다.

**Rationale:** 오타·오기입 시 즉시 실패(fail-fast); 의도치 않은 Phase 실행 방지.

### REQ-007: init 전제조건 검사

WHEN `/moai db init`이 호출될 때, 스킬은 전제조건 파일을 검사해야 한다: `.moai/project/product.md` AND `.moai/project/tech.md`가 모두 존재해야 한다.

**Rationale:** DB 스키마는 프로덕트 요구사항 및 기술 스택 결정 이후에만 의미가 있음; 선후 관계 강제.

### REQ-008: 전제조건 부재 시 중단

IF `/moai db init` 실행 중 어느 한 전제조건 파일이 누락되면, THEN 스킬은 `"Run /moai project first to generate product.md and tech.md"` 메시지와 함께 중단해야 하며 어떠한 파일도 생성해서는 안 된다.

**Rationale:** 불완전 상태에서 DB 아티팩트를 생성하면 rollback 비용 증가; 방어적 실패.

### REQ-009: refresh 마이그레이션 스캔 (언어 중립)

WHEN `/moai db refresh`가 호출될 때, 스킬은 `.moai/config/sections/language.yaml`의 `project_markers` 또는 MoAI 언어 자동 탐지 결과를 기반으로 해당 프로젝트의 마이그레이션 파일을 Glob 스캔해야 하며, 16개 지원 언어(`go`, `python`, `typescript`, `javascript`, `rust`, `java`, `kotlin`, `csharp`, `ruby`, `php`, `elixir`, `cpp`, `scala`, `r`, `flutter`, `swift`) 각각에 대해 canonical 마이그레이션 도구의 기본 경로를 매핑 테이블(`workflows/db.md`의 별도 부록 섹션)에서 참조해야 한다. 특정 언어의 경로 하드코딩은 금지되며, 사용자가 설정 파일로 경로를 override할 수 있어야 한다.

**Rationale:** 템플릿 파일은 16개 언어 동등 취급(CLAUDE.local.md §15); 하드코딩 대신 자동 탐지 + 매핑 테이블로 확장성 확보.

### REQ-010: verify drift 정의 및 검출

WHEN `/moai db verify`가 호출될 때, 스킬은 `.moai/project/db/schema.md`에 등록된 **테이블/컬렉션 이름 집합**과 REQ-009의 스캔으로 검출된 **마이그레이션 파일에서 추출 가능한 테이블/컬렉션 이름 집합** 간의 집합 대칭 차(symmetric difference)를 "drift"로 정의하고 계산해야 하며, 대칭 차가 공집합이 아니면 non-zero exit code(exit 1)로 종료해야 한다. drift 상세는 stdout에 추가된/누락된 이름을 분리해 출력한다.

**Rationale:** "drift"를 집합 대칭 차로 정량 정의하여 테스트 가능성 확보(plan-auditor D6 해소).

### REQ-010a: verify 스키마 부재 처리

IF `/moai db verify` 실행 시 `.moai/project/db/schema.md`가 존재하지 않으면, THEN 스킬은 `"Schema not found. Run /moai db init to initialize db schema."` 메시지와 함께 exit 2로 중단해야 하며 어떠한 파일도 생성·수정해서는 안 된다.

**Rationale:** init 선행 필수 경로 명시; `/moai db list`와 동일한 방어 패턴 공유.

### REQ-011: list 출력 포맷 및 읽기 전용성

WHEN `/moai db list`가 호출될 때, 스킬은 `.moai/project/db/schema.md`에 등록된 테이블/컬렉션을 **Markdown 정렬 테이블** 형식(열: `table_name`, `column_count`, `primary_key`, `last_migration_file`)으로 stdout에 출력해야 하며, 파일 시스템에 어떠한 쓰기(create/modify/delete)도 수행해서는 안 된다.

**Rationale:** 구체적 출력 스키마로 weasel phrase("가독성 있는 형식") 제거; `git status --porcelain`으로 읽기 전용 검증 가능(plan-auditor D8 해소).

### REQ-011a: list 스키마 부재 처리

IF `/moai db list` 실행 시 `.moai/project/db/schema.md`가 존재하지 않으면, THEN 스킬은 `"Schema not found. Run /moai db init to initialize db schema."` 메시지와 함께 exit 2로 중단해야 하며 어떠한 파일도 생성·수정해서는 안 된다.

**Rationale:** verify와 동일한 방어 경로; init이 선행되지 않은 상태에서의 오동작 방지.

### REQ-012: 작업 추적

WHILE 어떤 서브커맨드가 실행 중일 때, 스킬은 다단계 작업(Phase 2개 이상을 가로지르는 서브커맨드)에 대해 `TaskCreate`/`TaskUpdate`를 사용하여 진행 상태를 추적해야 한다.

**Rationale:** 장시간 실행되는 다단계 작업에서 사용자 가시성 확보; MoAI 표준 진행 추적 메커니즘 준수.

### REQ-013: init 인터뷰 구조

WHEN 스킬이 `init`을 시작할 때, 다운스트림 템플릿 호출 이전에 **단일 `AskUserQuestion` 호출**로 정확히 4개 질문(`engine`, `ORM`, `multi-tenant`, `migration tool`)을 한 번에 제시해야 한다(Claude Code의 AskUserQuestion 단일 호출 최대 4개 질문 제한 정확히 활용).

**Rationale:** "4 라운드" 애매성 제거; AskUserQuestion 한 번으로 모든 초기화 결정을 원자적으로 수집(plan-auditor D9 해소).

### REQ-014: 스킬 allowed-tools 선언

WHEN `workflows/db.md` 스킬의 frontmatter가 로드될 때, `allowed-tools` 필드는 정확히 `AskUserQuestion, Read, Write, Edit, Bash, TaskCreate, TaskUpdate, Glob, Grep` 9개 도구를 포함해야 한다.

**Rationale:** 스킬이 필요로 하는 도구 권한을 REQ로 명시적으로 고정(plan-auditor D11 orphan AC 해소); 권한 드리프트 방지.

### REQ-015: Phase 구조 완전성

WHEN `workflows/db.md`가 렌더링될 때, 본문은 `^## Phase` 헤딩을 최소 9개 포함해야 한다: Phase 0 (Router), Phase 1 (Preflight), Phase 2 (Interview), Phase 3 (Template Render), Phase 4 (Hook Registration), Phase 5 (Scan), Phase 6 (Regenerate), Phase 7 (Drift Detection), Phase 8 (List Tables).

**Rationale:** Router Table과 Phase 구조를 강제 일치; plan-auditor D5(AC-4 under-specified) 해소.

### REQ-016: 용어 및 HISTORY 포맷

WHEN SPEC 본문 또는 관련 문서에서 본 커맨드를 지칭할 때, 용어는 "Thin Wrapper"로 통일되어야 하며 "Single-File Command" 표기는 사용하지 않아야 한다. HISTORY 엔트리는 `YYYY-MM-DD vX.Y.Z — <change summary>` 포맷을 준수해야 한다.

**Rationale:** 용어 일관성 및 변경 이력 추적 가능성 확보(plan-auditor D12, D13 해소).

## Acceptance Criteria

- **AC-1 [REQ-002]**: `internal/template/templates/.claude/commands/moai/db.md` 파일이 존재하며 `wc -l`이 20 미만을 반환한다.
- **AC-2 [REQ-001, REQ-004, REQ-015]**: `internal/template/templates/.claude/skills/moai/workflows/db.md` 파일이 존재하며 YAML frontmatter(`name: moai-workflow-db`)와 본문 섹션(Phase 0 라우팅 + Phase 1~8 서브커맨드 본체)을 포함한다.
- **AC-3 [REQ-002]**: `go test ./internal/template/... -run TestCommandsAudit`가 통과한다(`commands_audit_test.go`가 thin wrapper 제약을 검증).
- **AC-4 [REQ-015]**: `grep -c "^## Phase" internal/template/templates/.claude/skills/moai/workflows/db.md`가 9 이상을 반환한다(Phase 0 router + Phase 1~8 서브커맨드).
- **AC-5 [REQ-005]**: Claude Code 세션에서 서브커맨드 없이 `/moai db`를 호출하면 4개 옵션(`init`/`refresh`/`verify`/`list`)을 가진 `AskUserQuestion`이 트리거되며 첫 번째 옵션이 `init (권장)`이다(수동 세션 검증).
- **AC-6 [REQ-007, REQ-008]**: `.moai/project/product.md`가 없는 디렉터리에서 `/moai db init`을 호출하면 `"Run /moai project first"` 부분 문자열이 포함된 메시지와 함께 중단되며 `.moai/project/db/` 하위에 어떤 파일도 생성되지 않는다.
- **AC-7 [REQ-003]**: `commands/moai/db.md`의 `argument-hint` 값이 정확히 `init|refresh|verify|list [args]`와 일치한다(grep으로 확인).
- **AC-8 [REQ-014]**: `workflows/db.md`의 frontmatter `allowed-tools` 필드가 정확히 `AskUserQuestion, Read, Write, Edit, Bash, TaskCreate, TaskUpdate, Glob, Grep` 9개 도구를 포함한다(순서 무관).
- **AC-9 [REQ-006]**: `/moai db xyz`(알 수 없는 서브커맨드) 호출 시 유효한 4개 서브커맨드 이름을 포함한 오류 메시지가 반환되고 어떤 Phase도 실행되지 않는다(exit code ≠ 0).
- **AC-10 [REQ-009]**: `workflows/db.md` 부록 섹션에 16개 언어(`go`, `python`, `typescript`, `javascript`, `rust`, `java`, `kotlin`, `csharp`, `ruby`, `php`, `elixir`, `cpp`, `scala`, `r`, `flutter`, `swift`) 각각에 대한 canonical 마이그레이션 경로 매핑 테이블이 존재한다(grep으로 언어 이름 16개 모두 검증).
- **AC-11 [REQ-010]**: `.moai/project/db/schema.md`에 테이블 `users`만 등록되고 마이그레이션에는 `users`, `orders`가 존재하는 테스트 픽스처에서 `/moai db verify`는 exit 1로 종료하며 stdout에 `orders`가 "added" 또는 "missing in schema.md"로 출력된다.
- **AC-12 [REQ-010a]**: `.moai/project/db/schema.md`가 없는 디렉터리에서 `/moai db verify`는 `"Schema not found"` 메시지와 함께 exit 2로 종료한다.
- **AC-13 [REQ-011]**: 테스트 픽스처에서 `/moai db list` 실행 후 `git status --porcelain`이 공백 출력을 반환한다(파일 시스템 mutation 없음). stdout에는 열 `table_name | column_count | primary_key | last_migration_file`을 가진 Markdown 정렬 테이블이 출력된다.
- **AC-14 [REQ-011a]**: `.moai/project/db/schema.md`가 없는 디렉터리에서 `/moai db list`는 `"Schema not found"` 메시지와 함께 exit 2로 종료한다.
- **AC-15 [REQ-012]**: `workflows/db.md` 본문의 `init`, `refresh` Phase 섹션에 `TaskCreate` 및 `TaskUpdate` 토큰이 각각 1회 이상 등장한다(grep 검증).
- **AC-16 [REQ-013]**: `workflows/db.md`의 Phase 2 (Interview) 섹션에 "AskUserQuestion" 토큰이 1회만 등장하며, 해당 호출이 정확히 4개 질문 주제(`engine`, `ORM`, `multi-tenant`, `migration tool`)를 모두 명시한다.
- **AC-17 [REQ-016]**: `spec.md` 본문 내 "Single-File Command" 표기 출현 횟수가 0이며, HISTORY 엔트리가 `YYYY-MM-DD vX.Y.Z` 포맷을 따른다.

## Router Table (구현 가이드)

스킬 본체 `workflows/db.md` Phase 0은 다음 라우터 테이블을 충실히 구현해야 한다.

| First $ARGUMENTS token | Action | Target Phase |
|---|---|---|
| `init` | Full interactive init | Phase 1 (Preflight) → 2 (Interview) → 3 (Template Render) → 4 (Hook Registration) |
| `refresh` | Rescan migrations, rebuild docs | Phase 5 (Scan) → Phase 6 (Regenerate) |
| `verify` | Drift check (read-only) | Phase 7 (Drift Detection) |
| `list` | Read-only table output | Phase 8 (List Tables) |
| (empty) | AskUserQuestion select | return to Phase 0 after selection |
| \<unknown\> | Structured error, exit | N/A |

## Scope

### IN SCOPE

- `commands/moai/db.md` thin wrapper 파일 작성
- `workflows/db.md` 스킬 본체 (라우터 + 4개 서브커맨드 Phase skeleton, 16개 언어 마이그레이션 경로 매핑 부록)
- `argument-hint` 사양 정의
- `init` 전제조건 검사 로직
- 알 수 없는 서브커맨드 / 빈 인수 / 스키마 부재 오류 처리
- `allowed-tools` 선언
- Phase 구조 (9개 Phase) 완전성

### OUT SCOPE (Exclusions — What NOT to Build)

- 실제 템플릿 렌더링 로직 (SPEC-DB-TEMPLATES-001에서 처리)
- `PostToolUse` 훅 통합 (SPEC-DB-SYNC-001에서 처리)
- `/moai project` 통합 힌트 (SPEC-PROJECT-DB-HINT-001에서 처리)
- `moai-domain-db-docs` subagent 본체 (SPEC-DB-SYNC-001에서 처리)
- DB 엔진별 상세 스키마 생성 로직
- 마이그레이션 파일 자동 생성 로직
- DB 연결 및 실제 스키마 introspection
- 사용자 커스텀 마이그레이션 경로 override 구현 디테일(본 SPEC은 선언만, 구현은 SPEC-DB-TEMPLATES-001)

## Risks

- **R-1**: 커맨드 파일이 20 LOC을 초과할 위험 — `commands_audit_test.go`가 CI에서 자동 차단.
- **R-2**: 중첩 서브커맨드 구조가 사용자에게 혼란을 줄 수 있음 — `argument-hint` 명시(REQ-003)와 `AskUserQuestion` 폴백(REQ-005)으로 완화.
- **R-3**: 16개 언어의 canonical 마이그레이션 도구가 미래에 추가·변경될 위험 — REQ-009의 부록 매핑 테이블을 single source로 유지하고, 사용자 override 옵션을 SPEC-DB-TEMPLATES-001에서 구현하여 확장성 확보.
- **R-4**: 집합 대칭 차 기반 drift 정의가 컬럼·인덱스 수준 변경을 잡아내지 못할 위험(MVP 한계) — 본 SPEC은 테이블/컬렉션 이름 수준만 커버함을 명시하고, 컬럼/인덱스 drift는 차기 SPEC에서 다룸.
