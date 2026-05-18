# Design — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.1   | 2026-05-15 | plan-audit remediation | P1 결함 반영: §3.2 검토 보류 패턴 표의 검증 안 된 `^build:` 인용을 `^refactor:` (transitions.go:32, 실측 검증된 prefix) 로 교체. §6.5 failure mode에 transitions.go 미등록 prefix가 fall-through할 수 있다는 ack 2-3문장 추가 (ops 가이드 옵션 포함). 동시에 sibling SPEC 문서 (spec.md REQ-002 재배치 / REQ-010 EARS Optional / research.md 라인 인용) 일괄 0.1.1 bump. |
| 0.1.0   | 2026-05-15 | manager-spec  | 초기 design. As-Is / To-Be 코드 분석 + skip pattern 정의 + walker depth 결정 + edge case 6건 + failure mode 분석. |

---

## 1. Current State (As-Is)

### 1.1 `getGitImpliedStatus` 현행 구현

`internal/spec/drift.go` lines 96-143:

```go
// getGitImpliedStatus determines the status implied by git log for a SPEC
// It scans git log on main for the latest commit mentioning the SPEC-ID
// and classifies that commit to determine the implied status
func getGitImpliedStatus(specID string) (string, error) {
    // 기본 브랜치 결정 — main 우선, 없으면 master
    branch := "main"
    if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
        branch = "master"
    }

    // 해당 SPEC-ID를 언급하는 가장 최근 commit 1건만 가져옴 — 본 SPEC이 수정할 핵심 지점
    cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges", "--grep="+specID, "-1")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("git log failed: %w", err)
    }

    if len(output) == 0 {
        return "", fmt.Errorf("no git history found for %s", specID)
    }

    // 첫 줄에서 commit title 추출 (commit hash 다음의 문자열)
    scanner := bufio.NewScanner(strings.NewReader(string(output)))
    if scanner.Scan() {
        line := scanner.Text()
        // commit hash 부분 분리
        parts := strings.SplitN(line, " ", 2)
        if len(parts) < 2 {
            return "", fmt.Errorf("invalid git log format")
        }
        commitTitle := parts[1]

        // commit title을 분류하여 status를 얻음
        _, status, err := ClassifyPRTitle(commitTitle)
        if err != nil {
            return "", fmt.Errorf("failed to classify commit: %w", err)
        }

        if status == "" {
            // ★ 핵심 결함 지점: Unknown prefix는 "in-progress"로 fallback —
            //  chore(spec): sweep commit이 이 경로로 진입하여 false-positive WARN을 만든다
            return "in-progress", nil
        }

        return status, nil
    }

    return "", fmt.Errorf("failed to parse git log output")
}
```

### 1.2 결함 동작의 트리거 시퀀스

1. PR #930 (`chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)`) 머지
2. 해당 commit (`bdcb57f8d`) 의 body에 SPEC-UTIL-001 / SPEC-V3R2-CON-001/002/003 / SPEC-V3R2-RT-001 / SPEC-V3R2-SPC-003 / SPEC-V3R4-HARNESS-003 모두 언급됨
3. `moai spec lint --strict` 실행 → `StatusGitConsistencyRule::Check` 호출 → `getGitImpliedStatus("SPEC-UTIL-001")` 호출
4. git log 명령이 `bdcb57f8d` 한 건만 반환 (이 commit이 SPEC-UTIL-001을 언급하는 가장 최근 commit이기 때문)
5. commit title `chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)`을 `ClassifyPRTitle`에 전달
6. `transitions.go` line 22 규칙 매칭 → `(category="skip-meta", status="", error=nil)` 반환
7. drift.go line 134-136 — 빈 status를 받아 `"in-progress"` 반환
8. frontmatter `status: implemented` 와 비교 → 불일치 → WARNING 생성
9. 7개 SPEC 모두 동일 시퀀스 → 7건 WARNING

이 동작은 **개별 함수 관점에서 모두 올바르다** (`ClassifyPRTitle`은 의도대로 chore를 skip-meta로 분류, drift.go는 unknown을 fallback 처리). 그러나 두 함수의 조합이 sweep commit의 본래 의도 ("이 commit은 lifecycle 추론에서 제외하라") 를 깨뜨린다. **bootstrapping bug**.

---

## 2. Target State (To-Be)

### 2.1 새 `getGitImpliedStatus` 의사 코드 (Go-like sketch)

