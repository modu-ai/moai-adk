# v0.36.1 - Skill Library Consolidation (2025-12-30)

## Summary

Patch release consolidating and optimizing the skill library, reducing from 50 to 47 skills by removing virtual implementations and merging duplicate functionalities. This release improves maintainability and eliminates hallucination-prone virtual skills.

## Changes

### Skill Refactoring

- **refactor(skills)**: Consolidate and optimize skill library (576697e8)
  - Delete moai-mcp-figma (virtual implementation with non-existent module)
  - Delete moai-mcp-notion (virtual implementation with non-existent module)
  - Merge moai-security-auth0 into moai-platform-auth0 (consolidate 36 security modules)
  - Rename moai-worktree to moai-workflow-worktree (consistent naming convention)
  - Update skill count from 50 to 47 in all README files (English, Korean, Japanese, Chinese)
  - Reduce skill file sizes for better maintainability:
    - Ruby: 688 → 424 lines (-38%)
    - C++: 650 → 422 lines (-35%)
    - PHP: 645 → 496 lines (-23%)
    - Elixir: 613 → 386 lines (-37%)
    - R: 580 → 381 lines (-34%)
  - Update technology version references:
    - Rust: 1.91 → 1.92 (latest stable)
    - Mermaid.js: 10.x → 11.12.2 (latest stable)
  - Update all agent references (expert-security, manager-docs, mcp-figma, mcp-notion)
  - Update command references (1-plan, 3-sync)
  - Consolidate MCP section to AI Integration section in documentation
  - Total changes: 165 files, 10,669 insertions(+), 12,727 deletions(-)

### Breaking Changes

⚠️ **Important**: The following changes may affect existing workflows:

- **moai-mcp-figma** skill removed
  - Migration: Use `moai-domain-uiux` skill with `mcp-figma` agent
  - Reason: Virtual implementation without actual module support

- **moai-mcp-notion** skill removed
  - Migration: Use `moai-workflow-project` skill with `mcp-notion` agent
  - Reason: Virtual implementation without actual module support

- **moai-security-auth0** renamed to **moai-platform-auth0**
  - Migration: Update skill references to `moai-platform-auth0`
  - Reason: Consolidate security and platform features into single comprehensive skill

- **moai-worktree** renamed to **moai-workflow-worktree**
  - Migration: Update skill references to `moai-workflow-worktree`
  - Reason: Consistent naming convention (all workflow skills use `moai-workflow-*` prefix)

## Installation & Update

### Fresh Install (uv tool - Recommended)
```bash
uv tool install moai-adk
```

### Update Existing Installation
```bash
uv tool update moai-adk
```

### Alternative Methods
- **pip**: `pip install moai-adk`
- **pipx**: `pipx install moai-adk`
- **uv (global)**: `uv pip install moai-adk`

## Quality Metrics

- ✅ All 10,037 tests passing (180 skipped, 26 xfailed, 36 xpassed)
- ✅ Test coverage: 86.90% (above 85% threshold)
- ✅ Ruff checks: All checks passed
- ✅ Ruff format: 191 files unchanged
- ✅ Mypy: No issues found in 146 source files

---

# v0.36.0 - JavaScript Skill & Merge Analyzer Refactoring (2025-12-29)

## Summary

Minor release adding comprehensive JavaScript/TypeScript skill with modern tooling support and refactoring the merge analyzer to use pure Python instead of Claude headless dependency.

## Changes

### New Features

- **feat(skills)**: Add JavaScript/TypeScript skill (317aea4c)
  - Add comprehensive `moai-lang-javascript` skill with ES2024+ support
  - Include Node.js 22 LTS, modern runtimes (Deno, Bun)
  - Add testing frameworks (Vitest, Jest), linting (ESLint 9, Biome)
  - Support backend frameworks (Express, Fastify, Hono)
  - Provide 973 lines of practical examples
  - Include 695 lines of reference documentation
  - Location: `src/moai_adk/templates/.claude/skills/moai-lang-javascript/`

### Refactoring

