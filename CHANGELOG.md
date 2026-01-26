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

### Memory MCP Integration

- **feat**: Add Memory MCP Server integration (99ab5273)
  - Persistent memory across Claude Code sessions
  - User preferences, project context, learned patterns storage
  - Agent-to-agent context sharing during workflows
  - Configuration: `.mcp.json`, `.mcp.windows.json`
  - New skill: `moai-foundation-memory` (420 lines)

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

# v1.9.0 - Memory MCP, SVG 스킬, Rules 마이그레이션 (2026-01-26)

## 요약

세션 간 지속적인 메모리, 포괄적인 SVG 스킬, 표준 준수 rules 시스템 마이그레이션을 도입한 마이너 릴리즈.

**주요 기능**:
- **Memory MCP 통합**: 사용자 선호도 및 프로젝트 컨텍스트의 지속적인 저장
- **SVG 스킬**: SVGO 최적화 패턴과 모범 사례가 포함된 포괄적인 스킬
- **Rules 마이그레이션**: `.moai/rules/*.yaml`에서 `.claude/rules/*.md`로 마이그레이션 (Claude Code 공식 표준)
- **버그 수정**: Rank batch sync 표시 문제 (#300)

**영향**:
- Memory MCP를 통한 에이전트 간 컨텍스트 공유 활성화
- 전문적인 SVG 생성 및 최적화 지원
- 더 깔끔하고 표준을 준수하는 프로젝트 구조
- 정확한 배치 동기화 통계 표시

## Breaking Changes

없음. 모든 변경 사항은 하위 호환됩니다.

## 추가됨

### Memory MCP 통합

- **feat**: Memory MCP Server 통합 추가 (99ab5273)
  - Claude Code 세션 간 지속적인 메모리
  - 사용자 선호도, 프로젝트 컨텍스트, 학습된 패턴 저장
  - 워크플로우 중 에이전트 간 컨텍스트 공유
  - 설정: `.mcp.json`, `.mcp.windows.json`
  - 새 스킬: `moai-foundation-memory` (420 lines)

### SVG 생성 및 최적화 스킬

- **feat**: `moai-tool-svg` 스킬 추가 (54c12a85)
  - W3C SVG 2.0 명세 및 SVGO 문서 기반
  - 포괄적인 모듈: 기본, 스타일링, 최적화, 애니메이션
  - 12개의 작동하는 코드 예제
  - SVGO 설정 패턴 및 모범 사례
  - 총 3,698 lines (SKILL.md: 410, modules: 2,288, examples: 500, reference: 500)

### 언어 규칙 개선

- **feat**: 향상된 툴링 정보로 언어 규칙 업데이트 (54c12a85)
  - Ruff 설정 패턴 (flake8+isort+pyupgrade 대체)
  - Mypy strict mode 가이드라인
  - 테스팅 프레임워크 권장 사항
  - 16개 언어 파일 업데이트

## 변경됨

### CLAUDE.md 최적화

- **refactor**: v1.9.0을 위한 대규모 정리 및 모듈화 (4134e60d)
  - CLAUDE.md를 ~60k에서 ~30k 문자로 축소 (40k 제한 준수)
  - 상세 내용을 `.claude/rules/`로 이동하여 구성 개선
  - 크로스 플랫폼 호환성을 위한 `shell_validator.py` 유틸리티 추가
  - CLI 명령어 향상 (doctor, init, update)
  - `moai-workflow-thinking` 스킬 추가
  - bug-report.yml 이슈 템플릿 추가
  - 영향: 가독성, 유지보수성, Claude Code 호환성 개선

### Rules 시스템 마이그레이션

- **feat**: `.moai/rules/*.yaml`에서 `.claude/rules/*.md`로 마이그레이션 (99ab5273)
  - 삭제: 6,959 lines의 YAML 규칙
  - 추가: Claude Code 공식 Markdown 규칙
  - 구조: `.claude/rules/{core,development,workflow,languages}/`
  - 영향: 표준 준수, 더 깔끔한 구성

## 수정됨

### Rank 명령어

- **fix(rank)**: batch sync를 위한 중첩된 API 응답 올바르게 파싱 (#300) (31b504ed)
  - 문제: `moai-adk rank sync`가 항상 "Submitted: 0" 표시
  - 근본 원인: 중첩된 `data` 필드 추출 누락
  - 수정: 필드 접근 전 `data = response.get("data", {})` 추가
  - 영향: 정확한 제출 통계 표시

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

# v1.8.13 - Statusline Context Window 수정 (2026-01-26)

## 요약

Statusline context window 계산 정확도 개선 패치 릴리즈.

**주요 수정**:
- Statusline context window 퍼센티지가 Claude Code의 사전 계산 값을 사용하도록 수정

**영향**:
- Context window 표시가 auto-compact와 출력 토큰 예약을 고려
- 더 정확한 남은 토큰 정보 제공

## 수정됨

### Statusline Context Window 계산

- **fix(statusline)**: Claude Code의 사전 계산된 context percentage 사용 (2dacecb7)
  - 우선순위 1: Claude Code의 `used_percentage`/`remaining_percentage` 사용 (가장 정확)
  - 우선순위 2: `current_usage` 토큰으로 계산 (fallback)
  - 우선순위 3: 데이터 없을 때 0% 반환 (세션 시작)
  - Auto-compact 활성화 또는 출력 토큰 예약 시 정확도 보장
  - 파일: `src/moai_adk/statusline/main.py`

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