# SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — Acceptance Criteria

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. 5개 binary AC (Given-When-Then 형식) + Edge Cases + Definition of Done + Quality Gates. 모든 AC는 단일 verification command로 PASS/FAIL 결정 가능. |

---

## 1. Overview

본 acceptance.md는 SPEC-V3R4-LINT-SPECID-GREP-FIX-001의 5개 EARS requirements가 implementation 완료 시점에 검증 가능함을 정의한다. 각 AC는 Given-When-Then 시나리오 + 구체적 verification command + 기대 출력을 포함한다.

### 1.1 AC ↔ REQ Mapping

| AC ID         | REQ ID(s)                              | Wave        | Verification scope |
|---------------|----------------------------------------|-------------|--------------------|
| AC-LSGF-001   | REQ-LSGF-001                           | W2 (GREEN)  | End-to-end: `moai spec lint --strict` |
| AC-LSGF-002   | REQ-LSGF-001, REQ-LSGF-002, REQ-LSGF-004 | W1 (RED) + W2 (GREEN) | Unit: walker direct call |
| AC-LSGF-003   | REQ-LSGF-005                           | W1 (RED) + W2 (GREEN) | Regression: chore-skip unchanged |
| AC-LSGF-004   | REQ-LSGF-002, REQ-LSGF-003, REQ-LSGF-004 | W1 (RED) + W2 (GREEN) | Unit: word-boundary precision |
| AC-LSGF-005   | (all)                                  | W3 (REFACTOR) | Suite: race-safe full test |

---

## 2. Acceptance Scenarios

### AC-LSGF-001 — End-to-end lint clean (binary AC)

**maps REQ-LSGF-001**

**Given**:
- main HEAD에 SPEC-V3R4-HARNESS-001/002/003 (각 frontmatter `status: completed`) 및 PR #944로 머지된 ea1c10647 commit (`plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 — ...`) 이 존재한다.
- 본 SPEC의 walker fix가 main에 적용된 직후 상태.

**When**:
- `moai spec lint --strict 2>&1 | tail -3` 을 실행한다.

**Then**:
- 출력 마지막 줄이 정확히 `0 error(s), 0 warning(s)` 이어야 한다.
- exit code = 0 이어야 한다.

**Verification Command**:
```bash
moai spec lint --strict 2>&1 | tail -3
# Expected last line: 0 error(s), 0 warning(s)
echo $?
# Expected: 0
```

**Pre-fix baseline (2026-05-16 측정)**: `0 error(s), 3 warning(s)` (HARNESS-001/002/003 각각 1건). Post-fix 기대: `0 error(s), 0 warning(s)`.

---

### AC-LSGF-002 — Walker returns genuine signal (not substring noise)

**maps REQ-LSGF-001, REQ-LSGF-002, REQ-LSGF-004**

**Given**:
- main git log에 다음 commit 들이 존재 (newest-first):
  1. `ea1c10647 plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 — ...` (substring noise)
  2. `19957efd8 sync(specs): V3 final status closeout (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed) (#927)` (genuine completion signal for HARNESS-001)
- walker가 `getGitImpliedStatus("SPEC-V3R4-HARNESS-001")` 으로 호출된다.

**When**:
- 새로 도입된 word-boundary 필터가 ea1c10647의 subject/body에서 `SPEC-V3R4-HARNESS-001`을 exact-match로 찾지 못해 skip한다.
- 19957efd8이 genuine completion signal commit으로 채택된다.

**Then**:
- walker 반환값이 `"completed"` 이어야 한다.
- error가 nil 이어야 한다.
- 결코 `"planned"` 를 반환하지 않아야 한다 (false-positive 차단).

**Verification (Go test)**:
```go
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
    // Given: 실제 repo에서 호출 (HEAD가 NAMESPACE plan commit 이후)
    status, err := getGitImpliedStatus("SPEC-V3R4-HARNESS-001")
    require.NoError(t, err)
    assert.Equal(t, "completed", status, "must resolve to genuine sync signal, not NAMESPACE substring noise")
}
```

---

### AC-LSGF-003 — chore-skip mechanism unchanged (backward compatibility)

**maps REQ-LSGF-005**

**Given**:
- SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001에서 도입한 `shouldSkipCommitTitle("chore(spec):", "chore(specs):")` 필터가 drift.go에 존재한다.
- 기존 회귀 가드 테스트 `TestClassifyPRTitle_ChoreSpecUnchanged` 가 `internal/spec/transitions_test.go` (또는 동등 위치)에 존재한다.

**When**:
- 본 SPEC walker fix 적용 후 기존 회귀 가드 테스트 suite를 실행한다.

