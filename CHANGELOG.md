# v1.14.0 - Major Refactor: Pencil MCP Design Tool Migration (2026-02-01)

## Summary

Major refactoring release migrating the entire design tooling stack from Google Stitch MCP and Figma to Pencil MCP (.pen files). This consolidation simplifies the design workflow and empowers `expert-frontend` to handle UI/UX design directly.

**Key Changes**:
- **Design Tool Migration**: Complete replacement of Stitch/Figma with Pencil MCP for design-to-code workflows
- **Simplified Architecture**: Removed `expert-stitch` agent, consolidated into `expert-frontend` (Expert Agents: 9 → 8)
- **Pencil MCP Integration**: Added 14 Pencil MCP tools with comprehensive documentation
- **Chrome Extension**: Includes v1.13.0 Chrome Extension platform skill (previously untagged)

## Breaking Changes

**⚠️ Stitch MCP Removal**: The `moai-platform-stitch` skill and `expert-stitch` agent have been removed. Projects using Google Stitch should migrate to Pencil MCP.

**Migration Guide**:
- Replace `.stitch` files with `.pen` files
- Update design references from `expert-stitch` to `expert-frontend`
- Remove Stitch MCP server from `.mcp.json`
- See [Pencil MCP documentation](https://docs.pencil.dev/) for migration details

## Added

### Pencil MCP Design Integration

- **refactor**: Replace Stitch/Figma with Pencil MCP for UI/UX design (acfdc794)
  - **Pencil MCP Tools** (14 tools added to `expert-frontend`):
    - `batch_design`: Execute multiple insert/copy/update/replace/move/delete operations
    - `batch_get`: Retrieve nodes by patterns or IDs in batches
    - `get_editor_state`: Determine active .pen file and user selection
    - `get_guidelines`: Design guidelines for code, tables, Tailwind, landing pages
    - `get_screenshot`: Capture visual screenshots of .pen file nodes
    - `get_style_guide`: Retrieve style guides by tags or name for design inspiration
    - `get_style_guide_tags`: List available style guide tags
    - `get_variables`: Extract design tokens and themes from .pen files
    - `set_variables`: Add or update variables in .pen files
    - `open_document`: Create new or open existing .pen files
    - `snapshot_layout`: Check computed layout rectangles for node insertion
    - `find_empty_space_on_canvas`: Find empty space for new elements
    - `search_all_unique_properties`: Recursively search node properties
    - `replace_all_matching_properties`: Bulk property updates
  - **Comprehensive Workflow Guide**: Added "UI/UX Design with Pencil MCP" section to `expert-frontend.md`
    - 5-step design workflow: Initialize → Style Foundation → Design Creation → Iteration → Code Export
    - Variables and Design Tokens guide
    - Available UI Kits: Shadcn UI, Halo, Lunaris, Nitro
    - Best practices for prompting, file management, design-to-code
  - **Design Philosophy**: `.pen` files are JSON-based, Git-friendly, human-readable design-as-code format
  - **Integration**: Pencil MCP auto-starts with IDE extension/desktop app (no .mcp.json entry needed)

### Chrome Extension Platform (from v1.13.0)

- **feat**: Add Chrome Extension platform skill and agent (be0ec50e)
  - **Skill**: `moai-platform-chrome-extension` with 8 modular guides
  - **Agent**: `expert-chrome-extension` for specialized development
  - **Coverage**: Manifest V3, service workers, content scripts, message passing, chrome.* APIs, security, publishing

## Changed

### Design Tool Architecture

- **Removed**: `moai-platform-stitch` skill directory (template + local)
- **Removed**: `expert-stitch` agent from agent catalog
- **Updated**: Expert Agents count from 9 to 8
- **Updated**: `expert-frontend` now handles UI/UX design directly
- **Updated**: All Figma/Stitch references replaced with Pencil MCP across:
  - Agent definitions: `expert-frontend`, `manager-docs`, `manager-spec`, `manager-strategy`
  - Skills: `moai-domain-uiux`, `moai-foundation-core`, `moai-foundation-claude`
  - Configurations: `multilingual-triggers.yaml` (4 languages), CLAUDE.md, README.zh.md
- **Removed**: Stitch MCP server from `.mcp.json` (all platforms: macOS, Windows)

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

# v1.14.0 - 주요 리팩토링: Pencil MCP 디자인 도구 마이그레이션 (2026-02-01)

## 요약

Google Stitch MCP 및 Figma에서 Pencil MCP (.pen 파일)로 전체 디자인 도구 스택을 마이그레이션하는 주요 리팩토링 릴리스입니다. 이 통합으로 디자인 워크플로우가 단순화되고 `expert-frontend`가 UI/UX 디자인을 직접 처리할 수 있게 되었습니다.

**주요 변경사항**:
- **디자인 도구 마이그레이션**: Stitch/Figma를 Pencil MCP로 완전 교체 (design-to-code 워크플로우)
- **단순화된 아키텍처**: `expert-stitch` 에이전트 제거, `expert-frontend`로 통합 (전문 에이전트: 9 → 8)
- **Pencil MCP 통합**: 종합 문서와 함께 14개 Pencil MCP 도구 추가
- **Chrome Extension**: v1.13.0 Chrome Extension 플랫폼 스킬 포함 (이전에 태그되지 않음)

## Breaking Changes

**⚠️ Stitch MCP 제거**: `moai-platform-stitch` 스킬과 `expert-stitch` 에이전트가 제거되었습니다. Google Stitch를 사용하는 프로젝트는 Pencil MCP로 마이그레이션해야 합니다.

**마이그레이션 가이드**:
- `.stitch` 파일을 `.pen` 파일로 교체
- 디자인 참조를 `expert-stitch`에서 `expert-frontend`로 업데이트
- `.mcp.json`에서 Stitch MCP 서버 제거
- 마이그레이션 세부사항은 [Pencil MCP 문서](https://docs.pencil.dev/) 참조

## 추가됨

### Pencil MCP 디자인 통합

- **refactor**: UI/UX 디자인을 위해 Stitch/Figma를 Pencil MCP로 교체 (acfdc794)
  - **Pencil MCP 도구** (`expert-frontend`에 14개 도구 추가):
    - `batch_design`: 여러 insert/copy/update/replace/move/delete 작업 실행
    - `batch_get`: 패턴 또는 ID로 노드를 일괄 검색
    - `get_editor_state`: 활성 .pen 파일 및 사용자 선택 확인
    - `get_guidelines`: 코드, 테이블, Tailwind, 랜딩 페이지를 위한 디자인 가이드라인
    - `get_screenshot`: .pen 파일 노드의 시각적 스크린샷 캡처
    - `get_style_guide`: 디자인 영감을 위한 태그 또는 이름으로 스타일 가이드 검색
    - `get_style_guide_tags`: 사용 가능한 스타일 가이드 태그 목록
    - `get_variables`: .pen 파일에서 디자인 토큰 및 테마 추출
    - `set_variables`: .pen 파일에 변수 추가 또는 업데이트
    - `open_document`: 새 .pen 파일 생성 또는 기존 파일 열기
    - `snapshot_layout`: 노드 삽입을 위한 계산된 레이아웃 사각형 확인
    - `find_empty_space_on_canvas`: 새 요소를 위한 빈 공간 찾기
    - `search_all_unique_properties`: 노드 속성 재귀 검색
    - `replace_all_matching_properties`: 속성 일괄 업데이트
  - **종합 워크플로우 가이드**: `expert-frontend.md`에 "UI/UX Design with Pencil MCP" 섹션 추가
    - 5단계 디자인 워크플로우: 초기화 → 스타일 기반 → 디자인 생성 → 반복 → 코드 내보내기
    - 변수 및 디자인 토큰 가이드
    - 사용 가능한 UI 킷: Shadcn UI, Halo, Lunaris, Nitro
    - 프롬프팅, 파일 관리, design-to-code를 위한 모범 사례
  - **디자인 철학**: `.pen` 파일은 JSON 기반, Git 친화적, 사람이 읽을 수 있는 design-as-code 형식
  - **통합**: Pencil MCP는 IDE 확장/데스크톱 앱과 함께 자동 시작 (.mcp.json 항목 불필요)

### Chrome Extension 플랫폼 (v1.13.0에서)

- **feat**: Chrome Extension 플랫폼 스킬 및 에이전트 추가 (be0ec50e)
  - **스킬**: `moai-platform-chrome-extension` - 8개의 모듈식 가이드 포함
  - **에이전트**: `expert-chrome-extension` - 전문 개발용
  - **범위**: Manifest V3, service workers, content scripts, 메시지 전달, chrome.* APIs, 보안, 배포

## 변경됨

### 디자인 도구 아키텍처

- **제거됨**: `moai-platform-stitch` 스킬 디렉토리 (template + local)
- **제거됨**: 에이전트 카탈로그에서 `expert-stitch` 에이전트
- **업데이트됨**: 전문 에이전트 수 9 → 8
- **업데이트됨**: `expert-frontend`가 이제 UI/UX 디자인을 직접 처리
- **업데이트됨**: 모든 Figma/Stitch 참조를 Pencil MCP로 교체:
  - 에이전트 정의: `expert-frontend`, `manager-docs`, `manager-spec`, `manager-strategy`
  - 스킬: `moai-domain-uiux`, `moai-foundation-core`, `moai-foundation-claude`
  - 구성: `multilingual-triggers.yaml` (4개 언어), CLAUDE.md, README.zh.md
- **제거됨**: `.mcp.json`에서 Stitch MCP 서버 (모든 플랫폼: macOS, Windows)

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.13.0 - Feature: Chrome Extension Platform Support (2026-02-01)

## Summary

Feature release adding comprehensive Chrome Extension Manifest V3 development support.

**Key Features**:
- **Chrome Extension Platform**: Full Manifest V3 development support with service workers, content scripts, and chrome.* APIs
- **Expert Agent**: New `expert-chrome-extension` agent for specialized extension development
- **Comprehensive Documentation**: 8 modular guides covering APIs, security, publishing, and best practices

## Breaking Changes

None. This is a backward-compatible feature addition.

## Added

### Chrome Extension Platform Skill

- **feat**: Add Chrome Extension platform skill and agent (be0ec50e)
  - **Skill**: `moai-platform-chrome-extension` with 8 modular guides
    - Manifest V3 reference and configuration
    - Service Worker patterns and lifecycle management
    - Content Scripts injection and communication
    - Message passing patterns (one-time, long-lived, native)
    - chrome.* APIs quick reference (storage, tabs, runtime, permissions)
    - Security and CSP configuration
    - UI components (popup, options, side panel)
    - Publishing and Chrome Web Store submission
  - **Agent**: `expert-chrome-extension` for specialized development
  - **Coverage**: Service workers, content scripts, message passing, chrome.* APIs, security, publishing
  - **Integration**: Seamless integration with MoAI-ADK workflow
  - **Dependencies**: Updated `uv.lock` for package consistency

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

# v1.13.0 - 기능 추가: Chrome Extension 플랫폼 지원 (2026-02-01)

## 요약

Chrome Extension Manifest V3 개발을 위한 포괄적인 지원을 추가하는 기능 릴리스입니다.

**주요 기능**:
- **Chrome Extension 플랫폼**: Service Worker, Content Scripts, chrome.* APIs를 포함한 전체 Manifest V3 개발 지원
- **전문 에이전트**: 확장 프로그램 전문 개발을 위한 `expert-chrome-extension` 에이전트 추가
- **종합 문서**: API, 보안, 배포, 모범 사례를 다루는 8개의 모듈식 가이드

## Breaking Changes

없음. 하위 호환 가능한 기능 추가입니다.

## 추가됨

### Chrome Extension 플랫폼 스킬

- **feat**: Chrome Extension 플랫폼 스킬 및 에이전트 추가 (be0ec50e)
  - **스킬**: `moai-platform-chrome-extension` - 8개의 모듈식 가이드 포함
    - Manifest V3 참조 및 구성
    - Service Worker 패턴 및 생명주기 관리
    - Content Scripts 주입 및 통신
    - 메시지 전달 패턴 (일회성, 장기 연결, 네이티브)
    - chrome.* APIs 빠른 참조 (storage, tabs, runtime, permissions)
    - 보안 및 CSP 구성
    - UI 컴포넌트 (popup, options, side panel)
    - 배포 및 Chrome Web Store 제출
  - **에이전트**: `expert-chrome-extension` - 전문 개발용
  - **범위**: Service workers, content scripts, 메시지 전달, chrome.* APIs, 보안, 배포
  - **통합**: MoAI-ADK 워크플로우와 원활한 통합
  - **의존성**: 패키지 일관성을 위한 `uv.lock` 업데이트

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.12.15 - Bug Fix: StatusLine Template Substitution (2026-02-01)

## Summary

Bug fix release resolving statusLine command overwrite issue during `moai-adk update`.

**Key Fix**:
- **StatusLine Update**: Fixed `_update_statusline_command()` incorrectly replacing template-substituted statusLine commands
- **Impact**: StatusLine now consistently includes PATH augmentation, ensuring reliable `moai-adk statusline` execution
- **Scope**: Template variable detection and command generation logic

## Breaking Changes

None. This is a backward-compatible bug fix.

## Fixed

### StatusLine Template Substitution Overwrite

- **fix**: Resolve statusLine template substitution overwrite issue (2afba5d7)
  - **Issue**: `_update_statusline_command()` replaced valid template-substituted commands
  - **Root cause**: Detection logic only accepted `${SHELL}` literal, not actual shell paths like `/bin/zsh -l -c`
  - **Fix**: Updated detection to accept both `${SHELL}` variable and login shell indicator (`-l -c`)
  - **Enhancement**: Use `build_hook_context()` for command generation (same as hooks)
  - **Consistency**: StatusLine now uses identical format to hooks (includes `_PATH_AUGMENT`)
  - **Files modified**:
    - `update.py`: Updated `_update_statusline_command()` detection and generation logic
  - **Impact**: StatusLine commands now match hook format with proper PATH augmentation

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

# v1.12.15 - 버그 수정: StatusLine 템플릿 치환 (2026-02-01)

## 요약

`moai-adk update` 중 statusLine 명령 덮어쓰기 문제를 해결하는 버그 수정 릴리스입니다.

**주요 수정**:
- **StatusLine 업데이트**: `_update_statusline_command()`가 템플릿 치환된 statusLine 명령을 잘못 교체하는 문제 수정
- **영향**: StatusLine이 이제 일관되게 PATH 보강을 포함하여 `moai-adk statusline` 실행 안정성 확보
- **범위**: 템플릿 변수 감지 및 명령 생성 로직

## Breaking Changes

없음. 하위 호환 가능한 버그 수정입니다.

## 수정됨

### StatusLine 템플릿 치환 덮어쓰기

- **fix**: StatusLine 템플릿 치환 덮어쓰기 이슈 해결 (2afba5d7)
  - **문제**: `_update_statusline_command()`가 유효한 템플릿 치환 명령을 교체함
  - **근본 원인**: 감지 로직이 `${SHELL}` 리터럴만 허용하고 `/bin/zsh -l -c` 같은 실제 쉘 경로는 거부
  - **해결**: `${SHELL}` 변수와 로그인 쉘 표시자(`-l -c`) 모두 허용하도록 감지 로직 업데이트
  - **개선**: 명령 생성에 `build_hook_context()` 사용 (hooks와 동일)
  - **일관성**: StatusLine이 이제 hooks와 동일한 형식 사용 (`_PATH_AUGMENT` 포함)
  - **수정된 파일**:
    - `update.py`: `_update_statusline_command()` 감지 및 생성 로직 업데이트
  - **영향**: StatusLine 명령이 이제 적절한 PATH 보강을 포함한 hook 형식과 일치

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.12.11 - Critical Bug Fix: Template Sync JSON Parsing Error (2026-02-01)

## Summary

Critical bug fix release resolving JSON parsing failure during `moai-adk update` on macOS, Linux, and WSL systems.

**Key Fix**:
- **Template Sync**: Fixed JSON parsing error when HOOK_SHELL_PREFIX contains shell syntax with quotes
- **Impact**: All users on macOS/Linux/WSL can now run `moai update` successfully
- **Scope**: Template variable substitution system

## Breaking Changes

None. This is a backward-compatible bug fix.

## Fixed

### Template Sync JSON Parsing Error

- **fix**: Resolve JSON parsing error in template sync (058a9e46)
  - **Issue**: `moai-adk update` failed with `Expecting ',' delimiter: line 9 column 55` error
  - **Root cause**: Unescaped quotes in `_PATH_AUGMENT` (`export PATH="$HOME/..."`) broke JSON when substituted into settings.json
  - **Fix**: Removed unnecessary quotes from `_PATH_AUGMENT` (PATH values have no spaces)
  - **Enhancement**: Added `json_safe` parameter to `_substitute_variables()` for JSON-aware escaping using `json.dumps()`
  - **WSL support**: Handles `_WSL_PATH_NORMALIZE` with quotes correctly
  - **Files modified**:
    - `hook_context.py`: Simplified `_PATH_AUGMENT` from `'export PATH="$HOME/..."'` to `"export PATH=$HOME/..."`
    - `processor.py`: Added JSON escaping with `json.dumps()[1:-1]` when `json_safe=True`
    - `update.py`: Applied JSON-safe escaping to `.json` files
  - **Testing**: 80/80 unit tests passing, verified with actual `HOOK_SHELL_PREFIX` values
  - **Impact**: Template sync now works correctly on all platforms (macOS/Linux/WSL)

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

# v1.12.11 - 중요 버그 수정: Template Sync JSON 파싱 오류 (2026-02-01)

## 요약

macOS, Linux, WSL 시스템에서 `moai-adk update` 실행 중 발생하는 JSON 파싱 실패를 해결하는 중요 버그 수정 릴리스입니다.

**주요 수정**:
- **Template Sync**: HOOK_SHELL_PREFIX에 따옴표가 포함된 셸 구문이 있을 때 발생하는 JSON 파싱 오류 수정
- **영향**: macOS/Linux/WSL의 모든 사용자가 이제 `moai update`를 성공적으로 실행할 수 있음
- **범위**: 템플릿 변수 치환 시스템

## Breaking Changes

없음. 하위 호환 가능한 버그 수정입니다.

## 수정됨

### Template Sync JSON 파싱 오류

- **fix**: Template sync의 JSON 파싱 오류 해결 (058a9e46)
  - **문제**: `moai-adk update` 실행 시 `Expecting ',' delimiter: line 9 column 55` 오류 발생
  - **근본 원인**: `_PATH_AUGMENT`의 이스케이프되지 않은 따옴표 (`export PATH="$HOME/..."`)가 settings.json에 치환될 때 JSON을 손상시킴
  - **해결**: `_PATH_AUGMENT`에서 불필요한 따옴표 제거 (PATH 값에는 공백이 없음)
  - **개선**: `_substitute_variables()`에 JSON 인식 이스케이핑을 위한 `json_safe` 파라미터 추가 (`json.dumps()` 사용)
  - **WSL 지원**: 따옴표가 포함된 `_WSL_PATH_NORMALIZE`를 올바르게 처리
  - **수정된 파일**:
    - `hook_context.py`: `_PATH_AUGMENT`를 `'export PATH="$HOME/..."'`에서 `"export PATH=$HOME/..."`로 단순화
    - `processor.py`: `json_safe=True`일 때 `json.dumps()[1:-1]`를 사용한 JSON 이스케이핑 추가
    - `update.py`: `.json` 파일에 JSON-safe 이스케이핑 적용
  - **테스트**: 80/80 unit 테스트 통과, 실제 `HOOK_SHELL_PREFIX` 값으로 검증 완료
  - **영향**: 모든 플랫폼 (macOS/Linux/WSL)에서 template sync가 올바르게 작동

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.12.10 - MoAI-GO Experimental Implementation & UTF-8 Encoding Fixes (2026-01-31)

## Summary

Experimental release introducing MoAI-GO, a high-performance Go-based CLI implementation, alongside critical UTF-8 encoding fixes for Windows compatibility.

**Key Features**:
- **MoAI-GO System**: Experimental Go implementation (Phases 1-4) for performance evaluation
- **UTF-8 Encoding**: Comprehensive UTF-8 encoding fixes for Windows (#314, #316)
- **Cross-Platform**: Improved PATH handling and hook execution

**Note**: MoAI-GO is experimental. Production use still relies on Python implementation.

## Breaking Changes

None. All changes are backward compatible.

## Added

### MoAI-GO Experimental System

- **feat(go)**: Implement MoAI-GO Phase 1 - Foundation and configuration system (b36ddd66)
  - Go-based configuration management
  - Cross-platform compatibility layer
  - Performance-optimized file operations

- **feat(go)**: Implement MoAI-GO Phase 2 - Hook system with Python compatibility (f3e75202)
  - Go-based hook execution engine
  - Backward compatible with Python hooks
  - Improved hook performance

- **feat(go)**: Implement MoAI-GO Phase 3 - CLI commands with direct binary paths (e47534d0)
  - Go-based CLI commands: `rank`, `switch`, `worktree`
  - Direct binary execution (no Python interpreter needed)
  - Cross-platform PATH handling improvements

- **feat(go)**: Implement MoAI-GO Phase 4 - Distribution and migration system (d61f381d)
  - Binary distribution system
  - Migration tools from Python to Go
  - Version compatibility checks

- **feat**: Add MoAI-GO CLI commands (rank, switch, worktree) and improve cross-platform PATH handling (05d7b7be)
  - Integrated CLI commands with improved performance
  - Enhanced PATH resolution for Windows/macOS/Linux
  - Automatic binary fallback to Python implementation

## Fixed

### UTF-8 Encoding Issues

- **fix**: Add UTF-8 encoding to all hook scripts and file operations (33c0a483)
  - Fixed UTF-8 encoding in all `.claude/hooks/moai/*.py` files
  - Added explicit `encoding='utf-8'` to all file operations
  - Ensures correct handling of non-ASCII characters

- **fix**: Resolve cross-platform hook/encoding/credential issues (#314, #316) (bf280eeb)
  - Issue #314: Hook execution failures on Windows due to encoding
  - Issue #316: Credential file corruption with non-ASCII characters
  - Fix: Consistent UTF-8 encoding across all platforms
  - Impact: Windows users can now use hooks without encoding errors

- **fix(cli)**: Resolve Windows UTF-8 encoding error on moai init (894abd3e)
  - Issue: `moai init` failed on Windows with UnicodeEncodeError
  - Root cause: Console output encoding mismatch
  - Fix: Force UTF-8 encoding for all console output
  - Impact: `moai init` now works correctly on Windows

## Changed

### Go Module Restructure

- **refactor**: Restructure Go module path to github.com/modu-ai/moai-adk/go (97be74a7)
  - Aligned Go module path with repository structure
  - Improved import clarity for Go developers
  - Prepared for future Go module distribution

- **revert**: Undo Go module restructure (testing needed) (324097d1)
  - Reverted Go module path change for further testing
  - Ensures stability before finalizing module structure

- **refactor**: Rename moai-adk to moai (2b9bf0e1)
  - Internal renaming for consistency
  - Updated import statements
  - Aligned with unified branding

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

# v1.12.10 - MoAI-GO 실험적 구현 및 UTF-8 인코딩 수정 (2026-01-31)

## 요약

MoAI-GO 고성능 Go 기반 CLI 구현과 Windows 호환성을 위한 UTF-8 인코딩 수정을 포함하는 실험적 릴리스입니다.

**주요 기능**:
- **MoAI-GO 시스템**: 성능 평가를 위한 실험적 Go 구현 (Phase 1-4)
- **UTF-8 인코딩**: Windows용 포괄적 UTF-8 인코딩 수정 (#314, #316)
- **크로스플랫폼**: 개선된 PATH 처리 및 hook 실행

**참고**: MoAI-GO는 실험적입니다. 프로덕션 환경에서는 Python 구현을 사용하세요.

## Breaking Changes

없음. 모든 변경사항은 하위 호환됩니다.

## 추가됨

### MoAI-GO 실험적 시스템

- **feat(go)**: MoAI-GO Phase 1 - Foundation 및 구성 시스템 구현 (b36ddd66)
  - Go 기반 구성 관리
  - 크로스플랫폼 호환성 레이어
  - 성능 최적화된 파일 작업

- **feat(go)**: MoAI-GO Phase 2 - Python 호환성을 갖춘 Hook 시스템 (f3e75202)
  - Go 기반 hook 실행 엔진
  - Python hook과 하위 호환성
  - 향상된 hook 성능

- **feat(go)**: MoAI-GO Phase 3 - 직접 바이너리 경로를 사용하는 CLI 명령 (e47534d0)
  - Go 기반 CLI 명령: `rank`, `switch`, `worktree`
  - 직접 바이너리 실행 (Python 인터프리터 불필요)
  - 크로스플랫폼 PATH 처리 개선

- **feat(go)**: MoAI-GO Phase 4 - 배포 및 마이그레이션 시스템 (d61f381d)
  - 바이너리 배포 시스템
  - Python에서 Go로 마이그레이션 도구
  - 버전 호환성 검사

- **feat**: MoAI-GO CLI 명령 (rank, switch, worktree) 추가 및 크로스플랫폼 PATH 처리 개선 (05d7b7be)
  - 성능이 향상된 통합 CLI 명령
  - Windows/macOS/Linux를 위한 향상된 PATH 해결
  - Python 구현으로의 자동 바이너리 폴백

## 수정됨

### UTF-8 인코딩 이슈

- **fix**: 모든 hook 스크립트 및 파일 작업에 UTF-8 인코딩 추가 (33c0a483)
  - 모든 `.claude/hooks/moai/*.py` 파일의 UTF-8 인코딩 수정
  - 모든 파일 작업에 명시적 `encoding='utf-8'` 추가
  - 비ASCII 문자의 올바른 처리 보장

- **fix**: 크로스플랫폼 hook/encoding/credential 이슈 해결 (#314, #316) (bf280eeb)
  - 이슈 #314: 인코딩으로 인한 Windows hook 실행 실패
  - 이슈 #316: 비ASCII 문자로 인한 credential 파일 손상
  - 해결: 모든 플랫폼에서 일관된 UTF-8 인코딩
  - 영향: Windows 사용자가 인코딩 오류 없이 hook을 사용할 수 있음

- **fix(cli)**: moai init의 Windows UTF-8 인코딩 오류 해결 (894abd3e)
  - 문제: Windows에서 `moai init` 실행 중 UnicodeEncodeError 발생
  - 근본 원인: 콘솔 출력 인코딩 불일치
  - 해결: 모든 콘솔 출력에 UTF-8 인코딩 강제
  - 영향: Windows에서 `moai init`가 올바르게 작동

## 변경됨

### Go 모듈 구조 재구성

- **refactor**: Go 모듈 경로를 github.com/modu-ai/moai-adk/go로 재구성 (97be74a7)
  - Go 모듈 경로를 리포지토리 구조와 정렬
  - Go 개발자를 위한 import 명확성 개선
  - 향후 Go 모듈 배포 준비

- **revert**: Go 모듈 구조 재구성 취소 (테스트 필요) (324097d1)
  - 추가 테스트를 위해 Go 모듈 경로 변경 되돌림
  - 모듈 구조 확정 전 안정성 보장

- **refactor**: moai-adk를 moai로 이름 변경 (2b9bf0e1)
  - 일관성을 위한 내부 이름 변경
  - import 문 업데이트
  - 통합 브랜딩과 정렬

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

