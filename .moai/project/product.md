---
id: PRODUCT-001
version: 0.1.2
status: active
created: 2025-10-01
updated: 2025-10-17
author: @Goos
priority: high
---

# MoAI-ADK Product Definition

## HISTORY

### v0.1.2 (2025-10-17)
- **UPDATED**: 에이전트 수 갱신 (9개 → 11개, v0.3.4 반영)
- **AUTHOR**: @Alfred
- **SECTIONS**: Mission (Alfred SuperAgent 팀 구성 업데이트)

### v0.1.1 (2025-10-17)
- **UPDATED**: 템플릿 기본값을 실제 MoAI-ADK 프로젝트 내용으로 갱신
- **AUTHOR**: @Alfred
- **SECTIONS**: Mission, User, Problem, Strategy, Success 실제 내용 반영

### v0.1.0 (2025-10-01)
- **INITIAL**: 프로젝트 제품 정의 문서 작성
- **AUTHOR**: @project-owner
- **SECTIONS**: Mission, User, Problem, Strategy, Success, Legacy

---

## @DOC:MISSION-001 핵심 미션

### 핵심 가치 제안

> **"SPEC이 없으면 CODE도 없다."**

MoAI-ADK는 **SPEC-First TDD 방법론**을 통해 플랑켄슈타인 코드를 근본적으로 방지하는 AI 에이전틱 개발 프레임워크입니다.

#### 4가지 핵심 가치

1. **일관성 (Consistency)**: SPEC → TDD → Sync 3단계 파이프라인으로 개발 품질 보장
2. **품질 (Quality)**: TRUST 5원칙 (Test First, Readable, Unified, Secured, Trackable) 자동 적용
3. **추적성 (Traceability)**: @TAG 시스템 (`@SPEC → @TEST → @CODE → @DOC`)으로 완벽한 이력 추적
4. **범용성 (Universality)**: Python, TypeScript, Java, Go, Rust 등 17개 주요 언어 지원

#### Alfred SuperAgent

**Alfred**는 12개 AI 에이전트 팀 (Alfred + 11개 전문 에이전트)을 조율하는 중앙 오케스트레이터입니다:
- **spec-builder** 🏗️: SPEC 작성 (EARS 방식)
- **implementation-planner** 📋: SPEC 분석 및 구현 전략 수립
- **tdd-implementer** 🔬: TDD RED-GREEN-REFACTOR 전문 구현
- **quality-gate** 🛡️: TRUST 원칙 통합 검증
- **doc-syncer** 📖: 문서 동기화 (Living Document)
- **tag-agent** 🏷️: TAG 시스템 관리
- **git-manager** 🚀: Git 워크플로우 자동화
- **debug-helper** 🔬: 런타임 오류 진단
- **trust-checker** ✅: TRUST 원칙 검증
- **cc-manager** 🛠️: Claude Code 설정 관리
- **project-manager** 📋: 프로젝트 초기화

## @SPEC:USER-001 주요 사용자층

### 1차 사용자: 실무 개발자

- **대상**: AI 도구를 활용하여 개발 생산성을 높이고 싶은 개인 개발자 및 팀
- **핵심 니즈**:
  - 일관성 있는 코드 품질 유지
  - 반복적인 TDD 사이클 자동화
  - 명세-코드-문서 간 동기화 자동화
  - Git 워크플로우 자동화 (브랜치, PR, 머지)
- **핵심 시나리오**:
  1. **신규 기능 개발**: `/alfred:1-spec` → `/alfred:2-build` → `/alfred:3-sync` 워크플로우
  2. **레거시 코드 리팩토링**: SPEC 기반 점진적 개선 및 TAG 추적
  3. **팀 협업**: GitFlow 통합, Draft PR 자동 생성, Living Document 동기화

### 2차 사용자: 팀 리더 및 아키텍트

- **대상**: 팀의 코드 품질과 개발 표준을 관리하는 리더십 역할
- **핵심 니즈**:
  - 팀 전체 코드 품질 표준 유지
  - SPEC 기반 요구사항 관리
  - TAG 시스템으로 전체 프로젝트 추적성 확보
  - TRUST 5원칙 준수 여부 자동 검증
- **핵심 시나리오**:
  1. **프로젝트 초기화**: `/alfred:0-project`로 product/structure/tech 문서 구축
  2. **품질 게이트 설정**: TRUST 원칙 기반 자동 검증 체계 구축
  3. **코드 리뷰 효율화**: @TAG 체인으로 변경 영향 범위 빠른 파악

## @SPEC:PROBLEM-001 해결하는 핵심 문제

### 우선순위 높음

1. **플랑켄슈타인 코드 (Frankenstein Code)**
   - **문제**: AI가 생성한 코드가 맥락 없이 조합되어 유지보수 불가능한 코드 생성
   - **해결**: SPEC-First 원칙 - 명세 없이는 코드 생성 불가, @TAG 시스템으로 전체 이력 추적

