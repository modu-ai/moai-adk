# SPEC-V3R3-DESIGN-FOLDER-FIX-001 — Implementation Plan

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial plan. Single-commit bug fix: design_folder.go warning 격하 + test 갱신 + constitution amendment. |

---

## 1. Implementation Strategy

### 1.1 Approach Selection

**선택**: 단일 commit, TDD 사이클 (RED → GREEN → 검증)

**Rationale**:
- 변경 범위 좁음 (1 Go 파일 + 1 test 파일 + 1 rule 파일 + 1 template mirror)
- 의존성 없는 독립 fix
- Bug fix 본질상 split하면 오히려 reviewer cognitive load 증가

**Alternative 고려 후 거부**:
- 2-commit split (impl 분리 + constitution 분리): rule 변경이 코드 변경과 의미적으로 결합되어 있어 분리 시 review 흐름 단절
- 3-commit split (test → impl → docs): TDD 의도는 좋지만 단일 PR cycle 내 간섭 우려

### 1.2 Code Change Pattern (Strict Mode 분기)

**Option A (선택): bool 파라미터 추가**
```go
// 시그니처 변경
func checkReservedCollision(projectRoot string, errOut io.Writer, strict bool) error
// strict=true → error 반환 (scaffold path)
// strict=false → warning 출력 후 nil 반환 (update path)
```

**Option B (거부): 별도 함수 분리**
- `checkReservedCollisionStrict` + `warnReservedCollision` 두 함수 도입
- 거부 이유: 호출처가 `updateDesignDir` 1곳뿐. 함수 2개 도입 overkill.

**Option C (거부): caller에서 분기**
- `checkReservedCollision`은 무조건 발견 list 반환, caller가 error/warning 결정
- 거부 이유: 시그니처 변경 폭 큼 (반환 타입 `error` → `[]string` + error). 기존 시그니처 보존 우선.

**최종 선택**: Option A — strict bool 파라미터로 명시적 분기.

### 1.3 함수 동작 (수정 후)

```
checkReservedCollision(projectRoot, errOut, strict bool) error
├── reserved files 발견 list 수집
├── if strict == true:
│     errOut에 "error: reserved filename: ..." 출력
│     return error (첫 번째 발견)
├── if strict == false:
│     errOut에 "warning: reserved filename: ... (preserved; rename to use canonical templates)" 출력
│     return nil (모든 발견 list warning 후)
```

**호출 변경**:
- `updateDesignDir`: `checkReservedCollision(projectRoot, errOut, false)` (warning mode)
- `scaffoldDesignDir`: 현재는 `checkReservedCollision`을 호출하지 않음. 변경 없음.

**확인 사항**: scaffold path가 reserved 검사를 호출하지 않음 → 신규 프로젝트는 reserved name 위반 가능성 사실상 없음 (template만 deploy). 따라서 REQ-DFF-003은 update path 동작에 의해 자동 보장. 그러나 향후 누가 scaffold 호출 추가할 때를 대비해 strict mode 시그니처를 분리 유지.

---

## 2. Milestones (Priority-Based, No Time Estimates)

### M1 (P0): Test 갱신 (RED Phase)

**Priority**: P0 Critical (TDD 시작점)

**Tasks**:
1. `internal/cli/design_folder_test.go`:
   - `TestDesignFolderReservedExact` 수정: error 기대 → warning + nil + 다른 파일 sync 검증
   - `TestDesignFolderReservedGlob` 수정: 동일 패턴
   - `TestDesignFolderReservedNotModified` 강화: 파일 보존 + 다른 templates sync 검증 추가
   - `TestDesignFolderUpdate_WarningIncludesGuidance` 신규: warning 메시지에 "preserved" + 우회 안내 포함
   - `TestDesignFolderUpdate_MultipleReservedConflicts` 신규: 다중 충돌 케이스
2. `go test ./internal/cli/... -run TestDesignFolder` 실행 → 전체 RED 확인

