# Acceptance Criteria: SPEC-WORKTREE-002

## Overview

This document defines acceptance criteria for the Worktree End-to-End Automation feature using Given-When-Then format.

---

## AC1: Execution Mode Selection Gate

### Scenario 1.1: Tmux Available with CC Mode

**Given** user has completed SPEC creation with `--worktree` flag
**And** tmux is available on the system
**And** active mode is CC (no GLM env)
**When** user selects "Start Implementation"
**Then** system presents three options:
  - "Worktree + CC (Recommended)"
  - "Team Mode"
  - "Sub-agent Mode"
**And** on selecting "Worktree + CC":
  - New tmux session created with name `moai-{ProjectName}-{SPEC-ID}`
  - Session changes to worktree directory
  - `/moai run {SPEC-ID}` command is sent to session
  - No GLM environment variables are injected

### Scenario 1.2: Tmux Available with CG Mode

**Given** user has completed SPEC creation with `--worktree` flag
**And** tmux is available on the system
**And** active mode is CG (Claude Leader + GLM Workers)
**When** user selects "Worktree + CG"
**Then** system:
  - Creates new tmux session
  - Injects GLM environment variables to session
  - Clears GLM env from settings.local.json (Leader isolation)
  - Changes to worktree directory
  - Executes `/moai run {SPEC-ID}`

### Scenario 1.3: Tmux Not Available

**Given** user has completed SPEC creation with `--worktree` flag
**And** tmux is NOT available on the system
**When** user selects "Start Implementation"
**Then** system presents two options:
  - "Sub-agent Mode (Recommended)"
  - "Team Mode (in-process)"
**And** displays manual navigation instructions:
```
Worktree created at: ~/.moai/worktrees/{ProjectName}/{SPEC-ID}/
Run: cd ~/.moai/worktrees/{ProjectName}/{SPEC-ID}/
Then: /moai run {SPEC-ID}
```

---

## AC2: Worktree Creation After SPEC Commit

### Scenario 2.1: Successful Worktree Creation

**Given** SPEC document is created with `--worktree` flag
**And** SPEC files are committed successfully
**When** worktree creation is triggered
**Then** system:
  - Creates worktree at `~/.moai/worktrees/{ProjectName}/{SPEC-ID}/`
  - Creates feature branch `feature/{SPEC-ID}`
  - Displays success message with path

### Scenario 2.2: Worktree Creation Failure

**Given** SPEC document is created with `--worktree` flag
**When** worktree creation fails
**Then** system:
  - Displays error message with cause
  - Provides manual creation command
  - Does NOT block SPEC creation (SPEC is already committed)

---

## AC3: Auto-Merge Default Behavior

### Scenario 3.1: Auto-Merge Triggered (Worktree Context)

**Given** implementation is complete in worktree
**And** user runs `/moai sync {SPEC-ID}` without flags
**And** worktree context is detected
**And** all CI/CD checks pass
**And** PR has zero merge conflicts
**When** sync phase reaches delivery step
**Then** system:
  - Automatically executes PR merge
  - Does NOT require `--merge` flag

### Scenario 3.2: Auto-Merge Skipped with --no-merge

**Given** implementation is complete in worktree
**When** user runs `/moai sync {SPEC-ID} --no-merge`
**Then** system:
  - Creates PR without auto-merge
  - Displays PR URL for manual review

### Scenario 3.3: Auto-Merge Blocked (CI Failure)

**Given** implementation is complete
**And** CI/CD checks have failed
**When** sync phase reaches auto-merge step
**Then** system:
  - Does NOT merge
  - Displays error: "Auto-merge blocked: CI checks failed"
  - Provides recovery command: `/moai sync {SPEC-ID}`

### Scenario 3.4: Auto-Merge Blocked (Conflicts)

**Given** implementation is complete
**And** PR has merge conflicts
**When** sync phase reaches auto-merge step
**Then** system:
  - Does NOT merge
  - Displays error: "Auto-merge blocked: Merge conflicts detected"
  - Instructs manual resolution

---

## AC4: Automatic Worktree Cleanup

### Scenario 4.1: Successful Cleanup After Merge

**Given** PR merge completed successfully
**And** `auto_cleanup` is enabled in config
**When** cleanup is triggered
**Then** system:
  - Removes worktree directory
  - Removes feature branch (if --delete-branch was set)
  - Updates worktree registry
  - Displays cleanup confirmation

### Scenario 4.2: Cleanup Failure Does Not Block Merge

