# v1.4.5 - StatusLine Preservation & Critical Bug Fixes (2026-01-19)

## Summary

This patch release resolves critical issues reported in GitHub issues #274 and #275, preserving user statusLine configuration during template updates. Additionally, it includes Windows MCP platform adaptation and various bug fixes.

## Fixed

- **fix(template)**: Preserve user statusLine configuration during template updates
  - Added `statusLine` to `preserve_fields` in merger.py
  - Fixes Issue #274: Windows statusLine loading issues
  - Fixes Issue #275: Template overwrites user's statusLine setting (claude-hud, etc.)
  - Users' custom statusLine settings now preserved across `moai update`
  - File: `src/moai_adk/core/template/merger.py` (line 205)

- **fix(windows)**: Add MCP config platform adaptation for Windows
  - Added `_adapt_mcp_config_for_windows()` method to `TemplateProcessor`
  - Automatically converts `npx` commands to `cmd /c npx` format on Windows
  - Fixes Issue #272: Windows MCP config warning
  - File: `src/moai_adk/core/template/processor.py` (lines 1524-1558)

- **fix(template)**: Prevent settings.json fallback from overwriting platform-specific file
  - Fixed template merge logic to respect platform-specific settings files
  - Prevents Unix settings from overwriting Windows settings and vice versa
  - File: `src/moai_adk/core/template/processor.py`

- **fix(hooks)**: Use $CLAUDE_PROJECT_DIR env var instead of {{PROJECT_DIR}}
  - Fixed hook environment variable resolution
  - Ensures consistent path resolution across different contexts
  - File: Hook configuration files

- **fix(docs)**: Correct install URLs and Windows instructions
  - Updated installation documentation with correct URLs
  - Fixed Windows-specific installation instructions

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Ruff: All checks passed
- Mypy: Success (no issues found in 169 source files)

## Installation & Update

```bash
# In your project folder, run:
moai update

# This will update both the MoAI-ADK package and your project templates
```

---

# v1.4.5 - statusLine 설정 보존 및 중요 버그 수정 (2026-01-19)

## 요약

이 패치 릴리스는 GitHub 이슈 #274, #275에서 보고된 중요한 문제들을 해결하여 템플릿 업데이트 중 사용자 statusLine 설정을 보존합니다. 또한 Windows MCP 플랫폼 적응과 다양한 버그 수정이 포함됩니다.

## 수정됨

- **fix(template)**: 템플릿 업데이트 중 사용자 statusLine 설정 보존
  - merger.py의 `preserve_fields`에 `statusLine` 추가
  - Issue #274 수정: Windows statusLine 로딩 문제
  - Issue #275 수정: 템플릿이 사용자의 statusLine 설정 덮어쓰기 (claude-hud 등)
  - 사용자의 커스텀 statusLine 설정이 `moai update` 시 보존됨
  - 파일: `src/moai_adk/core/template/merger.py` (205행)

- **fix(windows)**: Windows용 MCP 설정 플랫폼 적응 추가
  - `TemplateProcessor`에 `_adapt_mcp_config_for_windows()` 메서드 추가
  - Windows에서 `npx` 명령을 자동으로 `cmd /c npx` 형식으로 변환
  - Issue #272 수정: Windows MCP config 경고
  - 파일: `src/moai_adk/core/template/processor.py` (1524-1558행)

- **fix(template)**: settings.json fallback이 플랫폼별 파일 덮어쓰기 방지
  - 플랫폼별 설정 파일을尊重하도록 템플릿 병합 로직 수정
  - Unix 설정이 Windows 설정을 덮어쓰는 것과 그 반대를 방지
  - 파일: `src/moai_adk/core/template/processor.py`

- **fix(hooks)**: {{PROJECT_DIR}} 대신 $CLAUDE_PROJECT_DIR 환경 변수 사용
  - Hook 환경 변수 해결 수정
  - 다양한 컨텍스트에서 일관된 경로 해결 보장
  - 파일: Hook 설정 파일

- **fix(docs)**: 설치 URL 및 Windows 지침 수정
  - 올바른 URL로 설치 문서 업데이트
  - Windows별 설치 지침 수정

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

