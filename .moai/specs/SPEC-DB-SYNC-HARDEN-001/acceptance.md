---
id: SPEC-DB-SYNC-HARDEN-001
document: acceptance
version: 0.2.1
created_at: 2026-04-21
updated_at: 2026-04-21
---

# SPEC-DB-SYNC-HARDEN-001: Acceptance Criteria

본 문서는 spec.md의 AC-1 ~ AC-10을 Given-When-Then 시나리오로 전개하며, 각 AC를 5개 Warning 항목(H1~H5) 및 회귀 방어(AC-10)에 매핑한다. 최소 2개 이상의 경계 시나리오를 포함해야 하는 원칙에 따라, 각 H 패밀리에 정상 경로 + 경계 경로 시나리오를 모두 기술한다.

## Definition of Done

구현 PR은 다음 모든 조건을 만족해야 머지 가능하다:

1. spec.md의 REQ-H1-001 ~ REQ-H5-002 (14개 REQ) 모두 구현됨
2. 본 문서의 AC 시나리오(AC-1 ~ AC-10 + Edge 시나리오)가 모두 통과
3. `go test -race ./...` 전체 green
4. `golangci-lint run ./...` green
5. 패키지 커버리지 `internal/hook/dbsync` ≥ 85.0%
6. `make build` 성공, 임베디드 템플릿 재생성됨
7. SPEC-DB-SYNC-001 기존 AC-1 ~ AC-10 회귀 없음(AC-10에서 검증)
8. 본 SPEC의 Target Files 외 다른 파일 수정 없음(scope discipline)

---

## H1 — parseMigrationStub 파일 크기 가드

### Scenario AC-1: 크기 초과 파일은 로그 기록 후 non-error 반환 (`parsed_content=""` + `truncated=true` 동시 설정)

**Given**: 테스트 환경에서 `t.TempDir()` 내부에 2MB 크기의 합성 Prisma 마이그레이션 파일(`schema.prisma`)을 생성하고, `ErrorLogFile` 경로는 `t.TempDir()` 내 테스트 전용 로그 파일로 오버라이드된 상태이다.

**When**: `parseMigrationStub(path)`이 해당 파일 경로로 호출된다.

**Then**:
- 반환값 `error`는 `nil`이다(non-blocking).
- 반환 결과의 `parsed_content` 필드는 빈 문자열 `""`이다 (명확한 확정 값).
- 반환 결과의 `truncated` 필드는 `true`이다 (명확한 확정 값).
- **양쪽 필드가 모두 확정 설정되며, either/or 모호성은 없다** (W-2 해결).
- 테스트 전용 `ErrorLogFile`에 `"parseMigrationStub: file exceeds maxMigrationFileSize="` 접두사로 시작하는 로그가 정확히 1줄 기록된다.
- `os.ReadFile` 또는 동등한 전체 읽기가 호출되지 않는다(REQ-H1-003).

**AC Coverage**: REQ-H1-002, REQ-H1-003

### Scenario AC-2: maxMigrationFileSize 상수 존재 및 참조 확인

**Given**: 리포지토리 최신 HEAD.

**When**: 다음 두 grep 명령이 실행된다:
```
grep -n 'maxMigrationFileSize = 1 << 20' internal/hook/dbsync/db_schema_sync.go
grep -c 'maxMigrationFileSize' internal/hook/dbsync/db_schema_sync.go
```

**Then**:
- 첫 번째 명령은 정확히 1행을 반환한다(선언 라인).
- 두 번째 명령은 `>= 2`를 반환한다(선언 + `parseMigrationStub` 본문 참조 최소 1회).

**AC Coverage**: REQ-H1-001, REQ-H1-003

### Scenario AC-Edge-H1: 정상 크기 파일은 기존 경로 유지

**Given**: `t.TempDir()` 내부에 50KB 크기의 유효한 Prisma schema 파일이 존재하고, 기존 파서 스텁이 그 파일에 대해 정상 반환을 해왔다.

**When**: `parseMigrationStub(path)`이 호출된다.

**Then**:
- 반환 에러 `nil`, `parsed_content`가 비어있지 않고(파일 내용 포함), `truncated`가 `false`.
- `ErrorLogFile`에 크기 초과 관련 로그가 **기록되지 않는다**.
- 기존 동작(SPEC-DB-SYNC-001 REQ-008의 계약)이 유지된다.

