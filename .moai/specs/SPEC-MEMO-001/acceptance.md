# SPEC-MEMO-001 Acceptance Criteria

## AC-MEMO-001: PreCompact writes session-memo
Given a session with active SPEC-ID "SPEC-AUTH-001" in loop workflow
When PreCompact event fires
Then `.moai/state/session-memo.md` is created with P1 content (SPEC-ID, workflow phase, execution mode)
And file size does not exceed 2,200 tokens equivalent

## AC-MEMO-002: PostCompact restores via systemMessage
Given `.moai/state/session-memo.md` exists with valid content
When PostCompact event fires
Then HookOutput.SystemMessage contains the memo content
And systemMessage is non-empty

## AC-MEMO-003: Token budget enforcement
Given session-memo content exceeds 2,200 tokens
When memo/writer.go trims content
Then P4 content is removed first, then P3, then P2
And P1 content is always preserved

## AC-MEMO-004: Empty memo graceful handling
Given `.moai/state/session-memo.md` does not exist
When PostCompact event fires
Then HookOutput is empty (no error, no systemMessage)

## AC-MEMO-005: Persistent mode integration
Given persistent-mode is active (SPEC-PERSIST-001)
When PreCompact writes session-memo
Then memo includes persistent-mode status, workflow name, and spec_id

## AC-MEMO-006: Unit tests pass
Given all new code in internal/hook/memo/ package
When `go test ./internal/hook/memo/...` runs
Then all tests pass with >= 85% coverage
And `go test ./internal/hook/...` (parent) also passes
