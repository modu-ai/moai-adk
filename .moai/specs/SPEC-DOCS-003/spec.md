---
# 필수 메타데이터
id: DOCS-003
version: 0.0.1
status: draft
created: 2025-10-17
updated: 2025-10-17
author: @Goos
priority: high

# 선택 메타데이터
category: docs
labels:
  - documentation
  - user-journey
  - mkdocs
  - restructure

# 관계 필드
related_specs:
  - DOCS-001
  - DOCS-002

# 범위 필드
scope:
  packages:
    - docs/
    - mkdocs.yml
  files:
    - docs/index.md
    - docs/introduction.md
    - docs/getting-started.md
    - docs/configuration.md
    - docs/workflow.md
    - docs/commands.md
    - docs/agents.md
    - docs/hooks.md
    - docs/api-reference.md
    - docs/contributing.md
    - docs/security.md
    - docs/troubleshooting.md
---

# SPEC-DOCS-003: MoAI-ADK 문서 체계 전면 개선 - 사용자 여정 기반 재구성

## HISTORY

### v0.0.1 (2025-10-17) - INITIAL
- **STATUS**: draft
- **AUTHOR**: @Goos
- **CHANGES**: SPEC 문서 초안 작성
- **SECTIONS**: EARS 요구사항, 제약사항, 추적성

---

## @SPEC:DOCS-003 Overview

### 목적

MoAI-ADK 문서를 **사용자 여정 기반**으로 전면 재구성하여 다음을 달성:

1. **명확한 스토리라인**: 바이브 코딩의 한계 → MoAI-ADK 해결책 → 설치 → 설정 → 워크플로우 → 상세 기능
2. **체계적 구조**: 11단계 문서 구조로 초급자부터 고급 개발자까지 커버
3. **완벽한 참조**: 모든 코어 모듈의 API 문서, 에이전트 가이드, 문제 해결 가이드 제공
4. **일관성 유지**: README.md 기반 흐름과 MkDocs Material 테마 호환성

### 배경

**현재 상황**:
- docs/ 디렉토리에 26개 파일 존재
- 체계적 구조 부재 (랜덤한 파일명, 불명확한 계층)
- 사용자 여정 고려 부족 (무엇을 먼저 읽어야 할지 불명확)
- checkpoint-policies.md 같은 과도하게 세부적인 문서 포함

**개선 목표**:
- 11단계 사용자 여정 기반 구조
- 기존 26개 → 53개 파일 (8개 유지, 18개 삭제, 45개 신규)
- README.md 스토리라인과 일관성 유지
- MkDocs Material 네비게이션 최적화

---

## EARS Requirements Specification

### Ubiquitous Requirements (기본 기능)

**문서 구조**:
- **REQ-DOCS-003-001**: 시스템은 사용자 여정 기반 11단계 문서 구조를 제공해야 한다
  - @TAG: `@REQ:DOCS-003-STRUCTURE-001`
  - 단계: Introduction → Getting Started → Configuration → Workflow → Commands → Agents → Hooks → API Reference → Contributing → Security → Troubleshooting

- **REQ-DOCS-003-002**: 시스템은 바이브 코딩의 한계와 MoAI-ADK 해결책을 명확히 설명해야 한다
  - @TAG: `@REQ:DOCS-003-INTRO-001`
  - 위치: docs/introduction.md

- **REQ-DOCS-003-003**: 시스템은 모든 코어 모듈의 API 참조를 제공해야 한다
  - @TAG: `@REQ:DOCS-003-API-001`
  - 위치: docs/api-reference/

- **REQ-DOCS-003-004**: 시스템은 개발자 기여 가이드를 제공해야 한다
  - @TAG: `@REQ:DOCS-003-CONTRIB-001`
  - 위치: docs/contributing.md

**MkDocs 호환성**:
- **REQ-DOCS-003-005**: 시스템은 MkDocs Material 테마와 완벽히 호환되어야 한다
  - @TAG: `@REQ:DOCS-003-MKDOCS-001`
  - 기술: mkdocs-material >= 9.0, mkdocstrings[python] >= 0.24

- **REQ-DOCS-003-006**: 시스템은 자동 네비게이션 구조를 생성해야 한다
  - @TAG: `@REQ:DOCS-003-NAV-001`
  - 도구: mkdocs.yml nav 섹션 자동 생성

### Event-driven Requirements (이벤트 기반)

