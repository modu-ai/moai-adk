# Changelog

All notable changes to MoAI-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.11] - 2025-09-15 (CRITICAL HOTFIX)

### 🚨 Critical Bug Fixes
- **🛡️ CRITICAL: Fixed file deletion bug in `moai init .`**
  - `installer.py`: Modified `_create_project_directory()` to preserve existing files when initializing in current directory
  - **Issue**: `shutil.rmtree()` was unconditionally deleting ALL files in current directory
  - **Solution**: Added safe mode logic that preserves existing files and only creates MoAI directories
  - **Impact**: Prevents catastrophic data loss for users running `moai init .`

### ✅ Enhanced Safety Features
- **🔒 Added --force option with strong warnings**: Users must explicitly use `--force` to overwrite files
- **⚠️ Pre-installation warnings**: Clear messages about which files will be preserved
- **🛡️ Current directory protection**: Enhanced safety for current directory initialization
- **📋 File preservation confirmation**: User prompt showing exactly which files will be kept

### 🔧 Technical Improvements
- **config.py**: Added `force_overwrite` configuration flag
- **cli.py**: Enhanced init command with safety warnings and file preservation messages
- **installer.py**: Implemented intelligent directory handling based on context

### ⚡ Breaking Changes
- **NONE**: This hotfix is fully backward compatible while adding safety

### 🧪 Verified Fixes
- ✅ Current directory files are preserved during `moai init .`
- ✅ MoAI-ADK directories (.claude/, .moai/) are properly created
- ✅ Warning messages clearly inform users about file preservation
- ✅ --force option works as expected for explicit overwrite scenarios

## [0.1.10] - 2025-09-15

### 🚀 Enhanced Python Support & Documentation
- **🐍 Python 3.11+ Requirement**: Upgraded minimum Python version from 3.9 to 3.11+
- **🆕 Modern Python Features**: Enhanced templates to leverage Python 3.11+ features (match-case, exception groups, etc.)
- **📚 Comprehensive Memory System**: Improved documentation files in `.claude/memory/` and `.moai/memory/`
- **🏗️ Updated Architecture Standards**: Enhanced coding standards with Python 3.11+ best practices
- **📋 Refined Project Guidelines**: Updated 16-Core TAG system documentation
- **🤝 Enhanced Team Conventions**: Improved collaboration protocols and workflows

### 📖 Documentation Improvements
- **Constitution References**: Clear file path references to `@.claude/memory/` and `@.moai/memory/` files
- **TAG System Alignment**: Synchronized documentation with actual configuration
- **Workflow Optimization**: Updated CI/CD templates with latest security and performance practices

### 🔧 Template System Updates
- **Settings Optimization**: Streamlined `.claude/settings.json` permissions
- **Workflow Enhancement**: Updated GitHub Actions with Python 3.11+ compatibility
- **Configuration Refinement**: Improved MoAI config with enhanced indexing
- **🌐 ccusage Integration**: Added ccusage statusLine support for real-time Claude Code usage tracking
- **📊 Node.js Environment Check**: Automatic verification of Node.js/npm for ccusage compatibility

## [0.1.9] - 2025-09-15

### 🛡️ SECURITY - Removed Dangerous Installation Options

#### Removed
- **❌ Dangerous `--force` option**: Completely removed from all CLI commands
- **❌ Unsafe file overwriting**: No more destructive reinstallation

#### Added
- **🔒 Safe installation system**: Automatic conflict detection before installation
- **💾 Automatic backup system**: `--backup` option creates timestamped backups
- **🔍 Pre-installation checks**: Detects potential file conflicts and warns users
- **💬 Interactive confirmations**: User consent required for any changes
- **🏥 Recovery system**: New `moai doctor` and `moai restore` commands

#### New Commands
- `moai doctor`: Health check and backup listing
- `moai doctor --list-backups`: Show all available backups
- `moai restore <backup_path>`: Restore from backup
- `moai restore <backup_path> --dry-run`: Preview restoration

#### Safety Features
- **Git preservation**: Always preserves existing .git directories
- **Backup creation**: Automatic backup of .moai/, .claude/, and CLAUDE.md
- **Conflict warnings**: Lists potential file conflicts before proceeding
- **User confirmation**: Interactive prompts for all potentially destructive operations
- **Recovery info**: Detailed backup information with restoration instructions

#### Updated Installation Flow
```bash
# Safe installation with backup
moai init . --backup

# Interactive installation with safety checks
moai init . --interactive --backup

# Check installation health
moai doctor

# Restore from backup if needed
moai restore .moai_backup_20241215_143022
```

## [0.1.7] - 2025-09-12

### Added

- 🧠 **완전한 메모리 시스템**
  - `.claude/memory/` 디렉토리에 프로젝트 가이드라인, 코딩 표준, 팀 협업 규약 파일
  - `.moai/memory/` 디렉토리에 Constitution 헌법, 업데이트 체크리스트, ADR 템플릿
  - 메모리 파일 자동 설치 기능 (`_install_memory_files()`)

