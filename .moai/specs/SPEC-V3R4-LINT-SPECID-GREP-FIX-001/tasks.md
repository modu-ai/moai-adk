# SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — Task Breakdown

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. 3-Wave (RED/GREEN/REFACTOR) × 9 tasks. TDD discipline 엄격 준수. 각 task는 file + acceptance + estimated diff size + AC mapping. |

---

## 1. Task Summary

| Wave | Task ID | Title | Type | Files | AC mapping |
|------|---------|-------|------|-------|-----------|
| W1 (RED) | T1-001 | Create drift_specid_grep_test.go with failing tests | test | `internal/spec/drift_specid_grep_test.go` (new) | AC-LSGF-002, AC-LSGF-004 |
| W1 (RED) | T1-002 | Verify pre-fix RED state | test-verify | (run only) | AC-LSGF-002, AC-LSGF-004 |
| W1 (RED) | T1-003 | Identify/confirm chore-skip regression guards | test | `internal/spec/drift_test.go` (or transitions_test.go) | AC-LSGF-003 |
| W2 (GREEN) | T2-001 | Add commitMatchesSPECID helper | impl | `internal/spec/drift.go` | REQ-LSGF-001..004 |
| W2 (GREEN) | T2-002 | Wire helper into getGitImpliedStatus walker | impl | `internal/spec/drift.go` | REQ-LSGF-001..005 |
| W2 (GREEN) | T2-003 | Verify GREEN — Wave 1 tests pass | test-verify | (run only) | AC-LSGF-002..004 |
| W2 (GREEN) | T2-004 | End-to-end lint clean verification | verify | (run only) | AC-LSGF-001 |
| W3 (REFACTOR) | T3-001 | Update @MX:NOTE / @MX:REASON / @MX:ANCHOR tags | docs | `internal/spec/drift.go` | (governance) |
| W3 (REFACTOR) | T3-002 | Final lint sweep + race-safe full test | verify | (run only) | AC-LSGF-001, AC-LSGF-005 |

---

## 2. Wave 1 — RED

### T1-001: 신규 테스트 파일 `drift_specid_grep_test.go` 신설

**Goal**: 현재 walker가 false-positive를 생성하는 시나리오를 unit test로 capture. 두 테스트 모두 pre-fix에서 FAIL.

**File**: `internal/spec/drift_specid_grep_test.go` (new)

**Tests to add**:

```go
package spec

import (
    "testing"
)

// TestGetGitImpliedStatus_HARNESS001Resolution verifies that the walker
// returns the genuine sync signal for SPEC-V3R4-HARNESS-001, not the
// substring-noise NAMESPACE-001 plan commit.
//
// Pre-fix (substring matching) returns "planned" (from NAMESPACE plan commit).
// Post-fix (word-boundary matching) returns "completed" (from sync commit).
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
    status, err := getGitImpliedStatus("SPEC-V3R4-HARNESS-001")
    if err != nil {
        t.Fatalf("getGitImpliedStatus returned unexpected error: %v", err)
    }
    if status != "completed" {
        t.Errorf("expected status 'completed' (genuine sync signal), got %q (likely NAMESPACE substring noise)", status)
    }
}

// TestGetGitImpliedStatus_SPECIDWordBoundary verifies that commitMatchesSPECID
// (or equivalent walker filter) treats SPEC-ID boundaries with regex precision.
//
// This test exercises the post-filter logic directly via commitMatchesSPECID
// helper. The helper is introduced in Wave 2 GREEN — Wave 1 RED stages this
// test as expected-fail (helper not yet defined → compile error or behavior fail).
func TestGetGitImpliedStatus_SPECIDWordBoundary(t *testing.T) {
    tests := []struct {
        name        string
        commitTitle string
        specID      string
        want        bool
    }{
        {
            name:        "C1 exact match (plan)",
            commitTitle: "plan(spec): SPEC-V3R4-HARNESS-001 — initial",
            specID:      "SPEC-V3R4-HARNESS-001",
            want:        true,
        },
        {
            name:        "C2 substring noise (NAMESPACE)",
            commitTitle: "plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 — supersedes 001",
            specID:      "SPEC-V3R4-HARNESS-001",
            want:        false,
        },
        {
            name:        "C3 exact match (sync)",
            commitTitle: "sync(SPEC-V3R4-HARNESS-001): status transition",
            specID:      "SPEC-V3R4-HARNESS-001",
            want:        true,
        },
        {
            name:        "C4 chore-post token (no SPEC- prefix)",
            commitTitle: "chore(post-V3R4-HARNESS-001): cleanup",
            specID:      "SPEC-V3R4-HARNESS-001",
            want:        false,
        },
        {
            name:        "C5 closeout body (HARNESS-001 without SPEC- prefix)",
            commitTitle: "sync(specs): closeout (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed)",
            specID:      "SPEC-V3R4-HARNESS-001",
            want:        false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := commitMatchesSPECID(tt.commitTitle, tt.specID)
            if got != tt.want {
                t.Errorf("commitMatchesSPECID(%q, %q) = %v, want %v",
                    tt.commitTitle, tt.specID, got, tt.want)
            }
        })
    }
}
```

