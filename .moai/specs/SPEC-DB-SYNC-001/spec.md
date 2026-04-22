---
id: SPEC-DB-SYNC-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-21
author: moai-adk-go
priority: Medium
labels: [db, hook, post-tool-use, schema-sync, drift, moai-domain-db-docs, debounce]
issue_number: null
depends_on: [SPEC-DB-CMD-001, SPEC-DB-TEMPLATES-001]
related_specs: [SPEC-PROJECT-DB-HINT-001]
---

# SPEC-DB-SYNC-001: PostToolUse Hook + moai-domain-db-docs Skill + DB Drift Verification

## HISTORY

- 2026-04-21 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. Frontmatter 표준(version/labels/created_at), AC-1 traceability 보강, REQ-008 scope 정리(파서 구현 세부 제거), 중복 Excluded Patterns 블록 통합, AC-4 유효 frontmatter 기준 명시, 각 REQ에 Rationale 추가.
- 2026-04-20 v0.1.0: SPEC 최초 작성.

## Background

마이그레이션 파일(Prisma schema, Alembic version, Rails migration, raw SQL)이 수정될 때마다 `.moai/project/db/schema.md`, `erd.mmd`, `migrations.md` 문서가 자동 동기화되어야 한다. 본 SPEC은 다음 세 가지를 정의한다.

1. PostToolUse 훅 패턴: `Write`/`Edit` 이벤트를 감시하여 마이그레이션 파일 변경을 감지하는 bash wrapper + Go hook subcommand
2. `moai-domain-db-docs` 스킬: 마이그레이션 파일을 파싱하여 DB 문서를 갱신하는 로직
3. `/moai db refresh` 및 `/moai db verify` 서브커맨드 실제 로직 (SPEC-DB-CMD-001에서 스텁으로 정의된 명령을 본 SPEC에서 구현)

Debounce(10초) 및 recursion guard(`.moai/project/db/**` 제외)로 무한 루프를 방지한다. 사용자 승인은 AskUserQuestion을 통해 오케스트레이터 레벨에서 처리한다.

### Scope Boundary

- SPEC-DB-CMD-001과의 경계: SPEC-DB-CMD-001은 `/moai db` 명령 표면과 서브커맨드 라우팅을 정의(13 REQ)하며 `refresh`/`verify` 동작은 스텁이다. 본 SPEC은 훅 및 두 서브커맨드가 호출하는 실제 파싱/diff/문서 갱신 로직을 구현한다.
- SPEC-DB-TEMPLATES-001과의 경계: SPEC-DB-TEMPLATES-001은 정적 템플릿을 정의한다. 본 SPEC은 마이그레이션 파일 파싱 결과를 기반으로 해당 템플릿에 동적 콘텐츠를 기입한다.

### Migration File Detection

대상 glob(SPEC-DB-TEMPLATES-001의 `db.yaml` → `migration_patterns`):

- `prisma/schema.prisma`
- `alembic/versions/**/*.py`
- `db/migrate/**/*.rb`
- `migrations/**/*.sql`
- `supabase/migrations/**/*.sql`
- `sql/migrations/**/*.sql`

### Excluded Patterns (recursion guard) — Single Source of Truth

다음 glob 집합은 훅과 스킬 모두에서 제외 대상으로 적용한다. 본 섹션이 **유일한 정의처**이며, 하위 섹션/요구사항에서는 "Excluded Patterns 섹션 참조"로만 지칭한다.

- `.moai/project/db/**` (문서 스스로의 업데이트에 의한 재귀 방지)
- `.moai/cache/**`
- `.moai/logs/**`

## Requirements (EARS)

### Hook Configuration

1. **REQ-001**: WHEN moai installs `settings.json.tmpl`, it SHALL register a PostToolUse matcher for tool types `Write` and `Edit` with the exact command path `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` and timeout `30000`.
   - **Rationale**: 훅 스크립트가 `settings.json`에 등록되지 않으면 Claude Code가 이벤트를 디스패치하지 않으므로 전체 동기화 파이프라인이 작동하지 않는다.
2. **REQ-002**: WHEN the PostToolUse hook runs, it SHALL read `tool_input.file_path` from stdin JSON.
   - **Rationale**: Claude Code는 훅에 JSON을 stdin으로 전달하며, 변경된 파일 경로는 필터링 판정의 단일 입력이다.
3. **REQ-003**: WHEN the read `file_path` does NOT match any pattern in `db.yaml` `migration_patterns`, the hook SHALL exit `0` silently without further action.
   - **Rationale**: 모든 Write/Edit 이벤트에 대해 Go 바이너리를 기동하면 지연이 누적되므로, 마이그레이션 패턴이 아닐 때는 즉시 종료한다.
