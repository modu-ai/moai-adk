# Implementation Plan — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-15 | manager-spec | 초기 plan draft. Wave 1 단일 wave (작은 변경 범위). 7개 task: T-LSCS-001 ~ T-LSCS-007. BODP 평가 (A=¬ B=¬ C=¬ → main @ origin/main). Branch/PR 전략 + Test fixture 설계 + Verification pipeline 정의. |

---

## 1. 변경 대상 파일

| 파일 | 변경 유형 | 추정 LOC | 비고 |
|------|-----------|----------|------|
| `internal/spec/drift.go` | 수정 | +25 / -10 | `getGitImpliedStatus` 함수 body 만 수정. 시그니처 유지. |
| `internal/spec/lint.go` | 수정 (방어적) | +3 / -0 | `StatusGitConsistencyRule.Check` 의 `gitStatus == ""` 가드 명시 추가. 기존 `if err != nil` 분기로도 cover 되나 명시화. |
| `internal/spec/drift_test.go` | 신규 | +180 | 5+ 시나리오 table-driven test. `t.TempDir()` + git init fixture. |
| `internal/spec/transitions_test.go` | 수정 (보강) | +15 / -0 | `chore(spec):`, `chore(specs):`, `revert:` 회귀 케이스 추가 (이미 있으면 skip). |

총 변경 추정: 약 +220 / -10 LOC. 단일 PR 으로 충분.

---

## 2. Task 분할 (Wave 1, 단일 Wave)

### T-LSCS-001: drift.go getGitImpliedStatus 로직 개선 (P1, 핵심)

**파일**: `internal/spec/drift.go:99-143`

**변경 내용**:

```go
// getGitImpliedStatus determines the status implied by git log for a SPEC.
// It scans up to maxGitLogScan commits mentioning the SPEC-ID and returns
// the status of the first non-skip commit. Commits with prefix
// "chore(spec):", "chore(specs):", or "revert:" are skipped (skip-meta /
// no-op categories per ClassifyPRTitle).
//
// Returns:
//   - (status, nil) when a meaningful lifecycle commit is found
//   - ("", nil) when all scanned commits are skip-meta/no-op (no meaningful history)
//   - ("", err) on git command failure or parse failure
const maxGitLogScan = 10

func getGitImpliedStatus(specID string) (string, error) {
    branch := "main"
    if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
        branch = "master"
    }

    cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges",
        "--grep="+specID, fmt.Sprintf("-%d", maxGitLogScan))
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("git log failed: %w", err)
    }
    if len(output) == 0 {
        return "", fmt.Errorf("no git history found for %s", specID)
    }

    scanner := bufio.NewScanner(strings.NewReader(string(output)))
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.SplitN(line, " ", 2)
        if len(parts) < 2 {
            continue
        }
        commitTitle := parts[1]
        category, status, err := ClassifyPRTitle(commitTitle)
        if err != nil {
            continue
        }
        // Skip sweep/revert commits — they do not represent lifecycle transitions
        if category == "skip-meta" || category == "no-op" {
            continue
        }
        if status == "" {
            // Unknown prefix — default to "in-progress" for partial work
            return "in-progress", nil
        }
        return status, nil
    }

    // All scanned commits were skip-meta/no-op — signal "no meaningful history"
    return "", nil
}
```

**핵심 변경**:
1. `-1` → `-10` (scan range expansion)
2. for-loop 로 commit 들을 순회하며 ClassifyPRTitle 호출
3. `category == "skip-meta" || category == "no-op"` 인 commit 은 건너뛴다
4. 모든 scan 결과가 skip-meta/no-op 이면 `("", nil)` 반환

**의존성**: 없음 (단독 task)

---

### T-LSCS-002: lint.go gitStatus == "" 가드 명시 추가 (P2, 방어적)

**파일**: `internal/spec/lint.go:886-914`

**변경 내용**: `StatusGitConsistencyRule.Check` 에서 `gitStatus == ""` 시 명시적 skip:

```go
func (r *StatusGitConsistencyRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
    fm := doc.Frontmatter
    var findings []Finding

    if fm.ID == "" || fm.Status == "" {
        return nil
    }

    gitStatus, err := getGitImpliedStatus(fm.ID)
    if err != nil {
        return nil
    }

    // New guard: empty gitStatus signals no meaningful git history
    // (e.g., all recent commits are chore(spec)/revert sweeps).
    // Skip the check to avoid false-positive WARNs.
    if gitStatus == "" {
        return nil
    }

    if fm.Status != gitStatus {
        findings = append(findings, Finding{
            File:     doc.Path,
            Line:     1,
            Severity: SeverityWarning,
            Code:     "StatusGitConsistency",
            Message:  fmt.Sprintf("SPEC %s frontmatter status '%s' disagrees with git-implied status '%s'", fm.ID, fm.Status, gitStatus),
        })
    }

    return findings
}
```

**의존성**: T-LSCS-001 (drift.go 가 `""` 반환할 수 있도록 먼저 수정).

---

### T-LSCS-003: drift_test.go 신규 작성 (P1, 핵심)

**파일**: `internal/spec/drift_test.go` (신규)

**테스트 시나리오** (table-driven):

```go
package spec

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// setupGitRepo creates a temp git repo with given commit titles (oldest first).
// Returns the repo path.
func setupGitRepo(t *testing.T, commits []string) string {
    t.Helper()
    dir := t.TempDir()
    runGit := func(args ...string) {
        cmd := exec.Command("git", args...)
        cmd.Dir = dir
        if out, err := cmd.CombinedOutput(); err != nil {
            t.Fatalf("git %v: %v\n%s", args, err, out)
        }
    }
    runGit("init", "-b", "main")
    runGit("config", "user.email", "test@example.com")
    runGit("config", "user.name", "Test")
    runGit("commit", "--allow-empty", "-m", "initial commit")
    for _, title := range commits {
        runGit("commit", "--allow-empty", "-m", title)
    }
    return dir
}

func TestGetGitImpliedStatus(t *testing.T) {
    tests := []struct {
        name     string
        specID   string
        commits  []string // oldest first; latest commit is the last entry
        wantStatus string
        wantErr  bool
    }{
        {
            name:   "chore(spec) sweep followed by latest feat",
            specID: "SPEC-V3R4-TEST-001",
            commits: []string{
                "feat(SPEC-V3R4-TEST-001): implement feature",
                "chore(spec): status drift sweep mentioning SPEC-V3R4-TEST-001",
            },
            wantStatus: "implemented",
        },
        {
            name:   "chore(spec) sweep followed by sync (completed)",
            specID: "SPEC-V3R4-TEST-002",
            commits: []string{
                "sync(SPEC-V3R4-TEST-002): docs and codemap",
                "chore(spec): status drift sweep mentioning SPEC-V3R4-TEST-002",
            },
            wantStatus: "completed",
        },
        {
            name:   "revert followed by feat",
            specID: "SPEC-V3R4-TEST-003",
            commits: []string{
                "feat(SPEC-V3R4-TEST-003): implement feature",
                "revert: revert prior change mentioning SPEC-V3R4-TEST-003",
            },
            wantStatus: "implemented",
        },
        {
            name:   "all skip-meta commits returns empty",
            specID: "SPEC-V3R4-TEST-004",
            commits: []string{
                "chore(spec): sweep 1 mentioning SPEC-V3R4-TEST-004",
                "chore(spec): sweep 2 mentioning SPEC-V3R4-TEST-004",
                "chore(spec): sweep 3 mentioning SPEC-V3R4-TEST-004",
            },
            wantStatus: "",
        },
        {
            name:   "vanilla feat commit (no skip)",
            specID: "SPEC-V3R4-TEST-005",
            commits: []string{
                "feat(SPEC-V3R4-TEST-005): initial implementation",
            },
            wantStatus: "implemented",
        },
        {
            name:    "no commits mention SPEC-ID returns error",
            specID:  "SPEC-V3R4-TEST-MISSING",
            commits: []string{"feat: unrelated change"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repoDir := setupGitRepo(t, tt.commits)
            origDir, _ := os.Getwd()
            t.Cleanup(func() { _ = os.Chdir(origDir) })
            if err := os.Chdir(repoDir); err != nil {
                t.Fatalf("chdir: %v", err)
            }

            got, err := getGitImpliedStatus(tt.specID)
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error, got status %q", got)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.wantStatus {
                t.Errorf("status = %q, want %q", got, tt.wantStatus)
            }
        })
    }
}
```

