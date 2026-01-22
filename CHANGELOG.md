# v1.7.0 - UltraThink Mode Enhancement (2026-01-22)

## Summary

This minor release introduces **UltraThink mode** (`--ultrathink`), an enhanced analysis feature that automatically applies Sequential Thinking MCP for deep request analysis and optimal execution planning. Additionally, this release includes important agent tool permission updates and parallel execution safeguards for improved reliability.

## Added

- **feat(command)**: Add `--ultrathink` keyword guidance to complex commands
  - UltraThink mode automatically activates Sequential Thinking MCP for deep analysis
  - Applied to 6 commands: `/moai:alfred`, `/moai:0-project`, `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync`, `/moai:99-release`
  - Each command includes WHY/IMPACT documentation for the guidance
- **docs(claude)**: Add Parallel Execution Safeguards section
  - File write conflict prevention mechanisms
  - Agent tool requirements documentation
  - Loop prevention guards
  - Platform compatibility guidelines
- **docs**: Add Sequential Thinking MCP Support section to README files
  - UltraThink mode documentation for all language versions
  - Agent-specific UltraThink examples

## Fixed

- **fix(agent)**: Update all agent tool permissions to full access
  - All agents now have Read, Write, Edit, Grep, Glob, Bash, TodoWrite, Task, Skill tools
  - Ensures agents can perform code modifications without platform-specific issues
- **fix(agent)**: Add Edit and Write tools to expert-debug agent
  - Prevents fallback to Bash commands that may fail on different platforms
  - Enables cross-platform code editing capabilities
