---
id: DOCS-002
version: 0.1.0
status: completed
created: 2025-10-11
updated: 2025-10-11
author: @Goos
priority: high
category: docs
labels:
  - documentation
  - concepts
  - user-guide
related_specs:
  - DOCS-001
scope:
  packages:
    - docs/guides/concepts
  files:
    - ears-guide.md
    - trust-principles.md
    - tag-system.md
    - spec-first-tdd.md
---

# @SPEC:DOCS-002: MoAI-ADK 핵심 개념 문서

## HISTORY

### v0.1.0 (2025-10-11)
- **COMPLETED**: TDD 구현 완료 (RED → GREEN → REFACTOR)
- **AUTHOR**: @Goos
- **REVIEW**: pending
- **ARTIFACTS**:
  - docs/guides/concepts/ears-guide.md (298 LOC)
  - docs/guides/concepts/trust-principles.md (543 LOC)
  - docs/guides/concepts/tag-system.md (622 LOC)
  - docs/guides/concepts/spec-first-tdd.md (737 LOC)
- **METRICS**:
  - 총 코드 블록: 218개
  - 상호 참조 링크: 13개
  - SPEC 수락 기준 달성률: 100% (17/17)
  - TRUST 원칙 준수율: 95%
- **TAG CHAIN**: @SPEC:DOCS-002 → @CODE:DOCS-002 (8개 TAG)

### v0.0.1 (2025-10-11)
- **INITIAL**: MoAI-ADK 핵심 개념 문서 SPEC 작성
- **AUTHOR**: @Goos
- **REVIEW**: pending
- **REASON**: 사용자가 프레임워크의 핵심 개념을 이해하고 효과적으로 활용할 수 있도록 체계적인 가이드 제공 필요

---

## Environment (환경 및 가정사항)

### 프로젝트 컨텍스트
- **프로젝트**: MoAI-ADK (MoAI Agentic Development Kit)
- **목적**: SPEC-First TDD 방법론 기반 AI 에이전틱 개발 프레임워크
- **대상 사용자**: 실무 개발자, AI 코딩 학습자, 프로젝트 리더

### 기술 환경
- **문서 형식**: Markdown (.md)
- **배포 위치**: `docs/guides/concepts/` 디렉토리
- **참조 문서**: CLAUDE.md, development-guide.md, product.md
- **버전 관리**: Git + GitHub

### 가정사항
1. 사용자는 기본적인 소프트웨어 개발 경험이 있다
2. 사용자는 TDD(Test-Driven Development) 개념에 익숙하지 않다
3. 사용자는 Git 기본 명령어를 사용할 수 있다
4. 문서는 한국어로 작성되며, 기술 용어는 영어로 표기한다

---

## Assumptions (전제 조건)

### 선행 작업
- [x] SPEC-DOCS-001 완료 (TODO App 풀스택 예제)
- [x] product.md, development-guide.md 최신 상태
- [x] docs/guides/ 디렉토리 구조 확립

### 필요 리소스
- EARS 방법론 참조 자료
- TRUST 5원칙 상세 설명
- TAG 시스템 예시 코드
- SPEC-First TDD 워크플로우 다이어그램

---

## Requirements (기능 요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 EARS 요구사항 작성 방법론 가이드를 제공해야 한다
- 시스템은 TRUST 5원칙 상세 설명 문서를 제공해야 한다
- 시스템은 TAG 시스템 사용 가이드를 제공해야 한다
- 시스템은 SPEC-First TDD 워크플로우 문서를 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 EARS 가이드를 읽으면, 시스템은 5가지 구문(Ubiquitous, Event-driven, State-driven, Optional, Constraints)의 예시를 제공해야 한다
- WHEN 사용자가 TRUST 문서를 읽으면, 시스템은 각 원칙별 언어별 구현 방법을 제공해야 한다
- WHEN 사용자가 TAG 가이드를 읽으면, 시스템은 SPEC → TEST → CODE → DOC 체인의 실제 예시를 제공해야 한다
- WHEN 사용자가 SPEC-First TDD 문서를 읽으면, 시스템은 3단계 워크플로우(1-spec, 2-build, 3-sync)의 실습 가능한 예제를 제공해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 개념 학습 중일 때, 시스템은 실습 가능한 예제 코드를 함께 제공해야 한다
- WHILE 사용자가 문서를 읽는 중일 때, 시스템은 관련 문서 간 명확한 상호 참조(cross-reference)를 제공해야 한다

### Optional Features (선택적 기능)
- WHERE 사용자가 고급 기능을 원하면, 시스템은 복잡한 시나리오의 예시를 제공할 수 있다
- WHERE 사용자가 빠른 참조를 원하면, 시스템은 치트 시트(Cheat Sheet) 섹션을 제공할 수 있다

