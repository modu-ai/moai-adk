---
id: DOCS-002
version: 0.1.0
status: implementation-complete
created: 2025-10-06
updated: 2025-10-16
author: @Goos
priority: high
category: docs
labels:
  - documentation
  - markdown
  - guides
  - commands
depends_on:
  - DOCS-001
related_specs:
  - DOCS-001
scope:
  packages:
    - docs/
  files:
    - docs/guides/architecture.md
    - docs/commands/cli.md
    - docs/security-scanning.md
---

# @SPEC:DOCS-002: VitePress Phase 2 - 핵심 개념 페이지 3개 추가

## HISTORY

### v0.1.0 (2025-10-16)
- **COMPLETED**: 문서화 완료, Python 프로젝트에 맞게 마크다운 문서 작성
- **AUTHOR**: @Goos
- **CHANGES**:
  - docs/guides/architecture.md 작성 완료
  - docs/commands/cli.md 작성 완료
  - docs/security-scanning.md 작성 완료
  - Python 프로젝트 구조에 맞게 VitePress 대신 일반 마크다운 문서로 구현
  - 핵심 콘텐츠 문서화 완료

### v0.0.1 (2025-10-06)
- **INITIAL**: 문서화 명세 최초 작성
- **AUTHOR**: @Goos
- **ORIGINAL_SCOPE**: VitePress 기반 문서 (Python 프로젝트로 전환되면서 일반 마크다운으로 변경)

---

## Environment (환경 및 전제조건)

### 실행 환경
- **VitePress**: v1.x (Vue 기반 정적 사이트 생성기)
- **프레임워크**: Vue 3, TypeScript
- **빌드 도구**: Bun 또는 npm
- **Phase 1**: 5개 기본 페이지 완료 (index, getting-started, what-is-moai-adk, faq, spec-first-tdd)

### 콘텐츠 소스
- **README.md**: 1,749줄 (Phase 2 타겟: 608~932줄, 약 324줄)
- **development-guide.md**: 390줄 (TRUST 원칙 상세)
- **기존 SPEC**: SPEC-DOCS-001 (Phase 1 구조 참조)

### 제약사항
- **콘텐츠 소스 비율**: README 60% + dev-guide 30% + 신규 10%
- **페이지당 최대 LOC**: 400줄 (권장)
- **Sidebar 확장**: concepts 섹션에 3개 항목 추가
- **Mermaid 다이어그램**: 에이전트 협업, TAG 체인 시각화

---

## Assumptions (가정사항)

1. **Phase 1 완료 가정**:
   - VitePress 설정 완료 (config.mts 존재)
   - Sidebar 기본 구조 확립 (시작하기, 핵심 개념 섹션)
   - 테스트 프레임워크 설정 완료 (Vitest)

2. **콘텐츠 소스 가정**:
   - README.md 608~932줄: Alfred 에이전트 팀 상세 설명
   - README.md 813~931줄: @TAG 시스템 CODE-FIRST 원칙
   - dev-guide.md: TRUST 5원칙 상세 (Test, Readable, Unified, Secured, Trackable)

3. **Mermaid 다이어그램 가정**:
   - VitePress Mermaid 지원 기본 활성화
   - 다이어그램 렌더링 성공 (빌드 시 에러 없음)

4. **Phase 2 우선순위 가정**:
   - Priority 1 (Critical): Alfred, TAG, TRUST 페이지 우선 작성
   - Priority 2-3: Output Styles, CLI, 문제 해결 등은 Phase 3-4로 연기

---

## Requirements (EARS 요구사항)

### Ubiquitous Requirements (기본 기능)

**UR-001**: 시스템은 VitePress docs/concepts/ 디렉토리에 3개 핵심 개념 페이지를 제공해야 한다
- **alfred-agents.md**: Alfred SuperAgent + 9개 전문 에이전트 팀 구조 및 역할
- **tag-system.md**: @TAG 추적성 시스템 (CODE-FIRST 원칙, TAG 체인)
- **trust-principles.md**: TRUST 5원칙 상세 설명 (품질 기준)

