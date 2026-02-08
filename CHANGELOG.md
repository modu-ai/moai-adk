# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [2.1.0] - 2026-02-09

### Summary

Major update introducing SessionEnd hook support, Agent Teams enabled by default, and critical template system improvements. This release fixes cross-platform test failures and enhances the workflow execution system with intelligent mode selection.

### Breaking Changes

- `--auto` flag removed from workflow execution (auto-selection now default behavior)

### Added

- **SessionEnd Hook**: New `.claude/hooks/moai/handle-session-end.sh` wrapper for Claude Code session lifecycle management
- **Agent Hook System**: Dedicated agent-specific hook configuration in agent frontmatter with PreToolUse, PostToolUse, and SubagentStop support
- **Session Management**: Automatic session cleanup and state persistence through SessionEnd event handling

### Changed

- **Agent Teams Default**: Teams mode now enabled by default with complexity-based auto-selection (3+ domains, 10+ files, or score 7+)
- **Workflow Mode Selection**: Simplified execution mode logic — auto-selection analyzes task complexity to choose between team and sub-agent modes
- **Parallel Execution**: Enhanced efficiency with Agent Teams as primary execution mode for complex workflows

### Fixed

- **Cross-Platform Tests**: Resolved Windows path escaping, macOS Unicode NFD/NFC normalization, and non-git directory detection errors
- **Windows CI**: Fixed path separator issues, permission tests, and filesystem compatibility across Windows, macOS, and Linux
- **Template Filter**: `moai update` now correctly processes `.tmpl` files using rendered target paths instead of template paths
- **JSON Logging**: Merge confirmation now uses structured output, fixing JSON formatting issues during `moai update`
- **Config Cleanup**: Full configuration backup (including sections/) ensures complete v2.x-to-v2.x migration restore capability
- **Test Imports**: Removed unused `runtime` imports from shell and template test files

### Removed

- **Deprecated Flag**: `--auto` flag (auto-selection now default)
- **builder-command.md**: Removed 1,208-line agent definition in favor of skill-based command creation approach
- **Verbose Docs**: Cleaned up redundant documentation in hooks-system.md and workflow skills
- **Settings Bloat**: Removed unused settings from settings.json template

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.1.0] - 2026-02-09 (한국어)

### 요약

SessionEnd 훅 지원, Agent Teams 기본 활성화, 템플릿 시스템 개선을 포함한 주요 업데이트입니다. 크로스 플랫폼 테스트 실패를 수정하고 지능형 모드 선택으로 워크플로우 실행 시스템을 강화했습니다.

### 주요 변경 사항 (Breaking Changes)

- `--auto` 플래그 제거 (자동 선택이 이제 기본 동작)

### 추가됨 (Added)

- **SessionEnd Hook**: Claude Code 세션 생명주기 관리를 위한 `.claude/hooks/moai/handle-session-end.sh` 래퍼
- **Agent Hook System**: 에이전트별 훅 설정 지원 (PreToolUse, PostToolUse, SubagentStop)
- **세션 관리**: SessionEnd 이벤트를 통한 자동 세션 정리 및 상태 지속성

### 변경됨 (Changed)

- **Agent Teams 기본 활성화**: 복잡도 기반 자동 선택으로 Teams 모드가 기본값 (3개 이상 도메인, 10개 이상 파일, 또는 점수 7 이상)
- **워크플로우 모드 선택**: 실행 모드 로직 단순화 — 작업 복잡도를 분석하여 팀 모드와 서브 에이전트 모드 중 선택
- **병렬 실행 강화**: Agent Teams를 복잡한 워크플로우의 주요 실행 모드로 사용하여 효율성 향상

### 수정됨 (Fixed)

- **크로스 플랫폼 테스트**: Windows 경로 이스케이핑, macOS Unicode NFD/NFC 정규화, non-git 디렉토리 감지 오류 해결
- **Windows CI**: 경로 구분자 문제, 권한 테스트, Windows/macOS/Linux 파일시스템 호환성 수정
- **템플릿 필터**: `moai update`가 템플릿 경로 대신 렌더링된 대상 경로를 사용하여 `.tmpl` 파일을 올바르게 처리
- **JSON 로깅**: 병합 확인이 구조화된 출력을 사용하여 `moai update` 중 JSON 형식 문제 해결
- **설정 정리**: sections/를 포함한 전체 설정 백업으로 완전한 v2.x-to-v2.x 마이그레이션 복원 보장
- **테스트 import**: shell 및 template 테스트 파일에서 사용하지 않는 `runtime` import 제거