### Constraints (제약사항)
- IF 문서가 3,000자를 초과하면, 시스템은 별도 파일로 분리해야 한다
- IF 예시 코드가 포함되면, 시스템은 해당 코드가 실제 동작 가능함을 보장해야 한다
- 모든 문서는 한국어로 작성되어야 하며, 기술 용어는 영어 원문을 병기해야 한다

---

## Specifications (상세 명세)

### 1. EARS 요구사항 작성 가이드 (ears-guide.md)

#### 목표
사용자가 EARS 방법론을 이해하고 SPEC 작성 시 적용할 수 있도록 한다.

#### 구조
1. **EARS란?**
   - 정의: Easy Approach to Requirements Syntax
   - 목적: 명확하고 구조화된 요구사항 작성
   - 장점: 애매모호함 제거, 검증 가능성 향상

2. **5가지 구문 유형**
   - Ubiquitous: 시스템은 [기능]을 제공해야 한다
   - Event-driven: WHEN [조건]이면, 시스템은 [동작]해야 한다
   - State-driven: WHILE [상태]일 때, 시스템은 [동작]해야 한다
   - Optional: WHERE [조건]이면, 시스템은 [동작]할 수 있다
   - Constraints: IF [조건]이면, 시스템은 [제약]해야 한다

3. **실제 적용 예시**
   - 인증 시스템 요구사항 (AUTH-001)
   - 파일 업로드 기능 요구사항 (UPLOAD-001)
   - 결제 시스템 요구사항 (PAYMENT-001)

4. **안티 패턴**
   - "사용자 친화적이어야 한다" (모호함)
   - "빠르게 동작해야 한다" (측정 불가)
   - "보안이 강해야 한다" (구체성 부족)

5. **베스트 프랙티스**
   - 측정 가능한 기준 명시
   - 구체적인 액션 동사 사용
   - 제약사항 명확히 정의

#### 수락 기준
- [ ] EARS 5가지 구문 각각에 대해 3개 이상의 실제 예시 제공
- [ ] 안티 패턴 → 개선안 전환 예시 5개 이상 포함
- [ ] 실습 가능한 템플릿 제공
- [ ] 다른 개념 문서와의 상호 참조 링크 포함

---

### 2. TRUST 5원칙 가이드 (trust-principles.md)

#### 목표
사용자가 TRUST 5원칙을 이해하고 코드 품질 기준으로 활용할 수 있도록 한다.

#### 구조
1. **TRUST 개요**
   - 정의: Test First, Readable, Unified, Secured, Trackable
   - 목적: AI 시대의 일관된 코드 품질 보장
   - 적용 범위: 모든 주요 프로그래밍 언어

2. **T - Test First (테스트 주도 개발)**
   - SPEC → Test → Code 사이클
   - 언어별 테스트 프레임워크:
     - Python: pytest
     - TypeScript: Vitest
     - Java: JUnit
     - Go: go test
     - Rust: cargo test
   - 테스트 커버리지 ≥85% 기준

3. **R - Readable (가독성)**
   - 함수 ≤50 LOC
   - 매개변수 ≤5개
   - 복잡도 ≤10
   - 의도 드러내는 이름 사용
   - 언어별 린터:
     - Python: ruff, black
     - TypeScript: Biome, ESLint
     - Go: gofmt, golint
     - Rust: rustfmt, clippy

4. **U - Unified (통합 아키텍처)**
   - SPEC 기반 복잡도 관리
   - 모듈화 및 경계 정의
   - 일관된 패턴 사용

5. **S - Secured (보안)**
   - SPEC 보안 요구사항 정의
   - 언어별 보안 도구:
     - Python: bandit
     - TypeScript: npm audit
     - Java: OWASP Dependency-Check
     - Go: gosec
     - Rust: cargo audit

6. **T - Trackable (추적성)**
   - @TAG 시스템 (@SPEC, @TEST, @CODE, @DOC)
   - CODE-FIRST 원칙 (코드 직접 스캔)
   - 고아 TAG 자동 탐지

#### 수락 기준
- [ ] 각 원칙별 언어별 구현 예시 제공 (Python, TypeScript, Go 필수)
- [ ] TRUST 체크리스트 템플릿 제공
- [ ] 실제 코드 리뷰 시나리오 포함
- [ ] 자동화 도구 통합 가이드 포함

---

### 3. TAG 시스템 가이드 (tag-system.md)

#### 목표
사용자가 @TAG 시스템을 이해하고 코드 추적성을 확보할 수 있도록 한다.

#### 구조
1. **TAG 시스템 개요**
   - CODE-FIRST 철학
   - TAG 라이프사이클: SPEC → TEST → CODE → DOC
   - TAG ID 규칙: `<DOMAIN>-<3자리>` (예: AUTH-001)

