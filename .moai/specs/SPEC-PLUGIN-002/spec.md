---
# 필수 필드 (7개)
id: PLUGIN-002
version: 0.0.1
status: draft
created: 2025-10-10
updated: 2025-10-10
author: @Goos
priority: high

# 선택 필드 - 분류/메타
category: feature
labels:
  - claude-code
  - initialization
  - template
  - mustache

# 선택 필드 - 관계
depends_on:
  - PLUGIN-001

# 선택 필드 - 범위
scope:
  packages:
    - .claude/commands/alfred/
    - scripts/
  files:
    - 8-project.md
    - render-template.cjs
---

# @SPEC:PLUGIN-002: /alfred:8-project 커맨드 강화 (init 대체)

## HISTORY

### v0.0.1 (2025-10-10)
- **INITIAL**: /alfred:8-project 커맨드 강화 SPEC 작성
- **AUTHOR**: @Goos
- **SCOPE**: 프로젝트 초기화, 템플릿 복사, Mustache 렌더링, Git 초기화
- **CONTEXT**: npm 패키지 제거로 인한 moai init CLI 대체 기능 구현

---

## 📋 개요

npm 패키지 제거로 인해 `moai init` CLI가 사라졌으므로, `/alfred:8-project` 커맨드를 강화하여 프로젝트 초기화 기능을 완전히 대체합니다. 템플릿 복사, Mustache 렌더링, Git 초기화를 포함한 완전한 초기화 워크플로우를 제공합니다.

## 🎯 목표

1. **템플릿 복사**: `${CLAUDE_PLUGIN_ROOT}/templates/.moai` → `.moai` 디렉토리 복사
2. **Mustache 렌더링**: config.json.mustache, CLAUDE.md.mustache 자동 렌더링
3. **Git 초기화**: 선택적 Git 초기화 및 .gitignore 생성
4. **헬퍼 스크립트**: scripts/render-template.cjs (Mustache 렌더링 전담)
5. **변수 수집**: 사용자 입력을 통한 PROJECT_NAME, PROJECT_MODE, PROJECT_DESCRIPTION, LOCALE 수집

## 📝 EARS 요구사항

### Ubiquitous Requirements (기본 기능)

1. **프로젝트 초기화 기능**
   - 시스템은 `/alfred:8-project` 커맨드를 통해 완전한 프로젝트 초기화를 제공해야 한다
   - 시스템은 `${CLAUDE_PLUGIN_ROOT}/templates/.moai` 디렉토리의 모든 파일을 복사해야 한다
   - 시스템은 `.mustache` 확장자 파일을 자동 렌더링해야 한다

2. **템플릿 복사 규칙**
   - 시스템은 `.moai/` 디렉토리가 이미 존재하는지 확인해야 한다
   - 시스템은 product.md, structure.md, tech.md를 복사하거나 생성해야 한다
   - 시스템은 config.json.mustache → config.json 렌더링을 수행해야 한다

3. **Mustache 렌더링**
   - 시스템은 `scripts/render-template.cjs` Node.js 스크립트를 제공해야 한다
   - 시스템은 `mustache` npm 패키지를 사용하여 템플릿을 렌더링해야 한다
   - 시스템은 렌더링된 파일에서 `.mustache` 확장자를 제거해야 한다

4. **변수 매핑**
   - 시스템은 다음 변수를 지원해야 한다:
     - `{{PROJECT_NAME}}`: 프로젝트 이름
     - `{{PROJECT_MODE}}`: personal 또는 team
     - `{{PROJECT_DESCRIPTION}}`: 프로젝트 설명
     - `{{LOCALE}}`: 언어 설정 (ko, en, ja, zh)

### Event-driven Requirements (이벤트 기반)

1. **커맨드 실행 흐름**
   - WHEN 사용자가 `/alfred:8-project`를 실행하면, 시스템은 Phase 1 분석을 시작해야 한다
   - WHEN `.moai/` 디렉토리가 없으면, 시스템은 신규 초기화 모드로 진행해야 한다
   - WHEN `.moai/` 디렉토리가 존재하면, 시스템은 갱신 모드로 전환해야 한다

2. **템플릿 렌더링 트리거**
   - WHEN `.mustache` 파일이 발견되면, 시스템은 render-template.cjs를 실행해야 한다
   - WHEN 렌더링이 완료되면, 시스템은 원본 `.mustache` 파일을 삭제할 수 있다
   - WHEN 렌더링이 실패하면, 시스템은 에러 메시지를 표시하고 중단해야 한다