- **refactor(merge)**: Replace Claude headless with Pure Python analyzer (651e4dc7)
  - Migrate merge analyzer from Claude API dependency to pure Python implementation
  - Improve performance by eliminating external API calls
  - Maintain same functionality with better reliability
  - Update 944 lines in analyzer implementation
  - Update 545 lines in test coverage
  - Location: `src/moai_adk/core/merge/analyzer.py`, `tests/unit/core/test_merge_analyzer_cov.py`

- **refactor(skills)**: Cleanup deprecated skill modules (317aea4c)
  - Remove deprecated worktree management modules (701 lines)
  - Remove deprecated integration patterns (982 lines)
  - Remove deprecated parallel development guide (778 lines)
  - Remove deprecated worktree commands (782 lines)
  - Remove root-level SKILL.md, examples.md, reference.md
  - Streamline skill structure and reduce technical debt
  - Location: `.claude/skills/modules/`

### Other Changes

- **style**: Auto-fix lint and format issues (335dc7d6)
  - Apply ruff format to codebase
  - Clean up configuration backups
  - Update .gitignore patterns
  - Location: Multiple files

## Installation & Update

### Fresh Install (uv tool - Recommended)
```bash
uv tool install moai-adk
```

### Update Existing Installation
```bash
uv tool update moai-adk
```

### Alternative Methods
- **pip**: `pip install moai-adk`
- **pipx**: `pipx install moai-adk`
- **uv (global)**: `uv pip install moai-adk`

## Breaking Changes

None. This is a backward-compatible minor release.

## Quality Metrics

- ✅ All 10,037 tests passing
- ✅ Test coverage: 86.92% (above 85% threshold)
- ✅ Ruff checks: Passing
- ✅ Mypy: No issues found

---

# v0.35.2 - Plugin Documentation Update & YAML Parser Fix (2025-12-26)

## Summary

Patch release updating plugin documentation with official Claude Code standards and fixing a critical YAML parser bug in session startup hook that caused empty string values to be incorrectly parsed as empty dictionaries.

## Changes

### New Features

- **feat(docs)**: Update plugin documentation with official standards (01bb8c36)
  - Update `builder-plugin.md` agent to v1.1.0 with complete hook events and LSP options
  - Add PostToolUseFailure, SubagentStart, Notification, PreCompact hook events
  - Add agent hook type alongside command and prompt types
  - Document 12 LSP server advanced options (transport, initializationOptions, settings, etc.)
  - Add Plugin Caching and Security section with installation scopes
  - Add managed installation scope for enterprise deployments
  - Update `moai-plugin-builder` skill to v1.1.0 with synchronized changes
  - Location: `src/moai_adk/templates/.claude/agents/moai/builder-plugin.md`, `src/moai_adk/templates/.claude/skills/moai-plugin-builder/`

- **feat(docs)**: Cleanup CLAUDE.md and remove outdated sections (01bb8c36)
  - Remove MCP Integration and External Services section
  - Fix Chinese character "推测" to "speculative" in documentation
  - Update Context7 references to use WebSearch/WebFetch tools
  - Remove unverified percentage claim (40-60% → significantly)
  - Update version to 8.5.0
  - Location: `src/moai_adk/templates/CLAUDE.md`, `CLAUDE.md`

### Bug Fixes

- **fix(hooks)**: Fix empty string YAML values incorrectly parsed as empty dict (ad949745)
  - Add was_quoted flag to _simple_yaml_parse function
  - Properly handle quoted empty strings like `name: ""` in YAML
  - Prevent empty strings from being stored as empty dicts `{}`
  - Fix welcome message bug showing dict values instead of empty strings
  - Location: `.claude/hooks/moai/session_start__show_project_info.py`, `src/moai_adk/templates/.claude/hooks/moai/session_start__show_project_info.py`

## Installation & Update

### Fresh Install (uv tool - Recommended)
```bash
uv tool install moai-adk
```

### Update Existing Installation
```bash
uv tool update moai-adk
```

