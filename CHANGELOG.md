# v1.8.2 - Cross-Platform Compatibility & Settings.json Smart Merge Enhancement (2026-01-24)

## Summary

This patch release improves cross-platform compatibility for Windows, macOS, and Linux environments, fixes settings.json environment variable merge priority, and enhances Windows environment validation.

**Key Changes**:
- Fixed environment variable merge priority in settings.json smart merge
- Ensured template variables are always substituted in settings.json
- Added shell environment validation for Windows
- Improved cross-platform documentation and error messages
- Clarified Windows PowerShell requirement and WSL non-support

## Fixed

### Cross-Platform Compatibility

- **feat**: Fix env merge priority in settings.json smart merge (d02f8d88)
  - Fixed environment variable merge order: User ENV > Template defaults
  - Ensures user-provided environment variables always take precedence
  - Prevents template defaults from overriding user customizations
  - File: `src/moai_adk/core/template/merger.py`

- **fix**: Ensure template variables are always substituted in settings.json (b3cc2525)
  - Guarantees `{{PROJECT_DIR}}`, `{{MOAI_VERSION}}` substitution during init
  - Prevents raw placeholders from appearing in user settings
  - File: `src/moai_adk/core/template/merger.py`

- **feat**: Improve cross-platform compatibility (Windows/macOS/Linux) (0dc44d41)
  - Enhanced platform detection and path handling
  - Improved error messages for platform-specific issues
  - Files: `src/moai_adk/core/project/checker.py`, `src/moai_adk/core/project/phase_executor.py`, `src/moai_adk/core/git/event_detector.py`

- **feat**: Add shell environment validation for Windows (a1613654)
  - New module: `src/moai_adk/utils/shell_validator.py`
  - Validates Windows shell configuration (PowerShell requirement)
  - Detects unsupported environments (WSL, Cygwin, MinGW)
  - Provides clear error messages with resolution steps
  - File: `src/moai_adk/utils/shell_validator.py`

### Documentation

- **docs(platform)**: Improve cross-platform documentation and error messages (736440d7)
  - Enhanced error messages with platform-specific guidance
  - Improved documentation clarity for Windows users

- **docs**: Clarify WSL is also not supported on Windows (2b709740)
  - Updated all README files with WSL non-support notice
  - Files: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`

- **docs**: Add Windows PowerShell requirement notice to all README files (de2132b0)
  - Clear documentation of Windows PowerShell requirement
  - Installation instructions for PowerShell users
  - Files: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`

### Hooks

- **docs(hooks)**: Add comprehensive lib/ documentation (NEW)
  - New file: `src/moai_adk/templates/.claude/hooks/moai/lib/README.md`
  - Documents 13 utility modules with usage examples
  - Covers configuration, data structures, execution, I/O, and quality utilities
  - Provides import patterns and best practices

## Changed

- **chore**: Update dependency lock file (uv.lock 1.8.0 → 1.8.1) (b78a5324)
  - Updated dependency versions for compatibility

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 221 files unchanged
- Mypy: Success (no issues found in 173 source files)

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

# v1.8.2 - 크로스 플랫폼 호환성 및 Settings.json 스마트 병합 개선 (2026-01-24)

## 요약

이 패치 릴리스는 Windows, macOS, Linux 환경의 크로스 플랫폼 호환성을 개선하고, settings.json 환경 변수 병합 우선순위를 수정하며, Windows 환경 검증을 강화합니다.

**주요 변경사항**:
- settings.json 스마트 병합의 환경 변수 병합 우선순위 수정
- settings.json에서 템플릿 변수가 항상 치환되도록 보장
- Windows용 쉘 환경 검증 추가
- 크로스 플랫폼 문서 및 오류 메시지 개선
- Windows PowerShell 요구사항 및 WSL 미지원 명확화

## 수정됨

### 크로스 플랫폼 호환성

