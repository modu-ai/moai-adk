# Plan — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.1   | 2026-05-15 | plan-audit remediation | plan-auditor 0.924 P1+P2 cosmetic 결함 4건 반영 (sibling 문서 동기화). P1: design.md §3.2 build→refactor 교정 + §6.5 ack. P2: spec.md REQ-002 §3.4 재배치, REQ-010 EARS Optional 재작성, research.md transitions.go 라인 인용 보정. plan.md 본문 변경 없음. |
| 0.1.0   | 2026-05-15 | manager-spec  | 초기 plan. Walker filter 알고리즘 + 4-milestone TDD 시퀀스 + 7건 WARNING → 0 검증 계획 수립. BODP plan-in-main 원칙 적용 (signals A=¬ B=¬ C=¬). |

---

## 1. Implementation Strategy

### 1.1 핵심 알고리즘 — Walker Filter

현재 `getGitImpliedStatus` (drift.go:96-143)는 다음과 같다:

1. `git log <branch> --oneline --no-merges --grep=<specID> -1` → **단일** commit
2. commit title을 `ClassifyPRTitle`에 전달
3. status가 비어 있으면 `"in-progress"`로 fallback

본 SPEC은 이 함수를 다음과 같이 변경한다:

1. `git log <branch> --oneline --no-merges --grep=<specID> -<N>` → **최대 N개** commit
2. newest → oldest 순회하면서:
   - commit title이 skip pattern에 매칭되면 다음 commit으로 이동
   - 매칭되지 않으면 `ClassifyPRTitle` 호출
   - 반환된 status가 비어 있지 않으면 그 status 반환 + 종료
   - 반환된 status가 비어 있으면 다음 commit으로 이동 (안전망 — skip pattern 외에도 unknown prefix 보호)
3. 모든 N개 commit을 소진해도 의미 있는 분류를 못 찾으면 unknown signal 반환 (error)
4. `StatusGitConsistencyRule::Check`는 error를 받으면 finding을 emit하지 않는다 (현 line 897-900 동작 그대로)

### 1.2 Skip Pattern 정의

설계 결정 (design.md §3 상세):

- **항상 skip**: `^chore(spec):` (대소문자 무관, 옵션으로 `chore(specs):` 도 포함)
- **추가 skip 후보 (plan.md §7 OQ3에서 결정)**: `^sync(spec`, `^docs(sync)` — 후자는 이미 `ClassifyPRTitle`이 `(sync-merge, "completed")`를 반환하므로 skip 불요. `sync(spec)` 는 transitions.go에 없는 prefix이므로 unknown 분류되며 fallback의 영향을 받지 않는다 → skip 불요로 판정 (research.md §3 분석 기반).

따라서 v2.20.0-rc1 에서는 skip pattern을 **`chore(spec):` 단일 패턴** (case-insensitive, `chore(specs):` 변형 포함)으로 한정한다. externalization은 deferred (§7 OQ2).

### 1.3 Walk Depth (N)

추천 N = 50 (§7 OQ1 결정).

근거:

- SPEC별 git 이력은 평균 5-10건 commit (plan / run / sync 각 1건 + chore sweep 누적)
- 7건 영향 SPEC 중 SPEC-V3R2-CON-001 의 git log 매칭 commit이 최대 ~15건으로 추정 (research.md §2 보조 데이터)
- N=50은 약 3x 안전 여유 — 미래에 sweep이 더 누적되어도 의미 있는 commit을 놓치지 않음
- N이 너무 커지면 git log subprocess 비용이 증가하지만, 50건은 < 50ms 수준이며 lint 전체 시간 (수 분) 대비 무시 가능

### 1.4 Error vs Sentinel 신호

`getGitImpliedStatus`는 walker 소진 시 다음 옵션 중 선택:

- (a) Go `error` 타입을 반환 (예: `errors.New("no classifiable commit within window")`)
- (b) sentinel 문자열 `"unknown"` 을 반환하고 nil error

선택: **(a) error**. 근거:

- 기존 함수 signature `(string, error)` 호환
- `StatusGitConsistencyRule::Check` line 897-900가 이미 `err != nil`이면 skip 처리 → 추가 코드 변경 없음
- "unknown"이라는 자기설명적 문자열을 valid status enum과 섞으면 다른 호출처에서 혼선 위험

---

## 2. Milestones

### M1 — RED Phase: Failing Tests 작성

작업 내용:

- 새 test 파일 `internal/spec/drift_chore_skip_test.go` 작성 (또는 기존 `internal/spec/drift_test.go` 가 있다면 확장)
- 4 test case 모두 RED 상태로 작성 (AC-LSCSK-005 a/b/c/d)
- 각 test case는 `t.TempDir()` 하에 git 저장소를 초기화하고 commit fixture를 작성한 뒤 `getGitImpliedStatus` 호출
- `go test ./internal/spec/ -run TestGetGitImpliedStatus_ChoreSkip`로 실행 시 4건 모두 FAIL 확인

확인 지표:

- `go test ./internal/spec/ -run TestGetGitImpliedStatus_ChoreSkip -v` 출력에서 4건 FAIL
- 기존 테스트는 모두 PASS (regression 없음)

### M2 — GREEN Phase: Walker Filter 구현

작업 내용:

- `internal/spec/drift.go::getGitImpliedStatus` 함수 본문을 walker filter 로직으로 교체
- 새 helper 함수 `shouldSkipCommitTitle(title string) bool` 추가 (skip pattern 매칭)
- `git log ... -1` → `git log ... -50` 변경 (N=50)
- 결과 multi-line output을 `bufio.Scanner` 로 한 줄씩 처리 (newest first)
- 빈 status를 받으면 다음 commit으로 이동 (continue)
- 의미 있는 status를 받으면 즉시 반환
- 50개 모두 소진되면 `fmt.Errorf("no classifiable commit within window for %s", specID)` 반환

확인 지표:

- M1에서 작성한 4 test case 모두 PASS
- `go test ./internal/spec/...` 전체 PASS
- `go test -race ./internal/spec/` PASS

### M3 — REFACTOR + Edge Cases

작업 내용:

- godoc 갱신 (Korean — code_comments=ko)
- `shouldSkipCommitTitle` 의 regex 또는 prefix match 구현 결정 (간단한 strings.HasPrefix + ToLower 권장)
- Edge case 점검:
  - SPEC에 commit이 0건인 경우 (현재 line 113-115에서 처리 중) → walker 로직과 호환 유지
  - 모든 commit이 chore인 경우 (case c) → walker 소진 → error → 상위 rule이 skip
  - merge commit 처리 — 현재 `--no-merges` 플래그가 있으므로 merge commit은 자동 제외 → 변경 불요
- `golangci-lint run ./internal/spec/` 0 warning 확인

확인 지표:

- `go test ./internal/spec/...` 100% PASS
- `golangci-lint run` 0 warning
- godoc Korean 본문 확인

### M4 — End-to-End 검증

작업 내용:

- 현재 main (`bdcb57f8d`) 에서 `moai spec lint --strict` 실행 → 7건 WARNING 확인 (baseline)
- 본 PR feature 브랜치에서 `moai spec lint --strict` 실행 → 7건 WARNING 모두 사라짐 확인 (AC-LSCSK-001)
- `moai spec lint --strict` 전체 output이 0 ERROR + 0 WARNING (또는 다른 출처의 WARNING만 잔존하고 7건 status 관련은 0건) 확인
- 51개 SPEC의 `lint.skip: [StatusGitConsistency]` 엔트리는 그대로 유지 — 제거 작업은 별도 cleanup task로 분리 (REQ-LSCSK-009 / scope §6 out-of-scope #3)

확인 지표:

- 영향 7건 SPEC에 대한 `StatusGitConsistency` WARNING이 0건
- 다른 SPEC에서 신규 false-positive WARNING이 생기지 않음 (regression check)
- CHANGELOG `[Unreleased]` 에 본 SPEC 항목 추가됨

---

## 3. Files to be Modified

| Path | Line Range | Change Type | Description |
|------|-----------|-------------|-------------|
| `internal/spec/drift.go` | 96-143 | Modify | `getGitImpliedStatus` 함수 본문 교체 — walker filter 로직 적용 |
| `internal/spec/drift.go` | (new) | Add | `shouldSkipCommitTitle(title string) bool` helper 함수 추가 |
| `internal/spec/lint.go` | 879-914 | (no change expected) | `StatusGitConsistencyRule::Check` 는 이미 `err != nil`이면 skip → 코드 수정 없을 예정. 검증만 수행. |
| `internal/spec/transitions.go` | 22 | (no change) | `chore(spec)` 분류 규칙 보존 (AC-LSCSK-003 regression guard) |
| `CHANGELOG.md` | `[Unreleased]` section | Add | 본 SPEC 변경 entry 추가 (sync-phase 책임이나 run-phase에서 [Unreleased] entry 추가 권장) |

---

## 4. Files to be Created

| Path | Purpose |
|------|---------|
| `internal/spec/drift_chore_skip_test.go` | 4개 새 test case (AC-LSCSK-005 a/b/c/d). 기존 drift_test.go가 있다면 확장하는 것도 무방하지만, 신규 파일로 분리하면 PR diff 가독성이 향상된다. |

`drift_test.go` 가 이미 존재하면 그 파일에 새 test case를 추가하고 별도 파일은 생성하지 않는다 (M1 시점에 결정).

---

## 5. Testing Strategy

### 5.1 Test Framework

- Go 표준 `testing` 패키지
- Table-driven test 권장 (.claude/rules/moai/languages/go.md "Testing" 섹션 준수)
- `t.TempDir()` + `os/exec`로 임시 git 저장소 초기화 (CLAUDE.local.md §6 준수)
- `t.Parallel()` 적용 가능 — 각 테스트는 독립 임시 디렉토리 사용

### 5.2 Git Fixture 패턴

각 test case는 다음 단계를 수행한다:

1. `t.TempDir()` 으로 임시 디렉토리 확보
2. `git init` + `git config user.email/name` 로컬 설정 (CI environment 무관)
3. `git commit --allow-empty -m "<title>"` 으로 commit fixture 작성 (oldest → newest 순서)
4. `git checkout` 또는 `os.Chdir(tmpDir)`로 작업 디렉토리 전환 (단, 다른 테스트와 격리 필요 — `t.Helper()` + 원복)
5. `getGitImpliedStatus("SPEC-XXX")` 호출
6. 반환값 verify (status string + error nil/non-nil)

**중요**: `getGitImpliedStatus`는 현재 작업 디렉토리의 git 저장소를 사용한다 (drift.go line 102 `exec.Command("git", ...)`). 따라서 테스트 시 `os.Chdir()` 가 필요하다. 또는 함수 signature를 확장하여 dir 파라미터를 받는 리팩토링도 가능하나, 이는 scope 확장이므로 chdir 방식을 우선 적용한다 (run-phase에서 결정).

### 5.3 4 Test Cases (AC-LSCSK-005)

**Case (a) — sweep commit hides real impl commit**:

- Commit 1 (oldest): `impl(spec): SPEC-UTIL-001 implementation` → status="implemented"
- Commit 2 (newest): `chore(spec): status drift sweep` body에 `SPEC-UTIL-001` 언급
- Expected: walker가 chore commit을 skip하고 commit 1로 이동, status="implemented" 반환

**Case (b) — chore precedes feat**:

- Commit 1 (oldest): `chore(spec): metadata cleanup` body에 `SPEC-X-001` 언급
- Commit 2 (newest): `feat(SPEC-X-001): new feature` → status="implemented"
- Expected: walker가 commit 2를 newest first 로 보고 `feat`을 valid 분류 → status="implemented" 즉시 반환 (skip 무관)

Note: case (b)의 "preceds" 표현은 시간 순서이지 walker 순회 순서가 아니다. 실제 walker는 newest first.

**Case (c) — only chore commits**:

- Commit 1, 2, 3: 모두 `chore(spec): ...` 형태
- Expected: walker가 모두 skip → 50개 budget 내 모두 소진 → error 반환 → `StatusGitConsistencyRule::Check`는 finding을 emit하지 않음

**Case (d) — control case: latest is real impl**:

- Commit 1 (only): `impl(spec): SPEC-Z-001 implementation` → status="implemented"
- Expected: walker가 첫 commit에서 즉시 status="implemented" 반환, skip filter 미발동

### 5.4 Regression Test

`internal/spec/transitions_test.go` (기존 파일) 에 `TestClassifyPRTitle_ChoreSpecUnchanged` 추가:

```go
// AC-LSCSK-003 regression guard
func TestClassifyPRTitle_ChoreSpecUnchanged(t *testing.T) {
    // chore(spec): 분류는 skip-meta 카테고리 + 빈 status를 반환해야 한다 (의도된 설계)
    category, status, err := ClassifyPRTitle("chore(spec): status drift sweep")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if category != "skip-meta" {
        t.Errorf("category: got %q, want %q", category, "skip-meta")
    }
    if status != "" {
        t.Errorf("status: got %q, want empty string", status)
    }
}
```

### 5.5 Coverage Target

- 새 `shouldSkipCommitTitle` 함수: 100% line coverage (true/false 두 케이스 모두 테스트)
- `getGitImpliedStatus` 수정 부분: 기존 coverage 유지 + walker 분기 신규 커버
- 전체 `internal/spec/` package: 85% 이상 유지 (CLAUDE.local.md §6 기준)

---

## 6. Definition of Done

- [ ] AC-LSCSK-001 충족: feature 브랜치에서 `moai spec lint --strict` 실행 시 7건 WARNING이 0건으로 감소 (manual verification + CI 결과 첨부)
- [ ] AC-LSCSK-002 충족: `drift_chore_skip_test.go` test case (a) PASS — sweep commit이 real impl을 hide하는 시나리오에서 정확한 status 반환
- [ ] AC-LSCSK-003 충족: `transitions_test.go` regression test `TestClassifyPRTitle_ChoreSpecUnchanged` PASS
- [ ] AC-LSCSK-004 충족: walker N=50 boundary 검증 — 50건 모두 chore 인 경우 error 반환 + lint rule skip
- [ ] AC-LSCSK-005 충족: 4 test case (a/b/c/d) + 기존 `internal/spec/*_test.go` 모두 PASS, `go test ./...` 전체 GREEN
- [ ] `golangci-lint run ./internal/spec/` 0 warning, `go vet ./...` 0 issue, `go test -race ./internal/spec/` PASS
- [ ] CHANGELOG.md `[Unreleased]` section 에 entry 추가 — 형식 예시: `### Fixed - spec-lint: chore(spec) sweep commit이 status drift WARNING을 유발하던 bootstrapping bug 해소 (SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001)`
- [ ] @MX:NOTE 또는 @MX:ANCHOR 태그가 새 `shouldSkipCommitTitle` 함수와 수정된 `getGitImpliedStatus`에 적절히 부착 (mx-tag-protocol.md 준수, 한국어 description, fan_in ≥ 3 검토 후 ANCHOR 여부 결정)

---

## 7. Open Questions & Decisions

### OQ1: Walker Depth N

**Decision**: N = 50

**Rationale**: 평균 SPEC 당 git log 매칭 commit이 5-10건. 50은 약 5-10x 안전 여유로, 미래 sweep 누적에도 견딘다. 50건 git log subprocess 비용은 < 50ms (벤치 추정) — lint 전체 시간 (수 분) 대비 무시 가능. 100 이상으로 키울 명확한 근거가 없으므로 50으로 고정.

**Alternative considered**: N = 20 (보수적), N = 100 (공격적). 20은 deep history SPEC에서 위험, 100은 cost-benefit 정당화 어려움.

### OQ2: Skip Pattern 외부 설정화

**Decision**: Hard-coded in Go source for v2.20.0-rc1. Externalization to `.moai/config/sections/spec-lint.yaml` `git_status_skip_patterns:` is **deferred** to future SPEC.

**Rationale**:

- 현재 skip pattern은 `chore(spec):` 단일 — externalization으로 얻는 유연성보다 코드 단순성이 우선
- 미래에 새 prefix (예: `chore(specs):`, `rebase(spec):`, `revert(spec):`) 가 도입되면 별도 SPEC으로 패턴 확장 + 동시에 외부화 결정
- 외부화 시 `spec-lint.yaml` 파일 자체가 새 정합성 검증 대상이 되어 회귀 위험 증가

**Alternative considered**: 외부 설정화 즉시 도입 — rejected. KISS principle + 단일 패턴은 hard-code가 적절.

### OQ3: `sync(spec...)` 도 skip 대상인가?

**Decision**: 아니오. v2.20.0-rc1 에서는 `chore(spec):` 단일 패턴만 skip한다.

**Rationale** (research.md §3 분석 기반):

- `sync(spec)` 또는 `sync(spec-XXX)` prefix는 `transitions.go` 에 등록되지 않은 unknown prefix → `ClassifyPRTitle` 이 `("unknown", "", nil)` 반환 → drift.go line 134-136 fallback에 의해 `"in-progress"` 반환
- 그러나 sync-phase는 일반적으로 `docs(sync):` prefix를 사용하며 이는 transitions.go line 26-27에 의해 `("sync-merge", "completed", nil)` 로 정확히 분류됨 — fallback에 의존하지 않음
- 따라서 `sync(spec):` 형태의 prefix는 실무에서 거의 사용되지 않으며, 만약 사용된다면 별도 SPEC으로 transitions.go 규칙 추가 또는 skip pattern 확장으로 처리
- 본 SPEC scope를 좁게 유지하여 회귀 위험 최소화

**Alternative considered**: `^sync(` 도 skip pattern에 포함 — rejected. 위 분석에서 fallback 트리거 빈도 낮음 + scope creep 위험.

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Walker N=50이 일부 deep-history SPEC에서 부족 | Low | Low (false-positive WARNING → lint.skip로 우회 가능) | M4 검증 시 모든 SPEC에 대해 신규 false-positive 발생 여부 확인 |
| `os.Chdir()` 기반 test가 다른 parallel test와 충돌 | Low | Medium (flaky test) | t.Parallel() 미적용 또는 chdir 후 t.Cleanup으로 원복 |
| `chore(spec):` 외에 다른 prefix가 새로운 sweep 도구로 도입 | Low | Medium (재발) | OQ2 deferred externalization을 미래 SPEC에서 즉시 진행 |
| `git log --grep` 가 plaintext 매칭이라 false-positive (다른 SPEC body 언급) | Low | Low (해당 SPEC도 valid commit history 일부) | 추가 검증 불필요 — 현행 동작과 일관성 유지 |
| 새 helper 함수 `shouldSkipCommitTitle`이 향후 변경 시 회귀 | Medium | Low | 100% line coverage + table-driven tests로 보장 |

---

## 9. Verification Checklist (M4 Pre-Merge)

- [ ] `git log --oneline -5` 가 본 PR feature 브랜치에서 정상 출력됨
- [ ] `go test ./...` 전체 PASS
- [ ] `go test -race ./internal/spec/` PASS
- [ ] `go vet ./...` 0 issue
- [ ] `golangci-lint run` 0 warning
- [ ] `moai spec lint --strict` 실행 시 0 ERROR + 0 WARNING (또는 unrelated WARNING 만 잔존)
- [ ] 7건 영향 SPEC 모두에서 `StatusGitConsistency` WARNING 0건
- [ ] 51개 SPEC의 `lint.skip: [StatusGitConsistency]` entry는 그대로 유지 (제거 작업은 별도 task)
- [ ] CHANGELOG.md `[Unreleased]` entry 추가됨
- [ ] `internal/spec/drift.go::getGitImpliedStatus` godoc Korean 본문 갱신됨
- [ ] `shouldSkipCommitTitle` helper 함수에 @MX:NOTE 또는 적절한 태그 부착됨