**Then**:
- `TestClassifyPRTitle_ChoreSpecUnchanged` 및 LSCSK-001 관련 모든 회귀 가드 테스트가 PASS 한다.
- chore-skip + word-boundary 두 필터의 합성 결과는 chore-skip 단독 적용 결과의 상위 집합 (chore-skip이 거부한 commit은 word-boundary도 거부; 즉 chore-skip이 통과시킨 commit만 word-boundary가 추가로 평가).

**Verification Command**:
```bash
go test ./internal/spec/... -run 'TestClassifyPRTitle|TestShouldSkipCommitTitle|TestGetGitImpliedStatus' -v
# Expected: --- PASS for all LSCSK-001 regression tests
```

---

### AC-LSGF-004 — Word-boundary precision (synthetic-input unit test)

**maps REQ-LSGF-002, REQ-LSGF-003, REQ-LSGF-004**

**Given**:
- 새 unit test `TestGetGitImpliedStatus_SPECIDWordBoundary` 가 추가된다. 본 테스트는 walker의 word-boundary 매칭 로직을 직접 호출 가능한 helper (예: `commitMatchesSPECID(title, body, specID) bool`) 또는 walker 전체를 mock된 git log output에 대해 실행하여 검증한다.
- Test inputs (table-driven):

| Case | commit subject | target SPEC-ID | expected match |
|------|----------------|----------------|----------------|
| C1 | `plan(spec): SPEC-V3R4-HARNESS-001 — initial` | `SPEC-V3R4-HARNESS-001` | true |
| C2 | `plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 — supersedes 001` | `SPEC-V3R4-HARNESS-001` | false (substring noise) |
| C3 | `sync(SPEC-V3R4-HARNESS-001): status transition` | `SPEC-V3R4-HARNESS-001` | true |
| C4 | `chore(post-V3R4-HARNESS-001): cleanup` | `SPEC-V3R4-HARNESS-001` | false (no SPEC- prefix in this token) |
| C5 | `sync(specs): closeout (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed)` | `SPEC-V3R4-HARNESS-001` | false (HARNESS-001 without SPEC-V3R4- prefix is not exact match) |

> NOTE C5: subject body에 `HARNESS-001` 만 단독 존재할 경우 exact `SPEC-V3R4-HARNESS-001`과 매칭되지 않아야 함. 단, 동일 commit이 body에 full `SPEC-V3R4-HARNESS-001` 을 포함하면 true. Approach B의 `ExtractSPECIDs(title+body)` 가 자연스럽게 이를 처리.

**When**:
- 각 case를 word-boundary 매칭 함수에 입력한다.

**Then**:
- 각 case의 expected match와 실제 결과가 일치해야 한다.
- 모든 case PASS.

**Verification (Go test)**:
```bash
go test ./internal/spec/... -run TestGetGitImpliedStatus_SPECIDWordBoundary -v
# Expected: --- PASS, all 5 sub-tests
```

---

### AC-LSGF-005 — Full suite + race-safe (regression-wide guard)

**maps (all REQ)**

**Given**:
- 본 SPEC fix가 main에 push되기 직전.

**When**:
- `go test ./internal/spec/... -race -count=1` 을 실행한다.

**Then**:
- 모든 test PASS (0 fail, 0 race-detected).
- 새 test 2건 (`TestGetGitImpliedStatus_SPECIDWordBoundary`, `TestGetGitImpliedStatus_HARNESS001Resolution`) 도 포함하여 GREEN.

**Verification Command**:
```bash
go test ./internal/spec/... -race -count=1 -v 2>&1 | tail -10
# Expected: PASS, ok    github.com/modu-ai/moai-adk/internal/spec
```

---

## 3. Edge Cases

### EC-001 — SPEC-ID가 commit message에 단 한 번도 등장하지 않는 경우

**Scenario**: 새로 생성된 SPEC (예: 본 SPEC의 plan 단계 직후) 은 main에 commit이 0건이다.
**Expected**: walker가 `"", error("no git history found for ...")` 반환. lint rule이 이를 skip 처리. False-positive WARNING 0건.
**Existing behavior (pre-fix)**: 동일. Fix는 이 동작에 영향을 주지 않는다.

### EC-002 — 정확히 일치하는 commit + substring noise commit이 둘 다 N=50 window 내 존재

**Scenario**: AC-LSGF-002와 동일. Approach B 채택 시 word-boundary 필터가 substring noise를 skip하고 genuine commit에 도달.
**Expected**: walker가 genuine commit의 status 반환.

### EC-003 — 모든 N=50 commit이 word-boundary skip되는 경우 (모두 substring noise)

**Scenario**: SPEC-ID `SPEC-X-001` 검색 시 N=50 commit 모두가 `SPEC-X-NAMESPACE-001`/`SPEC-X-FOO-001` 등 prefix-collision commit만 존재.
**Expected**: walker가 LSCSK-001과 동일한 fail-safe path로 fallback — `error("no classifiable commit within window of 50 for ...")` 반환. lint rule이 skip 처리 (finding emit 안 함). False-positive 0건.
**Existing behavior**: LSCSK-001 도입 시 이미 동일 fail-safe 경로 존재.

