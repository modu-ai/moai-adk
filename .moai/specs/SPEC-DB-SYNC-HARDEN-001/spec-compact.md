---
id: SPEC-DB-SYNC-HARDEN-001
document: spec-compact
version: 0.2.1
created_at: 2026-04-21
updated_at: 2026-04-21
source: spec.md (Requirements + Acceptance Criteria + Exclusions sections)
---

# SPEC-DB-SYNC-HARDEN-001: Compact View

본 문서는 spec.md의 Requirements (EARS), Acceptance Criteria, Exclusions 세 섹션만 발췌한 컴팩트 뷰이다. 자동 생성된 참고용이며, 계약의 단일 원천은 spec.md이다.

## Requirements (EARS)

### H1 — parseMigrationStub 파일 크기 가드

- **REQ-H1-001** (Ubiquitous): `db_schema_sync.go`는 `const maxMigrationFileSize = 1 << 20` (1 MiB)을 패키지 상수로 선언하며, `parseMigrationStub`가 동일 파일 내 유일한 참조 site이다.
- **REQ-H1-002** (Event-driven): WHEN `parseMigrationStub`가 호출되고 `os.Stat` 선검사로 측정된 파일 크기가 `maxMigrationFileSize`를 초과할 때, THEN (a) `os.ReadFile` 등 전체 읽기를 호출하지 않고, (b) `ErrorLogFile`에 `parseMigrationStub: file exceeds maxMigrationFileSize=` 접두사 로그 1줄 기록, (c) `parsed_content=""` AND `truncated=true` 양쪽 필드를 확정 설정하여 non-error 반환.
- **REQ-H1-003** (Ubiquitous): `parseMigrationStub`는 `os.ReadFile`/`io.ReadAll` 등 전체 읽기 BEFORE `os.Stat`로 크기를 판정해야 한다. 크기 초과 경로에서는 전체 읽기가 발생하지 않는다.

### H2 — CheckDebounce 원자성 (시그니처 불변)

- **REQ-H2-001** (Ubiquitous): `CheckDebounce` 내부의 모든 `Config.StateFile` 쓰기는 동일 경로 대상 동시 `CheckDebounce` 호출자에 대해 원자적이어야 한다(부분/tear 상태 관측 불가). 구현 수단은 plan.md가 결정(권장: 임시 파일 + `os.Rename`).
- **REQ-H2-002** (State-driven): WHILE 둘 이상의 goroutine이 동일 `(stateFile, filePath, window)` 인수로 `CheckDebounce`를 동시에 호출하고 있을 때, 정확히 한 호출만 `debounced=false`를 반환하고 나머지는 `debounced=true`를 반환한다.
- **REQ-H2-003** (IF-THEN): IF 원자적 쓰기가 I/O 에러로 실패하면, (a) `ErrorLogFile`에 기록 후 (b) `(true, nil)` — `debounced=true` 안전 기본값 — 반환. 단일 복구 정책.

### H3 — handle-db-schema-change.sh Windows 플랫폼 분기

