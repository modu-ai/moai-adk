---
id: SPEC-DB-SYNC-HARDEN-001
document: plan
version: 0.2.0
created_at: 2026-04-21
updated_at: 2026-04-21
---

# SPEC-DB-SYNC-HARDEN-001: Implementation Plan

본 문서는 spec.md에 정의된 5개 견고화 항목(H1~H5)을 구현하기 위한 순차/병렬 작업 분해와 기술적 접근을 기술한다. 본 SPEC은 **기능 추가가 아닌 견고화**이므로 구현 위험이 낮지만, 회귀 방어(AC-10)가 최우선이다.

## Technical Approach

### 설계 원칙

- **계약 보존**: SPEC-DB-SYNC-001의 REQ 및 외부 계약은 불변. 본 SPEC은 구현 내부만 수정.
- **비침습**: 수정 범위는 4개 파일로 한정(spec.md Target Files 섹션).
- **테스트 우선 가능 항목**: H1/H2/H4는 테이블 드리븐 테스트를 먼저 작성하여 RED 상태 확인 후 GREEN으로 전환 가능. H3는 렌더링 결과 비교 테스트. H5는 godoc 추가(4개) + 기존 1개 verification이므로 테스트보다는 AST 기반/awk 기반 AC 검증이 주.
- **Signature 보존**: `CheckDebounce`의 현재 시그니처 `func CheckDebounce(stateFile, filePath string, window time.Duration) (bool, error)`를 그대로 유지한다. 시험 제어가 필요한 지점은 시그니처 변경이 아닌 **짧은 `window` 값**(예: `50*time.Millisecond`)과 실제 `time.Now()` 경합으로 동시성을 재현한다.
- **Atomicity 수단 선택**: spec.md REQ-H2-001은 atomicity 계약만 규정한다. 본 plan에서 수단은 (a) 임시 파일 + `os.Rename` 기본 채택. (b) `flock`은 R-1 이식성 약점으로 배제.

### 아키텍처 영향 범위

본 SPEC은 아키텍처를 바꾸지 않는다. 기존 데이터 흐름(PostToolUse 이벤트 → bash wrapper → `moai hook db-schema-sync` → 파서 → `proposal.json` → AskUserQuestion)은 그대로 유지되며, 각 단계에서 다음과 같은 내부 강화가 일어난다:

```
PostToolUse 이벤트
    ↓
handle-db-schema-change.sh (변경 없음)
    ↓  [H3: settings.json.tmpl에서의 등록 경로만 플랫폼 분기]
moai hook db-schema-sync
    ↓
HandleDBSchemaSync  [H5: @MX:NOTE 기존; verification]
    ↓
CheckDebounce  [H2: 원자적 교체, 시그니처 불변]
    ↓
parseMigrationStub  [H1: 크기 선검사]
    ↓
BuildProposal → proposal.json  [H5: @MX:NOTE 신규]
    ↓
(orchestrator decision 시그널)
```

5개 exported helper(H5 대상)는 본 플로우의 관측 가능 경계점(Observable Boundary)에 위치한다. `HandleDBSchemaSync`는 이미 `aa29a9316` 커밋에서 `@MX:NOTE` + `@MX:ANCHOR` 를 보유하므로, H5 작업은 나머지 4개(`BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce`)에 `@MX:NOTE` 신규 부착 + 5개 전수 존재 verification으로 구성된다.

## Task Decomposition

본 SPEC은 Priority 기반으로 분해하며 시간 추정은 포함하지 않는다(Agent Common Protocol §Time Estimation 준수). Priority는 회귀 위험과 의존성에 따라 High/Medium/Low로 구분한다.

### Milestone M1 (Priority: High) — 회귀 방어 기반 구축

**목표**: 기존 동작을 보존하면서 본 SPEC 수정이 안전하게 진행될 수 있는 테스트 베드 확보.

- **Task M1.1**: 현재 `internal/hook/dbsync` 패키지의 베이스라인 커버리지를 측정하여 기록(`go test -coverprofile=base.out ./internal/hook/dbsync/...`). 이후 H4 진행 시 비교 기준으로 사용.
- **Task M1.2**: SPEC-DB-SYNC-001 AC-5(Prisma schema 편집 → 5초 내 `proposal.json`) 재현이 가능한 회귀 테스트가 현재 스위트에 존재하는지 확인. 없다면 본 SPEC 구현 전에 추가(Target: `internal/hook/dbsync/db_schema_sync_test.go`).
- **Task M1.3**: 현재 `settings.json.tmpl` 내 PostToolUse/Stop/SessionStart/SubagentStop 엔트리 중 `{{- if eq .Platform "windows"}}` 분기를 사용하는 것을 3~5개 샘플링하여 Windows/Unix 플랫폼 분기 구조를 정확히 카피할 기준 블록으로 확보(H3 복제 원본). 이 샘플은 AC-5의 "다른 엔트리 일관성 증거" 그렙 대상이기도 하다.