### 제거됨 (Removed)

- **더 이상 사용되지 않는 플래그**: `--auto` 플래그 (자동 선택이 기본값)
- **builder-command.md**: 1,208줄 에이전트 정의를 스킬 기반 명령 생성 방식으로 대체
- **장황한 문서**: hooks-system.md 및 워크플로우 스킬에서 중복 문서 정리
- **불필요한 설정**: settings.json 템플릿에서 사용되지 않는 설정 제거

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.5] - 2026-02-08

### Summary

Add git installation check to `moai init`, remove TUI experimental feature, and add v1-to-v2 migration cleanup utility.

### Breaking Changes

- Removed TUI (Terminal UI) experimental feature from `moai init` — `--tui` flag no longer available, `internal/cli/tui/` package deleted
- TUI will be redeveloped in future releases with improved architecture

### Added

- Git installation check in `moai init` with OS-specific installation guidance (macOS, Windows, Linux)
- `GitInstallHint()` function providing platform-specific git installation instructions
- `cleanMoaiManagedPaths()` utility for v1-to-v2 migration path cleanup
- Test coverage for git installation hints (`TestGitInstallHint`, `TestCheckGit_DetailWhenMissing`)

### Removed

- TUI (Terminal UI) experimental feature — 6 files deleted from `internal/cli/tui/` package (~1600 lines)
- `--tui` flag from `moai init` command
- `RunInitWizardTUI()` and `RunInitWithTUI()` functions
- Bubble Tea dependency from init command (CLI wizard remains intact)

### Changed

- `moai init` now shows non-fatal warning when git is not installed instead of silently continuing
- Git check runs after binary update step, before flag parsing

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.5] - 2026-02-08 (한국어)

### 요약

`moai init`에 git 설치 확인 기능을 추가하고, TUI 실험 기능을 제거하며, v1-to-v2 마이그레이션 정리 유틸리티를 추가했습니다.

### 주요 변경 사항 (Breaking Changes)

- TUI (Terminal UI) 실험 기능 제거 — `--tui` 플래그 더 이상 사용 불가, `internal/cli/tui/` 패키지 삭제
- TUI는 향후 개선된 아키텍처로 재개발될 예정

### 추가

- `moai init`에 OS별 설치 안내가 포함된 git 설치 확인 기능 추가 (macOS, Windows, Linux)
- 플랫폼별 git 설치 지침을 제공하는 `GitInstallHint()` 함수 추가
- v1-to-v2 마이그레이션 경로 정리를 위한 `cleanMoaiManagedPaths()` 유틸리티 추가
- git 설치 힌트 테스트 커버리지 추가 (`TestGitInstallHint`, `TestCheckGit_DetailWhenMissing`)

### 제거

- TUI (Terminal UI) 실험 기능 — `internal/cli/tui/` 패키지에서 6개 파일 삭제 (~1600줄)
- `moai init` 명령에서 `--tui` 플래그 제거
- `RunInitWizardTUI()`와 `RunInitWithTUI()` 함수 제거
- init 명령에서 Bubble Tea 의존성 제거 (CLI wizard는 유지)

### 변경

- git이 설치되지 않은 경우 `moai init`이 치명적 오류 대신 경고 메시지 표시
- git 확인은 바이너리 업데이트 단계 후, 플래그 파싱 전에 실행

### 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.4] - 2026-02-08

### Summary

Fix version persistence in `moai update` and `moai init`, and exclude hook files from merge confirmation UI. Official documentation link added to all README files.

### Breaking Changes

None

### Fixed

- Template version not persisted after `moai update` — `WithVersion()` was missing from `TemplateContext` creation in both `update.go` and `initializer.go`, causing `config.yaml` to render with empty version fields
- Status line showing stale version (`v1.14.0`) and perpetual update indicator because `moai.version` was empty in config
- `.claude/hooks/moai/*` files incorrectly appearing in merge confirmation UI during `moai update` — added `hooks` to `isMoaiManaged()` filter

### Added

- Official documentation link (https://adk.mo.ai.kr) to all README files (EN, KO, JA, ZH)
- Test cases for hooks path in `TestIsMoaiManaged` (3 new cases)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.4] - 2026-02-08 (한국어)

### 요약