2. **추적성 부재 (Lost Traceability)**
   - **문제**: 코드 변경 이유와 설계 의도를 파악할 수 없어 리팩토링 시 위험 증가
   - **해결**: CODE-FIRST @TAG 시스템 - 코드 직접 스캔 방식으로 `@SPEC → @TEST → @CODE → @DOC` 체인 자동 검증

3. **품질 일관성 결여 (Quality Inconsistency)**
   - **문제**: 개발자/팀마다 다른 품질 기준으로 코드베이스 품질 편차 발생
   - **해결**: TRUST 5원칙 자동 적용 - Test First, Readable, Unified, Secured, Trackable

### 우선순위 중간

- **반복적 TDD 사이클**: RED → GREEN → REFACTOR 수동 반복의 피로도
- **Git 워크플로우 복잡도**: 브랜치 생성, PR 작성, 머지 작업의 반복
- **문서 동기화 누락**: 코드 변경 후 문서 업데이트 누락으로 인한 불일치

### 현재 실패 사례들

- **수동 TDD**: 테스트 작성 → 구현 → 리팩토링 사이클을 개발자가 직접 관리하여 일관성 저하
- **수동 TAG 관리**: 주석 기반 TAG를 수동 관리하여 동기화 누락 및 고아 TAG 발생
- **수동 Git 작업**: 브랜치명 규칙, 커밋 메시지 형식, PR 템플릿 등을 수동 관리하여 실수 발생

## @DOC:STRATEGY-001 차별점 및 강점

### 경쟁 솔루션 대비 강점

#### 1. Alfred SuperAgent - 10개 AI 에이전트 팀

- **차별점**: 단일 AI가 아닌 전문 에이전트 팀이 역할별 최적화된 작업 수행
- **발휘 시나리오**:
  - spec-builder는 EARS 방식 명세 작성에 특화
  - code-builder는 TDD 구현에 특화
  - doc-syncer는 Living Document 생성에 특화
  - 각 에이전트가 IT 전문가 직무에 매핑되어 전문성 극대화

#### 2. CODE-FIRST @TAG 시스템

- **차별점**: 중간 캐시 없이 코드를 직접 스캔하여 TAG 무결성 보장
- **발휘 시나리오**:
  - `/alfred:3-sync` 실행 시 `rg '@(SPEC|TEST|CODE|DOC):' -n`으로 전체 코드베이스 스캔
  - 고아 TAG, 끊어진 링크 자동 탐지
  - TAG의 진실은 코드 자체에만 존재 (단일 진실 원천)

#### 3. TRUST 5원칙 자동 적용

- **차별점**: 수동 품질 관리가 아닌 자동 검증 시스템
- **발휘 시나리오**:
  - Test First: pytest, Vitest 등 언어별 테스트 프레임워크 자동 적용
  - Readable: ruff, ESLint 등 언어별 린터 자동 실행
  - Unified: 복잡도 임계값 (함수 ≤50 LOC, 파일 ≤300 LOC) 자동 검증
  - Secured: bandit, pip-audit 등 보안 도구 자동 실행
  - Trackable: @TAG 체인 무결성 자동 검증

#### 4. Context Engineering 전략

- **차별점**: JIT Retrieval로 효율적인 컨텍스트 관리
- **발휘 시나리오**:
  - JIT: 필요한 순간에만 문서 로드 (초기 컨텍스트 최소화)
  - Haiku/Sonnet 전략적 배치로 응답 속도 2~5배 향상 + 비용 67% 절감

#### 5. 완전 자동화된 GitFlow 워크플로우

- **차별점**: 브랜치 생성부터 PR 자동 머지까지 완전 자동화
- **발휘 시나리오**:
  - `/alfred:1-spec`: develop에서 feature/SPEC-{ID} 브랜치 생성 + Draft PR 생성
  - `/alfred:2-build`: TDD 구현 + 단계별 커밋 (RED/GREEN/REFACTOR)
  - `/alfred:3-sync --auto-merge`: 문서 동기화 + PR Ready 전환 + CI/CD 확인 + 자동 머지

## @SPEC:SUCCESS-001 성공 지표

### 즉시 측정 가능한 KPI

#### 1. 테스트 커버리지
- **베이스라인**: 85% 이상
- **측정 방법**: pytest-cov (Python), Vitest (TypeScript) 등 언어별 도구
- **실패 시 대응**: `/alfred:2-build` 실행 시 커버리지 미달 경고 + 추가 테스트 케이스 제안

#### 2. TAG 체인 무결성
- **베이스라인**: 고아 TAG 0개, 끊어진 링크 0개
- **측정 방법**: `rg '@(SPEC|TEST|CODE|DOC):' -n` 코드 스캔
- **실패 시 대응**: `/alfred:3-sync` 실행 시 자동 탐지 + 수정 권장

#### 3. 코드 품질 (TRUST 5원칙 준수)
- **베이스라인**: 5가지 원칙 모두 100% 준수
- **측정 방법**: `@agent-trust-checker` 자동 검증
- **실패 시 대응**: 위반 항목별 구체적 수정 권장

#### 4. 개발 사이클 시간 단축
- **베이스라인**: SPEC 작성 → TDD 구현 → 문서 동기화 전체 사이클 50% 단축
- **측정 방법**: 수동 워크플로우 대비 시간 비교
- **실패 시 대응**: 병목 단계 분석 + 에이전트 최적화