### Milestone M2 (Priority: High) — H1 (parseMigrationStub 크기 가드)

**목표**: 대용량 마이그레이션 파일로 인한 메모리 압박 제거.

- **Task M2.1 [RED]**: `TestParseMigrationStub_OversizedFile` 테이블 드리븐 테스트 추가. `t.TempDir()` 내부에 2MB 합성 파일을 생성하고 `parseMigrationStub` 호출. 현 구현은 전체를 읽으므로 RED 확인. 테스트는 반환된 결과에서 `parsed_content == ""` AND `truncated == true`임을 검증.
- **Task M2.2 [GREEN]**: 패키지 상수 `maxMigrationFileSize = 1 << 20` 추가(REQ-H1-001). `parseMigrationStub` 진입점에서 `os.Stat(path)`로 크기 선검사 후 초과 시 조기 종료(REQ-H1-002/003). 로그 포맷: `"parseMigrationStub: file exceeds maxMigrationFileSize=%d path=%s"`. 반환 shape은 두 필드(`parsed_content=""`, `truncated=true`) 양쪽을 설정 — 결정적 assertion 가능.
- **Task M2.3 [REFACTOR]**: `ErrorLogFile` 경로를 테스트에서 `t.TempDir()`로 오버라이드할 수 있는지 확인. 기존 구조가 전역 경로면 테스트 훅으로 의존 주입(dependency injection) 최소 리팩토링(변수 `errorLogPath`를 패키지 변수로 추출하는 정도).
- **Task M2.4 [VERIFY]**: AC-1, AC-2 수동 실행 및 통과 확인.

### Milestone M3 (Priority: High) — H2 (CheckDebounce 원자성, 시그니처 불변)

**목표**: 동시 PostToolUse 이벤트에서 중복 `proposal.json` 생성 제거. **현재 시그니처 `func CheckDebounce(stateFile, filePath string, window time.Duration) (bool, error)` 유지**.

- **Task M3.1 [RED]**: `TestCheckDebounceConcurrency` 테스트 추가 (AC-3 명명 규칙). 핵심 구성:
  - `t.TempDir()` 내 `stateFile := filepath.Join(tmp, "last-seen.json")` 준비, `filePath := "prisma/schema.prisma"` 고정.
  - `window := 50 * time.Millisecond` (짧은 실제 윈도우).
  - `sync.WaitGroup`으로 2개 goroutine을 기동하여 동시에 `CheckDebounce(stateFile, filePath, window)`를 호출.
  - 각 goroutine의 `(debounced, err)`를 채널로 수집.
  - Assert: (i) 두 `err`가 모두 `nil`, (ii) 두 `debounced`의 multiset이 정확히 `{false, true}`, (iii) 종료 후 `stateFile`이 유효한 JSON.
  - 현재 구현(`os.WriteFile` 직접 호출)은 `go test -race`에서 데이터 레이스 또는 두 goroutine 모두 `false` 관측으로 RED.
  - `-count=10`에서 10회 모두 그린이어야 GREEN.
- **Task M3.2 [GREEN]**: `CheckDebounce` 내 상태 파일 갱신을 임시 파일 + `os.Rename` 패턴으로 교체(시그니처는 불변). 구현 스케치:
  ```
  // inside CheckDebounce, after computing `not debounced`:
  newState := DebounceState{FilePath: filePath, Timestamp: time.Now()}
  stateJSON, err := json.Marshal(newState)
  if err != nil { return true, nil }  // safe default per REQ-H2-003

  tmpFile, err := os.CreateTemp(filepath.Dir(stateFile), ".last-seen-*.json.tmp")
  if err != nil {
      logError(errorLogPath, "CheckDebounce: create temp: "+err.Error())
      return true, nil  // REQ-H2-003
  }
  renamed := false
  defer func() { if !renamed { _ = os.Remove(tmpFile.Name()) } }()

  if _, werr := tmpFile.Write(stateJSON); werr != nil { /* log + return true,nil */ }
  if cerr := tmpFile.Close(); cerr != nil { /* log + return true,nil */ }
  if rerr := os.Rename(tmpFile.Name(), stateFile); rerr != nil {
      logError(errorLogPath, "CheckDebounce: rename: "+rerr.Error())
      return true, nil
  }
  renamed = true
  return false, nil
  ```
  (파일시스템이 상이하면 `os.Rename`이 실패할 수 있으나 `stateFile` 디렉터리와 tmp 디렉터리가 동일하므로 안전.)
