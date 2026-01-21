# v1.5.6 - AST-Grep Integration & Performance Enhancements (2026-01-21)

## Summary

This feature release integrates AST-Grep structural code analysis into all quality check commands (`/moai:loop`, `/moai:fix`) and adds the `moai glm` shortcut command for quick LLM backend switching. Also includes statusline enhancements with visual context graph display.

## Added

- **feat(performance)**: Integrate AST-Grep into parallel diagnosis (11787f87)
  - AST-Grep now runs alongside LSP, Tests, and Coverage in parallel
  - 3.75x faster code quality diagnosis with concurrent structural analysis
  - Detects security vulnerabilities, code smells, anti-patterns, and best practice violations
  - Supports 40+ programming languages (Python, TypeScript, Go, Rust, Java, etc.)
  - Integration points: `/moai:loop`, `/moai:fix`, and `/moai:alfred` workflows

- **feat(ux)**: Add `moai glm` shortcut command for quick backend switching
  - Quick switch to GLM backend: `moai glm`
  - Update API key: `moai glm <your-api-key>`
  - Switch back to Claude: `moai claude`
  - Useful for Worktree parallel development workflows

## Changed

- **feat(statusline)**: Enhanced visual context graph display (519bd367)
  - Context window usage now shown as visual graph bar
  - Battery-style color-coded display (green/yellow/red based on usage)
  - Simplified statusline with focus on context visualization
  - ANSI graph support for terminal compatibility

## Fixed

- **fix(update)**: Protect statusline-config.yaml from overwrite during moai update (b8cf28dd)
  - User statusline settings now preserved during updates

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 215 files unchanged
- Mypy: Success (no issues found in 169 source files)

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

# v1.5.6 - AST-Grep 통합 및 성능 향상 (2026-01-21)

## 요약

이 기능 릴리스는 AST-Grep 구조적 코드 분석을 모든 품질 검사 명령어(`/moai:loop`, `/moai:fix`)에 통합하고 빠른 LLM 백엔드 전환을 위한 `moai glm` 단축 명령어를 추가합니다. 또한 시각적 컨텍스트 그래프 표시로 statusline을 향상시킵니다.

## 추가됨

- **feat(performance)**: 병렬 진단에 AST-Grep 통합 (11787f87)
  - AST-Grep이 이제 LSP, Tests, Coverage와 함께 병렬로 실행됨
  - 동시 구조 분석으로 3.75배 더 빠른 코드 품질 진단
  - 보안 취약점, 코드 스멀, 안티 패턴, 모범 사례 위반 감지
  - 40개 이상의 프로그래밍 언어 지원 (Python, TypeScript, Go, Rust, Java 등)
  - 통합 지점: `/moai:loop`, `/moai:fix`, `/moai:alfred` 워크플로우

- **feat(ux)**: 빠른 백엔드 전환을 위한 `moai glm` 단축 명령어 추가
  - GLM 백엔드로 빠른 전환: `moai glm`
  - API 키 업데이트: `moai glm <your-api-key>`
  - Claude로 다시 전환: `moai claude`
  - Worktree 병렬 개발 워크플로우에 유용

## 변경됨

- **feat(statusline)**: 향상된 시각적 컨텍스트 그래프 표시 (519bd367)
  - 컨텍스트 윈도우 사용량이 시각적 그래프 바로 표시됨
  - 배터리 스타일 색상 코딩 디스플레이 (사용량 기반 green/yellow/red)
  - 컨텍스트 시각화에 집중한 단순화된 statusline
  - 터미널 호환성을 위한 ANSI 그래프 지원

## 수정됨

- **fix(update)**: moai update 중 statusline-config.yaml 덮어쓰기 방지 (b8cf28dd)
  - 업데이트 중 사용자 statusline 설정 보존

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 215개 파일 변경 없음
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

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

# v1.5.5 - StatusLine Config Protection (2026-01-21)

## Summary

This patch release fixes an issue where user statusline settings (token usage display) were overwritten during `moai update`. The fix ensures that `statusline-config.yaml` is now protected from template overwrites, preserving user customizations.

## Fixed

- **fix(update)**: Protect statusline-config.yaml from overwrite during moai update (b8cf28dd)
  - Added `statusline-config.yaml` to `template_protected_paths` in template processor
  - User statusline settings (token usage display 💰, etc.) now preserved during `moai update`
  - Resolves issue where statusline customizations were lost after updates
  - File: `src/moai_adk/core/template/processor.py`

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 215 files unchanged
- Mypy: Success (no issues found in 169 source files)

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

# v1.5.5 - StatusLine 설정 보호 (2026-01-21)

## 요약

이 패치 릴리스는 `moai update` 실행 시 사용자 statusline 설정(토큰 사용량 표시)이 덮어써지는 문제를 수정합니다. 이제 `statusline-config.yaml`이 템플릿 덮어쓰기로부터 보호되어 사용자 커스터마이징이 보존됩니다.

## 수정됨