- **feat**: settings.json 스마트 병합에서 환경 변수 병합 우선순위 수정 (d02f8d88)
  - 환경 변수 병합 순서 수정: 사용자 ENV > 템플릿 기본값
  - 사용자가 제공한 환경 변수가 항상 우선하도록 보장
  - 템플릿 기본값이 사용자 커스터마이징을 덮어쓰지 않도록 방지
  - 파일: `src/moai_adk/core/template/merger.py`

- **fix**: settings.json에서 템플릿 변수가 항상 치환되도록 보장 (b3cc2525)
  - init 중 `{{PROJECT_DIR}}`, `{{MOAI_VERSION}}` 치환 보장
  - 사용자 설정에 원시 플레이스홀더가 나타나지 않도록 방지
  - 파일: `src/moai_adk/core/template/merger.py`

- **feat**: 크로스 플랫폼 호환성 개선 (Windows/macOS/Linux) (0dc44d41)
  - 플랫폼 감지 및 경로 처리 강화
  - 플랫폼별 문제에 대한 오류 메시지 개선
  - 파일: `src/moai_adk/core/project/checker.py`, `src/moai_adk/core/project/phase_executor.py`, `src/moai_adk/core/git/event_detector.py`

- **feat**: Windows용 쉘 환경 검증 추가 (a1613654)
  - 새 모듈: `src/moai_adk/utils/shell_validator.py`
  - Windows 쉘 구성 검증 (PowerShell 요구사항)
  - 미지원 환경 감지 (WSL, Cygwin, MinGW)
  - 해결 단계를 포함한 명확한 오류 메시지 제공
  - 파일: `src/moai_adk/utils/shell_validator.py`

### 문서화

- **docs(platform)**: 크로스 플랫폼 문서 및 오류 메시지 개선 (736440d7)
  - 플랫폼별 가이드를 포함한 오류 메시지 강화
  - Windows 사용자를 위한 문서 명확성 개선

- **docs**: Windows에서 WSL도 미지원임을 명확화 (2b709740)
  - WSL 미지원 공지로 모든 README 파일 업데이트
  - 파일: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`

- **docs**: 모든 README 파일에 Windows PowerShell 요구사항 공지 추가 (de2132b0)
  - Windows PowerShell 요구사항에 대한 명확한 문서화
  - PowerShell 사용자를 위한 설치 지침
  - 파일: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`

### 훅

- **docs(hooks)**: 포괄적인 lib/ 문서 추가 (NEW)
  - 새 파일: `src/moai_adk/templates/.claude/hooks/moai/lib/README.md`
  - 사용 예제를 포함한 13개 유틸리티 모듈 문서화
  - 구성, 데이터 구조, 실행, I/O, 품질 유틸리티 포함
  - import 패턴 및 모범 사례 제공

## 변경됨

- **chore**: 의존성 잠금 파일 업데이트 (uv.lock 1.8.0 → 1.8.1) (b78a5324)
  - 호환성을 위한 의존성 버전 업데이트

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 221개 파일 변경 없음
- Mypy: 성공 (173개 소스 파일에서 문제 없음)

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

# v1.8.1 - AST-Grep Parsing Fix & Documentation Improvements (2026-01-24)

## Summary

This patch release resolves ast-grep parsing errors by removing problematic multi-line patterns and improves CHANGELOG structure guidelines for consistent bilingual documentation.

**Key Changes**:
- Fixed ast-grep parsing errors in Python and TypeScript rules
- Improved CHANGELOG structure with clear English → Korean ordering
- Updated release workflow documentation

## Fixed

### AST-Grep Parsing Issues

- **fix(ast-grep)**: Resolve parsing errors by removing multi-line patterns (83fd640b)
  - Removed complex multi-line patterns from `rules/languages/python.yml`
  - Simplified TypeScript patterns in `rules/languages/typescript.yml`
  - Removed unused patterns from `rules/quality/complexity-check.yml` and `deprecated-apis.yml`
  - Updated `sgconfig.yml` with cleaner rule configuration
  - Resolves parsing errors that prevented ast-grep from functioning correctly

