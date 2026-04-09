# Research: SPEC-CC297-001

## Claude Code 2.1.94~2.1.97 Feature Adoption

### Source Analysis

#### 1. UserPromptSubmit sessionTitle (v2.1.94)

**New Feature**: `hookSpecificOutput.sessionTitle` field in UserPromptSubmit hooks.

**Current State**:
- `HookSpecificOutput` struct (`internal/hook/types.go:247-252`) has 4 fields: HookEventName, PermissionDecision, PermissionDecisionReason, AdditionalContext
- `SessionTitle` field does NOT exist yet
- UserPromptSubmit handler (`internal/hook/user_prompt_submit.go`) only logs prompts, returns empty `HookOutput{}`
- SPEC-ID regex already exists in codebase: `regexp.MustCompile("SPEC-[A-Z]+-\\d+")`  (task_completed.go:15)
- Language config accessible via `ConfigProvider` interface
- Session state at `.moai/state/last-session-state.json`

**Architecture**:
```
Claude Code → handle-user-prompt-submit.sh → moai hook user-prompt-submit
  → HookRegistry.Dispatch() → UserPromptSubmitHandler.Handle()
  → HookOutput { HookSpecificOutput: { SessionTitle: "..." } }
  → Claude Code sets session title
```

#### 2. refreshInterval Status Line (v2.1.97)

**New Feature**: `refreshInterval` setting for periodic status line refresh.

**Current State**:
- settings.json template (`settings.json.tmpl:383-386`): Only `type` and `command` fields
- statusline.yaml config: No refreshInterval field
- StdinData struct (`internal/statusline/types.go:57-69`): No refresh-related fields

#### 3. workspace.git_worktree Status Line Input (v2.1.97)

**New Feature**: `workspace.git_worktree` in status line JSON input.

**Current State**:
- WorkspaceInfo struct (`types.go:121-125`): Only `CurrentDir` and `ProjectDir`
- `GitWorktree` field does NOT exist
- Git worktree infrastructure complete at `internal/core/git/worktree.go`
- 12 segment constants exist, no `SegmentWorktree`

#### 4. Output Style keep-coding-instructions (v2.1.94)

**Already Applied**: All 3 output styles (moai.md, r2d2.md, yoda.md) have `keep-coding-instructions: true`.
No work needed.

#### 5. Worktree CWD Leak Fix (v2.1.97)

**Bug Fix**: Subagents with `isolation: worktree` leaked CWD back to parent.
Document minimum version requirement.

#### 6. Stop/SubagentStop Long Session Fix (v2.1.97)

**Bug Fix**: Stop/SubagentStop hooks failed on long sessions.
Document minimum version requirement.

### Existing SPECs

- SPEC-HOOK-001~009: Hook system specifications
- SPEC-STATUSLINE-001~002: Statusline specifications

### Test Infrastructure

- internal/hook/: 43 test files
- internal/statusline/: 12 test files
- internal/template/: 9 test files
