# BODP Audit Trail 경로 누수 진단 보고서

**작성일**: 2026-05-09
**작성자**: expert-debug
**대상 SPEC**: SPEC-V3R3-CI-AUTONOMY-001 Wave 7 (T8)
**PR**: #794 (분석 시점 OPEN)

---

## Summary

`internal/cli/worktree/new.go`의 `writeWorktreeAuditTrail()` 함수는 BODP audit trail 파일을 git 리포지터리 루트(repo root) 기준으로 작성해야 하지만, 실제로는 `os.Getwd()`의 반환값을 그대로 `repoRoot`로 사용한다. Go 테스트는 패키지 디렉터리를 cwd로 실행하므로, `go test ./internal/cli/worktree/...`를 실행하면 audit trail 파일이 `internal/cli/worktree/.moai/branches/decisions/` 경로에 생성된다.

누수를 일으키는 구체적인 테스트는 두 종류다. (1) `subcommands_test.go`의 `TestRunNew_Success`와 `TestRunNew_AddError` — 두 테스트 모두 `t.Chdir()` 없이 `cmd.RunE`를 직접 호출하여 `os.Getwd()`가 패키지 cwd(`internal/cli/worktree/`)를 반환하게 한다. (2) `new_test.go`의 `TestRunNew_WithTmuxFlag_TmuxAvailable`와 `TestRunNew_TmuxNotAvailable_GracefulDegradation` — 마찬가지로 `t.Chdir()`를 호출하지 않아 동일한 경로 누수가 발생한다. 그 결과 5개의 fixture 파일 (`feature-x.md`, `fix-something.md`, `feature-SPEC-AUTH-001.md`, `feature-SPEC-UI-042.md`, `feature-SPEC-TEST-001.md`)이 패키지 서브디렉터리에 untracked 상태로 생성되었다.

CLAUDE.local.md §6 "Test Isolation" 원칙("모든 테스트 임시 디렉터리는 `t.TempDir()` 사용, 자동 cleanup")을 위반한다.

---

## Root Cause

### 1차 원인: `writeWorktreeAuditTrail()` — `os.Getwd()` 를 repoRoot로 사용

**파일**: `internal/cli/worktree/new.go`
**라인**: 185–195

```go
// new.go:185
cwd, err := os.Getwd()
if err != nil {
    return fmt.Errorf("getwd: %w", err)
}
decision, _ := bodp.Check(bodp.CheckInput{
    ...
    RepoRoot: cwd,   // ← os.Getwd() 직접 주입
    ...
})
return bodp.WriteDecision(cwd, bodp.AuditEntry{ ... }) // ← os.Getwd()를 repoRoot로 전달
```

`bodp.WriteDecision(repoRoot, entry)` (audit_trail.go:31)는 `repoRoot`를 파일 경로의 base로 사용한다:
```go
fullPath := filepath.Join(repoRoot, auditTrailDir, fname)
```

테스트 실행 시 `os.Getwd()`가 반환하는 값은 패키지 디렉터리(`internal/cli/worktree/`)이므로, audit trail이 `internal/cli/worktree/.moai/branches/decisions/`에 작성된다. 이는 git repo root(프로젝트 최상위)가 아니다.

### 2차 원인: 테스트에 `t.Chdir()` 또는 `t.TempDir()` 격리 부재

**파일**: `internal/cli/worktree/subcommands_test.go`
**라인**: 68–108 (`TestRunNew_Success`), 110–133 (`TestRunNew_AddError`)

두 테스트 모두:
- `t.TempDir()`로 임시 루트를 생성하지 않음
- `t.Chdir(tempDir)`를 호출하지 않음
- `cmd.RunE`를 호출하면 `writeWorktreeAuditTrail()`이 내부적으로 `os.Getwd()` → 패키지 cwd 사용

**파일**: `internal/cli/worktree/new_test.go`
**라인**: 342–432 (`TestRunNew_WithTmuxFlag_TmuxAvailable`), 434–492 (`TestRunNew_TmuxNotAvailable_GracefulDegradation`)

두 테스트 모두 `t.TempDir()`로 `userHomeDirFunc`를 대체하지만, `t.Chdir()`를 호출하지 않아 `os.Getwd()` 경로 누수가 동일하게 발생한다.

### 왜 `TestNew_AuditTrailWritten`은 정상인가?

`new_test.go:588`의 `TestNew_AuditTrailWritten`은 `t.Chdir(tempDir)`를 호출하므로 `os.Getwd()`가 tempDir을 반환한다. 이 테스트는 정상이다.

---

## Reproduction Steps

```bash
# 재현: 테스트 실행 전 패키지 디렉터리 확인
ls internal/cli/worktree/.moai/branches/decisions/ 2>/dev/null && echo "EXISTS" || echo "NOT EXIST"

# 테스트 실행 (단, 이미 파일이 있으므로 동일 파일 재생성)
go test -count=1 -run 'TestRunNew_Success|TestRunNew_AddError' ./internal/cli/worktree/

# 재현 확인
ls internal/cli/worktree/.moai/branches/decisions/
```