3. **Git 초기화 프롬프트**
   - WHEN 초기화가 완료되면, 시스템은 Git 초기화 여부를 사용자에게 확인해야 한다
   - WHEN 사용자가 "예"를 선택하면, 시스템은 `git init` 및 `.gitignore` 생성을 수행해야 한다
   - WHEN 사용자가 "아니오"를 선택하면, 시스템은 Git 초기화를 건너뛰어야 한다

### State-driven Requirements (상태 기반)

1. **초기화 진행 중**
   - WHILE 초기화가 진행 중일 때, 시스템은 진행 상태를 표시해야 한다
   - WHILE 파일 복사 중일 때, 시스템은 복사된 파일 목록을 출력해야 한다
   - WHILE 렌더링 중일 때, 시스템은 렌더링 대상 파일을 표시해야 한다

2. **사용자 입력 대기**
   - WHILE 사용자 입력을 대기할 때, 시스템은 기본값을 제안해야 한다
   - WHILE 변수 수집 중일 때, 시스템은 각 변수의 설명을 제공해야 한다

3. **에러 상태**
   - WHILE 에러가 발생한 상태일 때, 시스템은 상세한 에러 메시지와 해결 방법을 표시해야 한다
   - WHILE 권한 에러일 때, 시스템은 필요한 권한 및 명령어를 안내해야 한다

### Optional Features (선택적 기능)

1. **고급 설정**
   - WHERE 사용자가 고급 설정을 요청하면, 시스템은 추가 변수 입력을 제공할 수 있다
   - WHERE `.mcp.json`이 템플릿에 존재하면, 시스템은 MCP 서버 설정을 복사할 수 있다

2. **Git 초기화**
   - WHERE Git이 설치되어 있으면, 시스템은 Git 초기화 옵션을 제공할 수 있다
   - WHERE .gitignore 템플릿이 존재하면, 시스템은 해당 템플릿을 사용할 수 있다

3. **기존 파일 보존**
   - WHERE 기존 product.md가 존재하면, 시스템은 백업 후 보존할 수 있다
   - WHERE 기존 config.json이 존재하면, 시스템은 사용자에게 덮어쓰기 여부를 확인할 수 있다

### Constraints (제약사항)

1. **경로 검증**
   - IF `${CLAUDE_PLUGIN_ROOT}`가 설정되지 않았으면, 시스템은 에러를 표시하고 중단해야 한다
   - IF `.moai/` 디렉토리 생성 권한이 없으면, 시스템은 권한 에러를 표시해야 한다

2. **필수 변수 검증**
   - IF PROJECT_NAME이 비어있으면, 시스템은 기본값(현재 디렉토리명)을 사용해야 한다
   - IF PROJECT_MODE가 "personal" 또는 "team"이 아니면, 시스템은 에러를 표시해야 한다
   - IF LOCALE이 지원되지 않는 값이면, 시스템은 기본값 "ko"를 사용해야 한다

3. **템플릿 무결성**
   - IF config.json.mustache가 없으면, 시스템은 경고를 표시하고 기본 config.json을 생성해야 한다
   - IF 필수 템플릿 파일이 누락되면, 시스템은 초기화를 중단해야 한다

4. **스크립트 실행 제약**
   - IF Node.js가 설치되지 않았으면, 시스템은 Mustache 렌더링을 스킵하고 경고를 표시해야 한다
   - IF render-template.cjs 실행이 실패하면, 시스템은 수동 렌더링 가이드를 제공해야 한다

5. **특수문자 제약**
   - IF PROJECT_NAME에 공백 또는 특수문자가 포함되면, 시스템은 kebab-case로 정규화해야 한다
   - IF PROJECT_DESCRIPTION에 개행 문자가 포함되면, 시스템은 공백으로 치환해야 한다

## 🔗 추적성 (Traceability)

- **@SPEC:PLUGIN-002** → 이 문서
- **@TEST:PLUGIN-002** → `tests/commands/project-init.test.ts` (예정)
- **@CODE:PLUGIN-002** → `.claude/commands/alfred/8-project.md` (기존, 강화)
- **@CODE:PLUGIN-002:SCRIPT** → `scripts/render-template.cjs` (예정)
- **@DOC:PLUGIN-002** → `docs/project-initialization.md` (예정)

---

## 참조

- **관련 SPEC**: @SPEC:PLUGIN-001 (Claude Code 플러그인 구조 설계)
- **기존 커맨드**: `.claude/commands/alfred/8-project.md`
- **템플릿 디렉토리**: `${CLAUDE_PLUGIN_ROOT}/templates/.moai/`
- **Mustache 라이브러리**: https://www.npmjs.com/package/mustache
