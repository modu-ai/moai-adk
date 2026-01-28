# v1.11.2 - CLAUDE.md Reference Fix (2026-01-29)

## Summary

Patch release fixing CLAUDE.md corruption issue during project initialization.

**Key Fix**:
- Disabled @path import processing during template copy
- @path references now preserved as-is in CLAUDE.md
- Claude Code handles @path references at runtime via `<system-reminder>` tags

**Impact**:
- Fixes CLAUDE.md corruption during `moai init`
- Aligns with Claude Code's standard @path reference behavior
- Simplified template processing logic

## Fixed

### Template Processing

- **fix(init)**: Disable @path import processing in CLAUDE.md copy (#308) (1f997b7a)
  - Issue: CLAUDE.md corrupted due to @path import expansion during template copy
  - Root cause: ClaudeMDImporter expanded @path references, causing circular/incomplete references
  - Fix: Removed import processing, preserve @path references as-is
  - Files affected:
    - `src/moai_adk/core/template/processor.py` (lines 1267-1314)
  - Impact: CLAUDE.md no longer corrupted, @path references work correctly at runtime

## Installation & Update

```bash
# Update to the latest version
claude install moai-adk

# Update project templates
moai update

# Verify version
moai --version
```

---

# v1.11.2 - CLAUDE.md 참조 수정 (2026-01-29)

## 요약

프로젝트 초기화 중 CLAUDE.md 손상 문제를 수정하는 패치 릴리스입니다.

**주요 수정**:
- 템플릿 복사 중 @path 가져오기 처리 비활성화
- @path 참조가 CLAUDE.md에 그대로 유지됨
- Claude Code가 런타임에 `<system-reminder>` 태그로 @path 참조 처리

**영향**:
- moai init 중 CLAUDE.md 손상 문제 해결
- Claude Code 표준 @path 참조 동작과 정렬
- 템플릿 처리 로직 단순화

## 수정됨

### 템플릿 처리

- **fix(init)**: CLAUDE.md 복사 시 @path 가져오기 처리 비활성화 (#308) (1f997b7a)
  - 문제: 템플릿 복사 중 @path 가져오기 확장으로 CLAUDE.md 손상 발생
  - 근본 원인: ClaudeMDImporter가 @path 참조를 확장하여 순환/불완전 참조 발생
  - 해결: 가져오기 처리 제거, @path 참조를 그대로 유지
  - 영향을 받는 파일:
    - `src/moai_adk/core/template/processor.py` (1267-1314행)
  - 영향: CLAUDE.md 손상 문제 해결, 런타임에 @path 참조가 올바르게 작동

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
claude install moai-adk

# 프로젝트 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.11.1 - CLAUDE.md English Only (2026-01-29)

## Summary

Patch release simplifying CLAUDE.md handling to always use English version.

**Key Change**:
- CLAUDE.md is now always English (instruction document)
- Removed language-specific CLAUDE files (ko/ja/zh)
- Simplified template processing logic

**Impact**:
- Consistent English documentation across all projects
- conversation_language setting still works for agent responses
- Users can create CLAUDE.local.md for personal instructions

## Breaking Changes

### CLAUDE.md Language Behavior Changed

- **Before**: `moai init` copies language-specific CLAUDE.md (ko/ja/zh) based on conversation_language
- **After**: `moai init` ALWAYS copies English CLAUDE.md regardless of conversation_language
- **Workaround**: Create CLAUDE.local.md for personalized instructions

## Fixed

### Template Processing

- **fix(cli)**: CLAUDE.md always English only, remove language-specific files (#307, #309) (e1a015f9)
  - Removed: CLAUDE.ko.md, CLAUDE.ja.md, CLAUDE.zh.md (998 lines)
  - Simplified `_copy_claude_md()` to always use English template
  - conversation_language config still works for agent responses
  - Files affected:
    - `src/moai_adk/core/template/processor.py` (lines 1267-1294)
    - `src/moai_adk/templates/CLAUDE.ko.md` (deleted)
    - `src/moai_adk/templates/CLAUDE.ja.md` (deleted)
    - `src/moai_adk/templates/CLAUDE.zh.md` (deleted)

## Installation & Update

```bash
# Update to the latest version
claude install moai-adk

# Update project templates
moai update

# Verify version
moai --version
```

---

# v1.11.1 - CLAUDE.md 영어 전용 (2026-01-29)

## 요약

CLAUDE.md 처리를 항상 영어 버전만 사용하도록 단순화하는 패치 릴리스입니다.

**주요 변경**:
- CLAUDE.md가 이제 항상 영어입니다 (지침 문서)
- 언어별 CLAUDE 파일 삭제 (ko/ja/zh)
- 템플릿 처리 로직 단순화

**영향**:
- 모든 프로젝트에서 일관된 영어 문서
- conversation_language 설정은 에이전트 응답에 여전히 적용됨
- CLAUDE.local.md를 사용하여 개인화된 지침 추가 가능

## Breaking Changes

### CLAUDE.md 언어 동작 변경

- **이전**: conversation_language에 따라 언어별 CLAUDE.md (ko/ja/zh) 복사
- **이후**: conversation_language 상관없이 항상 영어 CLAUDE.md 복사
- **해결책**: CLAUDE.local.md를 생성하여 개인화된 지침 추가

## 수정됨

### 템플릿 처리

- **fix(cli)**: CLAUDE.md 영어 전용, 언어별 파일 제거 (#307, #309) (e1a015f9)
  - 제거됨: CLAUDE.ko.md, CLAUDE.ja.md, CLAUDE.zh.md (998줄)
  - `_copy_claude_md()`를 단순화하여 항상 영어 템플릿 사용
  - conversation_language 설정은 에이전트 응답에 여전히 적용됨
  - 영향을 받는 파일:
    - `src/moai_adk/core/template/processor.py` (1267-1294행)
    - `src/moai_adk/templates/CLAUDE.ko.md` (삭제됨)
    - `src/moai_adk/templates/CLAUDE.ja.md` (삭제됨)
    - `src/moai_adk/templates/CLAUDE.zh.md` (삭제됨)

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
claude install moai-adk

# 프로젝트 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.11.0 - Template Variable Substitution Fix (2026-01-29)

## Summary

Patch release fixing template variable substitution during project initialization.

## 요약

프로젝트 초기화 중 템플릿 변수 치환 문제를 수정하는 패치 릴리스입니다.

**Key Fix**:
- Fixed template variables not being substituted during `moai init`
- Ensures proper cross-platform shell wrapper configuration

**주요 수정**:
- `moai init` 중 템플릿 변수가 치환되지 않는 문제 수정
- 모든 플랫폼에서 셸 래퍼 구성이 올바르게 작동하도록 개선

**Impact**:
- All template variables now correctly substituted during initialization
- Proper MCP server configuration on all platforms
- Consistent settings.json structure across platforms

**영향**:
- 초기화 중 모든 템플릿 변수가 올바르게 치환됨
- 모든 플랫폼에서 올바른 MCP 서버 구성
- 플랫폼 간 일관된 settings.json 구조

## Fixed / 수정됨

### Template Variable Substitution / 템플릿 변수 치환

- **fix(template)**: Resolve template variable substitution issue (#304) (25b87aef)
  - Issue: `{{HOOK_SHELL_PREFIX}}`, `{{HOOK_SHELL_SUFFIX}}`, `{{MCP_SHELL}}` not substituted during `moai init`
  - 문제: `moai init` 중 `{{HOOK_SHELL_PREFIX}}`, `{{HOOK_SHELL_SUFFIX}}`, `{{MCP_SHELL}}` 변수가 치환되지 않음
  - Root cause: `_merge_settings_json` called before variable substitution in `_copy_claude` method
  - 근본 원인: `_copy_claude` 메서드에서 변수 치환 전에 `_merge_settings_json`이 호출됨
  - Fix: Apply variable substitution BEFORE merging settings.json
  - 해결: settings.json 병합 전에 변수 치환 적용
    - Read template content and substitute variables
    - 템플릿 내용을 읽고 변수 치환
    - Write to temporary file for merging
    - 병합을 위해 임시 파일에 기록
    - Apply additional substitution after merge for remaining variables
    - 병합 후 남은 변수에 대해 추가 치환 적용
  - Impact: All shell wrapper variables now correctly substituted on all platforms
  - 영향: 모든 플랫폼에서 셸 래퍼 변수가 올바르게 치환됨
  - Files affected:
  - 영향을 받는 파일:
    - `src/moai_adk/core/template/processor.py` (lines 958-992)
    - `src/moai_adk/templates/.mcp.json` (formatting fix / 형식 수정)

### Test Improvements / 테스트 개선

- **test**: Fix pytest collection errors for release (1a63b4d8)
  - Removed obsolete test file: `test_context_manager_session_recovery.py`
  - 제거된 오래된 테스트 파일: `test_context_manager_session_recovery.py`
  - Fixed import in: `test_update_final.py`
  - 수정된 import: `test_update_final.py`

## Changed / 변경됨

### Documentation / 문서

- **docs**: Migrate /moai: commands to unified /moai syntax (15ea1dd8)
  - Deprecated `/moai:plan`, `/moai:run`, `/moai:sync` subcommands
  - Migrated to unified `/moai plan`, `/moai run`, `/moai sync` syntax
  - Updated all command references and skill definitions
  - 영어 번역된 README 파일 제거 (README.ko.md, README.ja.md, README.zh.md)
  - 모든 언어별 파일을 개별 리포지토리로 분리
  - 문서 정리 및 일관성 개선

- **docs**: Split CHANGELOG into separate language files (56fa9ecd)
  - Separated English and Korean CHANGELOG files
  - CHANGELOG.md (English), CHANGELOG.ko.md (Korean)
  - 문서 관리 효율성 개선

- **docs**: Remove Memory MCP section from translated CLAUDE files (02eb66c1)
  - Memory MCP 기능이 통합 스킬 시스템으로 마이그레이션되어 제거
  - Cleaned up obsolete references

- **chore**: Remove localized files from project root (66214271)
  - Removed: README.ko.md, README.ja.md, README.zh.md
  - 제거됨: 다국어 README 파일 (개별 리포지토리로 분리)

### Bug Fixes / 버그 수정

- **fix(statusline)**: Use template priority for statusLine to enable shell wrapper substitution (119c5b1c)
  - Fixed statusline template variable substitution
  - 템플릿 우선순위를 사용하여 셸 래퍼 치환 활성화

- **fix(types)**: Add type annotation for RTL_LANGUAGES (28713c33)
  - Added proper type hints for RTL language support
  - RTL 언어 지원을 위한 타입 힌트 추가

## Installation & Update / 설치 및 업데이트

```bash
# Update to the latest version / 최신 버전으로 업데이트
claude install moai-adk

# Update project templates / 프로젝트 템플릿 업데이트
moai update

# Verify version / 버전 확인
moai --version
```

---

# v1.10.5 - Template Variable Substitution Fix (2026-01-29)

## Summary

Patch release fixing template variable substitution during project initialization.

**Key Fix**:
- Fixed template variables not being substituted during `moai init`
- Ensures proper cross-platform shell wrapper configuration

**Impact**:
- All template variables now correctly substituted during initialization
- Proper MCP server configuration on all platforms
- Consistent settings.json structure across platforms

## Fixed

### Template Variable Substitution

- **fix(template)**: Resolve template variable substitution issue (#304) (25b87aef)
  - Issue: `{{HOOK_SHELL_PREFIX}}`, `{{HOOK_SHELL_SUFFIX}}`, `{{MCP_SHELL}}` not substituted during `moai init`
  - Root cause: `_merge_settings_json` called before variable substitution in `_copy_claude` method
  - Fix: Apply variable substitution BEFORE merging settings.json
    - Read template content and substitute variables
    - Write to temporary file for merging
    - Apply additional substitution after merge for remaining variables
  - Impact: All shell wrapper variables now correctly substituted on all platforms
  - Files affected:
    - `src/moai_adk/core/template/processor.py` (lines 958-992)
    - `src/moai_adk/templates/.mcp.json` (formatting fix)

### Code Cleanup

- **chore**: Remove unused hook library modules (25b87aef)
  - Removed: `checkpoint.py`, `language_detector.py`, `language_validator.py`, `timeout.py`
  - Consolidated into `unified_timeout_manager.py`
  - Reduced codebase by ~1,000 lines

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates
moai update

# Verify version
moai --version
```

---

# v1.9.0 - Memory MCP, SVG Skill, Rules Migration (2026-01-26)

## Summary

Minor release introducing persistent memory across sessions, comprehensive SVG skill, and standards-compliant rules system migration.

**Key Features**:
- **Memory MCP Integration**: Persistent storage for user preferences and project context
- **SVG Skill**: Comprehensive skill with SVGO optimization patterns and best practices
- **Rules Migration**: Migrated from `.moai/rules/*.yaml` to `.claude/rules/*.md` (Claude Code official standard)
- **Bug Fix**: Rank batch sync display issue (#300)

**Impact**:
- Enables agent-to-agent context sharing via Memory MCP
- Professional SVG creation and optimization support
- Cleaner, standards-compliant project structure
- Accurate batch sync statistics display

## Breaking Changes

None. All changes are backward compatible.

## Added

### SVG Creation and Optimization Skill

- **feat**: Add `moai-tool-svg` skill (54c12a85)
  - Based on W3C SVG 2.0 specification and SVGO documentation
  - Comprehensive modules: basics, styling, optimization, animation
  - 12 working code examples
  - SVGO configuration patterns and best practices
  - 3,698 lines total (SKILL.md: 410, modules: 2,288, examples: 500, reference: 500)

### Language Rules Enhancement

- **feat**: Update language rules with enhanced tooling information (54c12a85)
  - Ruff configuration patterns (replaces flake8+isort+pyupgrade)
  - Mypy strict mode guidelines
  - Testing framework recommendations
  - 16 language files updated

## Changed

### CLAUDE.md Optimization

- **refactor**: Major cleanup and modularization for v1.9.0 (4134e60d)
  - Reduced CLAUDE.md from ~60k to ~30k characters (40k limit compliance)
  - Moved detailed content to `.claude/rules/` for better organization
  - Added `shell_validator.py` utility for cross-platform compatibility
  - Enhanced CLI commands (doctor, init, update)
  - Added `moai-workflow-thinking` skill
  - Added bug-report.yml issue template
  - Impact: Improved readability, maintainability, and Claude Code compatibility

### Rules System Migration

- **feat**: Migrate from `.moai/rules/*.yaml` to `.claude/rules/*.md` (99ab5273)
  - Deleted: 6,959 lines of YAML rules
  - Added: Claude Code official Markdown rules
  - Structure: `.claude/rules/{core,development,workflow,languages}/`
  - Impact: Standards compliance, cleaner organization

## Fixed

### Rank Command

- **fix(rank)**: Correctly parse nested API response for batch sync (#300) (31b504ed)
  - Issue: `moai-adk rank sync` always showed "Submitted: 0"
  - Root cause: Missing nested `data` field extraction
  - Fix: Added `data = response.get("data", {})` before accessing fields
  - Impact: Accurate submission statistics display

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.8.13 - Statusline Context Window Fix (2026-01-26)

## Summary

Patch release improving statusline context window calculation accuracy.

**Key Fix**:
- Fixed statusline context window percentage to use Claude Code's pre-calculated values

**Impact**:
- Context window display now accounts for auto-compact and output token reservation
- More accurate remaining token information

## Fixed

### Statusline Context Window Calculation

- **fix(statusline)**: Use Claude Code's pre-calculated context percentages (2dacecb7)
  - Priority 1: Use `used_percentage`/`remaining_percentage` from Claude Code (most accurate)
  - Priority 2: Calculate from `current_usage` tokens (fallback)
  - Priority 3: Return 0% when no data available (session start)
  - Ensures accuracy when auto-compact is enabled or output tokens are reserved
  - Files: `src/moai_adk/statusline/main.py`

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates
moai update

# Verify version
moai --version
```

---

# v1.8.12 - Hook Format Update & Login Command (2026-01-26)

## Summary

Patch release with Claude Code hook format compatibility fix and UX improvements.

**Key Changes**:
- Fixed Claude Code settings.json hook format (new matcher-based structure)
- Renamed `moai rank register` to `moai rank login` (more intuitive)
- settings.json now always overwritten on update; use settings.local.json for customizations

**Impact**:
- MoAI Rank hooks now work with latest Claude Code
- `moai rank login` is the new primary command (register still works as alias)
- User customizations preserved in settings.local.json

## Breaking Changes

None. `moai rank register` still works as a hidden alias.

## Fixed