기대 결과: `feature-x.md` 등이 다시 생성됨.

파일 내용에서 `current_branch: feat/SPEC-V3R3-CI-AUTONOMY-001-wave-7` 가 기록된 것은, 테스트 실행 당시 실제 git cwd(패키지 디렉터리)에서 `git rev-parse --abbrev-ref HEAD`를 실행했기 때문이다. 즉 실제 git 상태가 audit trail에 그대로 노출된다.

---

## AC-CIAUT-019 Compliance Verdict

**판정**: **N/A** (해당 없음)

AC-CIAUT-019 및 AC-CIAUT-019b는 BODP의 _기능 동작_(stacked PR 권장, CLI audit trail 기록)을 검증하는 AC다. test isolation 관련 AC는 SPEC-V3R3-CI-AUTONOMY-001 acceptance.md에 별도 항목으로 정의되어 있지 않다.

적용 가능한 정책 기준은 **CLAUDE.local.md §6 Test Isolation**이다:

> [HARD] All test temp directories MUST be created under `/tmp` and cleaned up automatically.
> Use `t.TempDir()` for all temporary directories.

이 기준으로 판정하면 **FAIL**:
- `TestRunNew_Success` (subcommands_test.go:68): `t.TempDir()` / `t.Chdir()` 누락 → audit trail을 패키지 cwd에 작성
- `TestRunNew_AddError` (subcommands_test.go:110): 동일
- `TestRunNew_WithTmuxFlag_TmuxAvailable` (new_test.go:342): `t.Chdir()` 누락 → 동일
- `TestRunNew_TmuxNotAvailable_GracefulDegradation` (new_test.go:434): 동일

---

## Fix Proposal — Code Changes

### 방안 A (권장): `writeWorktreeAuditTrail()`에 repoRoot 파라미터 주입

`os.Getwd()`를 함수 내부에서 호출하는 대신, 호출자인 `runNew()`에서 git repo root를 결정하여 주입한다.

**Before** (`internal/cli/worktree/new.go:182`):
```go
func writeWorktreeAuditTrail(specID, branchName, base, _ string) error {
    currentBranch, _ := gitWorktreeCmd("rev-parse", "--abbrev-ref", "HEAD")
    currentBranch = strings.TrimSpace(currentBranch)
    cwd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("getwd: %w", err)
    }
    decision, _ := bodp.Check(bodp.CheckInput{
        CurrentBranch: currentBranch,
        NewSpecID:     specID,
        RepoRoot:      cwd,
        EntryPoint:    bodp.EntryWorktreeCLI,
    })
    return bodp.WriteDecision(cwd, bodp.AuditEntry{ ... })
}
```