## 설치 및 업데이트

```bash
# 프로젝트 폴더에서 다음을 실행하세요:
moai update

# 이 명령은 MoAI-ADK 패키지와 프로젝트 템플릿을 모두 업데이트합니다
```

---

# v1.6.2 - TDD to DDD Migration & StatusLine Preservation (2026-01-19)

## Summary

This release completes the comprehensive migration from TDD (Test-Driven Development) to DDD (Domain-Driven Development) methodology across the entire codebase, enhances Alfred as a professional orchestrator, and fixes critical issues with template settings merge behavior. Additionally, it includes Windows MCP platform adaptation and various bug fixes.

## Changed

- **refactor(methodology)**: Complete TDD to DDD migration across entire codebase
  - Migrated all development workflows from TDD to DDD methodology
  - Updated all test files to use DDD patterns (ANALYZE-PRESERVE-IMPROVE)
  - Removed ~50,000 lines of redundant coverage test files
  - DDD now supports both refactoring (legacy) and new development (Greenfield)
  - Files: All test files, methodology documentation, skills

- **refactor(alfred)**: Redesign Alfred as professional orchestrator
  - Removed manager-claude-code agent (functionality consolidated into Alfred)
  - Enhanced agent delegation patterns with progressive disclosure
  - Improved task decomposition and auto-parallel execution
  - Added comprehensive agent catalog management
  - Files: `.claude/agents/alfred.md`, `CLAUDE.md`

- **refactor(tests)**: Remove redundant coverage test files
  - Removed 18 redundant test coverage files (enhanced, comprehensive variants)
  - Consolidated test organization into single source of truth
  - Improved test maintainability and reduced CI/CD execution time
  - Files: Multiple test files removed

- **docs(skills)**: Update foundation skill examples and references
  - Updated skill documentation with modern examples
  - Improved cross-references between related skills
  - Enhanced progressive disclosure configuration

## Fixed

- **fix(template)**: Preserve user statusLine configuration during template updates
  - Added `statusLine` to `preserve_fields` in merger.py
  - Fixes Issue #274: Windows statusLine loading issues
  - Fixes Issue #275: Template overwrites user's statusLine setting (claude-hud, etc.)
  - Users' custom statusLine settings now preserved across `moai update`
  - File: `src/moai_adk/core/template/merger.py` (line 205)

- **fix(windows)**: Add MCP config platform adaptation for Windows
  - Added `_adapt_mcp_config_for_windows()` method to `TemplateProcessor`
  - Automatically converts `npx` commands to `cmd /c npx` format on Windows
  - Fixes Issue #272: Windows MCP config warning
  - File: `src/moai_adk/core/template/processor.py` (lines 1524-1558)

- **fix(template)**: Prevent settings.json fallback from overwriting platform-specific file
  - Fixed template merge logic to respect platform-specific settings files
  - Prevents Unix settings from overwriting Windows settings and vice versa
  - File: `src/moai_adk/core/template/processor.py`

- **fix(hooks)**: Use $CLAUDE_PROJECT_DIR env var instead of {{PROJECT_DIR}}
  - Fixed hook environment variable resolution
  - Ensures consistent path resolution across different contexts
  - File: Hook configuration files

- **fix(docs)**: Correct install URLs and Windows instructions
  - Updated installation documentation with correct URLs
  - Fixed Windows-specific installation instructions

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Ruff: All checks passed
- Mypy: Success (no issues found in 169 source files)

## Installation & Update

```bash
# In your project folder, run:
moai update

# This will update both the MoAI-ADK package and your project templates
```

---

# v1.6.2 - TDD에서 DDD로 마이그레이션 및 statusLine 설정 보존 (2026-01-19)

## 요약

이 릴리스는 전체 코드베이스에서 TDD(테스트 주도 개발)에서 DDD(도메인 주도 개발) 방법론으로의 포괄적인 마이그레이션을 완료하고, Alfred를 전문 오케스트레이터로 개선하며, 템플릿 설정 병합 동작의 중요한 문제들을 수정합니다. 또한 Windows MCP 플랫폼 적응과 다양한 버그 수정이 포함됩니다.

