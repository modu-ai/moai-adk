# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-02-06

### Summary

**Major Release: MoAI-ADK Go Edition**

This is the first official release of MoAI-ADK Go Edition, a complete rewrite of the Python-based MoAI-ADK in Go. This release delivers significantly improved performance, easier installation, and enhanced features while maintaining full compatibility with Claude Code workflows.

### Breaking Changes

- **Installation Method**: Changed from `uv tool install moai-adk` to single binary installation
- **Hook System**: Migrated from Python hooks to shell script wrappers
- **Configuration**: Updated configuration file structure and locations
- **Update Mechanism**: New automatic update system with GitHub releases integration

### Added

- **Go Edition Core**: Complete rewrite in Go for better performance and easier distribution
- **Multi-platform Binary Support**: Pre-built binaries for macOS (ARM64/Intel), Linux (ARM64/AMD64), Windows (AMD64)
- **Embedded Template System**: Templates now embedded using `go:embed` for faster startup
- **Web-based Installation UI**: Modern web interface for installation instructions
- **Korean Documentation**: Full Korean language documentation and migration guide
- **Go-specific Release Command**: `/moai:99-release` for automated release workflow
- **Transcript Parsing**: Support for Claude Code transcript analysis with MoAI Rank
- **LSP Quality Gates**: Integrated LSP diagnostics for quality validation
- **Security Scanner**: Hook-based security scanning for code changes
- **i18n Support**: Multi-language support in CLI commands
- **Agent Teams v2.0** (Experimental): Dual-mode execution engine with sub-agent and Agent Teams support
  - 5 team agents: researcher, backend-dev, frontend-dev, tester, quality
  - Team workflow skill with plan/run orchestration
  - `--team`, `--solo`, `--auto` execution mode flags
  - Complexity-based automatic mode selection
  - File ownership strategy for write conflict prevention
  - Workflow configuration (`workflow.yaml`) with team patterns
- **Hook Auto-Update**: Automatic update checking via session hooks
- **Update Cache**: Caching layer for update checks to reduce API calls

### Changed

- **Performance**: 10x faster startup time compared to Python version
- **Memory Usage**: Reduced memory footprint with Go runtime
- **Update System**: New update mechanism with GitHub releases integration
- **Template Deployment**: Automatic template deployment during initialization
- **Configuration Management**: Enhanced configuration with better validation
- **Development Methodology**: Hybrid (TDD+DDD) is now the default for new projects; DDD reserved for brownfield/legacy
- **CLI Update Command**: Refactored with extracted dependency management (`deps.go`)
- **StatusLine**: Improved version display and rendering with expanded test coverage
- **CLAUDE.md**: Updated to v12.0.0 with Agent Teams section (Section 15)
- **SKILL.md**: Updated to v2.0.0 with team mode support and execution mode selection

### Fixed

- **GitHub Issue #323**: Fixed PowerShell `irm | iex` installation failure
  - Wrapped install.ps1 script in `& { ... } @args` scriptblock for piping compatibility
  - Added ARM64 platform detection via ProcessArchitecture
  - Changed install location from `$env:USERPROFILE` to `$env:LOCALAPPDATA\Programs\moai`
  - Added SHA-256 checksum verification
- **GitHub Issue #324**: Fixed Linux/WSL2 installation 404 download error
  - Updated download URL to match goreleaser archive naming (`moai-adk_go-vX.Y.Z_OS_ARCH.tar.gz`)
  - Added tar.gz extraction step
  - Added SHA-256 checksum verification
  - Added WSL environment detection
- Windows CMD installation script improvements
  - Added ARM64 platform detection
  - Updated download URL to match goreleaser naming
  - Added extraction via PowerShell Expand-Archive
  - Fixed install location to `%LOCALAPPDATA%\Programs\moai`
- goreleaser configuration fixes
  - Fixed module path from `moai-adk-go` to `moai-adk` in ldflags
  - Fixed release target repository from `moai-adk-go` to `moai-adk`
- Windows hook execution improvements
  - Changed from `cmd.exe /c` to `bash` command (uses Git for Windows)
  - Ensures consistent hook execution across all platforms
- Cross-platform path construction
  - Replaced string concatenation with `filepath.Join()` in shell detection
  - Fixed path handling for PowerShell profile detection
- Update checker enhancements
  - Added `go-v` prefix support for version comparison
  - Updated archive naming to match goreleaser conventions
- StatusLine configuration
  - Changed from absolute path to relative path for better portability
  - Addresses GitHub Issue #7925 (StatusLine doesn't expand environment variables)
- Go bin path detection on Windows
  - Added fallback paths for Go installation directory detection
  - Checks `%PROGRAMFILES%\Go\bin` and `C:\Go\bin`
- Template synchronization issues in development builds
- Browser opening during automated tests
- Hook JSON output schema compliance
- API URL routing to correct repository

### Installation & Update

```bash
# Install MoAI-ADK Go Edition (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash

# Or download binary directly from GitHub Releases
# Visit: https://github.com/modu-ai/moai-adk/releases/tag/go-v2.0.0

# Update to the latest version
moai update

# Verify version
moai version
```

### Migration from Python Version

Users migrating from Python MoAI-ADK v1.x should:

1. Uninstall Python version: `uv tool uninstall moai-adk`
2. Install Go Edition using binary installation
3. Run `moai init` to update project templates

See [MIGRATION.ko.md](MIGRATION.ko.md) for detailed migration guide.

---

## [2.0.0] - 2026-02-06

### 요약

**메이저 릴리스: MoAI-ADK Go 에디션**

