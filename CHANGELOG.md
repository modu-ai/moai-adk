# v1.3.8 - Critical Bug Fixes for Hooks and ESC Key (2026-01-18)

## Summary

This patch release resolves critical issues reported in GitHub issues #265, #266, #267, and #268, including hooks schema incompatibility with Claude Code, ESC key terminal freeze, and DDD documentation clarity for Greenfield projects.

## Fixed

- **fix(hooks)**: Resolve SessionStart/SessionEnd hook format incompatibility (#265, #266)
  - Fixed SessionEnd hook schema to use flat structure (no matcher, no nested hooks array)
  - SessionStart/SessionEnd hooks now correctly use `{"type": "command", "command": "..."}` format
  - PreToolUse/PostToolUse continue to use matcher-based nested structure
  - Resolves "Expected array, but received undefined" validation errors in Claude Code
  - File: `src/moai_adk/rank/hook.py` (function: `_register_hook_in_settings()`)

- **fix(ui)**: Improve ESC key handling to prevent terminal freeze (#268)
  - Added `curses.set_escdelay(25)` to reduce ESC key response delay from 1000ms to 25ms
  - Prevents terminal input freeze when ESC key is pressed during migration UI
  - Enhanced terminal compatibility across different environments
  - File: `src/moai_adk/core/migration/interactive_checkbox_ui.py`

- **fix(docs)**: Clarify DDD support for Greenfield projects (#267)
  - Updated DDD documentation to explicitly include Greenfield project support
  - DDD adapts its cycle for new projects: ANALYZE (requirements) → PRESERVE (test-first) → IMPROVE (implement)
  - DDD is now documented as a superset of TDD, supporting both refactoring and new development
  - Removed "Greenfield projects (use TDD instead)" from exclusions
  - Files: `.claude/skills/moai-workflow-ddd/SKILL.md`, `src/moai_adk/templates/.claude/skills/moai-workflow-ddd/SKILL.md`

- **feat(database)**: Add Oracle support to moai-domain-database skill
  - Comprehensive Oracle Database support documentation
  - Covers Oracle-specific features, performance optimization, and security best practices
  - File: `.claude/skills/moai-domain-database/modules/oracle.md`

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Hook integration tests: 59 passed (100% pass rate)
- Rank tests: 197 passed (100% pass rate)
- Migration UI tests: 380 passed (100% pass rate)
- Ruff: All checks passed
- Mypy: Success (no issues found)

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version
moai --version
```

---

# v1.3.8 - Hooks 및 ESC 키 중요 버그 수정 (2026-01-18)

## 요약

이 패치 릴리스는 GitHub 이슈 #265, #266, #267, #268에서 보고된 중요한 문제들을 해결합니다. Claude Code와의 hooks 스키마 비호환성, ESC 키 터미널 프리즈, Greenfield 프로젝트를 위한 DDD 문서 명확화가 포함됩니다.

## 수정됨

- **fix(hooks)**: SessionStart/SessionEnd hook 형식 비호환성 해결 (#265, #266)
  - SessionEnd hook 스키마를 flat 구조로 수정 (matcher 없음, 중첩 hooks 배열 없음)
  - SessionStart/SessionEnd hook이 이제 `{"type": "command", "command": "..."}` 형식을 올바르게 사용
  - PreToolUse/PostToolUse는 계속 matcher 기반 중첩 구조 사용
  - Claude Code에서 "Expected array, but received undefined" 유효성 검사 오류 해결
  - 파일: `src/moai_adk/rank/hook.py` (함수: `_register_hook_in_settings()`)

- **fix(ui)**: 터미널 프리즈 방지를 위한 ESC 키 처리 개선 (#268)
  - ESC 키 응답 지연을 1000ms에서 25ms로 줄이기 위해 `curses.set_escdelay(25)` 추가
  - migration UI에서 ESC 키 누를 때 터미널 입력 프리즈 방지
  - 다양한 환경에서 향상된 터미널 호환성
  - 파일: `src/moai_adk/core/migration/interactive_checkbox_ui.py`

- **fix(docs)**: Greenfield 프로젝트를 위한 DDD 지원 명확화 (#267)
  - DDD 문서를 업데이트하여 Greenfield 프로젝트 지원을 명시적으로 포함
  - DDD가 새 프로젝트를 위해 사이클 적응: ANALYZE (요구사항) → PRESERVE (test-first) → IMPROVE (구현)
  - DDD가 이제 TDD의 상위 집합으로 문서화되어 리팩토링과 새 개발 모두 지원
  - 제외 목록에서 "Greenfield 프로젝트 (TDD 사용)" 제거
  - 파일: `.claude/skills/moai-workflow-ddd/SKILL.md`, `src/moai_adk/templates/.claude/skills/moai-workflow-ddd/SKILL.md`

- **feat(database)**: moai-domain-database 스킬에 Oracle 지원 추가
  - 포괄적인 Oracle Database 지원 문서
  - Oracle 특정 기능, 성능 최적화, 보안 모범 사례 포함
  - 파일: `.claude/skills/moai-domain-database/modules/oracle.md`

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Hook 통합 테스트: 59개 통과 (100% 통과율)
- Rank 테스트: 197개 통과 (100% 통과율)
- Migration UI 테스트: 380개 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 성공 (문제 없음)

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인
moai --version
```

---

# v1.3.7 - DDD Terminology Complete Migration (2026-01-17)

## Summary

This patch release completes the TDD to DDD (Domain-Driven Development) terminology migration across all test files and documentation.

## Changed

- **refactor(git)**: Complete TDD to DDD terminology migration in git module
  - Renamed `TDDCommitPhase` to `DDDCommitPhase`
  - Updated phase names: RED→ANALYZE, GREEN→PRESERVE, REFACTOR→IMPROVE
  - Updated `format_tdd_commit` to `format_ddd_commit`
  - Updated `TDD_PHASES` to `DDD_PHASES`
- **test(git)**: Updated all git tests with DDD terminology
  - `tests/foundation/test_git.py`
  - `tests/unit/foundation/test_git.py`
- **docs(release)**: Added config version files to release checklist
  - All 4 version files must be updated: pyproject.toml, version.py, config.yaml, system.yaml

## Quality

- All 73 git tests: PASSED
- Smoke tests: 6 passed (100% pass rate)
- Ruff: All checks passed

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version
moai --version
```

---

# v1.3.7 - DDD 용어 완전 마이그레이션 (2026-01-17)

## 요약

이 패치 릴리스는 모든 테스트 파일과 문서에서 TDD에서 DDD(도메인 주도 개발) 용어로의 마이그레이션을 완료합니다.

## 변경됨

- **refactor(git)**: git 모듈의 TDD to DDD 용어 마이그레이션 완료
  - `TDDCommitPhase`를 `DDDCommitPhase`로 이름 변경
  - 단계 이름 업데이트: RED→ANALYZE, GREEN→PRESERVE, REFACTOR→IMPROVE
  - `format_tdd_commit`을 `format_ddd_commit`으로 업데이트
  - `TDD_PHASES`를 `DDD_PHASES`로 업데이트
- **test(git)**: 모든 git 테스트를 DDD 용어로 업데이트
  - `tests/foundation/test_git.py`
  - `tests/unit/foundation/test_git.py`
- **docs(release)**: 릴리스 체크리스트에 config 버전 파일 추가
  - 4개 버전 파일 모두 업데이트 필요: pyproject.toml, version.py, config.yaml, system.yaml

## 품질

- 모든 73개 git 테스트: 통과
- Smoke 테스트: 6개 통과 (100% 통과율)
- Ruff: 모든 검사 통과

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인
moai --version
```

---

# v1.3.6 - MoAI Rank Session Sync Fix (2026-01-17)

## Summary

This patch release fixes a critical bug in MoAI Rank session synchronization where sessions with extremely high `cacheReadTokens` values (exceeding ~2.1 billion) were failing server-side validation.

## Fixed

- **fix(rank)**: Cap cacheReadTokens to INT32_MAX to prevent server validation error (#264)
  - Sessions with `cache_read_tokens` exceeding 2,147,483,647 (INT32_MAX) now have values capped before submission
  - Prevents `VALIDATION_ERROR: Cache read tokens exceed limit` errors from MoAI Rank server
  - Affects heavy Claude Code users with long sessions that accumulate massive prompt caching tokens
  - Client-side workaround until server-side INT type constraint is upgraded
  - Normal values (< INT32_MAX) are preserved unchanged
  - Added 4 comprehensive test cases for capping logic validation

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Rank tests: 114 passed (100% pass rate, +4 new tests)
- Ruff: All checks passed
- Mypy: No issues found in 169 source files

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version
moai --version
```

---

# v1.3.6 - MoAI Rank 세션 동기화 수정 (2026-01-17)

## 요약

이 패치 릴리스는 `cacheReadTokens` 값이 극단적으로 높은 세션(약 21억 초과)이 서버 측 유효성 검사에 실패하는 MoAI Rank 세션 동기화의 중요한 버그를 수정합니다.

## 수정됨

- **fix(rank)**: 서버 유효성 검사 오류 방지를 위해 cacheReadTokens를 INT32_MAX로 제한 (#264)
  - `cache_read_tokens`가 2,147,483,647(INT32_MAX)을 초과하는 세션의 값을 제출 전에 캡핑
  - MoAI Rank 서버의 `VALIDATION_ERROR: Cache read tokens exceed limit` 오류 방지
  - 긴 세션에서 대량의 프롬프트 캐싱 토큰을 축적하는 헤비 Claude Code 사용자에게 영향
  - 서버 측 INT 타입 제약이 업그레이드될 때까지의 클라이언트 측 임시 해결책
  - 정상 값(< INT32_MAX)은 변경 없이 보존됨
  - 캡핑 로직 검증을 위한 4개의 포괄적인 테스트 케이스 추가

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Rank 테스트: 114개 통과 (100% 통과율, +4개 신규 테스트)
- Ruff: 모든 검사 통과
- Mypy: 169개 소스 파일에서 문제 없음

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인
moai --version
```

---

# v1.3.5 - Worktree CLI Enhancement & DDD Migration (2026-01-17)

## Summary

This patch release improves the worktree CLI with cross-project search capabilities and better error handling, while completing the TDD→DDD naming migration in test files.

## Fixed

- **fix(worktree)**: Enhance CLI with cross-project search and improved error messages
  - `moai-wt go` now searches across all projects if worktree not found in current project
  - `moai-wt remove` applies the same cross-project search improvement
  - Added path existence validation with helpful recovery suggestions
  - Added `assert` type guard for `spec_id` None safety in `sync_worktree`
  - Fixed `cleaned` variable initialization in `clean_worktrees` to prevent mypy errors
  - Improved error messages showing available worktrees when not found

- **fix(tests)**: Update TDD→DDD naming in test files
  - Renamed `TDDCommitPhase` → `DDDCommitPhase`
  - Renamed `TDD_PHASES` → `DDD_PHASES`
  - Renamed `format_tdd_commit` → `format_ddd_commit`
  - Updated test files: `tests/foundation/test_git.py`, `tests/unit/foundation/test_git.py`
  - Fixed worktree test to create directory for path.exists() validation

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Worktree tests: 131 passed (100% pass rate)
- Ruff: All checks passed
- Mypy: No issues found

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version
moai --version
```

---

# v1.3.5 - Worktree CLI 개선 및 DDD 마이그레이션 (2026-01-17)

## 요약

이 패치 릴리스는 worktree CLI를 전 프로젝트 검색 기능과 향상된 오류 처리로 개선하며, 테스트 파일의 TDD→DDD 네이밍 마이그레이션을 완료합니다.

## 수정됨

- **fix(worktree)**: 전 프로젝트 검색 및 향상된 오류 메시지로 CLI 개선
  - `moai-wt go`가 현재 프로젝트에서 worktree를 찾지 못하면 모든 프로젝트에서 검색
  - `moai-wt remove`에 동일한 전 프로젝트 검색 개선 적용
  - 유용한 복구 제안과 함께 경로 존재 여부 검증 추가
  - `sync_worktree`에서 `spec_id` None 안전성을 위한 `assert` 타입 가드 추가
  - mypy 오류 방지를 위해 `clean_worktrees`의 `cleaned` 변수 초기화 수정
  - worktree를 찾지 못할 때 사용 가능한 worktree를 표시하는 향상된 오류 메시지

- **fix(tests)**: 테스트 파일에서 TDD→DDD 네이밍 업데이트
  - `TDDCommitPhase` → `DDDCommitPhase`로 이름 변경
  - `TDD_PHASES` → `DDD_PHASES`로 이름 변경
  - `format_tdd_commit` → `format_ddd_commit`로 이름 변경
  - 테스트 파일 업데이트: `tests/foundation/test_git.py`, `tests/unit/foundation/test_git.py`
  - path.exists() 검증을 위해 디렉토리를 생성하도록 worktree 테스트 수정

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Worktree 테스트: 131개 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 문제 없음

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인
moai --version
```

---

# v1.3.2 - Hooks Format Fix (2026-01-17)

## Summary

This patch release fixes the Claude Code hooks format by adding required `matcher: {}` field to SessionStart and SessionEnd hooks.

## Fixed

- **fix(hooks)**: Add required `matcher: {}` field to SessionStart and SessionEnd hooks
  - Claude Code now requires `matcher` field even for hooks without conditions
  - Fixed "Expected array, but received undefined" error in settings validation
  - Updated all settings.json templates (unix/windows variants)

- **fix(tests)**: Fix mock function signature in test_switch.py
  - Added missing `encoding` parameter to mock_read_text_func

---

# v1.3.2 - Hooks 형식 수정 (2026-01-17)

## 요약

이 패치 릴리스는 SessionStart 및 SessionEnd hooks에 필수 `matcher: {}` 필드를 추가하여 Claude Code hooks 형식을 수정합니다.

## 수정됨

- **fix(hooks)**: SessionStart 및 SessionEnd hooks에 필수 `matcher: {}` 필드 추가
  - Claude Code가 이제 조건이 없는 hooks에도 `matcher` 필드를 요구함
  - settings 유효성 검사에서 "Expected array, but received undefined" 오류 수정
  - 모든 settings.json 템플릿(unix/windows 변형) 업데이트

- **fix(tests)**: test_switch.py의 mock 함수 시그니처 수정
  - mock_read_text_func에 누락된 `encoding` 매개변수 추가

---

# v1.3.1 - Bugfix Release (2026-01-17)

## Summary

This patch release fixes the SessionStart hook error display issue (#263) and removes hardcoded user settings from templates.

## Fixed

- **fix(hooks)**: Correct SessionStart/SessionEnd hook structure (#263)
  - Fixed nested "hooks" array causing "hook error" display despite successful execution
  - Simplified hook configuration format for Claude Code compatibility

- **fix(templates)**: Remove hardcoded user name from user.yaml template
  - Template now defaults to empty string for proper user personalization
  - Users can set their own name in `.moai/config/sections/user.yaml`

---

# v1.3.1 - 버그 수정 릴리스 (2026-01-17)

## 요약

이 패치 릴리스는 SessionStart 훅 오류 표시 문제 (#263)를 수정하고 템플릿에서 하드코딩된 사용자 설정을 제거합니다.

## 수정됨

- **fix(hooks)**: SessionStart/SessionEnd 훅 구조 수정 (#263)
  - 성공적인 실행에도 불구하고 "hook error"가 표시되는 중첩된 "hooks" 배열 수정
  - Claude Code 호환성을 위한 훅 설정 형식 단순화

- **fix(templates)**: user.yaml 템플릿에서 하드코딩된 사용자 이름 제거
  - 템플릿이 이제 적절한 사용자 개인화를 위해 빈 문자열을 기본값으로 사용
  - 사용자는 `.moai/config/sections/user.yaml`에서 자신의 이름을 설정할 수 있습니다

---



# v1.3.0 - DDD-Only Methodology & Progressive Disclosure (2026-01-17)

## Summary

This release marks a major architectural shift to Domain-Driven Development (DDD) methodology exclusively, removing the TAG System entirely. It introduces Progressive Disclosure for 67% token reduction in skill loading.

## Changes

### Breaking Changes

- **refactor(methodology)**: Complete removal of TAG System v2.0
  - Shift to DDD-only methodology (Domain-Driven Development)
  - Removed all TAG-related code, hooks, and validation
  - Removed manager-tdd agent (superseded by manager-ddd)
  - Updated quality.yaml: `constitution.development_mode: ddd` (TDD removed)
  - Updated CLAUDE.md to v10.3.0 (DDD Only + Progressive Disclosure)

### Added

- **feat(progressive-disclosure)**: Implement 3-level skill loading system
  - Level 1: Metadata only (~100 tokens per skill)
  - Level 2: Full skill body (~5K tokens per skill)
  - Level 3+: On-demand bundled files
  - 67% reduction in initial token load (from ~90K to ~600 tokens for manager-spec)
  - Automatic trigger-based loading via keywords, phases, agents, languages
  - Integration with JIT Context Loader

- **feat(builder)**: Enhance agent/skill/command builder definitions
  - Improved builder-agent, builder-skill, builder-command
  - Enhanced creation workflows
  - Better template generation

### Changed

- **refactor(terminology)**: Rename DDR to DDD (Domain-Driven Development)
  - Updated all references from DDR (Domain-Driven Refactoring) to DDD
  - Aligned terminology with industry standards
  - Updated documentation across EN/KO/JA/ZH README files

- **chore(templates)**: Sync CLAUDE.md to templates (v10.2.0 → v10.3.0)
  - Updated Alfred execution directives
  - Added Progressive Disclosure documentation
  - Removed TAG System references

- **chore(agents)**: Update agent frontmatter with skills format
  - 18 agents updated with `skills:` format for Progressive Disclosure
  - Improved agent-to-skill loading efficiency

### Removed

- Removed TAG System v2.0 completely:
  - `src/moai_adk/tag_system/` directory
  - `.claude/hooks/moai/lib/tag_*.py` files
  - `.claude/hooks/moai/pre_commit__tag_validator.py`
  - `.claude/hooks/moai/post_tool__coverage_guard.py`
  - `.claude/hooks/moai/pre_tool__tdd_enforcer.py`
  - `.claude/agents/moai/manager-tdd.md`
  - All TAG-related tests
  - TAG mock translations from test files

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version
moai --version
```

## Migration Guide

### Breaking Changes

- **TAG System users**: TAG System has been completely removed. Projects using TAG validation should migrate to DDD methodology.
- **TDD workflow users**: Use `/moai:2-run` with DDD methodology instead of deprecated TDD mode.
- **manager-tdd users**: Use `manager-ddd` agent for implementation tasks.

### What's New

- All skills now support Progressive Disclosure (automatic, no action needed)
- CLAUDE.md updated to v10.3.0 (templates auto-sync on update)

## Quality

- 6 smoke/critical tests passing (100% pass rate)
- Ruff: All checks passed
- Mypy: No issues found in 169 source files
- 273 files changed (+7,938 additions, -2,620 deletions)

## Documentation

- Updated CLAUDE.md to v10.3.0 (DDD Only + Progressive Disclosure)
- Updated README files (EN/KO/JA/ZH) with DDD terminology
- Added Progressive Disclosure documentation

---

# v1.3.0 - DDD 전용 방법론 및 Progressive Disclosure (2026-01-17)

## 요약

이 릴리스는 Domain-Driven Development (DDD) 방법론으로 전환하는 주요 아키텍처 변경을 표시하며, TAG 시스템을 완전히 제거합니다. 스킬 로딩에서 67% 토큰 절감을 위한 Progressive Disclosure를 도입합니다.

## 변경 사항

### Breaking Changes

- **refactor(methodology)**: TAG 시스템 v2.0 완전 제거
  - DDD 전용 방법론으로 전환 (Domain-Driven Development)
  - 모든 TAG 관련 코드, 훅, 검증 제거
  - manager-tdd 에이전트 제거 (manager-ddd로 대체)
  - quality.yaml 업데이트: `constitution.development_mode: ddd` (TDD 제거)
  - CLAUDE.md를 v10.3.0으로 업데이트 (DDD Only + Progressive Disclosure)

### 추가됨

- **feat(progressive-disclosure)**: 3-level 스킬 로딩 시스템 구현
  - Level 1: 메타데이터만 (스킬당 ~100 토큰)
  - Level 2: 전체 스킬 본문 (스킬당 ~5K 토큰)
  - Level 3+: 온디맨드 번들 파일
  - 초기 토큰 로드 67% 절감 (manager-spec의 경우 ~90K → ~600 토큰)
  - 키워드, 단계, 에이전트, 언어를 통한 자동 트리거 기반 로딩
  - JIT Context Loader와 통합

- **feat(builder)**: 에이전트/스킬/명령어 빌더 정의 강화
  - builder-agent, builder-skill, builder-command 개선
  - 향상된 생성 워크플로우
  - 개선된 템플릿 생성

### 변경됨

- **refactor(terminology)**: DDR을 DDD로 이름 변경 (Domain-Driven Development)
  - 모든 참조를 DDR (Domain-Driven Refactoring)에서 DDD로 업데이트
  - 업계 표준에 맞춘 용어
  - EN/KO/JA/ZH README 파일의 문서 업데이트

- **chore(templates)**: CLAUDE.md를 템플릿으로 동기화 (v10.2.0 → v10.3.0)
  - Alfred 실행 지시문 업데이트
  - Progressive Disclosure 문서 추가
  - TAG 시스템 참조 제거

- **chore(agents)**: 에이전트 frontmatter를 skills 형식으로 업데이트
  - Progressive Disclosure를 위해 18개 에이전트를 `skills:` 형식으로 업데이트
  - 에이전트-스킬 로딩 효율성 개선

### 제거됨

- TAG 시스템 v2.0 완전 제거:
  - `src/moai_adk/tag_system/` 디렉토리
  - `.claude/hooks/moai/lib/tag_*.py` 파일들
  - `.claude/hooks/moai/pre_commit__tag_validator.py`
  - `.claude/hooks/moai/post_tool__coverage_guard.py`
  - `.claude/hooks/moai/pre_tool__tdd_enforcer.py`
  - `.claude/agents/moai/manager-tdd.md`
  - 모든 TAG 관련 테스트
  - 테스트 파일의 TAG mock 번역

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인
moai --version
```

## 마이그레이션 가이드

### Breaking Changes

- **TAG 시스템 사용자**: TAG 시스템이 완전히 제거되었습니다. TAG 검증을 사용하는 프로젝트는 DDD 방법론으로 마이그레이션해야 합니다.
- **TDD 워크플로우 사용자**: deprecated된 TDD 모드 대신 DDD 방법론으로 `/moai:2-run`을 사용하세요.
- **manager-tdd 사용자**: 구현 작업에 `manager-ddd` 에이전트를 사용하세요.

### 새로운 기능

- 모든 스킬이 이제 Progressive Disclosure를 지원합니다 (자동, 조치 불필요)
- CLAUDE.md가 v10.3.0으로 업데이트되었습니다 (업데이트 시 템플릿 자동 동기화)

## 품질

- 6개 smoke/critical 테스트 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 169개 소스 파일에서 문제 없음
- 273개 파일 변경 (+7,938 추가, -2,620 삭제)

## 문서

- CLAUDE.md를 v10.3.0으로 업데이트 (DDD Only + Progressive Disclosure)
- DDD 용어로 README 파일 업데이트 (EN/KO/JA/ZH)
- Progressive Disclosure 문서 추가

---

# v1.2.0 - Platform-Specific Templates & Enhanced Skills (2025-01-15)

## Summary

This release introduces platform-specific settings.json templates to resolve Windows hook compatibility issues, along with enhanced frontend and UI/UX skills including Vercel React Best Practices and Web Interface Guidelines.

## Changes

### Fixed

- **fix(windows)**: Use relative paths with backslash for Windows hooks
  - Windows now uses `.\.claude\hooks\...` (relative paths with backslash)
  - Unix/macOS continues using `$CLAUDE_PROJECT_DIR/.claude/hooks/...` (environment variables)
  - Resolves hook execution failures on Windows due to Claude Code not expanding `%CLAUDE_PROJECT_DIR%`
  - Automatic platform detection ensures correct template selection

- **feat(platform)**: Add platform-specific settings.json templates
  - Separate templates for Windows (`settings.json.windows`) and Unix (`settings.json.unix`)
  - Template processor automatically selects appropriate file based on OS
  - Eliminates cross-platform path separator issues

### Added

- **feat(skills)**: Vercel React Best Practices module
  - 45 rules across 8 categories from Vercel Engineering
  - Covers async patterns, bundle optimization, server/client performance
  - 1,131+ lines of detailed guidance
  - Added to `moai-domain-frontend/modules/vercel-react-best-practices.md`

- **feat(skills)**: Web Interface Guidelines module
  - Comprehensive web interface guidelines from Vercel Labs
  - Covers HTML, accessibility, forms, animation, typography, performance
  - 687+ lines of comprehensive guidelines
  - Added to `moai-domain-uiux/modules/web-interface-guidelines.md`

### Changed

- Updated system configuration templates
- Improved template processor with platform-aware file selection

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Apply platform-specific templates to your project
moai-adk update
```

## Notes for Windows Users

Windows users should update and run `moai-adk update` to apply the new Windows-specific template with relative paths. This resolves hook execution issues caused by Claude Code not expanding environment variables on Windows.

---

# v1.2.0 - 플랫폼별 템플릿 및 향상된 스킬 (2025-01-15)

## 요약

이 릴리스는 Windows 훅 호환성 문제를 해결하기 위한 플랫폼별 settings.json 템플릿을 도입하며, Vercel React Best Practices와 Web Interface Guidelines를 포함한 향상된 프론트엔드 및 UI/UX 스킬을 제공합니다.

## 변경 사항

### 수정됨

- **fix(windows)**: Windows 훅을 위한 상대 경로 및 백슬래시 사용
  - Windows는 이제 `.\.claude\hooks\...` (백슬래시가 포함된 상대 경로) 사용
  - Unix/macOS는 계속 `$CLAUDE_PROJECT_DIR/.claude/hooks/...` (환경 변수) 사용
  - Claude Code가 Windows에서 `%CLAUDE_PROJECT_DIR%` 확장하지 않는 문제로 인한 훅 실행 실패 해결
  - 자동 플랫폼 감지로 올바른 템플릿 선택 보장

- **feat(platform)**: 플랫폼별 settings.json 템플릿 추가
  - Windows용 (`settings.json.windows`)과 Unix용 (`settings.json.unix`) 별도 템플릿
  - 템플릿 프로세서가 OS에 따라 적절한 파일 자동 선택
  - 크로스 플랫폼 경로 구분자 문제 해결

### 추가됨

- **feat(skills)**: Vercel React Best Practices 모듈
  - Vercel Engineering의 8개 카테고리 45개 규칙
  - 비동기 패턴, 번들 최적화, 서버/클라이언트 성능 포괄
  - 1,131줄 이상의 상세 가이드
  - `moai-domain-frontend/modules/vercel-react-best-practices.md`에 추가

- **feat(skills)**: Web Interface Guidelines 모듈
  - Vercel Labs의 종합 웹 인터페이스 가이드라인
  - HTML, 접근성, 폼, 애니메이션, 타이포그래피, 성능 포괄
  - 687줄 이상의 종합 가이드라인
  - `moai-domain-uiux/modules/web-interface-guidelines.md`에 추가

### 변경됨

- 시스템 구성 템플릿 업데이트
- 플랫폼 인식 파일 선택을 위한 템플릿 프로세서 개선

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트에 플랫폼별 템플릿 적용
moai-adk update
```

## Windows 사용자 참고사항

Windows 사용자는 업데이트 후 `moai-adk update`를 실행하여 상대 경로를 사용하는 새 Windows용 템플릿을 적용해야 합니다. 이것은 Claude Code가 Windows에서 환경 변수를 확장하지 않는 문제로 인한 훅 실행 문제를 해결합니다.

---

# v1.2.0 - Enhanced Planning Experience & Bug Fixes (2025-01-15)

## Summary

This release introduces an enhanced PHASE 0 planning experience with x.com style interview format (v7.0.0) and includes important bug fixes for hook library syntax errors.

## Changes

### Added

- **feat(plan)**: Enhance PHASE 0 with x.com style interview (v7.0.0)
  - Modern interview-style requirements gathering
  - Improved user interaction flow
  - Enhanced documentation for planning workflow

### Fixed

- **fix(hooks)**: Correct syntax errors in hook library files
  - Fixed syntax issues in config_validator.py
  - Fixed syntax issues in git_operations_manager.py
  - Fixed syntax issues in timeout.py and unified_timeout_manager.py
  - Updated documentation for affected hooks

### Security

- Added new security guard hook for enhanced tool validation

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk
```

---

# v1.2.0 - 향상된 계획 수립 경험 및 버그 수정 (2025-01-15)

## 요약

이 릴리스는 x.com 스타일 인터뷰 형식(v7.0.0)을 도입한 향상된 PHASE 0 계획 수립 경험을 제공하며, 훅 라이브러리 문법 오류에 대한 중요한 버그 수정을 포함합니다.

## 변경 사항

### 추가됨

- **feat(plan)**: x.com 스타일 인터뷰로 PHASE 0 강화 (v7.0.0)
  - 현대적인 인터뷰 스타일 요구사항 수집
  - 개선된 사용자 상호작용 흐름
  - 계획 수립 워크플로우를 위한 향상된 문서

### 수정됨

- **fix(hooks)**: 훅 라이브러리 파일의 문법 오류 수정
  - config_validator.py 문법 이슈 수정
  - git_operations_manager.py 문법 이슈 수정
  - timeout.py 및 unified_timeout_manager.py 문법 이슈 수정
  - 영향받는 훅의 문서 업데이트

### 보안

- 향상된 도구 검증을 위한 새로운 보안 가드 훅 추가

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk
```

---

# v1.1.0 - Comprehensive Enhancement: Backup System, TAG System v2.0, Performance & Quality (2026-01-13)

## Summary

Major feature release introducing comprehensive backup system improvements with metadata tracking and automatic cleanup, TAG System v2.0 for flexible validation, 85%+ test coverage achievement, and critical performance optimizations including parallel execution as default. This release also includes important bug fixes for Windows emoji encoding and statusline UTF-8 support.

## Added

### Backup System Enhancement

- **feat(backup)**: Improve backup system with metadata and auto-cleanup
  - Add `backup_metadata.json` to track backup contents and excluded items
  - Unify backup exclusion paths across all modules (specs, reports, project, config/sections)
  - Add `list_backups()` and `cleanup_old_backups()` methods to `TemplateBackup`
  - Integrate automatic backup cleanup in `moai update` command (keep last 5 backups)
  - Add comprehensive tests for new backup functionality (43 tests)

### TAG System v2.0

- **feat(tag-system)**: Implement TAG System v2.0 with flexible validation
  - Flexible validation engine supporting multi-language tags
  - Pre-commit tag validation hook for quality assurance
  - Multi-language tag support (KO, JA, ZH, EN)
  - Improved tag parsing and linkage validation
  - Integration with `moai init` and `moai update` workflows

### Performance & Workflow

- **feat(workflow)**: Parallel execution is now the default mode (#255)
  - All independent tasks execute in parallel by default
  - Add `--sequential` option to opt-out and run tasks sequentially when needed
  - Significantly improves workflow performance for multi-step operations
  - Better utilization of system resources
  - Related PR: #255

- **feat(worktree)**: Add `moai-wt done` command for streamlined workflow completion
  - One-command completion: checkout main → merge branch → remove worktree
  - Optional `--push` flag to push merged changes to remote
  - Automatic feature branch cleanup after merge
  - Simplifies Phase 3 (Merge and Cleanup) workflow

### Test Coverage

- **test(coverage)**: Achieve 85%+ coverage across core modules
  - Add comprehensive TDD tests for `init.py` (88.12% coverage)
  - Add integration tests for core commands
  - Add edge case tests for statusline and config modules
  - Add model allocator tests with comprehensive coverage

### Statusline & CLI

- **feat(statusline)**: Add comprehensive statusline enhancements
  - UTF-8 encoding support for international characters
  - Enable memory and directory display
  - Enhanced output style detection
  - Improved multilingual support

- **feat(cli)**: Add multilingual prompt translations
  - KO/JA/ZH prompt translations for init workflow
  - Improved localization support

## Fixed

### Hooks & Performance

- **fix(hooks)**: Prevent session_start hook from hanging on slow git operations (#254)
  - Fixed blocking issue where slow git commands caused startup delays
  - Improved timeout handling for git operations
  - Enhanced reliability of session initialization
  - Related PR: #254

- **fix(hooks)**: Run session_end hook in background to prevent exit delays
  - Session exit now completes instantly without waiting for cleanup
  - Background processing for auto-cleanup and rank submission
  - Eliminates ~3 second delay when closing Claude Code sessions

### CLI & Encoding

- **fix(cli)**: Fix Windows emoji encoding error (#256)
  - Resolve emoji display issues on Windows platforms
  - Ensure consistent emoji rendering across platforms

- **fix(statusline)**: Add UTF-8 encoding support
  - Fix encoding issues in statusline display
  - Ensure proper character encoding for multilingual content

- **fix(commands)**: Align tool permissions with CLAUDE.md Command Types policy
  - Ensured all commands follow documented permission policies
  - Improved security and consistency across command execution
  - Better alignment with Type A, Type B, and Type C command classifications

### Type Safety

- **fix(types)**: Resolve mypy type checking errors
  - Fix type annotations across core modules
  - Improve type safety and IDE support

### Git Worktree

- **fix(gitignore)**: Add llm-configs to tracked directories
  - Fix `.gitignore` configuration to properly track `llm-configs/` directory in git worktrees
  - Resolve issue where `moai glm` command failed in worktree environments

## Changed

### Configuration

- **chore(config)**: Sync template and local configurations
  - Synchronize `.moai/config/` with latest templates
  - Update `system.yaml`, `quality.yaml`, `language.yaml` configurations
  - Add new configuration options for TAG system and coverage targets

### Documentation

- **docs(tag)**: Add TAG system activation step to installation wizard
  - Document TAG system setup process
  - Add TAG system usage examples

- **docs(project)**: Sync project documentation with TAG System v2.0
  - Update `product.md`, `structure.md`, `tech.md`
  - Document new TAG system features

- **docs(readme)**: Add Step 2 session sync to MoAI Rank guide (EN/JA/ZH)
  - Document session synchronization workflow

- **docs(release)**: Improve CHANGELOG generation guide
  - Update CHANGELOG generation process

## Installation & Update

```bash
# Install
uv tool install moai-adk
pip install moai-adk==1.1.0

# Update existing installation
uv tool update moai-adk
pip install --upgrade moai-adk

# Verify version
moai --version
```

## Migration Guide

No breaking changes. Existing workflows will automatically benefit from:

- Automatic backup cleanup (keeps last 5 backups)
- Enhanced backup metadata for better tracking
- Parallel execution as default (use `--sequential` to opt-out)
- TAG System v2.0 validation (opt-in via configuration)

## Quality

- All 43 new backup tests passing (100% pass rate)
- 85%+ test coverage achieved for core modules
- Comprehensive integration test suite added
- Type safety verified through mypy

## Documentation

- Updated CHANGELOG generation guide
- Added TAG system documentation
- Updated multilingual README files (KO, JA, ZH)
- Added TESTING_GUIDE.md for contributors

---

# v1.1.0 - 포괄적 개선: 백업 시스템, TAG 시스템 v2.0, 성능 및 품질 (2026-01-13)

## 요약

백업 시스템의 메타데이터 추적 및 자동 정리 기능이 포함된 포괄적인 개선과 TAG 시스템 v2.0의 유연한 검증 기능을 도입한 주요 기능 릴리스입니다. 또한 85%+ 테스트 커버리지 목표를 달성했으며, 병렬 실행을 기본값으로 하는 중요한 성능 최적화가 포함되어 있습니다. Windows 이모지 인코딩 및 statusline UTF-8 지원에 대한 중요한 버그 수정도 포함되어 있습니다.

## 추가됨

### 백업 시스템 개선

- **feat(backup)**: 백업 시스템 개선 (메타데이터 및 자동 정리)
  - 백업 내용 및 제외 항목 추적을 위한 `backup_metadata.json` 추가
  - 모든 모듈에서 백업 제외 경로 통일 (specs, reports, project, config/sections)
  - `TemplateBackup`에 `list_backups()` 및 `cleanup_old_backups()` 메서드 추가
  - `moai update` 명령어에 자동 백업 정리 통합 (최근 5개 백업 유지)
  - 새로운 백업 기능에 대한 포괄적인 테스트 추가 (43개 테스트)

### TAG 시스템 v2.0

- **feat(tag-system)**: TAG 시스템 v2.0 구현 (유연한 검증)
  - 다국어 태그 지원 유연한 검증 엔진
  - 품질 보증을 위한 pre-commit 태그 검증 훅
  - 다국어 태그 지원 (KO, JA, ZH, EN)
  - 향상된 태그 파싱 및 연결 검증
  - `moai init` 및 `moai update` 워크플로우와의 통합

### 성능 및 워크플로우

- **feat(workflow)**: 병렬 실행이 이제 기본 모드입니다 (#255)
  - 모든 독립적인 작업이 기본적으로 병렬로 실행됩니다
  - 필요할 때 순차 실행을 위한 `--sequential` 옵션 추가
  - 다단계 작업의 워크플로우 성능 대폭 향상
  - 시스템 리소스 활용 개선
  - 관련 PR: #255

- **feat(worktree)**: 워크플로우 완료를 위한 `moai-wt done` 명령어 추가
  - 한 번의 명령으로 완료: checkout main → 브랜치 병합 → worktree 제거
  - 병합된 변경사항을 원격에 푸시하는 `--push` 옵션
  - 병합 후 자동 feature 브랜치 정리
  - Phase 3 (병합 및 정리) 워크플로우 간소화

### 테스트 커버리지

- **test(coverage)**: 핵심 모듈에서 85%+ 커버리지 달성
  - `init.py`를 위한 포괄적인 TDD 테스트 (88.12% 커버리지)
  - 핵심 명령어에 대한 통합 테스트 추가
  - statusline 및 config 모듈에 대한 엣지 케이스 테스트
  - 포괄적인 커버리지를 갖는 모델 할당자 테스트

### Statusline 및 CLI

- **feat(statusline)**: 포괄적인 statusline 개선
  - 국제 문자를 위한 UTF-8 인코딩 지원
  - 메모리 및 디렉토리 표시 활성화
  - 향상된 출력 스타일 감지
  - 개선된 다국어 지원

- **feat(cli)**: 다국어 프롬프트 번역 추가
  - init 워크플로우를 위한 KO/JA/ZH 프롬프트 번역
  - 개선된 현지화 지원

## 수정됨

### 훅 및 성능

- **fix(hooks)**: 느린 git 작업으로 인한 session_start hook hang 방지 (#254)
  - 느린 git 명령어가 시작 지연을 유발하는 블로킹 문제 수정
  - git 작업에 대한 타임아웃 처리 개선
  - 세션 초기화의 안정성 향상
  - 관련 PR: #254

- **fix(hooks)**: 종료 지연 방지를 위해 session_end hook을 백그라운드에서 실행
  - 세션 종료가 정리 작업을 기다리지 않고 즉시 완료됨
  - auto-cleanup 및 rank 제출의 백그라운드 처리
  - Claude Code 세션 종료 시 ~3초 지연 제거

### CLI 및 인코딩

- **fix(cli)**: Windows 이모지 인코딩 오류 수정 (#256)
  - Windows 플랫폼에서의 이모지 표시 문제 해결
  - 모든 플랫폼에서 일관된 이모지 렌더링 보장

- **fix(statusline)**: UTF-8 인코딩 지원 추가
  - statusline 표시의 인코딩 문제 수정
  - 다국어 콘텐츠를 위한 적절한 문자 인코딩 보장

- **fix(commands)**: CLAUDE.md Command Types 정책에 맞춰 도구 권한 정렬
  - 모든 명령어가 문서화된 권한 정책을 따르도록 보장
  - 명령어 실행 전반의 보안 및 일관성 개선
  - Type A, Type B, Type C 명령어 분류와의 더 나은 정렬

### 타입 안전성

- **fix(types)**: mypy 타입 검사 오류 해결
  - 핵심 모듈의 타입 어노테이션 수정
  - 타입 안전성 및 IDE 지원 개선

### Git Worktree

- **fix(gitignore)**: llm-configs를 추적 디렉토리에 추가
  - git worktree에서 `llm-configs/` 디렉토리를 올바르게 추적하도록 `.gitignore` 구성 수정
  - worktree 환경에서 `moai glm` 명령어 실패 문제 해결

## 변경됨

### 구성

- **chore(config)**: 템플릿 및 로컬 구성 동기화
  - 최신 템플릿과 `.moai/config/` 동기화
  - `system.yaml`, `quality.yaml`, `language.yaml` 구성 업데이트
  - TAG 시스템 및 커버리지 목표를 위한 새로운 구성 옵션 추가

### 문서

- **docs(tag)**: 설치 마법사에 TAG 시스템 활성화 단계 추가
  - TAG 시스템 설정 프로세스 문서화
  - TAG 시스템 사용 예제 추가

- **docs(project)**: TAG 시스템 v2.0으로 프로젝트 문서 동기화
  - `product.md`, `structure.md`, `tech.md` 업데이트
  - 새로운 TAG 시스템 기능 문서화

- **docs(readme)**: MoAI Rank 가이드에 Step 2 세션 동기화 추가 (EN/JA/ZH)
  - 세션 동기화 워크플로우 문서화

- **docs(release)**: CHANGELOG 생성 가이드 개선
  - CHANGELOG 생성 프로세스 업데이트

## 설치 및 업데이트

```bash
# 설치
uv tool install moai-adk
pip install moai-adk==1.1.0

# 기존 설치 업데이트
uv tool update moai-adk
pip install --upgrade moai-adk

# 버전 확인
moai --version
```

## 마이그레이션 가이드

Breaking change 없음. 기존 워크플로우는 다음 기능의 혜택을 자동으로 받습니다:

- 자동 백업 정리 (최근 5개 백업 유지)
- 향상된 추적을 위한 백업 메타데이터
- 병렬 실행 기본값 (순차 실행 필요 시 `--sequential` 사용)
- TAG 시스템 v2.0 검증 (구성을 통해 opt-in)

## 품질

- 43개의 새로운 백업 테스트 모두 통과 (100% 통과율)
- 핵심 모듈에서 85%+ 테스트 커버리지 달성
- 포괄적인 통합 테스트 스위트 추가
- mypy를 통한 타입 안전성 검증

## 문서

- CHANGELOG 생성 가이드 업데이트
- TAG 시스템 문서 추가
- 다국어 README 파일 업데이트 (KO, JA, ZH)
- 기여자를 위한 TESTING_GUIDE.md 추가

---

# v1.0.0 - Production Ready Release (2026-01-12)

## Summary

Initial production-ready release of MoAI-ADK, featuring SPEC-First TDD workflow with Alfred SuperAgent, unified moai-core-* skills, and comprehensive project management capabilities.

## Added

- ** Alfred SuperAgent **: Strategic orchestration engine for automated SPEC-Plan-Run-Sync workflow
- ** SPEC-First TDD **: Complete Test-Driven Development methodology with EARS format requirements
- ** moai-core-* Skills **: Unified skill system for domain-specific expertise
- ** Project Management **: Full project lifecycle management from init to documentation
- ** Multilingual Support **: KO/EN/JA/ZH language support
- ** CI/CD Integration **: GitHub Actions workflows for automated testing and deployment

## Installation

```bash
pip install moai-adk==1.0.0
uv tool install moai-adk
```

---

# v1.0.0 - 프로덕션 준비 릴리스 (2026-01-12)

## 요약

MoAI-ADK의 초기 프로덕션 준비 릴리스입니다. Alfred SuperAgent를 통한 자동화된 SPEC-Plan-Run-Sync 워크플로우, 통합 moai-core-* 스킬, 포괄적인 프로젝트 관리 기능을 특징으로 합니다.

## 추가됨

- ** Alfred SuperAgent **: 자동화된 SPEC-Plan-Run-Sync 워크플로우를 위한 전략적 오케스트레이션 엔진
- ** SPEC-First TDD **: EARS 형식 요구사항을 포함한 완전한 테스트 주도 개발 방법론
- ** moai-core-* 스킬 **: 도메인별 전문 지식을 위한 통합 스킬 시스템
- ** 프로젝트 관리 **: init부터 문서화까지 전체 프로젝트 라이프사이클 관리
- ** 다국어 지원 **: KO/EN/JA/ZH 언어 지원
- ** CI/CD 통합 **: 자동화된 테스트 및 배포를 위한 GitHub Actions 워크플로우

## 설치

```bash
pip install moai-adk==1.0.0
uv tool install moai-adk
```

---
