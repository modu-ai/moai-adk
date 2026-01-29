# Changelog

All notable changes to MoAI-ADK Go implementation will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial Go implementation of MoAI-ADK
- Cross-platform binary distribution (macOS, Linux, Windows)
- goreleaser configuration for automated releases
- GitHub Actions release pipeline
- Install script with platform auto-detection
- Self-update command (`moai-adk self-update`)
- Migration command (`moai-adk migrate`)
- Dual-mode operation (Python + Go coexistence)
- Compatibility verification tool
- Hook execution logging
- Homebrew formula auto-generation

### Changed
- Performance: 10-50x faster than Python implementation
- Binary size: Static binary < 15MB per platform
- Memory usage: 5x reduction compared to Python

### Fixed
- N/A (initial release)

## [1.0.0] - 2026-01-29

### Added
- Core Hooks implementation:
  - `session_start` - Display project info on session start
  - `session_end` - Display session summary
  - `pre_tool_use` - Security checks before tool use
  - `post_tool_use` - Linting after tool use
  - `pre_compact` - Prepare for context compaction
  - `notification` - Handle notifications

- Project Management:
  - `moai-adk init` - Initialize new projects
  - `moai-adk update` - Update templates and configuration
  - `moai-adk doctor` - Check project health
  - `moai-adk status` - Show project status

- CLI Commands:
  - `moai-adk hook` - Execute hooks
  - `moai-adk statusline` - Display formatted statusline
  - `moai-adk version` - Show version information

- Distribution:
  - Cross-compilation for 5 platforms (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64)
  - goreleaser configuration with checksums and release notes
  - GitHub Actions automated release pipeline
  - Homebrew formula auto-generation
  - Install script with platform auto-detection

- Migration:
  - `moai-adk migrate` - Convert Python hooks to Go hooks
  - `moai-adk migrate --dry-run` - Preview migration changes
  - `moai-adk migrate --rollback` - Rollback migration
  - Dual-mode detection (Go binary found → use Go, else Python)
  - Manual override flags (`--force-go`, `--force-python`)
  - Hook execution logging (`.moai/logs/hook-implementation.log`)

- Self-Update:
  - `moai-adk self-update` - Update to latest version
  - `moai-adk self-update --check-only` - Check for updates without downloading
  - `moai-adk self-update --version <version>` - Update to specific version

- Compatibility:
  - `cmd/verify-compatibility` tool for Python vs Go comparison
  - JSON schema compatibility verification
  - Exit code compatibility verification
  - Error message compatibility verification

- Documentation:
  - README with installation instructions
  - Migration guide for Python users
  - CHANGELOG with release notes

### Performance
- Cold start: ~50ms (vs ~500ms Python)
- Hook execution: ~20ms (vs ~200ms Python)
- Project init: ~200ms (vs ~2s Python)
- Memory usage: ~20MB (vs ~100MB Python)

### Platform Support
- macOS Intel (amd64): ~15MB binary
- macOS Apple Silicon (arm64): ~14MB binary
- Linux AMD64: ~13MB binary
- Linux ARM64: ~13MB binary
- Windows AMD64: ~15MB binary

### Dependencies
- github.com/spf13/cobra v1.10.2
- github.com/charmbracelet/lipgloss v1.1.0
- gopkg.in/yaml.v3 v3.0.1

### Python Deprecation Timeline
- Go 1.0 release: Python enters maintenance-only mode
- 6 months after Go 1.0: Python deprecated
- 12 months after Go 1.0: Python support removed

---

**Release Date**: 2026-01-29
**Go Version**: 1.23+
**License**: Apache-2.0