**사용자 상호작용**:
- **REQ-DOCS-003-007**: WHEN 사용자가 처음 방문하면, 시스템은 바이브 코딩의 문제점을 설명해야 한다
  - @TAG: `@REQ:DOCS-003-FIRST-VISIT-001`
  - 페이지: docs/introduction.md
  - 내용: 플랑켄슈타인 코드, 추적성 부재, 품질 일관성 결여

- **REQ-DOCS-003-008**: WHEN 사용자가 시작을 원하면, 시스템은 설치부터 첫 프로젝트까지 단계별 가이드를 제공해야 한다
  - @TAG: `@REQ:DOCS-003-START-001`
  - 페이지: docs/getting-started/installation.md, docs/getting-started/quick-start.md
  - 포함: PyPI 설치, 템플릿 다운로드, 첫 SPEC 작성

- **REQ-DOCS-003-009**: WHEN 개발자가 특정 에이전트를 사용하려면, 시스템은 에이전트별 상세 가이드를 제공해야 한다
  - @TAG: `@REQ:DOCS-003-AGENT-001`
  - 페이지: docs/agents/spec-builder.md ~ docs/agents/project-manager.md (9개 파일)
  - 포함: 페르소나, 전문 영역, 호출 방법, 예제

- **REQ-DOCS-003-010**: WHEN 사용자가 에러를 만나면, 시스템은 문제 해결 절차를 제공해야 한다
  - @TAG: `@REQ:DOCS-003-ERROR-001`
  - 페이지: docs/troubleshooting/common-errors.md, docs/troubleshooting/faq.md
  - 포함: 에러별 진단 절차, 해결 방법, FAQ

### State-driven Requirements (상태 기반)

**문서 읽기 경험**:
- **REQ-DOCS-003-011**: WHILE 사용자가 문서를 읽을 때, 시스템은 자연스러운 스토리 흐름을 유지해야 한다
  - @TAG: `@REQ:DOCS-003-FLOW-001`
  - 흐름: 문제 인식 → 해결책 이해 → 설치 → 설정 → 실전 사용 → 고급 기능

- **REQ-DOCS-003-012**: WHILE 개발자가 API를 사용할 때, 시스템은 클래스/함수 시그니처 및 예제를 표시해야 한다
  - @TAG: `@REQ:DOCS-003-API-USAGE-001`
  - 도구: mkdocstrings, Python docstring 자동 파싱
  - 포함: 파라미터, 반환값, 예외, 사용 예제

**컨텍스트 인식**:
- **REQ-DOCS-003-013**: WHILE Personal 모드 설명 시, 시스템은 로컬 워크플로우에 집중해야 한다
  - @TAG: `@REQ:DOCS-003-PERSONAL-001`
  - 차이점: 로컬 브랜치, 수동 커밋, PR 생성 안 함

- **REQ-DOCS-003-014**: WHILE Team 모드 설명 시, 시스템은 GitFlow 자동화에 집중해야 한다
  - @TAG: `@REQ:DOCS-003-TEAM-001`
  - 차이점: Draft PR, 자동 브랜치 생성, 자동 머지

### Optional Features (선택적 기능)

- **REQ-DOCS-003-015**: WHERE 고급 사용자이면, 시스템은 Hook 커스터마이징 가이드를 제공할 수 있다
  - @TAG: `@REQ:DOCS-003-HOOK-001`
  - 페이지: docs/hooks/custom-hooks.md
  - 포함: SessionStartHook, PreToolUseHook, PostToolUseHook 상속 방법

- **REQ-DOCS-003-016**: WHERE 기여자이면, 시스템은 개발 환경 설정 가이드를 제공할 수 있다
  - @TAG: `@REQ:DOCS-003-DEV-ENV-001`
  - 페이지: docs/contributing/development-setup.md
  - 포함: Poetry 설정, pre-commit 훅, 테스트 실행

### Constraints (제약사항)

**일관성 제약**:
- **CONS-DOCS-003-001**: 문서 구조는 README.md의 흐름과 일관성을 유지해야 한다
  - @TAG: `@CONS:DOCS-003-README-001`
  - 검증: README.md 섹션 순서와 비교

- **CONS-DOCS-003-002**: ❌ checkpoint-policies.md 같은 과도하게 세부적인 문서는 제외해야 한다
  - @TAG: `@CONS:DOCS-003-DETAIL-001`
  - 삭제 대상: checkpoint-policies.md, advanced-optimization.md

**기술 제약**:
- **CONS-DOCS-003-003**: 모든 문서는 MkDocs Material 테마와 호환되어야 한다
  - @TAG: `@CONS:DOCS-003-THEME-001`
  - 호환성: Admonitions, Code blocks, Tabs, Tables

