---
id: SPEC-DB-SYNC-HARDEN-001
version: 0.2.1
status: draft
created_at: 2026-04-21
updated_at: 2026-04-21
author: moai-adk-go
priority: medium
labels: [db, hook, hardening, follow-up, debounce, atomicity, windows-compat, coverage, mx-tag]
issue_number: null
depends_on: [SPEC-DB-SYNC-001]
related_specs: [SPEC-DB-CMD-001, SPEC-DB-TEMPLATES-001]
---

# SPEC-DB-SYNC-HARDEN-001: dbsync 훅 견고화 (5 warning 통합)

## HISTORY

- 2026-04-21 v0.2.1: `/moai run` 실행 중 발견된 REQ-H3-002/AC-6 Windows 리터럴 오작성 정정. `settings.json.tmpl` 내 다른 16개 Windows 엔트리는 전부 `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-*.sh"` (bash 스타일 변수 + 슬래시) 패턴이며, `%CLAUDE_PROJECT_DIR%` 백슬래시 경로는 bash 내부에서 확장되지 않아 실제로 작동하지 않는다. REQ-H3-001("기존 엔트리와 동일 패턴")이 정확한 의도이며, REQ-H3-002/AC-6의 리터럴을 기존 컨벤션에 맞춰 정정. REQ/AC 개수 불변.
- 2026-04-21 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. F-1~F-5 blocking defects 해결(AC-9 multiline grep, REQ-H5-003 제거, HandleDBSchemaSync 5번째 대상 추가, CheckDebounce 실제 signature 반영, spec-compact에 Exclusions 추가). W-1~W-6 warnings 전부 반영. REQ 총 15 → 14 (REQ-H5-003 제거), AC 10 유지.
- 2026-04-21 v0.1.0: SPEC 최초 작성. SPEC-DB-SYNC-001 (commit `e22eb718d`) 병합 이후 발견된 5개 Warning-level 코드 리뷰 지적사항을 단일 SPEC으로 통합. Critical 수정(c6985e2fe / aa29a9316 / 8a4022c69)은 이미 적용 완료된 상태에서 잔여 견고화 항목을 다룸.

## Background

SPEC-DB-SYNC-001은 PostToolUse 훅, `moai-domain-db-docs` 스킬, `/moai db refresh`·`/moai db verify` 구현을 정의하여 마이그레이션 파일 편집 시 `.moai/project/db/**` 문서를 자동 동기화한다. 머지(commit `e22eb718d`) 이후 즉각적인 Critical 이슈(c6985e2fe / aa29a9316 / 8a4022c69)는 후속 commit으로 처리되었으나, 코드 리뷰에서 식별된 5개의 Warning-level 지적사항은 별도 견고화 SPEC으로 분리되어 본 SPEC이 그 통합 산출물이다.

본 SPEC의 목적은 **기능 추가가 아닌 견고화**이다. 각 항목은 기존 SPEC-DB-SYNC-001의 계약을 위반하지 않으면서 신뢰성·이식성·관측성을 향상시킨다.

### Scope Boundary

- SPEC-DB-SYNC-001과의 관계: 본 SPEC은 SPEC-DB-SYNC-001의 REQ를 개정하지 않는다. SPEC-DB-SYNC-001 REQ는 외부 행동 계약이며 그대로 유효하다. 본 SPEC은 구현 내부의 견고성(메모리 상한, 동시성 원자성, 크로스 플랫폼, 테스트 커버리지, godoc/MX 계약)을 다룬다.
- 파서 모듈(`internal/db/parser/`)과의 관계: 파서 모듈은 본 SPEC에서 손대지 않는다. 파서 행동 계약은 SPEC-DB-SYNC-001 REQ-008에 이미 위임되어 있으며, 본 SPEC은 파서 호출자 측면(크기 가드)에만 영향을 미친다.

### 5 Warning Findings Summary