## Changed

### Documentation

- **docs(release)**: Improve CHANGELOG structure and GitHub Release notes guidelines (425f3935)
  - Added [HARD] rule: Each version MUST have Korean section IMMEDIATELY after English section
  - Clarified correct structure: English vX.Y.Z → Korean vX.Y.Z → Previous versions
  - Updated GitHub Release notes update instructions with bilingual format

- **docs**: Fix CHANGELOG structure - Korean follows English per section (12391fe1)
  - Ensured proper ordering: English section → Korean section for each version
  - Prevents structure violations where all English sections come before all Korean sections

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

# v1.8.1 - AST-Grep 파싱 수정 및 문서화 개선 (2026-01-24)

## 요약

이번 패치 릴리즈는 문제가 있던 멀티라인 패턴을 제거하여 ast-grep 파싱 에러를 해결하고, 일관된 이중 언어 문서화를 위한 CHANGELOG 구조 가이드라인을 개선합니다.

**주요 변경사항**:
- Python 및 TypeScript 규칙의 ast-grep 파싱 에러 수정
- 영어 → 한국어 순서를 명확히 한 CHANGELOG 구조 개선
- 릴리즈 워크플로우 문서 업데이트

## 수정됨

### AST-Grep 파싱 이슈

- **fix(ast-grep)**: 멀티라인 패턴 제거로 파싱 에러 해결 (83fd640b)
  - `rules/languages/python.yml`에서 복잡한 멀티라인 패턴 제거
  - `rules/languages/typescript.yml`의 TypeScript 패턴 단순화
  - `rules/quality/complexity-check.yml` 및 `deprecated-apis.yml`의 미사용 패턴 제거
  - 더 깔끔한 규칙 설정으로 `sgconfig.yml` 업데이트
  - ast-grep이 올바르게 작동하지 않던 파싱 에러 해결

## 변경됨

### 문서화

- **docs(release)**: CHANGELOG 구조 및 GitHub Release 노트 가이드라인 개선 (425f3935)
  - [HARD] 규칙 추가: 각 버전은 영어 섹션 바로 다음에 한국어 섹션이 있어야 함
  - 올바른 구조 명시: 영어 vX.Y.Z → 한국어 vX.Y.Z → 이전 버전들
  - 이중 언어 형식으로 GitHub Release 노트 업데이트 지침 개선

- **docs**: CHANGELOG 구조 수정 - 각 섹션마다 한국어가 영어 뒤에 옴 (12391fe1)
  - 각 버전별로 적절한 순서 보장: 영어 섹션 → 한국어 섹션
  - 모든 영어 섹션이 모든 한국어 섹션보다 먼저 오는 구조 위반 방지

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

# v1.8.0 - Ralph-Style LSP Integration & Google Stitch MCP (2026-01-23)

## Summary

This major release introduces Ralph-style LSP integration for autonomous workflow execution, Google Stitch MCP integration for AI-powered UI/UX design generation, comprehensive language pattern support for 12 programming languages, and cross-platform path variable consolidation.

**Key Features**:
- Ralph-style LSP integration with autonomous workflow completion
- Google Stitch MCP integration with expert-stitch agent
- Language patterns for 12+ programming languages
- Cross-platform path variable consolidation (Windows/macOS/Linux)
- Star History integration for GitHub repository visibility

**Reference**: SPEC-LSP-001, SPEC-STITCH-001

## Breaking Changes

- **Path Variables**: `{{PROJECT_DIR_UNIX}}` and `{{PROJECT_DIR_WIN}}` are deprecated, use `{{PROJECT_DIR}}` instead
- **Autonomous Mode**: New execution mode requires LSP-capable IDE for optimal functionality
- Default execution mode remains **interactive** (backward compatible)
- Workflow configuration file added: `.moai/config/sections/workflow.yaml`
- New configuration section in `quality.yaml`: `lsp_quality_gates`