2. **TAG 유형별 사용법**
   - @SPEC:ID - SPEC 문서 (.moai/specs/)
   - @TEST:ID - 테스트 코드 (tests/)
   - @CODE:ID - 구현 코드 (src/)
   - @DOC:ID - 문서 (docs/)

3. **TAG 서브 카테고리**
   - @CODE:ID:API - REST API, GraphQL
   - @CODE:ID:UI - 컴포넌트, 뷰
   - @CODE:ID:DATA - 데이터 모델
   - @CODE:ID:DOMAIN - 비즈니스 로직
   - @CODE:ID:INFRA - 인프라

4. **TAG 검증 방법**
   - 중복 확인: `rg "@SPEC:AUTH" -n`
   - TAG 체인 검증: `rg '@(SPEC|TEST|CODE|DOC):' -n`
   - 고아 TAG 탐지: SPEC 없는 CODE 검색

5. **실전 예시**
   - JWT 인증 시스템 (AUTH-001)
   - 파일 업로드 기능 (UPLOAD-001)
   - 결제 시스템 (PAYMENT-001)

6. **문제 해결**
   - TAG 중복 발생 시 대응
   - 고아 TAG 발생 시 처리
   - TAG 체인 재구성

#### 수락 기준
- [ ] TAG 체계도 (다이어그램) 포함
- [ ] 실제 코드 예시 (Python, TypeScript) 제공
- [ ] rg 명령어 치트 시트 제공
- [ ] 문제 해결 시나리오 5개 이상 포함

---

### 4. SPEC-First TDD 워크플로우 가이드 (spec-first-tdd.md)

#### 목표
사용자가 MoAI-ADK의 3단계 워크플로우를 이해하고 실무에 적용할 수 있도록 한다.

#### 구조
1. **SPEC-First TDD란?**
   - 정의: 명세 우선, 테스트 주도 개발
   - 철학: "명세 없으면 코드 없다. 테스트 없으면 구현 없다."
   - Alfred SuperAgent 역할

2. **3단계 워크플로우**
   - 1단계: `/alfred:1-spec` - SPEC 작성
     - EARS 방식 요구사항 작성
     - Git 브랜치 생성 (feature/SPEC-XXX)
     - Draft PR 생성 (Team 모드)

   - 2단계: `/alfred:2-build` - TDD 구현
     - RED: 실패하는 테스트 작성
     - GREEN: 테스트 통과하는 최소 구현
     - REFACTOR: 코드 품질 개선

   - 3단계: `/alfred:3-sync` - 문서 동기화
     - Living Document 생성
     - TAG 체인 검증
     - PR Ready 전환 (Team 모드)

3. **모드별 차이점**
   - Personal 모드: 로컬 Git 워크플로우
   - Team 모드: GitHub PR 자동화

4. **실전 예제**
   - 전체 사이클 실습: TODO App 기능 추가
   - 단계별 상세 설명
   - 각 커밋별 변경 내역

5. **베스트 프랙티스**
   - SPEC 작성 시 주의사항
   - TDD 사이클 팁
   - 문서 동기화 타이밍

6. **문제 해결**
   - 테스트 실패 시 대응
   - TAG 체인 끊김 해결
   - PR 충돌 해결

#### 수락 기준
- [ ] 3단계 워크플로우 다이어그램 포함
- [ ] 전체 사이클 실습 가능한 예제 제공
- [ ] Personal/Team 모드별 차이점 명확히 설명
- [ ] 문제 해결 시나리오 3개 이상 포함
- [ ] Quick Start 가이드 링크 포함

---

## Traceability (추적성)

### 관련 SPEC
- @SPEC:DOCS-001 - TODO App 풀스택 예제 (선행 작업)
- @SPEC:PROJECT-001 - MoAI-ADK 프로젝트 정의

### 관련 문서
- @DOC:CLAUDE.md - 전체 프레임워크 개요
- @DOC:development-guide.md - 개발 가이드
- @DOC:product.md - 제품 정의

### TAG 체인
```
@SPEC:DOCS-002 (이 문서)
  ↓
@TEST:DOCS-002 (문서 검증 테스트)
  ↓
@CODE:DOCS-002 (문서 작성)
  ↓
@DOC:DOCS-002 (최종 문서)
```

---

## 비고

### 우선순위 결정 근거
- **priority: high**: 사용자가 프레임워크를 효과적으로 활용하기 위한 핵심 개념 문서
- SPEC-DOCS-001 (예제)과 함께 사용자 온보딩의 핵심 자료

### 향후 확장 계획
- SPEC-DOCS-003: 고급 기능 가이드 (에이전트 커스터마이징, 멀티 언어 지원)
- SPEC-DOCS-004: 베스트 프랙티스 케이스 스터디
- SPEC-DOCS-005: 문제 해결 FAQ 및 트러블슈팅

---

**작성자**: @Goos
**작성일**: 2025-10-11
**상태**: draft (v0.0.1)