### Alternative Methods
```bash
# Using uvx (no install needed)
uvx moai-adk --help

# Using pip
pip install --upgrade moai-adk
```

---

# v0.35.1 - Update Error Detection Improvement (2025-12-25)

## Summary

Patch release fixing the empty "Claude execution error" message when Claude Code fails silently during `moai-adk update`. Enhanced error detection with exit codes, stdout fallback, and new error patterns for better debugging.

## Changes

### Bug Fixes

- **fix(merge)**: Fix empty error message when Claude Code fails silently (6f8b3f2e)
  - Extend `_detect_claude_errors()` method with `returncode` and `stdout` parameters
  - Include exit code in error messages for better diagnostics
  - Add stdout hint extraction when stderr is empty
  - Add new error patterns: API key, rate limit, network errors
  - Add WARNING level logging for Claude Code failures
  - Update ANALYZED_FILES to use config.yaml (JSON→YAML migration)
  - Extend error message truncation from 200 to 300 characters
  - Location: `src/moai_adk/core/merge/analyzer.py`

- **test(merge)**: Add 3 new tests for enhanced error detection (6f8b3f2e)
  - Test empty stderr with helpful message
  - Test empty stderr with returncode context
  - Test stdout hint extraction
  - Location: `tests/unit/core/test_merge_analyzer_cov.py`

### Maintenance

- **chore(ci)**: Extract bilingual release notes from CHANGELOG.md (0ed395d2)
  - Simplify release workflow with CHANGELOG-based notes
  - Location: `.github/workflows/release.yml`

## Installation & Update

### Fresh Install (uv tool - Recommended)
```bash
uv tool install moai-adk
```

### Update Existing Installation
```bash
uv tool update moai-adk
```

### Alternative Methods
```bash
# Using uvx (no install needed)
uvx moai-adk --help

# Using pip
pip install moai-adk==0.35.1
```

## Quality Metrics

- Test Coverage: 86.72% (target: 85%)
- Tests Passed: 10,039 passed, 180 skipped, 26 xfailed
- CI/CD: All quality gates passing

## Breaking Changes

None

## Migration Guide

No migration required. Update will apply automatically.

---

# v0.35.1 - 업데이트 에러 감지 개선 (2025-12-25)

## 요약

`moai-adk update` 실행 중 Claude Code가 조용히 실패할 때 빈 "Claude execution error" 메시지를 수정하는 패치 릴리즈입니다. 더 나은 디버깅을 위해 exit 코드, stdout 폴백, 새로운 에러 패턴이 포함된 향상된 에러 감지 기능을 제공합니다.

## 변경 사항

### 버그 수정

- **fix(merge)**: Claude Code가 조용히 실패할 때 빈 에러 메시지 수정 (6f8b3f2e)
  - `_detect_claude_errors()` 메서드를 `returncode` 및 `stdout` 매개변수로 확장
  - 더 나은 진단을 위해 에러 메시지에 exit 코드 포함
  - stderr가 비어있을 때 stdout 힌트 추출 추가
  - 새로운 에러 패턴 추가: API 키, rate limit, 네트워크 에러
  - Claude Code 실패에 대한 WARNING 레벨 로깅 추가
  - JSON→YAML 마이그레이션을 위해 ANALYZED_FILES를 config.yaml로 업데이트
  - 에러 메시지 잘림을 200에서 300자로 확장
  - 위치: `src/moai_adk/core/merge/analyzer.py`

- **test(merge)**: 향상된 에러 감지를 위한 3개의 새 테스트 추가 (6f8b3f2e)
  - 도움이 되는 메시지가 있는 빈 stderr 테스트
  - returncode 컨텍스트가 있는 빈 stderr 테스트
  - stdout 힌트 추출 테스트
  - 위치: `tests/unit/core/test_merge_analyzer_cov.py`

### 유지보수

- **chore(ci)**: CHANGELOG.md에서 이중 언어 릴리즈 노트 추출 (0ed395d2)
  - CHANGELOG 기반 노트로 릴리즈 워크플로우 간소화
  - 위치: `.github/workflows/release.yml`