## Migration Guide

### For Existing Projects

1. **Update MoAI-ADK**:
   ```bash
   uv tool update moai-adk
   ```

2. **Sync Templates** (adds new files):
   ```bash
   # Sync new configuration files
   rsync -av src/moai_adk/templates/.moai/config/sections/workflow.yaml .moai/config/sections/

   # Sync updated agent and configuration files
   moai update
   ```

3. **Enable Autonomous Mode** (optional):
   Edit `.moai/config/sections/workflow.yaml`:
   ```yaml
   execution_mode:
     autonomous:
       user_approval_required: false
       continuous_loop: true
       completion_marker_based: true
       lsp_feedback_integration: true
   ```

### No Action Required for Interactive Mode

If you prefer manual approval (current behavior), no configuration changes are needed. The default is `execution_mode.interactive.user_approval_required: true`.

## Added

### Google Stitch MCP Integration (NEW)

- **feat(stitch)**: Add Google Stitch MCP integration with expert-stitch agent
  - New agent: `expert-stitch` for AI-powered UI/UX design generation
  - MCP tools: `mcp__stitch__create_project`, `mcp__stitch__generate_screen_from_text`, `mcp__stitch__fetch_screen_code`
  - Design DNA extraction and consistency checking
  - Code export in React, Vue, Angular, and Flutter formats
  - Added to Tier 1 Domain Experts (8 → 9 agents)

- **feat(mcp)**: Add Stitch MCP configuration
  - Updated `.mcp.json` with Google Stitch server configuration
  - Added Windows-compatible configuration in `.mcp.windows.json`

### Language Patterns (NEW)

- **feat(lang)**: Add language-specific code patterns for 12 languages
  - New files: `.moai/rules/language-patterns/*.yaml`
  - Languages: C++, C#, Elixir, Flutter, Go, JavaScript, Kotlin, PHP, R, Ruby, Rust, Scala, Swift
  - Framework-specific patterns for each language
  - Security best practices and anti-patterns

### Statusline Enhancement (NEW)

- **feat(statusline)**: Update statusline format with cleaner output
  - Improved version display format
  - Better visual hierarchy for status information

### Documentation (NEW)

- **docs(readme)**: Add Star History section to all README files
  - Added Star History chart for GitHub repository visibility
  - Synced across EN, KO, JA, ZH versions
  - Updated section numbering (License → 16, Made with → 17)

- **docs(readme)**: Sync expert-stitch agent to README files
  - Updated Tier 1 Domain Experts from 8 to 9
  - Added expert-stitch description in all language versions

### Core LSP Integration (SPEC-LSP-001 Implementation)

- **feat(lsp)**: Add LSP-based completion marker system
  - New module: `src/moai_adk/utils/completion_marker.py`
  - `LSPState` dataclass: Captures diagnostic state (errors, warnings, type_errors, lint_errors)
  - `CompletionMarker` class: Determines phase completion readiness
  - `LoopPrevention` class: Prevents infinite loops in autonomous mode
  - Integration with MCP tool: `mcp__ide__getDiagnostics`
  - 100 unit tests with 100% coverage

- **feat(config)**: Add workflow execution mode configuration
  - New file: `.moai/config/sections/workflow.yaml`
  - Interactive mode: Manual approval (default, backward compatible)
  - Autonomous mode: Ralph-style continuous execution
  - Completion markers per phase: plan, run, sync
  - Loop prevention: max_iterations, no_progress_threshold, stale_detection

- **feat(hooks)**: Add quality gate hook with LSP integration
  - New file: `.claude/hooks/moai/quality_gate_with_lsp.py`
  - Validates LSP diagnostics before sync phase
  - Provides clear error summaries for remaining issues
  - Exit codes: 0 (pass), 1 (fail), 2 (error)
  - CLI interface for standalone execution