**Acceptance**:
- 파일 생성 완료, build/compile 시도 시 `commitMatchesSPECID` undefined (Wave 2에서 정의) → 의도된 RED.
- 또는 stub helper로 (`func commitMatchesSPECID(title, id string) bool { return strings.Contains(title, id) }`) 정의해 두면 C2/C4/C5 sub-test FAIL → RED.

**Estimated diff**: +60-80 LOC

---

### T1-002: pre-fix RED 상태 검증

**Goal**: T1-001의 테스트가 실제로 실패함을 main HEAD 상태에서 확인.

**Commands**:
```bash
# Stub helper 정의 후 (또는 helper 정의 없이 compile-only)
go test ./internal/spec/... -run 'TestGetGitImpliedStatus_HARNESS001Resolution|TestGetGitImpliedStatus_SPECIDWordBoundary' -v 2>&1 | tail -30
```

**Acceptance**:
- `TestGetGitImpliedStatus_HARNESS001Resolution` FAIL: `expected status 'completed' ..., got "planned"`.
- `TestGetGitImpliedStatus_SPECIDWordBoundary` C2/C4/C5 FAIL (stub helper 사용 시).

**Estimated diff**: 0 (verify only)

---

### T1-003: chore-skip regression guard 확보

**Goal**: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 회귀 가드 테스트 식별. 없다면 anchor test 추가.

**Files to inspect**:
- `internal/spec/drift_test.go`
- `internal/spec/transitions_test.go`

**Tasks**:
1. `TestShouldSkipCommitTitle*` 또는 `TestClassifyPRTitle_ChoreSpecUnchanged` 존재 여부 grep.
2. 없으면 `TestShouldSkipCommitTitle_ChoreSpecUnchanged` 신설:
   ```go
   func TestShouldSkipCommitTitle_ChoreSpecUnchanged(t *testing.T) {
       cases := []struct {
           title string
           want  bool
       }{
           {"chore(spec): cleanup", true},
           {"chore(specs): bulk sweep", true},
           {"feat(spec): SPEC-X-001", false},
           {"plan(spec): SPEC-X-001", false},
       }
       for _, c := range cases {
           if got := shouldSkipCommitTitle(c.title); got != c.want {
               t.Errorf("shouldSkipCommitTitle(%q) = %v, want %v", c.title, got, c.want)
           }
       }
   }
   ```

**Acceptance**: chore-skip 회귀 가드 테스트가 존재 + GREEN (pre-fix).

**Estimated diff**: 0-20 LOC (anchor test if missing)

---

## 3. Wave 2 — GREEN

### T2-001: `commitMatchesSPECID` helper 신설