**AC Coverage**: REQ-H1-002(회귀 방어 측), REQ-H4-002(empty/nonexistent 경계 포함 테이블 케이스로 함께 커버)

---

## H2 — CheckDebounce 원자성 (시그니처 불변)

### Scenario AC-3: 동시 호출 시 정확히 한 호출만 not-debounced 관측

**Given**:
- 테스트 환경에서 `t.TempDir()` 내부에 `stateFile := filepath.Join(tmp, "last-seen.json")` 경로가 준비되어 있고, 해당 파일은 아직 존재하지 않는다.
- `filePath := "prisma/schema.prisma"`로 고정한다.
- `window := 50 * time.Millisecond` (짧은 실제 윈도우로 `time.Now()` 기반 디바운스를 재현).
- `sync.WaitGroup`에 2개 고루틴을 등록한다.

**When**: 두 고루틴이 동시에 `CheckDebounce(stateFile, filePath, window)`를 호출한다 — **live signature 그대로** (`stateFile, filePath string, window time.Duration) (bool, error)`). 테스트는 `go test -race -count=10 -run TestCheckDebounceConcurrency`로 10회 반복 실행된다.

**Then** (매 반복마다):
- 두 goroutine의 `err`는 모두 `nil`이다.
- 반환된 두 `debounced` 값의 multiset이 정확히 `{false, true}`이다(순서 무관).
- 종료 후 `stateFile`을 읽으면 유효한 JSON으로 파싱된다(부분 쓰기/손상 없음).
- `go test -race`에서 데이터 레이스 검출이 없다.
- 10회 반복 모두 통과한다.

**AC Coverage**: REQ-H2-001, REQ-H2-002

### Scenario AC-4: 원자적 접근 채택 AST 기반 정적 검증 (변수 rename false-pass 방지)

**Given**: 리포지토리 최신 HEAD와 `TestCheckDebounce_NoDirectWriteFile` Go 단위 테스트.

**When**: 테스트가 다음을 수행한다:
1. `go/parser.ParseFile`로 `internal/hook/dbsync/db_schema_sync.go`를 AST로 파싱.
2. Top-level declarations를 순회하여 `Name.Name == "CheckDebounce"`인 `*ast.FuncDecl`을 찾는다.
3. 해당 함수의 `Body`를 `ast.Inspect`로 순회하여 모든 `*ast.CallExpr`를 수집.
4. 각 `*ast.CallExpr`의 `Fun`이 `*ast.SelectorExpr`인 경우, `X.Name` 및 `Sel.Name`을 추출하여 `"os"` + `"WriteFile"` 조합 호출을 카운트.
5. 동일하게 `"os"` + `"Rename"` 또는 `flock` 관련 함수 호출을 카운트.

**Then**:
- `CheckDebounce` 함수 본문 내부의 `os.WriteFile(...)` 직접 호출 카운트가 정확히 `0`이다.
- `CheckDebounce` 함수 본문 내부의 `os.Rename(...)` 호출 카운트가 `>= 1`이다 (또는 flock 경로를 채택했다면 그쪽 호출이 `>= 1`).
- **변수 이름을 `stateFile`에서 `StateFile`/`cfg.StateFile`/`path`/`target` 등으로 바꾸어도 AST 구조 검증은 영향받지 않는다** (W-5 해결).

**AC Coverage**: REQ-H2-001

### Scenario AC-Edge-H2: I/O 실패 시 안전 기본값 `(true, nil)` 반환

**Given**: 테스트 환경에서 `stateFile`의 부모 디렉터리가 읽기 전용(`os.Chmod(dir, 0444)`)으로 설정되어 쓰기가 불가능하다.

**When**: `CheckDebounce(stateFile, filePath, window)`가 호출된다.

**Then**:
- 반환 `debounced`는 `true`(안전 기본값, REQ-H2-003).
- 반환 `err`는 `nil` (non-blocking — 파이프라인을 차단하지 않음).
- `ErrorLogFile`에 I/O 실패 로그가 1줄 이상 기록된다.
- 함수는 panic하지 않고 정상 종료한다.

**AC Coverage**: REQ-H2-003, REQ-H4-002(corrupt_state_recovery와 함께 카테고리 구성)

