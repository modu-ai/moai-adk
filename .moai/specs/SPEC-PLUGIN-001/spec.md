---
# 필수 필드 (7개)
id: PLUGIN-001
version: 0.0.1
status: draft
created: 2025-10-10
updated: 2025-10-10
author: @Goos
priority: high

# 선택 필드 - 분류/메타
category: refactor
labels:
  - claude-code
  - plugin
  - architecture

# 선택 필드 - 관계
depends_on:
  - STRUCTURE-001

# 선택 필드 - 범위
scope:
  packages:
    - .claude-plugin/
    - moai-adk-ts/templates/.claude/
  files:
    - plugin.json
    - hooks.json
---

# @SPEC:PLUGIN-001: Claude Code 플러그인 구조 설계

## HISTORY

### v0.0.1 (2025-10-10)
- **INITIAL**: Claude Code 플러그인 구조 설계 SPEC 작성
- **AUTHOR**: @Goos
- **SCOPE**: 플러그인 디렉토리 구조, 매니페스트 설계, hooks 변환
- **CONTEXT**: MoAI-ADK를 npm 패키지에서 Claude Code 플러그인 전용으로 전환

---

## 📋 개요

MoAI-ADK를 npm 패키지 기반에서 Claude Code 플러그인 전용 프로젝트로 전환합니다. 기존 CLI 기능을 플러그인 커맨드로 대체하고, 표준 플러그인 디렉토리 구조 및 매니페스트를 설계합니다.

## 🎯 목표

1. Claude Code 플러그인 표준 구조 정의
2. `.claude-plugin/plugin.json` 매니페스트 설계
3. `hooks/hooks.json` 변환 (settings.json → hooks.json)
4. `${CLAUDE_PLUGIN_ROOT}` 환경변수 기반 경로 체계 확립
5. 기존 템플릿과의 호환성 유지

## 📝 EARS 요구사항

### Ubiquitous Requirements (기본 기능)

1. **플러그인 구조 표준화**
   - 시스템은 `.claude-plugin/` 디렉토리를 플러그인 루트로 사용해야 한다
   - 시스템은 `plugin.json` 매니페스트를 필수로 제공해야 한다
   - 시스템은 `commands/`, `agents/`, `hooks/`, `templates/` 디렉토리 구조를 따라야 한다

2. **매니페스트 필드**
   - 시스템은 `name`, `version` 필수 필드를 포함해야 한다
   - 시스템은 `description`, `author`, `homepage`, `license` 메타데이터를 제공해야 한다
   - 시스템은 `commands`, `agents`, `hooks` 경로를 명시해야 한다

3. **환경변수 사용**
   - 시스템은 `${CLAUDE_PLUGIN_ROOT}` 변수를 모든 내부 경로 참조에 사용해야 한다
   - 시스템은 상대 경로는 `./`로 시작하도록 강제해야 한다

### Event-driven Requirements (이벤트 기반)

1. **플러그인 활성화**
   - WHEN 플러그인이 설치되면, 시스템은 `plugin.json`을 파싱해야 한다
   - WHEN 플러그인이 활성화되면, 시스템은 `commands/`, `agents/` 디렉토리를 로드해야 한다

2. **후크 실행**
   - WHEN Write 또는 Edit 도구가 실행되면, 시스템은 PostToolUse 후크를 트리거해야 한다
   - WHEN 후크 실행이 실패하면, 시스템은 에러 로그를 출력하고 작업을 계속해야 한다

### State-driven Requirements (상태 기반)

1. **개발 모드**
   - WHILE 개발 모드일 때, 시스템은 상세한 플러그인 로딩 로그를 출력해야 한다
   - WHILE 디버그 모드일 때, 시스템은 `${CLAUDE_PLUGIN_ROOT}` 경로를 표시해야 한다

2. **활성화 상태**
   - WHILE 플러그인이 활성화되어 있을 때, 시스템은 `/alfred:*` 커맨드를 제공해야 한다
   - WHILE 플러그인이 비활성화되어 있을 때, 시스템은 기본 Claude Code 기능만 제공해야 한다

### Optional Features (선택적 기능)

1. **MCP 서버 통합**
   - WHERE MCP 서버가 정의되어 있으면, 시스템은 `.mcp.json`을 로드할 수 있다

2. **스크립트 자동화**
   - WHERE `scripts/` 디렉토리가 존재하면, 시스템은 후크에서 스크립트를 실행할 수 있다

### Constraints (제약사항)

1. **경로 규칙**
   - IF 내부 파일 참조 시, 절대 경로는 `${CLAUDE_PLUGIN_ROOT}`를 사용해야 한다
   - IF 외부 파일 참조 시, 상대 경로는 `./`로 시작해야 한다

2. **필수 파일**
   - IF `plugin.json`이 없으면, 시스템은 플러그인 로드를 실패해야 한다
   - IF `commands/` 또는 `agents/` 디렉토리가 비어있으면, 경고를 표시해야 한다

3. **버전 호환성**
   - 플러그인 `version`은 Semantic Versioning을 따라야 한다
   - `name` 필드는 소문자, 하이픈만 사용해야 한다

## 🔗 추적성 (Traceability)

- **@SPEC:PLUGIN-001** → 이 문서
- **@TEST:PLUGIN-001** → `tests/plugin/structure.test.ts` (예정)
- **@CODE:PLUGIN-001** → `.claude-plugin/plugin.json` (예정)
- **@DOC:PLUGIN-001** → `docs/plugin-structure.md` (예정)

---

## 참조

- **Claude Code 플러그인 참조**: https://docs.claude.com/en/docs/claude-code/plugins-reference
- **관련 SPEC**: @SPEC:STRUCTURE-001
- **기존 템플릿**: `moai-adk-ts/templates/.claude/`
