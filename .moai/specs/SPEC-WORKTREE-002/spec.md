# SPEC-WORKTREE-002: Worktree End-to-End Automation

---
id: SPEC-WORKTREE-002
version: "1.0.0"
status: completed
created: "2026-03-11"
updated: "2026-03-11"
author: MoAI
priority: high
lifecycle_level: spec-anchored
---

## HISTORY

| Date | Version | Author | Changes |
|------|---------|--------|---------|
| 2026-03-11 | 1.0.0 | MoAI | Initial SPEC creation |

---

## Environment

**Target System**: MoAI-ADK Go CLI (`moai` binary)

**Technology Stack**:
- Go 1.26+
- Git (system binary via exec)
- tmux (optional, for session isolation)

**Affected Components**:
- `internal/cli/worktree/` - Worktree CLI commands
- `internal/tmux/` - Tmux session management
- `.claude/skills/moai/workflows/plan.md` - Plan workflow (Decision Point 3.5)
- `.claude/skills/moai/workflows/sync.md` - Sync workflow (auto-merge logic)
- `internal/core/git/worktree.go` - Worktree manager interface

**Integration Points**:
- `/moai plan --worktree` - Worktree creation after SPEC commit
- `/moai run SPEC-XXX` - Implementation in worktree environment
- `/moai sync SPEC-XXX` - Auto-merge and cleanup
- `moai worktree done SPEC-XXX` - Worktree completion

---

## Assumptions

### Technical Assumptions

| Assumption | Confidence | Validation Method |
|------------|------------|-------------------|
| Git worktree operations remain stable | High | Git 2.30+ API stability |
| tmux session-scoped env isolation works | High | Verified in CG Mode implementation |
| Claude Code hooks support WorktreeCreate/Remove | High | Already implemented in hook registry |

### Business Assumptions

| Assumption | Confidence | Validation Method |
|------------|------------|-------------------|
| Users prefer automation over manual steps | High | User feedback from existing workflows |
| Auto-merge is desired default behavior | Medium | Configurable via --no-merge flag |
| tmux is optional but preferred for isolation | High | Fallback to manual cd instructions |

---

## Requirements

### R1: Execution Mode Selection Gate

**EARS Pattern**: Event-Driven

**WHEN** user completes SPEC creation in `/moai plan` phase
**THEN** system shall automatically detect and present execution mode options

**Requirements**:

1.1 The system SHALL detect active LLM mode from `.moai/config/sections/llm.yaml` (cc, glm, cg)

1.2 The system SHALL detect tmux availability via `$TMUX` environment variable

1.3 **WHEN** tmux is available AND user selects "Start Implementation"
**THEN** system SHALL present three options:
- Worktree + {active_mode} (Recommended)
- Team Mode (in-process)
- Sub-agent Mode (sequential)

1.4 **WHEN** tmux is NOT available
**THEN** system SHALL present two options:
- Sub-agent Mode (Recommended)
- Team Mode (in-process)

1.5 **WHEN** user selects Worktree mode
**THEN** system SHALL:
- Create new tmux session with name `moai-{ProjectName}-{SPEC-ID}`
- Change directory to worktree path
- Execute `/moai run SPEC-XXX` in new session
- Display navigation instructions

### R2: Worktree Creation After SPEC Commit

**EARS Pattern**: Event-Driven

**WHEN** SPEC document is created with `--worktree` flag
**THEN** system shall automatically create worktree after SPEC commit

**Requirements**:

2.1 The system SHALL commit SPEC files before worktree creation

2.2 The system SHALL create worktree at path `~/.moai/worktrees/{ProjectName}/{SPEC-ID}/`

2.3 The system SHALL create feature branch with naming convention `feature/{SPEC-ID}`

2.4 **IF** worktree creation fails
**THEN** system SHALL display error and provide manual creation command

### R3: Auto-Merge Default Behavior

**EARS Pattern**: State-Driven

**WHILE** in worktree-based development flow
**IF** user executes `/moai sync SPEC-XXX` without flags
**THEN** system shall default to auto-merge behavior

**Requirements**:

3.1 The system SHALL treat auto-merge as DEFAULT for worktree flows

3.2 The system SHALL provide `--no-merge` flag to skip auto-merge

3.3 Auto-merge SHALL only execute when:
- All CI/CD checks pass
- PR has zero merge conflicts
- Minimum reviewer approvals obtained (Team mode)

3.4 **IF** auto-merge fails
**THEN** system SHALL report failure reason and NOT merge

### R4: Automatic Worktree Cleanup

**EARS Pattern**: Event-Driven

**WHEN** PR merge completes successfully
**THEN** system shall automatically execute cleanup

**Requirements**:

4.1 The system SHALL run `moai worktree done SPEC-XXX` after successful PR merge

4.2 Cleanup SHALL remove:
- Worktree directory
- Feature branch (if --delete-branch was specified)
- Registry entry from `.moai/worktrees/registry.json`

4.3 **IF** cleanup fails
**THEN** system SHALL log warning and provide manual cleanup commands

4.4 Cleanup failure SHALL NOT block PR merge completion

### R5: Tmux Integration

**EARS Pattern**: Optional

**WHERE** tmux is available
**THEN** system shall provide seamless session isolation

**Requirements**:

5.1 The system SHALL create named tmux session: `moai-{ProjectName}-{SPEC-ID}`

5.2 **IF** active mode is GLM or CG
**THEN** system SHALL inject GLM environment variables into tmux session

5.3 **IF** active mode is CC
**THEN** system SHALL NOT inject GLM environment variables

5.4 Session creation SHALL include:
- Change directory to worktree
- Execute `/moai run SPEC-XXX`
- Return focus to main session

