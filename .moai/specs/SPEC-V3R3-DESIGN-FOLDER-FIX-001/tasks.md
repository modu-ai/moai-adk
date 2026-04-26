# SPEC-V3R3-DESIGN-FOLDER-FIX-001 — Task Breakdown

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial task list. 9 tasks across 4 milestones. Single-commit bug fix. |

---

## 1. Task Overview

| Task ID  | Milestone | Priority | Owner            | Files Touched | Dependencies |
|----------|-----------|----------|------------------|----------------|--------------|
| T-01     | M1        | P0       | manager-tdd      | design_folder_test.go | none |
| T-02     | M1        | P0       | manager-tdd      | design_folder_test.go | T-01 |
| T-03     | M1        | P0       | manager-tdd      | design_folder_test.go | T-01 |
| T-04     | M1        | P1       | manager-tdd      | design_folder_test.go | T-01 |
| T-05     | M1        | P1       | manager-tdd      | design_folder_test.go | T-01 |
| T-06     | M2        | P0       | expert-backend   | design_folder.go | T-01..T-05 |
| T-07     | M3        | P1       | manager-docs     | constitution.md (×2) | T-06 |
| T-08     | M3        | P1       | expert-backend   | embedded.go via make | T-07 |
| T-09     | M4        | P0       | manager-quality  | full test suite | T-06, T-08 |

---

## 2. Detailed Tasks

### T-01 (M1, P0) — Update TestDesignFolderReservedExact

**File**: `internal/cli/design_folder_test.go`
**Owner**: manager-tdd

**Action**:
- 기존 `TestDesignFolderReservedExact` (line 153~190)을 `TestDesignFolderUpdate_ReservedExact_WarnsButContinues`로 rename
- expectation 변경:
  - error 반환 → nil 반환
  - errBuf에 "error: reserved filename" → "warning" + "tokens.json" + "preserved" 포함
  - tokens.json 내용 보존 (기존 검증 유지)
  - 추가: README.md user-edit 보존 검증 (다른 sync 정상 진행 증명)

**Acceptance**: AC-DFF-01

**Exit**:
- 테스트가 명확한 RED 상태 (현재 hard error 동작과 새 expectation 불일치)

---

### T-02 (M1, P0) — Update TestDesignFolderReservedGlob

**File**: `internal/cli/design_folder_test.go`
**Owner**: manager-tdd

**Action**:
- 기존 `TestDesignFolderReservedGlob` (line 194~226)을 `TestDesignFolderUpdate_ReservedGlob_WarnsButContinues`로 rename
- expectation을 T-01과 동일 패턴으로 변경
- BRIEF-LOCAL.md 보존 + 다른 templates sync 검증 추가

**Acceptance**: AC-DFF-02

**Exit**: RED 상태 확인

---

### T-03 (M1, P0) — Strengthen TestDesignFolderReservedNotModified

**File**: `internal/cli/design_folder_test.go`
**Owner**: manager-tdd

**Action**:
- 기존 `TestDesignFolderReservedNotModified` (line 230~256)을 `TestDesignFolderUpdate_ReservedNotModified`로 rename
- expectation 강화:
  - 두 reserved 파일 (components.json + import-warnings.json) 동시 존재
  - 모두 byte-identical 보존
  - 함수 nil 반환 (기존 코드는 `_ = updateDesignDir`로 무시했으나 명시적으로 검증)

**Acceptance**: AC-DFF-05

**Exit**: RED 상태 확인

---

### T-04 (M1, P1) — Add TestDesignFolderUpdate_WarningIncludesGuidance

**File**: `internal/cli/design_folder_test.go`
**Owner**: manager-tdd

**Action**: 신규 테스트 추가
```go
func TestDesignFolderUpdate_WarningIncludesGuidance(t *testing.T) {
    // components.json 작성 → updateDesignDir 호출 → errBuf에 4 keyword 포함 검증
}
```

**Required Keywords**: "warning", "components.json", "preserved", "rename"

**Acceptance**: AC-DFF-04

**Exit**: RED 상태 확인 (현재 메시지에 "rename" 없음)

---

### T-05 (M1, P1) — Add TestDesignFolderScaffold_ReservedExact_StillErrors + MultipleConflicts

**File**: `internal/cli/design_folder_test.go`
**Owner**: manager-tdd

