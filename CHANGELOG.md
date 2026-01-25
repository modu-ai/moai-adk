# v1.8.6 - Universal Cross-Platform Shell Wrapper (2026-01-25)

## Summary

This minor release implements a universal shell wrapper system that respects user's default shell on all platforms, resolving shell-specific PATH loading issues. The system automatically uses the user's preferred shell (zsh, bash, etc.) instead of forcing bash, ensuring environment variables are loaded correctly on macOS, Linux, Windows Git Bash, and WSL.

**Key Changes**:
- Universal shell wrapper: `${SHELL:-/bin/bash} -l -c` for all Unix platforms
- Template variables: `{{HOOK_SHELL_PREFIX}}` and `{{HOOK_SHELL_SUFFIX}}`
- Cross-platform path handling improvements
- Comprehensive cross-platform development guidelines

**Impact**:
- macOS users with zsh can now use MoAI-ADK without PATH issues
- bash users continue to work as before (backward compatible)
- Windows continues to use direct execution (no wrapper needed)
- WSL users benefit from auto-shell detection

## Breaking Changes

None. This release is backward compatible with v1.8.4.

## Fixed

### Shell-Specific PATH Loading Issues

- **fix(templates)**: Fix double slash in hook paths after PROJECT_DIR substitution (69f12f88)
  - Fixed template pattern: `{{PROJECT_DIR}}/.claude/` → `{{PROJECT_DIR}}.claude/`
  - `{{PROJECT_DIR}}` already includes trailing slash by design (v1.8.0)
  - Resolves hook execution failures with exit code 127
  - File: `src/moai_adk/templates/.claude/settings.json`

- **fix(templates)**: Fix double slash in agent hook paths after PROJECT_DIR substitution (e0e3c24e)
  - Updated 12 agent hook commands with corrected path pattern
  - Affected agents: builder-*, expert-backend, expert-frontend, expert-security, expert-refactoring, expert-debug, manager-ddd, manager-quality
  - Files: `src/moai_adk/templates/.claude/agents/moai/*.md`

### Cross-Platform Compatibility