## 변경됨

- **refactor(methodology)**: 전체 코드베이스의 TDD에서 DDD로 마이그레이션 완료
  - 모든 개발 워크플로우를 TDD에서 DDD 방법론으로 마이그레이션
  - 모든 테스트 파일을 DDD 패턴(ANALYZE-PRESERVE-IMPROVE)으로 업데이트
  - 약 50,000줄의 중복된 커버리지 테스트 파일 제거
  - DDD가 이제 리팩토링(레거시)과 새 개발(Greenfield)을 모두 지원
  - 파일: 모든 테스트 파일, 방법론 문서, 스킬

- **refactor(alfred)**: 전문 오케스트레이터로 Alfred 재설계
  - manager-claude-code 에이전트 제거 (기능이 Alfred에 통합됨)
  - 점진적 공개를 통한 에이전트 위임 패턴 개선
  - 작업 분해 및 자동 병렬 실행 개선
  - 포괄적인 에이전트 카탈로그 관리 추가
  - 파일: `.claude/agents/alfred.md`, `CLAUDE.md`

- **refactor(tests)**: 중복된 커버리지 테스트 파일 제거
  - 18개의 중복된 테스트 커버리지 파일 제거 (enhanced, comprehensive 변형)
  - 테스트 조직을 단일 정보 소스로 통합
  - 테스트 유지보수성 개선 및 CI/CD 실행 시간 단축
  - 파일: 여러 테스트 파일 제거됨

- **docs(skills)**: Foundation 스킬 예제 및 참조 업데이트
  - 최신 예제로 스킬 문서 업데이트
  - 관련 스킬 간의 상호 참조 개선
  - 점진적 공개 구성 개선

## 수정됨

- **fix(template)**: 템플릿 업데이트 중 사용자 statusLine 설정 보존
  - merger.py의 `preserve_fields`에 `statusLine` 추가
  - Issue #274 수정: Windows statusLine 로딩 문제
  - Issue #275 수정: 템플릿이 사용자의 statusLine 설정 덮어쓰기 (claude-hud 등)
  - 사용자의 커스텀 statusLine 설정이 `moai update` 시 보존됨
  - 파일: `src/moai_adk/core/template/merger.py` (205행)

- **fix(windows)**: Windows용 MCP 설정 플랫폼 적응 추가
  - `TemplateProcessor`에 `_adapt_mcp_config_for_windows()` 메서드 추가
  - Windows에서 `npx` 명령을 자동으로 `cmd /c npx` 형식으로 변환
  - Issue #272 수정: Windows MCP config 경고
  - 파일: `src/moai_adk/core/template/processor.py` (1524-1558행)

- **fix(template)**: settings.json fallback이 플랫폼별 파일 덮어쓰기 방지
  - 플랫폼별 설정 파일을尊重하도록 템플릿 병합 로직 수정
  - Unix 설정이 Windows 설정을 덮어쓰는 것과 그 반대를 방지
  - 파일: `src/moai_adk/core/template/processor.py`

- **fix(hooks)**: {{PROJECT_DIR}} 대신 $CLAUDE_PROJECT_DIR 환경 변수 사용
  - Hook 환경 변수 해결 수정
  - 다양한 컨텍스트에서 일관된 경로 해결 보장
  - 파일: Hook 설정 파일

- **fix(docs)**: 설치 URL 및 Windows 지침 수정
  - 올바른 URL로 설치 문서 업데이트
  - Windows별 설치 지침 수정

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

## 설치 및 업데이트

```bash
# 프로젝트 폴더에서 다음을 실행하세요:
moai update

# 이 명령은 MoAI-ADK 패키지와 프로젝트 템플릿을 모두 업데이트합니다
```

---

# v1.4.1 - Alfred Agent Delegation and Git Strategy Improvements (2026-01-19)

## Summary

This patch release strengthens Alfred's agent orchestration rules to ensure consistent delegation after auto compact, adds resume capability to the fix command for better session recovery, and refines personal mode git strategy defaults to be more conservative and user-friendly.

## Added