- **Task M3.3**: I/O 실패 경로(REQ-H2-003) 테스트 추가 — `AC-Edge-H2`에 해당. `stateFile`의 부모 디렉터리를 `os.Chmod(dir, 0444)`로 읽기 전용 설정 후 `CheckDebounce` 호출 → 반환값 `(true, nil)` 및 `ErrorLogFile` 한 줄 이상 기록 확인.
- **Task M3.4 [AC-4 AST 테스트]**: `TestCheckDebounce_NoDirectWriteFile` 추가. `go/parser.ParseFile`로 `db_schema_sync.go`를 AST로 파싱 → `CheckDebounce` 함수의 `*ast.FuncDecl`을 찾아 Body 내부를 `ast.Inspect`로 순회 → 모든 `*ast.CallExpr`에서 `Fun`이 `os.WriteFile`인 호출이 **0건**임을 assert. 동시에 동일 AST에서 `os.Rename` 호출 ≥ 1건 존재 확인. (변수 이름 rename에 대한 false-pass 방지 — W-6 해결.)
- **Task M3.5 [VERIFY]**: `go test -race ./internal/hook/dbsync/... -count=10 -run TestCheckDebounceConcurrency` 10회 반복 모두 통과. AC-3, AC-4 확인.

### Milestone M4 (Priority: Medium) — H3 (Windows 플랫폼 분기)

**목표**: Windows 사용자에게도 db-schema-change 훅이 정상 작동.

- **Task M4.1**: `internal/template/templates/.claude/settings.json.tmpl`에서 `handle-db-schema-change.sh` 등록 블록(현재 약 line 83-91)을 식별. 동일 파일 내 M1.3에서 선정한 기준 블록의 `{{- if eq .Platform "windows"}}...{{- else}}...{{- end}}` 구조를 정확히 카피. Windows 브랜치 command는 `bash "%CLAUDE_PROJECT_DIR%\.claude\hooks\moai\handle-db-schema-change.sh"`, Unix 브랜치는 `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 그대로 유지.
- **Task M4.2**: 템플릿 렌더링 테스트 추가:
  - `TestRender_DbSchemaChangeHook_Unix`: `Platform=darwin` 또는 `linux`로 렌더 → 결과에 `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 정확히 포함, `bash ` 접두사 없음.
  - `TestRender_DbSchemaChangeHook_Windows`: `Platform=windows`로 렌더 → 결과에 `bash ` 접두사(공백 포함)와 `%CLAUDE_PROJECT_DIR%\.claude\hooks\moai\handle-db-schema-change.sh` 포함.
- **Task M4.3 [VERIFY]**: `make build && make install`로 임베디드 템플릿 재생성 후, `/tmp/test-project`에서 `moai init`을 실행하고 생성된 `settings.json`을 Platform별로 검사(macOS 실행 환경에서는 렌더 단위 테스트로 갈음). AC-5, AC-6 통과.

### Milestone M5 (Priority: Medium) — H4 (커버리지 85%)

**목표**: 패키지 총 커버리지 ≥ 85% 달성.

의존성: M2, M3 완료 후 시작. M2/M3의 신규 테스트가 커버리지에 기여하므로 남은 갭만 채우면 됨.

- **Task M5.1**: M2/M3 완료 후 `go test -coverprofile=cover.out ./internal/hook/dbsync/...` 재측정하여 현재 커버리지와 85% 갭 계산.
- **Task M5.2**: 갭 항목별 테이블 케이스 추가(REQ-H4-002 명시 8개):
  - `parseMigrationStub`: `empty_file`, `utf8_bom`, `nonexistent` (oversized는 M2.1에서 이미 작성).
  - `matchGlob`: `trailing_slash`, `double_star_only`, `unicode_path`.
  - `CheckDebounce`: `corrupt_state_recovery` (상태 파일이 부분 쓰기로 잘린 JSON 시 안전 반환 `(true, nil)` — REQ-H2-003과 일치하는 단일 복구 정책).