## 설치 및 업데이트

### 신규 설치 (uv tool - 권장)
```bash
uv tool install moai-adk
```

### 기존 설치 업데이트
```bash
uv tool upgrade moai-adk
```

### 대체 방법
```bash
# uvx 사용 (설치 없이)
uvx moai-adk --help

# pip 사용
pip install moai-adk==0.35.1
```

## 품질 지표

- 테스트 커버리지: 86.72% (목표: 85%)
- 테스트 통과: 10,039 통과, 180 스킵, 26 xfailed
- CI/CD: 모든 품질 게이트 통과

## 중대 변경사항

없음

## 마이그레이션 가이드

마이그레이션 불필요. 업데이트 시 자동 적용됩니다.

---

# v0.35.0 - Security Skills & Image Generation (2025-12-25)

## Summary

Minor release adding comprehensive Auth0 security skill, image generation capabilities, improved git workflows, and plugin builder documentation. Includes configuration system cleanup and enhanced MCP integration.

## Changes

### New Features

- **feat(skills)**: Enhanced moai-platform-auth0 with comprehensive security modules (e4853270)
  - 36 comprehensive security modules covering MFA, attack protection, and compliance
  - Multi-factor authentication (WebAuthn, TOTP, SMS, Email, Push)
  - Attack protection (brute force, bot detection, breached passwords)
  - Compliance frameworks (GDPR, FAPI, Highly Regulated Identity)
  - Sender constraining (mTLS, DPoP) and continuous session protection
  - Location: `.claude/skills/moai-platform-auth0/`

- **feat(nano-banana)**: Add image generation scripts (216a36a7)
  - `generate_image.py` - Single image generation with Gemini 3 Pro
  - `batch_generate.py` - Batch image generation with parallel processing
  - Support for aspect ratios, safety settings, and error handling
  - Comprehensive test coverage (1,590+ lines)
  - Location: `.claude/skills/moai-ai-nano-banana/scripts/`

- **feat(git-workflow)**: Add main_direct and main_feature workflow options (f2a6e438)
  - `main_direct` - Work directly on main branch (single-developer workflow)
  - `main_feature` - Feature branches merged to main (team workflow)
  - Enhanced workflow configuration in project setup
  - Location: `.moai/config/questions/tab3-git.yaml`

- **feat(plugin-builder)**: Add comprehensive plugin builder skill
  - Plugin architecture documentation and validation guides
  - Migration patterns from loose files to organized plugins
  - 2,600+ lines of plugin development documentation
  - Location: `.claude/skills/moai-plugin-builder/`

### Bug Fixes

- **fix(hooks)**: Properly parse quoted YAML values with inline comments (da392e8b)
  - Fix git strategy parsing for workflow rules validation
  - Location: `.claude/hooks/moai/session_start__show_project_info.py`

- **fix(hooks)**: Prevent false positives in pre-push security check (567118fd)
  - Improve secret detection patterns
  - Location: `src/moai_adk/templates/.git-hooks/pre-push`

- **fix(tests)**: Use explicit initial_branch in temp_repo fixture (026ca759)
  - Ensure consistent test behavior across git versions

- **fix(tests)**: Update worktree tests for project_name parameter (e9cf5ccc)
  - Fix test compatibility with updated worktree API

- **fix(worktree)**: Add type annotation to fix mypy errors (003a9f68)
  - Improve type safety in worktree modules

### Maintenance

- **chore(mcp)**: Simplify MCP server configuration (174689fe)
  - Streamlined .mcp.json structure
  - Improved server registration patterns

- **chore(templates)**: Sync config.yaml version to 0.34.0 (1c53caa8)
  - Update template versioning

- **chore**: Remove session state from git tracking (9a7b0665)
  - Clean up .moai/memory/last-session-state.json from version control

- **refactor(tests)**: Relocate nano-banana skill tests to package test directory (c1a45def)
  - Organized test structure: `tests/skills/nano-banana/`

- **style(worktree)**: Apply ruff format to worktree modules (52e7a1a5)
  - Consistent code formatting

