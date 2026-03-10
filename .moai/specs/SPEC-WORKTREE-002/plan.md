# Implementation Plan: SPEC-WORKTREE-002

## Overview

Implement end-to-end automation for MoAI worktree workflow, connecting plan → run → sync phases with automatic worktree creation, tmux session setup, auto-merge, and cleanup.

---

## Phase 1: Execution Mode Selection Gate (Priority: High)

### 1.1 Plan Workflow Modification

**File**: `.claude/skills/moai/workflows/plan.md`

**Changes**:
- Enhance Decision Point 3.5 with complete automation flow
- Add Step 4 (Worktree Selection Execution) implementation details
- Add Step 5 (Gate Result Passing) to run workflow

**Implementation**:
1. Read `.moai/config/sections/llm.yaml` to detect `llm.team_mode`
2. Execute `test -n "$TMUX"` to check tmux availability
3. Present options via AskUserQuestion with mode-aware descriptions
4. On worktree selection:
   - CC: Create worktree, create tmux session, run claude
   - GLM: Create worktree, inject GLM env to tmux, run claude
   - CG: Create worktree, inject GLM env to tmux, clear GLM from settings.local.json, run claude

### 1.2 Tmux Session Naming

**Convention**: `moai-{ProjectName}-{SPEC-ID}`

**Implementation**:
- Use `detectProjectName()` from `internal/cli/worktree/project.go`
- Combine with SPEC-ID for session name
- Example: `moai-moai-adk-SPEC-AUTH-001`

---

## Phase 2: Auto-Merge Default Behavior (Priority: High)

### 2.1 Sync Workflow Modification

**File**: `.claude/skills/moai/workflows/sync.md`

**Changes**:
- Modify Phase 3 Step 3.4 (Auto-Merge)
- Change execution condition from `--merge flag` to `worktree context detected`
- Add `--no-merge` flag documentation

**Auto-merge Trigger Conditions**:
```
IF (worktree context detected) AND (--no-merge NOT set) AND (all checks pass)
THEN execute auto-merge
```

**Worktree Context Detection**:
- Check if current git directory path contains `.moai/worktrees/`
- Or check `.moai/worktrees/registry.json` for active worktree entry

### 2.2 Deprecate --merge Flag

**Behavior**:
- `--merge` flag becomes no-op (auto-merge is default)
- Log deprecation warning when `--merge` is used
- Document migration in help text

---

## Phase 3: Automatic Cleanup (Priority: High)

### 3.1 Post-Merge Cleanup Hook

**File**: `.claude/skills/moai/workflows/sync.md` (Phase 3 Step 3.4 extension)

**Implementation**:
After successful `gh pr merge`:
1. Detect if worktree exists for SPEC-ID
2. Execute `moai worktree done SPEC-ID --delete-branch` equivalent
3. Log cleanup result

### 3.2 Cleanup Failure Handling

**Requirements**:
- Do NOT block PR merge completion
- Log warning with manual cleanup commands
- Update sync report with cleanup status

---

## Phase 4: Tmux Integration Enhancement (Priority: Medium)

### 4.1 Worktree Creation with Tmux

**File**: `internal/cli/worktree/new.go`

**Changes**:
- Add `--tmux` flag to create tmux session after worktree creation
- Add `--session-name` flag for custom session naming
- On creation with `--tmux`:
  1. Create worktree (existing logic)
  2. Create tmux session
  3. Inject env vars based on active mode
  4. Send cd + run command

### 4.2 Environment Variable Injection

**Reference**: `internal/tmux/session.go` - `InjectEnv` method

**Active Mode Detection**:
```go
// Read llm.team_mode from config
mode := config.GetString("llm.team_mode")
switch mode {
case "glm":
    // Inject all GLM env vars
case "cg":
    // Inject GLM env vars for workers
    // Keep Leader clean (no GLM in settings.local.json)
case "", "cc":
    // No GLM env injection
}
```

---

## Phase 5: Configuration Schema (Priority: Medium)

### 5.1 New Workflow Configuration

**File**: `.moai/config/sections/workflow.yaml`

**Add Section**:
```yaml
workflow:
  worktree:
    auto_create: true          # Auto-create worktree after SPEC commit
    auto_merge: true           # Default auto-merge in sync phase
    auto_cleanup: true         # Auto-cleanup after PR merge
    tmux_preferred: true       # Prefer tmux for session isolation
    session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"
```

### 5.2 Default Values

**In `internal/config/types.go`**:
- Add WorktreeConfig struct
- Set defaults in `defaults.go`

