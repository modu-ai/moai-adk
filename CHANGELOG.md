# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- SecurityScanner and LSP diagnostics integration in hook system
- Local release folder update system for streamlined development
- Branch-specific release detection for version management

### Changed
- Module renamed from `moai-adk-go` to `moai-adk` with lint fixes

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

### Removed
- Unused `err` field from `confirmModel` struct in merge confirmation UI

---

## Release History

For previous releases, see [GitHub Releases](https://github.com/modu-ai/moai-adk/releases).