4. **REQ-004**: WHEN the read `file_path` matches the Excluded Patterns section (e.g., `.moai/project/db/**`), the hook SHALL exit `0` silently (recursion guard).
   - **Rationale**: 문서 자체가 변경될 때 훅이 재귀 호출되면 무한 루프가 발생하므로 훅 레벨에서 먼저 차단한다.
5. **REQ-005**: WHEN the read `file_path` matches a `migration_pattern`, the hook SHALL invoke `"moai hook db-schema-sync --file <path>"` and capture its output.
   - **Rationale**: bash wrapper는 경량 필터 역할만 하고, 파싱·상태 저장·JSON 출력은 Go 서브커맨드가 책임져야 테스트 가능성이 확보된다.

### Debounce

6. **REQ-006**: WHEN the same `file_path` is reported to the hook within 10 seconds of the previous invocation, the handler SHALL skip processing and log the debounce event.
   - **Rationale**: 편집기 자동 저장 또는 AI 도구의 연속 Edit가 수 초 간격으로 발생할 때 중복 파싱을 피해 CPU와 사용자 승인 피로를 줄인다.
7. **REQ-007**: WHEN debounce is triggered, the state file `.moai/cache/db-sync/last-seen.json` SHALL record the debounced event with `timestamp` and `file_path`.
   - **Rationale**: 디바운스 판정은 상태가 없으면 불가능하므로 `last-seen.json`을 유일한 진실로 고정한다.

### Go Hook Handler (`moai hook db-schema-sync`)

8. **REQ-008**: WHEN `moai hook db-schema-sync` receives a `--file <path>` argument, it SHALL parse the migration file into a normalized schema representation that covers the migration frameworks enumerated in `db.yaml` `migration_patterns`. The specific parser module used (resolved under `internal/db/parser/`) is an implementation detail out of scope for this SPEC.
   - **Rationale**: 본 SPEC은 파서 **행동 계약**(입력 = 마이그레이션 파일, 출력 = 정규화된 스키마 표현)만 규정한다. 파서 구현 기술은 별도 모듈에서 진화하므로 REQ 본문에 고정하지 않는다.
9. **REQ-009**: WHEN parsing a migration succeeds, the handler SHALL produce a normalized JSON representation of the schema changes and write it to `.moai/cache/db-sync/proposal.json`.
   - **Rationale**: 오케스트레이터와 스킬이 동일한 JSON 제안서를 소비해야 승인 플로우와 문서 갱신이 일관되게 작동한다.
10. **REQ-010**: WHEN `proposal.json` is written, the handler SHALL emit a JSON decision message to stdout signaling `"decision": "ask-user"` so the orchestrator presents an AskUserQuestion in the next turn.
    - **Rationale**: 서브에이전트는 사용자에게 직접 질문할 수 없으므로(Section 8 HARD), 명시적 decision 시그널로 오케스트레이터에 제어권을 위임한다.
11. **REQ-011**: IF parsing fails (unsupported syntax or corrupt file), THEN the handler SHALL log the error to `.moai/logs/db-sync-errors.log` and exit `0` (non-blocking).
    - **Rationale**: 파서 오류가 사용자의 `Write`/`Edit` 흐름을 차단하면 UX 저하가 크다. 에러는 기록하되 비차단으로 종료한다.

### User Approval Flow

12. **REQ-012**: WHEN the orchestrator sees the `db-sync` decision signal, it SHALL call AskUserQuestion with 3 options: "Apply proposed schema update", "Review diff first", "Skip this time".
    - **Rationale**: 스키마 문서는 프로젝트 산출물이므로 자동 덮어쓰기는 금지한다. 3-옵션 UX는 즉시 적용/검토 후 적용/생략의 실제 의사결정 경로를 모두 포함한다.
13. **REQ-013**: WHEN the user selects `Apply`, the orchestrator SHALL invoke the `moai-domain-db-docs` skill to rewrite `schema.md` / `erd.mmd` / `migrations.md` based on `proposal.json`.
    - **Rationale**: 문서 갱신은 파싱과 독립된 관심사이므로 전용 스킬이 담당해야 테스트와 진화가 용이하다.
14. **REQ-014**: WHEN the user selects `Review`, the orchestrator SHALL display the diff and then re-ask the `Apply`/`Skip` question.
    - **Rationale**: diff 확인 후 결정할 수 있는 경로가 없으면 사용자는 보수적으로 `Skip`을 선택하게 되어 동기화율이 하락한다.
15. **REQ-015**: WHEN the user selects `Skip`, the orchestrator SHALL delete `proposal.json` and take no action.
    - **Rationale**: 잔존 제안서는 다음 턴 오케스트레이션을 혼란시키므로 즉시 제거한다.

### moai-domain-db-docs Skill