**UR-002**: 시스템은 README.md 콘텐츠를 60% 이상 반영해야 한다
- Alfred 에이전트: README 608~932줄 기반 (~324줄)
- TAG 시스템: README 813~931줄 기반 (~118줄)
- TRUST 원칙: dev-guide.md + README 분산 섹션 통합

**UR-003**: 시스템은 Mermaid 다이어그램으로 시각화된 구조를 제공해야 한다
- Alfred 오케스트레이션 구조 (graph TD)
- 에이전트 협업 흐름 (graph LR)
- TAG 체인 무결성 (graph LR)

**UR-004**: 시스템은 Sidebar 네비게이션에 "핵심 개념" 섹션을 확장해야 한다
- 기존: spec-first-tdd.md (1개)
- 추가: alfred-agents.md, tag-system.md, trust-principles.md (3개)
- 총 4개 항목

---

### Event-driven Requirements (이벤트 기반)

**ER-001**: WHEN 사용자가 "Alfred 10개 에이전트" 링크를 클릭하면, 에이전트 팀 페이지를 표시해야 한다
- **트리거**: Sidebar 또는 홈페이지에서 링크 클릭
- **응답**: alfred-agents.md 렌더링
- **구성**:
  - Alfred SuperAgent 소개
  - 9개 전문 에이전트 표 (이름, 페르소나, 역할, 호출 방법)
  - Mermaid 오케스트레이션 다이어그램
  - 에이전트 협업 원칙

**ER-002**: WHEN 사용자가 "@TAG 추적성 시스템" 링크를 클릭하면, TAG 시스템 페이지를 표시해야 한다
- **트리거**: Sidebar에서 링크 클릭
- **응답**: tag-system.md 렌더링
- **구성**:
  - CODE-FIRST 원칙 설명
  - @SPEC → @TEST → @CODE → @DOC 체인
  - 언어별 TAG 사용 예시 (TypeScript, Python, Dart)
  - TAG 검증 방법 (rg 명령어)

**ER-003**: WHEN 사용자가 "TRUST 5원칙" 링크를 클릭하면, 품질 원칙 페이지를 표시해야 한다
- **트리거**: Sidebar에서 링크 클릭
- **응답**: trust-principles.md 렌더링
- **구성**:
  - T (Test First): 테스트 커버리지 ≥85%
  - R (Readable): 함수 ≤50 LOC, 복잡도 ≤10
  - U (Unified): 아키텍처 통합성
  - S (Secured): 보안 검증 (SQL Injection, XSS, CSRF)
  - T (Trackable): @TAG 추적성

**ER-004**: WHEN 콘텐츠가 업데이트되면, 시스템은 자동으로 사이트를 재빌드해야 한다
- **트리거**: Markdown 파일 저장
- **응답**: 핫 리로드 (개발 모드), 재빌드 (프로덕션 모드)
- **시간**: < 3초

---

### State-driven Requirements (상태 기반)

**SR-001**: WHILE VitePress 개발 모드일 때, 3개 페이지 변경 감지 → 자동 재빌드
- **상태**: `bun run docs:dev` 실행 중
- **동작**: 파일 변경 감지 → 자동 재빌드 → 브라우저 리로드
- **포트**: localhost:5173 (기본)

**SR-002**: WHILE Sidebar가 열려 있을 때, "핵심 개념" 섹션에 4개 항목을 표시해야 한다
- **상태**: 사용자가 문서 페이지 열람 중
- **동작**: Sidebar에 "핵심 개념" 섹션 표시
- **항목**:
  1. SPEC-First TDD
  2. Alfred 10개 에이전트 (NEW)
  3. @TAG 추적성 시스템 (NEW)
  4. TRUST 5원칙 (NEW)

**SR-003**: WHILE 프로덕션 빌드 시, 3개 페이지를 정적 HTML로 생성해야 한다
- **상태**: `bun run docs:build` 실행
- **동작**: Markdown → HTML 변환 + Mermaid 렌더링
- **출력**: docs/.vitepress/dist/concepts/ 디렉토리

---

### Optional Features (선택적 기능)

**OF-001**: WHERE 다이어그램 렌더링 실패 시, 정적 이미지로 대체할 수 있다
- **조건**: Mermaid 렌더링 에러 발생
- **동작**: 사전 준비된 PNG/SVG 이미지로 대체
- **현재**: Phase 2에서는 미구현 (Phase 3 검토)