| ID | 영역 | 핵심 문제 |
|----|------|-----------|
| H1 | Performance/Security | `parseMigrationStub`이 파일 크기 상한 검사 없이 `os.ReadFile`을 호출 → 대용량 파일 시 메모리 압박 + `proposal.json` 비대화 |
| H2 | Concurrency | `CheckDebounce`의 `os.WriteFile`이 원자적이지 않음 → 10초 윈도우 내 동시 PostToolUse 이벤트에서 중복 알림 발생 가능 |
| H3 | Cross-platform | 최근 추가된 `db-schema-change` 훅 엔트리가 `settings.json.tmpl` 내 다른 PostToolUse/Stop/SessionStart/SubagentStop 엔트리들의 Windows 플랫폼 분기 패턴을 누락(그 엔트리들이 일관되게 `{{- if eq .Platform "windows"}}` 분기를 사용하는 것과 달리 db-schema-change만 예외) |
| H4 | Quality Gate | `internal/hook/dbsync` 패키지 테스트 커버리지 77.8%로 프로젝트 목표 85% 미달 |
| H5 | Observability | 5개 exported 함수의 MX 계약(`@MX:NOTE`) 존재 여부를 CI에서 강제 검증하는 가드 부재 → 리팩토링 시 동작 계약 망실 위험 (※ `HandleDBSchemaSync`는 commit `aa29a9316`에서 이미 `@MX:NOTE` + `@MX:ANCHOR` 부착됨. 본 SPEC의 H5는 **5 함수 전수 존재 검증과 향후 누락 방지**에 초점) |

## Requirements (EARS)

### H1 — parseMigrationStub 파일 크기 가드

1. **REQ-H1-001** (Ubiquitous): The `db_schema_sync.go` source file SHALL declare a package-level constant `maxMigrationFileSize` with the literal value `1 << 20` (1 MiB), AND `parseMigrationStub` SHALL be the sole reference site of this constant within the package.
   - **Rationale**: 단일 원천(single source of truth)으로 한계값을 고정하여, 테스트와 런타임이 동일 값을 공유하도록 보장한다. 1MB는 Prisma schema/Alembic version 파일의 실측 상한(수십 KB)보다 훨씬 크지만 악성·우발 입력에는 충분한 경계이다.

2. **REQ-H1-002** (Event-driven): WHEN `parseMigrationStub` is called AND the target file size (obtained by `os.Stat` before any full read) exceeds `maxMigrationFileSize`, THEN the function SHALL (a) NOT invoke `os.ReadFile` or any equivalent whole-file read for that call, (b) append exactly one log line to `ErrorLogFile` matching the literal prefix `parseMigrationStub: file exceeds maxMigrationFileSize=`, AND (c) return a non-error result whose `parsed_content` is the empty string `""` AND whose truncation marker field (e.g., `truncated`) is `true` (both set, not an either/or).
   - **Rationale**: SPEC-DB-SYNC-001 REQ-011에 따라 파서 오류는 non-blocking이어야 하므로, 크기 초과도 동일 원칙으로 처리한다. ErrorLogFile에 남겨야 사후 관측 가능하며, 비-에러 반환이어야 호출 측 오케스트레이터가 사용자 승인 플로우를 이어갈 수 있다. 양쪽 필드를 모두 설정해야 테스트가 결정적으로 PASS/FAIL을 판정할 수 있다.

3. **REQ-H1-003** (Ubiquitous): `parseMigrationStub` SHALL determine file size via `os.Stat` BEFORE any call to `os.ReadFile`, `io.ReadAll`, or equivalent whole-file read. Thus when the size-guard branch is taken (REQ-H1-002), no whole-file read has occurred.
   - **Rationale**: "먼저 전부 읽고 나서 크기 검사"는 메모리 보호 목적을 달성하지 못하므로, 읽기 **이전** 단계에서 크기 판정이 수행되어야 한다.

### H2 — CheckDebounce 원자성

4. **REQ-H2-001** (Ubiquitous): Every write to `Config.StateFile` performed from within `CheckDebounce` SHALL be atomic with respect to concurrent `CheckDebounce` callers targeting the same `StateFile` path. "Atomic" means: a concurrent reader either observes the pre-write contents in full or the post-write contents in full, and never a partial/torn state.
   - **Implementation note (non-normative, belongs to plan.md)**: Preferred realization is temp-file + `os.Rename` in the same directory (POSIX rename atomicity). An alternative `flock`-based advisory lock is permitted but discouraged due to NFS/WSL weakness (see R-1).
   - **Rationale**: 특정 구현(`os.Rename` vs `flock`)을 REQ 텍스트에 박제하면 향후 구현이 제약된다. REQ는 atomicity 보장만 규정하고, 수단 선택은 plan.md에서 관리한다.

