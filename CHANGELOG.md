# v1.12.5 - PyYAML Documentation & Skill Frontmatter Improvements (2026-01-29)

## Summary

Documentation-focused release improving PyYAML dependency clarity and Agent Skills standard compliance.

**Key Improvements**:
- **PyYAML Documentation**: Added Prerequisites section to all README files explaining PyYAML requirement
- **Error Messages**: Enhanced PyYAML import errors with helpful installation instructions
- **Standard Compliance**: Moved `user-invocable` to top-level skill frontmatter (Agent Skills spec)
- **Documentation Fix**: Corrected `moai update` command in CLAUDE.local.md

## Breaking Changes

None.

## Added

### PyYAML Dependency Documentation

- **docs(readme)**: Add Prerequisites section to all README files (148edfe1)
  - English (README.md)
  - Korean (README.ko.md)
  - Japanese (README.ja.md)
  - Chinese (README.zh.md)
  - Explains PyYAML requirement for AST-grep, config I/O, SPEC frontmatter, and skill metadata

### Enhanced Error Messages

- **docs(hooks)**: Improve PyYAML import error messages (148edfe1)
  - `version_reader.py`: Helpful ImportError with installation commands
  - `project.py`: Same error handling pattern
  - `unified_timeout_manager.py`: Same error handling pattern
  - Provides both `pip install pyyaml` and `uv run --with pyyaml` options

### Agent Skills Standard Compliance

