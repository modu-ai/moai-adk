# Design — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 design. Status field model + detector contract + terminal-state exemption code sketch + bulk script architecture (LSKC-001 fork) + Pattern별 처ing 전략 + edge case 분석 + failure mode 분석 + backward compatibility 검토. |

---

## 1. Status Field Model

### 1.1 SPEC Lifecycle Status Enum (canonical 8-value)

```text
draft → planned → in-progress → implemented → completed
                                     ↓
                                 superseded   (terminal — supersession)
                                     ↓
                                 archived     (terminal — abandoned/legacy)
                                     ↓
                                 rejected     (terminal — explicitly refused)
```

Source: `internal/spec/status.go` (canonical enum) + `.claude/rules/moai/workflow/status-lifecycle.md` (lifecycle rules).

### 1.2 Detector 의 status field 의 두 종류

Detector (`StatusGitConsistencyRule`) 가 비교하는 두 status:

| 출처 | 변경 주체 | 본 SPEC 의 처리 방향 |
|------|----------|---------------------|
| frontmatter `status` (spec.md YAML) | 사람 / SPEC author / sync-phase agent | Pattern A/B/C/H: bulk script로 동기화 |
| git-implied (drift.go::getGitImpliedStatus) | walker filter + ClassifyPRTitle 출력 | 변경 없음 (walker filter 자체 미변경) |

본 SPEC 의 핵심 통찰: **frontmatter 를 git history 에 맞춤** (역방향 — git 을 frontmatter에 맞추는 것이 아님). 단 Pattern D/E/F/G 의 경우 frontmatter terminal state 의미가 정확하므로 detector 측에 narrow exemption 추가.

### 1.3 Terminal State 의 의미론적 특성

`superseded` / `archived` / `rejected` 3 terminal state 의 git history 와 frontmatter 관계:

- 실제 git history는 `implemented` 또는 `completed` 단계까지 발달했을 수 있음 (코드 작업이 완료된 후 supersede / archive 결정이 내려짐)
- frontmatter terminal state 는 lifecycle 결정의 명시적 기록 (lifecycle decision metadata)
- 따라서 frontmatter `superseded` + git-implied `completed` 는 mismatch가 정상 — false-positive
- detector exemption 으로 이를 인식

`rejected` 는 본 SPEC 의 64 drift 에는 포함되지 않으나 (현 main에 rejected SPEC 0건), 미래 호환성을 위해 exemption 에 포함.

---

## 2. Detector Contract (Current vs Target)

### 2.1 Current State (As-Is, post PR #933)

