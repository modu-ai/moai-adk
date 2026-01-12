# v1.0.1 - Git Worktree Support Fix (2026-01-12)

## Summary

Patch release that fixes `.gitignore` configuration to properly track `llm-configs/` directory in git worktrees. This resolves the issue where `moai glm` command failed in worktree environments because untracked files are not copied to worktrees.

## Fixed

- **Git Worktree Support**: Added exception rules for `llm-configs/` directory in `.gitignore`
  - Problem: `/.moai/*` rule excluded `llm-configs/`, causing `moai glm` command to fail in worktrees
  - Solution: Added `!/.moai/llm-configs/` and `!/.moai/llm-configs/**` exception rules

## Changed

- **Test Suite**: Updated test mocks to match current GLM-only authentication flow
  - Updated `test_init_prompts_core.py` mock structure
  - Updated `test_init_enhanced.py` ProjectSetupAnswers structure
  - Fixed `test_prompts_cov.py` KeyboardInterrupt handling

## Installation

```bash
pip install moai-adk==1.0.1
uv tool install moai-adk
```

---

# v1.0.1 - Git Worktree 지원 수정 (2026-01-12)

## 요약

git worktree에서 `llm-configs/` 디렉토리를 올바르게 추적하도록 `.gitignore` 설정을 수정한 패치 릴리스입니다. 추적되지 않는 파일은 worktree에 복사되지 않기 때문에 worktree 환경에서 `moai glm` 명령어가 실패하는 문제를 해결합니다.

## 수정

- **Git Worktree 지원**: `.gitignore`에 `llm-configs/` 디렉토리 예외 규칙 추가
  - 문제: `/.moai/*` 규칙이 `llm-configs/`를 제외하여 worktree에서 `moai glm` 명령어 실패
  - 해결: `!/.moai/llm-configs/` 및 `!/.moai/llm-configs/**` 예외 규칙 추가

## 변경

- **테스트 스위트**: 현재 GLM-only 인증 플로우에 맞게 테스트 mock 업데이트
  - `test_init_prompts_core.py` mock 구조 업데이트
  - `test_init_enhanced.py` ProjectSetupAnswers 구조 업데이트
  - `test_prompts_cov.py` KeyboardInterrupt 처리 수정

## 설치

```bash
pip install moai-adk==1.0.1
uv tool install moai-adk
```

---

# v1.0.0 - Production Ready Release (2026-01-12)

## Summary

MoAI-ADK reaches version 1.0.0, marking its **Production/Stable** status. This major release includes the Ralph Engine for intelligent code quality assurance, the Rank System with Dashboard TUI, a major CLI redesign with multilingual support, and comprehensive documentation updates. The package is now ready for production use with 9,800+ tests and 80%+ coverage.

## Highlights

- 🚀 **Ralph Engine**: Intelligent code quality assurance with LSP and AST-grep integration
- 📊 **Rank System**: Dashboard TUI for development progress tracking
- 🌐 **CLI Redesign**: Streamlined init flow with multilingual support
- 📦 **curl Installation**: One-line installation via `curl -fsSL https://moai-adk.github.io/MoAI-ADK/install.sh | sh`
- ✨ **Production Status**: Development Status upgraded from Beta to Production/Stable

## Added

### Ralph Engine (SPEC-RALPH-001)

- **LSP Integration Layer**: Real-time diagnostics with 16+ language support
  - `MoAILSPClient`: High-level LSP client interface
  - `LSPServerManager`: Server lifecycle management
  - `LSPProtocol`: JSON-RPC 2.0 implementation
- **AST-grep Analyzer**: Structural pattern matching and security scanning
  - Support for 20+ programming languages
  - Security rule scanning with severity levels
- **Loop Controller**: Ralph-style autonomous feedback loops
  - Completion detection and progress tracking
  - State persistence with `LoopState` and `LoopStorage`
- **Slash Commands**: `/moai:loop`, `/moai:fix`, `/moai:cancel-loop`
- **Test Coverage**: 302 tests for Ralph Engine

### Rank System

- Dashboard TUI with real-time progress tracking
- Hook-based development activity monitoring
- Stability improvements and test fixes

### CLI Improvements

- Major CLI redesign with streamlined init flow
- Multilingual support (English, Korean, Japanese, Chinese)
- `moai-wt` alias for `moai-worktree` command
- curl install script with GitHub Pages deployment

### Infrastructure

- Terminal PTY support with WebSocket integration
- Multi-LLM support with GLM config auto-copy
- Permission settings optimization for bypass/acceptEdits modes

## Changed

- **CLI Display**: Standardized CLI command display to use `moai` alias
- **Command Naming**: Renamed `moai:all-is-well` to `moai:alfred`
- **Skills**: Converted all 40 skills to CLAUDE.md Documentation Standards
- **Agent Ecosystem**: Consolidated from 28 to 20 agents, 50 to 47 skills
- **Development Status**: Upgraded from "4 - Beta" to "5 - Production/Stable"

## Fixed

- MongoDB deny pattern parse error in settings.json
- GitHub Pages URL case sensitivity
- Config sections sync with templates
- PTY read blocking issue with select-based non-blocking read

## Documentation

- Enhanced README with 9-step wizard and quality focus
- Synchronized all language READMEs with Korean master template
- Added Star History chart to Korean README
- Added official online documentation link to all README files

## Installation