**Exit Criteria**: 갱신된 모든 테스트가 명확한 RED 상태 (기존 hard error 동작과 새 expectation 불일치).

### M2 (P0): 구현 (GREEN Phase)

**Priority**: P0 Critical

**Tasks**:
1. `internal/cli/design_folder.go`:
   - `checkReservedCollision` 시그니처에 `strict bool` 추가
   - strict=true: 기존 error 반환 동작 유지
   - strict=false: 발견 list 누적 → warning 출력 → nil 반환
   - warning 메시지 포맷: `warning: reserved filename: <path> (preserved; rename to use canonical templates)`
   - `updateDesignDir`의 호출을 `checkReservedCollision(projectRoot, errOut, false)`로 변경
   - reserved 충돌 발견된 파일은 update loop에서 skip (canonical hash 비교 대상 외)

**Exit Criteria**:
- M1의 모든 테스트 GREEN
- `go vet ./...` clean
- `golangci-lint run` clean
- 기존 `TestDesignFolderUserEditPreserved` 등 unrelated 테스트도 모두 통과

### M3 (P1): Documentation Amendment

**Priority**: P1 High

**Tasks**:
1. `.claude/rules/moai/design/constitution.md`:
   - HISTORY 최상단에 v3.3.1 entry 추가
   - §3.2 footnote: "Reserved name violation은 update path에서 warning + skip, scaffold path에서 hard error"
   - Version footer: 3.3.0 → 3.3.1
2. `internal/template/templates/.claude/rules/moai/design/constitution.md`:
   - 동일한 변경 적용 (Template-First HARD 준수)
3. `make build` 실행 → `internal/template/embedded.go` 재생성

**Exit Criteria**:
- 두 constitution 파일 byte-identical (template mirror 검증)
- `make build` 성공
- diff 검증: `diff .claude/rules/moai/design/constitution.md internal/template/templates/.claude/rules/moai/design/constitution.md` empty

### M4 (P1): Full Test Suite + 검증

**Priority**: P1 High

**Tasks**:
1. `go test ./...` 전체 통과
2. `go test -race ./internal/cli/...` 통과
3. `make build && make install` (로컬 binary 검증)
4. 수동 검증: `~/MoAI/mo.ai.kr` 또는 임시 프로젝트에서 `moai update` 실행 → tokens.json 존재 시 warning + 다른 파일 sync 정상 확인

**Exit Criteria**: 모든 quality gate green, 사용자 보고 시나리오 재현 후 정상 동작 확인.

---

## 3. Technical Approach

### 3.1 Files Modified

| File | Change Type | LOC Estimate |
|------|-------------|--------------|
| `internal/cli/design_folder.go` | Modify | +20/-10 |
| `internal/cli/design_folder_test.go` | Modify + Add | +80/-30 |
| `.claude/rules/moai/design/constitution.md` | Modify (HISTORY + footnote + version) | +6/-1 |
| `internal/template/templates/.claude/rules/moai/design/constitution.md` | Mirror | +6/-1 |
| `internal/template/embedded.go` | Auto-regenerate via `make build` | (generated) |

총 신규 코드 < 100 LOC. 단일 commit 적합.

### 3.2 Function Signature Changes

**Before**:
```go
func checkReservedCollision(projectRoot string, errOut io.Writer) error
```

**After**:
```go
func checkReservedCollision(projectRoot string, errOut io.Writer, strict bool) error
```

**Caller 변경**:
- `updateDesignDir`: `checkReservedCollision(projectRoot, errOut, false)` (warning mode, REQ-DFF-001)
- 다른 caller 없음 (grep 검증 필요)

### 3.3 Update Loop Skip Logic

`updateDesignDir`은 현재 `designTemplateFiles` (README/research/system/spec)만 iterate. reserved files는 `designTemplateFiles`에 없으므로 별도 skip 로직 불필요. **단**, warning 출력 후 다른 template sync는 정상 진행되어야 함을 명시적으로 verify.

