# Plan — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 plan. 5-Wave 전략 (BASELINE / Pattern A bulk / Pattern B+C verification / Pattern D-G detector exemption / Pattern H 재귀 cleanup). LSKC-001 Go bulk script fork → `.moai/scripts/status-drift-cleanup.go`. plan-in-main + worktree 미사용. BODP A=¬ B=¬ C=¬ → main @ origin/main. 4 OQ + 6 risks. |

---

## 1. Implementation Strategy

### 1.1 핵심 접근

본 cleanup은 **frontmatter status field 동기화 + narrow detector exemption** 의 결합이다:

- **63 SPECs frontmatter 변경** (Pattern A: 50 + Pattern B+C: 10 + Pattern H: 3): bulk Go script 로 `status` / `version` / `updated_at` / HISTORY row 수정
- **1 detector code 변경** (Pattern D/E/F/G 4 SPECs 처리): `internal/spec/lint.go::StatusGitConsistencyRule::Check` 에 terminal state exemption 추가 + 4 unit tests
- **5 wave 분해** (병렬 금지, 엄격 순서): BASELINE → Pattern A bulk → Pattern B+C verification → Pattern D-G detector → Pattern H sync-phase 재귀

### 1.2 Wave 분해 전략 (priority-based, time-estimate 금지)

| Wave | Priority | 목적 | Volume |
|------|----------|------|--------|
| BASELINE | High (모든 후속 wave 의 기준선) | actual affected-list 산출 | 64 lines (총합) |
| Wave 1 (Pattern A) | High (volume) | bulk script로 50건 일괄 downgrade | 50 SPECs |
| Wave 2 (Pattern B+C) | High (correctness) | 개별 verification + 선택적 downgrade | 10 SPECs |
| Wave 3 (Pattern D-G detector) | Medium (코드 변경) | terminal state exemption + 4 tests | 1 Go file + 1 test file |
| Wave 4 (Pattern G 검증) | Low (Wave 3 자동 효과) | SPEC-V3R3-WEB-001 통과 검증만 | 0 변경 |
| Wave 5 (Pattern H, sync-phase) | Low (마지막 wave) | cleanup-chain 4 SPEC 재귀 정합화 | 4 SPECs (본 SPEC 자체 포함) |

### 1.3 Sequential vs Parallel

[HARD] **Sequential only**. 이유 (LSKC-001 §1.3 정신 계승):

- 각 Wave 완료 후 `moai spec lint --strict` 실행하여 누적 효과 측정
- 병렬 처리 시 lint baseline 검증 시점이 partial state → 잘못된 분류
- single PR 목표 (run-phase) → commit ordering 일관성

---

## 2. Milestones

priority-based ordering (per `agent-common-protocol.md` §Time Estimation):

### M1 — BASELINE

**Priority: High** (모든 후속 단계 기준선)

작업:
1. main HEAD 확인 (`git log --oneline -1` → `758341089` 또는 그 이후)
2. `moai spec lint --strict 2>&1 > /tmp/lint-baseline.txt` → 전체 출력 캡처
3. `grep "StatusGitConsistency" /tmp/lint-baseline.txt | wc -l` → 64 (또는 ±변동) 측정
4. SPEC ID 별 분류:
   - `grep "StatusGitConsistency" /tmp/lint-baseline.txt | awk '{print $2}' | sort -u` → unique SPEC ID list
   - 각 SPEC ID 의 frontmatter status + git-implied 매핑 → 8 패턴 (A-H) 분류
   - 4 affected-list 파일 작성: `affected-list-pattern-{A,B,C,H}.txt`
   - Pattern D/E/F/G 의 4 SPEC ID는 별도 list (코드 변경 검증용)

**완료 조건:**
- `affected-list-pattern-A.txt` (50 ± 변동 lines)
- `affected-list-pattern-B.txt` (4 ± lines)
- `affected-list-pattern-C.txt` (6 ± lines)
- `affected-list-pattern-H.txt` (3 lines, 사전 확정 — cleanup chain 본 SPEC 제외)
- `pattern-DEFG.txt` 4 SPEC ID list (검증 reference)
- baseline.txt 캡처 (lint --strict 전체 출력)

### M2 — Wave 1: Pattern A Bulk Downgrade

**Priority: High** (볼륨)

작업:
1. `.moai/scripts/status-drift-cleanup.go` 작성 (LSKC-001 fork, design.md §3.2)
2. dry-run: `go run .moai/scripts/status-drift-cleanup.go --pattern A --dry-run`
   - 50 SPECs 의 expected diff 출력
   - frontmatter parse 오류 사전 검출
