# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2025-01-10

### Fixed
- Version bump for npm republish (24-hour restriction workaround)

## [0.1.0] - 2025-01-10

### Added
- 🎯 SPEC-First TDD workflow with 3-step development process
- 🤖 Alfred SuperAgent with 9 specialized agents ecosystem
- 🏷️ 4-Core @TAG system (SPEC → TEST → CODE → DOC)
- 🌍 Universal language support (TypeScript, Python, Java, Go, Rust, Dart, Swift, Kotlin)
- 📱 Mobile framework support (Flutter, React Native, iOS, Android)
- ⚡ Intelligent project diagnostics and environment validation
- 🔒 TRUST 5 principles (Test, Readable, Unified, Secured, Trackable)
- 📊 Living Document synchronization
- 🚀 GitFlow integration with automated branching and PR management
- 🛠️ Claude Code complete integration (Agents, Commands, Hooks, Output Styles)

### Features
- **CLI Commands**:
  - `moai init` - Initialize new MoAI-ADK project
  - `moai doctor` - System environment diagnostics
  - `moai status` - Project status check
  - `moai update` - Update MoAI-ADK templates
  - `moai restore` - Restore from backup

- **Alfred Commands**:
  - `/alfred:1-spec` - EARS specification creation
  - `/alfred:2-build` - TDD implementation (Red-Green-Refactor)
  - `/alfred:3-sync` - Document synchronization
  - `/alfred:8-project` - Project initialization
  - `/alfred:9-update` - Package and template updates

- **Code Quality**:
  - TypeScript strict mode enabled
  - Biome linter and formatter integration
  - Comprehensive test coverage with Vitest
  - Automated security scanning

- **Documentation**:
  - VitePress documentation site
  - TypeDoc API documentation
  - Comprehensive README and guides

### Technical Stack
- TypeScript 5.9.2+
- Node.js 18.0+
- Bun 1.2.19+ (recommended)
- Vitest for testing
- Biome for linting/formatting
- tsup for building

### Known Issues
- Some installation flow edge cases require validation improvements
- Test coverage at ~95% (target: 100%)
- Documentation still being enhanced

### Breaking Changes
- First release - no breaking changes

[0.1.0]: https://github.com/modu-ai/moai-adk/releases/tag/v0.1.0