**Action**: 두 개 신규 테스트 추가
1. `TestDesignFolderScaffold_ReservedExact_StillErrors`: `checkReservedCollision(root, &errBuf, true)` 직접 호출, error 반환 검증
2. `TestDesignFolderUpdate_MultipleReservedConflicts`: 세 reserved 파일 동시 존재, 모두 warning + 모두 보존

**Acceptance**: AC-DFF-03, AC-DFF-06

**Exit**: 두 테스트 모두 RED (T-05.1: strict 파라미터 미존재로 컴파일 에러; T-05.2: 현재 hard error 동작)

---

### T-06 (M2, P0) — Implement Strict Mode Dispatch

**File**: `internal/cli/design_folder.go`
**Owner**: expert-backend

**Action**:
1. `checkReservedCollision` 시그니처 변경:
   ```go
   func checkReservedCollision(projectRoot string, errOut io.Writer, strict bool) error
   ```
2. 함수 본문 변경:
   - reserved 파일 발견 list 누적 (append 패턴)
   - strict == true: 첫 발견 시 즉시 error 반환 + errOut에 "error: reserved filename: ..." 출력 (기존 동작)
   - strict == false: 모든 발견 list iterate → errOut에 "warning: reserved filename: <path> (preserved; rename to use canonical templates)" 출력 → return nil
3. `updateDesignDir`의 호출 변경: `checkReservedCollision(projectRoot, errOut, false)`
4. 시그니처 변경 영향 caller grep 검증: `grep -rn "checkReservedCollision" internal/`

**Acceptance**: M1의 모든 AC GREEN 전환

**Exit**:
- M1의 6개 테스트 모두 GREEN
- 기존 `TestDesignFolderUserEditPreserved`, `TestDesignFolderSubdirs` 등 unrelated 테스트 모두 통과
- `go test ./internal/cli/...` 전체 통과
- `golangci-lint run` clean

---

### T-07 (M3, P1) — Constitution v3.3.1 Amendment

**Files**:
- `.claude/rules/moai/design/constitution.md`
- `internal/template/templates/.claude/rules/moai/design/constitution.md`

**Owner**: manager-docs

**Action**:
1. HISTORY 최상단에 entry 추가:
   ```
   - 2026-04-26 (SPEC-V3R3-DESIGN-FOLDER-FIX-001): §3.2 footnote — Reserved name violation은 update path에서 warning + skip, scaffold path에서 hard error. v3.3.0 → 3.3.1.
   ```
2. §3.2 마지막에 footnote 추가:
   ```
   > Note: Reserved name violation은 `moai update` (update path)에서는 warning 출력 + 해당 파일 skip + 다른 templates sync 계속 진행. `moai init` (scaffold path)에서는 hard error 반환. 사용자 데이터는 어떤 경우에도 수정·삭제되지 않음.
   ```
3. Version footer: `3.3.0` → `3.3.1`
4. `Last Updated`: `2026-04-20` → `2026-04-26`
5. 두 파일 byte-identical 검증: `diff .claude/rules/moai/design/constitution.md internal/template/templates/.claude/rules/moai/design/constitution.md` empty

**Acceptance**: D-1, D-2 (spec.md §5.3)

**Exit**: diff empty 확인

---

### T-08 (M3, P1) — Regenerate Embedded Templates

**File**: `internal/template/embedded.go` (auto-generated)
**Owner**: expert-backend

**Action**:
1. `make build` 실행
2. `git diff internal/template/embedded.go` 확인 → constitution.md 변경 반영 확인
3. `go test ./internal/template/...` 통과 확인 (embedded fs integrity)

**Acceptance**: Q-4 (spec.md §5.2)

**Exit**: build 성공 + embedded.go 갱신 확인

---

### T-09 (M4, P0) — Full Quality Gate Verification

**Owner**: manager-quality

**Action**:
1. `go test ./...` 전체 통과
2. `go test -race ./internal/cli/...` 통과
3. `golangci-lint run` 0 warnings
4. `go vet ./...` clean
5. `make install` (로컬 binary 갱신)
6. 수동 검증 시나리오:
   - 임시 디렉터리 생성: `mkdir -p /tmp/test-update-fix && cd /tmp/test-update-fix && moai init`
   - `.moai/design/tokens.json` 작성: `echo '{"primary":"#000"}' > .moai/design/tokens.json`
   - `moai update` 실행 → warning 출력 확인 + tokens.json 보존 확인 + 다른 design templates sync 확인
   - 또는 `~/MoAI/mo.ai.kr`에서 직접 재현 (사용자 보고 시나리오)