- **style**: Auto-fix lint and format issues (a81fdaae)
  - Pre-release code cleanup

### Configuration System Cleanup

Removed old monolithic config files in favor of modular sections:
- Deleted `.moai/config/config.yaml` (replaced with `sections/*.yaml`)
- Removed legacy SPEC-SYNC-QUALITY-001 artifacts
- Cleaner project initialization workflow

## Installation & Update

### Fresh Install (uv tool - Recommended)
```bash
uv tool install moai-adk
```

### Update Existing Installation
```bash
uv tool update moai-adk
```

### Alternative Methods
```bash
# Using uvx (no install needed)
uvx moai-adk --help

# Using pip
pip install moai-adk==0.35.0
```

## Quality Metrics

- Test Coverage: 86.78% (target: 85%)
- Tests Passed: 10,037 passed, 180 skipped, 26 xfailed
- CI/CD: All quality gates passing

## Breaking Changes

None - all changes are additive or internal improvements.

## Migration Guide

No migration required. New skills and features are available immediately after upgrade.

To use new features:
- Security guidance: Load `Skill("moai-platform-auth0")`
- Image generation: Use scripts in `moai-ai-nano-banana` skill
- Git workflows: Configure via `moai-adk init` or update `.moai/config/sections/git-strategy.yaml`

---

# v0.35.0 - 보안 스킬 및 이미지 생성 (2025-12-25)

## 요약

Auth0 보안 스킬, 이미지 생성 기능, 개선된 git 워크플로우, 플러그인 빌더 문서를 추가한 마이너 릴리즈입니다. 설정 시스템 정리 및 향상된 MCP 통합이 포함되어 있습니다.

## 변경 사항

### 신규 기능

- **feat(skills)**: moai-platform-auth0 스킬에 포괄적인 보안 모듈 강화 (e4853270)
  - MFA, 공격 방어, 컴플라이언스를 다루는 36개의 포괄적인 보안 모듈
  - 다중 인증 (WebAuthn, TOTP, SMS, Email, Push)
  - 공격 방어 (무차별 대입 공격, 봇 탐지, 침해된 비밀번호)
  - 컴플라이언스 프레임워크 (GDPR, FAPI, 고도 규제 신원)
  - 발신자 제약 (mTLS, DPoP) 및 지속적 세션 보호
  - 위치: `.claude/skills/moai-platform-auth0/`

- **feat(nano-banana)**: 이미지 생성 스크립트 추가 (216a36a7)
  - `generate_image.py` - Gemini 3 Pro를 사용한 단일 이미지 생성
  - `batch_generate.py` - 병렬 처리를 통한 배치 이미지 생성
  - 종횡비, 안전 설정, 오류 처리 지원
  - 포괄적인 테스트 커버리지 (1,590+ 줄)
  - 위치: `.claude/skills/moai-ai-nano-banana/scripts/`

- **feat(git-workflow)**: main_direct 및 main_feature 워크플로우 옵션 추가 (f2a6e438)
  - `main_direct` - main 브랜치에서 직접 작업 (단일 개발자 워크플로우)
  - `main_feature` - main으로 병합되는 기능 브랜치 (팀 워크플로우)
  - 프로젝트 설정에서 향상된 워크플로우 구성
  - 위치: `.moai/config/questions/tab3-git.yaml`

- **feat(plugin-builder)**: 포괄적인 플러그인 빌더 스킬 추가
  - 플러그인 아키텍처 문서 및 검증 가이드
  - 느슨한 파일에서 조직화된 플러그인으로의 마이그레이션 패턴
  - 2,600+ 줄의 플러그인 개발 문서
  - 위치: `.claude/skills/moai-plugin-builder/`

### 버그 수정

- **fix(hooks)**: 인라인 주석이 있는 따옴표로 묶인 YAML 값 적절히 파싱 (da392e8b)
  - 워크플로우 규칙 검증을 위한 git 전략 파싱 수정
  - 위치: `.claude/hooks/moai/session_start__show_project_info.py`