```bash
# One-line installation (recommended)
curl -fsSL https://moai-adk.github.io/MoAI-ADK/install.sh | sh

# Or via pip/uv
pip install moai-adk==1.0.0
uv tool install moai-adk

# Verify installation
moai --version
```

## Statistics

- **Commits since v0.41.2**: 74 commits
- **Files changed**: 433 files
- **Test Coverage**: 80%+ (9,800+ tests)

---

# v1.0.0 - 프로덕션 준비 릴리스 (2026-01-12)

## 요약

MoAI-ADK가 버전 1.0.0에 도달하여 **Production/Stable** 상태가 되었습니다. 이 주요 릴리스에는 지능형 코드 품질 보증을 위한 Ralph Engine, 대시보드 TUI가 포함된 Rank System, 다국어 지원이 포함된 주요 CLI 재설계, 그리고 포괄적인 문서 업데이트가 포함되어 있습니다. 이 패키지는 9,800개 이상의 테스트와 80% 이상의 커버리지로 프로덕션 사용 준비가 완료되었습니다.

## 하이라이트

- 🚀 **Ralph Engine**: LSP 및 AST-grep 통합을 통한 지능형 코드 품질 보증
- 📊 **Rank System**: 개발 진행 상황 추적을 위한 대시보드 TUI
- 🌐 **CLI 재설계**: 다국어 지원이 포함된 간소화된 init 플로우
- 📦 **curl 설치**: `curl -fsSL https://moai-adk.github.io/MoAI-ADK/install.sh | sh`로 한 줄 설치
- ✨ **프로덕션 상태**: 개발 상태가 Beta에서 Production/Stable로 업그레이드

## 추가됨

### Ralph Engine (SPEC-RALPH-001)

- **LSP 통합 레이어**: 16개 이상 언어 지원 실시간 진단
  - `MoAILSPClient`: 고수준 LSP 클라이언트 인터페이스
  - `LSPServerManager`: 서버 생명주기 관리
  - `LSPProtocol`: JSON-RPC 2.0 구현
- **AST-grep 분석기**: 구조적 패턴 매칭 및 보안 스캐닝
  - 20개 이상 프로그래밍 언어 지원
  - 심각도 레벨별 보안 규칙 스캐닝
- **Loop Controller**: Ralph 스타일 자율 피드백 루프
  - 완료 감지 및 진행률 추적
  - `LoopState` 및 `LoopStorage`를 통한 상태 영속성
- **슬래시 커맨드**: `/moai:loop`, `/moai:fix`, `/moai:cancel-loop`
- **테스트 커버리지**: Ralph Engine 302개 테스트

### Rank System

- 실시간 진행 상황 추적이 포함된 대시보드 TUI
- 훅 기반 개발 활동 모니터링
- 안정성 개선 및 테스트 수정

### CLI 개선

- 간소화된 init 플로우를 포함한 주요 CLI 재설계
- 다국어 지원 (영어, 한국어, 일본어, 중국어)
- `moai-worktree` 명령어를 위한 `moai-wt` 별칭
- GitHub Pages 배포가 포함된 curl 설치 스크립트

### 인프라

- WebSocket 통합이 포함된 터미널 PTY 지원
- GLM 설정 자동 복사가 포함된 Multi-LLM 지원
- bypass/acceptEdits 모드를 위한 권한 설정 최적화

## 변경됨

- **CLI 표시**: CLI 명령어 표시를 `moai` 별칭 사용으로 표준화
- **명령어 이름**: `moai:all-is-well`을 `moai:alfred`로 이름 변경
- **스킬**: 40개 모든 스킬을 CLAUDE.md 문서 표준으로 변환
- **에이전트 생태계**: 28개에서 20개 에이전트로, 50개에서 47개 스킬로 통합
- **개발 상태**: "4 - Beta"에서 "5 - Production/Stable"로 업그레이드

## 수정됨

- settings.json의 MongoDB deny 패턴 파싱 오류
- GitHub Pages URL 대소문자 구분 문제
- 템플릿과 config 섹션 동기화 문제
- select 기반 논블로킹 읽기로 PTY 읽기 블로킹 문제 해결

## 문서

- 9단계 마법사 및 품질 집중으로 README 개선
- 한국어 마스터 템플릿과 모든 언어 README 동기화
- 한국어 README에 Star History 차트 추가
- 모든 README 파일에 공식 온라인 문서 링크 추가

## 설치

```bash
# 한 줄 설치 (권장)
curl -fsSL https://moai-adk.github.io/MoAI-ADK/install.sh | sh

# 또는 pip/uv 사용
pip install moai-adk==1.0.0
uv tool install moai-adk

# 설치 확인
moai --version
```

## 통계

- **v0.41.2 이후 커밋 수**: 74개 커밋
- **변경된 파일 수**: 433개 파일
- **테스트 커버리지**: 80% 이상 (9,800개 이상 테스트)

---

# v0.41.2 - Skill Menu Visibility Enhancement (2026-01-09)

## Summary

Patch release improving user experience by hiding internal skills from the slash command menu while maintaining full functionality. All 45 skills now include `user-invocable: false` frontmatter to reduce menu clutter and improve discoverability of user-facing commands.

## Changes

### User Experience

- **feat(skills)**: Add user-invocable: false to all skill frontmatter (5cc6088e)
  - Add `user-invocable: false` to 45 skill files
  - Hide skills from slash command menu (`/` autocomplete)
  - Skills remain fully accessible via `Skill()` function and Agent invocation
  - Improves UX by reducing menu clutter
  - Affected files: All skill SKILL.md files in `.claude/skills/` and `src/moai_adk/templates/.claude/skills/`