`moai update`와 `moai init`에서 템플릿 버전이 저장되지 않던 버그를 수정하고, 훅 파일이 병합 확인 UI에 노출되던 문제를 해결했습니다. 모든 README에 공식 문서 링크를 추가했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- `moai update` 후 템플릿 버전이 저장되지 않는 버그 — `update.go`와 `initializer.go`에서 `TemplateContext` 생성 시 `WithVersion()`이 누락되어 `config.yaml`의 버전 필드가 빈 문자열로 렌더링됨
- 상태 표시줄에 이전 버전(`v1.14.0`)이 표시되고 업데이트 표시가 계속 나타나는 문제 — config의 `moai.version`이 비어있었기 때문
- `moai update` 중 `.claude/hooks/moai/*` 파일이 병합 확인 UI에 잘못 표시되는 문제 — `isMoaiManaged()` 필터에 `hooks` 추가

### 추가됨 (Added)

- 모든 README(EN, KO, JA, ZH)에 공식 문서 링크(https://adk.mo.ai.kr) 추가
- `TestIsMoaiManaged`에 hooks 경로 테스트 케이스 3개 추가

### 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.3] - 2026-02-07

### Summary

Binary-first self-update and configuration improvements. The `moai update` command now updates the binary before syncing templates, ensuring the latest template engine processes files. Agent hook definitions and settings schema have been corrected.

### Breaking Changes

None

### Added

- Binary self-update step in `moai update` and `moai init` commands with re-exec pattern
- 3-layer loop prevention for binary update: env var guard, dev build detection, version comparison
- `--templates-only` flag for skipping binary update during re-exec
- `plansDirectory` setting in settings.json for Claude Code plan storage

### Changed

- `moai update` now performs binary update before template sync
- Agent hook definitions converted from object to array format for SubagentStop events
- Removed Homebrew tap from GoReleaser configuration

### Fixed

- Invalid schema fields removed from settings.json template
- Missing configuration fields added to settings.json template
- SubagentStop hooks in 8 agent definitions corrected to valid array format

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.3] - 2026-02-07 (한국어)

### 요약

바이너리 우선 자체 업데이트 및 설정 개선. `moai update` 명령어가 이제 템플릿 동기화 전에 바이너리를 먼저 업데이트하여 최신 템플릿 엔진이 파일을 처리하도록 보장합니다. 에이전트 훅 정의와 설정 스키마가 수정되었습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- `moai update` 및 `moai init` 명령어에 re-exec 패턴을 활용한 바이너리 자체 업데이트 단계 추가
- 바이너리 업데이트를 위한 3중 루프 방지: 환경변수 가드, 개발 빌드 감지, 버전 비교
- re-exec 시 바이너리 업데이트 건너뛰기를 위한 `--templates-only` 플래그
- Claude Code 계획 문서 저장을 위한 settings.json에 `plansDirectory` 설정 추가

### 변경됨 (Changed)

- `moai update`가 이제 템플릿 동기화 전에 바이너리 업데이트를 수행
- SubagentStop 이벤트의 에이전트 훅 정의를 객체에서 배열 형식으로 변환
- GoReleaser 설정에서 Homebrew tap 제거

### 수정됨 (Fixed)

- settings.json 템플릿에서 잘못된 스키마 필드 제거
- settings.json 템플릿에 누락된 설정 필드 추가
- 8개 에이전트 정의의 SubagentStop 훅을 유효한 배열 형식으로 수정

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.2] - 2026-02-07

### Summary

Template system refactoring and cross-platform compatibility improvements. This patch release migrates settings.json generation from runtime-based to template-based approach, improves PATH handling, and fixes Windows CI test failures.

### Breaking Changes

None

### Added

- Template-based configuration files: settings.json.tmpl, .mcp.json.tmpl, handle-session-end.sh.tmpl
- SmartPATH and Platform fields in TemplateContext for better cross-platform support

### Changed

- Migrated settings.json generation from runtime JSON builder to template-based rendering
- Simplified SettingsGenerator by removing complex JSON construction logic
- Removed settings.json merge logic from update command (now handled by template deployment)
- Enhanced template rendering with SmartPATH and Platform context

### Fixed

- Resolved cross-platform test failures on Windows CI
- Restored .moai/project, specs, and config directories deleted in v2.0.0 cleanup
- Fixed PowerShell `$IsWindows` read-only variable conflict

### Technical Details