```go
const gitLogWindowSize = 50  // OQ1 결정: walker N=50

// getGitImpliedStatus determines the status implied by git log for a SPEC.
// 본 함수는 SPEC-ID를 언급하는 git commit을 newest-to-oldest 순회하면서
// chore(spec): sweep commit 등 lifecycle 추론에서 의도적으로 제외되는 commit을 건너뛰고,
// 의미 있는 분류(`ClassifyPRTitle`이 비어있지 않은 status를 반환)를 가진 첫 commit의 status를 채택한다.
// 모든 N개 commit이 skip 대상이면 error를 반환하고, 상위 lint rule은 이를 skip 조건으로 처리한다.
func getGitImpliedStatus(specID string) (string, error) {
    // 기본 브랜치 결정 — main 우선, 없으면 master (현행 동작 유지)
    branch := "main"
    if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
        branch = "master"
    }

    // SPEC-ID를 언급하는 최대 N개 commit을 newest-first 순서로 가져옴
    cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges",
        "--grep="+specID, fmt.Sprintf("-%d", gitLogWindowSize))
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("git log failed: %w", err)
    }

    if len(output) == 0 {
        return "", fmt.Errorf("no git history found for %s", specID)
    }

    // 한 줄씩 순회 — newest first
    scanner := bufio.NewScanner(strings.NewReader(string(output)))
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.SplitN(line, " ", 2)
        if len(parts) < 2 {
            continue  // 손상된 줄은 건너뛴다
        }
        commitTitle := parts[1]

        // skip pattern 매칭 — chore(spec): 등 lifecycle 추론에서 제외되는 commit
        if shouldSkipCommitTitle(commitTitle) {
            continue
        }

        // commit title 분류
        _, status, err := ClassifyPRTitle(commitTitle)
        if err != nil {
            // 분류 실패는 안전상 skip하고 다음 commit으로 이동
            continue
        }

        if status == "" {
            // unknown prefix 안전망 — 빈 status는 의미 있는 분류로 인정하지 않고 다음 commit 탐색
            continue
        }

        // 의미 있는 분류 발견 → 즉시 반환
        return status, nil
    }

    // N개 commit 모두 소진해도 의미 있는 분류를 못 찾음 → error
    // StatusGitConsistencyRule::Check (lint.go:897-900)는 err != nil이면 finding을 emit하지 않는다
    return "", fmt.Errorf("no classifiable commit within window of %d for %s", gitLogWindowSize, specID)
}

// shouldSkipCommitTitle returns true if the commit title matches a known skip pattern
// for lifecycle status inference. Skip-pattern commits represent metadata maintenance
// operations (frontmatter sweeps, lint.skip registrations) and must not influence
// the git-implied status.
//
// v2.20.0-rc1 의 skip pattern: `chore(spec):` 만 대상.
// 향후 추가 패턴은 별도 SPEC + plan.md §7 OQ2 externalization 결정 시 확장.
func shouldSkipCommitTitle(title string) bool {
    // 대소문자 무관 prefix 매칭
    lower := strings.ToLower(strings.TrimSpace(title))
    return strings.HasPrefix(lower, "chore(spec):") ||
        strings.HasPrefix(lower, "chore(specs):")
}
```

### 2.2 동작 변화 요약

| 시나리오 | As-Is 결과 | To-Be 결과 |
|---------|-----------|-----------|
| Latest commit이 `chore(spec):` sweep, 이전에 `impl(spec):` 존재 | "in-progress" (fallback) | "implemented" (이전 commit 채택) |
| Latest commit이 `feat(SPEC-X-001):` 단일 | "implemented" | "implemented" (변화 없음) |
| Latest commit이 `chore:` (non-spec) | "in-progress" (fallback) | "in-progress" (fallback — `chore:` 는 transitions.go line 35에 의해 `("run-partial", "in-progress")` 로 명시 분류 → status 비어있지 않음) |
| 모든 N개 commit이 `chore(spec):` | "in-progress" (fallback) | error 반환 → lint rule이 skip |
| SPEC에 git history 0건 | error: no git history | error: no git history (변화 없음) |

`chore:` (non-spec, scope 없음) 는 transitions.go 의 fallthrough 규칙에 의해 `("run-partial", "in-progress")` 로 분류됨에 주의. 따라서 status가 비어있지 않으므로 walker가 즉시 반환한다 — 현행 동작과 동일.

---

## 3. Skip Pattern Definition

### 3.1 v2.20.0-rc1 적용 패턴