---

## H3 — Windows 플랫폼 분기

### Scenario AC-5: settings.json.tmpl 플랫폼 분기 정적 구조 검증 + 다른 엔트리 일관성 증거

**Given**: 리포지토리 최신 HEAD.

**When**: 다음 두 검사가 실행된다:

검사 1 (db-schema-change 엔트리 자체):
```
grep -nA 10 'handle-db-schema-change.sh' internal/template/templates/.claude/settings.json.tmpl
```

검사 2 (다른 엔트리 일관성 증거):
```
grep -c '{{- if eq .Platform "windows"}}' internal/template/templates/.claude/settings.json.tmpl
```

**Then**:
- 검사 1의 출력 블록 안에 다음 세 행이 모두 포함된다(순서 유지):
  - `{{- if eq .Platform "windows"}}`
  - `{{- else}}`
  - `{{- end}}`
- 검사 2의 결과가 `>= 3` (db-schema-change 엔트리 + 최소 2개 이상의 기존 엔트리에서 동일 토큰 사용 — 다른 PostToolUse/Stop/SessionStart/SubagentStop 엔트리가 같은 패턴을 일관되게 사용한다는 증거; W-6 해결).

**AC Coverage**: REQ-H3-001

### Scenario AC-6: Windows/Unix 양방향 렌더 결과 검증

**Given**: Go 템플릿 렌더링 테스트(`internal/template/...` 내 신규 또는 기존 테스트).

**When-1**: `TestRender_DbSchemaChangeHook_Windows` 실행 시 `Platform="windows"`로 `settings.json.tmpl`을 렌더한다.

**Then-1**:
- 렌더 결과 db-schema-change 훅 블록의 command 필드가 정확히 `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 문자열을 포함한다 (JSON 이스케이프 형태: `bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh\"`).
- 동일 파일 내 다른 Windows 훅 엔트리(예: handle-session-start, handle-subagent-stop)가 사용하는 패턴과 일치한다.
- 결과는 유효한 JSON으로 파싱 가능하다.

**When-2**: `TestRender_DbSchemaChangeHook_Unix` 실행 시 `Platform="darwin"` 또는 `Platform="linux"`로 동일 템플릿을 렌더한다.

**Then-2**:
- 렌더 결과 db-schema-change 훅 command 필드는 `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` 문자열을 정확히 포함한다.
- `bash ` 접두사는 포함되지 않는다.
- SPEC-DB-SYNC-001 AC-2의 기존 형식과 문자열 일치한다(회귀 방어).

**AC Coverage**: REQ-H3-002, REQ-H3-003

---

## H4 — 커버리지 77.8% → 85%

### Scenario AC-7: 패키지 커버리지 임계 달성

**Given**: 리포지토리 최신 HEAD, 본 SPEC 구현 완료 상태.

**When**: 다음 명령이 실행된다:
```
go test -coverprofile=cover.out ./internal/hook/dbsync/... && go tool cover -func=cover.out | awk '/^total:/{print $3}' | sed 's/%//'
```

**Then**: 출력 값이 부동소수점으로 `>= 85.0`이다.

**AC Coverage**: REQ-H4-001

### Scenario AC-8: 지정된 테이블 케이스 이름 존재

**Given**: `internal/hook/dbsync/db_schema_sync_test.go` 최신 HEAD.

**When**: 각 케이스 이름에 대해 `grep -c '"empty_file"' internal/hook/dbsync/db_schema_sync_test.go` 등을 실행한다.

**Then**: 다음 8개 케이스 이름 모두 `>= 1`로 등장한다:
- `empty_file`
- `utf8_bom`
- `oversized`
- `nonexistent`
- `trailing_slash`
- `double_star_only`
- `unicode_path`
- `corrupt_state_recovery`

**AC Coverage**: REQ-H4-002, REQ-H4-003

### Scenario AC-Edge-H4: 테스트 격리 준수

**Given**: H4에서 추가된 테스트 케이스 코드.

**When**: 코드 리뷰 및 정적 검사.

**Then**:
- 모든 추가 테스트가 `t.TempDir()`을 사용하여 임시 파일 생성.
- `t.Setenv("HOME", ...)` 사용하지 않는다(CLAUDE.local.md §6 위반 금지).
- 테스트 종료 후 프로젝트 루트에 잔여 파일이 생기지 않는다.