### 측정 주기

- **실시간**: TAG 체인 무결성, 테스트 커버리지 (커밋 시점마다)
- **일간**: 코드 품질 게이트 통과율, 자동화된 PR 수
- **주간**: 전체 프로젝트 TRUST 원칙 준수도, 고아 TAG 발생 빈도
- **월간**: 개발 사이클 시간 단축률, 팀 생산성 지표

## Legacy Context

### 기존 자산 요약

#### 모두의AI 연구실 축적 경험

- **서적 집필**: "(가칭) 에이전틱 코딩" 별책 부록으로 MoAI-ADK 개발
- **AI 협업 설계**: GPT-5 Pro + Claude 4.1 Opus 공동 아키텍처 설계
- **100% AI 작성**: 모든 코드가 10개 AI 에이전트 팀에 의해 작성됨

#### 오픈소스 커뮤니티 피드백

- **PyPI 배포**: https://pypi.org/project/moai-adk/
- **GitHub 저장소**: https://github.com/modu-ai/moai-adk
- **커버리지 87.66%**: codecov 연동 자동 측정
- **CI/CD**: GitHub Actions 자동화 (moai-gitflow.yml, publish-pypi.yml)

## TODO:SPEC-BACKLOG-001 다음 단계 SPEC 후보

### Phase 1: 핵심 안정화 (v0.3.x)

1. **SPEC-TEMPLATE-001**: Template Processor 병합 로직 고도화
   - **우선순위**: high
   - **목표**: 사용자 커스터마이징 보존율 100%

2. **SPEC-CHECKPOINT-001**: Event-Driven Checkpoint 시스템 개선
   - **우선순위**: high
   - **목표**: 위험 작업 탐지율 100%, 복구 성공률 100%

3. **SPEC-MULTI-LANG-001**: 언어별 TRUST 검증 규칙 확장
   - **우선순위**: medium
   - **목표**: 17개 언어 모두 언어별 최적 도구 자동 선택

### Phase 2: 고급 기능 (v0.4.x)

4. **SPEC-TEAM-MODE-001**: Team 모드 고급 기능
   - **우선순위**: medium
   - **목표**: PR 자동 리뷰, Squash Merge, 브랜치 정리 자동화

5. **SPEC-LIVING-DOC-001**: Living Document 고도화
   - **우선순위**: medium
   - **목표**: 자동 다이어그램 생성, 아키텍처 문서 자동 동기화

6. **SPEC-TAG-GRAPH-001**: TAG 의존성 그래프 시각화
   - **우선순위**: low
   - **목표**: `depends_on`, `blocks`, `related_specs` 그래프 자동 생성

### Phase 3: 확장성 (v0.5.x)

7. **SPEC-PLUGIN-001**: Plugin 시스템 설계
   - **우선순위**: low
   - **목표**: 사용자 정의 에이전트, 커스텀 워크플로우 지원

8. **SPEC-WEB-UI-001**: Web UI 대시보드
   - **우선순위**: low
   - **목표**: 프로젝트 전체 현황, TAG 체인, TRUST 원칙 준수도 시각화

## EARS 요구사항 작성 가이드

### EARS (Easy Approach to Requirements Syntax)

SPEC 작성 시 다음 EARS 구문을 활용하여 체계적인 요구사항을 작성하세요:

#### EARS 구문 형식

1. **Ubiquitous Requirements**: 시스템은 [기능]을 제공해야 한다
2. **Event-driven Requirements**: WHEN [조건]이면, 시스템은 [동작]해야 한다
3. **State-driven Requirements**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
4. **Optional Features**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
5. **Constraints**: IF [조건]이면, 시스템은 [제약]해야 한다

#### 적용 예시

```markdown
### Ubiquitous Requirements (기본 기능)
- 시스템은 SPEC-First TDD 워크플로우를 제공해야 한다
- 시스템은 @TAG 시스템을 통한 추적성을 보장해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 `/alfred:1-spec`을 실행하면, 시스템은 SPEC 문서를 생성해야 한다
- WHEN 위험한 작업이 감지되면, 시스템은 자동으로 checkpoint를 생성해야 한다

### State-driven Requirements (상태 기반)
- WHILE TDD 구현 중일 때, 시스템은 RED → GREEN → REFACTOR 순서를 강제해야 한다
- WHILE Personal 모드일 때, 시스템은 로컬 브랜치만 생성해야 한다

### Optional Features (선택적 기능)
- WHERE Team 모드이면, 시스템은 자동으로 PR을 생성할 수 있다
- WHERE Haiku 모드이면, 시스템은 빠른 응답을 제공할 수 있다

### Constraints (제약사항)
- IF SPEC이 없으면, 시스템은 코드 생성을 거부해야 한다
- 테스트 커버리지는 85% 이상을 유지해야 한다
- 각 함수는 50 LOC를 초과하지 않아야 한다
```

---

_이 문서는 `/alfred:1-spec` 실행 시 SPEC 생성의 기준이 됩니다._