- 🐙 **GitHub CI/CD 시스템**
  - `moai-ci.yml`: Constitution 5원칙 자동 검증 파이프라인
  - `PULL_REQUEST_TEMPLATE.md`: MoAI Constitution 기반 PR 템플릿
  - 언어별 자동 감지 (Python, Node.js, Rust, Go)
  - 보안 스캔, 커버리지 검사, Constitution 검증 자동화

- 🚀 **지능형 Git 시스템**
  - 운영체제별 Git 자동 설치 제안 (Homebrew, APT, YUM, DNF)
  - 기존 .git 디렉토리 자동 보존 (--force 사용 시에도)
  - Git 상태별 적응형 메시지 (신규/기존/실패)
  - 포괄적 .gitignore 파일 자동 생성

- 🔀 **명령어 책임 분리**
  - `moai init`: MoAI-ADK 기본 시스템만 설치
  - `/moai:project init`: steering 문서 기반 프로젝트별 구조 생성
  - 명확한 설치 범위 구분 및 문서화

### Changed

- 📁 **스크립트 디렉토리 위치 수정**: `scripts/` → `.moai/scripts/`
- 🏗️ **설치 과정 확장**: 13단계 → 17단계 프로세스
- 📊 **진행률 표시 개선**: 상황별 동적 메시지 시스템
- 📋 **디렉토리 구조 정리**: 불필요한 docs, src, tests 디렉토리 생성 제거

### Fixed

- 🔧 Git 초기화 중복 실행 방지
- 🔧 --force 옵션 사용 시 .git 디렉토리 삭제 문제 해결
- 🔧 CLAUDE.md 파일 설치 누락 문제 해결
- 🔧 메모리 파일 설치 누락 문제 해결

### Enhanced

- ⚡ **에러 복구**: Git 설치 실패 시 graceful degradation
- 🎯 **사용자 경험**: Git 필요성 설명 및 설치 가이드 제공
- 🔒 **보안 강화**: 자동 시크릿 스캔 및 라이선스 검사
- 📖 **문서화 개선**: MoAI-ADK-Design-Final.md 대폭 업데이트

## [0.1.4] - 2025-09-01

### Fixed

- 🔧 Fixed hook file installation from .cjs templates to .js files
- 🔧 Updated hook command paths to use correct `.claude/hooks/` directory
- ✅ Resolved "Cannot find module" errors for hook files
- 📁 Fixed installer to copy template hooks properly

## [0.1.3] - 2025-09-01

### Fixed

- 🔧 Fixed hook matcher format to use string instead of object
- 🔧 Updated all settings.json files to use correct matcher syntax
- ✅ Resolved "Expected string, but received object" matcher errors
- 📚 Applied official Claude Code documentation format requirements

## [0.1.2] - 2025-09-01

### Fixed

- 🔧 Fixed installer to generate correct Claude Code settings.json format
- 🔧 Updated dynamic settings generation to use new hook matcher syntax
- ✅ Ensure all generated projects use compatible settings format
- 🗿 Fixed version consistency between CLI and installer

## [0.1.1] - 2025-09-01

### Fixed

- 🔧 Updated Claude Code settings.json format to use new hook matcher syntax
- 🔧 Fixed permissions format to use ":_" prefix matching instead of "_"
- ✅ Resolved compatibility issues with latest Claude Code version
- 📝 Updated all template files to use correct settings format

## [0.1.0] - 2025-09-01

### Added

- 🚀 Initial beta release of MoAI-ADK (MoAI Agentic Development Kit)
- 🤖 Complete Claude Code project initialization system
- 📋 16-Core TAG system for perfect traceability
- 🔧 Node.js native Hook system (pre-tool-use, post-tool-use, session-start)
- 🎯 AZENT methodology integration (SPEC → @TAG → TDD philosophy)
- 📊 Real-time document synchronization system
- 🔄 Hybrid TypeScript development + JavaScript deployment architecture

### Features

- **CLI Tool**: `moai-adk init` command for project initialization
- **Multiple Templates**: minimal, standard, enterprise project templates
- **Cross-Platform Support**: Windows, macOS, Linux compatibility
- **Zero Dependencies**: Hook system runs without compilation overhead
- **TypeScript Support**: Full type definitions and IDE integration
- **Auto-Updates**: Built-in `update` and `doctor` commands

### Technical Improvements

- ✅ Removed Bun dependency for better compatibility
- ✅ Node.js 18+ requirement with native module support
- ✅ ESM + CommonJS hybrid module system
- ✅ Optimized package size and distribution structure
- ✅ Complete TypeScript declaration files (.d.ts)

### Documentation

- 📖 Comprehensive README with usage examples
- 🔧 Complete API documentation for library usage
- 📋 Installation and setup guides
- 🚀 Getting started tutorials

## [Unreleased]

### Planned Features

- 🌐 Web dashboard for project management
- 📱 VS Code extension integration
- 🔗 GitHub Actions automation templates
- 🎨 Custom project template creation tools

---

**MoAI-ADK v0.1.16** - Making AI-driven development accessible to everyone! 🎉