3. 실제 실행: `go run .moai/scripts/status-drift-cleanup.go --pattern A`
4. 검증:
   - `git status .moai/specs/` → 50 modified
   - `awk '/^status:/' .moai/specs/<pattern A SPEC>/spec.md` → 모두 `implemented`
   - `moai spec lint --strict | grep -E "$(cat affected-list-pattern-A.txt | sed 's|.moai/specs/||;s|/spec.md||' | tr '\n' '|')" | wc -l` → 0
   - `git diff --stat .moai/specs/` → 50 files modified, frontmatter 변경 + HISTORY 1줄

**완료 조건:**
- 50 Pattern A SPECs frontmatter status `completed → implemented`
- version patch bump + updated_at 갱신 + HISTORY row
- lint --strict Pattern A 잔여 WARN 0건

### M3 — Wave 2: Pattern B+C Verification & Conditional Downgrade

**Priority: High** (correctness)

작업:
1. `run-verification.md` 초안 작성 (10 row, decision pending)
2. 각 SPEC 개별 verification (10건 sequential):
   - `git log --oneline --no-merges --grep=<SPEC-ID> -50` 출력 검토
   - project memory grep (`grep -lR "<SPEC-ID>" ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)
   - SPEC body review (`.moai/specs/<SPEC-ID>/spec.md`)
   - decision: `downgrade` (default) 또는 `keep + sync-pending` (예외)
3. `affected-list-pattern-B.txt` + `affected-list-pattern-C.txt` 에서 분기 (a) downgrade SPECs 만 추출
4. bulk script Operation 추가 (Pattern B: completed → in-progress, Pattern C: implemented → in-progress)
5. `go run .moai/scripts/status-drift-cleanup.go --pattern B --pattern C --dry-run` → 검증
6. 실제 실행: `go run .moai/scripts/status-drift-cleanup.go --pattern B --pattern C`
7. 검증:
   - `git status .moai/specs/` → Wave 2 영향 SPECs modified
   - `moai spec lint --strict | grep -E "$(cat affected-list-pattern-B.txt affected-list-pattern-C.txt | sed 's|.moai/specs/||;s|/spec.md||' | tr '\n' '|')" | wc -l` → 0 (분기 a SPECs)
   - 분기 (b) keep SPECs 는 run-verification.md 에 사유 + 기록
8. 분기 (b) 가 3건 이상이면 사용자 확인 (orchestrator AskUserQuestion)

**완료 조건:**
- Pattern B+C 분기 (a) downgrade 적용된 SPECs frontmatter 정합화
- run-verification.md 10 row 모두 decision 기록
- lint --strict Pattern B+C 잔여 WARN: 분기 (b) keep SPEC 수 만큼 (정상)

### M4 — Wave 3: Pattern D/E/F/G Detector Exemption

**Priority: Medium** (코드 변경)

작업:
1. `internal/spec/lint.go` 또는 `internal/spec/lint/checks/status_git_consistency.go` 정확한 위치 확인 (`grep -rn "StatusGitConsistency" internal/spec/`)
2. design.md §2.2 코드 sketch 적용:
   - `terminalStatusEnum` map 정의 (superseded, archived, rejected)
   - `StatusGitConsistencyRule::Check` 함수에 terminal state exemption 분기 추가
3. @MX 태그 부착 (mx-tag-protocol.md 준수, 한국어 description):
   - `terminalStatusEnum` 정의에 `@MX:NOTE`
   - `Check` 함수에 `@MX:ANCHOR` (fan_in 검토 후)
   - exemption 분기 line에 `@MX:NOTE`
4. 새 unit test 파일 (또는 기존 lint_test.go 확장):
   - `TestStatusGitConsistency_TerminalState_Superseded_Completed` (Pattern D)
   - `TestStatusGitConsistency_TerminalState_Superseded_Implemented` (Pattern E)
   - `TestStatusGitConsistency_TerminalState_Archived_Implemented` (Pattern F)
   - `TestStatusGitConsistency_TerminalState_Archived_InProgress` (Pattern G)
   - 각 test는 `t.TempDir()` + git fixture (CLAUDE.local.md §6 Test Isolation)
5. 검증:
   - `go test ./internal/spec/...` 전체 PASS
   - `go test -race ./internal/spec/` PASS
   - `golangci-lint run ./internal/spec/...` 0 warning
   - 4 SPEC (LSP-001, V3R3-HARNESS-001, I18N-001-ARCHIVED, V3R3-WEB-001) frontmatter status 변경 0
   - `moai spec lint --strict | grep -E "(SPEC-LSP-001|SPEC-V3R3-HARNESS-001|SPEC-I18N-001-ARCHIVED|SPEC-V3R3-WEB-001)" | wc -l` → 0

**완료 조건:**
- terminal state exemption 코드 적용
- 4 새 unit tests PASS
- 기존 테스트 회귀 0건 (또는 expected 갱신 적용)
- 4 terminal state SPECs lint --strict 통과

### M5 — Wave 4: Pattern G 통합 검증

**Priority: Low** (Wave 3 자동 효과)

작업:
- Wave 3 머지 후 (또는 같은 PR 내) Pattern G (SPEC-V3R3-WEB-001) 통과 확인
- `moai spec lint --strict | grep "SPEC-V3R3-WEB-001"` → 0 hit

**완료 조건:**
- Wave 3 의 sub-condition으로 자동 충족 (별도 작업 불요)

### M6 — Wave 5: Pattern H Recursive Cleanup (sync-phase)

**Priority: Low** (마지막 wave, sync-phase 책임)

작업:
- 본 SPEC sync-phase 시점에 bulk script 재실행 with `affected-list-pattern-H.txt`
- 4 cleanup-chain SPECs (LSKC-001, LINT-STATUS-CHORE-SKIP-001, SPECLINT-DEBT-001, **본 SPEC 자체**) frontmatter 정합화
- 본 SPEC 자체 포함 시 sync PR title 명시 ("Wave 5 자기참조 cleanup 포함")

**완료 조건:**
- 4 cleanup-chain SPECs frontmatter status 정합화
- `moai spec lint --strict | grep -E "(SPEC-V3R4-LINT-SKIP-CLEANUP-001|SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001|SPEC-V3R4-SPECLINT-DEBT-001|SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001)" | wc -l` → 0

### M7 — Final Verification & RUN-PR

**Priority: High** (전체 cleanup 종결)

작업:
1. 모든 Wave 적용 후 `moai spec lint --strict 2>&1 > /tmp/lint-final.txt`
2. `grep -c "StatusGitConsistency" /tmp/lint-final.txt` → 0 (AC-SDF-001)
3. `git diff main..HEAD --stat` 검증:
   - 63 SPECs `spec.md` modified (Pattern A + B+C 분기 a + H = 50 + N + 4 ≈ 60-64 — 분기 b keep 수에 따라 변동)
   - 1 detector code file modified (`internal/spec/lint.go` 또는 `lint/checks/status_git_consistency.go`)
   - 1 detector test file modified (`*_test.go`)
   - 1 bulk script (`.moai/scripts/status-drift-cleanup.go`) — Optional
   - 본 SPEC 자체 5 artifacts + progress.md + run-verification.md + 4 affected-list-*.txt
   - 외 다른 파일 0건
4. CI 검증:
   - `go test ./...` PASS
   - `go vet ./...` 0 issue
   - `golangci-lint run` 0 warning
   - `moai spec lint --strict` 0 ERROR + 0 StatusGitConsistency WARN
5. PR 생성:
   - branch: `feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001`
   - title: `feat(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — 64 SPEC status drift 일괄 동기화`
   - body: SPEC link + AC summary + lint --strict before/after 캡처 + Wave별 변경 요약

**완료 조건:**
- AC-SDF-001 ~ AC-SDF-008 모두 GREEN
- PR MERGED to main (Enhanced GitHub Flow §18.3 squash merge)
- main HEAD `moai spec lint --strict | grep -c "StatusGitConsistency"` → 0

---

## 3. Files to Modify

### 3.1 변경 대상 (frontmatter only)

63 SPECs `spec.md` 파일 (Pattern A: 50 + Pattern B+C 분기 (a): N + Pattern H: 3 = 60-64). 각 파일 평균 변경 4-5 lines:

- 수정: status (1줄) + version (1줄) + updated_at (1줄) = 3 lines
- 추가: HISTORY row (1줄) = 1 line
- Net diff per SPEC: ~4 lines

총 변경 추정: ~250 lines across 63 SPECs.

### 3.2 변경 대상 (Go code)

| Path | 변경 | 사유 |
|------|------|------|
| `internal/spec/lint.go` 또는 `internal/spec/lint/checks/status_git_consistency.go` | terminalStatusEnum 추가 + Check 함수 exemption 분기 추가 (~10 lines) | Pattern D/E/F/G 처리 |
| `internal/spec/lint_test.go` 또는 동등 test 파일 | 4 새 test cases (각 ~30 lines, 총 ~120 lines) | AC-SDF-004 검증 |

### 3.3 변경 없음 보장 (out-of-scope safety net)

- `internal/spec/drift.go` — **0 lines** (walker filter scope 미확장 per REQ-SDF-011)
- `internal/spec/transitions.go` — **0 lines** (ClassifyPRTitle prefix 매핑 보존)
- `internal/spec/status.go` — **0 lines** (lifecycle status enum 보존)
- `.github/workflows/*.yml` — **0 lines**
- `docs-site/**` — **0 files**
- `README.md` / `README.ko.md` / `CHANGELOG.md` — **0 lines** (CHANGELOG는 sync-phase 범위)
- 64-affected 외 `.moai/specs/*/spec.md` — **0 lines**
- 다른 `internal/spec/lint/checks/*.go` — **0 lines** (다른 lint rule 미변경)

### 3.4 신규 파일

| Path | 목적 |
|------|------|
| `.moai/scripts/status-drift-cleanup.go` | LSKC-001 fork, frontmatter status 정합화 bulk script (~250 LOC) |
| `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-A.txt` | Pattern A 50 SPECs |
| `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-B.txt` | Pattern B 4 SPECs |
| `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-C.txt` | Pattern C 6 SPECs |
| `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-H.txt` | Pattern H 3-4 SPECs (본 SPEC 포함 결정) |
| `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/run-verification.md` | Pattern B+C 개별 verification 기록 |
| `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/progress.md` | run-phase 진행 추적 |

---

## 4. Technical Approach

### 4.1 Bulk Script 설계 (LSKC-001 fork)

`.moai/scripts/status-drift-cleanup.go` 구조 (design.md §3.2 의사 코드 구체화):

```go
// pseudocode — run-phase 가 정확한 IO/import 결정
package main

import (
    "flag"
    "fmt"
    "os"
    "strings"
    
    "gopkg.in/yaml.v3"
)

type Operation struct {
    SpecPath     string  // .moai/specs/SPEC-XXX/spec.md
    SpecID       string  // SPEC-XXX
    Pattern      string  // "A" | "B" | "C" | "H"
    FromStatus   string  // 현재 frontmatter status
    ToStatus     string  // 목표 frontmatter status
    HistoryDescr string  // HISTORY row description
}

var (
    flagPattern = flag.String("pattern", "", "Pattern code (A, B, C, H) — comma-separated for multi")
    flagDryRun  = flag.Bool("dry-run", false, "Show diff without applying")
)

func main() {
    flag.Parse()
    
    patterns := strings.Split(*flagPattern, ",")
    var ops []Operation
    for _, p := range patterns {
        ops = append(ops, loadOperationsForPattern(p)...)
    }
    
    appliedCount := 0
    skippedCount := 0
    for _, op := range ops {
        applied, err := applyOperation(op, *flagDryRun)
        if err != nil {
            fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", op.SpecID, err)
            os.Exit(1)
        }
        if applied {
            appliedCount++
        } else {
            skippedCount++
        }
    }
    
    fmt.Printf("OK — applied=%d skipped=%d (idempotent)\n", appliedCount, skippedCount)
}

func loadOperationsForPattern(pattern string) []Operation {
    listPath := fmt.Sprintf(".moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-%s.txt", pattern)
    fromStatus, toStatus, descr := patternConfig(pattern)
    
    lines, _ := os.ReadFile(listPath)
    var ops []Operation
    for _, line := range strings.Split(string(lines), "\n") {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        specID := extractSpecID(line)  // .moai/specs/SPEC-XXX/spec.md → SPEC-XXX
        ops = append(ops, Operation{
            SpecPath:     line,
            SpecID:       specID,
            Pattern:      pattern,
            FromStatus:   fromStatus,
            ToStatus:     toStatus,
            HistoryDescr: descr,
        })
    }
    return ops
}

func patternConfig(pattern string) (from, to, descr string) {
    switch pattern {
    case "A":
        return "completed", "implemented",
            "status drift FOLLOWUP cleanup — Pattern A: completed → implemented (sync chore commit이 walker skip 대상이라 git-implied implemented). SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001."
    case "B":
        return "completed", "in-progress",
            "status drift FOLLOWUP cleanup — Pattern B: completed → in-progress (feat commit 부재로 git-implied in-progress; verification 분기 a). SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001."
    case "C":
        return "implemented", "in-progress",
            "status drift FOLLOWUP cleanup — Pattern C: implemented → in-progress (feat commit 부재로 git-implied in-progress; verification 분기 a). SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001."
    case "H":
        return "", "",  // H는 SPEC별로 다름 — 별도 dispatch 필요
            ""
    }
    return "", "", ""
}

func applyOperation(op Operation, dryRun bool) (bool, error) {
    data, err := os.ReadFile(op.SpecPath)
    if err != nil {
        return false, err
    }
    
    fmNode, body, err := parseFrontmatter(data)
    if err != nil {
        return false, err
    }
    
    // Idempotency: 현재 status가 이미 목표 status면 no-op
    currentStatus := getMapValue(fmNode, "status")
    if currentStatus == op.ToStatus {
        return false, nil  // skipped
    }
    if currentStatus != op.FromStatus {
        return false, fmt.Errorf("expected fromStatus=%q, got %q", op.FromStatus, currentStatus)
    }
    
    // 1. status 필드 변경
    setMapValue(fmNode, "status", op.ToStatus)
    
    // 2. version patch bump
    oldVersion := getMapValue(fmNode, "version")
    newVersion := bumpVersionPatch(oldVersion)
    setMapValue(fmNode, "version", newVersion)
    
    // 3. updated_at 갱신
    setMapValue(fmNode, "updated_at", "2026-05-16")
    
    // 4. HISTORY row 추가
    row := fmt.Sprintf("| %s | 2026-05-16 | manager-develop (run-phase) | %s |",
        newVersion, op.HistoryDescr)
    body = appendHistoryRow(body, row)
    
    if dryRun {
        fmt.Printf("DRY-RUN: %s — %s → %s, version %s → %s\n",
            op.SpecID, op.FromStatus, op.ToStatus, oldVersion, newVersion)
        return true, nil
    }
    
    // 5. 파일 재작성
    return true, os.WriteFile(op.SpecPath, serializeFrontmatter(fmNode, body), 0644)
}

// LSKC-001 fork: parseFrontmatter, serializeFrontmatter, getMapValue, setMapValue,
// bumpVersionPatch, appendHistoryRow — 동일 구현 재사용
```

### 4.2 Detector Exemption 코드 (Wave 3)

`internal/spec/lint.go` 또는 `internal/spec/lint/checks/status_git_consistency.go` 의 변경 (design.md §2.2 적용):

```go
// 신규 추가: package-level constant
// @MX:NOTE: [AUTO] terminal lifecycle state — git history와 mismatch가 정상으로 간주되는 status
// @MX:REASON: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Pattern D/E/F/G false-positive 해소
var terminalStatusEnum = map[string]bool{
    "superseded": true,
    "archived":   true,
    "rejected":   true,  // future-proof
}

// 기존 함수에 분기 추가:
func (r *StatusGitConsistencyRule) Check(ctx Context) []Finding {
    // ... 기존 코드 (frontmatter parse, lint.skip 검사) ...
    
    // ★ 신규 exemption 분기
    // @MX:NOTE: [AUTO] terminal state는 git-implied보다 ahead가 정상 — finding 미발생
    if terminalStatusEnum[fm.Status] {
        return nil
    }
    
    // ... 기존 코드 (getGitImpliedStatus 호출, mismatch 검증) ...
}
```

### 4.3 Frontmatter Parse 전략 (LSKC-001 패턴 계승)

`gopkg.in/yaml.v3` 의 `yaml.Node` API 사용 필수:

- `yaml.Unmarshal → map → Marshal` 사용 시 key ordering 손실 → **금지**
- `yaml.Node` 로 frontmatter parse → key 순서 보존된 `node.Content[0].Content[]` 순회 → 특정 key 의 value node 직접 수정 → `yaml.Marshal(node)` 로 serialize

LSKC-001 PR #937 에서 검증된 패턴 그대로 fork.

---

## 5. Testing Strategy

본 SPEC 은 frontmatter cleanup + narrow detector exemption이므로 테스트는 **AC-SDF-004 의 4 cases (terminal state exemption)** + bulk script idempotency 검증으로 한정 (테스트 부담 최소화).

### 5.1 Pre-Wave-1 BASELINE Test

```bash
# Body content sha256 capture (script-side)
go run .moai/scripts/status-drift-cleanup.go --pattern A --capture-baseline

# lint --strict 전체 출력 캡처
moai spec lint --strict 2>&1 > .moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/lint-baseline.txt
grep -c "StatusGitConsistency" .moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/lint-baseline.txt
# expected: 64 (또는 ±변동)
```

### 5.2 Wave 1 Post-Edit Verification

```bash
# Body sha256 unchanged
go run .moai/scripts/status-drift-cleanup.go --pattern A --verify-baseline

# lint --strict — Pattern A SPECs 잔여 WARN 0
moai spec lint --strict 2>&1 | \
  grep "StatusGitConsistency" | \
  grep -E "$(cat affected-list-pattern-A.txt | sed 's|.moai/specs/||;s|/spec.md||' | tr '\n' '|')" | \
  wc -l
# expected: 0
```

### 5.3 Wave 3 Detector Exemption Tests

새 unit tests (`internal/spec/lint_test.go` 또는 신규 `status_git_consistency_test.go` 확장):

```go
// pseudocode — run-phase 가 정확한 fixture / helper 사용
func TestStatusGitConsistency_TerminalState_Superseded_Completed(t *testing.T) {
    tmpDir := t.TempDir()
    setupGitFixture(tmpDir, []string{
        "feat(SPEC-X-001): implementation",
        "docs(sync): SPEC-X-001 sync close",
    })
    setupSpecFile(tmpDir, "SPEC-X-001", "superseded", "1.0.0")
    
    rule := &StatusGitConsistencyRule{}
    findings := rule.Check(testContext(tmpDir, "SPEC-X-001"))
    
    if len(findings) != 0 {
        t.Errorf("expected 0 findings (terminal state exemption), got %d", len(findings))
    }
}

func TestStatusGitConsistency_TerminalState_Superseded_Implemented(t *testing.T) { ... }
func TestStatusGitConsistency_TerminalState_Archived_Implemented(t *testing.T) { ... }
func TestStatusGitConsistency_TerminalState_Archived_InProgress(t *testing.T) { ... }
```

### 5.4 Idempotency Test (script-only)

```bash
# 1차 실행
go run .moai/scripts/status-drift-cleanup.go --pattern A
git diff --stat | tee /tmp/diff-1st.txt

# 2차 실행 (no-op 기대)
go run .moai/scripts/status-drift-cleanup.go --pattern A
git diff --stat | tee /tmp/diff-2nd.txt

# diff 비교
diff /tmp/diff-1st.txt /tmp/diff-2nd.txt && echo "PASS: idempotent" || echo "FAIL"
```

### 5.5 Final Verification (M7)

```bash
# 모든 Wave 적용 후
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"
# expected: 0  (AC-SDF-001)

go test ./internal/spec/... -v
# expected: PASS (4 새 cases + 기존 회귀 0)

go vet ./...
# expected: clean

golangci-lint run
# expected: 0 warning
```

### 5.6 Regression Guard

- LSKC-001 의 `internal/spec/drift.go` walker filter 테스트 (`TestGetGitImpliedStatus_SkipSweepCommit*`) 영향 없음 (walker filter 자체 미변경)
- LINT-STATUS-CHORE-SKIP-001 의 `transitions_test.go` regression test 영향 없음 (transitions.go 미변경)
- 본 SPEC은 lint.go (또는 lint/checks/) 만 변경하므로 위 두 SPEC 의 테스트 자산 보존

---

## 6. Definition of Done

- [ ] **AC-SDF-001 ~ AC-SDF-008 모두 GREEN** (M7 Final Verification 통과)
- [ ] **PR MERGED** to main (Enhanced GitHub Flow §18.3 squash merge)
- [ ] **`moai spec lint --strict | grep -c "StatusGitConsistency"`** → 0
- [ ] **63 SPECs frontmatter** (Pattern A: 50 + B+C 분기 a: N + H: 3) status 정합화, version patch bump, updated_at 갱신, HISTORY row 추가
- [ ] **detector exemption 코드** (`terminalStatusEnum` + Check 분기) 적용 + 4 새 unit tests PASS
- [ ] **bulk script `.moai/scripts/status-drift-cleanup.go`** 작성 + 보존 (LSKC-001 fork)
- [ ] **affected-list-pattern-{A,B,C,H}.txt** 4 파일 작성 (BASELINE milestone 산출)
- [ ] **run-verification.md** Pattern B+C 10 row decision 기록
- [ ] **CHANGELOG `[Unreleased]` row 1줄** 추가 (sync-phase 범위 — manager-docs 책임)
- [ ] **plan-auditor PASS** (≥ 0.85, 목표 iter 1)
- [ ] **MX tags**: `@MX:NOTE` (terminalStatusEnum + exemption 분기) + `@MX:ANCHOR` (Check 함수, fan_in 검토 후) — bulk script는 코드 수정 아니므로 별도 태그 불요
- [ ] **frontmatter status 전이**: `draft → in-progress → implemented → completed` (각 phase entry 시점에 manager-* agent 가 갱신)
- [ ] **`go test ./...`** 전체 PASS, `go vet ./...` 0 issue, `golangci-lint run` 0 warning
- [ ] **REQ-SDF-010 검증**: `git diff main..HEAD .moai/specs/ | grep -c "lint.skip"` → 0 (새 lint.skip 도입 0건)
- [ ] **REQ-SDF-011 검증**: `git diff main..HEAD internal/spec/drift.go | wc -l` → 0 (walker filter 미변경)

---

## 7. Open Questions

### OQ1 — Pattern A의 일괄 downgrade 정책 (해석 1 vs 해석 2)

**Question**: Pattern A 50건의 frontmatter `completed` 가 사실에 부합하는지 (sync 표준 prefix 부재만 문제) vs 사실에 부합하지 않는지 (sync 자체가 미완료) 의 두 해석이 가능. 각각의 처리 방향:

- **해석 1 keep**: frontmatter `completed` 유지 + sync chore commit 표준 prefix 재발행 (50건 새 commit) → walker filter 가 정확히 인식 → mismatch 해소
- **해석 2 downgrade**: frontmatter `completed → implemented` 일괄 downgrade → 사실에 맞춤

**Default Decision**: **해석 2 (downgrade)** — research.md §5.1 / spec.md §1.3 적용 이유:

- 50건 새 commit 작성은 sweep cycle 재시작 위험 (LSKC-001 정신 위배)
- frontmatter `implemented` 는 의미상 잘못 아님 (코드 구현 완료, 문서 sync 표준 prefix 부재)
- 미래 sync chore prefix 컨벤션 정착 시 일괄 재승격 가능 (별도 SPEC)

**Action Required**: spec.md §1.3 + REQ-SDF-004 + design.md §4.1 으로 확정. run-phase 가 변경 가능 (사용자 확인 시).

### OQ2 — Pattern B+C 분기 (b) keep 의 허용 기준

**Question**: 10 SPEC verification 결과 분기 (b) keep (frontmatter status 유지 + run-verification.md 사유 기록) 가 어느 임계 수까지 허용 가능한가?

**Default Decision**: 0건 ideal, 3건까지 허용. 3건 초과 시 사용자 확인 (orchestrator AskUserQuestion).

**Rationale**:
- 사용자 옵션 (a) "status field bulk synchronization" 이 핵심 → keep은 본 SPEC scope 위반
- 그러나 일부 SPEC은 frontmatter 가 정확하나 SPEC ID 매칭 불일치 (예: rename, multi-SPEC PR) 으로 git-implied 가 부정확한 케이스
- 0-3건은 수용 가능 (run-verification.md 에 사유 명시)
- 4건 이상은 분류 자체가 잘못됐을 가능성 → 재검토

**Action Required**: M3 Wave 2 milestone에서 분기 (b) 카운트 추적. 4건 이상 시 escalate.

### OQ3 — Bulk Script 보존 결정

**Question**: `.moai/scripts/status-drift-cleanup.go` 를 PR 에 포함하여 보존 vs 작업 후 폐기 (PR 미포함)?

**Default Decision**: **PR 에 포함하여 보존** (LSKC-001 정신 계승).

**Rationale**:
- LSKC-001 의 `lint-skip-cleanup.go` 가 보존되어 본 SPEC fork 의 기반이 됨 (재사용 가치 입증)
- 미래 status drift cleanup (불가피하게 다시 발생 가능) 시 본 script 재사용
- 보존 비용 ~250 LOC (`//go:build ignore` tag 로 정상 빌드 제외)

**Alternative**: 폐기 → script LOC 증가 부담 회피, 그러나 미래 cleanup 시 재작성 필요.

**Action Required**: 기본 보존. run-phase가 변경 가능.

### OQ4 — Pattern H Wave 5 의 sync-phase vs run-phase 처리

**Question**: Pattern H 4건 (cleanup chain SPECs) 처리를 run-phase (Wave 5) vs sync-phase 어디에서?

**Default Decision**: **sync-phase**.

**Rationale**:
- run-phase 시점에는 본 SPEC sync chore commit 이 main에 없음 → 본 SPEC 자체의 H 패턴 측정 불가
- sync-phase 에서 sync PR squash 후 본 SPEC 자체 frontmatter 재정합화 가능
- sync-phase agent (manager-docs) 에게 위임 가능
- 단, sync-phase 가 자체 sync chore commit 작성 → 새 H 패턴 entry 가능 → 재귀 한 번 더 필요? (드물게 발생, design.md §6.5 mitigation)

**Alternative**: run-phase 에 처리 → 본 SPEC 자체 미포함 처리 (3건만, LSKC-001 / LINT-STATUS-CHORE-SKIP-001 / SPECLINT-DEBT-001) → run-phase 종료. sync-phase 에서 본 SPEC 자체 정합화.

**Action Required**: sync-phase agent (manager-docs) 호출 시점에 결정. 기본은 sync-phase.

---

## 8. Risk Assessment

| Risk | Severity | Likelihood | Mitigation |
|------|----------|------------|-----------|
| R1: BASELINE 측정 결과가 64에서 큰 변동 (예: ±20건) | Low | Medium | Wave 1 시작 전 affected-list 재산출. ±10 이내 정상. |
| R2: bulk script가 yaml.Node key ordering 손상 | Low | Low | LSKC-001 패턴 계승 (검증된 코드). dry-run 시 frontmatter diff 검증. |
| R3: Pattern B+C 분기 (b) keep 4건 이상 발생 | Medium | Low | OQ2 정책 — 사용자 확인 후 진행 또는 별도 SPEC 분리. |
| R4: detector exemption 코드 변경이 기존 lint test 회귀 | Medium | Low | Wave 3 시작 전 `grep -rn "superseded\|archived" internal/spec/lint*` 영향 분석. expected 갱신 적용. |
| R5: Pattern H sync-phase 재귀 cleanup이 또 다른 자기참조 발생 | Low | Low | design.md §6.5 mitigation. 1-stage 재귀로 충분, 2-stage 발생 시 별도 SPEC. |
| R6: 새 SPEC이 plan-phase 와 run-phase 사이 main에 추가됨 | Low | Medium | BASELINE milestone에서 실시간 측정으로 정정. ±5 SPEC 변동은 정상. |
| R7: Wave 3 코드 변경이 정확한 file/line 위치 파악 어려움 (lint.go가 PR #929 등으로 refactor) | Low | Medium | run-phase 첫 작업으로 `grep -rn "StatusGitConsistency" internal/spec/` 으로 정확한 위치 확인. |
| R8: bulk script 가 일부 SPEC frontmatter parse 실패 (비표준 yaml) | Medium | Low | dry-run 시점에 모든 64 SPEC parse 검증. 실패 SPEC manual edit. |

---

## 9. Dependencies

### 9.1 Predecessor (HARD)

- ✅ `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` PR #933 + #934 merged → walker filter 활성
- ✅ `SPEC-V3R4-LINT-SKIP-CLEANUP-001` PR #937 merged (main `758341089`) → 51 lint.skip 제거 → 64 WARN 노출

### 9.2 Successor (None)

본 SPEC 머지 후 후속 SPEC 없음. 4-단계 cleanup chain 종결.

미래 가능 후속 (별도 SPEC, 본 SPEC 의존 아님):
- sync chore commit prefix 컨벤션 정착 SPEC
- 정기 status drift audit SPEC (분기 (b) keep 사유들 후속 처리)

---

## 10. Branch & PR Strategy

- **Branch name**: `feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (Enhanced GitHub Flow §18.2 — feat prefix; SPEC 기반 변경)
- **Base**: `origin/main` (BODP A=¬ B=¬ C=¬ → main 결정)
- **Merge strategy**: **squash** (Enhanced GitHub Flow §18.3 — feat → main 은 squash; release/* 만 merge commit)
- **Worktree**: 미사용 (per `feedback_worktree_never_use` 정책 — main checkout 단일 작업공간)
- **PR title**: `feat(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — 64 SPEC status drift 일괄 동기화 (cleanup chain 종결)`
- **Commit message**: `chore(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — 64 SPEC status drift 동기화 + terminal state exemption` (Conventional Commits)
- **Reviewers**: 자동 — Release Drafter autolabeler가 `type:chore` + `area:spec-lint` 부착
- **CI**: Lint + Test (ubuntu/macos/windows) + Build 5 + CodeQL all green 대기. spec-lint job 의 `StatusGitConsistency` WARN count = 0 검증.

---

## 11. Plan-Audit Pre-Check

본 plan 의 plan-auditor 통과 가능성 향상을 위한 사전 점검:

| 항목 | 본 plan 의 충족 |
|------|----------------|
| EARS 5 modality 모두 사용 | ✅ Ubiquitous 3 + Event 4 + State 2 + Unwanted 5 + Optional 2 (spec.md §3) |
| REQ ↔ AC 매핑 100% | ✅ 14 REQ × 8 AC 매트릭스 완비 (spec.md §5) |
| Exclusions 명시 | ✅ spec.md §10 + plan.md §3.3 + design.md §10 |
| Worktree 정책 명시 | ✅ plan.md §10 (worktree 미사용 — feedback_worktree_never_use) |
| BODP 평가 기록 | ✅ HISTORY 0.1.0 row + plan.md §10 (A=¬ B=¬ C=¬ → main) |
| 9-field frontmatter | ✅ id, version, status:draft, created_at, updated_at, author, priority, labels, issue_number 모두 (spec.md frontmatter) |
| Pre-Write Frontmatter Checklist | ✅ id 정규식 매칭, status enum, priority Title-case (P1), labels array, version quoted string, created_at/updated_at ISO date |
| 의존 SPEC 명시 | ✅ research.md §4 + plan.md §9 |
| 위험 평가 | ✅ plan.md §8 (8 risks + mitigation) |
| Definition of Done | ✅ plan.md §6 (15 항목) |
| Open Questions | ✅ plan.md §7 (4 OQ + default decision) |

목표: plan-auditor iter 1 PASS, score ≥ 0.85.