**Goal**: word-boundary SPEC-ID 매칭 helper 함수를 drift.go에 추가.

**File**: `internal/spec/drift.go`

**Position**: `shouldSkipCommitTitle` 함수 직후 (line ~188 이후).

**Code**:
```go
// commitMatchesSPECID는 commit title에 정확한 SPEC-ID 토큰이 포함되는지 확인한다.
//
// git log --grep=<specID>는 substring 매칭을 수행하므로,
// 예를 들어 specID="SPEC-V3R4-HARNESS-001" 검색이
// "plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 ..." commit에도 매칭되는 결함이 있다.
//
// 본 함수는 ExtractSPECIDs (transitions.go:97-116)를 사용해 정확한 SPEC-ID 토큰만 추출한 후,
// target specID가 그 set에 포함되어 있는지 확인한다.
//
// @MX:NOTE: [AUTO] commitMatchesSPECID — word-boundary SPEC-ID 필터 (LSGF-001)
// @MX:REASON: SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — git log --grep substring 매칭이
//   NAMESPACE supersede commit을 walker first match로 채택하던 결함을 차단.
//   ExtractSPECIDs 재사용으로 외부 의존성 0.
func commitMatchesSPECID(commitTitle, specID string) bool {
    extracted := ExtractSPECIDs(commitTitle)
    return slices.Contains(extracted, specID)
}
```

**Import**: `slices` (Go 1.21+ stdlib; Go 1.23 프로젝트이므로 안전).

**Acceptance**: T1-001 `TestGetGitImpliedStatus_SPECIDWordBoundary` 5 sub-cases 모두 GREEN.

**Estimated diff**: +18 LOC (function + 2 comments) + 1 import

---

### T2-002: `getGitImpliedStatus` walker에 필터 wire

**Goal**: scanner loop에서 `shouldSkipCommitTitle` 직후 `commitMatchesSPECID` 호출.

**File**: `internal/spec/drift.go:131-164` (scanner loop)

**Diff**:
```go
for scanner.Scan() {
    line := scanner.Text()
    parts := strings.SplitN(line, " ", 2)
    if len(parts) < 2 {
        continue
    }
    commitTitle := parts[1]

    if shouldSkipCommitTitle(commitTitle) {
        continue
    }

+   // word-boundary SPEC-ID 필터 (LSGF-001) — substring collision 차단
+   // 예: specID="SPEC-V3R4-HARNESS-001"이 "SPEC-V3R4-HARNESS-NAMESPACE-001"에 매칭되지 않도록
+   // @MX:NOTE: [AUTO] LSGF-001 word-boundary filter
+   // @MX:REASON: NAMESPACE supersede commit이 walker first match로 채택되던 false-positive 차단
+   if !commitMatchesSPECID(commitTitle, specID) {
+       continue
+   }
+
    _, status, err := ClassifyPRTitle(commitTitle)
    // ... rest unchanged
}
```

**Acceptance**: T1-001 `TestGetGitImpliedStatus_HARNESS001Resolution` GREEN.

**Estimated diff**: +6 LOC

---

### T2-003: Wave 1 tests GREEN 검증

**Commands**:
```bash
go test ./internal/spec/... -run 'TestGetGitImpliedStatus_HARNESS001Resolution|TestGetGitImpliedStatus_SPECIDWordBoundary|TestShouldSkipCommitTitle' -v
```

**Acceptance**: 모든 named tests PASS.

**Estimated diff**: 0

---

### T2-004: End-to-end lint clean (AC-LSGF-001)

**Commands**:
```bash
moai spec lint --strict 2>&1 | tail -5
echo "exit code: $?"
```

**Acceptance**:
- 마지막 줄 `0 error(s), 0 warning(s)`.
- exit code 0.

**Estimated diff**: 0

---

## 4. Wave 3 — REFACTOR

### T3-001: @MX 태그 갱신

**Goal**: `getGitImpliedStatus` 헤더 godoc + `gitLogWindowSize` MX 태그에 LSGF-001 cross-reference 추가.