5.5 **IF** tmux is NOT available
**THEN** system SHALL display manual cd instructions:
```
Worktree created at: ~/.moai/worktrees/{ProjectName}/{SPEC-ID}/
Run: cd ~/.moai/worktrees/{ProjectName}/{SPEC-ID}/
Then: /moai run SPEC-XXX
```

### R6: Error Handling and Recovery

**EARS Pattern**: Unwanted Behavior

**The system shall NOT** leave worktrees in inconsistent state

**Requirements**:

6.1 **IF** worktree creation fails after SPEC commit
**THEN** system SHALL display recovery commands

6.2 **IF** tmux session creation fails
**THEN** system SHALL proceed without tmux and display manual instructions

6.3 **IF** auto-merge fails
**THEN** system SHALL preserve worktree for manual intervention

6.4 All error messages SHALL include:
- Error description
- Recovery command(s)
- Reference to SPEC-ID

---

## Specifications

### S1: Command Flow

```
/moai plan "feature" --worktree
  |
  v
Phase 1: SPEC Document Created
  |
  v
Phase 2: SPEC Committed
  |  git add .moai/specs/SPEC-XXX/
  |  git commit -m "feat(spec): Add SPEC-XXX - feature"
  |
  v
Phase 3: Worktree Created
  |  moai worktree new SPEC-XXX
  |  Path: ~/.moai/worktrees/{ProjectName}/SPEC-XXX/
  |
  v
Phase 3.5: Execution Mode Selection Gate
  |  Detect: active_mode (cc/glm/cg), tmux_available
  |  Present options via AskUserQuestion
  |
  v
[User selects Worktree mode]
  |
  v
Phase 3.6: Tmux Session Setup (if available)
  |  tmux new-session -d -s moai-{ProjectName}-{SPEC-ID}
  |  tmux send-keys "cd {worktree-path}" Enter
  |  tmux send-keys "/moai run SPEC-XXX" Enter
  |  [GLM/CG modes only] injectTmuxSessionEnv()
  |
  v
User executes: /moai run SPEC-XXX
  |
  v
Implementation completes
  |
  v
User executes: /moai sync SPEC-XXX
  |  [Default: auto-merge enabled]
  |  [Optional: --no-merge to skip]
  |
  v
PR Created and Merged
  |
  v
Phase 4: Auto-Cleanup
  |  moai worktree done SPEC-XXX --delete-branch
  |
  v
Complete
```

### S2: Configuration Schema

New configuration in `.moai/config/sections/workflow.yaml`:

```yaml
workflow:
  worktree:
    auto_create: true          # Auto-create worktree after SPEC commit
    auto_merge: true           # Default auto-merge in sync phase
    auto_cleanup: true         # Auto-cleanup after PR merge
    tmux_preferred: true       # Prefer tmux for session isolation
    session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"
```

### S3: CLI Flag Changes

| Command | Current Behavior | New Behavior |
|---------|-----------------|--------------|
| `/moai sync SPEC-XXX` | No auto-merge | Auto-merge by default (worktree flow) |
| `/moai sync SPEC-XXX --merge` | Auto-merge | Deprecated (auto-merge is default) |
| `/moai sync SPEC-XXX --no-merge` | N/A | Skip auto-merge explicitly |
| `/moai plan --worktree` | Create worktree only | Create worktree + tmux session |

### S4: Error Messages

| Scenario | Error Message |
|----------|--------------|
| Worktree creation failed | "Worktree creation failed: {error}. Manual: `moai worktree new {SPEC-ID}`" |
| Tmux not available | "tmux not available. Run: `cd {path}` then `/moai run {SPEC-ID}`" |
| Auto-merge blocked (CI failed) | "Auto-merge blocked: CI checks failed. Fix issues and re-run: `/moai sync {SPEC-ID}`" |
| Auto-merge blocked (conflicts) | "Auto-merge blocked: Merge conflicts detected. Resolve manually in PR." |
| Cleanup failed | "Worktree cleanup failed: {error}. Manual: `moai worktree done {SPEC-ID}`" |

---

## Traceability

### TAG Block

```markdown
<!-- TAG: SPEC-WORKTREE-002 -->
<!-- @MX:SPEC: SPEC-WORKTREE-002 -->
<!-- @MX:PRIORITY: high -->
```

### Related SPECs

- SPEC-WORKTREE-001: Basic worktree management (implemented)
- SPEC-HOOK-007: Worktree lifecycle hooks
- SPEC-GIT-001: Git domain operations

### Files to Modify

| File | Change Type |
|------|-------------|
| `.claude/skills/moai/workflows/plan.md` | Modify - Add Execution Mode Selection Gate |
| `.claude/skills/moai/workflows/sync.md` | Modify - Change auto-merge default for worktree |
| `internal/cli/worktree/new.go` | Modify - Add tmux integration |
| `internal/cli/worktree/done.go` | Modify - Add auto-cleanup trigger |
| `internal/tmux/session.go` | No change - Already supports session naming |
| `.moai/config/sections/workflow.yaml` | New - Worktree automation settings |

---

## Constraints

### Technical Constraints

- C1: Must maintain backward compatibility with existing worktree commands
- C2: Must not require tmux (graceful degradation required)
- C3: Must work with all three LLM modes (CC, GLM, CG)

### Business Constraints

- C4: Auto-merge must respect Team mode approval requirements
- C5: Cleanup must not delete uncommitted work without confirmation

---

## Out of Scope

- Multi-SPEC parallel development (future enhancement)
- Worktree sync across machines
- Custom tmux layout configuration
- Worktree template presets