- **feat(fix)**: Add Resume pattern support to fix command
  - Implemented snapshot-based recovery system in `.moai/cache/fix-snapshots/`
  - Added `--resume [ID]` argument to resume from latest or specific snapshot
  - Preserves fix state including issues_found, issues_fixed, issues_pending, and todo_state
  - Enables seamless continuation after auto compact or session interruption
  - Files: `.claude/commands/moai/fix.md`, `src/moai_adk/templates/.claude/commands/moai/fix.md`

## Fixed

- **fix(alfred)**: Add AGENT DELEGATION MANDATE to prevent direct execution after auto compact
  - Added explicit [HARD] rules to Type A commands (1-plan, 2-run, 3-sync) requiring agent delegation
  - Added explicit [HARD] rules to Type B commands (alfred, fix, loop) for agent delegation
  - Prevents Alfred from executing implementation directly after context recovery
  - WHY: Auto compact recovery was causing Alfred to violate orchestrator role by implementing directly
  - Files: `.claude/commands/moai/1-plan.md`, `2-run.md`, `3-sync.md`, `alfred.md`, `fix.md`, `loop.md`, `CLAUDE.md`

- **fix(git-strategy)**: Disable auto-branch and auto-push in personal mode by default
  - Changed `branch_creation.prompt_always`: false → true (ask before creating)
  - Changed `branch_creation.auto_enabled`: true → false (disabled by default)
  - Changed `automation.auto_branch`: true → false (disabled by default)
  - Changed `automation.auto_push`: true → false (disabled by default)
  - More conservative defaults prevent unintended branch creation and remote pushes
  - Files: `.moai/config/sections/git-strategy.yaml`, `src/moai_adk/templates/.moai/config/sections/git-strategy.yaml`

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Ruff: All checks passed
- Mypy: Success (no issues found in 169 source files)

## Installation & Update

```bash
# In your project folder, run:
moai update
```

---

# v1.4.1 - Alfred 에이전트 위임 및 Git 전략 개선 (2026-01-19)

## 요약

이 패치 릴리스는 auto compact 후 일관된 위임을 보장하기 위해 Alfred의 에이전트 오케스트레이션 규칙을 강화하고, 세션 복구를 개선하기 위해 fix 명령에 재개 기능을 추가하며, personal 모드 git 전략 기본값을 보다 보수적이고 사용자 친화적으로 개선합니다.

## 추가됨

- **feat(fix)**: fix 명령에 Resume 패턴 지원 추가
  - `.moai/cache/fix-snapshots/`에 스냅샷 기반 복구 시스템 구현
  - 최신 또는 특정 스냅샷에서 재개하기 위한 `--resume [ID]` 인자 추가
  - issues_found, issues_fixed, issues_pending, todo_state를 포함한 fix 상태 보존
  - auto compact 또는 세션 중단 후 원활한 계속 가능
  - 파일: `.claude/commands/moai/fix.md`, `src/moai_adk/templates/.claude/commands/moai/fix.md`

## 수정됨

- **fix(alfred)**: auto compact 후 직접 실행 방지를 위한 AGENT DELEGATION MANDATE 추가
  - Type A 명령(1-plan, 2-run, 3-sync)에 에이전트 위임을 요구하는 명시적 [HARD] 규칙 추가
  - Type B 명령(alfred, fix, loop)에 에이전트 위임을 위한 명시적 [HARD] 규칙 추가
  - 컨텍스트 복구 후 Alfred가 직접 구현을 실행하는 것 방지
  - WHY: Auto compact 복구가 Alfred가 오케스트레이터 역할을 위반하고 직접 구현하도록 야기
  - 파일: `.claude/commands/moai/1-plan.md`, `2-run.md`, `3-sync.md`, `alfred.md`, `fix.md`, `loop.md`, `CLAUDE.md`