### Maintenance

- **chore**: Bump version to 0.41.2 (1d2b095a)

### Quality

- All tests passing (9,627 passed, 85.70% coverage)
- Zero linting or type checking issues
- Security checks passed

## Installation & Update

```bash
# Install
pip install moai-adk==0.41.2
# or
uv pip install moai-adk==0.41.2

# Update existing installation
pip install --upgrade moai-adk
# or
uv pip install --upgrade moai-adk
```

## Migration Guide

No action required. Skills that were previously visible in the slash command menu will now be hidden but remain fully functional when invoked by Alfred or agents.

---

# v0.41.2 - 스킬 메뉴 가시성 개선 (2026-01-09)

## 요약

내부 스킬을 슬래시 명령 메뉴에서 숨겨 사용자 경험을 개선한 패치 릴리즈입니다. 45개 모든 스킬에 `user-invocable: false` frontmatter를 추가하여 메뉴 혼잡도를 줄이고 사용자 대상 명령어의 발견성을 향상시켰습니다.

## 변경 사항

### 사용자 경험

- **feat(skills)**: 모든 스킬 frontmatter에 user-invocable: false 추가 (5cc6088e)
  - 45개 스킬 파일에 `user-invocable: false` 추가
  - 슬래시 명령 메뉴(`/` 자동완성)에서 스킬 숨김
  - `Skill()` 함수 및 Agent 호출을 통한 완전한 접근성 유지
  - 메뉴 혼잡도 감소로 UX 개선
  - 영향받는 파일: `.claude/skills/` 및 `src/moai_adk/templates/.claude/skills/`의 모든 SKILL.md 파일

### 유지보수

- **chore**: 버전을 0.41.2로 업데이트 (1d2b095a)

### 품질

- 모든 테스트 통과 (9,627개 통과, 커버리지 85.70%)
- 린팅 및 타입 체킹 이슈 없음
- 보안 검사 통과

## 설치 및 업데이트

```bash
# 설치
pip install moai-adk==0.41.2
# 또는
uv pip install moai-adk==0.41.2

# 기존 설치 업데이트
pip install --upgrade moai-adk
# 또는
uv pip install --upgrade moai-adk
```

## 마이그레이션 가이드

별도 조치 불필요. 이전에 슬래시 명령 메뉴에 표시되던 스킬들이 숨겨지지만 Alfred나 Agent를 통한 호출 시 정상 작동합니다.

---

# v0.41.1 - Critical Bug Fixes & File Read Enhancement (2026-01-09)

## Summary

Patch release resolving critical GitHub Issues #248 (statusline settings not reflected) and #249 (Windows UTF-8 encoding error), plus adding Claude Code 2.1.0 file read token enhancement for improved large file handling. This release includes comprehensive template synchronization with the latest Claude Code frontmatter format.

## Changes

### Bug Fixes

- **fix**: Resolve GitHub issues #248 and #249 (0e7e039c)
  - Issue #249: Add UTF-8 stdout/stderr reconfiguration for Windows terminals
    - Fixes `UnicodeEncodeError` with emoji characters on cp1252 encoding
    - Affected files: `session_start__show_project_info.py`, `session_end__auto_cleanup.py`
  - Issue #248: Statusline settings not reflected
    - Add `display_config.version` check before rendering Claude version
    - Add directory rendering support in `_build_compact_parts` and `_render_extended`
    - Ensures `statusline-config.yaml` settings are properly respected
    - Affected file: `renderer.py`

### New Features

- **feat**: Add CLAUDE_CODE_FILE_READ_MAX_OUTPUT_TOKENS env variable (f374f5b2)
  - Add Claude Code 2.1.0 file read token limit setting
  - Set default to 55555 tokens for enhanced file reading capability
  - Enables reading larger files without truncation
  - Reduces context loss during complex code analysis
  - Reference: [Claude Code 2.1.0 Release](https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md)

### Maintenance

