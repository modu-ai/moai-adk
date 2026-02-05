# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- SecurityScanner and LSP diagnostics integration in hook system
- Local release folder update system for streamlined development
- Branch-specific release detection for version management
- Hook wrapper scripts for all 5 MoAI hook events (session-start, compact, pre-tool, post-tool, stop)

### Changed
- Module renamed from `moai-adk-go` to `moai-adk` with lint fixes
- **Hook Path Syntax**: Updated to use `"$CLAUDE_PROJECT_DIR/.claude/hooks/..."` with proper quoting for paths with spaces
- **StatusLine Configuration**: Changed to relative path `.moai/status_line.sh` (statusLine does not support `$CLAUDE_PROJECT_DIR` expansion per GitHub Issue #7925)
- **Hook Wrapper Deployment**: Hook wrappers are now deployed to `.claude/hooks/moai/` during initialization with automatic moai binary path detection

### Fixed
- **Hook JSON Output Schema**: Added missing `hookEventName` field in `hookSpecificOutput` for protocol compliance
  - Affected events: `PreToolUse`, `PostToolUse`, `SessionStart`, `PreCompact`
  - Ensures full compliance with Claude Code hook validation requirements
  - Resolves JSON schema validation errors in hook execution
- Hook JSON output corrected for `Stop` and `SessionEnd` events
- Template synchronization now works on dev builds and when Go binary is unavailable
- Browser no longer opens during automated tests
- Template sync properly executes after binary updates
- API URL updated to point to `modu-ai/moai-adk` repository
- **Missing Hook Scripts**: Deployed 5 hook wrapper scripts that were missing from local project, causing "No such file or directory" errors

### Removed
- Unused `err` field from `confirmModel` struct in merge confirmation UI
- Duplicate `.tmpl` files from local `.moai/config/sections/` directory (template sources belong in `internal/template/templates/` only)

---

## Release History

For previous releases, see [GitHub Releases](https://github.com/modu-ai/moai-adk/releases).
