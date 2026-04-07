# SPEC-HOOKWAVE-001 Acceptance Criteria

## AC-HOOKWAVE-001: SubagentStop handler registered
Given moai binary is built
When `moai hook subagent-stop` is called with valid stdin JSON
Then handler logs agent_id and agent_name
And exits with code 0

## AC-HOOKWAVE-002: Missing EventTypes added
Given internal/hook/types.go
When types are checked
Then TaskCreated, PermissionDenied, ConfigChange, CwdChanged, FileChanged, Elicitation, ElicitationResult EventType constants exist

## AC-HOOKWAVE-003: 5 new handlers registered
Given deps.go handler registration
When all handlers are registered
Then SubagentStop, TaskCreated, PermissionDenied, ConfigChange, CwdChanged handlers are present

## AC-HOOKWAVE-004: Worktree registry
Given agent creates a worktree
When WorktreeCreate fires
Then `.moai/state/worktrees.json` contains entry with branch, agent_name, created_at
And when WorktreeRemove fires, the entry is removed

## AC-HOOKWAVE-005: UserPromptSubmit context
Given user types "loop" in a prompt
When UserPromptSubmit fires
Then additionalContext contains workflow context hint

## AC-HOOKWAVE-006: Unit tests pass
Given all new/modified handlers
When `go test ./internal/hook/...` runs
Then all tests pass with >= 85% coverage