5. **REQ-H2-002** (State-driven): WHILE two or more goroutines concurrently call `CheckDebounce(stateFile, filePath, window)` with identical `stateFile` and `filePath` arguments within the same `window`, THE system SHALL ensure that exactly one call returns `debounced=false` (the single "winner") AND all other concurrent calls return `debounced=true`.
   - **Rationale**: 동시 실행 시 "두 호출 모두 not-debounced"가 관측되면 `proposal.json`이 두 번 작성되어 사용자에게 중복 AskUserQuestion이 표시되는 근본 원인이 된다. 원자적 교체로 이 윈도우를 제거한다. REQ는 외부 관측 가능한 반환값 집합(`{false, true}`)을 규정하며, 구현 고루틴 순서는 무관하다.

6. **REQ-H2-003** (Unwanted behavior guard / IF-THEN): IF the atomic state-file write fails due to an I/O error (disk full, permission denied, ENOSPC, etc.), THEN `CheckDebounce` SHALL (a) append the error message to `ErrorLogFile`, AND (b) return `(true, nil)` — the "debounced=true" safe default — so the pipeline is not blocked.
   - **Rationale**: 상태 파일 쓰기 실패 시 "not-debounced"(`false`)를 반환하면 동일 이벤트가 영구히 작동하지 않으므로, 안전 측면에서 `debounced=true`(안전 기본값)로 처리해 다음 정상 실행을 기다린다. 단일 복구 정책을 선언하여 구현 분기와 테스트 assertion을 단순화한다.

### H3 — handle-db-schema-change.sh Windows 플랫폼 분기

7. **REQ-H3-001** (Ubiquitous): The `db-schema-change` PostToolUse hook entry registered in `internal/template/templates/.claude/settings.json.tmpl` SHALL use the same `{{- if eq .Platform "windows"}} … {{- else}} … {{- end}}` branching pattern that the other PostToolUse / Stop / SessionStart / SubagentStop hook entries in the same file already use consistently; the `db-schema-change` entry is currently the sole exception to this pattern and SHALL be brought into conformance.
   - **Rationale**: 일관성 회귀 방지. 최근 추가된 훅이 파일 내 다른 훅들과 다른 형식을 채택하면 Windows 사용자에게만 동기화가 작동하지 않아 회귀 테스트에서 누락되기 쉽다. 특정 엔트리 개수 대신 "기존 엔트리는 이 패턴을 사용한다"라는 구조적 사실을 참조한다.

8. **REQ-H3-002** (Event-driven): WHEN `settings.json.tmpl` is rendered with `Platform=windows`, THEN the `db-schema-change` hook `command` field SHALL be exactly the string `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` — identical to the Windows branch pattern used by the other 16 hook entries in the same file (handle-session-start/handle-compact/handle-session-end/handle-pre-tool/handle-post-tool/handle-stop/handle-subagent-stop/handle-post-tool-failure/etc.). The `bash ` prefix is required so that Git Bash / WSL executes the `.sh` wrapper, and `$CLAUDE_PROJECT_DIR` is expanded by bash at invocation time (NOT `%CLAUDE_PROJECT_DIR%` cmd.exe-style which does not expand inside a bash-quoted argument).
   - **Rationale**: Windows에서 `.sh` 파일을 직접 실행할 수 없으므로 `bash` 접두사가 필수이며, Git Bash/WSL가 실제로 변수를 전개할 수 있도록 `$CLAUDE_PROJECT_DIR` (bash 스타일)을 사용한다. `%CLAUDE_PROJECT_DIR%`는 cmd.exe 전개 문법이므로 bash-quoted 인자 내부에서는 확장되지 않아 경로가 리터럴 문자열로 넘어가는 버그가 된다. v0.2.0 초안은 이 점을 누락했다.

9. **REQ-H3-003** (Event-driven): WHEN `settings.json.tmpl` is rendered with `Platform=darwin` OR `Platform=linux`, THEN the `db-schema-change` hook `command` field SHALL be exactly the string `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` (double quotes included) — identical to the SPEC-DB-SYNC-001 AC-2 baseline, with no `bash ` prefix.
   - **Rationale**: macOS/Linux 경로는 SPEC-DB-SYNC-001 REQ-001이 이미 규정했으며 회귀가 있어서는 안 된다.

### H4 — dbsync 테스트 커버리지 77.8% → 85%