### 3.4 Warning Message Format

**English** (default `code_comments` setting):
```
warning: reserved filename: tokens.json (preserved; rename to use canonical templates if you want them auto-managed)
```

**Korean** (when `code_comments: ko`):
```
warning: 예약 파일명 충돌: tokens.json (사용자 파일 보존됨; canonical template 사용을 원하면 rename 필요)
```

**선택**: 단순화를 위해 영어 단일 메시지로 시작 (REQ-DFF-005 충족). i18n은 별도 SPEC 영역.

---

## 4. Risks and Mitigations

| 위험 | 가능성 | 영향 | 완화 |
|------|--------|------|------|
| `checkReservedCollision` 시그니처 변경이 다른 caller에 영향 | 낮음 | 중간 | grep으로 caller 전수 검증; 1곳만 발견 시 OK |
| Constitution amendment HISTORY 누락 | 중간 | 낮음 | review checklist에 추가; M3 exit criteria에 명시 |
| Template mirror 누락 | 중간 | 높음 | M3 exit criteria diff 검증; pre-commit hook 활용 |
| 기존 사용자 워크플로우 파괴 (warning이 noise로 인식) | 낮음 | 중간 | warning에 "preserved" 키워드 명시; 사용자가 의도한 동작 명확히 안내 |
| `errOut`이 nil인 케이스 (silent skip 위험) | 낮음 | 중간 | nil check 추가; nil인 경우 stderr fallback 또는 silent OK 정책 결정 (현재 코드 nil 허용) |

---

## 5. Dependencies and Pre-conditions

### 5.1 Dependencies

- 없음. 독립 fix.

### 5.2 Pre-conditions

- 현재 branch `fix/SPEC-V3R3-DESIGN-FOLDER-FIX-001` 활성
- `internal/cli/design_folder.go` 및 test 파일 read 권한 확보
- `make build` 정상 동작
- `~/.moai/.env.glm` 등 GLM 토큰은 이번 fix와 무관

### 5.3 Post-conditions

- 모든 quality gate green
- `moai update`가 reserved 충돌 시에도 다른 templates sync 정상 진행
- v2.15.x patch release candidate (단, release는 별도 SPEC/PR)

---

## 6. Quality Gates

### 6.1 Pre-Commit

- [ ] `go test ./...` 전체 통과
- [ ] `go test -race ./internal/cli/...` 통과
- [ ] `golangci-lint run` 0 warnings
- [ ] `go vet ./...` clean
- [ ] `make build` 성공 + embedded.go 재생성 확인

### 6.2 Pre-PR

- [ ] template mirror diff empty
- [ ] constitution.md HISTORY entry 추가됨
- [ ] commit message 한국어 body + Conventional Commits format
- [ ] `~/MoAI/mo.ai.kr` 시나리오 수동 재현 완료

### 6.3 Post-Merge

- [ ] CHANGELOG.md v2.15.x entry (별도 release SPEC에서)
- [ ] User feedback 수집 (다음 release cycle)

---

## 7. Out-of-Band Considerations

### 7.1 i18n (Korean Warning Message)

`code_comments: ko` 설정 시 한국어 warning 메시지 제공은 nice-to-have. 본 SPEC 범위 외. 향후 별도 enhancement SPEC.

### 7.2 Constitution Amendment 승인 절차

design constitution은 FROZEN amendment (SPEC-DESIGN-CONST-AMEND-001 패턴 따름). v3.3.0 → v3.3.1 bump는 footnote 추가 수준 minor amendment. 인간 승인 필요하나 P0 bug fix이므로 PR review 단계에서 일괄 승인 요청.

### 7.3 Release Note Hook

v2.15.1 또는 v2.16.0 release 시 release note에 "moai update가 더 이상 reserved filename 충돌 시 abort하지 않음" 명시 필요. 본 SPEC은 fix만 다루고, release SPEC에서 release note 작성.

---

Version: 0.1.0
Last Updated: 2026-04-26