**OF-002**: WHERE 검색 기능 사용 시, 3개 페이지 콘텐츠도 검색 대상에 포함할 수 있다
- **조건**: VitePress 내장 검색 활성화
- **동작**: 전체 문서에서 "Alfred", "TAG", "TRUST" 키워드 검색 가능
- **현재**: Phase 1에서 이미 활성화됨

---

### Constraints (제약사항)

**C-001**: VitePress 빌드가 성공해야 한다
- **조건**: `bun run docs:build` 실행 시
- **기준**: 빌드 에러 0개, 경고 0개
- **실패 시**: Mermaid 다이어그램 문법 오류, 깨진 링크 수정 필요

**C-002**: README.md 콘텐츠 반영률이 60% 이상이어야 한다
- **검증**: Phase 2에서 README 608~932줄 (324줄) 중 60% 이상 활용
- **측정**: 3개 페이지 총 LOC 대비 README 기반 LOC 비율
- **조치**: acceptance.md에서 반영률 기록

**C-003**: 모든 내부 링크가 유효해야 한다
- **검증**: VitePress 빌드 시 링크 체크
- **형식**: [텍스트](/path/to/page) 또는 [텍스트](./relative-path)
- **실패 시**: 빌드 경고 또는 중단 (설정 기준)

**C-004**: 페이지당 최대 파일 크기는 400 LOC를 초과하지 않아야 한다 (권장)
- **이유**: 빠른 로딩 시간, 가독성
- **예외**: 복잡한 다이어그램이나 코드 예제가 많은 페이지

**C-005**: Sidebar 구조는 일관성을 유지해야 한다
- **규칙**: "핵심 개념" 섹션 내 4개 항목 순서 유지
- **순서**: 1) SPEC-First TDD, 2) Alfred, 3) TAG, 4) TRUST

---

## Traceability (@TAG)

- **SPEC**: @SPEC:DOCS-002 (이 문서)
- **TEST**: @TEST:DOCS-002
  - moai-adk-ts/__tests__/docs/vitepress-build.test.ts (빌드 검증)
  - moai-adk-ts/__tests__/docs/link-validation.test.ts (3개 페이지 존재 확인)
- **CODE**: @CODE:DOCS-002
  - docs/concepts/alfred-agents.md
  - docs/concepts/tag-system.md
  - docs/concepts/trust-principles.md
  - docs/.vitepress/config.mts (Sidebar 확장)
- **DOC**: @DOC:DOCS-002 (메타 문서: Phase 2 완료 보고서)

---

## Dependencies (의존성)

**기존 SPEC 참조**:
- **SPEC-DOCS-001**: Phase 1 완료 필수 (VitePress 기본 설정)

**외부 의존성**:
- VitePress v1.x
- Mermaid 다이어그램 렌더러 (VitePress 내장)
- Node.js ≥18 또는 Bun ≥1.2

**콘텐츠 의존성**:
- README.md (1,749줄) - 핵심 콘텐츠 소스
- development-guide.md (390줄) - TRUST 원칙 상세

---

## Success Metrics (성공 지표)

- **Phase 2 완료**: 3개 핵심 개념 페이지 작성 및 빌드 성공
- **빌드 성공률**: 100% (에러 0개)
- **링크 유효성**: 100% (깨진 링크 0개)
- **콘텐츠 소스 비율**: README 60%, dev-guide 30%, 신규 10% 준수
- **Mermaid 다이어그램**: 3개 다이어그램 렌더링 성공
- **Sidebar 확장**: "핵심 개념" 섹션 4개 항목 표시
- **테스트 통과**: 12개 이상 (기존 9개 + 신규 3개)

---

## Notes (참고사항)

- Phase 2는 README.md 비교 분석 결과 Priority 1 (Critical) 콘텐츠를 우선 추가합니다
- Mermaid 다이어그램은 VitePress에서 기본 지원되므로 추가 설정 불필요
- Phase 3-4는 Priority 2-3 콘텐츠 (Output Styles, CLI, 문제 해결 등) 추가 예정
- README 콘텐츠 비율은 Phase 1-4 전체 평균으로 30% 목표 달성 예정