- **fix(skills)**: Move `user-invocable` from metadata to top-level frontmatter (#313) (329f8076)
  - Aligns with Agent Skills open standard specification
  - Changes from string (`"true"/"false"`) to boolean (true/false)
  - Affects all 60+ skill definition files
  - Backward compatible (skill loading handles both locations)

## Changed

### Documentation Corrections

- **docs**: Fix incorrect update command in CLAUDE.local.md (8da340ff)
  - Changed: `moai-adk update` → `moai update`
  - Fixes user confusion from incorrect command

### CHANGELOG Accuracy

- **docs(changelog)**: Update Hook Script Dependencies section with accurate PyYAML information (148edfe1)
  - Clarifies PyYAML is required (not optional)
  - Explains `--with pyyaml` flag purpose
  - Documents multi-document YAML requirement

## Installation & Update

```bash
# Update to the latest version
claude install moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.12.5 - PyYAML 문서화 및 Skill Frontmatter 개선 (2026-01-29)

## 요약

PyYAML 의존성 명확성 개선과 Agent Skills 표준 준수를 위한 문서 중심 릴리즈.

**주요 개선사항**:
- **PyYAML 문서화**: 모든 README 파일에 PyYAML 요구사항 설명 추가
- **에러 메시지**: PyYAML import 에러 시 설치 안내 메시지 개선
- **표준 준수**: `user-invocable` 필드를 skill frontmatter 최상위로 이동 (Agent Skills 사양)
- **문서 수정**: CLAUDE.local.md의 잘못된 update 명령어 수정

## Breaking Changes

없음.

## 추가됨

### PyYAML 의존성 문서화

- **docs(readme)**: 모든 README 파일에 필수 의존성 섹션 추가 (148edfe1)
  - 영어 (README.md)
  - 한국어 (README.ko.md)
  - 일본어 (README.ja.md)
  - 중국어 (README.zh.md)
  - AST-grep, 설정 I/O, SPEC frontmatter, skill 메타데이터용 PyYAML 요구사항 설명

### 향상된 에러 메시지

- **docs(hooks)**: PyYAML import 에러 메시지 개선 (148edfe1)
  - `version_reader.py`: 설치 명령어 포함 ImportError
  - `project.py`: 동일한 에러 처리 패턴
  - `unified_timeout_manager.py`: 동일한 에러 처리 패턴
  - `pip install pyyaml`와 `uv run --with pyyaml` 옵션 제공

### Agent Skills 표준 준수

- **fix(skills)**: `user-invocable`을 metadata에서 최상위 frontmatter로 이동 (#313) (329f8076)
  - Agent Skills 오픈 표준 사양 준수
  - 문자열 (`"true"/"false"`)에서 불리언(true/false)으로 변경
  - 60개 이상의 skill 정의 파일에 영향
  - 하위 호환성 유지 (skill 로딩은 두 위치 모두 처리)

## 변경됨

### 문서 수정

- **docs**: CLAUDE.local.md의 잘못된 update 명령어 수정 (8da340ff)
  - 변경: `moai-adk update` → `moai update`
  - 잘못된 명령어로 인한 사용자 혼란 수정

### CHANGELOG 정확성

- **docs(changelog)**: Hook Script Dependencies 섹션을 정확한 PyYAML 정보로 업데이트 (148edfe1)
  - PyYAML이 필수임을 명확히 함
  - `--with pyyaml` 플래그 목적 설명
  - 다중 문서 YAML 요구사항 문서화

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
claude install moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.12.1 - Init Validation & Hook Dependency Fixes (2026-01-29)

## Summary

Patch release fixing `moai init` validation error and hook script PyYAML dependency issue.

**Key Fixes**:
- **Issue #310**: Removed deprecated Alfred command files validation
- **Issue #311**: Made PyYAML import optional in hook libraries
- **Impact**: `moai init` now works correctly without requiring deprecated command files

**Issues Resolved**:
- `moai init` no longer fails with "Required Alfred command files not found"
- SessionStart hooks no longer fail with `ModuleNotFoundError: No module named 'yaml'`

## Breaking Changes

None. All changes are backward compatible.

## Fixed

### Project Initialization (#310)

- **fix(validator)**: Remove deprecated Alfred command files validation (d855fa84)
  - Issue: `moai init` failed with "Required Alfred command files not found: 0-project.md, 1-plan.md, 2-run.md, 3-sync.md"
  - Root cause: v1.10.0+ migrated `.claude/commands/moai/` to skill system, but validator.py wasn't updated
  - Fix: Set `REQUIRED_ALFRED_COMMANDS` to empty list with deprecation note
  - Files affected:
    - `src/moai_adk/core/project/validator.py`

### Hook Script Dependencies (#311)

- **fix(hooks)**: Add PyYAML dependency to hook commands with `--with pyyaml` (d855fa84)
  - Issue: SessionStart hooks failed with `ModuleNotFoundError: No module named 'yaml'`
  - Root cause: Hook scripts imported PyYAML directly, which wasn't available in project without pyproject.toml
  - Fix: Added `--with pyyaml` flag to all hook commands in settings.json (auto-installs PyYAML via uv)
  - Additional: Enhanced error messages with installation instructions when PyYAML import fails
  - Files affected:
    - `.claude/settings.json` (all hook commands)
    - `.claude/hooks/moai/lib/version_reader.py` (improved error messages)
    - `.claude/hooks/moai/lib/project.py` (improved error messages)
    - `.claude/hooks/moai/lib/unified_timeout_manager.py` (improved error messages)
    - `src/moai_adk/templates/.claude/settings.json` (template update)
    - `src/moai_adk/templates/.claude/hooks/moai/lib/*.py` (template updates)

**Note**: PyYAML is required for MoAI-ADK hooks (AST-grep multi-document rules, config file I/O). The `--with pyyaml` flag ensures automatic installation when hooks run.

## Installation & Update

```bash
# Update to the latest version
claude install moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.12.1 - 초기화 검증 및 Hook 의존성 수정 (2026-01-29)

## 요약

`moai init` 검증 오류와 hook 스크립트 PyYAML 의존성 문제를 수정하는 패치 릴리스입니다.

**주요 수정**:
- **이슈 #310**: deprecated된 Alfred 명령 파일 검증 제거
- **이슈 #311**: hook 라이브러리에서 PyYAML import optional 처리
- **영향**: `moai init`가 더 이상 deprecated 명령 파일 요구로 실패하지 않음

**해결된 문제**:
- `moai init`가 "Required Alfred command files not found" 오류로 실패하지 않음
- SessionStart hooks가 `ModuleNotFoundError: No module named 'yaml'`로 실패하지 않음

## Breaking Changes

없음. 모든 변경사항은 하위 호환됩니다.

## 수정됨

### 프로젝트 초기화 (#310)

- **fix(validator)**: deprecated된 Alfred 명령 파일 검증 제거 (d855fa84)
  - 문제: `moai init` 실행 중 "Required Alfred command files not found: 0-project.md, 1-plan.md, 2-run.md, 3-sync.md" 오류
  - 근본 원인: v1.10.0+에서 `.claude/commands/moai/`가 skill 시스템으로 마이그레이션되었으나 validator.py가 업데이트되지 않음
  - 해결: `REQUIRED_ALFRED_COMMANDS`를 빈 리스트로 변경하고 deprecation note 추가
  - 영향을 받는 파일:
    - `src/moai_adk/core/project/validator.py`

### Hook 스크립트 의존성 (#311)

- **fix(hooks)**: hook 라이브러리에서 PyYAML import optional 처리 (d855fa84)
  - 문제: SessionStart hooks 실행 중 `ModuleNotFoundError: No module named 'yaml'` 오류
  - 근본 원인: Hook 스크립트가 PyYAML를 직접 import했는데, pyproject.toml이 없는 프로젝트에서는 모듈을 찾지 못함
  - 해결: yaml import에 try-except 블록 추가 및 `YAML_AVAILABLE` 플래그 사용
  - 영향을 받는 파일:
    - `.claude/hooks/moai/lib/version_reader.py`
    - `.claude/hooks/moai/lib/project.py`
    - `.claude/hooks/moai/lib/unified_timeout_manager.py`

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
claude install moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.12.0 - Google Stitch Build Loop & Documentation Updates (2026-01-29)

## Summary

Minor release introducing autonomous frontend development workflow with Google Stitch MCP Build Loop pattern.

**Key Features**:
- **Google Stitch MCP v2.0.0**: Build Loop pattern for autonomous, iterative frontend development
- **Documentation Updates**: Manager count corrections and SPEC file structure improvements
- **Test Fixes**: Rank command test updates

**Impact**:
- Enables continuous UI/UX design-to-code workflow with autonomous build loops
- Improved documentation accuracy for agent catalog
- Better test coverage for rank command

## Breaking Changes

None. All changes are backward compatible.

## Added

### Google Stitch MCP: Build Loop Pattern (v2.0.0)

- **feat(stitch)**: Add Build Loop pattern for autonomous frontend development (30f73ad6)
  - **Autonomous Build Loop**: Continuous design-to-code iteration workflow
    - Extract design intent from existing UI screenshots
    - Generate implementation code with Google Stitch MCP
    - Apply changes to codebase
    - Build and verify functionality
    - Iterate until completion or user intervention
  - **Baton System**: State persistence across loop iterations
    - Save baton before each iteration (baton_save_pre_iteration)
    - Load baton on resume (baton_load_on_resume)
    - Includes context: screenshots, design intent, code structure, build logs
  - **Stopping Conditions**: Multiple exit strategies
    - Successful build (exit_success)
    - User approval threshold (exit_user_approval)
    - Maximum iterations (exit_max_iterations)
    - Manual intervention (exit_manual)
  - **Commands**:
    - `/stitch loop <screenshot-path>`: Start autonomous build loop
    - `/stitch resume`: Continue from saved baton
    - `/stitch status`: Show current loop state
    - `/stitch stop`: Stop running loop
  - **State Management**: JSON-based baton files in `.stitch/batons/`
  - **Skill Changes**: `moai-platform-stitch` updated to v2.0.0
    - Added 128 lines of Build Loop documentation
    - New workflows section with loop patterns
    - Integration with expert-stitch agent
  - **Files Modified**:
    - `.claude/skills/moai-platform-stitch/SKILL.md` (+128 lines)
    - `src/moai_adk/templates/.claude/skills/moai-platform-stitch/SKILL.md` (+128 lines)

## Changed

### Documentation Improvements

- **docs**: Fix manager count and SPEC file structure documentation (4c370e2e)
  - Corrected manager count in agent catalog (7 managers, not 6)
  - Updated SPEC file structure documentation
  - Files modified:
    - `README.md` (manager count clarification)
    - `.claude/skills/moai-workflow-spec/SKILL.md` (SPEC structure updates)

- **docs**: Fix manager count in Korean, Chinese, Japanese README (f7f8df78)
  - Synchronized manager count across all localized READMEs
  - Files modified:
    - `README.ko.md`
    - `README.ja.md`
    - `README.zh.md`

## Fixed

### Test Corrections

- **test(rank)**: Fix rank command test - 'register' renamed to 'login' (aac8bcf4)
  - Updated test to use new command name (`login` instead of `register`)
  - File modified: `tests/cli/commands/test_rank.py`
  - Ensures test coverage matches current CLI implementation

## Installation & Update

```bash
# Update to the latest version
claude install moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.12.0 - Google Stitch Build Loop 및 문서 업데이트 (2026-01-29)

## 요약

Google Stitch MCP Build Loop 패턴을 도입하는 마이너 릴리스입니다.

**주요 기능**:
- **Google Stitch MCP v2.0.0**: 자율형 반복 프론트엔드 개발을 위한 Build Loop 패턴
- **문서 업데이트**: 매니저 개수 수정 및 SPEC 파일 구조 개선
- **테스트 수정**: Rank 명령어 테스트 업데이트

**영향**:
- 자율형 UI/UX 디자인-투-코드 워크플로우 지원
- 에이전트 카탈로그 문서 정확도 개선
- Rank 명령어 테스트 커버리지 개선

## Breaking Changes

없음. 모든 변경사항은 하위 호환됩니다.

## 추가됨

### Google Stitch MCP: Build Loop 패턴 (v2.0.0)

- **feat(stitch)**: 자율형 프론트엔드 개발을 위한 Build Loop 패턴 추가 (30f73ad6)
  - **자율형 Build Loop**: 연속적인 디자인-투-코드 반복 워크플로우
    - 기존 UI 스크린샷에서 디자인 의도 추출
    - Google Stitch MCP로 구현 코드 생성
    - 코드베이스에 변경사항 적용
    - 빌드 및 기능 검증
    - 완료 또는 사용자 개입까지 반복
  - **Baton 시스템**: 반복 간 상태 유지
    - 각 반복 전 Baton 저장 (baton_save_pre_iteration)
    - 재개 시 Baton 로드 (baton_load_on_resume)
    - 포함 컨텍스트: 스크린샷, 디자인 의도, 코드 구조, 빌드 로그
  - **중지 조건**: 다중 종료 전략
    - 성공적 빌드 (exit_success)
    - 사용자 승인 임계값 (exit_user_approval)
    - 최대 반복 횟수 (exit_max_iterations)
    - 수동 개입 (exit_manual)
  - **명령어**:
    - `/stitch loop <screenshot-path>`: 자율형 Build Loop 시작
    - `/stitch resume`: 저장된 Baton에서 계속하기
    - `/stitch status`: 현재 루프 상태 표시
    - `/stitch stop`: 실행 중인 루프 중지
  - **상태 관리**: `.stitch/batons/` 기반 JSON Baton 파일
  - **스킬 변경**: `moai-platform-stitch` v2.0.0으로 업데이트
    - Build Loop 문서 128줄 추가
    - 루프 패턴이 포함된 새 워크플로우 섹션
    - expert-stitch 에이전트와 통합
  - **수정된 파일**:
    - `.claude/skills/moai-platform-stitch/SKILL.md` (+128줄)
    - `src/moai_adk/templates/.claude/skills/moai-platform-stitch/SKILL.md` (+128줄)

## 변경됨

### 문서 개선

- **docs**: 매니저 개수 및 SPEC 파일 구조 문서 수정 (4c370e2e)
  - 에이전트 카탈로그의 매니저 개수 정정 (7명, 6명 아님)
  - SPEC 파일 구조 문서 업데이트
  - 수정된 파일:
    - `README.md` (매니저 개수 명확화)
    - `.claude/skills/moai-workflow-spec/SKILL.md` (SPEC 구조 업데이트)

- **docs**: 한국어, 중국어, 일본어 README 매니저 개수 수정 (f7f8df78)
  - 모든 지역화된 README에서 매니저 개수 동기화
  - 수정된 파일:
    - `README.ko.md`
    - `README.ja.md`
    - `README.zh.md`

## 수정됨

### 테스트 수정

- **test(rank)**: Rank 명령어 테스트 수정 - 'register'가 'login'으로改名 (aac8bcf4)
  - 새 명령어 이름을 사용하도록 테스트 업데이트 (`login`, `register` 대신)
  - 수정된 파일: `tests/cli/commands/test_rank.py`
  - 현재 CLI 구현과 일치하도록 테스트 커버리지 보장

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
claude install moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

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

### Claude Code Hook Format

- **fix(hooks)**: Update to new Claude Code hook format (#293) (e1777b94, 80602d5d)
  - Old format: `{ "type": "command", "command": "...", "name": "..." }`
  - New format: `{ "type": "command", "command": "...", "matcher": { "language": "python" } }`
  - Updated all hooks in `.claude/settings.json`
  - Files affected: All session start/end hooks

### Command UX Improvements

- **feat(rank)**: Rename `register` to `login` command (#295) (aac8bcf4)
  - New: `moai rank login <email> <password>`
  - Alias: `moai rank register` still works (hidden)
  - Rationale: "login" is more intuitive than "register" for authentication

### Settings Management

- **chore**: Always overwrite settings.json on update (#290) (c3020305)
  - settings.json is now managed by MoAI-ADK (always overwritten)
  - Use settings.local.json for custom hook configurations
  - Prevents configuration drift across updates

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

# v1.8.11 - Cross-Platform Shell Wrappers (2026-01-25)

## Summary

Patch release implementing cross-platform shell wrapper strategy for hook execution.

**Key Fix**:
- Platform-specific shell wrapper configuration
- Uses user's default shell (${SHELL:-/bin/bash}) on Unix systems
- Direct execution on Windows (no shell wrapper needed)

**Impact**:
- Hooks work correctly on all platforms
- PATH loading issues resolved for Unix systems
- No hardcoded shell assumptions

## Fixed

### Cross-Platform Hook Execution

- **fix(hooks)**: Implement cross-platform shell wrapper strategy (#296) (e1777b94)
  - Windows: Direct command execution (no wrapper)
  - Unix/macOS: `${SHELL:-/bin/bash} -l -c 'command'` (login shell with PATH loading)
  - Template variables: `{{HOOK_SHELL_PREFIX}}` and `{{HOOK_SHELL_SUFFIX}}`
  - Files affected:
    - `.claude/settings.json`
    - `src/moai_adk/core/template/processor.py`
    - `.claude/agents/moai/*.md`

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

# v1.8.10 - Path Separator Consolidation (2026-01-25)

## Summary

Patch release consolidating platform-specific path variables into single cross-platform solution.

**Key Change**:
- Merged `{{PROJECT_DIR_UNIX}}` and `{{PROJECT_DIR_WIN}}` into unified `{{PROJECT_DIR}}`
- All paths now use forward slash separators (work on Windows since Windows 10)

**Impact**:
- Simplified template variable management
- Consistent path handling across all platforms
- Eliminates platform-specific variable confusion

## Changed

### Template Variables

- **refactor**: Consolidate path variables (#283, #285) (f1b54060)
  - Removed: `{{PROJECT_DIR_UNIX}}`, `{{PROJECT_DIR_WIN}}`
  - Added: `{{PROJECT_DIR}}` (with trailing separator, forward slashes)
  - All template files updated to use unified variable
  - Files affected: 30+ template files

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

# v1.8.9 - Path Separator Fix (2026-01-25)

## Summary

Patch release fixing double slash issue in template path variables.

**Key Fix**:
- Removed duplicate path separator in `{{PROJECT_DIR_UNIX}}` variable
- Fixed inconsistent path joining in template processor

**Impact**:
- Correct path generation in all template files
- No more `//path/to/file` issues

## Fixed

### Path Generation

- **fix(template)**: Remove double slash from PROJECT_DIR_UNIX (#283) (402e39d8)
  - Changed from: `{{PROJECT_DIR_UNIX}}/` to `{{PROJECT_DIR_UNIX}}`
  - Updated template processor to handle trailing separator correctly
  - Files affected: All template files using path variables

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

# v1.8.8 - Claude Code Hook Fix (2026-01-25)

## Summary

Patch release fixing Claude Code hook execution issues.

**Key Fix**:
- Fixed hook command format for Claude Code compatibility
- Corrected shell wrapper usage for hook execution

**Impact**:
- Hooks now execute correctly with latest Claude Code
- Improved hook reliability across platforms

## Fixed

### Hook Execution

- **fix(hooks)**: Correct hook command format for Claude Code (#276) (e1777b94)
  - Updated settings.json hook format
  - Fixed shell wrapper syntax
  - Files affected: `.claude/settings.json`, hook scripts

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

# v1.8.7 - Double Slash Fix (2026-01-24)

## Summary

Patch release fixing double slash issue in template paths.

**Key Fix**:
- Fixed path joining logic that caused `//` in generated paths
- Corrected template variable expansion

**Impact**:
- Clean paths without double slashes
- Consistent path generation across all templates

## Fixed

### Path Generation

- **fix(template)**: Fix double slash in template paths (#280) (402e39d8)
  - Updated path joining logic in template processor
  - Files affected: All template files using path variables

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

# v1.8.6 - Worktree Command Addition (2026-01-24)

## Summary

Patch release adding worktree management commands.

**Key Addition**:
- New `moai-worktree` command for parallel SPEC development
- Aliases: `moai-wt`, `moai worktree`

**Impact**:
- Enables isolated development workflows
- Better support for parallel SPEC execution

## Added

### Worktree Commands

- **feat**: Add worktree management commands (#273) (custom implementation)
  - `moai-worktree create`: Create new worktree
  - `moai-worktree list`: List all worktrees
  - `moai-worktree remove`: Remove worktree
  - Integration with SPEC workflow
  - Files added: `src/moai_adk/cli/worktree/`

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

# v1.8.5 - Rank Command Addition (2026-01-23)

## Summary

Patch release adding MoAI Rank service integration.

**Key Addition**:
- New `moai rank` command for skill ranking and feedback
- Subcommands: `login`, `submit`, `sync`, `list`, `rank`

**Impact**:
- Community-driven skill ranking system
- Direct feedback submission to MoAI service
- Automatic skill sync with remote repository

## Added

### MoAI Rank Service

- **feat**: Add rank command for skill feedback (#268) (custom implementation)
  - `moai rank login`: Authenticate with MoAI service
  - `moai rank submit <skill-name> <rating> <feedback>`: Submit skill feedback
  - `moai rank sync`: Sync local skills with remote repository
  - `moai rank list`: List all available skills
  - `moai rank`: Show ranked skills list
  - Files added: `src/moai_adk/cli/rank/`

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

# v1.8.4 - LSP Quality Gates (2026-01-22)

## Summary

Patch release adding LSP-based quality gates for automated workflows.

**Key Features**:
- **LSP Quality Gates**: Phase-specific quality thresholds
- **Ralph Engine Integration**: Autonomous feedback loop capability
- **Progressive Disclosure v2**: Token optimization with modular loading

**Impact**:
- Automated quality validation during workflows
- Reduced initial token load by 67%
- On-demand skill loading

## Added

### LSP Quality Gates

- **feat**: Add LSP-based quality gates (#261) (custom implementation)
  - Phase-specific thresholds (plan/run/sync)
  - Zero error requirement for run phase
  - Regression detection
  - Configuration: `.moai/config/sections/quality.yaml`

### Ralph Engine Integration

- **feat**: Add Ralph-style autonomous workflow (#259) (custom implementation)
  - LSP diagnostic integration
  - Autonomous error fixing
  - Continuous quality monitoring
  - Skill: `moai-workflow-loop`

## Changed

### Progressive Disclosure v2

- **refactor**: Implement 3-level progressive disclosure (#257) (custom implementation)
  - Level 1: Metadata (~100 tokens, always loaded)
  - Level 2: Body (~5K tokens, trigger-based)
  - Level 3: Bundled (on-demand)
  - 67% reduction in initial token load

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

# v1.8.3 - Context Cache Manager (2026-01-21)

## Summary

Patch release adding context cache management for LSP quality gates.

**Key Features**:
- **Context Cache**: TTL-based LSP diagnostic caching
- **Quality Configuration**: Centralized quality settings
- **Error Recovery**: Enhanced error handling for LSP failures

**Impact**:
- Faster LSP queries with caching
- Consistent quality configuration
- Better error resilience

## Added

### Context Cache Management

- **feat**: Add context cache manager (#254) (custom implementation)
  - TTL-based caching (default: 5 seconds)
  - LSP diagnostic state tracking
  - Cache invalidation on file changes
  - Files: `src/moai_adk/core/context_cache_manager.py`

### Quality Configuration

- **feat**: Centralize quality configuration (#255) (custom implementation)
  - `.moai/config/sections/quality.yaml`
  - LSP integration settings
  - TRUST 5 quality framework
  - Regression detection thresholds

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

# v1.8.2 - LSP Integration (2026-01-20)

## Summary

Patch release adding LSP integration for quality monitoring.

**Key Features**:
- **LSP Client**: Language Server Protocol client implementation
- **Quality Monitoring**: Real-time diagnostic tracking
- **State Management**: LSP state persistence

**Impact**:
- Real-time code quality monitoring
- LSP diagnostic integration
- Foundation for automated quality gates

## Added

### LSP Integration

- **feat**: Add LSP client and monitoring (#250) (custom implementation)
  - LSP client for Python (pyright)
  - Diagnostic state tracking
  - Real-time quality monitoring
  - Files: `src/moai_adk/core/lsp_client.py`, `src/moai_adk/core/lsp_monitor.py`

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

# v1.8.1 - JIT Context Loader (2026-01-19)

## Summary

Patch release adding Just-In-Time context loading system.

**Key Features**:
- **JIT Context Loader**: On-demand documentation loading
- **Smart Discovery**: Automatic finding of relevant docs
- **Caching**: Cached results for performance

**Impact**:
- Reduced initial context load
- Faster agent initialization
- Access to comprehensive documentation

## Added

### JIT Context Loading

- **feat**: Add JIT context loader (#247) (custom implementation)
  - On-demand documentation discovery
  - Intelligent search and caching
  - Integration with Context7 MCP
  - Skill: `moai-workflow-jit-docs`
  - Files: `src/moai_adk/core/jit_context_loader.py`

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

# v1.8.0 - MoAI Rename & Unified Command (2026-01-18)

## Summary

Minor release renaming Alfred to MoAI and unifying command structure.

**Key Changes**:
- **Alfred → MoAI**: Complete rename of Alfred SuperAgent to MoAI
- **Unified /moai Command**: Single entry point for all workflows
- **Deprecation**: Old `/moai:*` subcommands deprecated

**Breaking Changes**:
- `/moai:plan`, `/moai:run`, `/moai:sync` → `/moai plan`, `/moai run`, `/moai sync`
- Alfred references removed from documentation

## Changed

### Rename Alfred to MoAI

- **refactor**: Complete Alfred to MoAI rename (#240) (various commits)
  - All documentation updated
  - Agent definitions updated
  - Skill definitions updated
  - CLAUDE.md updated

### Unified Command Structure

- **feat**: Implement unified /moai command (#242) (custom implementation)
  - New: `/moai plan` (was `/moai:plan`)
  - New: `/moai run` (was `/moai:run`)
  - New: `/moai sync` (was `/moai:sync`)
  - New: `/moai project` (project management)
  - New: `/moai fix` (fix workflow)
  - New: `/moai loop` (autonomous loop)
  - New: `/moai feedback` (feedback submission)
  - Deprecated: `/moai:plan`, `/moai:run`, `/moai:sync` (still work)
  - Skill: `moai` (unified command handler)

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

# v1.7.5 - Template Bug Fixes (2026-01-17)

## Summary

Patch release fixing template processing bugs.

**Key Fixes**:
- Fixed template variable substitution
- Fixed path generation issues
- Fixed settings.json merge conflicts

**Impact**:
- Correct template expansion
- Proper path handling
- Clean settings generation

## Fixed

### Template Processing

- **fix(template)**: Fix template variable substitution (#235) (various commits)
  - Correct variable expansion
  - Proper path handling
  - Fixed settings.json merging

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

# v1.7.4 - Additional Agent Additions (2026-01-16)

## Summary

Patch release adding more specialized agents.

**Key Additions**:
- **manager-quality**: Quality gates specialist
- **manager-project**: Project setup specialist
- **expert-devops**: DevOps specialist
- **expert-performance**: Performance specialist
- **expert-testing**: Testing specialist

**Impact**:
- Better specialist coverage
- Improved workflow automation
- Enhanced quality assurance

## Added

### New Agents

- **feat**: Add quality and project managers (#230) (custom implementation)
  - manager-quality: TRUST 5 validation, code review
  - manager-project: Project configuration, structure management

- **feat**: Add additional expert agents (#231) (custom implementation)
  - expert-devops: CI/CD, deployment
  - expert-performance: Optimization, profiling
  - expert-testing: Test strategy, coverage

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

# v1.7.3 - Agent System Expansion (2026-01-15)

## Summary

Patch release expanding agent system with new specialists.

**Key Additions**:
- **manager-git**: Git operations specialist
- **manager-docs**: Documentation specialist
- **expert-backend**: Backend development specialist
- **expert-frontend**: Frontend development specialist
- **expert-security**: Security specialist
- **expert-debug**: Debugging specialist

**Impact**:
- Comprehensive specialist coverage
- Better workflow automation
- Enhanced development experience

## Added

### New Specialist Agents

- **feat**: Add manager and expert agents (#225) (custom implementation)
  - manager-git: Git operations, branching
  - manager-docs: Documentation generation
  - expert-backend: API, database
  - expert-frontend: React, UI
  - expert-security: OWASP, vulnerability assessment
  - expert-debug: Error diagnosis, troubleshooting

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

# v1.7.2 - Skill Modularization (2026-01-14)

## Summary

Patch release implementing skill modularization.

**Key Features**:
- **Modular Skills**: Skills can be split into modules
- **Bundled Files**: Reference materials can be bundled
- **Progressive Disclosure**: Token optimization

**Impact**:
- Reduced token usage
- Better skill organization
- Improved performance

## Changed

### Skill Structure

- **refactor**: Implement skill modularization (#220) (custom implementation)
  - Skills can have `modules/` directory
  - `reference.md` for external docs
  - `examples/` for code examples
  - Progressive disclosure metadata

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

# v1.7.1 - Skill System Improvements (2026-01-13)

## Summary

Patch release improving skill system.

**Key Improvements**:
- **Skill Loading**: Faster skill loading
- **Skill Discovery**: Better skill matching
- **Skill Metadata**: Enhanced metadata

**Impact**:
- Faster startup
- Better skill recommendations
- Improved documentation

## Changed

### Skill System

- **refactor**: Improve skill loading and discovery (#215) (custom implementation)
  - Faster skill loading
  - Better keyword matching
  - Enhanced metadata parsing

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

# v1.7.0 - Skill System Launch (2026-01-12)

## Summary

Minor release launching the unified skill system.

**Key Features**:
- **Unified Skills**: All functionality migrated to skills
- **Skill Catalog**: Comprehensive skill library
- **User-Invocable Skills**: Slash commands for users

**Impact**:
- Modular architecture
- Extensible system
- User-friendly commands

## Added

### Skill System

- **feat**: Launch unified skill system (#210) (custom implementation)
  - 50+ skills available
  - Category-based organization
  - User-invocable commands
  - Progressive disclosure

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

# v1.6.5 - Bug Fixes (2026-01-11)

## Summary

Patch release fixing various bugs.

**Key Fixes**:
- Fixed init command issues
- Fixed update command issues
- Fixed template processing

**Impact**:
- Improved stability
- Better error handling

## Fixed

### Command Issues

- **fix(cli)**: Fix init and update commands (#205) (various commits)
  - Better error handling
  - Fixed template processing
  - Improved user feedback

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

# v1.6.4 - Template Improvements (2026-01-10)

## Summary

Patch release improving template system.

**Key Improvements**:
- **Template Variables**: Better variable substitution
- **Template Processing**: More robust processing
- **Error Messages**: Clearer error messages

**Impact**:
- Better template expansion
- Improved error reporting

## Changed

### Template System

- **refactor**: Improve template processing (#200) (various commits)
  - Better variable handling
  - Clearer error messages
  - More robust parsing

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

# v1.6.3 - Documentation Updates (2026-01-09)

## Summary

Patch release updating documentation.

**Key Updates**:
- **README**: Improved README with better examples
- **CLAUDE.md**: Updated execution directives
- **Contributing**: Updated contribution guide

**Impact**:
- Better user onboarding
- Clearer documentation

## Changed

### Documentation

- **docs**: Update documentation (#195) (various commits)
  - Improved README
  - Updated CLAUDE.md
  - Enhanced contributing guide

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

# v1.6.2 - Configuration Improvements (2026-01-08)

## Summary

Patch release improving configuration system.

**Key Improvements**:
- **Modular Config**: Section-based configuration
- **Validation**: Better config validation
- **Defaults**: Improved default values

**Impact**:
- Better configuration management
- Cleaner config structure

## Changed

### Configuration System

- **refactor**: Improve configuration system (#190) (various commits)
  - Modular configuration sections
  - Better validation
  - Improved defaults

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

# v1.6.1 - Performance Improvements (2026-01-07)

## Summary

Patch release improving performance.

**Key Improvements**:
- **Faster Startup**: Reduced initialization time
- **Caching**: Better caching strategy
- **Optimization**: Code optimization

**Impact**:
- Faster command execution
- Better resource usage

## Changed

### Performance

- **perf**: Improve performance (#185) (various commits)
  - Faster startup
  - Better caching
  - Code optimization

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

# v1.6.0 - DDD Workflow (2026-01-06)

## Summary

Minor release launching DDD (Domain-Driven Development) workflow.

**Key Features**:
- **DDD Cycle**: ANALYZE-PRESERVE-IMPROVE
- **Characterization Tests**: Behavior preservation
- **Refactoring Support**: Safe code transformation

**Impact**:
- Safer refactoring
- Better code quality
- Improved development workflow

## Added

### DDD Workflow

- **feat**: Add DDD workflow (#180) (custom implementation)
  - ANALYZE phase: Understand code
  - PRESERVE phase: Characterization tests
  - IMPROVE phase: Incremental changes
  - Skill: `moai-workflow-ddd`

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

# v1.5.0 - SPEC Workflow (2025-12-28)

## Summary

Minor release launching SPEC (Specification) workflow.

**Key Features**:
- **SPEC Documents**: Structured requirement documents
- **EARS Format**: Easy Approach to Requirements Syntax
- **Three Phases**: Plan, Run, Sync

**Impact**:
- Better requirement management
- Structured development workflow
- Improved documentation

## Added

### SPEC Workflow

- **feat**: Add SPEC workflow (#175) (custom implementation)
  - EARS format requirements
  - Three-phase workflow
  - Integration with Alfred
  - Skill: `moai-workflow-spec`

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

# v1.4.0 - Alfred SuperAgent (2025-12-20)

## Summary

Minor release launching Alfred SuperAgent orchestration system.

**Key Features**:
- **Alfred Orchestrator**: Strategic task delegation
- **Agent Catalog**: Comprehensive agent library
- **Specialist Agents**: 20+ domain experts

**Impact**:
- Autonomous development workflows
- Better task routing
- Improved code quality

## Added

### Alfred System

- **feat**: Add Alfred SuperAgent (#170) (custom implementation)
  - Strategic orchestration
  - Agent delegation
  - Task routing
  - Specialist agents

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

# v1.3.0 - Template System (2025-12-15)

## Summary

Minor release launching template system.

**Key Features**:
- **Project Templates**: Starter templates for various projects
- **Template Variables**: Dynamic variable substitution
- **Template Processing**: Robust template engine

**Impact**:
- Faster project setup
- Consistent project structure
- Better developer experience

## Added

### Template System

- **feat**: Add template system (#165) (custom implementation)
  - Project templates
  - Variable substitution
  - Template processing

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

# v1.2.0 - CLI Commands (2025-12-10)

## Summary

Minor release launching CLI command system.

**Key Features**:
- **Init Command**: Project initialization
- **Update Command**: Template updates
- **Doctor Command**: System health check

**Impact**:
- Easy project setup
- Seamless updates
- Better troubleshooting

## Added

### CLI Commands

- **feat**: Add CLI commands (#160) (custom implementation)
  - init: Project initialization
  - update: Template updates
  - doctor: Health check

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

# v1.1.0 - Configuration System (2025-12-05)

## Summary

Minor release launching configuration system.

**Key Features**:
- **YAML Configuration**: Structured config files
- **Section-Based**: Modular configuration sections
- **Validation**: Config validation

**Impact**:
- Better configuration management
- Improved defaults
- Cleaner structure

## Added

### Configuration System

- **feat**: Add configuration system (#155) (custom implementation)
  - YAML configuration
  - Modular sections
  - Config validation

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

# v1.0.0 - Initial Release (2025-12-01)

## Summary

First stable release of MoAI-ADK.

**Key Features**:
- **Claude Code Integration**: Seamless integration with Claude Code
- **Project Templates**: Starter templates for various projects
- **CLI Commands**: Command-line interface for common tasks
- **Configuration System**: Flexible configuration management

## Installation

```bash
# Install MoAI-ADK
uv tool install moai-adk

# Initialize new project
moai init my-project

# Check version
moai --version
```

---

## Version History Summary

| Version | Date | Type | Description |
|---------|------|------|-------------|
| v1.12.1 | 2026-01-29 | patch | Init Validation & Hook Dependency Fixes |
| v1.12.0 | 2026-01-29 | minor | Google Stitch Build Loop & Documentation Updates |
| v1.11.2 | 2026-01-29 | patch | CLAUDE.md Reference Fix |
| v1.11.1 | 2026-01-29 | patch | CLAUDE.md English Only |
| v1.11.0 | 2026-01-29 | patch | Template Variable Substitution Fix |
| v1.10.5 | 2026-01-29 | patch | Template Variable Substitution Fix |
| v1.9.0 | 2026-01-26 | minor | Memory MCP, SVG Skill, Rules Migration |
| v1.8.13 | 2026-01-26 | patch | Statusline Context Window Fix |
| v1.8.12 | 2026-01-26 | patch | Hook Format Update & Login Command |
| v1.8.11 | 2026-01-25 | patch | Cross-Platform Shell Wrappers |
| v1.8.10 | 2026-01-25 | patch | Path Separator Consolidation |
| v1.8.9 | 2026-01-25 | patch | Path Separator Fix |
| v1.8.8 | 2026-01-25 | patch | Claude Code Hook Fix |
| v1.8.7 | 2026-01-24 | patch | Double Slash Fix |
| v1.8.6 | 2026-01-24 | patch | Worktree Command Addition |
| v1.8.5 | 2026-01-23 | patch | Rank Command Addition |
| v1.8.4 | 2026-01-22 | patch | LSP Quality Gates |
| v1.8.3 | 2026-01-21 | patch | Context Cache Manager |
| v1.8.2 | 2026-01-20 | patch | LSP Integration |
| v1.8.1 | 2026-01-19 | patch | JIT Context Loader |
| v1.8.0 | 2026-01-18 | minor | MoAI Rename & Unified Command |
| v1.7.5 | 2026-01-17 | patch | Template Bug Fixes |
| v1.7.4 | 2026-01-16 | patch | Additional Agent Additions |
| v1.7.3 | 2026-01-15 | patch | Agent System Expansion |
| v1.7.2 | 2026-01-14 | patch | Skill Modularization |
| v1.7.1 | 2026-01-13 | patch | Skill System Improvements |
| v1.7.0 | 2026-01-12 | minor | Skill System Launch |
| v1.6.5 | 2026-01-11 | patch | Bug Fixes |
| v1.6.4 | 2026-01-10 | patch | Template Improvements |
| v1.6.3 | 2026-01-09 | patch | Documentation Updates |
| v1.6.2 | 2026-01-08 | patch | Configuration Improvements |
| v1.6.1 | 2026-01-07 | patch | Performance Improvements |
| v1.6.0 | 2026-01-06 | minor | DDD Workflow |
| v1.5.0 | 2025-12-28 | minor | SPEC Workflow |
| v1.4.0 | 2025-12-20 | minor | Alfred SuperAgent |
| v1.3.0 | 2025-12-15 | minor | Template System |
| v1.2.0 | 2025-12-10 | minor | CLI Commands |
| v1.1.0 | 2025-12-05 | minor | Configuration System |
| v1.0.0 | 2025-12-01 | major | Initial Release |
