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