### EC-004 — SPEC-ID에 정규식 메타문자가 포함된 경우

**Scenario**: SPEC-ID 명명 규칙 (`^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) 에 따라 메타문자 (`.`, `*`, `+`, `(`, `)` 등) 는 등장할 수 없음 (영문 대문자 + 숫자 + hyphen만).
**Expected**: 정규식 escape 불필요. 단, defense-in-depth를 위해 `regexp.QuoteMeta(specID)` 를 사용 권장.

### EC-005 — chore-skip + word-boundary 동시 필터 순서

**Scenario**: walker 단일 iteration 안에서 두 필터가 순차 적용된다.
**Decision**: chore-skip이 word-boundary보다 cheaper (string prefix check) 이므로 **chore-skip 먼저, word-boundary 다음** 순서 적용. Performance 무의미 (둘 다 microsecond) 하지만 일관성을 위해 cheaper-first 원칙.

### EC-006 — 본 SPEC의 spec.md 자체가 lint 대상이 됨

**Scenario**: 본 SPEC plan-PR 머지 후 lint walker가 `SPEC-V3R4-LINT-SPECID-GREP-FIX-001` 을 검색한다. plan-PR 머지 직후에는 plan commit 1건만 존재 → walker가 `"planned"` 반환 → frontmatter `status: draft` (plan-phase 작성 시점) 와 mismatch.
**Mitigation**: plan-PR squash-merge 직후 frontmatter `status: draft → planned` 자동 sync는 본 SPEC 범위 외 (별도 mechanism). 본 SPEC 자체에 대한 WARNING은 일시적이며 run-PR 머지 후 `"in-progress"` → sync-PR 머지 후 `"completed"` 로 자연 해소.
**Acceptance**: 본 SPEC 자체의 lint 상태는 AC-LSGF-001 final verification 시점 (sync-PR 머지 후) 에 0 WARNING 이면 충족.

---

## 4. Definition of Done

본 SPEC의 implementation이 "완료"로 간주되려면 다음 모든 조건이 충족되어야 한다:

- [ ] **AC-LSGF-001** PASS: `moai spec lint --strict` 출력이 `0 error(s), 0 warning(s)`.
- [ ] **AC-LSGF-002** PASS: `TestGetGitImpliedStatus_HARNESS001Resolution` GREEN.
- [ ] **AC-LSGF-003** PASS: LSCSK-001 regression guard tests 전부 GREEN.
- [ ] **AC-LSGF-004** PASS: `TestGetGitImpliedStatus_SPECIDWordBoundary` 5 sub-cases 전부 GREEN.
- [ ] **AC-LSGF-005** PASS: `go test ./internal/spec/... -race -count=1` 전부 GREEN.
- [ ] HARNESS-001/002/003 spec.md frontmatter 미수정 (walker fix만으로 해소).
- [ ] `lint.skip` 추가 없음.
- [ ] `transitions.go` 변경 없음.
- [ ] `@MX:NOTE` / `@MX:REASON` 갱신 — walker filter 변경 사유 + 참조 SPEC-ID 명시.
- [ ] `.claude/rules/moai/development/spec-frontmatter-schema.md` (SSOT) 무수정.
- [ ] plan PR, run PR, sync PR 3건 모두 main에 squash-merge 완료.

---

## 5. Quality Gates

### QG-1 (plan-phase)
- plan-auditor score ≥ 0.85.
- 본 acceptance.md 의 5개 AC + 6개 Edge Cases가 모두 binary/testable.

### QG-2 (run-phase, before merge)
- `go vet ./...` clean.
- `golangci-lint run ./internal/spec/...` clean.
- `go test ./internal/spec/... -race -count=1` GREEN.
- AC-LSGF-001~005 모두 PASS.

### QG-3 (sync-phase, after merge)
- main HEAD에서 `moai spec lint --strict` 0 ERROR / 0 WARNING.
- 본 SPEC 자체의 frontmatter status가 `completed` 로 sync 됨 (sync-PR 머지 후).
- HARNESS-001/002/003 의 status drift WARNING 영구 해소 확인.

---

## 6. Out-of-AC (의도적 비검증)

- ea1c10647 commit 자체의 적절성 (NAMESPACE-001 plan-PR 메시지 형식) — git history는 immutable.
- 다른 SPEC walker (예: manager-docs 가 사용하는 git log 호출) — 별도 audit SPEC 후보.
- 197 SPECs 전체 frontmatter cleanup — 본 SPEC scope 외.
- ELT/통계 메트릭 (예: walker 평균 latency, P95) — 본 SPEC은 correctness만 검증.