**After** (`internal/cli/worktree/new.go:182`):
```go
// writeWorktreeAuditTrail records the BODP decision for a CLI-driven worktree
// creation. repoRoot MUST be the git repository root (from `git rev-parse
// --show-toplevel`), NOT os.Getwd() — the two diverge during test execution.
func writeWorktreeAuditTrail(repoRoot, specID, branchName, base string) error {
    currentBranch, _ := gitWorktreeCmd("rev-parse", "--abbrev-ref", "HEAD")
    currentBranch = strings.TrimSpace(currentBranch)
    decision, _ := bodp.Check(bodp.CheckInput{
        CurrentBranch: currentBranch,
        NewSpecID:     specID,
        RepoRoot:      repoRoot,
        EntryPoint:    bodp.EntryWorktreeCLI,
    })
    return bodp.WriteDecision(repoRoot, bodp.AuditEntry{ ... })
}
```

`runNew()` 내부에서 호출 시 (`new.go:120`):
```go
// Before:
if err := writeWorktreeAuditTrail(specID, branchName, effectiveBase, wtPath); err != nil {

// After:
repoRoot, err := gitRepoRoot()
if err != nil {
    _, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: cannot determine repo root: %v\n", err)
    repoRoot = "." // fallback: non-fatal
}
if err := writeWorktreeAuditTrail(repoRoot, specID, branchName, effectiveBase); err != nil {
```

`gitRepoRoot()` 헬퍼 추가 (`new.go` 하단):
```go
// gitRepoRoot returns the absolute path of the git repository root via
// `git rev-parse --show-toplevel`. Overridable in tests.
var gitRepoRootFunc = func() (string, error) {
    out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
    return strings.TrimSpace(string(out)), err
}

func gitRepoRoot() (string, error) {
    return gitRepoRootFunc()
}
```

이 접근 방식의 장점:
1. 프로덕션: `git rev-parse --show-toplevel`로 항상 정확한 repo root 사용
2. 테스트: `gitRepoRootFunc`를 `tempDir` 반환 함수로 대체 가능
3. 다른 테스트(`TestNew_AuditTrailWritten`)와 동일한 패턴 — `t.Chdir(tempDir)` 방식 유지

---

## Fix Proposal — Test Changes

### 방안 B: 기존 테스트에 `t.Chdir()` 추가 (단기 완화)

`writeWorktreeAuditTrail()`에 repoRoot 파라미터를 추가하지 않더라도, 기존 테스트에 `t.Chdir(t.TempDir())`만 추가하면 패키지 cwd 오염을 방지할 수 있다.

**`TestRunNew_Success` (subcommands_test.go:68) — Before**:
```go
func TestRunNew_Success(t *testing.T) {
    origProvider := WorktreeProvider
    defer func() { WorktreeProvider = origProvider }()
    // ... (t.Chdir 없음)
    err := cmd.RunE(cmd, []string{"feature-x"})
```

**After**:
```go
func TestRunNew_Success(t *testing.T) {
    t.Chdir(t.TempDir()) // audit trail을 tempDir에 격리
    origProvider := WorktreeProvider
    defer func() { WorktreeProvider = origProvider }()
    // ...
    err := cmd.RunE(cmd, []string{"feature-x"})
```

동일 패턴을 다음 테스트에도 적용:
- `TestRunNew_AddError` (subcommands_test.go:110)
- `TestRunNew_WithTmuxFlag_TmuxAvailable` (new_test.go:342): 각 subtest `t.Run` 내부에 `t.Chdir(t.TempDir())`
- `TestRunNew_TmuxNotAvailable_GracefulDegradation` (new_test.go:434)

### 방안 A+B 조합 (최적)

방안 A (코드 측 repoRoot 주입)와 방안 B (테스트 측 `t.Chdir()`)를 함께 적용하면:
- 프로덕션 코드가 항상 올바른 경로를 사용 (A)
- 테스트가 패키지 cwd를 오염시키지 않음 (B)
- 방어적 이중 보호 구조

**테스트에서 gitRepoRootFunc 오버라이드 (방안 A 선택 시)**:
```go
func TestRunNew_Success(t *testing.T) {
    tmpDir := t.TempDir()
    t.Chdir(tmpDir)

    // gitRepoRootFunc 오버라이드
    origGitRepoRoot := gitRepoRootFunc
    gitRepoRootFunc = func() (string, error) { return tmpDir, nil }
    defer func() { gitRepoRootFunc = origGitRepoRoot }()

    // ... 나머지 동일
}
```

---

## Recommended Application Window

### 권장: 이번 PR #794 내 fixup commit

**근거**:
1. 누수 파일 5개가 현재 untracked 상태로 존재하며 git status를 오염시키고 있음
2. PR #794는 Wave 7 (T8 BODP) 최종 PR이며, BODP 관련 테스트가 포함되어 있음 — 해당 테스트 수정이 PR 범위에 자연스럽게 포함됨
3. 수정 범위가 작음: `new.go` 함수 시그니처 변경 + 헬퍼 추가, 테스트 4개에 `t.Chdir()` 또는 `gitRepoRootFunc` 오버라이드 추가
4. PR #794가 아직 OPEN 상태이므로 force push 없이 추가 commit 가능

**대안: 별도 fixup PR**

만약 PR #794 범위 유지를 원한다면 별도 PR(`fix/bodp-audit-trail-cwd-leak`)을 생성하고 squash merge. 단, 이 경우 untracked 파일 5개가 PR #794 머지 후에도 계속 남아 있어 `git status`를 오염시킴. 권장하지 않음.

**적용 순서** (PR #794에 추가하는 경우):
1. `internal/cli/worktree/new.go`: `gitRepoRootFunc` var 추가, `writeWorktreeAuditTrail()` 시그니처 변경
2. `internal/cli/worktree/subcommands_test.go`: `TestRunNew_Success`, `TestRunNew_AddError`에 `t.Chdir(t.TempDir())` 추가
3. `internal/cli/worktree/new_test.go`: `TestRunNew_WithTmuxFlag_TmuxAvailable`, `TestRunNew_TmuxNotAvailable_GracefulDegradation`에 `t.Chdir(t.TempDir())` 추가
4. 로컬에서 `go test ./internal/cli/worktree/... ./internal/bodp/...` 실행 — 패키지 cwd에 새 파일이 생기지 않음을 확인
5. 5개 untracked 파일 수동 삭제: `rm -rf internal/cli/worktree/.moai/`

---

## 파일 위치 요약

| 파일 | 역할 | 관련 라인 |
|------|------|----------|
| `internal/cli/worktree/new.go` | `writeWorktreeAuditTrail()` — os.Getwd() 오용 | 182–203 |
| `internal/cli/worktree/subcommands_test.go` | `TestRunNew_Success`, `TestRunNew_AddError` — t.Chdir 누락 | 68–133 |
| `internal/cli/worktree/new_test.go` | tmux 테스트 2개 — t.Chdir 누락 | 342–492 |
| `internal/bodp/audit_trail.go` | `WriteDecision()` — repoRoot 파라미터로 경로 구성 (정상) | 31–45 |