- **feat(cross-platform)**: Implement universal shell wrapper for all platforms (f87df602)
  - New template variables: `{{HOOK_SHELL_PREFIX}}`, `{{HOOK_SHELL_SUFFIX}}`
  - Unix: `${SHELL:-/bin/bash} -l -c` (respects user's default shell)
  - Windows: Direct execution (no wrapper needed)
  - Automatically loads zsh config (`~/.zshenv`) or bash config (`~/.bash_profile`)
  - File: `src/moai_adk/cli/commands/update.py`

## Changed

### Documentation

- **docs(development)**: Add comprehensive cross-platform development guidelines (f87df602)
  - New section in `CLAUDE.local.md`: Cross-Platform Development Guidelines
  - Documents shell wrapper strategy and template variable usage
  - Provides examples for future hook and agent development
  - Critical rules for Windows/macOS/Linux compatibility
  - File: `CLAUDE.local.md` (Section 20, 240+ lines)

### Template Updates

- **sync**: Update local .claude/ from templates with cross-platform shell wrapper (3bf75656)
  - Synced all agent hook commands with new shell wrapper pattern
  - Updated settings.json with cross-platform shell wrapper
  - Removed `expert-stitch.md` (not in v1.8.6 template)
  - All hooks now use `${SHELL:-/bin/bash} -l -c` for PATH loading

- **chore**: Bump version to 1.8.6 (97e76ff2)
  - Updated: `pyproject.toml`, `src/moai_adk/version.py`, `.moai/config/config.yaml`, `.moai/config/sections/system.yaml`

## Technical Details

### Hook Command Pattern Evolution

**v1.8.4 Pattern (Issue: Forced bash on zsh users)**:
```json
{
  "command": "bash -l -c 'uv run \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/session_start.py\"'"
}
```

**v1.8.6 Pattern (Solution: Respects user's shell)**:
```json
{
  "command": "${SHELL:-/bin/bash} -l -c 'uv run \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/session_start.py\"'"
}
```

### Template Variables

| Variable | Usage | Example |
|----------|-------|---------|
| `{{HOOK_SHELL_PREFIX}}` | Shell wrapper opening | `${SHELL:-/bin/bash} -l -c '` (Unix) or `""` (Windows) |
| `{{HOOK_SHELL_SUFFIX}}` | Shell wrapper closing | `'` (Unix) or `""` (Windows) |
| `{{PROJECT_DIR}}` | Project path with trailing separator | `$CLAUDE_PROJECT_DIR/` (forward slash, cross-platform) |

### Cross-Platform Behavior

| Platform | Shell Wrapper | Loaded Config Files | Status |
|----------|---------------|---------------------|--------|
| macOS (zsh) | `zsh -l -c` | `~/.zshenv`, `~/.zprofile` | ✅ Fixed |
| macOS (bash) | `bash -l -c` | `~/.bash_profile`, `~/.bashrc` | ✅ Works |
| Linux (bash) | `bash -l -c` | `~/.bash_profile`, `~/.bashrc` | ✅ Works |
| WSL (bash/zsh) | Respects `$SHELL` | User's shell config | ✅ Fixed |
| Windows Git Bash | `bash -l -c` | `.bash_profile` | ✅ Works |
| Windows PowerShell | Direct execution | System PATH | ✅ Works |

### Files Modified

- `src/moai_adk/cli/commands/update.py`: Shell wrapper generation logic (lines 1845-1862)
- `src/moai_adk/templates/.claude/settings.json`: All 6 hooks updated
- `src/moai_adk/templates/.claude/agents/moai/*.md`: 12 agents updated
- `CLAUDE.local.md`: Cross-platform development guidelines added
- `pyproject.toml`, `src/moai_adk/version.py`: Version updated to 1.8.6
- `.moai/config/config.yaml`, `.moai/config/sections/system.yaml`: Version synced

## Migration Guide

### For New Users (Fresh Install)

```bash
# Install MoAI-ADK
uv tool install moai-adk

# Initialize project (automatically uses v1.8.6 templates)
moai init
```

### For Existing Users (Upgrade from v1.8.4 or earlier)

```bash
# Update MoAI-ADK
uv tool install moai-adk

# Update project templates - hooks automatically use shell wrapper
moai update

# Verify version
moai --version  # Should show 1.8.6
```

**Note**: The `moai update` command automatically updates all hook commands to use the new shell wrapper pattern. No manual intervention required.

## Quality

- Smoke Tests: 6/6 passed (100% pass rate)
- Ruff check: All checks passed
- Ruff format: 222 files unchanged
- Mypy: Success (no issues found in 174 source files)
- Test Coverage: Maintained at 88.12%

## Installation & Update

```bash
# Update to the latest version
uv tool install moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.8.6 - 범용 크로스 플랫폼 셸 래퍼 (2026-01-25)

## 요약

이 마이너 릴리스는 모든 플랫폼에서 사용자의 기본 셸을 존중하는 범용 셸 래퍼 시스템을 구현하여 셸별 PATH 로딩 문제를 해결합니다. 시스템은 bash를 강제하는 대신 사용자의 선호 셸(zsh, bash 등)을 자동으로 사용하여 macOS, Linux, Windows Git Bash, WSL에서 환경 변수가 올바르게 로드되도록 보장합니다.

**주요 변경사항**:
- 범용 셸 래퍼: 모든 Unix 플랫폼에 대해 `${SHELL:-/bin/bash} -l -c`
- 템플릿 변수: `{{HOOK_SHELL_PREFIX}}`와 `{{HOOK_SHELL_SUFFIX}}`
- 크로스 플랫폼 경로 처리 개선
- 포괄적인 크로스 플랫폼 개발 가이드라인

**영향**:
- zsh를 사용하는 macOS 사용자가 PATH 문제 없이 MoAI-ADK 사용 가능
- bash 사용자는 이전과 동일하게 작동 (하위 호환)
- Windows는 직접 실행 계속 사용 (래퍼 불필요)
- WSL 사용자는 자동 셸 감지 혜택

## Breaking Changes

없음. 이 릴리스는 v1.8.4와 하위 호환됩니다.

## 수정됨

### 셸별 PATH 로딩 문제

- **fix(templates)**: PROJECT_DIR 치환 후 hook 경로의 이중 슬래시 수정 (69f12f88)
  - 템플릿 패턴 수정: `{{PROJECT_DIR}}/.claude/` → `{{PROJECT_DIR}}.claude/`
  - `{{PROJECT_DIR}}`는 설계상 이미 후행 슬래시 포함 (v1.8.0)
  - exit code 127로 hook 실행 실패 해결
  - 파일: `src/moai_adk/templates/.claude/settings.json`

- **fix(templates)**: PROJECT_DIR 치환 후 agent hook 경로의 이중 슬래시 수정 (e0e3c24e)
  - 수정된 경로 패턴으로 12개 agent hook 명령 업데이트
  - 영향받는 에이전트: builder-*, expert-backend, expert-frontend, expert-security, expert-refactoring, expert-debug, manager-ddd, manager-quality
  - 파일: `src/moai_adk/templates/.claude/agents/moai/*.md`

### 크로스 플랫폼 호환성

- **feat(cross-platform)**: 모든 플랫폼을 위한 범용 셸 래퍼 구현 (f87df602)
  - 새로운 템플릿 변수: `{{HOOK_SHELL_PREFIX}}`, `{{HOOK_SHELL_SUFFIX}}`
  - Unix: `${SHELL:-/bin/bash} -l -c` (사용자의 기본 셸 존중)
  - Windows: 직접 실행 (래퍼 불필요)
  - zsh 설정(`~/.zshenv`) 또는 bash 설정(`~/.bash_profile`)을 자동으로 로드
  - 파일: `src/moai_adk/cli/commands/update.py`

## 변경됨

### 문서화

- **docs(development)**: 포괄적인 크로스 플랫폼 개발 가이드라인 추가 (f87df602)
  - `CLAUDE.local.md`에 새 섹션: 크로스 플랫폼 개발 가이드라인
  - 셸 래퍼 전략 및 템플릿 변수 사용 문서화
  - 향후 hook 및 agent 개발을 위한 예제 제공
  - Windows/macOS/Linux 호환성을 위한 핵심 규칙
  - 파일: `CLAUDE.local.md` (Section 20, 240+ 줄)

### 템플릿 업데이트

- **sync**: 크로스 플랫폼 셸 래퍼로 로컬 .claude/ 템플릿에서 업데이트 (3bf75656)
  - 새로운 셸 래퍼 패턴으로 모든 agent hook 명령 동기화
  - 크로스 플랫폼 셸 래퍼로 settings.json 업데이트
  - `expert-stitch.md` 제거 (v1.8.6 템플릿에 없음)
  - 모든 hook이 PATH 로딩을 위해 `${SHELL:-/bin/bash} -l -c` 사용

- **chore**: 버전을 1.8.6으로 업데이트 (97e76ff2)
  - 업데이트됨: `pyproject.toml`, `src/moai_adk/version.py`, `.moai/config/config.yaml`, `.moai/config/sections/system.yaml`

## 기술 세부사항

### Hook 명령 패턴 진화

**v1.8.4 패턴 (문제: zsh 사용자에게 bash 강제)**:
```json
{
  "command": "bash -l -c 'uv run \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/session_start.py\"'"
}
```

**v1.8.6 패턴 (해결: 사용자의 셸 존중)**:
```json
{
  "command": "${SHELL:-/bin/bash} -l -c 'uv run \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/session_start.py\"'"
}
```

### 템플릿 변수

| 변수 | 사용 | 예제 |
|------|------|------|
| `{{HOOK_SHELL_PREFIX}}` | 셸 래퍼 시작 | `${SHELL:-/bin/bash} -l -c '` (Unix) 또는 `""` (Windows) |
| `{{HOOK_SHELL_SUFFIX}}` | 셸 래퍼 종료 | `'` (Unix) 또는 `""` (Windows) |
| `{{PROJECT_DIR}}` | 후행 구분자가 있는 프로젝트 경로 | `$CLAUDE_PROJECT_DIR/` (슬래시, 크로스 플랫폼) |

### 크로스 플랫폼 동작

| 플랫폼 | 셸 래퍼 | 로드되는 설정 파일 | 상태 |
|--------|---------|-------------------|------|
| macOS (zsh) | `zsh -l -c` | `~/.zshenv`, `~/.zprofile` | ✅ 수정됨 |
| macOS (bash) | `bash -l -c` | `~/.bash_profile`, `~/.bashrc` | ✅ 작동 |
| Linux (bash) | `bash -l -c` | `~/.bash_profile`, `~/.bashrc` | ✅ 작동 |
| WSL (bash/zsh) | `$SHELL` 존중 | 사용자의 셸 설정 | ✅ 수정됨 |
| Windows Git Bash | `bash -l -c` | `.bash_profile` | ✅ 작동 |
| Windows PowerShell | 직접 실행 | 시스템 PATH | ✅ 작동 |

### 수정된 파일

- `src/moai_adk/cli/commands/update.py`: 셸 래퍼 생성 로직 (1845-1862줄)
- `src/moai_adk/templates/.claude/settings.json`: 모든 6개 hook 업데이트
- `src/moai_adk/templates/.claude/agents/moai/*.md`: 12개 agent 업데이트
- `CLAUDE.local.md`: 크로스 플랫폼 개발 가이드라인 추가
- `pyproject.toml`, `src/moai_adk/version.py`: 버전 1.8.6으로 업데이트
- `.moai/config/config.yaml`, `.moai/config/sections/system.yaml`: 버전 동기화

## 마이그레이션 가이드

### 새 사용자 (새로 설치)

```bash
# MoAI-ADK 설치
uv tool install moai-adk

# 프로젝트 초기화 (자동으로 v1.8.6 템플릿 사용)
moai init
```

### 기존 사용자 (v1.8.4 이하에서 업그레이드)

```bash
# MoAI-ADK 업데이트
uv tool install moai-adk

# 프로젝트 템플릿 업데이트 - hook이 자동으로 셸 래퍼 사용
moai update

# 버전 확인
moai --version  # 1.8.6이 표시되어야 함
```

**참고**: `moai update` 명령이 모든 hook 명령을 새로운 셸 래퍼 패턴으로 자동 업데이트합니다. 수동 개입이 필요하지 않습니다.

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff check: 모든 검사 통과
- Ruff format: 222개 파일 변경 없음
- Mypy: 성공 (174개 소스 파일에서 문제 없음)
- 테스트 커버리지: 88.12% 유지

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool install moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.8.4 - Hook PATH Fix for WSL & Cross-Platform Compatibility (2026-01-25)

## Summary

This patch release completely resolves **Issue #296** (moai update PATH loading issues) by implementing login shell execution for all hooks, ensuring PATH environment variables are correctly loaded on WSL, macOS, Linux, and Windows Git Bash.

**Key Changes**:
- All hooks now use `bash -l -c` wrapper for PATH loading
- MoAI Rank hooks automatically updated during `moai update`
- Settings.json consolidation for simplified maintenance
- Complete cross-platform compatibility (WSL/macOS/Linux/Windows)

**Reference**: Issue #296

## Fixed

### Hook PATH Loading Issues

- **fix(hooks)**: Use login shell for hooks to ensure PATH is loaded (e1777b94)
  - All template hooks now use `bash -l -c` wrapper pattern
  - Ensures `~/.bashrc`, `~/.zshrc`, and `.bash_profile` are loaded
  - Resolves `command not found: uv` errors in non-interactive shells
  - File: `src/moai_adk/templates/.claude/settings.json`

- **fix(hooks)**: Update agent hooks to use login shell for PATH loading (80602d5d)
  - Updated all agent hook commands with `bash -l -c` wrapper
  - Affected agents: expert-frontend, expert-security, expert-backend, manager-quality, builder-*
  - Files: `src/moai_adk/templates/.claude/agents/moai/*.md` (12 agents updated)

- **fix(rank)**: Update moai rank SessionEnd hook to use bash login shell (84cf8f02)
  - MoAI Rank installation now registers hooks with `bash -l -c` wrapper
  - File: `src/moai_adk/rank/hook.py`

## Changed

### Settings Configuration

- **refactor(settings)**: Consolidate platform-specific settings into unified settings.json (c3020305)
  - Removed: `settings.json.unix`, `settings.json.windows`
  - Unified: Single `settings.json` with cross-platform compatibility
  - Simplifies maintenance and reduces duplication
  - Files: `src/moai_adk/templates/.claude/settings.json`

### Documentation

- **docs(hooks)**: Add hook development guidelines and sync local agents (38cd4d6c)
  - New section in `CLAUDE.local.md`: Hook Development Guidelines
  - Documents bash -l -c pattern usage
  - Provides examples for future hook development
  - File: `CLAUDE.local.md`

## Added

### Automatic Hook Migration

- **feat(update)**: Auto-update moai rank hook command during moai update (550be2f3)
  - `moai update` now automatically converts old hook commands to `bash -l -c` format
  - Existing users automatically benefit from PATH fix
  - No manual re-installation required
  - Displays confirmation message on successful update
  - File: `src/moai_adk/cli/commands/update.py`

## Technical Details

### Hook Command Pattern

**Before (Issue #296 - PATH not loaded)**:
```json
{
  "command": "python3 ~/.claude/hooks/moai/session_end__rank_submit.py"
}
```

**After (PATH correctly loaded)**:
```json
{
  "command": "bash -l -c 'python3 ~/.claude/hooks/moai/session_end__rank_submit.py'"
}
```

### Cross-Platform Compatibility

| Platform | Shell | Loaded Files | Status |
|----------|-------|--------------|--------|
| macOS | bash | `.bash_profile`, `.bashrc` | ✅ Fixed |
| macOS | zsh | `.zprofile`, `.zshenv` | ✅ Fixed |
| Linux | bash | `.bash_profile`, `.bashrc` | ✅ Fixed |
| WSL | bash | `.bash_profile`, `.bashrc` | ✅ Fixed |
| Windows Git Bash | bash | `.bash_profile`, `.bashrc` | ✅ Fixed |

### Files Modified

- `src/moai_adk/templates/.claude/settings.json`: All hooks updated with `bash -l -c`
- `src/moai_adk/templates/.claude/agents/moai/*.md`: 12 agents updated
- `src/moai_adk/rank/hook.py`: MoAI Rank hook installation updated
- `src/moai_adk/cli/commands/update.py`: Auto-migration logic added
- `CLAUDE.local.md`: Hook development guidelines added

## Migration Guide

### For New Users (Fresh Install)

```bash
# Install MoAI-ADK
uv tool install moai-adk

# Initialize project
moai init

# MoAI Rank install (optional)
moai rank install  # Hooks automatically use bash -l -c
```

### For Existing Users (Upgrade from v1.8.3 or earlier)

```bash
# Update MoAI-ADK
uv tool install moai-adk

# Update project templates - hooks automatically converted
moai update
```

**Note**: No manual intervention required. The `moai update` command automatically updates existing hook commands to use the `bash -l -c` pattern.

## Quality

- Smoke Tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 222 files unchanged
- Mypy: Success (no issues found in 174 source files)

## Installation & Update

```bash
# Update to the latest version
uv tool install moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.8.4 - WSL 및 크로스 플랫폼 호환성을 위한 Hook PATH 수정 (2026-01-25)

## 요약

이 패치 릴리스는 모든 hook에 로그인 셸 실행을 구현하여 **Issue #296** (moai update PATH 로딩 문제)을 완전히 해결하며, WSL, macOS, Linux, Windows Git Bash에서 PATH 환경 변수가 올바르게 로드되도록 보장합니다.

**주요 변경사항**:
- 모든 hook이 PATH 로딩을 위해 `bash -l -c` 래퍼 사용
- `moai update` 실행 시 MoAI Rank hook 자동 업데이트
- 유지보수 간소화를 위한 Settings.json 통합
- 완전한 크로스 플랫폼 호환성 (WSL/macOS/Linux/Windows)

**참조**: Issue #296

## 수정됨

### Hook PATH 로딩 문제

- **fix(hooks)**: PATH 로딩을 위해 로그인 셸 사용 (e1777b94)
  - 모든 템플릿 hook이 `bash -l -c` 래퍼 패턴 사용
  - `~/.bashrc`, `~/.zshrc`, `.bash_profile` 로드 보장
  - Non-interactive 셸에서 `command not found: uv` 에러 해결
  - 파일: `src/moai_adk/templates/.claude/settings.json`

- **fix(hooks)**: PATH 로딩을 위해 에이전트 hook을 로그인 셸로 업데이트 (80602d5d)
  - 모든 에이전트 hook 명령을 `bash -l -c` 래퍼로 업데이트
  - 영향받는 에이전트: expert-frontend, expert-security, expert-backend, manager-quality, builder-*
  - 파일: `src/moai_adk/templates/.claude/agents/moai/*.md` (12개 에이전트 업데이트)

- **fix(rank)**: bash 로그인 셸을 사용하도록 moai rank SessionEnd hook 업데이트 (84cf8f02)
  - MoAI Rank 설치 시 `bash -l -c` 래퍼로 hook 등록
  - 파일: `src/moai_adk/rank/hook.py`

## 변경됨

### Settings 구성

- **refactor(settings)**: 플랫폼별 설정을 통합 settings.json으로 통합 (c3020305)
  - 제거됨: `settings.json.unix`, `settings.json.windows`
  - 통합됨: 크로스 플랫폼 호환성을 갖춘 단일 `settings.json`
  - 유지보수 간소화 및 중복 제거
  - 파일: `src/moai_adk/templates/.claude/settings.json`

### 문서화

- **docs(hooks)**: hook 개발 가이드라인 추가 및 로컬 에이전트 동기화 (38cd4d6c)
  - `CLAUDE.local.md`에 새 섹션: Hook 개발 가이드라인
  - bash -l -c 패턴 사용 문서화
  - 향후 hook 개발을 위한 예제 제공
  - 파일: `CLAUDE.local.md`

## 추가됨

### 자동 Hook 마이그레이션

- **feat(update)**: moai update 실행 시 moai rank hook 명령 자동 업데이트 (550be2f3)
  - `moai update`가 기존 hook 명령을 `bash -l -c` 형식으로 자동 변환
  - 기존 사용자가 PATH 수정 자동 적용
  - 수동 재설치 불필요
  - 성공적인 업데이트 시 확인 메시지 표시
  - 파일: `src/moai_adk/cli/commands/update.py`

## 기술 세부사항

### Hook 명령 패턴

**이전 (Issue #296 - PATH 로드 안 됨)**:
```json
{
  "command": "python3 ~/.claude/hooks/moai/session_end__rank_submit.py"
}
```

**이후 (PATH 올바르게 로드됨)**:
```json
{
  "command": "bash -l -c 'python3 ~/.claude/hooks/moai/session_end__rank_submit.py'"
}
```

### 크로스 플랫폼 호환성

| 플랫폼 | 셸 | 로드되는 파일 | 상태 |
|--------|------|--------------|------|
| macOS | bash | `.bash_profile`, `.bashrc` | ✅ 수정됨 |
| macOS | zsh | `.zprofile`, `.zshenv` | ✅ 수정됨 |
| Linux | bash | `.bash_profile`, `.bashrc` | ✅ 수정됨 |
| WSL | bash | `.bash_profile`, `.bashrc` | ✅ 수정됨 |
| Windows Git Bash | bash | `.bash_profile`, `.bashrc` | ✅ 수정됨 |

### 수정된 파일

- `src/moai_adk/templates/.claude/settings.json`: 모든 hook을 `bash -l -c`로 업데이트
- `src/moai_adk/templates/.claude/agents/moai/*.md`: 12개 에이전트 업데이트
- `src/moai_adk/rank/hook.py`: MoAI Rank hook 설치 업데이트
- `src/moai_adk/cli/commands/update.py`: 자동 마이그레이션 로직 추가
- `CLAUDE.local.md`: Hook 개발 가이드라인 추가

## 마이그레이션 가이드

### 새 사용자 (새로 설치)

```bash
# MoAI-ADK 설치
uv tool install moai-adk

# 프로젝트 초기화
moai init

# MoAI Rank 설치 (선택사항)
moai rank install  # hook이 자동으로 bash -l -c 사용
```

### 기존 사용자 (v1.8.3 이하에서 업그레이드)

```bash
# MoAI-ADK 업데이트
uv tool install moai-adk

# 프로젝트 템플릿 업데이트 - hook 자동 변환
moai update
```

**참고**: 수동 개입이 필요하지 않습니다. `moai update` 명령이 기존 hook 명령을 `bash -l -c` 패턴으로 자동 업데이트합니다.

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 222개 파일 변경 없음
- Mypy: 성공 (174개 소스 파일에서 문제 없음)

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool install moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.8.3 - WSL Support Restoration & Cross-Platform Path Handling (2026-01-26)