Python 기반 MoAI-ADK를 Go로 완전히 재작성한 첫 번째 공식 릴리스입니다. 성능이 크게 향상되고 설치가 간편해지며 기능이 향상되었습니다.

### Breaking Changes

- **설치 방법**: `uv tool install moai-adk`에서 단일 바이너리 설치로 변경
- **훅 시스템**: Python 훅에서 셸 스크립트 래퍼로 마이그레이션
- **설정**: 설정 파일 구조 및 위치 업데이트
- **업데이트 메커니즘**: GitHub 릴리스 통합 새 업데이트 시스템

### 추가됨

- **Go 에디션 코어**: 더 나은 성능과 배포를 위한 Go로 완전 재작성
- **멀티 플랫폼 바이너리 지원**: macOS (ARM64/Intel), Linux (ARM64/AMD64), Windows (AMD64)용 미리 빌드된 바이너리
- **임베디드 템플릿 시스템**: `go:embed`를 사용한 더 빠른 시작을 위한 템플릿 임베딩
- **웹 기반 설치 UI**: 설치 안내를 위한 현대적 웹 인터페이스
- **한국어 문서**: 완전한 한국어 문서 및 마이그레이션 가이드
- **Go 전용 릴리스 명령**: 자동화된 릴리스 워크플로우를 위한 `/moai:99-release`
- **트랜스크립트 파싱**: MoAI Rank를 위한 Claude Code 트랜스크립트 분석 지원
- **LSP 품질 게이트**: 품질 검증을 위한 통합 LSP 진단
- **보안 스캐너**: 코드 변경을 위한 훅 기반 보안 스캐닝
- **i18n 지원**: CLI 명령어의 다국어 지원

### 변경됨

- **성능**: Python 버전 대비 10배 더 빠른 시작 시간
- **메모리 사용량**: Go 런타임으로 감소된 메모리 사용량
- **업데이트 시스템**: GitHub 릴리스 통합 새 업데이트 메커니즘
- **템플릿 배포**: 초기화 중 자동 템플릿 배포
- **설정 관리**: 향상된 검증을 통한 개선된 설정

### 수정됨

- **GitHub Issue #323**: PowerShell `irm | iex` 설치 실패 수정
  - 파이핑 호환성을 위해 install.ps1 스크립트를 `& { ... } @args` 스크립트블록으로 래핑
  - ProcessArchitecture를 통한 ARM64 플랫폼 감지 추가
  - 설치 위치를 `$env:USERPROFILE`에서 `$env:LOCALAPPDATA\Programs\moai`로 변경
  - SHA-256 체크섬 검증 추가
- **GitHub Issue #324**: Linux/WSL2 설치 404 다운로드 오류 수정
  - goreleaser 아카이브 명명 규칙에 맞게 다운로드 URL 업데이트 (`moai-adk_go-vX.Y.Z_OS_ARCH.tar.gz`)
  - tar.gz 압축 해제 단계 추가
  - SHA-256 체크섬 검증 추가
  - WSL 환경 감지 추가
- Windows CMD 설치 스크립트 개선
  - ARM64 플랫폼 감지 추가
  - goreleaser 명명 규칙에 맞게 다운로드 URL 업데이트
  - PowerShell Expand-Archive를 통한 압축 해제 추가
  - 설치 위치를 `%LOCALAPPDATA%\Programs\moai`로 수정
- goreleaser 설정 수정
  - ldflags의 모듈 경로를 `moai-adk-go`에서 `moai-adk`로 수정
  - 릴리스 대상 저장소를 `moai-adk-go`에서 `moai-adk`로 수정
- Windows 훅 실행 개선
  - `cmd.exe /c`에서 `bash` 명령으로 변경 (Git for Windows 사용)
  - 모든 플랫폼에서 일관된 훅 실행 보장
- 크로스 플랫폼 경로 구성
  - 셸 감지에서 문자열 연결을 `filepath.Join()`으로 교체
  - PowerShell 프로필 감지를 위한 경로 처리 수정
- 업데이트 검사기 개선
  - 버전 비교를 위한 `go-v` 접두사 지원 추가
  - goreleaser 규칙에 맞게 아카이브 명명 업데이트
- StatusLine 설정
  - 이식성 향상을 위해 절대 경로에서 상대 경로로 변경
  - GitHub Issue #7925 해결 (StatusLine이 환경 변수를 확장하지 않음)
- Windows에서 Go bin 경로 감지
  - Go 설치 디렉터리 감지를 위한 대체 경로 추가
  - `%PROGRAMFILES%\Go\bin` 및 `C:\Go\bin` 확인
- 개발 빌드에서의 템플릿 동기화 문제
- 자동화된 테스트 중 브라우저 열림 문제
- 훅 JSON 출력 스키마 준수
- 올바른 저장소로의 API URL 라우팅

### 설치 및 업데이트

```bash
# MoAI-ADK Go 에디션 설치 (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash

# 또는 GitHub 릴리스에서 바이너리 직접 다운로드
# 방문: https://github.com/modu-ai/moai-adk/releases/tag/go-v2.0.0

# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

### Python 버전에서 마이그레이션

Python MoAI-ADK v1.x에서 마이그레이션하는 사용자는:

1. Python 버전 제거: `uv tool uninstall moai-adk`
2. 바이너리 설치로 Go 에디션 설치
3. `moai init` 실행으로 프로젝트 템플릿 업데이트

자세한 마이그레이션 가이드는 [MIGRATION.ko.md](MIGRATION.ko.md)를 참조하세요.

---

## Release History

For previous releases, see [GitHub Releases](https://github.com/modu-ai/moai-adk/releases).