| Pattern (case-insensitive prefix) | Rationale | Source |
|----------------------------------|-----------|--------|
| `chore(spec):` | transitions.go line 22 에서 `(skip-meta, "")` 로 분류되는 metadata-only commit | PR #926, #930 sweep commit |
| `chore(specs):` | `chore(spec):` 의 plural 변형 — 일부 sweep commit 에서 사용 가능 | 방어적 포함 |

### 3.2 검토했으나 적용 보류한 패턴

| 패턴 | 보류 사유 |
|------|----------|
| `^sync(spec` | sync-phase는 표준적으로 `docs(sync):` prefix 사용 → 이미 transitions.go line 26-27에 의해 `("sync-merge", "completed")` 로 정확히 분류됨. `sync(spec)` 형태는 실무에서 거의 사용 안 됨 (plan.md §7 OQ3) |
| `^docs(sync):` | transitions.go line 27 에서 정확한 status로 분류됨 → fallback 트리거 없음 |
| `^revert:` | transitions.go line 51 에서 `(no-op, "")` 분류 → 이론상 fallback 트리거하나, revert 자체는 history 일부로 정상 의미 있음. 추후 SPEC에서 별도 검토. |
| `^refactor:`, `^perf:`, `^ci:`, `^test:` | transitions.go line 32-37 에서 명시 분류됨 (예: `refactor:` → `("run-partial", "in-progress")`, `perf:` → 동일) → 빈 status 반환 안 함 → walker가 fallback 경로를 거치지 않으므로 skip 불요. `^build:` 등 transitions.go에 미등록된 prefix는 §6.5 ack 참조. |

### 3.3 정규식 vs prefix 매칭

`shouldSkipCommitTitle` 구현은 **`strings.HasPrefix` + `strings.ToLower`** 를 선택한다. 정규식 (`regexp.MustCompile("^chore\\(spec.*\\):")`) 대비 장점:

- 더 빠름 (regexp compile 비용 없음)
- 코드 가독성 향상
- 패턴 수가 2개 (chore(spec):, chore(specs):) 로 적어 prefix 매칭이 적절

미래에 패턴 수가 5개 이상으로 늘어나면 정규식 도입 재검토.

---

## 4. Walk Depth Decision (N)

### 4.1 N = 50 권장

**Empirical evidence** (research.md §2 / git log 직접 측정):

- SPEC-UTIL-001: git log --grep 매칭 commit ~5건 (plan, run, sync, fix, chore sweep)
- SPEC-V3R2-CON-001: 매칭 commit ~15건 (multi-wave SPEC 으로 chore + impl + sync 누적)
- SPEC-V3R4-SPECLINT-DEBT-001 (선행 SPEC, 가장 commit-rich): ~25건
- 모든 SPEC 평균: ~8건
- 평균 + 3σ 추정: ~20건

N = 50 은 약 2-3x 안전 여유. 미래 5년간 SPEC 라이프사이클이 길어져도 견딘다.

### 4.2 Cost Analysis

`git log --grep --oneline -50` 의 비용:

- 본 저장소 (`/Users/goos/MoAI/moai-adk-go`, 약 6000 commit 보유)에서 측정 시 ~30ms
- 188개 SPEC × 30ms = 5.6초 ← lint 전체 실행 시간 (수 분) 대비 < 2% — 무시 가능

N = 100 으로 늘려도 비용은 동일 수준 (60ms × 188 = 11.3초 — 여전히 무시 가능). 단, justification 부재로 50 으로 유지.

### 4.3 Configurability

N 값은 `const gitLogWindowSize = 50` 로 Go 소스에 hard-code한다. plan.md §7 OQ2 externalization 결정과 일관 — v2.20.0-rc1 에서 외부 설정 부재.

향후 일부 super-deep-history SPEC이 false-positive를 일으키면 const 값 상향 + 별도 SPEC으로 처리.

---

## 5. Edge Cases

### 5.1 SPEC has only chore commits in history

**Scenario**: 어떤 SPEC이 plan/run 단계 없이 sweep 으로만 생성됨 (예: 메타데이터 정리 목적). 모든 git log 매칭 commit이 `chore(spec):` prefix.

**Behavior**: walker가 N개 모두 skip → error 반환 → `StatusGitConsistencyRule::Check` 가 skip 처리 → WARN/ERROR 발생 안 함.

**Verification**: AC-LSCSK-005 case (c) test 가 검증.

