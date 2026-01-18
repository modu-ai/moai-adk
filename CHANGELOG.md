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
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version (should now show 1.4.0)
moai --version
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
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인 (이제 1.4.0으로 표시되어야 함)
moai --version
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
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version
moai --version
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
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인
moai --version
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
# Update to the latest version
uv tool update moai-adk

# Or using pipx
pipx upgrade moai-adk

# Verify version
moai --version
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
# 최신 버전으로 업데이트
uv tool update moai-adk

# 또는 pipx 사용
pipx upgrade moai-adk

# 버전 확인
moai --version
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