- **fix(hooks)**: pre-push 보안 검사에서 오탐지 방지 (567118fd)
  - 비밀 탐지 패턴 개선
  - 위치: `src/moai_adk/templates/.git-hooks/pre-push`

- **fix(tests)**: temp_repo 픽스처에서 명시적 initial_branch 사용 (026ca759)
  - git 버전 간 일관된 테스트 동작 보장

- **fix(tests)**: project_name 매개변수에 대한 worktree 테스트 업데이트 (e9cf5ccc)
  - 업데이트된 worktree API와 테스트 호환성 수정

- **fix(worktree)**: mypy 오류 수정을 위한 타입 주석 추가 (003a9f68)
  - worktree 모듈의 타입 안전성 개선

### 유지보수

- **chore(mcp)**: MCP 서버 구성 단순화 (174689fe)
  - .mcp.json 구조 간소화
  - 서버 등록 패턴 개선

- **chore(templates)**: config.yaml 버전을 0.34.0으로 동기화 (1c53caa8)
  - 템플릿 버전 관리 업데이트

- **chore**: git 추적에서 세션 상태 제거 (9a7b0665)
  - 버전 관리에서 .moai/memory/last-session-state.json 정리

- **refactor(tests)**: nano-banana 스킬 테스트를 패키지 테스트 디렉토리로 재배치 (c1a45def)
  - 조직화된 테스트 구조: `tests/skills/nano-banana/`

- **style(worktree)**: worktree 모듈에 ruff 포맷 적용 (52e7a1a5)
  - 일관된 코드 포맷팅

- **style**: 린트 및 포맷 이슈 자동 수정 (a81fdaae)
  - 릴리즈 전 코드 정리

### 설정 시스템 정리

모듈식 섹션을 위해 기존의 모놀리식 설정 파일 제거:
- `.moai/config/config.yaml` 삭제 (`sections/*.yaml`로 대체)
- 레거시 SPEC-SYNC-QUALITY-001 아티팩트 제거
- 깔끔한 프로젝트 초기화 워크플로우

## 설치 및 업데이트

### 신규 설치 (uv tool - 권장)
```bash
uv tool install moai-adk
```

### 기존 설치 업데이트
```bash
uv tool upgrade moai-adk
```

### 대체 방법
```bash
# uvx 사용 (설치 없이)
uvx moai-adk --help

# pip 사용
pip install moai-adk==0.35.0
```

## 품질 지표

- 테스트 커버리지: 86.78% (목표: 85%)
- 테스트 통과: 10,037개 통과, 180개 건너뜀, 26개 예상 실패
- CI/CD: 모든 품질 게이트 통과

## 중대 변경사항

없음 - 모든 변경사항은 추가 기능 또는 내부 개선입니다.

## 마이그레이션 가이드

마이그레이션 불필요. 업그레이드 후 즉시 새로운 스킬 및 기능 사용 가능.

새 기능 사용 방법:
- 보안 가이드: `Skill("moai-platform-auth0")` 로드
- 이미지 생성: `moai-ai-nano-banana` 스킬의 스크립트 사용
- Git 워크플로우: `moai-adk init`를 통해 구성하거나 `.moai/config/sections/git-strategy.yaml` 업데이트

---

# v0.34.1 - Windows Compatibility & UX Improvements (2025-12-25)

## Summary

Patch release improving Windows compatibility for Claude Code detection and statusline rendering, plus UX improvements for AskUserQuestion configuration prompts.

## Changes

### Bug Fixes

- **fix(windows)**: Improve Claude Code executable detection on Windows
  - Add `_find_claude_executable()` method with comprehensive path search
  - Search npm global directory (`%APPDATA%\npm\claude.cmd`)
  - Search Local AppData installation paths
  - Use `shutil.which()` with Windows fallback paths
  - Location: `src/moai_adk/core/merge/analyzer.py`