**AC Coverage**: REQ-H4-003

---

## H5 — MX 계약 godoc (5 함수 전수)

### Scenario AC-9: 5개 함수 godoc 블록 내부에 @MX:NOTE 마커 존재

**Given**: `internal/hook/dbsync/db_schema_sync.go` 최신 HEAD.

**When**: 다음 awk 기반 multiline 스캔 명령이 실행된다:
```bash
awk '
  /^\/\/ @MX:NOTE/ { seen_mx=1 }
  /^func (HandleDBSchemaSync|BuildProposal|MatchesMigrationPattern|IsExcluded|CheckDebounce)\(/ {
    if (seen_mx) { print "OK " $2 } else { print "MISSING " $2; exit 1 }
    seen_mx=0
  }
  /^[[:space:]]*$/ { seen_mx=0 }
' internal/hook/dbsync/db_schema_sync.go
```

또는 각 함수별 단순 형태로:
```bash
for fn in HandleDBSchemaSync BuildProposal MatchesMigrationPattern IsExcluded CheckDebounce; do
  count=$(grep -B 20 "^func ${fn}(" internal/hook/dbsync/db_schema_sync.go | grep -c '^// @MX:NOTE')
  [ "$count" -ge 1 ] || { echo "FAIL: $fn missing @MX:NOTE"; exit 1; }
done
```

**Then**:
- 5개 함수(`HandleDBSchemaSync`, `BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce`) 전부에 대해 함수 시그니처 이전 20줄 윈도우 내에 `// @MX:NOTE` 마커가 최소 1줄 존재한다.
- `HandleDBSchemaSync`는 이미 commit `aa29a9316`에서 `@MX:NOTE`가 부착되어 있으므로 기존 상태 유지 확인.
- 나머지 4개 함수는 본 SPEC M6에서 신규 부착.
- **multi-line godoc 블록 전체(첫 줄부터 `func` 직전까지)가 스캔 윈도우에 포함되므로 plan.md:L114-L131의 예시 구조와 일관** (F-1 해결).

**AC Coverage**: REQ-H5-001

### Scenario AC-Edge-H5: MX 주석 3요소 포함 검증 + 자연어 설정 준수

**Given**: 5개 함수의 godoc 블록(함수 정의 이전 연속 주석 범위). `.moai/config/sections/language.yaml`의 `code_comments` 값은 `ko`이다.

**When**: 코드 리뷰에서 다음 두 가지를 확인:

확인 1 (3요소 키워드 존재):
- 입력 요소: `입력:` 또는 `Inputs:` 또는 `Parameters:` 중 하나
- 출력 요소: `출력:` 또는 `Outputs:` 또는 `Returns:` 중 하나
- 부작용 요소: `부작용:` 또는 `Side Effects:` 중 하나

확인 2 (자연어 선택):
- godoc 블록의 산문 부분(prose)이 한국어로 작성됨 (REQ-H5-002에 따라 `code_comments: ko` 설정 준수).
- 기술 식별자(함수/파라미터/타입 이름, 리터럴 값)는 영어 그대로 유지.

**Then**:
- 5개 godoc 블록 모두 3요소 키워드를 각 1개 이상 포함한다.
- 산문 문장은 한국어 (W-4 해결: 하드코딩된 "한국어/영어 혼용" 대신 프로젝트 설정 준수).

**AC Coverage**: REQ-H5-002

---

## AC-10 — 회귀 방어

### Scenario AC-10: SPEC-DB-SYNC-001 기존 AC 전수 유지

**Given**: 리포지토리 최신 HEAD, 본 SPEC 구현 완료 상태.

**When**:
- `go test ./...` 실행.
- `.moai/specs/SPEC-DB-SYNC-001/spec.md`의 AC-1 ~ AC-10에 대응하는 테스트 케이스 전수 실행.