**File**: `internal/spec/drift.go`

**Diffs**:

1. `getGitImpliedStatus` 헤더 (line 101-112):
   ```go
   // getGitImpliedStatus는 SPEC-ID에 대한 git log를 분석하여 lifecycle status를 추론한다.
   //
   // walker는 두 필터를 순차 적용한다:
   //   1. chore-skip 필터 (LSCSK-001): chore(spec): sweep commit 제외
   //   2. word-boundary 필터 (LSGF-001): substring collision (예: HARNESS-001 vs HARNESS-NAMESPACE-001) 차단
   //
   // 모든 N개 commit이 skip되면 error를 반환하고, 상위 lint rule이 skip 조건으로 처리한다.
   //
   // @MX:ANCHOR: [AUTO] getGitImpliedStatus — git-implied status 추론 진입점
   // @MX:REASON: StatusGitConsistencyRule.Check + DetectDrift 두 곳에서 호출 (fan_in=2);
   //   LSCSK-001 (chore-skip) + LSGF-001 (word-boundary) 두 결함 fix가 적용된 core walker.
   ```

2. `gitLogWindowSize` (line 96-99):
   ```go
   // gitLogWindowSize는 getGitImpliedStatus 가 git log에서 최대 몇 개의 commit을 조회할지 결정한다.
   //
   // @MX:NOTE: [AUTO] N=50 결정 근거: SPEC당 평균 git log 매칭 commit 5-10건 +
   //   word-boundary 필터 (LSGF-001) 후 expected match 1-3건 → 5-10x 안전 여유.
   // @MX:REASON: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 OQ1 (원본 결정) +
   //   SPEC-V3R4-LINT-SPECID-GREP-FIX-001 (word-boundary 필터 영향 평가 — 변동 없음).
   ```

**Acceptance**: godoc + MX tag에 LSGF-001 명시 + LSCSK-001 cross-reference 유지.

**Estimated diff**: ~12 LOC (comment-only)

---

### T3-002: 최종 검증

**Commands**:
```bash
go vet ./internal/spec/...
golangci-lint run ./internal/spec/...
go test ./internal/spec/... -race -count=1 -v 2>&1 | tail -20
moai spec lint --strict 2>&1 | tail -3
```

**Acceptance**:
- 모든 명령 exit 0.
- `moai spec lint --strict` `0 error(s), 0 warning(s)`.
- AC-LSGF-001~005 모두 PASS.

**Estimated diff**: 0

---

## 5. Task Dependencies (DAG)

```
T1-001 (RED tests)
  ↓
T1-002 (verify RED)
  ↓
T1-003 (chore-skip guard) — parallel-safe
  ↓
T2-001 (helper)
  ↓
T2-002 (wire walker)
  ↓
T2-003 (verify GREEN)
  ↓
T2-004 (E2E lint)
  ↓
T3-001 (MX tags)
  ↓
T3-002 (final sweep)
```

T1-003은 T1-001/T1-002와 file이 다르므로 parallel-safe. 나머지는 sequential.

---

## 6. Estimated Total Diff

| Wave | Files | Net LOC | Comment |
|------|-------|---------|---------|
| W1 | `drift_specid_grep_test.go` (new), `drift_test.go` (optional) | +60-100 | test only |
| W2 | `drift.go` | +24 (18 helper + 6 wire) | implementation |
| W3 | `drift.go` | +12 (comment-only) | docs |
| **Total** | 2-3 files | **+96-136 LOC** | small surgical change |

---

## 7. Out of Scope

다음은 본 task list에서 의도적으로 배제된다 (spec.md §4 Exclusions 참조):

- `transitions.go` 수정
- `lint.skip` 추가
- HARNESS-001/002/003 spec.md frontmatter touch
- 새 lint rule
- body 매칭 확장
- grace window
- 새 외부 의존성
- 197 SPECs mass-update