**Template System Improvements:**
- Centralized configuration in templates for single source of truth
- Better cross-platform PATH handling via SmartPATH
- Consistent template rendering across init and update commands
- Reduced maintenance overhead with template-based approach

**Test Coverage:**
- All 30 packages pass race detection tests
- Zero linting issues
- Enhanced test coverage for template rendering

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.2] - 2026-02-07 (한국어)

### 요약

템플릿 시스템 리팩토링 및 크로스 플랫폼 호환성 개선. 이 패치 릴리스는 settings.json 생성을 런타임 기반에서 템플릿 기반 접근 방식으로 마이그레이션하고, PATH 처리를 개선하며, Windows CI 테스트 실패를 수정합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- 템플릿 기반 구성 파일: settings.json.tmpl, .mcp.json.tmpl, handle-session-end.sh.tmpl
- 더 나은 크로스 플랫폼 지원을 위한 TemplateContext의 SmartPATH 및 Platform 필드

### 변경됨 (Changed)

- settings.json 생성을 런타임 JSON 빌더에서 템플릿 기반 렌더링으로 마이그레이션
- 복잡한 JSON 구성 로직을 제거하여 SettingsGenerator 단순화
- update 명령에서 settings.json 병합 로직 제거 (이제 템플릿 배포로 처리)
- SmartPATH 및 Platform 컨텍스트로 템플릿 렌더링 강화

### 수정됨 (Fixed)

- Windows CI에서 크로스 플랫폼 테스트 실패 해결
- v2.0.0 정리 시 삭제된 .moai/project, specs, config 디렉토리 복원
- PowerShell `$IsWindows` 읽기 전용 변수 충돌 수정

### 기술 세부 사항

**템플릿 시스템 개선:**
- 단일 소스로서의 템플릿에 구성 중앙화
- SmartPATH를 통한 더 나은 크로스 플랫폼 PATH 처리
- init 및 update 명령에서 일관된 템플릿 렌더링
- 템플릿 기반 접근 방식으로 유지 관리 오버헤드 감소

**테스트 커버리지:**
- 30개 패키지 모두 race detection 테스트 통과
- linting 문제 0개
- 템플릿 렌더링에 대한 향상된 테스트 커버리지

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.1] - 2026-02-07

### 요약

Windows 설치 스크립트 버그 수정 및 릴리즈 워크플로우 개선

### 주요 변경 사항

없음

### 수정됨

- Windows PowerShell 6+ 환경에서 `$IsWindows` 읽기 전용 변수 충돌 해결
- `moai update` 실행 시 불필요한 JSON 로그 출력 제거 (merge confirmation)

### 변경됨

- 릴리즈 노트 이중언어 형식을 영어 우선으로 변경 (이전: 한국어 우선)
- CI/CD 워크플로우에 OAuth 토큰 설정 추가

### 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.1] - 2026-02-07 (English)

### Summary

Windows installer bugfix and release workflow improvements

### Breaking Changes

None

### Fixed

- Resolved PowerShell `$IsWindows` read-only variable conflict in Windows installer (PowerShell 6+)
- Removed unwanted JSON log output during `moai update` (merge confirmation)

### Changed

- Updated release notes bilingual format to English-first (previously Korean-first)
- Added OAuth token configuration to CI/CD workflows

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.0] - 2026-02-06

### Summary

**Major Release: MoAI-ADK Go Edition**

This is the first official release of MoAI-ADK Go Edition, a complete rewrite of the Python-based MoAI-ADK in Go. This release delivers significantly improved performance, easier installation, and enhanced features while maintaining full compatibility with Claude Code workflows.

### Latest Updates (2026-02-06)

**Template Synchronization:**
- Synchronized 17 agent definition files with updated skill frontmatter
- Updated workflow skills (SKILL.md v2.0.0, moai.md) with team mode support
- Updated workflow-modes.md with Hybrid methodology as default
- Synchronized workflow.yaml and status_line.sh templates
- Updated CLAUDE.md to v12.0.0 with Agent Teams documentation

**Agent Hooks System:**
- Added agent-specific hooks for workflow enforcement
- Implemented `SubagentStop` event type for agent completion hooks
- Created `handle-agent-hook.sh` wrapper script for agent hooks
- Added factory pattern for agent-specific handlers in `internal/hook/agents/`
- Implemented hook actions for DDD workflow (ddd-pre-transformation, ddd-post-transformation, ddd-completion)
- Implemented hook actions for TDD workflow (tdd-pre-implementation, tdd-post-implementation, tdd-completion)
- Added validation/verification hooks for expert agents (backend, frontend, testing, debug, devops)
- Added completion hooks for manager agents (quality, spec, docs)
- Updated hooks-system.md documentation with agent hooks reference
- Synchronized agent hook configuration to all template locations