- **REQ-H3-001** (Ubiquitous): `settings.json.tmpl`의 `db-schema-change` PostToolUse 엔트리는 동일 파일 내 다른 PostToolUse/Stop/SessionStart/SubagentStop 엔트리들이 일관되게 사용하는 `{{- if eq .Platform "windows"}} … {{- else}} … {{- end}}` 패턴을 준수해야 한다. db-schema-change 엔트리가 현재 유일한 예외이다.
- **REQ-H3-002** (Event-driven): WHEN `Platform=windows`로 렌더될 때, command 필드는 정확히 `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 문자열(JSON 이스케이프 `bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh\"`) — 동일 파일 내 다른 16개 Windows 훅 엔트리와 동일 패턴.
- **REQ-H3-003** (Event-driven): WHEN `Platform=darwin` OR `Platform=linux`로 렌더될 때, command 필드는 정확히 `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 문자열(SPEC-DB-SYNC-001 AC-2 baseline)을 유지한다.

### H4 — dbsync 테스트 커버리지 77.8% → 85%

- **REQ-H4-001** (Event-driven): WHEN `go test -coverprofile=cover.out ./internal/hook/dbsync/...`가 실행될 때, `go tool cover -func=cover.out`의 `total:` 행이 `>= 85.0%`.
- **REQ-H4-002** (Ubiquitous): 테스트 파일은 다음 8개 명명 케이스를 포함: `empty_file`, `utf8_bom`, `oversized`, `nonexistent` (for `parseMigrationStub`); `trailing_slash`, `double_star_only`, `unicode_path` (for `matchGlob`); `corrupt_state_recovery` (for `CheckDebounce`, 안전 기본값 `(true, nil)` 복구 — REQ-H2-003과 일치).
- **REQ-H4-003** (Ubiquitous): 모든 추가 테스트 케이스는 `t.TempDir()`로만 임시 파일 생성. `t.Setenv("HOME", …)` 금지.

### H5 — exported helper MX 계약 명시 (5 함수 전수)

- **REQ-H5-001** (Ubiquitous): `db_schema_sync.go`의 5개 exported 함수 — `HandleDBSchemaSync`, `BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce` — 각각의 godoc 블록 안(함수 시그니처 직전 연속 주석 범위)에 `// @MX:NOTE`로 시작하는 줄이 최소 1개 존재해야 한다. `HandleDBSchemaSync`는 commit `aa29a9316`에서 이미 부착됨(verification only). 나머지 4개는 신규 부착.
- **REQ-H5-002** (Ubiquitous): 각 `@MX:NOTE` 블록은 입력 파라미터, 출력/반환값, 부작용의 3요소를 명시해야 한다. 산문은 `.moai/config/sections/language.yaml`의 `code_comments` 값(현재 `ko`)을 따른다. 기술 식별자는 영어 유지.

---

## Acceptance Criteria

- **AC-1 (H1-a)**: `TestParseMigrationStub_OversizedFile` 통과. 2MB 합성 파일 입력 → `err == nil`, `parsed_content == ""` AND `truncated == true` 양쪽 확정, `ErrorLogFile`에 `parseMigrationStub: file exceeds maxMigrationFileSize=` 접두사 1줄.
- **AC-2 (H1-b)**: `grep -n 'maxMigrationFileSize = 1 << 20' internal/hook/dbsync/db_schema_sync.go` = 1행. `grep -c 'maxMigrationFileSize' …` ≥ 2.
- **AC-3 (H2-a)**: `go test -race ./internal/hook/dbsync/... -run TestCheckDebounceConcurrency -count=10` 10회 모두 통과. `window=50ms` 짧은 실제 윈도우 + 2 goroutine + `sync.WaitGroup` → 반환 multiset `{false, true}`, 상태 파일 유효 JSON.
- **AC-4 (H2-b)**: `TestCheckDebounce_NoDirectWriteFile` (신규 Go 테스트) 통과. `go/parser` AST로 `CheckDebounce` 본문 순회하여 `os.WriteFile(...)` 직접 호출 0건 + `os.Rename(...)` 호출 ≥ 1건 검증 (변수 rename false-pass 방지).
- **AC-5 (H3-a)**: `grep -nA 10 'handle-db-schema-change.sh' internal/template/templates/.claude/settings.json.tmpl` 결과 블록에 `{{- if eq .Platform "windows"}}`, `{{- else}}`, `{{- end}}` 3행 존재. `grep -c '{{- if eq .Platform "windows"}}' …` ≥ 3 (다른 엔트리 일관성 증거).
- **AC-6 (H3-b)**: 렌더 테스트 `TestRender_DbSchemaChangeHook_Windows`(`bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 정확 포함, 기존 16 엔트리 컨벤션) 및 `TestRender_DbSchemaChangeHook_Unix`(`bash ` 접두사 없음 + `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 정확 포함) 통과.
- **AC-7 (H4-a)**: `go test -coverprofile=cover.out ./internal/hook/dbsync/... && go tool cover -func=cover.out | awk '/^total:/{print $3}' | sed 's/%//'` ≥ 85.0.
- **AC-8 (H4-b)**: 테스트 파일에 8개 케이스 이름 문자열 리터럴 존재: `empty_file`, `utf8_bom`, `oversized`, `nonexistent`, `trailing_slash`, `double_star_only`, `unicode_path`, `corrupt_state_recovery`.
- **AC-9 (H5)**: 5개 함수(`HandleDBSchemaSync`, `BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce`) 각각의 godoc 블록(시그니처 이전 ~20줄 윈도우) 안에 `// @MX:NOTE` 라인 ≥ 1. awk multiline 스캔 또는 `grep -B 20 '^func X(' … | grep -c '^// @MX:NOTE' >= 1` per function.
- **AC-10 (회귀 방어)**: SPEC-DB-SYNC-001 AC-1 ~ AC-10 전수 유지. 본 SPEC 구현 PR은 `go test ./...` 전체 green.