**중요 사항**:
- `t.TempDir()` 자동 정리 (CLAUDE.local.md §6 의무)
- `os.Chdir` 사용 후 `t.Cleanup` 으로 원복. 다만 병렬 테스트와 충돌 가능 — `t.Parallel()` 호출 금지.
- `--grep` 매칭 검증 시나리오 ("no commits mention SPEC-ID") 도 커버.
- `filepath` import 는 setup 함수에서 사용되지 않으므로 실제 작성 시 제거.

**의존성**: T-LSCS-001 (drift.go 수정 선행).

---

### T-LSCS-004: transitions_test.go 회귀 케이스 보강 (P3, 보완)

**파일**: `internal/spec/transitions_test.go`

**변경 내용**: 기존 `TestClassifyPRTitle` 의 table 에 다음 케이스가 없으면 추가:

```go
{
    name:           "chore(spec) sweep returns skip-meta",
    title:          "chore(spec): status drift sweep + lint-skip",
    wantCategory:   "skip-meta",
    wantStatus:     "",
},
{
    name:           "chore(specs) plural form returns skip-meta",
    title:          "chore(specs): batch frontmatter sync",
    wantCategory:   "skip-meta",
    wantStatus:     "",
},
{
    name:           "revert prefix returns no-op",
    title:          "revert: revert prior change",
    wantCategory:   "no-op",
    wantStatus:     "",
},
```

**의존성**: 없음. T-LSCS-001/002 와 병행 가능.

---

### T-LSCS-005: 로컬 lint 검증 (C1)

**명령**:
```bash
moai spec lint --strict
echo "exit code: $?"
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"  # expected: 0
```

**Expected**: exit 0, StatusGitConsistency 카운트 0 (51 lint.skip suppress 유지 + 신규 7건 logic fix 로 해소).

**의존성**: T-LSCS-001 ~ T-LSCS-004 완료.

---

### T-LSCS-006: CI green 검증 (C2)

**명령**: PR 생성 후 `gh pr view <PR>` 으로 spec-lint job 상태 확인.

**Expected**: spec-lint CI job GREEN.

**의존성**: T-LSCS-005 통과 후 PR push.

---

### T-LSCS-007: 회귀 없음 검증 (C3)

**명령**:
```bash
# 51개 lint.skip 등록 SPEC 들이 여전히 silent 인지 확인
grep -l "skip:" .moai/specs/SPEC-*/spec.md | xargs grep -l "StatusGitConsistency" | wc -l  # 51 이어야 함
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"  # 0 이어야 함
```

**Expected**: 51개 SPEC 의 lint.skip 자산 보존, 신규 WARN 0건.

**의존성**: T-LSCS-005 통과 후 회귀 측정.

---

## 3. 의존성 그래프

```
T-LSCS-001 (drift.go fix) ──┬──> T-LSCS-002 (lint.go guard)
                            │
                            ├──> T-LSCS-003 (drift_test.go)
                            │
                            └──> T-LSCS-005 (local lint) ──> T-LSCS-006 (CI) ──> T-LSCS-007 (regression)
T-LSCS-004 (transitions_test.go) ──> T-LSCS-005
```

병렬 실행 가능: T-LSCS-002, T-LSCS-003, T-LSCS-004 (모두 T-LSCS-001 후).

---

## 4. BODP 평가

본 SPEC plan-phase 의 BODP signals (CLAUDE.local.md §18.12 알고리즘):

| Signal | 값 | 근거 |
|--------|-----|------|
| A: depends_on + diff overlap | ¬ (false) | SPEC-V3R4-SPECLINT-DEBT-001 (이미 머지된 SPEC) 만 dependency. 작업 트리에 diff overlap 없음 |
| B: `.moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/` 디렉토리 git status 매칭 | ¬ (false) | 현재 시점에는 SPEC 디렉토리가 작업 트리에 부재. plan PR 으로 처음 생성. |
| C: 현재 브랜치 (`claude/issue-932-20260515-1325`) open PR | ¬ (false) | 본 PR 이 첫 PR. |