- **fix(update)**: moai update 시 statusline-config.yaml 덮어쓰기 방지 (b8cf28dd)
  - 템플릿 프로세서의 `template_protected_paths`에 `statusline-config.yaml` 추가
  - 사용자 statusline 설정(토큰 사용량 표시 💰 등)이 `moai update` 시 보존됨
  - 업데이트 후 statusline 커스터마이징이 사라지던 문제 해결
  - 파일: `src/moai_adk/core/template/processor.py`

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 215개 파일 변경 없음
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

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

# v1.5.4 - SPEC Validation Guidelines Enhancement (2026-01-21)

## Summary

This patch release enhances SPEC creation guidelines with comprehensive validation rules, classification logic, and migration guides to prevent common SPEC organization issues. Also includes a minor bug fix for missing import in rank command.

## Added

- **docs(spec)**: Add SPEC validation and classification guidelines (f5115252)
  - Added PHASE 1.5 Pre-Creation Validation Gate to `1-plan.md`
    - SPEC Type Classification (SPEC vs Report vs Documentation)
    - Pre-Creation Validation Checklist (4 mandatory checks)
    - Allowed Domain Names (6 categories, 25+ domains)
    - Validation Failure Responses
  - Added SPEC vs Report Classification to `manager-spec.md`
    - Document Type Decision Matrix
    - Classification Algorithm (3-step process)
    - Report Creation Guidelines
    - Flat File Rejection (Enhanced)
  - Added SPEC Scope and Migration Guide to `SKILL.md`
    - What Belongs / Does NOT Belong in `.moai/specs/`
    - Migration scenarios for legacy files (4 scenarios)
    - Validation script reference
  - Files: `.claude/commands/moai/1-plan.md`, `.claude/agents/moai/manager-spec.md`, `.claude/skills/moai-workflow-spec/SKILL.md`
  - Package templates updated in `src/moai_adk/templates/.claude/`

## Fixed

- **fix(rank)**: Add missing import for `_safe_run_subprocess` (889a9f31)
  - Added import: `from moai_adk.core.claude_integration import _safe_run_subprocess`
  - Resolves F821 undefined name error
  - File: `src/moai_adk/cli/commands/rank.py`

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed (1 issue auto-fixed)
- Ruff format: 215 files unchanged
- Mypy: Success (no issues found in 169 source files)

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

# v1.5.4 - SPEC 검증 가이드라인 강화 (2026-01-21)

## 요약

이 패치 릴리스는 SPEC 생성 가이드라인을 포괄적인 검증 규칙, 분류 로직, 마이그레이션 가이드로 강화하여 일반적인 SPEC 조직 문제를 방지합니다. rank 명령어의 누락된 import에 대한 소규모 버그 수정도 포함됩니다.

## 추가됨

- **docs(spec)**: SPEC 검증 및 분류 가이드라인 추가 (f5115252)
  - `1-plan.md`에 PHASE 1.5 Pre-Creation Validation Gate 추가
    - SPEC 타입 분류 (SPEC vs Report vs Documentation)
    - Pre-Creation 검증 체크리스트 (4가지 필수 검사)
    - 허용된 도메인 이름 (6개 카테고리, 25개 이상 도메인)
    - 검증 실패 응답
  - `manager-spec.md`에 SPEC vs Report 분류 추가
    - 문서 타입 의사결정 매트릭스
    - 분류 알고리즘 (3단계 프로세스)
    - Report 생성 가이드라인
    - 플랫 파일 거부 (강화됨)
  - `SKILL.md`에 SPEC 범위 및 마이그레이션 가이드 추가
    - `.moai/specs/`에 포함되어야 할/포함되지 말아야 할 항목
    - 레거시 파일을 위한 마이그레이션 시나리오 (4가지 시나리오)
    - 검증 스크립트 참조
  - 파일: `.claude/commands/moai/1-plan.md`, `.claude/agents/moai/manager-spec.md`, `.claude/skills/moai-workflow-spec/SKILL.md`
  - `src/moai_adk/templates/.claude/`의 패키지 템플릿 업데이트됨

## 수정됨

- **fix(rank)**: `_safe_run_subprocess`에 대한 누락된 import 추가 (889a9f31)
  - import 추가: `from moai_adk.core.claude_integration import _safe_run_subprocess`
  - F821 정의되지 않은 이름 오류 해결
  - 파일: `src/moai_adk/cli/commands/rank.py`

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과 (1개 이슈 자동 수정됨)
- Ruff format: 215개 파일 변경 없음
- Mypy: 성공 (169개 소스 파일에서 문제 없음)

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

# v1.5.3 - Memory Leak Fixes and Critical Bug Fixes (2026-01-20)

## Summary

This patch release resolves critical memory issues reported in GitHub issues #282 and #284 that were causing crashes during agent execution. It also fixes missing context_window field in statusline config (#283) and includes various documentation improvements.

## Fixed