**Code Quality Improvements:**
- Fixed missing error checks in init_tui.go (added nolint comments for informational messages)
- Fixed missing error checks in init.go (added nolint comment for informational message)
- Simplified character validation logic in wizard_tui.go using De Morgan's law
- All 26 packages pass race detection tests
- Zero linting issues after fixes
- Fixed `.tmpl` file display in `moai update` (now shows rendered target paths)
- Fixed `permissions.allow` format (array instead of string per Claude Code IAM docs)

**Language Configuration:**
- Default conversation language set to Korean (ko) for improved user experience

**Additional Updates (Post v2.0.0 Tag):**
- **Documentation Restructuring**:
  - Made English the default README, moved Korean to README.ko.md (2e28f54f)
  - Maintained multilingual support (EN, JA, ZH, KO)
- **CI/CD Enhancements**:
  - Switched claude-code-action to GLM API Key (unofficial) (29d353ca)
  - Added open-source AI automation infrastructure (ffcaa6a2)
  - Improved CI/CD workflows with CodeQL, community automation
- **Project Organization**:
  - Untracked .moai local config, keeping only project/ and status_line.sh (8153bb19)
  - Cleaned up 38,895 lines of stale SPEC/project files
- **GitHub Flow Integration**:
  - Added /moai cpr command for issue-to-PR automation (081e5b7a)
  - Switched to GitHub Flow branch protection with feature/hotfix patterns (61f54378)
  - Made git delivery strategy-aware instead of GitHub Flow only (3fdec7aa)
- **Agent Teams Infrastructure** (a95e2a8d):
  - Added 8 team agents: team-researcher, team-analyst, team-architect, team-designer, team-backend-dev, team-frontend-dev, team-tester, team-quality
  - Created team workflow skills: team-plan, team-run, team-debug, team-review, team-sync
  - Implemented dual-mode execution (sub-agent vs Agent Teams)
  - Added complexity-based automatic mode selection
- **Settings Migration** (d01d16b8):
  - Migrated env, permissions, and teammateMode from global to project-level settings
  - Smart PATH capture instead of removing env.PATH (233f8907, 76500f84)
  - Added required type field to statusLine configuration (ad40b799)
- **Code Quality**:
  - Improved StatusLine version display format with config fallback (9a8183cc)
  - Fixed CI builds for Go 1.25 compatibility with golangci-lint (c72f4516, 542e146b, c58a61f7)
- **Community Infrastructure**:
  - Added CONTRIBUTING.md (KO/EN), SECURITY.md, LICENSE
  - GitHub issue/PR templates, dependabot, labeler, CodeQL

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
- **Agent Hooks System**: Agent-specific hooks for workflow enforcement
  - SubagentStop event type for agent lifecycle hooks
  - handle-agent-hook.sh wrapper script for consistent interface
  - Factory pattern for agent-specific handlers
  - DDD workflow hooks (pre/post-transformation, completion)
  - TDD workflow hooks (pre/post-implementation, completion)
  - Expert agent validation/verification hooks
  - Manager agent completion hooks

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

### 최신 업데이트 (2026-02-06)

**템플릿 동기화:**
- 업데이트된 스킬 프론트매터로 17개 에이전트 정의 파일 동기화
- 팀 모드 지원이 포함된 워크플로우 스킬 (SKILL.md v2.0.0, moai.md) 업데이트
- Hybrid 방법론을 기본값으로 사용하는 workflow-modes.md 업데이트
- workflow.yaml 및 status_line.sh 템플릿 동기화
- Agent Teams 문서가 포함된 CLAUDE.md v12.0.0 업데이트

**코드 품질 개선:**
- init_tui.go에서 누락된 오류 검사 수정 (정보 메시지에 nolint 주석 추가)
- init.go에서 누락된 오류 검사 수정 (정보 메시지에 nolint 주석 추가)
- 드 모르간 법칙을 사용한 wizard_tui.go의 문자 검증 로직 단순화
- 26개 패키지 모두 race detection 테스트 통과
- 수정 후 linting 문제 0개
- `moai update`에서 `.tmpl` 파일 표시 수정 (이제 렌더링된 대상 경로 표시)
- `permissions.allow` 형식 수정 (Claude Code IAM 문서에 따라 문자열 대신 배열 사용)