**결과**: ¬a ¬b ¬c → main @ origin/main (Decision Matrix 1행).

**Rationale**: 부모 의존 commit (SPEC-V3R4-SPECLINT-DEBT-001 머지) 이 main 에 안정적으로 있으며, 본 SPEC 의 plan-phase 는 markdown 만 추가하므로 main base 가 안전.

**Plan PR target branch**: `main`.
**Audit trail**: `.moai/branches/decisions/claude-issue-932-20260515-1325.md` (BODP audit) — orchestrator 가 자동 생성.

---

## 5. Run-Phase BODP (사전 결정)

Plan PR (#931) 머지 후 run-phase 진입 시 재평가 필요. 예상 signals:

| Signal | 예상 값 | 비고 |
|--------|---------|------|
| A | ¬ | 여전히 코드 의존 없음. |
| B | ¬ | SPEC 디렉토리는 main 에 머지됨. 작업 트리는 새 worktree 에서 clean. |
| C | ¬ | feat/ branch 신규 생성. |

→ main @ origin/main. `moai worktree new SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 --base origin/main` 으로 worktree 생성.

---

## 6. Verification Pipeline (Run-Phase 완료 직전)

```bash
# 1. Unit tests
go test -race ./internal/spec/...

# 2. Full test suite (Go test discipline: 항상 전체 실행)
go test -race ./...

# 3. Linting
golangci-lint run ./internal/spec/...

# 4. Build (template embedding 재생성 불요 — Go 코드만 변경)
go build ./...

# 5. SPEC lint
moai spec lint --strict
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"  # expected: 0

# 6. SPEC lint with --filter (specific rule)
moai spec lint --strict --filter StatusGitConsistency  # expected: no findings
```

모든 단계 통과 시 PR 생성.

---

## 7. 산출물 체크리스트

- [ ] `.moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/spec.md` (본 SPEC 본문)
- [ ] `.moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/plan.md` (본 파일)
- [ ] `.moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/acceptance.md` (Given-When-Then 시나리오)
- [ ] (run-phase 산출) `internal/spec/drift.go` 수정
- [ ] (run-phase 산출) `internal/spec/drift_test.go` 신규
- [ ] (run-phase 산출) `internal/spec/lint.go` 가드 추가
- [ ] (run-phase 산출) `internal/spec/transitions_test.go` 회귀 케이스 보강
- [ ] (run-phase 산출) `CHANGELOG.md` `[Unreleased]` 섹션 entry

---

## 8. Out of Scope (Plan 수준)

§spec.md §1.3 + §7 참조. plan-phase 에서 추가로 명시:

- 본 plan 은 단일 Wave 이므로 wave-split 의사결정 (lessons #9) 불필요.
- 본 plan 은 `--worktree` 옵션 사용 권장 (BODP signals 가 main @ origin/main 임에도 run-phase 에서는 worktree 사용이 spec-workflow.md Step 2 의무).
- evaluator-active scoring 적용 불필요 (harness level `standard`, 작은 변경 범위).

---

## 9. References

- spec.md §3 (EARS REQ-LSCS-001 ~ REQ-LSCS-010)
- spec.md §6 Risks (R1 성능, R2 word-boundary, R3 test 격리, R4 lint.skip 유지, R5 revert intent)
- `internal/spec/drift.go:99-143` (수정 대상)
- `internal/spec/transitions.go:22, 70-92` (skip-meta 매핑 참조)
- `internal/spec/lint.go:879-914` (StatusGitConsistencyRule 호출부)
- SPEC-V3R4-SPECLINT-DEBT-001 status-residuals.md §5 F1 (follow-up rationale)
- CLAUDE.local.md §6 Testing Guidelines (`t.TempDir()` 패턴)
- CLAUDE.local.md §18.3 Merge Strategy (squash for feature)
- CLAUDE.local.md §18.12 BODP (branch origin decision)