**판단**: 정상 동작. 만약 이 SPEC의 frontmatter status가 부정확하면 다른 메커니즘 (예: ParseStatus + 수동 검수) 으로 잡아내야 하며, status-git-consistency 검사는 본질적으로 skip이 정답.

### 5.2 SPEC has zero commits

**Scenario**: 어떤 SPEC이 디스크에 존재하지만 git log에 한 번도 commit으로 포함되지 않음 (예: untracked SPEC).

**Behavior**: `git log --grep=<specID> -50` 가 빈 output 반환 → drift.go line 113-115 의 기존 처리 경로 → "no git history found" error → 상위 rule이 skip.

**판단**: 현행 동작 유지. walker 로직과 호환.

### 5.3 Sweep commit body lists 11 SPECs

**Scenario**: PR #930 처럼 sweep commit 하나가 11개 SPEC을 언급. 11개 SPEC 모두에 대해 `git log --grep` 가 이 commit을 latest 로 매칭.

**Behavior**: 각 SPEC 검색마다 walker가 sweep commit을 skip → 다음 (각 SPEC의 실제 impl/feat/sync commit) 로 이동 → 올바른 status 반환.

**Verification**: AC-LSCSK-005 case (a) test 가 검증 (sweep commit이 real impl commit을 hide하는 시나리오).

### 5.4 Merge commits

**Scenario**: PR squash-merge가 아닌 일반 merge로 머지된 commit.

**Behavior**: `git log --no-merges` 플래그가 이미 merge commit을 제외 → walker 입력에서부터 배제. 변경 불필요.

**판단**: 현행 동작 유지 (drift.go line 107 의 `--no-merges` 플래그).

### 5.5 Mixed prefix in title (e.g., "chore(spec) + feat" combined)

**Scenario**: `chore(spec): metadata update, feat(SPEC-X-001): new feature` 같이 두 prefix가 혼합된 비표준 commit title.

**Behavior**: prefix matching은 `^chore(spec):` 만 검사 → 이 commit은 skip됨. 그 다음 commit으로 이동.

**판단**: 이런 commit은 실무에서 거의 없으며 (Conventional Commits 위반), 안전상 skip 처리가 합리적. 만약 이 commit이 실제 lifecycle 정보를 담고 있다면 별도 표준 commit으로 재발행 필요.

### 5.6 Case insensitivity

**Scenario**: `Chore(Spec):`, `CHORE(SPEC):` 같은 대소문자 변형.

**Behavior**: `shouldSkipCommitTitle` 가 `strings.ToLower` 적용 후 비교 → 모두 정상 skip.

**Verification**: 별도 test case 추가 권장 (M3 REFACTOR phase).

---

## 6. Failure Mode Analysis

### 6.1 Walker returns error (no classifiable commit within window)

**Trigger**: SPEC의 모든 N=50 매칭 commit이 skip pattern 또는 unknown prefix.

**Lint Rule Behavior**: `StatusGitConsistencyRule::Check` line 897-900:

```go
gitStatus, err := getGitImpliedStatus(fm.ID)
if err != nil {
    // git 이력 추론 불가 시 본 검사를 건너뜀
    return nil
}
```

→ error를 받으면 빈 findings slice 반환 → WARN/ERROR 미발생.

**판단**: 정상 fail-safe 동작. lint rule을 silently skip하는 것은 잘못된 false-positive 보다 안전.

### 6.2 git log subprocess failure

**Trigger**: `git` 명령이 존재하지 않거나 저장소가 손상.

**Behavior**: drift.go line 108-110 의 기존 처리 → `"git log failed"` error → 상위 rule skip.

**판단**: 현행 동작 유지.

### 6.3 Walker depth budget exhausted but real commit exists at N+1

**Trigger**: 매우 deep history SPEC에서 의미 있는 commit이 N=50 이후에만 존재 (이론적으로 발생 가능하나 실제로는 거의 불가능 — research.md §2 데이터 참조).

**Behavior**: walker가 N 소진 → error → lint rule skip → status mismatch가 잡히지 않음 (false-negative).

**Mitigation**:

- N=50의 안전 여유는 평균 + 3σ 의 약 2-3x → 실무에서 발생 가능성 < 0.1%
- 만약 발생하면 다른 메커니즘 (manual review, periodic audit) 으로 보완
- 발견 시 N 상향 또는 SPEC-specific override 도입을 별도 SPEC으로 처리

### 6.4 transitions.go 규칙 변경에 의한 회귀