`internal/spec/lint.go::StatusGitConsistencyRule::Check` (Wave 3 코드 변경 대상 함수). 현 코드는 reading 시점에서 정확히 line 879-914 가 아닐 수 있으며 (lint.go 가 PR #929 등으로 refactor 됨), run-phase 시점에 정확한 line range 재확인:

```go
// 현 동작 (의사 코드)
func (r *StatusGitConsistencyRule) Check(ctx Context) []Finding {
    fm, err := ParseFrontmatter(ctx.SpecPath)
    if err != nil {
        return nil
    }
    
    // lint.skip suppression 검사 (LSKC-001 PR #937 머지 후에도 검사 자체는 보존)
    if hasLintSkip(fm, "StatusGitConsistency") {
        return nil
    }
    
    gitStatus, err := getGitImpliedStatus(fm.ID)
    if err != nil {
        // git history 조회 실패 시 skip (PR #933 walker filter 의 unknown 신호 포함)
        return nil
    }
    
    if fm.Status != gitStatus {
        return []Finding{{
            Severity: SeverityWarning,
            Rule:     "StatusGitConsistency",
            Message:  fmt.Sprintf("status mismatch: frontmatter %q vs git-implied %q", fm.Status, gitStatus),
        }}
    }
    
    return nil
}
```

### 2.2 Target State (To-Be, after Wave 3)

```go
// terminalStatusEnum 은 git history 와의 mismatch가 정상으로 간주되는 terminal lifecycle state
// @MX:NOTE: [AUTO] terminal state는 SPEC lifecycle 의 명시적 종료 표시 — git-implied가 ahead가 정상
// @MX:REASON: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Pattern D/E/F/G false-positive 해소
var terminalStatusEnum = map[string]bool{
    "superseded": true,
    "archived":   true,
    "rejected":   true,
}

func (r *StatusGitConsistencyRule) Check(ctx Context) []Finding {
    fm, err := ParseFrontmatter(ctx.SpecPath)
    if err != nil {
        return nil
    }
    
    if hasLintSkip(fm, "StatusGitConsistency") {
        return nil
    }
    
    // ★ 신규 exemption: terminal lifecycle state 는 git-implied 와 mismatch 가 정상
    // Pattern D (superseded → completed), E (superseded → implemented),
    // F (archived → implemented), G (archived → in-progress) 모두 통과
    if terminalStatusEnum[fm.Status] {
        return nil
    }
    
    gitStatus, err := getGitImpliedStatus(fm.ID)
    if err != nil {
        return nil
    }
    
    if fm.Status != gitStatus {
        return []Finding{{
            Severity: SeverityWarning,
            Rule:     "StatusGitConsistency",
            Message:  fmt.Sprintf("status mismatch: frontmatter %q vs git-implied %q", fm.Status, gitStatus),
        }}
    }
    
    return nil
}
```

### 2.3 동작 변화 요약

| 시나리오 | As-Is 결과 | To-Be 결과 | Pattern |
|---------|-----------|-----------|---------|
| frontmatter `completed`, git-implied `implemented` | WARN | WARN (변화 없음) | A — bulk script로 frontmatter downgrade로 해소 |
| frontmatter `completed`, git-implied `in-progress` | WARN | WARN (변화 없음) | B — bulk script로 frontmatter downgrade로 해소 |
| frontmatter `implemented`, git-implied `in-progress` | WARN | WARN (변화 없음) | C — bulk script로 frontmatter downgrade로 해소 |
| frontmatter `superseded`, git-implied `completed` | WARN | **no finding** | D — detector exemption |
| frontmatter `superseded`, git-implied `implemented` | WARN | **no finding** | E — detector exemption |
| frontmatter `archived`, git-implied `implemented` | WARN | **no finding** | F — detector exemption |
| frontmatter `archived`, git-implied `in-progress` | WARN | **no finding** | G — detector exemption |
| frontmatter == git-implied (정상 일치) | no finding | no finding | normal — 변화 없음 |
| frontmatter `rejected`, any git-implied | (현재 발생 안 함) | **no finding** | future-proof — rejected 도 terminal |

---

## 3. Bulk Script Architecture

### 3.1 LSKC-001 Script 자산

`.moai/scripts/lint-skip-cleanup.go` (PR #937 보존, ~220 LOC) 의 핵심 헬퍼 (재사용 대상):

| 함수 | 역할 | 본 SPEC 재사용성 |
|------|------|----------------|
| `parseFrontmatter([]byte) (*yaml.Node, []byte, error)` | YAML frontmatter + body 분리, yaml.Node 보존 | ✅ 동일 |
| `serializeFrontmatter(node *yaml.Node, body []byte) []byte` | yaml.Node + body 재조립 (key ordering 보존) | ✅ 동일 |
| `bumpVersionPatch(versionStr string) string` | semver patch +1 (`"0.3.0"` → `"0.3.1"`) | ✅ 동일 |
| `appendHistoryRow(body []byte, row string) []byte` | markdown HISTORY table 마지막 row 추가 | ✅ 동일 |
| `verifyBodySha256(specPath string, baseline string) error` | body byte preservation 검증 | ✅ 동일 |
| `removeLintSkip(node *yaml.Node, ruleName string) bool` | lint.skip 엔트리 제거 (LSKC-001 specific) | ❌ 본 SPEC 미사용 |

### 3.2 본 SPEC Script Architecture

신규 파일: `.moai/scripts/status-drift-cleanup.go` (LSKC-001 헬퍼 fork + 신규 dispatch).

```go
// pseudocode — run-phase 가 정확한 line/import 결정
package main

import (
    "fmt"
    "os"
    "strings"
    
    "gopkg.in/yaml.v3"
)

// Wave dispatcher
type Operation struct {
    SpecID       string
    Pattern      string  // "A" | "B" | "C" | "H"
    FromStatus   string  // 현 frontmatter status
    ToStatus     string  // 목표 frontmatter status
    HistoryDescr string  // HISTORY row description
}

func main() {
    // affected-list-pattern-{A,B,C,H}.txt 4 파일에서 ops 읽기
    ops := loadOperations()
    
    for _, op := range ops {
        if err := applyOperation(op); err != nil {
            fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", op.SpecID, err)
            os.Exit(1)
        }
    }
    
    fmt.Println("OK — applied", len(ops), "operations")
}

func applyOperation(op Operation) error {
    specPath := fmt.Sprintf(".moai/specs/%s/spec.md", op.SpecID)
    data, err := os.ReadFile(specPath)
    if err != nil {
        return err
    }
    
    fmNode, body, err := parseFrontmatter(data)
    if err != nil {
        return err
    }
    
    // Idempotency: 현재 status가 이미 목표 status면 no-op
    currentStatus := getMapValue(fmNode, "status")
    if currentStatus == op.ToStatus {
        return nil  // no change
    }
    if currentStatus != op.FromStatus {
        return fmt.Errorf("expected fromStatus=%q, got %q", op.FromStatus, currentStatus)
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
    
    // 5. 파일 재작성
    return os.WriteFile(specPath, serializeFrontmatter(fmNode, body), 0644)
}
```

### 3.3 affected-list 파일 형식

`affected-list-pattern-A.txt`:
```text
# Pattern A: completed → implemented (50 SPECs)
# Source: research.md §10 Appendix Table A
.moai/specs/SPEC-AGENCY-ABSORB-001/spec.md
.moai/specs/SPEC-AGENT-002/spec.md
... (50 lines)
```

각 파일은 frontmatter ID 추출 후 Operation 으로 변환 (Wave별 hard-coded `FromStatus` / `ToStatus` / `HistoryDescr`).

### 3.4 Idempotency 보장

LSKC-001 idempotency 패턴 계승 (run-phase 실측에서 검증된 효과):

- 1차 실행: 64 SPECs 변경 발생 → `git diff --stat` 캡처
- 2차 실행: `currentStatus == op.ToStatus` 분기로 no-op → 0 추가 변경

`go run .moai/scripts/status-drift-cleanup.go` 두 번 실행 후 `git diff` 가 1차와 동일하면 PASS.

### 3.5 Mid-Run Crash Recovery

LSKC-001 design.md §5.4 의 mid-run crash recovery 패턴 계승:

- 각 Operation 적용 후 SPEC `version` field bump → 부분 실행 후 재시작 시점에 어디까지 처리됐는지 추적 가능
- `currentStatus` 확인 분기가 자연스러운 idempotency 제공 → 재시작은 이미 처리된 SPEC 자동 skip
- crash 발생 시: `git status` 로 변경된 SPECs 확인 → script 재실행 → 자동 resume

---

## 4. Pattern별 처리 절차 상세

### 4.1 Wave 1: Pattern A (50 SPECs, bulk)

전제: research.md §3.2 가설 (i) 검증됨 — 50 SPECs 모두 sync chore commit이 `chore(spec):` skip → impl commit이 latest non-skip → git-implied `implemented`.

처리:
- `affected-list-pattern-A.txt` 작성 (50 lines)
- script Operation:
  - FromStatus: "completed"
  - ToStatus: "implemented"
  - HistoryDescr: "status drift FOLLOWUP cleanup — Pattern A: frontmatter completed → implemented (sync chore commit이 walker skip 대상이라 git-implied implemented). SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001."
- 검증: 50 SPECs 모두 frontmatter `status: implemented`, `version` patch bump, HISTORY row 추가

위험: 일부 SPEC 이 실제 `completed` 가 정확하고 단지 sync prefix 컨벤션 문제일 수 있음 → 본 SPEC은 사실에 맞춤 정책 (해석 2). 미래 sync chore prefix 컨벤션 정착 후 재승격 가능.

### 4.2 Wave 2: Pattern B (4) + Pattern C (6, 합계 10 SPECs, 개별)

전제: research.md §3.3 가설 검증됨 — 본 SPEC implementation이 다른 SPEC 우산 아래 진행되어 grep matching에 `feat:` 없음.

처리:
- 각 SPEC 개별 verification:
  1. `git log --oneline --no-merges --grep=<specID> -50` 출력 직접 확인
  2. 본 SPEC의 implementation이 다른 SPEC PR (예: #745, #746) 에 들어갔는지 project memory 검토
  3. 두 결정 분기:
     - **분기 (a) downgrade**: frontmatter status 가 사실에 부합하지 않음 → bulk script로 downgrade
     - **분기 (b) keep**: frontmatter status 가 사실에 부합 → run-verification.md 에 사유 기록 + sync-phase 에서 정확한 chore commit 추가 권장 (별도 작업, 본 SPEC scope 외)
- run-verification.md 형식:

```markdown
# Run Verification — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Wave 2

| SPEC ID | Pattern | Decision | Rationale |
|---------|---------|----------|-----------|
| SPEC-CORE-001 | B | downgrade | feat commit 부재 (PR #745 우산); frontmatter completed 가 사실에 부합하지 않음 — implementation은 RT-XXX 우산 |
| SPEC-LOOP-001 | B | downgrade | 동상 |
| SPEC-V3R4-CATALOG-001 | B | keep | feat commit이 PR #862에 있으나 grep 매칭 안 됨; sync-phase에서 정확한 chore commit 추가 |
| ... |
```

- 분기 (a) → bulk script Operation 추가
- 분기 (b) → 본 SPEC frontmatter 변경 0건, run-verification.md 만 작성

### 4.3 Wave 3: Pattern D + E + F + G (4 SPECs, detector exemption)

처리:
- `internal/spec/lint.go::StatusGitConsistencyRule::Check` (또는 동등 위치) 에 §2.2 코드 sketch 적용
- 새 unit tests 4 cases 추가:
  - case D: `frontmatter status: superseded` + git fixture로 git-implied `completed` → no finding
  - case E: `frontmatter status: superseded` + git-implied `implemented` → no finding
  - case F: `frontmatter status: archived` + git-implied `implemented` → no finding
  - case G: `frontmatter status: archived` + git-implied `in-progress` → no finding
- `terminalStatusEnum` map 정의 + 적용 (rejected 도 future-proof로 포함)
- 검증: `go test ./internal/spec/lint/checks/... -run TestStatusGitConsistency_TerminalState` PASS
- 4 SPEC frontmatter는 변경 없음 (검증: `git diff` 에 4 SPECs 없음)

### 4.4 Wave 4: Pattern G 통합 검증

Wave 3 detector exemption 이 archived 도 포함하므로 Pattern G (SPEC-V3R3-WEB-001) 가 자동 통과. 별도 코드 변경 불요. 검증만 수행:

- Wave 3 머지 후 `moai spec lint --strict | grep "SPEC-V3R3-WEB-001"` → 0 hit
- AC-SDF-001 의 sub-condition

### 4.5 Wave 5: Pattern H (3 SPECs, 재귀 cleanup, sync-phase)

본 SPEC sync-phase 에서 자체 재실행 (run-phase 종료 시점에는 본 SPEC 자체 sync chore commit이 작성되지 않은 상태).

처리:
- bulk script 재실행 with `affected-list-pattern-H.txt`:
  ```text
  .moai/specs/SPEC-V3R4-LINT-SKIP-CLEANUP-001/spec.md
  .moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/spec.md
  .moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/spec.md
  .moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/spec.md  # 본 SPEC 자체
  ```
- 각 SPEC의 git-implied 상태 측정 후 frontmatter 동기화:
  - LSKC-001: implemented → in-progress (sweep close 후 git-implied)
  - LINT-STATUS-CHORE-SKIP-001: completed → in-progress
  - SPECLINT-DEBT-001: completed → planned
  - 본 SPEC: (sync-phase 시점 측정 — 예상 implemented → in-progress)
- 검증: `moai spec lint --strict` 4 SPECs 대상 WARN 0건

대안: Wave 5는 별도 sub-PR로 분리 가능 (sync-phase chain 의 자기참조 방지). run-phase 결정.

---

## 5. Edge Cases

### 5.1 Pattern A 의 일부 SPEC이 sync 표준 prefix로 commit됨

**Scenario**: 50 Pattern A SPECs 중 일부가 실제로 `docs(sync): SPEC-XXX sync` 표준 commit 보유 → walker가 skip하지 않고 hit → git-implied `completed` → frontmatter 와 일치 → drift 아님.

**Behavior**: BASELINE 측정 시점 (Wave 1 milestone) 에 `moai spec lint --strict` 출력의 정확한 SPEC list 를 affected-list-pattern-A.txt 로 산출. 가설로 만든 50 SPEC list (research.md §10) 와 차이 가능. run-phase 실시간 측정으로 정정.

**Mitigation**: 
- BASELINE milestone에서 actual list 캡처
- script Operation의 `FromStatus == currentStatus` precondition으로 의도하지 않은 SPEC 변경 차단
- 만약 실측 결과 50 미만이면 spec.md §1.3 표 의 count는 plan-phase 추정으로 명시 (research.md §10 Note 참조)

### 5.2 Pattern B/C verification 결과 모두 분기 (b) keep

**Scenario**: 10 SPECs 모두 frontmatter status 가 사실에 부합하나 chore commit 부재로 detector mismatch.

**Behavior**: 
- bulk script Operation 0건
- run-verification.md 에 10 row 모두 `decision: keep` + 사유 기록
- `moai spec lint --strict` WARN 10건 잔존 (AC-SDF-001 fail)

**Mitigation**:
- 사용자 옵션 (a) 정의에 따르면 "status field bulk synchronization" 이 핵심 → 분기 (b) keep은 실질적으로 본 SPEC scope 위반
- 정책: 모든 Pattern B/C SPEC 은 default 분기 (a) downgrade 적용. 분기 (b) keep 은 명시적 사유 (예: "PR #745 squash 시 frontmatter 가 정정됨, git history 만 stale") 가 있을 때만 허용
- run-phase에서 분기 (b) keep 결정이 3건 이상 발생 시 사용자 확인 (AskUserQuestion via orchestrator)

### 5.3 Detector Exemption 코드 변경이 다른 lint test 파괴

**Scenario**: `terminalStatusEnum` 추가가 기존 `internal/spec/lint_test.go` 테스트 일부를 fail시킴 (예: 기존 테스트가 `superseded → completed` mismatch 를 WARN으로 기대).

**Behavior**: `go test ./internal/spec/...` 에서 회귀 발생.

**Mitigation**:
- run-phase Wave 3 시점에 기존 테스트 영향 분석 (`grep -rn "superseded" internal/spec/`)
- 영향받는 테스트의 expected output 갱신 (terminal state는 이제 no finding)
- 새 테스트 (case D/E/F/G) 추가 + 영향받는 기존 테스트 expected 변경

### 5.4 Mid-Run Crash During Wave 1

**Scenario**: 50 Pattern A SPECs 처리 중 25번째에서 script crash (예: yaml parse 오류 — 일부 SPEC frontmatter가 비표준 형식).

**Behavior**:
- crash 직후: 24 SPECs 변경됨 (frontmatter status: implemented), 26 SPECs 미변경
- script 재실행 시 idempotency 분기:
  - 처리된 24 SPECs: `currentStatus == ToStatus` → no-op
  - 25번째 SPEC: yaml parse 오류 재발생 → 디버깅 후 fix 필요
  - 미처리 25 SPECs: 정상 처리

**Mitigation**:
- script error message에 SPEC ID + 정확한 line 출력
- BASELINE milestone에서 `go run .moai/scripts/status-drift-cleanup.go --dry-run` 옵션으로 사전 parse 검증 (모든 SPEC frontmatter 가 yaml.Node 로 파싱 가능한지)
- crash 시점부터 manual intervention 가이드 (run-verification.md 에 기록)

### 5.5 Pattern H 재귀 cleanup이 본 SPEC 자체를 변경

**Scenario**: Wave 5 sync-phase 시점에 본 SPEC frontmatter `status: implemented` 가 git-implied `in-progress` 와 mismatch → script 가 본 SPEC 자체 frontmatter 변경.

**Behavior**: 본 SPEC frontmatter `version` patch bump + status 변경 + HISTORY row 추가 → sync PR 의 스코프에 본 SPEC 자체 변경 포함

**Mitigation**:
- 정상 동작 (LSKC-001 design §5 의 mid-run resume 패턴 계승)
- sync PR title에 "Wave 5 자기참조 cleanup 포함" 명시
- 만약 본 SPEC sync 후에도 자기참조 잔존 (sync chore commit이 또다시 walker skip 대상) → terminal state exemption (Wave 3) 으로 보호 받지 않음 → 다음 SPEC scope (가능성 낮음 — H pattern은 2-stage cycle만 발생)

### 5.6 frontmatter status enum이 8 외 값 보유

**Scenario**: 일부 SPEC frontmatter `status` 가 enum 외 값 (예: 오타 `complted`).

**Behavior**: `terminalStatusEnum[fm.Status]` 가 false → 기존 mismatch 검사 진행 → WARN.

**Mitigation**: 별도 lint rule (`StatusEnumValidation`) 책임. 본 SPEC scope 외. 실측 시점에 해당 SPEC 발견되면 별도 수정.

---

## 6. Failure Mode Analysis

### 6.1 BASELINE 측정 결과가 64 ± 변동

**Trigger**: plan-phase 와 run-phase 사이에 새 commit / SPEC이 main에 추가됨 → 실측 WARN 수가 64에서 변동.

**Behavior**: affected-list 가 plan-phase 추정과 다름. Pattern 분류 재계산 필요.

**Mitigation**: 
- BASELINE milestone에서 `moai spec lint --strict` 실시간 출력으로 final affected-list 확정
- spec.md §1.3 의 count는 plan-phase 추정으로 명시 (research.md §10 Note)
- ±10 범위 변동은 정상; 큰 변동 시 (예: 100건 이상) plan 재검토

### 6.2 Pattern A bulk downgrade 후 WARN 잔존

**Trigger**: Wave 1 적용 후 50 Pattern A SPECs 처리됐으나 lint --strict 가 여전히 일부 SPEC 에 WARN 보고.

**Behavior**: Pattern A 분류가 정확하지 않거나 (`completed` → 다른 git-implied), 새로 노출된 case.

**Mitigation**:
- Wave 1 verification milestone에서 `moai spec lint --strict | grep -E "$(cat affected-list-pattern-A.txt | sed 's|.moai/specs/||;s|/spec.md||' | tr '\n' '|')"` → 예상 0건
- 잔존 시: SPEC 별 git history 재조사 + Pattern 재분류

### 6.3 Detector Exemption이 false-negative 유발

**Trigger**: Wave 3 적용 후 의도하지 않은 SPEC (terminal state로 표기됐으나 실제로 active 상태) 이 false-negative로 통과.

**Behavior**: lint --strict 가 정상으로 보고하나 실제 status drift 잔존.

**Mitigation**:
- terminal state 4 SPECs (LSP-001, V3R3-HARNESS-001, I18N-001-ARCHIVED, V3R3-WEB-001) 의 frontmatter `status` enum이 정확함을 plan-phase에서 verify (research.md §3.4)
- 새 terminal state SPEC 추가 시 별도 audit (별도 SPEC 또는 정기 review)

### 6.4 Bulk Script가 yaml.Node Key Ordering 손상

**Trigger**: `gopkg.in/yaml.v3` 의 yaml.Node API 사용 미숙으로 frontmatter key 순서 변경됨.

**Behavior**: `git diff` 에 frontmatter 영역 노이즈 발생, body 변경 0이지만 frontmatter 변경 라인 증가.

**Mitigation**:
- LSKC-001 script (`.moai/scripts/lint-skip-cleanup.go`) 의 yaml.Node 패턴 그대로 계승 (이미 PR #937에서 검증됨)
- script 작성 후 1개 SPEC 으로 dry-run → `git diff --stat` 의 frontmatter 변경 line 수 확인 (예상: status + version + updated_at + HISTORY = 4-5줄)

### 6.5 sync-phase 에서 cleanup 자체가 새 chore commit 발생 → 재회귀

**Trigger**: 본 SPEC sync PR squash merge 후 main에 `chore(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 sync` commit 발생 → walker skip → 본 SPEC frontmatter implemented 일 때 git-implied 가 다시 잘못된 단계.

**Behavior**: 본 SPEC 자체가 새 H pattern entry → `moai spec lint --strict` WARN 1건 재발.

**Mitigation**:
- Wave 5 sync-phase 에서 본 SPEC 자체 정합화 (frontmatter `status: completed` → 실제 sync 후 git-implied 와 정합)
- LSKC-001 design §5.4 mid-run resume 패턴: 본 SPEC sync 시점에 bulk script 재실행으로 자동 흡수
- 미해결 시: 별도 SPEC 또는 정기 audit (드물게 발생, 영향 1 SPEC 규모)

### 6.6 lint.skip 우회 시도 (REQ-SDF-010 위반 압력)

**Trigger**: 일부 SPEC 처리가 어려워 회피책으로 `lint.skip: [StatusGitConsistency]` 추가하고 싶은 유혹.

**Behavior**: LSKC-001 정신 위배, 회피책 cycle 재시작.

**Mitigation**:
- REQ-SDF-010 HARD 규정으로 절대 금지
- run-phase 점검: `git diff main..HEAD .moai/specs/ | grep "lint.skip"` → 0 hit 보장
- plan-auditor 검사 항목에 "no new lint.skip introduction" 추가

---

## 7. Backward Compatibility

본 SPEC 의 변경은 다음 측면에서 backward-compatible:

| 측면 | 호환성 보장 방법 |
|------|----------------|
| `getGitImpliedStatus` signature | 변경 없음 (drift.go 미수정) |
| `shouldSkipCommitTitle` 동작 | 변경 없음 (walker filter scope 미확장 — REQ-SDF-011) |
| `ClassifyPRTitle` prefix 매핑 | 변경 없음 (transitions.go 미수정) |
| `StatusGitConsistencyRule` 외부 인터페이스 | 변경 없음 (Check 함수 signature 보존) |
| `StatusGitConsistencyRule::Check` 동작 | terminal state 4건만 추가 pass-through. Active SPEC 전체에 대해 동작 동일 |
| SPEC frontmatter status enum | 변경 없음 (8-value 그대로) |
| `lint.skip` mechanism | 변경 없음 (LSKC-001 정신 보존, 새 entry 도입 0건) |
| CLI flag (`moai spec lint`, `moai spec lint --strict`) | 변경 없음 |
| Exit code | 변경 없음 (다만 같은 SPEC set에 대해 strict 모드 0 exit code 로 전환되는 효과) |

기존 호출자 또는 통합 도구가 본 SPEC 변경으로 깨질 가능성: **없음**.

---

## 8. @MX Tag Placement (Wave 3 코드 변경 가이드)

`internal/spec/lint.go` (또는 `internal/spec/lint/checks/status_git_consistency.go`) 수정 시 다음 태그 권장 (mx-tag-protocol.md 준수):

| 위치 | 태그 종류 | 사유 |
|------|----------|------|
| `terminalStatusEnum` map 정의 | `@MX:NOTE` | terminal state 의미 명시 + Pattern D/E/F/G 처리 의도 |
| `StatusGitConsistencyRule::Check` 함수 | `@MX:ANCHOR` (fan_in 검토 후) | lint rule 진입점, fan_in ≥ 3 가능 (lint engine 여러 곳 호출) |
| terminal state exemption `if terminalStatusEnum[fm.Status]` 분기 | `@MX:NOTE` | "terminal state 는 git history 와 mismatch 가 정상" 의도 |
| 새 unit tests (case D/E/F/G) | `@MX:TEST` | AC-SDF-004 의 testable assertion |

`@MX:REASON` 모든 ANCHOR/WARN 에 한국어 (code_comments=ko).

`@MX:SPEC: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` 모든 새 태그에 부착.

---

## 9. Dependencies and Sequencing

### 9.1 Predecessor Dependency (HARD)

본 SPEC 시작 전 필수 머지 상태:

- ✅ `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` PR #933 (run) + #934 (sync) merged → walker filter 활성
- ✅ `SPEC-V3R4-LINT-SKIP-CLEANUP-001` PR #937 merged (main `758341089`) → 51 lint.skip 제거 → 64 WARN 노출

### 9.2 Wave Ordering

엄격한 순서 (병렬 금지 — bulk lint 검증 baseline 충돌):

```
BASELINE → Wave 1 (Pattern A) → Wave 2 (Pattern B+C) → Wave 3 (Pattern D-G detector) → Wave 4 (G 통합 검증) → Wave 5 (Pattern H, sync-phase)
```

각 Wave 완료 후 `moai spec lint --strict` 실행으로 누적 효과 측정 (예상: BASELINE 64 → Wave 1 후 14 → Wave 2 후 ~4 → Wave 3 후 0 → Wave 5 후 0 유지).

### 9.3 Successor (None)

본 SPEC 머지 후 후속 SPEC 없음. 4-단계 cleanup chain 종결.

미래 가능 후속 (별도 SPEC, 본 SPEC 의존 아님):
- sync chore commit prefix 컨벤션 정착 SPEC (`docs(sync):` 표준화)
- terminal state exemption 코드를 별도 lint rule 로 분리 (refactor)

---

## 10. Configuration & Environment

본 SPEC 은 다음 외부 설정에 의존하나 변경하지 않음:

| 파일 / 설정 | 의존 사유 | 변경 여부 |
|------------|----------|----------|
| `.moai/config/sections/language.yaml` | code_comments=ko (코드 주석 한국어) | 변경 없음 |
| `internal/spec/transitions.go::transitionRules` | ClassifyPRTitle prefix 매핑 (Wave 1/2 검증) | 변경 없음 |
| `internal/spec/drift.go::shouldSkipCommitTitle` | walker filter (변경 금지 per REQ-SDF-011) | 변경 없음 |
| `internal/spec/status.go` | lifecycle status enum (terminal state 정의 source) | 변경 없음 |
| `.github/workflows/spec-lint.yml` | CI lint job 실행 환경 | 변경 없음 |