- **feat(agent)**: Update manager-ddd agent for Ralph-style autonomous execution
  - Updated: `.claude/agents/moai/manager-ddd.md` (v2.0.0)
  - Autonomous execution mode documentation
  - LSP completion marker integration
  - Quality gate enforcement before sync
  - Backward compatibility with manual approval mode

- **feat(config)**: Add LSP quality gates to quality configuration
  - Updated: `.moai/config/sections/quality.yaml`
  - `lsp_quality_gates` section with phase-specific thresholds
  - `lsp_integration` section with TRUST 5 alignment
  - `lsp_state_tracking` section for observability

### Documentation Enhancements

- **docs(claude)**: Update CLAUDE.md with LSP integration documentation (v10.6.0)
  - Autonomous Execution Mode section
  - Ralph-style workflow documentation
  - LSP quality gate enforcement
  - Backward compatibility notes
  - Troubleshooting guide

- **docs(agent)**: Update all agents with LSP integration patterns
  - 4 builder agents updated (builder-agent, builder-command, builder-plugin, builder-skill)
  - 4 expert agents updated (expert-backend, expert-frontend, expert-refactoring, expert-security)
  - 3 manager agents updated (manager-ddd, manager-docs, manager-quality)
  - Updated: `.claude/commands/moai/loop.md` with LSP integration

## Changed

- **refactor(core)**: Enhance context manager for LSP state tracking
  - Updated: `src/moai_adk/core/context_manager.py`
  - LSP state capture at phase boundaries
  - Baseline comparison for regression detection

- **chore(utils)**: Export completion_marker module
  - Updated: `src/moai_adk/utils/__init__.py`
  - Public API: `from moai_adk.utils import LSPState, CompletionMarker, LoopPrevention`

- **refactor(config)**: Add workflow configuration section loader
  - Updated: `src/moai_adk/templates/.moai/config/config.yaml`
  - Includes `workflow.yaml` section in main configuration

- **chore(template)**: Sync template files to local project
  - All template changes synced to local `.claude/` and `.moai/` directories

## Technical Details

### LSP State Tracking

The system tracks LSP diagnostic state at multiple points:

- **Phase Start**: Capture baseline diagnostic state
- **Post-Transformation**: Compare current state against baseline
- **Pre-Sync**: Validate all quality gates before PR

### Completion Markers

Each phase has specific completion criteria:

**Plan Phase**:
- SPEC document created
- LSP baseline recorded

**Run Phase**:
- Tests passing
- Behavior preserved (no regression from baseline)
- LSP errors: 0 (configurable threshold)
- Type errors: 0
- Lint errors: 0

**Sync Phase**:
- Documentation generated
- LSP clean (0 errors, 0 warnings for PR)
- Quality gate passed

### Loop Prevention

Autonomous mode includes safeguards:

- **Max Iterations**: 100 (configurable in `workflow.yaml`)
- **No Progress Threshold**: 5 iterations without error reduction
- **Stale Detection**: Same fix attempted twice, error count unchanged
- **Regression Detection**: Significant error increase triggers stop

### TRUST 5 Integration

LSP diagnostics align with TRUST 5 quality framework:

- **Tested**: Unit tests pass, type errors == 0
- **Readable**: Naming conventions followed, lint errors == 0
- **Unified**: Code formatted, warnings < threshold
- **Secured**: Security scan pass, security warnings == 0
- **Trackable**: LSP state changes logged, diagnostic history tracked

## Quality

- Unit Tests: 100 tests for `completion_marker.py` (100% coverage)
- Smoke Tests: All passing
- Ruff: All checks passed
- Mypy: Success
- Test Coverage: 88.12% overall (exceeds 85% target)

## Configuration Examples

### Strict Mode (Production)

```yaml
# .moai/config/sections/workflow.yaml
execution_mode:
  autonomous:
    user_approval_required: false
    continuous_loop: true

completion_markers:
  run:
    lsp_errors: 0
    type_errors: 0
  sync:
    lsp_clean: true
```

### Lenient Mode (Prototyping)