**언어 설정:**
- 개선된 사용자 경험을 위해 기본 대화 언어를 한국어(ko)로 설정

**추가 업데이트 (v2.0.0 태그 이후):**
- **문서 재구성**:
  - 영문 README를 기본으로 설정, 한국어를 README.ko.md로 이동 (2e28f54f)
  - 다국어 지원 유지 (EN, JA, ZH, KO)
- **CI/CD 개선**:
  - claude-code-action을 GLM API Key로 전환 (비공식) (29d353ca)
  - 오픈소스 AI 자동화 인프라 추가 (ffcaa6a2)
  - CodeQL, 커뮤니티 자동화를 포함한 CI/CD 워크플로우 개선
- **프로젝트 정리**:
  - .moai 로컬 설정 untrack, project/ 및 status_line.sh만 유지 (8153bb19)
  - 오래된 SPEC/project 파일 38,895줄 정리
- **GitHub Flow 통합**:
  - issue-to-PR 자동화를 위한 /moai cpr 명령어 추가 (081e5b7a)
  - feature/hotfix 패턴을 사용한 GitHub Flow 브랜치 보호 전환 (61f54378)
  - GitHub Flow만이 아닌 전략 인식 git 전달 방식으로 변경 (3fdec7aa)
- **에이전트 팀 인프라** (a95e2a8d):
  - 8개 팀 에이전트 추가: team-researcher, team-analyst, team-architect, team-designer, team-backend-dev, team-frontend-dev, team-tester, team-quality
  - 팀 워크플로우 스킬 생성: team-plan, team-run, team-debug, team-review, team-sync
  - 이중 모드 실행 구현 (sub-agent vs Agent Teams)
  - 복잡도 기반 자동 모드 선택 추가
- **설정 마이그레이션** (d01d16b8):
  - env, permissions, teammateMode를 global에서 project-level로 마이그레이션
  - env.PATH 제거 대신 Smart PATH 캡처 (233f8907, 76500f84)
  - statusLine 구성에 필수 type 필드 추가 (ad40b799)
- **코드 품질**:
  - config fallback을 사용한 StatusLine 버전 표시 형식 개선 (9a8183cc)
  - golangci-lint와 Go 1.25 호환성을 위한 CI 빌드 수정 (c72f4516, 542e146b, c58a61f7)
- **커뮤니티 인프라**:
  - CONTRIBUTING.md (KO/EN), SECURITY.md, LICENSE 추가
  - GitHub 이슈/PR 템플릿, dependabot, labeler, CodeQL

**에이전트 훅 시스템:**
- 워크플로우 강제를 위한 에이전트별 훅 추가
- 에이전트 완료 훅을 위한 `SubagentStop` 이벤트 타입 구현
- 에이전트 훅을 위한 `handle-agent-hook.sh` 래퍼 스크립트 생성
- `internal/hook/agents/`의 에이전트별 핸들러를 위한 팩토리 패턴 추가
- DDD 워크플로우 훅 구현 (ddd-pre-transformation, ddd-post-transformation, ddd-completion)
- TDD 워크플로우 훅 구현 (tdd-pre-implementation, tdd-post-implementation, tdd-completion)
- 전문가 에이전트를 위한 검증/확인 훅 추가 (backend, frontend, testing, debug, devops)
- 관리자 에이전트를 위한 완료 훅 추가 (quality, spec, docs)
- 에이전트 훅 참조가 포함된 hooks-system.md 문서 업데이트
- 모든 템플릿 위치에 에이전트 훅 구성 동기화

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
- **에이전트 훅 시스템**: 워크플로우 강제를 위한 에이전트별 훅
  - 에이전트 수명주기 훅을 위한 SubagentStop 이벤트 타입
  - 일관된 인터페이스를 위한 handle-agent-hook.sh 래퍼 스크립트
  - 에이전트별 핸들러를 위한 팩토리 패턴
  - DDD 워크플로우 훅 (pre/post-transformation, completion)
  - TDD 워크플로우 훅 (pre/post-implementation, completion)
  - 전문가 에이전트 검증/확인 훅
  - 관리자 에이전트 완료 훅

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