- **fix(git-strategy)**: personal 모드에서 auto-branch 및 auto-push 기본 비활성화
  - `branch_creation.prompt_always`: false → true로 변경 (생성 전 물어봄)
  - `branch_creation.auto_enabled`: true → false로 변경 (기본 비활성화)
  - `automation.auto_branch`: true → false로 변경 (기본 비활성화)
  - `automation.auto_push`: true → false로 변경 (기본 비활성화)
  - 더 보수적인 기본값으로 의도하지 않은 브랜치 생성 및 원격 push 방지
  - 파일: `.moai/config/sections/git-strategy.yaml`, `src/moai_adk/templates/.moai/config/sections/git-strategy.yaml`

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

## 설치 및 업데이트

```bash
# 프로젝트 폴더에서 다음을 실행하세요:
moai update
```

---

# v1.4.0 - Version Display Fix and Update Command Enhancement (2026-01-18)

## Summary

This minor release resolves a critical version display bug where `moai` command showed stale version "1.1.0" instead of the actual installed version. It also enhances the `moai update` command to automatically sync templates after package upgrade, eliminating the need for manual re-run.

## Fixed

- **fix(version)**: Resolve version display showing stale 1.1.0 instead of actual version
  - Implemented 3-tier version resolution with pyproject.toml priority reading
  - Priority chain: pyproject.toml → importlib.metadata → fallback constant
  - Fixes editable install metadata cache issue that returns outdated version
  - Added `_get_version_from_pyproject()` function using Python 3.11+ `tomllib`
  - File: `src/moai_adk/version.py` (lines 18-62)

- **feat(update)**: Auto-run template sync after package upgrade
  - Automatically executes `moai update --templates-only` via subprocess after package upgrade
  - Eliminates manual re-run requirement for template synchronization
  - Includes 5-minute timeout protection and graceful error handling
  - Preserves `--yes` flag for CI/CD automation
  - File: `src/moai_adk/cli/commands/update.py` (lines 2350-2381)

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Ruff: All checks passed (1 file reformatted)
- Mypy: Success (no issues found in 169 source files)

## Installation & Update

```bash
# In your project folder, run:
moai update
```

---

# v1.4.0 - 버전 표시 수정 및 업데이트 명령 개선 (2026-01-18)

## 요약

이 마이너 릴리스는 `moai` 명령이 실제 설치된 버전 대신 오래된 버전 "1.1.0"을 표시하는 중요한 버그를 해결합니다. 또한 `moai update` 명령을 개선하여 패키지 업그레이드 후 자동으로 템플릿을 동기화하므로 수동 재실행이 필요 없습니다.

## 수정됨

- **fix(version)**: 실제 버전 대신 오래된 1.1.0이 표시되는 버전 표시 문제 해결
  - pyproject.toml 우선 읽기를 사용한 3단계 버전 해결 구현
  - 우선순위 체인: pyproject.toml → importlib.metadata → 폴백 상수
  - 오래된 버전을 반환하는 editable install 메타데이터 캐시 문제 수정
  - Python 3.11+ `tomllib`을 사용하는 `_get_version_from_pyproject()` 함수 추가
  - 파일: `src/moai_adk/version.py` (18-62행)

- **feat(update)**: 패키지 업그레이드 후 템플릿 동기화 자동 실행
  - 패키지 업그레이드 후 subprocess를 통해 `moai update --templates-only` 자동 실행
  - 템플릿 동기화를 위한 수동 재실행 요구사항 제거
  - 5분 타임아웃 보호 및 우아한 오류 처리 포함
  - CI/CD 자동화를 위한 `--yes` 플래그 보존
  - 파일: `src/moai_adk/cli/commands/update.py` (2350-2381행)

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Ruff: 모든 검사 통과 (1개 파일 형식 수정)
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

## 설치 및 업데이트

```bash
# 프로젝트 폴더에서 다음을 실행하세요:
moai update
```

---

# v1.3.10 - Critical Bug Fixes for Worktree, Template Variables, and Hooks (2026-01-18)

## Summary

This patch release resolves critical bugs reported in GitHub issues #270 and #269, including worktree path duplication, template variable substitution failures, and hook import errors. Additionally, it improves Windows compatibility with UTF-8 encoding and consolidates loop control commands.

## Fixed