```yaml
# .moai/config/sections/workflow.yaml
execution_mode:
  autonomous:
    user_approval_required: false

completion_markers:
  run:
    lsp_errors: 5  # Allow some errors during prototyping
  sync:
    max_warnings: 50  # Allow more warnings
```

### Interactive Mode (Default)

```yaml
# .moai/config/sections/workflow.yaml
execution_mode:
  interactive:
    user_approval_required: true
```

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates in your folder
moai update

# Enable autonomous mode (optional)
edit .moai/config/sections/workflow.yaml
```

## Acknowledgments

This feature implements SPEC-LSP-001 (LSP-Based Autonomous Workflow Completion) following the ANALYZE-PRESERVE-IMPROVE DDD methodology with comprehensive characterization tests ensuring backward compatibility.

---

# v1.8.0 - Ralph-Style LSP 통합 & Google Stitch MCP (2026-01-23)

## 요약

이 메이저 릴리스는 자율 워크플로우 실행을 위한 Ralph-style LSP 통합, AI 기반 UI/UX 디자인 생성을 위한 Google Stitch MCP 통합, 12개 프로그래밍 언어에 대한 포괄적인 언어 패턴 지원, 크로스 플랫폼 경로 변수 통합을 도입합니다.

**주요 기능**:
- 자율 워크플로우 완료를 위한 Ralph-style LSP 통합
- expert-stitch 에이전트와 Google Stitch MCP 통합
- 12개 이상 프로그래밍 언어를 위한 언어 패턴
- 크로스 플랫폼 경로 변수 통합 (Windows/macOS/Linux)
- GitHub 저장소 가시성을 위한 Star History 통합

**참조**: SPEC-LSP-001, SPEC-STITCH-001

## Breaking Changes

- **경로 변수**: `{{PROJECT_DIR_UNIX}}`와 `{{PROJECT_DIR_WIN}}`은 더 이상 사용되지 않으며, `{{PROJECT_DIR}}`를 사용하세요
- **자율 모드**: 새로운 실행 모드는 최적의 기능을 위해 LSP 지원 IDE가 필요합니다
- 기본 실행 모드는 **대화형** 유지 (하위 호환)
- 워크플로우 구성 파일 추가: `.moai/config/sections/workflow.yaml`
- `quality.yaml`에 새 구성 섹션: `lsp_quality_gates`

## 추가됨

### Google Stitch MCP 통합 (NEW)

- **feat(stitch)**: expert-stitch 에이전트와 Google Stitch MCP 통합 추가
  - AI 기반 UI/UX 디자인 생성을 위한 새 에이전트: `expert-stitch`
  - MCP 도구: `mcp__stitch__create_project`, `mcp__stitch__generate_screen_from_text`, `mcp__stitch__fetch_screen_code`
  - Design DNA 추출 및 일관성 검사
  - React, Vue, Angular, Flutter 형식으로 코드 내보내기
  - Tier 1 도메인 전문가에 추가 (8 → 9개 에이전트)

### 언어 패턴 (NEW)

- **feat(lang)**: 12개 언어에 대한 언어별 코드 패턴 추가
  - 새 파일: `.moai/rules/language-patterns/*.yaml`
  - 언어: C++, C#, Elixir, Flutter, Go, JavaScript, Kotlin, PHP, R, Ruby, Rust, Scala, Swift
  - 각 언어별 프레임워크 패턴
  - 보안 모범 사례 및 안티패턴

### LSP 통합 (SPEC-LSP-001)

- **feat(lsp)**: LSP 기반 완료 마커 시스템 추가
  - 새 모듈: `src/moai_adk/utils/completion_marker.py`
  - 자율 모드에서 무한 루프 방지
  - MCP 도구와 통합: `mcp__ide__getDiagnostics`

### 문서 (NEW)

- **docs(readme)**: 모든 README 파일에 Star History 섹션 추가
  - EN, KO, JA, ZH 버전 동기화
  - expert-stitch 에이전트 설명 추가

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