**Trigger**: 누군가 `transitions.go` 의 `chore(spec):` 분류 규칙을 실수로 변경 (예: `(skip-meta, "implemented")` 로 변경).

**Behavior**: walker가 sweep commit을 skip하지 않게 됨 → 현행 결함 재현.

**Mitigation**: AC-LSCSK-003 regression test `TestClassifyPRTitle_ChoreSpecUnchanged` 가 `transitions_test.go` 에 영구 보존 → 회귀 즉시 검출.

### 6.5 skip pattern 누락 (새 chore prefix 도입)

**Trigger**: 누군가 새 sweep tool을 만들어 `style(spec):` prefix로 sweep commit 작성.

**Behavior**: walker가 `style(spec):` 를 skip하지 않음 → transitions.go line 37 의 `style:` 규칙과 매칭 시도. `style:` (no scope) → `("run-partial", "in-progress")` → status 비어있지 않음 → walker 즉시 반환 → "in-progress" 채택 → false-positive WARN 재발.

**Mitigation**:

- 본 SPEC scope는 `chore(spec):` 만 처리. 새 prefix 도입 시 별도 SPEC + skip pattern 확장 (plan.md §7 OQ2 externalization 결정 포함) 필요.
- 운영 절차로 "새 sweep tool 도입 시 본 SPEC 후속 조치 필요" 가이드 문서화 (별도 task).

### 6.6 transitions.go 미등록 prefix의 fall-through

`build:`, `style:` (scope 없음 변형) 처럼 transitions.go 의 transitionRules 에 등록되지 않은 prefix는 `ClassifyPRTitle` 이 unknown 분류 + 빈 status 를 반환하여 현재 walker의 안전망 (line `if status == ""` continue) 으로 자연히 skip된다. 그러나 walker 가 N=50 budget 내 모두 skip 처리할 경우 false-negative (실제 lifecycle 정보 손실) 위험이 있다. 운영 가이드: (a) 새 prefix 도입 시 `internal/spec/transitions.go::transitionRules` 에 명시 분류를 추가하거나, (b) 본 SPEC 의 `shouldSkipCommitTitle` skip-pattern 목록을 확장하여 walker가 의도적으로 다음 commit으로 이동하도록 한다 — 둘 중 하나를 사전 적용한 후에만 신규 prefix sweep tool을 도입할 것.

---

## 7. Backward Compatibility

본 SPEC의 변경은 다음 측면에서 backward-compatible:

| 측면 | 호환성 보장 방법 |
|------|----------------|
| `getGitImpliedStatus` signature | `(string, error)` 그대로 유지 |
| 호출자 (lint.go::StatusGitConsistencyRule::Check) | 코드 변경 불요 — err 처리 경로 그대로 |
| `ClassifyPRTitle` 의미 | 변경 없음 (AC-LSCSK-003) |
| SPEC frontmatter format | 변경 없음 (REQ-LSCSK-002) |
| `.moai/config/sections/spec-lint.yaml` | 변경 없음 (OQ2 deferred) |
| CLI flag (`moai spec lint`, `moai spec lint --strict`) | 변경 없음 |
| Exit code | 변경 없음 — 다만 같은 SPEC set에 대해 strict 모드 exit 0 으로 전환되는 효과 (의도된 fix) |

기존 호출자 또는 통합 도구가 본 fix로 인해 깨질 가능성: **없음**.

---

## 8. @MX Tag Placement (코드 주석 가이드)

`internal/spec/drift.go` 수정 시 다음 태그 권장 (mx-tag-protocol.md 준수):

| 위치 | 태그 종류 | 사유 |
|------|----------|------|
| `getGitImpliedStatus` 함수 head | `@MX:NOTE` | 함수 의도 명시 — walker filter 도입 배경 |
| `gitLogWindowSize` const 정의 | `@MX:NOTE` | N=50 결정 사유 + 변경 시 영향 |
| `shouldSkipCommitTitle` 함수 | `@MX:ANCHOR` (fan_in 검토 후) | skip pattern은 SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 invariant. fan_in이 3 미만이면 NOTE로 격하 |
| Walker loop 내 skip continue | `@MX:NOTE` | "chore(spec) commit 건너뛰기" 의도 |
| Walker exhaustion 의 error 반환 | `@MX:NOTE` + `@MX:REASON` | "N=50 소진 시 unknown 신호 — lint rule이 skip 처리" |

`@MX:REASON` 은 한국어 (code_comments=ko).

`@MX:SPEC: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` 모든 새 태그에 부착.