- **chore**: Sync templates to local (pre-release v0.41.1) (f5add2f9)
  - Sync 28 agent files with security guard hook updates
  - Sync 5 hook files including Windows UTF-8 fix (Issue #249)
  - Sync 53 skill files with new frontmatter format (`user-invocable`, array `allowed-tools`)
  - Add new `moai-tool-opencode` skill
  - Update all skills to Claude Code 2026-01 frontmatter specification

### Quality

- All tests passing (9,627 passed, 85.70% coverage)
- Zero linting or type checking issues
- Security checks passed

## Installation & Update

```bash
# Install
pip install moai-adk==0.41.1
# or
uv pip install moai-adk==0.41.1

# Upgrade
pip install --upgrade moai-adk
# or
uv pip install --upgrade moai-adk
```

---

# v0.41.1 - 중요 버그 수정 및 파일 읽기 향상 (2026-01-09)

## 요약

중요한 GitHub Issues #248 (statusline 설정 미반영)과 #249 (Windows UTF-8 인코딩 오류)를 해결하고, 대용량 파일 처리 개선을 위한 Claude Code 2.1.0 파일 읽기 토큰 향상 기능을 추가한 패치 릴리스입니다. 최신 Claude Code frontmatter 형식으로 포괄적인 템플릿 동기화가 포함되어 있습니다.

## 변경 사항

### 버그 수정

- **fix**: GitHub issues #248 및 #249 해결 (0e7e039c)
  - Issue #249: Windows 터미널용 UTF-8 stdout/stderr 재설정 추가
    - cp1252 인코딩에서 emoji 문자의 `UnicodeEncodeError` 수정
    - 영향받은 파일: `session_start__show_project_info.py`, `session_end__auto_cleanup.py`
  - Issue #248: Statusline 설정 미반영 문제 수정
    - Claude 버전 렌더링 전 `display_config.version` 검사 추가
    - `_build_compact_parts` 및 `_render_extended`에 디렉토리 렌더링 지원 추가
    - `statusline-config.yaml` 설정이 제대로 반영되도록 보장
    - 영향받은 파일: `renderer.py`

### 새로운 기능

- **feat**: CLAUDE_CODE_FILE_READ_MAX_OUTPUT_TOKENS 환경 변수 추가 (f374f5b2)
  - Claude Code 2.1.0 파일 읽기 토큰 한도 설정 추가
  - 향상된 파일 읽기 기능을 위해 기본값 55555 토큰 설정
  - 잘림 없이 더 큰 파일 읽기 가능
  - 복잡한 코드 분석 중 컨텍스트 손실 감소
  - 참고: [Claude Code 2.1.0 Release](https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md)

### 유지보수

- **chore**: 템플릿을 로컬로 동기화 (pre-release v0.41.1) (f5add2f9)
  - 보안 가드 훅 업데이트가 포함된 28개 에이전트 파일 동기화
  - Windows UTF-8 수정이 포함된 5개 훅 파일 동기화 (Issue #249)
  - 새로운 frontmatter 형식이 포함된 53개 스킬 파일 동기화 (`user-invocable`, 배열 `allowed-tools`)
  - 새로운 `moai-tool-opencode` 스킬 추가
  - 모든 스킬을 Claude Code 2026-01 frontmatter 사양으로 업데이트

### 품질

- 모든 테스트 통과 (9,627 passed, 85.70% coverage)
- 린트 및 타입 체크 이슈 없음
- 보안 검사 통과

## 설치 및 업데이트

```bash
# 설치
pip install moai-adk==0.41.1
# 또는
uv pip install moai-adk==0.41.1

# 업그레이드
pip install --upgrade moai-adk
# 또는
uv pip install --upgrade moai-adk
```

---

# v0.41.0 - Claude Code 2026-01 Compatibility & OpenCode Integration (2026-01-08)

## Summary

Minor release adding comprehensive support for Claude Code 2026-01 frontmatter fields and introducing the moai-tool-opencode skill for OpenCode AI editor integration. This release updates all 51 skills and agents to use modern YAML list format and adds specialized hooks for security, TDD, and quality workflows.

## Changes

### New Features

- **feat(skills)**: Add Claude Code 2026-01 frontmatter fields support (6870ea00)
  - Add `context: fork` and `agent` fields to 4 workflow skills
  - Add `user-invocable: false` to 4 foundation skills
  - Migrate 48 skills to YAML list format for `allowed-tools`
  - Add hooks field to 3 agents (expert-security, manager-tdd, manager-quality)
  - Update moai-foundation-claude reference docs with new field documentation

- **feat(skills)**: Add moai-tool-opencode skill for OpenCode AI editor integration (a2cc7e62)
  - Comprehensive OpenCode configuration and usage patterns
  - 22 detailed module files covering configuration, usage, and development
  - Support for agents, commands, skills, MCP servers, LSP servers, and custom tools
  - ~6,800 lines of documentation and examples

### Documentation

- **docs(readme)**: Add Tool category with moai-tool-opencode and moai-tool-ast-grep (8d6ecf92)
  - Reorganize skill categories to highlight tool integrations
  - Add Tool category alongside Workflow, Domain, Language, and Platform

### Refactoring

- **refactor(skills)**: Rename moai-platform-opencode to moai-tool-opencode (a2cc7e62)
  - Better categorization: OpenCode is a tool, not a platform
  - Consistent with other tool skills like moai-tool-ast-grep

### Quality

- All tests passing (9,627 passed, 85.63% coverage)
- Zero linting or type checking issues
- Security checks passed
- Verified against official Claude Code documentation

## Installation & Update

```bash
# Install
pip install moai-adk==0.41.0
# or
uv pip install moai-adk==0.41.0

# Upgrade
pip install --upgrade moai-adk
# or
uv pip install --upgrade moai-adk
```

---

# v0.41.0 - Claude Code 2026-01 호환성 및 OpenCode 통합 (2026-01-08)

## 요약

Claude Code 2026-01 frontmatter 필드에 대한 포괄적 지원을 추가하고 OpenCode AI 편집기 통합을 위한 moai-tool-opencode 스킬을 도입한 마이너 릴리스입니다. 이번 릴리스에서는 51개의 모든 스킬과 에이전트를 최신 YAML 리스트 형식으로 업데이트했으며, 보안, TDD, 품질 워크플로우를 위한 특화 hooks를 추가했습니다.

## 변경 사항

### 새로운 기능

- **feat(skills)**: Claude Code 2026-01 frontmatter 필드 지원 추가 (6870ea00)
  - 4개 워크플로우 스킬에 `context: fork`와 `agent` 필드 추가
  - 4개 foundation 스킬에 `user-invocable: false` 추가
  - 48개 스킬을 `allowed-tools` YAML 리스트 형식으로 마이그레이션
  - 3개 에이전트에 hooks 필드 추가 (expert-security, manager-tdd, manager-quality)
  - moai-foundation-claude reference 문서에 신규 필드 문서화 추가

- **feat(skills)**: OpenCode AI 편집기 통합을 위한 moai-tool-opencode 스킬 추가 (a2cc7e62)
  - 포괄적인 OpenCode 구성 및 사용 패턴 제공
  - 구성, 사용법, 개발을 다루는 22개 상세 모듈 파일
  - agents, commands, skills, MCP servers, LSP servers, custom tools 지원
  - 약 6,800줄의 문서 및 예제 제공

### 문서화

- **docs(readme)**: moai-tool-opencode와 moai-tool-ast-grep을 포함한 Tool 카테고리 추가 (8d6ecf92)
  - 도구 통합을 강조하기 위한 스킬 카테고리 재구성
  - Workflow, Domain, Language, Platform과 함께 Tool 카테고리 추가

### 리팩토링

- **refactor(skills)**: moai-platform-opencode를 moai-tool-opencode로 이름 변경 (a2cc7e62)
  - 더 나은 분류: OpenCode는 플랫폼이 아닌 도구
  - moai-tool-ast-grep과 같은 다른 도구 스킬과 일관성 유지

### 품질

- 모든 테스트 통과 (9,627 passed, 85.63% coverage)
- 린트 및 타입 체크 이슈 없음
- 보안 검사 통과
- Claude Code 공식 문서 기준 검증 완료

## 설치 및 업데이트

```bash
# 설치
pip install moai-adk==0.41.0
# 또는
uv pip install moai-adk==0.41.0

# 업그레이드
pip install --upgrade moai-adk
# 또는
uv pip install --upgrade moai-adk
```

---

# v0.40.2 - Test Isolation Fix (2026-01-08)

## Summary

Patch release fixing a critical bug in test suite where pytest would delete the real project `.moai` folder during test execution. This release improves test isolation using pytest's `tmp_path` and `monkeypatch` fixtures.

## Changes

### Bug Fixes

- **fix(tests)**: Prevent test_logger from deleting real .moai folder (e612d193)
  - Use tmp_path and monkeypatch.chdir for test isolation
  - Affected tests: test_setup_logger_default_log_dir, test_setup_logger_with_none_log_dir
  - Prevents accidental deletion of project configuration during testing

### Quality

- All tests passing (9,627 passed, 85.63% coverage)
- No linting or type checking issues
- Security checks passed

## Installation & Update

```bash
# Install
pip install moai-adk==0.40.2
# or
uv pip install moai-adk==0.40.2

# Upgrade
pip install --upgrade moai-adk
# or
uv pip install --upgrade moai-adk
```

---

# v0.40.2 - 테스트 격리 버그 수정 (2026-01-08)

## 요약

테스트 실행 중 실제 프로젝트 `.moai` 폴더가 삭제되는 심각한 버그를 수정한 패치 릴리스입니다. pytest의 `tmp_path`와 `monkeypatch` 픽스처를 사용하여 테스트 격리를 개선했습니다.

## 변경 사항

### 버그 수정

- **fix(tests)**: test_logger가 실제 .moai 폴더를 삭제하는 문제 수정 (e612d193)
  - tmp_path와 monkeypatch.chdir를 사용한 테스트 격리
  - 영향받은 테스트: test_setup_logger_default_log_dir, test_setup_logger_with_none_log_dir
  - 테스트 중 프로젝트 구성이 실수로 삭제되는 것을 방지

### 품질

- 모든 테스트 통과 (9,627개 통과, 커버리지 85.63%)
- 린트 및 타입 체크 이슈 없음
- 보안 체크 통과

## 설치 및 업데이트

```bash
# 설치
pip install moai-adk==0.40.2
# 또는
uv pip install moai-adk==0.40.2

# 업그레이드
pip install --upgrade moai-adk
# 또는
uv pip install --upgrade moai-adk
```

---

# v0.40.1 - Multilingual Agent Routing Enhancement (2026-01-07)

## Summary

Patch release enhancing Alfred's multilingual capabilities with comprehensive cross-lingual agent routing. Adds extensive keyword mappings across English, Korean, Japanese, and Chinese to enable seamless agent selection regardless of user's language preference.

## Changes

### New Features

- **feat**: Add multilingual agent routing with cross-lingual keyword mapping (409ecaf7)
  - English, Korean, Japanese, Chinese keyword support
  - 16 domain categories with comprehensive trigger patterns
  - Git operations, UI/UX design, quality gates, testing strategy
  - Project setup, implementation strategy, Claude Code configuration
  - Agent/command/skill/plugin creation workflows
  - Image generation with Nano-Banana AI integration
  - Cross-Lingual Thought (XLT) protocol for semantic bridging
  - Dynamic skill loading based on technology keywords

- **feat(agents)**: Add summary first line to all agent descriptions (24332d8f)
  - Enhanced agent discoverability and documentation
  - Applied to all 28 agents consistently

### Quality

- All tests passing (9,627 passed, 85.63% coverage)
- No linting or type checking issues
- Security checks passed

## Installation & Update

```bash
# Install
pip install moai-adk==0.40.1
# or
uv pip install moai-adk==0.40.1

# Upgrade
pip install --upgrade moai-adk
# or
uv pip install --upgrade moai-adk
```

---

# v0.40.1 - 다국어 에이전트 라우팅 개선 (2026-01-07)

## 요약

Alfred의 다국어 기능을 강화한 패치 릴리스입니다. 영어, 한국어, 일본어, 중국어 전반에 걸친 포괄적인 키워드 매핑을 통해 사용자의 언어 선호도와 관계없이 원활한 에이전트 선택을 가능하게 합니다.

## 변경 사항

### 새로운 기능

- **feat**: 다국어 에이전트 라우팅 및 교차 언어 키워드 매핑 추가 (409ecaf7)
  - 영어, 한국어, 일본어, 중국어 키워드 지원
  - 포괄적인 트리거 패턴을 갖춘 16개 도메인 카테고리
  - Git 작업, UI/UX 디자인, 품질 게이트, 테스트 전략
  - 프로젝트 설정, 구현 전략, Claude Code 구성
  - 에이전트/커맨드/스킬/플러그인 생성 워크플로우
  - Nano-Banana AI 통합을 통한 이미지 생성
  - 의미론적 브릿징을 위한 Cross-Lingual Thought (XLT) 프로토콜
  - 기술 키워드 기반 동적 스킬 로딩

- **feat(agents)**: 모든 에이전트 설명에 요약 첫 줄 추가 (24332d8f)
  - 향상된 에이전트 검색 가능성 및 문서화
  - 28개 모든 에이전트에 일관되게 적용

### 품질

- 모든 테스트 통과 (9,627개 통과, 85.63% 커버리지)
- 린팅 또는 타입 체킹 이슈 없음
- 보안 검사 통과

## 설치 및 업데이트

```bash
# 설치
pip install moai-adk==0.40.1
# 또는
uv pip install moai-adk==0.40.1

# 업그레이드
pip install --upgrade moai-adk
# 또는
uv pip install --upgrade moai-adk
```

---

# v0.40.0 - Large-Scale Module Optimization and Documentation Updates (2026-01-06)

## Summary

Major release featuring comprehensive module optimization, accurate documentation updates, standalone plugin mode, and enhanced skill library. This release significantly improves maintainability by removing 16,432 lines of obsolete code while adding critical features and fixing quality issues.

## Changes

### Documentation Updates

- **docs**: Update README files with accurate skill and agent counts (7cfd568c)
  - Skill count: 47 → 48
  - Agent count: 27 → 28 (consistent across all sections)
  - Added missing agents: expert-performance, expert-refactoring, expert-testing, builder-plugin
  - Updated tier counts and descriptions
  - Applied to all language versions (EN/KO/JA/ZH)

- **docs(readme)**: Update all README files with agent count and multilingual routing (7bcde13b)
- **docs(readme)**: Add multilingual agent routing feature documentation (bb31219b)
- **docs(config)**: Clarify TRUST 5 framework description and set 85% coverage default (26d08a36)

### New Features

- **feat(agents)**: Add Standalone plugin mode and testing section to builder-plugin (4ade0780)
  - Standalone mode for MoAI-independent plugins
  - Comprehensive testing section for plugin validation
  - Enhanced marketplace setup guidance

- **feat(hooks)**: Add PostToolUse/PreToolUse hooks and LSP config (245c26fa)
  - Enhanced hook system with pre/post tool execution
  - Language Server Protocol configuration support

- **feat(skills)**: Complete Context7 integration and module optimization (36652250)
- **feat(skills)**: Modularize Tier 2 language skills and add quality validator (7020a3f3)
- **feat(workflow-testing)**: Complete large-scale module optimization with progressive disclosure (7ae64f21)

### Refactoring and Optimization

- **refactor(skills)**: Large-scale module optimization and cleanup (117bf5d9)
  - Removed 16,432 lines of obsolete code
  - Improved module structure and organization
  - Enhanced maintainability and reduced complexity

- **refactor(hooks)**: Full hooks system refactoring with code consolidation and architecture improvements (f87009d0)
- **refactor(skills)**: Modularize all 7 platform skills with hybrid documentation pattern (fffb60fb)
- **refactor(skill)**: Modularize moai-platform-supabase with hybrid documentation pattern (123f1a4b)
- **refactor(templates)**: Comprehensive MoAI-ADK v4.0.0 template refactoring (c9dd6624)
- **refactor(hooks)**: Enhance path_utils with safe project root detection (321b5b39)

### Bug Fixes

- **fix(cli)**: Add reset_stdin() to fix interactive prompt after SpinnerContext (3045ab3d)
  - Fixed interactive prompt issues after spinner display
  - Improved terminal state management

- **fix**: Resolve quality issues and remove obsolete test files (5fe29d21)
- **fix(version)**: Sync all versions to 0.36.2 and add version management guidelines (5e02977e)
- **fix(hooks)**: Skip commits already on remote branches/tags (75d0b6f4)
- **fix**: Import bug in post_tool_auto_spec_completion.py (e4c34979)

### Code Cleanup

- **chore**: Remove dead code: Auto-Spec Completion System (3,515 lines) (3b0858b2)
- **chore**: Remove unused validate_skills.py utility script (e974f56c)
- **chore**: Remove local settings file from git tracking (fdeacf19)
- **chore**: Remove tracked backup file (now in .gitignore) (5f21c4c2)
- **chore**: Add .moai/config to version control and exclude backup files (6a0aa8cc)

### Code Quality

- **style**: Auto-fix lint and format issues (b8691da2)
  - Applied ruff formatter to 3 files
  - Improved code consistency

## Statistics

- **Total Changes**: 853 files changed
- **Code Changes**: +137,480 insertions, -69,084 deletions
- **Net Change**: +68,396 lines
- **Test Coverage**: 85.44% (9,627 tests passed)
- **Commits**: 26 commits since v0.36.2

## Installation & Update

```bash
# Install or update to v0.40.0
uv tool install moai-adk
# or
pip install --upgrade moai-adk

# Verify installation
moai --version
# Should show: 0.40.0
```

## What's Next

- v0.41.0: Enhanced agent coordination patterns
- v0.42.0: Advanced context management features
- v0.43.0: Performance optimization and benchmarking

---

# v0.40.0 - 대규모 모듈 최적화 및 문서 업데이트 (2026-01-06)

## 요약

포괄적인 모듈 최적화, 정확한 문서 업데이트, 독립형 플러그인 모드, 향상된 스킬 라이브러리를 포함하는 주요 릴리스입니다. 이 릴리스는 16,432줄의 obsolete 코드를 제거하면서 중요한 기능을 추가하고 품질 문제를 수정하여 유지보수성을 크게 향상시켰습니다.

## 변경 사항

### 문서 업데이트

- **docs**: 정확한 스킬 및 에이전트 수로 README 파일 업데이트 (7cfd568c)
  - 스킬 수: 47 → 48
  - 에이전트 수: 27 → 28 (모든 섹션에서 일관성 유지)
  - 누락된 에이전트 추가: expert-performance, expert-refactoring, expert-testing, builder-plugin
  - 계층 수 및 설명 업데이트
  - 모든 언어 버전에 적용 (EN/KO/JA/ZH)

- **docs(readme)**: 에이전트 수 및 다국어 라우팅으로 모든 README 파일 업데이트 (7bcde13b)
- **docs(readme)**: 다국어 에이전트 라우팅 기능 문서 추가 (bb31219b)
- **docs(config)**: TRUST 5 프레임워크 설명 명확화 및 85% 커버리지 기본값 설정 (26d08a36)

### 새로운 기능

- **feat(agents)**: builder-plugin에 독립형 플러그인 모드 및 테스팅 섹션 추가 (4ade0780)
  - MoAI 독립형 플러그인을 위한 Standalone 모드
  - 플러그인 검증을 위한 포괄적인 테스팅 섹션
  - 향상된 마켓플레이스 설정 가이드

- **feat(hooks)**: PostToolUse/PreToolUse 훅 및 LSP 설정 추가 (245c26fa)
  - 도구 실행 전후 훅 시스템 강화
  - Language Server Protocol 설정 지원

- **feat(skills)**: Context7 통합 완료 및 모듈 최적화 (36652250)
- **feat(skills)**: Tier 2 언어 스킬 모듈화 및 품질 검증기 추가 (7020a3f3)
- **feat(workflow-testing)**: 점진적 공개를 통한 대규모 모듈 최적화 완료 (7ae64f21)

### 리팩토링 및 최적화

- **refactor(skills)**: 대규모 모듈 최적화 및 정리 (117bf5d9)
  - 16,432줄의 obsolete 코드 제거
  - 모듈 구조 및 조직 개선
  - 유지보수성 향상 및 복잡도 감소

- **refactor(hooks)**: 코드 통합 및 아키텍처 개선을 통한 전체 훅 시스템 리팩토링 (f87009d0)
- **refactor(skills)**: 하이브리드 문서 패턴으로 모든 7개 플랫폼 스킬 모듈화 (fffb60fb)
- **refactor(skill)**: 하이브리드 문서 패턴으로 moai-platform-supabase 모듈화 (123f1a4b)
- **refactor(templates)**: 포괄적인 MoAI-ADK v4.0.0 템플릿 리팩토링 (c9dd6624)
- **refactor(hooks)**: 안전한 프로젝트 루트 감지로 path_utils 강화 (321b5b39)

### 버그 수정

- **fix(cli)**: SpinnerContext 이후 대화형 프롬프트 수정을 위한 reset_stdin() 추가 (3045ab3d)
  - 스피너 표시 후 대화형 프롬프트 문제 수정
  - 터미널 상태 관리 개선

- **fix**: 품질 문제 해결 및 obsolete 테스트 파일 제거 (5fe29d21)
- **fix(version)**: 모든 버전을 0.36.2로 동기화 및 버전 관리 가이드라인 추가 (5e02977e)
- **fix(hooks)**: 원격 브랜치/태그에 이미 있는 커밋 건너뛰기 (75d0b6f4)
- **fix**: post_tool_auto_spec_completion.py의 import 버그 (e4c34979)

### 코드 정리

- **chore**: 데드 코드 제거: Auto-Spec Completion System (3,515줄) (3b0858b2)
- **chore**: 미사용 validate_skills.py 유틸리티 스크립트 제거 (e974f56c)
- **chore**: git 추적에서 로컬 설정 파일 제거 (fdeacf19)
- **chore**: 추적된 백업 파일 제거 (이제 .gitignore에 포함) (5f21c4c2)
- **chore**: .moai/config를 버전 관리에 추가하고 백업 파일 제외 (6a0aa8cc)

### 코드 품질

- **style**: 린트 및 포맷 문제 자동 수정 (b8691da2)
  - 3개 파일에 ruff 포맷터 적용
  - 코드 일관성 개선

## 통계

- **총 변경**: 853개 파일 변경
- **코드 변경**: +137,480 삽입, -69,084 삭제
- **순 변경**: +68,396줄
- **테스트 커버리지**: 85.44% (9,627개 테스트 통과)
- **커밋**: v0.36.2 이후 26개 커밋

## 설치 및 업데이트

```bash
# v0.40.0 설치 또는 업데이트
uv tool install moai-adk
# 또는
pip install --upgrade moai-adk

# 설치 확인
moai --version
# 출력: 0.40.0
```

## 다음 계획

- v0.41.0: 향상된 에이전트 조정 패턴
- v0.42.0: 고급 컨텍스트 관리 기능
- v0.43.0: 성능 최적화 및 벤치마킹

---

# v0.36.2 - CLI Rename and Configuration System Improvements (2025-12-30)

## Summary

Patch release renaming the worktree CLI command for better user experience and migrating configuration system from monolithic JSON to modular YAML sections. This release improves usability and maintainability while fixing git hook issues.

## Changes

### CLI Improvements

- **refactor**: Rename CLI command from `moai-workflow-worktree` to `moai-worktree` (73c778de)
  - Shorter, more intuitive command name
  - Updated all documentation (English, Korean, Japanese, Chinese)
  - Updated pyproject.toml entry point
  - Updated all skill references and examples
  - Breaking Change: Users must reinstall package to use new command name

### Configuration System

- **refactor(config)**: Migrate from config.json to section YAML files (4f59c0d4)
  - Modular section-based configuration (user.yaml, language.yaml, project.yaml, etc.)
  - Improved token efficiency with on-demand loading
  - Enhanced configuration management and validation
  - Backward compatible migration system

### Bug Fixes

- **fix**: Restore .moai/config/ and clean up duplicate gitignore patterns (d736acdc)
  - Fixed missing .moai/config/ directory in distribution
  - Removed duplicate gitignore patterns

- **fix(hooks)**: Improve pre-push hook to skip already-pushed commits (069a0e5c, 9f44a754)
  - Only check commits not yet pushed to remote
  - Improved performance for large repositories
  - Better error messages and validation

## Breaking Changes

⚠️ **Important**: CLI command name changed

- **Old command**: `moai-workflow-worktree`
- **New command**: `moai-worktree`
- **Migration**: Reinstall package with `pip install --upgrade moai-adk`

All commands using the old name must be updated:

```bash
# Old
moai-workflow-worktree new SPEC-001

# New
moai-worktree new SPEC-001
```

## Installation & Update

```bash
# Update to latest version
pip install --upgrade moai-adk

# Or with uv
uv pip install --upgrade moai-adk

# Verify installation
moai --version  # Should show 0.36.2
moai-worktree --help  # Verify new CLI command works
```

## Quality Metrics

- Test Coverage: 85.99% (10,024 tests passed)
- Code Quality: All ruff and format checks passed
- Files Changed: 75 files (+3,083 / -1,910)

---

# v0.36.2 - CLI 이름 변경 및 설정 시스템 개선 (2025-12-30)

## 요약

더 나은 사용자 경험을 위해 worktree CLI 명령어 이름을 변경하고, 설정 시스템을 단일 JSON에서 모듈화된 YAML 섹션으로 마이그레이션한 패치 릴리스입니다. 이 릴리스는 사용성과 유지보수성을 개선하고 git hook 문제를 수정합니다.

## 변경 사항

### CLI 개선

- **리팩토링**: CLI 명령어를 `moai-workflow-worktree`에서 `moai-worktree`로 변경 (73c778de)
  - 더 짧고 직관적인 명령어 이름
  - 모든 문서 업데이트 (영어, 한국어, 일본어, 중국어)
  - pyproject.toml 진입점 업데이트
  - 모든 스킬 참조 및 예제 업데이트
  - Breaking Change: 새 명령어 이름을 사용하려면 패키지 재설치 필요

### 설정 시스템

- **리팩토링(config)**: config.json에서 섹션 YAML 파일로 마이그레이션 (4f59c0d4)
  - 모듈화된 섹션 기반 설정 (user.yaml, language.yaml, project.yaml 등)
  - 온디맨드 로딩으로 토큰 효율성 개선
  - 향상된 설정 관리 및 검증
  - 하위 호환 가능한 마이그레이션 시스템

### 버그 수정

- **수정**: .moai/config/ 복원 및 중복 gitignore 패턴 정리 (d736acdc)
  - 배포판에서 누락된 .moai/config/ 디렉토리 수정
  - 중복 gitignore 패턴 제거

- **수정(hooks)**: 이미 푸시된 커밋을 건너뛰도록 pre-push hook 개선 (069a0e5c, 9f44a754)
  - 아직 원격에 푸시되지 않은 커밋만 확인
  - 대규모 저장소의 성능 개선
  - 더 나은 오류 메시지 및 검증

## Breaking Changes

⚠️ **중요**: CLI 명령어 이름 변경

- **이전 명령어**: `moai-workflow-worktree`
- **새 명령어**: `moai-worktree`
- **마이그레이션**: `pip install --upgrade moai-adk`로 패키지 재설치

이전 이름을 사용하는 모든 명령어를 업데이트해야 합니다:

```bash
# 이전
moai-workflow-worktree new SPEC-001

# 이후
moai-worktree new SPEC-001
```

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
pip install --upgrade moai-adk

# 또는 uv 사용
uv pip install --upgrade moai-adk

# 설치 확인
moai --version  # 0.36.2 표시되어야 함
moai-worktree --help  # 새 CLI 명령어 작동 확인
```

## 품질 지표

- 테스트 커버리지: 85.99% (10,024개 테스트 통과)
- 코드 품질: 모든 ruff 및 format 체크 통과
- 변경된 파일: 75개 파일 (+3,083 / -1,910)

---