---

## Phase 6: Error Handling (Priority: High)

### 6.1 Error Message Templates

**Create**: `internal/cli/worktree/errors.go` (if not exists)

**Templates**:
```go
const (
    ErrWorktreeCreateFailed = "Worktree creation failed: %s. Manual: `moai worktree new %s`"
    ErrTmuxNotAvailable     = "tmux not available. Run: `cd %s` then `/moai run %s`"
    ErrAutoMergeCIFailed    = "Auto-merge blocked: CI checks failed. Fix issues and re-run: `/moai sync %s`"
    ErrAutoMergeConflicts   = "Auto-merge blocked: Merge conflicts detected. Resolve manually in PR."
    ErrCleanupFailed        = "Worktree cleanup failed: %s. Manual: `moai worktree done %s`"
)
```

### 6.2 Recovery Commands

**Display Pattern**:
```
[ERROR] {error description}

Recovery commands:
  1. {step 1 command}
  2. {step 2 command}

Reference: SPEC-{ID}
```

---

## Risk Analysis

### Risk 1: Tmux Session Orphaning

**Scenario**: User closes terminal without cleanup
**Probability**: Medium
**Impact**: Low (worktree persists)
**Mitigation**: Document `moai worktree list` for recovery

### Risk 2: Auto-Merge Conflicts

**Scenario**: Conflicts detected during auto-merge
**Probability**: Medium
**Impact**: Medium (blocks automation)
**Mitigation**: Clear error message + manual resolution guidance

### Risk 3: Environment Variable Leakage

**Scenario**: GLM env vars leak to CC mode session
**Probability**: Low
**Impact**: High (cost implications)
**Mitigation**: Strict mode detection + env isolation in CG mode

---

## Testing Strategy

### Unit Tests

| Test | File | Coverage Target |
|------|------|-----------------|
| Session naming | `worktree/new_test.go` | 100% |
| Worktree context detection | `worktree/context_test.go` | 100% |
| Auto-merge trigger logic | `workflow/sync_test.go` | 100% |

### Integration Tests

| Test | Scenario |
|------|----------|
| E2E plan → run → sync | Full workflow with worktree |
| Tmux session creation | With and without tmux |
| Auto-merge with conflicts | Conflict detection |
| Cleanup after merge | Successful cleanup |

### Manual Tests

1. Run `/moai plan "test feature" --worktree`
2. Verify worktree creation at correct path
3. Verify tmux session with correct name
4. Run `/moai run SPEC-XXX`
5. Run `/moai sync SPEC-XXX`
6. Verify auto-merge behavior
7. Verify cleanup after merge

---

## Dependencies

### Internal Dependencies

- `internal/core/git/worktree.go` - WorktreeManager interface
- `internal/tmux/session.go` - SessionManager interface
- `internal/config/` - Configuration loading

### External Dependencies

- Git 2.30+ (for worktree operations)
- tmux 3.0+ (optional, for session isolation)
- GitHub CLI (`gh`) for PR operations

---

## Milestones

### Milestone 1: Core Automation (Priority High)
- [ ] Execution Mode Selection Gate implementation
- [ ] Auto-merge default behavior change
- [ ] Automatic cleanup after PR merge

### Milestone 2: Enhanced Integration (Priority Medium)
- [ ] Tmux session creation with worktree
- [ ] Configuration schema addition
- [ ] Error message templates

### Milestone 3: Polish (Priority Low)
- [ ] Deprecation warnings for old flags
- [ ] Documentation updates
- [ ] Integration test coverage

---

## Reference Implementations

### Existing Code Patterns

**Tmux Session Creation**: `internal/cli/cg.go`
```go
// Reference: injectTmuxSessionEnv() for GLM mode
func injectTmuxSessionEnv(vars map[string]string) error
```

**Worktree Detection**: `internal/cli/worktree/project.go`
```go
// Reference: detectProjectName() for session naming
func detectProjectName(dir string) string
```

**Auto-merge Logic**: `.claude/skills/moai/workflows/sync.md` Step 3.4
```
// Reference: Existing auto-merge execution conditions
```

---

## Next Steps

1. Review and approve this SPEC
2. Create feature branch: `feature/SPEC-WORKTREE-002`
3. Implement Phase 1 (Execution Mode Selection Gate)
4. Implement Phase 2 (Auto-Merge Default)
5. Implement Phase 3 (Automatic Cleanup)
6. Add integration tests
7. Update documentation