16. **REQ-016**: WHEN `moai-domain-db-docs` is invoked with `proposal.json`, it SHALL update `.moai/project/db/schema.md` in-place preserving unmodified sections and existing `_TBD_` markers.
    - **Rationale**: 사용자 손필기 주석과 `_TBD_` placeholder는 팀 커뮤니케이션 산출물이므로 자동 덮어쓰기에서 보존되어야 한다.
17. **REQ-017**: WHEN `erd.mmd` is regenerated, the skill SHALL preserve the comment header and produce a valid Mermaid `erDiagram` body.
    - **Rationale**: 유효하지 않은 Mermaid는 docs-site 빌드 실패를 유발하므로 스펙 준수가 필수이다.
18. **REQ-018**: WHEN `migrations.md` is updated, the skill SHALL append a new entry to the `## Applied Migrations` table with `filename`, `applied_at` (ISO-8601), `checksum` (SHA-256 of the migration file), and `up_summary` (1-line human-readable).
    - **Rationale**: 마이그레이션 이력은 순서·체크섬·타임스탬프가 있어야 감사와 롤백 근거로 사용 가능하다.
19. **REQ-019**: WHEN the skill completes, it SHALL NOT trigger recursive hook invocation because `.moai/project/db/**` is in the Excluded Patterns section.
    - **Rationale**: 훅 레벨 차단(REQ-004)만으로 재귀가 해소되지 않는 경우(예: glob 오설정)에 대비한 스킬 레벨의 이중 보호이다.

### Drift Verification (`/moai db verify`)

20. **REQ-020**: WHEN `/moai db verify` is invoked, the skill SHALL compute the expected `schema.md` content from current migration files and compare it to the existing `schema.md`.
    - **Rationale**: drift는 "문서가 마이그레이션 실체와 다른가"로 정의되므로, 기댓값을 재계산해서 실물과 대조하는 방식이 유일하게 정확하다.
21. **REQ-021**: WHEN drift is detected, `/moai db verify` SHALL exit with code `1` and print a unified diff to stdout.
    - **Rationale**: CI에서 drift를 실패로 판정하려면 POSIX exit code 기반 시그널이 필요하고, 사람은 unified diff로 즉시 맥락을 파악한다.
22. **REQ-022**: WHEN no drift is detected, `/moai db verify` SHALL exit `0` and print `"Schema documentation is in sync"`.
    - **Rationale**: 성공 시에도 사용자에게 명시적 확인 메시지를 제공해 "정말로 검사가 실행됐는가" 의심을 제거한다.

### `/moai db refresh`

23. **REQ-023**: WHEN `/moai db refresh` is invoked, the skill SHALL rescan ALL migration files (not just the last changed one) and rebuild `schema.md` / `erd.mmd` / `migrations.md` from scratch.
    - **Rationale**: 증분 동기화 경로가 누락한 과거 마이그레이션을 복구하기 위한 "진실의 재구성" 모드이다.
24. **REQ-024**: WHEN `/moai db refresh` runs, the skill SHALL prompt AskUserQuestion `"Confirm full rebuild?"` with options `Apply` / `Cancel` before touching any file.
    - **Rationale**: 전체 재구성은 사용자 편집을 덮어쓸 수 있는 파괴적 연산이므로 명시적 확인 없이 실행해서는 안 된다.

## Acceptance Criteria

- **AC-1**: `internal/template/templates/.claude/hooks/moai/handle-db-schema-change.sh` exists, is executable (`+x`), and is under 30 lines.
- **AC-2**: `internal/template/templates/.claude/settings.json.tmpl` contains a PostToolUse matcher entry referencing `handle-db-schema-change.sh` with `timeout: 30000`.
- **AC-3**: Go CLI accepts `moai hook db-schema-sync --file <path>` and returns exit `0` on debounced invocations.
- **AC-4**: `internal/template/templates/.claude/skills/moai-domain-db-docs/SKILL.md` exists and its YAML frontmatter matches the Skill Frontmatter Template defined in this document, including the required top-level fields `name`, `description`, `license`, `compatibility`, `allowed-tools`, `user-invocable`, all `metadata.*` keys (`version`, `category`, `status`, `updated`, `tags`), and all `triggers.*` keys (`keywords`, `agents`, `phases`).
- **AC-5**: Editing `prisma/schema.prisma` in a sample project triggers the hook within 5 seconds and produces `.moai/cache/db-sync/proposal.json`.
- **AC-6**: The user approval flow shows 3 options (`Apply`, `Review`, `Skip`) and correctly honors each choice.
- **AC-7**: After `Apply`, `schema.md` contains updated table definitions matching the Prisma schema AND `.moai/project/db/**` is NOT re-triggered by the hook.
- **AC-8**: `/moai db verify` exits `1` and prints a unified diff when `schema.md` is out of sync.
- **AC-9**: `/moai db refresh` rebuilds all 3 docs (`schema.md`, `erd.mmd`, `migrations.md`) after user confirmation.
- **AC-10**: Debounce: editing the same migration file twice within 10 seconds triggers only 1 `proposal.json` creation.