- **CONS-DOCS-003-004**: API 문서는 Python docstring 표준을 따라야 한다
  - @TAG: `@CONS:DOCS-003-DOCSTRING-001`
  - 표준: Google/NumPy docstring 스타일

**범위 제약**:
- **CONS-DOCS-003-005**: 기존 26개 문서 중 8개만 유지하고 나머지는 재구성해야 한다
  - @TAG: `@CONS:DOCS-003-SCOPE-001`
  - 유지: workflow.md, trust-principles.md, tag-system.md, git-strategy.md, context-engineering.md, agent-design.md, template-processor.md, hooks.md
  - 삭제: 18개 (checkpoint-policies.md 포함)

**품질 제약**:
- **CONS-DOCS-003-006**: 모든 코드 예제는 실제 실행 가능한 코드여야 한다
  - @TAG: `@CONS:DOCS-003-CODE-001`
  - 검증: pytest doctest 또는 수동 검증

---

## Technical Specifications

### 11단계 문서 구조

#### 1. Introduction (docs/introduction.md)
- **내용**: 바이브 코딩의 한계, 플랑켄슈타인 코드, MoAI-ADK 해결책
- **TAG**: `@DOC:INTRO-001`

#### 2. Getting Started (docs/getting-started/)
- `installation.md`: PyPI 설치, 템플릿 다운로드
- `quick-start.md`: 첫 프로젝트 초기화, 첫 SPEC 작성
- `first-project.md`: 실전 예제 (TODO 앱)
- **TAG**: `@DOC:START-001`

#### 3. Configuration (docs/configuration/)
- `config-json.md`: .moai/config.json 구조 설명
- `personal-vs-team.md`: Personal/Team 모드 차이
- `advanced-settings.md`: 고급 설정 (TRUST 임계값, TAG 전략)
- **TAG**: `@DOC:CONFIG-001`

#### 4. Workflow (docs/workflow.md)
- **내용**: 3단계 워크플로우 (`/alfred:1-spec` → `/alfred:2-build` → `/alfred:3-sync`)
- **유지**: 기존 workflow.md
- **TAG**: `@DOC:WORKFLOW-001`

#### 5. Commands (docs/commands/)
- `cli-reference.md`: MoAI-ADK CLI 명령어
- `alfred-commands.md`: Alfred SuperAgent 명령어 (/alfred:0~3)
- `agent-commands.md`: 개별 에이전트 호출 (@agent-*)
- **TAG**: `@DOC:CMD-001`

#### 6. Agents (docs/agents/)
- 9개 에이전트별 가이드:
  - `spec-builder.md`
  - `code-builder.md`
  - `doc-syncer.md`
  - `tag-agent.md`
  - `git-manager.md`
  - `debug-helper.md`
  - `trust-checker.md`
  - `cc-manager.md`
  - `project-manager.md`
- **TAG**: `@DOC:AGENT-001`

#### 7. Hooks (docs/hooks/)
- `overview.md`: Hook 시스템 개요
- `session-start-hook.md`: SessionStartHook
- `pre-tool-use-hook.md`: PreToolUseHook
- `post-tool-use-hook.md`: PostToolUseHook
- `custom-hooks.md`: 커스텀 Hook 작성
- **TAG**: `@DOC:HOOK-001`

#### 8. API Reference (docs/api-reference/)
- `core-installer.md`: moai_adk.core.installer
- `core-git.md`: moai_adk.core.git_strategy
- `core-tag.md`: moai_adk.core.tag_system
- `core-template.md`: moai_adk.core.template_processor
- `agents.md`: moai_adk.agents.*
- **TAG**: `@DOC:API-001`

#### 9. Contributing (docs/contributing/)
- `overview.md`: 기여 가이드 개요
- `development-setup.md`: 개발 환경 설정
- `code-style.md`: 코딩 스타일, 린트 규칙
- `testing.md`: 테스트 작성 가이드
- `pull-request-process.md`: PR 프로세스
- **TAG**: `@DOC:CONTRIB-001`

#### 10. Security (docs/security/)
- `overview.md`: 보안 개요
- `template-security.md`: 템플릿 보안 검증
- `best-practices.md`: 보안 모범 사례
- `checklist.md`: 보안 체크리스트
- **TAG**: `@DOC:SEC-001`

#### 11. Troubleshooting (docs/troubleshooting/)
- `common-errors.md`: 자주 발생하는 에러
- `debugging-guide.md`: 디버깅 가이드
- `faq.md`: FAQ
- **TAG**: `@DOC:TROUBLESHOOT-001`