- **fix(windows)**: Fix statusline command for Windows compatibility
  - Add `{{STATUSLINE_COMMAND}}` template variable
  - Windows: Use `python -m moai_adk statusline` for better PATH compatibility
  - Unix: Use `moai-adk statusline` directly
  - Location: `src/moai_adk/core/project/phase_executor.py`, `src/moai_adk/cli/commands/update.py`

- **fix(ux)**: Improve AskUserQuestion prompts for text input
  - Replace confusing "Other" option with clear "Type something..." guidance
  - Remove deprecated `{{prompt_user}}` placeholder usage
  - Add preset options (4 max) with custom input field guidance
  - Location: `.moai/config/questions/` YAML files

### Maintenance

- **test**: Update `test_build_claude_command_structure` for new executable path format
- **style**: Fix unused variable warnings in batch_generate.py

## Installation & Update

### Fresh Install (uv tool - Recommended)
```bash
uv tool install moai-adk
```

### Update Existing Installation
```bash
uv tool update moai-adk
```

### Alternative Methods
```bash
# Using uvx (no install needed)
uvx moai-adk --help

# Using pip
pip install moai-adk==0.34.1
```

## Quality Metrics

- Test Coverage: 86.78% (target: 85%)
- Tests Passed: 10,037 passed, 180 skipped, 26 xfailed
- CI/CD: All quality gates passing

## Breaking Changes

None

## Migration Guide

Windows users should run `moai-adk update` after upgrading to apply the new statusline command format.

---

# v0.34.1 - Windows 호환성 및 UX 개선 (2025-12-25)

## 요약

Windows에서 Claude Code 감지 및 statusline 렌더링 호환성을 개선하고, AskUserQuestion 설정 프롬프트의 UX를 개선한 패치 릴리즈입니다.

## 변경 사항

### 버그 수정

- **fix(windows)**: Windows에서 Claude Code 실행 파일 감지 개선
  - 포괄적인 경로 검색을 포함한 `_find_claude_executable()` 메서드 추가
  - npm 전역 디렉토리 검색 (`%APPDATA%\npm\claude.cmd`)
  - Local AppData 설치 경로 검색
  - Windows 폴백 경로와 함께 `shutil.which()` 사용
  - 위치: `src/moai_adk/core/merge/analyzer.py`

- **fix(windows)**: Windows 호환성을 위한 statusline 명령어 수정
  - `{{STATUSLINE_COMMAND}}` 템플릿 변수 추가
  - Windows: PATH 호환성을 위해 `python -m moai_adk statusline` 사용
  - Unix: `moai-adk statusline` 직접 사용
  - 위치: `src/moai_adk/core/project/phase_executor.py`, `src/moai_adk/cli/commands/update.py`

- **fix(ux)**: 텍스트 입력을 위한 AskUserQuestion 프롬프트 개선
  - 혼란스러운 "Other" 옵션을 명확한 "Type something..." 안내로 대체
  - 더 이상 사용되지 않는 `{{prompt_user}}` 플레이스홀더 사용 제거
  - 커스텀 입력 필드 안내와 함께 프리셋 옵션 추가 (최대 4개)
  - 위치: `.moai/config/questions/` YAML 파일

### 유지보수

- **test**: 새로운 실행 파일 경로 형식에 맞게 `test_build_claude_command_structure` 업데이트
- **style**: batch_generate.py의 사용되지 않는 변수 경고 수정

## 설치 및 업데이트

### 신규 설치 (uv tool - 권장)
```bash
uv tool install moai-adk
```

### 기존 설치 업데이트
```bash
uv tool upgrade moai-adk
```

### 대체 방법
```bash
# uvx 사용 (설치 없이)
uvx moai-adk --help

# pip 사용
pip install moai-adk==0.34.1
```

## 품질 메트릭

- 테스트 커버리지: 86.78% (목표: 85%)
- 테스트 통과: 10,037 통과, 180 스킵, 26 xfailed
- CI/CD: 모든 품질 게이트 통과

## 호환성 변경

없음

## 마이그레이션 가이드

Windows 사용자는 업그레이드 후 `moai-adk update`를 실행하여 새로운 statusline 명령어 형식을 적용해야 합니다.

---