## Skill Frontmatter Template

```yaml
---
name: moai-domain-db-docs
description: >
  Parses DB migration files (Prisma, Alembic, Rails, raw SQL) and keeps
  .moai/project/db/schema.md, erd.mmd, migrations.md in sync. Powers the
  PostToolUse hook and /moai db refresh/verify subcommands.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TaskCreate, TaskUpdate
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-20"
  tags: "db, schema, migration, documentation, sync, drift"

triggers:
  keywords: ["db docs", "schema sync", "migration parse", "erd update"]
  agents: ["expert-backend"]
  phases: ["run", "sync"]
---
```

## Hook Script Skeleton

```bash
#!/bin/bash
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
[ -z "$FILE_PATH" ] && exit 0
moai hook db-schema-sync --file "$FILE_PATH" 2>>/tmp/moai-db-sync.log
exit 0
```

## Scope

### IN SCOPE

- PostToolUse 훅 인프라 (bash wrapper + settings.json.tmpl 매처 엔트리)
- Go hook subcommand (`moai hook db-schema-sync`) — 파싱 호출, proposal.json 생성, debounce 처리, 에러 로깅
- `moai-domain-db-docs` 스킬 — schema.md / erd.mmd / migrations.md 갱신 로직
- Drift detection 알고리즘 (`/moai db verify`)
- Full rebuild 로직 (`/moai db refresh`)
- 사용자 승인 플로우 (3-option AskUserQuestion)

### OUT OF SCOPE

- 실제 마이그레이션 파서 구현체 (`internal/db/parser/` 하위 별도 모듈, 본 SPEC은 동작 요구사항만 선언)
- `/moai db` 명령 라우팅 (SPEC-DB-CMD-001 책임)
- 정적 템플릿 콘텐츠 (SPEC-DB-TEMPLATES-001 책임)

## Risks

- **R-1 (훅 재귀 루프)**: recursion guard 설정 오류 시 `.moai/project/db/**` 쓰기가 다시 훅을 트리거할 위험. REQ-004(훅 레벨 exclude) + REQ-019(스킬 레벨 인식)로 이중 보호.
- **R-2 (파서 커버리지 공백)**: 비주류 프레임워크 또는 변형된 SQL dialect는 파싱 실패 가능. REQ-011에 따라 에러는 `.moai/logs/db-sync-errors.log`에 기록하고 non-blocking(exit 0)으로 사용자 경험을 해치지 않는다.
- **R-3 (대규모 diff로 인한 UX 저하)**: 대규모 마이그레이션의 diff가 AskUserQuestion 표시 한도를 초과할 경우, 오케스트레이터는 diff를 100줄로 truncate하고 `"see proposal.json for full diff"` 안내를 덧붙인다.

## Exclusions (What NOT to Build)

- 실제 파서 구현 (Prisma, Alembic, Rails, SQL) — 별도 모듈(`internal/db/parser/`)에서 다룸
- `/moai db` 최상위 명령 라우팅 — SPEC-DB-CMD-001 책임
- 정적 템플릿 파일 콘텐츠 정의 — SPEC-DB-TEMPLATES-001 책임
- DB 마이그레이션 실행 자체 (순수 문서 동기화만, 런타임 DB 변경 없음)
- ERD 시각화 렌더링(훅은 `.mmd` 텍스트만 생성하며 이미지 렌더링은 하지 않음)

## Traceability

- **REQ → AC 매핑**:
  - REQ-001, REQ-005 → AC-1, AC-2 (훅 스크립트 파일·매처 등록·호출 경로)
  - REQ-002 → AC-2 (stdin JSON 소비 계약은 매처 설정과 짝을 이룸)
  - REQ-003, REQ-004 → AC-7 (재귀 차단 및 비대상 파일 무시 검증)
  - REQ-006, REQ-007 → AC-10 (디바운스 윈도우·상태 파일 동작)
  - REQ-008, REQ-009, REQ-010, REQ-011 → AC-3, AC-5 (파싱·JSON 산출·에러 비차단)
  - REQ-012, REQ-013, REQ-014, REQ-015 → AC-6 (3-옵션 승인 플로우)
  - REQ-016, REQ-017, REQ-018, REQ-019 → AC-4, AC-7 (스킬 프런트매터·문서 갱신·재귀 방지)
  - REQ-020, REQ-021, REQ-022 → AC-8 (drift 검증·exit code·메시지)
  - REQ-023, REQ-024 → AC-9 (전체 재구성·사용자 확인)