**Given** PR merge completed successfully
**And** cleanup fails (e.g., permission error)
**When** cleanup is attempted
**Then** system:
  - Logs warning message
  - Provides manual cleanup command
  - Does NOT affect PR merge completion status

---

## AC5: Tmux Integration

### Scenario 5.1: Session Naming Convention

**Given** project name is "moai-adk"
**And** SPEC-ID is "SPEC-AUTH-001"
**When** tmux session is created
**Then** session name is `moai-moai-adk-SPEC-AUTH-001`

### Scenario 5.2: Session Isolation (CG Mode)

**Given** active mode is CG
**When** tmux session is created for worktree
**Then** system:
  - Injects GLM env vars to tmux session only
  - Clears GLM env vars from settings.local.json
  - Leader (main session) has no GLM env
  - Worker (worktree session) has GLM env

### Scenario 5.3: Graceful Degradation Without Tmux

**Given** tmux is not installed
**When** worktree is created
**Then** system:
  - Creates worktree successfully
  - Displays clear manual instructions
  - Does NOT error or block workflow

---

## AC6: Error Handling

### Scenario 6.1: All Error Messages Include Recovery

**Given** any error occurs during automation
**When** error is displayed to user
**Then** error message includes:
  - Error description
  - At least one recovery command
  - Reference to SPEC-ID

### Scenario 6.2: Partial Failure Recovery

**Given** worktree creation succeeded
**And** tmux session creation failed
**When** error is handled
**Then** system:
  - Preserves created worktree
  - Displays manual navigation instructions
  - Allows user to continue manually

---

## Quality Gates

### TRUST 5 Compliance

| Principle | Criteria | Verification |
|-----------|----------|--------------|
| Tested | 85%+ coverage on new code | `go test -cover ./internal/cli/worktree/...` |
| Readable | Clear naming, English comments | Code review |
| Unified | Consistent style with existing code | `golangci-lint run` |
| Secured | No secrets in env injection | Security review |
| Trackable | SPEC reference in all commits | Commit message check |

### Performance Criteria

| Metric | Target |
|--------|--------|
| Worktree creation | < 2 seconds |
| Tmux session creation | < 1 second |
| Worktree context detection | < 100ms |

---

## Edge Cases

### EC1: Multiple Worktrees for Same Project

**Given** project already has active worktree for SPEC-A
**When** creating worktree for SPEC-B
**Then** both worktrees coexist without conflict
**And** each has separate tmux session

### EC2: Worktree Already Exists

**Given** worktree for SPEC-A already exists
**When** user runs `/moai plan "new feature" --worktree` with same SPEC-ID
**Then** system:
  - Detects existing worktree
  - Displays warning
  - Offers options: Use existing / Create new / Abort

### EC3: Branch Already Exists

**Given** feature/SPEC-AUTH-001 branch already exists
**When** creating worktree for SPEC-AUTH-001
**Then** system:
  - Uses existing branch
  - Does NOT create duplicate branch

### EC4: Network Failure During PR Operations

**Given** network is unavailable
**When** auto-merge is attempted
**Then** system:
  - Displays clear network error
  - Preserves local state
  - Provides retry command

---

## Test Checklist

### Unit Tests

- [ ] `TestExecutionModeSelection_TmuxAvailable`
- [ ] `TestExecutionModeSelection_TmuxNotAvailable`
- [ ] `TestAutoMergeTrigger_WorktreeContext`
- [ ] `TestAutoMergeTrigger_NoMergeFlag`
- [ ] `TestWorktreeCleanup_Success`
- [ ] `TestWorktreeCleanup_Failure`
- [ ] `TestTmuxSessionNaming`
- [ ] `TestEnvInjection_CGMode`
- [ ] `TestEnvInjection_CCMode`

### Integration Tests

- [ ] `TestE2E_PlanRunSync_WithWorktree`
- [ ] `TestE2E_AutoMerge_Success`
- [ ] `TestE2E_AutoMerge_CIFailure`
- [ ] `TestE2E_AutoMerge_Conflicts`
- [ ] `TestE2E_Cleanup_AfterMerge`

### Manual Tests

- [ ] Full E2E: `/moai plan "test" --worktree` → run → sync
- [ ] Tmux session verification
- [ ] Manual cleanup verification
- [ ] Error recovery verification

---

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] Unit test coverage >= 85%
- [ ] Integration tests pass
- [ ] golangci-lint passes with zero errors
- [ ] Documentation updated (plan.md, sync.md)
- [ ] SPEC document marked as completed
- [ ] PR merged to main branch