- **fix(#288)**: session_start hook fails to detect moai-adk version with uv tool installation
  - Replace Python import with subprocess call to `moai --version` CLI command
  - Works correctly with uv tool isolated installations
  - Graceful fallback to config if CLI command fails
- **fix(#287)**: moai rank sync incorrectly counts duplicates as failed
  - Add `_is_duplicate_error()` helper function with robust pattern matching
  - Support 7 error patterns across 3 languages (EN, KO, ZH)
  - Replace 3 duplicate detection locations with helper function
  - Add 9 new tests for duplicate detection
- **fix(alfred)**: Optimize LSP diagnostics to prevent infinite loops
  - Reduce LSP timeout from 30 to 15 seconds for faster failure detection
  - Increase poll interval from 500ms to 1000ms to reduce CPU usage
  - Change GLM Haiku model from glm-4.7-flashx to glm-4.7

## Changed

- **chore(build)**: Exclude template README files from package distribution
  - Template README files are now local-only (not distributed in package)
  - Prevents confusion between template and user README files

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 216 files unchanged
- Mypy: Warning (3 minor type hint issues in non-critical code)

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

# v1.7.0 - UltraThink 모드 기능 강화 (2026-01-22)

## 요약

이 마이너 릴리스는 **UltraThink 모드**(`--ultrathink`)를 도입하여, Sequential Thinking MCP를 자동으로 적용하는 심층 분석 기능을 추가합니다. 또한 에이전트 도구 권한 업데이트 및 병렬 실행 안전장치를 포함하여 신뢰성을 개선했습니다.

## 추가됨

- **feat(command)**: 복잡한 커맨드에 `--ultrathink` 키워드 가이드 추가
  - UltraThink 모드는 심층 분석을 위해 Sequential Thinking MCP를 자동 활성화
  - 6개 커맨드에 적용: `/moai:alfred`, `/moai:0-project`, `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync`, `/moai:99-release`
  - 각 커맨드에는 WHY/IMPACT 문서 포함
- **docs(claude)**: 병렬 실행 안전장치 섹션 추가
  - 파일 쓰기 충돌 방지 메커니즘
  - 에이전트 도구 요구사항 문서화
  - 루프 방지 가드
  - 플랫폼 호환성 가이드라인
- **docs**: README 파일에 Sequential Thinking MCP 지원 섹션 추가
  - 모든 언어 버전의 UltraThink 모드 문서
  - 에이전트별 UltraThink 예제

## 수정됨

- **fix(agent)**: 모든 에이전트 도구 권한을 전체 액세스로 업데이트
  - 모든 에이전트가 Read, Write, Edit, Grep, Glob, Bash, TodoWrite, Task, Skill 도구 보유
  - 플랫폼별 문제 없이 코드 수정 수행 가능
- **fix(agent)**: expert-debug 에이전트에 Edit 및 Write 도구 추가
  - 다른 플랫폼에서 실패할 수 있는 Bash 명령 대체 방지
  - 크로스 플랫폼 코드 편집 기능 활성화
- **fix(#288)**: session_start hook이 uv tool 설치 시 moai-adk 버전 감지 실패
  - Python import를 `moai --version` CLI 명령 호출로 교체
  - uv tool 격리 설치 환경에서 올바르게 작동
  - CLI 실패 시 config로 우아한 fallback
- **fix(#287)**: moai rank sync가 중복을 실패로 잘못 카운트
  - 강건한 패턴 매칭을 위한 `_is_duplicate_error()` 헬퍼 함수 추가
  - 3개 언어(EN, KO, ZH)의 7개 에러 패턴 지원
  - 3곳의 중복 감지 위치를 헬퍼 함수로 통합
  - 중복 감지 테스트 9개 추가
- **fix(alfred)**: 무한 루프 방지를 위한 LSP 진단 최적화
  - 더 빠른 실패 감지를 위해 LSP 타임아웃 30→15초 감소
  - CPU 사용량 감소를 위해 폴링 간격 500→1000ms 증가
  - GLM Haiku 모델 glm-4.7-flashx → glm-4.7로 변경

## 변경됨

- **chore(build)**: 패키지 배포에서 템플릿 README 파일 제외
  - 템플릿 README 파일은 이제 로컬 전용 (패키지에 미포함)
  - 템플릿과 사용자 README 파일 간의 혼선 방지

## 품질

- Smoke tests: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 216 파일 변경 없음
- Mypy: 경고 (3개의 사소한 타입 힌트 문제)

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

# v1.6.4 - GLM API Endpoint Update (2026-01-22)

## Summary

This patch release updates the GLM API endpoint to the official BigModel API domain for improved service reliability and compatibility.

## Changed

- **fix(config)**: Update GLM API endpoint to official BigModel domain
  - Changed base URL: `https://api.z.ai/api/anthropic` → `https://open.bigmodel.cn/api/anthropic`
  - Ensures compatibility with latest GLM API infrastructure
  - Updated in both local and template configuration files

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 216 files unchanged
- Mypy: Warning (3 minor type hint issues in non-critical code)

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

# v1.6.4 - GLM API 엔드포인트 업데이트 (2026-01-22)

## 요약

이 패치 릴리스는 GLM API 엔드포인트를 공식 BigModel 도메인으로 업데이트하여 서비스 안정성과 호환성을 개선합니다.

## 변경됨

- **fix(config)**: 공식 BigModel 도메인으로 GLM API 엔드포인트 업데이트
  - 베이스 URL 변경: `https://api.z.ai/api/anthropic` → `https://open.bigmodel.cn/api/anthropic`
  - 최신 GLM API 인프라와의 호환성 보장
  - 로컬 및 템플릿 구성 파일 모두 업데이트

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 216개 파일 변경 없음
- Mypy: 경고 (비임계 코드 3개의 사소한 타입 힌트 문제)

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

# v1.6.3 - GLM API Endpoint Update (2026-01-22)

## Summary

This patch release updates the GLM API endpoint to the official BigModel API domain. The base URL has been changed from `https://api.z.ai/api/anthropic` to `https://open.bigmodel.cn/api/anthropic` for improved service reliability and performance.

## Changed

- **fix(config)**: Update GLM API endpoint to official BigModel domain
  - Changed base URL: `https://api.z.ai/api/anthropic` → `https://open.bigmodel.cn/api/anthropic`
  - Ensures compatibility with latest GLM API infrastructure
  - Updated in both local and template configuration files

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 216 files unchanged
- Mypy: Warning (3 minor type hint issues in non-critical code)

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

# v1.6.3 - GLM API 엔드포인트 업데이트 (2026-01-22)

## 요약

이 패치 릴리스는 GLM API 엔드포인트를 공식 BigModel 도메인으로 업데이트합니다. 개선된 서비스 안정성과 성능을 위해 베이스 URL이 `https://api.z.ai/api/anthropic`에서 `https://open.bigmodel.cn/api/anthropic`으로 변경되었습니다.

## 변경됨

- **fix(config)**: 공식 BigModel 도메인으로 GLM API 엔드포인트 업데이트
  - 베이스 URL 변경: `https://api.z.ai/api/anthropic` → `https://open.bigmodel.cn/api/anthropic`
  - 최신 GLM API 인프라와의 호환성 보장
  - 로컬 및 템플릿 구성 파일 모두 업데이트

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 216개 파일 변경 없음
- Mypy: 경고 (비임계 코드 3개의 사소한 타입 힌트 문제)

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

# v1.6.2 - GLM Model Name Standardization (2026-01-22)

## Summary

This patch release standardizes GLM model names to use lowercase naming convention for consistency with the API specification. The model names `GLM-4.7-FlashX`, `GLM-4.7` have been updated to `glm-4.7-flashx`, `glm-4.7` respectively.

## Changed

- **fix(config)**: Update GLM model names to lowercase (e31114cd)
  - Changed `GLM-4.7-FlashX` → `glm-4.7-flashx`
  - Changed `GLM-4.7` → `glm-4.7`
  - Ensures consistency with API specification
  - Updated in both local and template configuration files

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 216 files unchanged
- Mypy: Warning (3 minor type hint issues in non-critical code)

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

# v1.6.2 - GLM 모델명 표준화 (2026-01-22)

## 요약

이 패치 릴리스는 API 사양과의 일관성을 위해 GLM 모델명을 소문자 명명 규칙으로 표준화합니다. `GLM-4.7-FlashX`, `GLM-4.7` 모델명이 각각 `glm-4.7-flashx`, `glm-4.7`로 업데이트되었습니다.

## 변경됨

- **fix(config)**: GLM 모델명을 소문자로 업데이트 (e31114cd)
  - `GLM-4.7-FlashX` → `glm-4.7-flashx` 변경
  - `GLM-4.7` → `glm-4.7` 변경
  - API 사양과의 일관성 보장
  - 로컬 및 템플릿 구성 파일 모두 업데이트

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 216개 파일 변경 없음
- Mypy: 경고 (비임계 코드 3개의 사소한 타입 힌트 문제)

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

# v1.6.1 - StatusLine Command Simplification (2026-01-22)

## Summary

This patch release simplifies the statusline command configuration by using the `moai statusline` subcommand instead of the complex PYTHONPATH-based invocation. This makes the configuration more portable and easier to maintain across projects.

## Changed

- **refactor(config)**: Simplify statusline command to use subcommand (ae954835)
  - Changed from: `PYTHONPATH="$CLAUDE_PROJECT_DIR/src" python3 -m moai_adk.statusline.main`
  - Changed to: `moai statusline`
  - Benefits:
    - Removes `$CLAUDE_PROJECT_DIR` dependency
    - Works consistently across all projects
    - Only requires `uv tool install moai-adk`
    - Simpler and more maintainable configuration

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 216 files unchanged
- Mypy: Warning (3 minor type hint issues in non-critical code)

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

# v1.6.1 - StatusLine 명령어 단순화 (2026-01-22)

## 요약

이 패치 릴리스는 복잡한 PYTHONPATH 기반 호출 대신 `moai statusline` 서브명령어를 사용하여 statusline 명령 설정을 단순화합니다. 이를 통해 설정이 더욱 이식성 있고 프로젝트 간 유지보수가 쉬워집니다.

## 변경됨

- **refactor(config)**: statusline 명령어를 서브명령어로 단순화 (ae954835)
  - 변경 전: `PYTHONPATH="$CLAUDE_PROJECT_DIR/src" python3 -m moai_adk.statusline.main`
  - 변경 후: `moai statusline`
  - 장점:
    - `$CLAUDE_PROJECT_DIR` 의존성 제거
    - 모든 프로젝트에서 일관되게 동작
    - `uv tool install moai-adk`만 필요
    - 더 간단하고 유지보수하기 쉬운 설정

## 품질

- Smoke 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 216개 파일 변경 없음
- Mypy: 경고 (비임계 코드 3개의 사소한 타입 힌트 문제)

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

# v1.6.0 - Sequential Thinking MCP & Statusline Enhancements (2026-01-22)

## Summary

This feature release integrates Sequential Thinking MCP tool for complex problem-solving, enhances statusline with battery-style context graph and version display, and improves Explore agent performance with anti-bottleneck optimizations. Also includes AST-Grep security enhancements and various bug fixes.

## Added

- **feat(mcp)**: Integrate Sequential Thinking MCP across workflow (ebb3e73f)
  - Structured reasoning with step-by-step breakdown
  - Context maintenance across multiple reasoning steps
  - Ability to revise and adjust thinking based on new information
  - Automatic activation for complex decisions and architecture choices

- **feat(statusline)**: Battery icon with color-coded graph (957a8620)
  - Visual context window usage display with battery-style icon
  - Color-coded display (green/yellow/red) based on usage percentage
  - Improved visual feedback for token consumption

- **feat(statusline)**: Show used tokens percentage and add version display (9ded4fda)
  - Display token usage as percentage in statusline
  - Show MoAI-ADK version in statusline
  - Better visibility of resource utilization

- **feat(orchestration)**: Explore agent anti-bottleneck system (11787f87)
  - AST-Grep priority for structural code search
  - Search scope limitation to prevent unnecessary scanning
  - File pattern specificity for 50-80% reduction in scanned files
  - Parallel processing optimization

- **feat(performance)**: Enhance orchestration and integrate AST-Grep (74e5e3f0)
  - AST-Grep security rule enhancements
  - Performance optimizations for code quality checks
  - XSS prevention rules updated

## Changed

- **docs**: Add Sequential Thinking MCP to README files (014f08cf)
  - Documentation updates for new MCP integration
  - Usage examples and patterns documented

- **revert**: Remove unnecessary exploration constraints system (170c257a)
  - Simplified agent orchestration
  - Reduced overhead in explore operations

## Fixed

- **fix(statusline)**: Remove ANSI escape codes from graph rendering (a98cf1b6)
  - Clean graph display without escape codes

- **fix(statusline)**: Include output_tokens in context window calculation (b0702cdf)
  - Accurate token usage calculation including output tokens

- **fix(statusline)**: Add fallback calculation from tokens when percentages not provided (bdbc98d9)
  - Graceful fallback when percentage data unavailable

- **fix(statusline)**: Display context graph using percentage instead of string (5caec0c3)
  - Consistent graph display using percentage values

- **fix(statusline)**: Always show context graph instead of token count (519bd367)
  - Improved visual consistency in statusline

- **fix(statusline)**: Calculate percentage from tokens when not provided (b57197da)
  - Automatic percentage calculation from raw token counts

- **fix(commands)**: Restore 99-release.md command (cd9e3266)
  - Release command restored for local development

## Quality

- Smoke tests: 6/6 passed (100% pass rate)
- Ruff: All checks passed
- Ruff format: 216 files unchanged
- Mypy: Success (3 minor type hint issues in non-critical code)

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

# v1.6.0 - Sequential Thinking MCP 및 Statusline 개선 (2026-01-22)

## 요약

이 기능 릴리스는 Sequential Thinking MCP 도구를 통합하여 복잡한 문제 해결을 지원하고, 상태 표시줄(statusline)을 배터리 스타일 컨텍스트 그래프와 버전 표시로 개선합니다. 또한 Explore agent 성능을 최적화하고 AST-Grep 보안 규칙을 강화했습니다.

## 추가됨

- **feat(mcp)**: 워크플로우에 Sequential Thinking MCP 통합 (ebb3e73f)
  - 단계별 분석을 통한 구조화된 추론
  - 여러 추론 단계에서의 컨텍스트 유지
  - 새로운 정보에 기반한 추론 수정 및 조정 기능
  - 복잡한 의사결정 및 아키텍처 선택 시 자동 활성화

- **feat(statusline)**: 색상 코딩된 그래프와 배터리 아이콘 (957a8620)
  - 배터리 스타일 아이콘으로 시각적 컨텍스트 창 사용량 표시
  - 사용량 비율 기반 색상 코딩 (녹색/노란색/빨간색)
  - 토큰 소비에 대한 개선된 시각적 피드백

- **feat(statusline)**: 사용된 토큰 비율 표시 및 버전 표시 추가 (9ded4fda)
  - 상태 표시줄에 토큰 사용량을 백분율로 표시
  - MoAI-ADK 버전을 상태 표시줄에 표시
  - 리소스 활용도에 대한 더 나은 가시성

- **feat(orchestration)**: Explore agent 병목 방지 시스템 (11787f87)
  - 구조적 코드 검색을 위한 AST-Grep 우선순위
  - 불필요한 스캔 방지를 위한 검색 범위 제한
  - 50-80% 스캔 파일 감소를 위한 특정 파일 패턴
  - 병렬 처리 최적화

- **feat(performance)**: 오케스트레이션 강화 및 AST-Grep 통합 (74e5e3f0)
  - AST-Grep 보안 규칙 강화
  - 코드 품질 검사를 위한 성능 최적화
  - XSS 방지 규칙 업데이트

## 변경됨

- **docs**: README 파일에 Sequential Thinking MCP 추가 (014f08cf)
  - 새로운 MCP 통합을 위한 문서 업데이트
  - 사용 예제 및 패턴 문서화

- **revert**: 불필요한 탐색 제약 시스템 제거 (170c257a)
  - 단순화된 에이전트 오케스트레이션
  - 탐색 작업에서 오버헤드 감소

## 수정됨

- **fix(statusline)**: 그래프 렌더링에서 ANSI 이스케이프 코드 제거 (a98cf1b6)
  - 이스케이프 코드 없는 깨끗한 그래프 표시

- **fix(statusline)**: 컨텍스트 창 계산에 output_tokens 포함 (b0702cdf)
  - 출력 토큰을 포함한 정확한 토큰 사용량 계산

- **fix(statusline)**: 백분율이 제공되지 않을 때 토큰에서 대체 계산 추가 (bdbc98d9)
  - 백분율 데이터를 사용할 수 없을 때 우아한 대체 처리

- **fix(statusline)**: 문자열 대신 백분율을 사용하여 컨텍스트 그래프 표시 (5caec0c3)
  - 백분율 값을 사용한 일관된 그래프 표시

- **fix(statusline)**: 토큰 수 대신 컨텍스트 그래프 항상 표시 (519bd367)
  - 상태 표시줄에서 개선된 시각적 일관성

- **fix(statusline)**: 제공되지 않을 때 토큰에서 백분율 계산 (b57197da)
  - 원시 토큰 수에서 자동 백분율 계산

- **fix(commands)**: 99-release.md 명령 복원 (cd9e3266)
  - 로컬 개발을 위한 릴리스 명령 복원

## 품질

- 스모크 테스트: 6/6 통과 (100% 통과율)
- Ruff: 모든 검사 통과
- Ruff format: 216 파일 변경 없음
- Mypy: 성공 (비임계 코드 3개의 사소한 타입 힌트 문제)

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