- **Task M5.3 [VERIFY]**: AC-7, AC-8 통과 확인. `go tool cover -func=cover.out | awk '/^total:/{print $3}'`가 ≥ 85.0%.

### Milestone M6 (Priority: Low) — H5 (MX 계약 godoc)

**목표**: 5개 exported helper에 `@MX:NOTE` 계약 표기 존재 보장 (4개 신규 부착 + 1개 기존 verification).

의존성: M2/M3가 완료된 후 최종 구현 상태가 안정되면 수행(구현 변경 후 다시 주석 갱신하는 이중 작업 방지).

- **Task M6.1**: 5개 함수의 현재 상태 확인:
  - `HandleDBSchemaSync` — **이미** `@MX:NOTE` + `@MX:ANCHOR` 부착 (commit `aa29a9316`). Verification only.
  - `BuildProposal` — `@MX:NOTE` 신규 부착 필요.
  - `MatchesMigrationPattern` — `@MX:NOTE` 신규 부착 필요.
  - `IsExcluded` — `@MX:NOTE` 신규 부착 필요.
  - `CheckDebounce` — `@MX:NOTE` 신규 부착 필요. (시그니처는 M3 이후에도 불변.)
- **Task M6.2**: 4개 신규 대상 각각에 대해 multi-line godoc 블록 작성. REQ-H5-002의 3요소 체크리스트(입력 / 출력 / 부작용)를 포함하며, 자연어는 **한국어**(`code_comments: ko`에 따름). 기술 식별자(함수/파라미터/타입 이름)는 영어 유지. 예시 (`CheckDebounce`):

  ```go
  // @MX:NOTE CheckDebounce는 동일 filePath가 디바운스 윈도우 내에서 이미 관측되었는지 확인하고,
  // 관측되지 않았다면 stateFile을 원자적으로 갱신한다(임시 파일 + os.Rename).
  //
  // 입력:
  //   - stateFile: 디바운스 상태를 기록하는 JSON 파일의 절대 경로
  //   - filePath: PostToolUse 훅이 보고한 마이그레이션 파일 경로
  //   - window: 디바운스 윈도우 (기본 10초; 테스트에서는 짧은 값 사용 가능)
  //
  // 출력:
  //   - debounced: true이면 동일 filePath가 window 내에 이미 관측됨(호출 무시)
  //   - error: nil on success; stateFile JSON 마샬링 실패 시 non-nil
  //
  // 부작용:
  //   - stateFile을 temp-file + os.Rename으로 원자적 교체 (POSIX rename 원자성 활용)
  //   - I/O 실패 시 ErrorLogFile에 한 줄 append 후 안전 기본값 (true, nil) 반환
  func CheckDebounce(stateFile, filePath string, window time.Duration) (bool, error) {
  ```

  godoc 블록 내 `// @MX:NOTE` 마커는 블록의 첫 줄 또는 그 이후 어느 줄에 있든 AC-9 검증에 허용된다(spec.md REQ-H5-001).
- **Task M6.3 [VERIFY]**: AC-9의 awk 기반 multiline 스캔 명령이 5/5 함수에 대해 `OK`를 출력함을 확인. 또는 각 함수별 `grep -B 20 '^func Name(' internal/hook/dbsync/db_schema_sync.go | grep -c '^// @MX:NOTE'` 결과가 모두 `>= 1`.

### Milestone M7 (Priority: High) — 최종 회귀 검증

- **Task M7.1**: `go test ./...` 전체 통과.
- **Task M7.2**: `go test -race ./...` 통과.
- **Task M7.3**: `golangci-lint run ./...` 통과.
- **Task M7.4**: `make build` 통과 및 임베디드 템플릿 재생성 확인.
- **Task M7.5**: SPEC-DB-SYNC-001의 기존 AC-1~AC-10에 대응하는 테스트 케이스 전수 green 확인(AC-10 요구사항).

## Milestone Dependency Graph

```
M1 (baseline) → M2 (H1) ─┐
                         ├→ M5 (H4) → M6 (H5) → M7 (final regression)
              M3 (H2) ───┘
M4 (H3) ─────────────────────────────────↗
```

M2와 M3은 독립적으로 병렬 실행 가능(서로 다른 함수 수정). M4도 독립적(템플릿 수정). M5는 M2/M3 이후 시작. M6은 마지막에 구현이 안정된 후.

