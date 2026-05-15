# Acceptance Criteria — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-15 | manager-spec | 초기 draft. REQ-LSCS-001 ~ REQ-LSCS-010 각각에 대한 Given-When-Then 시나리오 정의. lint 카테고리별 카운트 + Go 단위 테스트 통과 + CI green 을 객관적 측정 지표로 사용. |

---

## 1. Acceptance Criteria 개요

본 문서는 `spec.md` §3 의 10개 REQ-LSCS-NNN 요구사항을 검증한다. 각 AC 는 독립적으로 테스트 가능하며, 다음을 객관적 측정 지표로 사용한다:

- `moai spec lint --strict` 의 카테고리별 카운트 (StatusGitConsistency 0 검증)
- `go test ./internal/spec/...` 단위 테스트 결과
- `go test -race ./internal/spec/...` race 검출 결과
- spec-lint CI job 의 GREEN 여부

검증 기준선: origin/main commit `bdcb57f` (PR #930 머지 직후) 기준 StatusGitConsistency 신규 7건 WARN + 51건 lint.skip suppress.

목표 상태: `moai spec lint --strict` exit 0, 신규 7건 WARN 0건 (logic fix), 51건 lint.skip suppress 유지 (회귀 없음).

---

## 2. Acceptance Criteria 상세

### AC-LSCS-001 (REQ-LSCS-001: chore(spec) sweep skip)

**Given** `internal/spec/drift.go::getGitImpliedStatus` 가 SPEC-ID `SPEC-V3R4-TEST-001` 에 대해 호출되고, git log 에 다음 commit 들이 있는 상태 (최신 → 오래된 순):

```
chore(spec): status drift sweep mentioning SPEC-V3R4-TEST-001
feat(SPEC-V3R4-TEST-001): implement feature
```

**When** `getGitImpliedStatus("SPEC-V3R4-TEST-001")` 가 호출되면,
**Then** 반환 값은 `"implemented"`, `nil` 이어야 한다. `chore(spec):` prefix 의 latest commit 은 건너뛰고 직전 `feat(...)` commit 이 status 분류 대상이 된다.

검증 명령:
```bash
go test -run "TestGetGitImpliedStatus/chore_sweep_followed_by_latest_feat" ./internal/spec/...
# expected: PASS
```

---

### AC-LSCS-002 (REQ-LSCS-002: revert commit skip)

**Given** SPEC-ID `SPEC-V3R4-TEST-003` 의 git log 가 다음 상태 (최신 → 오래된 순):

```
revert: revert prior change mentioning SPEC-V3R4-TEST-003
feat(SPEC-V3R4-TEST-003): implement feature
```

**When** `getGitImpliedStatus("SPEC-V3R4-TEST-003")` 가 호출되면,
**Then** 반환 값은 `"implemented"`, `nil` 이어야 한다. `revert:` prefix 의 latest commit 은 건너뛰고 직전 `feat(...)` commit 이 status 분류 대상이 된다.

검증 명령:
```bash
go test -run "TestGetGitImpliedStatus/revert_followed_by_feat" ./internal/spec/...
# expected: PASS
```

---

### AC-LSCS-003 (REQ-LSCS-003: Skip-meta/no-op 정직한 해석)

**Given** `internal/spec/transitions.go::ClassifyPRTitle` 가 `chore(spec):` 입력에 대해 `("skip-meta", "", nil)` 을 반환하고, `revert:` 입력에 대해 `("no-op", "", nil)` 을 반환하는 상태,

**When** `getGitImpliedStatus` 가 위 두 카테고리의 commit 을 만나면,
**Then** 그 commit 은 skip 되고 다음 commit 으로 진행되어야 한다. 기존의 "status == '' → 'in-progress' 기본값" 변환 로직은 `category != "skip-meta" && category != "no-op"` 인 경우에만 적용되어야 한다.

검증 명령:
```bash
# transitions_test.go 의 회귀 케이스 통과
go test -run "TestClassifyPRTitle" ./internal/spec/...
# expected: PASS (chore(spec)/chore(specs)/revert 모두 정의된 카테고리 반환)

# drift.go 의 skip 처리 통과
grep -n "skip-meta\|no-op" internal/spec/drift.go
# expected: 1+ matches (skip 조건 분기 존재)
```

---

### AC-LSCS-004 (REQ-LSCS-004: 스캔 범위 상한 10)

**Given** SPEC-ID `SPEC-V3R4-TEST-004` 의 git log 가 다음 상태 (최신 → 오래된 순, 모두 SPEC-ID mention):

```
chore(spec): sweep 1
chore(spec): sweep 2
chore(spec): sweep 3
```

**When** `getGitImpliedStatus("SPEC-V3R4-TEST-004")` 가 호출되면,
**Then** 반환 값은 `""`, `nil` 이어야 한다. 스캔된 모든 commit 이 skip-meta 이므로 의미 있는 lifecycle 추론 불가 → empty status signal.

또한, `maxGitLogScan` 상수가 `internal/spec/drift.go` 에 정의되어 있고 값이 `10` 이어야 한다.

검증 명령:
```bash
go test -run "TestGetGitImpliedStatus/all_skip_meta_commits" ./internal/spec/...
# expected: PASS

grep -n "maxGitLogScan" internal/spec/drift.go
# expected: const maxGitLogScan = 10
```

---

### AC-LSCS-005 (REQ-LSCS-005: lint 게이트 우회)

**Given** `getGitImpliedStatus` 가 특정 SPEC 에 대해 `""`, `nil` 을 반환하는 상태 (스캔된 모든 commit 이 skip-meta/no-op),

**When** `StatusGitConsistencyRule.Check` 가 해당 SPEC 에 대해 호출되면,
**Then** Finding 슬라이스는 빈 슬라이스 (`nil` 또는 `[]Finding{}`) 이어야 한다. StatusGitConsistency WARN 이 emit 되지 않아야 한다.

검증 명령:
```bash
# lint.go 가드 코드 존재 확인
grep -n "gitStatus ==" internal/spec/lint.go
# expected: 1+ matches (empty string guard 존재)

# 통합 시나리오: 모든 commit 이 chore(spec) 인 SPEC 에 대해 lint 가 silent
moai spec lint --strict --filter StatusGitConsistency 2>&1 | grep -c "no meaningful"
# expected: 0 (no spurious findings)
```

---

### AC-LSCS-006 (REQ-LSCS-006: 7건 신규 WARN 해소)

**Given** origin/main commit `bdcb57f` (PR #930 머지 직후) 시점에 `moai spec lint --strict` 가 신규 7건 StatusGitConsistency WARN 을 보고하는 상태,

**When** 본 SPEC 의 run-phase 완료 후 (drift.go + lint.go 수정 + PR 머지 후) origin/main HEAD 에서 `moai spec lint --strict` 를 실행하면,
**Then** stdout 의 StatusGitConsistency 카운트는 0 이어야 한다.

검증 명령:
```bash
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"
# expected: 0

moai spec lint --strict
echo "exit code: $?"
# expected: 0
```

---

### AC-LSCS-007 (REQ-LSCS-007: 회귀 방지 단위 테스트)

**Given** `internal/spec/drift_test.go` 신규 파일이 5+ 시나리오를 table-driven test 로 커버하는 상태:

1. chore(spec) 단독 latest + feat 직전 → implemented
2. chore(spec) 단독 latest + sync 직전 → completed
3. revert 단독 latest + feat 직전 → implemented
4. 모두 chore(spec) (max scan 범위 내) → "" (empty)
5. vanilla feat (skip 없음) → implemented
6. (선택) SPEC-ID mention 없는 commit 만 → error

**When** `go test -race ./internal/spec/...` 를 실행하면,
**Then** 모든 테스트가 통과해야 한다 (no race detected).

검증 명령:
```bash
go test -race ./internal/spec/...
# expected: PASS, 0 race detected

go test -v -run "TestGetGitImpliedStatus" ./internal/spec/... 2>&1 | grep -c "--- PASS"
# expected: 5+ (시나리오 개수)
```

---

### AC-LSCS-008 (REQ-LSCS-008: 기존 lint.skip 자산 보존)

**Given** SPEC-V3R4-SPECLINT-DEBT-001 Wave 2 결과로 51개 SPEC frontmatter 에 `lint.skip: [StatusGitConsistency]` 가 등록된 상태 (`lint-final.md` 기준),

**When** 본 SPEC 의 run-phase 완료 후 `moai spec lint --strict` 를 실행하면,
**Then** 51개 SPEC 모두에서 StatusGitConsistency Finding 이 emit 되지 않아야 한다 (suppress 효과 유지). 본 SPEC 의 lint logic 수정이 lint.skip 등록 메커니즘에 영향을 미치지 않아야 한다.

검증 명령:
```bash
# 51개 SPEC 의 lint.skip 등록 확인
grep -l "skip:" .moai/specs/SPEC-*/spec.md 2>/dev/null | wc -l
# expected: >= 51

# 51개 SPEC 모두에서 StatusGitConsistency Finding 없음
moai spec lint --strict --filter StatusGitConsistency 2>&1 | grep -E "SPEC-(AGENCY-ABSORB-001|V3R2-ORC-001|...)"
# expected: empty output

# 전체 lint 결과에서 StatusGitConsistency 0건
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"
# expected: 0
```

---

### AC-LSCS-009 (REQ-LSCS-009: CI Green 게이트)

**Given** 본 SPEC 의 run-phase PR 이 생성된 상태,

**When** GitHub Actions 의 spec-lint workflow 가 트리거되면,
**Then** spec-lint job 이 status 0 으로 종료해야 한다. PR 의 `Checks` 탭에서 spec-lint job 이 GREEN 으로 표시되어야 한다.

검증 명령:
```bash
gh pr checks <PR_NUMBER> --watch
# expected: spec-lint pass

gh pr view <PR_NUMBER> --json statusCheckRollup | jq '.statusCheckRollup[] | select(.name == "spec-lint")'
# expected: state == "SUCCESS"
```

PR 머지 조건: spec-lint job GREEN 필수.

---

### AC-LSCS-010 (REQ-LSCS-010: Self-coverage)

**Given** 본 SPEC (`SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001`) 자체가 spec-lint 규칙을 만족해야 하는 상태,

**When** `moai spec lint --strict --filter SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` 를 실행하면,
**Then** 본 SPEC 에 대한 ERROR 0건, WARNING 0건이어야 한다.

추가 self-check:
- `spec.md` frontmatter 가 7개 mandatory 필드 (`title`, `created`, `updated`, `phase`, `module`, `lifecycle`, `tags`) 모두 포함
- `spec.md` §1.3 Out of Scope 섹션에 최소 1개 explicit 항목 포함 (현재 6개)
- `spec.md` §3 의 REQ-LSCS-001 ~ REQ-LSCS-010 모두 본 `acceptance.md` 의 AC 에서 참조됨 (AC-LSCS-001 ~ AC-LSCS-010 으로 1:1 매핑)

검증 명령:
```bash
# Self lint
moai spec lint --strict 2>&1 | grep "SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001"
# expected: empty output (no findings)

# Frontmatter 7 fields 확인
head -25 .moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/spec.md | \
  grep -E "^(title|created|updated|phase|module|lifecycle|tags):" | wc -l
# expected: 7

# REQ-AC 매핑 확인 (1:1)
grep -c "^#### REQ-LSCS-" .moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/spec.md
# expected: 10
grep -c "^### AC-LSCS-" .moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/acceptance.md
# expected: 10
```

---

## 3. Definition of Done

본 SPEC 의 run-phase 가 완료되기 위해 다음 모두 충족:

- [ ] AC-LSCS-001 ~ AC-LSCS-010 모두 PASS
- [ ] `go test -race ./...` 전체 통과 (CLAUDE.local.md §6 의무)
- [ ] `golangci-lint run ./...` 통과
- [ ] `moai spec lint --strict` exit 0
- [ ] PR description 에 본 SPEC ID + Issue #932 reference 포함
- [ ] CHANGELOG.md `[Unreleased]` 섹션에 변경 entry 추가
- [ ] PR 머지 후 frontmatter `status: draft → implemented` 갱신 (sync-phase 에서 `completed` 로 최종 전이)
- [ ] frontmatter `version: 0.1.0 → 0.2.0` (sync-phase 에서 갱신)

---

## 4. References

- spec.md §3 (REQ-LSCS-001 ~ REQ-LSCS-010)
- plan.md §2 (T-LSCS-001 ~ T-LSCS-007)
- `internal/spec/drift.go:99-143` getGitImpliedStatus (수정 대상)
- `internal/spec/transitions.go:22, 70-92` ClassifyPRTitle (skip-meta 매핑 소스)
- `internal/spec/lint.go:879-914` StatusGitConsistencyRule (호출부)
- SPEC-V3R4-SPECLINT-DEBT-001 lint-final.md (51 lint.skip suppress baseline)
- CLAUDE.local.md §6 Testing Guidelines (`t.TempDir()` 의무)