- **fix(worktree)**: Resolve path duplication bug in moai-worktree (#270)
  - Removed project name from `legacy_roots` in `_detect_worktree_root()` function
  - Prevents duplicate project names in worktree paths (e.g., `~/worktrees/iRMS/iRMS/SPEC-XXX` → `~/worktrees/iRMS/SPEC-XXX`)
  - Updated legacy root detection to exclude paths with `main_repo_path.name`
  - File: `src/moai_adk/cli/worktree/cli.py` (lines 109-117)

- **fix(hooks)**: Resolve ToolType NameError in linter hook (#269)
  - Added explicit `from tool_registry import ToolType` to prevent NameError
  - ToolType is now properly defined in module scope regardless of import success
  - Fixes Windows linting hook failures with "NameError: name 'ToolType' is not defined"
  - File: `src/moai_adk/templates/.claude/hooks/moai/post_tool__linter.py` (line 40)

- **fix(template)**: Fix PROJECT_DIR_UNIX template variable not substituted in moai init
  - Added `PROJECT_DIR_WIN` and `PROJECT_DIR_UNIX` variable definitions in `execute_resource_phase()`
  - Both variables are now properly included in template substitution context
  - Resolves hook execution failures due to unsubstituted `{{PROJECT_DIR_UNIX}}` placeholders
  - File: `src/moai_adk/core/project/phase_executor.py` (lines 315-359)

- **fix(encoding)**: Add UTF-8 encoding to file operations for Windows compatibility
  - Explicitly specify `encoding="utf-8"` for worktree registry file reads
  - Prevents cp949 encoding errors on Korean Windows systems
  - File: `src/moai_adk/cli/worktree/cli.py` (lines 88, 122)

- **refactor(commands)**: Remove /moai:cancel-loop command (functionality merged into /moai:loop)
  - Simplified loop control by consolidating cancel functionality into main loop command
  - Users can now cancel loops directly from `/moai:loop` command
  - Removed 163 lines of redundant command code
  - File: `.claude/commands/moai/cancel-loop.md` (removed)

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Worktree tests: 131 passed (100% pass rate)
- Ruff: All checks passed (1 file reformatted)
- Mypy: Success (no issues found in 169 source files)

## Installation & Update

```bash
# In your project folder, run:
moai update
```

---

# v1.3.10 - Worktree, 템플릿 변수, Hooks 중요 버그 수정 (2026-01-18)

## 요약

이 패치 릴리스는 GitHub 이슈 #270, #269에서 보고된 중요한 버그들을 해결합니다. 워크트리 경로 중복, 템플릿 변수 치환 실패, hook import 오류가 포함됩니다. 추가로 UTF-8 인코딩으로 Windows 호환성을 개선하고 루프 제어 명령을 통합했습니다.

## 수정됨

- **fix(worktree)**: moai-worktree 경로 중복 버그 해결 (#270)
  - `_detect_worktree_root()` 함수의 `legacy_roots`에서 프로젝트 이름 제거
  - 워크트리 경로에서 프로젝트 이름 중복 방지 (예: `~/worktrees/iRMS/iRMS/SPEC-XXX` → `~/worktrees/iRMS/SPEC-XXX`)
  - `main_repo_path.name`이 포함된 경로를 제외하도록 legacy root 탐지 업데이트
  - 파일: `src/moai_adk/cli/worktree/cli.py` (109-117행)

- **fix(hooks)**: linter hook의 ToolType NameError 해결 (#269)
  - NameError 방지를 위해 명시적으로 `from tool_registry import ToolType` 추가
  - import 성공 여부와 관계없이 ToolType이 모듈 스코프에서 올바르게 정의됨
  - "NameError: name 'ToolType' is not defined" Windows linting hook 실패 수정
  - 파일: `src/moai_adk/templates/.claude/hooks/moai/post_tool__linter.py` (40행)

- **fix(template)**: moai init에서 PROJECT_DIR_UNIX 템플릿 변수 치환 안 되는 문제 수정
  - `execute_resource_phase()`에 `PROJECT_DIR_WIN` 및 `PROJECT_DIR_UNIX` 변수 정의 추가
  - 두 변수 모두 템플릿 치환 컨텍스트에 올바르게 포함됨
  - 치환되지 않은 `{{PROJECT_DIR_UNIX}}` placeholder로 인한 hook 실행 실패 해결
  - 파일: `src/moai_adk/core/project/phase_executor.py` (315-359행)

- **fix(encoding)**: Windows 호환성을 위한 파일 작업에 UTF-8 인코딩 추가
  - worktree registry 파일 읽기에 `encoding="utf-8"` 명시적 지정
  - 한국어 Windows 시스템에서 cp949 인코딩 오류 방지
  - 파일: `src/moai_adk/cli/worktree/cli.py` (88, 122행)

- **refactor(commands)**: /moai:cancel-loop 명령 제거 (기능은 /moai:loop에 통합)
  - cancel 기능을 main loop 명령에 통합하여 루프 제어 간소화
  - 사용자가 이제 `/moai:loop` 명령에서 직접 루프 취소 가능
  - 중복된 명령 코드 163줄 제거
  - 파일: `.claude/commands/moai/cancel-loop.md` (제거됨)

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Worktree 테스트: 131개 통과 (100% 통과율)
- Ruff: 모든 검사 통과 (1개 파일 형식 수정)
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

## 설치 및 업데이트

```bash
# 프로젝트 폴더에서 다음을 실행하세요:
moai update
```

---

# v1.3.9 - OS-Specific Settings Installation Fix (2026-01-18)

## Summary

This patch release fixes the OS-specific settings.json installation logic in the template processor. The `moai init` and `moai update` commands now correctly install platform-appropriate settings files based on the user's operating system.

## Fixed

- **fix(template)**: Install OS-specific settings.json correctly
  - Skip platform-specific settings files that don't match current OS
    - `settings.json.windows` is now skipped on macOS/Linux
    - `settings.json.unix` is now skipped on Windows
  - Fix destination path to always save as `settings.json` (was incorrectly saving as `settings.json.unix` or `settings.json.windows`)
  - Fix variable substitution to use correct destination path for merged settings file
  - File: `src/moai_adk/core/template/processor.py` (function: `_copy_claude()`)

- **chore**: Update builder-skill agent and simplify pre-push hook
  - Enhanced builder-skill agent configuration
  - Streamlined pre-push hook for better performance (-498 lines)
  - Removed deprecated SKILL_TEMPLATE.md

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Template processor tests: 190 passed (100% pass rate)
- Ruff: All checks passed
- Mypy: Success (no issues found in 169 source files)

## Installation & Update

```bash
# In your project folder, run:
moai update
```

---

# v1.3.9 - OS별 설정 파일 설치 수정 (2026-01-18)

## 요약

이 패치 릴리스는 템플릿 프로세서의 OS별 settings.json 설치 로직을 수정합니다. `moai init` 및 `moai update` 명령이 이제 사용자의 운영 체제에 따라 플랫폼에 적합한 설정 파일을 올바르게 설치합니다.

## 수정됨

- **fix(template)**: OS별 settings.json 올바른 설치
  - 현재 OS와 일치하지 않는 플랫폼별 설정 파일 건너뛰기
    - macOS/Linux에서 `settings.json.windows` 건너뛰기
    - Windows에서 `settings.json.unix` 건너뛰기
  - 목적지 경로를 항상 `settings.json`으로 저장하도록 수정 (이전에는 잘못하여 `settings.json.unix` 또는 `settings.json.windows`로 저장)
  - 병합된 설정 파일에 올바른 목적지 경로를 사용하도록 변수 치환 수정
  - 파일: `src/moai_adk/core/template/processor.py` (함수: `_copy_claude()`)

- **chore**: builder-skill 에이전트 업데이트 및 pre-push hook 간소화
  - builder-skill 에이전트 설정 개선
  - 더 나은 성능을 위해 pre-push hook 간소화 (-498줄)
  - 더 이상 사용되지 않는 SKILL_TEMPLATE.md 제거

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Template processor 테스트: 190개 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

## 설치 및 업데이트

```bash
# 프로젝트 폴더에서 다음을 실행하세요:
moai update
```

---

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
# In your project folder, run:
moai update
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
# 프로젝트 폴더에서 다음을 실행하세요:
moai update
```

---