10. **REQ-H4-001** (Event-driven): WHEN `go test -coverprofile=cover.out ./internal/hook/dbsync/...` is executed in CI or locally, THEN the aggregated coverage (the `total:` row reported by `go tool cover -func=cover.out`) SHALL be greater than or equal to `85.0%`.
    - **Rationale**: CLAUDE.md Quality Gates는 신규 코드에 85% 커버리지를 요구하며, 본 항목이 패키지 수준에서 이를 강제한다.

11. **REQ-H4-002** (Ubiquitous): The test file `internal/hook/dbsync/db_schema_sync_test.go` SHALL contain table-driven test cases covering the following boundary scenarios, each identified by the literal case names required in AC-8: (a) for `parseMigrationStub` — `empty_file`, `utf8_bom`, `oversized`, `nonexistent`; (b) for `matchGlob` — `trailing_slash`, `double_star_only`, `unicode_path`; (c) for `CheckDebounce` corrupt-state recovery — `corrupt_state_recovery`, whose recovery behavior SHALL be the safe default `debounced=true` (aligning with REQ-H2-003), never silent state regeneration without logging.
    - **Rationale**: 커버리지는 숫자 달성이 목표가 아니라 실제 경계 동작을 검증해야 가치가 있다. 본 항목은 검증 대상 시나리오를 명시하여 "null 테스트로 % 끌어올리기" 안티패턴을 차단한다. 복구 정책은 REQ-H2-003과 일치하는 단일 결과(`debounced=true`)로 고정한다.

12. **REQ-H4-003** (Ubiquitous): All test cases added under REQ-H4-002 SHALL create temporary files exclusively via `t.TempDir()`. No test SHALL read from or write to any path outside the `t.TempDir()` tree, and no test SHALL invoke `t.Setenv("HOME", …)`.
    - **Rationale**: CLAUDE.local.md §6 Test Isolation 원칙과 일치. 테스트가 프로젝트 파일을 수정하면 다른 SPEC의 회귀 테스트가 오염된다.

### H5 — exported helper MX 계약 명시

13. **REQ-H5-001** (Ubiquitous): For each of the five exported functions in `internal/hook/dbsync/db_schema_sync.go` — `HandleDBSchemaSync`, `BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce` — the godoc comment block that immediately precedes the `func Name(` signature line SHALL contain at least one line whose content starts with `// @MX:NOTE` (the `// @MX:NOTE` marker may appear anywhere within the godoc block, not necessarily on the single line directly above `func`).
    - **Rationale**: MX 프로토콜(`.claude/rules/moai/workflow/mx-tag-protocol.md`)에 따르면 exported API 계약은 `@MX:NOTE`로 표기하여 리팩토링 시 암묵적 행동 변경을 탐지 가능한 형태로 보존해야 한다. "godoc 블록 내부 어딘가" 허용은 plan.md:L114-L131의 다라인 godoc 구조와 일치한다. commit `aa29a9316`에서 `HandleDBSchemaSync`는 이미 `@MX:NOTE` + `@MX:ANCHOR`가 부착되어 있으므로, H5는 **5 함수 전수 존재 verification** 및 **누락 방지 guardrail** 역할이다.

14. **REQ-H5-002** (Ubiquitous): Each `@MX:NOTE` godoc block identified in REQ-H5-001 SHALL include documentation for the three contract elements — (a) input parameters and their semantics, (b) return value semantics and error conditions, (c) side effects (file writes, log appends, external command invocations). The natural language used in the godoc prose SHALL follow the `code_comments` value declared in `.moai/config/sections/language.yaml` (currently `ko` — Korean). Technical identifiers (function/parameter/type names, literal values) remain in English regardless of this setting.
    - **Rationale**: "함수 이름만으로 계약 추론" 안티패턴을 차단. 세 요소가 갖춰져야 향후 리팩토링 시 계약 위반 여부를 리뷰어가 즉시 판단할 수 있다. 자연어 선택은 프로젝트 전역 설정에 위임하여 이중 원천(dual source of truth)을 제거한다.

## Acceptance Criteria