- **fix(memory)**: Resolve JavaScript heap exhaustion during agent execution (#282, #284)
  - Added `_safe_run_subprocess()` helper function with memory protection
  - Added timeout (60s), max output size (1MB), and max lines (1000) limits
  - Applied to all subprocess calls in rank, update, issue_creator, claude_integration
  - Prevents unbounded memory accumulation from subprocess outputs

- **fix(cache)**: Add ContextCache memory limits (#282, #284)
  - Added total memory limit: 100MB for cache (increased from 50MB)
  - Added per-entry memory limit: 10MB per skill
  - Fixed memory calculation for strings, dicts, and lists
  - Added LRU eviction when memory limits exceeded
  - Added warning logs when approaching 90% capacity
  - File: `src/moai_adk/core/jit_context_loader.py`

- **fix(session)**: Add SessionManager result limits (#282, #284)
  - Added max_results limit: 100 results stored in memory
  - Added max_result_size_mb limit: 10MB per result
  - Implemented LRU eviction for old results
  - Added `_truncate_result()` for large result handling
  - Prevents unbounded result storage causing memory exhaustion
  - File: `src/moai_adk/core/session_manager.py`

- **fix(statusline)**: Add missing context_window field to DisplayConfig (#283)
  - Added `context_window: bool = True` field to DisplayConfig dataclass
  - Updated `get_display_config()` to read context_window from YAML
  - Updated `_get_default_config()` to include context_window: True
  - File: `src/moai_adk/statusline/config.py`

## Changed

- **docs(pip/uv)**: Add pip/uv conflict resolution to all README files
  - Added "Known Issues & Solutions" section with detailed troubleshooting
  - Covered symptoms, root causes, and three resolution options
  - Included platform-specific instructions for macOS/Linux/Windows
  - Files: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`

- **docs**: Remove docs/ from .gitignore and add documentation files
  - Documentation is now tracked in git
  - Added troubleshooting and installation guides

## Quality

- Smoke tests: 6 passed (100% pass rate)
- Ruff: All checks passed
- Mypy: Success (no issues found in 169 source files)
- Code coverage: Maintained at previous levels

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

# v1.5.3 - 메모리 누수 수정 및 중요 버그 수정 (2026-01-20)

## 요약

이 패치 릴리스는 GitHub 이슈 #282, #284에서 보고된 에이전트 실행 중 충돌을 일으키는 치명적인 메모리 문제를 해결합니다. 또한 statusline config의 누락된 context_window 필드를 수정(#283)하고 다양한 문서 개선이 포함됩니다.

## 수정됨

- **fix(memory)**: 에이전트 실행 중 JavaScript 힙 고갈 해결 (#282, #284)
  - 메모리 보호를 위한 `_safe_run_subprocess()` 헬퍼 함수 추가
  - 타임아웃(60초), 최대 출력 크기(1MB), 최대 라인 수(1000) 제한 추가
  - rank, update, issue_creator, claude_integration의 모든 subprocess 호출에 적용
  - subprocess 출력으로 인한 무제한 메모리 축적 방지

- **fix(cache)**: ContextCache 메모리 한도 추가 (#282, #284)
  - 전체 메모리 한도: 캐시 100MB (기존 50MB에서 증가)
  - 항목별 메모리 한도: 스킬당 10MB
  - string, dict, list에 대한 메모리 계산 수정
  - 메모리 한도 초과 시 LRU 퇴거 추가
  - 90% 용량 접근 시 경고 로그 추가
  - 파일: `src/moai_adk/core/jit_context_loader.py`

- **fix(session)**: SessionManager 결과 제한 추가 (#282, #284)
  - 최대 결과 수 제한: 메모리에 100개 결과 저장
  - 결과당 최대 크기 제한: 결과당 10MB
  - 오래된 결과를 위한 LRU 퇴거 구현
  - 대용량 결과 처리를 위한 `_truncate_result()` 추가
  - 메모리 고갈을 일으키는 무제한 결과 저장 방지
  - 파일: `src/moai_adk/core/session_manager.py`

- **fix(statusline)**: DisplayConfig에 누락된 context_window 필드 추가 (#283)
  - DisplayConfig dataclass에 `context_window: bool = True` 필드 추가
  - YAML에서 context_window를 읽도록 `get_display_config()` 업데이트
  - context_window: True를 포함하도록 `_get_default_config()` 업데이트
  - 파일: `src/moai_adk/statusline/config.py`

## 변경됨

- **docs(pip/uv)**: 모든 README 파일에 pip/uv 충돌 해결 방법 추가
  - 상세 문제 해결을 포함한 "알려진 문제 및 해결 방법" 섹션 추가
  - 증상, 근본 원인, 3가지 해결 방법 포함
  - macOS/Linux/Windows용 플랫폼별 지침 포함
  - 파일: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`

- **docs**: .gitignore에서 docs/ 제거 및 문서 파일 추가
  - 문서가 이제 git에서 추적됨
  - 문제 해결 및 설치 가이드 추가

## 품질

- Smoke 테스트: 6개 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Mypy: 성공 (169개 소스 파일에서 문제 없음)
- 코드 커버리지: 이전 수준 유지

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

# v1.5.2 - Critical Bug Fixes for Windows StatusLine, Hook Uninstall, and Feedback Language (2026-01-20)