### 기술 스택

**MkDocs 설정**:
```yaml
# mkdocs.yml
site_name: MoAI-ADK Documentation
theme:
  name: material
  features:
    - navigation.tabs
    - navigation.sections
    - navigation.expand
    - toc.integrate
    - search.suggest
    - search.highlight
    - content.code.copy

plugins:
  - search
  - mkdocstrings:
      handlers:
        python:
          options:
            show_source: true
            show_root_heading: true
            show_signature_annotations: true

markdown_extensions:
  - admonition
  - pymdownx.details
  - pymdownx.superfences
  - pymdownx.tabbed:
      alternate_style: true
  - pymdownx.highlight:
      anchor_linenums: true
  - pymdownx.inlinehilite
  - pymdownx.snippets
  - tables
```

**Python 패키지**:
- `mkdocs>=1.5.0`
- `mkdocs-material>=9.0.0`
- `mkdocstrings[python]>=0.24.0`
- `pymdown-extensions>=10.0`

---

## Traceability

### TAG Chain

```
@SPEC:DOCS-003
  ├─ @REQ:DOCS-003-STRUCTURE-001 → 11단계 문서 구조
  ├─ @REQ:DOCS-003-INTRO-001 → 바이브 코딩 한계 설명
  ├─ @REQ:DOCS-003-API-001 → API 참조 문서
  ├─ @REQ:DOCS-003-CONTRIB-001 → 기여 가이드
  ├─ @REQ:DOCS-003-MKDOCS-001 → MkDocs 호환성
  ├─ @REQ:DOCS-003-NAV-001 → 자동 네비게이션
  ├─ @REQ:DOCS-003-FIRST-VISIT-001 → 첫 방문 경험
  ├─ @REQ:DOCS-003-START-001 → 빠른 시작
  ├─ @REQ:DOCS-003-AGENT-001 → 에이전트 가이드
  ├─ @REQ:DOCS-003-ERROR-001 → 에러 해결
  ├─ @REQ:DOCS-003-FLOW-001 → 스토리 흐름
  ├─ @REQ:DOCS-003-API-USAGE-001 → API 사용 예제
  ├─ @REQ:DOCS-003-PERSONAL-001 → Personal 모드
  ├─ @REQ:DOCS-003-TEAM-001 → Team 모드
  ├─ @REQ:DOCS-003-HOOK-001 → Hook 커스터마이징
  └─ @REQ:DOCS-003-DEV-ENV-001 → 개발 환경 설정
```

### 관련 문서

- **README.md**: 메인 프로젝트 소개 (스토리라인 일관성 유지)
- **.moai/project/product.md**: 제품 정의 (미션, 사용자, 문제, 전략)
- **.moai/memory/development-guide.md**: 개발 가이드 (TRUST 원칙, TAG 시스템)
- **SPEC-DOCS-001**: 기존 문서 SPEC (중복 방지)
- **SPEC-DOCS-002**: 기존 문서 SPEC (중복 방지)

### 의존성

- **depends_on**: 없음 (독립적 문서 작업)
- **blocks**: 없음
- **related_specs**: DOCS-001, DOCS-002

---

## Definition of Done

### 문서 작성 완료 기준

- [ ] 11단계 문서 구조 모두 작성 (53개 파일)
- [ ] 모든 코어 모듈 API 문서 작성 (mkdocstrings 활용)
- [ ] 9개 에이전트 가이드 작성
- [ ] MkDocs 네비게이션 구조 완성
- [ ] 모든 코드 예제 검증 완료

### 품질 게이트

- [ ] MkDocs 빌드 성공 (`mkdocs build --strict`)
- [ ] 링크 검증 통과 (깨진 링크 0개)
- [ ] 모든 API 문서 자동 생성 성공
- [ ] README.md 스토리라인과 일관성 확인
- [ ] ❌ checkpoint-policies.md 같은 과도한 세부 문서 제거 확인

### 검증 방법

```bash
# MkDocs 빌드 테스트
mkdocs build --strict

# 링크 검증
pytest tests/test_docs_links.py

# API 문서 생성 확인
ls docs/api-reference/*.md | wc -l  # 5개 이상

# 스토리라인 일관성 검증 (수동)
diff <(grep '^## ' README.md) <(grep '^## ' docs/introduction.md)
```

---

**작성일**: 2025-10-17
**작성자**: @Goos
**버전**: v0.0.1 (INITIAL)
