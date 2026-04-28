## Task Decomposition
SPEC: SPEC-STATUS-AUTO-001

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | YAML frontmatter status update (Format A/B) | REQ-1 | - | internal/spec/status.go, internal/spec/status_test.go | pending |
| T-002 | Markdown list status update (Format D) | REQ-1 | T-001 | internal/spec/status.go | pending |
| T-003 | Table status update (Format E, KR/EN) | REQ-1 | T-001 | internal/spec/status.go | pending |
| T-004 | YAML frontmatter prepend (Format F) | REQ-1 | T-001 | internal/spec/status.go | pending |
| T-005 | ParseStatus + validation | REQ-1 | T-001 | internal/spec/status.go | pending |
| T-006 | CLI: moai spec status + --dry-run + --list | REQ-2 | T-005 | internal/cli/spec.go, internal/cli/spec_status.go, internal/cli/root.go | pending |
| T-007 | PostToolUse hook handler + wrapper script | REQ-3 | T-005 | internal/hook/spec_status.go, .claude/hooks/moai/handle-spec-status.sh | pending |
| T-008 | Sync workflow integration | REQ-4 | T-006 | .claude/skills/moai/workflows/sync.md | pending |
| T-009 | Batch --sync-git command | REQ-5 | T-005 | internal/cli/spec_status.go | pending |