- **AC-1 (H1-a)**: `go test ./internal/hook/dbsync/... -run TestParseMigrationStub_OversizedFile -v` 실행 시, `t.TempDir()` 내 2MB 합성 파일 입력에 대해 테스트가 통과하며, 테스트는 다음을 전부 검증한다: (1) 함수가 에러 없이 반환(`err == nil`), (2) 반환된 결과의 `parsed_content == ""` 이며 동시에 `truncated == true` (양쪽 필드 모두 확정 값), (3) `ErrorLogFile`(테스트 내부에서 `t.TempDir()`로 리디렉트됨)에 `parseMigrationStub: file exceeds maxMigrationFileSize=` 접두사 문자열이 정확히 1줄 포함.
- **AC-2 (H1-b)**: `grep -n 'maxMigrationFileSize = 1 << 20' internal/hook/dbsync/db_schema_sync.go` 결과가 정확히 1행이며, `grep -c 'maxMigrationFileSize' internal/hook/dbsync/db_schema_sync.go` 결과가 `>= 2` (선언 1회 + `parseMigrationStub` 본문 참조 최소 1회).
- **AC-3 (H2-a)**: `go test -race ./internal/hook/dbsync/... -run TestCheckDebounceConcurrency -count=10` 실행 시, 10회 반복 모두 통과. 테스트는 (i) 짧은 `window`(예: `50*time.Millisecond`) 값을 사용하여 실제 `time.Now()` 기반 디바운스 윈도우를 재현하고, (ii) `sync.WaitGroup`으로 2개 고루틴을 동시에 기동하여 동일 `stateFile`·`filePath`에 대해 `CheckDebounce`를 호출하며, (iii) 반환된 두 `debounced` 값의 multiset이 정확히 `{false, true}`임을 asert하고, (iv) 종료 후 `stateFile`이 유효한 JSON으로 parse 가능함을 확인한다. 데이터 레이스 미검출 포함.
- **AC-4 (H2-b)**: Go 단위 테스트 `TestCheckDebounce_NoDirectWriteFile` (신규)가 존재하고 통과해야 한다. 이 테스트는 `go/parser`로 `db_schema_sync.go`의 AST를 읽고 `CheckDebounce` 함수 본문 내부의 모든 `*ast.CallExpr`를 순회하며, `os.WriteFile(...)` 직접 호출이 존재하지 않음을 assert한다 (변수 이름 rename에 대해 false-pass를 방지). 동시에 동일 AST에서 최소 1개 이상의 `os.Rename(...)` 또는 `flock(...)` 호출이 존재함을 확인한다.
- **AC-5 (H3-a)**: `grep -nA 10 'handle-db-schema-change.sh' internal/template/templates/.claude/settings.json.tmpl` 결과 블록 안에 `{{- if eq .Platform "windows"}}`, `{{- else}}`, `{{- end}}` 세 행이 모두 존재한다. 또한 동일 파일 내 다른 기존 PostToolUse/Stop/SessionStart/SubagentStop 엔트리 중 임의의 2개 이상이 동일한 세 토큰을 사용하고 있음을 grep으로 확인한다 (일관성 증거).
- **AC-6 (H3-b)**: Go 템플릿 렌더링 테스트 `TestRender_DbSchemaChangeHook_Windows` 및 `TestRender_DbSchemaChangeHook_Unix` (신규 또는 기존 확장) 추가 및 통과. Windows 렌더 결과(`Platform="windows"`)는 db-schema-change 훅 엔트리에서 `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 문자열(JSON escape: `bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh\"`)을 정확히 포함하며, Unix 렌더 결과(`Platform="darwin"` 또는 `"linux"`)는 `bash ` 접두사 없이 `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"`를 정확히 포함한다. 두 결과 모두 유효한 JSON이다.
- **AC-7 (H4-a)**: `go test -coverprofile=cover.out ./internal/hook/dbsync/... && go tool cover -func=cover.out | awk '/^total:/{print $3}' | sed 's/%//'` 명령 출력이 부동소수점으로 `>= 85.0`.
- **AC-8 (H4-b)**: `internal/hook/dbsync/db_schema_sync_test.go`에서 다음 8개 테이블 케이스 이름이 문자열 리터럴로 최소 1회씩 등장해야 한다: `empty_file`, `utf8_bom`, `oversized`, `nonexistent`, `trailing_slash`, `double_star_only`, `unicode_path`, `corrupt_state_recovery`. (`grep -c '"empty_file"' …` 등 각 케이스 이름별 확인.)
- **AC-9 (H5)**: 5개 함수 각각에 대해, 함수 시그니처 라인 `^func Name(` 바로 앞의 godoc 블록(대략 20줄 윈도우) 안에 `// @MX:NOTE` 마커가 최소 1줄 존재해야 한다. 검증 방법은 다음 awk 기반 multiline 스캔이며 5/5 함수에 대해 `OK`가 출력되어야 한다:

  ```bash
  awk '
    /^\/\/ @MX:NOTE/ { seen_mx=1 }
    /^func (HandleDBSchemaSync|BuildProposal|MatchesMigrationPattern|IsExcluded|CheckDebounce)\(/ {
      if (seen_mx) { print "OK " $2 } else { print "MISSING " $2; exit 1 }
      seen_mx=0
    }
    /^[[:space:]]*$/ { seen_mx=0 }   # blank line resets (godoc block boundary)
  ' internal/hook/dbsync/db_schema_sync.go
  ```

  또는 동등한 단순 형태: 각 함수별로 `grep -B 20 '^func Name(' internal/hook/dbsync/db_schema_sync.go | grep -c '^// @MX:NOTE'` 결과가 `>= 1`.
- **AC-10 (회귀 방어)**: SPEC-DB-SYNC-001의 기존 AC-1 ~ AC-10은 전부 유지되어야 하며, 본 SPEC 구현 PR은 `go test ./...` 전체 green을 달성해야 한다. 특히 AC-5(Prisma schema 편집 → 5초 내 `proposal.json` 생성) 재현 테스트가 본 SPEC 구현 이후에도 통과한다.

## Scope

### IN SCOPE

- `internal/hook/dbsync/db_schema_sync.go` 내부: 크기 상수 추가, `parseMigrationStub` 선검사 로직, `CheckDebounce` 원자적 교체, 5개 exported 함수에 MX 계약 godoc 존재 보장(4개 신규 + 1개 기존 verification)
- `internal/hook/dbsync/db_schema_sync_test.go` 확장: H1/H2/H4의 테이블 드리븐 테스트 케이스 추가, AC-4의 AST 기반 정적 검증 테스트 추가
- `internal/template/templates/.claude/settings.json.tmpl`: `db-schema-change` 훅 엔트리 한 곳에 대한 플랫폼 분기 추가(다른 훅 엔트리는 건드리지 않음)
- `internal/template/` 관련 렌더 테스트: Windows/Unix 렌더 결과 검증 케이스 추가

### OUT OF SCOPE

- 실제 Prisma/Alembic/Rails/SQL 파서 구현(SPEC-DB-SYNC-001 REQ-008 Out of Scope — `internal/db/parser/` 모듈로 위임)
- `/moai db` 신규 서브커맨드 또는 기존 서브커맨드의 행동 변경
- `handle-db-schema-change.sh` 자체 내용 변경(본 SPEC은 템플릿 등록 경로만 개정하며, 스크립트 본문은 SPEC-DB-SYNC-001 버전 그대로 유지)
- `settings.json.tmpl` 내 `db-schema-change` 외의 다른 훅/PostToolUse 엔트리 리팩토링
- `internal/cli/design_folder.go`의 TOCTOU 미세 윈도우(리뷰 listed as lower-priority Warning, 별도 SPEC 대상)
- Windows CI 러너 구축(템플릿 렌더 테스트는 Go `text/template` 순수 함수 호출이므로 크로스 플랫폼 러너 불요)
- 재귀 가드 알고리즘 변경(SPEC-DB-SYNC-001 REQ-004/REQ-019 유지)
- 사용자 승인 플로우 변경(SPEC-DB-SYNC-001 REQ-012~015 유지)
- `CheckDebounce` 시그니처 변경(현재 `func CheckDebounce(stateFile, filePath string, window time.Duration) (bool, error)`가 live code이며 본 SPEC에서 유지; clock injection을 별도 REQ로 도입하지 않음)

## Risks

- **R-1 (flock 이식성)**: `golang.org/x/sys/unix flock`은 NFS/WSL 환경에서 advisory-only로 약한 보장을 제공할 수 있음 → **완화**: plan.md에서 (a) 임시 파일 + `os.Rename` 경로를 기본으로 권장. `os.Rename`은 POSIX에서 동일 파일시스템 내 원자적이며 Go 표준 라이브러리만 사용 가능하여 이식성이 우수.
- **R-2 (커버리지 상승 중 선재 버그 발견)**: H4의 새 테이블 케이스가 기존 코드의 잠복 버그(예: UTF-8 BOM 처리 누락)를 드러낼 수 있음 → **완화**: 이를 **긍정적 결과**로 취급하고, 발견된 버그 수정은 본 SPEC 범위 내에서 처리. 발견 → 수정 → 테스트 통과의 한 사이클로 종결하며, 별도 SPEC으로 분기하지 않음.
- **R-3 (플랫폼 분기 회귀 리스크)**: H3 수정이 macOS/Linux의 기존 동작을 변경해서는 안 됨 → **완화**: AC-6에서 Windows와 Unix 양 방향의 렌더 결과를 모두 검증. Unix 렌더 결과가 SPEC-DB-SYNC-001 AC-2 기존 문자열과 정확히 일치하도록 diff 기반 단정(assertion) 포함.
- **R-4 (MX 주석 drift)**: H5의 `@MX:NOTE` 블록과 실제 함수 시그니처/부작용 계약이 시간이 지나며 표류할 수 있음 → **완화**: 5 함수 중 하나의 시그니처(파라미터/반환 타입) 또는 부작용 계약(쓰기 경로, 외부 호출 추가 등)을 변경하는 PR은 **동일 PR 내에서** 해당 함수의 `@MX:NOTE` 블록을 동반 갱신해야 한다. 이 규칙은 SPEC 레벨 REQ가 아닌 리뷰어 체크리스트(v0.1.0의 REQ-H5-003이 해당 내용이었으나 자동 검증 불가로 Risk mitigation으로 이동). PR 리뷰 템플릿에 "H5 drift check" 항목을 추가하여 인적 게이트 강화.
- **R-5 (대기열 상태 파일 손상)**: `os.Rename` 실패 또는 부분 쓰기 후 재부팅 시 상태 파일이 잘린 JSON으로 남을 수 있음 → **완화**: H4의 `corrupt_state_recovery` 테스트 케이스(REQ-H4-002)가 이 경로를 커버하며, REQ-H2-003이 안전 기본값(debounced=true) 반환을 규정.

## Traceability

### REQ ↔ AC Matrix

| REQ | AC | 검증 초점 |
|-----|------|-----------|
| REQ-H1-001 | AC-2 | `maxMigrationFileSize` 상수 선언 + 참조 (2회 이상) |
| REQ-H1-002 | AC-1 | 크기 초과 시 로그 기록 + `parsed_content=""` AND `truncated=true` 동시 설정 |
| REQ-H1-003 | AC-1, AC-2 | `os.Stat` 선검사 → `os.ReadFile` 미호출 (메모리 보호 목적 달성) |
| REQ-H2-001 | AC-3, AC-4 | atomicity 관찰 가능(AC-3) + AST로 직접 `os.WriteFile` 부재 검증(AC-4) |
| REQ-H2-002 | AC-3 | 동시 호출 2-goroutine multiset = `{false, true}` |
| REQ-H2-003 | AC-3 (I/O 실패 variant), AC-8(`corrupt_state_recovery` 케이스) | I/O 실패 시 `(true, nil)` 반환 |
| REQ-H3-001 | AC-5 | 플랫폼 분기 3-token + 다른 엔트리 일관성 증거 |
| REQ-H3-002 | AC-6 | Windows 렌더: `bash "$CLAUDE_PROJECT_DIR/..."` (기존 16 엔트리 컨벤션) |
| REQ-H3-003 | AC-6, AC-10 | Unix 렌더: SPEC-DB-SYNC-001 AC-2 baseline 회귀 없음 |
| REQ-H4-001 | AC-7 | 패키지 커버리지 ≥ 85.0% |
| REQ-H4-002 | AC-8 | 8개 지정 테이블 케이스 이름 존재 |
| REQ-H4-003 | AC-3, AC-8 | 테스트 격리(`t.TempDir()` 사용, `t.Setenv("HOME", …)` 금지) |
| REQ-H5-001 | AC-9 | 5개 함수 godoc 블록 내 `@MX:NOTE` 최소 1회 |
| REQ-H5-002 | AC-9 (리뷰 기반) + `code_comments` 설정 준수 | 입력/출력/부작용 3요소 + 한국어 자연어 |

14개 REQ, 10개 AC. 모든 REQ가 최소 1개 AC로 검증. (v0.1.0의 REQ-H5-003은 자동 검증 불가하여 R-4 Risk mitigation으로 이동.)

### 5 Warning → REQ 그룹 매핑

| Warning | REQ Family | AC 주요 지점 |
|---------|------------|-------------|
| H1 (file size guard) | REQ-H1-001 ~ REQ-H1-003 | AC-1, AC-2 |
| H2 (CheckDebounce atomicity) | REQ-H2-001 ~ REQ-H2-003 | AC-3, AC-4 |
| H3 (Windows platform branch) | REQ-H3-001 ~ REQ-H3-003 | AC-5, AC-6 |
| H4 (coverage 77.8% → 85%) | REQ-H4-001 ~ REQ-H4-003 | AC-7, AC-8 |
| H5 (MX tag annotations) | REQ-H5-001 ~ REQ-H5-002 | AC-9 |
| — (회귀 방어) | — | AC-10 |

모든 5개 Warning 항목이 REQ 패밀리로 매핑되고, 매 REQ가 최소 1개 AC에 연결되며, 회귀 방어는 AC-10으로 별도 보장됨.

## Exclusions (What NOT to Build)

- 실제 마이그레이션 파서 구현(Prisma/Alembic/Rails/SQL) — SPEC-DB-SYNC-001 REQ-008이 `internal/db/parser/` 별도 모듈로 위임함. 본 SPEC은 파서 호출자(크기 가드)만 다룸.
- `/moai db` 신규 서브커맨드 또는 플래그 추가
- `handle-db-schema-change.sh` 스크립트 본문 변경 — 본 SPEC은 `settings.json.tmpl`의 훅 등록 경로만 수정
- `settings.json.tmpl` 내 다른 훅 엔트리(db-schema-change 외) 수정
- Windows 전용 CI 러너 구축 — H3 검증은 Go `text/template` 렌더링 단위 테스트로 충분
- `internal/cli/design_folder.go`의 TOCTOU 마이크로 윈도우 — 별도 하위-우선 Warning 대상, 본 SPEC에서 다루지 않음
- 재귀 가드 알고리즘(`.moai/project/db/**` 제외)의 변경 — SPEC-DB-SYNC-001 REQ-004/REQ-019 계약 유지
- 사용자 승인 플로우(3-옵션 AskUserQuestion) 변경 — SPEC-DB-SYNC-001 REQ-012~015 계약 유지
- 문서 갱신 로직(`schema.md` / `erd.mmd` / `migrations.md` 재작성) 변경 — SPEC-DB-SYNC-001 REQ-016~019 계약 유지
- `CheckDebounce` 시그니처 변경 — 현재 `(stateFile, filePath string, window time.Duration) (bool, error)`를 유지하며 `now time.Time` 클럭 주입 등 신규 파라미터 도입 없음

## Target Files (scope discipline)

구현 PR이 수정할 수 있는 파일은 다음 4개로 한정된다:

1. `internal/hook/dbsync/db_schema_sync.go` — H1(크기 상수/선검사), H2(원자적 교체), H5(MX 계약 godoc 4개 신규 + 1개 기존 verification)
2. `internal/hook/dbsync/db_schema_sync_test.go` — H1/H2/H4의 테이블 드리븐 테스트 케이스, AC-4 AST 기반 테스트
3. `internal/template/templates/.claude/settings.json.tmpl` — H3 플랫폼 분기(db-schema-change 훅 엔트리 한 블록만)
4. `internal/template/` 하위 렌더 테스트 파일(기존 또는 신규) — AC-6 Windows/Unix 렌더 검증

외부 파일 변경은 금지(Agent Core Behavior #5 Scope Discipline 준수). 특히 `handle-db-schema-change.sh` 스크립트 본문은 건드리지 않는다.

## Configuration Reference

- SPEC-DB-SYNC-001 (`.moai/specs/SPEC-DB-SYNC-001/spec.md`) — 외부 계약 기준
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — `@MX:NOTE` 표기 규칙(H5)
- `CLAUDE.md` Quality Gates — 85% 커버리지 목표(H4)
- `CLAUDE.local.md` §6 — `t.TempDir()` 테스트 격리 원칙(H4)
- `.moai/config/sections/quality.yaml` — harness level 결정; 본 SPEC은 `standard` 사용(plan-auditor 리뷰 활성)
- `.moai/config/sections/language.yaml` — `code_comments: ko` (REQ-H5-002 자연어 선택 기준)
