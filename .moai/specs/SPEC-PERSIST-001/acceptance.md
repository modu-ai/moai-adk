# SPEC-PERSIST-001 Acceptance Criteria

## AC-PERSIST-001: Block stop during active workflow
Given persistent-mode is active with workflow="loop"
And stop_hook_active is false
When Stop event fires without completion marker
Then HookOutput has Decision="block" and Reason contains workflow name

## AC-PERSIST-002: Allow stop on completion marker
Given persistent-mode is active
When Stop event fires with `<moai>DONE</moai>` in last_assistant_message
Then persistent-mode is deactivated (file deleted or active=false)
And HookOutput allows stop (empty Decision)

## AC-PERSIST-003: Prevent infinite loop
Given persistent-mode is active
And stop_hook_active is true
When Stop event fires
Then HookOutput allows stop (empty Decision)
And persistent-mode remains unchanged

## AC-PERSIST-004: Max duration timeout
Given persistent-mode activated 61 minutes ago with max_duration_minutes=60
When Stop event fires
Then persistent-mode is deactivated
And HookOutput allows stop

## AC-PERSIST-005: Lifecycle API
Given no persistent-mode file exists
When Activate("loop", "SPEC-AUTH-001", 60) is called
Then `.moai/state/persistent-mode.json` is created with active=true
And when Deactivate() is called, active becomes false

## AC-PERSIST-006: Unit tests pass
Given all new code in internal/hook/lifecycle/
When `go test ./internal/hook/lifecycle/...` runs
Then all tests pass with >= 85% coverage
And `go test ./internal/hook/...` (parent) also passes