---

## Exclusions (What NOT to Build)

- 실제 마이그레이션 파서 구현(Prisma/Alembic/Rails/SQL) — SPEC-DB-SYNC-001 REQ-008이 `internal/db/parser/` 별도 모듈로 위임. 본 SPEC은 파서 호출자(크기 가드)만 다룸.
- `/moai db` 신규 서브커맨드 또는 플래그 추가.
- `handle-db-schema-change.sh` 스크립트 본문 변경 — 본 SPEC은 `settings.json.tmpl`의 훅 등록 경로만 수정.
- `settings.json.tmpl` 내 다른 훅 엔트리(db-schema-change 외) 수정.
- Windows 전용 CI 러너 구축 — H3 검증은 Go `text/template` 렌더링 단위 테스트로 충분.
- `internal/cli/design_folder.go`의 TOCTOU 마이크로 윈도우 — 별도 하위-우선 Warning 대상.
- 재귀 가드 알고리즘(`.moai/project/db/**` 제외) 변경 — SPEC-DB-SYNC-001 REQ-004/REQ-019 계약 유지.
- 사용자 승인 플로우(3-옵션 AskUserQuestion) 변경 — SPEC-DB-SYNC-001 REQ-012~015 계약 유지.
- 문서 갱신 로직(`schema.md`/`erd.mmd`/`migrations.md` 재작성) 변경 — SPEC-DB-SYNC-001 REQ-016~019 계약 유지.
- `CheckDebounce` 시그니처 변경 — 현재 `(stateFile, filePath string, window time.Duration) (bool, error)`를 유지하며 `now time.Time` 클럭 주입 등 신규 파라미터 도입 없음.

---

## REQ ↔ AC Traceability (14 REQs × 10 ACs)

| REQ | Primary AC | 검증 초점 |
|-----|-----------|----------|
| REQ-H1-001 | AC-2 | 상수 선언 + 참조 ≥ 2 |
| REQ-H1-002 | AC-1 | 로그 1줄 + `parsed_content=""` AND `truncated=true` |
| REQ-H1-003 | AC-1, AC-2 | `os.Stat` 선검사, 전체 읽기 부재 |
| REQ-H2-001 | AC-3, AC-4 | atomicity 관찰 + AST 검증 |
| REQ-H2-002 | AC-3 | 동시 호출 multiset `{false, true}` |
| REQ-H2-003 | AC-3 Edge, AC-8 | I/O 실패 시 `(true, nil)` 안전 기본값 |
| REQ-H3-001 | AC-5 | 3-token + 일관성 증거 |
| REQ-H3-002 | AC-6 | Windows 렌더 |
| REQ-H3-003 | AC-6, AC-10 | Unix 렌더 baseline |
| REQ-H4-001 | AC-7 | 커버리지 ≥ 85% |
| REQ-H4-002 | AC-8 | 8 케이스 이름 |
| REQ-H4-003 | AC-3, AC-8 | `t.TempDir()` 격리 |
| REQ-H5-001 | AC-9 | 5 함수 godoc 내 `@MX:NOTE` |
| REQ-H5-002 | AC-9 (리뷰) | 3요소 + 한국어 산문 |

14/14 REQs covered. Regression shield at AC-10.

### 5 Warning → REQ 그룹 매핑

| Warning | REQ Family | Primary AC | 상태 |
|---------|------------|-----------|------|
| H1 (file size guard) | REQ-H1-001..003 | AC-1, AC-2 | 커버됨 |
| H2 (CheckDebounce atomicity) | REQ-H2-001..003 | AC-3, AC-4 | 커버됨 |
| H3 (Windows platform branch) | REQ-H3-001..003 | AC-5, AC-6 | 커버됨 |
| H4 (coverage 77.8% → 85%) | REQ-H4-001..003 | AC-7, AC-8 | 커버됨 |
| H5 (MX tag annotations, 5 함수) | REQ-H5-001..002 | AC-9 | 커버됨 |
| Regression shield | — | AC-10 | 커버됨 |

5/5 Warning 항목이 REQ 패밀리로 매핑되고 AC로 검증 가능. 회귀 방어는 AC-10으로 별도 보장.