**Acceptance**: spec.md §5.1 F-1 ~ F-4 + Q-1 ~ Q-4

**Exit**:
- 모든 quality gate green
- 사용자 시나리오 수동 재현 정상

---

## 3. Task Dependency Graph

```
T-01 ──┐
T-02 ──┤
T-03 ──┼── T-06 ── T-07 ── T-08 ── T-09
T-04 ──┤
T-05 ──┘
```

- T-01..T-05 는 독립 실행 가능 (병렬). 단, 같은 파일을 수정하므로 순차 권장.
- T-06은 T-01..T-05의 RED 상태 확인 후 시작.
- T-07은 T-06 GREEN 확인 후 시작 (코드와 docs 일관성).
- T-08은 T-07 mirror 동기화 확인 후.
- T-09는 모든 task 완료 후 final gate.

---

## 4. Commit Strategy

### 4.1 Single Commit (선택)

**Rationale**: 변경 범위 좁고 의존성 강하게 결합. PR review에서 한 번에 검토하는 것이 효율적.

**Commit Message**:
```
spec(design-folder): SPEC-V3R3-DESIGN-FOLDER-FIX-001 — reserved collision update path warning 격하

moai update 실행 시 .moai/design/ 내 reserved filename (tokens.json,
components.json, import-warnings.json, brief/BRIEF-*.md) 충돌 발견 시
hard error로 abort 대신 warning 출력 + 해당 파일 skip + 다른 templates
sync 계속 진행하도록 동작 격하.

scaffold path (moai init)는 기존 hard error 동작 유지하여 신규 사용자
혼란 방지. 사용자 데이터는 어떤 경우에도 수정·삭제되지 않음.

변경 사항:
- internal/cli/design_folder.go: checkReservedCollision에 strict bool
  파라미터 추가, updateDesignDir 호출을 strict=false로 변경
- internal/cli/design_folder_test.go: 6 ACs 커버하는 테스트 추가/갱신
- .claude/rules/moai/design/constitution.md: v3.3.1 amendment, §3.2
  footnote 추가
- internal/template/templates/.claude/rules/moai/design/constitution.md:
  template-first mirror

User-reported bug: ~/MoAI/mo.ai.kr 프로젝트에서 v2.15+ 업데이트 시
이전 버전에서 정상 생성된 tokens.json이 reserved name 정책 위반으로
판정되며 moai update 전체가 abort되던 문제 해결.

Refs: SPEC-DESIGN-CONST-AMEND-001 (reserved name 정책 출처)
Refs: SPEC-DESIGN-001 (scaffoldDesignDir/updateDesignDir 원본)

🗿 MoAI <email@mo.ai.kr>
```

### 4.2 Branch / PR

- Branch: `fix/SPEC-V3R3-DESIGN-FOLDER-FIX-001` (현재 활성)
- PR base: `main`
- Merge strategy: **squash** (CLAUDE.local.md §18.3 — feature/fix → main = squash)
- Labels: `type:fix`, `priority:P0`, `area:cli`, `area:templates`

---

## 5. Verification Checklist

### 5.1 Pre-Commit
- [ ] T-01 ~ T-05 RED 상태 확인 후 T-06 시작
- [ ] T-06 후 모든 design_folder 테스트 GREEN
- [ ] T-07 두 constitution 파일 byte-identical
- [ ] T-08 `make build` 성공 + embedded.go 갱신
- [ ] T-09 모든 quality gate green

### 5.2 Pre-PR
- [ ] commit message 한국어 body + Conventional Commits
- [ ] CHANGELOG.md entry 검토 (별도 release SPEC에서 작성 가능)
- [ ] 사용자 시나리오 수동 재현 완료
- [ ] 3축 label 부착 (type/priority/area)

### 5.3 Post-Merge
- [ ] 다음 release (v2.15.x patch 또는 v2.16.0) release note에 fix 명시
- [ ] User feedback 모니터링 (다음 release cycle)

---

Version: 0.1.0
Last Updated: 2026-04-26
Total Tasks: 9
Estimated Commits: 1 (single squash merge)