## Risks and Mitigations

spec.md Risks 섹션의 R-1~R-5에 더해 구현 관점의 추가 리스크:

- **R-6 (GREEN 전환 실패)**: M2.2의 `os.Stat` 선검사는 심볼릭 링크의 대상(target) 파일 크기를 반영한다. 이는 `parseMigrationStub`가 `os.ReadFile`로 읽는 대상 크기와 일치하므로 의도된 동작이다. `os.Lstat`은 링크 자체 메타데이터만 반환하여 링크 대상의 실제 크기를 놓치므로 사용하지 않는다.
- **R-7 (임시 파일 누수)**: M3.2의 `os.CreateTemp` 후 `os.Rename` 실패 시 임시 파일이 디렉터리에 잔존할 수 있음 → `renamed := false` 플래그와 `defer func() { if !renamed { os.Remove(tmpFile.Name()) } }()` 조합으로 누수 방지. Rename 성공 시 `renamed = true` 세팅.
- **R-8 (테스트 격리 실패)**: 전역 `errorLogPath`가 테스트마다 초기화되지 않으면 한 테스트의 로그가 다른 테스트에 유출 → 각 테스트 시작에 `t.TempDir()`로 경로 재설정하는 헬퍼(`withTempErrorLog(t)`) 추가.
- **R-9 (AC-4 AST 테스트 유지보수)**: `go/parser` API 변경 시 AC-4 테스트가 깨질 위험 → 표준 라이브러리 API는 매우 안정적이며 Go 1.x 호환성 약속으로 보호됨. `go/parser.ParseFile` + `ast.Inspect`는 최소 Go 1.0부터 존재한 공용 API.

## Validation Strategy

- **단위 테스트**: H1/H2/H4는 `go test -race -count=10` 반복 실행으로 경계 조건과 동시성 재현성 검증.
- **렌더 테스트**: H3는 Go `text/template` 단위 테스트로 Platform별 문자열 출력 검증. Windows 러너 불요.
- **AST 기반 정적 검증**: AC-4는 `go/parser` AST 순회로 `CheckDebounce` 내부의 `os.WriteFile` 직접 호출 부재를 검증(변수 rename false-pass 차단).
- **grep/awk 기반 정적 검증**: AC-2/AC-5/AC-8/AC-9는 코드/템플릿 구조 자체를 검증하므로 grep 또는 multi-line awk 명령 실행이 최종 관문.
- **회귀 테스트**: SPEC-DB-SYNC-001 AC-5 재현(Prisma schema 편집 → 5초 내 proposal.json)을 통합 테스트 레벨에서 1회 수행(M7.5).

## Commit Strategy

Conventional Commits 준수. 권장 분할:

1. `test(dbsync): add failing tests for oversized file and concurrent debounce` — M2.1 + M3.1 (RED)
2. `fix(dbsync): add file size guard in parseMigrationStub` — M2.2/M2.3 (GREEN for H1)
3. `fix(dbsync): atomic state file update in CheckDebounce` — M3.2/M3.3 (GREEN for H2)
4. `test(dbsync): AST guard against direct os.WriteFile in CheckDebounce` — M3.4 (AC-4)
5. `fix(template): add Windows platform branch to db-schema-change hook` — M4
6. `test(dbsync): expand edge case coverage to 85%` — M5
7. `docs(dbsync): add @MX:NOTE contracts to exported helpers (4 new + 1 verify)` — M6
8. `chore(dbsync): SPEC-DB-SYNC-HARDEN-001 regression validation` — M7

각 커밋은 독립적으로 빌드 green이며 단계별 롤백 가능.

## Out of Scope Reminder

다음 항목은 본 SPEC에서 다루지 않는다(spec.md Exclusions 재확인):

- 실제 파서 구현 (`internal/db/parser/` 별도 SPEC 대상)
- `/moai db` 서브커맨드 행동 변경
- `handle-db-schema-change.sh` 스크립트 본문 변경
- `settings.json.tmpl` 내 다른 훅 엔트리 수정
- `internal/cli/design_folder.go` TOCTOU (별도 Warning, 별도 SPEC)
- 재귀 가드 알고리즘 변경
- 사용자 승인 플로우(AskUserQuestion 3옵션) 변경
- 문서 갱신 로직(schema.md/erd.mmd/migrations.md) 변경
- `CheckDebounce` 시그니처 변경(현재 `(stateFile, filePath string, window time.Duration) (bool, error)` 유지)