**Then**:
- `go test ./...` 전체 green.
- SPEC-DB-SYNC-001 AC-1: `handle-db-schema-change.sh`이 존재, 실행 가능(`+x`), 30라인 미만.
- SPEC-DB-SYNC-001 AC-2: `settings.json.tmpl`에 PostToolUse 매처 엔트리 존재, timeout 설정 유효(SPEC-DB-SYNC-001이 선언한 값과 live 템플릿 값 일치). **추가로 본 SPEC H3에 따라 Windows/Unix 플랫폼 분기 포함**.
- SPEC-DB-SYNC-001 AC-3: `moai hook db-schema-sync --file <path>` 디바운스 경로에서 exit 0 반환. **추가로 본 SPEC H2에 따라 동시성 안전**.
- SPEC-DB-SYNC-001 AC-4~AC-9: 기존 동작 변경 없음.
- SPEC-DB-SYNC-001 AC-10(debounce): 단일 파일 2회 연속 편집 → 1회만 `proposal.json` 생성. **본 SPEC H2가 이 보증을 **동시 호출 조건에서도** 확장**.

**AC Coverage**: 회귀 방어 전반

---

## Quality Gate Criteria

### TRUST 5 적합성

| TRUST 항목 | 본 SPEC에서의 충족 방식 |
|-----------|----------------------|
| Tested | 커버리지 ≥ 85% (REQ-H4-001), 동시성 race 테스트 10회 반복, AST 기반 정적 검증(AC-4) |
| Readable | `@MX:NOTE` godoc으로 계약 명시(H5), 명확한 함수 이름 유지, 한국어 산문(`code_comments: ko`) |
| Unified | 기존 PostToolUse/Stop/SessionStart/SubagentStop 엔트리와 동일한 플랫폼 분기 패턴 채택(H3) |
| Secured | 대용량 파일 메모리 DoS 완화(H1), 동시성 데이터 레이스 제거(H2) |
| Trackable | Conventional Commits, SPEC-DB-SYNC-HARDEN-001 traceability matrix |

### MX Tag 요구사항

- 신규 exported 함수 없음 → `@MX:ANCHOR` 신규 추가 불필요 (`HandleDBSchemaSync`는 이미 부착됨)
- 4개 기존 exported 함수에 `@MX:NOTE` 신규 부착 (`BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce`)
- `HandleDBSchemaSync`는 기존 `@MX:NOTE` + `@MX:ANCHOR` 유지 verification (H5)
- 복잡도 ≥ 15 또는 danger zone 함수 없음 → `@MX:WARN` 불필요
- `@MX:TODO` 신규 생성 없음(본 SPEC은 완결형 견고화)

### 성능 / 관측성

- H1: 대용량 파일 입력 시 메모리 피크 감소(2MB → O(1) 선검사).
- H2: 동시 이벤트 시 중복 proposal.json 생성 제거.
- H3: Windows 사용자에게 훅 작동 회복.
- H4: 커버리지 격차 해소로 회귀 탐지력 향상.
- H5: 리팩토링 시 계약 위반 탐지력 향상 (5 함수 전수).

---

## Sign-off Checklist

- [ ] AC-1 (H1 정상 경로, `parsed_content=""` + `truncated=true` 동시 설정) 통과
- [ ] AC-2 (H1 상수 정적 검증) 통과
- [ ] AC-Edge-H1 (정상 크기 회귀) 통과
- [ ] AC-3 (H2 동시성, 실제 `time.Now()` + `window=50ms`) 통과 × 10회
- [ ] AC-4 (H2 AST 기반 `os.WriteFile` 부재 검증) 통과
- [ ] AC-Edge-H2 (I/O 실패 `(true, nil)` 반환) 통과
- [ ] AC-5 (H3 플랫폼 분기 정적 + 일관성 증거) 통과
- [ ] AC-6 (H3 Windows/Unix 렌더) 통과
- [ ] AC-7 (H4 커버리지 ≥ 85%) 통과
- [ ] AC-8 (H4 테이블 케이스 이름) 통과
- [ ] AC-Edge-H4 (테스트 격리) 통과
- [ ] AC-9 (H5 `@MX:NOTE` 프리픽스, 5 함수 awk multiline) 통과 × 5함수
- [ ] AC-Edge-H5 (MX 3요소 포함 + 한국어 산문) 통과
- [ ] AC-10 (회귀 방어) 통과
- [ ] `go test -race ./...` 전체 green
- [ ] `golangci-lint run ./...` green
- [ ] `make build` 성공
- [ ] Target Files 외 수정 없음 확인
